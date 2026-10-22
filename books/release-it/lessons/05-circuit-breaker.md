# 05 · Circuit breaker around control, with a TTL route cache as the degraded mode

Tag: `release-it/lesson-05` · Fixes: F09 · Verify: `make verify LESSON=05` (before: `REF=release-it/lesson-04`)
Book: Ch. 5 "Circuit Breaker"; Ch. 4 "Cascading Failures", "Integration Points"

## The idea

A **cascading failure** (Ch. 4) is one layer's failure becoming the next layer's, because callers
keep depending on a dependency that has stopped working. Timeouts make each call finite. They do
not stop callers from making the next doomed call, each costing a full timeout, and they do
nothing to protect the struggling dependency from being hammered while it tries to recover.

Chapter 5's **Circuit Breaker** puts a switch on the integration point. While calls succeed, it is
closed and they pass. When failures cross a threshold, it opens, and calls fail immediately without
touching the dependency. After a cool-down it lets one trial call through ("half-open"). Success
closes it; failure opens it again. Nygard pairs the breaker with a decision the business has to
make: what should the system do while the breaker is open? Fail, serve something cached, queue the
work? A breaker without a fallback just fails faster.

For a database platform the fallback is the product's central availability property. The control
plane decides where databases live; the data plane serves them. If control is down, the data plane
should keep serving every database it already knows how to reach. It should refuse only what
genuinely needs control: databases it has never routed, creates, drops.

## The flaw (F09, at `release-it/lesson-04`)

- Every tenant request resolves through control: `g.Routes.Resolve` in
  `internal/api/handlers.go:64` calls `internal/routes/client.go:46` every time. Nothing remembers a
  route.
- No breaker guards the calls (`client.go:98`): while control is down, each request spends control's
  500 ms budget plus any retry the budget allows.
- The gateway's health check (`cmd/gateway/health.go:63–65`) reports `failed: control unreachable`.
  That is true, because the gateway cannot serve anything without control, so during a control
  outage HAProxy takes every gateway out. The data plane is down because the control plane is.

## The patch

Two commits.

**Part 1: the breaker** (`internal/breaker/breaker.go`, ~150 lines). It opens after 5 consecutive
failures, or when at least 50 % of at least 20 calls in a 10 s window fail (`:69`). It moves to
half-open after a 5 s cool-down and lets exactly one trial call through; success closes it, failure
reopens it. Only failures that say something about control count: transport errors and 5xx, not
`401` or `404` answers. Every control call except the usage report goes through it
(`internal/routes/client.go:159`). The report is slow by nature and off the critical path, so it
must not trip the breaker for the data plane (`:139`). An open breaker becomes
`503 control_unavailable` with `Retry-After: 5` in microseconds (`handlers.go:238`). Resolve retries
stop as soon as the breaker is open. Metrics: `breaker_state{dep}`,
`breaker_transitions_total{dep,from,to}`, `breaker_rejected_total{dep}`.

**Part 2: the degraded mode** (`internal/routes/cache.go`). Each gateway caches
`(sha256(api_key), db_name) → {tenant, database, node, epoch}` for `TTL` = 60 s (`:20`).
`Resolve` (`client.go:68`) answers from the cache first. A hit on an entry older than
`RefreshAfter` = 10 s triggers one background refresh (`client.go:94`), so busy routes stay fresh
without a control call per request. A failed refresh leaves the entry serving until its TTL; an
answer saying the key or database is gone evicts it. While control is down, known routes keep
working until their entries expire, and nothing new is resolved. This is the route-grant idea from
the ts-db planes design (existing grants serve until their TTL; no new grants) at baby size.

A cached route is a hint, not the truth. The node still checks the epoch on every request (it has
since v0). If the database moved, the node answers `409 stale_epoch`, and the gateway evicts the
entry, resolves again and retries once (`handlers.go:78–87`). That is safe even for writes, because
the node refused before running anything. `TestStaleCachedRouteIsReresolved` covers it.

Gateway health changes to match (`cmd/gateway/health.go:67`): with control unreachable but routes
cached, the gateway is `degraded` ("serving N cached routes") and stays in HAProxy. It is `failed`
only when it has nothing to serve. Metrics: `route_cache_entries`, `route_cache_hits_total`,
`route_cache_misses_total`.

## Verify

```
make verify LESSON=05 REF=release-it/lesson-04      # before
make verify LESSON=05                               # after
```

Steady mix at 50 % of fleet capacity (300 req/s) for 60 s, plus 5 req/s to databases outside the
steady set (`cold`) and one create per second. After a 10 s warm-up, the control proxy accepts and
never answers from t=10 s to t=40 s (`timeout` toxic, upstream). The script records the highest
breaker state across the gateways and HAProxy's count of UP gateways every 0.5 s.

Before (`lesson-04`):

```
  known databases served during the outage (t=12-40 s)   0.0 % (p50 0.9 ms)
  cold databases and creates during the outage           157 of 157 answered 503, p50 0.8 ms
  HAProxy gateways UP during the outage (min)            0 of 3
FLAW REPRODUCED: outage_success_pct 0.0 (want >= 95); breaker_closed_after_s NaN (want <= 8); gateways_up_during_outage 0 (want == 3)
```

After:

```
  known databases served during the outage (t=12-40 s)   99.25 % (p50 6.8 ms)
  cold databases and creates during the outage           126 of 156 answered 503, p50 3.3 ms
  calls the gateways made to control, t=15-40 s          15 (one trial per gateway per 5 s cool-down: 15)
  HAProxy gateways UP during the outage (min)            3 of 3
  breaker transitions closed>open 3, open>half-open 18, half-open>closed 3
  all breakers closed again after the outage ended      4.8 s
IMPROVED: outage_success_pct 99.25 >= 95; outage_p50_ms 6.8 <= 10; miss_p50_ms 4.6 <= 10; control_calls_during_outage 15 <= 20; breaker_closed_after_s 4.8 <= 8; gateways_up_during_outage 3 == 3
```

Before the patch, the outage was total and fast. Within two health checks every gateway reported
`control unreachable` and HAProxy answered 503 for everyone. The plan expected each request to burn
its 500 ms on control, and that does happen, but only for the two seconds before HAProxy gives up on
the gateways. After the patch, tenants whose routes were cached (almost all of them after a 10 s
warm-up) did not notice the outage. Cold databases and creates failed in milliseconds with a
`Retry-After`. Control saw one trial call per gateway every five seconds, and each breaker closed on
its first trial after control came back.

The 1 % that failed during the outage are entries that expired. A tenant first seen at the start
of the warm-up has a 60 s TTL that ends inside a 30 s outage, and its refresh cannot reach control.
With a TTL longer than the longest control outage you want to ride out, that would be zero; a
longer TTL also means a longer window in which a moved database is first reached at its old node.
The epoch check is what makes that window safe.

## Try changing

- Set `TTL` to 20 s and rerun. Known routes start failing 20 s into the outage. Then try 5 minutes,
  and think about what a moved database would cost during those 5 minutes without the epoch check.
- Make the gateway's health `failed` whenever control is unreachable (the lesson-04 rule) while
  keeping the cache. The cache can serve, but HAProxy no longer sends it traffic. A health check is
  part of the degraded mode too.
- Route the usage report through the breaker and run lesson 6's `reports` scenario. Twenty slow
  reports open the breaker for the critical path.

## Read more

*Release It!* 2nd ed., Ch. 5 "Circuit Breaker"; Ch. 4 "Cascading Failures". Martin Fowler's
"CircuitBreaker" article. `sony/gobreaker` and `failsafe-go` are what you would use at work; the
state machine is hand-written here because it is the lesson. The ts-db planes design
(`tsdb-planes-fable`) for route grants with a TTL.

## Deviations from the plan

- **Two commits, about 540 lines with tests and verify (plan ~190).** The breaker and its tests are
  ~230; the cache, the stale-epoch path and the health change are the rest.
- **Control calls during the outage: ≤ 20, counted from t=15 s,** not ≤ 10 from t=10 s. Three
  gateways each send one trial per 5 s cool-down, which is 15 in 25 s. The failures that open the
  breaker, plus the calls in flight at that moment, made 41 when counted from t=12 s.
- **Background refresh** happens on a hit once an entry is older than 10 s, deduplicated per
  entry. The plan says "refreshed on hit in the background"; refreshing on *every* hit would be a
  control call per request again, just asynchronous.
- The fault is the `timeout` toxic on the **upstream** stream, removed by cycling the proxy (see
  lessons 2 and 4).
