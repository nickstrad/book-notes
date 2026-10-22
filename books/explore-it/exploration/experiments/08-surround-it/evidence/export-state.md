# Blind experiment state

Goal: learn queuelet through the assigned charter and record predictions, probes and model changes.
Spec/procedure: project/SPEC.md and exploration/PROTOCOL.md.
Current: experiment 08 (surround-it) **done** in one run against source checkpoint `explore-it/lesson-07`
(19:18:14Z–19:39:50Z, 2026-10-22); awaiting coordinator review/reconciliation. No commits or tags made.

Done (artifacts, uncommitted):
- Record: `exploration/experiments/08-surround-it/notes.md`, `revealed.md`, `tools/` (env.sh, q.py adapted from
  07-state-it, capture.py, p1–p8b probe scripts). Raw output: `project/results/e08/`.
- Reports: `exploration/defects/E08-F01` (API down after start without DB), `E08-F02` (slow/unreachable webhook
  rows starve all later notifications, all tenants), `E08-F03` (DST gap skip / overlap double fire; status `spec`),
  `E08-F04` (frozen DB hangs every request incl. /health).
- Models: E08 revisions appended to `exploration/models/ecosystem.md` and `rules.md`; E08 section appended to
  `exploration/hypotheses.md`. Prior records and reports unchanged; `defects/INDEX.md` not touched.
- Score: 21 stated / 20 confirmed / 1 refuted / 1 code-reading contradiction / 4 reports / 7 model cells.

In progress: nothing. The `queuelet-e08` stack is stopped (`docker compose stop`, volumes kept); the capture listener
has been stopped. Other projects (`queuelet-e01`, `tsdb`) were not touched.

Next (coordinator): reconcile E08-F01…F04 ids; decide Q-E08-1 (worker token purpose) and Q-E08-2 (DST rule for
non-existent/repeated local times); consider T-E08-1…3 testability hooks; open hypotheses H-E08-22…24.

Open decisions: Q-E08-1, Q-E08-2 (coordinator). Blockers: none.
Resume: this plain export is `/tmp/explore-it-experiment-08`; no git history is available. Check run.jsonl,
run.stderr, run-end.txt and exit-status.txt in the export root before restarting. Absence of an exit file alone is
not liveness evidence. The coordinator owns relaunch decisions. To replay probes: from
`exploration/experiments/08-surround-it/tools`, `. ./env.sh`, then `docker compose -f ../../../../project/deploy/compose.yml start`
(or `up -d --build` from `project/`) and `bash lab/reset.sh` from `project/` before each script.
