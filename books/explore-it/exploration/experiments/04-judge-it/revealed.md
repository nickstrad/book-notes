# Experiment 04 revealed — judge-it

Status: **done** (one run). Build: plain export of `explore-it/lesson-03`; isolated stack `queuelet-e04`
(API :18250, hooksink :18251, Postgres :15483), no fixture seed, lab clock frozen and walked per probe, reset at stop.

## Elapsed

Wall clock 12:39:51Z → 12:54:30Z ≈ 15 min of the 60-min budget.

| Phase | Time | Approx. |
|---|---|---|
| Reading (STATE, charter, protocol, skill, SPEC, README, lesson 03, models, source) | 12:39:51–12:44 | 4 min |
| Writing rules/predictions before probing; stack build ran in parallel | 12:44–12:45 | 1 min |
| Probing P1–P12 (log entries written between probes) | 12:45:06–12:50:33 | 5.5 min |
| Writing reports, models, hypotheses, this file, STATE | 12:50:33–12:54:30 | 4 min |

## What the system showed

I wrote thirteen Never/Always rules before touching the stack, then tried to break each one with the API, `psql` and the
lab clock. Eight held up under probing. Leasing stayed atomic under 30 parallel calls. Attempts stopped at max_attempts on
the nack path. Order held: priority first, then FIFO for batches and for jobs created at the same instant. Stats matched a
direct `psql` count exactly for both the queue view and the namespace view. Cancelled jobs never ran. Operator retry gave a
fresh attempt 1. Exponential backoff doubled exactly. Error classes were right for job verbs.

Five rules broke:

1. **Errors outside handlers aren't in the error format** ([E04-F01](../../defects/E04-F01-unmatched-route-errors-not-json.md)).
   A wrong method or an unknown path gets Go's plain-text 404/405 instead of the JSON error body.
2. **A retry can come before the failure, or 292 years later** ([E04-F03](../../defects/E04-F03-negative-delay-wraps-and-precedes-failure.md)).
   `delay_seconds` is only checked at the top, so a negative value on nack or enqueue goes through. −5 schedules the retry
   before the failure it reports. −9223372037 wraps to the year 2322: a more negative delay gives a later job. The lesson-03
   duration fix covered only the upper bound.
3. **An expired lease still works until the sweep runs** ([E04-F04](../../defects/E04-F04-expired-lease-honoured-until-sweep.md)).
   Heartbeat, ack and nack check the token and state but not `lease_expires_at`. Sent the same second the lease expired,
   a heartbeat brings the lease back and no expired attempt is ever recorded. After one 1-s tick the identical request gets 409.
4. **Queued age counts time spent leased** ([E04-F05](../../defects/E04-F05-queued-age-counts-leased-time.md)).
   A job released by lease expiry or worker deletion reports an age of 600 s or 100 s although it has just become queued.
5. **Attempt history is ambiguous after a retry** ([E04-F02](../../defects/E04-F02-attempt-history-after-retry-ambiguous.md)).
   Numbering restarts at 1, so "numeric order" lists the final success between two earlier failures, and two executions
   share n=1.

Judged not to be defects, with reasons in the notes surprise ledger:
- **FIFO inversion when the app clock steps back.** It is real but lab-only here, so it is a risk (H-E04-08), not a defect.
- **Default backoff reaches 30.34 days at attempt 20.** This is spec-compliant (the stated bound is 68 years) and was
  already disclosed in lesson 03 as A1. It goes to the owner as Q-E04-1.
- **A dead job gets a future `run_at`.** Harmless: a terminal row's `run_at` has no documented meaning and retry overwrites it.

The charter's prediction, from reading the code, was "the code honours every rule on my list". It was refuted on the first
probe (P1).

Comparable systems were used only to calibrate oracles; another product's behaviour is not this product's requirement:
- **SQS FIFO** orders by sequence number, so it would not invert on a clock step. I used it to frame H-E04-08.
- **SQS ApproximateAgeOfOldestMessage** counts age from the original send, the opposite of my reading. That is why
  E04-F05 is flagged as depending on what the spec's "readiness" means.
- **SQS 12 h visibility maximum and pg-boss's retry-delay cap** were recalled for Q-E04-1, not verified in this run.

## Model delta

- [rules.md](../../models/rules.md): E04 oracle table with R31–R43. Each row has authority, enforcing source, prediction,
  observation and judgment; 13 starred rows.
- [states.md](../../models/states.md): 13 new observed cells. They cover expired-before-sweep heartbeat/ack/nack,
  after-sweep heartbeat, negative-delay nack, cancel-then-holder verbs, retry history, release paths × stats age,
  concurrent lease, and default backoff at attempt 20.
- [variables.md](../../models/variables.md): 8 rows: delay sign on nack and enqueue, request timing vs sweep, HTTP method,
  attempt count across retry, default backoff count, concurrency, and app-clock step back.
- [causes.md](../../models/causes.md): C20–C24, with sibling predictions and their status.
- [hypotheses.md](../../hypotheses.md): H-E04-C, R31–R43, S1–S2 with outcomes; open items H-E04-08…11 and Q-E04-1…4.

## Unresolved uncertainty and next hypotheses

- Owner decisions:
  - Q-E04-1: default backoff cap.
  - Q-E04-2: what "readiness" means for queued age.
  - Q-E04-3: whether there is a lease grace window.
  - Q-E04-4: whether "completed" includes dead and cancelled jobs.
  - Whether retry history should survive a retry, and how it is numbered (E04-F02).
- H-E04-09: stats read in three statements with no snapshot. Concurrent consistency is not probed; TR-E04-2 proposes a hook.
- H-E04-10: ack is not atomic. Needs DB fault injection (ecosystem charter).
- H-E04-08: FIFO under real clock steps or replica skew.
- R42: linear backoff not re-probed in this run.
- TR-E04-1: there is no signal for sweep lag, so the real-time size of the E04-F04 window cannot be measured.

## Score

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 16 | H-E04-C (charter) + R31–R43 (12:44Z) + H-E04-S1, S2 (12:49:30Z), all before their probes. The E04-F05 worker-delete sibling was logged after its probe and is not counted |
| Hypotheses confirmed | 15 | R31–R43 and S1–S2 matched their predictions (R42 exponential half only; linear needs more) |
| Hypotheses refuted | 1 | H-E04-C: violated rules R35, R36, R38, R39, R40 |
| Code-reading predictions contradicted by the running system | 1 | H-E04-C, whose source was code reading of nack.go, stats and the lease query. None of my own code-based predictions was contradicted |
| Defects found | 5 | [E04-F01](../../defects/E04-F01-unmatched-route-errors-not-json.md), [E04-F02](../../defects/E04-F02-attempt-history-after-retry-ambiguous.md), [E04-F03](../../defects/E04-F03-negative-delay-wraps-and-precedes-failure.md), [E04-F04](../../defects/E04-F04-expired-lease-honoured-until-sweep.md), [E04-F05](../../defects/E04-F05-queued-age-counts-leased-time.md). Excluded: A1 (disclosed), regressions that passed, the lab-only FIFO inversion |
| Model cells filled | 18 | 13 new states.md cells, plus variables rows not covered there (HTTP method, enqueue negative delay, clock step back), plus rules-only observations (stats = psql count, batch/priority order). Repeated observations counted once |

Raw evidence: `project/results/e04/` (p1–p12, up.log, final-snapshot.txt, all-logs-at-stop.txt). Replay helpers:
[tools/e04env.sh](tools/e04env.sh), [tools/lease1.sh](tools/lease1.sh), [tools/walk-backoff.sh](tools/walk-backoff.sh).
