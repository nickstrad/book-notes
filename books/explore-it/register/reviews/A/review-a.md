# Independent checkpoint A review — Explore It! (E01)

- Target: `b839fdf` / `explore-it/lesson-01`, reviewed from fresh clone `/tmp/explore-it-review-a` (no source edits, no commits; `git status` clean at end).
- Reviewer mode: Muse Wonder Full, Standard. Source repo `/root/Software/booknotes-work/explore-it` was not written to; `/root/Software/book-notes` untouched.
- Date: 2026-10-22 UTC. Budget: 30 min target; verification runs dominated elapsed time (see timings).

## Verdict: NITS

No material correctness defect found. Both verification directions, the baseline, unit suites at both tags, and the A1–A4 preservation controls reproduced independently on an isolated stack. Findings below are documentation/harness nits and stated residual risks; none blocks checkpoint B. Per the authorizing instruction, failed/transient checks are reported plainly and later work is not gated here.

## What was actually run (isolated stack)

Env for every run: `COMPOSE_PROJECT_NAME=queuelet-review-a QUEUELET_IMAGE_TAG=review-a QUEUELET_API_PORT=18260 QUEUELET_HOOK_PORT=18261 QUEUELET_PG_PORT=15493 QUEUELET_URL=http://localhost:18260 QUEUELET_HOOK_URL=http://localhost:18261` (plus sandbox-only `GOCACHE`/`GOPATH`/`DOCKER_CONFIG` under `/tmp`). Container names confirmed as `queuelet-review-a-*`; shared `queuelet` (18200) and explorer `queuelet-e01` (18220) stacks left running and untouched. Review containers stopped at end (`make down`, exit 0; `queuelet-ref` also absent).

| Command | Outcome |
|---|---|
| `go test ./...` (lesson-01) | exit 0; `internal/jobs`, `internal/schedule` ok |
| `go test ./...` + `go vet ./...` on `git archive explore-it/v0-inherited` in `/tmp/review-a-v0` | both exit 0 (v0 compiles/tests pass — flaws latent) |
| `make up` | exit 0 after sandbox `DOCKER_CONFIG` fix (first attempt failed on read-only `/root/.docker`, reviewer-environment only) |
| `make verify LESSON=00` | exit 0 — `BASELINE HOLDS: enqueue→lease→ack→webhook; unit and happy-path checks pass` |
| `make verify LESSON=01` | exit 0, 365 s — 35 `ABSENT`, 0 `PRESENT`; checks `ok queuelet/test/checks 179.287s`; `FIXED AND CHECKED: 35 absent; regression checks pass`. Spot values: D13 produced=0 (88.9 s), U03 promotion 1.56 s vs control 1.70 s (bound 5 s) |
| `make verify LESSON=01 REF=explore-it/v0-inherited` attempt 1 | **exit 2 after 144 s** — REF api on 18210 never answered within `lab/wait.sh`'s 60 s; no PRESENT/ABSENT lines; trap removed the scratch stack so container logs were lost |
| Manual v0 `queuelet-ref` bring-up (same ports/image) | api healthy in seconds (`{"status":"ok"}`), so attempt 1 was a startup/infrastructure transient, not a v0 defect |
| `make verify LESSON=01 REF=explore-it/v0-inherited` attempt 2 | exit 0, 255 s — 35 `PRESENT`, 0 `ABSENT`; `DEFECTS REPRODUCED: 35 present` |
| `register/proof/A1.sh`–`A4.sh` (isolated current stack, after `lab/reset.sh`) | each `PRESENT`, each exit 0 — later questions genuinely undecided |
| Negative oracle probe: `QUEUELET_URL=http://localhost:18999 bash lab/repro/D02.sh` | `PROBE ERROR: D02 — URLError(...)`, exit 2; against live stack `ABSENT: D02 — delete absent queue HTTP 404`, exit 0 (error ≠ absence, independently confirmed) |
| Hash audits | all 57 manifest artifacts hash-match the lesson-01 tree; all 37 `01-instruments.sha256` entries verify OK |

Not run: standalone `make checks` (its L00 subset ran inside `LESSON=00` and its L01 subset inside `verify LESSON=01` above — same named functions, no additional coverage); `make dwell` (correctly not claimed by the builder either); v0 integration baseline on a v0 stack (v0 functioning instead evidenced by: v0 unit/vet pass, committed `00-baseline.log`, and the REF run whose 35 fixtures all require the v0 happy paths to work); full-text book citation check (see limits).

Reviewer logs: `/tmp/review-a-logs/{unit,unit-v0,up,baseline00,current01,historical01,historical01-retry,a-controls,down}.log`.

## Findings (nits, with repair)

1. **SUT diff-size numbers are wrong.** Commit body for `b839fdf` and [BUILD-STATE.md](/tmp/explore-it-review-a/books/explore-it/register/BUILD-STATE.md:73) claim "+375/-97 across 17 files". Measured: 17 source/spec files (ex-migration) = +364/-101; including `migrations/002.sql` = 18 files, +383/-101. Repair: correct the numbers in BUILD-STATE (commit body is history; note the correction in the next state update). No behavioural impact.
2. **REF failure destroys its own evidence.** [verify.sh](/tmp/explore-it-review-a/books/explore-it/project/lab/verify.sh:8) traps `down -v; rm -rf` on EXIT, so attempt 1's api-crash/startup cause is unrecoverable. Repair: on failure, dump `docker compose logs` to stderr/a kept file before teardown. Also [verify.sh](/tmp/explore-it-review-a/books/explore-it/project/lab/verify.sh:55) exits 1 silently when a defect is PRESENT on the current tree (no verdict line); print one.
3. **Regression "checks" are the same oracle, not independent assertions.** [lesson01_test.go](/tmp/explore-it-review-a/books/explore-it/project/test/checks/lesson01_test.go:14) shells out to the identical `lab/repro/*.sh`/`probe.py` and only asserts the `ABSENT:` prefix. This is explicitly justified in-code (shared oracle prevents drift) and the REF run proves each oracle can say PRESENT, so the suite still enforces the named behaviours; but it doubles the real-minute cost (D13/D27 run twice per verify) and a future probe bug weakening an oracle would pass both layers silently. Repair (later checkpoint is fine): keep the shared-oracle checks, but add a small number of direct Go assertions for the highest-severity invariants (D18/D19/D21) so one layer is implementation-independent.
4. **Skill worked example is stale at lesson-01.** [map-codebase-from-behavior/SKILL.md](/tmp/explore-it-review-a/books/explore-it/skills/map-codebase-from-behavior/SKILL.md) still says "Pending real experiments. E01 will supply the recon…". Final skills are checkpoint C per plan/STATE, so this is expected sequencing, but the draft now misdescribes the present. Repair at the checkpoint-C skill pass (already tracked); no action needed for A.
5. **Duplicate canonical slugs invite misreading.** e.g. `D02-queue-delete-orphans.md` vs `D15-queue-delete-orphans.md`, `D05`/`D07`, `D03`/`D20`, `D16`/`D27` share slugs. IDs and INDEX disambiguate correctly; cosmetic only. No repair required; do not rename (names are stable per AGENTS.md).

## Review-bar assessment (AGENTS.md, adapted by PLAN §1/§5.3)

1. **v0 baseline functioning:** PASS. v0 units/vet pass independently; happy path evidenced at both tags; flaws are behavioural, not breakage.
2. **Verify both directions:** PASS with the attempt-1 transient recorded above. Verdict lines match commit body and `01-acceptance.json`. Current 365 s / historical 255 s are consistent with the disclosed ~6.5 min measured budget for clock-less minute oracles (D13/D27); the plan's sub-minute estimate was for the original two-finding E01 and is honestly superseded.
3. **Diff confinement / no early fixes:** PASS as adapted. Recon legitimately found nearly the whole register in E01 (25 live, D23 by harness+service state, D07 partial/disclosed); plan §4 ("found early is fixed early") and reconciliation `01.md`/`01.json` assign all 35 to lesson 01 — no fictional later discoveries are manufactured. Conversely A1–A4 and T1–T3 were *not* decided early: A1–A4 re-verified PRESENT above, and the SPEC diff only adds the three decided clarifications (D12 pending, U03 non-blocking, U08 age reference) plus direct consequences of assigned fixes. Fix diff is confined to `books/explore-it/`.
4. **Citations/claims:** PASS within stated limits. Lesson v0 pointers spot-checked accurate (`cmd/qctl/main.go:15`, `internal/jobs/ack.go:8`/`:10`, `nack.go:7`, `expire.go:5`, `webhook/ship.go:16`, `api/queues.go:85`/`:108`), as were fixed-tree pointers (`ack.go:10`, `nack.go:14`, `ship.go:24`, `stats.go:18`, `migrations/002.sql:11`). Chapter/section claims (§11.1, §11.4, §12.4, §2.x) are explicitly limited to the publisher contents check with source-text review pending ([lessons/01-map-it.md](/tmp/explore-it-review-a/books/explore-it/lessons/01-map-it.md:12)) — assessed here only for internal consistency with PLAN's chapter map, which holds. Raw-vs-reconciled counts are kept distinct throughout (48/44/2/2, 27 reports, 110 cells vs 25+1+partial registered and 35 assigned faults; arithmetic checks: 44+2+2=48, model subcounts sum to 110). Dwell is claimed only v0→lesson-01 from paired logs, with `01-dwell.tsv` matching both logs — no fabricated dwell or scores.
5. **Skills executable, not summaries:** sampled draft meets the format and stays procedural; final worked examples are checkpoint C (nit 4).
6. **History/tags:** PASS. Linear, no merges; `explore-it/v0-inherited`→`847f1b3`, experiment `6e4dd4a`, fix+`explore-it/lesson-01`→`b839fdf`; subjects follow convention; experiment and fix are separate single-purpose commits as the adaptation requires.

Blindness/immutability: `prepare-export.py` archives only `project, exploration, skills, lessons` (no `.git`/`PLAN.md`/`register/`); transcript grep finds `PLAN.md` only inside the protocol's own export description; the root bypass launch failure and allowlist relaunch are preserved in `01-explorer-launch-failure-*`; all 57 original artifacts are byte-identical to `6e4dd4a` (git diff empty for F-reports/models/hypotheses; only coordinator D/U status lines changed). Historical repros genuinely exercise archived binaries — notably D01 runs the container's `qctl` ([probe.py](/tmp/explore-it-review-a/books/explore-it/project/lab/repro/probe.py:150)), SQL is SELECT-only with read-only sessions, D20 uses a public `max_attempts=1` lease rather than a fabricated fixture, and seeder `register/proof/` SQL-write fixtures are correctly labelled answer-key arrangements, not black-box replays.

## Material risks / limits

- U03's 5-second warm-lab bound and D11's physical tie-ordering oracle are timing/storage-layout sensitive; both passed here on a loaded VM (many unrelated stacks running), but they remain the flake-prone pair for future checkpoints. The D13/D27 real-minute waits make every full verify ~4–6.5 min until E02's clock hook lands.
- Source-text fidelity of chapter claims beyond publisher excerpts was not (and per instructions could not be) verified; checkpoint C/final review still owes that check. The predicted qctl/sample-worker numeric-precision siblings remain unprobed backlog, correctly excluded from E01 counts — D08's verified scope is API/store/lease/schedule only.
- One REF startup transient was observed and not root-caused (evidence destroyed by the trap, nit 2); the immediate manual bring-up and retry both succeeded.

## Bottom line

Checkpoint A is substantively sound: measured early discovery, honest accounting, oracles that demonstrably distinguish PRESENT/ABSENT/ERROR, and both verification directions green on independent rerun. Verdict **NITS** — repair the diff-size numbers and (when next touching the harness) preserve REF failure logs; nothing here needs to block E02.
