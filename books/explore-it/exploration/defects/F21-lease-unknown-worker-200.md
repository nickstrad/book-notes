# F21 — Lease with an unknown worker returns HTTP 200 with an error body

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H20 (01:44:07Z).

## Minimal reproduction
`POST /v1/queues/imports/lease {"worker_id":9999}` (acme-key). Also `qctl lease imports --worker-id 9999`.

## Expected (oracle)
SPEC: "Errors have the form {error:{…}} and a 4xx or 5xx HTTP status." "A missing resource returns 404."

## Observed
```
{"error":{"code":"worker_not_found","message":"unknown worker"}} [HTTP 200]
qctl prints the error JSON on stdout, exit 0
```
The reference worker (worker.go:42) decodes the 200 as an empty batch and polls forever without logging a problem (inferred from code).

## Variables isolated
Worker id unknown (9999) and foreign-namespace worker (globex key + acme worker 4) both yield the same 200.

## Sibling check
internal/api/jobs.go:132 `fail(w, 200, …)` is the only non-4xx/5xx fail() call (grep of fail( call sites by reading). 

## Severity / status
Low–medium: deregistered workers silently starve. Status: open.
