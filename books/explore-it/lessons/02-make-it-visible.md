# 02 — Make the invisible visible

A job's current state cannot explain how it got there. A partner's receipt cannot
prove the sender received its response. This experiment compared the API, worker
console, attempt rows and webhook partner, then requested the missing instruments.
The useful result was a richer way to ask what happened, alongside two new bugs.

The source is Hendrickson, *Explore It!* (2013), Chapter 3: 3.1 (the moonwalking
bear), 3.2 “Digging Deeper,” 3.3 “Testability and Making the Invisible Visible,”
and 3.4 “Consoles and Logs.” These citations follow the approved [chapter map](../PLAN.md#22-chapter-map-kept-sections-only);
the fixer did not independently validate the full book text. The particular hooks,
probe IDs and checks below are this project's implementation of those ideas.

## What the blind experiment learned

[The original record](../exploration/experiments/02-make-it-visible/revealed.md)
remains unchanged at experiment commit `e390c31`. The explorer stated 19
predictions, confirmed 19, refuted zero, recorded zero code-reading contradictions
and filled 46 model cells. Its three reports contain two new bugs and one spec
clarification. The broader charter's disagreement condition occurred when the
slow receiver recorded 200 while the sender timed out; that does not rewrite the
score of the individual predictions. Process timestamps put the run at about
22 minutes; reading/probing/writing estimates overlap.

| Classification | Actual outcome |
|---|---|
| Found as new in E02 | U09 default sample-worker kind registration; U10 malformed destination blocks independent delivery |
| Expected but found early | D03 expiry accounting and D04 2xx-only delivery were found and fixed in E01; E02's passing observations are regressions |
| Missed/disclosed this time | No additional planted defect is assigned or manufactured |
| Unplanned clarification | E02-F03: delivery acknowledgement, timeouts and duplicates; no third bug |
| Testability requests | T1/T2/T3 implemented as T01/T02/T03 instruments; TR-E02-1/2/5/6 covered |
| Deferred proposals | TR-E02-3 backlog/liveness metrics and TR-E02-4 eligibility explanation exceed the minimal hooks; coordinator owns later consideration |

States, causes and ecosystem models already appeared during E01. E02 refines
those models and their evidence; it does not pretend to create them for the first
time. A1–A4 stay with E04–E07. Structural transition guards remain E09.

## Compare the views before changing the system

The explorer used an actual crash worker, 500 responses, a malformed destination
and a slow receiver. It asked whether attempts increased, whether notifications
were marked delivered and which console could explain a waiting job. E01's
correctness repairs held, but sparse logs could not explain notification failures
or tie worker processing back to a particular job. The source exposed the gaps;
the observations showed their cost.

At `explore-it/lesson-01`, `cmd/worker/main.go:24` advertises `fail` but not the
README's `fail3`. Exact matching in `internal/jobs/jobs.go:44` consequently leaves
fail3 queued. `internal/webhook/ship.go:25` constructs the notification request;
line 27 returns from the whole batch when one URL cannot be parsed. Create/PATCH
in `internal/api/queues.go:41` and `:84` persist that destination unchecked.
`internal/jobs/ack.go:14` writes the completion timestamp into both attempt timing
fields. There is no events route, and the loop takes wall-clock ticker times
in `internal/api/api.go:120`. These are pointers into the previous tag, not the
fixed tree.

## The patch and the instruments it earns

The hook checkpoint is `3287bd9`. The final lesson tag points at the following
runtime-fix commit. The combined source, instruments and evidence exceed the
plan's approximately 400-line split threshold, so hooks and bug fixes are separate.

- `migrations/003.sql:17` records job changes in the same transaction as each
  write. The triggers observe; they contain no permission table or transition
  policy. Every lifecycle writer passes application `event_at`. Delete, purge
  and retention use a returning CTE so their records survive the removed job.
- `internal/api/observe.go:10` reads events by retained namespace ownership.
  Deleted jobs retain history; another tenant sees an empty list. Worker/attempt
  and outbox/status/error fields correlate the separate views. The API console
  logs committed pending events as JSON and marks them logged afterward; a
  process crash can repeat a console line, so consumers deduplicate event_id.
- Leasing records `leased_at`; ack/nack/expiry use it for `started_at`. Pre-hook
  leases have null start times, and no historical timing is synthesized. Existing
  pre-hook attempt rows are not retroactively reconstructed.
- `internal/clock/clock.go:19` supplies a synchronized API semantic clock.
  `POST /admin/clock` exists only under `QUEUELET_TEST_CLOCK=1` and requires an
  admin key. Production-disabled routing is checked directly. Compose enables it
  for this lab. All API time decisions and SQL timestamp arguments share it;
  polling, workers, heartbeats, network deadlines and hooksink timestamps stay
  real. Reset or API restart returns to real UTC. It does not speed up execution.
- U09 advertises `fail3` in standalone and qctl worker defaults, with an explicit
  README list. Other numbered kinds require explicit registration. The queue
  API continues exact matching. The fixer additionally measured the actual qctl
  path, which was only a sibling inference in the raw report.
- `internal/api/queues.go:190` validates destination shape and HTTP(S) schemes
  on create/PATCH. `internal/webhook/ship.go:25` sends request-construction
  failures through the per-row failed-attempt path at line 48. The next row still
  runs. A private-schema Go check seeds a malformed persisted row independently
  from the public historical repro.

The specification now states that outbox destination/body/secret snapshots are
intentional. Queue edits affect future snapshots; deleting a job does not erase
pending delivery. Delivered means the sender received 2xx before its two-second
deadline. A receiver may complete a timed-out request, so a retry may duplicate
its side effect. Five-second retries remain; no retry cap, timeout option or
dead-letter policy is added. Finite observations do not prove “forever.”

## Run the same oracles on both versions

From `books/explore-it/project`, run sequentially:

```sh
make up
make verify LESSON=02
make verify LESSON=02 REF=explore-it/lesson-01
make checks
make test
python3 lab/observe02.py
```

`lesson-02.ids` names T01/T02/T03/U09/U10. These are five assigned items,
comprising three observation gaps and two bugs. They are not five newly found
correctness defects. The same public scripts and bounds run on each binary.
T01 compares JSON console events for a completed job and an abandoned lease;
T02 checks history, failed delivery metadata, true lease start, deletion and
namespace scope; T03 moves a delayed job through promotion and lease expiry.
U09 uses an echo control beside fail3 in both actual worker modes. U10 compares
malformed destinations with an independent healthy delivery.

Historical missing clock/events routes are evidence of the respective instrument
gap, not absence of a lifecycle defect. The older mux returns 405 for GET events
because its generic POST job-verb route matches the path. The first REF run
stopped with PROBE ERROR rather than pretending the check passed; that output
and Compose diagnostics are retained. The corrected E02 decoder preserves text
HTTP errors and recognizes 404/405 only as missing event instruments. E01 repros
are unchanged. Another preliminary direct-check failure compared Go nanoseconds
with PostgreSQL microseconds; the fixture now truncates its input precision.

Measured verdicts, elapsed time, actual replay evidence and dwell follow in the
acceptance section below. The exact command exit statuses, including preliminary
failures, are machine readable in `register/runs/02-acceptance.json`.

## Try changing one variable

Register only `fail` and enqueue `fail3`, then register `fail3` explicitly. Compare
eligibility without assuming the prefix is a wildcard. Put hooksink in 500, slow
and ok modes; compare receiver records with sender status/error and outbox ID.
Advance the API clock during a real worker sleep and observe which time moves.
Repeat a clock request with acme-key, then with the lab flag disabled, to separate
authorization from route availability. Restore real time and ok mode afterward.

For further reading, use Chapter 3.2–3.4, the [make-invisible-visible skill](../skills/make-invisible-visible/SKILL.md),
the immutable [probe notes](../exploration/experiments/02-make-it-visible/notes.md),
and the [coordinator's reconciliation](../register/reconciliations/02.md).


## Measured acceptance and dwell

| Command | Exit | Elapsed | Result |
|---|---:|---:|---|
| `make up` | 0 | 78.833 s | current stack built and healthy |
| `make verify LESSON=02` | 0 | 54.279 s | `FIXED AND CHECKED: 5 absent; regression checks pass` |
| `make verify LESSON=02 REF=explore-it/lesson-01` | 0 | 61.646 s | `DEFECTS REPRODUCED: 5 present` |
| `make checks` | 0 | 226.464 s | all 46 integration checks pass; Go reports 225.161 s |
| `make test` | 0 | 1.223 s | unit suite passes |
| `python3 lab/observe02.py` | 0 | 14.860 s | actual crash/500/slow/recovery and container CLI replay pass |

Current T03 advanced 3,662 semantic seconds in 1.95 real seconds. U09's historical
Compose and actual qctl default workers both left fail3 at zero attempts while
echo succeeded; current fail3 had one attempt in both modes. U10's historical
create/PATCH statuses were 201/201/201/200/200/200 and the independent receiver
saw nothing within eight seconds. Current statuses were six 400s and the healthy
notification arrived. These bounds are identical across versions.

The actual replay used crash job 395: worker 49 exited 3; event history recorded
created, leased, lease_expired and notification_failed with outbox 128/status 500.
Slow job 396/outbox 129 had a receiver status-200 record while queuelet recorded
`context deadline exceeded`, status 0 and attempt 1. After ok mode, attempt 2
recorded notification_delivered/status 200. The real container qctl read events,
watched the queue and set/advanced/reset the clock. These observations explain the
previously missing history without promising exactly-once delivery.

| Item | Measured at lesson-01 | Measured after E02 |
|---|---|---|
| T01 structured console correlation | PRESENT | ABSENT |
| T02 retained event history | PRESENT | ABSENT |
| T03 semantic clock control | PRESENT | ABSENT |
| U09 sample-kind registration | PRESENT | ABSENT |
| U10 malformed notification isolation | PRESENT | ABSENT |

This is a paired-log dwell measurement, also in `register/runs/02-dwell.tsv`.
No v0 replay or full `make dwell` sweep was measured here. The table proves these
items survived E01; it does not invent later missed experiments. All 102 raw
exploration artifacts match `e390c31` (`02-record-integrity.json`); frozen oracle
hashes match through both final verifications and the full suite. E01 instruments
are unchanged. No required check remains failed or unrun. Checkpoint B independent
review belongs after E06, and detailed book-text citation review remains pending.

Authored implementation/document whitespace check exits 0. Full staged whitespace
check exits 2 for preserved raw Compose/diagnostic trailing spaces; this is a
non-gating evidence-format limitation, recorded with its exact output. Raw logs
were not trimmed to manufacture a clean result.
