// Package nodeclient is the gateway's client for node agents: it forwards a
// tenant's statement to the node that owns the database.
package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"litefleet/internal/httpx"
	"litefleet/internal/retry"
	"litefleet/internal/wire"
)

// Client forwards statements to nodes.
type Client struct {
	http  *http.Client
	reads retry.Policy
}

// New returns a node client. A node runs a statement for at most 2 s (its
// statement deadline), so an attempt gets 2.5 s for headers and body.
//
// Reads are idempotent and may be retried: at most twice, after a backoff
// drawn from [0, 50 ms·2^n] (capped at 400 ms), only with time left in the
// request's deadline, and only while this gateway's retries stay under 10 %
// of its node calls. Writes are never retried: after a timeout the gateway
// cannot know whether the transaction committed, and SQL in general cannot
// be de-duplicated, so a repeat could apply it twice.
func New() *Client {
	return &Client{
		http: httpx.NewClient(httpx.Timeouts{Dial: 200 * time.Millisecond, Header: 2500 * time.Millisecond, Total: 2500 * time.Millisecond}),
		reads: retry.Policy{Base: 50 * time.Millisecond, Cap: 400 * time.Millisecond, MaxRetries: 2,
			MinAttempt: 100 * time.Millisecond, Budget: retry.NewBudget(0.1, 10)},
	}
}

// Query starts a read on the route's node and returns once the node has
// begun its answer with the columns. The rows follow through Rows.WriteTo,
// one at a time, so neither side holds a whole result in memory. Retries
// happen only before an answer has begun.
func (c *Client) Query(ctx context.Context, rt wire.Route, req wire.QueryRequest) (*Rows, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var rows *Rows
	p := c.reads
	p.Dep = rt.Node
	err = p.Do(ctx, transient, func(ctx context.Context) error {
		var err error
		rows, err = c.openQuery(ctx, rt, body)
		return err
	})
	return rows, err
}

// Execute runs a write transaction on the route's node.
func (c *Client) Execute(ctx context.Context, rt wire.Route, req wire.ExecuteRequest) (*wire.ExecuteResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var out wire.ExecuteResult
	if err := c.do(ctx, rt, "execute", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// transient reports whether a failed read is worth repeating: a timeout, an
// unreachable node and node-side 5xx are, except a statement that ran out its
// deadline (it would again); a 4xx is the request's own fault, and a bad
// payload will be bad again.
func transient(err error) bool {
	if re, ok := err.(*httpx.RemoteError); ok {
		return re.Status >= 500 && re.Code != "statement_deadline" && re.Code != "draining"
	}
	return !errors.Is(err, httpx.ErrBadPayload)
}

func (c *Client) do(ctx context.Context, rt wire.Route, op string, body []byte, out any) error {
	resp, err := c.send(ctx, rt, op, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, httpx.MaxErrorBody))
	if err != nil {
		return httpx.Transport(rt.Node, err)
	}
	if err := decode(data, out); err != nil {
		return httpx.BadPayload(rt.Node, fmt.Errorf("bad rows: %w", err))
	}
	return nil
}

// send posts to the node and returns its response if it is a 200 with a JSON
// body; any other status becomes a *RemoteError read from at most 1 MB of the
// error body, and anything that is not JSON a bad payload.
func (c *Client) send(ctx context.Context, rt wire.Route, op string, body []byte) (_ *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rt.NodeURL+"/internal/db/"+rt.DatabaseID+"/"+op, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Epoch", strconv.Itoa(rt.Epoch))
	httpx.Propagate(ctx, req)
	start, status := time.Now(), 0
	defer func() { httpx.Dependency(ctx, rt.Node, start, status, err) }()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, httpx.Transport(rt.Node, err)
	}
	status = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, httpx.ReadError(resp)
	}
	if !httpx.IsJSON(resp) {
		resp.Body.Close()
		return nil, httpx.BadPayload(rt.Node, fmt.Errorf("answer is %q, not JSON", resp.Header.Get("Content-Type")))
	}
	return resp, nil
}

// decode reads a node response exactly: an unexpected field means the node
// and gateway disagree about the wire format.
func decode(data []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	dec.UseNumber()
	return dec.Decode(out)
}
