---
name: hunt-stability-antipatterns
description: Walk a codebase or design for the twelve stability antipatterns of Release It! (unbounded waits, leaked resources, retry storms, unbounded results and queues, shared pools, capacity mismatches) and report each finding with evidence, trigger, blast radius and severity. Use for design or code review, before a launch, or in a postmortem asking "why did this fall over".
---

# hunt-stability-antipatterns

Find the cracks that turn into outages before production finds them. An antipattern is a shape in
the code or design that lets a small fault (a slow partner, a burst, one bad row) travel through the
system and become an outage. This skill walks a fixed list of them with concrete signals, and for
every hit says what triggers it and where the strain goes.

## When to use

- A design or code review of a service that calls other services or databases.
- Before a launch, a migration, or a large traffic event.
- A postmortem: "it fell over and we don't know why", "it stayed down after the fault went away".
- After `review-operational-transparency` shows you can see the system, and before
  `suggest-stability-patterns` decides what to change.

## Inputs

- The code (a repo or a tag), or a design doc with a topology diagram.
- The deployment description: instance counts, CPU/memory limits, `nofile`, pool sizes, balancer
  config (`compose.yml`, Kubernetes manifests, HAProxy/nginx config).
- If available: a load profile (request mix and rate) and the latency SLO at the edge.

## Procedure

1. **Map the system before hunting.** List every tier, its instance count, and whether it is
   stateless (can be routed around) or stateful (its users are pinned to it). List every
   integration point: each outbound HTTP/gRPC client, database driver, cache, queue, blob store,
   and each inbound server. List every shared resource: connection pools, worker pools, locks,
   file descriptors, disk volumes, memory limits. Write the call chain from the edge inward with
   each hop's timeout, if one exists. Everything below is judged against this map.

2. **Walk the antipatterns.** For each one, run its signals; record every hit with `file:line`.

   | Antipattern (Ch. 4 unless noted) | Signals to look for |
   |---|---|
   | **Integration Points** | Clients built with no timeout (Go: `&http.Client{}`, `http.DefaultClient`, `http.NewRequest(` instead of `NewRequestWithContext`; Java: no `setConnectTimeout`/`setReadTimeout`); servers with no read/write/idle timeouts (`http.Server{` without `ReadHeaderTimeout`); balancer `timeout server` in minutes; error bodies read whole (`io.ReadAll(resp.Body)`) or decoded before the status is checked; no test where the partner is slow, resets, or sends garbage. |
   | **Blocked Threads** | A resource released only on the success path (a `return err` between `Begin()` and `Rollback()`/`Close()`; no `defer`/`finally`); pools with no maximum (`SetMaxOpenConns` absent) or no checkout timeout; locks held across a network call or a slow I/O; `context.Background()` inside a request path; mutexes around remote calls. |
   | **Slow Responses** | Work with no deadline that can outlive its caller (statements, scans, recursive queries); a server that keeps computing after the client hung up (no cancellation hook such as `sqlite3_interrupt`, `pg_cancel_backend`, context propagation). |
   | **Unbounded Result Sets** | Queries without `LIMIT` on tables that grow, or on tenant-supplied SQL; results built whole in memory (`append` every row, then marshal); a proxy that buffers a downstream body (`io.ReadAll`) before forwarding; aggregation done in application code over raw rows. |
   | **Dogpile / Self-Denial / Force Multiplier** | Retry loops with a fixed sleep (`time.Sleep(100 * time.Millisecond)` in a loop), no jitter, no budget; retries on non-idempotent calls (POST/execute/write) or after a timeout; tight spin loops on `BUSY`/`locked`; caches that all expire at once; synchronized cron; automation that acts fleet-wide with no rate limit. |
   | **Cascading Failures** | Every request of the data path depends on one upstream (a resolve, an auth call, a config fetch) with no breaker and no cached fallback; a health check that reports the upstream's state as the caller's own `failed`. |
   | **Chain Reactions** | A tier of identical instances with no in-flight limit and no 503 path: losing one instance moves its load onto the rest. Check N−1: at today's peak, can the survivors carry it? |
   | **Unbalanced Capacities** | Front tier's total offer (instances × per-instance capacity) exceeds the back tier's; a back tier with no limit of its own; a producer faster than its consumer with an unbounded buffer between them (slice/list/channel with no bound used as a queue). |
   | **Scaling Effects / Users** | Point-to-point fan-out that grows with instance count; per-user or per-tenant metric labels; expensive pages or reports anyone can trigger repeatedly. |
   | **Shared resource across tenants or paths** (Ch. 5 Bulkheads, as a hunt signal) | One pool serving the critical path, a heavy report, and the health check; no per-tenant or per-key concurrency limit on a multitenant node; a maintenance job (backup, vacuum, reindex) that holds a lock users need while it waits on something slow. |
   | **Accumulation** (Ch. 5 Steady State, as a hunt signal) | Anything that only grows: log or WAL files with no rotation/checkpoint (`wal_autocheckpoint=0` and no checkpointer), tables appended forever with no purge, caches or handle maps with no eviction, `TODO: schedule` comments next to disabled housekeeping. |
   | **Version hazards** (Ch. 12–14, as a hunt signal) | Strict readers (`DisallowUnknownFields`, exact-schema validation) on data another team or release produces; migrations run in `main()` at startup; in-place `ALTER ... TYPE` on large tables; 32-bit counters that grow with traffic; SIGTERM handled as immediate exit (no drain). |

3. **Trace each hit's chain of failure.** For every finding ask the Ch. 3 question: when this part
   fails, where does the strain go? Name the trigger (what load or fault exposes it), the first
   resource exhausted (goroutines, fds, connections, memory, a lock, disk), and who else shares
   that resource. A hit with no plausible trigger is noise; drop it or mark it low.

4. **Look for compounds.** Antipatterns multiply. Check every pair where one finding feeds another:
   retries on top of a leak (each retry leaks again), timeouts that create new error paths into an
   unreleased resource, a stateful tier behind a front tier with no limits. Report compounds as
   their own findings.

5. **Rate blast radius and severity.**
   - Blast radius: one request / one tenant or key / one instance / one stateful shard (everyone
     pinned to it) / the whole fleet.
   - Severity: **critical** = fleet-wide or a stateful shard down, or data damaged (duplicate or
     lost writes); **high** = one tier degraded for everyone or does not recover by itself after
     the fault ends; **medium** = one tenant/path degraded, or waste that grows (memory, disk) on a
     dated horizon; **low** = inefficiency without an outage path.
   - "Stays broken after the fault is removed" always raises severity one step.

6. **State how to demonstrate it.** For each critical or high finding, write the one-line
   experiment that would show it (load mix + rate + injected fault + expected symptom). That line
   is the input to `design-chaos-experiment`.

## Output

One table, most severe first, then a list of compounds:

```
| # | Antipattern | Evidence (file:line) | Trigger | Strain goes to | Blast radius | Severity | Demonstrate with |
```

Close with the count per antipattern and the antipatterns you checked and found nothing for (so a
reader knows they were checked, not skipped).

## Source

*Release It!* 2nd ed. (Nygard): Ch. 3 "Stabilize Your System" (impulse, stress, strain, chain of
failure); Ch. 4 "Stability Antipatterns" (all twelve sections); Ch. 2 case study (the leaked
connection); Ch. 5 "Bulkheads" and "Steady State" for the shared-resource and accumulation
signals; Ch. 12 case study and Ch. 14 "Handling Versions" for the version hazards. Demonstrated in
this repo by every lesson 01–12; the register in `../../PLAN.md` section 3 and
`../../lessons/00-the-flawed-start.md` is the answer key.

## Worked example: litefleet at `release-it/v0-flawed`

Run the procedure on `git checkout release-it/v0-flawed -- books/release-it/project` (or a
worktree at that tag). The register has 18 flaws, F01–F18. **A run that finds fewer than 13 of
them means the skill run needs work.** The signals in step 2 land as follows (line numbers at the
tag, from `lessons/00-the-flawed-start.md`):

| Signal hit | Where at v0 | Register |
|---|---|---|
| constant `/health`, no `/metrics`, `handled POST /path` logs | `internal/httpx/httpx.go:25`, `:73` | F01 |
| `&http.Client{}` on all four clients | `internal/routes/client.go:26`, `internal/nodeclient/client.go:25`, `internal/blob/client.go:21`, `internal/metering/client.go:22` | F02 |
| statements on `context.Background()` | `internal/engine/exec.go:72`, `:123` | F03 |
| `http.Server` with no timeouts; HAProxy `timeout server 5m` | `internal/httpx/httpx.go:18`; `deploy/haproxy.cfg:12–13` | F04 |
| `return` before `tx.Rollback()` on statement error | `internal/engine/exec.go:133` | F05 |
| no `SetMaxOpenConns`; unbounded Postgres pool | `internal/engine/handles.go:85`; `internal/registry/registry.go:32–37` | F06 |
| resolve and file open before validation | `internal/api/handlers.go:59–72`; `internal/engine/exec.go:60–68` | F07 |
| 5 retries at fixed 100 ms including `/execute`; 50-try `BUSY` spin | `internal/nodeclient/client.go:48–71`; `internal/engine/exec.go:156–176` | F08 |
| every request resolves through control, no breaker or cache | `internal/routes/client.go:30–33` | F09 |
| no per-tenant admission; report on control's only pool | `internal/engine/admission.go:13`; `cmd/control/usage.go:14` | F10 |
| backup holds the write lock across the upload | `internal/backup/job.go:66–89` | F11 |
| no in-flight limit, no `maxconn`, no 503 path | `cmd/gateway/main.go`, `cmd/node/main.go`, `deploy/haproxy.cfg:25–31` | F12 |
| every row built in memory; gateway `io.ReadAll`; report aggregates in Go | `internal/engine/exec.go:96–101`; `internal/nodeclient/client.go:87`; `cmd/control/usage.go:30–48` | F13 |
| unbounded slice as the usage queue, never persisted | `internal/metering/queue.go:18–29` | F14 |
| `wal_autocheckpoint(0)`, ledger never purged, handles never closed | `internal/engine/handles.go:71`, `:94`; `internal/registry/usage.go:33` | F15 |
| happy-path tests only; decode regardless of status | `internal/nodeclient/client_test.go`; `internal/blob/client.go:52–61` | F16 |
| `DisallowUnknownFields` in gateway and consumer | `internal/nodeclient/client.go:105–110`; `cmd/loadgen/main.go:449–454` | F17 |
| migrations in `main()`; `int` counters; SIGTERM = exit | `cmd/control/main.go:31–39`; `migrations/001_init.sql:29–30`; `cmd/node/main.go:76` | F18 |

The ones a hunt most often misses, and the step that catches them:

- **F11** looks like correct locking. Only the "maintenance job holds a lock while it waits on
  something slow" signal finds it. Measured at `lesson-05`: with the blob store at 50 KB/s, all
  250 writes to the database being backed up failed (p99 2 015 ms, the busy timeout).
- **F05 + F06 + F08** is the compound of step 4. Each retry of a `database is locked` 500 leaks
  another connection. Measured at `lesson-02`: both nodes reached their 1 024-fd wall within 15 s,
  14 750 node 5xx, 21.89 % of valid writes succeeded, and node-a stayed down after the load
  stopped. Severity: critical (a stateful shard, does not recover).
- **F13** reads as "a slow query" until you note the tier is stateful. Measured at `lesson-07`: ten
  concurrent whale scans OOM-killed node-a 3–5 times in 45 s, and node-a's other tenants lost about
  a sixth of their requests. Blast radius: the shard, not the request.
- **F18's `int` counters** are only found by the version-hazard signal; they fail on a date, not
  under load.

The lesson tags are the verification: each critical finding's "demonstrate with" line should
resemble the timeline in the matching `project/lab/verify/NN-*.sh`, and
`make verify LESSON=NN REF=<parent tag>` prints `FLAW REPRODUCED:` for it.
