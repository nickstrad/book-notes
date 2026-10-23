# E05-F03 — After one lease expiry, every later exit from `leased` (worker release, ack, cancel) is logged as `lease_expired`

Seen at: export of `explore-it/lesson-04` (stack `queuelet-e05`), experiment 05 (sequence-it).
Original observation: walk seed 11, job 21 event 134 (notes 13:26:45Z); minimal reproduction H-E05-15 (13:32:08Z, raw
`project/results/e05/f02-f03.txt`); sibling H-E05-18 (13:32:51Z, raw `project/results/e05/f03-sibling.txt`).

## Minimal reproduction

Actor: acme key. Clock: test clock frozen; advanced explicitly with `/admin/clock` (admin key).
Replay: `exploration/experiments/05-sequence-it/tools/f02.sh` (first block) and `tools/f03sib.sh`.

```sh
source exploration/experiments/05-sequence-it/tools/e05env.sh
api POST /v1/queues -d '{"name":"f3s","lease_seconds":2,"max_attempts":5}'
api POST /v1/queues/f3s/jobs -d '{"kind":"k"}'     # 111
api POST /v1/queues/f3s/jobs -d '{"kind":"k"}'     # 112
# worker 55 leases both (attempt 1); adm advance 3 s; after one loop tick both queued, attempts 1, last_error "lease expired"
# worker 55 leases both again (attempt 2, expires 14:35:43; clock 14:35:41)
api POST /v1/jobs/111/ack -d '{"lease_token":"<token>","result":{"ok":true}}'   # 200 -> succeeded
api POST /v1/jobs/112/cancel                                                    # 200 -> cancelled
api GET /v1/jobs/111/events ; api GET /v1/jobs/112/events
```

Observed (API `/events`; identical `job event` lines in `docker compose logs api`):

| Job | Transition actually made | Logged event | attempt column |
|---|---|---|---|
| 111 | ack → **succeeded** | `lease_expired` (id 353) | 2 |
| 112 | operator cancel → **cancelled** | `lease_expired` (id 354) | 1 |
| 108 | `DELETE /v1/workers/53` → queued, lease valid 2 s more | `lease_expired` (id 338) | 1 |
| 21 (walk 11) | worker 30 deleted at step 133 → queued | `lease_expired` (id 134) | 1 |

Controls: the same transitions without an earlier expiry are labelled by state — `succeeded` (walk seeds), `cancelled`,
and `queued` for a worker release (P6, four times). `GET /v1/jobs/111` itself is correct: `"state":"succeeded"`.

## Expected (named oracle)

The event console's contract: migrations/002.sql "Observation only" triggers whose label names the transition (`created`,
the new state, or `lease_expired` for an expiry); the console is the SUT's exposed job history (`GET /v1/jobs/{id}/events`
and the API log). Consistency approximation: the same command produces the same event label regardless of an earlier,
unrelated error text; and an event never contradicts the job's state (`lease_expired` for a job now `succeeded`).

## Cause (code reading, consistent with the runtime)

`observe_job()` (migrations/002.sql): `label := CASE WHEN OLD.state='leased' AND NEW.last_error='lease expired' THEN
'lease_expired' ELSE NEW.state END`. `last_error` is sticky: `expire.go:116` writes it, and Ack (`ack.go:25`), Cancel
(`operator.go:6`) and worker release (`internal/api/jobs.go:126`) do not touch it (SPEC keeps "last error" on later
success, E02 Q-E02-2). So any later exit from `leased` matches the expiry test.

## Variables isolated

- Matters: `last_error = 'lease expired'` (from an earlier expiry) at the moment the job leaves `leased`.
- Does not matter: which exit (ack, cancel, worker deletion all mislabelled); same or different worker (108: 52 then 53;
  111/112: 55 both times).
- Not mislabelled: nack (sets a new `last_error`); a second real expiry (correct label).
- Not tried: a worker nack whose error text is literally `lease expired` (same condition; would be labelled an expiry —
  prediction only).

## Sibling search

Predictions before probe: H-E05-18 (ack and cancel) — confirmed. Other triggers: `observe_notification()` labels by
`delivered_at`, no sticky-text test — none found. Retain's `retained` event is inserted directly (`expire.go:141`), not
via the CASE — none.

## Severity and status

Severity: **medium** (observability). After any lease timeout, the history and log can no longer show that the job
succeeded, was cancelled by an operator, or was released by worker deregistration; all read as further lease expiries.
An operator diagnosing slow workers or E05-F01-style redelivery is pointed at lease duration. No lifecycle effect.

Status: `open`.
