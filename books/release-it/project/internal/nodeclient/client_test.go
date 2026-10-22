package nodeclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"litefleet/internal/httpx"
	"litefleet/internal/wire"
)

// stubNode answers like a node agent, in about a millisecond.
func stubNode(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /internal/db/{id}/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Epoch") != "3" {
			httpx.WriteError(w, r, http.StatusConflict, "stale_epoch", "stale epoch")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, wire.QueryResult{Columns: []string{"id"}, Rows: [][]any{{1}, {2}}})
	})
	mux.HandleFunc("POST /internal/db/{id}/execute", func(w http.ResponseWriter, r *http.Request) {
		var req wire.ExecuteRequest
		json.NewDecoder(r.Body).Decode(&req)
		httpx.WriteJSON(w, http.StatusOK, wire.ExecuteResult{RowsAffected: int64(len(req.Statements)), LastInsertRowID: 7})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestQuery(t *testing.T) {
	srv := stubNode(t)
	rows, err := New().Query(context.Background(), wire.Route{NodeURL: srv.URL, DatabaseID: "db1", Epoch: 3}, wire.QueryRequest{SQL: "SELECT id FROM t"})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := rows.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	var res wire.QueryResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil || len(res.Rows) != 2 || res.Columns[0] != "id" {
		t.Fatalf("result = %s (%v)", buf.String(), err)
	}
}

func TestExecute(t *testing.T) {
	srv := stubNode(t)
	res, err := New().Execute(context.Background(), wire.Route{NodeURL: srv.URL, DatabaseID: "db1", Epoch: 3},
		wire.ExecuteRequest{Statements: []wire.Statement{{SQL: "INSERT 1"}, {SQL: "INSERT 2"}}})
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsAffected != 2 || res.LastInsertRowID != 7 {
		t.Fatalf("result = %+v", res)
	}
}

func TestTypedErrorPassesThrough(t *testing.T) {
	srv := stubNode(t)
	_, err := New().Query(context.Background(), wire.Route{NodeURL: srv.URL, DatabaseID: "db1", Epoch: 1}, wire.QueryRequest{SQL: "SELECT 1"})
	re, ok := err.(*httpx.RemoteError)
	if !ok || re.Status != http.StatusConflict || re.Code != "stale_epoch" {
		t.Fatalf("err = %#v; want 409 stale_epoch", err)
	}
}

// A node that speaks a newer wire version adds fields; the gateway reads
// what it needs and passes on every first-version field.
func TestReadsANewerNodeAnswer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"columns":["id"],"column_types":["INTEGER"],"rows":[[1],[2]],"truncated":false,"stats":{"rows_read":2,"bytes":40}}`))
	}))
	defer srv.Close()
	rows, err := New().Query(context.Background(), wire.Route{NodeURL: srv.URL, DatabaseID: "db1", Epoch: 7}, wire.QueryRequest{SQL: "SELECT id FROM t"})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := rows.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	var got struct {
		wire.QueryResult
		NodeEpoch int `json:"node_epoch"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil || len(got.Rows) != 2 || got.Columns[0] != "id" || got.NodeEpoch != 7 {
		t.Fatalf("answer %s (%v)", buf.String(), err)
	}
}
