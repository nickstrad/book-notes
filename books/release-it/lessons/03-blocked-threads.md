# 03 · Blocked threads: roll back on every path, bound the pools, fail fast

Tag: `release-it/lesson-03` · Fixes: F05, F06, F07 · Verify: `make verify LESSON=03` (before: `REF=release-it/lesson-02`)
Book: Ch. 2 case study ("The Exception That Grounded an Airline"); Ch. 4 "Blocked Threads",
"Unbalanced Capacities"; Ch. 5 "Fail Fast", "Timeouts"

## The idea

Chapter 2 of *Release It!* tells how an airline's check-in system went down for hours. An exception
thrown while cleaning up after a database call skipped the code that gave the connection back to
its pool. Each time it happened, one more connection was lost. When the pool ran dry, every request
thread blocked waiting for a connection that would never come back. The servers were up, the
database was fine, and nothing could check anyone in. The bug was one missing piece of cleanup on
an error path. It passed every test, because tests exercise the paths that work.

Chapter 4 generalises this as **Blocked Threads**: most outages are not crashes but threads stuck
waiting, usually on a resource someone failed to release or a call with no limit. The defences are
to release resources on every path (the language's cleanup idiom, not hand-written error branches),
to bound every pool, and to put a timeout on waiting for a pool as well as on the network.
Chapter 5's **Fail Fast** adds a cheap rule: if a request is going to fail, fail it before it has
used anything expensive. Validate input and check resources first, then do the work.

## The flaw (at `release-it/lesson-02`)

- **F05.** In `Execute` (`internal/engine/exec.go:150–172`), the transaction begins with
  `h.db.Begin()`, and a failing statement returns `classify(ctx, err)` at `:160` without
  `tx.Rollback()`. The connection never goes back to the pool. Worse, in SQLite the open write
  transaction keeps **the file's write lock**, so that database never accepts another write: every
  later write gets `SQLITE_BUSY`, spins in the busy loop, fails with `database is locked`, and takes
  the same early return. Each failure leaks one more connection.
- **F06.** Each database's `*sql.DB` (`internal/engine/handles.go:85`) has no `SetMaxOpenConns`.
  When a connection leaks, the pool opens another, and every WAL connection costs 3 file descriptors.
  The node has no budget of open handles either. Control's Postgres pool
  (`internal/registry/registry.go:45–47`) bounds idle connections but not open ones.
- **F07.** The gateway resolves the route through control (`internal/api/handlers.go:78`, `:102`)
  before it looks at the body (`:82`, `:106–111`). The node opens the tenant's file (`exec.go:85`)
  before checking that `/query` carries a read statement (`:92`).

The gateway's retries (F08) make F05 faster: a 500 `database is locked` is retried five times, and
each attempt leaks another connection.

## The patch

- **Roll back on every path.** `Execute` checks out a connection, begins the transaction, and
  immediately does `defer tx.Rollback()` (`exec.go:188–194`). Every return, error or success,
  releases the write lock and returns the connection. A rollback after a commit is a harmless
  no-op. `Query` already used `defer tx.Rollback()` and `defer rows.Close()`. Both now begin their
  transaction under the statement context. A constraint violation is still the typed
  `409 constraint`, now returned through the deferred cleanup.
- **Bound the pools.** Each database: at most 4 connections, 2 kept idle, idle ones closed after a
  minute (`handles.go:29`, `:115`). Each node: a budget of open handles,
  `MaxOpenDatabases(nofile)` = (1024 − 124 reserved) / (2 idle connections × 3 fds) = 150
  (`handles.go:38`). Past it, a new database gets `503 node_full` instead of an fd error deep in
  SQLite. Control: at most 20 Postgres connections, 20 idle, recycled every 30 minutes
  (`registry.go:49`).
- **Bound the waits for them.** A statement waits at most `PoolWait` = 250 ms for a connection
  (`exec.go:38`, `conn` at `:49`), then gets `503 pool_timeout`, counted in
  `sqlite_pool_timeouts_total`. Control gives each route a budget inside its caller's (`budget`,
  `cmd/control/main.go:63`; resolve 450 ms inside the gateway's 500 ms). A registry call that runs
  out of time waiting for the pool answers `503 pool_timeout` (`:232`).
- **Fail fast.** The gateway decodes and validates the body before it resolves
  (`handlers.go:78`). The node checks that `/query` is a single read statement, and that `/execute`
  has statements, before it opens the file (`exec.go:113`, `:172`).

`TestFailedWriteReleasesTheWriteLock` (`internal/engine/engine_test.go`) is the regression test:
ten constraint violations, then a write that must succeed. At `lesson-02` it fails with
`database is locked`.

## Verify

```
make verify LESSON=03 REF=release-it/lesson-02      # before
make verify LESSON=03                               # after
```

First, fail fast: 50 `execute` requests with an empty statement list, then 40 writes sent to
`/query` on databases nobody has opened (each tenant's `analytics` and `staging`). Then the `errs`
mix at 50 % of fleet capacity (300 req/s) for 60 s. A quarter of the requests are writes, and 30 %
of those insert an `items.seq` that already exists.

Before (`lesson-02`):

```
[t=15 s] fds node-a 1024 node-b 1023
  control resolves spent on 50 empty executes       50
  tenant files opened by 40 writes sent to /query   40
  max process_open_fds on a node                    1024 (limit 1024)
  5xx answered by the nodes                         14750
  valid writes that succeeded                       21.89 %
  "database is locked" in node logs                 945
  8 s after the load: node-a /health 000, node-b /health 200, max fds 1024
FLAW REPRODUCED: resolves_for_invalid 50 (want == 0); files_opened_for_invalid 40 (want == 0); max_node_fds 1024 (want <= 600); node_5xx 1.475e+04 (want == 0); valid_write_pct 21.89 (want >= 95); constraint_p50_ms 53.8 (want <= 20); locked_errors 945 (want <= 5); node_a_healthy_after 000 (want == 200)
```

After:

```
  control resolves spent on 50 empty executes       0
  tenant files opened by 40 writes sent to /query   0
  max process_open_fds on a node                    471 (limit 1024)
  5xx answered by the nodes                         0
  valid writes that succeeded                       99.93 %
  409 constraint p50                                9.4 ms
  "database is locked" in node logs                 2
  8 s after the load: node-a /health 200, node-b /health 200, max fds 510
IMPROVED: resolves_for_invalid 0 == 0; files_opened_for_invalid 0 == 0; max_node_fds 471 <= 600; node_5xx 0 == 0; valid_write_pct 99.93 >= 95; constraint_p50_ms 9.4 <= 20; locked_errors 2 <= 5; node_a_healthy_after 200 == 200
```

Before the patch, both nodes reached their 1 024-fd wall within 15 s. Every database that saw one
constraint violation stopped accepting writes. When a node could not accept connections, every
tenant on it was down, and it stayed down after the load stopped: the leaked connections are
never returned. This is the airline's outage on a tenant database, and on a stateful tier that
cannot be routed around, "the fleet is down" means half the customers. After the patch, a
constraint violation is a 9 ms 409 and nothing else happens.

## Try changing

- Remove `defer tx.Rollback()` and run `go test ./internal/engine -run WriteLock`.
- Set `PoolWait` to 2 s and run the `noisy` scenario from lesson 6 against this tag. Requests queue
  behind the noisy tenant for the full statement deadline instead of failing fast.
- Lower `nofile` in `deploy/compose.yml` to 256 and watch `503 node_full` replace fd errors once
  `MaxOpenDatabases` drops below the number of databases on a node. Lesson 10's LRU is the fix for
  that.

## Read more

*Release It!* 2nd ed., Ch. 2 (the airline case study), Ch. 4 "Blocked Threads", Ch. 5 "Fail Fast".
The `database/sql` documentation on `DB.SetMaxOpenConns`, `Conn` and `Tx`. SQLite's documentation
on locking in WAL mode.

## Deviations from the plan

- **Control's pool is 20 connections, not 10.** With every request resolving through control
  (until lesson 5) at 300 req/s, ten connections ran at 100 % in use with five waiters on average
  on this shared VM. The mean registry call was 42 ms, almost all of it waiting for the pool. Twenty
  covers the observed tail and is still well inside Postgres's 100.
- **The verdict reads the nodes' own answers (node 5xx = 0), not every tenant-visible error.**
  Until lessons 5 and 6, a 100 ms stall in Postgres still reaches tenants through every request's
  resolve. Control's health check pings through the shared pool, gateways report control
  unreachable, and HAProxy takes them all out for a few seconds. One run on this shared VM showed
  exactly that at t=50 s. Tenant-visible errors are printed, and valid writes must still succeed
  ≥ 95 % end to end.
- **"database is locked" ≤ 5, not 0.** Writes that collide with a backup's lock still fail through
  the tight `SQLITE_BUSY` loop (F08, lesson 4). The leaked-lock version of the error ran to 945
  before the patch.
- The fail-fast probes run before the load, so a node already at its fd wall cannot make them look
  passed.
