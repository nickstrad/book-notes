# 08 · Create Back Pressure (and when to shed instead)

> **STUB** — teaching outline from `../../PLAN.md` (topic 08). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 5 "Create Back Pressure", "Shed Load"; Ch. 9 "Demand
Control"; Ch. 4 "Unbalanced Capacities".
**Language:** Rust — bounded vs unbounded is `sync_channel` vs `channel`, and blocking vs shedding
is `send` vs `try_send`: the lesson is two function names.
**Run:** `./run.sh`.
**Budget:** 15 s.

## The idea

When a producer outruns its consumer, something must give: the queue grows without bound (memory,
then death), the producer is slowed (back pressure), or work is dropped (shed). Back pressure
belongs inside a system boundary where the producer is yours to slow; at the boundary, shed load
with a clear response instead of letting a stranger's demand fill your queue.

_Builder: 2–4 paragraphs; cite the book's rule that back pressure should be applied within a
system boundary and its examples (blocking sends, TCP flow control, bounded pools)._

## What you will see

A producer emitting 1000 msg/s; a consumer handling 200 msg/s. Every 500 ms: `produced`,
`consumed`, `depth`, `oldest_age_ms`, and (after) `producer_blocked_ms` or `shed`.

### Before

`std::sync::mpsc::channel()` (unbounded). Depth and oldest-age grow linearly; at 5 s depth ≈ 4000,
age ≈ 4 s.

```
UNBOUNDED: queue grows 800/s; this runs until OOM
```

### After

Two variants in one run: `sync_channel(100)` with blocking `send` (producer slows to 200/s, depth
≤ 100, age ≤ 0.5 s), then `try_send` at the "boundary" where overflow is counted and rejected
(`TrySendError::Full`) with the consumer's throughput untouched.

## How it works

_Builder: three short programs or one `after.rs` with two phases; `src/bin/before.rs`,
`src/bin/after.rs`._

## Try changing

_Builder: e.g. buffer 1 (lock-step), buffer 10 000 (bounded but a 50 s latency — bounded is not
the same as small)._

## Where to read more

Ch. 5 "Create Back Pressure", "Shed Load"; Ch. 9 "Demand Control"; topic 05.

## Deviations from plan

None yet.
