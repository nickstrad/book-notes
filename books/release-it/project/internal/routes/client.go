// Package routes is the gateway's client for the control plane: it resolves
// a tenant's API key and database name to the node that owns the database,
// and forwards database lifecycle calls.
package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"litefleet/internal/httpx"
	"litefleet/internal/wire"
)

// Client talks to control.
type Client struct {
	base string
	http *http.Client
}

// New returns a client for the control plane at base.
func New(base string) *Client {
	return &Client{base: base, http: &http.Client{}}
}

// Resolve authenticates apiKey and locates its database dbName.
func (c *Client) Resolve(apiKey, dbName string) (wire.Route, error) {
	var rt wire.Route
	body, _ := json.Marshal(wire.ResolveRequest{APIKey: apiKey, DBName: dbName})
	err := c.do(http.MethodPost, "/internal/resolve", "", body, &rt)
	return rt, err
}

// Create asks control to place and provision a new database.
func (c *Client) Create(apiKey, name string) (wire.Database, error) {
	var d wire.Database
	body, _ := json.Marshal(wire.CreateDatabaseRequest{Name: name})
	err := c.do(http.MethodPost, "/internal/databases", apiKey, body, &d)
	return d, err
}

// List returns the tenant's databases.
func (c *Client) List(apiKey string) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.do(http.MethodGet, "/internal/databases", apiKey, nil, &out)
	return out, err
}

// Stats returns one database's registry record.
func (c *Client) Stats(apiKey, name string) (wire.Database, error) {
	var d wire.Database
	err := c.do(http.MethodGet, "/internal/databases/"+url.PathEscape(name), apiKey, nil, &d)
	return d, err
}

// Drop deletes a database.
func (c *Client) Drop(apiKey, name string) error {
	return c.do(http.MethodDelete, "/internal/databases/"+url.PathEscape(name), apiKey, nil, nil)
}

// Usage fetches the fleet usage report (admin key).
func (c *Client) Usage(apiKey, window string) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.do(http.MethodGet, "/internal/usage?window="+url.QueryEscape(window), apiKey, nil, &out)
	return out, err
}

func (c *Client) do(method, path, apiKey string, body []byte, out any) (err error) {
	req, err := http.NewRequest(method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	start, status := time.Now(), 0
	defer func() { httpx.Dependency("control", start, status, err) }()
	resp, err := c.http.Do(req)
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
