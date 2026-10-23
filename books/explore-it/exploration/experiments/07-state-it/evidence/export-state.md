# Blind experiment state

Goal: learn queuelet through the assigned charter and record predictions, probes and model changes.
Spec/procedure: project/SPEC.md and exploration/PROTOCOL.md.
Current: experiment 07 (state-it) **done** on source checkpoint `explore-it/lesson-06` (run 17:59Z–≈18:18Z on 2026-10-22, about 20 minutes of the 60-minute budget). It found two open defects:
- [E07-F01](exploration/defects/E07-F01-worker-delete-racing-lease-leaves-zombie-lease.md): a worker DELETE racing that worker's lease leaves a zombie lease.
- [E07-F02](exploration/defects/E07-F02-purge-misses-rows-changing-state-concurrently.md): a purge racing lease, nack, release or expiry leaves jobs pending.

Score: 13 stated / 12 confirmed / 1 refuted / 1 code-reading contradiction / 2 defects / 41 cells.
Done (no commits; the coordinator commits the artifacts as produced):
- exploration/experiments/07-state-it/{notes.md, revealed.md, tools/, evidence/project-results/e07/}
- an E07 revision in models/states.md, rules.md and causes.md
- an E07 section in hypotheses.md and defects/README.md, plus the two E07 reports

Inherited records and reports are unchanged.
In progress: nothing.
Next:
1. The coordinator reconciles E07-F01/F02 ids.
2. A fixer run uses the reports as its brief. Repro helpers: tools/p2_races.py, p3_wdel_isolate.py, p5_purge_siblings.py.
3. Optional second run: H-E07-16/17, plus the inferred-only races listed in revealed.md.

Open decisions (owner: coordinator):
- Q-E07-1: may a heartbeat shorten a lease?
- Q-E07-2: nack 409 message text.
- Testability requests T-E07-1 (real-time commit stamp / loop-step duration) and T-E07-2 (transaction pause hook for races).

Resume: this plain export is `/tmp/explore-it-experiment-07`; no git history is available. The run's isolated stack is Compose project `queuelet-e07`. Its containers are stopped (not removed) and the volume is kept. To restart, run `. exploration/experiments/07-state-it/tools/env.sh` and then `make up` in project/. Never touch other `queuelet-*` stacks. Before restarting, check run.jsonl, run.stderr, run-end.txt and exit-status.txt in the export root. An absent exit file alone is not liveness evidence. The coordinator owns relaunch decisions.
