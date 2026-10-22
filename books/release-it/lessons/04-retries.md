# 04 · Retries: reads only, jittered backoff, a budget, `busy_timeout`

Tag: `release-it/lesson-04` · Fixes: F08 · Verify: `make verify LESSON=04` (before: `REF=release-it/lesson-03`)
Book: Ch. 4 "Dogpile", "Self-Denial Attacks", "Force Multiplier"; Ch. 5 "Governor", "Handshaking"

## The idea

A retry feels like resilience: the call failed, so try again. To the dependency on the other end,
it is more load, arriving just as it is struggling. Chapter 4 of *Release It!* has several names
for load a system inflicts on itself. A **dogpile** is many clients acting at the same moment, for
example all retrying on the same schedule as a dependency comes back, so recovery meets a wall of
traffic. **Self-denial** is a system whose own behaviour takes it down. A **force multiplier** is
automation that turns a small fault into a large one because it acts fast and everywhere. A retry
loop in every gateway is all three.

Retries also have a correctness problem the book's integration-point chapters keep returning to.
When a call times out, you do not know what happened on the other side. For a read that is fine:
read again. For a write it is not: the write may have committed, and doing it again does it twice.

The remedies are mostly about restraint. Chapter 5's **Governor** limits how fast automation may
act. Applied to retries: retry only operations that are safe to repeat; space retries out with
exponential backoff and random jitter so clients do not synchronise; cap the total with a budget,
so retries stay a small fraction of real traffic however bad things get; and never retry past the
caller's deadline. Inside the node, a writer that finds the file locked should wait its turn, not
spin.

## The flaw (F08, at `release-it/lesson-03`)

- `internal/nodeclient/client.go:50–74`: the gateway retries any transport error or node 5xx up to
  five times, 100 ms apart, **including `/execute`** (`:42`). Lesson 2 already ended the loop at the
  request's deadline, so a slow node gets at most one extra attempt per request. A node that fails
  *fast* gets six.
- `internal/engine/exec.go:223–237`: the node has no `busy_timeout`
  (`internal/engine/handles.go:94`). When a write finds the file's lock taken, it re-runs the
  statement 50 times in a tight loop with no sleep, then fails with `database is locked`, and the
  gateway retries that failure five more times.

## The patch

- **`internal/retry`** (new, ~120 lines). `Policy.Do` (`retry.go:75`) retries while the error is
  retryable, at most `MaxRetries` times, after a backoff drawn uniformly from
  [0, min(Cap, Base·2ⁿ)] ("full jitter"). It skips a retry when the deadline would not leave the
  backoff plus `MinAttempt` (`:80`). It also skips one when the client's **retry budget** is empty:
  a token bucket (`NewBudget`, `:44`) that every call refills by 0.1 and every retry drains by 1, so
  retries stay at or below 10 % of calls. Metrics: `retry_attempts_total{dep,outcome}`,
  `retry_budget_exhausted_total`, `retry_skipped_deadline_total`.
- **Reads only.** The node client retries `/query` through the policy (base 50 ms, cap 400 ms, two
  retries, 100 ms minimum attempt; `nodeclient/client.go:38`, `:52`). It no longer retries a
  `504 statement_deadline` (`:74`), because the same statement would hit the same deadline.
  `/execute` gets exactly one attempt (`:65`). After a timeout, the platform cannot know whether a
  tenant's transaction committed, and arbitrary SQL cannot be de-duplicated by the platform. A
  tenant that needs exactly-once writes has to make them idempotent itself (a unique key it
  generates, `INSERT ... ON CONFLICT DO NOTHING`) and retry on its side. Control's resolve is a
  read and gets the same policy (`internal/routes/client.go:40`, `:49`).
- **`busy_timeout` on the node.** Every tenant connection opens with `busy_timeout(2000)`
  (`handles.go:96`; it must be the first pragma). The `ncruces` driver implements the busy handler
  in Go: it backs off with random sleeps of up to ~2 ms, stops when the statement's context ends,
  and so stays inside the 2 s statement deadline. The 50-try loops are gone. A writer that still
  cannot get the lock answers `503 busy` and is counted in `sqlite_busy_total` (`exec.go:88`).
  `TestWriterWaitsForTheLock` holds the lock for 300 ms and checks that a write waits for it
  instead of failing.

## Verify

```
make verify LESSON=04 REF=release-it/lesson-03      # before
make verify LESSON=04                               # after
```

Steady mix at 60 % of fleet capacity (360 req/s) for 60 s. From t=10 s to t=30 s, every answer from
node-a is delayed by 1.5–3.5 s (Toxiproxy `latency` 2 500 ms, jitter 1 000 ms). About half of all
attempts therefore outlive the gateway's 2.5 s budget, after the node has already run them.
Throughout, a hot tenant on node-b sends 10 bulk writes per second (1 000 rows each) to one
database, so its writers contend for the file's lock. Then `make audit` uses DuckDB to read every
`notes` table on both node volumes and counts `seq` values that appear twice: writes applied twice.

Before (`lesson-03`):

```
  requests node-a received / tenant requests for node-a (t=14-28 s)   1.10
  seconds after the fault until node-a answers 99 % again            19
  rows applied twice (duplicate notes.seq across all tenant files)   308
  write attempts node-b received / writes the tenant sent           868 / 599 = 1.45
  writes node-b failed with 500 (database is locked)                  240.9
  writes node-b refused with 503 (pool or busy timeout)               321.5
FLAW REPRODUCED: recovery_s 19 (want <= 3); duplicate_rows 308 (want == 0); hot_db_load_ratio 1.45 (want <= 1.05); writes_failed_locked 240.9 (want == 0)
```

After:

```
  requests node-a received / tenant requests for node-a (t=14-28 s)   0.98
  retries the gateways made to node-a in that window                 376.9
  seconds after the fault until node-a answers 99 % again            2
  rows applied twice (duplicate notes.seq across all tenant files)   0
  write attempts node-b received / writes the tenant sent           599 / 599 = 1.00
  writes node-b failed with 500 (database is locked)                  0
  writes node-b refused with 503 (pool or busy timeout)               33.81
  retry_budget_exhausted_total                                        1532
IMPROVED: node_a_load_ratio 0.98 <= 1.3; recovery_s 2 <= 3; duplicate_rows 0 == 0; hot_db_load_ratio 1.00 <= 1.05; writes_failed_locked 0 == 0
```

What changed:

- **Duplicates.** Before the patch, a write that timed out at the gateway after committing on the
  node was sent again and committed again: hundreds of rows in tenants' tables that nobody wrote.
  After the patch, the tenant gets a 504 and the row exists once.
- **The hot database.** Before, colliding writers failed in microseconds, and each failure came
  back up to five more times, so the node received 1.45 attempts per write and refused hundreds.
  After, writers queue for the lock, and the node receives exactly one attempt per write. A few are
  refused fast with 503 when four bulk writers already hold the database's connections, which is
  the bounded pool from lesson 3 doing its job.
- **Recovery.** The before run took 19 s to get node-a back to 99 %, because the retries of
  requests that were already failing kept it busy after the fault ended. After the patch: 2 s. The
  budget did its work: 1 532 retries were refused.
- **What did not change much** is the load ratio on node-a during the slow phase (1.10 before,
  ~1.0 after). The plan expected 4–6× here. Lesson 2 already made every retry stop at the request's
  deadline, so an attempt that timed out at 2.5 s could have at most one more short try. The
  slow-dependency storm was mostly closed by deadlines. The fast-failure storm, and the duplicate
  writes, are what this lesson closes.

## Try changing

- Set `MaxRetries: 5` and `NewBudget(1, 100)` and rerun. Then watch `retry_attempts_total` against
  `http_requests_total` on node-b while the hot tenant runs.
- Remove the jitter (`wait := p.Cap`): retries synchronise, and recovery after the fault shows a
  sawtooth.
- Give `/execute` the read policy and run `make audit` after the verify: duplicates come back. Then
  change the loadgen's write to `INSERT ... ON CONFLICT(seq) DO NOTHING` on a table with a unique
  `seq`: retries become safe because the tenant made the write idempotent.

## Read more

*Release It!* 2nd ed., Ch. 4 "Dogpile", "Self-Denial Attacks", "Force Multiplier"; Ch. 5
"Governor". Marc Brooker, "Exponential Backoff And Jitter" (AWS Architecture Blog), for full jitter.
The Google SRE book's chapter on handling overload, for retry budgets. SQLite's `busy_timeout`
documentation.

## Deviations from the plan

- **The fault is `latency 2500 ms ± 1000 ms` on every connection, not `1500 ms` with
  toxicity 0.5.** Toxicity is chosen per connection. Slow connections get closed by timeouts and the
  gateways keep reusing the healthy ones, so the effect faded within seconds.
- **The toxic is removed by disabling and re-enabling the proxy** (`untoxic_busy` in `lab/lib.sh`).
  Removing a latency toxic directly from a proxy with hundreds of open links wedged Toxiproxy 2.12
  for the rest of the run.
- **The hot database on node-b is new.** It is what shows the fast-failure retry storm and the
  `busy_timeout` change. The plan's node-a scenario alone shows duplicates but, after lesson 2, only
  a ~1.1× load ratio. That is reported, but it is no longer the headline.
- **Node-a's received requests are counted from its request log, not Prometheus.** Counter
  increases over a 14 s window undercounted badly in several runs.
- **Lab:** Postgres, control and Toxiproxy run with `cpu_shares: 4096`, so on this shared VM
  (another project was load-testing on it during the build) they stand in for machines of their
  own. Postgres is scraped every 5 s instead of every 1 s, and Toxiproxy's per-chunk proxy metrics
  are off.
