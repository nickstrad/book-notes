# F02 — Nack checks neither lease token nor job state: revives queued, succeeded and cancelled jobs

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observations: [notes.md](../experiments/01-map-it/notes.md) H03 (01:42:02Z), H07 (01:42:13Z–01:42:16Z), H27 (01:42:16Z), H28 (01:42:34Z).

## Minimal reproduction
Actor acme-key, queue h02 (max_attempts 3).
A. Never-leased job: `POST /v1/queues/h02/jobs {"kind":"other"}` → 303 (queued). `POST /v1/jobs/303/nack {"lease_token":"x","error":"spoof"}`.
B. Succeeded job 87 (acked in H01): `POST /v1/jobs/87/nack {"lease_token":"x","error":"late"}`.
C. Cancelled job: lease 302 (token T), `POST /v1/jobs/302/cancel` → cancelled; `POST /v1/jobs/302/nack {"lease_token":"T","error":"…","delay_seconds":1}`; wait 2.5 s; lease again.

## Expected (oracles)
- "Only the current lease holder may complete or extend a job" (nack completes an attempt) → 409 for A and B.
- "Nack with retry moves a job to retry_wait" applies to leased jobs; SPEC "A command incompatible with the current state returns 409".
- "A cancelled job is never run." (C)

## Observed
```
A: {"ok":true} [HTTP 200]   303|retry_wait|1|spoof|t   attempts row 303|1||failed|spoof (worker_id NULL)
B: {"ok":true} [HTTP 200]   87|retry_wait|2|{"done": true}|late
C: cancel 200 → 302|cancelled|0 ; nack 200 ; 2.5 s later 302|queued|1 ; next lease returns {"attempt":2,"id":302,…}
```

## Variables isolated
- Token: arbitrary string works (A, B); real token works on a cancelled job (C).
- State: queued, succeeded, cancelled all accepted. Leased not required.
- `retry:false` variant sends any job to `dead` (H28, cross-namespace, 306 → dead).
- Did not matter: worker identity (attempts.worker_id NULL in A).

## Sibling check
Same missing filter predicted in nack.go:10 (read) and :22 (update) — both unfiltered. Ack (F01) filters state but not token. Cancel/retry (operator.go) see F03/F06. Expire (expire.go:6) filters state correctly.

## Severity / status
Critical: a stale or rogue worker can resurrect succeeded or cancelled work (duplicate side effects) or kill arbitrary jobs; violates "cancelled job is never run". Status: open.
