# Explore It! build state

Authorized by `/root/booknotes-sol-exploreit.md`; the approved v2 plan governs disagreements.
Scope: only `books/explore-it/` in the existing `plan/explore-it` worktree. No pushes.

## Current unit

Checkpoint A v0 complete: live baseline passes, all 27 planted defects reproduced, all four spec silences characterized, eight skill drafts validated. Evidence: `runs/00-baseline.log`, `00-proofs.log`, `00-silences.log`, `00-reset.log`. Standard seed restored. E01 closed: Opus 5.5 produced 27 reports in 21.8 minutes; raw score 48/44/2 stated/confirmed/refuted, one code-reading contradiction. Register reconciliation: 25 live-probed defects, D23 harness finding, partial D07; eight additional unplanned faults. See reconciliations/01.md and 01.json. Next: fresh Astra fixer for assigned recon findings, then checkpoint A review.

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
2. E01 blind exploration and reconciliation DONE; fresh fixer, both verify directions and lesson tag pending.
3. Checkpoint A review; E02–E06; checkpoint B review; E07–E09 and final skills/index; checkpoint C review.
4. Final clean tree and ordered commit, score, skills, deviation report.

## Measured corrections / deviations

- The plan’s DST illustration is corrected to actual IANA cron semantics: nonexistent 02:00 is skipped on spring-forward day. Opposite-season 02:00 is used to expose the fixed-offset defect.
- A1 at base=5 and completed attempt 20 waits 30.34 days (2^19), rather than the plan’s illustrative 60 days. The defect remains uncapped growth.
- Baseline has four meaningful integration checks, adding concurrent disjoint leases and priority behavior; no speculative expected test count is claimed.
- Specific brief isolation and branch rules take precedence over the global per-agent worktree default. Fresh agents work sequentially in the authorized branch; blind explorers receive plain temporary exports.

- E01 found essentially the full register early. Per plan, findings are fixed where found; no artificial dwell will be manufactured. Causes and other models also appeared early and are retained as produced. Later experiments will refine/test those models, remaining spec silences and new leads.
- First explorer launch failed before any model turn because Claude disallows the bypass flag as root. The actual run used an explicit Bash/Read/Write/Edit/Glob/Grep allowlist and no environment bypass.
