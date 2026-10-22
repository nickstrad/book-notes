# F20 — Unknown fields and invalid enum values are accepted instead of rejected

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H18 (01:43:55Z).

## Minimal reproduction (acme-key)
`POST /v1/queues {"name":"h18a","bogus":1}`; `POST /v1/queues {"name":"h18b","backoff":{"kind":"zigzag","base_seconds":1}}`; `POST /v1/queues/h15/jobs {"kind":"x","priorty":9}` (typo); `GET /v1/queues/imports/jobs?state=bogus`; `?limit=1000`.

## Expected (oracle)
SPEC: "Unknown fields and invalid enum values should be rejected as invalid commands." "Backoff kind is linear or exponential." "Valid state filters are …" "The limit … is at most 500."

## Observed
```
bogus field → 201; backoff_kind "zigzag" stored → 201; "priorty" typo → 201 (priority silently default 5)
state=bogus → 200 with all states ["cancelled","dead","queued","retry_wait","scheduled","succeeded"]
limit=1000 → 200 with 100 rows (silently clamped to default)
```

## Variables isolated
JSON decoder lacks DisallowUnknownFields (api.go:43) — affects every body. Enum check absent for backoff kind (queues.go:35) and state filter (enqueue.go:101-105 falls back to ""). A "zigzag" queue behaves as exponential (internal/jobs/jobs.go:17-22) — inferred.

## Sibling check
All decode() call sites share the decoder: queues, PATCH queue, enqueue, batch, ack/nack/heartbeat, workers, lease, schedules. Kind filter on list is free-text (fine).

## Severity / status
Low–medium: typos silently change behaviour (priority, retry flags); a client filtering by a misspelled state gets every job. Status: open.
