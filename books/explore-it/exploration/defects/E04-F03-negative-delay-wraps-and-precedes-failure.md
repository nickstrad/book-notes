# E04-F03 — Negative `delay_seconds` is unchecked: a nack retry is scheduled before the failure, and large negatives wrap 292 years into the future

- Seen at: `explore-it/lesson-03` plain export; stack `queuelet-e04`; lab clock frozen at 2030-01-01T00:00:00Z; worker 3.
- Original observation: [E04 notes](../experiments/04-judge-it/notes.md) P3 (12:45:57Z), hypothesis R35 stated 12:44Z.
  Raw: `project/results/e04/p3-negative-delay.txt`. Prediction source (code): `internal/api/jobs.go:67-70` and
  `internal/api/enqueue.go:54-58` check only `delay_seconds > MaxSeconds`; `internal/jobs/nack.go:23-25` and
  `enqueue.go:58` compute `time.Duration(*delay) * time.Second`, which overflows int64 nanoseconds below −9 223 372 036 s.

## Minimal reproduction
1. Queue `e04b` (any settings, max_attempts > 1); worker with kind echo; enqueue an echo job; lease it.
2. `POST /v1/jobs/{id}/nack -d '{"lease_token":"…","error":"x","delay_seconds":-9223372037}'`; read `run_at`.
3. Enqueue-side: `POST /v1/queues/e04b/jobs -d '{"kind":"echo","payload":{},"delay_seconds":-9223372037}'`.

## Expected (oracle)
- SPEC Queues (lesson 03 line): "A computed wait never overflows … so a retry is never scheduled before the failure." Rule
  R35: a retry comes after the failure, within a bounded delay. The override is the same "wait"; the spec gives it no
  separate meaning.
- Monotonicity (approximation): a smaller (more negative) delay must never produce a *later* time than a larger one. Whatever
  the owner decides for negatives (reject with 400, or clamp to "now"), wrapping to the far future is wrong.
- SPEC "Every seconds-valued field … accepts at most 2147483647" names only the upper bound; the lower bound is a spec
  silence (E03 backlog Q-E03-2).

## Observed
```text
nack delay -5            → 200; retry_wait run_at 2029-12-31 23:59:55   (5 s before the failure at 00:00:00)
nack delay -9223372036   → 200; run_at 1737-09-22 00:12:44
nack delay -9223372037   → 200; run_at 2322-04-12 23:47:16.709552      (≈ +292 years)
nack delay -18446744073  → 200; run_at 2030-01-01 00:00:00.709552      (wrapped by 2^64 ns to +0.71 s)
enqueue delay -5          → 201 queued, run_at 23:59:55
enqueue delay -9223372037 → 201 {"state":"scheduled"}, run_at 2322-04-12 23:47:16.709552
```

## Variables isolated
- Sign and magnitude only: same queue, same worker, same frozen clock. Above −9 223 372 036 s the value lands in the past
  (job runs at once); at −9 223 372 037 it wraps forward; at −2^64/1e9 it wraps to ~now again.
- Path: nack override and enqueue `delay_seconds` both affected. Heartbeat `extend_seconds` is not (lower bound 1 is checked,
  `internal/api/jobs.go:77`).
- Did not matter: retry flag default, backoff kind (the override replaces the computed wait).

## Sibling check
- Heartbeat `extend_seconds`: guarded both sides (`jobs.go:77`) — not affected.
- `run_at` (RFC 3339) on enqueue: no range check; a far-past/far-future timestamp is stored as given — no arithmetic, so no
  wrap (not probed).
- Lab clock `advance`: checks both signs (`internal/clock/clock.go:41`) — not affected.
- This is the one-sided remainder of E03-F03/U13 (the fix added only the upper bound).

## Severity / status
Medium: one malformed negative value from a worker or producer silently parks a job for 292 years (or retries it
immediately, before the failure it reports) with a 2xx; nothing in stats distinguishes it from a normal retry. Status: open.
