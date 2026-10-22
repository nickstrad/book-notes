// Package routes is the gateway's client for the control plane: it resolves
// a tenant's API key and database name to the node that owns the database,
// and forwards database lifecycle calls.
package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"litefleet/internal/httpx"
	"litefleet/internal/retry"
	"litefleet/internal/wire"
)

// Client talks to control. Control answers from an indexed registry in
// milliseconds, so 500 ms is generous; the usage report is a heavy read and
// gets its own client.
type Client struct {
	base    string
	http    *http.Client
	report  *http.Client
	resolve retry.Policy
}

// New returns a client for the control plane at base.
func New(base string) *Client {
	return &Client{
		base:   base,
		http:   httpx.NewClient(httpx.Timeouts{Dial: 200 * time.Millisecond, Header: 500 * time.Millisecond, Total: 500 * time.Millisecond}),
		report: httpx.NewClient(httpx.Timeouts{Dial: 200 * time.Millisecond, Header: 10 * time.Second, Total: 10 * time.Second}),
		// A resolve is a read and may be retried, within the same limits
		// as node reads.
		resolve: retry.Policy{Dep: "control", Base: 50 * time.Millisecond, Cap: 400 * time.Millisecond, MaxRetries: 2,
			MinAttempt: 100 * time.Millisecond, Budget: retry.NewBudget(0.1, 10)},
	}
}

// Resolve authenticates apiKey and locates its database dbName.
func (c *Client) Resolve(ctx context.Context, apiKey, dbName string) (wire.Route, error) {
	var rt wire.Route
	body, _ := json.Marshal(wire.ResolveRequest{APIKey: apiKey, DBName: dbName})
	err := c.resolve.Do(ctx, retryable, func(ctx context.Context) error {
		return c.do(ctx, http.MethodPost, "/internal/resolve", "", body, &rt)
	})
	return rt, err
}

// Create asks control to place and provision a new database.
func (c *Client) Create(ctx context.Context, apiKey, name string) (wire.Database, error) {
	var d wire.Database
	body, _ := json.Marshal(wire.CreateDatabaseRequest{Name: name})
	err := c.do(ctx, http.MethodPost, "/internal/databases", apiKey, body, &d)
	return d, err
}

// List returns the tenant's databases.
func (c *Client) List(ctx context.Context, apiKey string) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.do(ctx, http.MethodGet, "/internal/databases", apiKey, nil, &out)
	return out, err
}

// Stats returns one database's registry record.
func (c *Client) Stats(ctx context.Context, apiKey, name string) (wire.Database, error) {
	var d wire.Database
	err := c.do(ctx, http.MethodGet, "/internal/databases/"+url.PathEscape(name), apiKey, nil, &d)
	return d, err
}

// Drop deletes a database.
func (c *Client) Drop(ctx context.Context, apiKey, name string) error {
	return c.do(ctx, http.MethodDelete, "/internal/databases/"+url.PathEscape(name), apiKey, nil, nil)
}

// Usage fetches the fleet usage report (admin key).
func (c *Client) Usage(ctx context.Context, apiKey, window string) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.doWith(ctx, c.report, http.MethodGet, "/internal/usage?window="+url.QueryEscape(window), apiKey, nil, &out)
	return out, err
}

// retryable: transport errors and control-side 5xx; a 4xx is an answer.
func retryable(err error) bool {
	var re *httpx.RemoteError
	if errors.As(err, &re) {
		return re.Status >= 500
	}
	return true
}

func (c *Client) do(ctx context.Context, method, path, apiKey string, body []byte, out any) error {
	return c.doWith(ctx, c.http, method, path, apiKey, body, out)
}

func (c *Client) doWith(ctx context.Context, client *http.Client, method, path, apiKey string, body []byte, out any) (err error) {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	httpx.Propagate(ctx, req)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	start, status := time.Now(), 0
	defer func() { httpx.Dependency(ctx, "control", start, status, err) }()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("control: %w", err)
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("control: %w", err)
	}
	if resp.StatusCode >= 300 {
		var eb wire.ErrorBody
		if json.Unmarshal(data, &eb) != nil || eb.Error.Code == "" {
			return &httpx.RemoteError{Status: http.StatusBadGateway, Code: "control_error", Message: resp.Status}
		}
		return &httpx.RemoteError{Status: resp.StatusCode, Code: eb.Error.Code, Message: eb.Error.Message}
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}
