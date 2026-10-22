---
name: review-operational-transparency
description: Review a service's operations surface — health-check semantics (ok / degraded / failed and what each checks), metrics (traffic, errors, latency per dependency, pool/queue/breaker/shed state, physical limits, label cardinality), structured logs with propagated request ids, and version exposure — ending with a test that a slow dependency is visible before users see errors. Use before a service's first deploy, after "we could not tell what was wrong", or when adding a stability pattern that must report its state.
---

# review-operational-transparency

You cannot fix what you cannot see. A transparent instance tells an operator, from the outside,
whether it can do its work right now, what it is waiting on, and how close it is to a limit. This
skill reviews that surface as a checklist, then proves it with one injected fault.

## When to use

- A new service before its first production deploy.
- A postmortem with "we couldn't tell what was wrong" or "everything showed green".
- Adding a breaker, bulkhead, shed limit or queue: each must report its state.
- Changing a health check, or a balancer's use of it.

## Inputs

- The service's code for `/health` (or equivalent), metrics registration, logging setup, and its
  outbound clients.
- The balancer/orchestrator config that consumes health (what does `failed` cause?).
- Access to the metrics system and logs during a test run, and a way to inject latency on one
  dependency.

## Procedure

1. **Health semantics.** For each service, write down what `/health` checks and what it returns:
   - Three states, not two: `ok`; `degraded` (200: keep sending me traffic, but something is
     wrong); `failed` (503: take me out). The body names each check's state and reason.
   - `failed` only for what makes **this instance** unable to work: its own database unreachable,
     its disk full, its process unable to open what it serves. Report a dependency the instance can
     work around (a down peer, a cached upstream, shedding) as `degraded`.
   - Ask of every `failed` rule: if this fires on every instance at once, does the balancer take
     out the whole tier for a fault in one dependency? A health check can cause the cascade it
     reports.
   - The check uses the real path (the same handle path or a dedicated one) but must not starve:
     it gets its own small pool/connection, and a short timeout.
   - Check physical facts, not just liveness: open fds vs `nofile`, disk free, log/WAL size.
   - The balancer actually uses it (HTTP check, not a TCP connect).

2. **Metrics.** Check each item exists, has a sensible name, and is exported (scraped targets show
   `up = 1`):
   - Per route: requests by status, latency histogram, in-flight.
   - **Per dependency**: calls by outcome and a latency histogram. This is the metric that names
     the slow dependency.
   - State of every stability mechanism: pool in use / idle / waiters / wait timeouts; queue depth
     and oldest-item age; breaker state and transitions; shed and admission rejections; retry
     attempts and budget exhaustion; cancellations by reason.
   - Physical limits: open fds, goroutines/threads, memory, disk, WAL or log size, open handles.
   - Business counters where cheap (statements run, events delivered).
   - **Cardinality:** no unbounded label (tenant, user, database, URL path with ids). Labels are
     routes, dependencies, outcomes, reasons. Per-tenant questions go to logs and usage data.

3. **Logs.**
   - Structured (JSON), one schema, every line including the libraries' own.
   - A request id generated at the edge or accepted from the caller, returned as a response
     header, and **propagated on every outbound call**, so one id finds every hop.
   - One line per request with status and duration; one line per dependency call with dependency,
     duration and outcome (sample or log only slow/failed calls for very frequent ones; keep them
     all in metrics).
   - Tenant/user ids allowed in logs (not in metric labels). No secrets, no full bodies.
   - Levels used consistently; rotation or shipping so logs cannot fill the disk.

4. **Configuration and version exposure.** The running version (build, protocol/wire version,
   schema version) is visible per instance: in a metric, the health body, or a heartbeat to the
   control plane. During a mixed-version rollout you need to know which instance runs what.

5. **The test: is a slow dependency visible before users see errors?** Under steady load, add
   latency (below every timeout) to one dependency of one instance. Then answer from the
   operations surface alone, without reading code:
   - Which dependency is slow, and on which instances? (per-dependency p99)
   - Can you follow one slow request across services? (request id in logs)
   Then make a dependency fail outright and check: how many seconds until health reports it, and
   does the balancer react the way you intended (degraded instances stay in; failed ones leave)?

## Output

A checklist table with one row per item above: `item | present? | evidence (file:line or metric
name) | gap`. Then the step-5 results: `fault | visible as | seconds to visible | balancer
reaction`. Then the gaps, ordered by "would have changed the last incident".

## Source

*Release It!* 2nd ed.: Ch. 8 "Transparency" (designing for transparency, logging, instance metrics,
health checks); Ch. 10 "System-Wide Transparency";
Ch. 6 case study (a busy season diagnosed with few vital signs); Ch. 5 "Steady State" (log files);
Ch. 5 "Handshaking" (health as the instance's way of saying "not now"). Demonstrated in lesson 01
(`../../lessons/01-transparency.md`), refined in lessons 05, 06, 07 and 12.

## Worked example: litefleet

At `release-it/v0-flawed` (F01) the review is short: `/health` is a constant `200 ok`
(`internal/httpx/httpx.go:73`), no service exports `/metrics` (Prometheus shows `up = 0` for every
platform job), the only log line is `handled POST /path`, and HAProxy checks gateways with a TCP
connect. Step 5 at v0: node-a gets 800 ms of latency and nothing names it; Postgres is paused and
every request fails while HAProxy shows 3 of 3 gateways UP.

Lesson 01's patch is this skill's reference answer:

- **Health** (`internal/httpx/health.go`, `{status, checks}`): control is `failed` when a ping to
  Postgres takes > 500 ms; a node is `failed` when its canary cannot open or its volume is full,
  `degraded` above 80 % of `nofile` or 128 MB of WAL; a gateway is `degraded` when a node is
  unreachable (a down node costs its tenants, not the gateway). HAProxy switches to
  `option httpchk GET /health`.
- **Metrics** (`internal/httpx/metrics.go`): `http_requests_total`,
  `http_request_duration_seconds`, `http_inflight_requests`,
  `dependency_requests_total{dep,outcome}`, `dependency_duration_seconds{dep}` for control,
  node-a, node-b, blobstore, metering and postgres, `db_pool_*`, `sqlite_open_databases`,
  `sqlite_wal_bytes_total`, `usage_queue_depth`, `usage_queue_oldest_age_seconds`,
  `process_open_fds`. **No tenant or database label**: ~240 databases would be ~240 series per metric
  per instance.
- **Logs**: JSON through `log/slog`; `X-Request-Id` accepted or generated, returned, and
  propagated by every client; one line per dependency call (`dep`, `dur_ms`, `outcome`); registry
  queries logged only when slow (> 100 ms) or failed.

Step 5, measured (`make verify LESSON=01`, before → after):

| | v0 | lesson 01 |
|---|---|---|
| gateway → node-a p99 while node-a is slow | not measurable | 1.019 s (node-b 0.109 s) |
| control `/health` failed after Postgres paused | never | 0.82 s |
| gateway-a `/health` failed after Postgres paused | never | 1.95 s |
| HAProxy gateways UP during the pause | 3 of 3 | 0 of 3 |
| JSON log lines / request lines with an id | 0 % / — | 100 % / 100 % |

A slow node-a call is found by its log line (`{"request_id":"02cde18703fd127e","dep":"node-a",
"dur_ms":842.859,"outcome":"ok"}`) and the same id finds its gateway and node lines.

The later lessons are step 1's warnings in practice:

- **A `failed` rule that cascades.** In lesson 01, the gateway is `failed: control unreachable`,
  which was true then: every request needed control. Once lesson 05 added the route cache, that
  rule would have taken every gateway out during a control outage while the cache could serve.
  The rule became `degraded: serving N cached routes`. Result: 3 of 3 gateways stayed UP and
  99.25 % of known-database requests were served (0 % before).
- **The health check starving.** Through lesson 05, control's health pinged through the shared
  pool. Twenty usage reports made it fail 13 of 55 checks, so a busy pool looked like a dead
  database. Lesson 06 gave health its own one-connection pool: 0 of 96 checks failed.
- **Mechanism state**: `breaker_state` (05), `admission_rejected_total{reason}` (06), `shed_total`
  and health `degraded: shedding load` (07), `usage_outbox_rows` (09), `sqlite_checkpoints_total`
  (10).
- **Version exposure**: nodes report `wire_version` in their heartbeat (lesson 12), and the contract
  migration reads it to refuse while any node spoke v1 in the last day.
