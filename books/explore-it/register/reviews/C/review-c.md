# Checkpoint C review — Explore It! / queuelet

Reviewer: fresh Claude Opus 5.5 run, executing `/root/booknotes-opus-ckptC.md` (the checkpoint C
wrap). Scope: lessons 07–09 at their tags and at HEAD, both verification directions, the full
check suite, and reconciliation of the whole register (D01–D27, U01–U45, A1–A5, T1–T3) across all
nine experiments, the way the [checkpoint B review](../B/review-b.md) did for E03–E06.

**Reviewer family.** The E07 fix was finished by Opus, and E08 and E09 were explored and fixed by
Opus. This review is also Opus, so it is a fresh, independent execution, **not a cross-family
review**. The non-Claude end review (Codex Astra) is still to run and is the cross-family check
for lessons 07–09.

Runtime source: a fresh shared-object clone at `/tmp/explore-it-checkpoint-c/repo`, HEAD
`605c339`, plus worktrees at `explore-it/lesson-07` (`8787e43`), `lesson-08` (`9b1b1a1`) and
`lesson-09` (`130f199`). Each tag ran its own committed instruments and checks. HEAD's project
tree equals `lesson-09`'s (the only later commit is a knowledge note). REF runs used HEAD's
instruments against images built from the previous tag. Compose project `queuelet-review-c`,
ports 18270/18271/15503 with an explicit `QUEUELET_DATABASE_URL`; REF used the runner's own
`queuelet-ref` project (18210/18211/15443). Runner: [run.py](run.py); ledger with start times,
commits, exits and elapsed seconds: [acceptance.json](acceptance.json).

## Verdict

**PASS with one recorded instrument flake and NITS.** Every verification direction for 07–09
reproduced the committed verdicts. All 27 planted defects and all 45 unplanned faults are fixed,
each with a repro script and a named regression check. The only failure is one `make checks`
run, where U32's race sampler missed its batch-path window. Five isolated reruns passed, so it
is recorded as an open item rather than a defect. Three documentation nits were fixed in the
repair commit. No runtime source, repro script, check, raw record, score or tag was changed.

## Measured runtime results

| Run | Exit | Seconds | Verdict / result |
|---|---:|---:|---|
| [tag07-verify](logs/tag07-verify.log) | 0 | 255.2 | `FIXED AND CHECKED: 3 absent; regression checks pass` |
| [tag08-verify](logs/tag08-verify.log) | 0 | 140.4 | `FIXED AND CHECKED: 4 absent; regression checks pass` |
| [tag09-verify](logs/tag09-verify.log) | 0 | 281.7 | `FIXED AND CHECKED: 7 absent; regression checks pass` |
| [head00-verify](logs/head00-verify.log) | 0 | 24.8 | `BASELINE HOLDS: enqueue→lease→ack→webhook; unit and happy-path checks pass` |
| [head01-verify](logs/head01-verify.log) | 0 | 364.0 | `FIXED AND CHECKED: 35 absent; regression checks pass` |
| [head02-verify](logs/head02-verify.log) | 0 | 57.4 | `FIXED AND CHECKED: 5 absent; regression checks pass` |
| [head03-verify](logs/head03-verify.log) | 0 | 31.2 | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| [head04-verify](logs/head04-verify.log) | 0 | 66.0 | `FIXED AND CHECKED: 6 absent; regression checks pass` |
| [head05-verify](logs/head05-verify.log) | 0 | 106.6 | `FIXED AND CHECKED: 4 absent; regression checks pass` |
| [head06-verify](logs/head06-verify.log) | 0 | 117.1 | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| [head07-verify](logs/head07-verify.log) | 0 | 256.9 | `FIXED AND CHECKED: 3 absent; regression checks pass` |
| [head08-verify](logs/head08-verify.log) | 0 | 143.3 | `FIXED AND CHECKED: 4 absent; regression checks pass` |
| [head09-verify](logs/head09-verify.log) | 0 | 318.6 | `FIXED AND CHECKED: 7 absent; regression checks pass` |
| [ref07-verify](logs/ref07-verify.log) (REF lesson-06) | 0 | 125.9 | `DEFECTS REPRODUCED: 3 present` |
| [ref08-verify](logs/ref08-verify.log) (REF lesson-07) | 0 | 172.9 | `DEFECTS REPRODUCED: 4 present` |
| [ref09-verify](logs/ref09-verify.log) (REF lesson-08) | 0 | 146.0 | `DEFECTS REPRODUCED: 7 present` |
| [head-checks](logs/head-checks.log) | **2** | 791.8 | 93 of 94 check functions pass; `TestL06U32` PROBE ERROR (below) |
| [u32-rerun1–5](logs/) | 0 ×5 | 29.6–48.5 | `ABSENT: U32`, batch path 404s observed 32/48/48/112/80 of 128 |
| [head-test](logs/head-test.log) | 0 | 0.3 | unit and application tests pass (Go test cache) |
| [head-vet](logs/head-vet.log) | 0 | 0.8 | `go vet` clean |

The own-tag and HEAD counts for 07–09 are 3/4/7, and the REF counts are 3/4/7. They match the
fix commit bodies of `explore-it/lesson-07…09` and the lessons' result sections. The early-fixed
controls passed ABSENT in both directions: D18–D20 for 07, D21–D24 for 08, and D12, D14,
D18–D20 and D27 for 09. HEAD verify 01–06 reproduce checkpoint B's counts (35/5/8/6/4/8), so
lessons 07–09 regressed nothing earlier.

### The U32 flake

`make checks` failed once in `TestL06U32`:

```text
PROBE ERROR: U32 — RuntimeError(('no concurrent missing-owner observations', {'enqueue': Counter({404: 416, 201: 384}), 'schedule': Counter({201: 464, 404: 336}), 'batch': Counter({201: 128})}))
```

U32's probe (`lab/repro/probe06.py:317`) races enqueue, schedule creation and batch enqueue
against queue deletion. It requires every path to observe at least one request that lands after
the owner is gone (a 404). Without that, the race window was never sampled, and the probe calls
the run an instrument error instead of ABSENT. In this run all 128 batch requests landed before
the delete. No foreign-key error, orphan or wrong status was observed; the control simply did
its job. The same check passed inside `head06-verify` twelve minutes earlier and in five isolated
reruns. Earlier committed logs (`register/runs/06-*`, checkpoint B) show batch 404 counts of
32–64. This is a sampling instrument whose batch window can, rarely, be missed under a different
timing. Widening that window means changing a frozen instrument, which is outside a reviewer's
remit, so it is recorded as an open item. The final suite at `v1-explored` is recorded
separately in STATE.md.

## Static review (AGENTS review bar, items 3, 4 and 6; PLAN 5.2)

| Criterion | Outcome and evidence |
|---|---|
| Raw experiment records unchanged | PASS. `git diff <experiment commit> HEAD -- exploration/experiments/<NN>` is empty for 07 (`49d8d67`), 08 (`a2e7b01`) and 09 (`2fee853`). The E05 record's `clock.go:182-191` citation is wrong (the file has 50 lines at that tag), but raw records are immutable and the lesson-level citations are right. |
| Experiment commits contain no SUT change | PASS. `a2e7b01` and `2fee853` touch no `project/` file. Between `lesson-06` and `49d8d67` the only project change is checkpoint B's README clarification `b1aee49`. |
| Fix diffs confined to their notes | PASS. 07: jobs/api/worker internals, one migration, SPEC, lab, checks. 08: `cmd/api`, compose restart policy, api/webhook internals, SPEC, lab, checks. 09: jobs/schedule/api internals, SPEC, lab, checks. Each matches the lesson's "Focused implementation" and the reconciliation's assigned scope. No file outside `books/explore-it/` changed since `v0-inherited`. |
| Source pointers into the previous tag | PASS on spot checks: lesson 09's `internal/schedule/fire.go:73-74` (`return e` inside the loop) and `internal/jobs/expire.go:23-29` (the notify loop after `Commit`) at `lesson-08`; lesson 08's `cmd/api/main.go:19-22` (`log.Fatal`) and `internal/webhook/ship.go:17`, `:21` at `lesson-07`. Explorer-record line numbers sometimes differ from the lessons' (E08's `api.go:132` against `:128`, verified at `lesson-07`); the lesson notes are the corrected ones. |
| Commit bodies match verdicts | PASS. All three bodies carry `FIXED AND CHECKED` / `DEFECTS REPRODUCED` lines with the counts measured here. |
| Blindness | QUALIFIED PASS, as before. Audits `07/08/09-blindness-audit.json`: zero source modifications and zero forbidden export paths for each. E07 flagged 2 inputs (an attempted read of a runtime task-output path outside the export); E08 flagged 0; E09 flagged 2, both judged benign. The boundary is a plain export plus prompt, not OS isolation. E09's charter quoted the plan's expected habit counts, and lesson 09 says the explorer's matching counts are therefore not independent evidence. |
| History | PASS. Linear: no merge commits since `v0-inherited`. Tags `lesson-07…09` point at their fix commits. |
| Book citations | NOT CERTIFIED. Section numbers are consistent with PLAN 2.2 and the publisher's table of contents. Full book-text validation has not been done, and every lesson and skill says so. |
| Time budget | QUALIFIED. All explorer runs stayed within 60 minutes (about 18–22). Verify runs take 2–6 minutes and the full suite about 13, not "well under a minute"; each lesson records this deviation. |

## Register reconciliation (all nine experiments)

The [canonical index](../../../exploration/defects/INDEX.md) is the status source. Every id below
has a black-box repro under `project/lab/repro/` and a named check under `project/test/checks/`
(`TestL01D01…D27`, `TestL01U01…`, through `TestL09U45`, plus `T01–T03` and `A1–A5`). Every repro
was ABSENT at HEAD in this review.

| Items | Found by | Fixed / decided | Measured here |
|---|---|---|---|
| D01–D27 (planted) | E01 recon: 25 live, D23 by harness plus stored state, D07 partial (slash/space disclosed in lesson 01) | lesson 01 | head01 ABSENT ×27; controls in 07/08/09 ABSENT both directions |
| U01–U08 | E01 | lesson 01 | head01 |
| T1–T3 (T01–T03) | E02 testability requests | lesson 02 hooks | head02 |
| U09–U10 | E02 | lesson 02 | head02 |
| U11–U18 | E03 | lesson 03 | head03 |
| U19–U23; A1 | E04; A1 measured, not blind (an exported lesson named it) | lesson 04; A1 decided (default cap 3600 s) | head04 |
| U24–U26; A2 | E05; A2 chartered | lesson 05; A2 decided | head05 |
| U27–U33; A4 | E06; A4 policy disclosure | lesson 06; A4 decided (dead exempt from retention) | head06 (U32 flake above) |
| U34–U35; A3 | E07; A3 chartered | lesson 07; A3 decided (cancel immediate, cooperative stop) | tag07, head07, ref07 |
| U36–U38; A5 | E08; A5 a new silence found blind | lesson 08; A5 decided (classic cron DST) | tag08, head08, ref08 |
| U39–U45 | E09 (6 of 9 predicted sites plus 1 sibling) | lesson 09 + transition table | tag09, head09, ref09 |

Totals: 27 planted defects, all fixed (26 found blind, D07 partly), 45 unplanned faults, all
fixed, and five spec silences decided. Two ids have a historical observation copy beside the
canonical record (U09 and U10). The index now flags both copies; U10's was missing and is fixed
here. "Fixed" keeps checkpoint B's bounded meaning: the paths exercised by the repro and check,
not every sibling or upgrade path.

**Checkpoint A follow-up, partly met.** A asked for direct, implementation-independent Go
assertions for D18, D19 and D21. `TestL09ModelWalk` now covers D19 directly: it drives nack after
cancel against the transition table. D18's two-worker stale token and D21's cross-tenant access
are still checked only through the shared repro oracle. Carried as open.

## Documentation nits fixed in the repair commit

1. `exploration/defects/INDEX.md`: flags the historical U10 observation copy, whose wording still
   says "status open", the way checkpoint B flagged U09's.
2. `exploration/models/README.md`: no longer says "no system models have been recorded yet",
   though all eight have existed since E01.
3. `docs/knowledge/06-data-boundaries.md`: "Independent validation is pending" now points to the
   E06 revalidation PASS.

Reported but not changed, because raw or explorer-owned records are immutable or the claim is
qualified elsewhere:
- `models/causes.md` says "predicted 7 defect sites"; revealed, reconciliation and lesson 09 say
  nine (PD1–PD9).
- E07 and E09 elapsed times differ by about two minutes between the explorer's estimate and the
  process timestamps; the lessons note this.
- `variables.md` has 24 E03 rows where lesson 03 says 25.
- The E05 walker's invariant numbering skips I11.
- `hypotheses.md` still shows some items `open` (H43, Q-E04-1…3) that later reconciliations
  decided; the backlog keeps its original wording, as checkpoint B recorded.

## Open items carried forward (owners)

- **Spec questions for Nick:** Q-E09-1 (is `0-23` a wildcard hour under A5?) and Q-E09-2 (should
  success or retry clear `last_error`?).
- **Legacy stored schedules (Nick, upgrade policy):** a stored schedule that cannot advance keeps
  `/health` degraded until it is disabled. Legacy invalid templates are neither migrated nor
  revalidated (E06, E09).
- **U32 sampler window (instrument):** the batch-path control can miss its window, one observed
  miss in this review. A wider window or a per-request spinner is a future instrument change,
  with re-frozen hashes and a REF replay.
- **Direct assertions for D18 and D21** (checkpoint A follow-up, above).
- **Testability:** TR-E09-2, a deterministic post-commit fault hook (U43 and U44 sample races
  instead). TR-E02-3/4, TR-E04-1/2, T-E05-1, T-E06-1/2, T-E07-1/2 and T-E08-1…3 also remain open.
- **Forward hypotheses:** H-E09-18/19, H-E04-08 (backward clock, now with a lease-before-`run_at`
  symptom), Q-E08-1 worker identity, H-E08-22/24, the E07, E06, E05 and E04 forward questions,
  and the E06 http-worker float64 decoding path, all as listed in STATE.md.
- **Cross-family review of 07–09:** the Astra end review.

## Limits

No full dwell sweep, new blind experiment, v0 rerun or book-text validation is claimed. Checkpoint
A and B evidence is inherited. The skills were finalized in the same run; their worked examples
were checked against the records by link and number, not by re-executing them blind.
