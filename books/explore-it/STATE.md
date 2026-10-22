# STATE — Explore It! / queuelet

**Goal:** turn *Explore It!* (Hendrickson) into a teaching project scoped to Nick's
direction: skip the fluff and soft-skills material; keep the probing/testing/learning-a-
codebase core — a good system (`queuelet`, a baby job queue, Go + Postgres 17) with nine
structured exploration experiments run against it, plus 8 codebase-exploration skills.
Spec: `books/explore-it/PLAN.md` (sharpened pass) + repo `AGENTS.md`.
Builder: Claude Opus 5.5 per Nick 2026-10-22 (was Codex Sol); Astra reviews at the end.

**Status (2026-10-22): experiment 04 exploration committed; fix in progress (Opus builder).**

## Done
- Plan: `78e56de` + scoped revision `7dcfb05`.
- v0 queuelet: `847f1b3`, tag `explore-it/v0-inherited`.
- Experiment 01: `6e4dd4a`, fixes `b839fdf`, `0386a14`, `99d42fc` (tag `explore-it/lesson-01`).
- Experiment 02: `e390c31`, hooks `3287bd9`, fixes `9394782` (tag `explore-it/lesson-02`).
  **Checkpoint A complete.**
- Experiment 03 (Vary it): exploration `09dd0a9`; fix = the commit containing this state
  (tag `explore-it/lesson-03`). 8 findings, all 8 fixed (U11–U18), none deferred.
  Verify current `FIXED AND CHECKED: 8 absent`; REF lesson-02 `DEFECTS REPRODUCED: 8
  present`; 54/54 integration checks; A1–A4 still PRESENT for later experiments.

## In progress
- E04 fix: reconcile E04-F01..F05 and A1, patch, repro, checks, note, tag `explore-it/lesson-04`.
  The blind explorer run is closed (exit 0, 15 min); its stack `queuelet-e04` is down.

## Incidents (history)
- Sol run 1 stopped 2026-10-21 ~21:13 CDT by the Codex content filter after E01.
- Sol run 2 stopped the same way ~23:1x CDT after E03's exploration phase.
- Nick moved the build to the Opus seat 2026-10-22; the E03 explorer was also Claude, so
  E03's cross-family check rests on the Astra end review.

## Next
1. Experiment 04 (Judge it): fresh run; export from `explore-it/lesson-03`
   (`register/tools/prepare-export.py 4 explore-it/lesson-03`). Carries A1 (default backoff
   cap); its planned defects D09–D11 and D25 were already fixed in E01.
2. Experiments 05–06, then checkpoint B review.
3. Checkpoint C: 07–09, final skills, landing commit; Astra end review.
4. Merge into `main` adds-only after Release It! lands.

## Open decisions
- A1–A4 spec silences: E04–E07 (coordinator decides one line each in the fix commit).
- Explorer seat for E04+: the plan wants explorer and fixer from different families; with
  Opus building, a Muse or Codex explorer would restore that. Owner: Nick / Twin.

## Resume
- Worktree `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`; never touch
  `/root/Software/book-notes`. Fresh agent per chunk; `register/BUILD-STATE.md` has detail.
