# F03 — Job-id endpoints ignore the caller's namespace

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observations: [notes.md](../experiments/01-map-it/notes.md) H04 (01:42:24Z–01:42:25Z), H28 (01:42:34Z).

## Minimal reproduction
Fixture: acme job 1 (emails, payload {"fixture":0}). Actor globex-key.
1. `GET /v1/jobs/1` with globex-key.
2. `GET /v1/jobs/87/attempts` with globex-key.
3. acme enqueues 305; globex `POST /v1/jobs/305/cancel`, then `DELETE /v1/jobs/305`.
4. acme enqueues 306; globex `POST /v1/jobs/306/nack {"error":"from globex","retry":false}`.

## Expected (oracle)
SPEC Authentication: "A namespace key accesses only its namespace's resources." → 404 for every call (as done for queues and schedules: controls returned 404).

## Observed
```
1: 200 {"…","namespace_id":1,"payload":{"fixture":0},"queue_id":1,…}
2: 200 {"attempts":[…job 87…]}
3: cancel 200 {"ok":true}; delete 204; acme GET 305 → 404 job_not_found
4: 200; 306|1|dead|from globex
Controls: globex GET /v1/queues/h02 → 404; GET /v1/schedules/1 → 404
```

## Variables isolated
Key (globex vs acme) is the only variable; job ids are sequential and guessable. Queue- and schedule-scoped endpoints filter correctly.

## Sibling check
Predicted from code before H28: internal/api/jobs.go:11 (get), :19 (attempts), :23 (delete), jobVerb → jobs/operator.go:6,13 (cancel/retry), nack.go:10 — none filter namespace_id. Probed: get, attempts, cancel, delete, nack. Not probed: retry, ack (ack needs leased state; token unchecked per F01, so predicted exploitable). Lease worker lookup internal/api/jobs.go:130 does filter namespace (globex lease with acme worker 4 → worker_not_found, 01:44:07Z).

## Severity / status
Critical: cross-tenant read (payload/results), cancel, delete and kill. Status: open.
