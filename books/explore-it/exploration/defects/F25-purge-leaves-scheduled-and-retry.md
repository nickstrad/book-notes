# F25 — Purge removes only `queued` jobs; scheduled and retry_wait remain (spec silence)

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H24 (01:44:08Z).

## Minimal reproduction
Queue h24 with one queued, one scheduled (delay 3600) and one retry_wait (nack delay 3600) job; `POST /v1/queues/h24/purge`.

## Expected (oracle)
SPEC: "purge removes all pending jobs." The spec never defines "pending". Plausible readings: (a) queued only; (b) every not-yet-run state (scheduled, queued, retry_wait). The API's own stats fold all three into "queued" (F18), which leans toward (b).

## Observed
```
{"deleted":1} [HTTP 200]   remaining: s|scheduled, r|retry_wait
```

## Variables isolated
State is the only variable (internal/api/queues.go:109 filters state='queued').

## Sibling check
None found beyond queues.go:109.

## Severity / status
Low–medium if (b) is intended: purged queues refill when delayed work comes due. Status: `spec` — needs a spec line defining "pending" before it can be called a defect.
