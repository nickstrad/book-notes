# STATE — Explore It! / queuelet

**Goal:** turn *Explore It!* (Hendrickson) into a runnable exploration teaching project:
one inherited queuelet service, nine structured experiments and eight reusable skills.
Spec: [PLAN.md](PLAN.md), [exploration/PROTOCOL.md](exploration/PROTOCOL.md).

**Status (2026-10-22): checkpoint B review in progress; E03–E06 tag/HEAD and historical checks running in a fresh temporary clone. E07 not started.**

## Done

- Plan `78e56de`, scoped revision `7dcfb05`; v0 `847f1b3`, tag `explore-it/v0-inherited`.
- E01: exploration `6e4dd4a`, fixes `b839fdf` / `0386a14` / `99d42fc`, lesson-01 at
  `99d42fc`. Checkpoint A complete (Muse NITS addressed).
- E02: exploration `e390c31`, hooks `3287bd9`, fixes `9394782`, lesson-02.
- E03: exploration `09dd0a9`, fixes `8ed9177`, lesson-03; U11–U18 fixed.
- E04: exploration `dffb1ec`, fixes `171974a`, lesson-04; U19–U23 fixed, A1 decided.
- E05: exploration `01137b6`, fixes `0099bcd`, lesson-05; U24–U26 fixed, A2 decided.
- **E06:** exploration `b3b0559` (blind Claude Opus 5.5); amended fix `755d21c`,
  tagged `explore-it/lesson-06`, parent `b3b0559` (fresh Codex Sol one-round repair).
  Supersedes reviewed `7a38788`; immutable [Muse NITS review](register/reviews/E06/review.md).
  Seven reports → U27–U33, all assigned repairs implemented. A4 decided separately, not a blind
  finding. Original score stays **20 stated / 18 confirmed / 2 refuted / 2 qualified
  code-reading contradictions / 30 observed cells**. All 3,447 explorer artifacts remain
  byte-identical; canonical U statuses and index carry repairs separately.
- Original E06 acceptance: `FIXED AND CHECKED: 8 absent; regression checks pass`;
  REF lesson-05 `DEFECTS REPRODUCED: 8 present`; all 74 integration checks pass; unit suite passes.
  Exact commands/exits/timings, failed attempts and diagnostics:
  [register/runs/06-acceptance.json](register/runs/06-acceptance.json).
  Frozen identical final instruments rechecked after the pair and full suite. Dwell is
  paired only (lesson-05 present → lesson-06 absent); **no full sweep**.
- Review repair acceptance, all exit 0, sequential: `make up` 50.593 s; current verify
  121.898 s → `FIXED AND CHECKED: 8 absent; regression checks pass`; unchanged REF
  49.876 s → `DEFECTS REPRODUCED: 8 present`; `make checks` 399.806 s, all 74 checks;
  `make test` 5.668 s, including new actual-binary CLI, client HTTP and worker-default tests.
  [06-repair-acceptance.json](register/runs/06-repair-acceptance.json) retains commands,
  exits, timings and logs separately from earlier successful pairs and failures.
  All 3,447 raw artifact hashes, all 95 frozen instruments, original acceptance, repros,
  historic proofs and review report rechecked unchanged. Own lab stopped, `make down` exit 0.

- E06 project-knowledge commit `7148dee`; coordinator closure/review evidence is in the
  commit containing this final handoff state. Independent revalidation rechecked 3,447 raw
  artifacts, 95 instruments, actual application tests and vet, with no findings.

## In progress

Checkpoint B reviewer: fresh Codex Sol run authorized by `/root/booknotes-sol-ckptB.md`.
Runtime clone `/tmp/explore-it-checkpoint-b/repo`; runner `run.py`, progress and acceptance
JSON in its parent; isolated Compose project `queuelet-review-b` (18260/18261/15493).
The shell runner is active while `pgrep -af /tmp/explore-it-checkpoint-b/run.py` finds it.
Do not launch a second runner over this stack. Review evidence will be copied under
`register/reviews/B/`. The stale historical U09 copy is now explicitly disambiguated in
the canonical index without editing the original observation. No runtime fix so far.

Earlier E06 explorer, fixer, repair and reviewer runs are closed; their owned lab is stopped.
The initial Muse NITS report is preserved. Its runtime was stopped after report/result
completion when closing/reminder provider retries continued. A fresh Muse revalidation
then stalled on provider-stream failures before running tests (no verdict); fresh Claude
Opus 5.5 cross-company fallback returned **PASS, 0 findings**, exit 0, on fix `755d21c`.
Evidence: [repair revalidation](register/reviews/E06/revalidation/revalidation.md).
This is the bounded E06 repair review, not the full checkpoint B review. E07 not started.

## Decisions and corrected predictions

- A4: dead jobs remain until an operator retries or explicitly deletes them. Automatic
  retention applies only to succeeded/cancelled; real API-clock/storage controls prove both.
- U33 / Q-E06-2 attempts half: attempts belong to their job and leave on deletion, purge
  and retention. Job events and pending webhook snapshots retain independent lifetimes.
  Legacy preexisting orphan attempts remain inspectable; no bulk cleanup was performed.
- U28: schedule creation uses existing enqueue payload semantics (omitted `{}`, explicit
  null/non-object invalid). PATCH gains no payload field. U31: NUL result/error is a
  service-owned 400; valid leases remain usable.
- D15–D17 (and all planted defects) were fixed early in E01. E06 did not rediscover them.
  The plan's queue/schedule orphan prediction was contradicted by repaired owner guards;
  the lesson instead follows new client precision paths and the job→attempt lifetime.
- H-E06-C was already counter-predicted from migration 002 before probing; H-E06-17 was
  a batch sampling non-hit, not proof of absence. Preserve the explorer's qualified score.
  The fixer subsequently reproduced the batch sibling and repaired its shared error mapping;
  that is paired regression evidence, not another blind finding.
- Retained E06 failures: unguarded new constraint on restart (guard repaired), first current
  probe reading hooksink's rounded convenience body (instrument corrected to raw bytes),
  baseline lease check assuming float64 after UseNumber (exact int64 parsing, assertions
  retained). Final required sequence reran with a newly frozen manifest. No repro weakened.

## Next

1. Coordinator arranges **checkpoint B review, E02–E06**, using the amended lesson tag,
   both acceptance records and preserved exploration/review evidence. E06 repair validation
   is complete (Claude PASS after Muse provider failures); full checkpoint B remains unrun.
2. After that gate, a **fresh E07** chunk from `explore-it/lesson-06`; A3 remains undecided.
3. Later authorized checkpoint C / final skills / Astra end review; merge adds-only after
   Release It! lands. This state is a handoff, not authorization to begin those steps.

## Open questions and owners

Coordinator/Nick owns policy decisions; fresh explorers own future observations:

- E06 review finding 2: pre-existing sample-worker http-response number decoding still
  uses float64. Deferred, unprobed through the actual http worker in this repair; U29
  fixes the assigned echo-result/webhook path. Checkpoint B must retain this follow-up.
- E06 review observation: legacy null/non-object schedule templates are not migrated or
  revalidated. New creation is repaired; stored invalid templates can still differ between
  cron and manual run. Coordinator/Nick owns the explicit checkpoint-B upgrade policy.
  Neither deferred question is a new blind finding or a change to the seven-report score.
- A3 (cancel while running), including E05-F02's missing attempt for an in-flight cancel.
- E06 Q-E06-1 sample-worker recovery after deregistration; T-E06-1 job id in decode logs;
  Q-E06-2 delivered-outbox cleanup; T-E06-2 deterministic race instrument;
  H-E06-20 deleted-queue lease race; H-E06-21 worker NUL result consequence;
  H-E06-22 NUL registration. H-E06-17's blind non-hit remains in the original record despite
  the fixer's later paired batch reproduction. No broader policies or test hooks were added.
- E04 Q-E04-4 (whether stats “completed” includes dead/cancelled), H-E04-08 (backward-clock
  FIFO), H-E04-09 (stats without a snapshot), H-E04-10 (ack atomicity under DB fault),
  R42 linear backoff not re-probed, TR-E04-1/2 (sweep lag / stats snapshot hooks).
- E05 H-E05-21 (SIGTERM deregistration, bounded by U24), T-E05-1 (offset-and-tick clock),
  and sequences outside the E05 walker (schedules, queue deletion, retention, webhooks).
  E06 covers named CRUD/data paths, not all those sequences. E05 W1 never fired on lesson-04;
  U24's direct regression supplies that evidence.
- Full book-text citation validation and final skill worked examples remain pending.

## Resume

- Use only `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`.
  Read this file first, then [register/BUILD-STATE.md](register/BUILD-STATE.md),
  [lessons/06-relate-it.md](lessons/06-relate-it.md) and E06 acceptance.
- Resolve the fix SHA with `git rev-parse explore-it/lesson-06`; its parent is exploration
  `b3b0559`. Git wins if state disagrees. This repair's lab is stopped and no acceptance
  process needs relaunching. Coordinator owns the separate review lifecycle metadata.
- To repeat E06 acceptance, from `books/explore-it/project`, run sequentially `make up`,
  `make verify LESSON=06`, `make verify LESSON=06 REF=explore-it/lesson-05`, `make checks`,
  `make test`; stop the lab with `make down`. Repros run target-image qctl/worker binaries.
- Do not regenerate an E06 blind export or alter immutable records. Independent repair
  validation passed. Next is full checkpoint B review, then fresh E07 after its gate.
- Project knowledge: [docs/knowledge/README.md](docs/knowledge/README.md);
  commit `7148dee` records the streaming Decode/whole-document lesson after code
  fix `755d21c`; the lesson tag remains on the code fix. E06 reusable oracle is indexed in
  `project/lab/repro/README.md`. No VM kb write was needed.
