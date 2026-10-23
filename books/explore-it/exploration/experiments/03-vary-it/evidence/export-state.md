# Blind experiment state

Goal: learn queuelet through the assigned charter and record predictions, probes and model changes.
Spec/procedure: project/SPEC.md and exploration/PROTOCOL.md.
Current: experiment 03 (vary-it) **done**, one run, at source checkpoint `explore-it/lesson-02`; record closed ≈2026-10-22T04:33Z
(about 21 min of the 60-min budget).
Done (plain export, no git; nothing committed or tagged):
- Record: exploration/experiments/03-vary-it/notes.md (timestamped probes, dead ends) and revealed.md (score: 23 stated,
  21 confirmed, 2 refuted, 2 code-reading contradictions, 8 defects, 28 model cells).
- Reports: exploration/defects/E03-F01 … E03-F08 (provisional ids) and an E03 section appended to exploration/defects/README.md.
  INDEX.md and all earlier reports are unchanged.
- Models: E03 sections appended to exploration/models/variables.md, rules.md and causes.md. hypotheses.md has E03 rows
  appended and the H46 outcome updated.
- Helpers: exploration/experiments/03-vary-it/tools/ (e03env.sh adapted from the E02 copy, which is unchanged; namelen.py,
  p9.sh, p10.sh, p10b.sh). Raw output: project/results/e03/.
In progress: nothing.
Next: coordinator reconciles E03-F ids, then the fixer stage. Backlog for later charters: hypotheses H53–H56, Q-E03-1…5,
TR-E03-1/2.
Open decisions (coordinator/spec owner): nested PATCH merge semantics (E03-F08); reject or implement backoff.max_seconds
(E03-F07); document the 2 MiB body cap and negative delay (Q-E03-2/4).
Resume: this plain export is `/tmp/explore-it-experiment-03`; no git history is available. Check run.jsonl, run.stderr, run-end.txt
and exit-status.txt in the export root before restarting. Absence of an exit file alone is not liveness evidence. The coordinator
owns relaunch decisions. Lab: Compose project `queuelet-e03` is **stopped** (`docker compose -p queuelet-e03 stop` at 04:25:04Z).
Containers and volume are retained, holding the probe data. To inspect, source the tools/e03env.sh helper and run `dc start`.
To remove the stack, run `dc down -v` (it is ours only). Other Docker projects were never touched.
