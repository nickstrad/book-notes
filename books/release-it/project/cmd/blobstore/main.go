// Command blobstore is an honest fake of an S3-shaped object store used by
// nodes for database snapshots. It accepts PUT/GET/LIST under /b/, streams
// objects to disk at a capacity of about 50 MB/s, keeps the last three
// generations per database and exposes Prometheus metrics. Network-level
// misbehaviour is injected by Toxiproxy in front of it, never in here.
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	puts     = promauto.NewCounterVec(prometheus.CounterOpts{Name: "blobstore_puts_total", Help: "PUT requests by outcome."}, []string{"outcome"})
	putBytes = promauto.NewCounter(prometheus.CounterOpts{Name: "blobstore_put_bytes_total", Help: "Bytes stored."})
	inflight = promauto.NewGauge(prometheus.GaugeOpts{Name: "blobstore_inflight_puts", Help: "PUTs in progress."})
	objects  = promauto.NewGauge(prometheus.GaugeOpts{Name: "blobstore_objects", Help: "Objects stored."})
)

const bytesPerSecond = 50 << 20

var root = "/blobs"

func main() {
	if v := os.Getenv("BLOB_DIR"); v != "" {
		root = v
	}
	os.MkdirAll(root, 0o755)
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /b/{db}/{object}", put)
	mux.HandleFunc("GET /b/{db}/{object}", get)
	mux.HandleFunc("GET /b/{db}/", list)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	log.Printf("blobstore listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

func objectPath(db, object string) (string, bool) {
	if strings.Contains(db, "/") || strings.Contains(object, "/") || strings.HasPrefix(db, ".") || !strings.HasSuffix(object, ".db") {
		return "", false
	}
	return filepath.Join(root, db, object), true
}

// throttled reads at the store's capacity.
type throttled struct {
	r     io.Reader
	start time.Time
	n     int64
}

func (t *throttled) Read(p []byte) (int, error) {
	if len(p) > 256<<10 {
		p = p[:256<<10]
	}
	n, err := t.r.Read(p)
	t.n += int64(n)
	if due := time.Duration(float64(t.n) / bytesPerSecond * float64(time.Second)); due > time.Since(t.start) {
		time.Sleep(due - time.Since(t.start))
	}
	return n, err
}

func put(w http.ResponseWriter, r *http.Request) {
	inflight.Inc()
	defer inflight.Dec()
	path, ok := objectPath(r.PathValue("db"), r.PathValue("object"))
	if !ok {
		puts.WithLabelValues("bad_key").Inc()
		http.Error(w, `{"error":{"code":"bad_key","message":"bad object key"}}`, http.StatusBadRequest)
		return
	}
	os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := path + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		puts.WithLabelValues("error").Inc()
		http.Error(w, `{"error":{"code":"internal","message":"create failed"}}`, http.StatusInternalServerError)
		return
	}
	n, err := io.Copy(f, &throttled{r: r.Body, start: time.Now()})
	f.Close()
	if err != nil {
		os.Remove(tmp)
		puts.WithLabelValues("aborted").Inc()
		return
	}
	os.Rename(tmp, path)
	puts.WithLabelValues("ok").Inc()
	putBytes.Add(float64(n))
	prune(filepath.Dir(path))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"key": r.PathValue("db") + "/" + r.PathValue("object"), "bytes": n})
}

// prune keeps the newest three generations of a database.
func prune(dir string) {
	ents, _ := os.ReadDir(dir)
	var gens []int
	for _, e := range ents {
		if g, err := strconv.Atoi(strings.TrimSuffix(e.Name(), ".db")); err == nil {
			gens = append(gens, g)
		}
	}
	sort.Ints(gens)
	for len(gens) > 3 {
		os.Remove(filepath.Join(dir, strconv.Itoa(gens[0])+".db"))
		gens = gens[1:]
	}
	count := 0
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".db") {
			count++
		}
		return nil
	})
	objects.Set(float64(count))
}

func get(w http.ResponseWriter, r *http.Request) {
	path, ok := objectPath(r.PathValue("db"), r.PathValue("object"))
	if !ok {
		http.Error(w, "bad key", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, path)
}

func list(w http.ResponseWriter, r *http.Request) {
	ents, _ := os.ReadDir(filepath.Join(root, filepath.Base(r.PathValue("db"))))
	names := []string{}
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".db") {
			names = append(names, e.Name())
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"objects": names})
}
