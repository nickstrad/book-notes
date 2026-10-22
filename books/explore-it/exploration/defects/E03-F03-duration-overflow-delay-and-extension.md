# E03-F03 — `delay_seconds` beyond ≈292 years wraps around: a requested far-future job runs immediately (also nack delay and heartbeat extension)

- Seen at: `explore-it/lesson-02` plain export; stack `queuelet-e03`; real clock; fixture seed 04:16:24Z.
- Original observation: [E03 notes](../experiments/03-vary-it/notes.md) E03-H04 (stated 04:16Z, observed 04:17:32Z); siblings
  E03-H17, E03-H18 (stated 04:24Z, observed 04:24:21Z). Raw: `project/results/e03/p3-numbers.txt`, `p10-siblings.txt`.
  Prediction source: `internal/api/enqueue.go:53-55` (`time.Duration(*in.DelaySeconds) * time.Second`, no range check).

## Minimal reproduction
1. `POST /v1/queues/imports/jobs -d '{"kind":"echo","delay_seconds":9223372037,"payload":{}}'` (acme-key).
2. Compare with `delay_seconds: 9223372036`.
3. Read `state`/`run_at`.

## Expected (oracle)
SPEC:48-49 "Choose either delay in seconds … A job due now is `queued`; a future job is `scheduled`." A delay of
9 223 372 037 s is in the future; the service must either schedule it at now + delay or reject a delay it cannot represent
(SPEC:17 invalid command → 400). It must never run it now.

## Observed
```text
delay 9223372036          → 201 scheduled, run_at 2319-02-01 04:04:49 (honoured)
delay 9223372037          → 201 queued,    run_at 1734-07-13 04:30:16 (wrapped ≈292 y into the past)
delay 100000000000        → 201 scheduled, run_at 2272-11-28 16:11:24 (asked ≈3169 y; a different, wrapped future)
delay 9223372036854775807 → 201 queued,    run_at now − 1 s
delay 9223372036854775808 → 400 invalid_json (does not fit Go int)
```

## Variables isolated
- Boundary is exactly the Go `time.Duration` range (2^63 − 1 ns ≈ 9 223 372 036.85 s); 9 223 372 036 is honoured, +1 wraps.
- Outcome above the boundary depends on the value modulo 2^64 ns: past (runs now) or an arbitrary wrong future.
- Not affected: `run_at` (parsed as an absolute time; 9999-12-31 → scheduled correctly); negative delay (queued; SPEC silent, not claimed).

## Sibling check (predicted before probing, then observed)
- `internal/jobs/nack.go:23-24` nack `delay_seconds` override: 9223372037 → 200, `retry_wait` run_at 1734-07-13, queued 1.5 s
  later (E03-H17, job 1365) — a "retry much later" request becomes an immediate retry.
- `internal/jobs/jobs.go:84` heartbeat `extend_seconds`: 9223372037 → 200 `{"ok":true}`, lease_expires_at 1734-07-13; 1.5 s
  later the job was `queued`, attempts 1, last_error "lease expired" (E03-H18, job 1366) — asking for a longer lease ends it
  and consumes an attempt.
- `internal/jobs/jobs.go:17-22` queue backoff — same arithmetic family, separate report [E03-F06](E03-F06-backoff-overflow-immediate-retry.md).
- `jobs.go:61,65` lease expiry `Duration(sec)*Second`: `lease_seconds` is an `int` column (≤ 2 147 483 647), so it cannot
  overflow (inference; 2147483647 was accepted, expiry not observed).
- `internal/clock/clock.go:40-42` explicitly range-checks its `advance` — the one place that guards this arithmetic.

## Severity / status
Medium: a caller's explicit "not before" intent is inverted — work runs ~immediately and a heartbeat can kill its own lease.
Inputs are unusual but accepted with 2xx and no indication. Status: open.
