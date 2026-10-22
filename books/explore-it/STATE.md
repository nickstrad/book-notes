# STATE — Explore It! / queuelet

**Goal:** turn *Explore It!* (Hendrickson) into a runnable exploration teaching project:
one inherited queuelet service, nine experiments and eight reusable skills.
Spec: [PLAN.md](PLAN.md), [exploration/PROTOCOL.md](exploration/PROTOCOL.md). Landing page:
[README.md](README.md).

**Status (2026-10-22): build complete. `explore-it/v1-explored` tags the landing commit (final
skills, landing page, lessons index). Checkpoint C reviewed (PASS with one instrument flake and
nits). Next: Astra's (Codex) end review. Not merged; nothing pushed.**

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

## In progress

Nothing. No stack for this book is up from this run (`queuelet-review-c`, `queuelet-final` and
`queuelet-ref` are down with volumes removed). A stale `queuelet-e01` stack (ports 18220/18221/15453, up since early
2026-10-22) belongs to the E01 era and was left untouched; the coordinator can remove it
with `COMPOSE_PROJECT_NAME=queuelet-e01 docker compose -f deploy/compose.yml down -v` from
`project/`.

## Next

1. **Astra (Codex `gpt-6-astra`) end review** of the whole arc from a fresh clone, as the
   cross-family review that lessons 07–09 lack (their explorers, fixers and the C reviewer were
   all Claude). Brief: the AGENTS review bar plus PLAN 5.3, with this file and
   [review C](register/reviews/C/review-c.md) as the starting point.
2. Nick's decisions on the spec and policy questions below.
3. Merge adds-only after Release It! lands, applying PLAN section 10's notes outside `books/`.

## Open items (owner)

- **Spec questions (Nick):** Q-E09-1, whether `0 0-23 * * *` is a wildcard-hour schedule under A5
  (SPEC literally says no, and the code follows it); Q-E09-2, whether success or an operator retry
  should clear `last_error`.
- **Legacy stored schedules (Nick, upgrade policy):** a schedule stored before lesson 09 that can
  never advance keeps `/health` degraded until it is disabled or deleted. Legacy invalid payload
  templates (E06) are neither migrated nor revalidated. Same question: migrate, revalidate or
  leave.
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
- **Book citations:** section numbers checked against the publisher's contents; full book-text
  validation not done.

Decided: A1 default retry cap 3600 s; A2 leased-at-purge jobs that fail become dead; A3
cancellation immediate and terminal, cooperative worker stop, late ack/nack 409; A4 dead jobs
exempt from retention; A5 classic cron DST (gap time runs once at gap end, repeated time once).

## Resume

- Worktree `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`. Git wins over
  this file.
- Review C: runner `register/reviews/C/run.py` (fresh clone and tag worktrees under
  `/tmp/explore-it-checkpoint-c`, project `queuelet-review-c`, ports 18270/18271/15503, explicit
  `QUEUELET_DATABASE_URL`). Earlier acceptance runners: `register/runs/0N-opus-runner.py`.
- Project knowledge: [docs/knowledge](docs/knowledge/README.md).
