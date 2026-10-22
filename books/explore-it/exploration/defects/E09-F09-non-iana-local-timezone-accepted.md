# E09-F09 — Schedule timezone `Local` (not an IANA identifier) is accepted

Seen at: `explore-it/lesson-08` (stack `queuelet-e09`), experiment 09.
Original observation: [E09 notes](../experiments/09-explain-it/notes.md) H-E09-08 (20:37:47Z), firing check 20:38:30Z.
Predicted before probing (PD4).

## Minimal reproduction

```sh
tools/q.sh acme-key POST /v1/schedules '{"name":"local","queue":"imports","kind":"echo","cron":"0 9 * * *","timezone":"Local"}'
```

## Expected (named oracle)

SPEC.md:161 "Time zones use IANA identifiers, defaulting to UTC." `Local` is a Go alias for the server process's
zone, not an IANA identifier → 400 `invalid_timezone`.

## Observed

```text
201 {"cron":"0 9 * * *",…,"next_run_at":"2026-10-23T09:00:00Z",…,"timezone":"Local","zone_offset":0}
fired: job 309 run_at 2026-10-23 09:00Z
```

## Variables isolated

The API container's zone is UTC, so today it behaves like UTC; its meaning would change with the server's TZ.
Not varied: PATCH (same `schedule.Offset`/`Next` path, api/schedules.go:336-341 — predicted, not probed).

## Mechanism (reading)

`time.LoadLocation(zone)` (fire.go:19, :60) accepts `Local` and `""`. Create maps `""` to UTC
(api/schedules.go:245-247); PATCH does not (stores `""`; reading, not probed).

## Sibling check

Same habit as E09-F03 (input whose meaning comes from the runtime/library rather than the spec). None else found.

## Severity and status

Low. Status: open.
