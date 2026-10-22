// Package nodeclient is the gateway's client for node agents: it forwards a
// tenant's statement to the node that owns the database.
package nodeclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"litefleet/internal/httpx"
	"litefleet/internal/wire"
)

// Client forwards statements to nodes.
type Client struct {
	http *http.Client
}

// New returns a node client.
func New() *Client {
	return &Client{http: &http.Client{}}
}

// Query runs a read statement on the route's node.
func (c *Client) Query(rt wire.Route, req wire.QueryRequest) (*wire.QueryResult, error) {
	var out wire.QueryResult
	if err := c.call(rt, "query", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Execute runs a write transaction on the route's node.
func (c *Client) Execute(rt wire.Route, req wire.ExecuteRequest) (*wire.ExecuteResult, error) {
	var out wire.ExecuteResult
	if err := c.call(rt, "execute", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Nodes are reached over the network and blip; retry transient failures.
const (
	retries    = 5
	retryDelay = 100 * time.Millisecond
)

func (c *Client) call(rt wire.Route, op string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	for attempt := 0; ; attempt++ {
		err = c.do(rt, op, body, out)
		if err == nil || !transient(err) || attempt == retries {
			return err
		}
		time.Sleep(retryDelay)
	}
}

// transient reports whether a failed call is worth repeating: network
// errors and node-side 5xx are; a 4xx is the request's own fault.
func transient(err error) bool {
	if re, ok := err.(*httpx.RemoteError); ok {
		return re.Status >= 500
	}
	return true
}

func (c *Client) do(rt wire.Route, op string, body []byte, out any) (err error) {
	req, err := http.NewRequest(http.MethodPost, rt.NodeURL+"/internal/db/"+rt.DatabaseID+"/"+op, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Epoch", strconv.Itoa(rt.Epoch))
	start, status := time.Now(), 0
	defer func() { httpx.Dependency(rt.Node, start, status, err) }()
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
