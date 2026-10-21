# Release It! — breakdown plan v3: one flawed multitenant SQLite platform, lessons as patch commits

**Book:** *Release It! Design and Deploy Production-Ready Software*, 2nd edition, Michael T.
Nygard, Pragmatic Bookshelf, 2018. All chapter numbers are from the 2nd edition.

**Phase:** plan (v3). Supersedes `PLAN-v2-checkout.md` (same structure, a web-shop order service as
the vehicle), which is kept as an archive next to `PLAN-v1-per-topic-projects.md`. Nick confirms this
plan before the builder is delegated anything. Design only: no implementation code exists yet.

**Status of the tree:** the v1 scaffold (`topics/NN-*/README.md` stubs) is still in the repo. It is
superseded by section 8 and is removed in the build phase, not now.

---

## 0. What changed from v2 and why

Nick approved v2's structure (one deliberately flawed project that runs, the deliberate-issues register
mapped to chapters, lessons as patch commits with tags, per-lesson verification, Docker Compose with a
real database and Prometheus, the eight skills, the checkpoints) and rejected its vehicle: a shop's
checkout service is not something he builds or cares about. He listed what he does care about, in
order: a multitenant database platform backed by SQLite files, a job queue, a workflow platform, a
baby version of something Vercel offers. A baby version is explicitly fine.

v3 keeps everything structural from v2 and swaps the vehicle for the first item on that list: a baby
**multitenant SQLite database platform**, called `litefleet` here. Section 2.0 says why it won over the
other three. The lesson sequence is unchanged in order and count (twelve); the register gains one flaw
(eighteen, from seventeen) that only this vehicle has, and several lessons get a sharper demonstration
because the thing that breaks is a tenant's database rather than a shop's order. Sections 1, 4.1, 4.2,
5, 7 and the builder rules are carried over from v2 with only the names changed; the diff against
`PLAN-v2-checkout.md` should read as a re-skin, not a new plan.

Three consequences of v2's design still hold and are repeated because they constrain every flaw below:

1. The flaws must be the kind real codebases have (no timeout on a client, a transaction not rolled
   back on an error path, a query without a row cap), they must be **behavioral** (hangs, storms,
   growth, cascades), and the flawed version must run fine under gentle load, because that is how such
   systems ship.
2. Every flaw must be **demonstrable in isolation** on the tree one commit before its fix, and fixed by
   **one focused patch** a reader can diff in minutes. Flaws therefore need to be independent enough
   that the fixes compose in order.
3. The stack is **one real stack**: Docker Compose, Postgres for the registry, SQLite for tenant data,
   a load balancer, a fault injector, Prometheus, a load generator, and DuckDB for analyzing the runs
   and auditing tenant files. Each choice is justified in section 2.6.

## 1. The book's core argument (unchanged from v1 and v2)

Software is done when it survives production, not when it passes QA. Systems fail through their
integration points and their coupling: a small fault (a slow reply, an exception on `close()`, a query
that returns too much) becomes a crack, and the crack propagates along whatever path the architecture
leaves open until it is an outage. The program: learn the recurring ways cracks start (Ch. 4 stability
antipatterns), build the structures that stop them spreading (Ch. 5 stability patterns), design the
system so it can be seen and run (Part II: transparency, demand control), deploy without downtime
(Part III: phases of deployment, handling versions), and keep proving it (Part IV: chaos).

The unit of analysis is the impulse and the stress on the system, and the question is "where does the
strain go?" (Ch. 3). The project below is one system with several places for strain to go, and each
lesson closes one of them. The chapter table from v1 (section 2 of the archived v1 plan) still applies;
it is not repeated here.

## 2. The project: `litefleet`, a baby multitenant SQLite database platform

### 2.0 Why this vehicle

Judged by the design's own tests (every Ch. 4 antipattern needs a natural home; every lesson must
demonstrate in isolation on the pre-fix tree and fix in one focused patch; the topology must have
instances behind a balancer, integration points, a resource pool, a critical path and a noisy path
without forcing it), the four candidates come out:

| Candidate | Fit | Why it did or did not win |
|---|---|---|
| **Multitenant SQLite platform** | near 1:1 | Integration points are the gateway's calls to the control plane and to node agents, plus two partners (blob store for backups, metering for usage). The resource pool is SQLite connections and file descriptors per node, and the Postgres pool in the control plane. Bulkheads are per-tenant isolation, which is the product's whole promise. Unbounded result sets are a tenant's `SELECT *`. Back pressure is the usage-event path. Steady state is WAL files, the usage ledger and the handle cache. Versions are the gateway/node wire format and the registry schema. It also has a **stateful tier** (nodes) that cannot be routed around, which the shop did not have, and that makes three lessons (L03, L08, L10) more honest. And it is the domain Nick designs in: the control plane / data plane split, a Postgres registry with tenant SQLite files on nodes, route grants with a TTL, and "take the snapshot off the writer lock" are all taken from his ts-db planes proposal (`ts-db/pages/content/tsdb-planes-fable/inventory.md`, DP-01/DP-02), reduced to baby size. |
| Job queue | good | Back pressure, retries, poison messages and steady state are native; but Unbounded Result Sets, Bulkheads and Handling Versions have to be bolted on (a queue has no noisy read path and no public response shape that evolves), and the topology has only one integration point (the broker). |
| Workflow platform | good but large | Everything in the book has a home, but the minimum viable version (definitions, state machine, timers, retries, history) is 2–3× the budget before the first flaw is planted. The teaching would be about the engine, not the book. |
| Baby Vercel product (deploys/previews/edge config) | medium | Deploy pipelines teach Ch. 12–13 well, but Ch. 4's resource-pool and result-set antipatterns need an invented data path. Note that a multitenant SQLite platform *is* a Vercel-style product (Turso, Cloudflare D1, Vercel's own storage offerings), so the first candidate covers this one. |

The baby version: tenants get databases; a database is one SQLite file on one node; a gateway tier
authenticates, resolves the owning node and forwards SQL; a control plane keeps the registry in
Postgres and places new databases; nodes back up to a blob store and emit usage to a metering partner.
No moves, no replication, no fencing protocol beyond an epoch check (those are the parts of ts-db's
design that are out of scope here, and the lesson notes say where they would go).

### 2.1 What it does

Tenants (application developers; a load generator stands in for them) authenticate with an API key,
create databases, and run SQL against them: reads through `/query`, writes through `/execute`. Each
database is a single SQLite file in WAL mode on one of two **node agents**. Three stateless **gateway**
instances behind a load balancer authenticate the key, ask the **control plane** which node owns the
database, forward the statement to that node and stream the rows back. The control plane keeps the
registry (tenants, keys, databases, placement, usage ledger) in Postgres, places new databases on the
node with the fewest, and receives node heartbeats. Nodes snapshot every database to an external
**blob store** on a schedule and emit one usage event per statement to an external **metering**
partner. An operations dashboard polls a fleet-wide **usage report** from the control plane.

This vehicle has every home the archetype needs: four integration points (control, node, blob store,
metering), two resource pools (SQLite connections and file descriptors on a node; Postgres connections
in the control plane), a critical path (query/execute) and two noisy paths (one tenant's heavy queries;
the usage report), a horizontally scaled stateless layer (gateways) and a stateful one (nodes), data that
grows (WAL files, usage ledger, open handles), a queue (usage events), and two interfaces that will
outlive their current shape (the gateway/node wire format and the registry schema).

### 2.2 API (v0)

Tenant-facing, through the gateway, `Authorization: Bearer <api-key>`:

| Route | Purpose | Criticality |
|---|---|---|
| `POST /v1/databases` | Create a database `{name}`: control picks a node, the node provisions the file, placement is recorded. Returns `{id, name, node}`. | critical (the "deploy" path) |
| `GET /v1/databases` | List the tenant's databases. | normal |
| `DELETE /v1/databases/{name}` | Drop: node deletes the file, registry marks it dropped. | normal |
| `POST /v1/databases/{name}/query` | Read: `{sql, args}` on a `query_only` connection. Returns `{columns, rows}`. | critical |
| `POST /v1/databases/{name}/execute` | Write: `{statements:[{sql, args}]}` in one transaction. Returns `{rows_affected, last_insert_rowid}`. | critical |
| `GET /v1/databases/{name}/stats` | Size, row totals, last backup, from the registry. | normal |
| `GET /v1/admin/usage?window=24h` | Fleet-wide per-tenant usage (statements, rows, bytes, storage) for the ops dashboard, which auto-refreshes. | noncritical, heavy |
| `GET /health` | Load balancer health check (every service has one). | ops |

Internal (service to service):

| Call | Purpose |
|---|---|
| gateway → control `POST /internal/resolve {api_key, db_name}` | Authenticate and resolve in one call: `{tenant_id, database_id, node_url, epoch}`. In v0 every request makes this call (F09). |
| gateway → node `POST /internal/db/{database_id}/query\|execute`, header `X-Epoch` | Execute on the owning node. The node rejects a stale epoch with 409 (the baby version of ts-db's admission check). |
| control → node `POST /internal/provision {database_id}`, `DELETE …` | Lifecycle commands (control-plane identity only). |
| node → control `POST /internal/heartbeat` every 5 s | `{node, fds, disk_free, databases:[{id, bytes, wal_bytes, rows}]}`; control updates `nodes` and `databases` counters. |
| node → blobstore `PUT /b/{database_id}/{generation}.db` | Scheduled snapshot upload (every 60 s per database in the lab). |
| node → metering `POST /events [{event_id, tenant_id, database_id, kind, rows, bytes, at}]` | Usage events, batched by the node. |

Tenant-visible errors are typed (`400 bad_request`, `401`, `404 no_such_database`, `409 constraint`,
`429 tenant_busy` from L06, `503 overloaded|draining|control_unavailable` from L05/L07/L12,
`504 timeout` from L02) and carry the request id.

### 2.3 Data model

**Registry (Postgres), control plane only.** `tenants(id, name, plan)`, `api_keys(key_hash pk,
tenant_id)`, `nodes(name pk, url, last_heartbeat_at, fds int, disk_free_bytes bigint)`,
`databases(id uuid pk, tenant_id, name, state text, bytes_total int, rows_total int, created_at,
unique(tenant_id, name))`, `placement(database_id pk, owner_node, epoch int)`,
`usage_ledger(id bigserial, at timestamptz, tenant_id, database_id, kind text, rows int, bytes int)`.
Hand-written SQL, no ORM, placement published only after the node confirms provisioning: this is
ts-db's DP5 and the provisioner flow, at baby size. `bytes_total`, `rows_total` and the ledger's
`bytes`, `rows` are `int` on purpose (lesson 12 migrates them to `bigint`; the whale tenant crosses
2 GB).

**Tenant data (SQLite), nodes only.** One file per database under `/data/tenants/<database_id>.db`,
WAL mode. Each file carries a platform table `_meta(key, value)` (`database_id`, `schema_version`,
`epoch`) and whatever the tenant creates. The node keeps `/data/_node.db` for its own state (placement
rows it has accepted with their epoch; from L09 also the usage outbox).

**Seed** (`make seed`, ~20 s): 200 tenants with 240 databases (most tenants one, a few three; every
tenant's first database is named `default`, as in ts-db's selection contract) spread over the two
nodes, each file with a small application table `items(id integer primary key, seq integer unique,
body text, created_at text)` and 1 000 rows; one **whale** tenant whose `events` table has 500 k
rows (~60 MB file; one account dwarfs the rest, and `SELECT *` on it is the unbounded result set);
`usage_ledger` with 5 M historical rows via `generate_series` (the Godot table). Files are written
directly onto the node volumes by `cmd/seedgen` and registered in Postgres in the same run.

### 2.4 Partner dependencies (deliberately misbehaving downstreams)

Two small Go services in the same repo stand in for external partners. They are **honest fakes**: they
implement the partner's protocol and have realistic capacity limits, and they are reachable only
through **Toxiproxy**, so network-level misbehavior (latency, hangs, resets, slow bytes, bandwidth
limits, outages) is injected from outside the fake, the way it would arrive from a real partner.
Toxiproxy also sits between the gateways and the control plane and between the gateways and each node,
because those are the integration points the book cares most about (peers you own still fail like
partners).

- **blobstore** (`PUT/GET/LIST /b/{database_id}/{generation}.db`): an S3-shaped object store, ~50 MB/s,
  exposes `/metrics` (objects, bytes, in-flight PUTs). Through Toxiproxy it can be made slow in bytes
  (`bandwidth` toxic), which is how lesson 6 shows a backup holding a writer lock. From lesson 11 it
  also has `/admin/mode` for application-level misbehavior (HTTP 500s, malformed JSON, 10 MB error
  bodies, slow headers). MinIO was considered: it is real, but it cannot be told to return malformed
  JSON, and the fake is ~150 lines; the note names MinIO as what you would use at work.
- **metering** (`POST /events`): ~10 ms per batch, deduplicates by `event_id` and exposes `/metrics`
  (`events_received_total`, `events_duplicate_total`), so lost and duplicated usage are both visible
  from the partner's side. Through Toxiproxy it can be made slow (1–2 s) to show an unbounded queue
  growing.

### 2.5 Topology (Docker Compose)

```
 loadgen ──► haproxy :18080 ──┬─► gateway-a :8080 ─┬─► toxiproxy ─► control :8090 ─► postgres :5432
 (compose run)  stats :18404  ├─► gateway-b        ├─► toxiproxy ─► node-a :8070 ─► /data/tenants/*.db
                              └─► gateway-c        └─► toxiproxy ─► node-b :8070 ─► /data/tenants/*.db
 control ─► node-{a,b}  (provision / drop, direct)        node-{a,b} ─► control  (heartbeat, direct)
 node-{a,b} ─► toxiproxy ─► blobstore :8081 (backups)     node-{a,b} ─► toxiproxy ─► metering :8082
 prometheus :19090  scrapes  gateway-{a,b,c}, control, node-{a,b}, haproxy, blobstore, metering, postgres-exporter
 toxiproxy API :18474        (verify scripts add/remove toxics over HTTP; proxies: control, node-a, node-b, blobstore, metering)
 report (compose run)        DuckDB over results/*.csv written by loadgen, and over the node volumes (read-only) for audits
```

Host ports are all in the 1xxxx range because this VM already has 5432 and 8080 bound by another
project. Direct ports are exposed for `curl`-ing one instance: gateways 18001–18003, control 18090,
nodes 18071–18072, Postgres 15432, blobstore 18081, metering 18082.

Nodes are **stateful and not interchangeable**: a database lives on exactly one node, so a node that
is slow, full or restarting takes its tenants with it. HAProxy balances gateways only. This is the
one topological difference from v2 and it is deliberate: lessons 3, 8 and 10 show that a stateful
tier's failures are not routed around, which is the honest version of those lessons for this domain.

### 2.6 Stack and rationale

| Choice | Pick | Why this and not the alternatives |
|---|---|---|
| Language | **Go 1.26** (one module; `cmd/gateway`, `cmd/control`, `cmd/node`, `cmd/blobstore`, `cmd/metering`, `cmd/loadgen`, `cmd/seedgen`) | Goroutines make "one blocked thread per hung call" literal and countable (`go_goroutines`). `net/http` exposes connect/response-header/total timeouts separately, which is exactly the distinction Ch. 5 Timeouts makes. `database/sql` has a real pool with `SetMaxOpenConns` and context-bounded checkout for both Postgres and SQLite, so lessons 2–3 are production knobs, not toy code. Prometheus has a first-class Go client. Static binaries make tiny images and fast Compose builds. The ts-db research on this VM compared Deno-plus-workers against Go for SQLite dispatch and the kb entry `go-sqlite-drivers-cancellation-and-threads` has the numbers; Go won on cancellation and on bounding concurrency. Rust would spend the budget on build times; Node/Deno hides blocking behind the event loop and cannot interrupt a running statement from a worker. |
| SQLite driver | **`github.com/ncruces/go-sqlite3`** (wasm-compiled, no cgo) | Measured on this VM (kb, 2026-10-12): context cancellation stops a running statement in 0.2–0.5 ms, no OS-thread growth per blocked statement (cgo adds one thread per blocked statement and the runtime aborts at 10 000), 3 fds per WAL connection. Cancellation that actually lands is what makes lesson 2's statement deadline real; no cgo keeps the alpine static build. `modernc.org/sqlite` cancels in 1–17 ms (cooperative); `mattn/go-sqlite3` needs cgo and pins threads. The lesson notes cite the measurement. |
| Dependencies | `jackc/pgx` (stdlib adapter), `ncruces/go-sqlite3` (`database/sql` driver), `prometheus/client_golang`, `Shopify/toxiproxy/v2/client` (tests only), `hashicorp/golang-lru/v2` (lesson 10) | Real libraries a Go platform would use. Circuit breaker, retry, shedding and the per-tenant semaphore are hand-written (~100 lines each) because their state machines are the lesson; the notes name `sony/gobreaker` and `failsafe-go` as what you would reach for at work. |
| Registry database | **Postgres 17** (`postgres:17-alpine`, cached on this VM) with `max_connections=100` | The control plane's pool, the lock-queue migration (Godot), the unbounded usage report and the growing ledger are Postgres-shaped lessons. A Postgres registry with per-tenant SQLite files on nodes is Nick's own ts-db planes design (DP5: "Postgres, hand-written SQL, no ORM"); this project is that design at baby size, which is the point of choosing it. |
| Tenant storage | **SQLite** files in WAL mode, one per database, on a per-node volume | The product. WAL mode gives concurrent readers and one writer per file, which is what makes the per-file writer lock (F05, F11), `busy_timeout` (F08), WAL growth (F15) and the three-fds-per-connection budget (F06) real, measurable behaviors rather than metaphors. |
| Postgres metrics | `prometheuscommunity/postgres-exporter` | `pg_stat_activity` counts, lock waits and table sizes in Prometheus; lessons 6, 10 and 12 read them. |
| Load balancer | **HAProxy 3.x** | Active health checks (`option httpchk GET /health`) make lesson 1's health semantics matter and drive the chain reaction in lesson 7. Per-server `maxconn` and `timeout queue` are the Ch. 9 "load balancer refuses connections" mechanism, applied in lesson 7. Built-in Prometheus exporter and stats page. Caddy would be simpler but has no connection queue limit; nginx OSS has no active health checks. |
| Fault injection | **Toxiproxy** (`ghcr.io/shopify/toxiproxy`) | The book's Test Harness idea as a real tool: latency, timeout (accept and never answer), slicer, bandwidth, reset_peer, down. Five proxies (control, node-a, node-b, blobstore, metering). Driven over HTTP from the verify scripts and from Go tests in lesson 11. |
| Instrumentation | **Prometheus** (`prom/prometheus`) scraping everything | Pull-based metrics are the natural fit, and HAProxy, Toxiproxy, postgres-exporter all speak it. Verification scripts query the HTTP API for "a metric that changes". Metrics are never labelled by tenant or database (cardinality); per-tenant accounting is the usage ledger's job, and lesson 1's note says so. Grafana is an optional Compose profile (`--profile ui`) the builder adds only if the budget allows; nothing depends on it. |
| Run analysis and audits | **DuckDB** (CLI in a small image with the `sqlite` extension, run via `docker compose run report`) | The load generator writes one CSV row per request (`ts_ms, scenario, route, tenant_class, node, status, latency_ms, err_class, instance`); per-window p50/p99, error and timeout rates and before/after comparisons are analytical SQL over files. DuckDB's `sqlite` scanner also reads the node volumes (mounted read-only) so `make audit` can check every tenant file at once: duplicate `seq` values after a retry storm (lesson 4), unshipped usage events in the outbox (lesson 9), WAL sizes (lesson 10). |
| Load generator | own Go binary, **open-loop** (fixed arrival rate, not closed-loop) | Open-loop load is what exposes overload: queues grow instead of the client politely waiting. Scenarios by name (`steady`, `errs`, `noisy`, `reports`, `burst`, `whale`, `create`, `rollout`), client timeout 3 s, strict JSON reader in v0 (that strictness is flaw F17). Requests are spread over the 200 tenants with a mild skew so both nodes are busy. |
| Images | multi-stage `golang:1.26-alpine` → `alpine` | Seconds to rebuild after the first module download. First build needs network. |

Resource sizing on this VM (4 cores, 8 GB): gateways and nodes get CPU limits (`cpus: "0.5"` each) so
that one instance saturates at a load the generator can comfortably exceed; nodes get
`ulimits: nofile: 1024` and `mem_limit: 384m`, gateways `mem_limit: 256m`, and the node data volumes
are 256 MB tmpfs in lessons that need a disk to fill. In production those walls are 65 k fds and
terabytes; in the lab they are low so the same failure arrives in a minute instead of a quarter. The
notes say so. `make calibrate` measures one gateway's and one node's capacity and writes
`lab/capacity.env`; scenario rates are multiples of it. That is itself a Ch. 15 lesson (know your
capacity before you load test).

### 2.7 What v0 does right (so the flaws are believable)

v0 is not a strawman. It has a clean package layout, unit tests that pass, a Compose file that comes up
with one command, API keys, typed errors, WAL mode with `query_only` read connections, an epoch check
at the node, scheduled backups, usage metering, a health endpoint on every service, and it serves
`steady` load at half capacity with p99 under 50 ms. Its tests use `httptest` stubs for control and
nodes that always answer in 1 ms. It is the platform that passed QA.

## 3. Deliberate-issues register

Every flaw planted in v0, where it lives, how it shows, which chapter names it, and which lesson fixes
it. This is the answer key for the `hunt-stability-antipatterns` skill (section 5).

| ID | Flaw in v0 (file / area) | How it shows under the right load | Book | Fixed in |
|---|---|---|---|---|
| F01 | `/health` returns `200 ok` unconditionally on gateway, control and node; no `/metrics`; `log.Printf("handled %s", path)` with no request id, no durations, no dependency attribution (`cmd/*/main.go`, `internal/httpx`) | node-a becomes slow: users see latency, the balancer sees three healthy gateways, the logs say "handled", and nothing names the node. Pause Postgres: every request 500s (v0 resolves every request through control) while every instance stays UP. A node at 1 000 open fds looks identical to one at 10. | Ch. 8 Transparency; Ch. 6 case study; Ch. 10 System-Wide Transparency | L01 |
| F02 | `http.Client{}` with no timeout for control, nodes, blobstore and metering; default `Transport` with no dial/response-header timeout (`internal/routes/client.go`, `internal/nodeclient/client.go`, `internal/blob/client.go`, `internal/metering/client.go`) | node-a accepts and never answers: every gateway goroutine for a node-a tenant waits forever; in-flight and goroutines climb linearly; nothing recovers when the node does. A blob store that stops sending bytes holds a backup goroutine for hours. | Ch. 4 Integration Points, Blocked Threads; Ch. 5 Timeouts | L02 |
| F03 | The node runs statements with `context.Background()`; no per-statement deadline (`internal/engine/exec.go`). Control uses `context.Background()` for Postgres and sets no `statement_timeout` (`internal/registry`) | A tenant's runaway `WITH RECURSIVE` runs until it finishes; the node keeps computing for a client that gave up at 3 s; nothing can cancel it; the node's CPU share is gone for everyone on it. | Ch. 5 Timeouts; Ch. 4 Slow Responses | L02 |
| F04 | `http.Server{}` with no `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout` on all three services; HAProxy `timeout server 5m` (`deploy/haproxy.cfg`) | Slow or stalled clients hold goroutines and balancer sessions for minutes. | Ch. 4 Integration Points (socket-based protocols); Ch. 5 Timeouts | L02 |
| F05 | In `execute`, the statement-error branch returns before `tx.Rollback()`; the connection stays checked out with an open write transaction for the life of the process (`internal/engine/exec.go`) | 30 % of writes violate a `UNIQUE` constraint: each failure leaks one connection **and that file's write lock**, so the database never accepts another write (`database is locked`); leaked connections hold 3 fds each toward the container's `nofile` 1024; at the wall `accept()` fails and the node is down for every tenant it hosts, and stays down after the load stops. The airline bug (Ch. 2) on a tenant database. | Ch. 2 case study; Ch. 4 Blocked Threads | L03 |
| F06 | Per-database `*sql.DB` opened with no `SetMaxOpenConns`/`SetMaxIdleConns`, no node-wide budget of open handles (`internal/engine/handles.go`); control's Postgres pool unbounded, no pool metrics | Combines with F05 to hit the fd wall in under a minute; under a burst alone each database opens as many connections as it has concurrent requests. | Ch. 4 Blocked Threads, Unbalanced Capacities; Ch. 5 Fail Fast | L03 |
| F07 | The gateway resolves the route (a control round-trip) and forwards to the node before validating the body (empty statement list, SQL over the size limit, bad JSON); the node opens the file before checking that `/query` carries a single read statement (`internal/api/query.go`, `internal/engine/exec.go`) | Invalid requests cost a control call, a node round-trip and a connection before being rejected with 400. | Ch. 5 Fail Fast | L03 |
| F08 | The gateway retries any node error 5× with a fixed 100 ms sleep, including timeouts and including `/execute` (`internal/nodeclient/client.go`); the node handles `SQLITE_BUSY` with a tight 50-iteration loop and no `busy_timeout` (`internal/engine/exec.go`) | node-a half-degraded (latency above the budget on 50 % of connections): every timeout becomes 6 attempts; node-a's in-flight pegs; the healthy half is crowded out; after the fault is removed node-a stays saturated by the backlog (dogpile on recovery). Writes that timed out at the gateway but committed on the node are retried and **insert duplicate rows into tenants' tables**. Under write contention readers and the writer spin on BUSY. | Ch. 4 Dogpile, Self-Denial Attacks, Force Multiplier; Ch. 5 Governor | L04 |
| F09 | No circuit breaker around control and no route cache: every request resolves through control (`internal/routes/client.go`) | Control (or Postgres) down: the data plane is down with it, every request burns the control timeout plus retries, control is hammered while down, and recovery is noticed only by luck. | Ch. 5 Circuit Breaker; Ch. 4 Cascading Failures | L05 |
| F10 | No per-database or per-tenant concurrency limit on the node (every request is a goroutine straight into SQLite); `GET /v1/admin/usage` runs on control's one Postgres pool with no concurrency cap; control's `/health` pings through that pool; node's `/health` opens its canary through the shared handle path (`internal/engine/admission.go`, `cmd/control/usage.go`, `internal/registry`) | One tenant firing 20 concurrent 2 s queries at node-a makes every other tenant on node-a slow (the noisy neighbour). 20 concurrent usage-report refreshes hold every Postgres connection for seconds; creates and route lookups time out; control's health times out and the gateways mark control unreachable. | Ch. 5 Bulkheads; Ch. 4 Chain Reactions, Cascading Failures | L06 |
| F11 | The backup job takes the database's writer lock, runs `VACUUM INTO` a temp file, **uploads it, then releases the lock** (`internal/backup/job.go`) | A slow blob store (50 KB/s) blocks every write to that database for the whole upload: minutes for a 5 MB file, an hour for the whale. The same shape as ts-db's landed `Storage.snapshot` on the writer queue, which the planes review moved off the lock (M3 `lane-snapshot`). | Ch. 5 Bulkheads; Ch. 4 Blocked Threads, Integration Points | L06 |
| F12 | No limit on in-flight requests per gateway or per node; HAProxy has no per-server `maxconn`/`timeout queue`; no 503 is ever returned on purpose (`cmd/gateway/main.go`, `cmd/node/main.go`, `deploy/haproxy.cfg`) | Offered load above capacity: queues grow without bound, latency passes the client timeout, goodput falls toward zero while the servers are "up". Kill one of three gateways at 80 % utilization: the survivors go to 120 %, their goodput collapses, and the fleet delivers nothing. | Ch. 4 Chain Reactions, Slow Responses, Unbalanced Capacities; Ch. 5 Shed Load, Handshaking; Ch. 9 Demand Control | L07 |
| F13 | `/query` returns every row; the node builds the full `[][]any` and JSON body in memory; the gateway `io.ReadAll`s the node body before forwarding; `GET /v1/admin/usage` loads every ledger row of the window into Go and aggregates in memory (`internal/engine/exec.go`, `internal/nodeclient/client.go`, `cmd/control/usage.go`) | The whale's `SELECT * FROM events` (500 k rows) builds ~200 MB on the node and ~90 MB on the gateway per request; 10 concurrent exceed `mem_limit`: the gateway is OOM-killed and routed around, the **node is OOM-killed and not routed around**: every tenant on node-a fails until it restarts and recovers its WALs. | Ch. 4 Unbounded Result Sets, Slow Responses; Ch. 5 Let It Crash (what makes the restart survivable) | L08 |
| F14 | Usage events (one per statement) sit in an unbounded slice behind a mutex on the node, drained by one goroutine to metering; nothing is persisted; work is lost on restart (`internal/metering/queue.go`) | Metering at 1 s latency: producer 100/s vs consumer 1/s; depth grows 100/s, the oldest event ages into minutes, memory grows, a restart loses every unshipped event (unbilled usage). | Ch. 5 Create Back Pressure; Ch. 4 Unbalanced Capacities | L09 |
| F15 | The node opens every file with `PRAGMA wal_autocheckpoint=0` (added when checkpoints showed up as write-latency spikes, with a TODO for a scheduled checkpointer that was never written); `usage_ledger` rows are inserted per heartbeat batch and never purged; the per-database handle cache never closes an idle database (`internal/engine/handles.go`, `internal/registry/usage.go`) | WAL files grow with every write and never shrink: the 256 MB data volume fills during one run and every write on the node fails with `SQLITE_FULL`; the ledger grows linearly with traffic; every database ever touched keeps 3 fds and a page cache forever. Extrapolated, disk and fds run out in days. | Ch. 5 Steady State (data purging, log files, in-memory caching) | L10 |
| F16 | Tests cover only the happy path against `httptest` stubs; the node client reads the body with `io.ReadAll` and decodes it regardless of status code; no `Content-Type` check; the blobstore client likewise (`internal/nodeclient/client_test.go`, `client.go`, `internal/blob/client.go`) | A 500 with an HTML body is logged as "bad rows"; a 10 MB error body is held in memory; connection resets and slow headers were never exercised, so the first time they happen is in production. | Ch. 5 Test Harnesses; Ch. 4 Integration Points; Ch. 15 The Testing Gap | L11 |
| F17 | Strict JSON readers everywhere: the gateway's decoder for node responses and the load generator (the stand-in consumer) use `DisallowUnknownFields` and require exact field sets | The node's wire v2 adds `column_types`: every query through an old gateway fails during a mixed-version rollout. The gateway adds a field to its own response: the consumer fails. | Ch. 14 Handling Versions (handle others' versions; help others handle yours) | L12 |
| F18 | Control runs registry migrations inside `main()` at startup, so DDL ships in the same deploy as the code; `bytes_total`/`rows_total` on `databases` and `bytes`/`rows` on `usage_ledger` are `int`, and the planned fix is an in-place `ALTER COLUMN ... TYPE bigint` (full rewrite of the 5 M-row ledger under `ACCESS EXCLUSIVE`); nodes have no drain: SIGTERM is an immediate exit (`cmd/control/main.go`, `migrations/`, `cmd/node/main.go`) | The whale's heartbeat overflows `int` at 2 GB and the heartbeat fails. The in-place ALTER, queued behind one running usage report, blocks every create, every heartbeat and every route lookup that misses the cache for the rewrite's duration (Godot). A node restart aborts every in-flight transaction on that node. | Ch. 12 case study (Waiting for Godot); Ch. 13 Phases of Deployment; Ch. 14 | L12 |

Eighteen flaws, twelve lessons. F02–F04 are one lesson because they are one idea (bound every wait);
F05–F07 are one lesson because they are the one case study (Ch. 2); F10–F11 are one lesson because
both are "a shared resource that one path can monopolize"; F17–F18 are one lesson because
expand/contract and tolerant readers are two halves of one rollout.

**Delta from v2:** F11 (backup under the writer lock) is new; it exists only in this vehicle and comes
straight from the ts-db planes review. Everything else is the v2 flaw re-homed; the mapping is one to
one except that v2's F10 (shared pools) became F10, v2's F11–F17 became F12–F18.

## 4. The lesson sequence

### 4.1 Shape of the history

```
release-it/v0-flawed      release-it: add the flawed litefleet platform (v0)
release-it/lesson-01      litefleet: expose health, metrics and structured logs (lesson 01)
release-it/lesson-02      litefleet: bound every integration point and statement with deadlines (lesson 02)
release-it/lesson-03      litefleet: roll back on every path and bound the handle pools (lesson 03)
release-it/lesson-04      litefleet: retry reads only, with jittered backoff and a budget (lesson 04)
release-it/lesson-05      litefleet: circuit breaker around control with a TTL route cache (lesson 05)
release-it/lesson-06      litefleet: bulkhead tenants, the usage report and backups (lesson 06)
release-it/lesson-07      litefleet: shed load at the gateway, the node and the balancer (lesson 07)
release-it/lesson-08      litefleet: cap and stream result sets, aggregate usage in SQL (lesson 08)
release-it/lesson-09      litefleet: bounded usage queue with a SQLite outbox (lesson 09)
release-it/lesson-10      litefleet: reach steady state: checkpoint, purge, evict (lesson 10)
release-it/lesson-11      litefleet: chaos test harness against toxiproxy (lesson 11)
release-it/lesson-12      litefleet: expand/contract migration, node drain and tolerant readers (lesson 12)
release-it/v1-hardened    release-it: lessons index and landing page
```

**Order.** Transparency comes first, before the book's Part I order, because every later verification
reads the metrics it adds, and because that is what a senior engineer does with an opaque system
(Ch. 6: the retailer could not be diagnosed without vital signs). Then the Part I antipatterns in the
order their fixes compose: timeouts make failures finite; a bounded pool makes them local; backoff
stops them amplifying; the breaker stops them repeating; bulkheads stop them spreading between
tenants and paths; shedding stops them spreading between instances. Then the data lessons (results,
queue, steady state), then the harness that proves it all, then versions. Retries (L04) come before the
breaker (L05) on purpose: with a breaker already in place, a retry storm cannot be demonstrated. The
order is unchanged from v2; the vehicle did not force a reordering.

**Tags.** `release-it/v0-flawed` is the before; `release-it/lesson-NN` after each patch;
`release-it/v1-hardened` is the final state. During the build phase history may be rewritten (fixups
folded into their lesson) and tags re-pointed; after Nick accepts the final state, history is frozen.

**Lesson docs.** Both a `lessons/` folder and rich commit bodies, with different jobs. The commit
body (10–20 lines) is the historical artifact: flaw, fix, verify command; it must stand alone in
`git log`. The note `books/release-it/lessons/NN-<slug>.md` is the teaching document: the idea with
chapter citations, the flaw with `file:line` pointers into the v0 tree, the patch walkthrough, the
verification with expected before/after tables, "try changing", and where to read more. Notes live
outside `project/` so they can be revised later without touching the code history, and each lesson's
commit adds its own note so checking out a tag gives the reader that lesson's note too.
`lessons/00-the-flawed-start.md` is a tour of v0 and the register, the reader's entry point.

### 4.2 Verification model (same for every lesson)

- `make up` builds and starts the stack; `make seed` loads data; `make down` removes it. Images are
  cached after the first build, so a lesson's verify run is 1–3 minutes wall clock.
- `make verify LESSON=NN` runs `lab/verify/NN-<slug>.sh`: resets state, starts the load generator
  with the lesson's scenario, applies the lesson's chaos timeline through the Toxiproxy API (and
  `docker compose kill/pause/restart` where the lesson needs it), collects `results/NN-*.csv`, runs the
  DuckDB report (and `make audit` where the lesson names it), queries Prometheus for the metrics the
  lesson names, and prints a verdict with explicit thresholds (`IMPROVED: p99 during outage 1.02 s ≤
  1.2 s` or `FLAW REPRODUCED: ...`). Exit 0 when the lesson's thresholds hold, 1 when the flaw is
  reproduced.
- `make verify LESSON=NN REF=release-it/lesson-MM` runs the same script with the **service images
  built from another commit** (`git archive REF project/` into a temp build context) while everything
  that is harness rather than codebase (partner fakes, load generator, scenarios, report SQL,
  thresholds) comes from the current tree. `REF_SERVICES=gateway` restricts which services are built
  from `REF` (default: gateway, control, node), which is how lesson 12 runs a mixed-version fleet. So on
  the `lesson-05` checkout, `make verify LESSON=05` passes and `make verify LESSON=05
  REF=release-it/lesson-04` reproduces the flaw. This is the "test that failed before" for every
  lesson, and it is why lessons 11 and 12 can point the current partner fakes (new admin modes) and
  the current node (wire v2) at the previous lesson's client.
- Report format (DuckDB over the CSV), one row per 5 s window per route and tenant class:
  `window route class req 2xx 4xx 5xx timeouts p50_ms p99_ms`, followed by the Prometheus readings
  the lesson names and the verdict lines. The same shape for every lesson so before/after compare line
  by line. Tenant classes are `normal`, `whale`, `noisy`, `errs` so a lesson can show that other
  tenants were unaffected.
- Each lesson's thresholds are stated in its section below and in its script; builders may tune the
  numbers after calibration but must keep the direction and the named metric.

### 4.3 Lessons

Each lesson below is the v2 lesson with the vehicle's names, flaws and chaos substituted. A **Delta
from v2** line says whether the demonstration got weaker or stronger and why; "same" means the
mechanism and the thresholds carried over unchanged.

#### L01 · Transparency: health, metrics, structured logs

- **Source:** Ch. 8 "Transparency" (designing for transparency; logging; instance metrics; health
  checks); Ch. 10 "System-Wide Transparency"; Ch. 6 case study.
- **Fixes:** F01.
- **Patch:** `/metrics` on every service with `prometheus/client_golang`: `http_requests_total{route,
  method,status}`, `http_request_duration_seconds` histogram, `http_inflight_requests`,
  `dependency_requests_total{dep,outcome}` and `dependency_duration_seconds{dep}` (deps: `control`,
  `node-a`, `node-b`, `blobstore`, `metering`, `postgres`), `db_pool_*` gauges from `sql.DBStats` on
  control, on the node `sqlite_open_databases`, `sqlite_open_connections`, `sqlite_wal_bytes_total`,
  `statement_duration_seconds{kind=read|write}`, `usage_queue_depth`,
  `usage_queue_oldest_age_seconds`, `backup_duration_seconds`, plus Go runtime defaults
  (`process_open_fds` is the one lesson 3 reads). No `tenant` or `database` label anywhere; the note
  explains cardinality and points at the ledger and the logs for per-tenant questions. `/health`
  returns JSON `{status: ok|degraded|failed, checks:{...}}`, 200 for ok and degraded, 503 for failed:
  control `failed` when the Postgres ping (500 ms budget) fails; node `failed` when its canary open of
  `_node.db` fails or the data volume is at 100 %, `degraded` when fds exceed 80 % of the limit or WAL
  bytes cross a threshold; gateway `failed` when control is unreachable and no routes are cached,
  `degraded` when any node is unreachable. HAProxy switches to `option httpchk GET /health` with
  `http-check expect status 200`. `log/slog` JSON to stdout with request id (propagated as
  `X-Request-Id` through gateway → control/node), route, status, duration, instance, tenant id and
  database id (logs may carry them; metrics may not), and one line per dependency call with its
  duration and outcome. Prometheus scrape config already lists every target in v0 (ops set it up; the
  services never reported), so `up{job="gateway"}` flips 0 → 3.
- **Verify:** `steady` at 50 % capacity for 60 s; at t=15 s add 800 ms latency on the node-a proxy; at
  t=35 s `docker compose pause postgres` for 10 s. Before: the only signal is client-side latency for
  half the tenants, then 500s for everyone; HAProxy shows 3/3 UP throughout; logs are `handled POST
  /v1/databases/x/query`. After: `dependency_duration_seconds{dep="node-a"}` p99 rises at t=15 s while
  `http_request_duration` shows the same rise, attributing it to one node; control `/health` goes
  `failed: postgres timeout` within two checks of the pause and the gateways report `failed: control
  unreachable`; logs carry request ids and `dep=node-a dur_ms=812`. Threshold: `up{job="gateway"} ==
  3`, control health reaches `failed` within 3 s of the pause, every log line parses as JSON with a
  `request_id`.
- **Size:** ~350 lines (the largest patch after v0: three services share one `httpx` middleware and
  metrics package; the node's gauges are the extra over v2); reviewable because it touches middleware,
  a metrics package, and three health handlers and nothing else.
- **Delta from v2:** same mechanism; slightly stronger because a node's health has physical facts to
  report (fds, disk, WAL) that later lessons read.
- **Note for later lessons:** control's health check uses the shared pool on purpose; L06 revisits
  that.

#### L02 · Timeouts on every integration point and every statement

- **Source:** Ch. 4 "Integration Points" (socket-based protocols; HTTP protocols); Ch. 5 "Timeouts";
  Ch. 4 "Slow Responses".
- **Fixes:** F02, F03, F04.
- **Patch:** all four clients get a `Transport` with `DialContext` timeout 200 ms,
  `ResponseHeaderTimeout` 1 s, `TLSHandshakeTimeout`, and an overall `http.Client.Timeout` (control
  500 ms; node 2.5 s; metering 2 s; blobstore 5 min total with the bandwidth floor noted); every call
  takes a `context.Context`. The gateway derives a per-request deadline (`/query`, `/execute`: 3 s)
  and passes it to control and to the node; the node derives a **statement deadline** (2 s) inside
  that and runs every statement with `QueryContext`/`ExecContext`, so a runaway query is interrupted
  (`sqlite3_interrupt` through the driver, in under a millisecond per the kb measurement) and counted
  in `statements_cancelled_total`. Control passes `r.Context()` into every registry call and sets
  `statement_timeout=2s` on the pool. `http.Server` gets `ReadHeaderTimeout 5s, ReadTimeout 10s,
  WriteTimeout 15s, IdleTimeout 60s` on all three; HAProxy `timeout client/server 10s, timeout
  http-request 5s`. Timed-out calls return 504 `node timeout` / `statement deadline` with the request
  id.
- **Verify:** `steady` at 50 % for 45 s; at t=10 s add a `timeout` toxic (accept, never answer) on the
  node-a proxy; remove at t=30 s. Then a second phase: the script submits one tenant's runaway
  `WITH RECURSIVE` query (through the public API, as a tenant would) and watches node-a's CPU and
  `statement_duration_seconds`. Before (at `lesson-01`): every node-a request hangs until the generator
  gives up at 3 s; `http_inflight_requests` and `go_goroutines` on the gateways climb linearly (~1 500
  by t=30 s) and do not fall when the toxic is removed; the runaway query runs to completion (~30 s)
  after its client is gone. After: each node-a request fails at 2.5 s with 504; in-flight stays under
  2 × rate × 2.5 s; goroutines flat; recovery immediate on toxic removal; the runaway is cancelled at
  2.0 s and node-a's CPU returns. Threshold: p99 during the outage ≤ 2.7 s; `max(http_inflight_requests)`
  ≤ 300; `statements_cancelled_total` ≥ 1 and the cancel lands within 50 ms of the deadline.
- **Size:** ~170 lines across four clients, the engine, handlers, servers, haproxy.cfg.
- **Delta from v2:** stronger. The statement deadline is a real, measurable cancellation of a running
  SQLite statement, which the shop's "slow query" could only approximate.

#### L03 · Blocked threads: roll back on every path, bound the pools, fail fast

- **Source:** Ch. 2 case study (the exception that grounded an airline); Ch. 4 "Blocked Threads";
  Ch. 5 "Fail Fast", "Timeouts".
- **Fixes:** F05, F06, F07.
- **Patch:** `defer tx.Rollback()` and `defer rows.Close()` idioms throughout the engine; a constraint
  violation returns a typed `ErrConstraint` (→ 409) through the deferred cleanup, so the write lock is
  released on every path. Per-database pools: `SetMaxOpenConns(4)`, `SetMaxIdleConns(2)`,
  `SetConnMaxIdleTime(1m)`; a node-wide budget of open handles (`max_open_databases`, default from
  calibration against `nofile`), and `sqlite_open_connections` becomes meaningful. Control:
  `SetMaxOpenConns(10)`, `SetMaxIdleConns(10)`, `SetConnMaxLifetime(30m)`. Pool wait is bounded by the
  request context (from L02), so a saturated pool yields a fast 503 `pool timeout` instead of a hang.
  The gateway validates the body (non-empty, size cap, parseable) before resolving the route; the node
  checks `/query` carries one read statement before opening the file.
- **Verify:** `errs` scenario (30 % of writes violate `items.seq UNIQUE`) at 50 % capacity for 60 s.
  Before (at `lesson-02`): `process_open_fds{instance="node-a"}` rises three per failed write; the
  affected databases return `database is locked` on every later write (one leaked open transaction
  per file); at ~1 024 the node fails `accept()`, every node-a tenant gets 504 from the gateways,
  `/health` on node-a reports `failed: fds`, and it stays down after the load stops (the leaked
  connections are never returned). node-b's tenants are unaffected throughout, which is the stateful
  tier's version of "the fleet is down": half the customers. After: fds flat at ≤ 400, constraint
  violations return 409 in < 20 ms, success rate of valid writes ≥ 99 %, no `database is locked`.
  Threshold: `max(process_open_fds{job="node"}) ≤ 600`; no 5xx; zero `locked` errors.
- **Size:** ~130 lines.
- **Delta from v2:** stronger. The leaked resource is a tenant's write lock plus a file descriptor,
  both physical; and the blast radius (every tenant on that node, not routed around) is the honest
  one for a stateful tier. The Postgres pool bound moves to a supporting role.

#### L04 · Retries: reads only, jittered backoff, a budget, `busy_timeout`

- **Source:** Ch. 4 "Dogpile", "Self-Denial Attacks", "Force Multiplier"; Ch. 5 "Governor",
  "Handshaking".
- **Fixes:** F08.
- **Patch:** `internal/retry`: exponential backoff with full jitter (`rand(0, base·2^n)`, base
  50 ms, cap 400 ms), max 2 retries, never retry past the request deadline, retry only idempotent
  calls (`/query` and the control resolve; **never `/execute`**, whose outcome after a timeout is
  unknown; the note explains why a platform cannot de-duplicate arbitrary SQL and what tenants are told
  instead), and a per-gateway **retry budget** (token bucket: retries ≤ 10 % of recent requests). On
  the node, the BUSY loop is replaced by `PRAGMA busy_timeout=2000` (SQLite's own bounded,
  backing-off wait) under the statement deadline. Metrics: `retry_attempts_total{dep,outcome}`,
  `retry_budget_exhausted_total`, `sqlite_busy_total`.
- **Verify:** `steady` at 60 % for 60 s; at t=10 s add a `latency` toxic of 1 500 ms with toxicity
  0.5 on the node-a proxy (half the connections exceed the 2.5 s budget once queued); remove at
  t=30 s. Then `make audit`: DuckDB reads every tenant file on both volumes and counts duplicate
  `seq` values. Before (at `lesson-03`): node-a's request rate is 4–6× the gateways' rate for node-a
  tenants, its in-flight gauge is pegged, timeouts reach ~100 % (the healthy half is crowded out by
  retries of the slow half), after t=30 s node-a stays saturated for 20 s or more (dogpile on
  recovery), and the audit finds hundreds of duplicated `seq` rows (writes retried after a timeout
  that had committed). After: node-a request rate ≤ 1.3× the gateways' rate; errors ≈ the fraction
  the toxic actually breaks; recovery within 3 s of removal; the audit finds zero duplicates.
  Threshold: ratio ≤ 1.3; recovery ≤ 3 s; duplicates = 0.
- **Size:** ~150 lines.
- **Delta from v2:** same mechanism; the duplicate side effect is more damning (rows in a customer's
  table rather than a second email) and is audited directly in the files.

#### L05 · Circuit breaker around control, with a TTL route cache as the degraded mode

- **Source:** Ch. 5 "Circuit Breaker"; Ch. 4 "Cascading Failures", "Integration Points".
- **Fixes:** F09.
- **Patch:** `internal/breaker`: closed → open after 5 consecutive failures (or 50 % failures in a
  10 s window), open → half-open after 5 s, one trial call, success closes. Wraps the control
  client. The gateway gains a **route cache** `(api_key_hash, db_name) → {tenant_id, database_id,
  node_url, epoch}` with a 60 s TTL, refreshed on hit in the background; a cached route is a hint: the
  node still checks the epoch (v0 already does) and a 409 `stale epoch` evicts the entry and re-resolves.
  With the breaker open, cache hits keep serving and misses (and creates) return 503 `control
  unavailable` in microseconds with `Retry-After`. This is ts-db's route-grant idea (F1/M1 in the
  planes revision: "existing grants serve until their TTL; no new grants") at baby size. Metrics:
  `breaker_state{dep}`, `breaker_transitions_total`, `route_cache_entries`, `route_cache_hits_total`,
  `route_cache_misses_total`; `/health` reports `degraded: control breaker open`.
- **Verify:** `steady` at 50 % for 60 s with a 10 s warm-up; at t=10 s `timeout` toxic on the
  control proxy (full control-plane outage); remove at t=40 s. Before (at `lesson-04`): every
  request spends its 500 ms control budget plus budgeted retries before failing; control receives every
  attempt while down; 100 % failures for 30 s; the data plane is down because the control plane is.
  After: breaker opens within ~1 s; warm routes keep serving (0 % failures for tenants seen in
  warm-up); cold routes and creates fail in < 5 ms; control receives ≤ 1 probe per 5 s; breaker closes
  within 5 s of recovery and `breaker_transitions_total` shows closed→open→half-open→closed.
  Threshold: success rate during the outage ≥ 95 %; p50 during outage ≤ 10 ms; control requests during
  outage ≤ 10.
- **Size:** ~190 lines (breaker ~100, cache ~70, wiring).
- **Delta from v2:** stronger. v2's degraded mode was a "try changing"; here the degraded mode is the
  product's central availability property (control plane down, data plane up for known routes) and
  it is built, measured, and tied to Nick's design.

#### L06 · Bulkheads: per-tenant limits on the node, separate pools for the report and health, backups off the writer lock

- **Source:** Ch. 5 "Bulkheads"; Ch. 4 "Chain Reactions", "Cascading Failures", "Blocked Threads".
- **Fixes:** F10, F11.
- **Patch:** three small, independent changes. (1) Node admission (`internal/engine/admission.go`): a
  per-database semaphore (4 concurrent reads, 1 writer, a bounded wait queue of 8 → 429 `tenant_busy`
  with `Retry-After` beyond it) under a node-wide statement limit (16 from calibration) so one tenant
  can hold at most a quarter of the node; this is the lane idea from ts-db (`Semaphore(1)` per file,
  `WRITER_QUEUE_CAP`), and the note says so. (2) Control: two `sql.DB` pools from the same DSN,
  `oltp` (8 connections) for resolve/create/heartbeat and `reporting` (2 connections,
  `statement_timeout=5s`) for `/v1/admin/usage`, with a semaphore of 2 on the report route (503
  beyond); `/health` pings on a dedicated single connection with its own 200 ms budget; node `/health`
  uses a dedicated canary handle. Metrics gain a `pool` label. (3) Backup (`internal/backup/job.go`):
  take the writer lock, `VACUUM INTO` the temp file, **release the lock**, then upload; expose
  `writer_lock_hold_seconds` and `backup_duration_seconds` separately.
- **Verify:** three phases. (a) `noisy`: `steady` at 50 % plus the `noisy` tenant on node-a firing 20
  concurrent 2 s `WITH RECURSIVE` queries in a loop for 45 s. (b) `reports`: `steady` plus 20
  concurrent usage-report refreshes for 30 s. (c) `bandwidth` toxic 50 KB/s on the blobstore proxy,
  then the script triggers a backup of one 5 MB database through the node's ops endpoint
  (`POST /internal/backup/{id}`, the same call the scheduler makes) while the generator keeps writing
  to that database. Before (at `lesson-05`): (a) every node-a tenant's p99 climbs past 2 s and 504s
  appear, node-b's tenants are fine; (b) control's pool is exhausted, creates and cache misses 503
  `pool timeout`, control `/health` fails and the gateways report it; (c) writes to the backed-up
  database 504 for the whole upload (~100 s). After: (a) the noisy tenant gets 429s beyond its share
  and the other node-a tenants' p99 stays ≤ 2× baseline; (b) reports serve 2 at a time and 503 beyond,
  critical-path p99 unaffected, health never leaves `ok`; (c) writes to the backed-up database keep
  their normal latency while the upload proceeds in the background. Threshold: other-tenant p99 ≤ 2×
  baseline and zero other-tenant 5xx in (a); zero critical-path 5xx and `up{job="control"} == 1`
  throughout (b); write p99 ≤ 100 ms for the backed-up database during (c);
  `writer_lock_hold_seconds` ≤ 1 s.
- **Size:** ~220 lines over three files; if calibration pushes it past ~400, split into two commits
  (admission; pools + backup) with the tag on the second.
- **Delta from v2:** much stronger. The noisy neighbour is the one failure every multitenant platform
  is judged by, and F11 is a new flaw with a visible, physical fix. The note is honest that the
  per-tenant semaphore bounds concurrency, not CPU: with the node at `cpus: 0.5` the noisy tenant still
  costs the others up to a quarter of the node, which is why the threshold is 2× baseline, not 1×.

#### L07 · Shed load: bounded in-flight, 503 fast, balancer queue limits (chain reactions)

- **Source:** Ch. 4 "Chain Reactions", "Slow Responses", "Unbalanced Capacities", "Scaling Effects";
  Ch. 5 "Shed Load", "Handshaking", "Fail Fast"; Ch. 9 "Demand Control".
- **Fixes:** F12.
- **Patch:** `internal/shed`: a per-instance in-flight limit (default from calibration, ~1.5 × the
  concurrency needed at capacity) applied as middleware on the gateway **and** on the node: beyond it,
  respond 503 with `Retry-After: 1` in microseconds, counted in `shed_total{route}`; critical routes
  get a larger share than the report and the list route. HAProxy: per-server `maxconn` matching the
  gateway limit, `timeout queue 500ms` (503 from the edge when every server is at its limit),
  `retry-on conn-failure` only. `/health` reports `degraded` while shedding. The load generator treats
  503 as a fast failure (it is the handshake). The note covers the unbalanced capacity between the
  tiers: three gateways can offer a node more than it can take, and the node's own limit is what keeps
  a gateway burst from becoming a node outage.
- **Verify:** `burst` scenario: offered load 0.8 × fleet capacity for 15 s, then `docker compose
  kill gateway-b` at t=15 s, load unchanged to t=60 s, then `start` at t=60 s. Before (at
  `lesson-06`): after the kill the survivors take 1.2 × their capacity; `http_inflight_requests` and
  goroutines climb to the arrival rate times the 3 s deadline (thousands, bounded only by the
  clients giving up); latency passes 3 s; goodput (2xx within 3 s) falls to near zero across the
  fleet; when `gateway-b` returns it is flooded by the queued backlog and the fleet takes tens of
  seconds to recover. After: survivors serve at capacity (~2/3 of previous goodput), excess is
  rejected in < 5 ms with 503, p99 of successful requests stays ≤ 100 ms, in-flight is capped, the
  returning instance rejoins cleanly. Threshold: goodput after the kill ≥ 55 % of before; p99 of 2xx
  ≤ 150 ms; `max(http_inflight_requests)` ≤ limit on every instance.
- **Size:** ~130 lines plus haproxy.cfg.
- **Delta from v2:** same. The kill stays on the stateless tier on purpose; a node kill is L08's
  story (a stateful instance cannot be routed around, so what matters is how fast and cleanly it comes
  back).

#### L08 · Unbounded result sets: cap and stream rows, aggregate usage in the database

- **Source:** Ch. 4 "Unbounded Result Sets" (and "Slow Responses"); Ch. 5 "Let It Crash" for why the
  OOM restart is survivable only because of L01 and L07.
- **Fixes:** F13.
- **Patch:** the node caps a `/query` response at 1 000 rows or 1 MB, whichever first, sets
  `truncated: true` and streams rows with `json.Encoder` as they are scanned (no `[][]any`); the
  gateway streams the node body to the client with `io.Copy` instead of `io.ReadAll`; the tenant-facing
  docs (in `project/README.md`) say to paginate with keyset `WHERE id > ? LIMIT` and that the cap is
  the platform's bulkhead, not a suggestion. `GET /v1/admin/usage` becomes one SQL aggregate (`SUM`,
  `COUNT`, `GROUP BY tenant_id ORDER BY ... LIMIT 50`) on the reporting pool with its
  `statement_timeout`, with a covering index added by migration `002_ledger_index.sql`
  (`CREATE INDEX CONCURRENTLY`, run by `make migrate`, not at startup; this foreshadows L12).
- **Verify:** `whale` scenario: `steady` at 30 % plus 10 concurrent `SELECT * FROM events` on the
  whale's database for 45 s, with `mem_limit` 384m on nodes and 256m on gateways. Before (at
  `lesson-07`): each whale request builds ~200 MB on node-a and ~90 MB on the gateway that forwards
  it; `process_resident_memory_bytes` spikes on both; a gateway is OOM-killed and HAProxy routes
  around it; node-a is OOM-killed within ~10 s (`docker inspect` RestartCount > 0), and **every node-a
  tenant fails** until it restarts and recovers its WALs (~2–3 s), then the next whale request
  repeats it. After: whale requests return 1 000 rows with `truncated: true` in < 30 ms; RSS flat on
  both tiers; zero restarts; `/v1/admin/usage` returns in < 200 ms. Threshold: RestartCount = 0
  on every instance; whale p99 ≤ 100 ms; report p99 ≤ 500 ms; zero non-whale 5xx.
- **Size:** ~150 lines plus one migration.
- **Delta from v2:** stronger. The gateway OOM shows the v2 lesson (routed around); the node OOM shows
  what v2 could not: a stateful crash is an outage for every tenant on the node, and Let It Crash is
  only acceptable because the restart is fast and WAL recovery is automatic.

#### L09 · Back pressure: a bounded usage queue that cannot lose work

- **Source:** Ch. 5 "Create Back Pressure", "Shed Load" (inside vs at the boundary); Ch. 4
  "Unbalanced Capacities"; Ch. 5 "Decoupling Middleware" as the next step (an outbox is a tiny
  broker).
- **Fixes:** F14.
- **Patch:** bounded channel (1 024) with 4 shipper workers batching up to 100 events per `POST
  /events` (each with its own L02 timeout; event ids make redelivery idempotent, and metering
  de-duplicates). The engine enqueues with a 20 ms budget: on timeout (queue full) the event is
  appended to an **outbox table in `/data/_node.db`** (SQLite is already on the box; the database is
  the queue of record) and a **sweeper** (every 5 s, `LIMIT 500`, bounded rate) ships and deletes
  outbox rows. So back pressure is applied to the producer inside the boundary (the enqueue budget)
  without blocking the tenant's statement, and nothing is lost on restart. Metrics:
  `usage_queue_depth` (now ≤ 1 024), `usage_enqueue_timeouts_total`, `usage_outbox_rows`,
  `usage_sweeper_shipped_total`.
- **Verify:** `steady` at 50 % for 60 s with a 1 000 ms `latency` toxic on the metering proxy from
  t=0 to t=30 s; at t=20 s `docker compose restart node-a`. Before (at `lesson-08`): depth grows
  ~100/s to thousands; `usage_queue_oldest_age_seconds` reaches minutes; after the restart metering's
  `events_received_total` never catches up with the generator's count of statements executed on node-a
  (those events are gone). After: depth ≤ 1 024, age ≤ 10 s, statement latency unaffected by the slow
  partner, and within 30 s of removing the toxic metering has received every event exactly once,
  including the ones queued before the restart; `make audit` shows the outbox empty. Threshold:
  `max(usage_queue_depth) ≤ 1024`; metering received = statements executed and `events_duplicate_total
  = 0` at t=60 s; outbox rows = 0.
- **Size:** ~160 lines.
- **Delta from v2:** same mechanism; the outbox being a SQLite table on the node is native rather than
  a Postgres column, and the lost work is billable usage, which is what makes a platform care.

#### L10 · Steady state: checkpoint WALs, purge the ledger, evict idle handles

- **Source:** Ch. 5 "Steady State" (data purging; log files; in-memory caching; avoid fiddling).
- **Fixes:** F15.
- **Patch:** a checkpointer on the node (`PRAGMA wal_checkpoint(TRUNCATE)` per database when it has
  been idle for 30 s, and `wal_autocheckpoint` restored to the default 1 000 pages so busy databases
  checkpoint on their own); a purge job on control (`control purge`, also run by a Compose
  `cron`-style sidecar every minute in the lab, daily in production) that rolls the ledger into
  `usage_daily(tenant_id, day, statements, rows, bytes)` and deletes raw rows older than the retention
  window in batches of 5 000 with a sleep between batches; the handle cache becomes an LRU
  (`hashicorp/golang-lru/v2`, 64 entries, idle 5 min → close), which is ts-db's leased 64-entry handle
  cache, and `/health` reports its size. Metrics: `sqlite_wal_bytes_total`, `usage_ledger_rows` and
  `pg_total_relation_size` (from postgres-exporter), `sqlite_open_databases`.
- **Verify:** `steady` at 60 % for 90 s with the node data volumes mounted as 256 MB tmpfs (the whale
  excluded from the seed for this run). Before (at `lesson-09`): the script samples at 0/30/60/90 s
  and prints growth per minute for WAL bytes, the ledger's relation size and the open-handle gauge,
  then extrapolates time-to-exhaustion for the volume, the VM's disk and `nofile` (minutes, hours,
  days); the volume hits 100 % before 90 s and **every write on that node fails with `SQLITE_FULL`**
  (reads keep working, which is the misleading part). After: WAL total flat after the first
  checkpoint cycle, ledger size flat after the first purge, handle gauge plateaus at ≤ 64, every
  extrapolation reads "never". Threshold: growth over the last 30 s window = 0 for all three.
- **Size:** ~170 lines.
- **Delta from v2:** stronger. The thing that fills is the tenant data volume, the failure is a
  SQLite error every platform operator has met, and the "log file" of v2 is replaced by WAL files,
  which are the same lesson in the vehicle's own terms.

#### L11 · Test harness: chaos tests against Toxiproxy and a misbehaving partner

- **Source:** Ch. 5 "Test Harnesses"; Ch. 4 "Integration Points"; Ch. 15 "The Testing Gap";
  Ch. 17 (the verify scripts are chaos experiments; this lesson automates them).
- **Fixes:** F16.
- **Patch:** `test/chaos/` Go tests (build tag `chaos`, run by `make test-chaos` against the running
  stack) using the Toxiproxy Go client: for each of `latency`, `timeout`, `slicer`, `bandwidth 1 KB/s`,
  `reset_peer`, `down`, the test applies the toxic on the node-a proxy, calls the gateway's node client,
  and asserts the error class and that the call returned within its budget, plus the same matrix for
  the blobstore client, plus a test that the control breaker opens and recovers. The `blobstore` fake
  gains `/admin/mode` for application-level misbehavior (`500`, `malformed-json`, `huge-body` 10 MB,
  `slow-headers`), and both clients are hardened to match: `http.MaxBytesReader`/`io.LimitReader`
  (1 MB for error bodies; the row stream from L08 is already bounded), status-code check before
  decode, `Content-Type` check, classified errors (`ErrTimeout`, `ErrRemote5xx`, `ErrBadPayload`,
  `ErrStaleEpoch`). CI script `make test` runs unit tests; `make test-chaos` runs the harness suite in
  ~60 s.
- **Verify:** `make test-chaos`. Before (at `lesson-10`, running the new tests against the old
  clients): 4 of 12 tests fail (`huge-body` holds 10 MB per call, `500` with HTML is misclassified as
  bad rows, `malformed-json` and `slow-headers` give unclassified errors). After: 12 of 12 pass. The
  script also prints the harness matrix (misbehavior × client × outcome × elapsed) as its table.
- **Size:** ~320 lines, mostly tests; reviewable because tests are table-driven and the client changes
  are ~70 lines.
- **Delta from v2:** same.

#### L12 · Handling versions: expand/contract migration, node drain, tolerant readers

- **Source:** Ch. 14 "Handling Versions" (help others handle your versions; handle others'
  versions); Ch. 13 "Phases of Deployment" (relax between phases; expansion; rollout; cleanup) and
  "The Fallacy of Planned Downtime"; Ch. 12 case study.
- **Fixes:** F17, F18.
- **Patch (two commits, see size note):**
  1. *Migrations become a job, counters expand, nodes drain.* Migrations move out of control's
     `main()` into `control migrate`, run as a one-shot Compose service (`make migrate`) before a
     rollout, never at startup. `003_counters_expand.sql` adds nullable `bytes_total_v2 bigint`,
     `rows_total_v2 bigint` on `databases` and `bytes_v2`, `rows_v2` on `usage_ledger`; control
     dual-writes and reads new-or-old; `control backfill` fills old rows in batches of 5 000 with a
     sleep; `004_counters_contract.sql` (drops the `int` columns, renames, adds `NOT NULL`) is shipped
     but guarded: it refuses to run while any row is unbackfilled or while any heartbeat in the last
     day came from a node reporting the old field names. Nodes gain a **drain**: on SIGTERM stop
     admitting (503 `draining`, `Retry-After: 2`), finish in-flight statements (bounded by the
     statement deadline), checkpoint every WAL, close handles, exit; the gateway retries a 503
     `draining` on `/query` once after `Retry-After` (never `/execute`). `make rollout` restarts
     gateways a → b → c, then nodes a → b, waiting for `/health` between each. The note is honest that
     restarting a node in place still costs its tenants ~2 s of writes, and that the real platform
     answer is ts-db's offline move protocol (move the database, then restart the empty node), which
     is out of scope here.
  2. *Tolerant readers and the additive wire format.* The gateway's node-response decoder drops
     `DisallowUnknownFields`, requires only the fields it uses, and validates types and ranges
     instead of shape. The load generator (the consumer) does the same first. The node's wire v2
     (`NODE_WIRE_VERSION=2`) adds `column_types` and `stats{rows_read, bytes}` and keeps every v1
     field; the gateway's response adds `stats` and `node_epoch` and keeps every v1 field; the
     request accepts `statements` and the older `sql` form.
- **Verify:** three parts. (a) `make verify LESSON=12` runs the `rollout` scenario: `steady` at 50 %
  plus a trickle of creates while `make migrate && make rollout` executes the expand migration and the
  rolling restart; then node-b is restarted with `NODE_WIRE_VERSION=2`. After: zero errors through the
  migration and the gateway rollout; during each node restart ≤ 2 s of 503 `draining` on that node's
  writes, zero aborted transactions, zero read errors; zero errors on the mixed-version fleet; the
  consumer keeps working with the new fields present. (b) Before, at `lesson-11` with
  `REF_SERVICES=gateway` (old gateway, new node): the same script's wire-v2 step → 100 % query
  failures through the old gateway (`unknown field "column_types"`); and a plain `docker compose
  restart node-a` at `lesson-11` aborts every in-flight transaction on node-a. (c) `make godot` (runs
  on any tree) starts a loop of usage reports, then applies the naive in-place `ALTER TABLE
  usage_ledger ALTER COLUMN bytes TYPE bigint` under the `create` scenario with the gateways' route
  cache TTL set to 5 s for the run (a config override in the Compose profile, not a code path) and
  prints the stall: the ALTER waits for `ACCESS EXCLUSIVE` behind the report, every create, heartbeat
  and cache-miss resolve queues behind the ALTER (`pg_locks` waits visible through postgres-exporter),
  control `/health` goes `degraded`, and the queue drains only when the rewrite finishes. Threshold:
  zero 5xx during (a) other than the bounded `draining` window; (b) reproduces ≥ 99 % query failures
  and ≥ 1 aborted transaction.
- **Size:** ~290 lines over two commits (~190 + ~100). Two commits because the migration mechanism
  (with drain) and the API tolerance are separately reviewable and the book treats them as two
  topics; the tag `release-it/lesson-12` points at the second.
- **Delta from v2:** same strength overall, with one honest weakening and one addition. Weaker: with
  a single control instance there is no "three instances race to run the same DDL" symptom, and after
  L05's route cache the Godot stall hits creates, heartbeats and cache misses rather than every
  request, which is why the `create` scenario and the short TTL are part of (c). Added: the node drain
  is the phases-of-deployment lesson for a stateful tier, which v2 had no place for.

### 4.4 Coverage honesty

| v1 topic | v3 | Note |
|---|---|---|
| 01 integration points / timeouts | L02 | same idea, now on four real clients, three servers, the balancer, and a cancellable SQLite statement |
| 02 blocked threads / resource pool | L03 | the airline bug reproduced on a tenant's write lock and the node's fd budget |
| 03 circuit breaker | L05 | with the TTL route cache as the degraded mode |
| 04 bulkheads | L06 | per-tenant admission on the node, separate Postgres pools, backup off the writer lock; instance-level bulkhead (dedicated nodes per plan) described, not built |
| 05 chain reactions / shed load | L07 | real kill of one gateway behind HAProxy; node-side limit added |
| 06 dogpile / retries | L04 | moved before the breaker so the storm is demonstrable; duplicate rows audited in the files |
| 07 unbounded result sets | L08 | with a real OOM kill of both a stateless and a stateful instance |
| 08 back pressure | L09 | bounded queue plus a SQLite outbox on the node |
| 09 steady state | L10 | WAL files, usage ledger, handle cache |
| 10 test harness | L11 | Toxiproxy-driven tests plus a misbehaving blob store |
| 11 transparency | L01 | moved first; everything else is measured through it |
| 12 handling versions | L12 | adds the Ch. 12/13 migration lesson (Godot) and a node drain |

Nothing dropped. Merged inside lessons: Fail Fast (L03, L07), Handshaking (L07: 503 + `Retry-After`
+ health checks), Governor (L04 retry budget), Let It Crash (L08), Decoupling Middleware (L09 note).
Still deliberately omitted as lessons, for the v1 reasons: Foundations (Ch. 7), Security (Ch. 11),
Adaptation (Ch. 16); Chaos Engineering (Ch. 17) is the method of every verify script and the
`design-chaos-experiment` skill rather than a lesson of its own.

**What the vehicle adds beyond the book's list, and what it leaves out.** It adds a stateful tier
(nodes), multitenant isolation as a first-class bulkhead, and a control/data plane split whose
availability property is demonstrated (L05). It leaves out everything in ts-db's planes design that is
about moving a database between nodes (fencing, epochs beyond the baby check, attestation, backup
manifests as the restore contract, placement policy beyond "fewest databases"). Each lesson note names
where those would attach ("try changing") so the project can grow toward the real design without
reworking the lessons.

## 5. Skills phase, revisited

The v1 set of eight skills stands (`test-integration-points`, `hunt-stability-antipatterns`,
`suggest-stability-patterns`, `budget-timeouts-and-retries`, `review-capacity-headroom`,
`review-operational-transparency`, `plan-safe-release`, `design-chaos-experiment`; descriptions in
the archived v1 plan, section 4, and in `skills/README.md`). Each now cites the project as its worked
example, and two gain a built-in exercise:

| Skill | Worked example in the project |
|---|---|
| `hunt-stability-antipatterns` | Run it against `release-it/v0-flawed`; the register in section 3 is the answer key. A skill run that finds fewer than 13 of the 18 is a skill that needs work. |
| `suggest-stability-patterns` | Its output for the register should reproduce the lesson sequence's patterns and placements, including the per-tenant bulkhead and the route cache as a breaker fallback. |
| `test-integration-points` | L11's harness matrix and `test/chaos` are the template output. |
| `budget-timeouts-and-retries` | L02's deadline budget (3 s → 2.5 s node → 2 s statement; 500 ms control) and L04's policy, including "never retry a write whose outcome is unknown". |
| `review-capacity-headroom` | `make calibrate` and L07's thresholds; the gateway kill is the headroom check; the gateway/node capacity ratio is the unbalanced-capacities worked example. |
| `review-operational-transparency` | L01's metric list, the no-tenant-label rule, and the three health semantics as the checklist's reference answer. |
| `plan-safe-release` | L12's expand → rollout (with drain) → contract with its guard query. |
| `design-chaos-experiment` | Every `lab/verify/NN-*.sh` is an experiment card made executable; the skill's template mirrors their structure (steady state, hypothesis, injection, observation, threshold). |

Skills impact of the vehicle swap: none to the set or the contract; only the worked-example column and
two procedure signals change (`hunt` gains "per-tenant resource shared across tenants" and "backup or
maintenance under a lock" as things to look for; `review-capacity-headroom` gains "stateful tier that
cannot be routed around"). Skills are written after the lessons, by the builder or a following agent,
under the unchanged skills contract in `AGENTS.md`.

## 6. Build order and budget (brief for the Opus pass)

One delegated pass, three checkpoints, each reviewed by a fresh agent from a different family (Codex)
before the next starts. Fixups found in review are folded into the lesson they belong to and tags
re-pointed (history is rewritten freely until final acceptance).

**Checkpoint A — v0 and the lab (~3 500–4 000 lines; v2 was 2 500–3 000).** The vehicle has three
services where v2 had one, so A is split into two reviewable halves:

- **A1 — the platform runs.** `project/`: module, `cmd/control` (registry, placement, resolve,
  heartbeat, usage report, startup migrations), `cmd/node` (engine, handles, admission-free execution,
  backup job, usage queue, heartbeat), `cmd/gateway` (auth, resolve client, node client, handlers),
  `cmd/blobstore`, `cmd/metering`, `cmd/seedgen`, migrations + seed, Dockerfiles, `deploy/compose.yml`,
  `haproxy.cfg`, `prometheus.yml`, `toxiproxy.json`, `Makefile` (`up/down/seed/calibrate/verify/
  audit/report/test/logs`). Prove `make up && make seed` and a hand-run `steady` load; unit tests pass.
- **A2 — the lab and the flaws.** `lab/`: `cmd/loadgen` with the scenarios in 2.6, `lab/verify.sh`
  (the runner with `REF` and `REF_SERVICES` support), `lab/report.sql` and `lab/audit.sql` (DuckDB),
  `lab/calibrate.sh`, `lab/toxics.sh` helpers. Plant every flaw in the register exactly where it
  says; the flaws must be natural code, not commented "BUG HERE". `lessons/00-the-flawed-start.md` is
  the only place they are listed. Prove v0 runs: `make up && make seed && make verify LESSON=00` (the
  `steady` scenario at 50 %, passes), unit tests pass. Tag `release-it/v0-flawed`. Then L01 (needed by
  every later verify). Tag.

**Checkpoint B — L02 through L07 (~1 000 lines of patches; v2 was ~800).** Each: patch,
`lessons/NN-*.md`, `lab/verify/NN-*.sh`, tag. Before tagging, run `make verify LESSON=NN` (pass) and
`make verify LESSON=NN REF=release-it/lesson-(NN-1)` (reproduces the flaw) and paste both verdict
lines into the commit body. Calibrate once at the start of B; store `lab/capacity.env` in git. L06 is
the largest patch in B and may be two commits.

**Checkpoint C — L08 through L12, landing page, skills (~1 100 lines of patches, 8 skills; v2 was
~1 000).** Same per-lesson routine. Then `lessons/README.md` (index with tags),
`books/release-it/README.md` (landing page), `STATE.md`, and the skills from section 5. Tag
`release-it/v1-hardened`.

**Budget delta from v2:** roughly +35 % on checkpoint A (three services, two fakes, seedgen writing
SQLite files), +20 % on B (L05's cache, L06's three parts), +10 % on C (drain in L12). The lesson
count and the review cadence are unchanged. If the pass runs long, A1, A2 and B land fully before C
is started as a second pass.

**Large patches and how they stay reviewable.** L01 (~350 lines: one metrics package, one middleware,
three health handlers), L06 (~220, three independent files), L11 (~320 lines, table-driven tests + two
~35-line client changes), L12 (two commits). Everything else is 130–190 lines. A patch that grows past
~400 lines is split into two commits under the same lesson with the tag on the last.

**Rules for the builder.**
- Read `AGENTS.md`, then this plan, then `lessons/00` once it exists. The register and the lesson
  order are fixed; thresholds and sizes may be tuned after calibration, with the change recorded in
  the lesson note under "Deviations from plan" and in `STATE.md`.
- v0 must pass its own tests and `make verify LESSON=00`. A flaw that makes v0 fail under gentle
  load is planted wrong.
- Every verify script: idempotent, ≤ 3 minutes with warm images, prints the report table, the named
  Prometheus readings, and a verdict line that starts with `IMPROVED:` or `FLAW REPRODUCED:`.
- Chaos is injected only through Toxiproxy's API, `docker compose kill/pause/restart`, the partner
  fakes' admin endpoints, SQL run by the script (through the tenant API or directly on Postgres), the
  node's real ops endpoints (backup trigger, drain), and Compose config overrides for a run. Never
  through hidden code paths in the service.
- Tenant data is only ever read by the lab through the API or through DuckDB's `sqlite` scanner on a
  read-only mount; the lab never writes tenant files except through `seedgen`.
- Commit subjects as in 4.1; bodies carry flaw, fix, both verdict lines. Trailer per repo rules.
- Budget check: if the pass is running long, finish checkpoints A and B fully, land them, and report;
  C can be a second pass. Do not thin out lessons to fit.

## 7. Reader's path (what the final artifact looks like in use)

```
git checkout release-it/v0-flawed && cd books/release-it/project
make up && make seed && make verify LESSON=00        # it works
git log --oneline release-it/v0-flawed..release-it/v1-hardened   # the story
git checkout release-it/lesson-03
make verify LESSON=03 REF=release-it/lesson-02       # FLAW REPRODUCED: node-a fds 1024, 48 % of tenants down
make verify LESSON=03                                # IMPROVED: fds ≤ 400, 409 in 11 ms
git diff release-it/lesson-02 release-it/lesson-03 -- project/   # the fix, in minutes
```

## 8. Repo layout v3 and disposition of the v1 scaffold

```
books/release-it/
  PLAN.md                           # this document
  PLAN-v2-checkout.md               # archived v2 (kept, not maintained)
  PLAN-v1-per-topic-projects.md     # archived v1 (kept, not maintained)
  README.md                         # landing page: argument, lesson index with tags, skills index
  lessons/
    README.md                       # index: lesson, tag, flaw ids, verify command
    00-the-flawed-start.md          # tour of v0 and the register
    01-transparency.md … 12-handling-versions.md
  project/                          # the one codebase; the thing the commits improve
    README.md  Makefile  go.mod  go.sum
    cmd/gateway/  cmd/control/  cmd/node/  cmd/blobstore/  cmd/metering/  cmd/loadgen/  cmd/seedgen/
    internal/httpx/ (server, middleware, logging; metrics from L01)
    internal/api/ (gateway handlers)  internal/routes/ (control client; cache from L05)
    internal/nodeclient/ (gateway → node)
    internal/registry/ (control: Postgres)  internal/placement/
    internal/engine/ (node: handles.go, exec.go, admission.go)  internal/backup/
    internal/blob/ (node → blobstore)  internal/metering/ (client, queue; outbox from L09)
    internal/retry/ (L04)  internal/breaker/ (L05)  internal/shed/ (L07)
    migrations/  deploy/ (compose.yml, Dockerfile.*, haproxy.cfg, prometheus.yml, toxiproxy.json)
    lab/ (verify.sh, verify/NN-*.sh, report.sql, audit.sql, calibrate.sh, capacity.env, toxics.sh)
    results/ (gitignored CSVs)
    test/chaos/ (L11)
  skills/                           # unchanged index now; SKILL.md files in checkpoint C
```

The flawed v0 lives only in history (tag `release-it/v0-flawed`); the working tree at `main` is the
hardened final state. Readers reach any intermediate state by tag.

**Disposition of the v1 scaffold.** Unchanged from v2: remove `books/release-it/topics/` (twelve stub
READMEs) in checkpoint A; keep `books/release-it/skills/README.md` (the index stands; only the
worked-example column changes); `books/release-it/README.md` is rewritten now for the vehicle and
again in checkpoint C with real tags. Nothing in the tree is deleted in this phase.

## 9. Decisions Nick is asked to confirm

1. The vehicle (`litefleet`, a baby multitenant SQLite platform: gateways, control plane with a
   Postgres registry, two nodes with tenant SQLite files, blobstore and metering partners) over the
   job queue, the workflow platform and a Vercel-style product, per 2.0.
2. The stack in 2.6, in particular Go, `ncruces/go-sqlite3` (no cgo) over modernc/mattn, Postgres as
   the registry, HAProxy over Caddy/nginx, Prometheus over TimescaleDB, DuckDB with the `sqlite`
   scanner for audits, a Go fake blobstore over MinIO, Grafana optional.
3. The baby scope: no moves, no replication, placement = fewest databases, epoch check only; each
   note says where the ts-db planes mechanism would attach.
4. The register at eighteen flaws (F11, backup under the writer lock, added) and the twelve lessons in
   the unchanged order; L05's TTL route cache as the breaker's degraded mode; L06 as three parts.
5. Lesson docs as `lessons/NN-*.md` plus commit bodies; tags `release-it/v0-flawed`,
   `release-it/lesson-NN`, `release-it/v1-hardened`; history rewritable until final acceptance.
6. Lesson 12 as two commits and L06 possibly two; everything else one commit per lesson.
7. Checkpoint A split into A1/A2; three checkpoints with cross-family review between them.
8. Removal of `topics/` in the build phase; skills index kept.
