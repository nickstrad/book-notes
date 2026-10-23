# E06 fresh independent spot-check — third-family (Muse Wonder Full) review

**Verdict: NITS** (2 findings; no assigned repair shown wrong; one small repair round requested)

**Top finding:** the U29/U30 `UseNumber` conversion changed decoder completeness semantics — the new qctl input path silently accepts trailing garbage after a JSON value and enqueues the first value, where lesson-05 `json.Unmarshal` rejected it (payload became `null`, which the real API refuses). All eight assigned verifications (U27–U33 + A4) otherwise hold on the evidence I could independently corroborate.

## Scope of this review

- Fresh clone `/tmp/explore-it-review06/source`, branch `plan/explore-it`; fix tag `explore-it/lesson-06` = `7a38788`, parent exploration `b3b0559`, historical ref `explore-it/lesson-05` = `0099bcd`. Diff confined to `books/explore-it/`; history linear (0 merges); 9 SUT files + `probe06.py` repros, lesson note, register/state records. Evidence: `git-scope.log`.
- Scope is E06 only. No E07, no whole checkpoint-B review, no fixes, no commits, no other worktrees touched. Working tree clean at close.
- Records read: `STATE.md`, `AGENTS.md`, `PLAN.md` E06 section (lines 632–663) + register rows D15–D17/A4, `exploration/PROTOCOL.md`, `register/reconciliations/06.md` / `06.json` / `06-fixer-brief.md`, experiment 06 artifacts (`revealed.md` score, defect reports E06-F01…F07, canonical U27–U33), `lessons/06-relate-it.md`, full `lesson-05..lesson-06` project diff, acceptance bundle `register/runs/06-*`.
- Book claims cite the approved PLAN chapter map (Ch. 7.1–7.3) and say so; no full-text validation was available or fabricated. The lesson's own wording is honest on this point (`lessons/06-relate-it.md:9-12`).

## Findings (severity order)

### 1. [Medium-low] New `Decode` paths accept trailing data that `Unmarshal` rejected — invalid qctl input now fails open

- Where: `project/cmd/qctl/main.go:53-61` (`jsonOpt`, used for `--payload`, `--result`, `--json`), `project/cmd/qctl/main.go:279-288` (response printer), `project/internal/client/client.go:57-59` (`Client.JSON`).
- What changed: lesson-05 used `json.Unmarshal`, which requires the entire input to be one JSON value. The lesson-06 code uses `json.Decoder.Decode` with `UseNumber()` but no second-decode/EOF check, and `jsonOpt` still discards the error. A single `Decode` consumes exactly one value and ignores the rest.
- Observed with actual binaries (built from `git archive explore-it/lesson-05` and from the lesson-06 tree, run against a stub API; `qctl-stub.log`, corroborated by `edge-probe.log`):
  - lesson-05 qctl, `--payload '{"big":9007199254740993}xyz'` → wire body `{"payload":null,...}` (real API rejects null, per `objectPayload`), fail-closed.
  - lesson-06 qctl, same input → wire body `{"payload":{"big":9007199254740993},...}`, exit 0. Trailing second value `{"big":1}` is also silently dropped.
  - `Client.JSON`: response `{"id":1}\n{"id":2}` → old error `invalid character '{' after top-level value`; new returns nil.
- Why it matters: this is a new behavior regression introduced by the fix on a write path, in exactly the correctness domain the lesson repairs (CLI must not silently alter operator data). Truly malformed input (`{"big":}`) is unchanged (becomes `null` under both versions — pre-existing, explicitly noted "not counted" in E06-F04:38-39). The established correct pattern already exists in this codebase at `project/internal/api/api.go:59-64` (second `Decode` must return EOF).
- Requested repair (one repair round): add whole-input completeness to all three new decode sites (e.g. a small helper mirroring `api.go:62`, or `json.Valid` pre-check); make `jsonOpt` surface a parse error rather than enqueue on truncated input. Add one regression to `probe06.py` U30 (or a unit check): trailing-garbage and second-value `--payload` must exit non-zero / send nothing, valid values unchanged. Re-run current `make verify LESSON=06`; REF behavior for U30 must stay PRESENT (the historical binary rejects trailing input, so a strict expectation of "rejected" would change the historical oracle — the new check should be attached so that the paired PRESENT/ABSENT judgement for U30 remains about numeric precision, e.g. as an additional current-only control inside the probe's ABSENT branch).

### 2. [Low] The sample worker's documented `http` kind still rounds remote JSON numbers; sibling neither fixed nor recorded as deferred

- Where: `project/internal/worker/worker.go:99-114` (`http` kind: `json.Unmarshal(b, &v)` into `any`, unchanged since lesson-05).
- Evidence: the explorer's sibling search named this exact site — E06-F03:48 "`worker.go:98-110` `http` kind: response decoded with `json.Unmarshal` into `any` → same loss for remote results". My edge probe (`edge-probe.log`) confirms the roundtrip `{"big":9007199254740993,...}` → `{"big":9007199254740992,"fraction":0.12345678901234568}` through that line. The `http` kind is a documented worker kind (project README:39-41); no repro or check covers it.
- Calibration: the coordinator's U29 assignment text is narrow ("through echo result and webhook", `reconciliations/06.md:15`), so I do not score this as an assigned repair done wrong. But canonical U29 says simply "fixed", and a reader following F03's sibling list will believe the worker's decoding habit is repaired. Silence overstates the repair.
- Requested repair (either): (a) decode the http response with `UseNumber()` **plus** a completeness check so the existing fallback-to-string behavior on invalid/trailing bodies is preserved, and extend U29 with an http-kind control; or (b) explicitly record the http sibling as deferred in `STATE.md` open questions, U29's canonical record, and the lesson note's forward-questions list.

## What was checked and held

| Brief item | Evidence |
|---|---|
| Only U27–U33 + A4 repaired; nothing else | SUT diff = 9 files, each maps to an assigned ID (`git-scope.log`, `diff-project.txt`). No SUT/instrument edits beyond the assignment; hooksink untouched. |
| D15–D17 not claimed as new | Lesson 06:39-47, `reconciliations/06.json` `catalogue_expected_found_early_in`, commit body: all say found/fixed in E01, no rediscovery manufactured. |
| Policy A4 + attempt-lifetime explicit | `SPEC.md` lines (dead exempt; attempts belong to job), commit body one-liners, U33 canonical record, lesson decisions section 125–129. A4 never counted as an eighth blind finding anywhere. |
| 7 found vs 8 verify IDs honest | Score 7 reports; `lesson-06.ids` = 8 with A4 named as policy disclosure in lesson/commit/acceptance. |
| Score qualifications retained | 20/18/2/2/30 kept in lesson, STATE, acceptance, commit; H-E06-C counter-prediction and H-E06-17 sampling non-hit qualifications reproduced faithfully (checked against `revealed.md:100-103`). Fixer's later batch-race hit recorded as fixer evidence, not backfilled. |
| All 3,447 artifact hashes | Independently recomputed: 3447 entries, 0 missing, 0 mismatches (`hash-check-full.txt`). Exploration diff `b3b0559..lesson-06` = only the 8 allowed canonical status/index files. Blindness audit consistent (exit 0, zero flagged paths; boundary "prompt and plain export; no OS isolation" stated). |
| Frozen instruments, same checksums both directions | `sha256sum -c 06-instruments.sha256` in my clone: 95/95 OK (`instruments-recheck.txt`); manifest sha256 `379fef8…` matches acceptance. `pair1` vs final differ only in `baseline_test.go`; `attempt1` vs `pair1` only in `probe06.py` — matching the retained-failure narrative exactly. |
| Numbers through actual clients | Builder probe06 U29/U30 with target-image binaries (acceptance logs) + my actual-binary stub test: lesson-06 qctl input and display exact (`9007199254740993`, long fraction), lesson-05 rounded (…992, 0.12345678901234568) (`qctl-stub.log`). |
| Worker numeric consumers, optional/defaults | Code inspected (`worker.go:82-93`): missing/wrong-typed `seconds`/`n` → empty `json.Number` → `Float64()` = 0, so prior defaults survive (probed, `edge-probe.log`). Present values proven by U29 (sleep 0.3 s timing, fail n=1 attempts==2); omitted-`n` default additionally covered by the pre-existing U09 check through the full 74-check suite (fail3 with no payload). |
| Schedule payload consistency | `objectPayload` shared verbatim (`enqueue.go:149-158` → `schedules.go:31-36`); omitted → `{}`, null/non-object → 400; PATCH has no payload field; manual run preserves numbers via `json.RawMessage` store (`store.go:43-45`); cron copies stored payload bytes raw (`fire.go:76`). Probe06 U28 checks all five invalid shapes + manual/cron/PATCH with controls. |
| Tenant-safe delete (U27) | `RowsAffected` check within namespace-scoped WHERE (`schedules.go:144-160`); probe proves foreign schedule untouched after a foreign DELETE attempt, mixed semantics rejected. |
| NUL rejection + lease preservation (U31) | Recursive string/key check before any mutation (`jobs.go:66-75`, `containsNUL` 185–203); probe asserts lease fields byte-equal, no attempt inserted, clean retry succeeds, literal `\u0000` text stays valid. Validation-before-lease-check ordering matches the pre-existing nack-delay pattern. |
| Missing-owner race mapping (U32) | Only SQLSTATE 23503 + the two real constraint names (`jobs_queue_owner`/`schedules_queue_owner`, migrations/002) map to 404 (`enqueue.go:161-173`), wired into enqueue, batch, schedule create, schedule run. Probe requires known statuses only, cardinality of accepted rows, zero orphans, positive+missing controls, and actual FK-error observation for PRESENT. Builder historical run recorded real hits: enqueue 2×400, schedule 4×500, batch 1×400 with constraint text (`06-historical-acceptance.log`) — not inferred from non-hits. |
| Attempt cascade (U33) | `migrations/005.sql`: guarded (restart-safe), `ON DELETE CASCADE … NOT VALID` + index; inspector-verified that purge/retention/explicit delete are all `DELETE FROM jobs`, so one mechanism covers all three; probe proves 1→0 rows on each path with live-attempt control retained. Legacy orphans stay by explicit decision (no bulk cleanup). |
| Pending webhook lifetime | U33 probe: snapshot rows survive job deletion, remain pending, then actually deliver after sink recovery; job_events survive via API + storage. No change to outbox/events lifetimes in code. |
| Dead retention + expiry controls (A4) | `expire.go:91` state list reduced to succeeded/cancelled; A4 probe uses real API clock below (3599 s) and above (3602/7202 s) cutoff for all three states with API status **and** `SELECT` counts agreeing, plus retained-dead retry. |
| Raw webhook bytes, unexpected = error | U29 decodes hooksink `raw_body` with Decimal, checks signature and equality with stored result (`probe06.py:150-156`); hooksink itself unchanged. Probe stubs: all setup/tool/parse/status surprises raise → `PROBE ERROR` exit 2; `verify.sh` treats non-PRESENT/ABSENT verdicts and incomplete historical reproduction as failure. No SQL writes in probe06 (only `SELECT` via `p.sql`), no pause hooks; chaos limited to public API concurrency, API clock, hooksink modes, container lifecycle. |
| Historical binaries are real | `verify.sh` archives REF project, builds its image, copies only current `lab/` + `test/checks/` instruments in; U29/U30 exec the target container's qctl/worker. My stub test used identically archived/built lesson-05 vs lesson-06 `qctl` binaries. |
| Repros as regression checks | `TestL06U27…TestL06A4` run the same public oracles (`lesson06_test.go`); baseline test change is type-only adaptation (`float64`→`json.Number.Int64()`), behavioral assertions (priority + disjoint leases) retained, and the whole sequence was re-run after it. |
| Previous-tag `file:line` pointers | Verified each cited line at `explore-it/lesson-05`; all land on the operative statement or its enclosing line (e.g. `enqueue.go` fail at 87 cited as 88 — adjacent, not misleading). |

## Commands actually run in this review and results

| Command | Result | Log |
|---|---|---|
| `python3 /tmp/explore-it-review06/check_hashes.py` (3447 hashes + exploration diff) | exit 0; 3447/3447 match | `hash-check-full.txt` |
| `cd books/explore-it/project && sha256sum -c ../register/runs/06-instruments.sha256` | exit 0; 95/95 OK | `instruments-recheck.txt` |
| `go run /tmp/explore-it-review06/edge/main.go` (old vs new decoder semantics, verbatim logic) | exit 0; trailing regression + http rounding observed | `edge-probe.log` |
| `go vet ./...` and `go test ./...` (project) | exit 0 / exit 0 (clock, jobs, schedule ok) | `unit-vet.log` |
| `go build` of qctl from archived lesson-05 + lesson-06 trees (`-buildvcs=false`) | both OK | terminal, `qctl-stub.log` header |
| `python3 /tmp/explore-it-review06/qctl_stub_test.py` (actual binaries vs stub API) | exit 0; lesson-06 exact on valid input/display, fail-open on trailing input; lesson-05 rounded, fail-closed | `qctl-stub.log` |
| `make up` with prescribed isolated env (`queuelet-review-e06`, ports 18370/18371/15513) | **blocked**: exit 2, `permission denied while trying to connect to the docker API at unix:///var/run/docker.sock` | `lab-up-attempt.log` |
| `gofmt -l` on all changed Go files | exit 0, clean | terminal |
| `git` scope/history checks | confined, linear, clean tree | `git-scope.log` |

Builder's recorded acceptance (inspected, not re-run): `make up` 41.282 s, `make verify LESSON=06` → `FIXED AND CHECKED: 8 absent; regression checks pass` (151.177 s), `REF=explore-it/lesson-05` → `DEFECTS REPRODUCED: 8 present` (57.671 s), `make checks` 74/74 (398.913 s), `make test` pass — all exit 0 in `06-acceptance.json`, logs coherent with commit body and dwell table (paired only, no sweep claimed).

## Scope limitations and deviations (honest)

1. **Docker lab not run by me.** The session sandbox denies the Docker socket and escalation prompts are disabled; local PostgreSQL via `runuser` is likewise sandbox-denied. Per the brief's smaller-spot-check option, I substituted: full hash/instrument recomputation, unit tests, actual-binary client tests, verbatim decoder edge probes, and audit of builder acceptance logs. `make verify LESSON=06` current and REF were therefore **not independently re-executed** in this review; their verdicts rest on builder evidence (which is internally consistent and checksum-backed). No stack of mine was started, so no cleanup was needed; no default or foreign stack was touched.
2. **U32 race evidence in my run is builder-recorded**, not self-observed (same sandbox limit). The oracle requires actual historical FK-error hits, and the recorded REF run has them; a sampling non-hit in a future REF rerun would surface as a failed REF verdict, not a silent pass.
3. **NOT VALID cascade for pre-upgrade attempt rows** (attempts written at lesson-05 whose jobs are deleted after the 005 migration applies to an existing DB) was not independently executed. The probe proves rows created after migration; the upgrade-path behaviour follows PostgreSQL FK-action semantics by code inspection only — flagged as an observation below, not a finding.
4. Full-book text unavailable: chapter claims checked only against the approved map, as the lesson itself states.
5. Full previous suite (`make checks`, 74) not re-run by instruction; builder's 398.913 s pass inspected, and I confirmed no SUT changes outside E06 that would make that evidence stale (the freeze diff shows only `baseline_test.go` changed between the failed and final suite runs, with the full sequence repeated after).

## Observations (not findings)

- **Legacy invalid schedule templates:** templates stored before lesson-06 with null/non-object payloads are not cleaned or re-validated; cron `Fire` (`fire.go:76`) copies stored payload bytes directly into jobs, bypassing `objectPayload`, so a pre-upgrade poison template would keep behaving inconsistently (cron inserts, manual run rejects) after upgrade. Post-fix creation is consistent. Suggest the coordinator record an explicit legacy-template decision (parallel to the legacy orphan-attempts decision) at the repair round or checkpoint B.
- **Race oracle inherent asymmetry:** a current-side sampling non-hit is structurally indistinguishable from fixed code in probe06's return value; the paired REF requirement (actual hits, `verify.sh` fails on incomplete reproduction) is what keeps the claim honest. Bounded current sampling (800 + 800 + 128 writes, recorded) is a reasonable budget.
- `sleep` with absurd magnitudes (e.g. `seconds: 1e999`, now storable/deliverable exactly) converts via `Float64()` overflow to ±Inf; behaviour is no worse than the prior float64 decode failure mode and is outside the assigned findings.

## Requested repair round summary

1. Restore whole-input JSON completeness in `cmd/qctl/main.go` (jsonOpt + printer) and `internal/client/client.go`; surface qctl parse errors; add a trailing-input control that preserves the paired U30 oracle semantics. Files: `project/cmd/qctl/main.go:53-61,279-288`, `project/internal/client/client.go:57-59`.
2. Fix or explicitly defer the worker `http` number decoding at `project/internal/worker/worker.go:110-113`, and align the U29 record/lesson wording with whichever is chosen.
3. Re-run the standard sequence (`make up`, verify current + REF, `make checks`, `make test`) in the repair run and record it; that also closes my execution gap from limitation 1.

Stop point respected: this is the fresh E06 spot-check only. Checkpoint B review (E02–E06) and E07 remain future, separate fresh units.
