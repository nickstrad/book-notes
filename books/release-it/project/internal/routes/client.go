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

	"litefleet/internal/breaker"
	"litefleet/internal/httpx"
	"litefleet/internal/retry"
	"litefleet/internal/wire"
)

// Client talks to control. Control answers from an indexed registry in
// milliseconds, so 500 ms is generous; the usage report is a heavy read and
// gets its own client.
//
// Every call except the usage report goes through a circuit breaker: when
// control keeps failing, callers get breaker.ErrOpen at once instead of
// spending their deadline on it. Resolves are answered from a TTL route
// cache first, so the data plane keeps serving known databases while
// control is down.
type Client struct {
	base    string
	http    *http.Client
	admin   *http.Client // creates and drops provision on a node: control's budget for them is 2 s
	report  *http.Client
	resolve retry.Policy
	breaker *breaker.Breaker
	cache   *cache
}

// New returns a client for the control plane at base.
func New(base string) *Client {
	return &Client{
		base:   base,
		http:   httpx.NewClient(httpx.Timeouts{Dial: 200 * time.Millisecond, Header: 500 * time.Millisecond, Total: 500 * time.Millisecond}),
		admin:  httpx.NewClient(httpx.Timeouts{Dial: 200 * time.Millisecond, Header: 2500 * time.Millisecond, Total: 2500 * time.Millisecond}),
		report: httpx.NewClient(httpx.Timeouts{Dial: 200 * time.Millisecond, Header: 10 * time.Second, Total: 10 * time.Second}),
		// A resolve is a read and may be retried, within the same limits
		// as node reads.
		resolve: retry.Policy{Dep: "control", Base: 50 * time.Millisecond, Cap: 400 * time.Millisecond, MaxRetries: 2,
			MinAttempt: 100 * time.Millisecond, Budget: retry.NewBudget(0.1, 10)},
		breaker: breaker.New("control"),
		cache:   newCache(TTL),
	}
}

// BreakerState is the state of the breaker around control.
func (c *Client) BreakerState() breaker.State { return c.breaker.State() }

// CachedRoutes is the number of routes the gateway could serve without
// control.
func (c *Client) CachedRoutes() int { return c.cache.len() }

// Evict drops a cached route, after the node said its epoch is stale.
func (c *Client) Evict(apiKey, dbName string) { c.cache.evict(cacheKey(apiKey, dbName)) }

// Resolve authenticates apiKey and locates its database dbName, from the
// cache when it can.
func (c *Client) Resolve(ctx context.Context, apiKey, dbName string) (wire.Route, error) {
	k := cacheKey(apiKey, dbName)
	if rt, refresh, ok := c.cache.get(k); ok {
		if refresh {
			go c.refresh(k, apiKey, dbName)
		}
		return rt, nil
	}
	rt, err := c.resolveRemote(ctx, apiKey, dbName)
	if err == nil {
		c.cache.put(k, rt)
	}
	return rt, err
}

func (c *Client) resolveRemote(ctx context.Context, apiKey, dbName string) (wire.Route, error) {
	var rt wire.Route
	body, _ := json.Marshal(wire.ResolveRequest{APIKey: apiKey, DBName: dbName})
	err := c.resolve.Do(ctx, retryable, func(ctx context.Context) error {
		return c.do(ctx, http.MethodPost, "/internal/resolve", "", body, &rt)
	})
	return rt, err
}

// refresh re-resolves a cached route in the background. A key or database
// that no longer exists is evicted; anything else leaves the entry serving.
func (c *Client) refresh(k, apiKey, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rt, err := c.resolveRemote(ctx, apiKey, dbName)
	var re *httpx.RemoteError
	switch {
	case err == nil:
		c.cache.put(k, rt)
	case errors.As(err, &re) && re.Status < 500:
		c.cache.evict(k)
	default:
		c.cache.refreshFailed(k)
	}
}

// Create asks control to place and provision a new database.
func (c *Client) Create(ctx context.Context, apiKey, name string) (wire.Database, error) {
	var d wire.Database
	body, _ := json.Marshal(wire.CreateDatabaseRequest{Name: name})
	err := c.breaker.Call(func() error {
		return c.doWith(ctx, c.admin, http.MethodPost, "/internal/databases", apiKey, body, &d)
	}, retryable)
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
	return c.breaker.Call(func() error {
		return c.doWith(ctx, c.admin, http.MethodDelete, "/internal/databases/"+url.PathEscape(name), apiKey, nil, nil)
	}, retryable)
}

// Usage fetches the fleet usage report (admin key).
func (c *Client) Usage(ctx context.Context, apiKey, window string) (json.RawMessage, error) {
	var out json.RawMessage
	// Not on the critical path, and slow by nature: it bypasses the breaker
	// so a heavy report cannot trip it for the data plane.
	err := c.doWith(ctx, c.report, http.MethodGet, "/internal/usage?window="+url.QueryEscape(window), apiKey, nil, &out)
	return out, err
}

// retryable: transport errors and control-side 5xx; a 4xx is an answer, and
// an open breaker means control is known to be down.
func retryable(err error) bool {
	if errors.Is(err, breaker.ErrOpen) {
		return false
	}
	var re *httpx.RemoteError
	if errors.As(err, &re) {
		return re.Status >= 500
	}
	return true
}

func (c *Client) do(ctx context.Context, method, path, apiKey string, body []byte, out any) error {
	return c.breaker.Call(func() error {
		return c.doWith(ctx, c.http, method, path, apiKey, body, out)
	}, retryable)
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
