# Experiment 06 — relate-it: revealed

Build: plain export of `explore-it/lesson-05`, isolated stack `queuelet-e06` (API :18270, hooksink :18271, PG :15503,
image tag e06, test clock available but **not used** — real UTC throughout). Explorer: fresh headless Claude Opus 5.5
session. Record: [notes.md](notes.md); replayable probes in [tools/](tools/) (`env.sh`, `reset-e06.sh` = adapted copy of
`project/lab/reset.sh`, `p1`…`p10`); raw output in `project/results/e06/` (incl. `final-snapshot.txt`).

## Status

**done.** The charter prediction "deleting an owner leaves orphans" was **refuted** for the owners the API lets you delete
(queue → jobs, queue → schedules): the lesson-01 fix added foreign keys (002.sql:9-17) and a locked 409 check, and
70 race rounds produced no orphan. It **holds** for one pair the schema never linked: job → attempts (E06-F07).
Following the data turned up more than deleting owners did. Seven reports in total.

## Elapsed

15:28:42Z → 15:48Z (stack stopped 15:47:26Z; STATE written after), about 20 minutes of the 60-minute budget.

| Activity | Minutes (approx.) | Notes |
|---|---|---|
| Reading | 3 | 15:28:42–15:31:37: STATE, charter, brief, protocol, skill, SPEC, migrations, the API/jobs/schedule/webhook/worker/client source; the stack built in the background |
| Probing | 10 | 15:31:37 (reset/seed) → 15:41:43 (last log read): P1–P10, three race runs. Prediction tables were written to notes between probes |
| Writing | 7 | 15:41:43–15:48: reports, citation check and correction, models, hypotheses, this file, STATE (plus note-taking between probes, counted under probing) |

## What the system showed

- **Owners refuse rather than orphan.** A queue that has any job (even a cancelled one) or a schedule gets 409
  `queue_not_empty`. An empty queue gets 204, a missing or foreign one 404. Racing 16 enqueues against a DELETE never
  left an orphan. The writes that lose the race, though, are answered **400 invalid_job** (enqueue) or **500
  database_error** (schedule create), with the foreign-key constraint name in the message, instead of 404
  → [E06-F06](../../defects/E06-F06-enqueue-racing-queue-delete-400-fk-text.md).
- **Schedule DELETE always says 204.** It says 204 for an id that never existed, for one already deleted, and for
  another tenant's schedule (which is left alone). GET, PATCH and run of the same ids return 404
  → [E06-F01](../../defects/E06-F01-schedule-delete-missing-204.md).
- **The cron path creates jobs no other path allows.** A schedule's payload may be omitted, `null`, an array, a string
  or a number. `run` rejects these (400), but the cron fire stores them as job payloads. On `emails`, an array payload
  is a poison job for the documented worker: the worker can't decode the lease, never acks or nacks, and logs no job id.
  The job then burns 3 attempts by lease expiry and dies with "lease expired"
  → [E06-F02](../../defects/E06-F02-schedule-payload-not-validated.md).
- **Data journey (one payload through every hop).** The API, jsonb storage, lease bytes, ack, outbox and partner bytes
  keep `9007199254740993` and `1.10` exactly, and the signature is valid over the bytes sent. jsonb and Go's encoder
  normalise the form without changing any value (`1e2`→`100`, key order, last duplicate wins, `<` → `<`), so this is
  not a defect. The loss happens in queuelet's own clients:
  - the sample worker's `echo` returns `9007199254740992` / `1.1`, and that is what the partner is notified
    → [E06-F03](../../defects/E06-F03-sample-worker-echo-loses-precision.md);
  - `qctl enqueue --payload` **stores** the rounded number, and `qctl job get` displays rounded numbers
    → [E06-F04](../../defects/E06-F04-qctl-rounds-payload-numbers.md).
- **Values the store cannot hold turn into a lease error.** An ack result or nack error that contains U+0000 gets
  409 `state_conflict` with driver text, while the caller still holds a valid lease. Enqueue gets 400 for the same
  character → [E06-F05](../../defects/E06-F05-nul-in-result-or-error-409-conflict.md).
- **Job removal leaves attempts behind.** Purge, DELETE and retention remove the job, but its attempt rows (including
  worker error text) stay in psql forever, unreachable through the API. Delivered outbox rows also stay
  → [E06-F07](../../defects/E06-F07-attempts-outlive-their-job.md).
- **What held** (no new defect):
  - queue PATCH changes dependents exactly as specified: existing jobs keep `max_attempts`, while new and `run` jobs
    take the new value; a pending notification keeps its URL and secret (valid signature with the old secret); URL `""`
    creates no more rows;
  - a pending row for a deleted job in a deleted queue is still delivered;
  - deleting a worker counts the released attempt (lesson-05 fix, regression only);
  - job events stay readable through the API and the API log after purge, delete and retention, and those two paths
    agree with psql;
  - the background loop logged no error during any probe.
- **Surprises that are not defects** (reasons in the notes' surprise ledger):
  - jsonb/HTML normalisation;
  - a retention change applies to jobs that are already terminal (SPEC has no snapshot rule);
  - an idempotency key is free again once its job is deleted;
  - the sample worker becomes a zombie after its row is deleted: it polls with 404 every second, never exits or
    re-registers, and the stranded job waits until someone restarts the container. SPEC states no client obligation
    here (Q-E06-1).

## Model delta

- [entities.md](../../models/entities.md), E06 revision:
  - constraint table separating database FKs from application enforcement, with the observed outcome when the owner
    goes (the E01 cell "1→N jobs, schedules (no FK)" is now contradicted for lesson-05);
  - full entity × CRUD grid of observed outcomes;
  - per-hop data-journey table.
- [ecosystem.md](../../models/ecosystem.md), E06 revision: the sample worker, qctl, PostgreSQL acting as a validator,
  the FKs under a race, and webhook snapshots under owner change.
- [rules.md](../../models/rules.md), E06 revision: R13 holds (regression); R7 has new violations; new rules R-E06-1…6.
- [causes.md](../../models/causes.md), E06 additions: five cause groups with timestamped sibling predictions.
- [hypotheses.md](../../hypotheses.md): H-E06-C, H-E06-01…22, Q-E06-1/2, T-E06-1/2.

## Unresolved uncertainties and next hypotheses

- H-E06-17 (batch racing a queue delete): the code has the window, but it was not hit in 320 tries. T-E06-2 asks for a
  lab pause hook so this can be reproduced deterministically.
- H-E06-20: a lease on a queue deleted between lookup and transaction would get 500 `sql: no rows`.
- H-E06-21: the sample worker acking a U+0000 result (`http` kind) would turn a data error into death by expiry.
- H-E06-22: worker registration with NUL would get 500.
- The other qctl verbs (`ack --result`, `batch --json`, `schedule create --json`) are predicted to round numbers the same
  way; not probed one by one.
- T-E06-1: the worker's decode-failure log line names no job.
- Q-E06-2: should attempts and delivered outbox rows follow job retention?

## Score

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 20 | H-E06-C, H-E06-01…19, each written in notes before its probe (H-E06-09a–d counted once). H-E06-20…22 are unprobed backlog and not counted |
| Hypotheses confirmed | 18 | 01–16, 18, 19 (P1–P10) |
| Hypotheses refuted | 2 | H-E06-C: the charter's orphan prediction, refuted for queue→job/schedule (P1, P9b); it holds only for job→attempt. H-E06-17: stated condition "only 201/404" observed in P9c (a sampling non-hit) |
| Code-reading predictions contradicted by the running system | 2 | H-E06-C (charter source: reading migration 001; my own code reading of 002/queues.go had already predicted the refutation, which the run confirmed) and H-E06-17 (enqueue.go window not hit). None of my own code-based predictions about a deterministic behaviour was contradicted |
| Defects found | 7 | E06-F01 … E06-F07 (no passing regression counted) |
| Model cells filled | 30 | entities.md E06 revision: 8 constraint-table rows, 14 CRUD cells, 8 data-journey hops. The ecosystem and rules rows restate these observations and are not counted again; (r) regression and inferred (I) cells excluded |
