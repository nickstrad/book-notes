# 00 · The flawed start: a tour of litefleet v0

Tag: `release-it/v0-flawed` · Verify: `make verify LESSON=00` · Book: *Release It!* 2nd ed., Ch. 1–3

## The idea

Nygard opens the book with a claim that the rest of it defends: software is "done" when it survives
production, not when it passes QA (Ch. 1, "Living in Production"). Chapter 2 tells how one uncaught
exception on a connection's error path drained a resource pool and grounded an airline's check-in
for hours. Chapter 3 gives the vocabulary used throughout this repo. An **impulse** is a sudden
shock (a burst, a partner that stops answering). A **stress** is a force held over time. The
**strain** is what the system does in response. A small fault is a **crack**, and the architecture
decides whether the crack stops or propagates through a **chain of failure** until it becomes an
outage. The question every lesson asks is the one Ch. 3 asks: when this part fails, where does the
strain go?

`litefleet` v0 is the system that passed QA. It has a clean layout, unit tests that pass, typed
errors, API keys, WAL-mode SQLite with read-only query connections, an epoch check on every node
request, scheduled backups, usage metering and a health endpoint on every service. Under gentle load
it is fast. It also has eighteen flaws of the kind real platforms ship with. None of them is marked
in the code. This page lists them all and is the only place that does.

## The platform in one picture

```
 loadgen ──► haproxy :18080 ──┬─► gateway-a ─┬─► toxiproxy ─► control ─► postgres (registry)
                              ├─► gateway-b  ├─► toxiproxy ─► node-a  ─► /data/tenants/*.db
                              └─► gateway-c  └─► toxiproxy ─► node-b  ─► /data/tenants/*.db
 node-{a,b} ─► toxiproxy ─► blobstore (snapshots)     node-{a,b} ─► toxiproxy ─► metering (usage)
 node-{a,b} ─► control (heartbeat, direct)            control ─► node-{a,b} (provision, direct)
 prometheus scrapes every service · toxiproxy API :18474 · DuckDB reads results/*.csv and the volumes
```

A request for `POST /v1/databases/default/query` goes: HAProxy picks a gateway; the gateway sends
the key and database name to control (`POST /internal/resolve`); control answers with the database
id, its node and its epoch; the gateway forwards the statement to that node with `X-Epoch`; the node
opens the tenant's file (once, then keeps it), runs the statement in a read-only transaction and
returns `{columns, rows}`. Writes go the same way to `/execute`. The nodes are stateful and not
interchangeable: a database lives on exactly one node, so when a node is slow, full or restarting,
its tenants feel it and nothing routes around it.

Code worth reading first: `project/internal/api/handlers.go` (the gateway),
`project/internal/engine/exec.go` (the node's execution path), `project/cmd/control/main.go`
(control's routes) and `project/deploy/compose.yml` (the topology and its resource limits).

## Run it

```
cd books/release-it/project
make up && make seed            # first seed ~1 min; the lab snapshots it for later resets
make verify LESSON=00
```

Expected verdict (numbers vary a little run to run):

```
Prometheus: services reporting metrics
  up{job="gateway"} = 0
  up{job="control"} = 0
  up{job="node"} = 0
  ...
BASELINE HOLDS: success_pct 100.00 >= 99.5; p99_ms 30.4 <= 50; errors_5xx_or_timeout 0 == 0
```

The steady mix (75 % point-range reads, 25 % single-row inserts, spread over 197 tenants with a
mild skew) at half the fleet's calibrated capacity is served completely, with p99 under 50 ms. The
three `up = 0` lines are the first hint that something is wrong. Prometheus was told to scrape every
service, but gateway, control and node expose nothing.

## The lab

- `make calibrate` measured, on this VM, one gateway at 300 req/s and one node at 250 req/s
  (`lab/capacity.env`). The fleet sustains 500 req/s, bounded by the two nodes. Every scenario rate
  is a fraction of that. The instances are deliberately small (`cpus: 0.5`, `nofile: 1024`,
  `mem_limit` 512 MB for nodes and 256 MB for gateways) so a failure that takes a quarter to arrive
  in production arrives in a minute here.
- `cmd/loadgen` is open-loop: requests go out at a fixed arrival rate whatever happened to the
  previous ones, the way real users behave. A closed-loop generator waits politely and hides
  overload. It writes one CSV row per request: `ts_ms, scenario, route, tenant_class, node, status,
  latency_ms, err_class, instance`.
- `lab/report.sql` (DuckDB) turns a run into one row per 5 s window per route and tenant class:
  `window route class req 2xx 4xx 5xx timeouts p50_ms p99_ms`. Every lesson prints the same table so
  before and after compare line by line.
- Chaos comes only from outside the code: the Toxiproxy API, `docker compose kill/pause/restart`,
  the partner fakes, SQL run by the script, the node's ops endpoints.

## The deliberate-issues register

Line numbers are at `release-it/v0-flawed`. "Shows" says how the flaw looks under the load that
triggers it, which is never the gentle load above.

| ID | Where (v0) | What is wrong | Shows | Book | Fixed in |
|---|---|---|---|---|---|
| F01 | `internal/httpx/httpx.go:25`, `:73`; every `cmd/*/main.go` | `/health` is a constant `200 ok`; no `/metrics`; logs say `handled POST /path` with no request id, duration or dependency | node-a gets slow: tenants see latency, the balancer sees three healthy gateways, nothing names the node. Pause Postgres: every request fails, every instance stays UP | Ch. 8 Transparency; Ch. 6; Ch. 10 System-Wide Transparency | L01 |
| F02 | `internal/routes/client.go:26`, `internal/nodeclient/client.go:25`, `internal/blob/client.go:21`, `internal/metering/client.go:22` | `&http.Client{}`: no connect, response-header or total timeout on any integration point | node-a accepts and never answers: one gateway goroutine per request waits forever; nothing recovers when the node does | Ch. 4 Integration Points, Blocked Threads; Ch. 5 Timeouts | L02 |
| F03 | `internal/engine/exec.go:72`, `:123` (no context); `internal/registry` (no `statement_timeout`) | statements run with `context.Background()` | a tenant's runaway `WITH RECURSIVE` runs to completion after its client has gone, holding the node's CPU | Ch. 5 Timeouts; Ch. 4 Slow Responses | L02 |
| F04 | `internal/httpx/httpx.go:18`; `deploy/haproxy.cfg:12–13` | `http.Server` with no read/write/idle timeouts; HAProxy `timeout client/server 5m` | slow or stalled clients hold goroutines and balancer sessions for minutes | Ch. 4 Integration Points; Ch. 5 Timeouts | L02 |
| F05 | `internal/engine/exec.go:133` | on a statement error, `Execute` returns before `tx.Rollback()` | a `UNIQUE` violation leaks the connection **and the file's write lock**: that database never accepts another write (`database is locked`) | Ch. 2 case study; Ch. 4 Blocked Threads | L03 |
| F06 | `internal/engine/handles.go:85` (no `SetMaxOpenConns`), `internal/registry/registry.go:32–37` (unbounded open) | no bound on SQLite connections per database or open handles per node; control's Postgres pool unbounded; no pool metrics | with F05 the node reaches its 1 024-fd wall within seconds and stays there | Ch. 4 Blocked Threads, Unbalanced Capacities; Ch. 5 Fail Fast | L03 |
| F07 | `internal/api/handlers.go:59–72`, `internal/engine/exec.go:60–68` | the gateway resolves through control before validating the body; the node opens the file before checking the statement | invalid requests cost a control round trip, a node round trip and a connection before their 400 | Ch. 5 Fail Fast | L03 |
| F08 | `internal/nodeclient/client.go:48–71`; `internal/engine/exec.go:156–176` | the gateway retries any transport error or node 5xx five times, 100 ms apart, including `/execute`; the node meets `SQLITE_BUSY` with a tight 50-try loop and no `busy_timeout` | a half-degraded node gets several attempts per request; writes that committed but timed out are applied twice (duplicate rows in tenants' tables) | Ch. 4 Dogpile, Self-Denial Attacks, Force Multiplier; Ch. 5 Governor | L04 |
| F09 | `internal/routes/client.go:30–33`, `internal/api/handlers.go:50` | every request resolves through control; no breaker, no route cache | control or Postgres down means the whole data plane is down | Ch. 5 Circuit Breaker; Ch. 4 Cascading Failures | L05 |
| F10 | `internal/engine/admission.go:13`; `cmd/control/usage.go:14`; control's single pool | no per-tenant or per-database concurrency limit on a node; the usage report has no cap and shares control's only pool (from L01, so does control's health check) | one noisy tenant slows every tenant on its node; twenty dashboard refreshes starve resolves and creates | Ch. 5 Bulkheads; Ch. 4 Chain Reactions, Cascading Failures | L06 |
| F11 | `internal/backup/job.go:66–89` | the backup takes the database's write lock, snapshots, **uploads**, and only then releases the lock | a slow blob store blocks every write to that database for the whole upload | Ch. 5 Bulkheads; Ch. 4 Blocked Threads, Integration Points | L06 |
| F12 | `cmd/gateway/main.go`, `cmd/node/main.go`, `deploy/haproxy.cfg:25–31` | no in-flight limit anywhere; no per-server `maxconn` or queue timeout; nothing ever answers 503 on purpose | offered load above capacity: queues grow, latency passes the client timeout, goodput collapses while every server is "up" | Ch. 4 Chain Reactions, Slow Responses, Unbalanced Capacities; Ch. 5 Shed Load, Handshaking; Ch. 9 Demand Control | L07 |
| F13 | `internal/engine/exec.go:96–101`; `internal/nodeclient/client.go:87`; `cmd/control/usage.go:30–48` | `/query` returns every row, built in memory then marshalled; the gateway `io.ReadAll`s the node's body; the usage report loads every ledger row of the window into Go | the whale's `SELECT * FROM events` (60 MB) OOM-kills node-a, which takes every node-a tenant down with it | Ch. 4 Unbounded Result Sets; Ch. 5 Let It Crash | L08 |
| F14 | `internal/metering/queue.go:18–29`, `:41–44` | usage events wait in an unbounded slice, shipped 50 at a time by one goroutine, never persisted | a slow metering partner makes the queue grow without limit; a node restart loses every unshipped event (unbilled usage) | Ch. 5 Create Back Pressure; Ch. 4 Unbalanced Capacities | L09 |
| F15 | `internal/engine/handles.go:71`, `:94`; `internal/registry/usage.go:33` | `wal_autocheckpoint(0)` with no checkpointer; the ledger is appended every heartbeat and never purged; open handles are never closed | WAL files grow until the data volume is full (`SQLITE_FULL` on every write); the ledger and the handle count grow with traffic forever | Ch. 5 Steady State | L10 |
| F16 | `internal/nodeclient/client_test.go`; `internal/nodeclient/client.go:87–103`; `internal/blob/client.go:52–61` | tests cover the happy path against stubs that answer in 1 ms; clients read whole bodies and decode before or regardless of the status | a 500 with an HTML body is reported as "bad rows"; a 10 MB error body is held in memory; resets and slow headers were never exercised | Ch. 5 Test Harnesses; Ch. 4 Integration Points; Ch. 15 The Testing Gap | L11 |
| F17 | `internal/nodeclient/client.go:105–110`; `cmd/loadgen/main.go:449–454` | strict JSON readers (`DisallowUnknownFields`, exact field sets) in the gateway and in the reference consumer | a node that adds a response field breaks every query through an older gateway | Ch. 14 Handling Versions | L12 |
| F18 | `cmd/control/main.go:31–39`; `migrations/001_init.sql:29–30`, `:50–51`; `cmd/node/main.go:76` | migrations run inside control's `main()`; byte and row counters are `int`; a node exits immediately on SIGTERM | the whale's byte counter overflows at 2 GB and its node's heartbeat fails; the obvious `ALTER COLUMN ... TYPE bigint` rewrites the ledger under an exclusive lock; a node restart aborts its in-flight transactions | Ch. 12 Waiting for Godot; Ch. 13 Phases of Deployment; Ch. 14 | L12 |

## What the lab found while building v0

Two of these flaws showed up harder than the plan expected, which is useful to know before lesson 3.

- **F05 + F06 + F08 compound.** At 50 % capacity with 30 % of writes violating `items.seq UNIQUE`,
  both nodes hit 1 024 open fds within ten seconds, not a minute. The gateway retries a
  `database is locked` 500 five more times, and each attempt leaks another connection. Each
  leaked SQLite connection also holds about 1 MB, so each node was OOM-killed once on the way up.
  After the restart it climbed straight back to 1 024 and stayed there after the load stopped.
- **F13 is a node outage, not a slow request.** One `SELECT * FROM events` on the whale's database
  costs node-a over 500 MB at peak and kills it. Ten concurrent ones killed node-a four times in 40 s.
  Every node-a tenant failed during each restart, and node-b's tenants did not notice.

## Deviations from the plan

- **Node `mem_limit` is 512 MB, not 384 MB.** At 384 MB a leaking node is OOM-killed before it
  reaches the fd wall, which hides the stays-down failure lesson 3 is about. The gateways keep 256 MB.
- **The whale's `events` table is 15 000 rows of 4 KB (60 MB), not 500 000 small rows.** On a
  half-core node a 500 000-row scan takes about 7 s. From lesson 2 on, the 2 s statement deadline
  would cancel it before the result set grew, and F13 could not be reproduced at lesson 7. Fat rows
  keep the file at ~60 MB and the per-request memory in the hundreds of MB, and the scan finishes
  inside the deadline. Lesson 8's cap is "1 000 rows or 1 MB, whichever first", so the byte cap is
  what truncates the whale.
- **Each tenant file has a second application table, `notes(id, seq, body, created_at)`, with no
  unique constraint.** The steady writes go there. `items.seq` is `UNIQUE` (the `errs` scenario
  violates it), so a retried insert into `items` would fail rather than duplicate. Lesson 4's audit
  counts duplicate `notes.seq` values: a write applied twice.
- **The ledger has a `statements` column**, because the usage report counts statements and one
  heartbeat row covers many.
- **Tenant classes in the report** add `cold` (databases outside the warm set, lesson 5), `ops` (the
  usage-report refreshes, lessons 6 and 8) and `target` (writes to the database being backed up,
  lesson 6) to the plan's `normal`, `whale`, `noisy`, `errs`. Special tenants (whale, noisy, bulky)
  receive no steady traffic, so `normal` means "everyone else".
- **v0 control keeps 32 idle Postgres connections** (`SetMaxIdleConns(32)`) while leaving open
  connections unbounded. With Go's default of 2 idle, every resolve beyond the second concurrent one
  opened a fresh Postgres backend and paid ~10 ms of planning. v0 could not serve gentle load, and a
  v0 that falls over under gentle load is planted wrong.
- **Lesson 00's verdict words are `BASELINE HOLDS:` / `BASELINE BROKEN:`.** There is no fix to call
  improved.
