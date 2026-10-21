# 02 · Blocked Threads and a resource pool that fails fast

> **STUB** — teaching outline from `../../PLAN.md` (topic 02). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 4 "Blocked Threads" (spot the blocking; use caching
carefully; libraries); Ch. 2 case study (the `close()` exception that drained the pool); Ch. 5
"Fail Fast", "Timeouts".
**Language:** Rust — the remedy the book describes (return the resource on every path) IS a `Drop`
guard, and `Condvar::wait_timeout` makes the bounded wait a different function, not a flag.
**Run:** `./run.sh` (builds with cargo, std-only; then `before`, then `after`).
**Budget:** 15 s.

## The idea

Blocked threads are the proximate cause of most outages. Request threads block on a shared
resource pool; the pool does not refill because a holder leaked its resource or is stuck on a slow
call; soon every thread is blocked and even requests that never needed the resource are starved.
Checking out with no timeout turns one leak into a frozen process. This is the mechanism of the
book's opening case study: an exception thrown from `close()` left connections un-returned until
the pool was empty and every request thread was waiting.

_Builder: 2–4 paragraphs; cite Ch. 2 "The Smoking Gun" and Ch. 4 "Blocked Threads"._

## What you will see

A `Pool` of 4 connections guarded by `Mutex<Vec<Conn>>` + `Condvar`. 16 worker threads serve a
scripted stream of requests; 1 in 8 requests hits an error path. A monitor thread prints every
500 ms: `available`, `checked_out`, `waiting`, `served`.

### Before

The error path returns early without returning the connection; `checkout()` uses `Condvar::wait`
with no timeout. `available` falls to 0 within ~2 s; `waiting` climbs to 16; `served` stops. The
watchdog at 5 s exits 1:

```
t=5.0s available=0 checked_out=4 waiting=16 served=41
FROZEN: 16 threads waiting on 0 connections; nothing was ever going to wake them
```

### After

Two changes, visible separately in the output: (1) an RAII guard whose `Drop` returns the
connection, so the error path cannot leak; (2) `Condvar::wait_timeout(250 ms)` so `checkout()`
returns `Err(PoolTimeout)` and the request fails fast. `served` keeps rising; a `fast_failed`
column appears instead of waiting threads.

## How it works

_Builder: show the guard type and the `wait_timeout` loop; one cargo package with
`src/bin/before.rs` and `src/bin/after.rs` (shared code may live in `src/lib.rs` within this topic)._

## Try changing

_Builder: e.g. keep the guard but drop the timeout — the pool no longer leaks but a slow holder
still blocks; raise the error rate._

## Where to read more

Ch. 2 whole chapter; Ch. 4 "Blocked Threads"; Ch. 5 "Fail Fast"; topic 04 for isolating pools.

## Deviations from plan

None yet.
