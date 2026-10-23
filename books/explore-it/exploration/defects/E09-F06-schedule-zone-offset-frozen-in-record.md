# E09-F06 — Schedule record keeps reporting the creation-time `zone_offset` after the zone's offset changes

Seen at: `explore-it/lesson-08` (stack `queuelet-e09`), experiment 09.
Original observation: [E09 notes](../experiments/09-explain-it/notes.md) H-E09-13 (20:40:08Z). Predicted before the
probe (habit "console that reports a value it does not check").

## Minimal reproduction

```sh
tools/q.sh admin-key POST /admin/clock '{"action":"set","at":"2026-11-01T05:30:00Z"}'      # 01:30 EDT
tools/q.sh acme-key POST /v1/schedules '{"name":"nyhourly","queue":"imports","kind":"echo","cron":"0 * * * *","timezone":"America/New_York"}'
tools/q.sh admin-key POST /admin/clock '{"action":"set","at":"2026-11-01T07:10:00Z"}'      # EST
tools/q.sh acme-key GET /v1/schedules/9
```

(In the run the schedule was created disabled and then enabled by PATCH; the same columns are written on both
paths.)

## Expected (named oracle)

SPEC.md:162 "Schedule times follow the named zone's local civil time, including offset changes." Approximation: a
derived field the API returns must agree with the value it is derived from at read time, or not be returned. At
07:10Z New York's offset is −18000.

## Observed

```text
{"cron":"0 * * * *",…,"last_run_at":"2026-11-01T06:00:00Z","next_run_at":"2026-11-01T08:00:00Z",
 "timezone":"America/New_York","zone_offset":-14400}
```

Same on schedule 10 (`0 0-23 * * *`). The scheduler fires correctly (it no longer reads the column); only the record
is wrong.

## Variables isolated

Holds for wildcard-hour and fixed-hour schedules. Firing is unaffected (next_run_at advances by elapsed hours).

## Mechanism (reading)

`zone_offset` is written at create/PATCH (api/schedules.go:259, :282, :350); `Next` ignores its `offset` argument
(fire.go:14) and `fireOne` passes 0 (fire.go:102). The column is the remnant of the F27/D23 design.

## Sibling check

Other stored-but-unread derived values: none found in `schedules` or `jobs` (`event_at`, `finished_at` are written on
each transition).

## Severity and status

Low: misleading console field; an operator or client computing local time from it is wrong by an hour after DST.
Status: open.
