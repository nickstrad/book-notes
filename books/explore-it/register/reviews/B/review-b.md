# Checkpoint B review — Explore It! / queuelet

Reviewer: fresh Codex Sol, executing `/root/booknotes-sol-ckptB.md`. Scope: E03–E06
at their tags and HEAD, plus E02/historical reproduction and register continuity
for PLAN §8's E02–E06 checkpoint. Experiment 07 was not started.

Runtime source: fresh shared-object clone in `/tmp/explore-it-checkpoint-b/repo`,
initial HEAD `8aa2895`; tag targets `8ed9177`, `171974a`, `0099bcd`, `755d21c`.
Each tag ran its own committed instruments and regression checks. HEAD used the
initial HEAD's unchanged project tree. Runtime source and instruments still match the final repository; review changes under
project/ affect README only. Every REF used the final HEAD instruments
against archived previous-tag binaries, including target-container qctl/worker.
Compose project `queuelet-review-b`, API/hook/PG ports 18260/18261/15493; direct
checks use `QUEUELET_DATABASE_URL=postgres://queuelet:queuelet@localhost:15493/queuelet?sslmode=disable`. Historical
runner uses its existing `queuelet-ref` project and 18210/18211/15443.

## Verdict

**PASS for the authorized checkpoint B review scope, with the explicit limits below.** Two documentation nits corrected separately. In `d6c34d0`,
the canonical index identifies the old U09 copy as historical and points to the
fixed U09 record. In `b1aee49`, project README documents `QUEUELET_DATABASE_URL`
for direct checks on an isolated database port. No material runtime defect found in scope. No immutable report, runtime source,
repro, policy or tag was changed.

## Measured runtime results

Full commands, UTC starts, exits, elapsed times and log paths: [acceptance.json](acceptance.json).

| Run | Exit | Seconds | Exact verdict / result |
|---|---:|---:|---|
| [tag03-verify](logs/tag03-verify.log) | 0 | 25.546 | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| [tag04-verify](logs/tag04-verify.log) | 0 | 47.804 | `FIXED AND CHECKED: 6 absent; regression checks pass` |
| [tag05-verify](logs/tag05-verify.log) | 0 | 130.263 | `FIXED AND CHECKED: 4 absent; regression checks pass` |
| [tag06-verify](logs/tag06-verify.log) | 0 | 109.807 | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| [head02-verify-retry](logs/head02-verify-retry.log) | 0 | 45.205 | `FIXED AND CHECKED: 5 absent; regression checks pass` |
| [head03-verify-retry](logs/head03-verify-retry.log) | 0 | 22.552 | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| [head04-verify-retry](logs/head04-verify-retry.log) | 0 | 38.121 | `FIXED AND CHECKED: 6 absent; regression checks pass` |
| [head05-verify-retry](logs/head05-verify-retry.log) | 0 | 113.034 | `FIXED AND CHECKED: 4 absent; regression checks pass` |
| [head06-verify-retry](logs/head06-verify-retry.log) | 0 | 118.499 | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| [ref02-verify](logs/ref02-verify.log) | 0 | 84.152 | `DEFECTS REPRODUCED: 5 present` |
| [ref03-verify](logs/ref03-verify.log) | 0 | 66.777 | `DEFECTS REPRODUCED: 8 present` |
| [ref04-verify](logs/ref04-verify.log) | 0 | 65.188 | `DEFECTS REPRODUCED: 6 present` |
| [ref05-verify](logs/ref05-verify.log) | 0 | 67.2 | `DEFECTS REPRODUCED: 4 present` |
| [ref06-verify](logs/ref06-verify.log) | 0 | 96.603 | `DEFECTS REPRODUCED: 8 present` |
| [head-checks](logs/head-checks.log) | 0 | 370.885 | `74 integration checks pass (test inventory: check-list.log)` |
| [head-test](logs/head-test.log) | 0 | 3.413 | `unit/application tests pass` |
| [head-vet](logs/head-vet.log) | 0 | 0.696 | `go vet passes` |
| [head-A3-open](logs/head-A3-open.log) | 0 | 0.308 | `PRESENT: A3 — cancelled running work; completion ack=409, no cancellation signal in worker protocol` |
| [final-down-retry](logs/final-down-retry.log) | 0 | 2.921 | `owned lab containers/network/volume removed` |

The tag and HEAD counts are 8/6/4/8 for E03–E06. The historical counts are
5/8/6/4/8 for E02–E06. They agree with the lesson fix commit bodies. The full
suite runs all named tests without skip sites; no extra full dwell sweep is implied.

E06 historical U32 observed enqueue 2, schedule 13 and batch 1 foreign-key errors;
HEAD observed zero such errors and zero orphans with positive/missing-owner controls.
The retained A3 control reports cancellation followed by ack 409 and no worker
cancellation signal; that confirms the unresolved policy rather than closing it.

Raw logs are preserved byte-for-byte, including Docker progress trailing spaces.
Authored Markdown/JSON/scripts are checked separately for whitespace. Exact evidence
and audit entry points are retained here; SHA256SUMS hashes logs and machine records.

## Retained setup failure

The first HEAD E02 run exited 2 after its five public repros reported ABSENT.
`TestL02ProductionClockDisabled` and `TestL02LegacyMalformedOutboxIsolation` tried
localhost:15433 and failed to connect: this reviewer set the Compose custom port
but omitted `QUEUELET_DATABASE_URL`. No product failure or weakening occurred.
The runner cleaned up, then `resume.py` set the explicit URL and reran HEAD E02
through the remaining acceptance. `acceptance.json` and `logs/head02-verify.log`
retain the failure; the successful retry has a separate log. This is why the
README clarification above was warranted. The initial launcher also used unavailable
`python` before any clone/test; it was immediately relaunched with `python3`.

## Acceptance criteria

PLAN §8 describes checkpoint B as a procedure rather than a numbered acceptance
list. The procedure and §5.2/§5.3/AGENTS review bar are assessed separately here.

| Criterion | Outcome and evidence |
|---|---|
| Charter → blind explorer → reconciliation for E02–E06 | PASS with disclosed deviations. Five committed briefs/charters, completed single runs, reconciliations `02.md`–`06.md`. No second run was required; no remaining assigned planted defect was concealed. |
| Fresh export without plan, register or git; blindness check | QUALIFIED PASS. Export/audit records show the intended boundary. E04's bare `find /` listed external skill directory names without reading their contents; exported lesson 03 disclosed A1. Both are explicitly recorded in `04-blindness-audit.json` and the lesson; A1 is not counted blind. E05 A2 was chartered and not reported as a bug; E06 A4 was policy disclosure. Prompt isolation is not OS isolation. |
| Original observations, scores and model deltas preserved | PASS. `integrity.json`: all 3,657 manifest entries match their experiment commits; all 3,626 stable report/experiment entries match HEAD. Living models/backlog legitimately evolve. E05's 22 stated = 18 confirmed + 2 refuted + 2 unprobed; E06's two contradictions remain qualified, including the sampled batch non-hit. |
| Separate experiment and fix commits; E02 hook split permitted | PASS. `history.json` records linear commits and exact fix bodies. E02 has separate hooks and runtime repair; remaining lesson tags point to their fixes. E06's existing authorized amendment is preserved; this review rewrites no history. |
| Fixes follow actual findings and policy decisions | PASS within tested scope. U09/U10, U11–U18, U19–U23, U24–U26, U27–U33 plus T1–T3/A1/A2/A4 reconcile to named reports and spec lines. All D01–D27 were repaired in E01, so later checks are regressions, not manufactured discoveries. No central transition-policy refactor is smuggled into B. |
| Lesson-specific checks at tags 03–06 | PASS: all four own-tag verification commands exited 0; see measured table above. |
| HEAD keeps lesson checks and prior fixed register green | PASS: HEAD verify 02–06 and all 74 integration checks, units and vet exited 0 after the documented environment correction. |
| Reviewer replays previous-tag repros | PASS: all five REF commands exited 0, reproducing all 5/8/6/4/8 assigned items at the respective previous tags. |
| Repros remain black-box and frozen | PASS: static audit and both runtime directions. Public probes use API/CLI/read-only SQL and lab controls; migrations and explicit private-schema direct checks are not mislabeled as black-box exploration. All recorded public instrument hashes checked match; E03 instruments also have an empty tag→HEAD diff. Original failed attempts and frozen manifests are retained. |
| Notes explain probes, patches, decisions, before/after, try-changing | PASS. Lessons 02–06 cite named chapters, link reports and exact commands, distinguish found-early/unplanned/policy, and qualify dwell as paired only. Source pointers spot-checked against prior tag and fix code. |
| Time budget | QUALIFIED. Explorer runs are within 60-minute budgets. PLAN's “well under a minute” verify estimate is superseded by measured expanded checks, real waits and races; exact independent elapsed times are in acceptance.json. |
| Skills and book claims | PARTIAL / DEFERRED as already recorded. Eight draft procedures exist and are executed by the experiments; final worked examples and source-text citation validation belong to later work. Chapter mapping was checked for internal consistency only, not certified against full book text. |
| Reviewer family | EXPLICIT DEVIATION. This brief assigns Sol instead of PLAN's Muse checkpoint seat. E03–E05 fixes were Claude; E02/E06 were Codex. This is fresh independent execution, but it is not a new cross-company code review of the Codex-built portions. Existing E06 Muse NITS and fresh Claude repair PASS remain separate evidence. Do not relabel this run Muse or claim it fulfills a missing cross-company review by identity alone. |

The speculative defect placement/dwell examples in PLAN are design predictions.
The actual fixes follow its “found early is fixed early” rule and the measured
reconciliations, without moving already-repaired defects to later experiments.

## Register and open-item reconciliation

[Canonical INDEX](../../../exploration/defects/INDEX.md) is the current D/U status
source. Original F/E##-F reports and the older U09 observation keep their original
wording. All 27 D records and 33 current U records are fixed; “fixed” has bounded
meaning (not an assertion that every sibling or upgrade path is repaired).

| Item | Current disposition / owner |
|---|---|
| T1/T2/T3 | Closed E02, mapped to T01/T02/T03 checks. TR-E02-1/2/5/6 covered; E01 Q1 actual attempt start timing covered for new leases without invented legacy timing. |
| A1 / Q-E04-1 | Decided E04: default cap 3600 seconds or larger base; explicit cap preserved. |
| A2 | Decided E05: leased-at-purge jobs may complete; unsuccessful exits become dead; retry clears mark. |
| A4 | Decided E06: dead jobs exempt from automatic retention; succeeded/cancelled age out. |
| A3 / E05-F02's in-flight cancellation attempt | OPEN, Nick/coordinator; fresh E07 must model the running worker and missing attempt, not just recheck stale token rejection. |
| Q2 / Q-E04-2/3 / Q-E05-1/2 | Decided by U20/U23/U22/U24/U25 respectively. Raw backlog questions stay historical; reconciliation/spec/current records supply the answers. |
| Q-E06-2 | Attempts half decided U33; delivered-outbox cleanup OPEN. Events and pending notification snapshots remain intentionally independent. Legacy orphan attempts were not bulk-cleaned. |
| E06 review finding 2 | OPEN: pre-existing sample-worker http-response decoder uses float64. Echo/client/qctl precision repair does not establish precision of actual http-worker path. Future fresh run; Nick owns priority. |
| Legacy invalid schedule templates | OPEN upgrade-policy decision for Nick/coordinator: new creation is validated, old stored null/non-object templates are neither migrated nor revalidated. Cron/manual inconsistency is not claimed fixed. No policy invented in this review. |
| TR-E02-3/4 | OPEN, coordinator: richer backlog/liveness metrics and eligibility explanation. Minimal observation hooks did not close these proposals. |
| TR-E04-1/2; T-E05-1; T-E06-1/2 | OPEN testability requests: sweep lag, stats snapshot, ticking offset clock, job ids on worker decode failures, deterministic delete-race instrument. No hook added during review. |
| E04 forward hypotheses | Q-E04-4 stats completed definition; H-E04-08 backward-clock FIFO, H-E04-09 stats snapshot, H-E04-10 ack atomicity under DB fault, H-E04-11 NULL schedule kind, and R42 linear half remain unprobed/open as appropriate. Nick owns policy; fresh explorers own observations. |
| E05 forward hypotheses | H-E05-20 error-text sibling covered structurally by U26, not newly live-probed; H-E05-21 SIGTERM/deregistration scenario remains unprobed but is bounded by U24. The recorded walker never demonstrated W1 on the old tag: direct U24 reproduction supplies that evidence. Schedules/queue deletion/retention/webhooks remain outside its walk coverage. |
| E06 forward hypotheses | Q-E06-1 worker recovery after deregistration; H-E06-20 deleted-queue lease race, H-E06-21 NUL http-worker result consequence, H-E06-22 NUL registration remain OPEN. H-E06-17 remains the original blind sampling non-hit; later paired batch-race repair evidence does not change that score. |
| Earlier remaining backlog | Q3 driver-text error policy, Q4 admin mutation authority, Q5 returned webhook secrets, H45 outage catch-up remain in hypotheses.md. H43's worker-release behavior is covered by U24; H48 cross-tenant verb concerns are bounded by D21's regression coverage, not a new exploration claim. No question was deleted just because STATE's summary is shorter. |
| Checkpoint A follow-ups | Direct independent Go assertions for D18/D19/D21 remain E07/E09 work; final skill worked examples and full source-text citations remain pending. |

## Limits and next step

No full dwell sweep, new blind experiment, complete v0 rerun or final-book review is
claimed here. Checkpoint A baseline evidence remains inherited. Independent fresh
execution here supplements, rather than rewrites, the E06 repair review history.
Next unit is a fresh experiment 07 run from `explore-it/lesson-06`; the coordinator
owns its authorization and the open policy decisions above.
