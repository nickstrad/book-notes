# E08-F02 — Twenty slow or unreachable webhook rows starve every later notification, for every tenant, indefinitely

Seen at: export of `explore-it/lesson-07` (stack `queuelet-e08`), experiment 08 (surround-it).
Original observations: [E08 notes](../experiments/08-surround-it/notes.md) P5 (19:27:28Z, H-E08-08), P5b
(19:29:21–19:30:37Z, H-E08-15/16/17), P5c (19:30:57Z, H-E08-18).

## Minimal reproduction

Actors: tenant globex (`globex-key`) and tenant acme (`acme-key`). Clock: real. Lab partner: hooksink, or no
partner at all.

```sh
cd books/explore-it/exploration/experiments/08-surround-it/tools
. ./env.sh
# drain first: hooksink mode ok and webhook_outbox has no undelivered rows
python3 p5b_isolate.py ok 20 http://172.28.0.250/x 45      # tenant-only cause (unused address on the lab network)
python3 p5_partner_modes.py timeout 22 60                   # partner-timeout cause (hooksink mode timeout)
```

Steps (blackhole variant): globex creates a queue whose webhook URL is a well-formed `http://` URL to a host that
does not answer within 2 s (`http://172.28.0.250/x`); globex completes 20 jobs (enqueue, lease, ack). acme creates a
queue whose webhook answers 200 at once (`http://hooksink:8201/admin/records`) and completes one job. Watch the acme
outbox row with read-only SQL (`attempts`, `delivered_at`).

## Expected (named oracle)

- SPEC.md:179 "Failed notifications are retried. Delivery is at least once." — the acme notification has a healthy
  receiver and must eventually be attempted and delivered.
- SPEC.md:197-198 "A malformed legacy row is counted as a failed attempt and rescheduled independently, so it does
  not stop later rows." — the stated design intent: one bad row must not stop later rows.
- SPEC.md:10 tenant isolation — one namespace's configuration should not deny another namespace its service
  (approximation: an acme row to a healthy receiver is attempted within one delivery pass, bound ≤ 45 s).

## Observed

```text
P5 (hooksink timeout, 22 globex rows 74..95, acme row 96):
19:27:39 … 19:28:33 acme 96|0|-1|f   globex attempts cycle over rows 74..93 only; 94, 95 never attempted
19:28:33 mode ok → 19:28:35 acme delivered 1.5 s after mode ok
P5b blackhole (hooksink healthy): after 45.6 s acme: 146|0|-1|f|
   globex attempts 3,3,3,3,2,2,1,…,1  err=Post "http://172.28.0.250/x": dial tcp …: connect: no route to host
                                       / context deadline exceeded (Client.Timeout exceeded while awaiting headers)
P5c: globex PATCH webhook url "" → 200; DELETE its 20 jobs → 204; DELETE queue → 204
19:31:03 … 19:31:43 acme 146|0|f | globex pending 20, max attempts 4→5 | /health 200
```

The acme notification was never attempted. Nothing either tenant can do through the API ends it, and `/health`
reports ok throughout.

## Variables isolated

- Matters: **slowness** of earlier rows, not failure. 22 rows failing fast (mode 500) → acme delivered after 1.4 s
  (H-E08-15).
- Matters: **count** ≥ 20. With 5 timing-out rows the acme row is reached in the first pass and delivered after
  11.1 s ≈ 5 × 2 s (H-E08-16): a delay, not starvation.
- Does not matter: the partner. A tenant's own unreachable URL is enough (H-E08-17); create/PATCH validation accepts
  it (SPEC.md:194 only requires a well-formed HTTP(S) URL).
- Does not matter: the tenant boundary itself — the outbox is one global `ORDER BY id` list, so the same happens
  between two queues of one tenant; the cross-tenant case is simply the most harmful.
- Not affected: core lifecycle (delayed jobs promoted on time during P5 and P5 slow) — U03 holds.
- Mechanism (source reading, consistent with every outcome): `internal/webhook/ship.go:17` selects the 20 lowest due
  ids; each slow row costs the 2 s client deadline, so a pass takes ≈40 s; a failure sets `next_at = now + 5 s`
  where `now` is the tick time taken before the pass (`internal/api/api.go:212`, ship.go:45), so all 20 are due again
  when the next pass starts and the same 20 are selected forever. Rows have no attempt cap (SPEC.md:204), snapshots
  ignore PATCH and job/queue deletion (SPEC.md:196-197), so the stall is permanent.

## Sibling check

- `schedule.Fire` returns on the first failing schedule (`internal/schedule/fire.go:37-40`): one schedule that cannot
  advance would starve later ones every tick. Predicted, not probed — no external trigger found (cron and timezone
  are validated at create/PATCH). `#offcharter`.
- `Expire`/`Promote`/`Retain` run in the core loop with bounded SQL statements, not network calls; none found.
- Event console (`internal/api/observe.go:36`) reads 500 rows per 200 ms and does no network I/O; none found.

## Related prior records

- [F13](F13-slow-webhook-stalls-background-loop.md) (E01) inferred the same arithmetic ("20 rows × 2 s … rows re-qualify
  every 5 s") for the *core loop*; that stall was removed by the separate shipper goroutine (U03 holds here: delayed
  jobs promoted on time). The starvation of other rows inside the shipper is what remains.
- [E02-F02](E02-F02-unparsable-webhook-url-blocks-all-notifications.md) (fixed as U10): one *malformed* row blocked
  all later rows; the fix reschedules malformed rows independently. Slow/unreachable well-formed rows are the
  remaining path to the same outcome.

## Severity and status

High: any tenant can, by configuration alone and 20 completed jobs, halt webhook delivery for every namespace
permanently. No API action ends it and no health/metric signal shows it (TR-E08-2). Partner outages that time out
(the commonest kind) do the same for as long as they last.

Status: open.
