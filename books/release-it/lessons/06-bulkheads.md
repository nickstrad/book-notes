# 06 · Bulkheads: per-tenant limits on the node, separate pools for the report and health, backups off the writer lock

Tag: `release-it/lesson-06` · Fixes: F10, F11 · Verify: `make verify LESSON=06` (before: `REF=release-it/lesson-05`)
Book: Ch. 5 "Bulkheads"; Ch. 4 "Chain Reactions", "Cascading Failures", "Blocked Threads"

## The idea

A ship's hull is divided into watertight compartments, so a breach floods one compartment and the
ship stays afloat. Chapter 5's **Bulkheads** applies the same idea to resources: partition them so
that one consumer, one path, or one failure can exhaust only its own share. The unit can be whole
servers dedicated to a customer class, separate connection pools for separate kinds of work, or
separate thread pools per dependency. Without partitions, the resource is shared, and whoever uses
it hardest decides how everyone else fares.

For a multitenant platform, bulkheads are the product's promise: your database is not slowed down
by someone else's. Three places in litefleet share a resource across work that should be kept
apart:

- **Tenants on a node.** One tenant's expensive queries compete with everyone else's for the
  node's processor and statement slots.
- **Paths in control.** The ops dashboard's heavy usage report, the critical path (resolve, create,
  heartbeat), and the health check all draw from one Postgres pool.
- **A backup and the writes it copies.** The backup holds the database's write lock while it waits
  on the blob store.

## The flaw (at `release-it/lesson-05`)

- **F10, node.** `internal/engine/admission.go:13` only checks the epoch. A tenant can run as many
  statements as its database's connection pool allows (4, from lesson 3) on each of its databases,
  and nothing limits a tenant or the node as a whole.
- **F10, control.** `GET /v1/admin/usage` (`cmd/control/usage.go:14`) runs with no limit on how many
  at once, on the same pool as everything else (`internal/registry/registry.go:49`,
  `usage.go:55–76`). Control's `/health` pings through that pool (`registry.go:62`,
  `cmd/control/main.go:72`), so a busy pool looks like a dead database. The node's health canary
  opens through the shared handle path (`internal/engine/metrics.go:56`).
- **F11, backup.** `internal/backup/job.go:75–98` takes the write lock (`BEGIN IMMEDIATE`), writes a
  snapshot, **uploads it** (`:92`), records the generation, and only then commits. A slow blob store
  holds the lock for the whole upload, and every write to that database waits out its 2 s busy
  timeout and fails.

## The patch

Three independent parts.

1. **Admission on the node** (`internal/engine/admission.go`). `acquire` (`:67`) takes three
   bulkheads in order. First, the database's lane: 2 concurrent reads or 1 write, with at most 8
   waiting; beyond that, `429 tenant_busy` with `Retry-After: 1` immediately. Next, the tenant's
   share: at most 4 statements across all of its databases, with a 250 ms wait, else 429
   (`tenantSlots`, `:116`). Last, the node: at most 16 statements at once, else `503 node_busy`. So
   one tenant can hold at most a quarter of the node. This is the lane idea from the ts-db planes
   design (a `Semaphore(1)` writer per file and a bounded writer queue) at baby size. Metrics:
   `admission_rejected_total{reason}`, `statements_running`.
2. **Separate pools in control** (`internal/registry/registry.go:48–67`). One DSN, three pools:
   `oltp` (8 connections, 2 s statements) for resolve, create and heartbeat; `reporting`
   (2 connections, `statement_timeout=5s`) for the usage report; `health` (1 connection, 200 ms)
   for `/health`. The report route admits two at a time and answers `503 report_busy` with
   `Retry-After` to the rest (`cmd/control/usage.go:14–24`). `db_pool_*` metrics gain a `pool`
   label. The node's canary gets a dedicated one-connection handle outside the tenant handle cache
   (`internal/engine/metrics.go:59`).
3. **Backups off the writer lock** (`internal/backup/job.go:70–107`). The lock now covers only
   `VACUUM INTO`, a consistent copy that takes milliseconds. It is released before the upload, and
   the generation is recorded after the upload succeeds. `writer_lock_hold_seconds` is measured
   separately from `backup_duration_seconds`.

Two small supporting changes. The gateway passes `Retry-After` through on upstream 429 and 503, so
refused clients back off instead of spinning. Creates and drops get a 2.5 s client
(`internal/routes/client.go:47`): they provision a file on a node, control's own budget for them is
2 s, and lesson 2's 500 ms was a resolve-sized budget.

## Verify

```
make verify LESSON=06 REF=release-it/lesson-05      # before
make verify LESSON=06                               # after
```

Steady load at 50 % in each phase:

- **(a)** The noisy tenant, spread over its 4 databases on node-a, runs 20 concurrent ~1.5 s
  `WITH RECURSIVE` queries in a loop from t=10 s to t=45 s. Its client honours `Retry-After`.
- **(b)** The ops dashboard refreshes the usage report 20 at a time for 30 s, while cold reads
  (5/s) and creates (2/s) exercise the paths that still need control. The script polls control's
  `/health` every 0.25 s.
- **(c)** The blob store's proxy drops to 50 KB/s (`bandwidth` toxic, upstream). The script calls
  the node's backup endpoint (`POST /internal/backup/{id}`, the same call the scheduler makes) for
  the 5 MB `bulky` database while its tenant writes to it at 10/s.

Before (`lesson-05`):

```
a.  other node-a tenants p99: before the noisy tenant 39.0 ms, during 534.0 ms (13.69x)
b.  control answered 503 pool_timeout to resolves or creates 264
    tenant-visible 5xx or timeouts (steady, cold, creates)    156 of 8590
    control /health not 200                                13 of 55 checks
c.  writes to the database being backed up: 250, failed 250, p99 2015.0 ms
FLAW REPRODUCED: noisy_other_p99_ms 534.0 (want <= 200); reports_control_pool_timeouts 264 (want == 0); reports_control_health_bad 13 (want == 0); backup_target_failures 250 (want == 0); backup_target_p99_ms 2015.0 (want <= 250); writer_lock_hold_s NaN (want <= 1)
```

After:

```
a.  other node-a tenants p99: before the noisy tenant 21.0 ms, during 87.0 ms (4.14x)
    other node-a tenants 5xx or timeouts during      0
    noisy tenant answered 429 tenant_busy            280
b.  control answered 503 pool_timeout to resolves or creates 0
    tenant-visible 5xx or timeouts (steady, cold, creates)    0 of 8596
    control /health not 200                                0 of 96 checks
    reports served 21 (p99 3977.0 ms), refused 503 report_busy 540
c.  writes to the database being backed up: 250, failed 0, p99 158.0 ms
    longest write-lock hold by a backup                    0.512 s
IMPROVED: noisy_other_p99_ms 87.0 <= 200; noisy_other_5xx 0 == 0; reports_control_pool_timeouts 0 == 0; reports_control_health_bad 0 == 0; backup_target_failures 0 == 0; backup_target_p99_ms 158.0 <= 250; writer_lock_hold_s 0.512 <= 1
```

- **(a)** Before, the noisy tenant ran 16 expensive statements at once across its databases, and
  everyone else on node-a waited behind them. After, it holds four and is told to come back later.
  Other tenants still see their p99 rise about fourfold, because four CPU-bound statements still
  take their share of a half-core node. Bulkheads bound concurrency, not CPU, as the plan warned.
  What they removed is the unbounded part.
- **(b)** Before, twenty reports took the whole pool. Creates and cold reads failed, and control's
  health check failed with them, because a busy pool looked like a dead database. Thanks to lesson
  5's cache, the gateways stayed up and known routes kept working. After, two reports run, the rest
  are refused in milliseconds, and the critical path and health check do not notice. The served
  reports take 4 s, because the report itself is still unbounded: lesson 8.
- **(c)** Before, every write to the database being backed up failed for the length of the upload.
  After, the lock is held for half a second and the upload proceeds in the background.

## Try changing

- Set `statementsTenant` to 16 and rerun phase (a): the per-database lanes alone do not stop a
  tenant with four databases.
- Give `reporting` 8 connections and make the report semaphore 20 (one shared pool in effect)
  and rerun phase (b).
- Move the upload back inside `snapshot`, keep the new metrics, and compare
  `writer_lock_hold_seconds` with `backup_duration_seconds`.
- The instance-level bulkhead, dedicated nodes for a "pro" plan, is described but not built. It
  would be a placement rule in `internal/placement` keyed on the tenant's plan.

## Read more

*Release It!* 2nd ed., Ch. 5 "Bulkheads"; Ch. 4 "Chain Reactions" and "Cascading Failures". The
ts-db planes design (lanes and the writer queue; the M3 `lane-snapshot` change that moved snapshots
off the writer lock).

## Deviations from the plan

- **2 concurrent reads per database, not 4, plus a per-tenant cap of 4 statements.** On a half-core
  node, Go runs with GOMAXPROCS 2 under a CFS quota, so any CPU-bound statement drains the quota.
  Four per database kept other tenants' p99 at 4–5× baseline. With a single database, lesson 3's
  4-connection pool already confined a noisy tenant almost as well, and the flaw barely showed. The
  noisy tenant now has four databases (the seed has 243 databases, not 240), and the per-tenant cap
  is what makes "at most a quarter of the node" true.
- **(a)'s threshold is an absolute 200 ms p99 for other tenants**, with the ratio to the pre-noise
  baseline printed. The plan's "≤ 2× baseline" cannot hold for CPU-bound statements on this node
  size (the plan says as much), and the 10 s baseline itself varied 21–85 ms between runs on the
  shared VM.
- **(b)'s verdict reads control's own answers** (no 503 `pool_timeout` on resolve or create, health
  never leaving 200) rather than "zero critical-path 5xx at the tenant". Tenant-visible errors are
  printed; on the shared VM a node stall occasionally failed a handful of requests in either tree.
- **(c)'s write p99 limit is 250 ms, not 100 ms.** Before, every write fails at the 2 s busy
  timeout, so the bar separates the two cleanly either way.
- **The oltp pool is 8 connections, as planned.** Lesson 3 raised the single pool to 20, because
  every request resolved through control; with lesson 5's cache the critical path needs far fewer.
- **Lab:** Postgres is capped at 1.5 CPUs (it keeps its higher weight), so report scans cannot
  starve the nodes on the shared VM. The loadgen's dashboard client waits 10 s for a report, and
  its extra loops honour `Retry-After`.
