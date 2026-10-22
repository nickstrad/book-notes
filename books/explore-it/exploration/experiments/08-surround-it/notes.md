# E08 surround-it — notes

Export: `/tmp/explore-it-experiment-08` (checkpoint `explore-it/lesson-07`, no git).
Stack: `COMPOSE_PROJECT_NAME=queuelet-e08`, image tag `e08`, API :18290, hooksink :18291, PG :15523
(`tools/env.sh`). Client helper `tools/q.py` (copied from `07-state-it/tools/q.py`, defaults and
compose path adapted; inherited copy unchanged). Other projects seen running and left untouched:
`queuelet-e01`, `tsdb`.

All times UTC. Paths in `#read` are relative to `books/explore-it/project/`.

## Setup

- 19:18:14 run start (`run-start.txt`); 19:18:20 first command (read STATE.md).
- 19:18:47 `docker compose -f deploy/compose.yml up -d --build` (env from tools/env.sh) → 19:19:33 `HEALTHY`
  (`project/results/e08/up.log`).
- 19:20:55 `bash lab/reset.sh` (env-driven; no absolute paths, used in place) → 19:21:06 seeded:
  `Seeded 300 jobs, 4 queues, 2 schedules`. Queues: 1 acme/emails, 2 acme/reports (webhook
  `http://hooksink:8201/hook`), 3 acme/imports, 4 globex/emails. `/health` 200 ok; hooksink mode `ok`.
  First `curl: (56) Recv failure` in wait.sh is the poll before the API listened (expected).

## Reading (19:18–19:21)

- `#read` `internal/api/api.go:149-170` — auth middleware: every path except `/health` and `/metrics`
  hashes the bearer key and looks up `api_keys(namespace_id, admin)`; `sql.ErrNoRows` → 401, any other
  error → 503 `auth_unavailable`. Identity is a namespace id plus admin flag; nothing per-worker.
- `#read` `migrations/001.sql:16` — `admin-key` is a row with `namespace_id=1` (acme) and `admin=true`:
  the administrative key is also a full acme tenant key. `/admin/clock` (observe.go:15) needs admin.
- `#read` `internal/api/jobs.go:200-213` `ownJob` — every `/v1/jobs/{id}` route checks
  `id AND namespace_id`; events (observe.go:12) filters by `namespace_id`. Queue routes look up
  `namespace_id AND name` (api.go:236). Schedules filter by namespace (schedules.go). Worker lease checks
  `workers.namespace_id` (jobs.go:174). Namespace stats: 403 unless same namespace or admin (queues.go:194).
- `#read` `internal/api/api.go:205-245` — two goroutines: shipper (1 s ticker, `webhook.Ship`) and core
  loop (Promote, Expire, Fire, Retain). Errors are logged and set `shipFailed`/`coreFailed`, loop continues.
  `/health` = DB ping + both flags (api.go:131-139).
- `#read` `cmd/api/main.go:19-22` — `store.Open` pings the DB and `log.Fatal`s on failure.
  `deploy/compose.yml`: `api` has **no restart policy** (workers have `restart: on-failure`);
  `depends_on: service_healthy` only applies at `up`.
- `#read` `internal/webhook/ship.go:17-60` — one serial pass: `SELECT … WHERE delivered_at IS NULL AND
  next_at <= now ORDER BY id LIMIT 20`; 2 s client timeout per POST; failure → `next_at = now+5s` where
  `now` is the tick time captured **before** the pass (api.go:212). Signs `json.Marshal(r["body"])`; store
  returns jsonb `body` as `json.RawMessage` (store.go:44-49) so the signed bytes are the sent bytes.
  The outbox is global (all namespaces in one ORDER BY id).
- `#read` `internal/jobs/ack.go:13-26` — ack is three autocommit statements: UPDATE job → INSERT attempt
  → `Notify` INSERT outbox (jobs.go:91-104). Any error after the UPDATE is returned as 409
  `state_conflict` by jobs.go:107-110 although the job already succeeded. Not deterministically
  injectable from outside (window is between statements) → testability request TR-E08-1.
- `#read` `cmd/hooksink/main.go` — modes ok/500/slow(3 s sleep)/timeout(10 s sleep); records
  `at, body, raw_body, signature_valid, status`; **does not record the received signature header**
  (E01 TR1 still open). Verifies with its own `WEBHOOK_SECRET=report-secret`.
- `#read` `internal/schedule/fire.go:13-23,64-84` — next fire = robfig `cron/v3` `Next(now.In(loc))`;
  stored `zone_offset` is not used by `Next` (parameter ignored). Job `run_at=due`; next computed from
  `now`.
- `#read` `internal/clock/clock.go` — test clock (`POST /admin/clock` set/advance/reset) controls API
  semantic time only; Postgres and the workers and hooksink use real time. No SQL uses `now()`.
- `#read` SPEC.md:162 "Schedule times follow the named zone's local civil time, including offset
  changes." Silent on non-existent (spring-forward) and repeated (fall-back) local times.
- `#read` SPEC.md:186-189 loop recovers after DB restart; webhook latency must not block core progress;
  503 for unavailable auth store. SPEC.md:192-205 snapshot/at-least-once/2 s deadline, malformed rows
  "rescheduled independently, so it does not stop later rows". No statement on cross-tenant delivery
  isolation.

## Context diagram (drawn before probes; ★ = boundary)

```text
              producers/operators (curl, qctl)            other tenant (globex-key)
                         │ ★ bearer key → namespace             │ ★ same API, key → ns 2
                         ▼                                       ▼
 admin-key ★──────▶  queuelet API (Go, :8200) ◀──── ★ worker containers (acme-key; worker_id, lease_token)
 (ns 1 + admin)      │ auth mw │ handlers │ test clock ★ (semantic time, /admin/clock)
                     │ core loop (promote/expire/fire/retain) │ shipper loop
                     ▼ ★ SQL (lib/pq, 16 conns)               │ ★ outbound HTTP, HMAC, 2 s deadline,
                  PostgreSQL 17 (state, api_keys, outbox)     ▼   tenant-chosen URL, shared serial pass
                                                         webhook partner (hooksink :8201, modes)
 docker/compose ★ process supervisor: workers restart on-failure; api: none; postgres: none
 IANA tz database ★ (Go time.LoadLocation in the API image) — cron in local civil time
```

## Hypotheses (stated before the probes)

| Id | Boundary | Prediction | Source | I would be wrong if |
|---|---|---|---|---|
| H-E08-01 | tenant | globex-key on an acme job id gets 404 for GET, attempts, events, cancel, retry, delete, ack, nack, heartbeat; acme job unchanged | jobs.go:200 ownJob; observe.go:12 | any 2xx/409, or acme job state/attempts change (regression of D21 — not a new find if it holds) |
| H-E08-02 | tenant | globex lease on its own `emails` with an acme worker_id → 404 worker_not_found; acme-key on `/v1/namespaces/globex/stats` → 403; globex schedule ids/worker delete of acme ids → 404 | jobs.go:174; queues.go:194; schedules.go | any 2xx or mutation |
| H-E08-03 | admin | admin-key reads acme job 200 (it *is* acme), gets 404 on a globex job, 200 on globex namespace stats | 001.sql:16 | admin-key 404 on acme job or 200 on globex job |
| H-E08-04 | Postgres restart | `docker compose restart postgres`: health 503 and authenticated calls 503 `auth_unavailable` for a few seconds, then 200; a job delayed to fire during the outage is promoted afterwards; workers resume | api.go:149-170, 205-245; E02 H15 | loop does not resume (job stays scheduled > 10 s after DB healthy) or 401 during outage (regression) |
| H-E08-05 | Postgres + supervisor | stop postgres, restart api → api exits (Fatal on ping); starting postgres again does **not** bring api back (no restart policy) — service stays down until manual start | main.go:19-22; compose.yml api has no `restart:` | api serves again without manual action within 30 s of Postgres healthy |
| H-E08-06 | partner 500 | mode 500: a reports job's outbox row is attempted every ~5 s, `delivered_at` null, hooksink records status 500; after mode ok it is delivered once | ship.go:40-48 | delivered_at set during 500 (D04 regression) |
| H-E08-07 | partner slow | mode slow (3 s): queuelet records failure `context deadline exceeded`/timeout and retries; hooksink records the request with status 200 (duplicate side-effects); a delayed job still promotes on time (core not blocked) | ship.go:32; api.go:205 | marked delivered in slow mode, or promotion delayed by > 2 s |
| H-E08-08 | partner timeout × tenants | with ≥ 20 pending globex rows to a timing-out partner, an acme row with a higher outbox id to an instantly-200 destination is **never attempted** (attempts 0, delivered_at null) for ≥ 60 s; after mode ok it is delivered | ship.go:17 `ORDER BY id LIMIT 20`; next_at=tick+5 s < pass length 40 s | acme row attempted/delivered within 60 s while timeout persists |
| H-E08-09 | data journey | one reports job's webhook: outbox `body` jsonb = hooksink `body`; HMAC-SHA256("report-secret", raw_body) computed by hand equals the signature header captured by a listener; hooksink `signature_valid=true` | ship.go:22-30; store.go:44 | hand HMAC ≠ received header, or body semantically differs from outbox row |
| H-E08-10 | clock/DST gap | clock at 2027-03-14T05:00Z; schedule `30 2 * * *` America/New_York: robfig skips the non-existent 02:30 → `next_run_at` = 2027-03-15T06:30Z (no run on 3/14) | fire.go:20 robfig/cron v3 | next_run_at on 2027-03-14 (07:00Z or 06:30Z) |
| H-E08-11 | clock/DST gap | hourly `0 * * * *` NY from 2027-03-14T06:30Z (01:30 EST): next fires at 07:00Z (03:00 EDT), i.e. no lost real hour | fire.go:20 | next fire ≠ 2027-03-14T07:00Z |
| H-E08-12 | clock/DST overlap | `30 1 * * *` NY across 2027-11-07 fall-back: fires at 05:30Z (01:30 EDT) and again at 06:30Z (01:30 EST) — twice on one civil date (low confidence) | fire.go:20 | only one fire on 2027-11-07 |
| H-E08-13 | clock/offset | seed `daily-report` `0 9 * * *` NY advanced across 2027-03-14: fires 2027-03-13T14:00Z then 2027-03-14T13:00Z (offset change honoured; D23 regression) | fire.go:20; SPEC.md:162 | 14:00Z on 3/14 |

Oracles named per hypothesis in the reports if violated. H-E08-01/03/04/06/13 are regression checks of earlier fixes; holding them is not a new discovery.

## Probe log

### P1 tenant/admin boundary — 19:22:08–19:22:16 (H-E08-01..03)

- `#obs` 19:22:08 `python3 tools/p1_tenant.py` (raw: `project/results/e08/p1.txt`). Fixture: acme queue
  `e08p1`, worker, jobs 301 leased, 302 dead, 303 queued; globex job created in globex `emails`.
  ```text
  GET/attempts 301,302,303 (globex-key) → 404 ×6
  events 301/302/303 (globex-key) → 200 {"events": []}
  cancel 303 / retry 302 / delete 302 / ack 301 (right token) / nack 301 / heartbeat 301 → 404 ×6
  globex lease own emails w/ acme worker → 404 worker_not_found "unknown worker"
  globex DELETE acme worker → 404 {"code":"worker_not_found","message":"state conflict"}
  globex schedule GET/PATCH/run/DELETE 1 → 404 ×4
  acme-key /v1/namespaces/globex/stats → 403 forbidden "namespace mismatch"; globex→acme 403
  admin GET acme job 200; admin GET globex job 404; admin globex ns stats 200
  no key → 401; acme-key POST /admin/clock → 403 "admin key required"
  acme jobs unchanged: True {301:'leased', 302:'dead', 303:'queued'}
  ```
- `#obs` 19:22:16 `curl …/v1/jobs/999999/events` with acme-key and globex-key → `{"events":[]}` 200 both;
  `…/302/events` acme-key → full history; globex-key → `[]`.
- Judgement: H-E08-01 **confirmed** (regression of D21 holds), H-E08-02 **confirmed**, H-E08-03
  **confirmed** (admin key = acme tenant + stats-anywhere, per 001.sql:16).
- `#surprise` (minor) events for a non-owned or never-existing job id → 200 `[]`, not 404.
  **Not a bug, reason:** observe.go:10 deliberately scopes by retained event ownership so history
  survives job deletion (SPEC.md:85 "durable job events … independent lifetime"); the answer is the
  same `[]` for a foreign id and a nonexistent id, so no existence or content crosses the tenant
  boundary. Noted as a model cell, not counted as a defect.
- `#surprise` (minor) DELETE of a foreign worker: message `"state conflict"` under code
  `worker_not_found`. **Not a bug, reason:** status/code are correct (404 worker_not_found); the message
  is the internal `jobs.Changed` text — cosmetic, no information crosses.

### P2 Postgres restart — 19:22:38–19:22:55 (H-E08-04)

- `#obs` `python3 tools/p2_pg_restart.py` (raw `project/results/e08/p2.txt`): job 305 enqueued in
  `e08p1` with `delay_seconds=6` (run_at 19:22:44.135); poller every 0.5 s on `/health` and GET job.
  ```text
  19:22:39Z restart postgres → returned 0 at 19:22:40Z
  19:22:38.1 200 (200,'scheduled')
  19:22:39.7 503 (503, auth_unavailable "authentication store unavailable")
  19:22:40.7 200 (200,'scheduled')
  19:22:44.8 200 (200,'queued')        ← promoted ≤0.7 s after run_at
  final queued 19:22:55Z
  ```
  api log: `event console … connection reset by peer`, `webhook task: dial tcp … connection refused`,
  `background task: … connection refused`, `lookup postgres … server misbehaving`; worker-a logged
  `HTTP 503 auth_unavailable` and kept polling.
- Judgement: H-E08-04 **confirmed** (regressions D22/U04 hold; outage window ≈1 s here vs ≈4–6 s in E02).

### P3 API (re)start while Postgres is down — 19:23:02–19:23:50 (H-E08-05)

- `#obs` 19:23:02 first attempt failed before acting: zsh does not word-split `$C="docker compose -f …"`
  (`no such file or directory: docker compose -f deploy/compose.yml`); nothing was stopped (api
  `exit=0 restarts=0 policy=no`, health 200 at 19:23:05). Dead end kept; rewritten as a bash script.
- `#obs` 19:23:13 `tools/p3_api_restart_db_down.sh` (raw `project/results/e08/p3.txt`):
  ```text
  19:23:13.6 stop postgres → Stopped
  19:23:14.2 restart api → Started
  19:23:18.2 ps: api exited  Exited (1) 3 seconds ago; postgres exited (0)
  api exit=1 restarts=0 policy=no
  api log: 19:23:15.208 "dial tcp: lookup postgres on 127.0.0.11:53: server misbehaving"   (log.Fatal)
  19:23:18.8 start postgres → Started
  19:23:50.0 last health=000;  ps: api exited (1) 35 s ago; postgres running (healthy) 31 s
  worker-a: Post "http://api:8200/v1/queues/emails/lease": dial tcp: lookup api … (polling forever)
  ```
- Judgement: H-E08-05 **confirmed** (from code reading; prediction of a defect confirmed).
- `#surprise` → `#bug` [E08-F01](../../defects/E08-F01-api-does-not-recover-from-start-without-database.md).

### P3b isolate restart variables — 19:24:03–19:24:42

- Prediction (formed before running and stated in the script header `p3b_isolate.sh`, but written
  into this log only at 19:25 after the run — counted as post-hoc, not scored): control `restart api` with DB up → health 200 within 5 s;
  sibling SIGKILL api with DB up → stays down (policy `no`); SIGKILL postgres → API stays up but 503
  until postgres is started by hand (no policy on postgres either); recovery after manual start.
- `#obs` `tools/p3b_isolate.sh` (raw `project/results/e08/p3b.txt`):
  ```text
  19:24:03.2 recover: start api → health 200 after 2s
  19:24:05.5 control: restart api with postgres up → health 200 after 2s
  19:24:07.4 SIGKILL api → health still 000 after 15 s; api state=exited exit=137 restarts=0
  19:24:23.1 start api → 200 after 2 s
  19:24:25.2 SIGKILL postgres → health still 503 after 15 s; postgres exit=137 restarts=0 policy=no;
             GET /v1/queues → {"error":{"code":"auth_unavailable",…}}
  19:24:40.9 start postgres → health 200 after 2s (API process survived the DB crash)
  ```
- Variables isolated: the trigger is *API process start while the DB is unreachable* (control with DB
  up restarts fine); the persistence is the missing restart policy (an API crash with DB up also stays
  down). A DB crash alone is survived by the running API (it only needs the DB to come back, which also
  has no policy). Sibling recorded in E08-F01.

### P4 data journey + hand signature check — prediction written 19:25:45 (H-E08-09, H-E08-14)

- Instrument: `tools/capture.py` listener bound to the e08 network gateway `172.28.0.1:18292` (host
  process, started 19:25:39, pid 2628168, output `project/results/e08/capture.jsonl`) records the
  received headers and raw bytes — stand-in for E01 TR1 because hooksink does not record the header.
  Queues: acme `e08cap` → `http://172.28.0.1:18292/cap` secret `e08-secret`; acme `e08hs` →
  hooksink, secret `report-secret`. Datum acked as raw JSON text:
  `{"trace":"E08-ü- -<&>","n":12345678901234567890,"f":1.10,"z":[],"nested":{"b":1,"a":2}}`.
- `#hypothesis` H-E08-14 (source: store.go:44 jsonb→RawMessage; Go `json.Marshal(RawMessage)` compacts
  with HTML escaping; jsonb key ordering): the sent bytes differ textually from the ack input — `<&>`
  arrive as `<&>`, U+2028 as ` `, keys reordered by jsonb (shorter first) — while
  remaining semantically equal to the outbox row; `n` stays exact, `1.10` keeps its text.
  Wrong if: semantic inequality, `n` rounded, or the signature does not match the sent bytes.
- `#obs` 19:25:48 `python3 tools/p4_journey.py project/results/e08/capture.jsonl` (raw `project/results/e08/p4.txt`):
  ```text
  e08cap ack 200 job 306; e08hs ack 200 job 307
  outbox 70 (job 306): attempts 1, status 200, delivered; body::text {"state": "succeeded", "job_id": 306,
     "result": {"f": 1.10, "n": 12345678901234567890, "z": [], "trace": "E08-ü-<U+2028>-<&>", "nested": {"a": 2, "b": 1}}, "last_error": ""}
  capture raw bytes: {"state":"succeeded","job_id":306,"result":{"f":1.10,"n":12345678901234567890,"z":[],
     "trace":"E08-ü- -<&>","nested":{"a":2,"b":1}},"last_error":""}
  received X-Queuelet-Signature: c3678dd229ee1d7f57c18cfa7f14fc35502268111918336eddc80bb50a76dcc6
  hand HMAC-SHA256("e08-secret", raw)  c3678dd229ee1d7f57c18cfa7f14fc35502268111918336eddc80bb50a76dcc6  equal: True
  same bytes with "report-secret": False
  hooksink (job 307): status 200, signature_valid True, raw_body same shape as above
  semantic equal hooksink body == outbox body: True; capture raw == outbox: True; n exact; "f":1.10 kept
  ```
  (The script's `job.result (API)` line printed `"f": 1.1` — that is Python's `json.loads`; a raw
  `curl /v1/jobs/306` at 19:25:57 returns `"f":1.10,…"trace":"E08-ü- -<&>"`.)
- Judgement: H-E08-09 **confirmed** (signature verified by hand over the received bytes; D24 holds).
  H-E08-14 **confirmed**: transformations along the path are jsonb key reordering (`state, job_id, result,
  last_error`; `nested` a before b), whitespace removal and Go HTML/line-separator escaping; all
  semantics-preserving, and the signature covers the transformed bytes. Not a bug: receivers must
  verify the raw bytes (as SPEC.md:177 says "over the body"), never a re-serialisation.

### P5 partner modes — 19:26:35–19:28:35 (H-E08-06, -07, -08)

- `#obs` 19:26:35 `python3 tools/p5_partner_modes.py 500` (raw `project/results/e08/p5-500.txt`):
  ```text
  mode 500; acked job 308 (acme e08hs → hooksink)
  19:26:39 outbox 72|1|500|f|HTTP 500|19:26:36.265
  19:26:43 outbox 72|2|500|f|HTTP 500|19:26:42.265
  19:26:48 outbox 72|3|500|f|HTTP 500|19:26:48.265
  hooksink: (500,'19:26:36.268'), (500,'19:26:42.267'), (500,'19:26:48.289')
  19:26:52 mode ok → 19:26:59 outbox 72|4|200|t||19:26:53.267; records + (200,'19:26:53.288')
  ```
  H-E08-06 **confirmed** (D04 holds; retry spacing 6 s = 5 s + 1 s tick).
- `#obs` 19:27:03 `python3 tools/p5_partner_modes.py slow` (raw `p5-slow.txt`):
  ```text
  mode slow; acked job 309; delayed job 310 run_at 19:27:07.478
  19:27:08 delayed job -> queued after 4.8 s (enqueue+4 s delay; on time)
  19:27:14 outbox 73|2|0|f|Post "http://hooksink:8201/hook": context deadline exceeded |19:27:10.265
  hooksink: (200,'19:27:04.319'), (200,'19:27:10.269')
  19:27:14 mode ok → 19:27:22 outbox 73|3|200|t; partner side-effects: 3 × 200 for one notification
  ```
  H-E08-07 **confirmed**. Duplicates are the documented at-least-once semantics (SPEC.md:200-204,
  E02-F03) — not a new bug.
- `#obs` 19:27:28 `python3 tools/p5_partner_modes.py timeout 22 60` (raw `p5-timeout22.txt`). Fixture:
  globex queue `e08slow` → hooksink (mode timeout, 10 s sleep), 22 terminal globex jobs 311..332 =
  outbox 74..95; acme queue `e08fast` → `http://hooksink:8201/admin/records` (answers 200 instantly in
  every mode), acme job 333 = outbox 96. Delayed acme job 334 as core check.
  ```text
  19:27:39 acme 96|0|-1|f|| | globex attempts 2,2,2,0,0,…,0          | delayed queued
  19:27:50 acme 96|0|-1|f|| | globex attempts 2,2,2,1,1,1,1,1,0,…     | delayed queued
  19:28:01 acme 96|0|-1|f|| | … 1 ×10, 0 ×9
  19:28:12 acme 96|0|-1|f|| | … 1 ×16, 0 ×3
  19:28:22 acme 96|0|-1|f|| | globex attempts 3,3,3,2,1,…,1,0,0      (pass restarted at row 74)
  19:28:33 acme 96|0|-1|f|| | globex attempts 3,3,3,2,2,2,2,2,2,1,…,1,0,0
  19:28:33 mode ok → 19:28:35 acme delivered 1.5 s after mode ok: 96|1|200|t||19:28:34.608; globex pending 0
  ```
- Judgement: H-E08-08 **confirmed** (prediction of a defect from code reading). Over 64 s the acme row
  and the last two globex rows (94, 95) were never attempted: each pass takes the same 20 lowest ids,
  costs 20 × 2 s = 40 s, and their `next_at` (= pass start + 5 s) is already due when the next pass
  starts. Core promotion was not affected (delayed job queued).
- `#surprise` → `#bug` [E08-F02](../../defects/E08-F02-slow-destination-starves-all-later-notifications.md).

### P5b isolate the starvation — predictions written 19:29:10, before running

- `#hypothesis` H-E08-15 (control: slowness, source ship.go:41 `next_at=now+5s`): same 22 globex rows but
  hooksink mode **500** (fast failure) → acme row delivered within 3 s of its ack. Wrong if > 5 s.
- `#hypothesis` H-E08-16 (control: count, source ship.go:17 LIMIT 20): only **5** globex rows in timeout
  → acme row is reached in the first pass and delivered after ≈ 5 × 2 s = 10–12 s (delayed, not
  starved). Wrong if never attempted within 30 s, or delivered in < 8 s.
- `#hypothesis` H-E08-17 (sibling, tenant-only cause, no partner misbehaviour): 20 globex rows whose
  queue URL points at an unused address on the e08 network (`http://172.28.0.250/x`, hooksink in ok)
  → acme row not attempted for 40 s. Wrong if acme delivered within 40 s.
- `#obs` 19:29:21 `python3 tools/p5b_isolate.py 500 22 http://hooksink:8201/hook 15` (raw `p5b-500x22.txt`):
  `acme job 357 outbox 119 … after 1.4 s acme: 119|1|200|t|`; globex 22 × attempts 1 `HTTP 500`.
  H-E08-15 **confirmed**: slowness, not failure, is the variable.
- `#obs` 19:29:33 drained (mode ok, pending 0), then `p5b_isolate.py timeout 5 … 30` (raw `p5b-timeoutx5.txt`):
  `after 11.1 s acme: 125|1|200|t|`; globex `1,1,1,1,1 err=… context deadline exceeded`.
  H-E08-16 **confirmed**: delay = 2 s × earlier slow rows; starvation needs ≥ 20.
- `#obs` 19:29:50 drained, then `p5b_isolate.py ok 20 http://172.28.0.250/x 45` (raw `p5b-blackholex20.txt`):
  hooksink healthy (mode ok); `after 45.6 s acme: 146|0|-1|f|`; globex attempts
  `3,3,3,3,2,2,1,…,1 err=Post "http://172.28.0.250/x": dial tcp 172.28.0.250:80: conn…`.
  H-E08-17 **confirmed**: one tenant's own configuration (a well-formed URL that accepts the create/PATCH
  validation) is enough; no partner needs to misbehave.
- `#hypothesis` H-E08-18 (19:30:45, source SPEC.md:196-198 snapshots + no attempt cap; ship.go): the
  stall is permanent and no API action available to either tenant ends it: globex PATCH of its queue
  webhook to `""` and DELETE of its terminal jobs leave the 20 pending rows retrying; acme row 146 stays
  unattempted for a further 45 s. Wrong if acme 146 is attempted after those actions.
- `#obs` 19:30:57 `python3 tools/p5c_recovery.py e08b97390g 146` (raw `p5c-recovery.txt`):
  ```text
  errors: Post "http://172.28.0.250/x": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
          Post "http://172.28.0.250/x": dial tcp 172.28.0.250:80: connect: no route to host
  globex PATCH webhook url "" → 200; DELETE its 20 terminal jobs → {204}; DELETE queue → 204
  19:31:03 … 19:31:43 (every ~5 s): acme 146|0|f | globex pending 20, max attempts 4→5 | health 200
  ```
  H-E08-18 **confirmed**: no tenant-level API action ends the stall; `/health` stays 200 (shipper
  `shipFailed` only reflects DB errors, api.go:214), so nothing signals it. Rows have no attempt cap
  (SPEC.md:204), so it lasts until an operator edits the database. Added to E08-F02.
- `#obs` 19:32 lab poisoned by design of the probe; `bash lab/reset.sh` next (truncates outbox, restarts
  api → test clock back to real time).

### P6 test clock across DST — plan and refined predictions appended ≈19:32:50 in the same command, immediately before the run (run 19:32:55)

Fixture: acme queue `e08dst` (no worker consumes it), schedules all `America/New_York`, kinds name the
schedule: S1 `30 2 * * *` (dst-0230), S2 `0 * * * *` (dst-hourly), S3 `0 2 * * *` (dst-0200), S6
`0 9 * * *` (dst-0900), created with the clock set to 2027-03-13T12:00:00Z; later S4 `30 1 * * *`
(fb-0130) and S7 `0 1 * * *` (fb-0100) created at 2027-11-07T04:00:00Z. Clock is `set` (frozen) at each
step, then 2.5 s real time for the 1 s loop. Hand derivation of robfig/cron v3 `Next` (hour loop adds
1 h in local time; minute loop wraps):
- H-E08-10: S1 next_run_at at creation = **2027-03-15T06:30Z**; S3 = **2027-03-15T06:00Z** (no 3/14 run).
- H-E08-11: after the 2027-03-14T06:00:30Z step, S2 next = **07:00Z** (03:00 EDT); fires at 07:00:30 step.
- H-E08-12: S4 fires at both the 05:30:30Z and 06:30:30Z steps on 2027-11-07 (two jobs, run_at 05:30Z
  and 06:30Z); S7 likewise at 05:00Z and 06:00Z.
- H-E08-13: S6 fires at the 2027-03-13T14:00:30Z step (run_at 14:00Z), next = 2027-03-14T13:00Z, fires
  at the 13:00:30Z step.
- Side effect expected and not judged: clock jumps promote seeded delayed/retry jobs; S2 catches up
  once per jump (fire.go:66 "At most one due execution can catch up").
- `#obs` 19:32:55–19:33:31 `python3 tools/p6_dst.py` (raw `project/results/e08/p6-dst.txt`; clock frozen by
  `set` at each step):
  ```text
  clock 2027-03-13T12:00:00Z
  created dst-0230 next_run_at 2027-03-15T06:30:00Z zone_offset -18000
  created dst-hourly next 2027-03-13T13:00Z; dst-0200 next 2027-03-15T06:00:00Z; dst-0900 next 2027-03-13T14:00Z
  [03-13T14:00:30Z] jobs: dst-hourly@03-13T13:00Z dst-0900@03-13T14:00Z; dst-0900 next=03-14T13:00Z
  [03-14T06:00:30Z] dst-hourly next=03-14T07:00Z (+ catch-up job dst-hourly@03-13T15:00Z)
  [03-14T07:00:30Z] + dst-hourly@03-14T07:00Z ; dst-0230/dst-0200 still next 03-15
  [03-14T13:00:30Z] + dst-hourly@03-14T08:00Z (catch-up) + dst-0900@03-14T13:00Z ; dst-0900 next=03-15T13:00Z
  [03-15T06:00:30Z] + dst-hourly@03-14T14:00Z + dst-0200@03-15T06:00Z ; dst-0200 next=03-16T06:00Z
  [03-15T06:30:30Z] + dst-0230@03-15T06:30Z
  → no dst-0200 / dst-0230 job for civil date 2027-03-14 at all
  clock 2027-11-07T04:00:00Z: catch-up jobs dst-0230@03-16T06:30Z dst-hourly@03-15T07:00Z dst-0200@03-16T06:00Z dst-0900@03-15T13:00Z (one each)
  created fb-0130 next 2027-11-07T05:30:00Z zone_offset -14400; fb-0100 next 05:00Z
  [11-07T05:00:30Z] + dst-hourly@11-07T05:00Z fb-0100@11-07T05:00Z ; fb-0100 next=11-07T06:00Z
  [11-07T05:30:30Z] + fb-0130@11-07T05:30Z ; fb-0130 next=11-07T06:30Z
  [11-07T06:00:30Z] + dst-hourly@11-07T06:00Z fb-0100@11-07T06:00Z ; fb-0100 next=11-08T06:00Z
  [11-07T06:30:30Z] + fb-0130@11-07T06:30Z ; fb-0130 next=11-08T06:30Z
  [11-07T07:00:30Z] + dst-hourly@11-07T07:00Z dst-0200@11-07T07:00Z
  ```
- Judgements: H-E08-10 **confirmed** (gap: 02:00 and 02:30 schedules skip 2027-03-14 entirely);
  H-E08-11 **confirmed** (hourly 01:00 EST → 03:00 EDT, no lost real hour); H-E08-12 **confirmed**
  (fall-back: `30 1` and `0 1` each fire twice on 2027-11-07, at EDT and EST instants); H-E08-13
  **confirmed** (09:00 local follows the offset change: 14:00Z then 13:00Z — D23 holds).
- `#surprise`-class oracle question (predicted, so not a surprise in the strict sense): skip and double-fire
  are robfig/cron defaults; SPEC.md:162 is silent; classic cron (Vixie/ISC/cronie) runs a fixed-time job
  skipped by spring-forward once after the change and does not re-run fixed-time jobs in the repeated
  hour. → [E08-F03](../../defects/E08-F03-dst-gap-skips-and-overlap-double-fires.md) (status `spec`).
- `#obs` jump catch-up jobs carry the old due `run_at` (e.g. `dst-0230@03-16T06:30Z` created at clock
  2027-11-07) — documented one-catch-up rule (fire.go:66), not judged.
- `#read` `zone_offset` is stored (-18000 / -14400, the offset at creation) but `schedule.Next` ignores its
  `offset` argument (fire.go:13). Dead data, not a behaviour fault; noted for the model.

### P7 worker identity boundary — prediction appended ≈19:34:10 in the same command, immediately before the run

- `#read` api/jobs.go:128-135 issues a worker `token`; nothing reads `workers.token` again (grep of
  internal/ and cmd/). Lease authorisation = namespace key + `worker_id` (jobs.go:173-178).
  SPEC.md:91 "returns a worker id and token" without saying what the token authorises.
- `#hypothesis` H-E08-19: with acme-key, leasing as the running container worker-a's `worker_id` (taken
  from the DB, never told to me) returns 200 and the job's `leased_by` is worker-a's id; no token is
  asked for. Wrong if 401/403/404 or leased_by differs.
- `#obs` 19:34:14 `python3 tools/p7_worker_identity.py` (raw `project/results/e08/p7.txt`):
  `worker-a id 2 lease as worker-a → 200 [(339, 1)]`; `job leased_by: 2 state: leased`; nack cleanup 200;
  attempt history `[(1, 2, 'failed')]` — the failure is attributed to worker-a, which never saw the job.
- Judgement: H-E08-19 **confirmed**. `#question` for the coordinator: what does the worker `token`
  authorise? **Not counted as a bug, reason:** SPEC.md:9-10 defines the bearer key as the only trust
  boundary and one key serves the whole namespace, so every holder is equally trusted; the token's
  purpose is unspecified (spec silence), and attribution in attempt history is only as good as the
  caller's honesty. Recorded as a boundary cell and an open question.

### P8 frozen Postgres (`docker pause`) — prediction written before running

- `#read` api.go:153 auth `QueryRow` has no context/deadline; store.go:16 pool `SetMaxOpenConns(16)`;
  `/health` uses a 1 s `PingContext` (api.go:132).
- `#hypothesis` H-E08-20: while postgres is paused for 12 s, authenticated requests (GET /v1/queues)
  neither fail nor 503 — they hang until unpause and then complete 200 (latency ≈ remaining pause);
  `/health` answers 503 within ≈1 s; loop resumes after unpause. Oracle in question: SPEC.md:189
  "an unavailable authentication store returns 503". Wrong if authenticated requests get a prompt 503.
- `#obs` 19:34:38 `tools/p8_pg_pause.sh` (raw `project/results/e08/p8.txt`):
  ```text
  19:34:38.5 pause postgres
  19:34:50.8 unpause postgres
  19:34:51.0 req health → 200 after 11.79s
  19:34:51.0 req auth1 (acme GET /v1/queues) → 200 after 11.80s
  19:34:51.0 req auth2 (globex GET /v1/jobs/1) → 404 after 11.80s
  19:34:51.0 req badkey → 401 after 11.79s
  19:34:51.1 health-after → 200 after .026s
  ```
- Judgement: H-E08-20 **refuted** in its health clause (code-reading prediction from api.go:132's 1 s
  `PingContext` contradicted: `/health` hung for the whole freeze and then reported `ok`); the auth clause
  held (requests hang, no 503).
- `#surprise` `/health` cannot report a frozen database. Characterising below.
- `#hypothesis` H-E08-21 (written before P8b): pause 25 s; `/health` sent at +1 s and +10 s and an
  authenticated GET at +5 s all return only after unpause (≈24 s, ≈15 s, ≈20 s), all 200; i.e. no
  deadline anywhere on the DB path. Wrong if any returns before unpause.
- `#obs` 19:35:12 `tools/p8b_pg_pause_long.sh` (raw `project/results/e08/p8b.txt`):
  ```text
  19:35:12.7 pause postgres
  19:35:38.1 unpause postgres
  19:35:38.4 req health+1 → 200 after 24.32s
  19:35:38.4 req auth+5  → 200 after 20.33s
  19:35:38.4 req health+10 → 200 after 15.33s
  ```
- Judgement: H-E08-21 **confirmed**: no request on the DB path returns before the database answers;
  latency = remaining freeze. Contrast P2/P3b: a *stopped* DB (connection refused) gives prompt 503s.
  Variable isolated: unresponsive-but-connected vs refusing database. Mechanism for the ignored 1 s ping
  deadline is an inference about lib/pq's cancellation (it waits for the server's reply after issuing a
  cancel request); the driver source is outside this export and was not read.
- `#bug` [E08-F04](../../defects/E08-F04-frozen-database-hangs-health-and-api.md).
- 19:35:50 capture listener (pid 2628168) stopped.

### Not run / testability requests

- TR-E08-1 (ack non-atomicity, `#read` ack.go:13-26): question — can a DB fault after the job UPDATE leave
  a succeeded job with no attempt row / no outbox row while the worker gets 409? Missing signal: a way to
  fail one statement deterministically. Minimal hook: a test-only fault switch in the store (fail the N-th
  statement of the next request) or a SQL-level lab trigger installed by the lab, not the SUT. Validation
  probe: ack with the fault armed after the UPDATE; expect either full rollback (job still leased) or full
  success (attempt + outbox), and a 5xx (not 409) on failure.
- TR-E08-2 (shipper visibility): per-row backlog age / oldest undelivered `next_at` lag as a metric, and
  `/health` degraded when the shipper lag exceeds a bound — P5c showed `/health` 200 during a permanent
  stall. Validation: P5b blackhole run should flip the signal within one pass.
- TR-E08-3 (hooksink): record the received `X-Queuelet-Signature` header (E01 TR1, still open) and add a
  redirect mode (302/307) — `#offcharter`: Go's client follows redirects for webhook POSTs (ship.go:31,
  default `http.Client`), so a 302→GET 200 may count as "delivered" without the body; not probed.
- `#offcharter` schedule.Fire returns on the first failing schedule (fire.go:37-40), so one schedule that
  cannot advance would starve later schedules each tick — sibling pattern of E08-F02 inside the core loop;
  no externally reachable trigger found in this run (cron/timezone validated at create/PATCH).

## Close

- 2026-10-22T19:39:50Z stack `queuelet-e08` stopped at 19:38:33Z (`docker compose stop`; logs saved to `project/results/e08/api-log-final.txt`, `neighbours-log-final.txt`). Reports E08-F01…F04, models, hypotheses, [revealed.md](revealed.md) and STATE.md written.
