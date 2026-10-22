// Package httpx is the HTTP plumbing shared by gateway, control and node:
// server construction, request logging and metrics, JSON helpers and typed
// errors.
package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"litefleet/internal/wire"
)

// NewServer returns the HTTP server a litefleet service listens with.
func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{Addr: addr, Handler: h}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// Logged counts and times every request by route, and logs it after it has
// been handled.
func Logged(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		inflight.Inc()
		h.ServeHTTP(sw, r)
		inflight.Dec()
		if sw.status == 0 {
			sw.status = http.StatusOK
		}
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		requests.WithLabelValues(route, r.Method, strconv.Itoa(sw.status)).Inc()
		duration.WithLabelValues(route).Observe(time.Since(start).Seconds())
		log.Printf("handled %s %s", r.Method, r.URL.Path)
	})
}

// RequestID returns the request's id, taken from X-Request-Id or generated.
func RequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-Id"); id != "" {
		return id
	}
	id := NewID()
	r.Header.Set("X-Request-Id", id)
	return id
}

// NewID returns a random 16-hex-digit identifier.
func NewID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// WriteJSON writes v as a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a typed error: {"error": {"code", "message", "request_id"}}.
func WriteError(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	WriteJSON(w, status, wire.ErrorBody{Error: wire.Error{Code: code, Message: msg, RequestID: RequestID(r)}})
}

// DecodeJSON decodes the request body into v.
func DecodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// Bearer returns the API key from "Authorization: Bearer <key>".
func Bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}

// RemoteError is a typed error returned by another litefleet service.
type RemoteError struct {
	Status  int
	Code    string
	Message string
}

func (e *RemoteError) Error() string {
	return e.Code + ": " + e.Message
}
