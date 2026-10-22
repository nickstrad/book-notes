# 04 — Judge it: write the oracles down, then break them

An output is only wrong against something. This experiment wrote that something down first: thirteen
Never/Always rules for a job queue, each with its authority (a spec line, an approximation, or a
comparable system), the line of code that should enforce it, and the observation that would break it.
Then it tried to break each one. Eight held. Five broke, and none of the five was a defect the plan had
put here.

The source is Hendrickson, *Explore It!* (2013), Chapter 5: 5.1 "Never and Always", 5.2 "Alternative
Resources", 5.3 "Approximations"; and 10.3 on status codes versus bodies. These citations follow the
approved [chapter map](../PLAN.md#22-chapter-map-kept-sections-only); the builder did not check them
against the book text.

## What the blind experiment learned

[The original record](../exploration/experiments/04-judge-it/revealed.md) is unchanged from the
experiment commit. The explorer stated 16 hypotheses, confirmed 15 and refuted 1: the charter's own
prediction, "the code honours every rule on my list". That was the one code-reading prediction the
running system contradicted. It filed 5 reports and filled 18 model cells, in about 15 minutes of a
60-minute budget.

The record shows a change in how the explorer reads. Before any probe it wrote the rules table in
[`models/rules.md`](../exploration/models/rules.md) (R31–R43). For each rule it wrote the enforcing line
from the source, and for five of them it wrote down in advance that the line was missing: no lower bound
on `delay_seconds`, no `lease_expires_at` in the worker guards, no key on `attempts`, `run_at` left alone
on release, and no `fail()` in front of the mux. All five predictions came true. In each case the
rule told the reader which line to look for, and the line was not there.

| Classification | Outcome |
|---|---|
| Expected here, found early | D09 off-by-one retries, D10 stats counts, D11 same-second FIFO, D25 lease error with 200: all found and fixed in E01. E04 re-probed D09–D11 as R32, R34 and R33, and they held. D25's route was not re-probed; `TestL01D25` still passes |
| Spec silence, resolved here | A1, no default backoff cap: measured (30.34 days at attempt 20) and raised as owner question Q-E04-1, **not blind** (see below) |
| Found as new in E04 | U19–U23 (below) |
| Missed / disclosed | none: no planted defect is left for this experiment |

**A1 was not a blind find.** The lesson 03 note, which is in every later export, says that A1 "(a
default backoff cap) stays with E04". The explorer read it, cited it, measured the 30-day wait and
handed the decision to the owner. It judged the wait compliant with the spec as written, since the only
stated bound is the 68-year overflow cap. That judgement is right, which is why A1 is a spec silence and
not a bug. Naming a register id in an exported note was a leak, though. From lesson 04 on, notes describe
open silences without saying which experiment the plan assigned them to.

**The comparable systems did their job without becoming requirements.** SQS FIFO orders by sequence
number, not by a timestamp, so the explorer could frame a backward clock step as a FIFO risk (H-E04-08)
without calling it a bug. SQS's `ApproximateAgeOfOldestMessage` counts age from the original send. That
is the opposite of the reading behind E04-F05, so the report says plainly that its oracle depends on what
this spec means by "readiness". The explorer recalled both systems' behaviour from memory and marked it
"not verified in this run".

## The five faults, with pointers into `explore-it/lesson-03`

| ID | Report | Where it lived | What happened | Rule |
|---|---|---|---|---|
| U19 | [E04-F01](../exploration/defects/E04-F01-unmatched-route-errors-not-json.md) | `internal/api/api.go:172` hands every authenticated request to `ServeMux`; only handlers call `fail()` | `PUT /v1/queues` → 405 `text/plain` "Method Not Allowed"; `GET /v1/nope` → "404 page not found" | R36 an error is always 4xx/5xx *in the error form* |
| U20 | [E04-F02](../exploration/defects/E04-F02-attempt-history-after-retry-ambiguous.md) | `ack.go:14`, `nack.go:30` insert `n = attempts`; `operator.go:13` resets `attempts` | after dead → retry → ack: `n=1 failed, n=1 succeeded, n=2 failed`; two executions share `n=1` | R38 one row per execution, numeric order = execution order |
| U21 | [E04-F03](../exploration/defects/E04-F03-negative-delay-wraps-and-precedes-failure.md) | `internal/api/jobs.go:67`, `enqueue.go:55` check only `> MaxSeconds` | nack `delay_seconds: -5` scheduled the retry 5 s before the failure; `-9223372037` wrapped to the year 2322 | R35 a retry comes after the failure, within a bound |
| U22 | [E04-F04](../exploration/defects/E04-F04-expired-lease-honoured-until-sweep.md) | `ack.go:10`, `nack.go:14`, `jobs.go:97` guard on token and state, never on `lease_expires_at` | a heartbeat sent just after expiry, before the 1 s sweep, revived the lease (5/5); one tick later the same request got 409 | R39 an expired lease is not current |
| U23 | [E04-F05](../exploration/defects/E04-F05-queued-age-counts-leased-time.md) | `stats.go:18` uses `GREATEST(created_at, run_at)`; `expire.go:41` and `internal/api/jobs.go:126` requeue without touching `run_at` | a job released by expiry reported `oldest_queued_age` 600 s; by worker deletion, 100 s | R40 queued age measures waiting since the job became ready |

U21 is the other half of U13: lesson 03 bounded `delay_seconds` from above only. U23 is a release path
that U08's "later of creation and readiness" (lesson 01) never reached. In both cases an earlier fix
asked part of the question, and here a written rule asked all of it.

## The patch

About 90 lines under `project/`, excluding instruments; the lines that matter:

- **One error form for unmatched requests** (`internal/api/api.go`, `unmatched`). The auth wrapper asks
  `m.Handler(r)` which pattern would serve the request. If none would, it runs the mux's own 404/405
  handler into `statusOnly`, which keeps the status and the `Allow` header and drops the plain-text
  body. Then it writes `not_found` or `method_not_allowed` with `fail()`.
- **History that survives a retry** (`ack.go`, `nack.go`, `expire.go`). Each new attempt row is
  numbered `(SELECT count(*)+1 FROM attempts WHERE job_id=$1)` instead of copying `jobs.attempts`. The
  budget counter still resets on retry, as the spec says; the history no longer does.
- **No negative delays** (`internal/api/jobs.go`, `enqueue.go`). `delay_seconds < 0` is 400 next to
  the existing upper bound.
- **A lease ends when it says it ends** (`ack.go`, `nack.go`, `jobs.go` `Heartbeat`). Each guard gains
  `AND lease_expires_at > $now`. The sweep still records the expired attempt; it just no longer decides
  whether a late worker wins.
- **Release resets readiness** (`expire.go`, `deleteWorker`). `run_at = now` when a leased job goes
  back to `queued`, as operator retry already did, so the stats query is right without changing it.
- **A default backoff cap** (`jobs.DefaultMaxSeconds = 3600`, `jobs.Backoff`). Without
  `backoff.max_seconds`, each wait is capped at an hour, or at `base_seconds` if that is longer, so a
  queue that asks for long waits still gets them. `TestBackoffSaturates` now pins both caps.

Spec lines added to `project/SPEC.md`, one decision each, from
[reconciliation 04](../register/reconciliations/04.md):

- Unmatched paths are 404 and unmatched methods are 405, in the error form.
- `delay_seconds` is never negative.
- Without `max_seconds`, each wait is capped at 3600 s, or at `base_seconds` if that is larger.
- Attempt `n` numbers every execution and is never reused across a retry.
- A lease ends at `lease_expires_at`, with no grace period.
- Release by expiry or worker deletion makes a job ready at the release.

## Rules become checks

The plan's point for this experiment is that a written rule is already half a test. `TestL04Invariants`
runs [`lab/repro/invariants04.py`](../project/lab/repro/invariants04.py). It enqueues 200 jobs at
random priorities. Each round, three workers lease random batches, then ack, nack, dead-letter, cancel
or abandon the jobs at random. The frozen lab clock then advances past the cap so retries come due.
After every round the walker asserts the rules list, R31–R43 compressed into seven invariants:

| Invariant | From |
|---|---|
| I1 a job is never leased by two workers at once | R31 |
| I2 attempts never exceed `max_attempts`; attempt `n` is 1..k, never reused | R32, R38 |
| I3 within a priority, leases are FIFO | R33 |
| I4 stats equal a direct `psql` count, state by state | R34 |
| I5 a retry comes after the failure and within the cap | R35, A1 |
| I6 a stale ack and an unknown route are 4xx with the error envelope | R36, R43 |
| I7 a cancelled job is never leased again | R37 |

Measured on the fixed tree: `INVARIANTS HOLD: I1–I7 over 200 jobs, 12 rounds, seed 4`, in about 34 s.
The walker was not run against `lesson-03`. On that tree I5 must break, because the base is 2000 s and
the second wait would be 4000 s, past the 3600 s cap. I6 would break on the unknown route. That has not
been measured, so treat it as a prediction.

## Verify

From `books/explore-it/project`, sequentially:

```sh
make up
make verify LESSON=04
make verify LESSON=04 REF=explore-it/lesson-03
make checks
make test
```

`lesson-04.ids` lists U19–U23 and A1. Each wrapper runs `lab/repro/probe04.py` through the public API
and the documented lab clock. The probes are raw requests for U19, a dead → retry → ack history for
U20, `-5` on nack and enqueue for U21, a heartbeat one second past `lease_expires_at` for U22, stats
right after each kind of release for U23, and two attempts at base 3000 s for A1. `TestL04U19` …
`TestL04A1` run the same oracles as regression checks.

Measured on 2026-10-22 (logs under `register/runs/04-*`):

| Command | Exit | Elapsed | Result |
|---|---:|---:|---|
| `make up` | 0 | 90 s | current stack built and healthy |
| `make verify LESSON=04` | 0 | 50 s | `FIXED AND CHECKED: 6 absent; regression checks pass` |
| `make verify LESSON=04 REF=explore-it/lesson-03` | 0 | 36 s | `DEFECTS REPRODUCED: 6 present` |
| `go test -tags checks ./test/checks -count=1 -v` (`make checks` plus `-v`) | 0 | 273 s | all 61 integration checks pass (54 earlier, 7 new) |
| `make test` | 0 | <1 s | unit suite passes, including the default cap in `TestBackoffSaturates` |

The first REF attempt stopped with `PROBE ERROR: U21` after U19 and U20 had printed PRESENT. The old
binary accepted the probe's `delay_seconds: -5` enqueue, which made that job due at once. The lease then
returned it ahead of the job the probe meant to nack. This is the FIFO interference that E03's note
warned about, from a new direction. The probe now enqueues the negative delay last, and every E04 probe
requires the lease to return its own job. The failed log and its Compose diagnostics are kept as
`register/runs/04-historical-attempt1*`.

Paired results, the same oracle on both binaries:

| Fault | lesson-03 | lesson-04 |
|---|---|---|
| U19 | `PUT /v1/queues`, `GET /v1/nope`, `PATCH /v1/jobs/1`: 405/404/405 `text/plain` | same statuses, `application/json` error envelope |
| U20 | attempts after dead → retry → ack: `n=[1, 1]` | `n=[1, 2]` |
| U21 | `-5`: enqueue 201, nack 200, retry `run_at` 5.0 s before the failure | 400, 400 |
| U22 | heartbeat 1 s past expiry: `[200, 200, 200]` | `[409, 409, 409]` |
| U23 | `oldest_queued_age` after expiry 600 s, after worker deletion 100 s | 0 s, 0 s |
| A1 | attempt 2 at base 3000 s, no cap: waits 6000 s | 3600 s |

The seeder's answer-key control for A1 now prints ABSENT: attempt 20 at base 5 waits 0.04 days, which
is one hour. The other answer-key controls are unchanged (`04-unresolved-controls.log`).

**Dwell.** Measured only as the pair above (`register/runs/04-dwell.tsv`): all six present at
lesson-03 and absent at lesson-04. No `make dwell` sweep was run, so this note does not claim when each
fault first appeared. Several earlier experiments walked right past them. E01 acked and nacked jobs and
retried a dead one. E02 watched leases expire on three consoles. E03 sent `delay_seconds` at the top of
its range. None of them asked "after the lease has ended?" or "below zero?", because no rule said those
questions mattered.

## Try changing one variable

- Heartbeat at exactly `lease_expires_at` (advance the clock by 30 on a 30 s lease) and one second
  before it: the boundary is `>`, so the first gets 409 and the second 200.
- Set `max_seconds` explicitly below 3600 and confirm it wins over the default.
- Create a queue with `base_seconds: 7200` and no cap: the first wait is two hours, not one.
- Retry a dead job twice and read `attempts`: `n` keeps counting while each lease reports `attempt: 1`.
- Run `python3 lab/repro/invariants04.py I04 7` with another seed.

For further reading: Chapter 5, the [derive-oracles skill](../skills/derive-oracles/SKILL.md), the
[probe notes](../exploration/experiments/04-judge-it/notes.md) (including the instrument slip in P2,
where a helper leased as the wrong worker, and the open questions Q-E04-1…4), and the
[reconciliation](../register/reconciliations/04.md).
