# Fresh E07 fixer — one chunk only

Work only in `/root/Software/booknotes-work/explore-it`, branch `plan/explore-it`.
Never touch `/root/Software/book-notes` or another worktree. All deliverables under
`books/explore-it/`; scratch `/tmp/explore-it-coordinator/e07/fixer/`. This specific
worktree restriction overrides the global fresh-worktree rule. You are a NEW headless
Codex Astra fixer, from the seeder's family, following a completed blind Claude run.
Permission profile: unrestricted local authorized book files and queuelet lab, local
commits/tags; approval never. No pushes, API keys, external messages, additional agents,
future experiments or plan restructuring. Read STATE first, AGENTS, PLAN, PROTOCOL,
register/reconciliations/07.md, original E07 reports, notes, revealed, state model.

Complete E07 end to end: smallest patches only for U34, U35 and policy A3; meaningful
regressions per finding; black-box public repros current/REF; lesson note; canonical
statuses/indexes; state; logical fix commit; tag `explore-it/lesson-07`; clean tree.
Do not change raw explorer records, scores, instruments or old tags. D18–D20 were fixed
in E01; preserve that fact and verify them as controls rather than claiming discovery.
Do not work on E06 http-number decoding, broad errors, performance bounds, E08 or E09.

A3 coordinator policy is explicit in reconciliation: immediate terminal cancellation
and lease revocation remain, worker cooperatively observes `cancel_requested` through
heartbeat and cancels ongoing sleep/HTTP; late ack/nack cannot succeed or revive it.
Document precise heartbeat wire semantics and limits in SPEC. Preserve old wrong-token,
nonleased and D19 checks; a cancellation observation must not extend a revoked lease.
Use the plan's cancel-request/heartbeat/ack design with a minimal representation that
fits existing immediate-cancel semantics. No OS process kill masquerading as cooperation.
Cancellation cannot roll back already-performed remote effects. Add actual historical
and current sample-worker probes (built target binaries, never current worker source
against historical API). Capture continuing work/stop evidence; process survival alone
or terminal job state alone is insufficient. Quiet unrelated job control must complete.
A3 was not probed by this blind run; do not backfill raw findings.

U34 must cover both zombie-leases and 500 worker-disappearance symptom. U35 must cover
purge racing lease, nack, worker release and expiry; positive sequential controls and
attempt/state outcomes. Fix locking order carefully: avoid introduced deadlocks and
unnecessary global serialization. Do not invent a test hook when existing black-box
races reproduce frequently. Bounded robust sampling is acceptable but preserve failure
logs and require genuine historical PRESENT; setup/runtime errors are PROBE ERROR,
never ABSENT. Checks may use internals, repros use public API/CLI/read-only SQL/test
clock/container controls only, never fixture SQL writes or SUT hidden paths.

State-table verification: run a test per rejected state/event cell, preserving the
observed 49-cell matrix and identifying extra actor/expiry/cancel-request cells explicitly.
Plan's illustrative 44 cells is not a target to fabricate. Record actual counts/results.
Keep check assertions independent enough to catch bad transition/state/lease/attempt effects.

Acceptance sequentially: isolated `make up`; `make verify LESSON=07` (U34,U35,A3);
`make verify LESSON=07 REF=explore-it/lesson-06`; `make checks`; `make test`; `go vet ./...`.
Set COMPOSE_PROJECT_NAME=queuelet-fix07, image fix07, API=18290, hook=18291, PG=15523,
QUEUELET_URL=http://localhost:18290, QUEUELET_HOOK_URL=http://localhost:18291,
QUEUELET_DATABASE_URL=postgres://queuelet:queuelet@localhost:15523/queuelet?sslmode=disable.
REF runner owns queuelet-ref ports; no concurrent verification. Stop only owned stacks.
Save exact commands, exits, elapsed and raw output in register/runs/07-*.log and
07-acceptance.json. Freeze repro/check hashes before current/REF pair and verify after.
Capture failure diagnostics and repeat only for new failures/changes. If verification
cannot pass within plan, record blocker honestly and finish clean per user brief.
Dwell optional: paired current/lesson-06 evidence sufficient, explicitly no full sweep.
Old planted D18/D19/D20 controls should pass at both versions, not enter historical
new-finding list. Never weaken them. Spec change requires exact one-line A3 decision
in commit body and both verdict lines. Include actual state-table count and deviations.

Write lessons/07-state-it.md with Chapter 8.1–8.4 citations as approved-plan references
(full text validation still pending), actual score, findings/early fixes, before/after
state table, previous-tag file:line pointers, focused patch walkthrough, commands and
verdicts, try-changing, read-more, paired dwell and honest counterfactual departure.
Update relevant lesson/readme indexes under this book if needed.
STATE must record E07 done, exploration SHA plus fix identity, 2 fixes vs deferrals,
A3 resolved, 1 contradicted reading prediction, forward questions, next fresh E08,
review pending and exact resume/liveness. No self-review claim.
Assess durable project knowledge and useful helpers, but leave them for a separately
committed knowledge step after the fix. If useful, write your assessment in final report.

Commit subject: `queuelet: fix experiment 07 findings; resolve A3; add checks (lesson 07)`.
Body includes U34/U35/A3, flaw/fix, exact current/REF verdicts, checks/matrix count,
A3 decision, paired dwell/no sweep, deviations and appropriate Co-Authored-By trailer.
Tag lesson-07 at this fix. End clean, report commits/checks/scope/knowledge assessment.


## Fresh repair handoff after provider termination

The first fresh Astra fixer ended exit 1 with a provider possible-cybersecurity-risk
error; see register/runs/07-fixer-initial-run.jsonl. Do not reuse that session, increase
permissions, change authentication or seek another provider. You are a fresh local
repair run on the same authorized queue lifecycle work and same model/permissions.
Partial source/tests/repros are uncommitted in the worktree. Read and review those
changes, then finish the original authorized unit rather than restarting exploration.

Current acceptance passed: 07-current-acceptance.log (280.111 s). Historical
07-ref-acceptance.log stopped at probe07.py's duplicate lease assertion. Worker
deletion can release a job, then another in-flight historical lease can take it with
a new token/attempt. Track separate lease epochs and current row state/attempt history
to distinguish that from duplicate active ownership. Keep U34's actual oracle strong,
validate the observed tokens and attempts; do not silently discard evidence. Correct
only this fixture problem if verified; freeze new hashes and replay current/REF.
Other partial implementation may need review. Original initial current test interrupted
when make up was still rebuilding the own stack; original failed logs remain. No
hidden fixture writes or instrument weakening. Historical diagnostics are in ignored
project/results/ref-failure.CCotQc; copy relevant text under register/runs if needed.

Prior exact runner /tmp/explore-it-coordinator/e07/fixer/run.py can be read/copied; use
new scratch /tmp/explore-it-coordinator/e07/fixer-repair/. Finish acceptance full checks,
units/vet, lesson, canonical records, indexes, state and commit/tag exactly as original
brief. Verify actual current behavior and source/record integrity, clean up owned stack.
Record provider interruption, fixture correction, measured runtime deviation, 49 rather
than illustrative44 matrix cells and no full dwell honestly. No future work.
