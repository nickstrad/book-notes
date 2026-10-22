# F27 — Schedules freeze the UTC offset at creation; DST changes are ignored

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H25b (01:52:02Z–01:52:04Z).
- **Evidence route:** running-service state (stored `zone_offset`) + the SUT's own `schedule.Next` executed in a scratch harness (project/results/dst/main.go). Not an end-to-end fire across a DST boundary: the lab has no clock control (testability request TR2).

## Minimal reproduction
1. Seed fixture schedule 1: `daily-report`, cron `0 9 * * *`, timezone America/New_York. DB: `zone_offset = -14400`, next_run_at 2026-10-22 13:00Z.
2. `go run -buildvcs=false ./results/dst` → calls `schedule.Next("0 9 * * *","America/New_York",-14400, 2026-11-02T00:00Z)` and `schedule.Offset` at that date.

## Expected (oracle)
SPEC: "Schedule times follow the named zone's local civil time, including offset changes." After DST ends (2026-11-01), 09:00 EST = 14:00Z.

## Observed
```
stored-offset next: 2026-11-02 13:00:00 +0000 UTC = 08:00 EST | offset now would be -18000
1|daily-report|America/New_York|-14400|2026-10-22 13:00:00+00
```

## Variables isolated
Fire always passes the stored zone_offset (fire.go:48) and Next builds a FixedZone (fire.go:17). PATCH timezone does not update zone_offset either (F07).

## Sibling check
Create computes offset at "now" (schedules.go:33) — a schedule created in winter is wrong all summer. No other zone handling.

## Severity / status
Medium: every non-UTC schedule drifts by an hour twice a year. Status: open (end-to-end confirmation pending TR2).
