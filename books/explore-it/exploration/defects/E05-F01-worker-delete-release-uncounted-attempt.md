# E05-F01 — Releasing a lease by deleting the worker hands the job out again under the same attempt number, without limit

Seen at: export of `explore-it/lesson-04` (stack `queuelet-e05`), experiment 05 (sequence-it).
Original observations: [E05 notes, P6](../experiments/05-sequence-it/notes.md) (13:23:10Z), isolation H-E05-11/12
(13:24:41Z), random walk seed 11 steps 26→78 and 133→136/146 ([walk log](../experiments/05-sequence-it/notes.md)).

Prior record: E01 observed the release itself (`leased | worker deleted | queued, no attempt`, states.md, H22b) and
listed "whether release counts an attempt" as a read-only question (E01 notes line 1460); no earlier experiment drove a
repeated release or judged it against the max-attempts rule, and no earlier report covers it.

## Minimal reproduction

Actor: one namespace key (acme). Clock: any (observed both on the real clock and with the test clock frozen). No
compose worker involved (queue name not polled).

```sh
source exploration/experiments/05-sequence-it/tools/e05env.sh
api POST /v1/queues -d '{"name":"p6","max_attempts":2}'
api POST /v1/queues/p6/jobs -d '{"kind":"k"}'                         # job 8
for i in 1 2 3 4; do
  W=$(api POST /v1/workers -d "{\"name\":\"p6w$i\",\"kinds\":[\"k\"]}" | sed 's/ \[.*//' | jq -r .worker_id)
  api POST /v1/queues/p6/lease -d "{\"worker_id\":$W,\"max_jobs\":1}"  # "attempt":1 every time
  api DELETE /v1/workers/$W                                            # 200, job back to queued
done
api GET /v1/jobs/8/attempts                                            # {"attempts":[]}
```

Replay script: `exploration/experiments/05-sequence-it/tools/pairs.sh P6`; variants `tools/isolate.sh`.

## Expected (named oracle)

SPEC "Worker protocol": *"No job may execute more than its maximum attempts."* SPEC "Jobs": *"Attempt `n` numbers every
execution of the job and is never reused …; the lease's `attempt` counts toward `max_attempts`."* rules.md R3/R32 (Never
rule). Comparable system (recalled, not verified in this run): SQS increments `ApproximateReceiveCount` on every receive,
and the redrive policy's `maxReceiveCount` counts receives regardless of how the message became visible again.

So each lease handed out should consume an attempt (or the release should otherwise be counted): with max_attempts 2, a
third delivery must not happen, and two deliveries must not both be told `attempt: 1`.

## Observed

| Variant | Deliveries observed | Lease `attempt` values | `attempts` column | attempts history |
|---|---|---|---|---|
| max 2, fresh job (P6, job 8) | 4 (workers 6, 7, 8, 9) | 1, 1, 1, 1 | 0 | `[]` |
| max 3 after one nack (H-E05-11, job 13) | 3 (w13, w14, w15) | 1, 2, **2** | 1 | n1 failed only |
| max **1** (H-E05-12, job 14) | 2 (w16, w17) then nack → dead | 1, 1 | 1 | n1 failed (w17) only |
| walk seed 11, max 2 (jobs 21, 24, 25) | 3 each | job 21: 1, 1, 2, 2 | ≤ 2 | expiry rows only |

`job_events` for job 8: `created | leased 6 (1) | queued 6 (0) | leased 7 (1) | queued 7 (0) | leased 8 (1) | queued 8 (0) |
leased 9 (1) | queued 9 (0)` — the only place the deliveries appear.

Source route (code reading, consistent with the runtime): `internal/api/jobs.go:126` releases with
`UPDATE jobs SET state='queued',…,run_at=$2,lease_token=NULL,leased_by=NULL,lease_expires_at=NULL` — `attempts` unchanged
and no `attempts` row; `internal/jobs/jobs.go:77` reports `attempt = attempts+1`. Contrast `expire.go:116-120`, which for the
same "holder is gone" situation increments `attempts`, records an `expired` row and dead-letters at max.

## Variables isolated

- Matters: a `DELETE /v1/workers/{id}` while the worker holds the lease. Nothing else between the leases (no expiry,
  no nack) in P6.
- Does not matter: max_attempts (1, 2, 3); whether earlier counted attempts exist (H-E05-11 reuses attempt 2); real
  vs frozen test clock; how many jobs the worker holds (walk step 133 released three jobs at once, all reused).
- Old holder's token after release: heartbeat 409 (P6 ×4) — the lease *is* revoked; the issue is only accounting.
- Not tried: a worker process that deregisters itself on shutdown (the sample worker in `cmd/worker` does not call
  DELETE — grep found no deregistration), so the realistic trigger is operator cleanup or a third-party worker that
  deregisters on SIGTERM while executing.

## Sibling search

Predicted before probing (notes 13:24:30Z, H-E05-13): other release/exit paths. Checked:

- `expire.go:116` — counts (P7: max 1 → dead with `expired` row). Not a sibling.
- `nack.go:60` — counts. Not a sibling.
- `operator.go:6` Cancel of a leased job — no attempt row for the in-flight delivery (H-E05-14), but the job is terminal
  and cannot run again, so no budget bypass; reported with its lease-field symptom in
  [E05-F02](E05-F02-terminal-rows-keep-lease-fields.md).
- `operator.go:13` Retry resets attempts by design (SPEC: "resets its attempt count").
- Event console mislabels this release (and any later ack/cancel) as `lease_expired` once `last_error` is "lease expired":
  [E05-F03](E05-F03-worker-release-logged-as-lease-expired.md).

## Severity and status

Severity: **medium-high**. A poison job whose workers are deregistered while running (operator cleanup of stuck workers,
autoscaler scale-in that deregisters) is redelivered forever and never reaches `dead`; every redelivery is reported to
the worker as the same attempt number, and the attempts API shows none of them, so the retry budget and the history both
understate execution. Breaks a SPEC Never rule.

Status: `open`.
