# 10 · Steady state: checkpoint WALs, purge the ledger, evict idle handles

Tag: `release-it/lesson-10` · Fixes: F15 · Verify: `make verify LESSON=10` (before: `REF=release-it/lesson-09`)
Book: Ch. 5 "Steady State" (Data Purging, Log Files, Limit Caching)

## The idea

Chapter 5's **Steady State** says that a system should be able to run indefinitely without a
person fiddling with it. Every mechanism that accumulates something (data, log files, cache
entries, open resources) needs a matching mechanism that removes it, running automatically, so the
amount levels off instead of growing until a disk, a heap or a file-descriptor table runs out.
Anything that only grows is a slow-motion outage with a date already set; you just have not
computed it yet. And humans fiddling with production to clean up is itself a source of outages,
which is why the cleanup must be automatic.

A database platform accumulates in three places:

- **WAL files.** In WAL mode, SQLite appends a full page image for every page a write touches, so
  a WAL grows tens of times faster than the data does. A checkpoint copies the pages back into the
  database file, and a truncating checkpoint shrinks the WAL. Without checkpoints, the WAL is a
  log file that is never rotated.
- **The usage ledger.** One row per database per kind per heartbeat, forever.
- **Open handles.** Every database ever touched keeps a connection pool, three file descriptors
  per connection and a page cache.

## The flaw (F15, at `release-it/lesson-09`)

- `internal/engine/handles.go:103`: every tenant file opens with `wal_autocheckpoint(0)`, which
  turns SQLite's own checkpoints off. The comment says they belong in a maintenance loop ("TODO:
  schedule them"), and the loop was never written.
- `internal/engine/handles.go:132`: the handle map only grows; nothing ever closes a database.
  Lesson 3's budget (150 handles) turns the growth into `503 node_full` instead of an fd error, but
  it is still a wall.
- `internal/registry/usage.go:15`: the ledger gains rows on every heartbeat and nothing ever removes
  them.

## The patch

- **The checkpointer** (`internal/engine/steady.go`). Tenant files open with SQLite's default
  autocheckpoint again (1 000 pages, on a writer's commit). Every 2 s, `checkpoint` (`:46`) runs
  `PRAGMA wal_checkpoint(TRUNCATE)` on any database whose WAL has passed `CheckpointBytes` = 64 KB,
  or whose database has been idle for 30 s. This runs in the background, off the tenants' write
  path, which was the original reason for turning autocheckpoints off. Metric:
  `sqlite_checkpoints_total{reason,outcome}`.
- **The handle cache** becomes an LRU of `HandleCache` = 128 (`hashicorp/golang-lru/v2`), and
  `closeIdle` (`:70`) closes any handle idle for 5 minutes. Callers take and `Release` a handle
  (`handles.go:161`), so an evicted handle that is in use closes with its last user, never under
  it. A database evicted before its next backup is remembered, and the backup job picks it up
  (`TakeEvictedDirty`, `:188`). This is the ts-db design's leased handle cache at baby size.
- **The ledger purge** (`cmd/control/purge.go`, `control purge`). Each pass moves raw rows older
  than the retention window (25 h, because the usage report reads 24 h) into
  `usage_daily(tenant_id, day, statements, rows, bytes)`, in batches of 5 000 with a 250 ms pause.
  Each batch is one statement: `DELETE ... RETURNING` feeding an upsert into the rollup
  (`PurgeBatch`, `internal/registry/usage.go:72`), so a row is never rolled up twice or lost between
  the two. The table comes from `migrations/002_usage_daily.sql`. In the lab, a Compose sidecar
  (`purger`, profile `maintenance`) runs a pass every minute; production would run it daily.

## Verify

```
make verify LESSON=10 REF=release-it/lesson-09      # before
make verify LESSON=10                               # after
```

The node data volumes are 48 MB tmpfs (`deploy/compose.tmpfs.yml`, a Compose override for this
run), and the whale is left out of the seed. The load is the steady mix at 60 % with half the
requests writes, cold reads at 5/s, and a new database every second (created, then read once, as a
new app would), for 90 s. The script samples at 1, 30, 60 and 90 s and extrapolates.

Before (`lesson-09`):

```
  t s  WAL node-a MB  WAL node-b MB  handles node-a  handles node-b  ledger rows  node-a volume used MB (of 48)
  1    0.5403         0.5206         97              97              5000840      23.5
  30   10.87          10.77          130             135             5002848      34.7
  60   22.35          21.74          145             150             5005259      46.5
  90   25.18          30.89          148             150             5007299      48.0
    node-a volume     +13.3 MB/min    full (48 MB) in 0 minutes
    handles node-a     +6.0 /min      handle budget (150) in 0 minutes; nofile (1024, 3 fds each) in 32 minutes
    ledger            +4080 rows/min
  writes failed: 2810 (node log lines saying the disk is full or failing: 2898); requests refused node_full: 4
FLAW REPRODUCED: wal_node_a_mb 25.18 (want <= 8); handles_growth_last_30s 3 (want <= 0); ledger_rows_growth_last_30s 2040 (want <= 0); disk_full_errors 2898 (want == 0)
```

After:

```
  1    0.5876         0.7056         96              95              5000611      23.2
  30   3.13           3.162          127             128             5002976      25.4
  60   3.079          3.457          128             128             4065387      26.1
  90   3.115          3.645          128             128             3157800      26.2
    WAL node-a         -0.0 MB/min
    handles node-a     +0.0 /min      handle budget (150) in never; nofile (1024, 3 fds each) in never
  writes failed: 0 (node log lines saying the disk is full or failing: 0); requests refused node_full: 0
IMPROVED: wal_node_a_mb 3.127 <= 8; handles_growth_last_30s 0 <= 0; ledger_rows_growth_last_30s -417586 <= 0; disk_full_errors 0 == 0
```

(The after table is from one run and the verdict line from the next, after the purge's pause was
lengthened. The numbers are close.)

Before the patch, the WAL grew 14 MB a minute per node while the data itself grew about half a
megabyte. The 48 MB volume filled after about 80 s, and from then on every write on node-a failed.
SQLite reports the full tmpfs as `disk I/O error`, an ENOSPC on the WAL append. Reads kept
working, which is the misleading part: the node still looked half alive. Handles reached the
150-handle budget, so new databases were refused. The ledger grew by 4 000 rows a minute; at a
platform's real scale that is the table whose `ALTER` lesson 12 is afraid of.

After the patch, the WAL holds at about 3 MB per node and the handles at the cache size. The ledger
shrinks while the first purge pass works through 30 days of history, and then holds at one
retention window. What still grows, 0.5–0.8 MB a minute on node-a, is the tenants' actual data.
That is supposed to grow: it is a question for capacity planning, not a leak.

## Try changing

- Set `CheckpointBytes` to 4 MB (SQLite's default autocheckpoint size) and rerun. With ~120
  databases each taking a few writes a second, none goes idle, and the volume fills almost as fast
  as before.
- Set `HandleCache` to 64, as in the plan. The cache is smaller than the steady working set (~123
  databases per node), so it thrashes: watch `sqlite_open_databases` stay at 64 while open and
  close counts climb.
- Run `control purge -every 0 -pause 0` against a full ledger during a load test, and watch the
  critical path's p99 and lesson 5's breaker.

## Read more

*Release It!* 2nd ed., Ch. 5 "Steady State". SQLite's documentation on WAL mode and
`wal_checkpoint`. PostgreSQL's documentation on `DELETE ... RETURNING` in a CTE.

## Deviations from the plan

- **The checkpointer also truncates any WAL past 64 KB**, not only those idle for 30 s. The lab's
  steady traffic touches every database every second or two, so none is ever idle, and SQLite's
  default autocheckpoint lets each WAL reach ~4 MB before it reacts: ~480 MB per node.
- **The handle cache is 128 entries, not 64.** The steady mix's working set is about 123 databases
  per node, so a 64-entry cache would evict databases that are about to be used again. 128 sits
  under lesson 3's 150-handle budget.
- **The volumes are 48 MB, not 256 MB.** At this VM's write rates, 256 MB takes several minutes to
  fill, beyond the verify's budget. 48 MB holds the seeded data with room for a bounded WAL.
- **A full volume shows as `disk I/O error`, not `SQLITE_FULL`**, and the check counts both. Ledger
  rows are counted with `count(*)`, because `pg_stat_user_tables` restarts from zero when the
  registry is restored from the seed template.
- **The purge pauses 250 ms between batches, not "a sleep" of 50 ms.** With 50 ms, the first pass
  over 30 days of history competed with resolves on a Postgres capped at 1.5 CPUs, and the gateways'
  breaker opened for a few seconds.
- **Lab:** loadgen's creates now read the new database once, so a new database adds a handle as it
  would for a real app.
