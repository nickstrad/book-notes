# F24 — Resuming a paused queue enqueues every missed schedule tick at once

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H23 (setup 01:44:20Z, resume 01:48:04Z–01:48:06Z).

## Minimal reproduction
1. Queue h23; `POST /v1/queues/h23/pause`.
2. `POST /v1/schedules {"name":"h23-every-minute","queue":"h23","kind":"echo","cron":"* * * * *"}` (next 01:45:00).
3. Wait until after 01:48:00; `POST /v1/queues/h23/resume`; wait 2 s.

## Expected (oracle)
SPEC: "POST …/pause prevents new leases and scheduled deliveries." "Missed periodic ticks should not create an unbounded burst of catch-up jobs."

## Observed
```
879|run_at 01:45:00|created 01:48:05|queued
880|01:46:00|01:48:05
881|01:47:00|01:48:05
882|01:48:00|01:48:05
schedule next_run_at 01:49:00, last_run_at 01:48:00
```
One job per missed tick, all in one loop tick. Burst size grows linearly with pause length (inferred: a day paused on `* * * * *` → 1440 jobs).

## Variables isolated
Pause duration (3+ minutes) and cron frequency set the burst; nothing was enqueued while paused (pause itself works). Mechanism: fire.go:37-38 skips without advancing next_run_at, fire.go:41 loops until next > now.

## Sibling check
Same catch-up after an API outage or a dead loop (F08) — predicted, not probed; after the F08 recovery (01:49:42Z) schedule 4 had been deleted so no burst was visible. Disabled-via-PATCH schedules (F07) keep advancing, so no burst there.

## Severity / status
Medium: a long pause or outage floods the queue and (with F13) can stall the loop. Status: open.
