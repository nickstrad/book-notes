# 05 · Chain Reactions and Shedding Load

> **STUB** — teaching outline from `../../PLAN.md` (topic 05). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 4 "Chain Reactions", "Cascading Failures", "Unbalanced
Capacities", "Scaling Effects"; Ch. 5 "Shed Load", "Fail Fast", "Handshaking"; Ch. 9 "Demand
Control" (how systems fail under load; preventing disaster).
**Language:** Go — a small simulation whose shedding loop has the same shape as topic 04's
bulkhead rejection, so one idea shows up in two places.
**Run:** `./run.sh`.
**Budget:** 10 s.

## The idea

In a horizontally scaled layer, when one instance dies its share of load lands on the survivors.
If they were near capacity, the next one falls, then the next: a chain reaction to zero. A layer
that refuses work it cannot do (bounded queues, immediate 503s) survives at reduced capacity.
Unbalanced capacities between layers make the first fall likely.

_Builder: 2–4 paragraphs; distinguish chain reaction (within a layer) from cascading failure
(across layers) as the book does; cite Ch. 9 "Demand Control" for "how systems fail under load"._

## What you will see

A discrete-time simulation, no network: 4 instances at 100 req/s capacity each, offered load
320 req/s, round-robin balancer. At t = 3 s instance 1 is killed. An ASCII timeline per second
shows each instance's state (`ok`, `overloaded`, `dead`) and totals `served`, `rejected`, `lost`.

### Before

Unbounded queues. Survivors take 107 req/s each; queues grow; an instance dies when its queue
exceeds a memory threshold; load re-spreads; by about t = 8 s every instance is dead.

```
t=8s  [dead][dead][dead][dead]  served=0 rejected=0 lost=320
CHAIN REACTION: 4 of 4 instances lost
```

### After

Each instance bounds its queue and sheds the excess with an immediate rejection. Survivors serve
300 req/s and reject 20; no further loss.

```
t=8s  [dead][ ok ][ ok ][ ok ]  served=300 rejected=20 lost=0
HELD: 3 of 4 instances serving at capacity; excess shed, not absorbed
```

## How it works

_Builder: the per-instance queue model and the single `if len(queue) >= max { reject }` line that
is the whole remedy; `before/main.go`, `after/main.go`._

## Try changing

_Builder: e.g. offered load 390 (97% utilisation) so even shedding is marginal; a 10× front-end
layer to show Unbalanced Capacities._

## Where to read more

Ch. 4 "Chain Reactions", "Unbalanced Capacities"; Ch. 5 "Shed Load"; Ch. 9 "Demand Control".

## Deviations from plan

None yet.
