# F14 — During a database outage valid keys get 401 "invalid API key"; /health stays ok

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observations: worker logs during H26 ([notes.md](../experiments/01-map-it/notes.md) 01:52:16Z "worker-a/b last log lines": `HTTP 401 … invalid API key` at 01:49:47), reproduced as H42 (01:52:25Z–01:52:26Z).

## Minimal reproduction
1. `docker compose -f deploy/compose.yml stop postgres` (isolated project).
2. `GET /v1/queues` with acme-key; `GET /health`.
3. Start postgres; restart api (F08 recovery).

## Expected (oracle)
HTTP semantics (RFC 9110 §15.5.2): 401 means the request lacks valid credentials; a backend failure is 5xx (500/503). SPEC: "Errors have the form … and a 4xx or 5xx HTTP status" — the class must reflect the cause. SPEC "GET /health returns service health".

## Observed
```
{"error":{"code":"unauthorized","message":"invalid API key"}} [HTTP 401]
{"status":"ok"} [HTTP 200]
```

## Variables isolated
Same key succeeds when DB is up. DB availability is the only variable. Cause read at api.go:86-88: any QueryRow error → 401.

## Sibling check
/health (api.go:56) never touches the DB → reports ok while DB is down and while the loop is dead (F08). Other handlers map DB errors to 500 `database_error` via `output` (api.go:125) — but queueOrError maps any error to 404 (api.go:117-123), and getJob maps any error to 404 (internal/api/jobs.go:12-14): predicted to return 404 during an outage if auth were bypassed (not probeable since auth fails first).

## Severity / status
Medium: clients may discard or rotate valid credentials and alerting sees auth failures instead of an outage; health checks cannot detect it. Status: open.
