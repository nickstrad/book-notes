# E07-F02 — Purge misses jobs that enter or leave `leased` while it runs: they survive unflagged and become pending again

Seen at: export of `explore-it/lesson-06` (stack `queuelet-e07`), experiment 07 (state-it).
Original observations: [E07 notes](../experiments/07-state-it/notes.md) P2 purgelease/purgenack (18:06:43Z,
H-E07-04/04b), P4 control (18:07:37Z, H-E07-09), P5 siblings (18:08:36Z, H-E07-10/11), reduction (18:11:33Z, H-E07-14).

## Minimal reproduction

Actor: acme key, one worker W, one queue Q (lease 300 s). Clock: real. One leased job suffices (6/20 trials).

```sh
cd books/explore-it/exploration/experiments/07-state-it/tools
. ./env.sh
N=1 python3 -I p5_purge_siblings.py wdel <tag> 20   # purge ‖ worker-delete release (1 job)
python3 -I p2_races.py purgenack <tag> 10           # purge ‖ 60 nacks(retry)
python3 -I p2_races.py purgelease <tag> 10          # purge ‖ 4 workers leasing 200 jobs
python3 -I p5_purge_siblings.py expire <tag> 8      # purge ‖ expiry sweep (lab clock +10 s per trial, reset at end)
python3 -I p4_purge_control.py <tag>                # sequential control
```

Shortest form: enqueue 1 job, lease it with W, then start together `DELETE /v1/workers/W` and, 0–20 ms later,
`POST /v1/queues/Q/purge`. Read the job's `state, purge_requested`.

## Expected (named oracle)

SPEC:54 "purge removes all pending jobs: scheduled, queued and retry_wait"; SPEC:55-56 "A job leased when its queue is
purged keeps its lease: an ack completes it, and a nack, lease expiry or worker deletion makes it `dead`, never pending
again." Under any serial order of purge and the concurrent transition the job ends deleted (transition first, then
purge removes the now-pending job) or flagged/dead (purge first); it never ends pending after both requests return.

## Observed

```text
control (sequential): leased 10, purge deleted 10, leased:true:10 → 10 nacks → dead:10
N=1 purge ‖ worker delete: queued:false:1 ×6, dead:true:1 ×1, deleted ×13           (20 trials)
60 jobs purge ‖ worker delete: trial 0 delete 200 purge {'deleted': 0} -> ['queued:false:60']  (9 of 10)
purge ‖ nack(retry): trial 8 purge {'deleted': 10} nack 200×60 states ['dead:4', 'retry_wait:46']  (10 of 10)
purge ‖ expiry sweep: trial 0 purge {'deleted': 0} -> ['queued:false:300']  (7 of 8)
purge ‖ lease: trial 7 purge {'deleted': 0} leased_total=200 leased_flagged=30 leased_unflagged=170
   nack(retry) on unflagged 3345 (200) → retry_wait                              (10 of 10)
```

After a purge that answered 200, jobs remain `queued` or `retry_wait` (they will run), or `leased` without the purge
flag (a later nack, expiry or worker deletion re-queues them). Purge's `deleted` count does not include them and
nothing tells the caller they were missed.

## Variables isolated

- Matters: overlap between the purge statement and a transaction that moves a row into or out of `leased` (lease,
  nack, expiry sweep, worker-delete release). The sequential control and the serial outcomes inside the race runs
  (dead, deleted) behave as specified.
- Does not matter: number of rows (1 job is enough), which release path, real vs lab clock (worker-delete and nack
  variants use real time).
- Window size: the purge-lease run left 77–170 unflagged jobs, more than the ≤40 rows four leasers can hold at once,
  so the window is the whole purge statement, stretched by its own lock waits (inference).
- Mechanism (inference from `api/queues.go:170`): `marked` (`UPDATE … WHERE state='leased'`) and `removed`
  (`DELETE … WHERE state IN pending`) select rows by the statement snapshot. A row `leased` in the snapshot that a
  concurrent transaction makes pending is skipped by `marked` on re-check and was never a `removed` candidate; a row
  `queued` in the snapshot that a concurrent lease makes `leased` is skipped by `removed` on re-check and was never a
  `marked` candidate.

## Sibling search

Predictions for siblings were written before P5 (notes ≈18:08Z, ahead of the 18:08:36Z run) and both reproduced (expiry sweep, worker release).
Other transitions racing purge, reasoned from guards, not probed: promote scheduled/retry_wait → queued (both pending:
still deleted — none), cancel (terminal either way — none), operator retry dead → queued (linearises as retry after
purge — acceptable), enqueue/schedule fire (new job after purge — acceptable). Same snapshot pattern elsewhere:
`Retain` (`expire.go:73`) and `deleteJob` (`api/jobs.go:33`) only touch terminal rows, which have no outgoing
transitions except delete/retry — none found. Related race class: [E07-F01](E07-F01-worker-delete-racing-lease-leaves-zombie-lease.md).

## Severity and status

Severity: **medium–high**. Purge is the operator's "stop this work" action (e.g. a poisoned or mistaken batch); on a
queue with active workers it silently leaves some — in the lease race, most — of the jobs to run, and reports a
count that hides them. The lesson-05 A2 rule ("never pending again") holds only without concurrency.
Status: **open**.
