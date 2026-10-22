# Experiment 01 — map-it — revealed

**Status: done.** Charter prediction H00 ("the README's happy path works and the spec is complete") is **refuted**.
The happy path works. The spec is not complete, and much of what it says has no code behind it.

Build: blind export of `explore-it/v0-inherited`, run in an isolated Compose project `queuelet-e01` (API :18220, hooksink :18221,
PG :15453, image `queuelet:e01`) built from this export. The pre-existing `queuelet` project, built from a path outside the export, was not touched
(reason in [notes](notes.md) 01:38:42Z). Evidence: [notes.md](notes.md), scratch in `project/results/`.

## Elapsed (UTC, from `date -u` in the log)
| Phase | Span | Minutes |
|---|---|---|
| Reading (charter, protocol, skill, README, SPEC, lesson 00, all source; stack build ran in background) | 01:38:30–01:41:03 | 2.6 |
| Probing (seed → H01…H41, H25b harness) | 01:41:03–01:52:16, plus H42 01:52:25–01:52:40 and the Q5 check ≈01:57 | ≈11.7 |
| Writing (27 reports, erratum, 8 models, backlog, this file) | 01:52:16–02:00 minus the probes above | ≈ 7.6 |
| **Total** | 01:38:30–02:00:20 | 21.8 of 60 |

The planned split was 20/30/10. Actual reading was much shorter because the codebase is about 2 kLOC. Writing took close to as long as probing.

## What the system showed, in plain words
- **The happy path works.** enqueue → lease → ack → get/attempts/stats via qctl (H01), worker mode with retries (H41), and every documented qctl verb (H40) succeed on the seed.
- **The lease token protects heartbeat and nothing else.** Anyone can ack (F01). Anyone can nack any job in any state (F02), which resurrects succeeded and cancelled work. In practice, job ids are a global namespace across tenants (F03).
- **max_attempts is not enforced.** Nack allows N+1 executions (F04). Lease expiry never counts an attempt, so a crashing job runs forever (F05). Retry also works on succeeded jobs (F06).
- **The background loop is a single point of failure.** One error ends it for the life of the process: a DB restart or one PATCHed bad cron (F08, F09). It also blocks on webhook I/O (F13). /health reports ok throughout, and during a DB outage valid keys get 401 (F14).
- **Schedules are computed once at create.** PATCH does not move next_run_at. Disabling a schedule does not stop it, and enabling one never starts it (F07). The UTC offset is frozen, so DST is ignored (F27, harness route). Resuming a paused queue releases every missed tick at once (F24).
- **Webhooks fail all three of their promises.** The signature is computed over pretty-printed JSON while compact JSON is sent (F10). A 500 response counts as delivered (F11). The result is sent as a truncated string (F12).
- **Input contracts are loose.** Unknown fields, bad enums, out-of-range values and bad name charsets are all accepted (F19, F20). Batch bounds and atomicity are broken (F16). Numbers lose precision (F17). Lease reports an error with a 200 (F21). Deleting a queue orphans its jobs and schedules and returns 204 even when the queue doesn't exist (F15). qctl exits 0 on failure (F22).
- **Reporting/order:** stats count future and retrying jobs as queued (F18), and age is measured from run_at (F26). FIFO breaks among jobs created in the same second (F23).
- **Spec silences:** "pending" is undefined, so purge may be correct as written (F25, status spec). Several others are listed in [notes](notes.md) ("spec silences" table) with code predictions.
- **Holds as specified:** heartbeat token check, idempotency, run_at/delay exclusivity, queue PATCH leaving existing jobs' max_attempts alone, namespace stats authorization, cancel/delete state filters, non-terminal delete 409, auth 401, transport-error webhook retries.

## Model delta (all files new; previously empty)
- [models/map.md](../../models/map.md): processes, routes→handlers, loop steps with order and failure mode.
- [models/nouns-verbs.md](../../models/nouns-verbs.md): 11 nouns, every verb walked; tried vs untested sequences.
- [models/variables.md](../../models/variables.md): 16 variable families with observed effects.
- [models/rules.md](../../models/rules.md): 24 named oracles: 19 SPEC rules violated (R1–R19), 3 hold (R20–R22), 2 non-SPEC oracles violated (R23–R24).
- [models/entities.md](../../models/entities.md), [models/states.md](../../models/states.md) (job and schedule state/event should/predicted/observed), [models/ecosystem.md](../../models/ecosystem.md).
- [models/causes.md](../../models/causes.md): 10 causes grouping F01–F27, with sibling predictions.
- [hypotheses.md](../../hypotheses.md): H00–H42 with outcomes; backlog Q1–Q5 and H43–H48.
- [defects/](../../defects/README.md): F01–F27.

## Process notes and corrections
- **Citation error, corrected.** My early file:line citations came from `cat -n` over concatenated files, so they were cumulative. I appended an erratum to the notes (the log itself is unchanged) and put corrected numbers in the reports and models. No prediction or outcome depended on them.
- **Dead ends kept in the log:** go build VCS stamping, jq quoting, a glob expansion in my sweep output, and `pkill -f` killing my own shell.
- **Lab state:** the isolated stack is still running with its probe data (two API restarts, one postgres restart and one stop, all recovered). To remove it: `COMPOSE_PROJECT_NAME=queuelet-e01 docker compose -f project/deploy/compose.yml down -v`. I did not run that; it is the coordinator's call.

## Unresolved uncertainty and next hypotheses
- F23 reproduced inversions in one of two attempts. The tie-break mechanism (heap vs index order) is inferred.
- F27 is evidenced by the SUT function plus stored state, not by an end-to-end fire. That needs TR2 (clock control).
- F13 uses an approximation oracle, because the spec gives no latency bound.
- Next, ranked: H48 retry/ack across namespaces; H47 crash-kind loop under F05; H44 a webhook URL that stops the loop; H45 catch-up after a loop outage; H43 accounting on worker-delete release; H46 negative retention. Also TR1–TR3 in the notes. Spec decisions are needed for Q1, Q2, Q4, Q5 and F25.

## Score
| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 48 | H00–H42 plus H09c, H10b, H11b, H22b, H25b, each written in notes before its probe ([hypotheses.md](../../hypotheses.md)) |
| Hypotheses confirmed | 44 | All except H00, H10b (refuted) and H22, H25 (needs more). Includes 11 "expected-correct" predictions that held (H01, H29–H33, H36, H37, H39–H41) |
| Hypotheses refuted | 2 | H00 (charter, by F01–F27); H10b (job 892 promoted after API restart, notes 01:49:26Z) |
| Code-reading predictions contradicted by the running system | 1 | H10b: I missed the Promote-before-Fire step order (api.go:101) |
| Defects found | 27 reports: 26 open + 1 `spec` | [defects/README.md](../../defects/README.md) F01–F27; F25 is a spec silence; F27's evidence is harness-based |
| Model cells filled | 110 | Observed table rows: map 20, nouns-verbs 11, variables 16, rules 24, entities 7, states 26, ecosystem 6. Rows marked inferred (Retain, namespace, retention transition) and causes.md groupings are excluded |
| Needs more | 2 | H22 (no inversion in 40 untouched jobs), H25 (no clock control → TR2) |
