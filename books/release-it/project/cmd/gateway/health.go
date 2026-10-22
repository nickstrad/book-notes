package main

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"litefleet/internal/breaker"
	"litefleet/internal/httpx"
	"litefleet/internal/routes"
)

// prober checks control and every node once a second, so /health answers
// from fresh state without making the balancer wait on a dependency.
type prober struct {
	client  *http.Client
	control string
	nodes   map[string]string
	routes  *routes.Client

	mu   sync.Mutex
	down map[string]bool
}

func newProber(control, nodeSpec string, rc *routes.Client) *prober {
	p := &prober{client: &http.Client{Timeout: time.Second}, control: control, nodes: map[string]string{}, routes: rc, down: map[string]bool{}}
	for _, kv := range strings.Split(nodeSpec, ",") {
		if name, url, ok := strings.Cut(strings.TrimSpace(kv), "="); ok {
			p.nodes[name] = url
		}
	}
	return p
}

func (p *prober) run() {
	for {
		p.probe("control", p.control)
		for name, url := range p.nodes {
			p.probe(name, url)
		}
		time.Sleep(time.Second)
	}
}

func (p *prober) probe(name, base string) {
	resp, err := p.client.Get(base + "/health")
	ok := err == nil && resp.StatusCode == http.StatusOK
	if err == nil {
		resp.Body.Close()
	}
	p.mu.Lock()
	p.down[name] = !ok
	p.mu.Unlock()
}

func (p *prober) isDown(name string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.down[name]
}

// The gateway cannot serve anything without control: every request resolves
// through it. An open breaker is reported as it happens.
func (p *prober) controlCheck() (string, string) {
	switch {
	case p.isDown("control"):
		return httpx.Failed, "control unreachable"
	case p.routes.BreakerState() != breaker.Closed:
		return httpx.Degraded, "control breaker " + p.routes.BreakerState().String()
	}
	return httpx.OK, ""
}

// A node being down costs its tenants, not the gateway.
func (p *prober) nodesCheck() (string, string) {
	var down []string
	for name := range p.nodes {
		if p.isDown(name) {
			down = append(down, name+" unreachable")
		}
	}
	if len(down) == 0 {
		return httpx.OK, ""
	}
	sort.Strings(down)
	return httpx.Degraded, strings.Join(down, ", ")
}
