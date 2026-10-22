package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct{ URL, Key string }

func Env() *Client {
	url := os.Getenv("QUEUELET_URL")
	if url == "" {
		url = "http://localhost:18200"
	}
	key := os.Getenv("QUEUELET_KEY")
	if key == "" {
		key = "acme-key"
	}
	return &Client{url, key}
}
func (c *Client) Do(method, path string, data any) ([]byte, int, error) {
	var body io.Reader
	if data != nil {
		b, e := json.Marshal(data)
		if e != nil {
			return nil, 0, e
		}
		body = bytes.NewReader(b)
	}
	req, e := http.NewRequest(method, c.URL+path, body)
	if e != nil {
		return nil, 0, e
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	resp, e := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if e != nil {
		return nil, 0, e
	}
	defer resp.Body.Close()
	b, e := io.ReadAll(resp.Body)
	return b, resp.StatusCode, e
}
func (c *Client) JSON(method, path string, data, out any) error {
	b, status, e := c.Do(method, path, data)
	if e != nil {
		return e
	}
	if status >= 400 {
		return fmt.Errorf("HTTP %d: %s", status, b)
	}
	return json.Unmarshal(b, out)
}
