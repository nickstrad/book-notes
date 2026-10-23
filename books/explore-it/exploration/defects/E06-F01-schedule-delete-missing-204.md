# E06-F01 — DELETE schedule answers 204 for a missing id and for another tenant's schedule

Seen at: export of `explore-it/lesson-05` (stack `queuelet-e06`), experiment 06 (relate-it).
Original observation: [E06 notes](../experiments/06-relate-it/notes.md) P2, 15:32:15Z (H-E06-03, predicted from code).

## Minimal reproduction

Actor: acme key (and globex key for the cross-tenant variant). Clock: real. Fixture: seeded stack (schedule 1 belongs to acme).

```sh
source exploration/experiments/06-relate-it/tools/env.sh
k acme-key   DELETE /v1/schedules/99999        # [204]
k acme-key   GET    /v1/schedules/99999        # schedule_not_found [404]
k acme-key   PATCH  /v1/schedules/99999 '{"enabled":false}'   # [404]
k acme-key   POST   /v1/schedules/99999/run    # [404]
k globex-key DELETE /v1/schedules/1            # [204]  — acme's schedule, untouched:
sql "select id,namespace_id,name from schedules where id=1"   # 1 | 1 | daily-report
k acme-key   DELETE /v1/schedules/3            # [204] again for an id deleted earlier
```
Replay: `tools/p2-schedules.sh` (first three blocks).

## Expected (named oracle)

SPEC:15 "Deletion returns 204 … A missing resource returns 404." The sibling verbs on the same path (GET, PATCH, run)
and the other delete endpoints (`DELETE /v1/queues/{name}` → 404 queue_not_found; `DELETE /v1/jobs/{id}` → 404 via
`ownJob`; `DELETE /v1/workers/{id}` → 404 worker_not_found) all follow it. For another tenant's id the namespace rule
(SPEC:10) makes the schedule invisible, so 404 is also the expected answer there.

## Observed

`204` with an empty body for (a) an id that never existed, (b) an id already deleted, (c) another namespace's live
schedule (which correctly survives). GET/PATCH/run of the same ids → 404 `schedule_not_found`.

## Variables isolated

- Matters: only the DELETE handler (`internal/api/schedules.go:136-143` executes the DELETE and never checks
  RowsAffected).
- Does not matter: whether the id ever existed, which tenant owns it, schedule state.
- Isolation holds: no cross-tenant deletion occurred (the WHERE includes `namespace_id`); the defect is the answer, not
  the effect.

## Sibling search

Other DELETE handlers: `queues.go:109-151` (404 checked), `api/jobs.go:28-38` (ownJob 404 then 409 via `jobs.Changed`),
`api/jobs.go:114-124` (404 via `jobs.Changed`). Pause/resume (`queues.go:165-167`) run after `queueOrError`, so 404
holds. None found beyond the schedule DELETE. This is the schedule analogue of the inherited queue-delete 204 (F15 /
D02, fixed in lesson 01 for queues only).

## Severity and status

Severity: **low**. A client cannot tell "deleted" from "never existed / not yours"; idempotent-retry clients are not
harmed, but scripts that verify cleanup or tenants probing ids get a misleading success. Status: **open**.
