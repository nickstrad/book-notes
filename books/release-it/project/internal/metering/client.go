// Package metering ships usage events from a node to the metering partner.
package metering

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"litefleet/internal/wire"
)

// Client posts batches of usage events.
type Client struct {
	base string
	http *http.Client
}

// NewClient returns a client for the metering partner at base.
func NewClient(base string) *Client {
	return &Client{base: base, http: &http.Client{}}
}

// Send posts one batch.
func (c *Client) Send(events []wire.UsageEvent) error {
	body, err := json.Marshal(events)
	if err != nil {
		return err
	}
	resp, err := c.http.Post(c.base+"/events", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("metering: %s", resp.Status)
	}
	return nil
}
