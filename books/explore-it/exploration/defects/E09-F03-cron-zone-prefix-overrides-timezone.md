# E09-F03 — A `CRON_TZ=`/`TZ=` prefix in `cron` is accepted and overrides the schedule's `timezone`

Seen at: `explore-it/lesson-08` (stack `queuelet-e09`), experiment 09.
Original observations: [E09 notes](../experiments/09-explain-it/notes.md) H-E09-07 (20:37:47Z), H-E09-07b and firing
check (20:38:11Z–20:38:30Z). Predicted before probing (PD3) from fire.go:12/:23 plus my recollection that robfig/cron
accepts the prefix (library behaviour, confirmed only by this run).

## Minimal reproduction

Clock at 2026-10-22T21:17:30Z, then 2026-11-01T05:30:00Z.

```sh
tools/q.sh acme-key POST /v1/schedules '{"name":"crontz","queue":"imports","kind":"echo","cron":"CRON_TZ=Asia/Tokyo 0 9 * * *","timezone":"UTC"}'
tools/q.sh acme-key POST /v1/schedules '{"name":"crontz-ny","queue":"imports","kind":"echo","cron":"CRON_TZ=Asia/Tokyo 0 9 * * *","timezone":"America/New_York"}'
tools/q.sh acme-key POST /v1/schedules '{"name":"tz-plain","queue":"imports","kind":"echo","cron":"TZ=Asia/Tokyo 0 9 * * *","timezone":"UTC"}'
```

## Expected (named oracle)

SPEC.md:160 "Cron uses five fields"; SPEC.md:161-162 zones are IANA identifiers and schedule times follow the
*named zone's* local time. A six-token cron is an invalid command (400), or at least the `timezone` member governs.

## Observed

```text
crontz    (timezone UTC)              → 201 next_run_at 2026-10-23T00:00:00Z  (= 09:00 Tokyo, not 09:00 UTC)
                                        fired: job 308 run_at 2026-10-23 00:00Z; next 2026-11-02 00:00Z
crontz-ny (timezone America/New_York) → 201 next_run_at 2026-11-02T05:00:00Z  (= 00:00 New York = 14:00 Tokyo)
tz-plain  (TZ= variant, UTC)          → 201 next_run_at 2026-11-02T00:00:00Z
```

GET of these schedules shows `"timezone":"UTC"` / `"America/New_York"` while they fire on another clock; with a
non-UTC member the instant matches neither zone's 09:00.

## Variables isolated

Prefix spelling (`CRON_TZ=`, `TZ=`) does not matter. The `timezone` member matters: UTC → embedded zone wins;
non-UTC → mixture. Firing uses the same `Next` (observed job 308).

## Mechanism (reading)

The raw string goes to robfig's parser (fire.go:12, :15), which accepts a zone prefix. The fixed/wildcard test reads
`strings.Fields(expression)[1]` (fire.go:23) — the minute field when a prefix is present — and the fixed-hour path
builds a fake-UTC wall time (fire.go:29-30) that robfig then converts into its own zone.

## Sibling check

Other places the cron string is interpreted: create (api/schedules.go:264), PATCH (:341), fireOne (fire.go:102) —
all go through `Next`, all affected. Other robfig extensions: descriptors (`@hourly`) are refused by the parser flags
(fire.go:12) — not probed.

## Severity and status

Medium: schedules silently run at hours that the stored record contradicts. Status: open.
