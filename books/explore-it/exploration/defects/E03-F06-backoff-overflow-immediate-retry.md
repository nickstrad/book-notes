# E03-F06 — Retry backoff overflows: from attempt 32 with the default base (or attempt 4 with a large base) a failed job is retried immediately

- Seen at: `explore-it/lesson-02` plain export; stack `queuelet-e03`; real clock; fixture seed 04:16:24Z; worker 4 (`e03-prober`, raw API).
- Original observation: [E03 notes](../experiments/03-vary-it/notes.md) E03-H08 (stated 04:16Z, observed 04:19:44Z–04:21:14Z).
  Raw: `project/results/e03/p6-backoff.txt`, `p6b-backoff-controls.txt`. Prediction source: `internal/jobs/jobs.go:17-22`
  (`Duration(base)*Second*Duration(1<<(attempt-1))`, linear `Duration(base*attempt)*Second`), `internal/jobs/nack.go:22-26`.

## Minimal reproduction
1. `POST /v1/queues -d '{"name":"e03-bodef","max_attempts":40}'` (defaults: exponential, base 5).
2. Register a worker (`POST /v1/workers {"name":"p","kinds":["echo"]}`), enqueue an echo job.
3. Repeat 31 times: lease, then `nack {"lease_token":…,"error":"x","delay_seconds":0}` (delay override only to avoid waiting;
   it does not change the attempt counter).
4. Lease once more and nack **without** `delay_seconds` (attempt 32 uses the queue backoff). Read `state`, `run_at`.
Faster variant: queue `{"max_attempts":10,"backoff":{"kind":"exponential","base_seconds":2147483647}}`, 3 zero-delay nacks,
4th default nack.

## Expected (oracle)
SPEC:27-28 "Failed jobs retry with exponential backoff from `base_seconds` or linear backoff when configured." A backoff
wait is nonnegative and nondecreasing in the attempt number; a configuration whose wait cannot be represented must be capped
or rejected at configuration time (SPEC:17), never turned into a retry-now.

## Observed
```text
e03-bodef job 845 attempt 31 default nack → retry_wait, run_at 2196-12-07 00:26:03 (5·2^30 s ≈ 170 y — honoured, control)
e03-bodef job 846 attempt 32 default nack → retry_wait, run_at 1782-07-04 20:57:20 → promoted to queued on the next tick
e03-boexp (exp, base 2147483647) job 840 attempt 4 → run_at 1986-08-30 06:38:09 → queued 2 s later
e03-bolin (linear, base 2147483647) job 841 attempt 5 → run_at 1782-07-04 20:55:55 → queued 2 s later
```

## Variables isolated
- Kind: both exponential and linear overflow; base × attempt determines the first wrapped attempt.
- Configuration: the defaults overflow at attempt 32, so any queue or job with `max_attempts ≥ 32` (accepted without limit,
  up to 2 147 483 647) reaches it.
- By arithmetic (source inference, not probed): exponential attempt 33 wraps to ≈+96 y, and from attempt 65 the shift
  `uint64(1)<<64` is 0, so the wait becomes 0 — behaviour alternates rather than saturating.
- Not involved: `backoff.max_seconds` (stored but unused, see [E03-F07](E03-F07-backoff-max-seconds-ignored.md)); a cap
  that worked would have prevented this.

## Sibling check
- `nack.go:23-24` explicit `delay_seconds` override and heartbeat extension: same overflow, reported in
  [E03-F03](E03-F03-duration-overflow-delay-and-extension.md).
- `expire.go` lease-expiry retry sets state `queued` with no backoff (by design per SPEC:88-89; not this defect).

## Severity / status
Medium: long-retry configurations silently degrade into a hot retry loop (or a centuries-long stall) exactly at the attempt
counts operators choose for "keep retrying with backoff". Status: open.
