# Explore It! — independent end review

**Verdict: PASS WITH FIXES.** Findings: **0 high, 4 medium, 2 low**.

Completed: 2026-10-23 UTC. Reviewer: Codex Astra, under `/root/booknotes-astra-ei-review.md`. Target:
`e750542457e8b42fce28108714d72352ff19ae71` on `plan/explore-it`, including the
`explore-it/v1-explored` landing and subsequent walkthrough. Runtime validation used
an independent local clone, not the working checkout. This supplies the non-Claude
end review of the Claude-built E07–E09 work. It does not retroactively change the
identity or scope of earlier reviewers.

The project delivers nine useful, evidenced experiments and eight executable skills.
The early discovery of nearly the entire planted register is real and is permitted
by PLAN §4's found-early/fixed-early rule. Later experiments found new faults instead
of staging the planned discoveries. The original records survived intact. The fixes
below concern acceptance evidence, overstatements in the teaching narrative, a
misleading run recipe, and current-status accuracy. No runtime repair is proposed by
this review. Final book-source acceptance remains incomplete until R1 is addressed.

## Findings and executable follow-up list

Paths below are relative to `books/explore-it/`, except the repo-root `AGENTS.md`;
line numbers refer to reviewed HEAD. Preserve the original experiment records, logs and tags when correcting prose.

### R1 — Medium: book-source verification has never reached the required review bar

Evidence: `skills/README.md:50`, `lessons/01-map-it.md:11`,
`lessons/09-explain-it.md:11`, and the Source sections of all eight skills explicitly
defer full-text validation. AGENTS.md's review bar requires checking the actual
chapter claims, including at least one citation per lesson and skill. Matching a
section title is not that check. The deferral is honestly disclosed, but it has now
survived A, B, C and the final deliverable.

I independently checked the publisher's contents and available excerpts (scope
below). They support the broad charter, observable-state and API-exploration themes;
they do not establish all claims attributed to chapters 3–7, 9, 11–13 or Appendix A2.
The state excerpt also explicitly emphasizes an external behavioral perspective,
whereas `skills/probe-sequences-and-states/SKILL.md:33` begins from database states
and writers. Source reading is an authorized adaptation, not a violation of Nick's
plan, but its relationship to the book needs to be explained rather than inferred
from a matching chapter title.

**Follow-up:** use an authorized full copy of the named edition to make a compact
citation ledger, one substantive claim per lesson and skill at minimum, with section
and page/e-book location, checked claim, and whether it is book guidance or this
project's adaptation. Review the stronger 13.3/13.5 and 10.3 attributions in
particular. Correct unsupported attributions in authored notes/skills; retain raw
explorer wording with an erratum where needed. Until then, retain the source-fidelity
qualification rather than declaring the entire AGENTS review bar passed.

### R2 — Medium: reader-facing blindness claims omit material exceptions

Evidence: `README.md:35` and `WALKTHROUGH.md:17` describe all nine runs as seeing only
the export and never the plan/register. The boundary excluded those files, but not
all their answers: `lessons/04-judge-it.md:36` acknowledges the A1 leak;
`exploration/experiments/09-explain-it/charter.md:14` supplies the expected cause
groups and counts directly from the plan. E04 also listed directories outside the
export, and E04/E07/E09 accessed or attempted their own runtime output outside it.
The summaries are too categorical for a teaching artifact whose central discipline
is distinguishing independent discovery from supplied knowledge.

The 26/27 planted-defect result is **not** invalidated by these later exceptions:
E01 precedes them, D23 is explicitly a harness finding, and D07's unprobed
slash/space manifestation is excluded. A1–A4 are also correctly excluded from
blind-find counts in `skills/README.md`. E09's actual new-site probes remain useful;
its agreement with supplied group counts is not independent evidence.

**Follow-up:** qualify the landing and walkthrough at their first blindness claim:
plain export and prompt boundary, with documented leaks and runtime-output exceptions.
Link the exception table below. Keep the 26/27 result with its existing qualifications;
do not call A1–A4 independent discoveries or the E09 group totals independent results.
For future exports, stop copying unrevealed assignments into lesson notes and remove
expected answers from generated charters. Do not sanitize the existing raw charters.

### R3 — Medium: the claimed learning trend is not measured, and E08's chronology is wrong

Evidence: `README.md:52` says the share of findings predicted before probing rose;
`lessons/09-explain-it.md:227` repeats that conclusion. Six of nine predicted E09
sites producing reports measures prediction yield, not the share of all findings
that were predicted. There is no comparable per-experiment series establishing a
rise. E01 already logged predictions before probes and created all eight models
(`exploration/experiments/01-map-it/revealed.md:32`).

Relatedly, `lessons/08-surround-it.md:50` says E01–E07 stayed inside the system and
E08 first put neighbors on the page. `exploration/models/ecosystem.md:3` already
records E01 PostgreSQL restart/stop, webhook failures, workers, other tenants and
clock probes; its E02 revision adds more. This locally restores the staged progression
that the landing otherwise correctly rejects.

**Follow-up:** retain the measured E09 statement—six of nine predicted sites produced
reports, plus a sibling finding—and remove the comparative rise unless a consistent,
evidence-linked numerator and denominator can be reconstructed for every experiment.
Describe E08 as extending/revisiting the existing ecosystem model with new neighbor
states. Preserve original scores and distinguish hypotheses, reports, canonical
faults and prediction sites throughout. No fresh exploration is required to fix these
claims.

### R4 — Medium: the landing recipe switches source tags without rebuilding the running service

Evidence: `README.md:63` starts HEAD, line 65 checks out lesson-07, and line 67 runs
its current verification without another `make up`. `project/Makefile:12` only runs
the verifier; `project/lab/verify.sh:55` onward resets fixtures and probes the already
running stack. The REF branch builds its own temporary historical service and tears
it down; it does not rebuild the main stack. Thus the apparent lesson-07 check can
exercise HEAD binaries with lesson-07 instruments. The following `make checks` also
runs that tag's 81 functions, not the final 94. A green result would misidentify the
version tested. The walkthrough's separate tag itinerary correctly rebuilds.

**Follow-up:** use the walkthrough's disposable-worktree sequence and run
`make up && make seed` after each tag switch. Return to the final tag and rebuild
before claiming to run the final suite. Rehearse the corrected recipe while recording
source SHA and image identity at each step. This is a documentation change; no verifier
or oracle needs weakening.

### R5 — Low: the interruption and boundary audit summaries are not internally consistent

Evidence: `register/reconciliations/07.md:75` says the provider filter stopped two
Astra runs. `lessons/07-state-it.md:236` gives a different, more precise sequence:
one provider stop, then a second run stopped by Twin under Nick's handoff rule.
`register/runs/07-fixer-initial-run.jsonl:139` and `:140` are an error and turn-failed
notification for **one** run, whose exit file says 1. The repair brief acknowledges
the first stop; the unfinished repair REF log is correctly excluded from acceptance.
I found no retained second provider terminal-error record establishing the
reconciliation's stronger wording. Two terminal notifications from one failed run
do not establish two provider-stopped runs.

Also, `register/runs/04-blindness-audit.json` corrects the bare-root `find`, but omits
E04's own task-output reads. The original transcript records them at lines 174,
179, 185, 189 and 191. They are analogous to the explicitly disclosed E07/E09 runtime
exceptions, not evidence of an answer-key read.

**Follow-up:** reconcile the stop descriptions to retained lifecycle evidence: first
provider rejection; second run interrupted for handoff, unless a second terminal
provider event can be linked. Record unknowns as unknowns. Add the E04 runtime-output
exception to its audit interpretation without editing the transcript. Preserve the
incomplete run and avoid counting its verdict fragments as acceptance.

### R6 — Low: the living backlog still presents resolved questions as open

Evidence: `exploration/hypotheses.md:60` leaves H43 open although E05/U24 resolves
worker-release accounting; lines 153–155 leave Q-E04-1/2/3 open although A1/U23/U22
settle them. Line 3 of that very file says to update outcomes without rewriting
initial statements. `register/BUILD-STATE.md:8` still labels checkpoint B the current
unit and says E07 has not started. These are mutable handoff surfaces, not immutable
probe logs. Review C already noticed the backlog problem but justified retaining
original wording; preserving a statement does not require preserving a stale status.

**Follow-up:** append dated resolution/evidence links or update status/outcome cells
while preserving the original question. Clearly mark BUILD-STATE as historical and
point its current-unit header to the book STATE. Leave genuinely unresolved items
open with owners. Verify that known resolved questions no longer appear actionable
from the living backlog alone.

## Plan and register assessment

| Experiment | What the record actually demonstrates | Assessment |
|---|---|---|
| E01 | Whole-service recon; 25 planted faults live, D23 via executed harness and stored state, D07 partial; eight additional faults | Strong recon; 26/27 clears the 19/27 bar. All 27 repaired early under the plan's explicit rule |
| E02 | Compare API, rows, logs and partner; two faults, one clarification; lifecycle events, history and lab clock | Observation charter met; three reports are not three canonical bugs |
| E03 | Raw input/request-shape sweep; eight new faults | Variation charter met; negative delay and client-decoder paths honestly remained |
| E04 | Written rules and named oracles; five faults plus A1 decision | Oracle charter met; A1 observation was informed by a leak, not blind |
| E05 | Directed sequences and random walks, reduced to three release-path faults; A2 decision | Sequence charter met; the short regression walk's failure to exercise W1 on the old tag is explicitly retained |
| E06 | Entity lifetime and data across API/client/worker boundaries; seven faults; A4 disclosure | Entity charter met; sampled batch non-hit and HTTP-worker precision sibling not hidden |
| E07 | 49 sequential cells held; overlapping actors exposed two faults; A3 supplied by fixer policy work | State charter met with a meaningful concurrency extension, not rediscovery of D18–D20 |
| E08 | Startup absence, frozen DB, shared shipper and DST; three faults and new silence A5 | Ecosystem charter met; chronology claim needs R3 |
| E09 | Nine site predictions, six yielding reports plus a sibling; seven faults; transactional notifications and centralized transition guards | Cause-driven probing and structural repair achieved; supplied grouping and unsupported trend require R2/R3 |

The canonical defect index, lesson ID lists and named regression wrappers reconcile:

| Register class | Found / disposition | Remaining in that assigned register |
|---|---|---|
| D01–D27 | 26 fully counted blind, D07 partial/disclosed; all fixed E01 | 0 unfixed |
| U01–U45 | 8 in E01; 37 in E02–E09; all fixed in their assigned lessons | 0 unfixed |
| Original silences A1–A4 | A1 E04, A2 E05, A4 E06, A3 E07; all decided, none claimed blind | 0 undecided |
| New silence A5 | Found E08, DST policy decided E08 | 0 undecided in this ID; related Q-E09-1 remains open |
| T1–T3 (repro IDs T01–T03) | Structured lifecycle logs, event history, lab clock in E02 | 0 unclosed in the original trio |

“Fixed” is bounded by each finding's paths and oracle. It does not close the HTTP
worker's numeric decoder, old stored invalid templates/schedules, delivered-outbox
cleanup, Q-E09-1/2, backward-clock questions, or deferred testability proposals.
Those are explicitly carried forward. The nine verification lists contain
35/5/8/6/4/8/3/4/7 items (80 total: 27 D + 45 U + 5 A + 3 T).

The history is linear with distinct experiment/fix steps. E02's hook split and
lesson-01's diagnostic repair tag are documented exceptions to the simple two-commit
illustration. Project code is unchanged between lesson-09, v1-explored and reviewed
HEAD. I inspected the late patches and their guards, transaction boundaries,
cancellation observation, scheduling and shipper code; their scope matches the
reported findings. No merge, tag rewrite or product change was made in this review.

## Independent record and blindness audit

All **3,844 manifest entries** hash-match their experiment commits. All **3,788
manifest entries for raw experiment artifacts and original F/E reports** also match
HEAD. These are entry counts, not a claim of 3,788 unique files. Living models can
legitimately evolve. The copied older canonical files excluded at E07 collection are
identified separately in its manifest. Evidence: `register/reviews/end/integrity.json`.

I inspected the exporter, all blindness audits, the original tool-input records,
leaked charter text and affected lessons. The exporter rejects forbidden path names,
not forbidden knowledge in otherwise allowed files. Its extraction of PLAN's Probe
paragraph explains how E09's expected answers entered the charter.

| Run / exception | Independent evidence and effect |
|---|---|
| E01 | No hidden-plan/history/other-checkout read identified in tool inputs; manifest matches. Later leaks do not taint this run |
| E04 A1 | Exported lesson03 names it; original notes acknowledge reading that lesson. Exclude A1 from blind discoveries |
| E04 external paths | Root `find` lists skill directories; no subsequent use of those external skill paths identified. Own runtime-task output reads also occur, omitted from summary audit |
| E05/A2, E06/A4, E07/A3 | Charter/policy knowledge, not independent bug discoveries; canonical accounting keeps them separate |
| E07 runtime path | Attempted own task-output read outside export; `.gitignore` search is not git-history access |
| E09 runtime path | Harness-persisted output from its own session; no hidden answer-key content identified |
| E09 expected counts | Explicitly supplied in charter; matching grouping is not independent. New-site predictions and subsequent probes are still evidence |

The retained transcripts support a qualified prompt/export boundary, not OS isolation
or an assertion that no information could have escaped a regex audit. Failed probes,
post-hoc predictions excluded from scores, sampling misses, fixture corrections and
provider/seat changes are generally unusually well recorded. R5 is a precision fix,
not evidence that acceptance was fabricated. The Sol-to-Opus coordinator/fixer change
and same-company C review are disclosed; interrupted Codex REF output was not counted
as a successful completed acceptance run.

## Skills and book-source assessment

All eight skills have triggers, inputs, executable procedures, output shapes, source
sections and concrete worked examples. They teach the experiments' actual lessons:
raw-client/path coverage after decoder mistakes; independent consoles and hook
cross-checks; per-path oracles; successful-action coverage in random walks; second
actors in state tables; owner races and exact bytes across boundaries; and reducing
surprises before searching siblings. Final skills specifically warn about the leaked
E09 answers and stale backlog. Their revised procedures were synthesized from the
runs; they were not all re-executed blind in their final form. Second-target litefleet
examples are explicitly unrun sketches.

Source checks in this review used the publisher's [contents](https://media.pragprog.com/titles/ehxta/toc.pdf)
and three authorized excerpts: [charters](https://media.pragprog.com/titles/ehxta/charters.pdf),
[states](https://media.pragprog.com/titles/ehxta/states.pdf), and
[non-GUI exploration](https://media.pragprog.com/titles/ehxta/explore.pdf).
The contents supports the named section mapping. The charter excerpt supports bounded
questions and new charters from discoveries; the state excerpt supports observable
states, transient conditions and external/system/time events; the API excerpt supports
varying inputs and observing effects without a GUI. These are useful spot checks,
not full-source validation of all nine lessons/eight skills. No book purchase or
unavailable full-text access is implied. R1 remains open.

## Independent runtime validation

Fresh clone: `/tmp/explore-it-astra-review/repo`. Logs and command/exit/time
ledgers: [end-review evidence](register/reviews/end/README.md). The HEAD lab used
`queuelet-astra` (18280/18281/15513); the seeded retry used
`queuelet-astra-seeded` (18284/18285/15515); v0 used `queuelet-astra-v0`
(18282/18283/15514). Each direct-check database URL was set explicitly. Historical
verifies used their own serial `queuelet-ref` stack.

| Independently executed | Result | Seconds |
|---|---|---:|
| HEAD `make checks`, initial setup | **Exit 2**: 93 pass; U09 fixture failure (below) | 689.183 |
| HEAD `make seed` then complete `make checks` | Exit 0; all 94 check functions pass | 813.920 |
| HEAD `go test -count=1 ./...` | Exit 0; uncached unit/application tests | 4.929 |
| HEAD `go vet ./...` | Exit 0 | 0.953 |
| HEAD `make verify LESSON=08 REF=explore-it/lesson-07` | Exit 0; `DEFECTS REPRODUCED: 4 present` | 136.938 |
| HEAD `make verify LESSON=09 REF=explore-it/lesson-08` | Exit 0; `DEFECTS REPRODUCED: 7 present` | 101.722 |
| v0 `make verify LESSON=00` | Exit 0; `BASELINE HOLDS` | 12.545 |
| HEAD D18 instrument against v0 binaries | Exit 0; stale A ack 200, current B ack 409, A's result stored: PRESENT | 1.981 |
| HEAD D21 instrument against v0 binaries | Exit 0; foreign namespace access/mutations admitted: PRESENT | 0.316 |

**My initial setup failure is retained, not a product finding.** I started a fresh
stack and called `make checks` without seeding it. U09 posts to the seeded `emails`
queue, which returned 404. All other 93 functions completed without reported failure;
there are no skip sites in this check package. The corrective run explicitly seeded
an independently named stack and passed the complete suite with unchanged code and
instruments. The first failure, commands and elapsed times remain in `acceptance.json`
and `logs/head-checks.log`; the seeded run has a separate ledger and logs.
All three review stacks were taken down with volumes removed; the historical
verifier also removed its temporary stack.

In REF08 I observed the absent-at-start API fail to recover, requests wait through
the frozen database, missing/double DST firings, and a healthy tenant's notification
starve. REF09 reproduced schedule validation/iteration faults, stale reported offsets,
and lost notifications/post-commit failures under concurrent deletion. D21–D24 stayed
ABSENT in REF08, and D12/D14/D18–D20/D27 stayed ABSENT in REF09. These were archived
historical binaries with unchanged HEAD instruments, not source-reading predictions.

I did **not** separately repeat HEAD verify00–09: the full check suite executes their
80 assigned repro oracles plus baseline, direct and model checks, so those wrappers
would mostly duplicate work. I did not rerun every lesson at its own tag, the entire
v0 proof sweep, full dwell, mutation campaigns or all alternate random seeds. A/B/C's
recorded own-tag results remain inherited evidence. The fresh v0 baseline, two
planted-defect reproductions and two complete late historical replays provide the
independent historical spot checks authorized by the brief.

## Residual limits and handoff

The known U32 race-sampler miss remains an instrument reliability risk despite this
run passing. A future instrument change needs preserved failed evidence, new frozen
hashes and a previous-tag replay. D18 and D21 still lack the separate direct Go
assertions suggested at checkpoint A; historical/current shared-oracle evidence is
valuable but is not two independent assertion implementations. D19 additionally has
the model walk. These carried risks do not turn a sampled zero-hit into proof of all
interleavings.

Checkpoint B explicitly identifies its E02 review as same-company with that fixer.
This Codex review adds independent execution but does not become a new cross-company
E02 review by changing its scope. Any remaining per-item family requirement belongs
to the coordinator; the designated non-Claude end-review gap for E07–E09 is filled.

A follow-up run should address R1–R6, with R4 first for reader reproducibility, and
record its corrections without altering original observations. Nick retains the
already-listed policy decisions. No full dwell sweep, fresh blind skill trial,
production-readiness certification or exhaustive concurrency proof is claimed.
The durable review result and reusable validation runners are retained with this
review; no separate VM knowledge entry is warranted.
