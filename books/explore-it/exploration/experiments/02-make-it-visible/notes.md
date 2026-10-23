# Experiment 02 notes — make-it-visible

Build: plain export of `explore-it/lesson-01` (no git metadata). Stack: Compose project `queuelet-e02`,
image tag `e02`, API :18230, hooksink :18231, postgres :15463. All commands sourced
[`tools/e02env.sh`](tools/e02env.sh) first (defines `dc`, `api`, `sql` (read-only txn as `explorer`), `ts`).
Raw outputs: `project/results/e02/`. Times are UTC.

Run start (coordinator file `run-start.txt`): 2026-10-22T03:18:01Z. Planned split: reading 10 min, probing 30 min, writing 20 min.

## 03:18:09Z–03:20:21Z reading

- 03:18Z #read STATE.md, charter, PROTOCOL, skill, brief, lesson 01 (current checkpoint's lesson, not a future one),
  hypotheses.md, all models. Prior record: E01 found F01–F27 at v0; lesson 01 claims all 35 fixed. Those are regressions,
  not discoveries for E02.
- 03:19Z #read `internal/jobs/expire.go:12-45` — Expire now selects `attempts+1,max_attempts,leased_by` for expired
  leases, sets `queued` or (`n >= max`) `dead`, writes `last_error='lease expired'`, inserts an attempt row
  `(id,n,worker,now,now,'expired','lease expired')`; Notify only for dead (after commit, `:53-56`).
  Interpretation: attempt row `started_at` = `finished_at` = expiry tick, not lease start (E01 Q1 persists for expiry).
- 03:19Z #read `internal/webhook/ship.go:16-49` — selects undelivered due rows `ORDER BY id LIMIT 20`; on transport
  error or non-2xx: `attempts+1, next_at=now+5s`, **no log line**, `continue`; no attempt cap, no growth (fixed 5 s).
  Success logs `webhook delivered status=%d` **without outbox id or job id** (`:45`). `http.NewRequest` error still
  `return e` (`:26-28`), aborting the batch (E01 H44 still open).
- 03:19Z #read `internal/api/api.go:111-145` — separate shipper goroutine; core steps Promote, Expire, Fire, Retain each
  logged as `background task: %v` on error and the loop continues; `/health` 503 only when DB ping fails or the last
  tick of either loop returned an error (`:62-70`). Partner 500s do not make `Ship` return an error, so health stays ok.
  `/metrics` is the constant `queuelet_up 1` (`:71`).
- 03:19Z #read `internal/worker/worker.go:47-71` — worker logs `processing kind=%s attempt=%d` (no job id) and
  `completed status=%d error=%v` (no job id); heartbeat every 5 s, response ignored; `crash` = `os.Exit(3)` (`:84-85`).
  `deploy/compose.yml`: workers `restart: on-failure`; worker-a consumes `emails`, worker-b `reports`.
  Each restart calls `POST /v1/workers` again (`:29`) → a new worker row each incarnation; nothing deletes the old one.
- 03:19Z #read `internal/api/jobs.go` — no request logging anywhere in the API; `cmd/api/main.go:41` logs only
  "queuelet listening". Consoles inventory below.
- 03:19Z #read seed `lab/seed/seed.py`: emails lease 10 s / max 3; reports lease 20 s / max 4 with webhook
  `http://hooksink:8201/hook` secret `report-secret`; backoff exponential base 2.
- 03:20Z #read `migrations/001.sql:7` — `attempts` has no key; nothing prevents duplicate `n`.

### Console inventory (skill step 1)

| Console | Producer (source) | Identifiers it carries | Clock |
|---|---|---|---|
| API HTTP status/body | `internal/api/*.go` | job id, state, attempts, lease fields, last_error | API `time.Now()` |
| API process log | `log.Printf` in api.go:124,139; ship.go:45; main.go:41 | none (no job/outbox/worker id) | log timestamp, second resolution (Go default) |
| Worker stdout | worker.go:32,43,48,71 | worker name, kind, attempt; **no job id** | worker container clock (same host) |
| Docker container state | compose restart policy | RestartCount, exit code | docker daemon |
| DB `jobs`, `attempts`, `workers`, `webhook_outbox` | SQL in jobs/*.go, ship.go | ids, n, worker_id, outcome, outbox attempts/next_at/delivered_at | API ticker `now` (Loop), request `now` |
| hooksink `/admin/records` | cmd/hooksink/main.go:67 | body (job_id,state), status sent, signature_valid, `at` | hooksink clock |

## 2026-10-22T03:20:58Z #hypothesis — stated before any probe

| Id | Source | Prediction | Would be wrong if |
|---|---|---|---|
| E02-H01 | charter; expire.go:35-45; compose restart:on-failure; seed emails lease 10 s/max 3 | A `crash` job (priority 9) on emails is leased by worker-a, the worker exits 3, Docker restarts it; each 10 s expiry writes an `expired` attempt row; after the 3rd expiry the job is `dead`, attempts=3, last_error `lease expired`, exactly 3 executions (~30–45 s) | the worker logs `processing kind=crash` more than 3 times, or the job ends other than dead/attempts=3/3 expired rows, or jobs.attempts disagrees with attempt rows |
| E02-H02 | grep of log.Printf in internal/api, internal/jobs | The API log shows **no** line for lease, expiry, dead transition of the crash job | API log names any of these transitions |
| E02-H03 | worker.go:29-48; compose restart policy | worker-a stdout shows `processing kind=crash attempt=1..3` with no `completed` line between, a `registered` line per incarnation; RestartCount rises by 3; `workers` gains one row named worker-a per incarnation (old rows remain) | RestartCount ≠ +3, or workers rows don't grow |
| E02-H04 | expire.go:45 | Each expired attempt row has worker_id = the incarnation that held the lease (distinct ids) and started_at = finished_at = expiry tick (≈ lease start + 10 s), so no console shows when the crashed execution started | started_at precedes finished_at, or worker_id repeats across incarnations |
| E02-H05 | ship.go:30-37; hooksink main.go:60-67 | hooksink mode 500 + echo job acked on reports: one outbox row; POST repeated every ≈5–6 s; each recorded by hooksink with status 500 and signature_valid true; outbox.attempts == hooksink record count; delivered_at NULL; API log silent during failures | delivered_at set while 500, or counts disagree, or API logs failures |
| E02-H06 | ship.go:43-46 | After mode→ok the next POST succeeds: delivered_at set, API logs one `webhook delivered status=200` (no id), hooksink count = outbox attempts | not delivered within 10 s, or extra deliveries after delivered_at |
| E02-H07 | ship.go:33 (fixed +5 s, no cap) | Retry interval stays ≈5–6 s across ≥5 failures (no backoff), no give-up | intervals grow or retries stop |
| E02-H08 | api.go:62-70 | `/health` stays 200 `ok` during sustained partner 500s; `/metrics` unchanged | health 503/degraded |
| E02-H09 | worker.go:86-95; nack.go:157-165 | `fail5` job on emails (max 3): worker nacks attempts 1,2 (retry_wait, backoff 2 s, 4 s), 3rd nack → dead; 3 `failed` attempt rows; worker logs `completed status=200` ×3 | 4th execution, or not dead after 3 |

## 2026-10-22T03:23:57Z #obs stack and fixture

- 03:21:03Z `dc up -d --build` → exit 0; `lab/wait.sh` healthy 03:23:06Z (raw: results/e02/up.log). `dc config` showed
  `name: queuelet-e02`, published 18230/18231/15463 (other stacks queuelet, queuelet-e01 untouched).
- 03:23:12Z `bash lab/reset.sh` (env exported) → seeded; per-queue counts (queue 1 emails, 2 reports, 3 imports) each
  cancelled 6/dead 6/retry_wait 6/scheduled 25; workers rows: 1 fixture-writer, 2 worker-b, 3 worker-a; RestartCount 0/0.
- 03:23:40Z emails drained (no queued/leased on queue 1), reports 1 queued; worker log `processing kind=echo attempt=1` /
  `completed status=200 error=<nil>` — no job id in either line.

## 2026-10-22T03:23:57Z #hypothesis (formed while preparing H09)

| Id | Source | Prediction | Would be wrong if |
|---|---|---|---|
| E02-H10 | jobs.go:44 `kind=ANY(kinds)`; cmd/worker/main.go:138 default kinds `echo,sleep,fail,crash,http`; README "use kind fail3" | A job of kind `fail3` on emails is **never leased** by compose worker-a (exact-kind match), though worker.go:86 handles the `fail` prefix; it stays `queued` indefinitely with no console saying why | worker-a processes it within 10 s |

H09 is therefore run with kind `fail` and payload `{"n":5}`.

## 2026-10-22T03:24:03Z #obs crash probe (E02-H01..H04) — three consoles

Command (03:24:03.080Z): `api POST /v1/queues/emails/jobs -d '{"kind":"crash","payload":{"probe":"e02-crash"},"priority":9}'`
→ `{"id":301,"state":"queued"}`; then `tools/watch-job.sh 301 50 results/e02/crash-watch.txt` (1 s samples of API GET,
DB row, attempts, worker rows, RestartCount). Raw: `project/results/e02/crash-watch.txt`.

API GET transitions (deduplicated samples):
```text
03:24:03.153Z queued attempts 0
03:24:04.666Z leased attempts 0 leased_by 3 lease_expires_at 03:24:13.17337Z  last_error ""
03:24:13.963Z leased attempts 1 leased_by 4 lease_expires_at 03:24:23.538163Z last_error "lease expired"
03:24:24.840Z leased attempts 2 leased_by 5 lease_expires_at 03:24:34.520581Z last_error "lease expired"
03:24:36.130Z dead   attempts 3 leased_by null                                 last_error "lease expired"
```
DB `attempts` (read-only):
```text
301 | 1 | 3 | 03:24:13.50681  | 03:24:13.50681  | expired | lease expired
301 | 2 | 4 | 03:24:24.506809 | 03:24:24.506809 | expired | lease expired
301 | 3 | 5 | 03:24:35.506808 | 03:24:35.506808 | expired | lease expired
```
jobs 301: dead, attempts 3/max 3, finished_at 03:24:35.506808. workers: 3 worker-a (last_seen 03:23:28.36), 4 worker-a
(03:24:04.15), 5 worker-a (03:24:14.59), 6 worker-a (03:24:25.52) — **last_seen_at equals registration time** for each.
Worker-a stdout (`dc logs -t worker-a`, echo lines removed):
```text
03:24:03.186 processing kind=crash attempt=1
03:24:04.174 worker worker-a registered
03:24:13.551 processing kind=crash attempt=2
03:24:14.593 worker worker-a registered
03:24:24.538 processing kind=crash attempt=3
03:24:25.528 worker worker-a registered
```
`docker inspect` worker-a: RestartCount 3, ExitCode 0 (current incarnation), StartedAt 03:24:25.04. API log since 03:24:00: **empty**.

Judgments:
- E02-H01 **confirmed**: 3 executions, dead, attempts 3, three `expired` rows; ~32.5 s from enqueue to dead. API, DB and worker
  agree on count. The charter prediction "retries up to max_attempts" holds for crash at lesson-01 (F05 regression, not a discovery).
- E02-H02 **confirmed**: the API log has no line for lease, expiry or dead. The death is visible only in DB/API state.
- E02-H03 **confirmed**: 3 `processing` lines, no `completed`, 3 re-registrations, RestartCount 3, workers rows 4–6 added; old rows 3–5 remain.
- E02-H04 **confirmed**: worker ids 3,4,5 distinct; started_at = finished_at = Expire tick (ticks land at .506 s; e.g.
  lease 1 expired at 03:24:13.173 but is recorded 03:24:13.507). Lease start (03:24:03.17) appears in no console except the
  transient API `lease_expires_at - lease_seconds`; the exit code 3 is nowhere (Docker reports the *current* container's 0).
- 03:25Z #surprise S1: at 03:24:13.963 the job is `leased` (attempt 2 in progress) while GET shows `last_error:"lease expired"`.
  **Not a bug**: SPEC:63 "GET returns the current job … and last error"; the previous attempt's error is the last error.
  Recorded as a visibility note: a reader cannot tell from GET whether the error belongs to the current execution.
- 03:25Z #surprise S2: crashed incarnations' worker rows (3,4,5) remain registered with `last_seen_at` = registration time
  and still pass the lease worker-exists check (jobs.go api:143). **Not a bug against SPEC** (no worker liveness or expiry is
  specified; SPEC:69-71 only create/delete); it is a testability gap → TR-E02-3.
- #question Q-E02-1: which process exit caused each expiry? No console correlates exit code 3 to job 301 — the worker log line
  lacks a job id and Docker keeps only the current exit code. "Did not ack" is visible; "crashed" is inferred.

## 2026-10-22T03:26:10Z #obs hooksink 500 setup
`POST /admin/mode {"mode":"500"}` → `{"mode":"500"}`; `DELETE /admin/records`. Outbox before: 69 rows, 0 undelivered, max id 69.

## 2026-10-22T03:26:10Z #obs webhook 500 probe (E02-H05..H08) and fail probes (E02-H09, H10)

03:26:10.304Z `api POST /v1/queues/reports/jobs {"kind":"echo","payload":{"probe":"e02-hook500"},"priority":9}` → id 302;
`api POST /v1/queues/emails/jobs {"kind":"fail","payload":{"n":5},"priority":9}` → 303;
`api POST /v1/queues/emails/jobs {"kind":"fail3","payload":{},"priority":8}` → 304. Then `tools/watch-outbox.sh 302 40`
(raw: results/e02/hook500-watch.txt). 03:27:06.955Z mode → ok; `watch-outbox.sh 302 12` (results/e02/hookok-watch.txt).

hooksink `/admin/records` for job 302 (results/e02/hook-302-records.jsonl), `at` per POST:
```text
03:26:11.516 500 | 16.546 500 | 22.511 500 | 27.510 500 | 32.514 500 | 38.513 500 | 44.517 500 | 49.509 500
03:26:54.510 500 | 59.511 500 | 03:27:04.509 500 | 03:27:09.510 200   (all signature_valid=true)
```
outbox row 70 at end: `attempts 12, next_at 03:27:09.506847, delivered_at 03:27:09.50685`, body
`{"state":"succeeded","job_id":302,"result":{"probe":"e02-hook500"},"last_error":""}`. Outbox samples during 500 mode:
att 0 → 1 (next 03:26:16.507) → 2 (next 03:26:21.507) …; delivered_at `-` until 03:27:09.
API log since 03:26:00: exactly one line `2026/10/22 03:27:09 webhook delivered status=200`. `/health` 52/52 samples
`{"status":"ok"}` 200; `/metrics` = `queuelet_up 1`.

- E02-H05 **confirmed**: 11 POSTs with 500, hooksink count 12 = outbox attempts 12, delivered_at NULL throughout 500 mode,
  API log silent for every failure. (F11 regression holds: non-2xx is not delivered.)
- E02-H06 **confirmed**: first POST after mode→ok (2.6 s later) delivered; one log line, no id; no extra POST after delivery
  (12 samples to 03:27:19).
- E02-H07 **confirmed**: gaps 5.0–6.0 s (5,6,5,5,6,6,5,5,5,5,5 s) over 11 failures; no growth, no give-up. 6 s gaps are
  tick granularity (next_at = tick+5 s exactly; equality at a later tick sometimes misses by µs). SPEC:121 "Failed
  notifications are retried" sets no cap/backoff → **not a bug**; an always-failing partner is retried forever with nothing
  in the API log (TR-E02-1/2).
- E02-H08 **confirmed**: health ok during sustained partner failure; SPEC:128-129 ties health to DB/loop degradation, not
  partner outcome → not a bug, but no console in queuelet shows a notification backlog except SQL.
- E02-H09 **confirmed** (job 303): attempts rows n=1,2,3 `failed` at 03:26:10.705, 13.866, 18.790 (gaps 3.2 s, 4.9 s ≈ backoff
  2 s, 4 s + tick), dead attempts 3/3, last_error `requested failure 3`. Worker log: `processing kind=fail attempt=N` then
  `completed status=200 error=<nil>` ×3 — **the worker says "completed" for each nack**; its log cannot distinguish ack from
  nack. #surprise S3 → not a product bug (sample worker logging), recorded as TR-E02-1 (job id + verb in worker log).
  Also #read/#obs: the dead job's `run_at` is 03:26:26.79 = last nack + 8 s backoff (nack.go:169 always writes run_at).
  #surprise S4 → **not a bug**: run_at is not specified for terminal jobs and Promote only reads scheduled/retry_wait.
- E02-H10 **confirmed** (job 304, kind `fail3`): at 03:27:30.969 still `queued`, attempts 0, no attempt rows, no worker
  line; 80 s after enqueue. #surprise S5 → characterized below (E02-F01).

## 2026-10-22T03:28:29Z #hypothesis — next probes

| Id | Source | Prediction | Would be wrong if |
|---|---|---|---|
| E02-H11 | queues.go:21-41,79-84 (no URL validation); ship.go:26-28 `return e`; api.go:121-124 | A queue whose webhook URL fails `http.NewRequest` (`http://bad host/x`) accepts create (201). Its first terminal job's outbox row makes every Ship tick return the parse error: API log `webhook task: parse …` each second, `/health` 503 degraded, the row's attempts stay 0 and every later-id outbox row — **including another tenant's** (acme reports) — is never POSTed | create → 400; or acme reports job is delivered while the bad row is undelivered; or health stays ok |
| E02-H12 | ship.go:35 (2 s client timeout); hooksink `slow` sleeps 3 s then 200 | In `slow` mode every POST is recorded by hooksink as status 200 but queuelet times out at 2 s and retries every ~5 s; outbox never delivered while slow; hooksink records n ≥ 3 duplicate 200s in 20 s; API log silent | outbox delivered in slow mode |
| E02-H13 | expire.go:53-56; Notify jobs.go:69-81 | A `crash` job on reports (lease 20 s, max 4) dies by expiry after 4 executions and produces one outbox entry with state dead, last_error `lease expired`, delivered to hooksink | no outbox row for the expiry-dead job |

## 2026-10-22T03:28:39Z #obs E02-H11 unparsable webhook URL

globex-key (`g` = api with globex key, prints [status]):
```text
03:28:39.283 g POST /v1/queues {"name":"e02bad","webhook":{"url":"http://bad host/x","secret":"s"}} → 201 (queue id 5, webhook_url "http://bad host/x")
             g POST /v1/workers {"name":"e02-manual","kinds":["echo"]} → 201 worker_id 7
             g POST /v1/queues/e02bad/jobs {"kind":"echo","payload":{"probe":"h11"}} → 201 id 305
             g POST /v1/queues/e02bad/lease {"worker_id":7,"max_jobs":1} → job 305 + token
03:28:39.531 g POST /v1/jobs/305/ack {...} → 200 {"ok":true}
03:28:42.643 outbox 71 job 305 url "http://bad host/x" attempts 0 delivered NULL;  GET /health → 503 {"status":"degraded"}
api log, once per second from 03:28:40.513:  webhook task: parse "http://bad host/x": invalid character " " in host name
```
03:28:50.961 acme `POST /v1/queues/reports/jobs {"kind":"echo","payload":{"probe":"e02-h11-acme"},"priority":9}` → 306; worker-b
acked it (job 306 succeeded attempts 1). `watch-outbox.sh 306 20` (results/e02/h11-acme-watch.txt): 20/20 samples
`OUTBOX: 72 att=0 … deliv=-`, `HOOK n=0`, `HEALTH degraded` through 03:29:15.

- E02-H11 **confirmed** → #surprise S6 / #bug E02-F02: a globex queue's bad URL stops delivery of acme's notification.
  The API log names the URL but neither the queue, namespace nor outbox row; health says "degraded" without a cause.

## 2026-10-22T03:29:35Z #hypothesis — E02-F02 variable isolation (before probes)

| Id | Prediction | Would be wrong if |
|---|---|---|
| E02-H11a | PATCH e02bad webhook → valid `http://hooksink:8201/hook` does **not** unblock: outbox row 71 keeps its copied URL (Notify copies url, jobs.go:80); 306 stays undelivered | 306 delivered within 8 s of the PATCH |
| E02-H11b | DELETE /v1/jobs/305 (terminal) → 204 but outbox row 71 remains (no FK, deleteJob only touches jobs) and blocking continues | row 71 disappears or 306 delivered |
| E02-H11c | A URL that parses but cannot be dialled (`ftp://x/y`, unsupported scheme at Do) is treated as a transport failure (attempts+1, +5 s) and does **not** block later rows | ftp row blocks or logs `webhook task:` |

## 2026-10-22T03:29:42Z #obs E02-H11a/H11b

```text
03:29:42.122 g PATCH /v1/queues/e02bad {"webhook":{"url":"http://hooksink:8201/hook","secret":"s"}} → 200 (webhook_url now valid)
03:29:50.178 outbox 71 url "http://bad host/x" att 0 undelivered; 72 (job 306) att 0 undelivered; hooksink records 0
03:29:50.335 g DELETE /v1/jobs/305 → 204
03:29:58.393 jobs 305: 0 rows; outbox 71 still present (job_id 305, bad URL); 72 undelivered; hooksink 0; /health 503 degraded
```
- E02-H11a **confirmed**: fixing the queue configuration does not repair the queued outbox entry.
- E02-H11b **confirmed**: deleting the owning job leaves an orphan outbox row that keeps blocking (C10 sibling: outbox has no FK).
- No public verb removes or skips an outbox row; recovery needs DB surgery or a lab reset. API log 03:28:30–03:29:58:
  every line is the parse error (results/e02/h11-api.log).
- E02-H11c **cannot be probed in this state**: row 71 is always first (ORDER BY id), so no later row is reached. Lab reset next.

## 2026-10-22T03:30:14Z #obs lab reset (own stack only)
`bash lab/reset.sh` 03:30:14.640Z → done 03:30:33.855Z; /health 200 ok. New fixture identity: same seed, ids 1–300 again.

## 2026-10-22T03:30:46Z #hypothesis — sibling: PATCH path
| E02-H11d | queues.go:79-84 — PATCH accepts `webhook.url` `http://bad host/x` with 200 (no validation), same as create | PATCH → 400 |

## 2026-10-22T03:30:46Z #obs E02-H11c/H11d (after reset)

```text
03:30:46.847 g POST /v1/queues {"name":"e02ftp","webhook":{"url":"ftp://x/y",...}} → 201
             g POST /v1/queues {"name":"e02patch","webhook":{"url":"http://hooksink:8201/hook",...}} → 201
             g PATCH /v1/queues/e02patch {"webhook":{"url":"http://bad host/x",...}} → 200, webhook_url "http://bad host/x"
             globex worker; enqueue e02ftp echo → job 301; lease; ack → 200
             acme POST reports/jobs {"probe":"e02-h11c-acme"} → 302 (worker-b acks)
03:30:59.341 outbox 69 job 301 ftp://x/y attempts 3, next 03:31:03.27, undelivered
             outbox 70 job 302 delivered 03:30:48.27 (attempts 1); outbox 71 job 251 (seed backlog) delivered
             hooksink: jobs 71,131,191 (seed backlog after reset), 302, 251 all 200; /health 200 ok
api log: only "webhook delivered status=200" lines; nothing for the ftp failures
```
- E02-H11c **confirmed**: a URL that parses but fails at Do() is a per-row failure (attempts+1/5 s, forever, silent) and does
  not block other rows. So the blocking defect is specific to NewRequest-time errors (ship.go:26-28). Variable isolated.
- E02-H11d **confirmed**: PATCH stores an unparsable URL with 200; it would block once a terminal job is notified
  (mechanism same as create; not re-triggered to save a reset — inference, not observation).

## 2026-10-22T03:31:23Z #obs E02-H12 slow partner (prediction above, 03:28Z table)
`POST /admin/mode {"mode":"slow"}`, `DELETE /admin/records`; 03:31:23.304 `api POST /v1/queues/reports/jobs
{"kind":"echo","payload":{"probe":"e02-slow"},"priority":9}` → 303; `watch-outbox.sh 303 22` (results/e02/h12-watch.txt);
03:31:47.986 mode → ok.
```text
hooksink records for job 303 (results/e02/h12-records.jsonl):
  03:31:24.273 200 | 03:31:29.278 200 | 03:31:34.274 200 | 03:31:40.275 200 | 03:31:45.274 200
outbox 72 at 03:31:48: attempts 5, next 03:31:50.269, delivered_at NULL;  /health 22/22 ok
after mode ok: outbox 72 attempts 6, delivered_at 03:31:51.269; api log 03:31:51 "webhook delivered status=200" (only line)
```
- E02-H12 **confirmed**. #surprise S7 — consoles disagree: hooksink says it answered 200 five times; queuelet records five
  failures and logs nothing. Six POSTs, one "delivered". Characterized as E02-F03 (status `spec`).

## 2026-10-22T03:33:25Z #read erratum (citations)
Some 03:19–03:28Z line numbers were read from concatenated `cat -n` output. Per-file lines (checked with `grep -n`):
ship.go select :17, secret :23, NewRequest :25 (return :26-28), client 2 s timeout :31, failure branch :32-41 (UPDATE +5 s :37),
delivered log :45, delivered UPDATE :46; cmd/worker/main.go default kinds :24 (not :138); nack.go select :14, dead test :19,
run_at write :26 (not :157-169). Hypothesis rows above keep their original text; defect reports use the corrected lines.

## 03:34Z Console comparison (skill output) and testability requests

| Object / interval | Prediction | API | DB | Process logs | Partner | Disagreement |
|---|---|---|---|---|---|---|
| job 301 crash, 03:24:03–03:24:36 | 3 executions → dead | dead, attempts 3, last_error "lease expired" | 3 `expired` rows, worker 3/4/5, started=finished | worker: 3 × processing, 3 × registered; Docker RestartCount 3, exit code lost; API: nothing | — | none in counts; *start* times and *cause* (exit 3) invisible |
| job 303 fail n=5 | 3 nacks → dead | dead 3/3 | 3 `failed` rows | worker: "completed status=200" × 3 (nack reported as completed) | — | worker wording vs DB outcome |
| job 304 fail3 | never leased | queued forever | no rows | silent | — | nothing explains the wait (E02-F01) |
| job 302 webhook in 500 mode | 11 fail + 1 ok | succeeded (no notification field) | outbox attempts 12, delivered 03:27:09 | API: 1 line, no id | 12 records (11×500, 1×200) | API log hides 11 failures |
| outbox 71 bad URL | blocks all | — | attempts 0 forever; later rows 0 | API: parse error each second, no row/queue id | 0 records | health "degraded" without cause (E02-F02) |
| job 303 (2nd fixture) slow partner | 5 timeouts | succeeded | attempts 5 undelivered | silent | 5 × status 200 | partner success counted as failure (E02-F03) |

Questions no console answered (did-not-happen vs cannot-observe):
1. When did a crashed execution start, and which process exit ended it? (cannot observe: attempt started_at = expiry tick;
   worker log has no job id; Docker keeps only the latest exit code.)
2. Which outbox row / job did "webhook delivered status=200" refer to; what status did each failed POST get?
   (cannot observe in queuelet: only `attempts` count and last `next_at` are stored; per-POST history only at hooksink.)
3. Is any notification stuck, for which tenant, and since when? (cannot observe without SQL; /health is boolean, /metrics constant.)
4. Is a worker incarnation alive? (cannot observe: `last_seen_at` = registration time, heartbeats do not update it.)
5. Why is a queued job not leased? (cannot observe: no eligibility reason anywhere.)

Testability requests (proposals for the fixer; none implemented here):

| Id | Question | Missing signal | Minimal hook | Validation probe |
|---|---|---|---|---|
| TR-E02-1 | Which job/attempt/row did each event concern? | correlated ids in logs | one structured log line per lifecycle event: `event=lease|ack|nack|expire|dead|notify_attempt` with `job_id, attempt, worker_id, outbox_id, status, error, at` (API) and job id + verb in worker lines | rerun the crash probe: api log alone shows lease ×3, expire ×3, dead ×1 for job 301 with worker ids matching `attempts` |
| TR-E02-2 | What happened on each notification POST? | per-attempt delivery history | outbox `last_status`, `last_error`, `last_attempt_at` columns (or `webhook_attempts` table) exposed via `GET /v1/jobs/{id}` (or `/notifications`) | 500-mode probe: 11 rows/fields with status 500 then one 200, equal to hooksink records |
| TR-E02-3 | Is the shipper/worker healthy, and what is stuck? | backlog gauges and liveness | `/metrics`: `queuelet_outbox_undelivered`, `queuelet_outbox_oldest_age_seconds`, `queuelet_ship_errors_total`, `queuelet_core_errors_total`; `/health` body names the failing loop and last error; heartbeats update `workers.last_seen_at` | bad-URL probe: metrics show oldest age growing and ship errors incrementing; health body says `ship: parse … outbox 71` |
| TR-E02-4 | Why is a queued job waiting? | eligibility explanation | stats/queue read listing queued kinds with no registered worker, or a lease debug field | `fail3` probe shows "no worker supports kind fail3" |
| TR-E02-5 | Lease start for attempts | start timestamp | attempts.started_at = lease time (store `leased_at` on jobs) | crash probe: started_at 03:24:03.17, finished_at 03:24:13.51 |
| TR-E02-6 | Real-time waits (lease 10–20 s × attempts, retry 5 s) | controllable clock | lab-only clock offset for the API loop (E01 TR2 restated) | crash probe completes in < 5 s with clock advance; note that worker/partner clocks stay real |

## 2026-10-22T03:32:05Z #obs E02-H13 crash on reports (webhook queue)

03:32:05.026 `api POST /v1/queues/reports/jobs {"kind":"crash","payload":{"probe":"e02-crash-reports"},"priority":9}` → 304
(second fixture). `watch-job.sh 304 100` (results/e02/h13-watch.txt; its WRK column tracks worker-a and is irrelevant here).
```text
attempts 304: n1 w3 03:32:25.269 | n2 w5 03:32:46.269 | n3 w6 03:33:07.269 | n4 w7 03:33:28.269  (all expired, started=finished)
jobs: dead attempts 4 (03:33:29.591 GET); worker-b RestartCount 4; 4 × "processing kind=crash attempt=N", 4 × "registered"
outbox 73: job 304 attempts 1 next_at 03:33:28.269 delivered 03:33:29.269 body {"state":"dead","job_id":304,"result":null,"last_error":"lease expired"}
hooksink: 03:33:29.280 status 200 signature_valid true, same body; api log: one "webhook delivered status=200"
```
- E02-H13 **confirmed**: expiry-dead notifies exactly once with state dead and last_error. ~84 s wall clock for one probe (TR-E02-6).
- 03:33:28.178 sample shows API `leased attempts 3` while the next SQL line already has 4 attempt rows: my watcher queries
  consoles sequentially and an Expire tick (03:33:28.269) fell between them. **Not a bug**: instrument artifact; a
  same-snapshot read (TR-E02-2 per-job history) would avoid it.

## 2026-10-22T03:34:54Z #hypothesis — frozen worker (container action)

| Id | Source | Prediction | Would be wrong if |
|---|---|---|---|
| E02-H14 | worker.go:49-71 (heartbeat 5 s, response ignored; ack result only logged); expire.go; ack.go token guard | `sleep` job (seconds 4) on emails; `docker pause` worker-a ≈1 s after lease for 14 s: lease expires at ≈+10 s (attempt 1 `expired`); after unpause the worker finishes, ack → **409**, logged only as `completed status=409` in worker stdout; then it leases the same job again (attempt=2), acks 200 → succeeded attempts 2; rows 1 expired + 2 succeeded. The work ran twice; the API log shows nothing | first ack accepted, or job not re-run, or attempts ≠ 2 |

## 2026-10-22T03:34:54Z #obs E02-H14 frozen worker

```text
03:34:54.900 api POST /v1/queues/emails/jobs {"kind":"sleep","payload":{"seconds":4,"probe":"e02-freeze"},"priority":9} → 305
03:34:55.989 docker pause queuelet-e02-worker-a-1; job 305 leased by 2, lease_expires_at 03:35:04.970
03:35:10.207 job 305 queued, attempts 1, leased_by NULL;  docker unpause (03:35:10.649)
worker-a: 03:35:10.654 completed status=409 error=<nil> | 03:35:10.987 processing kind=sleep attempt=2 | 03:35:15.022 completed status=200 error=<nil>
03:35:20.661 job 305 succeeded attempts 2, result {"probe":"e02-freeze","seconds":4}, last_error "lease expired"
attempts: n1 worker 2 03:35:05.269 expired | n2 worker 2 03:35:14.986 succeeded;  api log since 03:35:00: empty
```
- E02-H14 **confirmed**: the stale ack is refused (409, F01 regression holds), the job re-runs and succeeds; the work ran
  twice; the only trace of the rejected completion is a worker line whose `error=<nil>` contradicts its 409
  (client.Do returns no error for non-2xx). #surprise S8 → **not a product bug** (sample-worker logging; TR-E02-1).
- #surprise S9: succeeded job 305 keeps `last_error:"lease expired"` (Ack, ack.go:10, does not clear it). **Not a bug against
  SPEC:63/117** ("last error" is the last error that occurred), but a reader or webhook consumer cannot tell whether it
  applies to the final outcome → #question Q-E02-2 for the coordinator (should success clear or scope last_error?).

## 2026-10-22T03:35:51Z #hypothesis — DB restart consoles (regression context, not counted as discovery)

| Id | Source | Prediction | Would be wrong if |
|---|---|---|---|
| E02-H15 | api.go:62-70,121-143; F08/F14 fixed per lesson 01 | Enqueue echo with delay_seconds 8 on emails, then `dc restart postgres`: during the outage /health 503 degraded, acme requests 503 auth_unavailable, API log `background task:`/`webhook task:` lines; after postgres is back /health returns ok within 3 s and job is promoted and acked by worker-a ≤ 5 s after due | job not succeeded 15 s after restart end, or health stays degraded |

## 2026-10-22T03:35:51Z #obs E02-H15 postgres restart (own stack)

`api POST /v1/queues/emails/jobs {"kind":"echo","payload":{"probe":"e02-dbrestart"},"delay_seconds":8}` → 306 scheduled
(03:35:51.791); `dc restart postgres` at 03:35:51.875; 1 s samples (results/e02/h15-watch.txt):
```text
03:35:51.9–53.1 health 200 auth 200 | 03:35:54.3–56.7 health 503 auth 503 (GET job → auth_unavailable)
03:35:57.9 health 503 auth 200 | 03:35:59.5 health 200 | 03:36:00.7 job succeeded (run_at 03:35:59.846, finished 03:36:00.360)
api log (results/e02/h15-api.log, 24 lines 03:35:53–03:35:5x): "background task: pq: terminating connection…",
  "…connection reset by peer", "…database system is shutting down", "background task|webhook task: dial tcp …: connection refused"
worker-a: "HTTP 503: {"error":{"code":"auth_unavailable",…}}" once per second 03:35:53–03:35:57
```
- E02-H15 **confirmed** (regression of F08/F14 fixes; not a discovery). Health lags the DB by one tick (03:35:57.9 auth ok,
  health still 503) because it reports the *previous* tick's failure flag — expected from api.go:62-70, not a bug.
  Log lines carry the cause but not which step (Promote/Expire/Fire/Retain) failed (TR-E02-1).

## 2026-10-22T03:36:35Z #obs final snapshot and stop
`results/e02/final-snapshot.txt`: workers 1 fixture-writer, 2 worker-a, 3 worker-b, 4 e02-manual, 5–8 worker-b (each
last_seen_at = registration); outbox 73 rows, 1 undelivered (ftp row 69); worker-a restarts 0 (Docker reset the count at
the lab reset's stop/start), worker-b restarts 4. `dc logs -t` → results/e02/all-logs-at-stop.txt (815 lines).
03:36:36.697Z `dc stop` → all queuelet-e02 containers stopped 03:36:39Z; `docker compose ls` still shows queuelet and
queuelet-e01 running (untouched). Volume kept for the coordinator.

## 2026-10-22T03:39:40Z run end
Writing finished. Outputs: revealed.md, E02-F01..F03, model and hypothesis updates, defects/README E02 table.
