# E09-F01 — One schedule that cannot advance stops every later schedule, in every namespace

Seen at: export of `explore-it/lesson-08` (stack `queuelet-e09`), experiment 09 (explain-it).
Original observations: [E09 notes](../experiments/09-explain-it/notes.md) H-E09-02 (20:36:44Z), H-E09-02b
(20:36:58Z), H-E09-03 (20:37:06Z), H-E09-06 (20:37:26Z). Predicted from the cause model before probing (PD1).

## Minimal reproduction

Actors: acme (`acme-key`), globex (`globex-key`), admin for the test clock. Fixture: `make seed`. Clock pinned
(`POST /admin/clock {"action":"set","at":"2026-10-22T21:00:30Z"}`).

```sh
cd books/explore-it/exploration/experiments/09-explain-it; . tools/env.sh
tools/q.sh acme-key   POST /v1/schedules '{"name":"feb30","queue":"imports","kind":"echo","cron":"0 0 30 2 *","timezone":"UTC"}'
tools/q.sh globex-key POST /v1/schedules '{"name":"victim","queue":"emails","kind":"echo","payload":{"v":1},"cron":"* * * * *","timezone":"UTC"}'
tools/q.sh admin-key  POST /admin/clock '{"action":"advance","seconds":60}'   # twice, 3 s apart
tools/psql.sh "select id,next_run_at,last_run_at from schedules order by id"
tools/q.sh acme-key GET /health
```

The bad schedule must have a lower id than the victim. Any cron that never matches works (see E09-F02).

## Expected (named oracle)

- SPEC.md:189 "The background loop … fires cron schedules"; SPEC.md:166 enabled schedules enqueue jobs — the
  globex every-minute schedule must fire.
- SPEC.md:10 "A namespace key accesses only its namespace's resources" (approximation: one tenant's schedule must
  not deny another tenant its schedules); same design intent as SPEC.md:204-205 for webhook rows ("does not stop
  later rows").

## Observed

```text
victim id 4 next_run_at 2026-10-22 21:01:00 last_run_at NULL; jobs for it: 0 (after 21:01:30Z and 21:02:30Z)
GET /health → 503 {"status":"degraded"}
api log every second: "msg":"background task: schedule cannot advance"
PATCH /v1/schedules/3 {"enabled":false} → next tick: victim fires (job 303, run_at 21:01:00), /health 200 ok
```

## Variables isolated

- Matters: schedule id order. Schedule 2 (lower id) kept firing while 3 was broken (H-E09-02b).
- Does not matter: namespace (cross-tenant shown), or the bad schedule's queue being paused (H-E09-06: paused
  reports queue, victim2 still blocked), or the cron branch (fixed-hour `0 0 30 2 *` and wildcard `* * 30 2 *`).
- Only the bad schedule matters: disabling or deleting it alone releases the victims at the next tick (H-E09-03,
  20:37:36Z recovery).
- Not affected: Promote, Expire, Retain (the loop continues past a failed step, api.go:249-253).

## Mechanism (reading)

`Fire` loops over due schedules and `return e` on the first `fireOne` error (internal/schedule/fire.go:72-75).
`fireOne` errors "schedule cannot advance" when `Next` returns a time not after now (fire.go:102-108); `Next`
returns the zero time without error when the cron never matches (fire.go:32-33, :25), so the bad schedule stays due
forever.

## Sibling check

Predicted before probing (habit "one row's error aborts a set-based pass"): Expire's notify loop
(jobs/expire.go:26-29) and worker DELETE's notify loop (api/jobs.go:154-158) — both confirmed, see E09-F08. Earlier
fixed members of this habit: E02-F02, E03-F05, E08-F02. Already predicted as H-E08-23 ("no external trigger found"); this run supplies the trigger. Other `fireOne` error sources (queue lookup fire.go:98, job
INSERT fire.go:111): queue delete refuses queues with schedules (api/queues.go:133-139), so none found.

## Severity and status

High: a single tenant, through the public API, silently stops all later schedules of all tenants until an operator
finds the culprit; the log line names no schedule (TR-E09-1). Status: open.
