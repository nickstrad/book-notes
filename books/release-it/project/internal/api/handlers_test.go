package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"litefleet/internal/httpx"
	"litefleet/internal/nodeclient"
	"litefleet/internal/routes"
	"litefleet/internal/wire"
)

// stubs returns a gateway wired to an httptest control plane and node.
func stubs(t *testing.T) http.Handler {
	t.Helper()
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, wire.QueryResult{Columns: []string{"n"}, Rows: [][]any{{42}}})
	}))
	t.Cleanup(node.Close)
	control := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req wire.ResolveRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.APIKey != "good" {
			httpx.WriteError(w, r, http.StatusUnauthorized, "unauthorized", "unknown api key")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, wire.Route{TenantID: "t1", DatabaseID: "db1", NodeURL: node.URL, Node: "node-a", Epoch: 1})
	}))
	t.Cleanup(control.Close)
	g := &Gateway{Routes: routes.New(control.URL), Nodes: nodeclient.New()}
	mux := http.NewServeMux()
	g.Register(mux)
	return mux
}

func post(h http.Handler, key, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestQueryForwardsToNode(t *testing.T) {
	rec := post(stubs(t), "good", "/v1/databases/default/query", `{"sql":"SELECT 42 AS n"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"rows":[[42]]`) {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestUnknownKey(t *testing.T) {
	rec := post(stubs(t), "bad", "/v1/databases/default/query", `{"sql":"SELECT 1"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestMissingKey(t *testing.T) {
	rec := post(stubs(t), "", "/v1/databases/default/query", `{"sql":"SELECT 1"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestEmptyStatementsRejected(t *testing.T) {
	rec := post(stubs(t), "good", "/v1/databases/default/execute", `{"statements":[]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}
