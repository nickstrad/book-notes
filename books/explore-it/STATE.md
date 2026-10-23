# STATE — Explore It! / queuelet

**Goal:** turn *Explore It!* (Hendrickson) into a runnable exploration teaching project:
one inherited queuelet service, nine experiments and eight reusable skills.
Spec: [PLAN.md](PLAN.md), [exploration/PROTOCOL.md](exploration/PROTOCOL.md). Landing page:
[README.md](README.md).

**Status (2026-10-23): build, end review and its fix pass complete. Astra reviewed
`e750542`: PASS WITH FIXES (0 high, 4 medium, 2 low); see [REVIEW.md](REVIEW.md). The fix
pass (fresh Opus) fixed R2–R6 and recorded R1, which needs the book's full text (Nick).
No code changed, so `explore-it/v1-explored` stays on the landing commit `b6df7db`. Ready
for Twin to merge; not merged, nothing pushed.**

## Done

- Plan `78e56de`, scoped revision `7dcfb05`; v0 `847f1b3`, tag `explore-it/v0-inherited`.
- Nine experiments, each an exploration commit plus a fix commit tagged `explore-it/lesson-NN`
  ([index with scores](lessons/README.md)): E01 `6e4dd4a` → `99d42fc`; E02 `e390c31` → `9394782`;
  E03 `09dd0a9` → `8ed9177`; E04 `dffb1ec` → `171974a`; E05 `01137b6` → `0099bcd`;
  E06 `b3b0559` → `755d21c`; E07 `49d8d67` → `8787e43`; E08 `a2e7b01` → `9b1b1a1`;
  E09 `2fee853` → `130f199`.
- Register: all 27 planted defects found in E01 (D07 partly) and fixed in lesson 01; 45
  unplanned faults U01–U45 fixed in their lessons; spec silences A1–A5 decided; T1–T3 closed in
  E02. Canonical status: [defect index](exploration/defects/INDEX.md).
- Checkpoint reviews: [A](register/reviews/A/review-a.md) (Muse, NITS addressed),
  [B](register/reviews/B/review-b.md) (Sol, PASS), [C](register/reviews/C/review-c.md) (fresh
  Opus, `42ce139`; doc nits fixed in `c5bb76e`). C measured own-tag and HEAD verify 07/08/09 at
  3/4/7 absent and REF at 3/4/7 present, and HEAD verify 00–09 all pass. `make checks` failed once
  on the U32 sampler (instrument flake, five reruns ABSENT); `make test` and `go vet` pass.
- Final skills: all eight under [skills/](skills/README.md), each with its procedure revised from
  the experiments, a pitfalls section and a worked example from queuelet. Measured blind find rate
  against the register: 26 of 27, against the plan's bar of 19.
- Landing commit `b6df7db`: final skills, landing page, lessons index; tag `explore-it/v1-explored`.
- Final suite at `explore-it/v1-explored`, fresh clone, project `queuelet-final`
  ([ledger and logs](register/runs/final/acceptance.json)): verify 00 `BASELINE HOLDS`; verify
  01–09 `FIXED AND CHECKED` 35/5/8/6/4/8/3/4/7 absent; `make checks` (94 check functions) exit 0
  in 706.9 s; `go test -count=1 ./...` and `go vet ./...` exit 0. Every step exited 0; the stack
  was taken down.
- Walkthrough: [WALKTHROUGH.md](WALKTHROUGH.md), a guided tour for a cold reader (reading order,
  lab commands, tag itinerary); its commands checked live on 2026-10-22 through lesson 07.

- Independent end review (this review commit), target `e750542`: [REVIEW.md](REVIEW.md),
  [evidence](register/reviews/end/README.md). Fresh-clone final suite passed all 94
  checks after explicit seeding (813.920 s); uncached unit tests and vet passed.
  The first run's missing-seed U09 failure is retained as reviewer setup error.
  Historical REF08/09 reproduced 4/7 defects; v0 baseline passed and D18/D21
  reproduced against v0 binaries. All review stacks were taken down. Manifest and
  original-transcript audits are retained. The designated non-Claude end review of
  E07–E09 is complete; source-fidelity acceptance remains qualified by R1.

- End-review fix pass (fresh Opus, one commit per finding; docs and records only, project
  unchanged since `v1-explored`):

  | Finding | Outcome | Commit |
  |---|---|---|
  | R1 citations | **recorded, open (Nick)**: [citation ledger](register/citations.md) checks every lesson and skill citation against the publisher's contents and three excerpts; unsupported attributions corrected (lesson 09's 13.3 reading, 10.3, the state skill's departure from Chapter 8's external view) | `8a5d3a4` |
  | R2 blindness claims | fixed: landing and walkthrough name the boundary and its leaks; [exceptions table](register/blindness-exceptions.md) | `f378692` |
  | R3 learning trend, E08 chronology | fixed: E09's measured prediction yield only; E08 extends E01's ecosystem model | `d078712` |
  | R4 run recipe | fixed: final-tree commands stay on the final tree; own-tag check in a worktree with rebuilds; [rehearsed](register/runs/r4-recipe/README.md) | `1d6a64b`, `3a7f324` |
  | R5 stop and audit records | fixed: one provider stop, one handoff stop (provider event unknown); E04 runtime-output reads added to its audit | `8317ebc` |
  | R6 living backlog | fixed: eleven settled questions marked resolved with evidence; BUILD-STATE marked historical; walkthrough review status updated | `b91b531` |

- Fix-pass verification, fresh clones on 2026-10-23:
  - R4 recipe as written at `1d6a64b`, stack `queuelet-r4`, with source commit and image ID logged at every step ([ledger](register/runs/r4-recipe/acceptance.json)). Results: `BASELINE HOLDS`; REF06 3 present; verify 07 3 absent on the final build and again on lesson-07's own build (new image); final rebuilt, then `make checks` exit 0, all 94 in 716.7 s. The first `make checks`, run while another session held the host at load 12–14, failed two checks (U29, A3) on probe time limits. That run is retained as evidence and is not a product finding.
  - Final suite at HEAD `3a7f324`, stack `queuelet-fixpass` ([ledger](register/runs/fixpass-final/acceptance.json)). Results: verify 00 `BASELINE HOLDS`; verify 01–09 `FIXED AND CHECKED` 35/5/8/6/4/8/3/4/7 absent; `make checks` exit 0 in 702.7 s; `go test -count=1 ./...` and `go vet ./...` exit 0. All stacks were taken down.

## In progress

None. No runner is active; every fix-pass stack and worktree was removed.

## Next

1. Twin: merge adds-only after Release It! lands, applying PLAN section 10's notes outside `books/`.
2. Nick: R1 full-text citation check (below), and the spec and policy questions.

## Open items (owner)

- **Review R1 (Nick): "book-source verification has never reached the required review bar."**
  Review bar item 4 is met only at section-title level. Closing it needs an authorized full copy
  of the P1.0 edition: work the [citation ledger](register/citations.md)'s **title** and
  **unchecked** rows, adding page locations. Do 10.3, 13.3/13.5, 8.3–8.4 and A2.1 first.
  Correct unsupported attributions in the notes and skills, and add errata to the ledger
  rather than editing raw records. R2–R6 are fixed (table above).

- **Spec questions (Nick):** Q-E09-1, whether `0 0-23 * * *` is a wildcard-hour schedule under A5
  (SPEC literally says no, and the code follows it); Q-E09-2, whether success or an operator retry
  should clear `last_error`.
- **Legacy stored schedules (Nick, upgrade policy):** a schedule stored before lesson 09 that can
  never advance keeps `/health` degraded until it is disabled or deleted. Legacy invalid payload
  templates (E06) are neither migrated nor revalidated. Same question: migrate, revalidate or
  leave.
- **Probe time limits under host load (instrument, coordinator):** in the fix pass, U29's 10 s `docker stop` timeout and A3's 10 s worker-start wait both expired at load 12–14 ([R4 record](register/runs/r4-recipe/README.md)). They pass on a quiet host. Widening them is an instrument change, handled like U32 below.
- **U32 sampler window (instrument, coordinator):** the batch-path control can miss its race
  window (one miss in review C). Widening it is an instrument change: re-freeze hashes, replay REF.
- **Direct Go assertions for D18 and D21** (checkpoint A follow-up): D19 is now covered by the
  model walk; D18's stale token and D21's cross-tenant access are checked only through the
  shared repro oracle.
- **Testability requests (coordinator):** TR-E09-2 (deterministic post-commit fault hook),
  TR-E02-3/4, TR-E04-1/2, T-E05-1, T-E06-1/2, T-E07-1/2, T-E08-1…3.
- **Forward hypotheses (fresh explorers):** H-E09-18/19; H-E04-08 (backward clock, now also lease
  before `run_at`); Q-E08-1 worker identity within a tenant; H-E08-22 redirects counted as
  delivered; H-E08-24 pool exhaustion during a freeze; E07's Q-E07-1/2, H-E07-17 and the unprobed
  cancel/retry/promote/retention interleavings; E06's Q-E06-1/2, H-E06-20…22 and the
  http-worker float64 decoding path; E05's H-E05-20/21; E04's Q-E04-4, H-E04-09…11 and R42;
  earlier Q3–Q5 and H45. Detail: [hypotheses](exploration/hypotheses.md) and the review B and C
  tables.

Decided: A1 default retry cap 3600 s; A2 leased-at-purge jobs that fail become dead; A3
cancellation immediate and terminal, cooperative worker stop, late ack/nack 409; A4 dead jobs
exempt from retention; A5 classic cron DST (gap time runs once at gap end, repeated time once).

## Resume

- Worktree `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`. Git wins over
  this file.
- Fix-pass runners: `register/runs/r4-recipe/run.py` and `register/runs/fixpass-final/run.py`.
  They use fixed paths under `/tmp/ei-fixes/` (a fresh clone at `repo/`) and their own stacks.
  Scratch has been deleted; recreate it before rerunning.
- End-review evidence: `register/reviews/end/`; runners and reconstruction commands
  are indexed in its README. No background run needs resuming. Scratch clone:
  `/tmp/explore-it-astra-review/repo`, reviewed HEAD `e750542`.
- Review C: runner `register/reviews/C/run.py` (fresh clone and tag worktrees under
  `/tmp/explore-it-checkpoint-c`, project `queuelet-review-c`, ports 18270/18271/15503, explicit
  `QUEUELET_DATABASE_URL`). Earlier acceptance runners: `register/runs/0N-opus-runner.py`.
- Project knowledge: [docs/knowledge](docs/knowledge/README.md).
