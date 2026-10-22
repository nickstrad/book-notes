# E09-F08 — During expiry or worker deletion, one vanished job stops notifications for every later dead job of that pass

Seen at: `explore-it/lesson-08` (stack `queuelet-e09`), experiment 09.
Original observations: [E09 notes](../experiments/09-explain-it/notes.md) H-E09-15 (20:43:02Z–20:43:13Z), H-E09-17
(20:44:30Z). Predicted before probing (PD7) from the habit that produced E09-F01.

## Minimal reproduction

Queue e09dead (acme): `max_attempts:1`, `lease_seconds:30`, webhook to hooksink. Test clock fixed.

```sh
. tools/env.sh
python3 -I tools/race_expire_delete.py e09dead <worker_id> 40 5    # expiry pass makes 5 jobs dead; DELETE spinner on the lowest id
python3 -I tools/race_workerdel_delete.py e09dead 40 5             # DELETE worker releases 5 jobs dead; spinner on the lowest id
tools/psql.sh "select j.id,j.state from jobs j where j.queue_id=(select id from queues where name='e09dead') and not exists(select 1 from webhook_outbox o where o.job_id=j.id)"
```

## Expected (named oracle)

SPEC.md:177 every terminal transition of a webhook queue produces an outbox entry; SPEC.md:204-205 design intent
that one row "does not stop later rows"; SPEC.md:15-16 status codes (a committed DELETE is not a 500).

## Observed

```text
expiry, 40 trials: 8 dead jobs, never deleted, with no outbox row: 714–717 (X=713), 752,753,755,756 (X=751)
   714 events: created, leased, lease_expired — no notification_queued; unchanged 5 s (≥5 ticks) later
   api log at exactly those ticks: "background task: sql: no rows in result set"
worker delete, 40 trials: 39×200, 1×500 {"error":{"code":"database_error","message":"sql: no rows in result set"}}
   worker 14 gone; its jobs 1177–1180 dead with no outbox row
```

## Variables isolated

Matters: the DELETE of the first dead job landing before its Notify. Then all jobs after it in that pass lose
their notification; no later pass revisits them (they are already dead). Frequency 2/40 (expiry), 1/40 (worker
delete) with this spinner. The jobs that lose notifications are not the deleted one and nobody touched them.

## Mechanism (reading)

`Expire` commits, then `for _, id := range dead { if e = j.Notify(id, now); e != nil { return e } }`
(jobs/expire.go:26-29); `deleteWorker` likewise stops at the first error and answers it as 500 after commit
(api/jobs.go:150-159, output() api.go:274-277). Notify errors when its job is gone (jobs.go:94-97).

## Sibling check

Same habit, confirmed this run: Fire (E09-F01). Fixed earlier: E02-F02, E03-F05, E08-F02. Nack → dead notifies a
single job (nack.go:38-39): no batch. Purge/Retain emit events, not notifications: none found.

## Severity and status

Medium–high: notifications for jobs that are still present are lost permanently; worker DELETE reports failure for
a committed deletion. Needs a concurrent deletion; frequency low in this lab. Status: open.
