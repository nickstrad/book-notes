---
name: review-capacity-headroom
description: Review whether a system has the capacity and headroom to survive losing an instance, a burst, or its own self-inflicted peaks — measuring per-layer capacity open-loop, checking front-to-back capacity ratios and pool math against real limits, verifying N−1 survival, shed thresholds and balancer limits, and extrapolating anything that grows. Use when scaling a layer, sizing pools, planning a launch or campaign, or configuring autoscaling.
---

# review-capacity-headroom

Capacity is not "how fast is it", it is "what happens when demand passes what it can do, and how
close are we". This skill measures each layer, checks that the layers fit each other, and checks
that the system degrades by refusing the excess instead of collapsing, including when one instance
dies, when a cache empties, and when its own jobs run.

## When to use

- Scaling a tier up or down, or changing instance size.
- Sizing connection pools, worker pools, handle caches, queue bounds.
- Planning a launch, a marketing campaign, a migration window.
- Configuring autoscaling or balancer limits.
- After a chain reaction: "one instance died and then they all did".

## Inputs

- Topology with instance counts and resource limits (CPU, memory, `nofile`, disk).
- A way to drive load **open-loop** (a fixed arrival rate, independent of responses) with a
  realistic mix, and per-request results (status, latency, instance).
- Current and expected peak demand, and the latency SLO.
- Pool sizes and downstream limits (database `max_connections`, partner rate limits).

## Procedure

1. **Measure each layer's capacity, open-loop.** For one instance of each tier, step the offered
   rate up until success < 99.5 % or p99 > the SLO; the last good step is its capacity. A
   closed-loop generator (waits for each answer) hides overload: the offered rate falls as latency
   rises. Record the numbers in a file the rest of the work reads.

2. **Find the bottleneck and the ratios.** Fleet capacity = min over tiers of (instances ×
   per-instance capacity). Compute front:back ratios. Where the front can offer more than the back
   can take (ratio > 1), the back tier needs a limit of its own; it cannot rely on the front to
   protect it.

3. **Do the pool math against real limits.**
   - Σ (pool size × instances) ≤ the downstream's connection limit, with margin.
   - Per-process descriptor budget: `nofile` − reserved ÷ fds per pooled resource = how many
     resources you may hold. Enforce it in code; turn the wall into a fast, named refusal.
   - Pool size vs work: a pool smaller than (arrival rate × hold time) queues; bound the wait.
   - Cache sizes ≥ the working set, or the cache thrashes.

4. **Check N−1 (and N−k) headroom.** For each stateless tier: at expected peak with one instance
   gone, each survivor carries peak / (N−1). If that is above its capacity, the tier will either
   shed (good) or collapse (chain reaction). For stateful tiers (users pinned to a shard): nothing
   routes around a lost instance, so the question is the shard's own headroom and how fast and
   cleanly it comes back.

5. **Verify shedding and handshaking.**
   - Every instance has an in-flight limit; past it, `503` with `Retry-After`, fast, without
     touching the work. Health and metrics endpoints are never shed.
   - Size the limit from Little's law: latency budget × per-instance throughput (the concurrency
     the instance actually sustains at its calibrated capacity), not a round number.
   - Low-priority routes (reports, listings) share a fraction of the limit so they cannot crowd out
     the critical path.
   - The balancer: per-server `maxconn` a little **above** the instance's own limit (the instance
     refuses first, cheaply), a short queue timeout, retry only connection failures on another
     server, never a request a server has accepted.
   - Health says `degraded` while shedding, not `failed`; a shedding instance must stay in rotation.

6. **Estimate self-inflicted peaks.** List what your own system does that adds load all at once:
   cache expiry (all TTLs from one warm-up), restarts (cold caches, reopening every resource),
   deploys (one instance out at a time = N−1 for the duration), migrations and backfills, purges,
   backups, cron jobs on the hour, retries during an incident, campaigns that deep-link to
   expensive pages. For each: how much extra load, on which tier, and what limits it (jittered
   refresh, singleflight on open, batches with pauses, rate limits).

7. **Extrapolate everything that grows.** For each accumulator (disk, WAL/log, table rows, open
   handles, memory), sample at three points under load and compute when it hits its wall. Data
   that should grow is a capacity-plan input; anything else that grows is a leak.

8. **Governors on automation.** Autoscalers and other automation that adds or removes capacity:
   bounded step size, bounded rate, a floor and a ceiling, so a bad signal cannot remove the fleet.

## Output

```
| Tier | Instances | Capacity / instance (measured) | Tier capacity | Peak demand | Utilization | N−1 utilization | In-flight limit | Shed behaviour |
```

then the pool table (`pool | size × instances | downstream limit | margin`), the self-inflicted
peaks (`event | extra load | tier | limiter`), the growth table
(`resource | rate | wall | time to wall`), and a verdict per tier: OK / sheds under N−1 / collapses
under N−1.

## Source

*Release It!* 2nd ed.: Ch. 4 "Scaling Effects", "Unbalanced Capacities", "Chain Reactions",
"Users", "Self-Denial Attacks"; Ch. 5 "Shed Load", "Handshaking", "Governor", "Steady State";
Ch. 9 "Load Balancing" and "Demand Control"; Ch. 15 case study (capacity under real customers).
Demonstrated in lessons 03, 07 and 10 of `../../lessons/`.

## Worked example: litefleet

**Step 1.** `make calibrate` (`project/lab/calibrate.sh`, open-loop `cmd/loadgen`) writes
`project/lab/capacity.env`: one gateway 350 req/s, one node 300 req/s at `cpus: 0.5`. Every verify
offers a fraction of these numbers.

**Step 2.** Three gateways offer 1 050 req/s; two nodes take 600. Front:back is 1.75 : 1, and the
nodes are the fleet's bottleneck (`FLEET_CAPACITY=600`). That is the unbalanced-capacities case:
the node gets its own in-flight limit (48, `cmd/node/main.go`) because three gateways at their
limit of 24 can have 72 requests in flight toward it.

**Step 3.** The node's handle budget (lesson 03, `internal/engine/handles.go`):
(1 024 − 124 reserved) ÷ (2 idle connections × 3 fds) = **150** databases; past it, `503 node_full`
instead of an fd error inside SQLite. Lesson 10's LRU of 128 sits under that, and above the
measured working set (~123 databases per node); the plan's 64 would have thrashed. Control's
Postgres pools sum to 8 + 2 + 1 per control instance, well inside `max_connections` 100.

**Steps 4–5.** Lesson 07 runs N−1 for real. The gateways are made the bottleneck
(`L07_GATEWAY_CPUS=0.15`, 215 req/s each), offered 80 % of three (516 req/s), then
`docker compose kill gateway-b`, so the two survivors are offered 120 %:

| | before (`lesson-06`) | after (`lesson-07`) |
|---|---|---|
| goodput after the kill (share of before) | **11.4 %** | **68.5 %** (2/3 is the arithmetic ceiling) |
| p99 of successful requests | 2 784 ms | 184 ms |
| max gateway in-flight / goroutines | 997 / 3 228 | 13 / 136 |
| shed by the gateways (`shed_total`) | 0 | 8 288 |
| killed gateway serving again after restart | — | 1 671 requests |

Before, the survivors accepted everything, their queues grew to ~1 000, latency passed the 3 s
client timeout, their health checks waited in the same queue, and HAProxy ended with no server:
the chain reaction completed. After: `internal/shed` (limit 24 at half a CPU, ~60 ms of work;
scaled to 12 for the 0.15-CPU run), HAProxy `maxconn 32` (a little above 24) and
`timeout queue 500ms`, `retry-on conn-failure` with `option redispatch`. The lesson's "Try
changing" is step 5's rule in reverse: set the limit to 200 and p99 climbs toward the client
timeout while goodput falls. The limit is the latency budget expressed as concurrency.

**Step 4, stateful.** The nodes cannot be routed around: lesson 08's OOM-killed node took ~1/6 of
its tenants' requests with it at each crash, and lesson 12 measures a planned node restart at
3–8 s of refusals for that node's tenants.

**Step 6, self-inflicted peaks found by the lessons.** A restarted node reopened ~120 databases
under one lock and timed out every gateway (fixed with a singleflight, lesson 12). The first ledger
purge pass with 50 ms pauses competed with resolves and opened the gateways' breaker (pause raised
to 250 ms, lesson 10). Twenty dashboard refreshes of the usage report took control's whole pool
(lesson 06). Route-cache refresh is spread by refreshing on hit after 10 s rather than letting
every entry expire together (lesson 05).

**Step 7.** Lesson 10's verify samples at 1/30/60/90 s on 48 MB volumes. Before: WAL +13–14 MB/min
per node, the volume full in ~80 s, handles +6/min (`nofile` in 32 minutes), ledger +4 080
rows/min. After: WAL flat at ~3 MB, handles flat at 128, ledger shrinking; what still grows (0.5–0.8
MB/min) is tenant data, a capacity-plan input rather than a leak.
