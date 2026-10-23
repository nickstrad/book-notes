# E07 notes — state-it (build `explore-it/lesson-06`, export `/tmp/explore-it-experiment-07`)

Stack: Compose project `queuelet-e07`, image tag `e07`, API `http://localhost:18280`, hooksink
`http://localhost:18281`, Postgres `localhost:15513`. Env file: [tools/env.sh](tools/env.sh).
Clock: lab test clock (admin-key). Actors: workers created per probe (A, B) in namespace acme.

## 2026-10-22T17:59:17Z — run start (reading)

#read STATE.md, charter, PROTOCOL.md, skill probe-sequences-and-states, models/states.md (E01–E05 tables),
defects/INDEX.md (all D/U records fixed through lesson 06), SPEC.md, README.md.

## 2026-10-22T18:00:30Z — stack built and started

`make up` with the e07 variables exported → exit 0 (results/e07/up.log). Other stacks on the host
(queuelet-e01-*, tsdb-*) were listed read-only and not touched.

## 2026-10-22T18:02Z — #read guards (source route; evidence for predictions only)

| Verb | Guard (`WHERE`) | file:line |
|---|---|---|
| lease | `state='queued' AND kind=ANY(worker.kinds)` `FOR UPDATE SKIP LOCKED`; queue not paused; worker kinds read **without lock**; per-row `UPDATE … WHERE id=$1` (no state re-check) | internal/jobs/jobs.go:50-77 |
| lease (API) | worker exists in namespace — checked in a **separate statement before** the lease tx | internal/api/jobs.go:160-166 |
| heartbeat | `lease_token=$2 AND state='leased' AND lease_expires_at > now`; sets expiry = now+extend (may shorten) | jobs.go:100 |
| ack | `state='leased' AND lease_token AND lease_expires_at > now`; attempt row inserted in a **separate** autocommit statement | ack.go:14-21 |
| nack | `SELECT … FOR UPDATE` same guard, reads `purge_requested`; retry_wait unless !retry / attempts+1≥max / purged | nack.go:13-19 |
| expire sweep | `state='leased' AND lease_expires_at <= now FOR UPDATE SKIP LOCKED`; queued or dead (n≥max or purged) | expire.go:12,44-57 |
| worker delete | `DELETE worker` then `state='leased' AND leased_by=$1 FOR UPDATE` in one tx | api/jobs.go:123-145, expire.go:20 |
| cancel | `state NOT IN (succeeded,dead,cancelled)`; clears lease fields | operator.go:6 |
| retry | `state='dead'` → queued, attempts 0, purge flag cleared | operator.go:13 |
| delete | `state IN (succeeded,dead,cancelled)` | api/jobs.go:33 |
| purge | one statement: `UPDATE … purge_requested=true WHERE state='leased'` + `DELETE … state IN (queued,scheduled,retry_wait)` (same snapshot) | api/queues.go:170 |
| promote | `state IN (scheduled,retry_wait) AND run_at <= now` | expire.go:69 |
| retain | `state IN (succeeded,cancelled)` and age > retention_hours | expire.go:73 |

Reading note: the purge statement's two CTEs both evaluate rows against the statement snapshot. A row whose
committed state changes *while purge waits on its lock* is re-checked against the new version only if the snapshot
version already matched. So a row that was `queued` in the snapshot and becomes `leased` (lease racing purge) is
matched by neither CTE; a row `leased` in the snapshot that becomes `retry_wait`/`queued` (nack/expiry/worker
release racing purge) is skipped by `marked` and never a candidate for `removed`. This is a code-reading prediction,
not an observation.

## State diagram (from SPEC + guards; arrows labelled I = inferred until driven)

```text
 enqueue(future)─► scheduled ─promote(run_at≤now)─► queued ◄─promote── retry_wait
 enqueue(now) ───────────────────────────────────► queued                   ▲
                                         lease(W) │  ▲ expiry n<max / worker-delete n<max
                                                  ▼  │                       │ nack(retry, n<max, !purged)
                                                leased(W, token, exp) ───────┘
   leased ─ack(holder, before exp)─► succeeded     leased ─nack(!retry | n≥max | purged)─► dead
   leased ─expiry/worker-delete (n≥max | purged)─► dead
   {scheduled,queued,leased,retry_wait} ─cancel─► cancelled
   dead ─retry─► queued (attempts 0)     {succeeded,dead,cancelled} ─delete─► (gone)
   {scheduled,queued,retry_wait} ─purge─► (gone);  leased ─purge─► leased+purge flag
   {succeeded,cancelled} ─retention─► (gone)
```

## State table (should = SPEC line / rule; predicted = guard reading). ✱ = should and predicted disagree or
the spec is silent; these are driven first. Cells E01–E05 already observed are marked "prior" and only
re-driven where a new sequence/actor context matters.

| State | Event | Should (oracle) | Predicted (code) | Plan |
|---|---|---|---|---|
| queued | lease | leased (SPEC 96-100) | leased | prior |
| queued | heartbeat/ack/nack | 409 (SPEC 16, 107) | 409 | prior (F02 fixed) |
| queued | retry | 409 (SPEC 123-125) | 409 | prior |
| queued | delete | 409 | 409 | prior |
| queued | purge | gone (SPEC 54) | gone | prior |
| queued ✱ | purge **racing** lease(W) | either gone (purge first) or leased+flag (lease first) (SPEC 54-56, linearizability) | **leased, no flag** possible → later nack → retry_wait (pending again) | H-E07-04 |
| scheduled | lease/ack/nack/heartbeat/retry/delete | 409 / not leased | same | prior |
| scheduled | cancel → due | stays cancelled | stays | prior (E05 P8) |
| leased(A) | ack/nack/heartbeat by holder before exp | succeeded/retry_wait/extended | same | baseline in H-E07-01 |
| leased(A), exp passed, pre-sweep | A ack/nack/heartbeat | 409 (SPEC 108-109) | 409 | prior (U22 fixed) |
| leased(A) → expired → leased(B) ✱ | A ack / A nack / A heartbeat with A's token | 409 each; B unaffected (SPEC 107) | 409 (token mismatch) | H-E07-01 |
| leased(B) after A-expiry | B ack | succeeded; attempts rows n1 expired(A), n2 succeeded(B) (SPEC 80-81, 116) | same | H-E07-01 |
| leased(A) → expired → leased(A again) ✱ | A ack with **first** token | 409 | 409 | H-E07-02 |
| leased(A) ✱ | heartbeat extend_seconds smaller than remaining | spec: "extend its lease" — silent on shortening | expiry = now+extend (shortens) | H-E07-05 |
| leased(A) ✱ | DELETE worker A **racing** a lease call by A | no job left leased by a deregistered worker (SPEC 93) | job can stay leased_by deleted worker | H-E07-03 |
| leased ✱ | purge **racing** nack(retry) | dead (purge first) or deleted (nack first) — never pending after purge (SPEC 54-56) | retry_wait survives | H-E07-04b |
| leased, purged | ack | succeeded (SPEC 55) | succeeded | prior A2 regression |
| leased, purged | worker delete / expiry | dead | dead | prior A2 regression |
| dead (purge-killed) ✱ | retry → lease → nack(retry) | retry_wait (flag reset; operator retry is a fresh start) | retry_wait | H-E07-06 |
| leased | cancel → holder ack | 409 | 409 | prior (E04 P5) |
| succeeded/dead/cancelled | heartbeat/ack/nack | 409 | 409 | prior |
| dead | retry | queued, attempts 0, n continues | same | prior (U20) |
| leased ✱ | concurrent ack+nack by holder | exactly one 200; one attempt row | one 200 | H-E07-07 |
| leased(A), concurrent ack(A) ✱ ×2 | — | exactly one 200 | one 200 | H-E07-07 |
| succeeded/cancelled | retention | gone | gone | I (not driven; #offcharter) |

## Hypotheses (stated before any probe; written ≈18:03Z, before the 18:03:30Z P1 command — label corrected at 18:14Z)

- **H-E07-01** (charter deep cell). Lease job J with worker A (lease 30 s); advance lab clock 40 s; wait for sweep
  (event `lease_expired`, state queued); lease with worker B; A's ack, nack and heartbeat with A's token each return
  409; then B's ack 200. Attempts: n1 `expired` worker A, n2 `succeeded` worker B. Source: ack.go:14, nack.go:13,
  jobs.go:100 (token guard), SPEC 107. Wrong if any A verb returns 200 or changes J, or attempts differ.
- **H-E07-02**. Same as 01 but the second lease is by A again (same worker, new token): A's first token → 409 for
  ack; second token → 200. Source: token guard. Wrong if the first token is accepted.
- **H-E07-03**. Worker-delete vs lease race (two concurrent requests): over ≤40 trials, at least one trial ends with a
  job `leased` whose `leased_by` worker no longer exists (lease returned 200 with a job after or concurrently with the
  204/200 delete). Should: SPEC 93 "deregisters the worker and releases its active leases". Source: jobs.go:57
  (unlocked kinds read), api/jobs.go:160 (existence check outside tx), expire.go:20 (release sees committed rows).
  Wrong if all trials end with no job leased by a deleted worker.
- **H-E07-04**. Purge vs lease race: with many queued jobs and parallel leasing, a purge leaves ≥1 job `leased` with
  `purge_requested=false` that existed before the purge; a nack(retry) on it yields `retry_wait` (pending again after
  a purge). Should: SPEC 54-56. Source: queues.go:170 snapshot semantics. Wrong if every surviving leased job is
  flagged (or deleted) across ≥10 trials.
- **H-E07-04b**. Purge vs nack race: with N leased jobs and N parallel nacks(retry) overlapping a purge, ≥1 job ends in
  `retry_wait` (or later `queued`) after both finish. Should: none pending (SPEC 54-56). Wrong if all end dead or
  deleted across ≥10 trials.
- **H-E07-05**. Heartbeat extend_seconds=1 on a 30 s lease sets expiry to now+1 (shortens). Should: spec silent
  ("extend"). Wrong if expiry stays at the later value.
- **H-E07-06**. Purge-killed dead job → operator retry → lease → nack(retry) → `retry_wait` (flag cleared). Source:
  operator.go:13. Wrong if dead.
- **H-E07-07**. Concurrent duplicate verbs by holder (ack+ack, ack+nack): exactly one 200, the rest 409; exactly one
  new attempt row. Source: single guarded UPDATE / FOR UPDATE. Wrong if two 200s or two attempt rows.

Planned split: reading ~10 min (done), probing ~35 min, writing ~15 min.

## 2026-10-22T18:03:30Z — #obs P1 stale holder / heartbeat / purge-killed retry (H-E07-01, 02, 05, 06)

Command: `. tools/env.sh && python3 -I tools/p1_stale_holder.py r1` → raw [results/e07/p1-r1.txt](evidence/project-results/e07/p1-r1.txt).
Fixture: fresh queue `e07p1r1` (lease 30 s, max 3), workers A=4, B=5 (kinds echo). Clock: advance 40 s per expiry
step (frozen from first advance), reset at the end (clock moved back from 18:05:22 to real 18:03:34).

```text
H01 lease A 301 attempt 1 exp 2026-10-22T18:04:00.859795936Z
H01 clock {'now': '2026-10-22T18:04:10.86795961Z'}
H01 after sweep: ('queued', 0.85)
H01 lease B -> [(301, 2)]
H01 A ack (409, {'error': {'code': 'state_conflict', 'message': 'state conflict'}})
H01 A nack (409, {'error': {'code': 'state_conflict', 'message': 'sql: no rows in result set'}})
H01 A heartbeat (409, {'error': {'code': 'state_conflict', 'message': 'state conflict'}})
H01 job after A verbs: leased leased_by 5 token==B True
H01 B ack (200, {'ok': True})
H01 final {'state': 'succeeded', 'attempts': 2, 'result': {'by': 'B'}, 'leased_by': None, 'lease_token': None}
H01 attempts [(1, 4, 'expired'), (2, 5, 'succeeded')]
H01 events [('created', None, 0), ('leased', 4, 1), ('lease_expired', 4, 1), ('leased', 5, 2), ('succeeded', 5, 2)]
H02 relet attempt 2 ; ack first token 409 ; ack second token 200 ; attempts [(1, 4, 'expired'), (2, 4, 'succeeded')]
H05 lease exp 18:05:20.87 now 18:04:50.87 ; heartbeat extend 1 → 200 ; job exp 18:04:51.87
H05 after +2s: ack 409 ; state queued (0.95 s)
H06 purge (200, {'deleted': 0}) ; flag leased|t ; nack(retry) 200 → dead
H06 retry 200 → queued|f|0 ; lease, nack(retry) 200 → retry_wait ; attempts [(1,'failed'),(2,'failed')]
```

- H-E07-01 **confirmed**: all three A verbs 409, B's lease untouched, attempts n1 expired(A) n2 succeeded(B).
- H-E07-02 **confirmed**.
- H-E07-05 **confirmed** (heartbeat sets expiry = now+extend and can shorten the lease).
- H-E07-06 **confirmed** (A2 regression held, and operator retry clears the purge flag).

#surprise S1 — the A nack 409 message is the driver text `sql: no rows in result set`, while ack/heartbeat say
`state conflict`. Judgement: **not a lifecycle defect** — status 409 and code `state_conflict` match SPEC 16; SPEC
does not constrain message text. Saved as #offcharter (error-message consistency, nack.go:13 returns the raw
`sql.ErrNoRows`).

#surprise S2 — heartbeat can shorten a lease. Judgement: **not a bug, spec silence** — SPEC 102-103 says "positive
`extend_seconds` to extend its lease" without defining the reference point; the holder controls its own lease and
can only hurt itself (SQS-style "visibility = now+N"). Recorded as question Q-E07-1 for the coordinator.

## 2026-10-22T18:04:27Z — #obs P2 worker-delete racing lease (H-E07-03)

Command: `python3 -I tools/p2_races.py wdel r1 30` → [results/e07/p2-wdel-r1.txt](evidence/project-results/e07/p2-wdel-r1.txt).
Per trial: fresh worker W (kinds echo), 50 queued jobs in queue `e07wdr1` (lease 300 s), barrier-start
`lease(W, max_jobs 100)` and `DELETE /v1/workers/W` (delete delayed 0–20 ms); then read-only SQL
`count(*) WHERE leased_by=W AND state='leased' AND worker W absent`; leftovers cancelled.

```text
trial 0 worker 6 lease 200 jobs=50 err= delete 200 orphan_leased=50
trial 5 worker 11 lease 500 jobs=None err={'error': {'code': 'database_error', 'message': 'sql: no rows in result set'}} delete 200 orphan_leased=0
trial 15 worker 21 lease 200 jobs=50 err= delete 200 orphan_leased=0
… (27 other trials: lease 200 jobs=50, delete 200, orphan_leased=50)
2026-10-22T18:05:24Z hits 28 of 30
```

H-E07-03 **confirmed**: 28/30 trials end with all 50 jobs `leased` by a worker that no longer exists. #surprise S3
(rate far higher than expected) and S4 (lease answers **500 database_error** `sql: no rows in result set` when the worker
vanishes between the API existence check and the lease tx — should be 404 `worker_not_found`, SPEC 15).

## 2026-10-22T18:05:51Z — #obs P3 isolate worker-delete race (C1 control, C2 one job, C3 consequences)

#hypothesis H-E07-08 (**logged after P3 ran** — reasoned before running but written ≈18:08Z; *not counted as stated*, see 18:14Z correction): (C1) sequential lease→delete leaves 0 orphans (release works without
overlap); (C2) with max_jobs=1 the race still orphans ≥1 of 20; (C3) the deregistered worker's token still heartbeats
and acks (200) and a second DELETE of the worker is 404, so the operator has no way left to release the job. Wrong if
C1 orphans, C2 never orphans, or C3 verbs return 409.

Command: `python3 -I tools/p3_wdel_isolate.py r1` → [results/e07/p3-wdel-isolate-r1.txt](evidence/project-results/e07/p3-wdel-isolate-r1.txt);
C1 check [results/e07/p3-c1-check.txt](evidence/project-results/e07/p3-c1-check.txt).

```text
C1 lease 200 50 then delete 200 orphan_leased 0 states          (states query had a GROUP BY 1 bug → empty; re-checked below)
C1 check: cancelled|50|released|36      (all 50 got a `released` attempt row by worker 36, then my cleanup cancelled them)
C2 trial 1 lease 200 n=1 delete 200 orphan_leased=1   … C2 hits 16 of 20
C2 trial 12 lease 404 worker_not_found (delete won before the existence check) orphan 0
C3 worker 57 exists? 0 orphan jobs 50
C3 lease by deleted worker id -> (404, worker_not_found)
C3 heartbeat by deleted worker (200, {'ok': True}) lease_expires_at 2026-10-22 19:06:23 (+1 h)
C3 ack by deleted worker (200, {'ok': True}) succeeded [(1, 57, 'succeeded')]
C3 events j1 [('created', None), ('leased', 57), ('succeeded', 57)]
C3 second DELETE of same worker (404, {'error': {'code': 'worker_not_found', 'message': 'state conflict'}})
```

H-E07-08 **confirmed**. #bug → [E07-F01](../../defects/E07-F01-worker-delete-racing-lease-leaves-zombie-lease.md).

## 2026-10-22T18:06:43Z — #obs P2 purge racing lease (H-E07-04) and purge racing nack (H-E07-04b)

Commands: `python3 -I tools/p2_races.py purgelease r1 10`, `python3 -I tools/p2_races.py purgenack r1 10` →
[p2-purgelease-r1.txt](evidence/project-results/e07/p2-purgelease-r1.txt), [p2-purgenack-r1.txt](evidence/project-results/e07/p2-purgenack-r1.txt).
purgelease: fresh queue per trial, 200 queued jobs, 4 workers each looping 8× lease(max 10); purge after 10–80 ms.
purgenack: fresh queue per trial (linear backoff base 3600 s so retry_wait cannot promote), 60 jobs leased by one
worker, 60 concurrent nack(retry) (0–50 ms jitter) and a purge (10–40 ms).

```text
trial 0 queue e07plr1-0 purge (200, {'deleted': 54}) leased_total=146 leased_flagged=50 leased_unflagged=96
   nack(retry) on unflagged 1975 (200, {'ok': True}) retry_wait
trial 7 queue e07plr1-7 purge (200, {'deleted': 0}) leased_total=200 leased_flagged=30 leased_unflagged=170
… hits 10 of 10 (every trial: unflagged leased jobs; nack → retry_wait)
trial 0 queue e07pnr1-0 purge (200, {'deleted': 32}) nack codes {200: 60} states ['dead:13', 'retry_wait:15']
trial 8 queue e07pnr1-8 purge (200, {'deleted': 10}) nack codes {200: 60} states ['dead:4', 'retry_wait:46']
… hits 10 of 10 (5–46 jobs left retry_wait after the purge returned)
```

H-E07-04 and H-E07-04b **confirmed**. #surprise S5: the lease race leaves far more unflagged jobs (77–170) than the
≤40 rows the four leasers could hold at one instant. Interpretation (inference): the purge statement waits on each
row lock it meets; while it waits, leasers keep committing leases of rows ahead of its scan; every such row is
`queued` in the purge snapshot (so `marked` ignores it) and `leased` when re-checked (so `removed` skips it). The window
is the whole purge statement, which lock waits stretch. Unflagged count > concurrency bound is observed; the mechanism is
inferred from the guards.

## 2026-10-22T18:07:37Z — #obs P4 sequential purge control and P2 duplicate verbs (H-E07-07)

#hypothesis H-E07-09 (**logged after P4 ran**, ≈18:08Z; *not counted as stated*): without overlap, lease 10 of 20 → purge → `deleted 10`, 10 leased flagged → 10
nacks(retry) → 10 dead. Wrong if any job ends pending. Source: A2 fix, queues.go:170.

```text
2026-10-22T18:07:37Z leased 10 purge (200, {'deleted': 10})
after purge ['leased:true:10']
nack codes [200 ×10]
after nacks ['dead:10']
ack+nack job 4582 codes [409, 200] state retry_wait attempts_field 1 attempt_rows 1
ack+ack job 4583 codes [409, 200] state succeeded attempts_field 1 attempt_rows 1
… 40 pairs (20 ack+ack, 20 ack+nack): every pair exactly one 200, one attempt row; 0 violations
```

H-E07-09 **confirmed** (the purge defects need overlap). H-E07-07 **confirmed** (no double completion).
#bug → [E07-F02](../../defects/E07-F02-purge-misses-rows-changing-state-concurrently.md).

## ≈2026-10-22T18:08:3xZ — #hypothesis sibling predictions for E07-F02 (stated before P5; written in the same shell command, ahead of the 18:08:36Z run; label corrected at 18:14Z)

- **H-E07-10** purge racing worker-delete release: 60 jobs leased by W; barrier-start `DELETE worker W` and purge
  (purge jitter 0–20 ms). Prediction: in ≥1 of 10 trials some jobs end `queued` (released by the delete, unflagged)
  after both return. Should: dead or deleted (SPEC 55-56). Source: expire.go:20 + queues.go:170. Wrong if all trials
  end with only dead/deleted rows.
- **H-E07-11** purge racing the expiry sweep: 300 jobs leased (lease 5 s); clock +10 s; purge after a random 0–1.2 s real
  delay (the background loop ticks every 1 s, api.go:221-230). Prediction: in ≥1 of 8 trials some jobs end `queued`
  (unflagged) after the purge returns and the sweep has run. Wrong if every trial ends all-dead or all-deleted.

## 2026-10-22T18:08:36Z — #obs P5 purge siblings (H-E07-10, H-E07-11)

Commands: `python3 -I tools/p5_purge_siblings.py wdel r1 10`, `python3 -I tools/p5_purge_siblings.py expire r1 8` →
[p5-wdel-r1.txt](evidence/project-results/e07/p5-wdel-r1.txt), [p5-expire-r1.txt](evidence/project-results/e07/p5-expire-r1.txt).
Clock: expire trials advance +10 s each (8 × 10 s, frozen), reset at end (18:09:15 real).

```text
trial 0 delete 200 purge (200, {'deleted': 0}) -> ['queued:false:60']
trial 8 delete 200 purge (200, {'deleted': 0}) -> ['dead:true:60']        (purge won: flagged, then released dead)
… wdel hits 9 of 10
trial 0 purge (200, {'deleted': 0}) -> ['queued:false:300']
trial 2 purge (200, {'deleted': 300}) -> []                                  (purge after sweep commit)
… expire hits 7 of 8
```

H-E07-10 and H-E07-11 **confirmed**: a purge that overlaps a worker-delete release or an expiry sweep reports
`deleted 0` and leaves every released job `queued` (unflagged, will run). The serial outcomes appear too (trial 8 dead,
trial 2 deleted), so the outcome depends only on overlap. The high expire rate suggests the 300-row sweep transaction
occupies much of each 1 s tick (inference; sweep duration not measured — see testability request T-E07-1).
Added to E07-F02 as siblings.

## ≈2026-10-22T18:09:2xZ — #hypothesis pre-tick availability (stated before P6; same command, ahead of the run; label corrected at 18:14Z)

- **H-E07-12** leased(A) with `lease_expires_at` passed but before the sweep: B's lease returns `[]` (state still
  `leased`, lease guard `state='queued'`), and B obtains it with attempt 2 within ≤1.5 s real (one tick). A's ack is 409.
  Should: SPEC 108-109 ends the holder's rights at the instant; availability to others is via the loop (SPEC 116, 159) —
  latency not specified. Wrong if B gets it immediately (state already queued) in all trials, or not within 1.5 s.
- **H-E07-13** retry_wait with run_at passed but before Promote: B's lease returns `[]`, then gets it within ≤1.5 s.

## ≈2026-10-22T18:09:2xZ (command start not stamped; ended before 18:10:28Z) — #obs P6 pre-tick availability, run r1 (H-E07-12, H-E07-13) — partly spoiled by lab load

Command: `python3 -I tools/p6_pre_sweep.py r1` → [p6-pre-sweep-r1.txt](evidence/project-results/e07/p6-pre-sweep-r1.txt).
Fixture: queue `e07psr1` (lease 30 s, max 5, linear backoff 60 s), workers A, B. Clock advance +31 s (H12) / +61 s (H13).

```text
H12 trial 0 job 7585: immediate B lease 200 [] state then leased; B got [(7585, 2)] after 0.27 s; A ack 409
H12 trial 1 job 7586: … B got [(7586, 2)] after 1.03 s; A ack 409
H12 trial 2 job 7587: … B got [(7587, 2)] after 1.5 s; A ack 409
H13 trial 0 job 7588: immediate B lease [] state then retry_wait; B got [(7588, 2)] after 0.2 s
H13 trial 1 job 7589: immediate B lease [] state then retry_wait; B got [] after 3.04 s
IndexError (my script assumed a job; dead end, fixed with a guard before r2)
```

#surprise S6 — job 7589 was due but not promoted for >3 s. Follow-up (2026-10-22T18:10:28Z, read-only):
job 7589 later `queued` (event `queued` at app time 18:13:48.41); `/health` 200 `ok`. Lease-expired events per sweep
timestamp ([p6-sweep-batch.txt](evidence/project-results/e07/p6-sweep-batch.txt)):

```text
2026-10-22 18:09:03.290913+00|300   … (7 sweeps of 300 from the P5 expire trials)
2026-10-22 18:11:15.411227+00|49
2026-10-22 18:11:46.411227+00|242
2026-10-22 18:12:47.411227+00|1110
```

The +61 s advance in H13 trial 0 crossed the 300 s expiries of ~1,100 leases left by my P2/P3 race probes. One
Expire transaction released all 1,110; the single background loop (api.go:221-236: Promote → Expire → Fire → Retain
in sequence per 1 s tick) could not promote job 7589 until that sweep ended. Judgement: **not counted as a defect** —
SPEC 159 names the loop's duties but gives no promotion latency bound, and the load was my own pile-up (1,110
expiries in one tick). Saved as #offcharter (unbounded sweep batch stalls promotion/cron) and testability request
T-E07-1 (loop tick duration / real-time stamp on events). Leftover leases checked: none still `leased`.

## 2026-10-22T18:11:02Z — #obs P6 run r2 on a quiet stack

```text
H12 trial 0 job 7590: immediate B lease 200 [] state then leased; B got [(7590, 2)] after 0.68 s; A ack 409
H12 trial 1 job 7591: … after 0.98 s ; trial 2 job 7592: … after 1.01 s; A ack 409
H13 trial 0 job 7594: immediate B lease [] state then retry_wait; B got [(7594, 2)] after 0.76 s
H13 trial 1 job 7595: … after 0.86 s ; trial 2 job 7596: … after 1.01 s
clock reset {'now': '2026-10-22T18:11:08.73999029Z'}
```

H-E07-12 and H-E07-13 **confirmed** (one tick, ≤1.01 s; immediately after the instant the job is neither
leasable by B nor completable by A — a ≤1 s "nobody may run it" gap, consistent with SPEC 108-109 and 116).

## 2026-10-22T18:11:33Z — #obs reduction of E07-F02 (purge racing worker-delete) by row count

#hypothesis H-E07-14 (**logged after the run**, ≈18:12Z; *not counted as stated*): the purge/worker-delete race needs no batch — with one leased job ≥1 of 20 trials
ends `queued`. Wrong if 0/20.

Commands: `N=1 python3 -I tools/p5_purge_siblings.py wdel n1 20`, `N=5 … wdel n5 20` →
[p5-wdel-n1.txt](evidence/project-results/e07/p5-wdel-n1.txt), [p5-wdel-n5.txt](evidence/project-results/e07/p5-wdel-n5.txt).

```text
N=1: queued:false:1 ×6 ; dead:true:1 ×1 ; [] (deleted) ×13      → hits 6 of 20
N=5: queued:false:5 ×13 ; [] (deleted 5) ×7                      → hits 13 of 20
```

H-E07-14 **confirmed**: one job is enough. All three outcomes occur — deleted (release committed before the purge
snapshot), dead (purge flagged first) and **queued** (overlap), the last being the defect.

## ≈2026-10-22T18:13:2xZ — #hypothesis H-E07-15 (label corrected at 18:14Z; written ahead of the 18:13:24Z run) full state × verb matrix at lesson-06 (stated before P7)

Drive every (state × operator/worker verb) cell once with a stale/foreign token where a token applies.
States: scheduled, queued, leased (holder token known), retry_wait, succeeded, dead, cancelled. Verbs: lease (by
another worker), heartbeat/ack/nack with a wrong token, cancel, retry, delete.
Predicted (guards table above): lease takes only `queued`; wrong-token heartbeat/ack/nack 409 everywhere; cancel 200
for scheduled/queued/leased/retry_wait and 409 for terminal; retry 200 only for dead; delete 204 only for terminal and
409 otherwise; every rejected verb leaves state, attempts and lease fields unchanged. Wrong if any cell differs.

## 2026-10-22T18:14Z — #question log-discipline correction (kept, not rewritten)

Several timestamp labels above were my estimates and some ran ahead of the shell clock; I corrected them in place
to the command stamps (corrections marked "label corrected at 18:14Z"). H-E07-08, H-E07-09 and H-E07-14 were reasoned
before their probes but written to this log only after the probes ran; they are kept for the record and **excluded
from the stated/confirmed counts**. The E07-F02 report's reference to "notes 18:10Z" is corrected to 18:08Z.

## 2026-10-22T18:13:24Z — #obs P7 state × verb matrix (H-E07-15)

Command: `python3 -I tools/p7_matrix.py r1` → [p7-matrix-r1.txt](evidence/project-results/e07/p7-matrix-r1.txt).
Fixture: queue `e07mxr1` (lease 300 s, max 3, linear 3600 s), A=172 (holder), B=173 (other leaser). Each cell built a
fresh job in the target state (queued leftovers cancelled first). Wrong token = 32 zeros. Real clock.
`=` means state, attempts, leased_by, token presence and attempt-row count all unchanged.

```text
state | lease_B | heartbeat_wrong | ack_wrong | nack_wrong | cancel | retry | delete
scheduled | 200:[] = | 409 = | 409 = | 409 = | 200 scheduled→cancelled | 409 = | 409 =
queued | 200:[7724] queued→leased | 409 = | 409 = | 409 = | 200 queued→cancelled | 409 = | 409 =
leased | 200:[] = | 409 = | 409 = | 409 = | 200 leased→cancelled | 409 = | 409 =
retry_wait | 200:[] = | 409 = | 409 = | 409 = | 200 retry_wait→cancelled | 409 = | 409 =
succeeded | 200:[] = | 409 = | 409 = | 409 = | 409 = | 409 = | 204 succeeded→gone
dead | 200:[] = | 409 = | 409 = | 409 = | 409 = | 200 dead→queued | 204 dead→gone
cancelled | 200:[] = | 409 = | 409 = | 409 = | 409 = | 409 = | 204 cancelled→gone
```

H-E07-15 **confirmed**: all 49 cells match the guard predictions; no rejected verb had a side effect. Without
concurrency the lifecycle at lesson-06 has no arrow the state table lacks. The only extra arrows found in this run
need two actors overlapping in time (E07-F01, E07-F02).

## 2026-10-22T18:15Z — probing stopped; writing

Remaining candidate races reasoned from guards (not driven): cancel ‖ lease/nack/release, retry ‖ purge, promote ‖
purge, ack ‖ ack across workers — each linearises to a serial order by its re-checked guard (see E07-F02 sibling list).

## 2026-10-22T18:16:26Z — evidence preserved; own stack stopped

Raw outputs copied from the git-ignored `project/results/e07/` to [evidence/project-results/e07/](evidence/project-results/e07/)
(links above repointed there). `docker compose stop` for project `queuelet-e07` only (containers exited, volume kept for
inspection); the unrelated `queuelet-e01-*` and `tsdb-*` containers were left running and untouched.
