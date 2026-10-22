# Blind experiment state

Goal: learn queuelet through the assigned charter and record predictions, probes and model changes.
Spec/procedure: project/SPEC.md and exploration/PROTOCOL.md.
Current: experiment 04 (judge-it) run **done** at 2026-10-22T12:54:30Z against source checkpoint `explore-it/lesson-03`. It took
about 15 of 60 minutes and filed 5 new reports, E04-F01…F05. Nothing is committed: the coordinator copies and commits.
Done:
- Record and raw data: `exploration/experiments/04-judge-it/notes.md`, `revealed.md` (with the score table), `tools/`
  (e04env.sh, lease1.sh, walk-backoff.sh). Raw output is in `project/results/e04/`.
- Reports: `exploration/defects/E04-F01…E04-F05`.
- Models revised (E04 sections appended): `models/rules.md`, `states.md`, `variables.md`, `causes.md`, plus
  `exploration/hypotheses.md`.
- Inherited records, reports and helpers are unchanged.
In progress: none. The Compose project `queuelet-e04` is **stopped** (`docker compose -p queuelet-e04 stop`). Its containers
and volume are kept; the lab clock was reset before the stop.
Next:
1. The coordinator reconciles E04-F01…F05 identifiers.
2. Owner decisions Q-E04-1…4 (default backoff cap, the meaning of queued-age "readiness", a lease grace window, what
   throughput counts) and retry-history numbering (E04-F02).
3. Open hypotheses for later charters: H-E04-08…11.
Open decisions: the spec silences above belong to the coordinator/owner. Testability requests TR-E04-1 (sweep-lag metric)
and TR-E04-2 (snapshot stats) go to the fixer.
Resume: this plain export is `/tmp/explore-it-experiment-04`; no git history is available. Check run.jsonl, run.stderr,
run-end.txt and exit-status.txt in the export root before restarting. Absence of an exit file alone is not liveness evidence.
To inspect the stack: `source exploration/experiments/04-judge-it/tools/e04env.sh; dc start`. Remove it with `dc down -v`
only when the coordinator says so. The coordinator owns relaunch decisions.
