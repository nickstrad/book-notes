# STATE — Explore It! / queuelet

**Goal:** turn *Explore It!* (Hendrickson) into a runnable exploration teaching project:
one inherited queuelet service, nine experiments and eight reusable skills.
Spec: [PLAN.md](PLAN.md), [exploration/PROTOCOL.md](exploration/PROTOCOL.md).

**Status (2026-10-22): E09 exploration committed (blind Opus, 7 reports reconciled U39–U45); E09 fix in progress. Independent non-Claude reviews of E07 and E08 pending.**

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

- E07: exploration `49d8d67`, fix commit tagged `explore-it/lesson-07`. U34 and U35 fixed,
  A3 resolved (cancel revokes immediately; worker stops cooperatively on the heartbeat
  `cancel_requested` signal; late ack/nack stay 409). Original score: 13 stated /
  12 confirmed / 1 refuted / 1 contradicted reading prediction / 41 new cells; 49-cell
  matrix plus 13 named extra cells. D18–D20 stay early-fixed (E01) controls, ABSENT at both tags.
- Acceptance (`register/runs/07-opus-acceptance.json`): current `FIXED AND CHECKED: 3 absent`
  (251 s), REF lesson-06 `DEFECTS REPRODUCED: 3 present` (86 s), `make checks` (81 functions),
  `make test`, `go vet`, integrity all exit 0. Paired dwell only; no sweep.
- **Provider interruption and seat change:** the first Codex Astra fixer ended exit 1 on the
  provider `possible cybersecurity risk` filter. A second Astra repair run was stopped
  mid-REF by Twin per Nick's rule (Codex stops, Opus takes over, no retries around the
  filter). Claude Opus 5.5 reviewed the uncommitted partial work and finished the fix step.
  REF probe fixture corrected (lease epochs after worker deletion), not an oracle; see the
  [lesson deviations](lessons/07-state-it.md#deviations).

- E08: exploration `a2e7b01` (blind Opus, exit 0, 19:18–19:40Z), fix tagged
  `explore-it/lesson-08`. Score: 21 stated / 20 confirmed / 1 refuted / 1 contradicted
  reading prediction (`/health` under a frozen database) / 4 reports / 7 new ecosystem cells.
  U36 (API start without database), U37 (slow destinations starve the webhook shipper),
  U38 (frozen database hangs `/health` and auth) fixed; A5 (DST gap/overlap) decided:
  classic cron (skipped time once at gap end, repeated time once). D21–D24 early-fixed
  controls ABSENT at both tags; no planted miss to disclose.
- Acceptance (`register/runs/08-opus-acceptance.json`): current `FIXED AND CHECKED: 4 absent`
  (141 s), REF lesson-07 `DEFECTS REPRODUCED: 4 present` (135 s), `make checks` (85 check
  functions, 574 s), `make test`, `go vet`, integrity all exit 0. Earlier attempts retained
  (`08-opus-*-attempt1/2.log`): a cosmetic message fix, then U37's leftover rows broke U03's
  drain control and the suite exceeded Go's 10-minute default; U37 now drains its rows and
  `make checks` passes `-timeout 30m`. Paired dwell only. Fixer seat was Opus per the brief.

All 27 planted defects were fixed in E01. Later experiments tested remaining questions
and new faults, not fictional late discoveries. E06's original score remains
20 stated / 18 confirmed / 2 refuted / 2 qualified reading contradictions / 30 cells.
No raw report, score, historical evidence, behavioral instrument or lesson tag was rewritten.
Dwell remains paired evidence only; no full sweep was run.

## In progress

Nothing is running; no stack is up for this book (`queuelet-e08`, `queuelet-fix08` and
`queuelet-ref` are down with volumes removed).

## Decisions and open items

Coordinator/Nick owns policy; fresh explorers own future observations. The review's
[reconciliation table](register/reviews/B/review-b.md#register-and-open-item-reconciliation)
is the complete checkpoint disposition, alongside the original
[hypothesis backlog](exploration/hypotheses.md), reconciliations and canonical
[defect index](exploration/defects/INDEX.md). Original “open” wording in raw records
is historical, not an override of later documented decisions.

- **E08 forward questions:** Q-E08-1 worker identity within a tenant (any key holder can
  lease as any worker id), H-E08-22 redirects counted as delivered, H-E08-23 one schedule
  that cannot advance stops later schedules (`Fire` returns on first error), H-E08-24 pool
  exhaustion during a freeze, T-E08-1…3 (fault hook, shipper lag metric/health, partner
  signature and redirect modes); handlers past auth and the loop are still unbounded
  against a frozen database.
- **E07/E08 review seat:** both fixers were Claude (Opus) rather than the planned Codex
  seeder family, so their independent reviews should be Codex or Muse, not Claude.
- E07 forward questions: Q-E07-1 heartbeat shortening, Q-E07-2 error message consistency,
  H-E07-17 large-sweep latency, T-E07-1/2 sweep duration and deterministic race
  instruments; unprobed cancel/retry/promote/retention interleavings.
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

Decided: A3 cancellation is immediate and terminal; a leased job records one `cancelled`
attempt; the revoked holder's heartbeat answers 409 with `cancel_requested:true` (read-only);
remote effects already performed are not rolled back. A1 default retry cap 3600 seconds or larger base; A2 unsuccessful exits of
leased-at-purge jobs become dead; A4 dead jobs exempt from automatic retention.
U33 attempts follow jobs; events and pending snapshots keep independent lifetimes;
legacy orphan attempts were not bulk-cleaned. U28 uses enqueue's object/default rule
on schedule creation, with no new PATCH payload field. U31 NUL result/error is a
service-owned 400 and leaves a valid lease usable. Q1/Q2 and E04/E05 policy questions
resolved by these fixes are distinguished from the remaining questions in the review table.

## Next

1. Independent reviews of E07 (`explore-it/lesson-07`) and E08 (`explore-it/lesson-08`)
   from a non-Claude family.
2. E09 (Explain it) as a separate fresh explorer run from `explore-it/lesson-08`
   (`prepare-export.py 9 explore-it/lesson-08`); then checkpoint C review, final skills
   and landing page.
3. Merge adds-only after Release It! lands; no push or merge was performed here.

## Resume

- Worktree `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`.
  `explore-it/lesson-07` is the E07 fix commit; its parent is `49d8d67`.
  Git wins over stale state. Read this file and the checkpoint B review first.
- B acceptance used a fresh temporary clone and project `queuelet-review-b`, ports
  18260/18261/15493, with explicit `QUEUELET_DATABASE_URL` for direct integration checks.
  Its first HEAD E02 attempt omitted that variable and failed against default port
  15433; retained logs show five public repros passed before two direct checks failed.
  Corrected acceptance reran with the explicit database URL. No SUT change was needed.
- E08: explorer export `/tmp/explore-it-experiment-08` (coordinator metadata in
  `/tmp/explore-it-coordinator/current.json`); fix acceptance `register/runs/08-opus-runner.py`,
  project `queuelet-fix08` (ports 18390/18391/15623); REF `queuelet-ref`. All stacks down.
- E07 acceptance: `register/runs/07-opus-runner.py` with project `queuelet-fix07` (ports
  18290/18291/15523); REF uses `queuelet-ref` (18210/18211/15443). Both stacks were taken
  down. Earlier Codex runs' logs (`07-*.log` without `opus`) are retained as history.
- Nothing is active or needs relaunching. `register/reviews/B/` retains the exact runner,
  resume script, logs, integrity/history audits and final cleanup status. Do not relaunch
  a completed reviewer as the E07 explorer. Coordinator owns blind export lifecycle.
- [E06 lesson](lessons/06-relate-it.md), original `register/runs/06-acceptance.json`
  and `06-repair-acceptance.json` preserve earlier successes and failures. Own B evidence
  supplements them. Project knowledge: [index](docs/knowledge/README.md), including
  the separately committed [review database configuration](docs/knowledge/review-isolation.md).
  Knowledge assessment found no new VM fact; reusable project guidance belongs here.
