# Explore It! — project state

Goal: turn Hendrickson's *Explore It!* into queuelet, nine honest exploration
experiments and eight reusable procedures for probing and learning a codebase.
Approved scope and order: [PLAN.md](PLAN.md), [AGENTS.md](../../AGENTS.md),
[continuation brief](/root/booknotes-sol-exploreit2.md).

## Status

2026-10-22 UTC: E02 complete. Hooks T1–T3 closed; U09/U10 repaired; timeout and
snapshot semantics clarified. Current verification: five assigned items absent;
lesson-01 replay: five present. All 46 integration checks and unit tests pass.
Actual crash/500/slow-partner replay passes. Acceptance: register/runs/02-acceptance.json.
No E03 work was performed by this fixer.

## Done

- Checkpoint A reviewed NITS and repaired: 99d42fc, explore-it/lesson-01.
- E02 blind exploration and reconciliation: e390c31; raw 19 stated / 19 confirmed /
  zero refuted, 46 model cells; two new bugs plus one spec clarification.
- T1–T3 hooks checkpoint: 3287bd9.
- Commit containing this state: final U09/U10 fix, checks, lesson and acceptance;
  tag explore-it/lesson-02. No earlier commit or original report rewritten.
- All 102 raw exploration artifacts unchanged; E01 repro instruments unchanged.
  Detailed results, retained setup failures and measured dwell: register/BUILD-STATE.md.

## Next, in order

1. Coordinator exports explore-it/lesson-02 and dispatches fresh blind E03.
2. E03–E06 refine/probe the stronger existing models and remaining spec silences;
   checkpoint B receives independent cross-family review afterward.
3. E07–E09, structural transition guards, final skills/index/landing, checkpoint C
   and requested final Astra review/report.
4. Merge/push requires separate owner authorization. All work stays in this book.

## Open decisions and owners

- A1–A4 remain assigned to E04–E07; coordinator decides after actual probes.
- TR-E02-3 broader backlog/liveness metrics and TR-E02-4 eligibility explanation
  are deferred proposals beyond minimal T1–T3; coordinator owns consideration.
- E01 review note 3 direct lease/token/tenant invariants: E07/E09. Final worked
  examples and detailed book-text citation validation: later author/reviewer.
- qctl/sample-worker numeric precision remains an unprobed lead for later
  variation/data-journey exploration; do not count it as an evidenced new defect.
- No E02 blocker. Notification cap/dead-letter/timeout configuration is not
  decided here. Existing snapshots survive configuration edits and job deletion.

## Resume

Branch plan/explore-it; worktree /root/Software/booknotes-work/explore-it.
Read this state, register/BUILD-STATE.md and actual git status/tags. Git wins over
stale prose. Only books/explore-it writes and /tmp scratch are authorized; no new
branch/worktree, push, or raw-record rewrites. Stop this fixer after E02.

E02 explorer and fixer commands have ended; do not relaunch or reuse either.
No E02 background verification is active. Completed command status and logs are
register/runs/02-acceptance.json; native-agent lifecycle is 02-fixer-dispatch.json
(process_exit_status intentionally null). Scratch runners remain under
/tmp/explore-it-e02-fix for diagnosis only. The queuelet lab can remain running;
that is not an active agent. Historical scratch stack was automatically removed.
Coordinator owns future exports and must check process/log liveness before any
fresh dispatch. Do not touch the concurrent Release It! tree.
