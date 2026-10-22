# Explore It! — project state

Goal: turn Hendrickson's *Explore It!* into queuelet, nine honest exploration
experiments and eight reusable skills about probing and learning a codebase.
The approved scope omits the people/process material. Specification and order:
[PLAN.md](PLAN.md), [AGENTS.md](../../AGENTS.md), and
[continuation brief](/root/booknotes-sol-exploreit2.md).

## Status

2026-10-22 UTC: checkpoint A complete. Independent Muse Wonder Full review of
`b839fdf` returned **NITS**; current/historical verification, baseline, unit suites
and A1–A4 controls passed. Its first REF startup attempt failed and is retained,
with cause unproven. Review nits 1/2 are repaired in the commit containing this
state; `explore-it/lesson-01` moves to that repair commit. SUT and oracles remain
identical to reviewed `b839fdf`. Branch `plan/explore-it`, worktree
`/root/Software/booknotes-work/explore-it`. E02 has not started; its prepared export
must be regenerated for the repaired tag. Evidence and exact limits:
[BUILD-STATE.md](register/BUILD-STATE.md), [review](register/reviews/A/review-a.md),
[repair validation](register/reviews/A/a-nits-validation.md).

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
