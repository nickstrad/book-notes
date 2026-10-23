# 01 — Map it: reading makes predictions; running tests them

An inherited service gives you several competing descriptions: the README, the spec,
SQL guards, CLI output, database rows and partner records. Recon connects them. Read a
path, predict what it will do, name an observation that would disprove that prediction,
and run it. Keep disagreements in the record; they are the next questions to pursue.

This experiment applies Hendrickson's *Explore It!* (2013), §11.1 “Starting with a
Recon Session”, §11.4 “Selecting Techniques”, §12.4 “Active Reading”, and §2.1–2.4,
§2.6 on charters and their backlog. The explicit prediction/refutation fields, blind
agent export, scores and split experiment/fix commits are this project's adaptations.
Section references follow the approved plan's publisher-contents check; detailed
source-text citation review remains pending. The skill drafts are still drafts.

Read the [charter](../exploration/experiments/01-map-it/charter.md),
[original probe log](../exploration/experiments/01-map-it/notes.md), and
[revealed record](../exploration/experiments/01-map-it/revealed.md) before this answer
key. Those files and the original F reports have not been rewritten by the fixer.

## What this recon actually found

The happy path worked. The description of who may complete work, how attempts are
counted, what a schedule edit means and what a webhook delivers often did not.
Comparing consoles exposed failures that the successful HTTP response hid. Reading
SQL suggested hypotheses; public calls supplied the observations. The recon produced
all eight models, including the causes model originally expected in experiment 09.
Later experiments must challenge those models rather than pretend they do not exist.

| Outcome | Actual evidence and accounting |
|---|---|
| Found in its planned experiment | D01 and D02, both live black-box observations |
| Found early | 23 other registered defects demonstrated live, D23 by executed source harness plus stored service state, and partial D07 charset evidence |
| Missed / disclosed | D07's particular slash/space manifestation was not probed by the explorer; it is disclosed here and included in the repair and replay |
| Unplanned | U01–U08: schedule PATCH validation, webhook result type, shipper blocking progress, auth/health outage reporting, atomic batch insertion, queue limits, unknown backoff, and queued-age reference |

These are reconciled fault counts, not the explorer's report count. Several faults
share an F report: F05 covers D03 and D20, F07 covers D16 and D27, F12 covers D17
and U02. See the [canonical finding index](../exploration/defects/INDEX.md).
The reconciliation assigns all 27 registered defects and eight unplanned faults to
this fix. Deferring a discovered defect to manufacture the planned learning curve
would erase what happened.

The original score, in its own counting units:

| Measure | Count |
|---|---:|
| Hypotheses stated | 48 |
| Confirmed | 44 |
| Refuted | 2 |
| Needs more | 2 |
| Code-reading predictions contradicted by running behavior | 1 |
| Reports | 27: 26 open and one spec question |
| Observed model cells | 110 |

H10b is the code-reading contradiction: promotion happens before the failing schedule
step, so restarting the API can promote a job once even though the loop then stops.
A correct prediction of a bug counts as a confirmed hypothesis, not as correct product
behavior. D23's original finding is a harness result, not an observed DST firing.
The fixer's public opposite-season schedule probe now tests the returned next fire;
actual firing across a DST transition still needs experiment 02's clock instrument.
D08's measured repair is API enqueue/store/GET/lease (and the corresponding
schedule storage path). F17 also predicted numeric loss in the sample worker;
qctl's input/output JSON decoding has the same unprobed concern. Those client-side
predictions remain a later variation/data-journey backlog, not additional E01
finds or a claim of full CLI-to-worker numeric preservation.

## Where the inherited behavior lives

All pointers in this table refer to **`explore-it/v0-inherited`**, relative to
`books/explore-it/project/`. They were read from individual files at that tag, not
from cumulative line numbers in concatenated command output.

| Finding | Previous-tag source pointer | Mismatch |
|---|---|---|
| D01 | `cmd/qctl/main.go:15` | main prints errors without failing the process |
| D02, D15 | `internal/api/queues.go:85` | owner deletion ignores absence and dependents |
| D03, D20 | `internal/jobs/expire.go:5` | expiry does not count/history-record or exhaust an attempt |
| D04, D24 | `internal/webhook/ship.go:16` | any response succeeds; signing uses different bytes |
| D05, D08 | `internal/api/enqueue.go:22`; `internal/store/store.go:19` | missing priority bounds; float decoding loses integers |
| D06, U05 | `internal/api/enqueue.go:73` | truncate oversized requests; commit elements independently |
| D07, U06, U07 | `internal/api/queues.go:25`, `:56` | name/limit/backoff validation is inconsistent |
| D09, D19 | `internal/jobs/nack.go:7` | off-by-one exhaustion and missing token/state guards |
| D10, U08 | `internal/stats/stats.go:8` | queued includes other states; readiness predates existence |
| D11 | `internal/jobs/jobs.go:24`; `migrations/001.sql:4` | no deterministic tie-break; creation timestamps rounded |
| D12 | `internal/api/queues.go:108` | purge only removes queued work |
| D13, D23, D27 | `internal/schedule/fire.go:12`, `:27` | fixed offset, missed-tick burst and missing enabled guard |
| D14 | `internal/jobs/operator.go:12` | retry accepts non-dead states |
| D16, U01 | `internal/api/schedules.go:75` | PATCH neither validates nor recomputes the next fire |
| D17, U02 | `internal/jobs/jobs.go:69` | notification silently truncates and stringifies result |
| D18 | `internal/jobs/ack.go:8` | state guard alone lets a stale worker complete another lease |
| D21 | `internal/api/jobs.go:10`, `:30` | job-id routes do not check namespace ownership |
| D22, U03, U04 | `internal/api/api.go:54`, `:93` | loop stops on error; network shipping blocks core work; auth/health conceal outage |
| D25 | `internal/api/jobs.go:116` | unknown worker error has HTTP 200 |
| D26 | `internal/api/api.go:42`; `internal/api/enqueue.go:95` | unknown write fields and state filter are silently accepted |

For example:

```sh
git show explore-it/v0-inherited:books/explore-it/project/internal/jobs/ack.go | nl -ba
git diff explore-it/v0-inherited explore-it/lesson-01 -- books/explore-it/project/internal/jobs/ack.go
```

## The patch, grouped by the promise it restores

The inherited uncommitted patch was an interrupted **E01 fix**, not a new blind
experiment. Its 17 modified source/spec files and new migration were preserved and
completed. No inherited change was set aside. The final diff is larger than the
plan's two-finding estimate because the assignment includes 35 faults.

- Worker completion now checks the current token; nack also requires leased state.
  Expiry records an unsuccessful attempt and dead-letters at exhaustion. Nack uses
  `>=` at the boundary. Retry requires dead state. These guards remain local;
  experiment 09 still owns the structural transition-table change.
- Queue deletion distinguishes missing from nonempty owners and refuses dependents.
  The migration also enforces ownership for new writes when deletion races an
  already-started enqueue/schedule request. Existing orphan rows are retained for
  inspection with `NOT VALID` constraints; the patch does not invent lost parents.
- Strict write decoding, bounded batch/priority/name/config validation and preserved
  raw JSON repair input handling. Batch insertion uses one transaction and retains
  existing-key and intra-batch idempotency. Queue list-limit clamping is unchanged:
  the explorer's `limit=1000` observation did not establish a contract violation.
- Schedule edits validate cron and IANA location before commit, recompute the next
  fire on cron/zone/enable changes and clear it on disable. Firing checks enabled
  state, advances while paused and permits at most one catch-up execution. `Next`
  uses the IANA location rather than a captured fixed offset.
- Notifications retain the JSON result type and complete result/error, sign exactly
  the transmitted bytes and retry non-2xx responses. A separate serial shipper keeps
  partner timeouts out of promotion/expiry/scheduling. Core errors are logged and
  retried on subsequent ticks. Health reports unavailable storage/degraded loops;
  auth-store failure returns 503 while invalid keys still return 401.
- Stats count exact states. Queued age starts at `greatest(created_at, run_at)` for
  queued rows. Creation timestamps retain microseconds and leasing orders ties by id.
  qctl now returns a failing process status on errors.

Key lines in the repaired tree (same relative project paths):

| Promise | Patch entry point |
|---|---|
| A holder completes only its lease | `internal/jobs/ack.go:10`; `internal/jobs/nack.go:14` |
| Exhaustion and history | `internal/jobs/nack.go:19`; `internal/jobs/expire.go:37`, `:45` |
| Strict numbers/commands; atomic batch | `internal/api/api.go:48`; `internal/api/enqueue.go:15`, `:88`, `:115` |
| Owner isolation and deletion | `internal/api/jobs.go:152`; `internal/api/queues.go:87`; `migrations/002.sql:11` |
| Correct next fire and paused advancement | `internal/api/schedules.go:118`; `internal/schedule/fire.go:13`, `:33`, `:44` |
| Whole typed notification and byte signature | `internal/jobs/jobs.go:79`; `internal/webhook/ship.go:24`, `:32` |
| Independent shipper and surviving core loop | `internal/api/api.go:113`, `:143` |
| FIFO and age | `internal/jobs/jobs.go:24`; `internal/stats/stats.go:18` |

D12's coordinator decision is now a spec sentence: **pending means scheduled,
queued and retry_wait**. Purge removes those states. The leased-job outcome (A2),
backoff cap (A1), running cancellation outcome (A3), and dead-letter retention (A4)
remain undecided. Sparse logs, absent event history and absent clock control
(T1–T3) also remain; none was closed early.

## Verify the same behavior on both versions

From `books/explore-it/project/`:

```sh
make up
make verify LESSON=01
make verify LESSON=01 REF=explore-it/v0-inherited
make checks
make test
```

`lab/repro/lesson-01.ids` names 35 independent cases. Each `ID.sh` calls the shared
black-box oracle in `lab/repro/probe.py`. Current and historical runs use identical
instruments; REF builds API, worker **and qctl** from the archived source. D01 runs
the container's binary, never a newly built current CLI. SQL observations enforce
read-only transactions. Setup/status/transport failures are `PROBE ERROR`, not
absence. Every `TestL01<ID>` repeats its behavior oracle and requires `ABSENT`;
`make checks` also retains the four L00 checks; all 39 named functions pass
(191.586 seconds).

Actual final verdicts (exit 0 in both directions):

```text
FIXED AND CHECKED: 35 absent; regression checks pass
DEFECTS REPRODUCED: 35 present
```

Exact outputs: [current](../register/runs/01-current.log),
[historical](../register/runs/01-historical.log),
[unit tests](../register/runs/01-test.log), and
[full checks](../register/runs/01-checks.log). The current run repeats all 35
named checks (167.102 seconds). With both real-time repros and their regression
repeats, this run took about 6.5 minutes on the warm lab; the original sub-minute
estimate for a two-finding E01 does not describe this expanded assignment.

Measured individual probes:

| Probe | v0 | Repaired |
|---|---|---|
| D13, two missed minute ticks | 2 jobs; 109.9 seconds | 0 jobs; 110.0 seconds |
| U03, six timeout rows | promotion 11.47 seconds; healthy control 1.63 seconds | promotion 1.43 seconds; healthy control 2.03 seconds |

The U03 bound is five seconds for the warm local lab, an approximation agreed
in the reconciliation, not a production latency guarantee. Original discovery
measurements (13.3 vs 2.1 seconds) remain in F13.

D11 uses a batch with tied timestamps, then a public nack/promote cycle that moves
the oldest tuple. D20 starts at attempt one with `max_attempts=1`; it does not
fabricate exhaustion or borrow a prior nack. D23 asks for 02:00 in the season with
the opposite New York offset. D13 waits across two real minute ticks; D27 waits for
one. Their real-time cost is deliberate because v0 has no clock hook.

The first current attempt is retained in `01-current-attempt1.log`, but is invalid
as full reproduction evidence: Docker stdin consumed later ids. Its subsequent
checks also exposed a restart-readiness transport-reset handling defect in the
harness. The verifier now preloads and validates all IDs, isolates stdin, enforces
nonempty/unique IDs and counts every result. Restart readiness retries transport
loss while still requiring a subsequent successful request. Dwell now retains errors
instead of hiding them and permits mixed present/absent outcomes at historical tags.
These are verification infrastructure repairs, not new product findings.
A second current attempt reached all 35 absent observations but its shell failed
after the loop because the fixer edited the running verifier. That mistake is
retained in `01-current-attempt2.log`; acceptance was restarted with frozen
instruments. The first historical attempt likewise stopped at restart readiness
and is retained as incomplete evidence. Only the final complete runs count.

## What the earlier experiments walked past

There are no earlier exploratory experiments. E00 checked a gentle happy path;
it did not explore these risks. The measured dwell is only **v0 → lesson 01**.
Dwell numbers for lessons 02–08 would be fabricated: those tags did not exist
when these findings were discovered and repaired.

| Findings | v0-inherited | lesson-01 fixed tree |
|---|---|---|
| D01–D27 | 27 PRESENT | 27 ABSENT |
| U01–U08 | 8 PRESENT | 8 ABSENT |

The [per-finding dwell table](../register/runs/01-dwell.tsv) is derived from
these same paired verification logs, not an additional `make dwell` campaign.
The [instrument hashes](../register/runs/01-instruments.sha256) match before
and after the final runs. No intervening lesson tags are claimed.

The original seeder's fixtures under `register/proof/` are separate answer-key
evidence, including SQL-written arrangements used at v0. They are not black-box
replays and were not copied into `project/`. Remaining A1–A4 are checked separately
there to confirm this fix has not decided the later experiments' questions.
All four remain PRESENT, each exit 0; see
[the control log](../register/runs/01-unresolved-controls.log).

## Try changing one thing

1. Replace the priority tie-break with `created_at` alone and rerun D11. Why can a
   few untouched jobs appear FIFO even though a batch after a state update does not?
2. Keep the ack state guard but remove its token predicate. Drive lease A → expiry
   → lease B → ack A. Which HTTP response looked plausible without the state model?
3. Move webhook shipping back into the core loop. Compare U03's six-timeout-row
   latency with its ok-partner control; do not infer a universal production SLA
   from the local five-second approximation.
4. Before resolving A2, write what lease → purge → nack should mean. Separate what
   the current code does from the missing product decision.

Read more: the [map skill](../skills/map-codebase-from-behavior/SKILL.md),
[charter skill](../skills/charter-as-probe/SKILL.md), and
[follow-the-surprise skill](../skills/follow-the-surprise/SKILL.md). The next
experiment uses Chapter 3 to make lifecycle behavior observable, then tests the
stronger model this recon already produced.
