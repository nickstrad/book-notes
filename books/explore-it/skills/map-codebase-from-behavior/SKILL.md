---
name: map-codebase-from-behavior
description: Recon an unfamiliar codebase in a fixed budget (entry points, nouns and verbs, spec silences with the code's predicted answer, probes that test those predictions), leaving a ranked hypothesis backlog and living models; later, group findings by code habit to predict new defects. Use on day one with an inherited service or when asked "how does this actually work?".
---

# Map a codebase from its behavior

This skill turns a first session with an inherited system into a map you can test: where work enters, which nouns and verbs exist, what the spec promises and leaves unsaid, what the code predicts in each gap, and which predictions the running system refutes. It leaves a ranked backlog and living models that later probes revise. Once findings accumulate, its last model groups them by cause and predicts where the same cause still lives. In this repo's queuelet arc, also follow [the protocol](../../exploration/PROTOCOL.md).

## When to use

- "We inherited this"; the first day in a repository; "how does this thing actually work?"
- A service with a README or spec but nobody who remembers why it behaves as it does.
- Before choosing a first probe: the backlog this skill leaves feeds `charter-as-probe`.
- After several experiments, when the defect list is long enough to ask which habits produced it (step 10).

Not for: one known bug (use `follow-the-surprise`), or one input surface or lifecycle that is already mapped (use the probe skills).

## Inputs

- Source tree, README, spec or API doc, existing tests and checks, deploy description (compose file, Procfile).
- A lab you can run and reset without touching anyone else's stack; read-only database access; partner fakes' records.
- A known seed or fixture and a known clock (host UTC or a lab clock).
- A budget (default 60 minutes) with a planned reading / probing / writing split.
- The existing record, if any (models, backlog, reports): prior knowledge to test, not a briefing.

## Procedure

1. **Set the boundary.** Write one charter: target, resources, information sought, *I predict*, *I'd be wrong if*, budget. Start a timestamped log (`date -u`) and record build or tag, fixture and clock. If a stack with the default project name is already running, start your own under a separate project name and ports. Reserve writing time: it can rival probing time (queuelet E01: 7.6 vs 11.7 minutes, [revealed](../../exploration/experiments/01-map-it/revealed.md)).
2. **Read the promises before the code.** Read the README, spec and test entry points. Turn each spec sentence into a candidate rule, then ask what it does not say and what would make it false. Signals of a silence: an undefined noun ("pending"); a quantity with no bound, unit or cap; an unstated side effect ("releases its leases": does that count an attempt?); an open reference point ("age" from creation or from readiness?); a role whose rights beyond one sentence are unstated. Put each silence in a table with the code's answer (step 3) and its status.
3. **Map entry points to code.** List processes (`cmd/*/main.go`, compose services), routes (`grep -n 'HandleFunc\|mux.Handle\|\.GET(\|\.POST(\|@app\.route\|router\.'`), CLI verbs and background work (`grep -n 'NewTicker\|time.Tick\|setInterval\|cron\|for {'`). For each background loop record the step order and what one error does: a `return` inside the loop ends it for the life of the process. From the migrations record tables, foreign keys or their absence, unique constraints and column precision (`timestamp(0)` rounds to seconds). Take line numbers per file (`grep -n`, `nl -ba <file>`), never from `cat -n` over concatenated files.
4. **List nouns and verbs.** A noun is each table or resource; a verb is each route or CLI command. For each mutating verb, note the terms its `WHERE` clause carries: `UPDATE … WHERE id=$1` with no state, owner or token term is a hypothesis, not yet a finding. Mark each pair walked, read only or untested, and list sequences not yet tried.
5. **Write hypotheses before probing.** One row each: id, source (README, spec sentence or code `file:line`), prediction, refuted-if, probe family (variation, oracle, sequence/state, entity/data journey, ecosystem/boundary, console), rank. Rank by consequence × uncertainty: cross-tenant access, double or lost execution, silent data loss and a dead background loop come before cosmetics. Add a few predictions you expect to hold, as controls. A source pointer predicts a path; it does not prove the path runs.
6. **Walk, then probe in rank order.** First the README happy path, then every documented verb once against the seed (scripted, output kept), then the ranked hypotheses. Record the exact command, trimmed output, exit status and persistent effects. Compare consoles (response, database rows, partner records, logs); a disagreement is the strongest signal (queuelet: stats said 80 queued while the table held 49, [F18](../../exploration/defects/F18-stats-queued-includes-other-states.md)). Write a sibling as a hypothesis before probing it. Label each result's route: live black-box, source harness or read only. When no console answers, write a testability request (question, missing signal, minimal hook, validation probe) and go on.
7. **Close every surprise.** Each ends in a report or a written reason it is not one; keep a disposition table. Name each broken promise separately even when one report covers several: report count and fault count differ. Do not count unprobed siblings as finds.
8. **Write the models and the backlog.** Start with `map` and `nouns-verbs`. Add `variables`, `rules`, `entities`, `states` and `ecosystem` only where you have evidence, marked observed, inferred or unknown; count only observed cells. Every read-only silence, unprobed sibling and off-charter lead becomes a backlog row with source, prediction, refutation and probe family. (queuelet: E01's read-only "does worker release count an attempt?" became [U24](../../exploration/defects/U24-worker-delete-release-uncounted-attempt.md) in E05; F17's unprobed client-side precision became [U29](../../exploration/defects/U29-sample-worker-echo-loses-precision.md) and [U30](../../exploration/defects/U30-qctl-rounds-payload-numbers.md) in E06.)
9. **Revise; never re-create.** A later probe appends a dated, tag-stamped revision section to each model it touched and updates each backlog row it settled. When fixes change the code under a model, re-walk the affected rows (queuelet: E02 re-walked the background loop at lesson-01, [map](../../exploration/models/map.md)).
10. **Build the cause model once findings exist** (ten or more). Group findings by the code shape that produced them, not by subsystem. Inventory every state-changing statement and selection or firing query in the core packages with the guard it carries. Look for each habit where it has not been probed: missing predicate (a statement without the state, owner, token, enabled or tiebreak term its rule needs); silent acceptance (lenient decoder, parse result unchecked for its zero value); clock reference (time computed once, from the wrong field, or too coarse); a console that lies (error body with 200, exit 0 on failure, health that checks nothing); per-row abort (`return` inside a loop over rows in a background pass); post-commit tail (writes after the commit). Timestamp the predicted sites, each with `file:line` and a confirming probe, then probe. Track findings per experiment as the stop signal. (queuelet: E09, [causes](../../exploration/models/causes.md).)

## Output

- `notes.md`: timestamped log tagged `#read #hypothesis #obs #surprise #bug #question #offcharter`.
- Spec silences: | Spec sentence | What it does not say | Code predicts (`file:line`) | Status (observed → id, or read only) |
- `models/map.md`: | Process | Entry | Role | Evidence |, | Route | Handler | Domain / SQL | Walked |, | Loop step, in order | Code | On error | Observed |
- `models/nouns-verbs.md`: | Noun | Verbs | Walked | Sequences tried | Untested sequences |
- `hypotheses.md`: | Id | Statement / prediction | Source | Would be wrong if | Probe family / experiment | Status | Outcome evidence |
- Surprise dispositions: | Surprise | Report id, or reason it is not a bug |
- `models/causes.md` (step 10): | Habit | Code shape | Findings | Where it can still live |, then a timestamped | Id | Habit | `file:line` | Prediction | Confirming probe |
- `revealed.md` ending in the protocol's score table: stated, confirmed, refuted, code-reading predictions contradicted, defects found, model cells filled, needs more.

## Pitfalls the experiments exposed

- **Skipping the loop's step order.** E01's one contradicted code-reading prediction, H10b, said an API restart would not get past a poisoned schedule. Promotion runs before firing, so one job was promoted before the loop died again ([hypotheses](../../exploration/hypotheses.md)). Step 3 records step order.
- **Cumulative line numbers.** `cat -n` over several files gave wrong citations in E01; the explorer appended an erratum ([notes](../../exploration/experiments/01-map-it/notes.md) 01:57:10Z). Step 3 takes numbers per file.
- **A harness read as a live observation.** E01's DST finding (F27) called the SUT function with the stored offset; no schedule fired. The reconciliation kept it as a harness finding ([reconciliation 01](../../register/reconciliations/01.md)). Step 6 labels each result's route.
- **A read-only note is not a finding.** E08 read that `zone_offset` was unused but named no oracle for the field the API returns; it became [U42](../../exploration/defects/U42-schedule-zone-offset-frozen.md) in E09 ([lesson 09](../../lessons/09-explain-it.md)). Steps 2 and 8 turn read-only items into backlog rows.
- **Stale backlog statuses.** `hypotheses.md` still lists H43 (worker-delete release should count an attempt) as `open`, though E05 reported that fault (E05-F01, now U24). Step 9 updates the row.
- **Expected answers in a cause charter.** E09's charter carried the plan's expected group counts, and the explorer matched them exactly; lesson 09 says that agreement is not independent evidence ([charter](../../exploration/experiments/09-explain-it/charter.md)). Keep expected groupings out of the brief.

## Source

Elisabeth Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (Pragmatic Bookshelf, 2013): §11.1 "Starting with a Recon Session" (steps 1–6); §11.4 "Selecting Techniques" (probe family, step 5); §12.4 "Active Reading" (step 2); §2.4 and §2.6 on generating charters and keeping a backlog (steps 5 and 8); §13.3 "Discovering Systemic Sources of Problems" (step 10); one line from §13.5 for the stop signal. Section numbers and titles follow the plan's chapter map, checked against the publisher's table of contents; the book text itself has not been validated against these steps. Procedure wording and record formats are this repo's, not quotations.

Demonstrated in [lesson 01](../../lessons/01-map-it.md) (the recon) and [lesson 09](../../lessons/09-explain-it.md) (the cause model); the models live in [exploration/models](../../exploration/models/README.md).

## Worked example

**E01, recon of `explore-it/v0-inherited`.** Charter: *Explore queuelet end to end with the source tree, `qctl`, the README and `SPEC.md` to discover what it does, what it promises, where the spec is silent, and what I predict the code does there* ([charter](../../exploration/experiments/01-map-it/charter.md)). It took 21.8 of 60 minutes on about 2 kLOC. Score ([revealed](../../exploration/experiments/01-map-it/revealed.md)): 48 hypotheses stated, 44 confirmed, 2 refuted, 2 needs more, 1 code-reading prediction contradicted, 27 reports (26 open, one spec question), 110 observed model cells.

- **Found:** 26 of the 27 planted defects, D07 partially (charset evidence), and unplanned faults [U01](../../exploration/defects/U01-schedule-patch-validation.md)–[U08](../../exploration/defects/U08-queued-age-reference.md). D23's evidence was a harness plus stored state. See the [index](../../exploration/defects/INDEX.md) and [reconciliation 01](../../register/reconciliations/01.md).
- **Wrong or missed:** H10b (step order). D07's slash/space manifestation was not probed; lesson 01 disclosed it. H22 (FIFO ties) and H25 (DST without clock control) stayed `needs more`.
- **Model delta:** all eight models created, including a ten-cause `causes.md` the plan expected only in E09, plus backlog H00–H48 and Q1–Q5 ([hypotheses](../../exploration/hypotheses.md)).

Excerpt, the spec-silence table ([notes](../../exploration/experiments/01-map-it/notes.md) 01:57:10Z); the read-only row became U24:

| Spec sentence / gap | What it does not say | Code predicts | Status |
|---|---|---|---|
| purge "removes all pending jobs" | which states are pending | queued only (queues.go:109) | observed H24 → F25 |
| "oldest queued age" | measured from created_at, run_at or promotion | min(run_at) (stats.go:19) | observed → F26 |
| "DELETE /v1/workers/{id} releases its active leases" | whether release counts an attempt | no attempt (internal/api/jobs.go:110) | read only |

**E09, the cause model at `explore-it/lesson-08`.** Charter, in short: read the reports, the models and every `UPDATE` and firing query in `internal/jobs` and `internal/schedule`, find the habits behind the 26 E01 findings, and predict before probing where they still hide a defect ([charter](../../exploration/experiments/09-explain-it/charter.md)). Score ([revealed](../../exploration/experiments/09-explain-it/revealed.md)): 21 stated, 19 confirmed, 0 refuted, 0 code-reading predictions contradicted, 7 defects, 20 model cells. At 20:35Z, before any probe, the explorer wrote an 18-row statement-and-guard inventory and nine predicted sites, PD1–PD9 ([notes](../../exploration/experiments/09-explain-it/notes.md)). Six produced reports and a sibling added a seventh: [U39](../../exploration/defects/U39-one-schedule-stops-later-schedules.md), [U40](../../exploration/defects/U40-never-matching-cron-accepted.md), [U41](../../exploration/defects/U41-cron-zone-prefix-overrides-timezone.md), [U42](../../exploration/defects/U42-schedule-zone-offset-frozen.md), [U43](../../exploration/defects/U43-terminal-notification-after-commit.md), [U44](../../exploration/defects/U44-vanished-job-drops-pass-notifications.md), [U45](../../exploration/defects/U45-non-iana-local-timezone.md). PD5, PD8 and PD9 behaved as predicted but are not defects under the spec or earlier judgements. Nobody predicted the worst outcome: an ack or cancel answered 200 whose notification was never written. It named a sixth habit, the post-commit tail. Model delta: the E09 regrouping in [causes.md](../../exploration/models/causes.md), plus 5 states, 5 rules, 6 variables and 4 ecosystem cells. The stop signal did not fall: raw blind reports for E01–E09 were 27/3/8/5/3/7/2/4/7 ([lesson 09](../../lessons/09-explain-it.md)).

**Second target: litefleet's tenant API (not yet run; planned once both arcs land).** Export only `books/release-it/project` at a chosen tag; Release It!'s plan and lessons list its flaws and must stay out of reach. With no `SPEC.md`, read the tenant API table in its [README](../../../release-it/project/README.md) actively: "`execute` runs its statements in one transaction" (and if the node dies mid-transaction?); "`query` takes one read statement" (and a write, or two statements?); `DELETE /v1/databases/{name}` → 204 (with a query in flight? its backups and metered usage?). Map HAProxy :18080 → `cmd/gateway` (`internal/api/handlers.go`) → control's Postgres registry (`internal/routes`, `migrations/001_init.sql`) → `cmd/node` (`internal/engine`, one SQLite file per database), plus the background work: heartbeats, snapshots to the blobstore, one usage event per statement to metering. Nouns: tenant, API key, database, node, statement, backup, usage event. Rank cross-tenant name resolution and lost or duplicated usage events first.
