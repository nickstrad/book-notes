# STATE — Explore It! / queuelet

**Goal:** turn *Explore It!* (Hendrickson) into a teaching project scoped to Nick's
direction: skip the fluff and soft-skills material; keep the probing/testing/learning-a-
codebase core — a good system (`queuelet`, a baby job queue, Go + Postgres 17) with nine
structured exploration experiments run against it, plus 8 codebase-exploration skills.
Spec: `books/explore-it/PLAN.md` (sharpened pass) + repo `AGENTS.md`.
Builder: Claude Opus 5.5 per Nick 2026-10-22 (was Codex Sol); Astra reviews at the end.

**Status (2026-10-22): experiment 03 exploration committed; fix step in progress (Opus).**

## Done
- Plan: `78e56de` + scoped revision `7dcfb05`.
- v0 queuelet: `847f1b3`, tag `explore-it/v0-inherited`.
- Experiment 01: `6e4dd4a`, fixes `b839fdf`, `0386a14`, `99d42fc` (tag `explore-it/lesson-01`).
- Experiment 02: `e390c31`, hooks `3287bd9`, fixes `9394782` (tag `explore-it/lesson-02`).
  **Checkpoint A complete.**
- Experiment 03 exploration record + reconciliation (U11–U18): the commit containing this state.

## In progress
- Experiment 03 fix: U11–U18 per `register/reconciliations/03.md`, repros, checks, lesson 03.

## Incidents
- Sol run 1 stopped 2026-10-21 ~21:13 CDT by the Codex content filter after E01.
- Sol run 2 stopped the same way ~23:1x CDT after E03's exploration phase.
- Nick moved the build to the Opus seat 2026-10-22.

## Next
1. Finish the E03 fix; tag `explore-it/lesson-03`.
2. Experiments 04–06 (checkpoint B), one fresh run each.
3. Checkpoint C: 07–09, final skills, landing commit; Astra end review.
4. Merge into `main` adds-only after Release It! lands.

## Resume
- Worktree `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`; never touch
  `/root/Software/book-notes`. Fresh agent per chunk; `register/BUILD-STATE.md` has detail.
