# E04-F05 — `oldest_queued_age` counts time a job spent leased when the lease ends by expiry or worker deletion

- Seen at: `explore-it/lesson-03` plain export; stack `queuelet-e04`; lab clock frozen (2031-07-01, 2031-08-01); worker 3/5.
- Original observation: [E04 notes](../experiments/04-judge-it/notes.md) P6 (12:47:17Z) and P8 (12:48:00Z); hypothesis R40
  stated 12:44Z; the worker-delete sibling was predicted after P6 but logged only after its probe. Raw: `project/results/e04/p6-expiry.txt`, `p8-age-sibling.txt`; side observation in
  `p5-cancel-stats.txt`. Prediction source (code): `internal/stats/stats.go:18` uses `GREATEST(created_at, run_at)`;
  `internal/jobs/expire.go:41` and `internal/api/jobs.go:126` return a job to `queued` without touching `run_at`.

## Minimal reproduction
1. Admin clock set `2031-07-01T00:00:00Z`. Queue `{"name":"e04f","lease_seconds":30}`; enqueue one echo job; lease it.
2. Clock advance 600; wait ≥1 s for the sweep (job → `queued`, attempts 1, `event_at` 00:10:00).
3. `GET /v1/queues/e04f/stats`.
Sibling: queue with lease 3600, lease the job, advance 100, `DELETE /v1/workers/{id}`, read stats.

## Expected (oracle)
SPEC Notifications/ops: "Queued age uses the later of creation and readiness, over queued rows only." A job returned by
expiry or worker deletion becomes ready again at that moment (or at the latest at `lease_expires_at`), so the oldest queued
age of a queue whose only job was just released is ≈0 (bound: ≤ time since `lease_expires_at`, here ≤570 s; release by
deletion: 0 s). Approximation rule R40: the metric measures waiting-to-be-leased time. (Comparable: SQS's
ApproximateAgeOfOldestMessage counts from the original send, so a different product could choose creation — but this spec
already chose "later of creation and readiness", and the U08 fix applied it to scheduled/retry rows.)

## Observed
```text
expiry:   57 | queued | attempts 1 | created 00:00:00 | run_at 00:00:00 | event_at 00:10:00 → oldest_queued_age 600
deletion: 64 | queued | attempts 0 | created 00:00:00 | run_at 00:00:00 | event_at 00:01:40 → oldest_queued_age 100
control:  65 untouched queued for 100 s                                         → oldest_queued_age 100   (correct)
```
Namespace view (P5): after a one-year clock jump expired 40 leases, acme `oldest_queued_age` = 31 536 000 s.

## Variables isolated
- Release path: lease expiry and worker deletion both affected; `retry_wait`→`queued` (Promote, run_at = readiness) and
  operator retry (run_at := now, `operator.go:13`) are not, by reading.
- Lease length only scales the error (600 s vs 100 s); one job per queue so `min()` is that job.
- Counts by state were exact (R34 held) — only the age is wrong.

## Sibling check
- `deleteWorker` release: confirmed (above).
- Lease ordering uses `created_at`, not `run_at` — not affected.
- `throughput_last_hour` uses `finished_at` — not affected.

## Severity / status
Low: a queue-backlog alarm on oldest age fires (or a dashboard shows hours of waiting) right after a long lease expires or a
worker is removed, although nothing is waiting. Status: open; the oracle depends on reading "readiness" as "became
queued" — flagged for the owner.
