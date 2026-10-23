# E09-F02 — A cron that never matches is accepted and reported as due at 0001-01-01

Seen at: `explore-it/lesson-08` (stack `queuelet-e09`), experiment 09.
Original observations: [E09 notes](../experiments/09-explain-it/notes.md) H-E09-01 (20:36:29Z), H-E09-04, H-E09-05
(20:37:16Z). Predicted before probing (PD1a, PD2).

## Minimal reproduction

```sh
tools/q.sh acme-key POST /v1/schedules '{"name":"feb30","queue":"imports","kind":"echo","cron":"0 0 30 2 *","timezone":"UTC"}'
tools/q.sh acme-key POST /v1/schedules '{"name":"feb30w","queue":"reports","kind":"echo","cron":"* * 30 2 *"}'
tools/q.sh globex-key PATCH /v1/schedules/4 '{"cron":"0 0 31 4 *"}'
```

## Expected (named oracle)

- SPEC.md:19 invalid values "should be rejected as invalid commands" (approximation for an impossible date); SPEC.md:170-171 "the next fire must reflect the current
  configuration". A schedule that can never fire has no next fire; reporting one in year 1 is false. Comparable
  system: classic cron accepts such a line but never runs it — it never reports it as overdue.
- Acceptable outcomes: 400 `invalid_cron`, or acceptance with no next fire (null) that the scheduler ignores.

## Observed

```text
POST feb30   → 201 … "next_run_at":"0001-01-01T00:00:00Z"
POST feb30w  → 201 … "next_run_at":"0001-01-01T00:00:00Z"
PATCH 4 (April 31) → 200 … "next_run_at":"0001-01-01T00:00:00Z"
```

Consequence: the scheduler then fails on it every tick (E09-F01).

## Variables isolated

Both `Next` branches (fixed-hour and wildcard-hour), create and PATCH, two impossible dates (Feb 30, Apr 31).
Timezone irrelevant (UTC). Disabled at create → next_run_at null, harmless until enabled.

## Mechanism (reading)

`schedule.Next` returns `(zero, nil)` when robfig finds no match (fire.go:25, :32-33); create and PATCH store the
result unchecked (api/schedules.go:264-282, :341-350). The fireOne check `!next.After(now)` (fire.go:106) exists only
in the scheduler.

## Sibling check

Same habit (validation in one writer, not another): E06-F02 (schedule payload), F09 (PATCH cron). Other
values that parse but cannot be used: `CRON_TZ=` prefix (E09-F03), non-IANA zone (E09-F09). None other found in
schedule input (name/kind are free text in SPEC).

## Severity and status

Medium on its own (false console value), high through E09-F01. Status: open.
