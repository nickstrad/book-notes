# F15 — DELETE queue removes a nonempty queue, orphans its jobs and schedules, and returns 204 for missing queues

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H14 (01:43:35Z–01:43:36Z), H14b.

## Minimal reproduction
1. Queue h02 holding 4 jobs. `DELETE /v1/queues/h02`.
2. `GET /v1/jobs/304` (a former h02 job).
3. `DELETE /v1/queues/never-existed`.
4. Queue h14s + schedule 3 on it; delete the queue; `GET /v1/schedules/3`; `POST /v1/schedules/3/run`.

## Expected (oracle)
SPEC: "DELETE … removes a queue; owned jobs and schedules must not become orphaned. Callers must handle a conflict for a nonempty queue." + "A missing resource returns 404."

## Observed
```
DELETE h02 → 204; orphan jobs (no queue row) = 4; GET /v1/jobs/304 → 200 {"queue_id":5,"state":"queued"}
DELETE never-existed → 204
schedule 3 survives with queue_id 8; run → 404 queue_not_found
```

## Variables isolated
Nonempty vs missing queue both 204. Job states in h02 included queued/retry_wait (all orphaned). Orphan queued jobs remain counted in namespace stats (inferred from stats.go query by namespace_id).

## Sibling check
No FK constraints in migrations/001.sql (jobs.queue_id, schedules.queue_id, webhook_outbox.job_id). Fire on an orphan schedule: fire.go:33-35 `continue`s without advancing (silent). DELETE schedule (schedules.go:92) also returns 204 for missing ids — predicted, not probed.

## Severity / status
Medium: data left unreachable by queue name but still visible by id; callers cannot detect the conflict the spec promises. Status: open.
