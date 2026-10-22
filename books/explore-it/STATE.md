# Explore It! — project state

Goal: turn Hendrickson's *Explore It!* into queuelet, nine honest exploration
experiments and eight reusable skills about probing and learning a codebase.
The approved scope omits the people/process material. Specification and order:
[PLAN.md](PLAN.md), [AGENTS.md](../../AGENTS.md), and
[continuation brief](/root/booknotes-sol-exploreit2.md).

## Status

2026-10-22 UTC: checkpoint A reviewed NITS and repaired (`99d42fc`, lesson-01).
E02 blind exploration closed done (Opus 5.5, exit 0):19/19 confirmed predictions,
0 refuted, 2 new bugs + 1 spec clarification, 46 model cells. Raw records preserved;
source/blindness audit clean. Coordinator reconciliation is register/reconciliations/02.md.
Next fresh fixer closes T1–T3 and repairs U09/U10; A1–A4 remain later.
Branch/worktree: plan/explore-it, /root/Software/booknotes-work/explore-it.

## Done

- `78e56de`, `7dcfb05`: proposal and scoped plan revision.
- `847f1b3`, `explore-it/v0-inherited`: inherited system, protocol and eight skill drafts.
- `6e4dd4a`: E01 blind experiment and canonical reconciliation, original evidence preserved.
- `b839fdf`: reviewed E01 SUT fix, prior target of `explore-it/lesson-01`.
  Acceptance: `register/runs/01-acceptance.json`; 35 findings absent now/present at v0.
- Commit containing this state: checkpoint A review evidence and nits 1/2 repair;
  new `explore-it/lesson-01` target. No earlier commit rewritten.

## Next, in order

1. Coordinator regenerates the unrun E02 export from repaired `explore-it/lesson-01`.
2. E02 closes T1–T3; then E03–E06 refine/probe the stronger model and remaining
   spec silences, followed by checkpoint B review.
3. E07–E09, structural transition guards, final skill revisions/index/landing,
   checkpoint C review, and the requested final Astra review/report.
4. Merge/push only under separate owner authorization. All edits stay in this book.

## Open work and owners

- Checkpoint A: complete, verdict NITS; notes 1/2 repaired. Direct Go invariant
  assertions (note 3) assigned to E07/E09, final worked examples (note 4) to E09.
  Canonical filenames remain stable (note 5).
- A1–A4 policy decisions: coordinator after their actual later probes; not decided here.
- T1–T3 visibility/clock instruments: next fresh E02 fixer after blind exploration.
- Detailed source-text citation validation and final skills: later author/reviewer.
- Predicted qctl/sample-worker numeric loss: later variation/data-journey explorer;
  still unprobed and excluded from E01 find counts. API-side D08 repair is verified.
- No active E01 blocker. Incomplete harness attempts remain in the evidence logs;
  the frozen final current/historical runs both passed. Independent review is
  complete with its initial startup failure retained.

## Resume

Read this file, `register/BUILD-STATE.md`, the latest reconciliation/fixer brief and
`git status --short`; inspect actual commits/tags before trusting prose. Do not
amend original experiment/F-report records. Use the existing branch/worktree and
only `books/explore-it/`; scratch goes under `/tmp`. Coordinator edits to
`register/tools/` belong to a separate unit and must not enter the E01 fix commit.

Completed review evidence is committed under `register/reviews/A/`. There is no
active reviewer. Prepared E02 export `/tmp/explore-it-experiment-02` and metadata
`/tmp/explore-it-coordinator/current.json` are stale until the coordinator
regenerates them from the repaired lesson tag. Repair simulations and helper:
`/tmp/explore-it-a-nits/`; no E02 explorer launched in this repair unit.

Before launching a duplicate worker, check the coordinator's active run and local
process/log liveness. The continuation runner uses `/tmp/bn-sol-exploreit2.log`
and `/tmp/bn-sol-exploreit2.exit`; an absent exit file alone does not prove it is
still running. Check process state and recent log progress. Docker queuelet
services can remain running after a completed unit; they are not an active agent.
Use project Compose actions only; do not change the concurrent Release It! tree.

Active E02 liveness: `/tmp/explore-it-experiment-02/run.jsonl`, `run.stderr`,
`run-start.txt`, `run-end.txt`, `exit-status.txt`; launch wrapper
`register/tools/run-explorer.sh`. Check matching `claude --model claude-opus-5-5`
process and recent log progress before relaunching. Preserve an interrupted run;
never create a duplicate or reuse a stalled agent. Metadata is
`/tmp/explore-it-coordinator/current.json`; collector is `collect-record.py`.

E02 explorer ended; do not relaunch. Fresh fixer brief: register/reconciliations/02-fixer-brief.md (to be dispatched after this experiment commit). Tools and current run metadata remain under /tmp/explore-it-coordinator.
