# F09 — Schedule PATCH accepts an unparseable cron (and an unknown time zone)

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observations: [notes.md](../experiments/01-map-it/notes.md) H10 (01:48:56Z), H38 (01:50:29Z).

## Minimal reproduction
`PATCH /v1/schedules/4 {"cron":"bogus"}` (acme-key) on any schedule; `PATCH /v1/schedules/5 {"timezone":"Mars/Base"}`.

## Expected (oracle)
SPEC: "Cron uses five fields"; "Time zones use IANA identifiers"; "Unknown fields and invalid enum values should be rejected as invalid commands." Create already enforces both (create with Mars/Base → 400 invalid_timezone; unit test TestInvalidCron).

## Observed
```
{"id":4,"cron":"bogus","next_run_at":"2026-10-22T01:49:00Z"}  (200)
"timezone":"Mars/Base" HTTP 200
```
At the next fire the stored bogus cron stops the background loop (F08).

## Variables isolated
Create vs PATCH is the variable: create validates (schedules.go:33-42), PATCH does not (schedules.go:84).

## Sibling check
Queue PATCH also skips create's validation for max_attempts/retention (F19). No other update endpoints.

## Severity / status
High (because it enables F08 from a single authenticated request); Medium on its own. Status: open.
