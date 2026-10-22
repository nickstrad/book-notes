// Package metering ships usage events from a node to the metering partner.
package metering

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"litefleet/internal/httpx"
	"litefleet/internal/wire"
)

// Client posts batches of usage events.
type Client struct {
	base string
	http *http.Client
}

// NewClient returns a client for the metering partner at base.
func NewClient(base string) *Client {
	// Metering answers a batch in ~10 ms; a slow partner gets 2 s for the
	// whole exchange (headers included) before a batch counts as failed and
	// goes to the outbox. A shorter header timeout than the partner's latency
	// turns every delivered batch into a "failure" and a duplicate.
	return &Client{base: base, http: httpx.NewClient(httpx.Timeouts{Dial: 200 * time.Millisecond, Header: 2 * time.Second, Total: 2 * time.Second})}
}

// Send posts one batch.
func (c *Client) Send(events []wire.UsageEvent) (err error) {
	start, status := time.Now(), 0
	defer func() { httpx.Dependency(context.Background(), "metering", start, status, err) }()
	body, err := json.Marshal(events)
	if err != nil {
		return err
	}
	resp, err := c.http.Post(c.base+"/events", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("metering: %s", resp.Status)
	}
	return nil
}
