# STATE — Release It! / litefleet

**Goal:** turn *Release It!* (Nygard, 2nd ed) into a teaching project: one deliberately
flawed platform (`litefleet`, a baby multitenant SQLite database platform, Go) improved
lesson-by-lesson as patch commits, plus skills that codify the book. Spec:
`books/release-it/PLAN.md` (v3) + repo `AGENTS.md`. Approved by Nick 2026-10-21.

**Status (2026-10-22):** build and skills COMPLETE on `main`; Fable's final review DONE
(`REVIEW.md`, verdict PASS with fixes recommended). Nothing is running. Next: Nick rules on the
findings.

## Done
- Plans: v1 `9b221c1` (superseded), v2 `a5b49d4` (superseded), v3 `d4382c0` (approved).
- v0 flawed platform: `6076a82`, tag `release-it/v0-flawed` — all 18 deliberate flaws in.
- Lessons, one tag each (`release-it/lesson-NN`): 01 `c8b9174`+`0bd3c1c`, 02 `8e71627`,
  03 `9f52c85`, 04 `3629bf1`, 05 `ae4efc4`+`fd7bee4`, 06 `9028409`, 07 `c147832`,
  08 `e602a08`+`196a295`, 09 `80c9d8c`, 10 `6f83f9e`, 11 `2ed890b`+`cf1f712`,
  12 `f8bb556`+`996c596`.
- Landing: lessons index + landing page `a338b8d`, tag `release-it/v1-hardened`.
- Skills: the eight `skills/<name>/SKILL.md` of PLAN.md section 5, commit `f27e8aa`
  (after the v1-hardened tag; tags were not moved).
- Deviations from the plan are recorded per lesson in each note and summarised in the root
  `STATE.md`.
- Final review: `REVIEW.md` (Fable, 2026-10-22). Full verify matrix at `main` and at every parent
  tag; 1 high, 6 medium, 10 low findings; nothing fixed.

## In progress
- Nothing.

## Next (in order — Twin drives; Nick: run straight through)
1. Nick rules on `REVIEW.md`: H1 (lesson 04's verify at `main`) first, then M1–M6; fixes fold
   into the lessons they belong to and tags are re-pointed (history is still rewritable).
2. Twin verification (history/tags spot-check, demo runs) and report to Nick.

## Open decisions / blockers
- **Cross-family review (Nick).** AGENTS.md requires a reviewer from a different family than
  the builder (Claude Opus); PLAN.md section 6 asked for a Codex review per checkpoint, and none
  is recorded in git. Fable is also Claude. Nick decides whether Fable's review stands alone or a
  Codex/Muse review is added (REVIEW.md M5).

## Resume instructions
- Branch `main` in `/root/Software/book-notes`; no worktree or background run is in use.
- Reproduce any lesson: `cd books/release-it/project && make up && make seed && make verify
  LESSON=NN [REF=<parent tag>]`.
- Update this file when a review starts or finishes, and when fixups re-point tags.
