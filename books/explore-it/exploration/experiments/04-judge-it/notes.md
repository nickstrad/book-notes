# Experiment 04 notes — judge-it

Build under test: plain export of `explore-it/lesson-03` at `/tmp/explore-it-experiment-04`.
Stack: Compose project `queuelet-e04`, image tag `e04`, API :18250, hooksink :18251, Postgres :15483.
Every lab command below runs after `source exploration/experiments/04-judge-it/tools/e04env.sh`
(adapted copy of E03's `e03env.sh`; inherited copy unchanged). `api`/`adm` = raw curl as acme/admin key
printing `body [HTTP status]`; `sql` = read-only explorer psql.

Charter: explore retries, ordering, stats and error responses with a written list of Never/Always
rules, the spec, and SQS/pg-boss documentation as comparable systems, to discover where queuelet breaks
rules nobody wrote down. Charter prediction (from `nack.go`, `stats`, lease query): "the code honours
every rule on my list". Refuted by a single violated rule.

Budget plan: 60 min wall clock — reading ~12, probing ~30, writing ~18.

## 2026-10-22T12:39:51Z #read start

Read STATE.md, charter, PROTOCOL.md, skills/derive-oracles/SKILL.md, SPEC.md, README.md, lesson 03
(current checkpoint's disclosed lesson), defects/INDEX.md, models/rules.md, models/states.md, hypotheses.md.

Boundary note (honest record): at 12:40 a `find / -name derive-oracles` I ran to locate the skill listed
directory *names* of other exports/checkouts on this host. I opened none of them and used none of them;
the skill was read from this export.

## 2026-10-22T12:41Z #read source

- `internal/jobs/jobs.go:25-35` `Backoff`: exponential = base·2^(attempt−1), linear = base·attempt, clamped at
  queue cap or MaxSeconds (2^31−1). No default cap (lesson 03 "A1" spec silence, disclosed).
- `jobs.go:37,57` lease order `priority DESC, created_at, id`, `FOR UPDATE SKIP LOCKED`, kind filter, no
  attempts<max check (relies on nack/expire making dead).
- `jobs.go:96-98` Heartbeat guard `lease_token=$2 AND state='leased'` — no `lease_expires_at > now` check.
- `ack.go:10` Ack guard `state='leased' AND lease_token` — no expiry check; attempts+1; then a *separate*
  INSERT into `attempts` and Notify (not one transaction).
- `nack.go:33-45` Nack: `attempts+1 >= max` → dead; `delay` override replaces wait:
  `time.Duration(*delay) * time.Second` — **no lower bound**. API (`internal/api/jobs.go:67`) checks only
  `*delay > MaxSeconds`. Enqueue (`enqueue.go:221-226`) likewise only upper bound. int64 ns overflow for
  delay ≤ −9 223 372 037 s would wrap to a large *positive* duration (U13 fix was one-sided).
- `expire.go:72-101` Expire `lease_expires_at <= now`, n=attempts+1, n≥max → dead; does not touch `run_at`.
- `expire.go:121-123` Promote does not touch run_at. `operator.go:13-15` Retry: dead only, attempts=0,
  run_at=now, created_at unchanged.
- `migrations/001.sql:7` `attempts` table has no key: after operator retry, attempt numbers restart at 1
  and `GET attempts` `ORDER BY n` (`jobs.go(api):25`) has duplicate n values.
- `stats/stats.go:8-26` counts by exact state; oldest_queued_age = now − min(GREATEST(created_at,run_at))
  over queued rows; throughput = succeeded with finished_at ≥ now−1h. Three separate statements (no snapshot).
- `api.go:119-172` ServeMux: unmatched routes/methods are answered by Go's mux, not `fail()`.
- `api.go:43-52` `fail()` is the only JSON error writer; `internal/api/jobs.go:87-89` maps every ack/nack/
  heartbeat/cancel/retry error (including DB errors) to 409 `state_conflict`.

## 2026-10-22T12:44Z #hypothesis — written Never/Always rules (before any probe)

Authority key: S = SPEC.md line; A = approximation; C = comparable system (SQS / pg-boss docs, as
recalled — not this product's requirement); — = spec silence (route to owner).

| Id | Rule | Authority | Enforcing source | Predicted (code) | Probe | Would be wrong if |
|---|---|---|---|---|---|---|
| R31 | A job is never leased by two workers at once | S "Leasing is atomic" | jobs.go:57 SKIP LOCKED | holds | 30 parallel lease calls on 40 jobs, 2 workers | any job id returned twice |
| R32 | Attempts never exceed max_attempts (nack path) | S Worker | nack.go:38 | holds (regression) | max 2: lease/nack ×2 → dead; lease → empty | third lease returns the job |
| R33 | Within a priority, always FIFO by creation order | S Worker; C SQS FIFO | LeaseOrder created_at,id | holds for ids; created_at comes from the app clock | enqueue A,B,C same priority; batch; lease 1 at a time | order ≠ A,B,C |
| R34 | Stats always equal a direct count | S stats "exact state"; A psql | stats.go:8 | holds exactly at rest | stats vs `SELECT state,count(*)` after mixed transitions | any mismatch at rest |
| R35 | A retry always comes after the failure and within a bounded delay | S backoff ("never scheduled before the failure"); — default cap | nack.go:41-44 | computed wait holds; **override `delay_seconds` negative → run_at before failure; ≤ −9 223 372 037 → wraps far into the future** | nack with delay −5, −9223372036, −9223372037 | 400 for these, or run_at within [now, now+?] |
| R36 | An error is always 4xx/5xx with `{"error":{code,message}}` | S Responses | api.go:50 fail() | **unmatched method/path → Go mux plain-text 404/405, not the error form** | PUT /v1/queues, GET /v1/nope, POST /v1/jobs/1/bogus | JSON error body |
| R37 | A cancelled job never runs | S Operator | operator.go:8; ack/nack/heartbeat guards | holds: cancel leased → ack/nack/heartbeat 409; cancel queued → not leased | cancel queued & leased jobs then lease/ack | any 200 / job leased |
| R38 | Attempt history: one row per execution, numbered so numeric order = execution order | S "completed attempts in numeric order"; A consistency | 001.sql:7 no key; Retry attempts=0 | **after operator retry, n restarts at 1: duplicate numbers, history interleaves** | nack to dead (max 2), retry, lease, ack; GET attempts | unique, monotone numbering |
| R39 | Only the current lease holder completes/extends; an expired lease is not current | S Worker; C SQS visibility timeout | ack.go:10, jobs.go:97 | **ack/heartbeat after expiry but before the Expire tick succeed** (spec silent on grace) | clock advance past expiry, immediately heartbeat | 409 |
| R40 | Queued age measures time since the job became ready | S stats "later of creation and readiness" | stats.go:17 | **a job requeued by lease expiry reports age from original run_at (includes leased time)** | lease, clock +3600, expire → stats age | age ≈ 0 |
| R41 | Retry of a dead job makes it immediately available, attempts 0 | S Operator | operator.go:13 | holds | dead → retry → lease attempt 1 | not leasable / attempt ≠ 1 |
| R42 | Backoff follows configured kind/base/cap | S Queues backoff | jobs.go:25 | holds: exp base 2 → 2,4,8; linear base 3 → 3,6,9; cap respected | nack sequences, compare run_at−now | any other wait |
| R43 | Error status classes: 404 missing resource, 409 state conflict, 400 invalid | S Responses | various | holds for verbs checked | nack missing job (404), nack wrong token (409), retry queued (409) | other class |

Charter-level prediction (as stated in charter): the code honours every rule. My own reading predicts
it does **not**: R35, R36, R38 likely violate; R39, R40 are spec-silence/judgment cases.

## 2026-10-22T12:45:06Z #obs stack up

`dc up -d --build` exit 0, `lab/wait.sh` exit 0 at 12:45:06.751Z (results/e04/up.log). No seed run: the stack
holds only migration rows (namespaces acme/globex, keys). Compose workers worker-a/worker-b registered as
worker ids 1 and 2 (acme, polling `emails`/`reports`, which do not exist in this unseeded stack).

## 2026-10-22T12:45:13Z #obs P1 — R36/R43 error forms (results/e04/p1-errors.txt)

```
$ curl -i -X PUT  -H 'Authorization: Bearer acme-key' :18250/v1/queues     → 405, Content-Type: text/plain, body "Method Not Allowed"
$ curl -i        -H 'Authorization: Bearer acme-key' :18250/v1/nope         → 404, text/plain, "404 page not found"
$ curl -i -X DELETE ... /v1/schedules                                       → 405 text/plain (Allow: GET, HEAD, POST)
$ curl -i -X GET    ... /v1/jobs/1/ack                                      → 405 text/plain (Allow: POST)
$ curl -i -X PUT /v1/queues  (no key)                                       → 401 application/json {"error":{"code":"unauthorized",...}}
$ api POST /v1/jobs/999999/nack                                             → 404 {"error":{"code":"job_not_found",...}}
$ api POST /v1/jobs/1/bogus  (job 1 absent)                                 → 404 job_not_found
$ api POST /v1/queues/e04a/bogus                                            → 404 {"error":{"code":"not_found","message":"unknown operation"}}
```
R36 prediction confirmed: status classes are 4xx, but four authenticated unmatched routes answer in Go's
plain-text form, not the SPEC error envelope. #surprise → #bug [E04-F01](../../defects/E04-F01-unmatched-route-errors-not-json.md).
R43 confirmed for these verbs (404 for missing job / unknown operation).

## 2026-10-22T12:45:26Z #obs P2 — R32/R38/R41/R42 nack→dead→retry→ack (results/e04/p2-retry-history.txt)

Clock: `adm POST /admin/clock {"action":"set","at":"2030-01-01T00:00:00Z"}` (frozen). Queue e04a: max 2,
exponential base 2. Instrument slip: worker `w1` got id 3, but `tools/lease1.sh e04a 1` leased as worker id 1
(the Compose worker-a, acme, kinds include echo). It does not poll e04a, so no interference; worker id
appears as 1 in attempts. Later probes use id 3.

```
enqueue → {"id":1,"state":"queued"}; lease → 1 attempt 1; nack f1 → 200
  jobs: 1 | retry_wait | attempts 1 | run_at 00:00:02 | event_at 00:00:00      (wait 2 s = base·2^0 ✔ R42)
clock advance 2 → queued; lease → attempt 2; nack f2 → 200
  jobs: 1 | dead | 2 | run_at 00:00:06 | finished_at 00:00:02                  (R32 ✔; run_at still gets a wait on dead)
lease → {"jobs": []}                                                            (R32 ✔)
retry → 200; lease → 1 attempt 1 (R41 ✔); clock +5; ack → 200
GET /v1/jobs/1/attempts →
  n=1 failed  f1  started 00:00:00 finished 00:00:00
  n=1 succeeded   started 00:00:02 finished 00:00:07
  n=2 failed  f2  started 00:00:02 finished 00:00:02
```
R38 prediction confirmed: after operator retry, numbering restarts; the API's "numeric order" lists the
final success *between* the two failures of the earlier incarnation, and two different executions share
n=1. #surprise → #bug [E04-F02](../../defects/E04-F02-attempt-history-after-retry-ambiguous.md) (judgment
in report). Minor #surprise: a dead job's `run_at` is set to failure+backoff (00:00:06) although it will
never run — not a bug: `run_at` of a terminal row has no documented meaning and Retry overwrites it.

## 2026-10-22T12:45:57Z #obs P3 — R35 negative delay override and enqueue (results/e04/p3-negative-delay.txt)

Queue e04b (max 10, exp base 2), worker 3, clock frozen at 2030-01-01T00:00:00Z. Each nack applies to the job
the lease returned (FIFO handed back promoted job 2/3 on later loops — the same FIFO interference E03 logged;
it does not affect the judgment because each row is read by the leased id).

```
nack delay_seconds=-5            → 200; job 2 retry_wait run_at 2029-12-31 23:59:55   (5 s BEFORE the failure)
nack delay_seconds=-9223372036   → 200; job 3 retry_wait run_at 1737-09-22 00:12:44   (292 years before)
nack delay_seconds=-9223372037   → 200; job 2 retry_wait run_at 2322-04-12 23:47:16.709552  (292 years AFTER)
nack delay_seconds=-18446744073  → 200; job 3 retry_wait run_at 2030-01-01 00:00:00.709552  (+0.71 s: 2^64 ns wrap)
enqueue delay_seconds=-5          → 201 {"id":6,"state":"queued"}  run_at 23:59:55
enqueue delay_seconds=-9223372037 → 201 {"id":7,"state":"scheduled"} run_at 2322-04-12
```
R35 refuted as a product rule (prediction confirmed): a nack override can schedule the retry before the
failure, and a large negative number wraps int64 nanoseconds into a retry/run 292 years in the future — a
*more negative* delay produces a *later* job. #surprise → #bug [E04-F03](../../defects/E04-F03-negative-delay-wraps-and-precedes-failure.md).
Note: E03 judged plain negative enqueue delay "not a bug (spec silent), Q-E03-2"; the new facts here are the
wrap to the future (monotonicity broken) and the retry-before-failure on the nack path, which SPEC:30-31
("a retry is never scheduled before the failure") speaks to for computed waits.

## 2026-10-22T12:46:17Z #obs P4 — R31 concurrent leasing (results/e04/p4-concurrent.txt, p4-l*.txt)

Queue e04c (lease 600 s), batch of 40 echo jobs (ids 8–47), workers 3 and 4 (w1, w2). 30 background
`lease1.sh e04c {3|4} 3` calls in parallel:
```
distinct leased: 40  rows: 40  none-replies: 16   (w3 19 jobs, w4 21 jobs; 8+8 empty replies)
DB: leased | 40
```
R31 confirmed: no job id returned twice; empty replies only after all 40 were taken.

## 2026-10-22T12:46:50Z #obs P5 — R37 cancel, R34 stats (results/e04/p5-cancel-stats.txt)

Clock set 2031-01-01T00:00:00Z. Queue e04d defaults; jobs 48–53 queued, 54 delay 100 (scheduled).
```
lease 48 → ack 200; lease 49 → nack 200 (retry_wait); lease 50 → cancel 200
  heartbeat(t50) 409 state_conflict "state conflict"; ack(t50) 409; nack(t50) 409 "sql: no rows in result set"; retry 409
cancel 51 (queued, next in FIFO) → 200; lease → 52 (not 51)
GET /v1/queues/e04d/stats → cancelled 2, dead 0, leased 1, queued 1, retry_wait 1, scheduled 1, succeeded 1, throughput 1, oldest_queued_age 0
psql e04d by state      → cancelled 2, leased 1, queued 1, retry_wait 1, scheduled 1, succeeded 1          (equal)
GET /v1/namespaces/acme/stats (admin) → cancelled 2, leased 1, queued 45, retry_wait 2, scheduled 2, succeeded 1, throughput 1, oldest_queued_age 31536000
psql namespace 1        → cancelled 2, leased 1, queued 45, retry_wait 2, scheduled 2, succeeded 1; succeeded in last hour 1  (equal)
```
R37 confirmed (cancelled jobs did not run; holder verbs 409). R34 confirmed at rest for queue and namespace.
Side observation: the 2031 clock jump expired all 40 e04c leases (600 s) back to queued; namespace
`oldest_queued_age` = 31 536 000 s (exactly 2030-01-01 → 2031-01-01), i.e. measured from *creation* — but
e04c's jobs spent that year leased, and e04b jobs 4–6 were genuinely queued that year, so this cannot isolate
R40. Isolated probe next. The nack 409 message `sql: no rows in result set` is a driver-text leak already
recorded as E01 Q3 — not new.

## 2026-10-22T12:47:17Z #obs P6 — R39 expired-lease grace, R40 queued age (results/e04/p6-expiry.txt)

Queues e04e/e04f lease 30 s. Clock set 2031-06-01T00:00:00Z; jobs 55 (x), 56 (y) leased, expiry 00:00:30.
```
clock advance 31 → now 00:00:31 (real 12:47:18.2)
heartbeat x extend 30 → 200 {"ok":true}
ack y result "late"  → 200 {"ok":true}            (real 12:47:18.29; both before the next sweep tick)
after 2 s: 55 | leased | attempts 0 | lease_expires_at 00:01:01 ;  56 | succeeded | attempts 1
events 55: created 00:00:00 / leased 00:00:00 / heartbeat 00:00:31   (no lease_expired, no attempt)
```
R39 prediction confirmed: a lease whose `lease_expires_at` has passed is still honoured until the Expire sweep
runs; a heartbeat *resurrects* it (no expired attempt is ever recorded) and an ack completes it. #surprise —
characterize (control with sleep before heartbeat next).

R40: clock set 2031-07-01T00:00:00Z; job 57 enqueued and leased; advance 600; after 2 s:
```
57 | queued | attempts 1 | created_at 00:00:00 | run_at 00:00:00 | event_at 00:10:00
GET /v1/queues/e04f/stats → queued 1, oldest_queued_age 600
```
R40 prediction confirmed: the only queued job has been queued for 0 s (since the 00:10:00 expiry sweep; its
lease ended at 00:00:30) but stats report 600 s — the leased time is counted as queued waiting. #surprise —
judgment and sibling next.

## 2026-10-22T12:47:47Z #obs P7 — R39 control and repeat (results/e04/p7-grace-control.txt)

```
(55 had been expired→queued by the 2031-07-01 clock set; lease → 55 attempt 3... see below)
leased 55; advance 31; sleep 1.5 (one sweep tick); heartbeat → 409 state_conflict; 55 | queued | attempts 2 | lease expired
repeat ×5 (enqueue, lease, advance 31, heartbeat immediately): 5/5 → 200 {"ok":true}  (jobs 55,58,59,60,61)
```
Control confirms the variable is *whether the 1-s sweep has run*, not the token or the extension: identical
requests give 409 after a tick and 200 before it. → #bug [E04-F04](../../defects/E04-F04-expired-lease-honoured-until-sweep.md).

## 2026-10-22T12:48:00Z #obs P8 — R40 sibling and control (results/e04/p8-age-sibling.txt)

Sibling prediction (formed after P6 from reading `internal/api/jobs.go:126`; **not written down before this probe**, so
not counted as a stated hypothesis): worker deletion releases a lease without moving `run_at`, so age counts the leased
time too; control: an untouched queued job reports its true wait.
```
clock 2031-08-01T00:00:00Z; job 64 (e04g lease 3600) leased by worker 5; advance 100; DELETE /v1/workers/5 → 200
64 | queued | attempts 0 | created 00:00:00 | run_at 00:00:00 | event_at 00:01:40
e04g stats → queued 1, oldest_queued_age 100          (queued for 0 s)
control e04h: enqueue 65, advance 100 → oldest_queued_age 100   (queued for 100 s: correct)
```
Sibling confirmed. → #bug [E04-F05](../../defects/E04-F05-queued-age-counts-leased-time.md) (oracle ambiguity stated there).

## 2026-10-22T12:48:31Z #obs P9 — R33 FIFO (results/e04/p9-fifo.txt)

```
clock 2031-09-01T00:00:00Z; batch [b1,b2,b3] → ids 66,67,68; three single leases → 66, 67, 68   (batch order ✔)
enqueue A-first → 69 (created 2031-09-01 00:00:00)
clock advance -60; enqueue B-second → 70 (created 2031-08-31 23:59:00)
lease → 70   (B leased before A, though A was created first)
```
R33 holds for batch order and same-instant ties (id tiebreak). The app-clock step-back inverts FIFO, as read
from `LeaseOrder = priority DESC, created_at, id`. #surprise — judged **not a defect in this lab**: it needs the
documented lab clock to move backwards, which the README warns against mixing; the production analogue (a
wall-clock step or multi-instance skew) is a risk hypothesis (H-E04-08), not demonstrated. SQS FIFO orders by
sequence number per message group, which would not invert; that informs the risk, not this product's
requirement.

## 2026-10-22T12:48:49Z #obs P10 — R35b bounded retry delay, default backoff (results/e04/p10-backoff-walk.txt; tools/walk-backoff.sh)

Queue e04j defaults (exponential base 5, no cap), max_attempts 21; clock walked to each run_at:
```
attempt 1 wait 5 s; 2: 10; 3: 20; … 10: 2560 s; 15: 81920 s (0.95 d); 18: 7.59 d; 19: 15.17 d; 20: 30.34 d (run_at 2031-11-30T16:21:15Z)
attempt 21 → dead (attempts 21 = max ✔ R32)
```
R42 confirmed (exact base·2^(n−1) doubling). "A retry always comes within a bounded delay" holds only with the
SPEC bound of 2 147 483 647 s (68 years). This is the spec silence lesson 03 disclosed as A1 (default cap);
my walk re-observes it. Not counted as a discovery; routed to the owner as open decision Q-E04-1. Comparable
systems (as recalled, unverified in this run): SQS caps a visibility timeout at 12 h; pg-boss offers a
maximum retry delay option.

## 2026-10-22T12:49:20Z #read panic scan

`grep '\.(int64|string…)'` over internal/api, schedule, webhook: unguarded assertions only on non-null
columns (`queues.max_attempts`, `queues.id`, outbox url/secret) plus `schedules.go:158 v["kind"].(string)`
(nullable column; schedules are #offcharter). No panic-to-dropped-connection path found in charter routes.

## 2026-10-22T12:49:30Z #hypothesis siblings (before probe)

- H-E04-S1 (sibling of E04-F04): nack after expiry but before the sweep → 200 and records a `failed` attempt
  instead of `expired`. Wrong if 409.
- H-E04-S2 (sibling of E04-F01): HEAD/OPTIONS on an API route and a wrong method on `/v1/jobs/{id}` are
  also plain text. Wrong if JSON envelope.

## 2026-10-22T12:49:59Z #obs P11 — siblings (results/e04/p11-siblings.txt)

```
H-E04-S1: clock 2031-12-01; job 58 leased (attempt 2); advance 31; nack immediately → 200
  attempts 58: n1 expired "lease expired" (earlier) ; n2 failed "late nack"     → confirmed: the late nack is recorded as a worker failure, not an expiry
H-E04-S2: OPTIONS /v1/queues → 405 text/plain (Allow: GET, HEAD, POST); PATCH /v1/jobs/66 → 405 text/plain (Allow: DELETE, GET, HEAD)
          POST /v1/jobs/66/events → 404 application/json {"error":{"code":"not_found","message":"unknown operation"}}
```
S1 confirmed (added to E04-F04). S2 confirmed for the mux-level cases; the `{verb}` wildcard instead turns a
wrong-method request on `/events` into a JSON 404 rather than 405 — inconsistent class but in the envelope;
recorded in E04-F01's sibling section, not counted separately.

## 2026-10-22T12:50:18Z #obs P12 — R33b priority then FIFO (results/e04/p12-priority.txt)

Batch [5a, 0a(p0), 9a(p9), 5b, 9b(p9)] → ids 73–77; five single leases → `9a 9b 5a 5b 0a`. Prediction confirmed.

## 2026-10-22T12:50:33Z #obs stop

Final snapshot `results/e04/final-snapshot.txt`; Compose logs `results/e04/all-logs-at-stop.txt` (1006 lines; no
`background task`/`webhook task`/`panic` lines). Lab clock reset (`{"now":"2026-10-22T12:50:33.739976669Z"}`), then
`dc stop` — only `queuelet-e04-*` containers stopped; volume kept for inspection. No other Compose project touched.

## 2026-10-22T12:51Z #question / testability

- TR-E04-1: no console shows *when* the Expire sweep last ran or how far behind it is. Question: how long can an expired
  lease stay honoured (E04-F04)? Missing signal: last-sweep timestamp/lag. Minimal hook: `/metrics` line
  `queuelet_core_tick_seconds` (time of last completed core loop) and `queuelet_expired_unswept` (leased rows with
  `lease_expires_at <= now`). Validation probe: advance clock past a lease, read both before and after one tick.
- TR-E04-2: no way to read `stats` and a direct count in one snapshot (H-E04-09). Minimal hook: return `as_of` (app time)
  and run the three statements in one REPEATABLE READ transaction. Validation: concurrent lease/ack load while comparing sums.
- Open owner decisions: Q-E04-1 (default backoff cap, lesson-03 A1), Q-E04-2 (meaning of "readiness"), Q-E04-3 (lease
  grace), Q-E04-4 (does "completed" include dead/cancelled in `throughput_last_hour`).

## 2026-10-22T12:51Z #offcharter

- `schedules.go:158` unguarded `v["kind"].(string)` on a nullable column (H-E04-11).
- Ack is three separate statements; a DB fault after the state UPDATE would be reported as 409 (H-E04-10).

## Surprise ledger (every #surprise judged)

| Surprise | Outcome |
|---|---|
| Mux plain-text 404/405 (P1) | E04-F01 |
| Attempt numbers restart after retry (P2) | E04-F02 |
| Dead job gets a future `run_at` (P2) | not a bug: terminal `run_at` has no documented meaning; Retry overwrites it |
| Negative delays precede failure / wrap forward (P3) | E04-F03 |
| Expired lease honoured until sweep (P6/P7/P11) | E04-F04 |
| Queued age counts leased time (P6/P8) | E04-F05 |
| FIFO inverted by app-clock step back (P9) | not a defect here: needs the documented lab clock to move backward; risk H-E04-08 |
| Default backoff 30 d at attempt 20 (P10) | spec-compliant (68-year bound); disclosed A1; Q-E04-1 |
| nack 409 message `sql: no rows in result set` (P5) | known E01 Q3 message leak; not new |
