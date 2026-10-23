# F23 — FIFO within a priority is not by creation order for jobs created in the same second

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H22 (01:45:17Z, inconclusive), H22b (01:45:35Z–01:45:48Z).

## Minimal reproduction
1. Queue h22b; enqueue 10 echo jobs priority 5 back-to-back (payload i=1..10; ids 862–871; created_at 01:45:35 ×5, 01:45:36 ×5).
2. New worker W leases 1 job; `DELETE /v1/workers/W` (releases it); worker 4 leases 10 times, one at a time.

## Expected (oracle)
SPEC: "Higher priorities run first. Within a priority, jobs run FIFO by creation order."

## Observed
```
W's first lease → {"id":863,"i":2}   (id 862 / i=1 was created earlier, same priority, queued)
subsequent order of i: 3 4 5 2 1 6 7 8 9 10
```
i=1 ran fifth. The inversion appeared on the very first lease, before any release.

## Variables isolated
- H22 (40 untouched jobs, 2 created_at buckets): no inversion — ties often fall in insertion order, so the defect is order-dependent, not deterministic per run.
- created_at is `timestamp(0)` (migrations/001.sql) → one-second resolution; LeaseOrder "priority DESC, created_at" (internal/jobs/jobs.go:24) has no id tie-breaker. Inferred mechanism: ties resolved by physical/index order; not proven which.
- Lease/release (UPDATE) of i=2 did not restore order.

## Sibling check
Schedule-fired jobs use created_at=now for every catch-up job (fire.go:44) → identical created_at within a burst (seen in F24: 4 jobs with created_at 01:48:05) → same tie problem. List ordering by id is unaffected.

## Severity / status
Medium: order-sensitive producers see reordering under bursty load. Reproduction is probabilistic; one clear inversion observed in one of two attempts. Status: open.
