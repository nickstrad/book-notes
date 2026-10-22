# E08 surround-it — revealed

Status: **done** (one run). Checkpoint `explore-it/lesson-07`, stack `queuelet-e08` (stopped at 19:38:33Z).
Record: [notes](notes.md); tools in [tools/](tools/); raw output in `books/explore-it/project/results/e08/`.

## Elapsed

Process timestamps: run start 19:18:14Z (`run-start.txt`), first command 19:18:20Z, stack stopped 19:38:33Z,
notes closed 19:39:50Z. Total ≈ 22 minutes of the 60-minute budget.

| Phase | Wall clock | Approx. |
|---|---|---|
| Reading (STATE, charter, protocol, skill, lesson 07, source; stack built in parallel) | 19:18:20–19:21:10 | 3 min |
| Pre-probe writing (diagram, 13 hypotheses) | 19:21:10–19:22:05 | 1 min |
| Probing (P1–P8b, with predictions and log entries interleaved) | 19:22:08–19:35:50 | 14 min |
| Writing (reports, models, backlog, this file, STATE) | 19:35:50–19:39:50 | 4 min |

## What the system showed

- **Tenant routes are scoped.** The other tenant's key gets 404 on every job, worker and schedule route and 403 on
  foreign namespace stats; nothing changed (D21 holds). Events answer `[]` for any id that isn't yours, nonexistent
  ones included. That gives away nothing, and it follows from events outliving their jobs. The admin key is an acme key
  that can also read stats anywhere. Within a tenant, the worker token is never checked, so any key holder can lease
  as any worker (spec silence, Q-E08-1).
- **Postgres: the running API survives, a starting one does not.** `compose restart postgres` costs ≈1 s of 503 and
  the loop carries on (D22/U04 hold). An API start while the database is unreachable `log.Fatal`s, and with no
  restart policy the service stays down after the database returns (E08-F01). An API or DB crash also stays down.
- **A frozen database is invisible.** With Postgres paused, every request, `/health` included, hangs for the whole
  freeze and then succeeds. The 1 s health ping deadline does not bound it, and no 503 is ever returned (E08-F04).
  The ping result contradicted my code reading.
- **The webhook shipper is one global queue.** It is separated from the core loop, so lifecycle progress is safe (U03
  holds). Inside it, though, 20 rows whose destination is slow or unreachable take every pass. Later rows from every
  tenant are never attempted, and a tenant can cause this through configuration alone. No API action ends it, and
  `/health` stays 200 (E08-F02). 500s, 3 s slowness and fewer than 20 slow rows only cause delay. That matches the
  documented at-least-once behaviour.
- **Following one webhook body:** ack → jsonb row → Go re-marshal (key reorder, compaction, `<&>`,
  ` `) → partner. The values match, and I verified the HMAC by hand over the received bytes (D24 holds;
  transformations as already seen in E06).
- **Clock × DST:** hourly and 09:00 schedules follow the offset change. Fixed-time schedules inside the spring-forward
  gap do not run that day, and those inside the fall-back hour run twice (E08-F03, spec silence).

## Model delta

- [models/ecosystem.md](../../models/ecosystem.md) E08 revision: context diagram and per-boundary table (what
  crosses, where checked, what-if, observed, recovery).
- [models/rules.md](../../models/rules.md) E08 revision: R-E08-1…6.
- [hypotheses.md](../../hypotheses.md) E08 section: H-E08-C, H-E08-01…24, Q-E08-1/2, T-E08-1…3.
- Charter prediction "every route is namespace-scoped and the loop survives a database restart": **confirmed** for
  routes and for a restart of a running API. **Refuted per neighbour** at three boundaries the sentence didn't cover:
  DB absent at start (E08-F01), DB frozen (E08-F04), and the partner/tenant boundary inside the shipper (E08-F02).

## Surprises and their dispositions

| Surprise | Disposition |
|---|---|
| API stays down after a start without DB | [E08-F01](../../defects/E08-F01-api-does-not-recover-from-start-without-database.md) |
| Later notifications never attempted behind 20 slow rows | [E08-F02](../../defects/E08-F02-slow-destination-starves-all-later-notifications.md) |
| DST gap skip / overlap double fire | [E08-F03](../../defects/E08-F03-dst-gap-skips-and-overlap-double-fires.md) (status `spec`) |
| `/health` hangs and then says ok during a DB freeze | [E08-F04](../../defects/E08-F04-frozen-database-hangs-health-and-api.md) |
| Events 200 `[]` for foreign/nonexistent ids | not a bug: ownership-scoped durable events, nothing crosses (notes P1) |
| Foreign worker DELETE message "state conflict" under 404 `worker_not_found` | not a bug: cosmetic message, correct status/code (notes P1) |
| Lease as another worker by id | not a bug under the current spec: bearer key is the only defined boundary; Q-E08-1 |
| 3 partner side effects for 1 notification in slow mode | documented at-least-once (SPEC.md:200-204, E02-F03) |

## Unresolved and next hypotheses

- H-E08-22 a redirecting partner (302 → GET 200) counted as delivered without the body (needs T-E08-3).
- H-E08-23 one schedule that cannot advance starves later schedules (fire.go:37-40); no trigger found.
- H-E08-24 pool exhaustion (16 connections) during a DB freeze.
- H-E04-10 / T-E08-1 ack non-atomicity under a single-statement DB fault (still needs a hook).
- E08-F04's mechanism (lib/pq cancellation) is an inference. The driver source was not in the export.
- The workers' own timeouts against a frozen API were not read or probed.
- Process notes: the P3b sibling predictions were written into the log after the run, so they are excluded from
  the score. In P6 and P7 the predictions were appended in the same command as the run, immediately before it, and
  the timestamps in those headings were corrected to say so. P3's first attempt failed on zsh word-splitting and
  took no action (kept in the notes).

## Score

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 21 | H-E08-01…21 stated before their probes (table 19:21Z; H-14 19:25:45Z; H-15…17 19:29Z; H-18 19:30:45Z; H-19…21 before P7/P8/P8b). Charter sentence H-E08-C counted separately; H-E08-22…24 were stated but not probed and are left open; P3b post-hoc excluded |
| Hypotheses confirmed | 20 | all except H-E08-20 (notes P1–P8b). Six of them are regression holds, not discoveries: H-01 (D21), H-04 (D22/U04), H-06 (D04), H-09 (D24), H-13 (D23), plus H-07 (E02-F03 semantics) |
| Hypotheses refuted | 1 | H-E08-20: `/health` hung 11.8 s and then returned 200 instead of 503 within ≈1 s (P8) |
| Code-reading predictions contradicted by the running system | 1 | H-E08-20, from the 1 s `PingContext` at `internal/api/api.go:132` |
| Defects found | 4 | E08-F01, E08-F02, E08-F03 (spec silence, status `spec`), E08-F04 |
| Model cells filled | 7 | new ★ ecosystem cells with E08 evidence: admin-key scope; worker identity within a tenant; DB absent at API start; DB/API crash supervision; DB frozen; slow/unreachable destination × other rows/tenants; clock × DST gap/overlap. Re-observed cells (tenant routes, DB restart, partner 500/slow, data-journey transformations from E06) are not counted |
