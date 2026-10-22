# E02-F03 — A partner slower than 2 s receives the same notification forever; queuelet and partner disagree on outcome

- Seen at: `explore-it/lesson-01` plain export; stack `queuelet-e02`; fixture `lab/reset.sh` 03:30:14Z.
- Original observation: [E02 notes](../experiments/02-make-it-visible/notes.md) E02-H12 (stated 03:28Z; observed 03:31:23Z–03:31:51Z);
  related E02-H05/H07 (500 mode, 03:26:10Z–03:27:09Z) and H11c (ftp row).

## Minimal reproduction
1. `POST :18231/admin/mode {"mode":"slow"}` (hooksink sleeps 3 s, then answers 200).
2. `POST /v1/queues/reports/jobs {"kind":"echo","payload":{"probe":"e02-slow"},"priority":9}` → 303 (worker-b acks).
3. Watch outbox, hooksink records and api log for 22 s; then mode ok.

## Expected (oracle)
SPEC:120-121 "Delivery is successful only on a 2xx response. Failed notifications are retried. Delivery is at least once."
The partner did answer 2xx; the spec defines no timeout, retry cap or backoff. Comparable practice (approximation): webhook
senders bound retries with growing backoff and give up/mark failure visibly. **Spec silence** — hence status `spec`.

## Observed
```text
hooksink: 03:31:24.273 200 | 29.278 200 | 34.274 200 | 40.275 200 | 45.274 200   (5 receipts, all 2xx)
outbox 72: attempts 5, delivered_at NULL at 03:31:48; api log: nothing; /health ok 22/22
after mode ok: 6th POST delivered 03:31:51.269; api log "webhook delivered status=200"
```
Client timeout is fixed at 2 s (`ship.go:31`); failure path adds 5 s with no cap and no log (`ship.go:32-41`).

## Variables isolated
- Partner status irrelevant to the duplication: in 500 mode (H05/H07) 11 failures at a constant 5–6 s; ftp URL (H11c)
  also retried every 5 s indefinitely. Only latency > 2 s turns a partner *success* into a queuelet failure.
- Signature valid on every receipt; body identical across receipts (same outbox row).
- Not probed: partner at exactly 2 s; a long run to show "forever" (inference from code: no cap).

## Sibling check
Single delivery path (`internal/webhook/ship.go`). The worker `http` kind uses a 5 s timeout (`worker.go:98`), unrelated.
No attempt cap or dead-letter state exists for outbox rows (`migrations/001.sql:10` has only attempts/next_at/delivered_at).

## Severity / status
Medium: a consistently slow partner receives unbounded duplicates while queuelet reports zero successes and logs nothing;
operators cannot see the backlog without SQL. Within "at least once", so the coordinator must decide the timeout/cap.
Status: spec.
