// Package blob is the node's client for the blob store that keeps database
// snapshots.
package blob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"litefleet/internal/httpx"
)

// Client uploads snapshots to the blob store.
type Client struct {
	base string
	http *http.Client
}

// New returns a client for the blob store at base. An upload of a large
// database legitimately takes a while, so the total is 5 minutes; that is
// also a floor on bandwidth: the 60 MB whale must move at 200 KB/s or more.
func New(base string) *Client {
	return &Client{base: base, http: httpx.NewClient(httpx.Timeouts{Dial: 200 * time.Millisecond, Header: time.Second, Total: 5 * time.Minute})}
}

// PutResult is the blob store's answer to an upload.
type PutResult struct {
	Key   string `json:"key"`
	Bytes int64  `json:"bytes"`
}

// Put uploads the file at path as generation gen of database dbID.
func (c *Client) Put(dbID string, gen int, path string) (_ *PutResult, err error) {
	start, status := time.Now(), 0
	defer func() { httpx.Dependency(context.Background(), "blobstore", start, status, err) }()
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/b/%s/%d.db", c.base, dbID, gen), f)
	if err != nil {
		return nil, err
	}
	req.ContentLength = fi.Size()
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, httpx.Transport("blobstore", err)
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	// The status first: an error page is not a bad answer, it is an error.
	if resp.StatusCode != http.StatusOK {
		return nil, httpx.ReadError(resp)
	}
	if !httpx.IsJSON(resp) {
		return nil, httpx.BadPayload("blobstore", fmt.Errorf("answer is %q, not JSON", resp.Header.Get("Content-Type")))
	}
	var out PutResult
	if err := json.NewDecoder(io.LimitReader(resp.Body, httpx.MaxErrorBody)).Decode(&out); err != nil {
		return nil, httpx.BadPayload("blobstore", err)
	}
	return &out, nil
}
