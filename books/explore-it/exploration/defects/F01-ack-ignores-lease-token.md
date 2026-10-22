# F01 — Ack completes a leased job without checking the lease token

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H02, entries 01:41:54Z–01:41:55Z.

## Minimal reproduction
Actor acme-key; any queue; worker registered with kind `echo`.
1. `POST /v1/queues/h02/jobs {"kind":"echo","payload":{"a":1}}` → 201 id 302.
2. `POST /v1/queues/h02/lease {"worker_id":4,"max_jobs":1}` → job 302 leased with token `75de57db…`.
3. `POST /v1/jobs/302/ack {"lease_token":"not-the-token","result":{"by":"impostor"}}`.

## Expected (oracle)
SPEC.md, Worker protocol: "Only the current lease holder may complete or extend a job." → 409 `state_conflict`, job stays `leased`.

## Observed
```
{"ok":true} [HTTP 200]
302|succeeded|1|{"by": "impostor"}
```
The real holder's later ack would get 409; the impostor's result is stored and notified.

## Variables isolated
- Token value: any string (here `not-the-token`). Job must be in state `leased` (ack.go filter); state did matter for ack, unlike nack (F02).
- Did not matter: payload, queue config.
- Control: heartbeat with a wrong token → 409 (H31, 01:46:19Z), so the token check exists for heartbeat only.

## Sibling check
Predicted before probing H03/H27/H31: nack (internal/jobs/nack.go:10) has no token filter → confirmed, see F02. heartbeat (internal/jobs/jobs.go:91) filters token → probe H31 confirmed it rejects. Lease release on worker delete (internal/api/jobs.go:110) is by worker id, not token — by design.

## Severity / status
High: any caller in the namespace (or with F03, any namespace) can complete another worker's job with a forged result; double-execution protection rests on the token. Status: open.
