// Package nodeclient is the gateway's client for node agents: it forwards a
// tenant's statement to the node that owns the database.
package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
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

// Query runs a read statement on the route's node.
func (c *Client) Query(ctx context.Context, rt wire.Route, req wire.QueryRequest) (*wire.QueryResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	var out wire.QueryResult
	p := c.reads
	p.Dep = rt.Node
	if err := p.Do(ctx, transient, func(ctx context.Context) error { return c.do(ctx, rt, "query", body, &out) }); err != nil {
		return nil, err
	}
	return &out, nil
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

// transient reports whether a failed read is worth repeating: network errors
// and node-side 5xx are, except a statement that ran out its deadline (it
// would again); a 4xx is the request's own fault.
func transient(err error) bool {
	if re, ok := err.(*httpx.RemoteError); ok {
		return re.Status >= 500 && re.Code != "statement_deadline"
	}
	return true
}

func (c *Client) do(ctx context.Context, rt wire.Route, op string, body []byte, out any) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rt.NodeURL+"/internal/db/"+rt.DatabaseID+"/"+op, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Epoch", strconv.Itoa(rt.Epoch))
	httpx.Propagate(ctx, req)
	start, status := time.Now(), 0
	defer func() { httpx.Dependency(ctx, rt.Node, start, status, err) }()
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("node %s: %w", rt.Node, err)
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("node %s: %w", rt.Node, err)
	}
	if resp.StatusCode != http.StatusOK {
		var eb wire.ErrorBody
		if err := decode(data, &eb); err != nil {
			return fmt.Errorf("node %s: bad rows: %w", rt.Node, err)
		}
		return &httpx.RemoteError{Status: resp.StatusCode, Code: eb.Error.Code, Message: eb.Error.Message}
	}
	if err := decode(data, out); err != nil {
		return fmt.Errorf("node %s: bad rows: %w", rt.Node, err)
	}
	return nil
}

// decode reads a node response exactly: an unexpected field means the node
// and gateway disagree about the wire format.
func decode(data []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	dec.UseNumber()
	return dec.Decode(out)
}
