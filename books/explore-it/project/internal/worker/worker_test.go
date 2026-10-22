package worker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"queuelet/internal/client"
	"testing"
	"time"
)

func TestDecodedWorkerDefaults(t *testing.T) {
	for _, tc := range []struct {
		name, kind, payload string
		attempt             int
		wantFailure         bool
	}{
		{"sleep omitted", "sleep", `{}`, 1, false},
		{"sleep wrong type", "sleep", `{"seconds":"not a number"}`, 1, false},
		{"sleep numeric", "sleep", `{"seconds":0}`, 1, false},
		{"fail3 omitted", "fail3", `{}`, 3, true},
		{"fail3 exhausted", "fail3", `{}`, 4, false},
		{"fail3 wrong type", "fail3", `{"n":true}`, 3, true},
		{"fail3 zero default", "fail3", `{"n":0}`, 3, true},
		{"fail omitted", "fail", `{}`, 1, false},
		{"fail numeric", "fail", `{"n":1}`, 1, true},
		{"fail numeric exhausted", "fail", `{"n":1}`, 2, false},
		{"echo exact", "echo", `{"big":9007199254740993,"nested":[0.12345678901234567890123456789]}`, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"payload":` + tc.payload + `}`))
			}))
			defer srv.Close()
			var job Job
			if err := (&client.Client{URL: srv.URL}).JSON("GET", "/job", nil, &job); err != nil {
				t.Fatal(err)
			}
			job.Kind, job.Attempt = tc.kind, tc.attempt
			result, err := perform(context.Background(), job)
			if (err != nil) != tc.wantFailure {
				t.Fatalf("failure=%v, want %v", err, tc.wantFailure)
			}
			if !tc.wantFailure {
				b, err := json.Marshal(result)
				if err != nil || string(b) != tc.payload {
					t.Fatalf("result=%s err=%v, want %s", b, err, tc.payload)
				}
			}
		})
	}
}

func TestCancellationInterruptsSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := perform(ctx, Job{Kind: "sleep", Payload: map[string]any{"seconds": json.Number("60")}})
		done <- err
	}()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("sleep continued")
	}
}
func TestCancellationInterruptsHTTPBody(t *testing.T) {
	started := make(chan struct{})
	stopped := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(" "))
		w.(http.Flusher).Flush()
		close(started)
		<-r.Context().Done()
		close(stopped)
	}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := perform(ctx, Job{Kind: "http", Payload: map[string]any{"url": srv.URL}})
		done <- err
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("HTTP body continued")
	}
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("remote connection remained open")
	}
}
