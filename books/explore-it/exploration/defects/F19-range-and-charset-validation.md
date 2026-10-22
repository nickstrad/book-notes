# F19 — Range/charset rules not enforced: queue names, priority, PATCHed limits

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H18/H19 (01:43:55Z).

## Minimal reproduction (acme-key)
`POST /v1/queues {"name":"a.b"}`; `{"name":"ü-q"}`; `POST /v1/queues/h15/jobs {"kind":"x","priority":42}` and `-3`; `PATCH /v1/queues/h15 {"max_attempts":-1,"retention_hours":-5}` then enqueue.

## Expected (oracle)
SPEC: "Names use ASCII letters, digits, underscores and hyphens." "Priority ranges from 0 through 9." "Lease duration and maximum attempts must be positive."

## Observed
```
"a.b" → 201; "ü-q" → 201
priority 42 → 201 (id 815); -3 → 201 (id 816)
PATCH → {"max_attempts":-1,"retention_hours":-5}; next enqueue → 400 "max_attempts must be positive"
```

## Variables isolated
Create validates lease/max (queues.go:35) but not name charset; PATCH validates lease only (queues.go:73). Priority unchecked (enqueue.go:29-32). Negative retention: terminal jobs then match the retention sweep immediately (inferred from expire.go:14, not probed).

## Sibling check
Schedule PATCH (F09). Worker name/kinds — not checked for content (read). lease max_jobs is range-checked (read).

## Severity / status
Low–medium: priority 42 jobs jump all valid work; a PATCH can make a queue unusable for enqueue. Status: open.
