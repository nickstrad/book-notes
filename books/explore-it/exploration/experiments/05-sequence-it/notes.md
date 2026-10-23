# Experiment 05 notes — sequence-it

Build: plain export of `explore-it/lesson-04` at /tmp/explore-it-experiment-05 (no git). Explorer: fresh headless
Claude Opus 5.5 session. Charter: [charter.md](charter.md). Skill: probe-sequences-and-states. Protocol: ../../PROTOCOL.md.
Stack: COMPOSE_PROJECT_NAME=queuelet-e05, image tag e05, API :18260, hooksink :18261, PG :15493, test clock enabled
(compose `QUEUELET_TEST_CLOCK: "1"`). Env helper: [tools/e05env.sh](tools/e05env.sh) (adapted copy of
04-judge-it/tools/e04env.sh; inherited copy unchanged). Raw output: `project/results/e05/`.

Planned split (60 min): reading 12 / probing 35 / writing 13.

## 13:17:51Z start #read

Run started 13:17:51Z (`date -u` before reading). Read STATE.md, charter, PROTOCOL.md, SKILL.md, SPEC.md, models/rules.md,
states.md, nouns-verbs.md, hypotheses tail, lessons/04 (previous lesson; the checkpoint this export is cut from).

- #read `internal/jobs/operator.go:6` Cancel: `UPDATE … state='cancelled',finished_at=now WHERE state NOT IN (terminal)` —
  does not clear `lease_token`, `leased_by`, `lease_expires_at`. A cancelled-from-leased row keeps lease fields.
- #read `operator.go:13` Retry: only `state='dead'`; sets queued, attempts=0, run_at=now; keeps last_error/result.
- #read `internal/api/jobs.go:126` deleteWorker: leased jobs of that worker → queued, run_at=now, lease cleared,
  **attempts unchanged and no attempts row**. Lease returns `attempt = attempts+1` (`jobs.go:77`). So a release by
  worker deletion is invisible to the attempt budget: lease → delete worker → lease … can repeat without bound.
- #read `jobs.go:53` Lease returns [] when paused. `expire.go:137` Promote ignores `paused` (scheduled/retry_wait → queued
  while paused). `expire.go:87` Expire ignores paused (lease can expire → dead while paused).
- #read `schedule/fire.go:74-81` while paused, a due schedule advances `next_run_at` without inserting; unpaused, at most
  one job with `run_at=due`, `created_at=now`.
- #read `queues.go:169` purge deletes queued/scheduled/retry_wait only (leased survives).
- #read `ack.go:25`, `nack.go:48`, `jobs.go:100` all guard `state='leased' AND lease_token AND lease_expires_at > now`
  (lesson-04 fix of E04-F04).
- #read `clock.go` /admin/clock (admin key): `set` / `advance` / `reset`; API semantic time only; the loop also uses it.
- #read no FK from jobs.leased_by → workers (001.sql); attempts has no key.

## 13:22:19Z hypotheses (written before any probe) #hypothesis

Rules asserted after every step (walker `check`), drawn from rules.md (E01–E04) and restated as row invariants:

| Inv | Rule | Source / authority | Check (read-only SQL) |
|---|---|---|---|
| I1 | state ∈ {scheduled, queued, leased, retry_wait, succeeded, dead, cancelled} | SPEC Jobs state list | state not in list |
| I2 | finished_at set ⇔ terminal | SPEC terminal + retention uses finished_at | mismatch |
| I3 | attempts ≤ max_attempts | R3/R32 SPEC Never rule | attempts > max |
| I4 | lease grants (executions) since the last operator retry ≤ max_attempts | R3 "No job may execute more than its maximum attempts"; "the lease's attempt counts toward max_attempts" | walker counts grants per job (reset on retry) |
| I5 | attempt rows numbered 1..k, unique | R38 (lesson-04 fix) | dup / gap |
| I6 | a cancelled job is never leased again | R5/R37 | walker: lease returns a job it saw cancelled |
| I7 | leased ⇔ lease_token/leased_by/lease_expires_at non-null | approximation (model consistency; SPEC silent on fields) | mismatch |
| I8 | lease attempt number is strictly increasing across consecutive leases of a job (without operator retry) | SPEC "Attempt n numbers every execution and is never reused" | walker: attempt ≤ previous |
| I9 | retry_wait/scheduled run_at ≥ event_at (never before the failure) | R35 | run_at < event_at |
| I10 | stats per state = direct count | R34 | mismatch |

| Id | Statement / prediction | Source | Would be wrong if | Probe |
|---|---|---|---|---|
| H-E05-C | Charter: "every verb either does what it says from any starting point or is refused" | charter (handler reading) | a sequence leaves a job in a state the spec calls impossible | whole run |
| H-E05-01 | purge (scheduled job with idempotency key K) → enqueue delay 60 with K → new id, scheduled; purge reports deleted ≥1 | queues.go:169; SPEC purge | old id returned or purge leaves scheduled | P1 |
| H-E05-02 | pause → advance clock 3600 → resume, with a 60 s delayed job and a `* * * * *` schedule: during pause lease [] ; delayed job promoted to queued while paused (Promote ignores paused); schedule inserts 0 jobs while paused and next_run_at advances; after resume 0 catch-up jobs and the delayed job leases | fire.go:74, expire.go:137, jobs.go:53 | lease during pause, or >1 catch-up job | P2 |
| H-E05-03 | lease → purge → nack: purge deleted 0 (leased not pending), nack 200 → retry_wait, job later leasable (attempt 2) | queues.go:169, nack.go | purge deletes leased, or nack 409 | P3 |
| H-E05-04 | cancel → retry: 409; and dead → cancel: 409; succeeded → retry 409 | operator.go:6,13 | any 200 | P4 |
| H-E05-05 | lease → cancel: the cancelled row keeps lease_token/leased_by/lease_expires_at (violates I7) | operator.go:6 | fields cleared | P5 |
| H-E05-06 | lease → DELETE worker → lease (new worker) repeated N = max_attempts+2 times: every lease reports attempt 1, attempts stays 0, no attempt rows, job never dead → I4 and I8 violated | api/jobs.go:126, jobs.go:77 | attempt increments or job dead/unleasable after max | P6 |
| H-E05-07 | lease (lease_seconds 2, max 1) → pause → let lease expire: expiry still records attempt and job goes dead while paused | expire.go:87 | job stays leased while paused | P7 |
| H-E05-08 | cancel scheduled job → advance clock past run_at → stays cancelled; cancel retry_wait → stays cancelled | expire.go:137 | becomes queued | P8 |
| H-E05-09 | lease → delete worker → old holder heartbeat/ack with old token → 409 | jobs.go:100, ack.go:25 | 200 | P6 tail |
| H-E05-10 | random walk (seeded) over the verbs {enqueue, enqueue-delay, lease, ack, nack, nack-noretry, heartbeat, cancel, retry, purge, pause, resume, delete-worker, new-worker, advance-clock, delete-job} asserting I1–I10 after every step finds only the I4/I7/I8 class violations predicted above | above | any other invariant violation | walk.sh |

## 13:22:22Z stack check #obs

`dc up -d --build` exit 0 (results/e05/up.log, start 13:20Z in up-start.txt). `curl /health` → `{"status":"ok"}`;
`adm POST /admin/clock {"action":"reset"}` → `{"now":"2026-10-22T13:22:22.560Z"}` 200. `sql "select count(*) from jobs"` → 0.
Compose workers `worker-a`/`worker-b` registered as ids 1, 2 (acme, polling `emails`/`reports`, which this run never creates).
All probe queues are named `p*` / `w*` so they cannot be leased by the compose workers.

## 13:22:50Z P1, P3, P4, P5 — `bash tools/pairs.sh P1 P3 P4 P5` (raw: results/e05/p1-p5.txt) #obs

Actor: acme key; fresh workers per probe (kinds `["k"]`). Clock real (reset at 13:22:22Z).

- **P1** purge → enqueue-with-delay, same idempotency key `K1`:
  `enqueue delay 60 K1` → `{"id":1,"state":"scheduled"}` 201; `purge` → `{"deleted":1}`; re-enqueue → `{"id":2,"state":"scheduled"}` 201;
  SQL: only `2 | scheduled | K1`. **H-E05-01 confirmed** (purge of scheduled holds — F25 regression not re-counted).
- **P3** lease → purge → nack: lease job 3 attempt 1; `purge` → `{"deleted":0}`; `nack` → 200; row `retry_wait | attempts 1 |
  run_at = +5 s | last_error "after purge"`; second purge → `{"deleted":1}`, row gone. **H-E05-03 confirmed**. Judgment: the
  leased job outliving the purge is consistent with SPEC ("removes all pending jobs"); not a defect.
- **P4** (max_attempts 1): cancel(queued) 200 → retry **409**; lease → nack → `dead | attempts 1 | finished_at set |
  run_at = failure+5 s` → cancel **409**; lease → ack → retry **409**, cancel **409**. **H-E05-04 confirmed**.
  Note: a dead job keeps a future `run_at` (backoff computed even when dead, nack.go:56-60). No reader uses run_at for dead
  rows; I9 holds (run_at ≥ event_at). Not a defect.
- **P5** lease (job 7, worker 5) → cancel → row: `cancelled | attempts 0 | finished_at 13:22:54.907 | leased_by 5 | tok t |
  lease_expires_at 13:23:24.81` and `GET /v1/jobs/7` returns `"state":"cancelled","lease_token":"edbb…","leased_by":5,
  "lease_expires_at":"2026-10-22T13:23:24.810896Z"`. `GET attempts` not taken here (see P5b). **H-E05-05 confirmed** (I7 violated).
  #surprise (predicted from code, but it is the only exit from `leased` that leaves a live-looking lease): characterized in
  [E05-F02](../../defects/E05-F02-cancel-leased-keeps-live-lease-fields.md).

## 13:23:10Z P6, P7 — `bash tools/pairs.sh P6 P7` (raw: results/e05/p6-p7.txt) #obs

- **P6** queue `p6` max_attempts 2, job 8. Four rounds of (new worker → lease → `DELETE /v1/workers/{id}` → old-token heartbeat):

  | round | worker | lease response | DELETE worker | old-token heartbeat | row after |
  |---|---|---|---|---|---|
  | 1 | 6 | `"attempt":1,"id":8` | 200 `{"ok":true}` | 409 | queued, attempts 0 |
  | 2 | 7 | `"attempt":1,"id":8` | 200 | 409 | queued, attempts 0 |
  | 3 | 8 | `"attempt":1,"id":8` | 200 | 409 | queued, attempts 0 |
  | 4 | 9 | `"attempt":1,"id":8` | 200 | 409 | queued, attempts 0 |

  `GET /v1/jobs/8/attempts` → `{"attempts":[]}`. job_events: `created | leased 6 attempt 1 | queued 6 0 | leased 7 1 | queued 7 0 |
  leased 8 1 | queued 8 0 | leased 9 1 | queued 9 0`.
  A job with max_attempts 2 was handed to four workers, each told "attempt 1"; nothing records the deliveries.
  **H-E05-06 confirmed; H-E05-09 confirmed** (old token 409). **H-E05-C refuted**: the spec says "No job may execute more than
  its maximum attempts" and "Attempt n numbers every execution … never reused; the lease's attempt counts toward
  max_attempts". #bug → [E05-F01](../../defects/E05-F01-worker-delete-release-uncounted-attempt.md).
- **P7** queue p7 lease_seconds 2, max 1: lease job 9 (worker 10) → pause → sleep 4 → row `dead | attempts 1 | last_error
  "lease expired"`, attempt row n1 `expired`. **H-E05-07 confirmed**: expiry is independent of pause. Not a defect (pause is
  about new leases; the holder's lease ran out).

## 13:23:29Z P8, P2 — `bash tools/pairs.sh P8 P2` (raw: results/e05/p8-p2.txt) #obs

- **P8** linear base 30: job 10 scheduled (+30 s), job 11 leased → nack → retry_wait; cancel both 200; `advance 120` →
  `{"now":"2026-10-22T13:25:29.414039622Z"}`; after 3 s both still `cancelled` (10: attempts 0; 11: attempts 1, last_error x);
  lease → `{"jobs":[]}`. **H-E05-08 confirmed**.
- **P2** (clock already advanced +120 s by P8): job 12 enqueued delay 60 → scheduled, run_at 13:26:29.414; schedule p2s
  `* * * * *` next_run_at 13:26:00; pause; `advance 3600` → now 14:25:29.414; after 3 s: job 12 **queued** (promoted while
  paused), schedule next_run_at 14:26:00, last_run_at null, no schedule job; lease while paused → `{"jobs":[]}`; resume; after
  3 s still one job; lease max 10 → only job 12 (attempt 1). **H-E05-02 confirmed** (0 catch-up; promotion ignores pause but
  lease refuses). Not a defect: SPEC's pause promise is "prevents new leases and scheduled deliveries"; nothing was delivered.
- #surprise (instrument): job 12 `created_at` = 13:25:29.41404 = the instant returned by P8's advance, 3 s later in real
  time. `clock.go:182-191`: `advance` and `set` store a **fixed** time; the semantic clock is frozen until `reset`. Not a
  SUT defect — documented test-clock semantics ("API semantic time only"); it means leases never expire by real time
  while the clock is frozen. Consequence for this run: from 13:23:29Z the clock is frozen; the walker advances it explicitly.
  Testability note: a "advance and keep ticking" (offset) mode would let lease expiry and real timing coexist with jumps.

## 13:24:30Z (approx.; written just before the probe stamp 13:24:37.790Z — first draft of this header mistyped 13:27:10Z) hypotheses for isolating E05-F01 and P5 #hypothesis

| Id | Prediction | Source | Wrong if |
|---|---|---|---|
| H-E05-11 | After one counted failure (lease → nack → retry_wait → advance 10 s → queued), lease reports attempt 2; DELETE worker; next lease again reports attempt 2; attempts rows only n=1 | api/jobs.go:126 | second lease says 3 |
| H-E05-12 | max_attempts 1: lease → delete worker → lease succeeds (2 deliveries with max 1); then nack → dead with attempts 1 | same | second lease empty |
| H-E05-13 | Sibling: DELETE worker 5 (holder of cancelled job 7 from P5) → 200, cancelled row unchanged, still leased_by 5 | api/jobs.go:126 state filter | row changes |
| H-E05-14 | P5b: `GET /v1/jobs/7/attempts` → `[]` (the cancelled in-flight delivery is not in history) | operator.go:6 | a row exists |

## 13:24:41Z isolation results — `bash tools/isolate.sh` (raw: results/e05/isolate.txt) #obs

- **H-E05-11 confirmed**: job 13 (max 3, linear 5): lease (w13) attempt 1 → nack → advance 10 → queued attempts 1 → lease (w14)
  `"attempt":2` → DELETE w14 200 → lease (w15) `"attempt":2` again; row leased, attempts 1; attempts list only n1 failed.
  Attempt number 2 was issued twice: the defect is not tied to the first attempt.
- **H-E05-12 confirmed**: job 14 max_attempts **1**: lease (w16) attempt 1 → DELETE w16 → lease (w17) attempt 1 → nack → dead,
  attempts 1, history one row (n1 failed, worker 17). Two deliveries of a max-1 job; the first is in no record except job_events.
- **H-E05-14 confirmed**: `GET /v1/jobs/7/attempts` → `{"attempts":[]}` for the job cancelled while leased.
- **H-E05-13 confirmed**: DELETE worker 5 → 200; job 7 unchanged (`cancelled | leased_by 5 | tok t | lease_expires_at 13:23:24`).
  The cancelled row now names a worker that no longer exists.

## 13:26:19Z walker written — [walk.sh](walk.sh) #hypothesis

Design: seeded bash `RANDOM`; per seed a queue `w<seed>` (lease 20 s, max 2, linear 5 s), 2 workers; each step picks one of
enqueue(now/delay 15), lease(random worker, max_jobs 1–3), ack, nack, nack retry=false, heartbeat (current token), ack with a
**stale** token, cancel, retry, DELETE job, pause, resume, purge, worker create, worker delete, clock advance (+6 or +25 s,
then 1.6 s real sleep so the loop ticks). After every step: SQL invariants I1–I3, I5, I7 (I7c when cancelled), I9; after
clock steps also I12 (due scheduled/retry_wait not promoted) and I13 (expired lease not swept); I10 stats = direct count;
walker-side I4 (grants since retry > max), I6 (cancelled job leased), I8 (attempt not increasing), R2 (stale-token ack 200).
Each (invariant, job) is printed once at first sight with its step.

Prediction (H-E05-10, refined): across seeds, violations only of I4 and I8 (via worker.delete → lease, E05-F01) and I7c
(cancel of a leased job, E05-F02). Zero I1, I2, I3, I5, I6, I9, I10, I12, I13, R2.
Would be wrong if: any other invariant fires, or I4/I8 fire on a job with no worker.delete between its leases.

## 13:26:45Z walk seed 11, 150 steps — `bash walk.sh 11 150` (raw: results/e05/walk-11.txt, 1 min 19 s) #obs

Smoke run `bash walk.sh 1 12` first died at the summary (`SEEN: unbound variable` under `set -u` with no violations);
removed `set -u`, re-ran seed 1 (discarded) and then seed 11. Clock frozen throughout; walk advances it explicitly.

```
VIOLATION step=24 inv=I7 job=19 state=succeeded
VIOLATION step=78 inv=I8 job=21 attempt 1 <= previous 1
VIOLATION step=86 inv=I7 job=22 state=succeeded
VIOLATION step=100 inv=I7 job=23 state=succeeded
VIOLATION step=119 inv=I7 job=29 state=succeeded
VIOLATION step=128 inv=I4 job=21 grant 3 > max 2 (attempt=2)
VIOLATION step=146 inv=I4 job=24 grant 3 > max 2 (attempt=2)
VIOLATION step=146 inv=I8 job=24 attempt 2 <= previous 2
VIOLATION step=146 inv=I4 job=25 grant 3 > max 2 (attempt=2)
VIOLATION step=146 inv=I8 job=25 attempt 2 <= previous 2
SUMMARY seed=11 steps=150 jobs=23 violations=10 keys= 3 I4  4 I7  3 I8
```

- I4/I8 reduction (job_events for 21, 24, 25): job 21 leased by w25 (attempt 1) → step 26 `worker.delete w=25` → leased by w27
  attempt 1 (I8). Jobs 21/24/25 expired once (attempt row), leased by w30 attempt 2 → step 133 `worker.delete w=30` →
  re-leased attempt 2 (I8) and a 3rd grant on max 2 (I4). Every I4/I8 has a worker.delete between the two leases; none
  without. Same cause as E05-F01 — counted once, not as new defects.
- #surprise **I7 on succeeded rows** (not predicted: I predicted I7 only for cancel). Row 22:
  `succeeded | attempts 1 | leased_by 27 | tok t | lease_expires_at 14:28:44 | finished_at 14:28:24`; `GET /v1/jobs/22` shows
  `"lease_token":"bd4a…","lease_expires_at":"2026-10-22T14:28:44.41404Z","leased_by":27` on a succeeded job.
  #read `ack.go:25` sets state/result/finished_at/attempts but no lease columns. Code-reading prediction gap (I had read this
  line and not noticed). Folded into [E05-F02](../../defects/E05-F02-terminal-rows-keep-lease-fields.md) (renamed from the
  cancel-only file name linked at 13:23:10Z, which was never created).
- #surprise **event console mislabels a worker-deletion release**: `GET /v1/jobs/21/events` shows event 134
  `"event":"lease_expired","worker":30,"attempt":1` at 14:29:01 for the release at step 133, although that lease was valid
  until 14:29:21. #read migrations/002.sql trigger: label `lease_expired` when `OLD.state='leased' AND NEW.last_error='lease
  expired'`; worker release (api/jobs.go:126) leaves `last_error` from the earlier expiry. Hypothesis H-E05-15 below.

## 13:28Z (approx.; header written 13:29:22Z) seeds 12, 13 (200 steps each) launched in background → results/e05/walk-12.txt, walk-13.txt #obs

## 13:29:40Z hypotheses (before probe) #hypothesis

| Id | Prediction | Source | Wrong if |
|---|---|---|---|
| H-E05-15 | Minimal: queue lease 2 s; lease → advance 3 s (expiry, attempts 1, last_error "lease expired") → lease (w2) → DELETE w2 → `/events` last event is `lease_expired` (worker w2) although `lease_expires_at` was in the future; control without prior expiry (P6) labels `queued` | 002.sql trigger; api/jobs.go:126 | label `queued` |
| H-E05-16 | Minimal ack: lease → ack → `GET /v1/jobs/{id}` shows non-null `lease_token`, `leased_by`, future `lease_expires_at` on `succeeded` | ack.go:25 | fields null |
| H-E05-17 | Sibling: a nack(retry=false) → dead row has lease fields cleared (nack.go:60 clears) | nack.go:60 | fields kept |

## 13:32:08Z H-E05-15/16/17 — `bash tools/f02.sh` (raw: results/e05/f02-f03.txt) #obs

- **H-E05-15 confirmed**: queue f3 (lease 2 s, max 5) job 108: lease w52 attempt 1 → advance 3 → `queued | attempts 1 |
  last_error "lease expired"` → lease w53 attempt 2 (expires 14:35:40, clock 14:35:38) → DELETE w53 → row queued attempts 1.
  `/events`: `336 lease_expired w52 attempt 1` (true expiry), `337 leased w53 attempt 2`, **`338 lease_expired w53 attempt 1`**
  (worker deletion, lease not expired). Same lines in `docker compose logs api` ("job event", event_id 336–338).
  #bug → [E05-F03](../../defects/E05-F03-worker-release-logged-as-lease-expired.md).
- **H-E05-16 confirmed**: job 109 ack → `{"state":"succeeded","leased_by":54,"lease_token":"dccc…","lease_expires_at":
  "2026-10-22T14:36:08.41404Z","finished_at":"2026-10-22T14:35:38.41404Z"}`.
- **H-E05-17 confirmed**: job 110 nack retry=false → dead, `leased_by/lease_token/lease_expires_at` null.
  → [E05-F02](../../defects/E05-F02-terminal-rows-keep-lease-fields.md).

## 13:34Z hypothesis H-E05-18 (sibling of E05-F03, before probe) #hypothesis

Prediction (002.sql CASE tests only `OLD.state='leased' AND NEW.last_error='lease expired'`): after one expiry, a
successful **ack** is logged as `lease_expired` (not `succeeded`), and a **cancel** of the leased job also as
`lease_expired`. Wrong if labels are `succeeded` / `cancelled`.

## 13:32:51Z H-E05-18 — `bash tools/f03sib.sh` (raw: results/e05/f03-sibling.txt) #obs

Jobs 111, 112 (lease 2 s): both leased w55 → advance 3 → both `queued | attempts 1 | lease expired` → re-leased attempt 2 →
ack 111 (200) → `GET` `{"state":"succeeded","last_error":"lease expired"}`, event **353 `lease_expired`** attempt 2;
cancel 112 (200) → `cancelled`, event **354 `lease_expired`** attempt 1. **H-E05-18 confirmed** — a success after one timeout
is never logged as `succeeded`. E05-F03 widened (title and table) to cover ack and cancel.

## Walks 12, 13 (200 steps each) #obs

```
walk-12: VIOLATION step=13 inv=I7c job=44 state=cancelled
walk-12: VIOLATION step=55 inv=I7 job=48 state=succeeded
walk-12: SUMMARY seed=12 steps=200 jobs=28 violations=2 keys= 1 I7  1 I7c
walk-13: VIOLATION step=65 inv=I8 job=81 attempt 1 <= previous 1
walk-13: VIOLATION step=67 inv=I7c job=82 state=cancelled
walk-13: VIOLATION step=139 inv=I4 job=81 grant 3 > max 2 (attempt=1)
walk-13: SUMMARY seed=13 steps=200 jobs=30 violations=3 keys= 1 I4  1 I7c  1 I8
```

Job 81 events: leased w46 → queued (step 60 delete w46) → leased w44 → queued (step 69 delete w44) → leased w50 → queued (step
142 delete w50) → purged. Three deliveries with max 2, all `attempt 1` — E05-F01 again (not recounted).

Coverage per seed (op/HTTP counts from the logs): holder verbs mostly refused — successful ack 4/1/0, nack 0/2/1, nack
retry=false 0/2/0, heartbeat 0/1/0 (seeds 11/12/13); stale-token ack 0 accepted of 12/10/8 attempts (R2 holds); retry 200 in
seed 12 only (5); seed 13 had 18 lease calls with no workers left (404, correct). Limitation: the walker picks the target
job uniformly from all jobs it ever created, so holder verbs rarely hit a held lease. A skipped or refused operation is
not coverage of the success arrow. Walker v2 below biases holder verbs toward held jobs.

## 13:33:47Z walker v2 (BIAS=1) — hypothesis before running #hypothesis

Change: `pickh()` — ack, nack, nack retry=false and heartbeat pick (3 in 4) a job the walker has held a token for.
Note on the instrument: `check` reads the clock with `advance 0`, which also freezes it if it were running; the clock is
already frozen in this run, so this is not a new effect.
H-E05-19: with BIAS=1 (seeds 21, 22; 200 steps), successful holder verbs rise to ≥ 10 per seed; violations stay within the
E05-F01 (I4/I8, each after a worker.delete), E05-F02 (I7/I7c) classes. Zero I1, I2, I3, I5, I6, I9, I10, I12, I13 and no
stale-token ack accepted. Wrong if any other invariant fires.

## 13:33–13:35Z walker v2 seed 21 — dead end #obs

`BIAS=1 bash walk.sh 21 200` (and 22 queued after it), raw `results/e05/walk-b21-killed-starved.txt`. At about step 190 the
op tally was: lease 404 ×14 (all workers deleted; walker falls back to worker 999999), lease 200 ×7, ack 409 ×18, nack 409 ×20,
no violations. The bias could not help because the walker had no workers and the queue was paused about half the time.
I stopped it with `pkill -f 'walk.sh …'`; the second, broader `pkill -f walk.sh` also matched my own shell command
(exit 144) and stopped seed 22 at step 8 (`walk-b22-killed.txt`). Neither run is counted as coverage.
Fix (walk.sh v2): with BIAS=1, lease first auto-creates a worker when none remain (logged `worker.create auto`).
H-E05-19 (stated 13:33Z) carried over unchanged to seeds 31, 32.

## 13:36:03–13:39:50Z walker v2 seeds 31, 32 — `bash tools/run-walks.sh 31 32` (BIAS=1, 200 steps each) #obs

Raw: results/e05/walk-b31.txt, walk-b32.txt.

```
SUMMARY bias=1 seed=31 steps=200 jobs=32 violations=14 keys= 3 I4  5 I7  1 I7c  5 I8
SUMMARY bias=1 seed=32 steps=200 jobs=44 violations=13 keys= 2 I4  8 I7  3 I8
```

- Successful holder verbs: seed 31 ack 5, nack 1, nack retry=false 1, heartbeat 0 (= 7); seed 32 ack 8, nack 3, retry=false 1
  (= 12). Stale-token acks: 0 accepted of 1 and 4. Lease 200: 25 and 29. Operator retry 200: 0 and 1.
- Every I4/I8 job traced through job_events to a worker.delete between the two leases:
  seed 31 — 168/169 (w69, step 95; second worker acked, so the first execution is invisible), 174 (w70, step 166), 176/177
  (w71, step 185); seed 32 — 212 (w79 step 118, w78 step 133, w81 step 171; **4 deliveries, max 2, then dead**), 213, 214 (w81).
  In the event history these releases show as `queued` or, after an earlier expiry, as `lease_expired` with the attempt
  column going *down* (e.g. 174: `leased(70:2) > lease_expired(70:1)`) — E05-F03 seen from the walk.
- No I1, I2, I3, I5, I6, I9, I10, I12, I13 or R2 violation in any walk (5 completed walks, 950 steps).
- **H-E05-19**: on its stated refutation condition (another invariant fires) → **confirmed**. Its coverage sub-prediction
  (≥ 10 successful holder verbs per seed) was missed for seed 31 (7): an instrument prediction, recorded, not scored as a
  code-reading contradiction.

## 13:40:11Z stop and close #obs

`final-snapshot.txt` (job counts per queue/state, worker count, frozen semantic clock 14:47:34.414Z; snapshot taken 13:40:29Z), `dc logs --no-color >
all-logs-at-stop.txt` (3019 lines), `stop-time.txt`, then `dc stop` → all five `queuelet-e05-*` containers `exited`
(volume kept; no other compose project touched). Models, hypotheses, defects E05-F01..F03, revealed.md and STATE.md
written 13:40–13:42Z. Elapsed: 13:17:51Z → 13:42Z ≈ 24 min (reading ≈5, probing ≈12, writing ≈7).
