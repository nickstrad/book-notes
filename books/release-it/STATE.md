# STATE — Release It! / litefleet

**Goal:** turn *Release It!* (Nygard, 2nd ed) into a teaching project: one deliberately
flawed platform (`litefleet`, a baby multitenant SQLite database platform, Go) improved
lesson-by-lesson as patch commits, plus skills that codify the book. Spec:
`books/release-it/PLAN.md` (v3) + repo `AGENTS.md`. Approved by Nick 2026-10-21.

**Status (2026-10-22):** build, skills and Fable's final review COMPLETE on `main`; the review's
findings were worked by an Opus fix run: 14 fixed, M3 and L7 partly (thresholds left), M5 open.
Nothing is running. Next: Nick rules on the open decisions below.

## Done
- Plans: v1 `9b221c1` (superseded), v2 `a5b49d4` (superseded), v3 `d4382c0` (approved).
- v0 flawed platform: `6076a82`, tag `release-it/v0-flawed` — all 18 deliberate flaws in.
- Lessons, one tag each (`release-it/lesson-NN`): 01 `c8b9174`+`0bd3c1c`, 02 `8e71627`,
  03 `9f52c85`, 04 `3629bf1`, 05 `ae4efc4`+`fd7bee4`, 06 `9028409`, 07 `c147832`,
  08 `e602a08`+`196a295`, 09 `80c9d8c`, 10 `6f83f9e`, 11 `2ed890b`+`cf1f712`,
  12 `f8bb556`+`996c596`.
- Landing: lessons index + landing page `a338b8d`.
- Skills: the eight `skills/<name>/SKILL.md` of PLAN.md section 5, commit `f27e8aa`.
- Tag `release-it/v1-hardened` now points at `afb7274`, the last review fix, so the "final state"
  includes the skills and the fixes. It was `a338b8d` (before the skills; REVIEW.md M6), moved
  2026-10-22 by the fix run. No other tag moved.
- Deviations from the plan are recorded per lesson in each note and summarised in the root
  `STATE.md`.
- Final review: `REVIEW.md` (Fable, 2026-10-22). Full verify matrix at `main` and at every parent
  tag; 1 high, 6 medium, 10 low findings.
- Review fixes (Opus 5.5, 2026-10-22), forward commits on `main`, one per finding; verification
  outcomes are in each commit body:
  - H1 `555a958`: lesson 04's verify selects the hot database by node (an `analytics` or
    `staging` database on node-b) and exits 2 if none. At `main` IMPROVED; at
    `REF=release-it/lesson-03` FLAW REPRODUCED (232 duplicates, ratio 1.09, 59 locked).
    Not folded into lesson 06's commit (see open decisions).
  - M1 `a35f09b`: lesson 12 and `plan-safe-release` describe the book's Godot story separately
    from the lab's lock queue.
  - M2 `6edbd15`: lessons 01–03's checks accept later lessons' behaviour (degraded health, 503
    shedding, streamed answers, `or vector(0)`). Each IMPROVED at `main` and FLAW REPRODUCED at
    its parent; 02 also IMPROVED at its own tag.
  - M3 `08d70cf`: lessons 04 and 06 say their before numbers need a loaded host; 04 gets a
    "Try changing" knob. Thresholds unchanged (open decisions).
  - M4 `389d257`: `reset_stack` retries the first `psql` for up to 30 s.
  - L1 `e58ba0c`, L2 `8351211`, L3 `b315fdc`, L4 `a29d477`, L5 `a84ceca`, L6 `3432703`,
    L10 `949d5ed`: citation and wording fixes in the notes, index, landing page and skills.
  - L7 `afb7274`: lesson 00's p99 line 80 ms, lesson 08's report line 0.8 s. Lesson 02's
    `outage_p99_s < 3.0` left (its before run is 3.02 s).
  - L8 `ba2f0fc`, L9 `a30c810`: `docs/knowledge/` started (Toxiproxy, shared-VM lab, resets and
    seeds, plan size estimates).

## In progress
- Nothing.

## Next (in order)
1. Nick rules on the open decisions below.
2. Twin verification (history/tags spot-check, demo runs) and report to Nick.

## Open decisions / blockers
- **Cross-family review (Nick; REVIEW.md M5).** AGENTS.md phase 4 and PLAN.md section 6 require
  reviews by a different family than the builder; none was run. This is not backfilled: the
  review trail is Fable's `REVIEW.md` plus the fix run's commits, all Claude. Accepted as an open
  gap until Nick either waives it here or orders a Codex/Muse review of the whole.
- **H1 not folded into history (Nick).** The review suggested folding the selector fix into
  lesson 06's commit and re-pointing the tags. The fix run made a forward commit instead, so the
  `lesson-06` … `lesson-12` tags still carry the old selector: from a checkout of those tags,
  `make verify LESSON=04` still dies without a verdict. From `main` (or `v1-hardened`), with any
  `REF=`, it works. Rewriting history is Nick's call.
- **M3 thresholds (Nick).** The review's ratio check for lesson 06 (a) (≥ 3×) would fail the
  lesson's own after run (4.14×), and making lesson 04's busy-loop spin on an idle host means
  raising its hot-tenant load, which changes the demo and needs tuning runs. Both notes now say
  the before numbers need a loaded host; the thresholds are unchanged.
- **Latency lines under heavy host load.** At load average 9–16 on 4 cores the fix run saw
  lesson 00's v0 baseline break on p99 (823 ms, 120 ms; 100 % served) and lesson 08 at `main`
  miss whale p99 and the widened report line (latency only; correctness held). On a quieter
  host (the review's runs) both held. Run the lab on a quiet VM, or widen further.
- **Unconfirmed book summaries (Nick).** M1's and L3's case-study sentences were rewritten to
  the conservative wording the review proposed; Nick should check them against his copy.
  PLAN.md's F18 "(Godot)" and L10 heading rows were left as approved.

## Resume instructions
- Branch `main` in `/root/Software/book-notes`; no worktree or background run is in use.
- Reproduce any lesson: `cd books/release-it/project && make up && make seed && make verify
  LESSON=NN [REF=<parent tag>]`.
- Update this file when a review starts or finishes, and when fixups re-point tags.
