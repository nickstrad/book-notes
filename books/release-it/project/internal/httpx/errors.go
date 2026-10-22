package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
)

// The ways a call to another service can fail, so callers handle classes of
// failure rather than strings. A *CallError carries one of these and the
// underlying error (errors.Is and errors.As see both); *RemoteError is a
// typed answer from the other service.
var (
	ErrTimeout     = errors.New("timeout")
	ErrUnreachable = errors.New("unreachable")
	ErrBadPayload  = errors.New("bad payload")
)

// CallError is a failed call to dep.
type CallError struct {
	Dep   string
	Class error // ErrTimeout, ErrUnreachable or ErrBadPayload
	Err   error
}

func (e *CallError) Error() string   { return e.Dep + ": " + e.Class.Error() + ": " + e.Err.Error() }
func (e *CallError) Unwrap() []error { return []error{e.Class, e.Err} }

// Kind names the class: "timeout", "unreachable" or "bad_payload".
func (e *CallError) Kind() string {
	switch e.Class {
	case ErrTimeout:
		return "timeout"
	case ErrUnreachable:
		return "unreachable"
	}
	return "bad_payload"
}

// Kind classifies a typed answer: "stale_epoch", "remote_5xx" or "remote_4xx".
func (e *RemoteError) Kind() string {
	switch {
	case e.Code == "stale_epoch":
		return "stale_epoch"
	case e.Status >= 500:
		return "remote_5xx"
	}
	return "remote_4xx"
}

// Transport classifies an error from sending a request or reading its
// answer: out of time, or not reachable at all.
func Transport(dep string, err error) error {
	var ne net.Error
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &ne) && ne.Timeout() {
		return &CallError{Dep: dep, Class: ErrTimeout, Err: err}
	}
	return &CallError{Dep: dep, Class: ErrUnreachable, Err: err}
}

// BadPayload is an answer that does not parse as what was promised.
func BadPayload(dep string, err error) error {
	return &CallError{Dep: dep, Class: ErrBadPayload, Err: err}
}

// MaxErrorBody is the most of an error answer a client reads: enough for any
// typed error, not enough for a partner's 10 MB stack trace to cost memory.
const MaxErrorBody = 1 << 20

// ReadError turns a non-2xx answer into a *RemoteError. It reads at most
// MaxErrorBody, takes the typed error from a JSON body, and from anything
// else (an HTML error page from a proxy) only the status.
func ReadError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, MaxErrorBody))
	if IsJSON(resp) {
		var eb struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(data, &eb) == nil && eb.Error.Code != "" {
			return &RemoteError{Status: resp.StatusCode, Code: eb.Error.Code, Message: eb.Error.Message}
		}
	}
	return &RemoteError{Status: resp.StatusCode, Code: "upstream_error", Message: resp.Status}
}

// IsJSON reports whether the answer says it is JSON.
func IsJSON(resp *http.Response) bool {
	mt, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	return err == nil && mt == "application/json"
}
