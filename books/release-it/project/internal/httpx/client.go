package httpx

import (
	"net"
	"net/http"
	"time"
)

// Timeouts bound each phase of a call to another service: connecting,
// waiting for the response headers once the request is sent, and the whole
// exchange including reading the body.
type Timeouts struct {
	Dial, Header, Total time.Duration
}

// NewClient returns an HTTP client that gives up on a dependency at each
// timeout instead of waiting for it forever.
func NewClient(t Timeouts) *http.Client {
	return &http.Client{
		Timeout: t.Total,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: t.Dial, KeepAlive: 30 * time.Second}).DialContext,
			TLSHandshakeTimeout:   t.Dial,
			ResponseHeaderTimeout: t.Header,
			MaxIdleConns:          1000,
			MaxIdleConnsPerHost:   1000,
			IdleConnTimeout:       60 * time.Second,
		},
	}
}
