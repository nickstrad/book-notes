# E09 explain-it — revealed

Status: **done** (one run). Checkpoint `explore-it/lesson-08`, stack `queuelet-e09` (stopped with `down -v` at
20:48:46Z; no other Compose project touched). Record: [notes](notes.md). Tools: [tools/](tools/). Raw output:
[results/](results/) (race summaries, API log `results/api.log`, `make up` log).

## Elapsed

Run start 20:31:13Z (`run-start.txt`), first command 20:31:18Z, last probe 20:44:30Z, stack stopped 20:48:46Z, record
closed ≈20:51Z. Total ≈ 20 minutes of the 60-minute budget.

| Phase | Wall clock | Approx. |
|---|---|---|
| Reading (STATE, charter, protocol, skill, defect index, causes; all of internal/jobs, internal/schedule, related API handlers, SPEC; stack built in parallel) | 20:31:18–20:35:10 | 4 min |
| Pre-probe writing (statement inventory, habit grouping, PD1–PD9 table, causes.md E09 section) | 20:35:10–20:35:55 | 1 min |
| Probing (H-E09-01…17, predictions logged before each) | 20:36:15–20:44:30 | 8 min |
| Writing (reports, prior-record check and withdrawal, models, backlog, this file, STATE) | 20:44:58–20:51 | 6 min |

## What the system showed

- **Five habits explain the 26 E01 findings**, which is close to the charter's expectation. Missing predicate:
  F01, F02, F03, F06, F07, F15, F18, F23 (8). Silent acceptance: F09, F16, F17, F19, F20 (5). Clock/time: F24,
  F26, F27 (3). Consoles that lie: F14, F22 (2). Singletons: 8. A recurring background habit also runs through
  E02–E08: one row's error aborts a set-based pass.
- **Scheduler: one bad row stops the rest.** `0 0 30 2 *` (Feb 30) is accepted with `next_run_at` 0001-01-01. Each
  tick, `Fire` then stops at that schedule, so every enabled schedule with a higher id, in any namespace, never fires.
  `/health` goes 503 `degraded` and the log says only "schedule cannot advance". Pausing that queue does not help;
  disabling or deleting the schedule releases the others at the next tick (E09-F01, E09-F02). E08 predicted this
  (H-E08-23) but found no trigger.
- **The cron string is interpreted by the library, not by the spec.** A `CRON_TZ=`/`TZ=` prefix is accepted and
  overrides the `timezone` member. With a non-UTC member the result matches neither zone (E09-F03). `Local` is
  accepted as a zone (E09-F09). The record's `zone_offset` stays at its creation value after DST (E09-F06).
- **Terminal transitions have a post-commit tail.** Ack, cancel, expiry and worker delete commit first. They write
  the attempt row and the outbox snapshot afterwards, as separate statements. If a concurrent DELETE removes the job
  in that window:
  - a successful ack is answered 409 (68/300, with a foreign-key message);
  - a committed worker delete is answered 500;
  - the terminal notification is lost (ack 125/300, cancel 14/300), sometimes after a 200, because `INSERT … SELECT`
    of a vanished job inserts nothing without error;
  - in an expiry or worker-delete pass, every later dead job loses its notification, even though those jobs still
    exist (E09-F07, E09-F08).
  The silent-200 case was not predicted. It was the main surprise of the run.
- **Predicted, observed, not filed.**
  - A job is leased before its `run_at` after a backward clock step. FIFO inverts the same way. E04 already judged
    that clock class lab-only (risk H-E04-08), so I withdrew my two draft reports (E09-F04, E09-F05).
  - `last_error` survives success, in the job, its events and the webhook (also seen in E05); the spec only says
    "last error" (Q-E09-2).
  - `0 0-23 * * *` is treated as a fixed hour and skips the repeated 01:00, while `0 * * * *` does not. That follows
    SPEC:163-165 literally (Q-E09-1).
- Lesson-08 behaviour seen incidentally and holding: the fire catch-up rule produced one job per missed window
  (job 303; nyhourly at 07:10Z). Queue deletion refusing queues with schedules was read in source only
  (api/queues.go:133-139), not probed.

## Model delta

- [causes.md](../../models/causes.md): new E09 habit table, written before probes (H-A…H-E with E01 counts and
  later members), and an outcomes table per predicted defect. It adds habit **H-F, the post-commit tail**, and links
  H-E/H-F to E09-F08.
- [states.md](../../models/states.md): 5 new cells (schedule stuck at year 1, blocked later schedules, terminal ‖
  DELETE for ack/cancel and for expiry/worker delete, leased-before-due after a clock step).
- [rules.md](../../models/rules.md): R-E09-1…5. [variables.md](../../models/variables.md): 6 rows.
  [ecosystem.md](../../models/ecosystem.md): 4 rows (robfig prefix and zero time, `Local`, the cleanup actor).
- [hypotheses.md](../../hypotheses.md): H-E08-23 → confirmed; H-E09-01…19, Q-E09-1/2, TR-E09-1/2 added.

## Unresolved uncertainty and next hypotheses

- H-E09-18: nack → dead ‖ DELETE (same tail, single job). Not probed.
- H-E09-19: retention with `retention_hours` 0 on a real clock deleting a just-acked job. This would trigger E09-F07
  with no operator involved. Not probed.
- The race frequencies are specific to this lab and spinner. A deterministic hook (TR-E09-2) would make the repros
  binary.
- `robfig` prefix support was a recollection, not read in this export. The run confirms the behaviour but not which
  library code path produces it.
- Q-E09-1 (what counts as a wildcard hour) and Q-E09-2 (whether `last_error` resets) need spec decisions.

## Score

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 21 | H-E09-01, 02, 02b, 03, 04, 05, 06, 07, 07b, 08, 09, 10, 11, 12, 13, 14, 15, 16, 17 (probed) + H-E09-18, 19 (stated, not probed); [notes](notes.md), [hypotheses.md](../../hypotheses.md) |
| Hypotheses confirmed | 19 | all 19 probed hypotheses; 4 of them (H-E09-09, -10, -11, -12) confirm behaviour that is judged not a defect |
| Hypotheses refuted | 0 | — (H-E09-18, 19 remain `needs more`) |
| Code-reading predictions contradicted by the running system | 0 | none contradicted. One unpredicted extra: silent 200 with a lost notification (E09-F07) |
| Defects found | 7 | [E09-F01](../../defects/E09-F01-one-unadvanceable-schedule-stops-all-later-schedules.md), [F02](../../defects/E09-F02-never-matching-cron-accepted-year-one.md), [F03](../../defects/E09-F03-cron-zone-prefix-overrides-timezone.md), [F06](../../defects/E09-F06-schedule-zone-offset-frozen-in-record.md), [F07](../../defects/E09-F07-ack-cancel-error-after-commit-and-lost-notification.md), [F08](../../defects/E09-F08-one-vanished-job-drops-other-dead-notifications.md), [F09](../../defects/E09-F09-non-iana-local-timezone-accepted.md); F04/F05 withdrawn |
| Model cells filled | 20 | states 5, rules 5, variables 6, ecosystem 4 (★ rows in the E09 sections) |

Charter verdict: **not refuted**. Of the 9 predicted defect sites (PD1–PD9), 6 produced defect reports
(PD1→F01+F02, PD2→F02, PD3→F03, PD4→F09, PD6→F07, PD7→F08). PD5, PD8 and PD9 behaved as predicted but are not
defects by the spec or the prior record. A sibling added F06. Stop signal (13.5): 7 findings in this experiment
against 4 in E08, so the habit-driven search is not yet exhausted.
