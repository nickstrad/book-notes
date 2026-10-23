# Blind experiment state

Goal: learn queuelet through the assigned charter and record predictions, probes and model changes.
Spec/procedure: project/SPEC.md and exploration/PROTOCOL.md.
Current: experiment 06 (relate-it) **done** on source checkpoint `explore-it/lesson-05` (2026-10-22 15:28:42Z–15:48Z, ~20 of 60 min).
Done: charter prediction "deleting an owner leaves orphans" refuted for queue→job/schedule (FK + 409), holds for
job→attempt; 20 hypotheses stated, 18 confirmed, 2 refuted; 7 new reports E06-F01…E06-F07 (provisional ids) indexed in
exploration/defects/README.md; models entities/ecosystem/rules/causes and hypotheses.md carry E06 revisions. Record:
exploration/experiments/06-relate-it/{notes.md,revealed.md,tools/}; raw output project/results/e06/.
In progress: nothing. Stack `queuelet-e06` is stopped (`docker compose -p queuelet-e06 -f project/deploy/compose.yml stop`),
volume retained.
Next: coordinator reconciles E06-F ids against the register and commits the artifacts as produced; fixer stage per protocol.
Open decisions: Q-E06-1 (sample worker reaction to deregistration), Q-E06-2 (attempts/outbox retention) — owner:
coordinator/spec. Testability requests T-E06-1, T-E06-2 — owner: fixer. Backlog H-E06-17, H-E06-20…22.
Resume: this plain export is `/tmp/explore-it-experiment-06`; no git history is available. Check run.jsonl, run.stderr,
run-end.txt and exit-status.txt in the export root before restarting. Absence of an exit file alone is not liveness
evidence. The coordinator owns relaunch decisions. To re-run probes: `source exploration/experiments/06-relate-it/tools/env.sh`,
`dc start` (or `dc up -d`), then `tools/reset-e06.sh` and `tools/pN-*.sh`.
