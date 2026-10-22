package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONResponseCompleteness(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		valid          bool
	}{
		{"exact with whitespace", " \n {\"big\":9007199254740993,\"nested\":[0.12345678901234567890123456789]} \t\n", true},
		{"malformed", `{"big":}`, false},
		{"garbage", `{"big":1}xyz`, false},
		{"second object", `{"big":1}{"big":2}`, false},
		{"second scalar", "{\"big\":1}\ntrue", false},
		{"empty", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tc.response))
			}))
			defer srv.Close()
			var out map[string]any
			err := (&Client{URL: srv.URL}).JSON("GET", "/response", nil, &out)
			if !tc.valid {
				if err == nil {
					t.Fatal("accepted an incomplete JSON document")
				}
				if out != nil {
					t.Fatalf("decoded a prefix on error: %#v", out)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if out["big"] != json.Number("9007199254740993") || out["nested"].([]any)[0] != json.Number("0.12345678901234567890123456789") {
				t.Fatalf("lost exact numbers: %#v", out)
			}
		})
	}
}
