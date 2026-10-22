package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"litefleet/internal/wire"
)

// Route cache settings: an entry serves for TTL after control last confirmed
// it, and a hit on an entry older than RefreshAfter refreshes it in the
// background. While control is unreachable, known databases keep working
// until their entries expire; nothing new is resolved.
const (
	TTL          = 60 * time.Second
	RefreshAfter = 10 * time.Second
)

var (
	cacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "route_cache_hits_total", Help: "Resolves answered from the route cache.",
	})
	cacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "route_cache_misses_total", Help: "Resolves that had to ask control.",
	})
	cacheEntries = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "route_cache_entries", Help: "Routes in the cache.",
	})
)

// cache maps (api key hash, database name) to a route. A cached route is a
// hint: the node still checks the epoch, and a stale one is evicted.
type cache struct {
	ttl time.Duration

	mu sync.Mutex
	m  map[string]*entry
}

type entry struct {
	rt         wire.Route
	fetched    time.Time
	refreshing bool
}

func newCache(ttl time.Duration) *cache {
	c := &cache{ttl: ttl, m: map[string]*entry{}}
	go c.sweep()
	return c
}

func cacheKey(apiKey, dbName string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(sum[:]) + "/" + dbName
}

// SetCacheTTL changes how long a cached route serves (default TTL).
func (c *Client) SetCacheTTL(ttl time.Duration) {
	c.cache.mu.Lock()
	c.cache.ttl = ttl
	c.cache.mu.Unlock()
}

// get returns a live entry and whether the caller should refresh it (the
// first caller to see it old enough gets true).
func (c *cache) get(k string) (wire.Route, bool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[k]
	if !ok || time.Since(e.fetched) > c.ttl {
		cacheMisses.Inc()
		return wire.Route{}, false, false
	}
	cacheHits.Inc()
	refresh := !e.refreshing && time.Since(e.fetched) > RefreshAfter
	if refresh {
		e.refreshing = true
	}
	return e.rt, refresh, true
}

func (c *cache) put(k string, rt wire.Route) {
	c.mu.Lock()
	c.m[k] = &entry{rt: rt, fetched: time.Now()}
	cacheEntries.Set(float64(len(c.m)))
	c.mu.Unlock()
}

// refreshFailed lets the next hit try again; the entry keeps serving until
// its TTL.
func (c *cache) refreshFailed(k string) {
	c.mu.Lock()
	if e, ok := c.m[k]; ok {
		e.refreshing = false
	}
	c.mu.Unlock()
}

func (c *cache) evict(k string) {
	c.mu.Lock()
	delete(c.m, k)
	cacheEntries.Set(float64(len(c.m)))
	c.mu.Unlock()
}

func (c *cache) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.m)
}

func (c *cache) sweep() {
	for range time.Tick(c.ttl / 2) {
		c.mu.Lock()
		for k, e := range c.m {
			if time.Since(e.fetched) > c.ttl {
				delete(c.m, k)
			}
		}
		cacheEntries.Set(float64(len(c.m)))
		c.mu.Unlock()
	}
}
