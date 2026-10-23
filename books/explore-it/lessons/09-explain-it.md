# 09 — Explain it: the findings reduce to habits, and the habits predict

Eight experiments left a long defect list. Read one at a time, each report is a separate
mistake. Read together, they show how the code was written. Hendrickson's 13.3 is titled
"Discovering Systemic Sources of Problems"; this project reads that as: a defect list is
evidence about habits, and a habit that put a bug in one place probably put one somewhere
else. That reading, and the predictions below, are this repo's adaptation, not checked
against the book's text ([citation ledger](../register/citations.md)). The experiment turns
it into a test. Group the findings by cause, list the places each habit could still live, write the
predictions down, and only then probe them. If the predictions find nothing, the model is
wrong.

The sources are Hendrickson, *Explore It!* (2013), Chapter 13.3 “Discovering Systemic Sources
of Problems”, with 8.4 for the state table written as code and 13.5 for the stop signal.
These are **approved-plan references**
([E09](../PLAN.md#e09--explain-it-the-pattern-the-prediction-and-the-cause)), checked
against the publisher's contents; full book-text validation is still pending (see the
[citation ledger](../register/citations.md)).

The blind [exploration record](../exploration/experiments/09-explain-it/revealed.md),
committed in the exploration commit before this one, is unchanged. It contains
21 stated hypotheses (19 probed), 19 confirmed, 0 refuted, 0 contradicted code-reading
predictions, 7 reports and 20 newly filled model cells. Process timestamps give about
19 minutes of the 60-minute budget.

## The prediction, made before the probe

At 20:35Z, four minutes into the run and before any probe, the explorer wrote the
[cause model](../exploration/models/causes.md)
and nine predicted defect sites (PD1–PD9), each with a `file:line`. The 26 E01 defect
reports fell into five habits:

| Habit | E01 reports | Where the explorer predicted it still lived |
|---|---|---|
| H-A missing predicate: an `UPDATE`/`SELECT` without the state, token or namespace term its rule needs | 8 | lease ignores `run_at`; success keeps `last_error` |
| H-B silent acceptance: a value stored without the rule the spec states | 5 | cron that never matches; `CRON_TZ=` prefix; `Local` zone |
| H-C clock and time reference | 3 | wildcard-hour detection by string prefix |
| H-D a console that lies | 2 | an error after commit answered 409 |
| H-E one row's error aborts a set-based pass | (loop level) | `Fire`; the expiry and worker-delete notify loops |
| singletons | 8 | — |

The plan's expected shape was "eight missing guards, five silent acceptances, three clock
mistakes, two lying consoles, the rest singletons". The blind explorer arrived at exactly
those counts; the plan text was in its charter, so this agreement is not independent
evidence. The probes are.

Six of the nine predictions produced reports, and a sibling check added a seventh. Three
behaved as predicted but are not defects under the spec or earlier judgements (PD5, PD8,
PD9; see "What remains open"). The habits held as a model. The one outcome nobody predicted
was the worst one: an ack or cancel answered **200** whose notification was never written.
That surprise named a sixth habit, H-F, the **post-commit tail**: terminal transitions
commit first and write their attempt row and outbox row in later statements.

## The faults at lesson-08

All source pointers refer to `explore-it/lesson-08`, relative to `project/`.

| Item | Report | Previous-tag location | Before | Repair |
|---|---|---|---|---|
| U39 | [E09-F01](../exploration/defects/E09-F01-one-unadvanceable-schedule-stops-all-later-schedules.md) | `internal/schedule/fire.go:73-74` | one schedule that cannot advance makes `Fire` return, so every later schedule of every tenant stops; the log names none | try every due schedule; join the errors, each naming its schedule |
| U40 | [E09-F02](../exploration/defects/E09-F02-never-matching-cron-accepted-year-one.md) | `fire.go:25`, `:32-33` | `0 0 30 2 *` is 201 with `next_run_at` 0001-01-01, on create and PATCH | 400 `invalid_cron`; disable-only PATCH still succeeds |
| U41 | [E09-F03](../exploration/defects/E09-F03-cron-zone-prefix-overrides-timezone.md) | `fire.go:12`, `:15`, `:23` | `CRON_TZ=Asia/Tokyo 0 9 * * *` is accepted and fires on Tokyo time while the record says UTC | exactly five fields, else 400 |
| U42 | [E09-F06](../exploration/defects/E09-F06-schedule-zone-offset-frozen-in-record.md) | `internal/api/schedules.go:61`, `:129` | `zone_offset` stays at its creation value after DST | reported for the time of the request |
| U43 | [E09-F07](../exploration/defects/E09-F07-ack-cancel-error-after-commit-and-lost-notification.md) | `internal/jobs/ack.go:14-25`, `operator.go:35-38`, `nack.go:35-39`, `jobs.go:91-103` | a job deleted right after its ack or cancel commits: the ack is answered 409, or the notification is lost, sometimes after a 200 | transition, attempt row and outbox row in one transaction |
| U44 | [E09-F08](../exploration/defects/E09-F08-one-vanished-job-drops-other-dead-notifications.md) | `internal/jobs/expire.go:23-29`, `internal/api/jobs.go:150-158` | in an expiry or worker-delete pass, one vanished job drops every later job's notification; worker DELETE answers 500 after committing | each dead job notified inside the pass's transaction |
| U45 | [E09-F09](../exploration/defects/E09-F09-non-iana-local-timezone-accepted.md) | `fire.go:19`, `:60` | `timezone: "Local"` (the server's zone) is accepted; PATCH stores `""` | IANA names only, else 400 `invalid_timezone` |

## The structural fix: the state table as code

The biggest cause group is the missing predicate. Each handler's `UPDATE … WHERE` carried
whatever guard its author remembered, and E01 repaired the planted members of that group
(D12, D14, D18, D19, D20, D27) one `WHERE` clause at a time. That leaves the habit in place.
The plan's remedy is to move the guards into one table, so that a missing guard becomes a
missing table row, which is visible on one screen.

`internal/jobs/transition.go` holds the E07 state table as data: for each event (`promote`,
`lease`, `heartbeat`, `ack`, `nack`, `expire`, `release`, `cancel`, `retry`, `purge`,
`delete`, `retain`), the states it is accepted in and the states it may lead to (`:46`).
`Guard` (`:104`) renders the state term of a `WHERE` clause from the table, and every
state-changing statement in `internal/jobs` and the job delete route now takes its guard
from it: fourteen call sites, from `Lease` to `Retain`. Lease-holder checks (token, expiry)
stay in the statements, because they are about who is acting, not which state the job is in.
Asking `Guard` for a target an event never reaches panics, so a handler cannot ask for an
arrow the table does not have. The guards' SQL is unchanged except for spelling: Cancel's
`state NOT IN ('succeeded','dead','cancelled')` became the equivalent positive list.

Two checks hold the table in place:

- `TestTransitionContract` (`internal/jobs/transition_test.go`) compares the table with the
  public contract written out by hand (SPEC and the E07 matrix), so editing the table alone
  fails a test.
- `TestL09ModelWalk` (`test/checks/lesson09_test.go`) is the E05 walker turned into a
  model-based check. It runs 1 000 random events against six one-job queues through the
  public API. Before each event it reads the job's state and asks `jobs.Targets` what the
  event may do. It then requires a 2xx and an allowed target, or a 409 and no change.
  Promote, expire and retain are driven by advancing the frozen clock, with every queue's
  move predicted from its row first. As a mutation test, a `Retry` hard-coded to accept
  `cancelled` (D14 reintroduced beside the table) failed the walk at step 241.

## Focused implementation

**U43, U44: no post-commit tail.** `notify` (`internal/jobs/jobs.go:95`) now takes the
caller's transaction. `Ack` became one transaction (`ack.go`): the guarded `UPDATE`, the
attempt row and the outbox row commit together. `Cancel` (`operator.go:35`) and `Nack`
(`nack.go:36`) queue the notification before their commit. `release` (`expire.go:72`)
notifies each dead job inside the expiry or worker-delete transaction, so `Expire` and
`deleteWorker` no longer loop after committing. A concurrent `DELETE` either lands first,
in which case the job was never terminal and the transition is refused, or it waits on the
row lock and deletes after the commit. The outbox row survives (it has no foreign key, and
SPEC already says deleting a job does not erase pending deliveries).

**U39.** `Fire` (`internal/schedule/fire.go:91`) tries every due schedule and returns
`errors.Join` of the failures, each prefixed `schedule <id>:`. The loop still marks the tick
failed, so `/health` stays degraded while a broken schedule is stored. That is the honest
signal for a legacy row, and the log now names it (part of TR-E09-1).

**U40, U41, U45.** `Next` (`fire.go:27`) refuses anything but five fields, which rules out
the parser's `CRON_TZ=`/`TZ=` prefix, and returns `ErrNeverMatches` when the parser finds no
date. `location` (`:20`) refuses `Local` and the empty name that `time.LoadLocation` would
otherwise accept. Create and PATCH already mapped these errors to 400. A PATCH that only
disables a schedule skips validation (`internal/api/schedules.go:130`), so an operator can
still switch off a row stored before this lesson. `TestL09LegacyUnadvanceableSchedule`
inserts such a row with SQL and checks both paths.

**U42.** `liveOffset` (`schedules.go:21`) replaces `zone_offset` in every schedule response
with the named zone's offset at the time of the request. The column is no longer read by
anything. Dropping it would be a migration for a field the API still returns.

The SPEC gains the five-field, never-matching, IANA-only, live-offset and disable-always
lines, plus one line under notifications: a terminal entry commits with its transition.

## Reproduce and check

From `books/explore-it/project`, set the isolated environment and run sequentially:

```sh
export COMPOSE_PROJECT_NAME=queuelet-fix09 QUEUELET_IMAGE_TAG=fix09
export QUEUELET_API_PORT=18400 QUEUELET_HOOK_PORT=18401 QUEUELET_PG_PORT=15633
export QUEUELET_URL=http://localhost:18400 QUEUELET_HOOK_URL=http://localhost:18401
export QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15633/queuelet?sslmode=disable'
make up
make verify LESSON=09
make verify LESSON=09 REF=explore-it/lesson-08
make checks
make test
go vet ./...
make down
```

The public [probe](../project/lab/repro/probe09.py) uses public routes, the lab clock and
read-only SQL. U43 and U44 sample a race with a delete spinner per job: historical runs stop
at the first hits, current runs exhaust their bounds (150 acks and 150 cancels; 20 trials of
eight jobs per path). Missing controls, unexpected statuses and transport failures are
PROBE ERROR, not ABSENT. See the [repro README](../project/lab/repro/README.md).

### Results (final run, sequential, isolated stack `queuelet-fix09`)

Ledger: [09-opus-acceptance.json](../register/runs/09-opus-acceptance.json); one
`09-opus-<step>.log` per step. Instruments were frozen in
[09-instrument-hashes.json](../register/runs/09-instrument-hashes.json) before the final
sequence and rechecked after it (`09-opus-integrity.log`: no mismatches).

| Step | Exit | Seconds |
|---|---|---|
| `make up` | 0 | 7.7 |
| `make verify LESSON=09` | 0 | 273.4 |
| `make verify LESSON=09 REF=explore-it/lesson-08` | 0 | 78.9 |
| `make checks` (94 check functions) | 0 | 697.1 |
| `make test` | 0 | 2.1 |
| `go vet ./...` | 0 | 0.3 |

`make up` reused images built by the first attempt, hence 7.7 s. At lesson-09:

```text
ABSENT: U39 — unadvanceable schedule stored=False; later globex every-minute schedule fired jobs=1 over two minute ticks
ABSENT: U40 — accepted={'create 0 0 30 2 *': False, 'create * * 30 2 *': False, 'patch 0 0 31 4 *': False}; valid control 201
ABSENT: U41 — accepted={'CRON_TZ=Asia/Tokyo 0 9 * * * in UTC': False, 'TZ=Asia/Tokyo 0 9 * * * in UTC': False, 'CRON_TZ=Asia/Tokyo 0 9 * * * in America/New_York': False}; plain 0 9 * * * UTC next=09:00Z
ABSENT: U42 — created at EDT offset -14400; at 07:10Z (EST, -18000) GET zone_offset=-18000, list=-18000
ABSENT: U43 — committed transitions answered non-2xx or never notified: {'ack': 0, 'cancel': 0} (ack 150, cancel 150 trials max; delete spinner per job)
ABSENT: U44 — dead jobs without notification or 500 for a committed delete: expire 0 trial(s) [], worker delete 0 trial(s) [] (20 trials × 8 jobs max)
ABSENT: U45 — accepted={'create Local': False, "patch 'Local'": False, "patch ''": False}; America/New_York control 201
ABSENT: D12 — purged scheduled/retry_wait/queued GET statuses=[404, 404, 404]
ABSENT: D14 — non-dead retry outcomes=[('cancelled', 409, 'cancelled'), ('succeeded', 409, 'succeeded')]
ABSENT: D18 — expired A ack=409; current B ack=200; stored result={'owner': 'B'}
ABSENT: D19 — nack incompatible outcomes=[('wrong-token', 409, 'leased'), ('cancelled', 409, 'cancelled'), ('queued', 409, 'queued')]
ABSENT: D20 — first crash max_attempts=1 leaves dead; completed attempts=1
ABSENT: D27 — disabled schedule after real minute boundary produced=0; enabled control fired
FIXED AND CHECKED: 7 absent; regression checks pass
```

At lesson-08 (historical images and Compose file, current scripts):

```text
PRESENT: U39 — unadvanceable schedule stored=True; later globex every-minute schedule fired jobs=0 over two minute ticks
PRESENT: U40 — accepted={'create 0 0 30 2 *': '0001-01-01T00:00:00Z', 'create * * 30 2 *': '0001-01-01T00:00:00Z', 'patch 0 0 31 4 *': '0001-01-01T00:00:00Z'}; valid control 201
PRESENT: U41 — accepted={'CRON_TZ=Asia/Tokyo 0 9 * * * in UTC': '2033-04-02T00:00:00Z', 'TZ=Asia/Tokyo 0 9 * * * in UTC': '2033-04-02T00:00:00Z', 'CRON_TZ=Asia/Tokyo 0 9 * * * in America/New_York': '2033-04-01T04:00:00Z'}; plain 0 9 * * * UTC next=09:00Z
PRESENT: U42 — created at EDT offset -14400; at 07:10Z (EST, -18000) GET zone_offset=-14400, list=-14400
PRESENT: U43 — committed transitions answered non-2xx or never notified: {'ack': 5, 'cancel': 5} (ack 150, cancel 150 trials max; delete spinner per job)
PRESENT: U44 — dead jobs without notification or 500 for a committed delete: expire 1 trial(s) [{'trial': 0, 'missing': [357, 358, 359, 360, 361, 362], 'worker_delete': None}], worker delete 1 trial(s) [{'trial': 0, 'missing': [366, 367, 368, 369, 370, 371, 372], 'worker_delete': 500}] (20 trials × 8 jobs max)
PRESENT: U45 — accepted={'create Local': True, "patch 'Local'": True, "patch ''": True}; America/New_York control 201
ABSENT: D12 — purged scheduled/retry_wait/queued GET statuses=[404, 404, 404]
ABSENT: D14 — non-dead retry outcomes=[('cancelled', 409, 'cancelled'), ('succeeded', 409, 'succeeded')]
ABSENT: D18 — expired A ack=409; current B ack=200; stored result={'owner': 'B'}
ABSENT: D19 — nack incompatible outcomes=[('wrong-token', 409, 'leased'), ('cancelled', 409, 'cancelled'), ('queued', 409, 'queued')]
ABSENT: D20 — first crash max_attempts=1 leaves dead; completed attempts=1
ABSENT: D27 — disabled schedule after real minute boundary produced=0; enabled control fired
DEFECTS REPRODUCED: 7 present
```

The L09 checks run inside the current verify (165 s): the seven wrappers, the legacy
schedule check and the 1 000-step model walk (about 90 s on its own). At lesson-08 the
race campaigns stopped soon after their first hits: 5 acks and 5 cancels were committed
but answered 409 or never notified, and in the first trial of each pass a
deleted job dropped the notifications of every later job in it.

## The stop signal

Hendrickson's 13.5 asks when exploring has done enough. The plan's signal is the number of
new findings per experiment, which it expected to fall from three to five early on to one or
two by the end. Raw blind reports per experiment:

| E01 | E02 | E03 | E04 | E05 | E06 | E07 | E08 | E09 |
|---|---|---|---|---|---|---|---|---|
| 27 | 3 | 8 | 5 | 3 | 7 | 2 | 4 | 7 |

The count did not fall. E09 found as many as E06, and in code the earlier experiments had
all read. The explorer's conclusion stands: the habit-driven search is not exhausted. What
E09 adds is a measured prediction yield: of nine sites predicted in writing before any
probe, six produced reports, and probing them turned up one sibling finding besides. The
plan expected the share of findings that were predicted to rise across experiments. This
record cannot say whether it did: there is no consistent numerator and denominator for
E01–E08 (E01 itself logged predictions before its probes and built all eight models), so
the comparison is left unmade rather than claimed.

## Dwell and what the earlier experiments walked past

Dwell is paired lesson-08/current evidence only, as in E07 and E08; no full historical sweep
was run. Two of the findings had been seen and left:

- **U39** was H-E08-23 in E08: the explorer read `return e` in `Fire`, predicted that one
  schedule could stop the rest and found no way to make a schedule fail. E09 supplied the
  trigger from a different habit (a cron the parser accepts and can never satisfy). The
  bug needed two habits to meet.
- **U42**'s column was read in E08 (`zone_offset` is unused by `Next`), but no one judged
  the field the API returns. A note in a probe log is not a finding until an oracle is
  named.

The post-commit tail (U43, U44) was invisible to every probe that ran one actor at a time.
It shows only when a second client acts on the same job in the gap. The E07 state matrix
probed every state × event cell, but each cell had one actor.

## What remains open

- **Q-E09-1** (spec): is `0 0-23 * * *` a wildcard-hour schedule under the A5 rule? SPEC
  reads literally as no, and the code follows it.
- **Q-E09-2** (spec): should success or an operator retry clear `last_error`? Today the
  job, its events and its webhook keep the previous attempt's error.
- **H-E04-08** gained a second symptom: after a backward step of the lab clock a queued job
  is leased before its `run_at` and FIFO inverts. The explorer withdrew its two drafts
  (E09-F04, E09-F05) because E04 had judged that clock class lab-only.
- **TR-E09-2**, a deterministic post-commit fault hook, was not built; U43 and U44 sample
  a race instead. H-E09-18 (nack → dead ‖ delete) and H-E09-19 (retention on a real clock)
  are structurally covered by the same transaction change and were not separately probed.
- A schedule stored before this lesson that cannot advance keeps `/health` degraded until
  it is disabled or deleted. No migration revalidates stored schedules, the same policy
  question as E06's legacy templates.
- Earlier forward questions (Q-E08-1 worker identity, H-E08-22, H-E08-24, E07 and E06
  items) are unchanged.

## Try changing and read more

Add a thirteenth event to the table, for example `pause_job` from `queued` to a new
`held` state, and watch `TestTransitionContract` fail before any handler exists. Then
move `notify` in `Cancel` back after the commit and run U43 with 300 trials; count how
often a 200 comes back without an outbox row. Finally, put `return e` back inside `Fire`'s
loop and store a legacy schedule with `TestL09LegacyUnadvanceableSchedule`'s SQL. Decide
whether `/health` should stay degraded for a schedule that can never fire, or whether the
scheduler should disable it.

Read Chapter 13.3 and 8.4 through the approved-plan references above, the
[cause model](../exploration/models/causes.md), the
[notes](../exploration/experiments/09-explain-it/notes.md) and the
[reconciliation](../register/reconciliations/09.md).

## Deviations

- **Fixer seat.** The plan names the seeder's family (Codex) as fixer. This chunk's brief
  assigned the whole experiment, fix included, to Claude Opus 5.5, as in E07 and E08. The
  explorer was a separate, fresh blind Opus process, and the fix was written after its
  record was committed. Independent review should come from a non-Claude family.
- **D27 is a control, not a prediction.** The plan made D27 (a disabled schedule keeps
  firing) E09's predicted target. E01 found and fixed it. It runs here with D12, D14, D18,
  D19 and D20 as unchanged early-fixed controls, all ABSENT at both tags.
- **Scope of the fix commit.** PLAN's E09 fix also lists final skill revisions, the lessons
  index and landing page and the `v1-explored` tag, with the verify line "all 27 defects
  absent at v1-explored; checks 30/30". The plan's own history (section 5.1) puts those in
  a separate `v1-explored` commit, and this chunk's brief assigns them to the checkpoint C
  wrap. This commit adds the lesson-09 row to the lessons index only. "Checks 30/30"
  predates the 30-odd checks per lesson that the experiments grew; the actual count is in
  the results above.
- **Order of work.** The transition table and walk were drafted while the blind explorer
  ran, from the plan and the E07 table, without its findings. The finding-specific fixes
  were written after its record was collected and committed.
- **Instrument interaction.** The first full sequence (`09-opus-*-attempt1.log`) passed
  current and REF with the same verdicts, then `make checks` failed `TestL01U03`. U44 froze
  the clock at 2033, so its notifications were due in 2033 and never delivered after the
  reset, and U03's control waits for the whole outbox to drain (E08 met the same trap with
  U37). U44 now waits for its own rows to deliver while the clock is still frozen. No oracle
  changed; hashes were re-frozen and the whole sequence was rerun.
- **Runtime.** Verify takes about four and a half minutes, not "well under a minute": two real minute
  ticks for U39, two race campaigns and a 1 000-step walk dominate.
