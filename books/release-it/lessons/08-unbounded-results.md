# 08 · Unbounded result sets: cap and stream rows, aggregate usage in the database

Tag: `release-it/lesson-08` · Fixes: F13 · Verify: `make verify LESSON=08` (before: `REF=release-it/lesson-07`)
Book: Ch. 4 "Unbounded Result Sets", "Slow Responses"; Ch. 5 "Let It Crash"

## The idea

Chapter 4's **Unbounded Result Sets** is about the query that worked in development, where the
table had a hundred rows, and then meets production, where it has millions. Code that reads "all
the rows" assumes something about the data that nobody enforces: that it is small. When the
assumption breaks, the result is not an error but memory exhaustion, long pauses, timeouts, or a
crash, and it tends to happen to the instance that serves everyone else too. The remedies are to
bound every result (limit the rows, page through the rest), to stream what you return instead of
building it whole, and to push aggregation to where the data lives instead of pulling the data to
where the code is.

A multitenant platform cannot trust its tenants' queries to be bounded, so the bound has to be the
platform's. The "whale", the one account that dwarfs the rest, is exactly where an unbounded read
turns into an outage for everyone on the same node.

Chapter 5's **Let It Crash** explains why a crash is not the end of the world when it is fast and
clean: an instance that dies and restarts in a known good state is better than one that limps. That
is only survivable here because of earlier lessons. Health checks report the node honestly (1), the
gateways shed instead of queueing behind it (7), and SQLite recovers its WAL on open. On a stateful
tier, though, "let it crash" still means every tenant on that node is down while it restarts.

## The flaw (F13, at `release-it/lesson-07`)

- The node builds the whole result in memory: every row appended to a `[][]any`
  (`internal/engine/exec.go:165`), then the whole thing marshalled into one JSON body (`:170`).
- The gateway reads the node's whole body (`internal/nodeclient/client.go:97`), decodes it into
  `[][]any` again, and re-encodes it for the tenant (`internal/api/handlers.go:117`).
- The usage report pulls every ledger row of the window into control (`cmd/control/usage.go:42`,
  `internal/registry/usage.go:53`), about 170 000 rows for 24 hours, and adds them up in Go
  (`usage.go:53`).

## The patch

Two commits.

**Part 1: cap and stream.**

- The node streams rows into the response as it scans them and stops at `MaxRows` = 1 000 rows or
  `MaxBytes` = 1 MB of JSON, whichever comes first, adding `"truncated": true`
  (`internal/engine/exec.go:121`, `:130–200`). One reused `[]any` holds the current row; nothing
  holds the result. If the statement fails after the answer has begun, the handler cuts the
  connection (`http.ErrAbortHandler`) instead of sending a short answer that looks complete.
- The gateway streams too (`internal/nodeclient/rows.go`). `openQuery` (`:24`) sends the query and
  reads the node's answer only up to its first row. The fields before `"rows"` must be ones the
  gateway knows (`preamble`, `:38–55`), so the wire format is still checked as strictly as before.
  That matters for lesson 12. Retries happen only before an answer has begun
  (`client.go:47`). `WriteTo` (`:86`) then forwards the rows one at a time as raw JSON.
  `handlers.go:120` streams them to the tenant, or cuts the connection if the node's answer breaks
  off.
- `README.md` documents the cap and keyset paging (`WHERE id > ? ORDER BY id LIMIT n`). The cap is
  the platform's bulkhead for its nodes' memory, not a suggestion.

**Part 2: the usage report in SQL.** `UsageReport` (`internal/registry/usage.go:44`) is one
statement: `SUM` per tenant over the window, joined to current storage,
`ORDER BY bytes DESC LIMIT 50`. It runs on lesson 6's reporting pool with its 5 s
`statement_timeout`. Fifty rows cross the wire however large the ledger grows. A covering index,
`migrations/online/002_ledger_index.sql`
(`CREATE INDEX CONCURRENTLY ... ON usage_ledger (at) INCLUDE (tenant_id, statements, rows, bytes)`),
turns the window into an index-only scan. It is applied by `make migrate`, online and idempotent,
and never at startup: an index build does not belong in a deploy's critical path, and
`CONCURRENTLY` cannot run in the startup migrator's transaction. That is a first look at lesson 12.

## Verify

```
make verify LESSON=08 REF=release-it/lesson-07      # before
make verify LESSON=08                               # after
```

The script applies the tree's online migrations (none exist before this lesson), then runs the
steady mix at 30 % of fleet capacity for 45 s while the whale tenant runs `SELECT * FROM events`
ten at a time. That is 15 000 rows of 4 KB, a 60 MB table. Nodes have 512 MB and gateways 256 MB.
It counts container restarts and peak RSS, and afterwards times ten refreshes of the usage report
over the 5 M-row ledger.

Before (`lesson-07`):

```
  container restarts during the run (nodes, gateways)   3
  peak RSS: node 497.8 MB (limit 512), gateway 17.46 MB (limit 256)
  whale scans answered 0 of 918, p99 NULL ms
  other tenants: 5xx or timeouts 858; node-a tenants served 83.36 %
  slowest of ten usage reports                           1.150878 s
FLAW REPRODUCED: restarts 5 (want == 0); whale_answered_pct 0.0 (want >= 99); whale_p99_ms NaN (want <= 1000); other_tenant_5xx 1244 (want == 0); report_max_s 1.164289 (want <= 0.5)
```

(The table is from one run, and the verdict line is from a second run with the final thresholds.)

After:

```
  container restarts during the run (nodes, gateways)   0
  peak RSS: node 71.78 MB (limit 512), gateway 16.77 MB (limit 256)
  whale scans answered 1004 of 1004
  other tenants: 5xx or timeouts 0; node-a tenants served 100.0 %
  slowest of ten usage reports                           0.29 s
IMPROVED: restarts 0 == 0; whale_answered_pct 100.0 >= 99; whale_p99_ms 760.0 <= 1000; other_tenant_5xx 0 == 0; report_max_s 0.293905 <= 0.5
```

Before the patch, node-a was OOM-killed every few seconds, three to five times in 45 s. It never
answered a single whale scan, because each one died with the node. Every tenant on node-a lost
about a sixth of its requests, the ones in flight at each crash. Node-b's tenants, on a node nobody
scanned, did not notice. The gateways survived only because the node never got far enough to send
them 60 MB. After the patch, a whale scan is a capped answer of about 250 rows (the 1 MB cap
arrives before the 1 000-row cap) at ~150 ms of work. Node memory stays at 72 MB, and the usage
report is an index-only aggregate in a quarter of a second.

The whale's p99 of 760 ms is mostly waiting, not working. Ten concurrent scans share the database's
two read lanes from lesson 6, so most scans queue behind others.

## Try changing

- Raise `MaxBytes` to 64 MB and rerun: the node survives longer, but each scan holds a lane and a
  connection for seconds.
- Drop the index (`DROP INDEX usage_ledger_at_covering`) and time the report again: same SQL,
  but now a pass over 5 M rows.
- Page the whale's table with a keyset loop through the API (`WHERE id > ? ORDER BY id LIMIT 200`)
  and compare total time and node memory with one capped `SELECT *`.

## Read more

*Release It!* 2nd ed., Ch. 4 "Unbounded Result Sets"; Ch. 5 "Let It Crash". PostgreSQL
documentation on `CREATE INDEX CONCURRENTLY` and index-only scans (`INCLUDE`). Markus Winand,
*Use The Index, Luke*, on keyset pagination.

## Deviations from the plan

- **The gateway streams by transcoding, not with `io.Copy`.** It reads the node's fields before the
  rows strictly and forwards the rows one at a time. A plain byte copy would have removed the
  gateway's strict reader for queries here, four lessons before lesson 12 replaces it on purpose.
  The gateway still never holds more than one row.
- **Whale threshold: ≥ 99 % of scans answered and p99 ≤ 1 s** (plan: p99 ≤ 100 ms). Lesson 6's two
  read lanes per database queue the whale's ten concurrent scans; each takes ~150 ms of work.
  Before, no scan was answered at all.
- **Two commits** (about 430 lines with tests and the verify; plan ~150): streaming on both sides
  took a small JSON transcoder at the gateway (`rows.go`, ~130 lines).
- `make migrate` here applies `migrations/online/*.sql` with `psql`; lesson 12 turns migrations
  into a `control migrate` job.
