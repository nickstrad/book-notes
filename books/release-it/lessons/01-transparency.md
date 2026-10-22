# 01 · Transparency: health, metrics, structured logs

Tag: `release-it/lesson-01` · Fixes: F01 · Verify: `make verify LESSON=01` (before: `REF=release-it/v0-flawed`)
Book: Ch. 8 "Transparency" (designing for transparency, logging, instance metrics, health checks);
Ch. 10 "System-Wide Transparency"; Ch. 6 case study

## The idea

You cannot fix what you cannot see. Chapter 6 of *Release It!* is a case study of a retailer whose
site struggled through its busiest season while the team had few vital signs to go on. They
diagnosed it from scraps and guesses, because the system said almost nothing about itself. Chapter 8
names the remedy **transparency**: a running instance should show its state from the outside, in
three ways.

- **Instance metrics.** Numbers that change when the instance's situation changes: requests,
  latencies, pool usage, queue depths, and above all the time spent waiting on each integration
  point.
- **Logs.** Written for the people who read them during an incident: consistent, searchable, and
  saying what happened and where.
- **Health checks.** More than "the process is alive". A health check should report whether this
  instance can do its work right now, including the state of the things it depends on, so a load
  balancer can act on it.

Chapter 10 widens the view from one instance to the system. Metrics and logs are useful once they
are collected in one place where you can compare instances and follow one request through several
of them.

This lesson comes first, before the book's Part I order, because every later lesson is measured
through what it adds.

## The flaw (F01, at `release-it/v0-flawed`)

- `internal/httpx/httpx.go:73`: `Health` answers `200 {"status":"ok"}` whatever is going on.
  Gateway, control and node all use it (`cmd/*/main.go`).
- No service has a `/metrics` endpoint. `deploy/prometheus.yml` was written by ops to scrape every
  service, so Prometheus shows `up{job="gateway"} == 0`: the targets exist and have nothing to say.
- `internal/httpx/httpx.go:25`: the only log line is `handled POST /v1/databases/default/query`.
  It has no request id, no status, no duration, and no hint of which dependency a slow request
  waited on.
- `deploy/haproxy.cfg:25–31`: HAProxy checks gateways with a bare TCP connect, so a gateway that
  accepts connections counts as healthy.

Run the lesson's timeline on v0 and this is all an operator gets. At t=15 s node-a gets 800 ms of
latency: half the tenants slow down, and nothing says why. At t=35 s Postgres is paused for 10 s:
every request fails (each resolves through control), while HAProxy shows 3 of 3 gateways UP and
every `/health` says `ok`.

## The patch

Two commits. Part 1 adds metrics and health, part 2 adds logs. Both stay inside
`internal/httpx`, three health implementations, and one-line hooks in the clients.

**Metrics** (`internal/httpx/metrics.go`). `Logged` (`httpx.go:83`) counts and times every request
by its route pattern (`http_requests_total`, `http_request_duration_seconds`,
`http_inflight_requests`). Every client calls `httpx.Dependency` once per attempt
(`internal/routes/client.go:84`, `internal/nodeclient/client.go:85`, blob and metering clients).
That produces `dependency_requests_total{dep,outcome}` and `dependency_duration_seconds{dep}` for
`control`, `node-a`, `node-b`, `blobstore`, `metering`, and `postgres`, the last from every registry
call through `observe` (`internal/registry/registry.go:55`). Control exports its pool as `db_pool_*`
(`internal/registry/metrics.go:11`). Nodes export `sqlite_open_databases`,
`sqlite_open_connections`, `sqlite_wal_bytes_total`, `statement_duration_seconds{kind}`,
`usage_queue_depth`, `usage_queue_oldest_age_seconds` and `backup_duration_seconds`. The Go
client's default collectors add `process_open_fds` and `go_goroutines`.

No metric has a tenant or database label (`metrics.go:16`). With 240 databases that would be
240 series per metric per instance, and a real platform has many more. Per-tenant questions go to
the usage ledger and to the logs, which may carry tenant and database ids.

**Health** (`internal/httpx/health.go:24`). Each service registers named checks. The answer is
`{status, checks}` with `ok`, `degraded` (200: keep sending traffic) or `failed` (503: take me out).

- control (`cmd/control/main.go:62`): `failed: postgres timeout` when a ping through its
  connection pool takes more than 500 ms. The ping goes through the shared pool on purpose; lesson
  6 shows why that is a problem.
- node (`cmd/node/health.go`): `failed` when the canary database cannot be opened through the same
  handle path tenants use, or the data volume is full. `degraded` when open fds pass 80 % of the
  limit, or WAL bytes pass 128 MB. These are physical facts that lessons 3 and 10 read.
- gateway (`cmd/gateway/health.go`): a background prober checks control and both nodes every
  second. `failed: control unreachable`, because in v0 every request needs control.
  `degraded: node-a unreachable`, because a down node costs its tenants, not the gateway.

HAProxy switches to `option httpchk GET /health` (`deploy/haproxy.cfg:28`).

**Logs** (part 2, `internal/httpx/httpx.go`). `SetupLogging` (`:29`) routes every log line,
including the standard library's, through `log/slog` as JSON. `Logged` gives each request an id: it
takes the caller's `X-Request-Id` or generates one, returns it as a response header, and stores it
in the context. Clients copy it onto outgoing calls (`Propagate`, `metrics.go:98`), so a gateway
request, its control call and its node call share one id. Handlers `Annotate` the line with
`tenant_id`, `database_id` and `node`. Every dependency call writes one line: `dep`, `dur_ms`,
`outcome`. Registry queries are too frequent for a line each, so only slow (>100 ms) or failed ones
are logged (`DependencyMetric`, `metrics.go:82`).

The clients gain a `ctx context.Context` parameter in this lesson, but only to carry the request id.
Outgoing requests are still built with `http.NewRequest`, not bound to the context, and nothing has
a timeout yet. That is lesson 2.

## Verify

```
make verify LESSON=01 REF=release-it/v0-flawed      # before
make verify LESSON=01                               # after
```

Timeline: steady mix at 50 % of fleet capacity (250 req/s) for 60 s; `latency 800ms` toxic on the
node-a proxy from t=15 s to t=33 s; `docker compose pause postgres` from t=35 s to t=45 s. The script
polls control and gateway-a `/health` every 0.25 s and HAProxy's server states every second. It
reads Prometheus at t=32 s and checks every log line written during the run.

Before (v0):

```
  up{job="gateway"}                                 0
  gateway → node-a p99 at t=32 s (latency on node-a) NaN s
  control /health first failed after pause           NaN s
  HAProxy gateways UP during pause (min)             3 of 3
  log lines that are JSON                            0 of 39954 (0.0 %)
FLAW REPRODUCED: up_gateway 0 (want == 3); node_a_dep_p99_s NaN (want >= 0.7); node_b_dep_p99_s NaN (want <= 0.2); control_health_failed_after_s NaN (want <= 3); gateway_health_failed_after_s NaN (want <= 5); log_json_pct 0.0 (want == 100); request_id_pct 0.0 (want == 100)
```

After:

```
  up{job="gateway"}                                 3
  gateway → node-a p99 at t=32 s (latency on node-a) 1.019 s
  gateway → node-b p99 at t=32 s                     0.109 s
  control /health first failed after pause           0.82 s  {"checks":{"postgres":"failed: postgres timeout"},...}
  gateway-a /health first failed after pause         1.95 s  {"checks":{"control":"failed: control unreachable",...}}
  HAProxy gateways UP during pause (min)             0 of 3
  log lines that are JSON                            67575 of 67575 (100.0 %)
  request log lines with a request_id                37573 of 37573 (100.0 %)
  a slow node-a call, found by its log line          {"request_id":"02cde18703fd127e","dep":"node-a","dur_ms":842.859,"outcome":"ok"}
IMPROVED: up_gateway 3 == 3; node_a_dep_p99_s 1.019 >= 0.7; node_b_dep_p99_s 0.109 <= 0.2; control_health_failed_after_s 0.82 <= 3; gateway_health_failed_after_s 1.95 <= 5; log_json_pct 100.0 == 100; request_id_pct 100.0 == 100
```

The latency is now attributed. The gateway's own p99 rises, and so does
`dependency_duration_seconds{dep="node-a"}`, while node-b stays near 100 ms. The request id in the
example finds the matching `request` line, the control call and the node's own line for the same
request. During the pause the health checks say exactly what is wrong. HAProxy takes the gateways
out, so tenants get an immediate 503 from the edge instead of a 3 s hang, which is honest about an
outage that already exists. Lesson 5 makes the data plane survive it.

## Try changing

- Add `tenant` as a label to `http_requests_total` and watch `prometheus_tsdb_head_series` after a
  run. Then answer "which tenant is busiest?" from the usage report instead.
- Make the gateway's health `failed` when a node is down (instead of `degraded`) and rerun: one
  slow node takes every gateway out of HAProxy, and node-b's tenants lose service too. A health
  check can cause the cascade it is meant to report.
- Log every registry query (`DependencyMetric` → `Dependency`) and compare the log volume and the
  steady-state p99.

## Read more

*Release It!* 2nd ed., Ch. 8 "Transparency" and Ch. 10 "System-Wide Transparency"; Ch. 6 for the
story. On metric cardinality, the Prometheus documentation on naming and labels.

## Deviations from the plan

- **Two commits, about 650 lines of Go instead of ~350.** The node's health reports physical facts
  (canary, fds, disk, WAL), the gateway's health needs a background prober, and control's pool
  gauges are spelled out. That roughly doubled the plan's estimate. Following the ~400-line rule,
  part 1 (metrics and health) and part 2 (logs, verify, this note) are separate commits; the tag is
  on part 2.
- **Registry queries are logged only when slow or failed;** every query is still in the metrics.
- The verify removes node-a's latency at t=33 s, before the Postgres pause, so the two signals do
  not overlap.
