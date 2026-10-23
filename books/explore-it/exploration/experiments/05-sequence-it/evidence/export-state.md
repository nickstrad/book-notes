# Blind experiment state

Goal: learn queuelet through the assigned charter and record predictions, probes and model changes.
Spec/procedure: project/SPEC.md and exploration/PROTOCOL.md.
Current: experiment 05 (sequence-it) **done** against source checkpoint `explore-it/lesson-04`; run 13:17:51Z–13:42Z
2026-10-22 (~24 of 60 min). Charter prediction refuted. 3 new reports: E05-F01 (worker deregistration release uncounted
→ unbounded redelivery), E05-F02 (ack/cancel keep lease fields), E05-F03 (event console mislabels exits after an expiry).
Done: exploration/experiments/05-sequence-it/{notes.md, revealed.md, walk.sh, tools/}; defects/E05-F01..F03; models
states/rules/nouns-verbs/causes and hypotheses.md appended (E05 sections). Prior records and reports unchanged.
Raw output: project/results/e05/ (probe outputs, walk logs, final-snapshot.txt, all-logs-at-stop.txt, stop-time.txt).
In progress: nothing. Stack `queuelet-e05` stopped with `docker compose stop` (containers and volume kept; no other
project touched).
Next (coordinator): copy/commit artifacts as produced; reconcile E05-F0x identifiers; owner decisions Q-E05-1, Q-E05-2;
testability request T-E05-1 (clock offset mode).
Open decisions: Q-E05-1 (should worker-delete release count an attempt), Q-E05-2 (terminal-row lease fields) — owner.
Resume: this plain export is `/tmp/explore-it-experiment-05`; no git history is available. Check run.jsonl, run.stderr,
run-end.txt and exit-status.txt in the export root before restarting. Absence of an exit file alone is not liveness
evidence. The coordinator owns relaunch decisions. To replay: `source exploration/experiments/05-sequence-it/tools/e05env.sh;
dc start` then `bash exploration/experiments/05-sequence-it/tools/pairs.sh P6` or `BIAS=1 bash .../walk.sh <new seed> 200`.
