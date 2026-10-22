# STATE — Explore It! / queuelet

**Goal:** turn *Explore It!* (Hendrickson) into a teaching project scoped to Nick's
direction: skip the fluff and soft-skills material; keep the probing/testing/learning-a-
codebase core — a good system (`queuelet`, a baby job queue, Go + Postgres 17) with nine
structured exploration experiments run against it, plus 8 codebase-exploration skills.
Spec: `books/explore-it/PLAN.md` (sharpened pass) + repo `AGENTS.md`.
Builder: Claude Opus 5.5 per Nick 2026-10-22 (was Codex Sol); Astra reviews at the end.

**Status (2026-10-22): experiment 05 complete, tag `explore-it/lesson-05`. Next: experiment 06.**

## Done
- Plan: `78e56de` + scoped revision `7dcfb05`.
- v0 queuelet: `847f1b3`, tag `explore-it/v0-inherited`.
- Experiment 01: `6e4dd4a`, fixes `b839fdf`, `0386a14`, `99d42fc` (tag `explore-it/lesson-01`).
- Experiment 02: `e390c31`, hooks `3287bd9`, fixes `9394782` (tag `explore-it/lesson-02`).
  **Checkpoint A complete.**
- Experiment 03 (Vary it): exploration `09dd0a9`, fix `8ed9177` (tag `explore-it/lesson-03`).
  8 findings, all fixed (U11–U18).
- Experiment 04 (Judge it): exploration `dffb1ec`; fix = the commit containing this state
  (tag `explore-it/lesson-04`). 5 findings, all 5 fixed (U19–U23), none deferred; A1
  resolved by decision (default backoff cap 3600 s). Verify current `FIXED AND CHECKED: 6
  absent`; REF lesson-03 `DEFECTS REPRODUCED: 6 present`; 61/61 integration checks incl.
  the rules-as-invariants walk (200 jobs). Prediction contradicted: only the charter's
  ("the code honours every rule").
- Experiment 05 (Sequence it): exploration `01137b6`; fix = the commit containing this state
  (tag `explore-it/lesson-05`). 3 findings, all 3 fixed (U24–U26), none deferred; A2 resolved
  by decision (purge marks leased jobs; a later nack/expiry/release dead-letters). Verify
  current `FIXED AND CHECKED: 4 absent`; REF lesson-04 `DEFECTS REPRODUCED: 4 present`;
  66/66 integration checks incl. the 100-step random walk. Predictions contradicted: the
  charter's ("every verb does what it says or is refused") and the walk's (which invariants
  break). A3–A4 still PRESENT. The walk check does not catch U24 at lesson-04 (W1 never fired).

## In progress
- Nothing running. The `queuelet` lab stack may still be up (current tree); that is not an agent.

## Incidents (history)
- Sol run 1 stopped 2026-10-21 ~21:13 CDT by the Codex content filter after E01.
- Sol run 2 stopped the same way ~23:1x CDT after E03's exploration phase.
- Nick moved the build to the Opus seat 2026-10-22; the E03 explorer was also Claude, so
  E03's cross-family check rests on the Astra end review.

## Next
1. Experiment 06 (Relate it): fresh run; export from `explore-it/lesson-05`
   (`register/tools/prepare-export.py 6 explore-it/lesson-05`). Carries A4; its planned
   defects D15–D17 were already fixed in E01.
2. Checkpoint B review (E02–E06).
3. Checkpoint C: 07–09, final skills, landing commit; Astra end review.
4. Merge into `main` adds-only after Release It! lands.

Handed forward (open in `exploration/hypotheses.md`): from E04, Q-E04-4 (does "completed"
include dead/cancelled), H-E04-08 (FIFO under a backward clock step), H-E04-09 (stats read
without a snapshot), H-E04-10 (ack not atomic: needs DB fault injection), R42 linear backoff
not re-probed, TR-E04-1/2 (sweep-lag signal, stats snapshot hook). From E05: H-E05-21
(worker deregistering on SIGTERM; now bounded by U24's fix), T-E05-1 (an "offset and keep
ticking" clock mode), the untested sequences in `models/nouns-verbs.md` (schedules, queue
DELETE, retention and webhooks were outside the walk), and E05-F02's missing attempt row for
a delivery cancelled in flight (left with A3).

## Open decisions
- A3–A4 spec silences: E06–E07 (coordinator decides one line each in the fix commit).
- Explorer seat for E06+ (E04 and E05 ran on Opus again; Muse CLI not on PATH this session): the plan wants explorer and fixer from different families; with
  Opus building, a Muse or Codex explorer would restore that. Owner: Nick / Twin.

## Resume
- Worktree `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`; never touch
  `/root/Software/book-notes`. Fresh agent per chunk; `register/BUILD-STATE.md` has detail.
