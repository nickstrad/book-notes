# Queuelet exploration protocol

An experiment teaches us about the inherited system by comparing a prediction with observation. Source reading is allowed and valuable; the register, proof scripts and SUT history are withheld. This document defines the local record and roles. It does not describe expected defects.

## Roles and export

The coordinator prepares a charter and an empty export directory. The seeder/fixer and explorer must be from different model families. Every experiment uses a fresh headless explorer. The reviewer independently checks the checkpoint and blindness boundary.

Export only the previous tag’s `books/explore-it/project`, `books/explore-it/exploration`, `books/explore-it/skills` and `books/explore-it/lessons`, preserving their relative paths. From the repository root, with the chosen tag and an existing empty destination:

```sh
git archive "$experiment_tag" books/explore-it/project books/explore-it/exploration books/explore-it/skills books/explore-it/lessons | tar -x -C "$export_dir"
```

The export has no `.git`, `PLAN.md`, `register/`, seeder proofs or hidden copy of repository history. Inspect its file list before dispatch. Existing experiment records, disclosed previous lessons and repro scripts may be present at later tags; they are prior knowledge, never a briefing about unfound targets. Keep the new charter and brief free of planted hints. The explorer can read any file in the export, including all source, and run the lab; it must not read outside the export or consult the originating checkout. Where external tools/services are necessary, their permitted use is specified in the brief rather than granting access to outside source or answer keys.

The committed `experiments/NN-<slug>/brief.md` names the export path, charter, relevant skill, this protocol, budget, permission profile and acceptance checks. It states the export-only reading rule and contains no information about what is planted. The coordinator supplies the charter; it is not assembled by the explorer from the plan.

## Before the probe

Each charter has one target, named resources/instruments and one kind of information:

```text
Explore <target> with <resources> to discover <information>.
I predict <observable behavior>.
Source: <spec line, code file:line, model cell, previous surprise or risk hypothesis>.
I would be wrong if <observable refutation>.
Budget: 60 minutes wall clock; planned reading / probing / writing split.
```

Write predictions and refutation conditions before executing the corresponding probe, including new hypotheses formed mid-run. A source reading is evidence for a prediction, not its confirmation. Separate should (named oracle), predicted (what code appears to do) and observed (what ran). Later observations can shape the next probe within this charter. Backlog adjacent questions as `#offcharter`; the explorer never widens the assignment.

The run budget includes reading, probing, resetting and writing. Track elapsed time and the actual reading/probing/writing split. No probe is worth losing the record to a time-out.

## During the run

Keep `experiments/NN-<slug>/notes.md` as timestamped entries. Use UTC timestamps and these tags:

| Tag | Record |
|---|---|
| `#read` | Source observation with file:line and its interpretation |
| `#hypothesis` | Id, source, prediction and refutation before the probe |
| `#obs` | Exact command, trimmed output and observed effects |
| `#surprise` | A discrepancy requiring characterization |
| `#bug` | Evidence and report link for an oracle violation |
| `#question` | An unresolved question or missing observation |
| `#offcharter` | Adjacent lead saved without expanding the target |

Retain dead ends, failed commands and timestamps. Trim repetitive output but retain evidence that determines the judgment. Preserve raw output or a linked artifact when trimming would lose meaning. Record fixture/reset identity, actors and relevant clock state. Read-only database observations are permitted; inject failures through lab/partner controls, container actions or documented clock controls, not hidden service paths or ad hoc source changes.

When no console answers a question, record a testability request with the question, missing signal, minimal hook and validation probe. The explorer proposes hooks; the fixer implements them in the separate fix stage.

## Characterize surprises

Every `#surprise` ends in a defect report or an explicit written reason it is not one. Isolate variables and reduce the reproduction. If it will not reproduce, preserve exact original state, vary one thing at a time, suspect timing and clock sources, and keep attempts and uncertainty in the log.

Reports under `defects/` contain:

- Title and tag/build seen at; link to the original observation.
- Exact minimal setup and steps, including actor, clock and fixture assumptions.
- Expected behavior with its named oracle: spec line, Never/Always rule, comparable system/standard or approximation with bounds.
- Observed behavior and evidence.
- Variables isolated, including what did not matter.
- Sibling check: where else the same cause could live, with file:line, or “none found.” Predictions about siblings must precede their probes.
- Severity with impact, and status: `open`, `fixed in lesson NN`, `won't fix: <why>` or `spec`.

The explorer writes reports in its own words with local provisional identifiers if necessary. The coordinator reconciles them with the hidden register: retain the planted `D##` id when matched; assign `U##` to an unplanned finding. Identifier reconciliation must not insert the register’s explanation into the explorer’s account.

## Close and score the run

Write `revealed.md` with charter status (`done`, `needs more` or `blocked`), elapsed time and reading/probing/writing split, what the system showed in plain words, model delta with file/cell links, unresolved uncertainties and next hypotheses. End with this score table, using actual counts and evidence links:

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | | Distinct ids stated before probing |
| Hypotheses confirmed | | Predictions supported by the stated observation |
| Hypotheses refuted | | Stated refutation condition observed |
| Code-reading predictions contradicted by the running system | | Refuted subset whose recorded source was code |
| Defects found | | Distinct evidenced reports, excluding disclosure |
| Model cells filled | | Previously unobserved cells now supported by evidence |

Record hypotheses left `needs more` explicitly; confirmed plus refuted need not equal stated. Neither running a command nor adding an inferred diagram cell establishes observed coverage. Count a repeated observation once, and link separate defect findings even when one probe exposes several. A prediction of a defect can be confirmed while the product behavior violates its oracle; keep these judgments separate.

An experiment may need a second run with a narrowed charter. It closes after at most two runs. The coordinator discloses remaining planted misses in the lesson note under “What the experiment missed, and why,” and the fixer flags them `disclosed`. Never backfill these into the explorer’s notes, scores or finds. Preserve out-of-order and unplanned discoveries as they happened.

## Backlog and living models

`hypotheses.md` retains every proposed hypothesis: id, statement/prediction, source, refutation, intended probe/experiment and status (`open` or `experiment NN: confirmed|refuted|needs more`). Keep original statements when updating outcomes.

Models under `models/` grow from observations: `map.md`, `nouns-verbs.md`, `variables.md`, `rules.md`, `entities.md`, `states.md`, `ecosystem.md` and, after evidence supports cause analysis, `causes.md`. Label inference and uncertainty. Later experiments revise earlier models; their diffs are evidence of learning. In a cause-prediction run, commit or timestamp the predicted-defects table before running the confirming probes.

## Handoff and verification

The coordinator commits experiment artifacts as produced, formatting only. Fixes belong in a separate fresh fixer run from the seeder’s family, using the reports as brief. The fixer changes only named findings, adds a regression check per defect under `project/test/checks/` and black-box repro scripts under `project/lab/repro/`, and writes the lesson note. A spec silence also needs a spec line and the coordinator’s one-line decision in the fix commit body. The lesson tag points at the fix commit. Fix history remains separate from the experiment record.

`make verify LESSON=NN` resets and runs that experiment’s repros and checks. Its final verdict is `FIXED AND CHECKED:` after repair or `DEFECTS REPRODUCED:` against the earlier version. Repros report `PRESENT:` or `ABSENT:` per defect; thresholds are binary and may not be weakened to hide a defect. `REF=<tag>` uses the current repros against earlier SUT code. `make checks` runs all regression checks; `make test` runs unit tests. `make dwell LESSON=NN` optionally shows present/absent across earlier tags; its result does not gate the verdict. Lesson 00 uses `BASELINE HOLDS:`.

Lesson notes include actual found/missed/found-early/unplanned outcomes, the score and before/after evidence. Review checks that previous-version repros still reproduce the finding, models match evidence and no answer key leaked into the brief or export.

## Source and local adaptations

Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (2013): 1.1–1.3, 2.1–2.4 and 2.6, Chapter 3, 10.4, 11.5–11.6 and 13.3. Section numbering checked against the publisher’s contents extract; detailed text review remains pending. Blind exports, split agent roles, explicit prediction/refutation lines, score format and the two-commit lesson are this repo’s implementation of the approved plan, not claims of prescribed book formats.
