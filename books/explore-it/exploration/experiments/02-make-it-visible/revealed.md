# Experiment 02 — make-it-visible: revealed

**Status: done.** Build: plain export of `explore-it/lesson-01`, isolated Compose project `queuelet-e02` (image `queuelet:e02`,
API :18230, hooksink :18231, postgres :15463). Raw log: [notes.md](notes.md); raw outputs: `project/results/e02/`; replay
helpers: [tools/](tools/). The stack was stopped (`dc stop`) at 03:36:39Z. Other stacks were not touched.

## Elapsed

Run 03:18:01Z → 03:40Z, about 22 min of the 60 min budget.

| Activity | Approx. | Notes |
|---|---|---|
| Reading | 5 min | 03:18–03:21 plus about 2 min of source checks during probing (queues.go URL handling, per-file line erratum) |
| Probing | 13 min | 03:21:03–03:36:39 minus interleaved writing; includes the 2 min image build and two lab resets |
| Writing | 4 min interleaved + 3 min final | defect reports while H13 ran; models, hypotheses, this file after the stack stopped |

## What the system showed

- **Crash lifecycle (charter prediction, first half): confirmed.** A `crash` job on emails was executed exactly 3 times
  by 3 Docker incarnations of worker-a. Each lease expiry wrote an `expired` attempt row, and the job died at attempts 3/3
  32.5 s after enqueue. On reports (lease 20 s, max 4) the job died after 4 executions and produced exactly one signed
  `state:dead` notification. API GET, `attempts`, worker stdout and Docker RestartCount agree on the counts.
  The API log contains **nothing** about any of it.
- **Webhook in 500 mode (charter prediction, second half): confirmed.** 11 POSTs answered 500 and a 12th answered 200.
  The outbox has `attempts 12`, and `delivered_at` is set only at the 200. hooksink has 12 records with valid signatures.
  The API log has exactly one line, `webhook delivered status=200`, with no job or outbox id. Retries run every 5–6 s with
  no backoff and no cap. `/health` stayed `ok` the whole time.
- **Consoles disagree when the partner is slow** (the charter's refutation condition, met in a variation). hooksink in
  `slow` mode answered 200 five times, but queuelet's 2 s client timeout counted five failures and logged none.
  [E02-F03](../../defects/E02-F03-timeout-shorter-than-partner-unbounded-duplicates.md) (spec).
- **One tenant can stop every tenant's notifications.** A globex queue with webhook URL `http://bad host/x` is accepted
  (201 on create, 200 on PATCH). Its first outbox row makes every shipper tick return a parse error. Acme's notification
  then waited with 0 attempts, and hooksink received nothing. Fixing the URL or deleting the job does not recover; only a
  DB reset did. An `ftp://` URL fails per row and does not block anything.
  [E02-F02](../../defects/E02-F02-unparsable-webhook-url-blocks-all-notifications.md) (high).
- **A README-documented kind never runs.** `fail3` is never leased by the sample worker with the documented kind list,
  and no console says why. [E02-F01](../../defects/E02-F01-sample-worker-ignores-failN-kinds.md) (low).
- **A frozen worker's work runs twice, and only the worker knows.** Its stale ack got 409, which is correct (the F01 fix
  holds). The only trace is `completed status=409 error=<nil>` in worker stdout. The job then succeeded on attempt 2 but
  keeps `last_error:"lease expired"`.
- Regression context only, not discoveries: the expiry counting (F05), nack exhaustion (F04), 2xx-only delivery (F11) and
  DB-restart recovery (F08/F14) fixes all held. After the restart, health lags the DB by one tick.

### Surprises and their dispositions

| # | Surprise | Disposition |
|---|---|---|
| S1 | GET shows `last_error:"lease expired"` while the next attempt is leased | Not a bug: SPEC:63 says "last error". Visibility note |
| S2 | Crashed incarnations stay registered, and `last_seen_at` stays at registration time | Not a bug against SPEC (no liveness rule). TR-E02-3 |
| S3 | Worker logs "completed status=200" for every nack | Not a product bug (sample worker logging). TR-E02-1 |
| S4 | A dead job's `run_at` = last nack + backoff | Not a bug: run_at is unspecified for terminal jobs, and Promote ignores dead |
| S5 | `fail3` stays queued | [E02-F01](../../defects/E02-F01-sample-worker-ignores-failN-kinds.md) |
| S6 | A bad URL blocks all tenants | [E02-F02](../../defects/E02-F02-unparsable-webhook-url-blocks-all-notifications.md) |
| S7 | Partner 200 is counted as a failure (slow) | [E02-F03](../../defects/E02-F03-timeout-shorter-than-partner-unbounded-duplicates.md) |
| S8 | `completed status=409 error=<nil>` | Not a product bug (the sample client reports no error for non-2xx). TR-E02-1 |
| S9 | A succeeded job keeps `last_error` | Not a bug against SPEC:63/117. Coordinator question Q-E02-2 |
| — | The watcher showed `leased attempts 3` beside 4 attempt rows | Instrument artifact: sequential queries straddled a tick |
| — | Docker RestartCount was 0 again after `lab/reset.sh` (stop/start) | Docker semantics. Restart history is another console that does not persist |

## Questions no console answered → testability requests

All six requests are specified in [notes.md](notes.md) under "Console comparison" (question, missing signal, minimal hook,
validation probe):

- **TR-E02-1:** correlated structured lifecycle events with ids.
- **TR-E02-2:** per-POST outbox history (status, error, time).
- **TR-E02-3:** backlog and liveness metrics, plus a health cause.
- **TR-E02-4:** a reason why a queued job is not eligible.
- **TR-E02-5:** lease start time in attempts.
- **TR-E02-6:** a controllable loop clock. The crash probe on reports took 84 s of real time; advancing an API clock
  would not advance the worker or partner clocks.

## Model delta

These are new lesson-01 sections, each marked ★; the E01 content is kept unchanged above them:

- [states.md](../../models/states.md): 13 observed cells. New outbox lifecycle; the crash/freeze/expiry-dead arrows
  replace the v0 "Expire (no attempt!)" arrow.
- [ecosystem.md](../../models/ecosystem.md): 7 cells. Docker restart policy, frozen worker, partner 500/slow, URL classes,
  DB restart.
- [map.md](../../models/map.md): 5 cells. Two loops, health composition, console producers.
- [rules.md](../../models/rules.md): 7 cells. R3/R4/R17/R19 at lesson-01; new R25 tenant isolation of the shipper
  (approximation), R26 URL validation, R27 README worker.
- [entities.md](../../models/entities.md): 4 cells (worker incarnations, attempt timestamps, outbox fields, job fields).
- [variables.md](../../models/variables.md): 5 cells (process fate, partner latency, URL class, kind string, queue config).
- [nouns-verbs.md](../../models/nouns-verbs.md): 5 sequences.
- [causes.md](../../models/causes.md): new causes C11–C14. These are cause analysis, so they are not counted as observed cells.
- [hypotheses.md](../../hypotheses.md): E02-H01…H15 and H11a–d, Q-E02-1/2, and backlog H49–H52. Inherited H44 is
  *confirmed*. Inherited H47 is *refuted* at lesson-01.

## Unresolved uncertainty and next hypotheses

- **Not observed:** whether core-loop promotion and expiry continue while the shipper is blocked (E02-F02). Code says
  they run independently (api.go:113-145), but that is an inference only.
- **Not re-triggered:** blocking through the PATCH path (only its 200 acceptance was observed).
- **H49:** ship.go UPDATE errors abort the batch.
- **H50:** other per-row errors stay per row.
- **H51:** `qctl worker run` with the README kinds also skips `fail3`.
- **H52:** attempt numbering after an operator retry of a crash-dead job.
- **Q-E02-1:** observability of process exits.
- **Q-E02-2:** `last_error` semantics after success.
- **Carried over:** E01 Q1 (attempt `started_at`) persists for every outcome at lesson-01.

## Score

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 19 | E02-H01…H15, H11a–d, each written in notes.md before its probe. Inherited E01 backlog also tested: H44, H47 (not counted here) |
| Hypotheses confirmed | 19 | All 19 predictions matched the stated observations (H11d: 200 acceptance observed; its blocking consequence inferred). Inherited H44 confirmed |
| Hypotheses refuted | 0 | No E02 refutation condition was observed. Inherited H47 refuted (job goes dead at lesson-01) |
| Code-reading predictions contradicted by the running system | 0 | All E02 predictions came from lesson-01 code and held. H47 came from v0 code that has since changed, so it is not a misreading |
| Defects found | 3 | [E02-F01](../../defects/E02-F01-sample-worker-ignores-failN-kinds.md) low/open; [E02-F02](../../defects/E02-F02-unparsable-webhook-url-blocks-all-notifications.md) high/open; [E02-F03](../../defects/E02-F03-timeout-shorter-than-partner-unbounded-duplicates.md) medium/spec. Passing regressions excluded |
| Model cells filled | 46 | ★ cells: states 13, ecosystem 7, map 5, rules 7, entities 4, variables 5, nouns-verbs 5. They come from the 19 E02 probes; some observations appear in more than one model, so this is a per-model cell count, not a count of distinct facts |
