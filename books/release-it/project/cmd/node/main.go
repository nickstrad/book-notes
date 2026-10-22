// Command node is a litefleet node agent: it owns a set of tenant SQLite
// files, runs tenant SQL routed to it by the gateways, snapshots databases to
// the blob store, ships usage to metering and heartbeats to control.
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"litefleet/internal/backup"
	"litefleet/internal/blob"
	"litefleet/internal/engine"
	"litefleet/internal/httpx"
	"litefleet/internal/metering"
	"litefleet/internal/wire"
)

type node struct {
	name   string
	url    string
	data   string
	eng    *engine.Engine
	backup *backup.Job
	usage  *metering.Queue
	self   *sql.DB
}

func main() {
	n := &node{
		name: env("NODE_NAME", "node-a"),
		url:  env("NODE_URL", "http://node-a:8070"),
		data: env("DATA_DIR", "/data"),
	}
	httpx.SetupLogging(n.name)
	self, err := sql.Open("sqlite3", "file:"+filepath.Join(n.data, "_node.db")+"?_pragma=journal_mode(wal)")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := self.Exec(`CREATE TABLE IF NOT EXISTS node_info (key text PRIMARY KEY, value text)`); err != nil {
		log.Fatal(err)
	}
	n.self = self

	n.usage = metering.NewQueue(metering.NewClient(env("METERING_URL", "http://metering:8082")))
	n.eng, err = engine.New(filepath.Join(n.data, "tenants"), n.usage)
	if err != nil {
		log.Fatal(err)
	}
	interval, _ := time.ParseDuration(env("BACKUP_INTERVAL", "60s"))
	tmp := filepath.Join(n.data, "tmp")
	os.MkdirAll(tmp, 0o755)
	n.backup = backup.New(n.eng, blob.New(env("BLOBSTORE_URL", "http://blobstore:8081")), interval, tmp)

	n.exportMetrics()
	go n.usage.Run()
	go n.backup.Run()
	go n.heartbeats(env("CONTROL_URL", "http://control:8090"))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /internal/db/{id}/query", n.query)
	mux.HandleFunc("POST /internal/db/{id}/execute", n.execute)
	mux.HandleFunc("POST /internal/provision", n.provision)
	mux.HandleFunc("DELETE /internal/db/{id}", n.drop)
	mux.HandleFunc("POST /internal/backup/{id}", n.backupNow)
	mux.HandleFunc("GET /health", httpx.HealthHandler(map[string]httpx.Check{
		"canary": n.canaryCheck, "fds": n.fdCheck, "disk": n.diskCheck, "wal": n.walCheck,
	}))
	mux.Handle("GET /metrics", httpx.MetricsHandler())

	addr := env("LISTEN", ":8070")
	log.Printf("%s listening on %s", n.name, addr)
	log.Fatal(httpx.NewServer(addr, httpx.Logged(mux)).ListenAndServe())
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func epoch(r *http.Request) int {
	e, _ := strconv.Atoi(r.Header.Get("X-Epoch"))
	return e
}

func (n *node) query(w http.ResponseWriter, r *http.Request) {
	var req wire.QueryRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	httpx.Annotate(r.Context(), "database_id", r.PathValue("id"))
	_, body, err := n.eng.Query(r.PathValue("id"), epoch(r), req)
	if err != nil {
		writeEngineError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (n *node) execute(w http.ResponseWriter, r *http.Request) {
	var req wire.ExecuteRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	httpx.Annotate(r.Context(), "database_id", r.PathValue("id"))
	res, err := n.eng.Execute(r.PathValue("id"), epoch(r), req)
	if err != nil {
		writeEngineError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (n *node) provision(w http.ResponseWriter, r *http.Request) {
	var req wire.ProvisionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil || req.DatabaseID == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "bad_request", "bad provision request")
		return
	}
	if err := n.eng.Provision(req); err != nil {
		writeEngineError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (n *node) drop(w http.ResponseWriter, r *http.Request) {
	if err := n.eng.Drop(r.PathValue("id")); err != nil {
		writeEngineError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// backupNow is the ops endpoint the scheduler's call is also exposed as.
func (n *node) backupNow(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	gen, err := n.backup.Backup(r.PathValue("id"))
	if err != nil {
		writeEngineError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"generation": gen, "seconds": time.Since(start).Seconds()})
}

func writeEngineError(w http.ResponseWriter, r *http.Request, err error) {
	var se *engine.StatementError
	switch {
	case errors.As(err, &se):
		httpx.WriteError(w, r, se.Status, se.Code, se.Message)
	case errors.Is(err, engine.ErrNoSuchDatabase):
		httpx.WriteError(w, r, http.StatusNotFound, "no_such_database", err.Error())
	case errors.Is(err, engine.ErrExists):
		httpx.WriteError(w, r, http.StatusConflict, "exists", err.Error())
	case errors.Is(err, engine.ErrStaleEpoch):
		httpx.WriteError(w, r, http.StatusConflict, "stale_epoch", err.Error())
	default:
		log.Printf("engine: %v", err)
		httpx.WriteError(w, r, http.StatusInternalServerError, "internal", err.Error())
	}
}

// heartbeats reports liveness, fds, disk and per-database activity to
// control every 5 seconds.
func (n *node) heartbeats(controlURL string) {
	client := &http.Client{}
	for range time.Tick(5 * time.Second) {
		hb := wire.Heartbeat{Node: n.name, URL: n.url, FDs: openFDs(), DiskFree: diskFree(n.data), Databases: n.eng.Report()}
		body, _ := json.Marshal(hb)
		resp, err := client.Post(controlURL+"/internal/heartbeat", "application/json", bytes.NewReader(body))
		if err != nil {
			log.Printf("heartbeat: %v", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("heartbeat: %s", resp.Status)
		}
	}
}

func openFDs() int {
	ents, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return -1
	}
	return len(ents)
}

func diskFree(path string) int64 {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return -1
	}
	return int64(st.Bavail) * int64(st.Bsize)
}
