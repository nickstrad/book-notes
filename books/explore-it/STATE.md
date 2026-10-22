# Explore It! — project state

Goal: turn Hendrickson's *Explore It!* into queuelet, nine honest exploration
experiments and eight reusable skills about probing and learning a codebase.
The approved scope omits the people/process material. Specification and order:
[PLAN.md](PLAN.md), [AGENTS.md](../../AGENTS.md), and
[continuation brief](/root/booknotes-sol-exploreit2.md).

## Status

2026-10-22 UTC: checkpoint A's E01 fix is complete and ready for independent
review. Current/historical verification, all 39 regression checks, unit suite and
A1–A4 preservation controls pass. Branch `plan/explore-it`, worktree
`/root/Software/booknotes-work/explore-it`. The inherited patch belonged to E01,
as git and reconciliation evidence show. This fresh unit stops here; E02 has not
started. Git history/tags take precedence over this file. Detailed evidence and
deviations: [BUILD-STATE.md](register/BUILD-STATE.md).

## Done

- `78e56de`, `7dcfb05`: proposal and scoped plan revision.
- `847f1b3`, `explore-it/v0-inherited`: inherited system, protocol and eight skill drafts.
- `6e4dd4a`: E01 blind experiment and canonical reconciliation, original evidence preserved.
- E01 fix: the commit containing this state, tagged `explore-it/lesson-01`.
  Acceptance: `register/runs/01-acceptance.json`; 35 findings absent now/present at v0.

## Next, in order

1. Coordinator obtains independent checkpoint A review; retain the actual verdict.
2. E02 closes T1–T3; then E03–E06 refine/probe the stronger model and remaining
   spec silences, followed by checkpoint B review.
3. E07–E09, structural transition guards, final skill revisions/index/landing,
   checkpoint C review, and the requested final Astra review/report.
4. Merge/push only under separate owner authorization. All edits stay in this book.

## Open work and owners

- Checkpoint A review: coordinator dispatches the fresh independent reviewer.
- A1–A4 policy decisions: coordinator after their actual later probes; not decided here.
- T1–T3 visibility/clock instruments: next fresh E02 fixer after blind exploration.
- Detailed source-text citation validation and final skills: later author/reviewer.
- Predicted qctl/sample-worker numeric loss: later variation/data-journey explorer;
  still unprobed and excluded from E01 find counts. API-side D08 repair is verified.
- No active E01 blocker. Incomplete harness attempts remain in the evidence logs;
  the frozen final current/historical runs both passed. Independent review remains
  pending before checkpoint B proceeds.

## Resume

Read this file, `register/BUILD-STATE.md`, the latest reconciliation/fixer brief and
`git status --short`; inspect actual commits/tags before trusting prose. Do not
amend original experiment/F-report records. Use the existing branch/worktree and
only `books/explore-it/`; scratch goes under `/tmp`. Coordinator edits to
`register/tools/` belong to a separate unit and must not enter the E01 fix commit.

Before launching a duplicate worker, check the coordinator's active run and local
process/log liveness. The continuation runner uses `/tmp/bn-sol-exploreit2.log`
and `/tmp/bn-sol-exploreit2.exit`; an absent exit file alone does not prove it is
still running. Check process state and recent log progress. Docker queuelet
services can remain running after a completed unit; they are not an active agent.
Use project Compose actions only; do not change the concurrent Release It! tree.
