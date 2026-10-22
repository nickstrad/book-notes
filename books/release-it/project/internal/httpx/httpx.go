// Package httpx is the HTTP plumbing shared by gateway, control and node:
// server construction, request logging and metrics, JSON helpers and typed
// errors.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"litefleet/internal/wire"
)

// NewServer returns the HTTP server a litefleet service listens with. A
// client that sends its headers slowly, never finishes its body or never
// reads the answer is cut off instead of holding a goroutine and a socket.
func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// SetupLogging sends every log line, including the standard library's, to
// stdout as JSON tagged with the instance name.
func SetupLogging(instance string) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("instance", instance))
}

type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyFields
)

// fields collects attributes a handler adds to its request's log line.
type fields struct {
	mu    sync.Mutex
	attrs []any
}

// Annotate adds key=value to the request's log line (tenant and database
// ids belong in logs, never in metric labels).
func Annotate(ctx context.Context, key string, value any) {
	if f, ok := ctx.Value(keyFields).(*fields); ok {
		f.mu.Lock()
		f.attrs = append(f.attrs, key, value)
		f.mu.Unlock()
	}
}

// RequestIDFrom returns the request id carried by ctx.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(keyRequestID).(string)
	return id
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

// Logged gives every request an id (taken from X-Request-Id or generated),
// counts and times it by route, and writes one JSON log line when it ends.
func Logged(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := RequestID(r)
		w.Header().Set("X-Request-Id", id)
		f := &fields{}
		ctx := context.WithValue(context.WithValue(r.Context(), keyRequestID, id), keyFields, f)
		r2 := r.WithContext(ctx)
		sw := &statusWriter{ResponseWriter: w}
		inflight.Inc()
		defer inflight.Dec() // also when a streaming handler aborts
		// Count and log in a defer: a handler that cuts its answer off
		// panics with http.ErrAbortHandler, and that request must be
		// counted and logged like any other. The panic is passed on, so
		// the server still aborts the connection.
		defer func() {
			p := recover()
			finish(r2, sw, id, f, time.Since(start), p != nil)
			if p != nil {
				panic(p)
			}
		}()
		h.ServeHTTP(sw, r2)
	})
}

func finish(r *http.Request, sw *statusWriter, id string, f *fields, d time.Duration, aborted bool) {
	if sw.status == 0 {
		sw.status = http.StatusOK
	}
	route := r.Pattern
	if route == "" {
		route = "unmatched"
	}
	requests.WithLabelValues(route, r.Method, strconv.Itoa(sw.status)).Inc()
	duration.WithLabelValues(route).Observe(d.Seconds())
	if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
		return
	}
	attrs := []any{"request_id", id, "method", r.Method, "path", r.URL.Path, "route", route,
		"status", sw.status, "duration_ms", ms(d)}
	if aborted { // the status went out, but the answer was cut off
		attrs = append(attrs, "aborted", true)
	}
	f.mu.Lock()
	attrs = append(attrs, f.attrs...)
	f.mu.Unlock()
	slog.Info("request", attrs...)
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

// RequestID returns the request's id, taken from X-Request-Id or generated.
func RequestID(r *http.Request) string {
	if id := RequestIDFrom(r.Context()); id != "" {
		return id
	}
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
	Annotate(r.Context(), "error_code", code)
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
