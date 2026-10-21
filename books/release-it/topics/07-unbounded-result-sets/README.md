# 07 · Unbounded Result Sets and Slow Responses

> **STUB** — teaching outline from `../../PLAN.md` (topic 07). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 4 "Unbounded Result Sets" (black Monday; limits on other
application-level protocols) and "Slow Responses".
**Language:** Node — `process.memoryUsage()` and a drifting `setInterval` make memory and
event-loop blocking visible with no tooling; async iterators are the natural streaming form.
**Run:** `./run.sh`.
**Budget:** 20 s.

## The idea

A query written for today's data returns "all matching rows". Production data grows; one day the
result set is a million rows and the service stalls or dies of memory. Slow responses are worse
than fast failures because they hold the caller's resources while they drag on. Bound what you ask
for (`LIMIT`, pagination), bound what you return, and stream rather than materialize.

_Builder: 2–4 paragraphs; include the book's point that the server must protect itself regardless
of what the client asked for, and that this applies beyond SQL to any list-returning API._

## What you will see

An in-memory `orders` table stepping 1 k → 10 k → 100 k → 1 M rows, simulating months passing. A
handler answers "all orders for customer X" at each step. A 100 ms heartbeat logs its drift. Per
step: `rows`, `response_ms`, `heap_mb`, `max_drift_ms`.

### Before

`query()` returns an array of every match; `JSON.stringify` the lot. Heap and response time scale
with the table; drift reaches hundreds of ms.

```
UNBOUNDED: response grew 900x with data; heap +180 MB; event loop blocked 640 ms
```

(README should also show `node --max-old-space-size=64 before.mjs` to make it crash for real.)

### After

`query({ limit: 100, after: cursor })` returning an async iterator; responses stream pages. Heap
and drift are flat across all four steps.

## How it works

_Builder: the `query` signature difference and the `for await` loop; `before.mjs`, `after.mjs`;
plain in-memory array, no SQLite dependency (mention that a real database behaves the same)._

## Try changing

_Builder: e.g. limit 10 000 (still bounded, still fine); remove the heartbeat to see why blocking
is invisible without one._

## Where to read more

Ch. 4 "Unbounded Result Sets", "Slow Responses"; topic 08 for bounding queues the same way.

## Deviations from plan

None yet.
