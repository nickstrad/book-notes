# E03-F05 — One queue's large `retention_hours` stops terminal-job retention for every queue in every namespace and holds /health at 503

- Seen at: `explore-it/lesson-02` plain export; stack `queuelet-e03`; real clock; fixture seed 04:16:24Z.
- Original observation: [E03 notes](../experiments/03-vary-it/notes.md) E03-H09 (stated 04:16Z, observed 04:18:27Z–04:19:23Z).
  Raw: `project/results/e03/p4-queuecfg.txt`, `p5-retention.txt`, `p5b-retention-isolate.txt`, `all-logs-at-stop.txt`.
  Prediction source: `internal/jobs/expire.go:66` `Retain` (one statement over all queues:
  `finished_at < $1 - q.retention_hours * interval '1 hour'`), `internal/api/queues.go:38` (retention only `>= 0`),
  `internal/api/api.go:148-154`.

## Minimal reproduction
1. acme: `POST /v1/queues -d '{"name":"e03-ret-1e8","retention_hours":1e8}'` → 201. (`1e308` also works; see below.)
2. Enqueue one job there and cancel it (`POST /v1/jobs/{id}/cancel` → 200), so the queue owns a terminal job.
3. Any namespace: a queue with `retention_hours: 0`; enqueue + cancel a job (control that should be removed on the next tick).
4. Poll the control row, `GET /health`, and the API log for ≥5 s.

## Expected (oracle)
SPEC:98 "Terminal jobs are deleted after `retention_hours`" (for each queue, by its own setting); SPEC:25 retention hours must
be nonnegative — 1e8 is a valid nonnegative value, so it must either work (keep that queue's jobs ~11 400 years) or be
rejected with 400 at create/PATCH (SPEC:17). SPEC:10 approximation: one namespace's configuration must not deny service to
another. SPEC:127-130: health reports real loop degradation.

## Observed
```text
before poison: control job 832 (retention 0) cancelled 04:18:27.686 → removed by 04:18:28.404 (event "retained" 04:18:28.010)
poison job 833 cancelled 04:18:28.582; control job 834 cancelled immediately after
04:18:29.700 … 04:18:35.224 (6 samples): /health {"status":"degraded"} [503]; job 834 still present
api log each second: {"msg":"background task: pq: timestamp out of range"}
globex queue gx-ret0 (retention 0): job 836 cancelled 04:18:57 → still present 04:19:01.859
Promote unaffected: delayed job 835 promoted to queued within 3 s
PATCH e03-ret-1e8 retention 24 → 04:19:04.580 health ok; acme 834 and globex 836 removed on that tick
```

## Variables isolated
- Entry point: create and PATCH (`PATCH {"retention_hours":1e8}` → 200 → degraded again 04:19:06.893).
- Magnitude: 5e7 h + terminal job → ok; 6e7 h → degraded (bracket observed). This matches the timestamp minimum (4713 BC
  ≈ 6 739 years ≈ 5.9e7 h before 2026; reason inferred). `1e308` fails differently: log "pq: interval out of range" → degraded.
- Trigger requires a terminal job in the poisoned queue: the queues sat harmless while empty (04:17:59–04:18:28).
- Namespace: globex retention blocked by an acme setting. Promotion, expiry and webhook shipping continued (Promote observed;
  others inferred from loop order `api.go:148`).
- Recovery: lowering retention (04:19:04) or deleting the terminal job (`DELETE /v1/jobs/833` → 204 → ok 04:19:08.947).

## Sibling check
- `queues.go:80` PATCH accepts the same values (observed).
- Other SQL interval arithmetic on stored config: none found — grep `interval` over `project/internal/` matches only
  `internal/jobs/expire.go:66` (Retain itself); expiry/backoff are
  computed in Go (`jobs.go:61`, `nack.go:22-26`, see [E03-F06](E03-F06-backoff-overflow-immediate-retry.md)).
- Schedules (`schedule/fire.go`) evaluate per row in Go — not probed (#offcharter).

## Severity / status
High: a single accepted configuration value in any tenant silently disables retention for all tenants (unbounded table growth,
SPEC:98 violated everywhere) and pins health at 503 with no cause named, until someone finds and edits that queue or job.
Status: open.
