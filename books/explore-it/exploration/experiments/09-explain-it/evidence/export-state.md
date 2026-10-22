# Blind experiment state

Goal: learn queuelet through the assigned charter and record predictions, probes and model changes.
Spec/procedure: project/SPEC.md and exploration/PROTOCOL.md; charter exploration/experiments/09-explain-it/charter.md.
Current: experiment 09 (explain-it) **done**, one run, at source checkpoint `explore-it/lesson-08`; ≈20 of 60 minutes used.
Done: habit grouping and predicted-defect table (exploration/models/causes.md E09 section, notes 20:35:43Z), probes
H-E09-01…17, 7 reports (exploration/defects/E09-F01, F02, F03, F06, F07, F08, F09; F04/F05 drafted and withdrawn),
models/hypotheses updated, revealed.md scored (21 stated / 19 confirmed / 0 refuted / 0 contradicted / 7 defects /
20 cells). No commits or tags (coordinator owns them).
In progress: nothing. Stack `queuelet-e09` stopped with `down -v` at 2026-10-22T20:48:46Z (image `queuelet:e09` left).
Next: coordinator reconciles E09-F ids with the register and adds E09 rows to defects/INDEX.md; decides Q-E09-1
(wildcard-hour classification) and Q-E09-2 (`last_error` on success); considers TR-E09-1/2. Open hypotheses
H-E09-18/19 for a later run.
Open decisions: Q-E09-1, Q-E09-2 (owner: coordinator); whether the backward-clock class (H-E04-08, H-E09-11/12)
stays a risk (owner: coordinator).
Resume: plain export `/tmp/explore-it-experiment-09`, no git. Run record in
exploration/experiments/09-explain-it/{notes.md,revealed.md,tools/,results/}. To replay, source
tools/env.sh, `make up && make seed` from project/, then the commands in notes.md. Check run.jsonl, run.stderr,
run-end.txt and exit-status.txt in the export root; the coordinator owns relaunch decisions.
