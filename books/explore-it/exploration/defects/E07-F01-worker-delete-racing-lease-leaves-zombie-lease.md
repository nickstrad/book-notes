# E07-F01 — Deregistering a worker while its lease request is in flight leaves the job leased by the deleted worker (a "zombie" lease no API call can release)

Seen at: export of `explore-it/lesson-06` (stack `queuelet-e07`), experiment 07 (state-it).
Original observations: [E07 notes](../experiments/07-state-it/notes.md) P2 (18:04:27Z, H-E07-03) and P3 (18:05:51Z,
H-E07-08).

## Minimal reproduction

Actor: acme key, one worker W (kinds `echo`), one queue (lease 300 s). Clock: real (no lab clock needed).
Timing-dependent but frequent: 16/20 trials with one job, 28/30 with 50 jobs.

```sh
cd books/explore-it/exploration/experiments/07-state-it/tools
. ./env.sh
python3 -I p3_wdel_isolate.py <tag>      # C1 control, C2 one-job race ×20, C3 consequences
python3 -I p2_races.py wdel <tag> 30     # 50-job variant
```

Each trial: `POST /v1/workers` → W; enqueue 1 job (`echo`); then start together
`POST /v1/queues/Q/lease {"worker_id":W,"max_jobs":1}` and, 0–10 ms later, `DELETE /v1/workers/W`.
Then read `jobs.state, jobs.leased_by` (read-only SQL) and the worker table.

## Expected (named oracle)

SPEC:93 "`DELETE /v1/workers/{id}` deregisters the worker and releases its active leases"; SPEC:94-95 each release
records an unsuccessful attempt; SPEC:107 "Only the current lease holder may complete or extend a job". Either serial
order is acceptable: lease first → the delete releases the job (queued, `released` attempt row); delete first → the
lease is 404 `worker_not_found`. In no order does a deregistered worker still hold a lease.

## Observed

```text
C1 (sequential control) lease 200 50 then delete 200 orphan_leased 0   (all 50 have a `released` attempt row)
C2 trial 1 lease 200 n=1 delete 200 orphan_leased=1      … C2 hits 16 of 20
C2 trial 12 lease 404 worker_not_found … orphan_leased=0               (delete won: correct)
P2 trial 0 worker 6 lease 200 jobs=50 delete 200 orphan_leased=50      … hits 28 of 30
P2 trial 5 worker 11 lease 500 {"error":{"code":"database_error","message":"sql: no rows in result set"}} delete 200
C3 worker 57 exists? 0 orphan jobs 50
C3 lease by deleted worker id -> (404, worker_not_found)
C3 heartbeat by deleted worker (200, {'ok': True}) lease_expires_at +1 h
C3 ack by deleted worker (200, {'ok': True}) succeeded [(1, 57, 'succeeded')]
C3 events j1 [('created', None), ('leased', 57), ('succeeded', 57)]
C3 second DELETE of same worker (404, {'error': {'code': 'worker_not_found', 'message': 'state conflict'}})
```

Both requests return success (lease 200 with the job, delete 200), the job stays `leased` by an id that no longer
exists in `workers`, there is no `released` attempt or event, the deregistered worker keeps heartbeating (extending
the lease by up to 2147483647 s) and can ack, and a second DELETE is 404 — so the operator's only release tool is gone.
The job is stuck until the lease expires, which the zombie can postpone indefinitely; `cancel` is the only remedy.

Secondary symptom (same window, other order): when the delete commits between the API's existence check and the lease
transaction, the lease answers **500 `database_error` `sql: no rows in result set`** instead of 404 `worker_not_found`
(SPEC:15) — P2 trial 5.

## Variables isolated

- Matters: overlap of the lease transaction and the delete transaction. Sequential lease→delete (C1) releases
  correctly; delete-then-lease is 404 (C2 trial 12).
- Does not matter: number of jobs (1 or 50 both reproduce), queue lease length, lab clock (real time used).
- Mechanism (inference from source, consistent with all outcomes): `api/jobs.go:160-166` checks the worker in a
  separate statement; `jobs.go:57` reads `workers.kinds` with no row lock, so the delete is not blocked; the delete's
  release query `state='leased' AND leased_by=$1 FOR UPDATE` (`expire.go:20`) evaluates the committed (still `queued`)
  row version and skips the row the lease transaction is updating; the lease then commits.

## Sibling search

Source-reading search (not probed unless stated): other read-then-act paths that do not lock the row they depend on.
- `jobs.go:50` Lease reads `queues.paused` without a lock — pause racing lease linearises as lease-before-pause; not a
  violation (none).
- `jobs.go:50`, queue deleted between `queueOrError` and the lease tx → 500 `sql: no rows` (E06's H-E06-20, same class
  as the secondary symptom; not probed here).
- Ack/nack/heartbeat (`ack.go:14`, `nack.go:13`, `jobs.go:100`) check only the token, never that the holder is still
  registered — this is what lets the zombie keep working (observed C3).
- Purge has the analogous snapshot problem with transitions in flight: [E07-F02](E07-F02-purge-misses-rows-changing-state-concurrently.md).

## Severity and status

Severity: **medium–high**. A routine operator action (deregistering a lost or decommissioned worker, e.g. while it is
still polling) silently fails to release work, reports success, and leaves jobs held by an identity that no longer
exists; the deregistered process can keep extending and completing them. No data loss, but the release guarantee and
the "only current holder" rule are violated and the operator cannot detect or repeat the release.
Status: **open**.
