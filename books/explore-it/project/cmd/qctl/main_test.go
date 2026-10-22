package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestQctlJSON(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "qctl")
	build := exec.Command("go", "build", "-o", exe, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build qctl: %v\n%s", err, out)
	}
	run := func(t *testing.T, url string, args ...string) (string, string, error) {
		t.Helper()
		t.Setenv("QUEUELET_URL", url)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, exe, args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}
	const exact = `{"big":9007199254740993,"nested":[0.12345678901234567890123456789,{"id":9223372036854775807}]}`
	var requests atomic.Int32
	var lastBody atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		b, _ := io.ReadAll(r.Body)
		lastBody.Store(b)
		io.WriteString(w, " \n"+exact+"\t\n")
	}))
	defer srv.Close()
	for _, tc := range []struct {
		option string
		args   []string
		field  string
	}{
		{"payload", []string{"enqueue", "emails"}, "payload"},
		{"result", []string{"job", "ack", "1"}, "result"},
		{"json", []string{"batch", "emails"}, ""},
		{"json", []string{"schedule", "create"}, ""},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			args := append(append([]string{}, tc.args...), "--"+tc.option, " \n"+exact+" \t")
			stdout, stderr, err := run(t, srv.URL, args...)
			if err != nil {
				t.Fatalf("valid options: %v: %s", err, stderr)
			}
			wire := lastBody.Load().([]byte)
			if tc.field != "" {
				var body map[string]json.RawMessage
				if err := json.Unmarshal(wire, &body); err != nil {
					t.Fatal(err)
				}
				wire = body[tc.field]
			}
			if string(wire) != exact {
				t.Fatalf("wire numbers changed: %s", wire)
			}
			var compact bytes.Buffer
			if err := json.Compact(&compact, []byte(stdout)); err != nil || compact.String() != exact {
				t.Fatalf("display changed: %q (%v)", stdout, err)
			}
			for _, invalid := range []string{"", " \t", `{"big":}`, exact + "xyz", exact + `{}`, exact + "\ntrue"} {
				before := requests.Load()
				args[len(args)-1] = invalid
				stdout, stderr, err := run(t, srv.URL, args...)
				if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
					t.Fatalf("invalid %q: expected CLI exit 1, got %v", invalid, err)
				}
				if !strings.Contains(stderr, "--"+tc.option+": invalid JSON") || stdout != "" {
					t.Fatalf("invalid %q: stdout=%q stderr=%q", invalid, stdout, stderr)
				}
				if requests.Load() != before {
					t.Fatalf("invalid %q sent a request", invalid)
				}
			}
		})
	}
	t.Run("omitted options", func(t *testing.T) {
		_, stderr, err := run(t, srv.URL, "enqueue", "emails")
		if err != nil {
			t.Fatalf("defaults: %v: %s", err, stderr)
		}
		if body := string(lastBody.Load().([]byte)); body != `{"kind":"echo","priority":5}` {
			t.Fatalf("changed defaults: %s", body)
		}
	})
	for _, result := range []string{"null", "true", `"text"`, "9007199254740993", "0.12345678901234567890123456789", "[1,null]"} {
		t.Run("result "+result, func(t *testing.T) {
			_, stderr, err := run(t, srv.URL, "job", "ack", "1", "--result", result)
			if err != nil {
				t.Fatalf("valid result: %v: %s", err, stderr)
			}
			if body := string(lastBody.Load().([]byte)); body != `{"result":`+result+`}` {
				t.Fatalf("result changed: %s", body)
			}
		})
	}
	for _, response := range []string{`{"id":1}xyz`, "{\"id\":1}\n{\"id\":2}", `{"id":}`} {
		t.Run("raw response "+response, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, response)
			}))
			defer server.Close()
			stdout, stderr, err := run(t, server.URL, "job", "get", "1")
			if err != nil || stdout != response {
				t.Fatalf("must retain complete raw response: stdout=%q stderr=%q err=%v", stdout, stderr, err)
			}
		})
	}
}
