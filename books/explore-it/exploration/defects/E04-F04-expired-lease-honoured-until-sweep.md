# E04-F04 — An expired lease is still honoured until the next sweep: heartbeat resurrects it, ack/nack complete it

- Seen at: `explore-it/lesson-03` plain export; stack `queuelet-e04`; lab clock frozen (2031-06-01, 2031-07-01, 2031-12-01);
  queue lease 30 s; worker 3.
- Original observation: [E04 notes](../experiments/04-judge-it/notes.md) P6 (12:47:17Z), control P7 (12:47:47Z), sibling
  P11 (12:49:59Z); hypothesis R39 stated 12:44Z. Raw: `project/results/e04/p6-expiry.txt`, `p7-grace-control.txt`,
  `p11-siblings.txt`. Prediction source (code): `internal/jobs/jobs.go:96-98` (Heartbeat guard `lease_token AND
  state='leased'`), `internal/jobs/ack.go:10`, `internal/jobs/nack.go:14` — none compares `lease_expires_at` with now; only
  `expire.go:12` (sweep, once per real second, `api.go:195-210`) does.

## Minimal reproduction
1. Queue `{"name":"e04e","lease_seconds":30}`; enqueue an echo job; lease it (note token, `lease_expires_at` = T+30).
2. Admin: `POST /admin/clock {"action":"advance","seconds":31}`.
3. Immediately (within the same second): `POST /v1/jobs/{id}/heartbeat {"lease_token":…,"extend_seconds":30}`.
4. Control: repeat 1–2, sleep 1.5 s, then step 3.

## Expected (oracle)
SPEC Worker: "Only the current lease holder may complete or extend a job" and "An expired lease records an unsuccessful
attempt and releases the job for retry". The lease response publishes `lease_expires_at` as the lease's end. Rule R39: after
`lease_expires_at` the worker is no longer the current holder, so extend/complete should get 409, and the expiry attempt
should be recorded — deterministically, not depending on when a background tick ran. (Comparable: SQS documents the
visibility timeout as the in-flight window; whether a late DeleteMessage can still succeed there is not verified in this run.
Spec silence on any grace period — the owner may choose one, but then it should be stated and bounded.)

## Observed
```text
before the sweep (immediate):  heartbeat → 200; job stays leased, lease_expires_at 00:01:01; events: created, leased,
                               heartbeat — no lease_expired, no attempt row        (5/5 repeats → 200)
                               ack (other job) → 200 succeeded, attempts 1
                               nack (sibling) → 200; attempt recorded as "failed: late nack", not "expired"
after one sweep (sleep 1.5 s): heartbeat → 409 state_conflict; job queued, attempts+1, last_error "lease expired"
```

## Variables isolated
- The only variable that changes the outcome is whether the 1-s Expire tick has run since the expiry instant (control P7).
- Token, extension size, job and worker were the same between the 200 and 409 cases.
- The lab clock only makes the window easy to hit; in real time the window is up to one loop period, and longer whenever the
  core loop is slow or degraded (it runs Promote, Expire, Fire, Retain serially, `api.go:204`).
- Not a double-lease: no other worker can lease the job until the sweep sets it `queued` (R31 held), so exclusivity holds.

## Sibling check
- `ack.go:10`, `nack.go:14`, `jobs.go:97` — all three confirmed by probe.
- `Cancel` (`operator.go:5-6`) is not lease-bound — n/a. Worker deletion releases regardless of expiry — n/a.

## Severity / status
Low–medium: a worker that stalled past its lease can extend it and its expired attempt is never recorded (attempt accounting
and the "expired" history depend on timing); results are nondeterministic for identical requests. Status: open (owner may
instead specify a grace window).
