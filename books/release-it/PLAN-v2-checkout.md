# Release It! — breakdown plan v2: one flawed project, lessons as patch commits

**Book:** *Release It! Design and Deploy Production-Ready Software*, 2nd edition, Michael T.
Nygard, Pragmatic Bookshelf, 2018. All chapter numbers are from the 2nd edition.

**Phase:** plan (v2). Supersedes `PLAN-v1-per-topic-projects.md` (twelve standalone
mini-projects), which is kept as an archive. Nick confirms this plan before the builder is
delegated anything. Design only: no implementation code exists yet.

**Status of the tree:** the v1 scaffold (`topics/NN-*/README.md` stubs) is still in the repo. It is
superseded by section 8 and is removed in the build phase, not now.

---

## 0. What changed from v1 and why

v1 taught each idea with its own tiny program. Nick redirected: one realistic service, built badly on
purpose with real technology, and every lesson from the book is a **patch commit** that improves that
same codebase. The commit history is the teaching artifact: `git log` reads as the system getting
better, the first tag is the "before", the last tag implements the book, and any lesson can be checked
out and its flaw/fix demonstrated with one command.

Three consequences for the design:

1. The flaws must be the kind real codebases have (no timeout on a client, a leaked connection on an
   error path, a query without `LIMIT`), they must be **behavioral** (hangs, storms, growth, cascades),
   and the flawed version must run fine under gentle load, because that is how such systems ship.
2. Every flaw must be **demonstrable in isolation** on the tree one commit before its fix, and fixed by
   **one focused patch** a reader can diff in minutes. Flaws therefore need to be independent enough
   that the fixes compose in order.
3. The stack is **one real stack**: Docker Compose, Postgres, a load balancer, a fault injector,
   Prometheus, a load generator, and DuckDB for analyzing the runs. Each choice is justified in section 2.

## 1. The book's core argument (unchanged from v1)

Software is done when it survives production, not when it passes QA. Systems fail through their
integration points and their coupling: a small fault (a slow reply, an exception on `close()`, a query
that returns too much) becomes a crack, and the crack propagates along whatever path the architecture
leaves open until it is an outage. The program: learn the recurring ways cracks start (Ch. 4 stability
antipatterns), build the structures that stop them spreading (Ch. 5 stability patterns), design the
system so it can be seen and run (Part II: transparency, demand control), deploy without downtime
(Part III: phases of deployment, handling versions), and keep proving it (Part IV: chaos).

The unit of analysis is the impulse and the stress on the system, and the question is "where does the
strain go?" (Ch. 3). The project below is one system with several places for strain to go, and each
lesson closes one of them. The chapter table from v1 (section 2 of the archived plan) still applies;
it is not repeated here.

## 2. The project: `checkout`, an order service for a small shop

### 2.1 What it does

A web shop's order service. Customers (a load generator stands in for the web and mobile clients)
place orders; the service validates the items, asks an external **pricing** partner for a quote
(prices and tax), reserves stock in its own Postgres, writes the order, and queues an order-confirmation
call to an external **notify** partner (email/webhook). The back office reads per-customer order lists
and a daily sales report. Three instances run behind a load balancer, as a real deployment would.

This archetype is chosen because every antipattern in Ch. 4 has a natural home in it: two integration
points (pricing, notify), a resource pool (Postgres connections), a critical path and a noisy path
sharing resources (checkout vs reports), a horizontally scaled layer (three instances) behind a
balancer, data that grows (orders, audit rows, logs), a queue (notifications), and an API with
consumers that will outlive its current shape (versions).

### 2.2 API (v0)

| Route | Purpose | Criticality |
|---|---|---|
| `POST /orders` | Create an order: validate items, quote pricing, reserve stock, insert order + items + audit row, enqueue notification. Body `{customer_id, items:[{sku, qty}]}`, optional `Idempotency-Key` header. Returns `{id, status, total_cents}`. | critical |
| `GET /orders/{id}` | Read one order. | critical |
| `GET /orders?customer_id=X` | List a customer's orders. | normal |
| `GET /reports/daily?date=` | Sales totals for a day: count, revenue, top SKUs. Used by an auto-refreshing back-office dashboard. | noncritical, heavy |
| `GET /health` | Load balancer health check. | ops |

### 2.3 Data model (Postgres)

`inventory(sku pk, on_hand int)`, `orders(id bigserial pk, customer_id text, status text,
total_cents int, created_at timestamptz, notified_at timestamptz null)`, `order_items(order_id,
sku, qty, unit_cents)`, `audit_log(id bigserial, at timestamptz, request_id text, route text,
request_body jsonb, response_body jsonb)`.

Seed (applied once by `make seed`, ~5 s with `generate_series`): 500 SKUs; 50 k orders spread over
2 000 customers; one **whale** B2B customer with 200 k orders (realistic: one account dwarfs the
rest); `audit_log` empty. `total_cents` is `int` on purpose (lesson 12 migrates money to `bigint`
split into subtotal/tax, a classic breaking change).

### 2.4 Partner dependencies (deliberately misbehaving downstreams)

Two small Go services in the same repo stand in for external partners. They are **honest fakes**: they
implement the partner's protocol and have realistic capacity limits, and they are reachable only
through **Toxiproxy**, so network-level misbehavior (latency, hangs, resets, slow bytes, bandwidth
limits, outages) is injected from outside the fake, the way it would arrive from a real partner.

- **pricing** (`POST /quote` → `{quote_id, lines:[{sku, unit_cents}], tax_cents, total_cents}`):
  ~20 ms per quote, a fixed concurrency of 8 with a bounded internal queue (a real tax API has a
  capacity, and this one is small enough that a retry storm from three checkout instances saturates it).
  Exposes `/metrics` (requests, in-flight, queue depth, rejected) so a storm is visible from the
  partner's side. From lesson 11 on it also has an admin endpoint for application-level
  misbehavior (HTTP 500s, malformed JSON, 10 MB bodies, slow headers) and from lesson 12 a `v2`
  response shape (additive fields).
- **notify** (`POST /order-confirmation`): ~10 ms per call, exposes `/metrics`. Through Toxiproxy it
  can be made slow (1–2 s) to show an unbounded queue growing.

### 2.5 Topology (Docker Compose)

```
 loadgen ──► haproxy :18080 ──┬─► checkout-a :8080 ─┬─► postgres :5432
 (compose run)  stats :18404  ├─► checkout-b        │
                              └─► checkout-c        ├─► toxiproxy ─► pricing :8081
                                                     └─► toxiproxy ─► notify  :8082
 prometheus :19090  scrapes  checkout-{a,b,c}, haproxy, pricing, notify, postgres-exporter
 toxiproxy API :18474        (verify scripts add/remove toxics over HTTP)
 report (compose run)        DuckDB over results/*.csv written by loadgen
```

Host ports are all in the 1xxxx range because this VM already has 5432 and 8080 bound by another
project. Direct per-instance ports (18001–18003) are exposed for `curl`-ing one instance's `/health`.

### 2.6 Stack and rationale

| Choice | Pick | Why this and not the alternatives |
|---|---|---|
| Language | **Go 1.26** (one module; `cmd/checkout`, `cmd/pricing`, `cmd/notify`, `cmd/loadgen`) | Goroutines make "one blocked thread per hung call" literal and countable (`go_goroutines`). `net/http` exposes connect/response-header/total timeouts separately, which is exactly the distinction Ch. 5 Timeouts makes. `database/sql` has a real pool with `SetMaxOpenConns` and context-bounded checkout, so lessons 2–3 are production knobs, not toy code. Prometheus has a first-class Go client. Static binaries make tiny images and fast Compose builds. Rust would spend the budget on build times and ownership ceremony that is not the lesson here; Node hides blocking behind the event loop. |
| Dependencies | `jackc/pgx` (stdlib adapter), `prometheus/client_golang`, `Shopify/toxiproxy/v2/client` (tests only), `hashicorp/golang-lru/v2` (lesson 10) | Real libraries a Go service would use. The v1 "std-only" rule is dropped; this is a real codebase. Circuit breaker, retry and shedding are hand-written (~100 lines each) because their state machines are the lesson; the notes name `sony/gobreaker` and `failsafe-go` as what you would reach for at work. |
| Database | **Postgres 17** (`postgres:17-alpine`, cached on this VM) with `max_connections=100` | The pool, the lock-taking migration, the unbounded query and the growing table are all Postgres-shaped lessons. `max_connections` is the hard wall the unbounded pool hits. |
| Postgres metrics | `prometheuscommunity/postgres-exporter` | `pg_stat_activity` counts and table sizes in Prometheus; lessons 3 and 10 read them. |
| Load balancer | **HAProxy 3.x** | Active health checks (`option httpchk GET /health`) make lesson 1's health semantics matter and drive the chain reaction in lesson 7. Per-server `maxconn` and `timeout queue` are the Ch. 9 "load balancer refuses connections" mechanism, applied in lesson 7. Built-in Prometheus exporter and stats page. Caddy would be simpler to configure but has no connection queue limit; nginx OSS has no active health checks. |
| Fault injection | **Toxiproxy** (`ghcr.io/shopify/toxiproxy`) | The book's Test Harness idea as a real tool: latency, timeout (accept and never answer), slicer, bandwidth, reset_peer, down. Driven over HTTP from the verify scripts and from Go tests in lesson 11. |
| Instrumentation | **Prometheus** (`prom/prometheus`) scraping everything | Pull-based metrics are the natural fit for an HTTP service in Go, and HAProxy, Toxiproxy, postgres-exporter all speak it. Verification scripts query the HTTP API (`/api/v1/query`) for "a metric that changes". **TimescaleDB** was considered and not used: nothing in the design produces time-series the service itself must store; it would mean writing our own ingestion for what Prometheus does out of the box. Grafana is an optional Compose profile (`--profile ui`) the builder adds only if the budget allows; nothing depends on it. |
| Run analysis | **DuckDB** (CLI in a small image, run via `docker compose run report`) | The load generator writes one CSV row per request (`ts_ms, scenario, route, status, latency_ms, err_class, instance`). Per-window p50/p99, error and timeout rates, and before/after comparisons are analytical SQL over files, which is exactly DuckDB's use (`read_csv_auto`, `quantile_cont`). |
| SQLite | **not used** | Considered as the results store; the use is analytical, not transactional, and DuckDB fits it better. Nothing in the service needs an embedded transactional store: the service's data must be shared by three instances (Postgres). Saying so here so the omission is a decision. |
| Load generator | own Go binary, **open-loop** (fixed arrival rate, not closed-loop) | Open-loop load is what exposes overload: queues grow instead of the client politely waiting. Scenarios by name (`steady`, `oos`, `reports`, `whale`, `burst`, `rollout`), client timeout 3 s, strict JSON reader in v0 (that strictness is flaw F16). |
| Images | multi-stage `golang:1.26-alpine` → `alpine` | Seconds to rebuild after the first module download. First build needs network. |

Resource sizing on this VM (4 cores, 8 GB): checkout containers get a CPU limit (`cpus: "0.5"`) so
that one instance saturates at a load the generator can comfortably exceed. `make calibrate` measures
one instance's capacity and writes `lab/capacity.env`; scenario rates are multiples of it. That is
itself a Ch. 15 lesson (know your capacity before you load test).

### 2.7 What v0 does right (so the flaws are believable)

v0 is not a strawman. It has a clean package layout, unit tests that pass, a Compose file that comes up
with one command, idempotency keys, an audit trail, a health endpoint, and it serves `steady` load at
half capacity with p99 under 50 ms. Its tests use an `httptest` stub for pricing that always answers
in 1 ms. It is the service that passed QA.

## 3. Deliberate-issues register

Every flaw planted in v0, where it lives, how it shows, which chapter names it, and which lesson fixes
it. This is the answer key for the `hunt-stability-antipatterns` skill (section 6).

| ID | Flaw in v0 (file / area) | How it shows under the right load | Book | Fixed in |
|---|---|---|---|---|
| F01 | `/health` returns `200 ok` unconditionally; no `/metrics`; `log.Printf("handled %s", path)` with no request id, no durations, no dependency attribution (`cmd/checkout/main.go`, `internal/api`) | Pricing becomes slow: users see latency, the balancer sees three healthy instances, the logs say "handled", and nothing names the cause. Pause Postgres: every request 500s while every instance stays UP. | Ch. 8 Transparency; Ch. 6 case study; Ch. 10 System-Wide Transparency | L01 |
| F02 | `http.Client{}` with no timeout for pricing and notify; default `Transport` with no dial/response-header timeout (`internal/pricing/client.go`, `internal/notify/client.go`) | Pricing accepts and never answers: every checkout goroutine waits forever; in-flight and goroutines climb linearly; nothing recovers when the partner does. | Ch. 4 Integration Points, Blocked Threads; Ch. 5 Timeouts | L02 |
| F03 | All DB calls use `context.Background()`; the request's context is never passed down; no `statement_timeout` (`internal/store`) | A slow query or a saturated pool holds the handler past the client's patience; the server keeps working on requests nobody is waiting for. | Ch. 5 Timeouts; Ch. 4 Slow Responses | L02 |
| F04 | `http.Server{}` with no `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout`; HAProxy `timeout server 5m` (`deploy/haproxy.cfg`) | Slow or stalled clients hold goroutines and balancer sessions for minutes. | Ch. 4 Integration Points (socket-based protocols); Ch. 5 Timeouts | L02 |
| F05 | In `ReserveStock`, the out-of-stock branch returns before `rows.Close()`; the connection stays checked out for the life of the process (`internal/store/inventory.go`) | 30 % out-of-stock orders: one connection leaks per rejected order; `pg_stat_activity` climbs; at 100 Postgres says `too many clients already` and every instance (they share the DB) starts failing. The airline bug (Ch. 2) in Go clothing. | Ch. 2 case study; Ch. 4 Blocked Threads | L03 |
| F06 | `sql.DB` pool unbounded (`SetMaxOpenConns` never called), no `SetConnMaxLifetime`, no pool metrics | Combines with F05 to hit `max_connections`; combined with load, every instance opens as many connections as it has goroutines. | Ch. 4 Blocked Threads, Unbalanced Capacities; Ch. 5 Fail Fast | L03 |
| F07 | `POST /orders` calls pricing before validating the items and before checking the idempotency key | Invalid and duplicate requests cost a partner call and a pool connection before being rejected. | Ch. 5 Fail Fast | L03 |
| F08 | Pricing client retries 5× on any error with a fixed 100 ms sleep, including on timeouts, with no budget; notify client retries a non-idempotent POST (`internal/pricing/client.go`, `internal/notify/client.go`) | Pricing half-degraded (latency above the timeout on 50 % of connections): every timeout becomes 6 attempts; pricing's queue fills; everything times out; after the fault is removed pricing stays saturated by the retry backlog (dogpile on recovery). Duplicate confirmations are sent. | Ch. 4 Dogpile, Self-Denial Attacks, Force Multiplier; Ch. 5 Governor | L04 |
| F09 | No circuit breaker; no degraded mode when pricing is down | Full pricing outage: every checkout burns the whole timeout plus its retries, pricing is hammered while down, the thread budget is spent on a dead call, and recovery is noticed only by luck. | Ch. 5 Circuit Breaker; Ch. 4 Cascading Failures | L05 |
| F10 | One `sql.DB` pool and no concurrency partition shared by `POST /orders` and `GET /reports/daily`; `/health` pings the DB through that same pool (`internal/store`, `internal/api`) | 20 concurrent report refreshes (the back-office dashboard) hold every pool connection for seconds; checkouts queue for a connection and time out; the health check times out too and the balancer drops the instance. | Ch. 5 Bulkheads; Ch. 4 Chain Reactions, Cascading Failures | L06 |
| F11 | No limit on in-flight requests per instance; HAProxy has no per-server `maxconn`/`timeout queue`; no 503 is ever returned on purpose (`cmd/checkout/main.go`, `deploy/haproxy.cfg`) | Offered load above capacity: queues grow without bound, latency passes the client timeout, goodput falls toward zero while the servers are "up". Kill one of three instances at 80 % utilization: the survivors go to 120 %, their goodput collapses, and the fleet delivers nothing. | Ch. 4 Chain Reactions, Slow Responses, Unbalanced Capacities; Ch. 5 Shed Load, Handshaking; Ch. 9 Demand Control | L07 |
| F12 | `GET /orders?customer_id=` has no `LIMIT`; `GET /reports/daily` loads every row of the day into Go and aggregates in memory (`internal/store/orders.go`, `internal/api/reports.go`) | The whale customer's list is 200 k rows / 40 MB of JSON; 20 concurrent requests exceed the container's `mem_limit`, the instance is OOM-killed, Compose restarts it, the balancer sends the next whale request to the next instance. | Ch. 4 Unbounded Result Sets, Slow Responses; Ch. 5 Let It Crash (what makes the restart survivable) | L08 |
| F13 | Order-confirmation queue is an unbounded slice behind a mutex drained by one goroutine; nothing is persisted; work is lost on restart (`internal/notify/queue.go`) | Notify partner at 1 s latency: producer 100/s vs consumer 1/s; depth grows 100/s, the oldest message ages into minutes, memory grows, a restart loses every queued confirmation. | Ch. 5 Create Back Pressure; Ch. 4 Unbalanced Capacities | L09 |
| F14 | `audit_log` rows written on every request, never purged; request log with full bodies written to a file in the container with no rotation; in-process idempotency map never evicted (`internal/store/audit.go`, `cmd/checkout/main.go`, `internal/api/idempotency.go`) | Everything grows linearly with traffic; the log volume (a small tmpfs in the lab) fills during one run and logging stops silently; extrapolated, the table and the map exhaust disk and memory in days. | Ch. 5 Steady State (data purging, log files, in-memory caching) | L10 |
| F15 | Tests cover only the happy path against an `httptest` stub; the pricing client reads the body with `io.ReadAll` and decodes it regardless of status code; no `Content-Type` check (`internal/pricing/client_test.go`, `client.go`) | A 500 with an HTML body is logged as "invalid quote"; a 10 MB body is held in memory; connection resets and slow headers were never exercised, so the first time they happen is in production. | Ch. 5 Test Harnesses; Ch. 4 Integration Points; Ch. 15 The Testing Gap | L11 |
| F16 | Strict JSON readers everywhere: the pricing client and the load generator (the stand-in consumer) use `DisallowUnknownFields` and require exact field sets | The partner adds a field (`currency`) in its v2: every checkout fails. The service adds a field to its own response: the consumer fails. | Ch. 14 Handling Versions (handle others' versions; help others handle yours) | L12 |
| F17 | Migrations run inside `main()` at startup, so DDL ships in the same deploy as the code; the planned money change is an in-place `ALTER COLUMN ... TYPE bigint` (full table rewrite under `ACCESS EXCLUSIVE`) | A rolling deploy of the money change blocks every checkout for the rewrite's duration, health checks fail, instances are dropped, and the three instances race to run the same DDL. | Ch. 12 case study (Waiting for Godot); Ch. 13 Phases of Deployment; Ch. 14 | L12 |

Seventeen flaws, twelve lessons. F02–F04 are one lesson because they are one idea (bound every wait);
F05–F07 are one lesson because they are the one case study (Ch. 2); F16–F17 are one lesson because
expand/contract and tolerant readers are two halves of one rollout.

## 4. The lesson sequence

### 4.1 Shape of the history

```
release-it/v0-flawed      release-it: add the flawed checkout project (v0)
release-it/lesson-01      checkout: expose health, metrics and structured logs (lesson 01)
release-it/lesson-02      checkout: bound every integration point with timeouts and deadlines (lesson 02)
release-it/lesson-03      checkout: return connections on every path and bound the pool (lesson 03)
release-it/lesson-04      checkout: retry with jittered backoff and a budget (lesson 04)
release-it/lesson-05      checkout: add a circuit breaker around pricing (lesson 05)
release-it/lesson-06      checkout: bulkhead reports and health away from checkout (lesson 06)
release-it/lesson-07      checkout: shed load at the instance and at the balancer (lesson 07)
release-it/lesson-08      checkout: bound result sets and aggregate in SQL (lesson 08)
release-it/lesson-09      checkout: bounded notification queue with back pressure (lesson 09)
release-it/lesson-10      checkout: reach steady state: purge, rotate, evict (lesson 10)
release-it/lesson-11      checkout: chaos test harness against toxiproxy (lesson 11)
release-it/lesson-12      checkout: expand/contract migration and tolerant readers (lesson 12)
release-it/v1-hardened    release-it: lessons index and landing page
```

**Order.** Transparency comes first, before the book's Part I order, because every later verification
reads the metrics it adds, and because that is what a senior engineer does with an opaque system
(Ch. 6: the retailer could not be diagnosed without vital signs). Then the Part I antipatterns in the
order their fixes compose: timeouts make failures finite; a bounded pool makes them local; backoff
stops them amplifying; the breaker stops them repeating; bulkheads stop them spreading between
paths; shedding stops them spreading between instances. Then the data lessons (results, queue,
steady state), then the harness that proves it all, then versions. Retries (L04) come before the
breaker (L05) on purpose: with a breaker already in place, a retry storm cannot be demonstrated.

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
  `docker compose kill/pause` where the lesson needs it), collects `results/NN-*.csv`, runs the DuckDB
  report, queries Prometheus for the metrics the lesson names, and prints a verdict with explicit
  thresholds (`IMPROVED: p99 during outage 1.02 s ≤ 1.2 s` or `FLAW REPRODUCED: ...`). Exit 0 when
  the lesson's thresholds hold, 1 when the flaw is reproduced.
- `make verify LESSON=NN REF=release-it/lesson-MM` runs the same script with the **`checkout`
  image built from another commit** (`git archive REF project/` into a temp build context) while
  everything that is harness rather than codebase (partner fakes, load generator, scenarios, report
  SQL, thresholds) comes from the current tree. So on the `lesson-05` checkout, `make verify
  LESSON=05` passes and `make verify LESSON=05 REF=release-it/lesson-04` reproduces the flaw. This is
  the "test that failed before" for every lesson, and it is why lessons 11 and 12 can point the
  current partner fakes (new admin modes, partner v2) at the previous lesson's client.
- Report format (DuckDB over the CSV), one row per 5 s window per route:
  `window route req 2xx 4xx 5xx timeouts p50_ms p99_ms`, followed by the Prometheus readings the
  lesson names and the verdict lines. The same shape for every lesson so before/after compare line
  by line.
- Each lesson's thresholds are stated in its section below and in its script; builders may tune the
  numbers after calibration but must keep the direction and the named metric.

### 4.3 Lessons

#### L01 · Transparency: health, metrics, structured logs

- **Source:** Ch. 8 "Transparency" (designing for transparency; logging; instance metrics; health
  checks); Ch. 10 "System-Wide Transparency"; Ch. 6 case study.
- **Fixes:** F01.
- **Patch:** `/metrics` with `prometheus/client_golang`: `http_requests_total{route,method,status,
  client_version}`, `http_request_duration_seconds` histogram, `http_inflight_requests`,
  `dependency_requests_total{dep,outcome}` and `dependency_duration_seconds{dep}`, `db_pool_*`
  gauges from `sql.DBStats` (open, in use, idle, wait count, wait duration), `notify_queue_depth`,
  `notify_queue_oldest_age_seconds`, `idempotency_cache_entries`, Go runtime defaults. `/health`
  returns JSON `{status: ok|degraded|failed, checks:{db, pool, notify_queue, ...}}`, 200 for ok and
  degraded, 503 for failed; `failed` when the DB ping (500 ms budget) fails; `degraded` when pool wait
  p99 or queue age cross a threshold. HAProxy switches to `option httpchk GET /health` with
  `http-check expect status 200`. `log/slog` JSON to stdout with request id (propagated as
  `X-Request-Id`), route, status, duration, instance, and one line per dependency call with its
  duration and outcome. Prometheus scrape config already lists the checkout targets in v0 (ops set it
  up; the service never reported), so `up{job="checkout"}` flips 0 → 1.
- **Verify:** `steady` at 50 % capacity for 60 s; at t=15 s add 800 ms latency on pricing; at t=35 s
  `docker compose pause postgres` for 10 s. Before: the only signal is client-side latency, then
  500s; HAProxy shows 3/3 UP throughout; logs are `handled POST /orders`. After:
  `dependency_duration_seconds{dep="pricing"}` p99 rises at t=15 s while `http_request_duration`
  shows the same rise, attributing it; `/health` goes `failed: db timeout` on all three within two
  checks of the pause and HAProxy marks them DOWN with the reason; logs carry request ids and
  `dep=pricing dur_ms=812`. Threshold: `up{job="checkout"} == 3`, health reaches `failed` within 3 s of
  the pause, every log line parses as JSON with a `request_id`.
- **Size:** ~300 lines (the largest patch after v0); reviewable because it touches middleware, a
  metrics package, and the health handler and nothing else.
- **Note for later lessons:** the health check uses the shared pool on purpose; L06 revisits that.

#### L02 · Timeouts on every integration point

- **Source:** Ch. 4 "Integration Points" (socket-based protocols; HTTP protocols); Ch. 5 "Timeouts";
  Ch. 4 "Slow Responses".
- **Fixes:** F02, F03, F04.
- **Patch:** pricing and notify clients get a `Transport` with `DialContext` timeout 200 ms,
  `ResponseHeaderTimeout` 1 s, `TLSHandshakeTimeout`, and an overall `http.Client.Timeout`; every
  call takes a `context.Context`. Handlers derive a per-request deadline (`POST /orders`: 3 s) and
  pass `r.Context()` into the store (`QueryContext`, `ExecContext`, `BeginTx`) and into the partner
  clients; the pricing call gets a 1 s budget inside that. Postgres pool gets
  `statement_timeout=2s` via connection parameters. `http.Server` gets
  `ReadHeaderTimeout 5s, ReadTimeout 10s, WriteTimeout 15s, IdleTimeout 60s`; HAProxy
  `timeout client/server 10s, timeout http-request 5s`. Timed-out dependency calls return 504
  `pricing timeout` with the request id.
- **Verify:** `steady` at 50 % for 45 s; at t=10 s add a `timeout` toxic (accept, never answer) on
  pricing; remove at t=30 s. Before (at `lesson-01`): every checkout hangs until the generator gives
  up at 3 s; `http_inflight_requests` and `go_goroutines` climb linearly (~1 500 by t=30 s) and do
  not fall when the toxic is removed. After: each checkout fails at 1.0 s with 504; in-flight stays
  under 2 × rate × 1 s; goroutines flat; recovery immediate on toxic removal. Threshold: p99 during
  the outage ≤ 1.2 s; `max(http_inflight_requests)` ≤ 120.
- **Size:** ~150 lines across clients, store, handlers, server, haproxy.cfg.

#### L03 · Blocked threads: return the connection on every path, bound the pool, fail fast

- **Source:** Ch. 2 case study (the exception that grounded an airline); Ch. 4 "Blocked Threads";
  Ch. 5 "Fail Fast", "Timeouts".
- **Fixes:** F05, F06, F07.
- **Patch:** `defer rows.Close()` and `defer tx.Rollback()` idioms throughout the store; the
  out-of-stock path returns a typed `ErrOutOfStock` (→ 409) through the deferred cleanup.
  `SetMaxOpenConns(10)`, `SetMaxIdleConns(10)`, `SetConnMaxLifetime(30m)`, `SetConnMaxIdleTime(5m)`.
  Pool wait is bounded by the request context (from L02), so a saturated pool yields a fast 503
  `pool timeout` instead of a hang; `db_pool_wait_duration_seconds` becomes meaningful. Handler
  validates the body and checks the idempotency key before any partner call or connection checkout.
- **Verify:** `oos` scenario (30 % of orders request more than on-hand) at 50 % capacity for 60 s.
  Before (at `lesson-02`): `pg_stat_activity_count` rises one per rejected order; at ~100 Postgres
  refuses connections, `5xx` goes to 100 % on all three instances, `/health` reports `failed: db`, the
  fleet is down and stays down after the load stops (the leaked connections are never returned).
  After: connections flat at ≤ 30 (3 × 10), out-of-stock orders return 409 in < 20 ms, success rate
  of in-stock orders ≥ 99 %. Threshold: `max(pg_stat_activity_count) ≤ 35`; no 5xx.
- **Size:** ~120 lines.

#### L04 · Retries: jittered backoff, a budget, idempotent only

- **Source:** Ch. 4 "Dogpile", "Self-Denial Attacks", "Force Multiplier"; Ch. 5 "Governor",
  "Handshaking".
- **Fixes:** F08.
- **Patch:** `internal/retry`: exponential backoff with full jitter (`rand(0, base·2^n)`, base
  50 ms, cap 400 ms), max 2 retries, never retry past the request deadline, retry only idempotent
  calls (the pricing quote; not the notify POST, which gets an idempotency key instead), and a
  per-instance **retry budget** (token bucket: retries ≤ 10 % of recent requests). Metrics:
  `retry_attempts_total{dep,outcome}`, `retry_budget_exhausted_total`.
- **Verify:** `steady` at 60 % for 60 s; at t=10 s add a `latency` toxic of 1 500 ms with toxicity
  0.5 on pricing (half the connections exceed the 1 s budget); remove at t=30 s. Before (at
  `lesson-03`): `pricing_requests_total` rate is 4–6× the checkout rate, pricing's in-flight gauge is
  pegged at its queue limit, timeouts reach ~100 % (the healthy half is crowded out by retries of the
  slow half), and after t=30 s pricing stays saturated for 20 s or more before the error rate drops
  (dogpile on recovery). After: pricing request rate ≤ 1.3× checkout rate; errors ≈ the fraction the
  toxic actually breaks; recovery within 3 s of removal; zero duplicate confirmations (notify counts
  one call per order). Threshold: ratio ≤ 1.3; recovery ≤ 3 s.
- **Size:** ~140 lines.

#### L05 · Circuit breaker around pricing

- **Source:** Ch. 5 "Circuit Breaker"; Ch. 4 "Cascading Failures", "Integration Points".
- **Fixes:** F09.
- **Patch:** `internal/breaker`: closed → open after 5 consecutive failures (or 50 % failures in a
  10 s window), open → half-open after 5 s, one trial call, success closes. Wraps the pricing client.
  Open breaker returns 503 `pricing unavailable` in microseconds with `Retry-After`. Exposes
  `breaker_state{dep}` (0 closed, 1 open, 2 half-open) and `breaker_transitions_total`, logs each
  transition, and `/health` reports `degraded: pricing breaker open`. Optional degraded mode shown
  in the note as "try changing": quote from the last-known price table and flag the order for
  re-pricing.
- **Verify:** `steady` at 50 % for 60 s; at t=10 s `timeout` toxic on pricing (full outage); remove
  at t=40 s. Before (at `lesson-04`): every checkout spends its full 1 s budget (plus budgeted
  retries) before failing; pricing receives every attempt while down; p50 = 1 s for 30 s. After:
  breaker opens within ~1 s of the outage; checkouts fail in < 5 ms; pricing receives ≤ 1 probe per
  5 s; breaker closes within 5 s of recovery and `breaker_transitions_total` shows
  closed→open→half-open→closed. Threshold: p50 during outage ≤ 10 ms; pricing requests during
  outage ≤ 10.
- **Size:** ~150 lines.

#### L06 · Bulkheads: separate pools for checkout, reports and health

- **Source:** Ch. 5 "Bulkheads"; Ch. 4 "Chain Reactions", "Cascading Failures", "Blocked Threads".
- **Fixes:** F10.
- **Patch:** two `sql.DB` pools from the same DSN: `oltp` (8 connections) for orders, `reporting`
  (2 connections, `statement_timeout=5s`) for `/reports/daily`; a per-route in-flight limit for
  reports (semaphore 2, 503 beyond). `/health` pings on a dedicated single connection with its own
  200 ms budget so a saturated pool cannot make a working instance look dead. Metrics gain a `pool`
  label. The note explains the instance-level bulkhead (route `/reports` to a separate instance set
  with one HAProxy ACL) as the next step up.
- **Verify:** `reports` scenario: `steady` checkouts at 50 % plus 20 concurrent report requests in a
  loop for 45 s. Before (at `lesson-05`): checkout p99 climbs past 2 s and `pool timeout` 503s appear;
  HAProxy marks instances DOWN because `/health` times out waiting for a connection; the dashboard
  took the shop down. After: checkout p99 ≤ 60 ms throughout; reports serve 2 at a time and return
  503 beyond that; health never leaves `ok`. Threshold: checkout p99 ≤ 100 ms; zero checkout 5xx;
  HAProxy `haproxy_server_up == 1` for all three throughout.
- **Size:** ~130 lines.

#### L07 · Shed load: bounded in-flight, 503 fast, balancer queue limits (chain reactions)

- **Source:** Ch. 4 "Chain Reactions", "Slow Responses", "Unbalanced Capacities", "Scaling Effects";
  Ch. 5 "Shed Load", "Handshaking", "Fail Fast"; Ch. 9 "Demand Control".
- **Fixes:** F11.
- **Patch:** `internal/shed`: a per-instance in-flight limit (default from calibration, ~1.5 × the
  concurrency needed at capacity) applied as middleware: beyond it, respond 503 with `Retry-After: 1`
  in microseconds, counted in `shed_total{route}`; critical routes get a larger share than reports.
  HAProxy: per-server `maxconn` matching the instance limit, `timeout queue 500ms` (503 from the
  edge when every server is at its limit), `retry-on conn-failure` only. `/health` reports
  `degraded` while shedding. Load generator treats 503 as a fast failure (it is the handshake).
- **Verify:** `burst` scenario: offered load 0.8 × fleet capacity for 15 s, then `docker compose
  kill checkout-b` at t=15 s, load unchanged to t=60 s, then `start` at t=60 s. Before (at
  `lesson-06`): after the kill the survivors take 1.2 × their capacity; `http_inflight_requests` and
  goroutines climb to the arrival rate times the 3 s deadline (thousands, bounded only by the
  clients giving up); latency passes 3 s; goodput (2xx within 3 s) falls to near zero
  across the fleet; when `checkout-b` returns it is flooded by the queued backlog and the fleet takes
  tens of seconds to recover. After: survivors serve at capacity (~2/3 of previous goodput), excess is
  rejected in < 5 ms with 503, p99 of successful requests stays ≤ 100 ms, in-flight is capped, the
  returning instance rejoins cleanly. Threshold: goodput after the kill ≥ 55 % of before; p99 of 2xx
  ≤ 150 ms; `max(http_inflight_requests)` ≤ limit.
- **Size:** ~120 lines plus haproxy.cfg.

#### L08 · Unbounded result sets: limit, paginate, aggregate in the database

- **Source:** Ch. 4 "Unbounded Result Sets" (and "Slow Responses"); Ch. 5 "Let It Crash" for why the
  OOM restart is survivable only because of L01 and L07.
- **Fixes:** F12.
- **Patch:** `GET /orders?customer_id=` gets `LIMIT 100` with keyset pagination (`?after=<id>`,
  `next` link in the response); `GET /reports/daily` becomes one SQL aggregate (`SUM`, `COUNT`,
  `GROUP BY sku ORDER BY ... LIMIT 20`) on the reporting pool with its `statement_timeout`, with a
  covering index added by migration `002_report_index.sql` (`CREATE INDEX CONCURRENTLY`, run by
  `make migrate`, not at startup; this foreshadows L12). Response bodies are streamed with
  `json.Encoder`, not built in memory.
- **Verify:** `whale` scenario: `steady` at 30 % plus 20 concurrent `GET /orders?customer_id=whale`
  for 45 s, checkout containers at `mem_limit: 256m`. Before (at `lesson-07`): each whale request
  builds ~40 MB; `process_resident_memory_bytes` spikes; an instance is OOM-killed within ~10 s
  (`docker inspect` RestartCount > 0), HAProxy routes around it, the next instance gets the whale,
  and checkouts fail during each restart. After: whale requests return 100 rows in < 30 ms with a
  `next` cursor; RSS flat; zero restarts; `/reports/daily` returns in < 200 ms. Threshold: RestartCount
  = 0; whale p99 ≤ 100 ms; report p99 ≤ 500 ms.
- **Size:** ~140 lines plus one migration.

#### L09 · Back pressure: a bounded notification queue that cannot lose work

- **Source:** Ch. 5 "Create Back Pressure", "Shed Load" (inside vs at the boundary); Ch. 4
  "Unbalanced Capacities"; Ch. 5 "Decoupling Middleware" as the next step (an outbox is a tiny
  broker).
- **Fixes:** F13.
- **Patch:** bounded channel (256) with 4 workers calling notify (each with its own L02 timeout and
  L04 idempotency key). The handler enqueues with a 50 ms budget: on success the order is marked
  `notified_at` when the worker finishes; on timeout (queue full) the handler returns 201 anyway and
  leaves `notified_at NULL`, which a **sweeper** (every 5 s, `LIMIT 100`, bounded rate) re-queues.
  So back pressure is applied to the producer inside the boundary (the enqueue budget) without
  blocking the customer, and nothing is lost on restart because the database is the queue of record.
  Metrics: `notify_queue_depth` (now ≤ 256), `notify_enqueue_timeouts_total`, `notify_sweeper_requeued_total`.
- **Verify:** `steady` at 50 % for 60 s with a 1 000 ms `latency` toxic on notify from t=0 to t=30 s;
  at t=20 s `docker compose restart checkout-a`. Before (at `lesson-08`): depth grows ~100/s to
  thousands; `notify_queue_oldest_age_seconds` reaches minutes; after the restart `SELECT count(*)
  FROM orders WHERE notified_at IS NULL` never returns to 0 (those confirmations are gone). After:
  depth ≤ 256, age ≤ 10 s, checkout latency unaffected by the slow partner, and within 30 s of
  removing the toxic every order has `notified_at` set, including the ones queued before the
  restart. Threshold: `max(notify_queue_depth) ≤ 256`; unnotified count = 0 at t=60 s.
- **Size:** ~150 lines.

#### L10 · Steady state: purge data, rotate logs, bound the cache

- **Source:** Ch. 5 "Steady State" (data purging; log files; in-memory caching; avoid fiddling).
- **Fixes:** F14.
- **Patch:** a purge job (`cmd/checkout purge`, also run by a Compose `cron`-style sidecar every
  minute in the lab, daily in production) deletes `audit_log` rows older than a retention window in
  batches of 5 000 with a sleep between batches, and the audit row stops storing full bodies
  (request id, route, status, summary fields only). Request logging goes to stdout only; the
  container's log driver sets `max-size: 10m, max-file: 3`; the in-container log file goes away.
  The idempotency map becomes an LRU with TTL (`hashicorp/golang-lru/v2`, 10 000 entries, 24 h) and
  `/health` reports its size. Metrics: `audit_log_rows` (from postgres-exporter),
  `idempotency_cache_entries`.
- **Verify:** `steady` at 60 % for 90 s with the log volume mounted as a 16 MB tmpfs. Before (at
  `lesson-09`): the script samples at 0/30/60/90 s and prints growth per minute for the table size
  (`pg_total_relation_size('audit_log')`), the log directory, and the cache gauge, then extrapolates
  time-to-exhaustion for the VM's disk and the container's memory limit (hours to days); the tmpfs
  hits 100 % before 90 s and the request log stops silently (the log you would need at 5 a.m. is the
  first casualty). After: table size flat after the first purge, log directory empty, cache gauge
  plateaus at ≤ 10 000, every extrapolation reads "never". Threshold: growth over the last 30 s window
  = 0 for all three.
- **Size:** ~160 lines.

#### L11 · Test harness: chaos tests against Toxiproxy and a misbehaving partner

- **Source:** Ch. 5 "Test Harnesses"; Ch. 4 "Integration Points"; Ch. 15 "The Testing Gap";
  Ch. 17 (the verify scripts are chaos experiments; this lesson automates them).
- **Fixes:** F15.
- **Patch:** `test/chaos/` Go tests (build tag `chaos`, run by `make test-chaos` against the running
  stack) using the Toxiproxy Go client: for each of `latency`, `timeout`, `slicer`, `bandwidth 1 KB/s`,
  `reset_peer`, `down`, the test applies the toxic, calls the pricing client, and asserts the error
  class and that the call returned within its budget, plus a test that the breaker opens and
  recovers. The `pricing` fake gains `/admin/mode` for application-level misbehavior (`500`,
  `malformed-json`, `huge-body` 10 MB, `slow-headers`), and the client is hardened to match:
  `http.MaxBytesReader`/`io.LimitReader` 1 MB, status-code check before decode, `Content-Type`
  check, classified errors (`ErrTimeout`, `ErrRemote5xx`, `ErrBadPayload`). CI script
  `make test` runs unit tests; `make test-chaos` runs the harness suite in ~60 s.
- **Verify:** `make test-chaos`. Before (at `lesson-10`, running the new tests against the old
  client): 4 of 10 tests fail (`huge-body` holds 10 MB per call, `500` with HTML is misclassified as a
  bad quote, `malformed-json` and `slow-headers` give unclassified errors). After: 10 of 10 pass. The
  script also prints the harness matrix (misbehavior × outcome × elapsed) as its table.
- **Size:** ~300 lines, mostly tests; reviewable because tests are table-driven and the client
  change is ~60 lines.

#### L12 · Handling versions: expand/contract migration and tolerant readers

- **Source:** Ch. 14 "Handling Versions" (help others handle your versions; handle others'
  versions); Ch. 13 "Phases of Deployment" (relax between phases; expansion; rollout; cleanup) and
  "The Fallacy of Planned Downtime"; Ch. 12 case study.
- **Fixes:** F16, F17.
- **Patch (two commits, see size note):**
  1. *Migrations become a job, money expands.* Migrations move out of `main()` into `cmd/checkout
     migrate`, run as a one-shot Compose service (`make migrate`) before a rollout, never at
     startup. `003_money_expand.sql` adds nullable `subtotal_cents bigint`, `tax_cents bigint`,
     `currency text`; the service dual-writes (old `total_cents` and the new columns) and reads
     new-or-old; `cmd/checkout backfill` fills old rows in batches of 5 000 with a sleep;
     `004_money_contract.sql` (drops `total_cents`, adds `NOT NULL`) is shipped but guarded: it
     refuses to run while any row is unbackfilled or while `http_requests_total{client_version="v1"}`
     has been non-zero in the last day. `make rollout` performs a rolling restart a → b → c waiting
     for `/health` between each.
  2. *Tolerant readers and the additive API.* The pricing client drops `DisallowUnknownFields`,
     requires only the fields it uses, and validates types and ranges instead of shape. The load
     generator (the consumer) does the same first. The response adds `subtotal_cents`, `tax_cents`,
     `currency` and keeps `total_cents`; the request accepts `customer` and `customer_id`. The
     `pricing` fake gains `PRICING_VERSION=2` (adds `currency`, `tax_lines`, `quote_expires_at`).
- **Verify:** three parts. (a) `make verify LESSON=12` runs `rollout` scenario: `steady` at 50 %
  while `make migrate && make rollout` executes the expand migration and rolling restart; then
  `PRICING_VERSION=2` is applied to the partner. After: zero errors through the migration and the
  rollout, zero errors on the partner's v2, consumer keeps working with the new fields present.
  (b) Before, at `lesson-11`: the same script applies the partner v2 → 100 % checkout failures
  (`unknown field "currency"`). (c) `make godot` (runs on any tree) applies the naive in-place
  `ALTER COLUMN total_cents TYPE bigint` under `steady` load and prints the stall: checkouts block
  for the rewrite's duration, `/health` fails, HAProxy drops instances, Postgres lock waits visible.
  Threshold: zero 5xx during (a); (b) reproduces ≥ 99 % failures.
- **Size:** ~250 lines over two commits (~150 + ~100). Two commits because the migration mechanism
  and the API tolerance are separately reviewable and the book treats them as two topics; the tag
  `release-it/lesson-12` points at the second.

### 4.4 Coverage honesty

| v1 topic | v2 | Note |
|---|---|---|
| 01 integration points / timeouts | L02 | same idea, now on real clients, server and balancer |
| 02 blocked threads / resource pool | L03 | the airline bug reproduced on `database/sql` + Postgres `max_connections` |
| 03 circuit breaker | L05 | |
| 04 bulkheads | L06 | DB pools and per-route limits instead of semaphores; instance-level bulkhead described, not built |
| 05 chain reactions / shed load | L07 | real kill of one instance behind HAProxy |
| 06 dogpile / retries | L04 | moved before the breaker so the storm is demonstrable |
| 07 unbounded result sets | L08 | with a real OOM kill and restart |
| 08 back pressure | L09 | bounded queue plus durable fallback (outbox-style) |
| 09 steady state | L10 | audit table, log volume, idempotency cache |
| 10 test harness | L11 | Toxiproxy-driven tests plus a misbehaving partner fake |
| 11 transparency | L01 | moved first; everything else is measured through it |
| 12 handling versions | L12 | adds the Ch. 12/13 migration lesson (Godot) that v1 only cited |

Nothing dropped. Merged inside lessons: Fail Fast (L03, L07), Handshaking (L07: 503 + `Retry-After`
+ health checks), Governor (L04 retry budget), Let It Crash (L08 note), Decoupling Middleware (L09
note). Still deliberately omitted as lessons, for the v1 reasons: Foundations (Ch. 7), Security
(Ch. 11), Adaptation (Ch. 16); Chaos Engineering (Ch. 17) is the method of every verify script and
the `design-chaos-experiment` skill rather than a lesson of its own.

## 5. Skills phase, revisited

The v1 set of eight skills stands (`test-integration-points`, `hunt-stability-antipatterns`,
`suggest-stability-patterns`, `budget-timeouts-and-retries`, `review-capacity-headroom`,
`review-operational-transparency`, `plan-safe-release`, `design-chaos-experiment`; descriptions in
the archived plan, section 4, and in `skills/README.md`). Each now cites the project as its worked
example, and two gain a built-in exercise:

| Skill | Worked example in the project |
|---|---|
| `hunt-stability-antipatterns` | Run it against `release-it/v0-flawed`; the register in section 3 is the answer key. A skill run that finds fewer than 12 of the 17 is a skill that needs work. |
| `suggest-stability-patterns` | Its output for the register should reproduce the lesson sequence's patterns and placements. |
| `test-integration-points` | L11's harness matrix and `test/chaos` are the template output. |
| `budget-timeouts-and-retries` | L02's deadline budget (3 s → 1 s pricing → 2 s statement) and L04's policy. |
| `review-capacity-headroom` | `make calibrate` and L07's thresholds; the 3-instance kill is the headroom check. |
| `review-operational-transparency` | L01's metric list and health semantics as the checklist's reference answer. |
| `plan-safe-release` | L12's expand → rollout → contract with its guard query. |
| `design-chaos-experiment` | Every `lab/verify/NN-*.sh` is an experiment card made executable; the skill's template mirrors their structure (steady state, hypothesis, injection, observation, threshold). |

Skills are written after the lessons, by the builder or a following agent, under the unchanged skills
contract in `AGENTS.md`.

## 6. Build order and budget (brief for the Opus pass)

One delegated pass, three checkpoints, each reviewed by a fresh agent from a different family (Codex)
before the next starts. Fixups found in review are folded into the lesson they belong to and tags
re-pointed (history is rewritten freely until final acceptance).

**Checkpoint A — v0 and the lab (~2 500–3 000 lines).**
1. `project/`: module, `cmd/checkout` (handlers, store, pricing + notify clients, queue, idempotency,
   startup migrations), `cmd/pricing`, `cmd/notify`, migrations + seed, Dockerfiles,
   `deploy/compose.yml`, `haproxy.cfg`, `prometheus.yml`, `toxiproxy.json`, `Makefile`
   (`up/down/seed/calibrate/verify/report/test/logs`).
2. `lab/`: `cmd/loadgen` with the scenarios in 2.6, `lab/verify.sh` (the runner with `REF` support),
   `lab/report.sql` (DuckDB), `lab/calibrate.sh`, `lab/toxics.sh` helpers.
3. Plant every flaw in the register exactly where it says; the flaws must be natural code, not
   commented "BUG HERE". `lessons/00-the-flawed-start.md` is the only place they are listed.
4. Prove v0 runs: `make up && make seed && make verify LESSON=00` (the `steady` scenario at 50 %,
   passes), unit tests pass. Tag `release-it/v0-flawed`.
5. L01 (needed by every later verify). Tag.

**Checkpoint B — L02 through L07 (~800 lines of patches).** Each: patch, `lessons/NN-*.md`,
`lab/verify/NN-*.sh`, tag. Before tagging, run `make verify LESSON=NN` (pass) and
`make verify LESSON=NN REF=release-it/lesson-(NN-1)` (reproduces the flaw) and paste both verdict
lines into the commit body. Calibrate once at the start of B; store `lab/capacity.env` in git.

**Checkpoint C — L08 through L12, landing page, skills (~1 000 lines of patches, 8 skills).**
Same per-lesson routine. Then `lessons/README.md` (index with tags), `books/release-it/README.md`
(landing page), `STATE.md`, and the skills from section 5. Tag `release-it/v1-hardened`.

**Large patches and how they stay reviewable.** L01 (~300 lines: one metrics package, one middleware,
one health handler), L11 (~300 lines, table-driven tests + a 60-line client change), L12 (two
commits). Everything else is 100–160 lines. A patch that grows past ~400 lines is split into two
commits under the same lesson with the tag on the last.

**Rules for the builder.**
- Read `AGENTS.md`, then this plan, then `lessons/00` once it exists. The register and the lesson
  order are fixed; thresholds and sizes may be tuned after calibration, with the change recorded in
  the lesson note under "Deviations from plan" and in `STATE.md`.
- v0 must pass its own tests and `make verify LESSON=00`. A flaw that makes v0 fail under gentle
  load is planted wrong.
- Every verify script: idempotent, ≤ 3 minutes with warm images, prints the report table, the named
  Prometheus readings, and a verdict line that starts with `IMPROVED:` or `FLAW REPRODUCED:`.
- Chaos is injected only through Toxiproxy's API, `docker compose kill/pause/restart`, the partner
  fakes' admin endpoints, and SQL run by the script. Never through hidden code paths in the service.
- Commit subjects as in 4.1; bodies carry flaw, fix, both verdict lines. Trailer per repo rules.
- Budget check: if the pass is running long, finish checkpoints A and B fully, land them, and report;
  C can be a second pass. Do not thin out lessons to fit.

## 7. Reader's path (what the final artifact looks like in use)

```
git checkout release-it/v0-flawed && cd books/release-it/project
make up && make seed && make verify LESSON=00        # it works
git log --oneline release-it/v0-flawed..release-it/v1-hardened   # the story
git checkout release-it/lesson-03
make verify LESSON=03 REF=release-it/lesson-02       # FLAW REPRODUCED: ...
make verify LESSON=03                                # IMPROVED: ...
git diff release-it/lesson-02 release-it/lesson-03 -- project/   # the fix, in minutes
```

## 8. Repo layout v2 and disposition of the v1 scaffold

```
books/release-it/
  PLAN.md                           # this document
  PLAN-v1-per-topic-projects.md     # archived v1 (kept, not maintained)
  README.md                         # landing page: argument, lesson index with tags, skills index
  lessons/
    README.md                       # index: lesson, tag, flaw ids, verify command
    00-the-flawed-start.md          # tour of v0 and the register
    01-transparency.md … 12-handling-versions.md
  project/                          # the one codebase; the thing the commits improve
    README.md  Makefile  go.mod  go.sum
    cmd/checkout/  cmd/pricing/  cmd/notify/  cmd/loadgen/
    internal/api/  internal/store/  internal/pricing/  internal/notify/
    internal/metrics/ (L01)  internal/retry/ (L04)  internal/breaker/ (L05)  internal/shed/ (L07)
    migrations/  deploy/ (compose.yml, Dockerfile.*, haproxy.cfg, prometheus.yml, toxiproxy.json)
    lab/ (verify.sh, verify/NN-*.sh, report.sql, calibrate.sh, capacity.env, toxics.sh)
    results/ (gitignored CSVs)
    test/chaos/ (L11)
  skills/                           # unchanged index now; SKILL.md files in checkpoint C
```

The flawed v0 lives only in history (tag `release-it/v0-flawed`); the working tree at `main` is the
hardened final state. Readers reach any intermediate state by tag.

**Disposition of the v1 scaffold.** Remove `books/release-it/topics/` (twelve stub READMEs) in
checkpoint A: they describe standalone programs that will not be built, and keeping them would
mislead a reader into expecting twelve `run.sh` files. Their content survives in the archived plan,
and every idea they carried is in a lesson. Keep `books/release-it/skills/README.md` (the index
stands; only the worked-example column changes). Keep `books/release-it/README.md` but rewrite it in
checkpoint C. The root `README.md` and `AGENTS.md` are updated now to describe the v2 pattern, with
the scaffold removal noted as pending. Nothing in the tree is deleted in this phase.

## 9. Decisions Nick is asked to confirm

1. The archetype (checkout service with pricing and notify partners) and the stack in 2.6, in
   particular Go, HAProxy over Caddy/nginx, Prometheus over TimescaleDB, DuckDB yes, SQLite no,
   Grafana optional.
2. Transparency first (L01), retries before the breaker (L04 before L05), and the rest of the order.
3. Lesson docs as `lessons/NN-*.md` plus commit bodies; tags `release-it/v0-flawed`,
   `release-it/lesson-NN`, `release-it/v1-hardened`; history rewritable until final acceptance.
4. Lesson 12 as two commits; everything else one commit per lesson.
5. Removal of `topics/` in the build phase; skills index kept.
6. Three checkpoints with cross-family review between them, rather than a single uninterrupted pass.
