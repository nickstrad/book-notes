# E02-F02 — One unparsable webhook URL stops notification delivery for every queue and namespace

- Seen at: `explore-it/lesson-01` plain export; stack `queuelet-e02`, image `queuelet:e02`; fixture `lab/reset.sh` 03:23:12Z
  (blocking run) and 03:30:14Z (variable isolation).
- Original observation: [E02 notes](../experiments/02-make-it-visible/notes.md) E02-H11 (stated 03:28Z; observed 03:28:39Z–03:29:58Z),
  E02-H11a–d. Source of the prediction: E01 backlog H44 + `internal/webhook/ship.go:26-28`.

## Minimal reproduction
1. Any namespace (here globex-key): `POST /v1/queues {"name":"e02bad","webhook":{"url":"http://bad host/x","secret":"s"}}` → 201.
2. Create a worker, enqueue an echo job, lease it, ack it (job 305) → 200; outbox row 71 is created with that URL.
3. In another namespace (acme), complete a job on a webhook queue: `POST /v1/queues/reports/jobs {"kind":"echo",…}` → 306,
   acked by worker-b.
4. Watch `webhook_outbox`, hooksink `/admin/records`, `/health`, api log for 20 s.

## Expected (oracle)
SPEC:121 "Failed notifications are retried. Delivery is at least once." SPEC:10 "A namespace key accesses only its
namespace's resources" (tenant isolation, approximation: one tenant's configuration must not deny service to another).
SPEC:17 commands with invalid values "should be rejected as invalid commands" (a URL that can never be requested).

## Observed
```text
create/PATCH with "http://bad host/x" → 201 / 200 (no validation, queues.go:21-41, :79-84)
outbox 71 (job 305, globex) attempts 0 undelivered; outbox 72 (job 306, acme) attempts 0 undelivered for 24 s (20/20 samples)
hooksink records: 0
/health: 503 {"status":"degraded"} (cause not stated)
api log, every second: webhook task: parse "http://bad host/x": invalid character " " in host name
```
`Ship` selects due rows `ORDER BY id LIMIT 20` and `return e` on `http.NewRequest` error before updating attempts/next_at
(`ship.go:17-28`), so the same row is first on every tick and nothing after it is attempted.

## Variables isolated
- Recovery via the API does not exist: PATCH the queue to a valid URL → still blocked (outbox copies the URL at Notify,
  `jobs.go:80`); DELETE the job (204) → orphan outbox row 71 remains and still blocks (H11a/b, 03:29:42–03:29:58Z).
- Failure class matters: `ftp://x/y` (parses, fails in `Do`) is a per-row failure, attempts+1 every 5 s, does **not** block
  (H11c, 03:30:46–03:30:59Z). Only NewRequest-time errors block.
- Namespace does not matter (cross-tenant shown). Hooksink mode irrelevant (ok throughout).
- Lab reset (TRUNCATE) cleared it; no restart was tried (row persists in DB, so restart is predicted not to help — not probed).

## Sibling check
- `queues.go:79-84` PATCH accepts the same URL (observed 200, H11d) — same entry point to the blocking state.
- Other `return e` inside the Ship row loop: the two `UPDATE webhook_outbox` errors (`ship.go:37-40`, `:46-49`) — DB errors,
  would also abort the batch (not probed). `secret` type assertion `r["secret"].(string)` (`ship.go:23`) would panic on NULL;
  Notify always writes text (inference).
- Outbox rows survive job deletion (no FK, migrations/001.sql:10) — orphan source (C10 sibling).

## Severity / status
High: any tenant can, by mistake or on purpose, silently stop all webhook notifications for all tenants until an operator
edits the database; health shows only "degraded". Status: open.
