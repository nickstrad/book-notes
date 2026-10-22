# E06-F02 — Schedules accept a non-object payload and cron fires it into jobs that no other path can create; the sample worker cannot decode them

Seen at: export of `explore-it/lesson-05` (stack `queuelet-e06`), experiment 06 (relate-it).
Original observations: [E06 notes](../experiments/06-relate-it/notes.md) P2 (15:32:15Z), P2b (15:32:24Z, H-E06-04),
P3 (15:33:35Z, H-E06-12). All three outcomes were predicted from code before probing.

## Minimal reproduction

Actor: acme key. Clock: real (cron `* * * * *`, wait for the next minute). Fixture: seeded stack (worker-a serves `emails`,
lease 10 s, max_attempts 3).

```sh
source exploration/experiments/06-relate-it/tools/env.sh
k acme-key POST /v1/schedules '{"name":"arr","queue":"emails","kind":"echo","payload":[1,2],"cron":"* * * * *"}'  # [201] "payload":[1,2]
k acme-key POST /v1/schedules '{"name":"nul","queue":"emails","kind":"echo","cron":"* * * * *"}'                  # [201] "payload":null
k acme-key POST /v1/schedules/<arr id>/run      # {"code":"invalid_job","message":"payload must be a JSON object"} [400]
# after the next minute boundary:
sql "select id,state,payload,jsonb_typeof(payload) from jobs where jsonb_typeof(payload)<>'object'"
```
Replay: `tools/p2-schedules.sh`, `tools/p3-poison-fire.sh` (disables the schedules after one fire).

## Expected (named oracle)

SPEC:60 a job accepts "a JSON object `payload`"; enqueue (`enqueue.go:64-70`), batch and `POST /v1/schedules/{id}/run`
enforce it (400). A schedule is a stored job template (SPEC "takes name, queue, kind, payload …"; `run` "enqueues one
immediate execution"), so the same payload rule applies to the template, and the cron path must not create a job the
enqueue contract forbids. Expected: schedule create (and therefore fire) rejects a non-object payload with 400, or —
at least — the fire path applies the same validation as `run`. An omitted payload should behave as on enqueue
(`{}`, `enqueue.go:64-65`).

## Observed

- Create: payload omitted → stored JSON `null`; `[1,2]`, `"x"`, `null`, `7` → all `201`, stored as given (ids 4–8).
- `run` of the same schedules: `400 invalid_job "payload must be a JSON object"`.
- Cron fire at 15:33:00: jobs 302–306 with payloads `null`, `[1, 2]`, `"x"`, `null`, `7` (`jsonb_typeof` null/array/
  string/null/number), state `queued`; listed by `GET /v1/queues/e6-pay/jobs`.
- Downstream (P3, emails/worker-a): job 307 (`[1,2]`) was leased three times; each time worker-a logged
  `json: cannot unmarshal array into Go struct field Job.jobs.payload of type map[string]interface {}` (no job id),
  never acked or nacked; three expiries 15:34:10→15:34:34 → `dead`, `last_error` "lease expired", attempts
  `[1,2,"expired"] [2,2,"expired"] [3,2,"expired"]`. The job was executed zero times. Job 308 (payload omitted →
  `null`) was acked with result `null`.

## Variables isolated

- Matters: the payload's JSON type at schedule creation (object vs anything else) and the creation path (cron fire vs
  run).
- Does not matter: kind, cron, time zone; queue settings (e6-pay and emails both); null vs omitted (identical).
- The poison effect needs a worker built on `internal/worker` (Payload `map[string]any`, worker.go:21); a worker that
  decodes `any` would execute it.

## Sibling search

- `schedules.go:42-60` createSchedule: no type check; `schedule/json.go:5` turns a nil RawMessage into `null`.
- `fire.go:76` INSERT copies `payload` bytes with no validation (the only job writer that bypasses `insertJob`).
- `schedules.go:144-165` runSchedule: validated (400) — the inconsistent sibling.
- PATCH schedule cannot change payload (`schedules.go:74-80`), so a bad template cannot be repaired, only deleted.
- Kind: also only checked non-empty at create; fire copies it; same as enqueue — no extra divergence found.

## Severity and status

Severity: **medium**. A tenant configuration error is accepted silently, then turns every tick into a job that the
documented worker cannot decode: it burns all attempts by lease expiry (lease_seconds × max_attempts of worker time per
tick), ends `dead` with a misleading "lease expired", and the worker log cannot name the job. Status: **open**.
