# E08-F03 — Across DST, a fixed-time schedule in the skipped hour does not run that day, and one in the repeated hour runs twice

Seen at: export of `explore-it/lesson-07` (stack `queuelet-e08`), experiment 08 (surround-it).
Original observation: [E08 notes](../experiments/08-surround-it/notes.md) P6 (19:32:55Z, H-E08-10/11/12/13).

## Minimal reproduction

Actor: acme key (schedules), admin key (test clock). Clock: lab test clock (`POST /admin/clock`, `set`), stack started
with `QUEUELET_TEST_CLOCK=1` (compose default). Fixture: an acme queue no worker consumes.

```sh
cd books/explore-it/exploration/experiments/08-surround-it/tools
. ./env.sh && python3 p6_dst.py
```

Gap: clock `set` 2027-03-13T12:00:00Z; create `30 2 * * *` and `0 2 * * *` in `America/New_York`; step the clock to
2027-03-14T06:00:30Z, 07:00:30Z, 13:00:30Z, 2027-03-15T06:00:30Z, 06:30:30Z (2.5 s real time per step); list the
queue's jobs. Overlap: clock `set` 2027-11-07T04:00:00Z; create `30 1 * * *` and `0 1 * * *`; step through 05:00:30Z,
05:30:30Z, 06:00:30Z, 06:30:30Z.

## Expected (named oracle)

SPEC.md:162 "Schedule times follow the named zone's local civil time, including offset changes." It is **silent** on
local times that do not exist (spring forward) or occur twice (fall back). Comparable systems disagree (stated from
the explorer's general knowledge, not verified against sources inside this export): classic cron (Vixie/ISC cron,
cronie) runs a fixed-time job whose time was skipped once right after the change and does not run a fixed-time job
twice in the repeated hour, while the robfig/cron v3 documentation warns that jobs in a skipped interval are not run.
Under the classic-cron oracle, a daily job runs exactly once per civil day.

## Observed

```text
created dst-0230 (30 2) next_run_at 2027-03-15T06:30:00Z   ← 2027-03-14 skipped at creation
created dst-0200 (0 2)  next_run_at 2027-03-15T06:00:00Z
jobs after stepping through 2027-03-14: none from dst-0200/dst-0230 for that civil date;
  dst-0200@03-15T06:00Z, dst-0230@03-15T06:30Z next
fall back 2027-11-07:
  fb-0100@11-07T05:00Z (01:00 EDT)  fb-0100@11-07T06:00Z (01:00 EST)   ← twice
  fb-0130@11-07T05:30Z (01:30 EDT)  fb-0130@11-07T06:30Z (01:30 EST)   ← twice
controls: hourly 0 * * * * → 06:00Z, 07:00Z (01:00 EST → 03:00 EDT), and 05:00Z, 06:00Z, 07:00Z on 11-07 (one per real hour);
          0 9 * * * → 2027-03-13T14:00Z, 2027-03-14T13:00Z (offset change honoured)
```

## Variables isolated

- Matters: a fixed local hour inside the transition (02:xx on spring-forward day; 01:xx on fall-back day).
- Does not matter: wildcard-hour schedules (hourly behaves correctly in both directions); schedules outside the
  transition hour (09:00 follows the new offset — D23 regression holds).
- Does not matter: how the clock arrives (computed at creation from a clock 18 h earlier, and by stepping).
- Mechanism (source reading): `internal/schedule/fire.go:13-23` delegates to robfig/cron v3 `Next(now.In(loc))`,
  which adds hours in local time (02:00 never occurs) and matches 01:30 again after the repeated hour. The stored
  `zone_offset` is passed but ignored (`Next`'s `offset` parameter is unused).

## Sibling check

- `POST /v1/schedules/{id}/run` is immediate and does not use `Next` — not affected.
- PATCH recomputes via the same `Next` (`internal/api/schedules.go:115-120`) — same behaviour expected; not probed.
- No other zone arithmetic: `LoadLocation` appears only in `internal/schedule/fire.go:18,25`; `Next`/`Offset` callers
  are `internal/api/schedules.go:38,43` (create) and `:115,120` (PATCH) plus `fire.go:67` (firing).

## Severity and status

Medium (skip): a daily 02:00–02:59 job in a DST zone silently does not run one day a year, with no log or event.
Low–medium (double): a 01:xx daily job runs twice once a year; at-least-once consumers may tolerate it.

Status: spec — SPEC.md:162 needs a rule for non-existent and repeated local times (coordinator decision); this
report records the observed behaviour and the conflicting comparable oracles.
