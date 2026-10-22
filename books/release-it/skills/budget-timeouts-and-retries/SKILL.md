---
name: budget-timeouts-and-retries
description: Build a nested deadline budget along a call chain from the edge SLO inward (each hop's limit inside its parent's remaining time; connect, response-header and total split), give long-running work real cancellation, and set a retry policy that is idempotent-only, bounded, jittered, budgeted and stops at the deadline. Use when setting any timeout, writing retry logic, tuning a client, or adopting a new latency SLO.
---

# budget-timeouts-and-retries

Every wait needs a limit, and the limits must fit inside each other. An inner wait that outlasts
its caller does work nobody will receive; an inner wait equal to its caller's races it. Retries
spend the same budget and add load at the worst moment, so they get a policy, not a loop.

## When to use

- Setting or reviewing any timeout: client, server, pool checkout, statement, balancer.
- Writing or reviewing a retry loop.
- A new or changed latency SLO at the edge.
- After a storm: "the dependency came back but we stayed down", "rows were written twice".

## Inputs

- The call chain from the edge inward: every hop, including pool waits and the work itself
  (statements, scans, uploads).
- The client-facing timeout or SLO at the edge.
- For each operation: is it idempotent? Can the callee tell a repeat from a new request?
- Normal latency of each hop (p50/p99) under load, to know the legitimate maximum.

## Procedure

1. **Start at the edge.** Write the time the end client will wait (its timeout, or the SLO). The
   edge's own deadline sits a little **inside** it (e.g. 2.9 s inside 3 s), so the edge's
   answer arrives before the client hangs up.

2. **Walk inward, nesting.** For each hop, its total limit < the parent's remaining time at the
   moment of the call, leaving room for the parent to turn the failure into an answer. The work at
   the bottom (statement, scan) gets the tightest limit, and the callee enforces it itself.
   Check: no inner limit ≥ outer limit anywhere on the chain.

3. **Split each client limit into three numbers.**
   - **Connect (dial):** short (100–300 ms on a LAN). A refused or unroutable peer should fail
     fast.
   - **Response header:** ≥ the longest *legitimate* work the callee does before its first byte
     (a server that sends headers only after the work finishes needs this as long as the work's
     own deadline).
   - **Total:** the hop's budget from step 2. Bind the request to the caller's context so the
     total also shrinks when the parent's deadline is closer.

4. **Bound the servers and the balancer.** Read-header, read, write and idle timeouts on every
   server; balancer client/server/queue timeouts in seconds, not minutes. A stalled client must
   not hold a goroutine and a socket indefinitely.

5. **Make cancellation real.** A deadline that only stops *waiting* leaves the work running.
   Check the callee actually stops: database statement interrupts (`sqlite3_interrupt`,
   `statement_timeout`, query cancel), context checks in loops, sleeps that wake on cancellation.
   Measure how late the work stops after its deadline.

6. **Bound pool and lock waits too.** A checkout from a pool, a lock wait (`busy_timeout`), a queue
   enqueue: each gets a limit inside the hop's budget, and fails fast with a distinct error.

7. **Set the retry policy.**
   - **Only idempotent operations.** After a timeout the caller does not know whether a write
     happened; never retry a write whose outcome is unknown unless the callee de-duplicates
     (idempotency key, `ON CONFLICT DO NOTHING`, event id).
   - **Not when retrying cannot help:** the callee's own deadline exceeded (the same work will hit
     it again), bad payloads, 4xx.
   - **Few:** 1–3 retries.
   - **Exponential backoff with full jitter:** sleep uniform in [0, min(cap, base·2ⁿ)].
   - **A budget:** a token bucket so retries stay a fixed fraction (≈10 %) of calls however bad
     things get.
   - **Inside the deadline:** skip a retry if the remaining time is less than backoff + a minimum
     useful attempt.
   - **Paired with a breaker:** stop retrying a dependency the breaker has declared down.
   - Inside a process, a writer that finds a lock taken should *wait* (busy timeout with backoff),
     not spin.

8. **Sanity-check with Little's law.** Requests in flight at a hop ≈ arrival rate × time each
   holds a slot. With limits, the worst case is rate × deadline; compare it to the instance's
   concurrency limits and memory.

9. **Expose it.** Count timeouts per dependency and per reason, cancellations, retries by outcome,
   budget exhaustions, deadline-skipped retries.

## Output

A budget table and a retry table:

```
| Wait | Limit (dial / header / total) | Inside | Enforced at (file:line) | On expiry the caller gets |
| Operation | Idempotent? | Retries | Backoff (base, cap, jitter) | Budget | Not retried when | Breaker |
```

Plus the list of violations found (inner ≥ outer, missing limits, retried writes, spinning lock
waits) and Little's-law worst-case in-flight per hop.

## Source

*Release It!* 2nd ed.: Ch. 5 "Timeouts", "Circuit Breaker", "Governor"; Ch. 4 "Integration Points",
"Slow Responses", "Blocked Threads", "Dogpile", "Self-Denial Attacks". Demonstrated in lessons 02
(`../../lessons/02-timeouts.md`) and 04 (`../../lessons/04-retries.md`), with the breaker of 05.

## Worked example: litefleet lessons 02 and 04

**The budget** (lesson 02, `internal/httpx/client.go` builds every client from dial / header /
total), edge to bottom:

| Wait | Limit | Inside | Where |
|---|---|---|---|
| client gives up | 3 s | — | loadgen |
| gateway request | 2.9 s | 3 s | `internal/api/handlers.go` `RequestDeadline` |
| gateway → node attempt | dial 200 ms, header = total 2.5 s | 2.9 s | `internal/nodeclient/client.go` |
| statement on the node | 2 s (`sqlite3_interrupt`) | 2.5 s | `internal/engine/exec.go` `StatementDeadline` |
| pool checkout on the node | 250 ms | 2 s | `exec.go` `PoolWait` (lesson 03) |
| gateway → control resolve | dial 200 ms, total 500 ms; control budgets itself 450 ms | 2.9 s | `internal/routes/client.go` |
| servers | header 5 s, read 10 s, write 15 s, idle 60 s | — | `internal/httpx/httpx.go` |
| HAProxy | client/server 10 s, `http-request` 5 s | — | `deploy/haproxy.cfg` |

Step 3's header rule changed the plan: it asked for a 1 s header timeout everywhere, but the node
sends headers only after its statement finishes, so the header limit equals the 2.5 s total. Lesson
09 hit the same rule from the other side: a 1 s header timeout against a metering partner that
answers in 1 s counted every delivered batch as failed and sent it again.

Measured with node-a black-holed for 20 s at 300 req/s (`make verify LESSON=02`):

- Gateway in-flight: **1 130 → 149**; goroutines 4 650 → 632. Step 8 predicts it: node-a gets
  half the 300 req/s, a third of that per gateway, so 50 req/s × 2.9 s ≈ 145 per gateway.
- Node-a failures answered `504` by the gateway: **0 % → 99.4 %**, p99 2.914 s (< 3 s).
- 1 200 requests were still stuck 6 s after the fault ended before; 2 after.
- A runaway `WITH RECURSIVE`: 0.495 cores held for a departed client before; cancelled **0.5 ms**
  after its 2 s deadline after.
- A half-sent request held the gateway 9.0 s before (gave up), 5.0 s after.

**The retry policy** (lesson 04, `internal/retry/retry.go`): reads only, base 50 ms, cap 400 ms,
2 retries, full jitter, 100 ms minimum attempt, budget `NewBudget(0.1, 10)` (`nodeclient/client.go:39`,
`routes/client.go:51`). `/execute` gets exactly one attempt. A `504 statement_deadline` is not
retried. SQLite connections open with `busy_timeout(2000)` instead of a 50-try spin.

Measured with node-a slowed to 1.5–3.5 s for 20 s and a hot database on node-b
(`make verify LESSON=04`):

- Rows applied twice (DuckDB audit of every tenant's `notes`): **308 → 0**.
- Recovery to 99 % after the fault: **19 s → 2 s**; the budget refused 1 532 retries.
- Attempts per write on the hot database: **1.45 → 1.00**; `database is locked` failures 240.9 → 0.
- Load ratio on the slow node: 1.10 → 0.98. Small, because step 2's nesting (lesson 02) had already
  capped each request at one extra try. The deadline closed the slow-dependency storm; the policy
  closed the fast-failure storm and the duplicates.

One more thing the after-numbers of lesson 02 teach: a node-a request took 2.9 s to fail, not
2.5 s, because the old retry loop spent the last 0.4 s on a second attempt. That is step 7's
"inside the deadline" rule, measured before it existed.
