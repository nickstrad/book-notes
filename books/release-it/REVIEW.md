# Release It! — final review (Fable, 2026-10-22)

**Verdict: PASS with fixes recommended.** The build delivers what PLAN.md v3 promised (one
flawed platform, eighteen planted flaws, twelve lessons as patch commits on a linear tagged
history, a lab whose scripts reproduce every flaw on the parent tag, eight general skills with
litefleet as the worked example), and a reader can learn the book's method from it. Three things
should be fixed before the history is frozen: lesson 04's verify script is broken at `main` (a
seed-placement flip introduced by lesson 06), lessons 01–03's scripts only pass at their own tags,
and lesson 12 attributes the lab's Godot mechanism to the book's case study. The rest is nits.

Reviewer: Claude Fable 5.1 (planner seat). Builder: Claude Opus 5.5. Per Nick's direction these
findings are reported, not gating. Scope and method are at the end.

## What was run

Every lesson's verify, from the working tree at `main` (`3f6cb12`), on this VM (4 cores, 8 GB,
two other Compose projects idle alongside), 2026-10-22 07:25–08:45 UTC. "At main" builds the
hardened tree and judges it with the lesson's thresholds; "at parent" is
`make verify LESSON=NN REF=<parent tag>` and must print `FLAW REPRODUCED`. Durations are wall
clock including the stack reset. Full logs: `/tmp/release-it-review/` on vm-1 (not kept).

| Lesson | At `main` | At parent tag | Time (main / ref) |
|---|---|---|---|
| 00 | first attempt died in the reset (M4); rerun: **BASELINE BROKEN on p99 only**, 56.3 ms against the 50 ms line, 100 % served, p50 4.9 ms. See L7. | at `v0-flawed`: BASELINE HOLDS, success 100 %, p50 7.3 ms, p99 48.4 ms (line: 50) | 35 s failed, 117 s / 227 s |
| 01 | FLAW REPRODUCED on 1 of 7: `gateway_health_failed_after_s NaN`; everything else held (3 up, node-a p99 1.4 s attributed, control health failed in 0.92 s, 100 % JSON logs). See M2. | FLAW REPRODUCED, all 7 checks | 225 s / 149 s |
| 02 | FLAW REPRODUCED on 2 of 8: `outage_504_pct 16.2`, `cancel_late_ms NaN`; in-flight 25, goroutines 127, cancellation counter 1, stalled client cut at 5 s. See M2. | FLAW REPRODUCED, all 8 (in-flight 1 215, goroutines 4 868, runaway 0.5 cores) | 122 s / 125 s |
| 03 | FLAW REPRODUCED on 1 of 8: `node_5xx NaN`; fds 474, zero tenant-visible errors, 100 % valid writes, 409 in 3.5 ms. See M2. | FLAW REPRODUCED, 6 of 8 (fds 1 024, 16 630 node 5xx, 9.4 % valid writes) | 141 s / 140 s |
| 04 | **crashed, no verdict** ("hot database null"). See H1. With a one-word temporary selector change (uncommitted, reverted): IMPROVED, duplicates 0, recovery 2 s. | **crashed, no verdict.** With the same change: FLAW REPRODUCED on `duplicate_rows 352`; the hot-database spin and the 19 s recovery did not show this run (see M3). | 129 s / 133 s; reruns 142 s / 139 s |
| 05 | IMPROVED: 98.0 % served during the control outage, p50 5 ms, 15 control calls, breaker closed after 4.6 s, 3 of 3 up | FLAW REPRODUCED: 0 % served, 0 gateways up | 136 s / 136 s |
| 06 | IMPROVED, all 7 (neighbours' p99 121 ms, lock held 0.26 s, backup target p99 25 ms) | FLAW REPRODUCED on 5 of 7: pool timeouts 289, health failed 8 of 60, 250 backup-target writes failed at 2 008 ms. **The noisy-neighbour phase did not reproduce** (105 ms, 9.6× baseline, under the 200 ms line). See M3. | 198 s / 205 s |
| 07 | IMPROVED: goodput after the kill 71.9 %, p99 174 ms, in-flight 24, rejoined gateway served 1 700 | FLAW REPRODUCED: goodput 29.8 %, p99 2 861 ms, in-flight 675 | 165 s / 154 s |
| 08 | IMPROVED: 0 restarts, 100 % whale scans answered, p99 697 ms, report 0.457 s (line: 0.5) | FLAW REPRODUCED: 6 restarts, 0 % whale answered, 1 603 other-tenant errors, report 1.16 s | 112 s / 113 s |
| 09 | IMPROVED: depth 153, oldest 0.97 s, 0 missing, outbox 0 | FLAW REPRODUCED: depth 4 143, oldest 28.7 s, 2 925 missing | 117 s / 154 s |
| 10 | IMPROVED: WAL 3.4 MB, handles flat, ledger −347 580 rows, 0 disk-full | FLAW REPRODUCED: WAL 25.2 MB, 3 026 disk-full errors | 227 s / 210 s |
| 11 | IMPROVED: 12 of 12 | FLAW REPRODUCED: 5 of 12 fail | 82 s / 126 s |
| 12 | IMPROVED: 0 rollout errors, longest restart window 3.6 s, 0 cut off, mixed fleet 0 % | FLAW REPRODUCED: 3 cut off, 3 rollout errors. With `REF_SERVICES=gateway`: 100 % mixed-fleet query failures, 4 356 errors | 170 s / 178 s / 181 s |

Reading of the matrix: **every lesson's flaw reproduces on its parent tag** (lesson 04 only after
the selector fix), and **lessons 05–12 are IMPROVED on the hardened tree**. Lessons 01–03 fail at
`main` on checks that encode a rule a later lesson deliberately changed or a query that confuses
"zero" with "no data", not on anything that regressed. The numbers agree with the commit bodies
within run-to-run noise; where they differ (lesson 04's hot database, lesson 06's noisy phase,
lesson 07's before-goodput 29.8 % vs 11.4 %) the builder's VM was busier, which the notes say. Unit
tests and `go vet` pass at `main`.

## Findings, by severity

**high** = a deliverable is broken or a reader would be misled about the book or the code;
**medium** = a plan-fidelity gap or a fragility a reader will hit; **low** = wording, consistency,
process.

### High

**H1. Lesson 04's verify is broken at `main` and at every tree from lesson 06 on.**
`lab/verify/04-retries.sh:17` picks "a normal tenant's `analytics` database on node-b" from the
seed manifest. `cmd/seedgen/main.go:98` assigns nodes by database-count parity
(`nodes[dbIndex%2]`); lesson 06 (`9028409`) gave the noisy tenant three more databases, which
flipped the parity for every later tenant, so all twenty `analytics` databases now land on node-a
and all twenty `staging` on node-b. The selector returns `null`, loadgen is started with
`-focus null`, and the script dies after the toxic is removed with no verdict (observed twice:
`04-main`, `04-ref`, 129 s and 133 s, exit 2). The harness always comes from the working tree, so
this also breaks `REF=release-it/lesson-03` from `main`. It works at the `lesson-04` and
`lesson-05` tags, which is why the builder never saw it. With `.name == "staging"` substituted
(temporary, uncommitted, reverted), the lesson runs: IMPROVED at `main` (duplicates 0, recovery
2 s) and FLAW REPRODUCED at `lesson-03` (352 duplicate rows). *Fix:* select by node, not by name
(`select(.node == "node-b" and (.name == "analytics" or .name == "staging"))`), and have the script
fail loudly if the selector is empty. Consider folding into lesson 06's commit, since that is the
commit that broke it, and re-pointing the tags from there.

### Medium

**M1. Lesson 12 attributes the lab's Godot mechanism to the book's case study.**
`lessons/12-handling-versions.md`, "The idea", last paragraph: "Chapter 12's case study, Waiting
for Godot, shows what happens without phases. One in-place `ALTER` on a big table is queued behind a
long report. Every query on that table then queues behind the `ALTER`, and the site waits for a
rewrite nobody can see." The lock queue behind a running report is `lab/godot.sh`'s design (its
header says so correctly). As far as I can verify, Nygard's chapter 12 is a deployment night spent
waiting on schema scripts that ran far longer than planned, the story that motivates Ch. 13's "The
Fallacy of Planned Downtime" and "Phases of Deployment"; it does not describe an ALTER blocked
behind a report. PLAN.md's register row F18 has the same conflation in miniature ("(Godot)" after
the lab symptom). *Fix:* reword to "`make godot` replays the shape of Chapter 12's night on
litefleet: a migration that takes a lock and holds everyone behind it", and describe the book's
story in one accurate sentence. Nick should confirm against his copy; I could check the chapter's
contents, not its text.

**M2. Lessons 01, 02 and 03's verify scripts pass only at their own tags.** `lessons/README.md`
says `make verify LESSON=NN` "judges the checked-out tree", and the landing page's "Run it" has
the reader check out a tag first, so the documented path works. But at `main` (the state a reader
lands in) three lessons print `FLAW REPRODUCED` for behaviour that is better than at the tag:
- 01 (`lab/verify/01-transparency.sh:71`): expects gateway `/health` to go `failed` within 5 s of
  the Postgres pause. Lesson 05 changed that rule to `degraded: serving N cached routes` on
  purpose, so HAProxy kept 3 of 3 up (the improvement) and the check reads NaN.
- 02 (`:74`, `:77`): counts only 504s as "answered by the gateway"; at `main` lesson 07's limiter
  sheds with 503 once 24 requests are waiting on the black-holed node (4 881 `overloaded:503` vs
  503 `timeout:504`), so the share is 16 %. And `cancel_late_ms` reads a node request log line with
  `error_code == "statement_deadline"` that no longer appears in that shape, although
  `statements_cancelled_total` incremented.
- 03 (`:40`): `sum(increase(http_requests_total{...status=~"5.."}[75s]))` returns no series when
  there were no node 5xx at all; `check` treats NaN as a failure. "Zero" and "no data" are
  conflated. Lesson 04's script already does this right (`or vector(0)`).
*Fix:* make the three checks monotone (01: accept `degraded` or `failed`; 02: count 503 with
`Retry-After` as a fast failure and read the cancellation from the metric; 03: `or vector(0)`), or
state in `lessons/README.md` and the landing page that each lesson's verify is meant for its own
tag and may fail later.

**M3. Two headline demonstrations are host-load sensitive.** The plan's thresholds were tuned on a
VM that was running other projects' load tests; on a quieter VM two "before" symptoms did not
appear:
- Lesson 06 phase (a) at `lesson-05`: other node-a tenants' p99 105 ms during the noisy tenant
  (9.6× the 11 ms baseline), under the absolute 200 ms line; the builder saw 534 ms. The ratio
  reproduces the flaw, the absolute number does not. The plan's threshold was the ratio.
- Lesson 04 at `lesson-03` (with the H1 selector fixed): node-a recovered in 1 s (builder: 19 s)
  and the hot database on node-b saw 599 of 599 writes with zero `database is locked` (builder:
  240.9). Only the duplicate rows (352) reproduced. The 50-try busy loop only spins when bulk
  writes overlap, which at 10/s needs a slow host.
Both lessons' *after* numbers hold, and the correctness and pool/backup checks reproduce every
time, so the lessons are not wrong, but the notes' before numbers are not what a reader on a quiet
machine will see. *Fix:* for 06 (a), check the ratio (≥ 3×) as well as the absolute; for 04, raise
the hot tenant's write size or rate until the spin shows on an idle VM (the note's "Try changing"
already points there), and say in both notes that the before numbers came from a loaded host.

**M4. The first `make verify` on a fresh Postgres volume can fail with a psql socket error.**
`lab/lib.sh` `reset_stack`: `compose up -d --no-build --wait postgres` then `psql_ "DROP DATABASE
..."` through `docker compose exec`. On a brand-new volume the image's initdb runs a temporary
server, the healthcheck can pass before the real one listens, and the next `psql` fails:
`make verify LESSON=00` from a never-started stack died at 35 s with `psql: error: connection to
server on socket "/var/run/postgresql/.s.PGSQL.5432" failed`; the next run worked. `make up` first
(the documented path) avoids it because `lab/up.sh` waits on control's health. *Fix:* in
`reset_stack`, retry the first `psql_` for up to 30 s.

**M5. The required cross-family review did not happen.** PLAN.md section 6: three checkpoints,
"each reviewed by a fresh agent from a different family (Codex)". AGENTS.md phase 4 requires a
reviewer from a different model family than the builder. Builder, landing, skills and this review
are all Claude. Both STATE.md files record this as an open decision. *Fix:* a Codex or Muse review
of the whole (history, tags, two or three verify runs, the skills), or Nick's explicit waiver in
STATE.md. This review does not satisfy the rule.

**M6. `release-it/v1-hardened` points at the landing page, before the skills.** Tag `a338b8d`;
the skills commit `f27e8aa` and the STATE update `3f6cb12` come after. The plan's checkpoint C puts
the skills before the tag, and the README at the tag says "Skills: not written yet; they are the
next phase" (`git show release-it/v1-hardened:books/release-it/README.md`, line 77). A reader who
checks out the "final state" gets no skills and a page that says they do not exist. *Fix:* once H1
is folded in and the tags move anyway, point `v1-hardened` at the last content commit.

### Low

**L1. Steady State citation uses the first edition's heading.** `lessons/10-steady-state.md`
header and PLAN.md L10: "Steady State (data purging, log files, in-memory caching, avoid
fiddling)". In the 2nd edition the subsections are "Data Purging", "Log Files" and "Limit Caching";
"in-memory caching" was the 1st edition's heading and "avoid fiddling" is the section's argument,
not a heading. The content taught is right. Confidence moderate: the publisher's contents page lists
sections, not subsections.

**L2. Lesson 12 renames the book's phases.** "The idea" lists Chapter 13's phases as Expand, Roll
out, Relax, Contract. The book's section "Phases of Deployment" has "Relax Between Phases",
"Expansion", "Rollout", "Cleanup"; expand/contract is Fowler's ParallelChange vocabulary (which
"Read more" credits). *Fix:* use the book's names and say expand/contract is the same idea.

**L3. Lesson 01's summary of the Ch. 6 case study could not be verified.** "a retailer whose site
struggled through its busiest season while the team had few vital signs to go on. They diagnosed it
from scraps and guesses, because the system said almost nothing about itself." The chapter's
sections ("Taking the Pulse", "Vital Signs", "Diagnostic Tests") suggest the team had
instrumentation and the story is about using it under pressure. The transparency argument the
lesson teaches is in Ch. 8 and Ch. 10, which it cites correctly. *Fix:* Nick checks the sentence;
if in doubt, cite Ch. 6 as "diagnosing a live system from its vital signs".

**L4. Lesson 00 reports the superseded calibration.** "one gateway at 300 req/s and one node at
250 req/s (`lab/capacity.env`). The fleet sustains 500 req/s." `lab/capacity.env` at `main` says
350 / 300 / 600 (recalibrated at checkpoint B; lesson 02's deviations). The note was never touched
after v0. *Fix:* one sentence pointing at the recalibration.

**L5. `project/README.md` and the landing page say 240 databases; the seed makes 243** (lesson
06's three extra for the noisy tenant). *Fix:* "243 (240 plus the noisy tenant's three)".

**L6. Lessons index wording "0 % timeouts" for lesson 02's before** (`lessons/README.md`, row 02).
The number is the share of node-a failures answered 504 by the gateway; a reader will read it as
"no timeouts happened". *Fix:* "0 % answered 504 (every failure was the client giving up)".

**L7. Thin margins on a shared VM.** Lesson 00's 50 ms p99 line: 48.4 ms at `v0-flawed` and
56.3 ms at `main` (the note shows 30.4 ms), so the hardened tree "breaks" the baseline while
serving 100 % with a better median (4.9 ms vs 7.3 ms); the extra tail is the admission, metrics
and streaming work the lessons added, plus host noise. Lesson 08 at `main`: report 0.457 s against
0.5 s (note: 0.29 s). Lesson 02 at `lesson-02` (builder): outage p99 2.914 s against `< 3.0`. Each
fails a run where nothing regressed. Lesson 07's window-median approach is the model; the notes
already say numbers vary. *Fix:* widen the three lines (lesson 00 to ~80 ms, or make it success
rate plus p50) or switch them to window medians.

**L8. No `docs/knowledge/` was written.** AGENTS.md and the global fresh-agent lifecycle expect
durable knowledge in `docs/knowledge/`. The build produced plenty (Toxiproxy 2.12 wedging on toxic
removal and `untoxic_busy`; the Postgres init race; `cpu_shares` on a shared VM; the seed snapshot
trick; "make the target the bottleneck or the experiment proves nothing"), all of it only inside
notes' "Deviations" and the chaos skill's closing list. *Fix:* three or four short entries cut from
the notes, committed as `docs: ...`.

**L9. Plan sizes versus delivered sizes.** Code-only lines per lesson (tests, lab, loadgen, seedgen
excluded): 01 679 / 02 207 / 03 115 / 04 175 / 05 390 / 06 263 / 07 119 / 08 306 / 09 265 / 10 273
/ 11 194 / 12 507, against plan estimates of 350 / 170 / 130 / 150 / 190 / 220 / 130 / 150 / 160 /
170 / 320 / 290. Every lesson over ~400 total was split as the rule says. "Diff in minutes" holds
for 02, 03, 04, 06, 07, 09, 10, 11; 01, 05, 08 and 12 are closer to twenty. A note for the next
plan's estimates, not a defect.

**L10. Run time.** The landing page says "1–3 minutes with warm images". Measured here: 82–227 s,
median ~140 s; lessons 01, 06 and 10 run 3.3–3.8 minutes including the reset. *Fix:* "2–4 minutes".

## Plan fidelity: the recorded deviations, ruled

| Deviation | Ruling | Why |
|---|---|---|
| v0 + lab one commit (A1/A2 merged) | acceptable | the plan's own gate (`make verify LESSON=00` before the tag) needs the lab |
| node `mem_limit` 512 MB (plan 384) | acceptable | keeps L03's "stays down at the fd wall" visible instead of an OOM; documented in lesson 00 |
| whale = 15 000 × 4 KB rows (plan 500 k small) | acceptable, clever | a 7 s scan would be cancelled by L02's deadline before F13 showed |
| `notes` table without a unique key; `statements` column; extra report classes | acceptable | L04's duplicate audit needs a table where a repeated insert succeeds |
| v0 keeps 32 idle Postgres connections | acceptable | "v0 must serve gentle load" is the plan's rule; open connections stay unbounded (F06 intact) |
| calibration 300/250/500 then 350/300/600 | acceptable; see L4 | |
| L01 two commits, ~650 lines | acceptable | node health reports physical facts later lessons read |
| Toxiproxy upstream `timeout` toxics and proxy cycling | acceptable; knowledge entry (L8) | injector limitation found and worked around |
| CPU weights on the shared VM; Postgres capped at 1.5 CPUs | acceptable | lab plumbing; recorded |
| L02 phase 1 reads only; outage p99 < 3.0 (plan ≤ 2.7) | acceptable; see L7 | F05 would leak during the run; F08's retry legitimately runs to 2.9 s until L04 |
| L03 control pool 20 (plan 10) | acceptable | measured; L06 cuts it to 8 once the cache exists |
| L04 headline moved to the hot database; node-a ratio ~1.1 (plan 4–6×) | acceptable and honest; see M3 | the note says L02's deadlines had already closed the slow-dependency storm; duplicates remain the damning number |
| L05 two commits; probe count ≤ 20 from t=15 s | acceptable | arithmetic given |
| L06 2 reads/db + per-tenant cap 4; noisy tenant 4 databases; absolute 200 ms | acceptable; see M3 and H1 | the per-tenant cap is what makes "a quarter of the node" true; the seed change broke L04's script |
| L07 gateways at 0.15 CPU, nodes 0.75; window-median p99 ≤ 300 ms | acceptable with a caveat | the contrast is real (29.8 % → 71.9 % here); the reader should know the default topology is node-bound; the note says so |
| L08 whale p99 ≤ 1 s (plan 100 ms); transcoding instead of `io.Copy` | acceptable; the second is better than the plan | L06's lanes queue the scans; transcoding keeps the strict reader alive for L12 |
| L09 spill on SIGTERM; metering header timeout 2 s | acceptable | the plan's own check needs it |
| L10 checkpoint any WAL > 64 KB; LRU 128; 48 MB volumes; `disk I/O error` counted | acceptable | measured and explained |
| L11 HTTP API, not the Go client; 5 of 12 fail (plan 4) | acceptable | REF trees lack the dependency; the extra failures are real |
| L12 refuse instead of retrying `draining`; gateways drain; window ≤ 10 s (plan ≤ 2 s); contract next release; `make godot` outside the verify; singleflight | acceptable, the largest numeric weakening | a Compose restart is a new container; contract-in-the-next-release is more correct than the plan; `ROUTE_CACHE_TTL` is the config knob the plan asked for |
| seed snapshot restore | acceptable; knowledge entry (L8) | |
| skills after the `v1-hardened` tag | gap; M6 | |
| Codex reviews per checkpoint not done | gap; M5 | |

Everything the plan listed is present: 18 flaws (lesson 00's register has `file:line` pointers,
spot-checked at the v0 tag for F01, F02, F05, F06, F08, F11, F13, F15, F17, F18 and at parent tags
for F11, F14, F15, F17, F18: all point at the code described), 12 lessons with notes, verify
scripts and tags, `lessons/README.md`, the landing page, `make calibrate / audit / rollout / godot
/ test-chaos`, and the eight skills of section 5.

## Teaching quality

Could a reader learn the book's method from this repo? Yes.

- **The history reads as the story.** `git log --oneline release-it/v0-flawed..release-it/v1-hardened`
  is 23 commits, linear, no merges, subjects exactly as PLAN.md section 4.1 wrote them, every lesson
  commit body carrying flaw, fix and both verdict lines. Split lessons (01, 05, 08, 11, 12) say
  "part 2 ... the tag is there" in part 1's body.
- **Each note has the same five parts** and the before/after numbers support each claim, both in
  the builder's runs and in this review's (table above): in-flight 1 215 → 25, node fds 1 024 →
  474, duplicate rows 352 → 0, served during a control outage 0 % → 98 %, goodput after a kill
  29.8 % → 71.9 %, node restarts 6 → 0, events lost 2 925 → 0, disk-full errors 3 026 → 0, chaos
  failures 5 → 0, mixed-fleet failures 100 % → 0 %.
- **The notes are honest where the plan was optimistic.** Lesson 04 says the plan's 4–6× storm
  was already closed by lesson 02; lesson 06 says bulkheads bound concurrency, not CPU; lesson 12
  says a stateful restart costs 3–8 s and names ts-db's move protocol as the real answer.
- **"Try changing" names real knobs.** Every identifier mentioned exists (`RequestDeadline`,
  `StatementDeadline`, `PoolWait`, `statementsTenant`, `CheckpointBytes`, `HandleCache`,
  `Shippers`, `MaxBytes`, `TTL`, `RefreshAfter`, `GATEWAY_SHED_LIMIT`, `NODE_A_WIRE_VERSION`), and
  the cited regression tests exist (`TestFailedWriteReleasesTheWriteLock`,
  `TestWriterWaitsForTheLock`, `TestStaleCachedRouteIsReresolved`).
- **No flaw is fixed early by accident.** Lesson 02 leaves `Execute`'s `Begin()` unbound so F05
  survives to lesson 03 and says why. The small cross-cutting hunks in later lessons (06's 2.5 s
  create client and `Retry-After` pass-through, 08's `defer inflight.Dec()`, 10's `Release()`
  calls, 12's `SetCacheTTL` and `InFlight()`) are each named in the commit body or note.
- **Weak spots:** H1 and M2 (a reader at `main` who runs the early lessons sees failures that are
  not failures); M3 (two before-symptoms need a loaded host); the two largest diffs (01, 12) are a
  long read; lesson 08's and 10's "table from one run, verdict from another" are honest but a
  reader may stumble.

## The skills

All eight follow the AGENTS.md contract (frontmatter; When to use, Inputs, Procedure, Output,
Source, Worked example) and are procedures, not summaries. They are usable on an unfamiliar
codebase: the procedures give grep-able signals (`&http.Client{}`, `return err` between `Begin()`
and `Rollback()`, `DisallowUnknownFields`, `wal_autocheckpoint=0`), thresholds (≈10 % retry
budget, Little's-law sizing, `lock_timeout` in seconds) and output tables before litefleet is
mentioned. The worked examples are confined to a last section and quote measured numbers that
match the lesson notes and commit bodies (checked line by line for `budget-timeouts-and-retries`,
`review-capacity-headroom`, `plan-safe-release` and `design-chaos-experiment`). The hunt's worked
example is the register with line numbers, which is the answer key the plan asked for, and the
"≥ 13 of 18" bar is stated.

Citations across the skills and notes, checked against the 2nd edition's published contents:
Ch. 8 "Transparency" (designing for transparency, logging, instance metrics, health checks); Ch. 10
"System-Wide Transparency"; Ch. 9 "Load Balancing", "Demand Control"; Ch. 13 "Phases of
Deployment", "The Fallacy of Planned Downtime"; Ch. 14 "Help Others Handle Your Versions", "Handle
Others' Versions"; Ch. 15 "The Testing Gap"; Ch. 17 "Adopting Your Own Monkey" and "Disaster
Simulations"; all twelve Ch. 4 antipattern names and all twelve Ch. 5 pattern names: **accurate**.
The quote "integration points are the number-one killer of systems" (lesson 02) is Nygard's.
Chapter numbers match the 2nd edition throughout. The inaccuracies found are M1, L1 and L2; L3 is
unverified.

Two suggestions, not defects: `hunt-stability-antipatterns` could state the expected
false-positive rate (a `context.Background()` in a startup path is not a finding), and
`review-operational-transparency` step 5 could give the latency to inject as a fraction of the
tightest timeout rather than "below every timeout".

## Anything broken or misleading

- Broken: H1 (lesson 04's verify at `main`). Misleading at `main`: M2 (lessons 01–03's verdict
  lines), L7 (lesson 00's baseline), M6 (the tag's README).
- Untrue statements found: lesson 12's Godot sentence (M1); 240 vs 243 databases (L5); lesson 00's
  calibration (L4); "0 % timeouts" (L6).
- `STATE.md` (root) and `books/release-it/STATE.md` match git: every SHA and tag listed resolves
  to the commit described. Unit tests and `go vet` pass at `main`.

## Scope and method

Read: PLAN.md v3, AGENTS.md, both STATE.md files, all thirteen lesson notes, the two indexes and
the landing page, all eight skills, every lesson commit body, the Makefile and lab runner, every
verify script's checks, the migrations and `godot.sh`. Spot-checked `file:line` pointers at the
tags named above and the small cross-cutting hunks of lessons 06, 08, 10 and 12. Ran `go vet` and
`go test ./...` at `main`. Ran the matrix above from the working tree at `main`, not a fresh clone;
images were built from the tree and REF images from `git archive` of the parent tags, which is what
a clone would do. The one temporary edit (lesson 04's selector, for evidence) was reverted and the
tree left clean. Book citations were checked against the publisher's contents page for the 2nd
edition (chapter and section titles); subsection titles and case-study content from memory, marked
with their confidence.
