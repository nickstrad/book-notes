# E09-F07 — Ack/cancel racing a job DELETE: the transition commits but the API answers 409, and the terminal notification is lost (sometimes with 200)

Seen at: `explore-it/lesson-08` (stack `queuelet-e09`), experiment 09.
Original observations: [E09 notes](../experiments/09-explain-it/notes.md) H-E09-14 (20:41:15Z), surprise at
20:41:40Z, H-E09-16 (20:43:54Z). Predicted before probing (PD6); the silent-200 variant was a surprise.

## Minimal reproduction

Queue e09lerr (acme) with `webhook:{url:"http://hooksink:8201/hook",secret:"s"}`. Test clock fixed. One actor
issues the transition, a second (same tenant) deletes the job as soon as it is terminal.

```sh
. tools/env.sh
python3 -I tools/race_cancel_delete.py e09lerr 300          # cancel vs DELETE spinner, queued jobs
python3 -I tools/race_ack_delete.py e09lerr <worker_id> 300  # ack vs DELETE spinner, leased jobs (worker kinds [e09a])
tools/psql.sh "select count(*) from job_events where job_id between <first> and <last> and event='succeeded'"
tools/psql.sh "select count(distinct job_id) from webhook_outbox where job_id between <first> and <last>"
```

## Expected (named oracle)

- SPEC.md:177 "Terminal transitions produce a webhook outbox entry when the queue has a URL." and SPEC.md:203
  "deleting a job does not erase pending deliveries" — every committed succeeded/cancelled transition gets an entry.
- SPEC.md:16 409 means "A command incompatible with the current state"; an ack or cancel that took effect is not
  incompatible (RFC 9110: 2xx when the action succeeded).

## Observed

```text
cancel race, jobs 319–618: cancel=200/delete=204: 298   cancel=409/delete=204: 2
   409 body {"error":{"code":"state_conflict","message":"sql: no rows in result set"}}
   300 'cancelled' events, 286 outbox rows → 14 lost (12 of them after cancel answered 200, e.g. job 330)
ack race, jobs 841–1140:   ack=200: 232   ack=409: 68
   409 bodies: pq: insert or update on table "attempts" violates foreign key constraint "attempts_job_owner"
               / sql: no rows in result set
   300 'succeeded' events, 300 'deleted' events, 175 outbox rows → 125 lost (57 after ack answered 200)
```

Job 841 events: created, leased, succeeded, deleted — and the worker was told 409.

## Variables isolated

- Matters: a DELETE committing between the transition's commit and the Notify statements. Without the concurrent
  DELETE every ack/cancel was 200 and notified (all earlier probes; H-E09-10).
- Matters: the verb. Ack's window is much wider (68+57/300) than cancel's (14/300) because ack's attempts INSERT is
  a separate autocommit statement after its UPDATE (ack.go:14, :21).
- Does not matter: queue settings beyond having a webhook URL; tenant (single tenant suffices).

## Mechanism (reading)

The terminal UPDATE commits first (ack.go:14 autocommit; operator.go:35 Commit). Afterwards: ack's attempts
`INSERT … SELECT FROM jobs` (ack.go:21, FK error if the job is gone), then `Notify` (jobs.go:91-103): its SELECT
returns `sql.ErrNoRows` if the job is gone (→ error), and its `INSERT … SELECT … WHERE j.id=$1` inserts **zero rows
without error** if the job vanished between the two statements (→ silent loss with 200). Any error becomes 409
state_conflict (api/jobs.go:110-112).

## Sibling check

Predicted before probing: Expire and worker DELETE notify loops — confirmed as E09-F08. Nack → dead (nack.go:35-39)
has the same post-commit Notify: predicted, not probed. Retention deleting a just-finished job (expire.go:91,
retention_hours 0, real clock) is a production trigger of the same window: predicted, not probed.

## Severity and status

Medium–high: workers are told their ack failed when it succeeded (they may redo work); terminal notifications are
lost permanently, sometimes with a success status. Requires concurrent deletion of the just-terminal job (an
operator cleanup or retention race). Status: open.
