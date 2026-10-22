# STATE — Explore It! / queuelet

**Goal:** turn *Explore It!* (Hendrickson) into a runnable exploration teaching project:
one inherited queuelet service, nine experiments and eight reusable skills.
Spec: [PLAN.md](PLAN.md), [exploration/PROTOCOL.md](exploration/PROTOCOL.md).

**Status (2026-10-22): E07 exploration complete; U34/U35 and A3 assigned to a fresh fixer; checkpoint B is complete.**

## Done

- Plan `78e56de`, scoped revision `7dcfb05`; v0 `847f1b3`, tag `explore-it/v0-inherited`.
- E01: exploration `6e4dd4a`, fixes `b839fdf` / `0386a14` / `99d42fc`, lesson-01 at
  `99d42fc`. Checkpoint A complete (Muse NITS addressed).
- E02: exploration `e390c31`, hooks `3287bd9`, fixes `9394782`, lesson-02; T1–T3 and U09/U10.
- E03: exploration `09dd0a9`, fixes `8ed9177`, lesson-03; U11–U18.
- E04: exploration `dffb1ec`, fixes `171974a`, lesson-04; U19–U23, A1 decided.
- E05: exploration `01137b6`, fixes `0099bcd`, lesson-05; U24–U26, A2 decided.
- E06: exploration `b3b0559`, amended fix `755d21c`, lesson-06; U27–U33, A4 decided.
  Project knowledge `7148dee`; independent review/closure evidence `8aa2895`.
  Original [Muse NITS review](register/reviews/E06/review.md) and fresh
  [Claude repair PASS](register/reviews/E06/revalidation/revalidation.md) remain separate,
  immutable evidence. Fresh Muse revalidation stalled on provider streams without a verdict.
- Checkpoint B: fresh Sol review under `/root/booknotes-sol-ckptB.md`; documentation nit
  `d6c34d0` disambiguates the historical U09 copy without editing raw observations.
  [Review record](register/reviews/B/review-b.md),
  [commands/exits/timings](register/reviews/B/acceptance.json), original logs and audits
  are committed in `8435db8`. This brief explicitly assigned Sol rather than
  PLAN's Muse seat; the record states the cross-company coverage limit.
  Own-tag and HEAD E03–E06 verify counts 8/6/4/8; HEAD E02 count 5. All five
  previous-tag REF replays reproduced 5/8/6/4/8 items. All 74 integration checks,
  unit/application tests and vet pass; A3 control stays PRESENT. Lab cleanup passed.
  README clarification `b1aee49` documents the direct-check database URL.

All 27 planted defects were fixed in E01. Later experiments tested remaining questions
and new faults, not fictional late discoveries. E06's original score remains
20 stated / 18 confirmed / 2 refuted / 2 qualified reading contradictions / 30 cells.
No raw report, score, historical evidence, behavioral instrument or lesson tag was rewritten.
Dwell remains paired evidence only; no full sweep was run.

## In progress

E07 exploration closed successfully; artifacts and boundary audit retained under
`exploration/experiments/07-state-it/` and `register/runs/07-*`.
Score: 13 stated / 12 confirmed / 1 refuted / 1 code-reading contradiction / 2 reports /
41 newly observed model cells; full matrix 49 cells. U34/U35 assigned; A3 is a coordinator
policy step, not a blind find. See `register/reconciliations/07.md` for the decision and
all audit/deviation details. Explorer's own stack is stopped. Fresh fixer dispatch is next.

## Decisions and open items

Coordinator/Nick owns policy; fresh explorers own future observations. The review's
[reconciliation table](register/reviews/B/review-b.md#register-and-open-item-reconciliation)
is the complete checkpoint disposition, alongside the original
[hypothesis backlog](exploration/hypotheses.md), reconciliations and canonical
[defect index](exploration/defects/INDEX.md). Original “open” wording in raw records
is historical, not an override of later documented decisions.

- **A3 stays open for E07:** cancel while running, including E05-F02's missing attempt.
- **E06 review finding 2:** actual sample-worker http-response number decoding still
  uses float64; the echo repair does not prove that path. Deferred for a fresh run.
- **Legacy invalid schedule templates:** no migration/revalidation; cron/manual consistency
  remains an upgrade-policy decision for Nick. New creation alone is repaired.
- Q-E06-1 worker recovery after deregistration; Q-E06-2 delivered-outbox cleanup;
  T-E06-1 job id in decode logs; T-E06-2 deterministic race instrument;
  H-E06-20 deleted-queue lease race; H-E06-21 NUL http-worker result consequence;
  H-E06-22 NUL registration. H-E06-17 stays a blind sampling non-hit despite later
  paired batch-race repair evidence.
- TR-E02-3/4 backlog/liveness metrics and eligibility explanation; TR-E04-1/2 sweep
  lag/stats snapshot; T-E05-1 ticking offset clock remain open testability proposals.
- E04 Q-E04-4 stats “completed” definition; H-E04-08 backward-clock FIFO,
  H-E04-09 stats without snapshot, H-E04-10 ack atomicity under DB fault,
  H-E04-11 NULL schedule kind; R42 linear half remain forward questions.
- E05 H-E05-21 SIGTERM deregistration remains unprobed but bounded by U24;
  H-E05-20 error-text sibling is structurally covered by U26 without a new live probe.
  Sequences outside its walker (schedules, queue deletion, retention, webhooks) remain
  outside walker coverage. W1 never fired on lesson-04; U24's direct repro supplies evidence.
- Earlier Q3 driver text, Q4 admin authority, Q5 returned secrets and H45 outage
  catch-up remain in the backlog. Direct Go assertions for D18/D19/D21 from checkpoint A,
  final skill worked examples and full book-text citation validation remain later work.

Decided: A1 default retry cap 3600 seconds or larger base; A2 unsuccessful exits of
leased-at-purge jobs become dead; A4 dead jobs exempt from automatic retention.
U33 attempts follow jobs; events and pending snapshots keep independent lifetimes;
legacy orphan attempts were not bulk-cleaned. U28 uses enqueue's object/default rule
on schedule creation, with no new PATCH payload field. U31 NUL result/error is a
service-owned 400 and leaves a valid lease usable. Q1/Q2 and E04/E05 policy questions
resolved by these fixes are distinguished from the remaining questions in the review table.

## Next

1. Fresh Codex fixer completes E07 U34/U35 and A3, verifies current/lesson-06, updates
   this state and tags lesson-07; independent different-company review follows.
2. Later authorized checkpoint C, final skills/landing page and final independent review.
3. Merge adds-only after Release It! lands; no push or merge was performed here.

## Resume

- Worktree `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`.
  Resolve `git rev-parse explore-it/lesson-06` → `755d21c`; parent `b3b0559`.
  Git wins over stale state. Read this file and the checkpoint B review first.
- B acceptance used a fresh temporary clone and project `queuelet-review-b`, ports
  18260/18261/15493, with explicit `QUEUELET_DATABASE_URL` for direct integration checks.
  Its first HEAD E02 attempt omitted that variable and failed against default port
  15433; retained logs show five public repros passed before two direct checks failed.
  Corrected acceptance reran with the explicit database URL. No SUT change was needed.
- Nothing is active or needs relaunching. `register/reviews/B/` retains the exact runner,
  resume script, logs, integrity/history audits and final cleanup status. Do not relaunch
  a completed reviewer as the E07 explorer. Coordinator owns blind export lifecycle.
- [E06 lesson](lessons/06-relate-it.md), original `register/runs/06-acceptance.json`
  and `06-repair-acceptance.json` preserve earlier successes and failures. Own B evidence
  supplements them. Project knowledge: [index](docs/knowledge/README.md), including
  the separately committed [review database configuration](docs/knowledge/review-isolation.md).
  Knowledge assessment found no new VM fact; reusable project guidance belongs here.
