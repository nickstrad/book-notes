// Package blob is the node's client for the blob store that keeps database
// snapshots.
package blob

import (
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

// New returns a client for the blob store at base.
func New(base string) *Client {
	return &Client{base: base, http: &http.Client{}}
}

// PutResult is the blob store's answer to an upload.
type PutResult struct {
	Key   string `json:"key"`
	Bytes int64  `json:"bytes"`
}

// Put uploads the file at path as generation gen of database dbID.
func (c *Client) Put(dbID string, gen int, path string) (_ *PutResult, err error) {
	start, status := time.Now(), 0
	defer func() { httpx.Dependency("blobstore", start, status, err) }()
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
		return nil, err
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out PutResult
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("blobstore: bad response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blobstore: %s", resp.Status)
	}
	return &out, nil
}
