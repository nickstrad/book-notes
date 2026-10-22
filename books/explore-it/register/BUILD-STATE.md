# Explore It! build state

Authorized by `/root/booknotes-sol-exploreit2.md` and the E01 fixer assignment; the approved v2 plan governs disagreements.
Scope: only `books/explore-it/` in the existing `plan/explore-it` worktree. No pushes.

## Current unit

Checkpoint A build complete through E01 fix, ready for independent review.
v0: `847f1b3`, tag `explore-it/v0-inherited`; blind E01: `6e4dd4a`; fix:
commit containing this state, tag `explore-it/lesson-01`. All D01–D27 and U01–U08
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
revisions and checkpoint review remain pending. Next unit: coordinator's fresh
independent checkpoint A review, then fresh E02 exploration/fix. This fixer does
not continue into E02.

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
3. Checkpoint A review; E02–E06; checkpoint B review; E07–E09 and final skills/index; checkpoint C review.
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
  Measured SUT-source/migration diff: 375 added and 97 removed lines across
  17 files, before the separate repro/check/lesson evidence.
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
