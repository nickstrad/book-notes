# E06-F07 — Attempt rows (with error text) outlive their job after purge, DELETE and retention; nothing ever removes them

Seen at: export of `explore-it/lesson-05` (stack `queuelet-e06`), experiment 06 (relate-it).
Original observation: [E06 notes](../experiments/06-relate-it/notes.md) P10, 15:41:24Z (H-E06-11, H-E06-18; this is the
charter's "deleting an owner leaves orphans" prediction, confirmed for job → attempt only).

## Minimal reproduction

Actor: acme key, worker 4 (`manual`). Clock: real. Queue `e6-res` (max 2, linear backoff base 0).

```sh
source exploration/experiments/06-relate-it/tools/env.sh
bash exploration/experiments/06-relate-it/tools/p10-residue.sh
```
1. purge: job 885 nacked once ("secret-ish detail", delay 3600 → retry_wait), `POST …/purge` → `{"deleted":1}`.
2. DELETE: job 886 nacked twice → dead; `DELETE /v1/jobs/886` → 204.
3. retention: job 887 acked; PATCH retention_hours 0 → removed by the sweep.

## Expected (named oracle)

Approximation, stated with its bounds. SPEC:122 "Terminal jobs are deleted after `retention_hours`"; SPEC:80-81 attempts
are part of a job's record (`GET /v1/jobs/{id}/attempts`); SPEC:49-50 states the ownership principle "owned … must not
become orphaned" for queues. Read together: when a job is removed, its owned attempt records should go with it (or a
documented retention should apply, as README does explicitly for `job_events`: "retained until lab reset … for
exploration"). The SPEC does not name attempts in the retention rule, so this is a consistency/retention oracle, not a
spec line; the owner may decide otherwise (→ `spec`).

## Observed

```text
GET /v1/jobs/885 and /attempts → job_not_found [404]
psql attempts: 885 | 1 | failed | secret-ish detail
               886 | 1 | failed | fail 1 ; 886 | 2 | failed | fail 2
               887 | 1 | succeeded
orphan_attempts | jobs = 4 | 3    (no attempts table cleanup anywhere in code)
```
Events record `purged`, `deleted`, `retained` (by design). Delivered outbox rows for missing jobs: 3 (SPEC keeps
*pending* deliveries; delivered rows are likewise never removed — sibling).

## Variables isolated

- Matters: any job removal path — `queues.go:170` (purge), `api/jobs.go:32` (DELETE), `expire.go:90-92` (Retain).
- Does not matter: state at removal, outcome kind. Job ids are bigserial and were not reused (887 is a new id for the
  freed idempotency key K1), so the rows are unreachable through the API but never collide.

## Sibling search

- `webhook_outbox`: delivered rows never deleted (no DELETE on the table in code); pending rows must survive (SPEC).
- `workers`: rows for dead incarnations persist (E02 entities) — not job-owned.
- `job_events`: retained by README design (not a defect).
- No FK `attempts.job_id → jobs` (001.sql:7), so the database does not help.

## Severity and status

Severity: **low**. Unbounded growth of `attempts` and indefinite retention of worker error text after the job's
retention period; invisible to API clients. Status: **open** (candidate for `spec` if the owner decides attempts are
history like `job_events`).
