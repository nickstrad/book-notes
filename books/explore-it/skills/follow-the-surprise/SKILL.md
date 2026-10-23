---
name: follow-the-surprise
description: Turn an odd, flaky or once-seen observation into a characterized defect report (or a written non-defect reason), then predict and probe where the same code habit put siblings; use on "that's odd", "I saw it once", a race, or a defect to write up.
---

# Follow the surprise

A surprise is a gap between what you expected and what ran. This skill takes it from a log line to a
durable result: a minimal reproduction with its variables isolated and its oracle named, a mechanism
read in the code, a sibling search driven by the code habit behind it, and a report, check and model
change others can reuse. In this repo's queuelet arc, also follow [the protocol](../../exploration/PROTOCOL.md)
(section "Characterize surprises").

## When to use

- You wrote "that's odd", "huh" or `#surprise` in a probe log, or two consoles disagree.
- A behaviour appeared once, or only under load, concurrency or a moved clock.
- A source reading shows something suspect (an unused field, a `return` inside a loop), but nobody has judged it.
- A defect must be written up, or a list of findings needs grouping by cause to predict where more live.

Not for: choosing what to explore first (use `charter-as-probe`) or deriving the oracles themselves
(use `derive-oracles`); this skill consumes an oracle and a surprise.

## Inputs

- The original command, output, UTC timestamp, ids, actor or key, build or tag, fixture and clock state.
- A way to reset to a known state, and read-only access to the system's own records (database rows,
  event tables, logs, partner sinks).
- The spec, the source tree, and the prior record: earlier reports, hypotheses, notes and models.

## Procedure

1. **Freeze the original.** Before simplifying anything, log the exact command and trimmed output, the
   time, every id, the actor, the configuration, the clock and what else was running. Tag it `#surprise`.
   Rule: every surprise ends in a report or a written reason it is not one. Nothing stays "odd".
2. **Name the oracle, then check the prior record.** Write *expected* with its source: a spec line, a
   Never/Always rule, a comparable system or standard (e.g. RFC 9110 for status codes) or an
   approximation with bounds. No oracle means a `#question` or spec silence, not a defect. Then search
   earlier reports, hypotheses and notes (`grep -rnE '<symptom>|<file>'`); if a run already judged it,
   add your evidence to that judgment instead of filing. A suspect source line is not a finding until judged.
   (queuelet: E09 drafted two reports, then withdrew them against E04's earlier lab-only judgment; E08 read
   the unused `zone_offset` but filed nothing until E09 judged it as U42 — [lesson 09](../../lessons/09-explain-it.md).)
3. **Reproduce from a known state and shrink.** Reset, rerun, then remove steps and data and vary one
   factor at a time. Record what mattered *and* what did not. Keep a control that proves the instrument
   can see both outcomes (the correct path, or the sequential order of a race). Judge by what committed:
   after each trial, read the stored state and side records, not just the response status.
   (queuelet: a 200 ack whose notification was never written showed only in `webhook_outbox` —
   [E09-F07](../../exploration/defects/E09-F07-ack-cancel-error-after-commit-and-lost-notification.md).)
4. **If it will not reproduce, assume timing.** Restore the exact original state and vary one thing at
   a time; suspect the clock, ordering ties, and a second actor on the same row. For a suspected race,
   do not wait for a fault hook: run the action against a concurrent spinner (e.g. repeat `DELETE` on
   the same row until it stops answering 409) for a bounded number of trials, and report the frequency
   as k/n with the lab and tool named; a spinner on every row of a batch widens the window. Separate
   lab-only triggers (a clock stepped backward) from production ones (retention, operator cleanup). Zero
   hits in n trials with working controls is "not reproduced in n", not "absent". Request the missing hook.
   (queuelet: PD6/PD7 were marked "not probed, needs a hook"; a delete spinner confirmed both minutes later —
   [notes](../../exploration/experiments/09-explain-it/notes.md), [race notes](../../docs/knowledge/09-post-commit-races.md).)
5. **Read the mechanism.** Follow the path to `file:line` and say which statement does what. Label each
   claim as observed or read, and as inference where the code is outside your reach (a driver, a library).
   Name the mechanism as a code shape (the *habit*), not as the symptom.
6. **Predict siblings by habit, before probing.** Write the habit as something you can search for, list
   every site with `file:line` and the guard it carries, and write for each a prediction, a refutation
   and a confirming probe, all timestamped before the first probe. Useful habit signals:
   - a write keyed only by id: `UPDATE|DELETE … WHERE id = $1` without the state, token or tenant term;
   - a post-commit tail: statements after `Commit()` or after an autocommit `UPDATE` that the outcome depends on;
   - a per-row abort: `return err` inside a `for … range` loop of a background pass or batch;
   - a silent no-op: `INSERT … SELECT` or `UPDATE` whose `RowsAffected` is never checked;
   - a silent acceptance: a value stored that a parser or library accepts but the spec does not;
   - a console that lies: a status, exit code or health answer that does not check what it reports.
   Probe the predictions and record hit, miss, or "behaves as predicted, not a defect". A predicted site
   with no trigger stays open; look for the trigger in another habit before closing it. When you group
   many findings, count them per habit; an outcome no habit predicted names a new habit.
   (queuelet: E08 predicted one schedule could stop the rest but had no trigger; E09's never-matching cron
   supplied it — [U39](../../exploration/defects/U39-one-schedule-stops-later-schedules.md).)
7. **Write the report.** Title; tag seen at; link to the original log line; minimal setup and steps with
   actor, clock and fixture; expected with the named oracle; observed output; variables isolated (matters,
   does not matter); mechanism; sibling check with `file:line` or "none found"; severity with impact;
   status (`open`, `fixed in lesson NN`, `won't fix: <why>`, `spec`).
8. **Leave durable forms.** A repro that prints PRESENT before the fix and ABSENT after it, where a
   failed control is a probe error, never ABSENT; a race repro samples a bounded number of trials. A
   regression check per defect. A model delta (the cause model gains the habit; state and rule models
   gain the cells you observed). Backlog entries for every sibling left unprobed.

## Output

- One report per defect (fields as in step 7), plus a surprise ledger: `time | surprise | disposition`
  (report id, or "not a defect" with its reason).
- A prediction table: `id | habit | file:line | prediction | confirming probe | outcome | report`.
- A cause-model entry: `habit | code shape | findings | where it can still live`; repro and check names;
  hypothesis ids left `needs more`.

## Pitfalls the experiments exposed

- **Trusting the status code.** 57 of 300 acks answered 200 and never got their notification; only the
  outbox table showed it ([E09-F07](../../exploration/defects/E09-F07-ack-cancel-error-after-commit-and-lost-notification.md)).
  Step 3 reads what committed.
- **Calling a site unprobeable for want of a hook.** PD6 and PD7 were written "needs DB fault … not
  probed" at 20:35:43Z and confirmed by races at 20:41Z and 20:43Z ([notes](../../exploration/experiments/09-explain-it/notes.md)).
  Step 4 samples the race first.
- **A predicted sibling parked without a trigger.** H-E08-23 sat open for a whole experiment until a
  second habit produced the trigger ([hypotheses](../../exploration/hypotheses.md)). Step 6 searches other habits.
- **A reading noted, never judged.** E08 recorded that `zone_offset` is unused; nobody named an oracle for
  the field the API returns until E09 ([U42](../../exploration/defects/U42-schedule-zone-offset-frozen.md)). Step 2.
- **One actor per probe.** The post-commit tail was invisible to every single-actor probe, including E07's
  full state × event matrix ([lesson 09](../../lessons/09-explain-it.md)). Step 4 adds a second actor on the same row.

## Source

Elisabeth Hendrickson, *Explore It!* (Pragmatic Bookshelf, 2013): Ch. 10 "Explore When There Is No User
Interface", 10.4 "Characterizing Bugs"; Ch. 11 "Explore an Existing System", 11.5 (recording findings) and
11.6 (unreproducible bugs), topics as named in this repo's plan; Ch. 13 "Integrate Exploration Throughout",
13.3 "Discovering Systemic Sources of Problems". Section numbers were checked against the publisher's
table of contents; full book-text validation has not been done. The procedure wording, record formats and
habit signals are this repo's, not quotations from the book. Demonstrated in [lesson 01](../../lessons/01-map-it.md)
(every report characterized, D18), [lesson 07](../../lessons/07-state-it.md) (a race reported as k/n) and
[lesson 09](../../lessons/09-explain-it.md) (habits, predictions, the post-commit tail).

## Worked example

**Characterization (E01, D18).** The blind recon's [F01](../../exploration/defects/F01-ack-ignores-lease-token.md)
(reconciled as [D18](../../exploration/defects/D18-ack-ignores-lease-token.md)) shows the report shape: three
steps, the oracle "Only the current lease holder may complete or extend a job", `{"ok":true} [HTTP 200]` with
the impostor's result stored, a wrong-token heartbeat control (409), and a sibling prediction written before
probing (nack lacks the token filter), confirmed as F02. Fixed in lesson 01.

**Prediction from habits (E09, explain-it).** Charter: *explore the defect reports, the models and every
`UPDATE` and firing query in `internal/jobs` and `internal/schedule` to find the habits behind the 26
findings and predict, before probing, where they put an unfound defect*
([charter](../../exploration/experiments/09-explain-it/charter.md)). The blind explorer ran this skill's
draft. Score ([revealed](../../exploration/experiments/09-explain-it/revealed.md)): 21 hypotheses stated,
19 confirmed, 0 refuted, 0 code-reading predictions contradicted, 7 defects found, 20 model cells filled.
Found: [U39](../../exploration/defects/U39-one-schedule-stops-later-schedules.md), [U40](../../exploration/defects/U40-never-matching-cron-accepted.md),
[U41](../../exploration/defects/U41-cron-zone-prefix-overrides-timezone.md), [U42](../../exploration/defects/U42-schedule-zone-offset-frozen.md),
[U43](../../exploration/defects/U43-terminal-notification-after-commit.md), [U44](../../exploration/defects/U44-vanished-job-drops-pass-notifications.md),
[U45](../../exploration/defects/U45-non-iana-local-timezone.md); all fixed in lesson 09 ([reconciliation](../../register/reconciliations/09.md)).
Of nine predicted sites, six produced reports; PD5, PD8 and PD9 behaved as predicted but are not defects
under the spec or the prior record; a sibling added U42. Missed by the predictions: the silent 200 with a
lost notification (inside U43), which named a sixth habit, H-F the post-commit tail. Left unprobed:
H-E09-18 (nack → dead) and H-E09-19 (retention). Model delta: [causes.md](../../exploration/models/causes.md)
(habits H-A…H-F, outcome per prediction), plus `states.md` 5 cells, `rules.md` 5, `variables.md` 6 and
`ecosystem.md` 4. From the outcome table:

| Predicted defect (20:35:43Z) | Habit | Outcome | Report |
|---|---|---|---|
| PD1 never-matching cron accepted → Fire stops at it every tick | H-B + H-E | confirmed; cross-namespace; pause irrelevant; only higher ids blocked | E09-F01, E09-F02 |
| PD6 error after commit answered 409 | H-D | confirmed for cancel and ack under a concurrent DELETE; **plus** silent 200 with no notification … — not predicted | E09-F07 |
| PD7 notify loop aborts | H-E | confirmed for Expire and worker DELETE (500 after commit) | E09-F08 |

**Second target (not yet run).** PLAN section 7 names no litefleet target for this skill; this is a
sketch. In Release It!'s `litefleet` (gateways, a control plane with a Postgres registry, SQLite node
agents, backups and usage events sent to partners), take the first surprise another skill produces, such
as a usage event missing for a statement that ran. Freeze it, name the oracle from that arc's spec or
lesson notes, and judge by the registry rows and the node's SQLite file, not the gateway's status. Seed
step 6 with queuelet's habit signals (statements after a commit that the outcome depends on, `return err`
inside background passes such as backups, answers that do not check what they report), write the
predictions before probing, and score them as E09 did.
