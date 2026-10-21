# 03 · Circuit Breaker

> **STUB** — teaching outline from `../../PLAN.md` (topic 03). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 5 "Circuit Breaker"; Ch. 4 "Integration Points", "Slow
Responses", "Cascading Failures".
**Language:** Rust — an exhaustive `match` over the state enum is the whole pattern, and the
compiler forces every transition to be handled.
**Run:** `./run.sh`.
**Budget:** 20 s (two 8 s timelines plus build).

## The idea

When a dependency is failing, continuing to call it harms both sides: each call costs the caller a
full timeout, and the dependency gets no room to recover. A breaker counts failures, trips open so
calls fail immediately, and after a cooldown lets one trial call through (half-open) to test for
recovery. The book insists state changes be logged and the current state exposed to operations.

_Builder: 2–4 paragraphs; include the book's note that a breaker is for protecting the caller and
that what to do when it is open (fallback, cached answer, error) is a business decision._

## What you will see

A simulated dependency scripted on a timeline: healthy (5 ms) for 2 s, down for 4 s (every call
consumes the 300 ms timeout then fails), healthy again. A caller issues 20 req/s. One row per
second: `state`, `attempted`, `reached_dependency`, `failed`, `mean_latency_ms`.

### Before

No breaker. During the outage every call takes 300 ms; mean latency jumps 60×; the dependency
absorbs ~80 calls while down. Last line names the cost:

```
NO BREAKER: 80 calls hit a dead dependency; caller latency 5ms -> 300ms for 4s
```

### After

`enum State { Closed { failures }, Open { since }, HalfOpen }`; trips at 5 consecutive failures;
cooldown 1 s. Calls fail in under 1 ms with `Err(CircuitOpen)`; the dependency receives ~4 probe
calls; recovery detected within one cooldown. Transition lines appear:

```
breaker: closed -> open (5 consecutive failures)
breaker: open -> half-open (cooldown elapsed)
breaker: half-open -> closed (probe succeeded)
```

## How it works

_Builder: the `match` on state in `call()`; a simulated clock so the timeline is deterministic;
`src/bin/before.rs`, `src/bin/after.rs`._

## Try changing

_Builder: e.g. threshold 1 (flappy), cooldown 100 ms (hammering), make the probe fail once._

## Where to read more

Ch. 5 "Circuit Breaker"; Ch. 4 "Cascading Failures"; topic 11 for exposing breaker state.

## Deviations from plan

None yet.
