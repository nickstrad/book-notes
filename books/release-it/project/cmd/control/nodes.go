package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"litefleet/internal/wire"
)

// nodeCommands sends lifecycle commands (provision, drop) to node agents
// directly, not through the data-plane proxies.
type nodeCommands struct {
	urls map[string]string
	http *http.Client
}

func newNodeCommands(spec string) *nodeCommands {
	n := &nodeCommands{urls: map[string]string{}, http: &http.Client{}}
	for _, kv := range strings.Split(spec, ",") {
		if name, url, ok := strings.Cut(strings.TrimSpace(kv), "="); ok {
			n.urls[name] = url
		}
	}
	return n
}

func (n *nodeCommands) provision(node string, req wire.ProvisionRequest) error {
	body, _ := json.Marshal(req)
	return n.send(node, http.MethodPost, "/internal/provision", body)
}

func (n *nodeCommands) drop(node, dbID string) error {
	return n.send(node, http.MethodDelete, "/internal/db/"+dbID, nil)
}

func (n *nodeCommands) send(node, method, path string, body []byte) error {
	base, ok := n.urls[node]
	if !ok {
		return fmt.Errorf("unknown node %q", node)
	}
	req, err := http.NewRequest(method, base+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: %s %s", method, path, resp.Status, bytes.TrimSpace(msg))
	}
	return nil
}
