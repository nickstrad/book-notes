# Blind experiment state

Goal: learn queuelet through the assigned charter and record predictions, probes and model changes.
Spec/procedure: project/SPEC.md and exploration/PROTOCOL.md.
Current: experiment 02 (make-it-visible) **done** against checkpoint `explore-it/lesson-01`. Run 2026-10-22 03:18:01Z–03:39:40Z.
Done: probe log `exploration/experiments/02-make-it-visible/notes.md`; record `…/revealed.md` (19 hypotheses stated and
confirmed; 3 reports; 46 model cells); new reports `exploration/defects/E02-F01..F03` (local ids, not reconciled); E02
sections appended to all models, hypotheses.md and defects/README.md. Prior E01 records and reports are unchanged.
Replay helpers are in `…/02-make-it-visible/tools/`; raw outputs are in `project/results/e02/`. No source changes, no git writes.
In progress: nothing. The isolated stack `queuelet-e02` is stopped (`docker compose -p queuelet-e02 … stop`) with its volume kept.
Next (coordinator): reconcile E02-F01..F03 ids; decide E02-F03 (spec: webhook timeout/retry cap) and Q-E02-2 (last_error
after success); route testability requests TR-E02-1..6 to the fixer. Open backlog: H49–H52, Q-E02-1.
Open decisions: E02-F03 and Q-E02-2 are spec questions owned by the coordinator.
Resume: this plain export is `/tmp/explore-it-experiment-02`; no git history is available. Check run.jsonl, run.stderr,
run-end.txt and exit-status.txt in the export root before restarting. The absence of an exit file alone is not liveness
evidence. The coordinator owns relaunch decisions. To restart the lab: source
`exploration/experiments/02-make-it-visible/tools/e02env.sh`, then `dc start` (or `dc up -d`) from `project/`.
