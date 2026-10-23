# Litefleet fidelity inventory

Snapshot: `release-it/v1-hardened` = `936786399d7bab3ebbfa845b9a07df02b41ed363`. Source code is the authority; the original plan describes some behavior that did not ship. This inventory and the JSON bundle cover only this snapshot. No book text is exported.

[inventory.json](inventory.json) is the complete audit: each element and relationship has a stable ID, boundary, protocol, pinned source selector and line, and its containing views. Every source selector is checked against the pinned git object by `tests/litefleet.test.ts`. Links in the bundle point directly at public commit-pinned code. This is builder fidelity evidence; independent Fable review is item 6.

The SDK requires external tags for supporting stores/containers in component views. Here that means outside the focused process, not outside the Litefleet system. Parent IDs and the Containers system boundary specify ownership. Deployment instances reuse logical containers; lab infrastructure nodes represent real tools and volumes. No schema extensions or SDK changes are needed.

Dynamic steps label conditional and alternative paths explicitly. The sequence shows the cache-miss read path, then documents the cache-hit shortcut. Usage steps show normal delivery followed by the alternative outbox path, not two required deliveries. Reverse steps are responses on the named edge.

## Views

| ID | Kind | Elements | Relationships |
|---|---|---:|---:|
| `context` | context | 5 | 4 |
| `containers` | container | 11 | 12 |
| `gateway-components` | component | 9 | 9 |
| `node-components` | component | 10 | 11 |
| `query-flow` | dynamic | 7 | 6 |
| `usage-flow` | dynamic | 5 | 5 |
| `lab-deployment` | deployment | 25 | 48 |

## Lesson-to-view map

| Lesson | Hardened behavior | Views | Pinned source paths and note |
|---|---|---|---|
| L01 | Metrics, health, structured logs | gateway-components, lab-deployment | [cmd/gateway/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/gateway/main.go), [deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml), [books/release-it/lessons/01-transparency.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/01-transparency.md) |
| L02 | Request, dependency and statement deadlines | gateway-components, node-components, query-flow | [internal/api/handlers.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/api/handlers.go), [internal/nodeclient/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/nodeclient/client.go), [internal/engine/exec.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/exec.go), [books/release-it/lessons/02-timeouts.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/02-timeouts.md) |
| L03 | Rollback and resource bounds | containers, node-components | [internal/engine/exec.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/exec.go), [internal/engine/handles.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/handles.go), [internal/registry/registry.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/registry/registry.go), [books/release-it/lessons/03-blocked-threads.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/03-blocked-threads.md) |
| L04 | Read retry policy; no uncertain write retry | gateway-components, query-flow | [internal/retry/retry.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/retry/retry.go), [internal/nodeclient/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/nodeclient/client.go), [books/release-it/lessons/04-retries.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/04-retries.md) |
| L05 | Route cache and circuit breaker | gateway-components, query-flow | [internal/routes/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/routes/client.go), [internal/routes/cache.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/routes/cache.go), [internal/breaker/breaker.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/breaker/breaker.go), [books/release-it/lessons/05-circuit-breaker.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/05-circuit-breaker.md) |
| L06 | Admission isolation, pool isolation, backup lock | containers, node-components | [internal/engine/admission.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/admission.go), [internal/backup/job.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/backup/job.go), [internal/registry/registry.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/registry/registry.go), [books/release-it/lessons/06-bulkheads.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/06-bulkheads.md) |
| L07 | Shedding at gateway, node and balancer | gateway-components, node-components, lab-deployment | [cmd/gateway/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/gateway/main.go), [cmd/node/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/main.go), [deploy/haproxy.cfg](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/haproxy.cfg), [books/release-it/lessons/07-shed-load.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/07-shed-load.md) |
| L08 | Capped streaming and SQL aggregation | node-components, query-flow, containers | [internal/engine/exec.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/exec.go), [internal/nodeclient/rows.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/nodeclient/rows.go), [internal/registry/usage.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/registry/usage.go), [books/release-it/lessons/08-unbounded-results.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/08-unbounded-results.md) |
| L09 | Bounded queue, durable overflow and graceful spill | node-components, usage-flow | [internal/metering/queue.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/metering/queue.go), [internal/metering/outbox.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/metering/outbox.go), [cmd/node/drain.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/drain.go), [books/release-it/lessons/09-back-pressure.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/09-back-pressure.md) |
| L10 | Checkpoint, idle eviction, ledger rollup/purge | node-components, lab-deployment | [internal/engine/steady.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/steady.go), [internal/registry/usage.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/registry/usage.go), [cmd/control/purge.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/control/purge.go), [books/release-it/lessons/10-steady-state.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/10-steady-state.md) |
| L11 | External fault harness | lab-deployment | [test/chaos/chaos_test.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/test/chaos/chaos_test.go), [lab/toxics.sh](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/lab/toxics.sh), [books/release-it/lessons/11-test-harness.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/11-test-harness.md) |
| L12 | Migrations, tolerant readers and drain (no draining retry) | gateway-components, node-components, lab-deployment | [cmd/control/jobs.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/control/jobs.go), [internal/registry/migrate.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/registry/migrate.go), [internal/nodeclient/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/nodeclient/client.go), [cmd/node/drain.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/drain.go), [books/release-it/lessons/12-handling-versions.md](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/lessons/12-handling-versions.md) |

## Behavior checked against code

- The registry usage ledger is populated by node heartbeats, not by metering callbacks. The metering partner has process-local deduplication, not durable global exactly-once semantics.
- The outbox protects overflow, failed batches and graceful shutdown spill. A hard crash can lose events in memory; outbox SQL errors are logged. Do not present a stronger delivery guarantee.
- Node-local state contains `node_info` and `usage_outbox`. Placement/epoch metadata lives in each tenant file's `_meta`; the original plan's claim of accepted placement rows in `_node.db` is not implemented.
- Remote resolve and pre-response reads use bounded retries. Unknown write outcomes and `draining` responses are not retried. A stale-epoch refusal permits re-resolve because execution did not start.
- Streaming is JSON transcoding in the gateway, not a raw `io.Copy`. Mid-stream errors abort the answer. Row and byte caps are on the node, with gateway validation too.
- Pools are 8 OLTP / 2 reporting / 1 health in the final registry; node cache is 128 entries. No replication, move, fencing protocol or node failover is modeled.
- Component responsibilities are grouped for reading: engine includes admission; backup includes blob client; outbox includes sweeper; maintenance includes drain. These are not separate services.
- Profile labels reflect Compose declarations. Hardened `lab/up.sh` runs migration before startup and enables maintenance through `lab/lib.sh`; the tool profiles are invoked on demand.

## Review limits

All explanation STE fields are drafts, with no certified language-compliance claim. Site tests verify rendering and navigation; pinned-object tests verify source coverage and selectors. Neither replaces the independent architectural review in item 6. The book's open review/history decisions remain in `books/release-it/STATE.md`.

## Complete ID index

| ID | Kind | Boundary | Protocol | Views | Sources |
|---|---|---|---|---|---|
| `tenant` | element | system/role | human role | context, containers, gateway-components, query-flow | src-1 |
| `operator` | element | system/role | human role | context, containers | src-2 |
| `litefleet` | element | system/role | Go | context | src-3 |
| `blob` | element | external | Go | context, containers, node-components | src-4 |
| `meter` | element | external | Go | context, containers, node-components, usage-flow | src-5 |
| `ctx-tenant` | relationship | process | HTTP/JSON | context | src-1 |
| `ctx-ops` | relationship | process | HTTP/JSON | context | src-6 |
| `ctx-backup` | relationship | process | HTTP PUT | context | src-7 |
| `ctx-usage` | relationship | process | HTTP/JSON | context | src-8 |
| `ingress` | element | litefleet | HAProxy 3.2 | containers, query-flow | src-9 |
| `gateway` | element | litefleet | Go / net/http | containers, query-flow | src-10 |
| `control` | element | litefleet | Go / net/http | containers, gateway-components, query-flow | src-11 |
| `registry` | element | litefleet | PostgreSQL 17 | containers, query-flow | src-12 |
| `node` | element | litefleet | Go / net/http | containers, gateway-components, query-flow | src-13 |
| `tenant-files` | element | litefleet | SQLite / filesystem | containers, node-components, query-flow | src-14 |
| `node-state` | element | litefleet | SQLite / filesystem | containers, node-components, usage-flow | src-15 |
| `tenant-edge` | relationship | process | HTTP/JSON | containers, query-flow | src-16 |
| `ops-edge` | relationship | process | HTTP/JSON | containers | src-6 |
| `balance` | relationship | process | HTTP/JSON | containers, query-flow | src-9 |
| `resolve` | relationship | process | HTTP/JSON | containers, query-flow | src-17 |
| `forward` | relationship | process | HTTP/JSON | containers, query-flow | src-18 |
| `provision` | relationship | process | HTTP/JSON | containers | src-19 |
| `heartbeat` | relationship | process | HTTP/JSON | containers | src-20 |
| `node-backup` | relationship | process | HTTP PUT / SQLite snapshot | containers | src-21 |
| `node-meter` | relationship | process | HTTP/JSON | containers | src-8 |
| `sql-registry` | relationship | process | PostgreSQL wire / SQL | containers, query-flow | src-22 |
| `sql-tenant` | relationship | node-local storage | SQLite / local file I/O | containers, query-flow | src-23 |
| `sql-state` | relationship | node-local storage | SQLite / local file I/O | containers | src-24 |
| `gw-api` | element | gateway | Go | gateway-components | src-25 |
| `gw-routes` | element | gateway | Go | gateway-components | src-17 |
| `gw-breaker` | element | gateway | Go | gateway-components | src-26 |
| `gw-node` | element | gateway | Go | gateway-components | src-27 |
| `gw-retry` | element | gateway | Go | gateway-components | src-28 |
| `gw-shed` | element | gateway | Go | gateway-components | src-29 |
| `g-tenant-shed` | relationship | process | HTTP/JSON | gateway-components | src-30 |
| `g-shed-api` | relationship | gateway process | Go call | gateway-components | src-30 |
| `g-api-route` | relationship | gateway process | Go call | gateway-components | src-31 |
| `g-route-breaker` | relationship | gateway process | Go call | gateway-components | src-32 |
| `g-breaker-control` | relationship | process | HTTP/JSON | gateway-components | src-33 |
| `g-route-retry` | relationship | gateway process | Go call | gateway-components | src-34 |
| `g-api-node` | relationship | gateway process | Go call | gateway-components | src-35 |
| `g-node-retry` | relationship | gateway process | Go call | gateway-components | src-36 |
| `g-client-node` | relationship | process | HTTP/JSON | gateway-components | src-18 |
| `nd-engine` | element | node | Go | node-components, usage-flow | src-37 |
| `nd-handles` | element | node | Go | node-components | src-38 |
| `nd-backup` | element | node | Go | node-components | src-21 |
| `nd-usage` | element | node | Go | node-components, usage-flow | src-39 |
| `nd-outbox` | element | node | Go | node-components, usage-flow | src-40 |
| `nd-maintain` | element | node | Go | node-components | src-41 |
| `n-engine-handles` | relationship | node process | Go call | node-components | src-42 |
| `n-handles-files` | relationship | node process / dependency | SQLite / file I/O | node-components | src-43 |
| `n-backup-handles` | relationship | node process | Go call | node-components | src-44 |
| `n-backup-blob` | relationship | node process / dependency | HTTP PUT / SQLite snapshot | node-components | src-45 |
| `n-engine-usage` | relationship | node process | Go call | node-components, usage-flow | src-46 |
| `n-usage-outbox` | relationship | node process | Go call | node-components, usage-flow | src-47 |
| `n-outbox-state` | relationship | node process / dependency | SQLite / file I/O | node-components, usage-flow | src-24 |
| `n-queue-meter` | relationship | node process / dependency | HTTP/JSON | node-components, usage-flow | src-48 |
| `n-sweep-meter` | relationship | node process / dependency | HTTP/JSON | node-components, usage-flow | src-40 |
| `n-maintain-handles` | relationship | node process | Go call | node-components | src-49 |
| `n-drain-queue` | relationship | node process | Go call | node-components | src-50 |
| `lab` | element | system/role | Docker Compose | lab-deployment | src-51 |
| `gateway-a` | element | lab | Go | lab-deployment | src-52 |
| `gateway-b` | element | lab | Go | lab-deployment | src-53 |
| `gateway-c` | element | lab | Go | lab-deployment | src-54 |
| `node-a` | element | lab | Go | lab-deployment | src-55 |
| `node-b` | element | lab | Go | lab-deployment | src-56 |
| `control-run` | element | lab | Go | lab-deployment | src-57 |
| `postgres-run` | element | lab | Go | lab-deployment | src-58 |
| `haproxy-run` | element | lab | Go | lab-deployment | src-59 |
| `blob-run` | element | lab | Go | lab-deployment | src-60 |
| `meter-run` | element | lab | Go | lab-deployment | src-61 |
| `toxiproxy` | element | lab | Toxiproxy 2.12 | lab-deployment | src-62 |
| `prometheus` | element | lab | Prometheus 3.7 | lab-deployment | src-63 |
| `exporter` | element | lab | postgres-exporter 0.18.1 | lab-deployment | src-64 |
| `pgdata` | element | lab | Docker named volume | lab-deployment | src-65 |
| `node-a-data` | element | lab | Docker named volume | lab-deployment | src-66 |
| `node-b-data` | element | lab | Docker named volume | lab-deployment | src-67 |
| `blobs` | element | lab | Docker named volume | lab-deployment | src-68 |
| `migrate` | element | lab | Go / optional job | lab-deployment | src-69 |
| `purger` | element | lab | Go / optional sidecar | lab-deployment | src-70 |
| `seedgen` | element | lab | Go / one-shot lab tool | lab-deployment | src-71 |
| `loadgen` | element | lab | Go / one-shot lab tool | lab-deployment | src-72 |
| `report` | element | lab | DuckDB / lab tool | lab-deployment | src-73 |
| `harness` | element | lab | Shell / Go / Docker CLI | lab-deployment | src-2 |
| `results` | element | lab | Host filesystem | lab-deployment | src-74 |
| `d-lb-gateway-a` | relationship | process | HTTP :8080 | lab-deployment | src-75 |
| `d-gateway-a-proxy` | relationship | process | HTTP over TCP proxies | lab-deployment | src-76 |
| `d-lb-gateway-b` | relationship | process | HTTP :8080 | lab-deployment | src-77 |
| `d-gateway-b-proxy` | relationship | process | HTTP over TCP proxies | lab-deployment | src-78 |
| `d-lb-gateway-c` | relationship | process | HTTP :8080 | lab-deployment | src-79 |
| `d-gateway-c-proxy` | relationship | process | HTTP over TCP proxies | lab-deployment | src-80 |
| `d-proxy-control` | relationship | process | TCP / HTTP | lab-deployment | src-81 |
| `d-proxy-node-a` | relationship | process | TCP / HTTP | lab-deployment | src-82 |
| `d-proxy-node-b` | relationship | process | TCP / HTTP | lab-deployment | src-83 |
| `d-proxy-blobstore` | relationship | process | TCP / HTTP | lab-deployment | src-84 |
| `d-proxy-metering` | relationship | process | TCP / HTTP | lab-deployment | src-85 |
| `d-node-a-partners` | relationship | process | HTTP over TCP proxies | lab-deployment | src-86 |
| `d-control-node-a` | relationship | process | HTTP :8070 | lab-deployment | src-19 |
| `d-node-a-hb` | relationship | process | HTTP :8090 | lab-deployment | src-20 |
| `d-node-a-disk` | relationship | volume mount | Local file I/O | lab-deployment | src-66 |
| `d-node-b-partners` | relationship | process | HTTP over TCP proxies | lab-deployment | src-86 |
| `d-control-node-b` | relationship | process | HTTP :8070 | lab-deployment | src-19 |
| `d-node-b-hb` | relationship | process | HTTP :8090 | lab-deployment | src-20 |
| `d-node-b-disk` | relationship | volume mount | Local file I/O | lab-deployment | src-67 |
| `d-control-pg` | relationship | process | PostgreSQL :5432 | lab-deployment | src-22 |
| `d-pg-volume` | relationship | volume mount | Local file I/O | lab-deployment | src-65 |
| `d-blob-volume` | relationship | volume mount | Local file I/O | lab-deployment | src-87 |
| `d-exporter-pg` | relationship | process | PostgreSQL :5432 | lab-deployment | src-88 |
| `d-scrape-gateway-a` | relationship | process | HTTP / Prometheus text | lab-deployment | src-89 |
| `d-scrape-gateway-b` | relationship | process | HTTP / Prometheus text | lab-deployment | src-89 |
| `d-scrape-gateway-c` | relationship | process | HTTP / Prometheus text | lab-deployment | src-89 |
| `d-scrape-control-run` | relationship | process | HTTP / Prometheus text | lab-deployment | src-90 |
| `d-scrape-node-a` | relationship | process | HTTP / Prometheus text | lab-deployment | src-91 |
| `d-scrape-node-b` | relationship | process | HTTP / Prometheus text | lab-deployment | src-91 |
| `d-scrape-haproxy-run` | relationship | process | HTTP / Prometheus text | lab-deployment | src-92 |
| `d-scrape-blob-run` | relationship | process | HTTP / Prometheus text | lab-deployment | src-93 |
| `d-scrape-meter-run` | relationship | process | HTTP / Prometheus text | lab-deployment | src-94 |
| `d-scrape-toxiproxy` | relationship | process | HTTP / Prometheus text | lab-deployment | src-95 |
| `d-scrape-exporter` | relationship | process | HTTP / Prometheus text | lab-deployment | src-96 |
| `d-migrate-pg` | relationship | process | PostgreSQL :5432 | lab-deployment | src-69 |
| `d-purger-pg` | relationship | process | PostgreSQL :5432 | lab-deployment | src-70 |
| `d-seedgen-pg` | relationship | process | PostgreSQL :5432 | lab-deployment | src-71 |
| `d-seed-node-a` | relationship | volume mount | Local file I/O | lab-deployment | src-97 |
| `d-report-node-a` | relationship | read-only volume mount | DuckDB SQLite scanner | lab-deployment | src-98 |
| `d-seed-node-b` | relationship | volume mount | Local file I/O | lab-deployment | src-97 |
| `d-report-node-b` | relationship | read-only volume mount | DuckDB SQLite scanner | lab-deployment | src-98 |
| `d-load-edge` | relationship | process | HTTP :8080 | lab-deployment | src-99 |
| `d-seedgen-results` | relationship | bind mount | Local file I/O | lab-deployment | src-71 |
| `d-loadgen-results` | relationship | bind mount | Local file I/O | lab-deployment | src-72 |
| `d-report-results` | relationship | bind mount | Local file I/O | lab-deployment | src-73 |
| `d-harness-tox` | relationship | process | HTTP admin :8474 | lab-deployment | src-100 |
| `d-harness-load` | relationship | host / Docker | Docker Compose CLI | lab-deployment | src-101 |
| `d-harness-blob` | relationship | process | HTTP admin | lab-deployment | src-102 |

## Source selectors

| ID | Pinned path | Symbol / selector | Line |
|---|---|---|---:|
| src-1 | [books/release-it/project/internal/api/handlers.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/api/handlers.go#L40) | `func (g *Gateway) Register` | 40 |
| src-2 | [books/release-it/project/Makefile](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/Makefile#L35) | `verify:` | 35 |
| src-3 | [books/release-it/project/cmd/node/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/main.go#L1) | `// Command node` | 1 |
| src-4 | [books/release-it/project/cmd/blobstore/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/blobstore/main.go#L45) | `func main()` | 45 |
| src-5 | [books/release-it/project/cmd/metering/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/metering/main.go#L49) | `func post` | 49 |
| src-6 | [books/release-it/project/internal/api/handlers.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/api/handlers.go#L226) | `func (g *Gateway) usage` | 226 |
| src-7 | [books/release-it/project/internal/blob/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/blob/client.go#L37) | `func (c *Client) Put` | 37 |
| src-8 | [books/release-it/project/internal/metering/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/metering/client.go#L33) | `func (c *Client) Send` | 33 |
| src-9 | [books/release-it/project/deploy/haproxy.cfg](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/haproxy.cfg#L18) | `backend gateways` | 18 |
| src-10 | [books/release-it/project/cmd/gateway/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/gateway/main.go#L24) | `func main()` | 24 |
| src-11 | [books/release-it/project/cmd/control/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/control/main.go#L27) | `func main()` | 27 |
| src-12 | [books/release-it/project/internal/registry/registry.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/registry/registry.go#L34) | `type Registry struct` | 34 |
| src-13 | [books/release-it/project/cmd/node/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/main.go#L42) | `func main()` | 42 |
| src-14 | [books/release-it/project/internal/engine/handles.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/handles.go#L100) | `func New(` | 100 |
| src-15 | [books/release-it/project/cmd/node/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/main.go#L49) | `self, err := sql.Open` | 49 |
| src-16 | [books/release-it/project/internal/api/handlers.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/api/handlers.go#L50) | `func (g *Gateway) key` | 50 |
| src-17 | [books/release-it/project/internal/routes/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/routes/client.go#L70) | `func (c *Client) Resolve` | 70 |
| src-18 | [books/release-it/project/internal/nodeclient/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/nodeclient/client.go#L110) | `func (c *Client) send` | 110 |
| src-19 | [books/release-it/project/cmd/control/nodes.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/control/nodes.go#L43) | `func (n *nodeCommands) send` | 43 |
| src-20 | [books/release-it/project/cmd/node/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/main.go#L224) | `func (n *node) heartbeats` | 224 |
| src-21 | [books/release-it/project/internal/backup/job.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/backup/job.go#L70) | `func (j *Job) Backup` | 70 |
| src-22 | [books/release-it/project/internal/registry/registry.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/registry/registry.go#L48) | `func Open(` | 48 |
| src-23 | [books/release-it/project/internal/engine/exec.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/exec.go#L130) | `func (e *Engine) Query` | 130 |
| src-24 | [books/release-it/project/internal/metering/outbox.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/metering/outbox.go#L25) | `func (o *Outbox) Add` | 25 |
| src-25 | [books/release-it/project/internal/api/handlers.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/api/handlers.go#L89) | `func (g *Gateway) query` | 89 |
| src-26 | [books/release-it/project/internal/breaker/breaker.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/breaker/breaker.go#L49) | `type Breaker struct` | 49 |
| src-27 | [books/release-it/project/internal/nodeclient/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/nodeclient/client.go#L48) | `func (c *Client) Query` | 48 |
| src-28 | [books/release-it/project/internal/retry/retry.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/retry/retry.go#L75) | `func (p Policy) Do` | 75 |
| src-29 | [books/release-it/project/cmd/gateway/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/gateway/main.go#L40) | `limiter := shed.New` | 40 |
| src-30 | [books/release-it/project/cmd/gateway/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/gateway/main.go#L60) | `limiter.Wrap(served)` | 60 |
| src-31 | [books/release-it/project/internal/api/handlers.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/api/handlers.go#L65) | `g.Routes.Resolve` | 65 |
| src-32 | [books/release-it/project/internal/routes/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/routes/client.go#L164) | `func (c *Client) do(` | 164 |
| src-33 | [books/release-it/project/internal/routes/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/routes/client.go#L170) | `func (c *Client) doWith(` | 170 |
| src-34 | [books/release-it/project/internal/routes/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/routes/client.go#L88) | `c.resolve.Do` | 88 |
| src-35 | [books/release-it/project/internal/api/handlers.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/api/handlers.go#L120) | `rows.WriteTo(w)` | 120 |
| src-36 | [books/release-it/project/internal/nodeclient/client.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/nodeclient/client.go#L56) | `p.Do(ctx` | 56 |
| src-37 | [books/release-it/project/internal/engine/admission.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/admission.go#L67) | `func (e *Engine) acquire` | 67 |
| src-38 | [books/release-it/project/internal/engine/handles.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/handles.go#L128) | `func (e *Engine) Handle` | 128 |
| src-39 | [books/release-it/project/internal/metering/queue.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/metering/queue.go#L40) | `type Queue struct` | 40 |
| src-40 | [books/release-it/project/internal/metering/queue.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/metering/queue.go#L155) | `func (q *Queue) sweep` | 155 |
| src-41 | [books/release-it/project/cmd/node/drain.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/drain.go#L22) | `func (n *node) drainOnSignal` | 22 |
| src-42 | [books/release-it/project/internal/engine/exec.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/exec.go#L137) | `h, err := e.Handle` | 137 |
| src-43 | [books/release-it/project/internal/engine/handles.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/handles.go#L170) | `func (e *Engine) open` | 170 |
| src-44 | [books/release-it/project/internal/backup/job.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/backup/job.go#L72) | `j.eng.Handle` | 72 |
| src-45 | [books/release-it/project/internal/backup/job.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/backup/job.go#L85) | `j.blob.Put` | 85 |
| src-46 | [books/release-it/project/internal/engine/exec.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/exec.go#L309) | `func (e *Engine) emit` | 309 |
| src-47 | [books/release-it/project/internal/metering/queue.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/metering/queue.go#L70) | `q.outbox.Add` | 70 |
| src-48 | [books/release-it/project/internal/metering/queue.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/metering/queue.go#L142) | `err := q.client.Send(batch)` | 142 |
| src-49 | [books/release-it/project/internal/engine/steady.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/internal/engine/steady.go#L34) | `func (e *Engine) Maintain` | 34 |
| src-50 | [books/release-it/project/cmd/node/drain.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/node/drain.go#L34) | `n.usage.Spill` | 34 |
| src-51 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L4) | `name: litefleet` | 4 |
| src-52 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L122) | `  gateway-a:` | 122 |
| src-53 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L127) | `  gateway-b:` | 127 |
| src-54 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L132) | `  gateway-c:` | 132 |
| src-55 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L98) | `  node-a:` | 98 |
| src-56 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L110) | `  node-b:` | 110 |
| src-57 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L63) | `  control:` | 63 |
| src-58 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L37) | `  postgres:` | 37 |
| src-59 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L137) | `  haproxy:` | 137 |
| src-60 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L150) | `  blobstore:` | 150 |
| src-61 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L158) | `  metering:` | 158 |
| src-62 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L143) | `  toxiproxy:` | 143 |
| src-63 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L165) | `  prometheus:` | 165 |
| src-64 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L57) | `  postgres-exporter:` | 57 |
| src-65 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L50) | `    volumes: [pgdata:` | 50 |
| src-66 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L108) | `    volumes: [node-a-data:/data]` | 108 |
| src-67 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L120) | `    volumes: [node-b-data:/data]` | 120 |
| src-68 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L156) | `    volumes: [blobs:/blobs]` | 156 |
| src-69 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L79) | `  migrate:` | 79 |
| src-70 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L89) | `  purger:` | 89 |
| src-71 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L173) | `  seedgen:` | 173 |
| src-72 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L182) | `  loadgen:` | 182 |
| src-73 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L191) | `  report:` | 191 |
| src-74 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L189) | `    volumes: [../results:/results]` | 189 |
| src-75 | [books/release-it/project/deploy/haproxy.cfg](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/haproxy.cfg#L41) | `    server gateway-a` | 41 |
| src-76 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L124) | `    environment: { INSTANCE: gateway-a` | 124 |
| src-77 | [books/release-it/project/deploy/haproxy.cfg](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/haproxy.cfg#L42) | `    server gateway-b` | 42 |
| src-78 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L129) | `    environment: { INSTANCE: gateway-b` | 129 |
| src-79 | [books/release-it/project/deploy/haproxy.cfg](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/haproxy.cfg#L43) | `    server gateway-c` | 43 |
| src-80 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L134) | `    environment: { INSTANCE: gateway-c` | 134 |
| src-81 | [books/release-it/project/deploy/toxiproxy.json](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/toxiproxy.json#L2) | `"name": "control"` | 2 |
| src-82 | [books/release-it/project/deploy/toxiproxy.json](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/toxiproxy.json#L3) | `"name": "node-a"` | 3 |
| src-83 | [books/release-it/project/deploy/toxiproxy.json](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/toxiproxy.json#L4) | `"name": "node-b"` | 4 |
| src-84 | [books/release-it/project/deploy/toxiproxy.json](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/toxiproxy.json#L5) | `"name": "blobstore"` | 5 |
| src-85 | [books/release-it/project/deploy/toxiproxy.json](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/toxiproxy.json#L6) | `"name": "metering"` | 6 |
| src-86 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L105) | `      BLOBSTORE_URL:` | 105 |
| src-87 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L156) | `    volumes: [blobs:` | 156 |
| src-88 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L60) | `      DATA_SOURCE_NAME:` | 60 |
| src-89 | [books/release-it/project/deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml#L8) | `  - job_name: gateway` | 8 |
| src-90 | [books/release-it/project/deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml#L14) | `  - job_name: control` | 14 |
| src-91 | [books/release-it/project/deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml#L17) | `  - job_name: node` | 17 |
| src-92 | [books/release-it/project/deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml#L20) | `  - job_name: haproxy` | 20 |
| src-93 | [books/release-it/project/deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml#L23) | `  - job_name: blobstore` | 23 |
| src-94 | [books/release-it/project/deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml#L26) | `  - job_name: metering` | 26 |
| src-95 | [books/release-it/project/deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml#L29) | `  - job_name: toxiproxy` | 29 |
| src-96 | [books/release-it/project/deploy/prometheus.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/prometheus.yml#L32) | `  - job_name: postgres` | 32 |
| src-97 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L179) | `    volumes: [node-a-data:/nodes/node-a` | 179 |
| src-98 | [books/release-it/project/deploy/compose.yml](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/deploy/compose.yml#L197) | `    volumes: [../results:/results, ../lab:/lab:ro` | 197 |
| src-99 | [books/release-it/project/cmd/loadgen/main.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/cmd/loadgen/main.go#L125) | `func main()` | 125 |
| src-100 | [books/release-it/project/lab/toxics.sh](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/lab/toxics.sh#L14) | `curl` | 14 |
| src-101 | [books/release-it/project/lab/lib.sh](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/lab/lib.sh#L18) | `compose()` | 18 |
| src-102 | [books/release-it/project/test/chaos/chaos_test.go](https://github.com/nickstrad/book-notes/blob/936786399d7bab3ebbfa845b9a07df02b41ed363/books/release-it/project/test/chaos/chaos_test.go#L224) | `/admin/mode` | 224 |
