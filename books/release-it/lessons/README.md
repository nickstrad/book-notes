# Release It! lessons — index

One note per lesson. Each lesson is one tag on `main`. The note has five parts: the idea with
chapter citations, the flaw with `file:line` pointers into the parent tag, the patch walkthrough,
the before and after verdicts, and its deviations from `../PLAN.md`.

Run every command from `books/release-it/project`. `make verify LESSON=NN` judges the checked-out
tree. Add `REF=<parent tag>` to build the parent's images and run today's script against them: that
reproduces the flaw. Every lesson's verify, as it is at `main`, also passes on the `main` tree:
lessons 01–03's checks accept what later lessons changed on purpose (lesson 05's `degraded` gateway
health, lesson 07's `503 overloaded` shedding). A lesson tag carries the scripts as they were at
that tag. One defect was found after the review: from lesson 08 a cut-off query left no request
log line. It is fixed at `release-it/v1-hardened`; see lesson 01's note, "After the review".

| # | Tag | Note | Flaws | Verify | Before (`REF=` parent) | After |
|---|---|---|---|---|---|---|
| 00 | `release-it/v0-flawed` | [The flawed start](00-the-flawed-start.md) | the register, F01–F18 | `make verify LESSON=00` | — | baseline holds: 100 % served at 50 % load, p99 ≈ 30 ms; `up = 0` for every platform service |
| 01 | `release-it/lesson-01` | [Transparency](01-transparency.md) | F01 | `make verify LESSON=01` | nothing scraped, no health, 0 % JSON logs | 3 gateways up; a slow dependency visible as p99 1.02 s; health fails in 0.8 s |
| 02 | `release-it/lesson-02` | [Timeouts](02-timeouts.md) | F02–F04 | `make verify LESSON=02` | 1 130 requests in flight, 0 % answered 504 (every failure was the client giving up), runaway statements | 149 in flight; 99.4 % answer 504 inside 2.9 s; a statement cancels in 0.5 ms |
| 03 | `release-it/lesson-03` | [Blocked threads](03-blocked-threads.md) | F05–F07 | `make verify LESSON=03` | node at 1 024 fds, 14 750 5xx, 22 % valid writes | 471 fds, no node 5xx, 99.9 % valid writes |
| 04 | `release-it/lesson-04` | [Retries](04-retries.md) | F08 | `make verify LESSON=04` | 308 duplicate rows, 19 s recovery | no duplicates, 2 s recovery, hot-database load ratio 1.00 |
| 05 | `release-it/lesson-05` | [Circuit breaker](05-circuit-breaker.md) | F09 | `make verify LESSON=05` | 0 % served while control is down, gateways unhealthy | 99.25 % served from the route cache; breaker closes 4.8 s after recovery |
| 06 | `release-it/lesson-06` | [Bulkheads](06-bulkheads.md) | F10–F11 | `make verify LESSON=06` | neighbours' p99 534 ms, 264 control pool timeouts, backup p99 2 s | neighbours' p99 87 ms, zero pool timeouts, writer lock held 0.5 s |
| 07 | `release-it/lesson-07` | [Shed load](07-shed-load.md) | F12 | `make verify LESSON=07` | 11 % goodput after a gateway dies, p99 2.8 s | 68.5 % goodput, p99 184 ms, gateway in-flight ≤ 13 |
| 08 | `release-it/lesson-08` | [Unbounded result sets](08-unbounded-results.md) | F13 | `make verify LESSON=08` | node OOM-killed 5 times, no whale scan answered | zero restarts, every scan answered capped, report 0.29 s |
| 09 | `release-it/lesson-09` | [Back pressure](09-back-pressure.md) | F14 | `make verify LESSON=09` | queue 4 076 deep, 2 921 usage events lost on restart | queue ≤ 155, zero events lost, outbox empty |
| 10 | `release-it/lesson-10` | [Steady state](10-steady-state.md) | F15 | `make verify LESSON=10` | WAL +14 MB/min, volume full in 80 s, handles and ledger grow | WAL steady at ~3 MB, handles at the cache size, ledger shrinking |
| 11 | `release-it/lesson-11` | [Test harness](11-test-harness.md) | F16 | `make verify LESSON=11` (also `make test-chaos`) | 5 of 12 chaos tests fail | 12 of 12 pass |
| 12 | `release-it/lesson-12` | [Handling versions](12-handling-versions.md) | F17–F18 | `make verify LESSON=12` (`REF=… REF_SERVICES=gateway` for old gateways); `make godot` | 99.9 % of queries fail on a mixed fleet; 5 requests cut off by a node restart | zero errors through the migration and the rollout, nothing cut off, the mixed fleet serves every query |
| — | `release-it/v1-hardened` | this index and `../README.md` | | | | |

The full verdict lines are in each note's "Verify" section and in its lesson commit's message
(`git log release-it/v0-flawed..release-it/v1-hardened`). Numbers vary a little from run to run,
more so on a busy VM. The thresholds include headroom for that, and each note says where a threshold
differs from the plan's.
