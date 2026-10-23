# E03-F04 — Out-of-range queue settings and over-long names are reported as 409 conflict (create) or 500 (PATCH) with driver text

- Seen at: `explore-it/lesson-02` plain export; stack `queuelet-e03`; fixture seed 04:16:24Z.
- Original observation: [E03 notes](../experiments/03-vary-it/notes.md) E03-H06 (stated 04:16Z, observed 04:17:59Z);
  name-length cells from E03-H11 (04:22:00Z–04:22:28Z). Raw: `project/results/e03/p4-queuecfg.txt`, `p8-misc.txt`,
  `p8b-namelen.txt`; helper `exploration/experiments/03-vary-it/tools/namelen.py`. Prediction source: `internal/api/queues.go:46-49`
  (any INSERT error → 409), `:93-94` (any UPDATE error → 500), `migrations/001.sql:3` (`int` columns, unique (namespace_id,name)).

## Minimal reproduction
1. `POST /v1/queues -d '{"name":"e03-o","lease_seconds":2147483648}'` (acme-key).
2. `PATCH /v1/queues/<existing> -d '{"lease_seconds":2147483648}'`.
3. `python3 -I tools/namelen.py rand 2800` (POST a 2 800-character random alphanumeric name).

## Expected (oracle)
SPEC:13-16: errors are 4xx/5xx JSON; "A command incompatible with the current state returns 409." An out-of-range number or a
name the store cannot index is an invalid command, not a state conflict and not a server fault → 400 with a service-level
message. (Approximation: RFC 9110 §15.5.10 defines 409 as a conflict with the target resource's current state.)

## Observed
```text
1 → 409 {"error":{"code":"queue_conflict","message":"pq: value \"2147483648\" is out of range for type integer"}}
    same 409 for max_attempts, backoff.base_seconds, backoff.max_seconds = 2147483648; 2147483647 → 201
2 → 500 {"error":{"code":"database_error","message":"pq: value \"2147483648\" is out of range for type integer"}} (also max_attempts)
3 → 409 {"error":{"code":"queue_conflict","message":"pq: index row size 2824 exceeds btree version 4 maximum 2704 for index \"qu…"}}
    random 10 000 chars → 409 "pq: index row requires 10024 bytes, maximum size is 8191"; random 2 600 → 201 (GET 200, DELETE 204)
    "q"×100 000 → 201 (GET 200, DELETE 204); "q"×900 000 / ×1 100 000 → 409 "index row requires …"
control: duplicate name → 409 queue_conflict (the one legitimate conflict uses the same code)
```

## Variables isolated
- Field: every `int`-backed queue field fails the same way above 2^31 − 1; per-job `max_attempts` 2147483648 on enqueue is a
  correct 400 `invalid_job` (still driver text) because enqueue maps all insert errors to 400.
- Verb: create → 409, PATCH → 500 for the identical value.
- Name length: the effective limit is undocumented and depends on compressibility of the name (≈2 700 bytes random, ≥100 000
  repetitive) because it is the btree index-row limit after TOAST compression.
- A client cannot distinguish "name taken" from "value too large" by status or code.

## Sibling check
- `internal/api/queues.go:93-94` PATCH `output()` → 500 for any DB error (observed).
- `internal/api/api.go:173-179` `output` is shared by list/stats/attempts reads — DB errors there are true 500s (fine).
- `internal/api/enqueue.go:81-84` maps every insert error (including real DB outages) to 400 `invalid_job` — the opposite
  misclassification for a DB fault (inference, not probed).
- Schedules create (`schedules.go`) — not probed (#offcharter).
- Driver text in messages: E01 Q3 (open question) — same leak.

## Severity / status
Low: no data corruption; clients and operators get a misleading error class (409 suggests retry-after-rename; 500 suggests a
server incident) and internal driver text. Status: open.
