# F07 — Schedule PATCH does not recompute the next fire; disabled schedules keep firing

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observations: [notes.md](../experiments/01-map-it/notes.md) H09 (01:44:21Z), H09b/H09c (01:44:21Z, outcome 01:45:05Z), H38 (01:50:29Z).

## Minimal reproduction
Actor acme-key, queue imports, host clock UTC ≈ 01:44:20.
A. `POST /v1/schedules {"name":"h09","queue":"imports","kind":"h09kind","cron":"0 0 1 1 *"}` → id 5; `PATCH /v1/schedules/5 {"cron":"* * * * *","timezone":"Asia/Tokyo"}`.
B. Create `* * * * *` schedule 6 (kind h09b); `PATCH {"enabled":false}`; wait for the next minute.
C. Create `* * * * *` schedule 7 with `"enabled":false`; `PATCH {"enabled":true}`; wait for the next minute.

## Expected (oracle)
SPEC: "PATCH … updates cron, timezone or enabled status; the next fire must reflect the current configuration." and "disabled schedules do not enqueue jobs."

## Observed
```
A: {"cron":"* * * * *","timezone":"Asia/Tokyo","zone_offset":0,"next_run_at":"2027-01-01T00:00:00Z"}
B: PATCH → {"enabled":false,"next_run_at":"2026-10-22T01:45:00Z"}; after 01:45: 6|h09b|f|01:46:00|last 01:45:00 ; job kind h09b queued in imports (1)
C: PATCH → {"enabled":true,"next_run_at":null}; after 01:45: 7|h09c|t|(null)|(null), no h09c jobs
```
A every-minute schedule would not fire until next New Year; a disabled schedule enqueued a job and advanced; an enabled one never fires.

## Variables isolated
Each of cron, timezone, enabled tested separately in A/B/C. Invalid timezone on PATCH is also accepted (`Mars/Base` → 200, H38) while create rejects it (400) — validation gap shared with F09.

## Sibling check
internal/api/schedules.go:84 updates columns only; schedule/fire.go:28 selects on next_run_at without `enabled`. Create path (schedules.go:52-56) computes next and nulls it when disabled — consistent with the observations. Queue PATCH (queues.go:82) does not touch schedules.

## Severity / status
High: schedule edits silently do nothing or the opposite; disabling cannot stop a runaway schedule (only DELETE does). Status: open.
