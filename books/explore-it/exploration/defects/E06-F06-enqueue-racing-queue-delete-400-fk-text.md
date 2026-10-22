# E06-F06 — A write racing queue deletion is answered 400 invalid_job (enqueue) or 500 database_error (schedule create) with foreign-key driver text instead of 404

Seen at: export of `explore-it/lesson-05` (stack `queuelet-e06`), experiment 06 (relate-it).
Original observations: [E06 notes](../experiments/06-relate-it/notes.md) P9 (15:39:03Z, dead end), P9b (15:39:55Z,
H-E06-10), P9c (15:40:33Z, H-E06-16 sibling, H-E06-17 not hit).

## Minimal reproduction

Actor: acme key. Clock: real. Timing-dependent; reproduced 1/480 enqueues and 5/320 schedule creates.

```sh
source exploration/experiments/06-relate-it/tools/env.sh
ROUNDS=30 bash exploration/experiments/06-relate-it/tools/p9b-race.sh        # enqueue variant
ROUNDS=40 bash exploration/experiments/06-relate-it/tools/p9c-race-siblings.sh  # schedule/batch variant
```
Each round: create queue `e6-race`, issue `DELETE /v1/queues/e6-race`, then 16 enqueues staggered 0–9.9 ms
(P9c: 8 schedule creates + 8 two-job batches); tidy up; repeat.

## Expected (named oracle)

SPEC:15 "A missing resource returns 404" (the queue is gone when the write is applied); SPEC:49-50 no orphans. A
valid command must not be called invalid (400) nor be reported as a server fault (500) with internal constraint names.

## Observed

```text
P9b  delete 26×204, 4×409 | enqueue 64×201, 415×404, 1×400
     {"error":{"code":"invalid_job","message":"pq: insert or update on table \"jobs\" violates foreign key constraint \"jobs_queue_owner\""}}
P9c  delete 39×204, 1×409 | schedule create 8×201, 307×404, 5×500 | batch 8×201, 312×404
     {"error":{"code":"database_error","message":"pq: insert or update on table \"schedules\" violates foreign key constraint \"schedules_queue_owner\""}}
orphan jobs 0, orphan schedules 0 (psql, after both runs)
```
Ownership itself holds: the 002 foreign keys plus the `LOCK TABLE` in deleteQueue never let an orphan in.

## Variables isolated

- Matters: the handler looked the queue up (`queueOrError`, `api.go:240-252`; `createSchedule` `schedules.go:42`) before
  the delete committed, and inserted after (FK check fails). Ordering DELETE first and staggering by a few ms exposes it;
  with simultaneous start (P9) enqueues always won (480×201, 30×409).
- Does not matter: payload, kind. Batch: window present in code but not hit in 320 tries.
- Error mapping: `enqueue.go:87` maps every insertJob error to 400 `invalid_job`; `schedules.go:57` maps to 500.

## Sibling search

`enqueue.go:112-114` batch (400, not observed), `schedules.go:144-165` runSchedule (400 via insertJob — predicted),
`jobs.go:42-50` Lease on a just-deleted queue (`sql: no rows` → 500 — predicted, not probed), `fire.go` (schedule
exists ⇒ queue cannot be deleted; none).

## Severity and status

Severity: **low**. Rare, transient, no data harm; misleads clients into treating a valid request as malformed or a
server fault, and leaks schema names. Status: **open**.
