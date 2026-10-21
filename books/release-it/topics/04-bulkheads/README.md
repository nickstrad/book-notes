# 04 · Bulkheads

> **STUB** — teaching outline from `../../PLAN.md` (topic 04). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 5 "Bulkheads"; Ch. 4 "Chain Reactions", "Cascading
Failures", "Blocked Threads".
**Language:** Go — buffered channels as semaphores are the exact pool model, `net/http` gives a
real server, and the before/after difference is a dozen lines.
**Run:** `./run.sh`.
**Budget:** 20 s.

## The idea

Partition capacity so a failure in one compartment cannot sink the whole ship. The common sinking:
one shared thread or connection pool serves both a critical path and a slow, noisy one; the slow
one occupies every slot and the critical path dies of starvation. Bulkheads decide the blast radius
of every other failure.

_Builder: 2–4 paragraphs; include the book's examples of bulkheads at different granularities
(separate pools per caller, dedicated instances for key customers, process/thread isolation)._

## What you will see

An in-process HTTP server with `/checkout` (5 ms) and `/report` (calls a dependency hanging to a
2 s timeout). A worker semaphore of 20 models the request thread pool. A generator sends 50 req/s
to each path for 8 s. Per second, per path: `ok`, `rejected`, `p50_ms`.

### Before

One shared semaphore. `/report` fills all 20 slots within a second; `/checkout` success falls
toward 0 and its p50 climbs to the dependency timeout.

```
CHECKOUT STARVED: 20/20 workers held by /report; checkout ok 100% -> 4%
```

### After

Separate semaphores per path (10 and 10), rejecting with 503 when a compartment is full. `/report`
degrades (503s) while `/checkout` stays at 100% with 5 ms p50.

## How it works

_Builder: the two semaphores and the `select` with `default` that turns "full" into 503;
`before/main.go`, `after/main.go`, one `go.mod`; prefer real loopback HTTP, fall back to in-process
calls only if timing is unreliable._

## Try changing

_Builder: e.g. give `/report` 2 slots instead of 10; add a third path._

## Where to read more

Ch. 5 "Bulkheads"; Ch. 4 "Chain Reactions"; topic 05 (what happens across instances).

## Deviations from plan

None yet.
