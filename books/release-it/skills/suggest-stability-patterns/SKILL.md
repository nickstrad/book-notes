---
name: suggest-stability-patterns
description: Turn stability findings (or symptoms) into concrete remedies from Release It!'s stability patterns — timeouts, circuit breakers, bulkheads, shed load, back pressure, steady state, fail fast, governors — deciding for each where it goes, what its parameters are, what it does when it trips, and what signal it must expose. Use after hunt-stability-antipatterns, or when designing a service's integration layer.
---

# suggest-stability-patterns

A finding says what can break. This skill says what to build so that it breaks small. Each
pattern is a decision with four parts: **where** it sits (caller, callee, or the boundary between
them), **what numbers** it needs, **what happens when it trips** (the fallback), and **which
signal** tells operators it tripped. A pattern missing any of the four is not done.

## When to use

- After `hunt-stability-antipatterns`, with its findings table in hand.
- Designing a new service's clients, servers and background jobs.
- A postmortem whose cause is known and whose remedy is being chosen.

## Inputs

- Findings: antipattern, evidence, trigger, blast radius (the hunt's output), or a described
  symptom ("one tenant slows everyone", "it stays down after the partner recovers").
- The system map: tiers, stateful vs stateless, call chain with SLO, pool sizes.
- Which operations are idempotent, and what the business accepts as degraded behaviour (stale
  data? refusals? queued work?). If nobody has decided, that is the first open question.

## Procedure

1. **Order the work.** Remedies depend on each other. Use this order unless a finding is on fire:
   transparency (you must see each pattern work) → timeouts → resource release and bounded pools
   → retries with restraint → breakers with fallbacks → bulkheads → shed load → bounded results →
   back pressure → steady state → test harness → version handling.

2. **Map each finding to candidate patterns** (Ch. 5 unless noted):

   | Finding | First remedy | Usually paired with |
   |---|---|---|
   | Integration point with no limit | **Timeouts** (connect, response header, total; server read/write/idle) | Circuit Breaker; deadline propagation |
   | Work that outlives its caller | **Timeouts** on the work itself, with real cancellation | Fail Fast |
   | Resource released only on success | Release on every path (`defer`/`finally`); bounded pool; checkout timeout | **Fail Fast** (refuse when the pool is empty) |
   | Invalid work reaching expensive steps | **Fail Fast**: validate, then check resources, then work | — |
   | Retry storm / dogpile / self-denial | **Governor** on retries: idempotent only, jittered backoff, budget | Circuit Breaker; `budget-timeouts-and-retries` |
   | Everything depends on one upstream | **Circuit Breaker** + an explicit fallback (cache, default, queue, refuse) | Handshaking (`Retry-After`), health semantics |
   | One consumer exhausts a shared resource | **Bulkheads**: per-tenant / per-path / per-dependency partitions | Shed Load inside each partition |
   | Overload in a tier; chain reaction | **Shed Load** (bounded in-flight, 503 early) | **Handshaking** (503 + `Retry-After`, degraded health); Ch. 9 Demand Control at the balancer |
   | Front tier can out-offer the back tier | A limit on the back tier itself | Shed Load; capacity review |
   | Unbounded result | Cap and page/stream; aggregate where the data lives | **Let It Crash** only if restart is fast and clean |
   | Producer faster than consumer | **Create Back Pressure**: bounded queue + a full-queue policy (wait briefly / refuse / persist) | **Decoupling Middleware** when work must survive restarts |
   | Something only grows | **Steady State**: an automatic purge, rotation, checkpoint or eviction for every accumulator | Metrics that show the level |
   | Failure modes never exercised | **Test Harnesses** | `test-integration-points` |
   | Strict readers, in-place migrations, no drain | Ch. 13–14: tolerant reader, expand/contract, drain | `plan-safe-release` |

3. **Decide placement.** For each pattern, state where it lives and why there:
   - Timeouts and breakers sit at the **caller**, around the integration point they guard.
   - Shed load and per-tenant bulkheads sit at the **callee** (it protects itself from the
     combined offer of all callers), and at the **balancer** as a second line.
   - Back pressure sits **between** producer and consumer, owned by the producer's process.
   - Steady-state jobs sit **beside** the thing that accumulates, running without a human.
   - In a stateful tier, place protection so the shard survives: a node that sheds stays up for
     its tenants; a node that crashes takes them all.

4. **Set the parameters from measurements, not taste.**
   - Timeouts: inner < outer along the chain; response-header ≥ the longest legitimate work.
   - Pools and lanes: from the resource's real limit (fds, connections, CPU).
   - Shed limits: latency budget × per-instance throughput (Little's law), a little above the
     concurrency at calibrated capacity.
   - Breakers: consecutive-failure and failure-ratio thresholds, minimum calls, window, cool-down.
   - Mark every number "initial" and name the experiment that will tune it.

5. **Decide the fallback.** For every breaker, bulkhead and shed point: what does the caller get?
   A cached answer, a refusal with `Retry-After`, a queued write, a default. Name who decided (a
   product decision, not an engineering default). Which failures count toward tripping (5xx and
   transport errors, not a 404 that is the caller's fault).

6. **Name the signal.** Every pattern exposes its state: a gauge (breaker state, queue depth, in
   use), a counter of trips (`shed_total`, `breaker_rejected_total`, `retry_budget_exhausted_total`),
   and a health effect (`degraded` while tripping, never `failed` for a dependency the instance can
   work around).

7. **Check for new failure paths.** A timeout creates an error path: does it release everything?
   A retry adds load: is it budgeted? A breaker's fallback serves stale data: what makes stale data
   safe (a version or epoch check)? A shed limit refuses: does the client honour `Retry-After`?

## Output

One row per finding, in the order of step 1:

```
| Finding | Pattern(s) | Placement | Parameters (initial) | When tripped, the caller gets | Signal / health effect | Verify by |
```

Then the open product decisions (fallback behaviour nobody has chosen yet).

## Source

*Release It!* 2nd ed.: Ch. 5 "Stability Patterns" (Timeouts, Circuit Breaker, Bulkheads, Steady
State, Fail Fast, Let It Crash, Handshaking, Test Harnesses, Decoupling Middleware, Shed Load,
Create Back Pressure, Governor); Ch. 9 "Demand Control"; Ch. 13–14 for the version remedies.
Demonstrated by lessons 02–12 in `../../lessons/`.

## Worked example: litefleet's register → lessons 02–12

Run on the hunt's output for `release-it/v0-flawed`, this skill should reproduce the lesson
sequence. The step-1 order *is* the lesson order (transparency was moved first, lesson 01, because
every later pattern is measured through it). The output, condensed from the lesson notes:

| Finding | Pattern | Placement | Parameters as built | Caller gets | Signal |
|---|---|---|---|---|---|
| F02–F04 | Timeouts | every client (`internal/httpx/client.go`), every statement, every server, HAProxy | 2.9 s request ⊃ 2.5 s node attempt ⊃ 2 s statement; control 500 ms | `504 timeout` | `statements_cancelled_total{reason}` |
| F05–F07 | release on every path; bounded pools; Fail Fast | node engine (`defer tx.Rollback()`), handle budget 150, 4 conns/db, `PoolWait` 250 ms | from `nofile`: (1024−124)/(2×3) = 150 | `409`, `503 pool_timeout`, `503 node_full` | `sqlite_pool_timeouts_total` |
| F08 | Governor on retries | gateway node client and control client (`internal/retry`) | reads only, full jitter 50–400 ms, 2 retries, budget 10 % | one attempt for writes | `retry_budget_exhausted_total` |
| F09 | Circuit Breaker + **route cache as fallback** | gateway, around control (`internal/breaker`, `internal/routes/cache.go`) | 5 consecutive or ≥50 % of ≥20 in 10 s; 5 s cool-down; TTL 60 s | cached route, or `503 control_unavailable` + `Retry-After: 5` | `breaker_state`, health `degraded: serving N cached routes` |
| F10 | **Per-tenant bulkhead** + separate pools | node admission (`internal/engine/admission.go`); control pools `oltp`/`reporting`/`health` | 2 reads or 1 write per db, 4 per tenant, 16 per node; pools 8/2/1 | `429 tenant_busy` + `Retry-After` | `admission_rejected_total{reason}` |
| F11 | Bulkhead in time: lock only the copy | backup job | lock covers `VACUUM INTO` only | — | `writer_lock_hold_seconds` |
| F12 | Shed Load + Handshaking + balancer limits | gateway and node (`internal/shed`), HAProxy | gateway 24, node 48; `maxconn 32`, `timeout queue 500ms` | `503 overloaded` + `Retry-After: 1` | `shed_total`, health `degraded: shedding load` |
| F13 | cap + stream; aggregate in SQL | node, gateway, control report | 1 000 rows or 1 MB; `LIMIT 50` report | `"truncated": true` | — |
| F14 | Back Pressure + outbox | node metering queue | 1 024 events, 20 ms enqueue wait, 4 shippers | (statement delayed ≤ 20 ms) | `usage_queue_depth`, `usage_outbox_rows` |
| F15 | Steady State | node checkpointer + LRU; `control purge` | WAL > 64 KB truncated every 2 s; LRU 128; ledger 25 h | — | `sqlite_checkpoints_total` |
| F16 | Test Harness | `test/chaos`, blobstore `/admin/mode` | 12 tests | classified errors | test matrix |
| F17–F18 | tolerant reader; expand/contract; drain | gateway, consumer, migration job, node/gateway SIGTERM | `lock_timeout 3s` | `503 draining` + `Retry-After: 2` | `wire_version` in heartbeat |

Three things the worked example teaches that a pattern catalogue does not:

- **Placement is the decision.** The per-tenant bulkhead (lesson 06) had to be per tenant, not per
  database: with per-database lanes alone a tenant with 4 databases still held the whole node.
  Other tenants' p99 under the noisy tenant: 534 ms before, 87 ms after.
- **The fallback is the value.** A breaker alone (lesson 05, part 1) only fails faster. With the
  route cache as its fallback, 99.25 % of known-database requests were served through a 30 s
  control outage (0 % before), and the gateway health had to change to `degraded` to match: a
  health rule is part of the fallback.
- **Parameters move after measurement.** Lesson 03's control pool went 10 → 20; lesson 06 cut it
  to 8 once the route cache took load off it; lesson 10's handle cache went 64 → 128 because the
  working set was ~123 databases per node. Each change is recorded in its note's "Deviations".
