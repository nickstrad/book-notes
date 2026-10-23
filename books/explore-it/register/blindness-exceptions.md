# Blindness — the boundary and its documented exceptions

Each experiment ran from a plain export of the previous tag (source, exploration record,
skills, lessons) plus a charter in the prompt. The export excluded `PLAN.md`, `register/`
and git history by path. That is a **prompt and export boundary, not OS isolation**: the
exporter rejects forbidden file names, not forbidden knowledge inside allowed files, and the
audits are pattern scans of tool inputs. The exceptions below are documented in the run
records; the end review ([REVIEW.md](../REVIEW.md), R2 and R5) reconciled them.

| Run | Exception | Effect on the claims |
|---|---|---|
| E01 | None identified in tool inputs | The 26/27 planted-defect result stands: it precedes every exception below. D07 counts as partial; D23 was a harness finding |
| E04 | Lesson 03's note (exported) named A1 as the backoff-cap silence; the explorer read and cited it ([lesson 04](../lessons/04-judge-it.md), [audit](runs/04-blindness-audit.json)) | A1 is not a blind find |
| E04 | A root `find` listed skill directory names outside the export; no later use of them | Directory names only; no answer-key read |
| E04 | Read its own background task output outside the export (original transcript `runs/04-explorer-run.jsonl` lines 174, 179, 185, 189, 191) | Own runtime output, like E07/E09 below; no answer-key content |
| E05, E06, E07 | A2, A4 and A3 entered through charter or fixer policy work | Decisions, not independent discoveries; counted separately |
| E07 | Attempted to read its own task output outside the export ([audit](runs/07-blindness-audit.json)) | Own runtime output; no answer-key content |
| E09 | The charter, generated from PLAN's Probe paragraph, supplied the expected cause groups and counts ([charter](../exploration/experiments/09-explain-it/charter.md)) | E09's agreement with those group totals is not independent. Its new-site predictions and probes still count |
| E09 | Read a harness-persisted copy of its own oversized tool output ([audit](runs/09-blindness-audit.json)) | Own session output; no answer-key content |

A1–A4 are excluded from every blind-find count ([skills index](../skills/README.md)).

**For future exports:** do not name unrevealed assignments (which experiment owns which
silence or defect) in exported lesson notes, and do not copy expected answers from the plan
into generated charters. The existing raw charters and notes stay as they were run.
