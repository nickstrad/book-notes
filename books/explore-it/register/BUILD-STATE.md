# Explore It! build state

Authorized by `/root/booknotes-sol-exploreit2.md` and the E01 fixer assignment; the approved v2 plan governs disagreements.
Scope: only `books/explore-it/` in the existing `plan/explore-it` worktree. No pushes.

## Current unit

**E06 one-round review repair complete**, amended code fix `755d21c`, tagged
`explore-it/lesson-06`, parent `b3b0559`; supersedes reviewed `7a38788`.
Exploration `b3b0559`: fresh blind Claude Opus 5.5, exit 0, seven reports reconciled
U27–U33. Fresh Codex Sol fixer repaired all seven assigned paths; A4 decided separately.
Fresh Muse Wonder Full review of `7a38788` returned NITS; its original evidence is
immutable under `reviews/E06/`. This fresh Sol repair addresses finding 1's introduced
JSON prefix-acceptance regression with full-document validation plus UseNumber in
qctl options/output and shared client. Invalid CLI JSON exits before any request;
the printer retains its full raw-response fallback. Actual-binary CLI/server tests,
client HTTP tests and worker-default tests exercise the application, not copied decoders.
Finding 2's pre-existing worker http-response precision sibling is explicitly deferred,
unprobed through the actual http worker and unchanged here (outside the narrow echo repair).
Legacy invalid schedule templates are not migrated/revalidated; creation is repaired,
but upgrade policy remains coordinator/Nick's checkpoint-B question. Neither is a new find.
Independent repair validation is complete: fresh Claude Opus 5.5 (different company
from Codex builder) returned PASS, 0 findings, exit 0. Prior Muse NITS report remains
immutable; fresh Muse revalidation failed provider streams before tests and produced
no verdict. The fallback and lifecycle evidence are in `reviews/E06/revalidation/` and
`reviews/E06/revalidation-muse/`. Initial Muse Docker checks were sandbox-blocked;
reviewers independently ran actual-code tests, vet, hash checks and actual CLI probes,
while current/REF/74 integration verdicts rest on the builder's recorded acceptance.
Next is full checkpoint B (E02–E06), then fresh E07 after that gate. All owned runs are
closed and the lab is stopped. Root book STATE.md is the current handoff.

- Repair sequence, sequential and all exit 0: `make up` 50.593 s; current verify
  121.898 s → `FIXED AND CHECKED: 8 absent; regression checks pass`; unchanged REF
  lesson-05 49.876 s → `DEFECTS REPRODUCED: 8 present`; `make checks` 399.806 s,
  all 74 integration checks; `make test` 5.668 s, including new strictness/default tests.
  Own `make down` 3.124 s, exit 0. Details: `runs/06-repair-acceptance.json` and
  linked `06-repair-*.log`. Original acceptance and all earlier failures/pairs stay unchanged.
  Final integrity: 3,447/3,447 raw artifacts, 95/95 frozen instruments, original review
  hash, repros and historical proofs unchanged. No numeric oracle or score change.

- Raw score unchanged: 20 stated / 18 confirmed / 2 refuted / 2 qualified reading
  contradictions / 30 cells. All 3,447 original explorer artifact hashes match.
  Only canonical U status records and their index changed under exploration/.
- Original required sequence, all exit 0: `make up` 41.282 s;
  `make verify LESSON=06` 151.177 s →
  `FIXED AND CHECKED: 8 absent; regression checks pass`;
  `make verify LESSON=06 REF=explore-it/lesson-05` 57.671 s →
  `DEFECTS REPRODUCED: 8 present`; `make checks` 398.913 s, 74/74 checks;
  `make test` 1.239 s. Exact output/failures/elapsed in `runs/06-*.log` and
  `runs/06-acceptance.json`. Frozen final instrument hashes are rechecked after
  acceptance; target-image qctl and worker binaries exercised in both directions.
- Paired dwell only, lesson-05 PRESENT → lesson-06 ABSENT; no full sweep.
  Historical owner-race errors actually hit both reported surfaces; batch sibling
  also reproduced. That fixer evidence does not rewrite H-E06-17's blind non-hit.
- A4: automatic retention exempts dead, with succeeded/cancelled age-out controls
  observed through API clock and real storage. U33: attempts follow job deletion,
  purge and retention; events/pending snapshots remain independent. Legacy orphan
  attempts are left inspectable, without bulk cleanup. Schedule create uses the
  existing enqueue object/default rule; PATCH gains no payload field. NUL result/
  error is 400 and does not mutate a valid lease. SPEC records these decisions.
- Retained failures: missing constraint restart guard (repaired); first current
  probe compared hooksink's rounded convenience body (corrected to raw bytes);
  baseline check assumed float64 after UseNumber (exact int64 parsing, same lease
  assertions). Final full sequence repeated after a new freeze; public repro hashes
  unchanged between the two successful pairs. No weakening or SQL fixture writes.
- D15–D17 and all planted defects were already fixed in E01. The E06 orphan
  prediction changed: guarded queue owners held; client decoders and attempt lifetime
  supplied new findings. Qualified contradictions, forward questions, policy owners
  and exact resume steps are in `../STATE.md` and `../lessons/06-relate-it.md`.
- A3, delivered-outbox cleanup, worker recovery/decode-log questions, deterministic
  race hooks and remaining unprobed leads are deferred to fresh authorized chunks.
  Project knowledge: `../docs/knowledge/README.md`; E06 helper indexed in
  `../project/lab/repro/README.md`. A separate documentation commit after `755d21c`
  appends streaming Decode versus whole-document semantics; lesson-06 stays on the code fix.
  No VM kb write, push or external message.

Resume: read `../STATE.md` first. Resolve the fix SHA with
`git rev-parse explore-it/lesson-06`; its parent is `b3b0559`. Nothing needs relaunching.
Do not start E07 before checkpoint B review. Earlier sections below are historical
checkpoint evidence, not the next dispatch.

## E04 unit (historical)

E04 complete, tag explore-it/lesson-04 (the fix commit containing this state). Builder: Claude
Opus 5.5 (explorer also Opus; the cross-family check rests on the Astra end review).

- Exploration record `dffb1ec` (blind Opus explorer, 15 min, 67/67 manifest hashes rechecked).
  Blindness audit: zero flagged tool inputs; corrected afterwards for one `find /` that listed
  skill directory names in other exports (none opened). The exported lesson 03 note named A1,
  so the explorer's A1 observation is not counted blind.
- Reconciliation `reconciliations/04.md`: five reports → U19–U23, all fixed; A1 resolved by
  decision (default cap 3600 s, or base if larger). D09–D11 held as regressions; D25 not re-probed.
- Acceptance (`runs/04-acceptance.json`): verify current `FIXED AND CHECKED: 6 absent`;
  REF lesson-03 `DEFECTS REPRODUCED: 6 present` (second attempt; the first stopped on a probe
  ordering bug, retained with diagnostics); 61/61 integration checks incl. TestL04Invariants;
  unit tests pass; seeder A1 control ABSENT, A2–A4 still PRESENT.
- Process notes: probes lease with `lease_job`, which demands its own job back; an accepted bad
  input can be due first. Lesson notes must not name which experiment an open silence belongs to.

Resume: fresh agent for E05 from lesson-04; coordinator prepares the export with
`register/tools/prepare-export.py 5 explore-it/lesson-04`.

## E03 unit (historical)

E03 complete at explore-it/lesson-03: blind record 09dd0a9, fix 8ed9177. Eight reports → U11–U18,
all fixed; verify current 8 absent, REF lesson-02 8 present, 54 checks (runs/03-acceptance.json).

## E02 unit (historical)

E02 complete at explore-it/lesson-02: blind record e390c31, hooks 3287bd9, fixes 9394782.
Verify current 5 absent, REF lesson-01 5 present, 46 checks; details in runs/02-acceptance.json.
Process knowledge kept from E02: a missing route can answer 405 when a different-method
wildcard matches; fixture timestamp precision must match the storage boundary
(microseconds); the API clock is semantic and resets on API restart, while workers and
network timers stay real; capture failure diagnostics before scratch-stack teardown.

## Prior checkpoint A acceptance


Checkpoint A complete through E01 and the narrow review-nit repair. Independent
Muse Wonder Full review of `b839fdf` returned **NITS**; evidence is preserved in
[reviews/A/review-a.md](reviews/A/review-a.md).
v0: `847f1b3`, tag `explore-it/v0-inherited`; blind E01: `6e4dd4a`; fix:
`b839fdf`; diagnostic/docs repair: commit containing this state, the new target
of `explore-it/lesson-01`. SUT and behavioral oracles are unchanged from reviewed
`b839fdf`; the tag moves only for this authorized pre-acceptance repair. All D01–D27 and U01–U08
assigned by reconciliation are repaired. Original discovery remains 25 live-probed
registered defects, D23 harness evidence and partial D07; eight additional faults
across 27 original reports. Raw score stays 48 stated / 44 confirmed / 2 refuted /
2 needs more, one code-reading contradiction, 110 observed model cells.

Final acceptance (exit 0):
- `make up`: `runs/01-up.log`.
- `make verify LESSON=01`: **FIXED AND CHECKED: 35 absent; regression checks pass**.
- `make verify LESSON=01 REF=explore-it/v0-inherited`: **DEFECTS REPRODUCED: 35 present**.
- `make checks`: all 39 named integration checks pass, 191.586 seconds.
- `make test`: unit suite passes.
- Later-policy controls A1–A4 still PRESENT, each exit 0; they are answer-key
  arrangements, not historical black-box repros.

Exact output and machine-readable acceptance live under `runs/01-*`.
Measured dwell is only v0 → E01: all 35 present at v0, absent on the repaired
E01 tree. The TSV is derived from those paired verification logs, not an extra
`make dwell` run. Current full verification takes about 6.5 minutes because the
real-time D13/D27 oracles run again as regression checks. D13 measured 109.9 s
historically (2 resumed jobs) and 110.0 s currently (0). U03 measured 11.47 s vs
1.63 s healthy control historically, and 1.43 s vs 2.03 s currently.

All 57 original explorer artifacts are byte-identical to `6e4dd4a`; only canonical
D/U statuses and their index changed. Source-text citation review, final skill
revisions remain pending. Checkpoint A review is complete. Next unit: coordinator
regenerates the unrun E02 export at the repaired lesson tag, then dispatches fresh
E02 exploration/fix. This fixer does not continue into E02.

## Execution discipline

- Plan checkpoint boundaries govern: A ends after E01; B covers E02–E06; C covers E07–E09.
- The coordinator knows the register and never acts as explorer. Each explorer is a fresh different-family headless agent in a plain export of only project, exploration, skills and lessons.
- Scratch coordination helper: `/tmp/explore-it-coordinator/prepare.py`. Its generated brief permits the explorer to read only the export. Charters are written before invocation. Artifacts are copied back unchanged; any ID reconciliation is recorded separately.
- Source-controlled progress lives here because root STATE.md edits are excluded by the brief; root merge notes remain in PLAN.md section 10.
- Review failures are recorded and repaired within scope; the brief explicitly says reviews do not stop subsequent work.

## Pool snapshot before build

2026-10-22: Claude `/usage` reports 21% current window used and 12% weekly used. Codex `/status` reports 98% weekly remaining. Muse `/usage` reports subscription gauge “Currently unavailable”. No paid API fallback is authorized or used.

## Pending

1. DONE: baseline, 27 proofs and four silence probes; commit and tag v0.
2. DONE: E01 experiment, reconciliation, fix, both verification directions and lesson tag.
3. DONE: checkpoint A review and nits 1/2. Next E02–E06; checkpoint B review; E07–E09 and final skills/index; checkpoint C review.
4. Final clean tree and ordered commit, score, skills, deviation report.

## Measured corrections / deviations

- The plan’s DST illustration is corrected to actual IANA cron semantics: nonexistent 02:00 is skipped on spring-forward day. Opposite-season 02:00 is used to expose the fixed-offset defect.
- A1 at base=5 and completed attempt 20 waits 30.34 days (2^19), rather than the plan’s illustrative 60 days. The defect remains uncapped growth.
- Baseline has four meaningful integration checks, adding concurrent disjoint leases and priority behavior; no speculative expected test count is claimed.
- Specific brief isolation and branch rules take precedence over the global per-agent worktree default. Fresh agents work sequentially in the authorized branch; blind explorers receive plain temporary exports.

- E01 found essentially the full register early. Per plan, findings are fixed where found; no artificial dwell will be manufactured. Causes and other models also appeared early and are retained as produced. Later experiments will refine/test those models, remaining spec silences and new leads.
- First explorer launch failed before any model turn because Claude disallows the bypass flag as root. The actual run used an explicit Bash/Read/Write/Edit/Glob/Grep allowlist and no environment bypass.

## E01 deviations, limits and durable process knowledge

- Preserved and completed every inherited modification (17 source/spec files plus
  the migration); no part was discarded or shifted into a fictional E02 experiment.
  Thirty-five assigned faults make this larger than the plan's two-finding estimate.
  Corrected measured diff (`6e4dd4a` → `b839fdf`): 17 source/spec files
  excluding migration = +364/-101; including `migrations/002.sql` = 18 files,
  +383/-101, before separate repro/check/lesson evidence. The historical
  `b839fdf` commit body incorrectly says +375/-97 across 17 files; this note
  corrects that claim without rewriting history.
- D12 decision: pending = scheduled/queued/retry_wait. U03: partner latency must
  not block core progression; the warm-lab bound is five seconds. U08: queued age
  starts at greatest(created_at, run_at). A1–A4 remain undecided; T1–T3 remain E02.
- Added guarded migration execution and owner foreign keys alongside timestamp
  precision: new orphan writes/deletions are prevented even after a stale owner
  lookup; existing orphan rows remain inspectable via NOT VALID constraints.
- Historical CLI probes must run the archived container's qctl; running current
  Go source would falsely repair D01 in a REF experiment. All SQL in repros is
  SELECT-only and read-only by session default. D20 starts with a public first
  lease at max_attempts=1; no fabricated exhaustion fixture is used.
- Docker exec can consume a shell loop's input stream. The initial current attempt
  therefore skipped IDs. Preload/validate unique nonempty IDs, isolate child stdin,
  require a verdict for each ID, and fail setup errors explicitly. Both first
  attempts also exposed expected transport loss during API restart readiness;
  readiness now waits for a subsequent live 200 rather than classifying loss.
- Editing a running shell script shifts its read offset: the second current attempt
  emitted all 35 absent observations, then failed. Keep the raw failed log, freeze
  instruments and restart. Final instrument hashes match through both complete
  acceptance runs. Dwell propagates errors and supports mixed historical outcomes.
- Required reusable helper: project/lab/repro/probe.py, indexed by that directory's
  README and lesson-01.ids, shared by 35 named Go checks. No duplicate speculative
  utility or VM kb entry was appropriate; this knowledge is project-specific.
- F17 predicts client-side precision loss; qctl's JSON option/pretty-print and the
  sample worker's float decoding remain unprobed siblings. Coordinator explicitly
  keeps the measured API-side D08 scope here; later variation/data-journey runs
  should probe those paths and record an evidenced new finding if warranted.

## Checkpoint A review repair and handoff

- Independent verdict **NITS**, reviewed target `b839fdf`: baseline, current
  verification, historical retry, both unit suites, v0 vet and A1–A4 controls
  passed. Exact verdicts: `FIXED AND CHECKED: 35 absent; regression checks pass`
  and `DEFECTS REPRODUCED: 35 present`. Full evidence: `reviews/A/`.
- Retained failure: first independent REF attempt exited 2 after 144 seconds
  with no probe verdicts because the API did not answer within 60 seconds. Its
  container logs were lost by the old trap; manual bring-up and retry passed.
  The cause remains unproven. The original failed log is preserved unchanged.
- Nit 1 corrects the diff counts above. Nit 2 saves REF Compose ps/logs and run
  metadata in the original project's ignored `results/ref-failure.*` before
  teardown, preserves the original status despite diagnostic/cleanup errors,
  and emits `VERIFICATION FAILED:` for current-tree PRESENT results. No SUT,
  repro, regression oracle or immutable experiment record changed.
- Validation: `bash -n project/lab/verify.sh`, `git diff --check`, and six
  controlled command simulations passed (details in
  `reviews/A/a-nits-validation.md`). No full behavioral suite rerun for this
  diagnostics/docs-only patch; the independent reviewed runs supply that evidence.
  Full staged whitespace check exits 2 for preserved raw Compose log trailing
  whitespace; the scoped check of authored repair files passes (see validation).
- Deferred note 3: E07/E09 adds direct Go lease/token/tenant invariant assertions
  alongside shared-oracle checks. Note 4: E09 final skills pass updates worked
  examples. Note 5: intentionally retain stable canonical filenames. Source-text
  citation validation and unprobed numeric sibling leads remain later work.
- Resume: existing branch/worktree; coordinator regenerates
  `/tmp/explore-it-experiment-02` using the repaired `explore-it/lesson-01` and
  refreshes `/tmp/explore-it-coordinator/current.json` before fresh E02 launch.
  Completed review logs are under `reviews/A/`; no active reviewer remains.
- Durable finding: failure diagnostics must be captured before scratch-stack
  teardown and must not replace the failing status. This project-specific note
  is sufficient; the simulation helper stays in `/tmp/explore-it-a-nits/`.
