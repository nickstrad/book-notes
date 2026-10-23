# Experiment 06 — relate-it: notes

Build: plain export of `explore-it/lesson-05` at `/tmp/explore-it-experiment-06`; isolated Compose project
`queuelet-e06` (API :18270, hooksink :18271, PG :15503, image tag e06, `QUEUELET_TEST_CLOCK=1`).
Every command below ran after `source tools/env.sh` (helpers: [tools/env.sh](tools/env.sh) defines `k KEY METHOD PATH [BODY]`
= curl with bearer key printing `body [status]`, `sql` = read-only psql as `explorer`, `dc` = compose pinned to
`-p queuelet-e06`). [tools/reset-e06.sh](tools/reset-e06.sh) is an adapted copy of `project/lab/reset.sh` pinned to this
stack (inherited copy unchanged). Raw output: `project/results/e06/`.

## 15:28:42Z start / reading

- 15:28:42Z `date -u` at start. Read STATE.md, charter, brief, PROTOCOL, skill probe-entities-and-boundaries, SPEC.md,
  migrations 001–004, defects INDEX/README, models entities/ecosystem, hypotheses (E04/E05 tail), E05 revealed.
- 15:29Z `dc up -d --build` (background) → `results/e06/up.log` ends `exit 0`; 5 containers up.

#read migrations (database constraints vs application enforcement)
- `migrations/001.sql:1-9` — no foreign keys at all in the original schema; `attempts`, `webhook_outbox`, `workers`,
  `schedules` have no NOT NULL on owner columns.
- `migrations/002.sql:9-17` — **contradicts the charter premise**: `jobs_queue_owner` (jobs.queue_id → queues.id) and
  `schedules_queue_owner` (schedules.queue_id → queues.id) were added `NOT VALID`, no `ON DELETE` clause ⇒ NO ACTION
  (a delete of a referenced queue fails). Comment: "Existing orphan rows are retained for inspection".
- No FK: attempts.job_id→jobs, attempts.worker_id→workers, jobs.leased_by→workers, webhook_outbox.job_id→jobs,
  job_events.job_id→jobs, workers/schedules/api_keys/queues.namespace_id→namespaces.
#read application enforcement
- `internal/api/queues.go:109-151` deleteQueue: LOCK queues,jobs,schedules SHARE ROW EXCLUSIVE; 404 if missing;
  409 `queue_not_empty` if any job (any state) or schedule; else delete.
- `internal/api/schedules.go:136-143` deleteSchedule: `DELETE … WHERE id AND namespace_id`, **no RowsAffected check**,
  always 204.
- `internal/api/jobs.go:28-38` deleteJob: only terminal states; inserts `deleted` event; attempts rows untouched.
- `internal/api/jobs.go:114-138` deleteWorker: delete row + ReleaseWorker in one tx; notify dead after commit.
- `internal/api/schedules.go:19-63` createSchedule: payload is `json.RawMessage`, stored via `schedule.JSON` =
  `json.Marshal` ⇒ omitted payload becomes JSON `null`; no "must be an object" check (enqueue has one,
  `enqueue.go:64-70`). `fire.go:76` copies schedule payload bytes into jobs **without** the enqueue validation.
  `schedules.go:144-165` runSchedule goes through insertJob ⇒ validated.
- `internal/client/client.go:57` `json.Unmarshal` without UseNumber; `internal/worker/worker.go:21` Payload
  `map[string]any` ⇒ sample worker decodes numbers as float64; echo returns `j.Payload`.
- `internal/api/api.go:53-62` request decode uses UseNumber ⇒ ack `result` keeps digits; `ack.go:11` `json.Marshal(result)`.
- `internal/jobs/jobs.go:85-97` Notify snapshots queue URL/secret + body {job_id,state,result,last_error} at terminal time.
- `internal/webhook/ship.go:22-25` re-marshals stored jsonb body (compact; HTML-escaped) and signs those bytes.
- `internal/jobs/expire.go:90-92` Retain uses the queue's **current** retention_hours for existing terminal jobs.
- Postgres jsonb rejects `\u0000` in strings; `text` columns reject NUL bytes (general PG knowledge, to be observed).

## 15:31Z entity diagram (drawn before probing; also in models/entities.md E06 revision)

```text
namespace 1─┬─N api_key            (no FK; app: auth lookup)
            ├─N queue ─┬─N job ─┬─N attempt        (no FK)
            │          │        ├─N webhook_outbox (no FK; spec: survives job delete)
            │          │        └─N job_events     (no FK; retained by design)
            │          │   FK jobs_queue_owner NOT VALID, NO ACTION + app 409
            │          └─N schedule   FK schedules_queue_owner NOT VALID, NO ACTION + app 409
            ├─N worker ── job.leased_by (no FK; app release on DELETE), attempt.worker_id (kept by spec)
            └─N schedule ──(fires)──▶ job   (payload copied, not validated)
```

CRUD grid (C/R/U/D), cells to be filled with observed outcomes during the run: see models/entities.md.

## 15:31Z hypotheses (stated before any probe; file written before the 15:31:37Z reset/seed — earlier draft headings said 15:33/15:34, corrected to the real clock)

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-C | Charter: "deleting an owner leaves orphans" (no FK with ON DELETE). **My code-reading counter-prediction:** for queue→job/schedule it is refused (409), never cascades and never orphans; orphan *rows* remain only for job→attempts (job delete/purge/retention), which the API cannot reach | charter; 002.sql:9-17; queues.go:109-151 | queue delete returns 204 with dependents, or jobs/schedules with a missing queue appear in psql |
| H-E06-01 | DELETE queue: any job (also only terminal) → 409 `queue_not_empty`; only a schedule → 409; empty → 204; missing → 404; another tenant's queue name → acts on own namespace only | queues.go:109-151 | 204 with dependents; 204 on missing |
| H-E06-02 | DELETE terminal job: GET job 404, GET attempts 404, psql attempts rows remain; events keep `deleted`; pending outbox row still delivered after job delete and even after its queue is deleted | jobs.go:28-38; SPEC "deleting a job does not erase pending deliveries" | outbox row removed/undelivered |
| H-E06-03 | DELETE /v1/schedules/{missing id} → 204; DELETE of another namespace's schedule → 204 and the schedule survives | schedules.go:136-143 | 404 |
| H-E06-04 | Schedule created with payload omitted (or `[1]`, `"x"`) → 201 with payload null/array; `POST …/run` → 400 invalid_job; cron fire → job row with non-object payload | schedules.go:42-60, fire.go:76, enqueue.go:64-70 | create 400, or run 201, or fired payload is `{}` |
| H-E06-05 | Data journey via sample worker: payload `{"big":9007199254740993,…}` on reports (worker-b, webhook) → lease/GET job exact; echo result `9007199254740992`; webhook body carries the lossy result | client.go:57, worker.go:21 | result keeps 9007199254740993 |
| H-E06-06 | Manual ack with result big number → stored and notified exact; key order/whitespace normalized by jsonb (not a defect: RFC 8259 objects unordered) | api.go:53-62, ack.go:11 | digits lost |
| H-E06-07 | Ack result containing `\u0000` → non-2xx, probably 409 state_conflict with driver text, job stays leased; nack error with `\u0000` likewise; enqueue payload with `\u0000` → 400 with driver text | ack.go:10-19, api/jobs.go:88 | 200 |
| H-E06-08 | DELETE worker-a (sample worker's row) while it runs: its leases released; the process logs worker_not_found each second and never re-registers; emails stop being processed | worker.go:34-46 | worker re-registers |
| H-E06-09 | Queue PATCH while dependents exist: max_attempts → existing jobs keep old, new and schedule-fired jobs get new; webhook URL/secret change → pending outbox rows keep old snapshot; retention change → applies to existing terminal jobs at next sweep | SPEC; Notify; expire.go:90-92 | pending row re-targeted / old jobs change max |
| H-E06-10 | Concurrent enqueue vs DELETE queue: never an orphan job in psql; losers see 404 or a 4xx/5xx carrying FK text | queues.go:118 LOCK; 002 FK | orphan job row |
| H-E06-11 | Purge / retention remove jobs but leave attempts rows (orphans) and events; no outbox for purged jobs | queues.go:170, expire.go:90-92 | attempts removed |

## 15:32:06Z P1 queue deletion — #obs (`bash tools/p1-queue-delete.sh`, raw `results/e06/p1.txt`)

Fixture: fresh `tools/reset-e06.sh` at 15:31:37Z (300 seeded jobs, queues 1–4, workers 1–3, schedules 1–2); real clock.

```text
# empty queue → 204                                  [204]
# missing queue                                       {"error":{"code":"queue_not_found",…}} [404]
# globex DELETE /v1/queues/e6-own (acme-only name)    queue_not_found [404]; acme GET e6-own → [200]
# queue with one queued job (job 301)                 {"code":"queue_not_empty",…} [409]
# cancel 301 → [200]; DELETE queue                    queue_not_empty [409]
# DELETE /v1/jobs/301 → [204]; DELETE queue           [204]
# queue with only schedule 3                          queue_not_empty [409]
# DELETE schedule 3 → [204]; DELETE queue             [204]
# psql: orphan_jobs | orphan_schedules | orphan_attempts = 0 | 0 | 0
# events job 301: created, cancelled, deleted (queue "e6-q1"), still readable after queue delete
```
H-E06-01 **confirmed**. H-E06-C (charter "deleting an owner leaves orphans") **refuted** for queue→job and
queue→schedule: the code refuses (409) rather than cascading or orphaning. The charter's premise "no foreign keys"
was already contradicted by reading 002.sql (see #read).

## 15:32:15Z P2 schedules — #obs (`bash tools/p2-schedules.sh`, raw `results/e06/p2.txt`)

```text
# DELETE /v1/schedules/99999                       [204]
# GET / PATCH / run of 99999                       schedule_not_found [404] each
# globex DELETE /v1/schedules/1 (acme's)           [204]; psql: 1 | 1 | daily-report (still there)
# DELETE /v1/schedules/3 again (deleted in P1)     [204]
# POST /v1/schedules {…,"queue":"e6-pay"} with payload omitted → [201] "payload":null (id 4)
#   [1,2] → [201] id 5 ; "x" → [201] id 6 ; null → [201] id 7 ; 7 → [201] id 8
psql: 4 | null ; 5 | [1, 2] ; 6 | "x" ; 7 | null ; 8 | 7    (all enabled, next_run_at 15:33:00)
```
#surprise H-E06-03 **confirmed** — DELETE schedule answers 204 for a missing id and for another tenant's id (no effect);
GET/PATCH/run answer 404 for the same id. Oracle SPEC:15 "A missing resource returns 404" → #bug [E06-F01](../../defects/E06-F01-schedule-delete-missing-204.md).

## 15:32:24Z P2b schedule payload through run and cron fire — #obs

```text
$ for i in 4 5 6; do k acme-key POST /v1/schedules/$i/run; done
{"error":{"code":"invalid_job","message":"payload must be a JSON object"}} [400]   ×3
(waited to 15:33:03Z for the * * * * * fire)
$ sql "select id,kind,state,payload,jsonb_typeof(payload) t,run_at from jobs where queue_id=9 order by id"
302 | echo | queued | null   | null   | 15:33:00
303 | echo | queued | [1, 2] | array  | 15:33:00
304 | echo | queued | "x"    | string | 15:33:00
305 | echo | queued | null   | null   | 15:33:00
306 | echo | queued | 7      | number | 15:33:00
GET /v1/queues/e6-pay/jobs → [302,null] [303,[1,2]] [304,"x"] [305,null] [306,7]
```
(A `jq` exit 5 at the end came from my wrapper's trailing ` [200]` line — tool noise, not SUT output.)
#surprise H-E06-04 **confirmed**: the same schedule datum is rejected by `run` (400) but turned into stored jobs by the
cron path. Jobs with non-object payloads exist although SPEC:60 defines `payload` as a JSON object and enqueue/batch/run
reject them. → #bug [E06-F02](../../defects/E06-F02-schedule-payload-not-validated.md).

## 15:33Z H-E06-12 (stated before P3; written by 15:33:27Z)

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-12 | A cron-fired job with payload `[1,2]` on `emails` (worker-a, lease 10 s, max 3) is leased by worker-a, whose decode of the lease response fails (`Payload map[string]any`); the worker logs a JSON error and never acks/nacks; the lease expires 3× → `dead` with last_error `lease expired`; meanwhile worker-a keeps serving other jobs. A `null` payload decodes to an empty map and is acked with result `null`. | worker.go:21,43-46; client.go:57 | worker acks [1,2] as result, or nacks with a decode error |

## 15:33:35Z P3 cron-fired non-object payload → sample worker — #obs (`bash tools/p3-poison-fire.sh`, raw `results/e06/p3.txt`)

```text
schedules arr=9 (payload [1,2], queue emails) nul=10 (payload omitted); both disabled at 15:34:02Z after one fire
fired jobs: 307,308
15:34:10 307 | leased | 0 | [1, 2] |      |              308 | succeeded | 1 | null | null |
15:34:18 307 | leased | 1 | [1, 2] |      | lease expired
15:34:26 307 | leased | 2 | [1, 2] |      | lease expired
15:34:34 307 | dead   | 3 | [1, 2] |      | lease expired   (unchanged at 15:34:50)
attempts 307: [1,2,"expired","lease expired"] [2,2,…] [3,2,…]     attempts 308: [1,2,"succeeded",""]
$ dc logs --since 15:33:58Z --until 15:34:40Z worker-a
15:34:00.288 "msg":"json: cannot unmarshal array into Go struct field Job.jobs.payload of type map[string]interface {}"
15:34:01.301 processing job 308 attempt 1 → 15:34:01.316 ack status 200
15:34:11.181 same unmarshal message ; 15:34:22.177 same unmarshal message   (no job id in these lines)
```
H-E06-12 **confirmed** in every detail: the fired array payload is a poison job for the documented sample worker —
leased, never processed or nacked, executed 0 times, burned 3 attempts by expiry (≈30 s), `dead` with "lease expired"
rather than a decode error, and the worker log line names no job. Consequence recorded in E06-F02. `null` payload
was delivered as JSON `null` and echoed as result `null`.

## 15:35Z hypotheses before P4 (data journey)

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-05 | (restated) reports job payload `{"big":9007199254740993,…}` → lease/GET exact; worker-b echo result has `9007199254740992`; webhook raw_body carries the lossy number; signature valid | client.go:57 | result exact |
| H-E06-13 | jsonb transforms but preserves values: `1e2`→`100`, `1.10` stays `1.10`, keys reordered `{"a":2,"b":1}`, duplicate key keeps last; `<&>` HTML-escaped as `<&>` in API/webhook bytes (still equal JSON strings) | PG jsonb; encoding/json escapeHTML | a value changes |
| H-E06-06 | (restated) manual ack of a reports job with result `{"big":9007199254740993}` → GET job and webhook body exact | api.go:53-62 | digits lost |
| H-E06-07 | (restated) result containing `\u0000` → 409 state_conflict (driver text), job stays leased; nack error with `\u0000` → 409/500, job stays leased; enqueue payload with `\u0000` → 400 invalid_job with driver text | ack.go:10-19; nack.go | 2xx |

## 15:35:24Z P4 data journey — #obs (`bash tools/p4-journey.sh`, raw `results/e06/p4.txt`)

Payload P = `{"big":9007199254740993,"dec":1.10,"exp":1e2,"html":"<&>","z":{"b":1,"a":2}}` on `reports` (webhook → hooksink).
Job 309 (echo, worker-b), 310 (echo, payload `{"d":1,"d":2}`), 311 (kind manual, my worker 4, acked by curl).

| Hop | 309 (sample worker echo) | 311 (manual ack) |
|---|---|---|
| lease response bytes (311) | — | `"payload":{"z":{"a":2,"b":1},"big":9007199254740993,"dec":1.10,"exp":100,"html":"<&>"}` |
| psql jobs.payload | `{"z": {"a": 2, "b": 1}, "big": 9007199254740993, "dec": 1.10, "exp": 100, "html": "<&>"}` | same |
| psql jobs.result | `{… "big": 9007199254740992, "dec": 1.1, "exp": 100 …}` | `{"big": 9007199254740993, "dec": 1.10}` |
| GET /v1/jobs result | `"big":9007199254740992,"dec":1.1` | `"big":9007199254740993,"dec":1.10` |
| outbox body (psql) | `"big": 9007199254740992, "dec": 1.1` | exact |
| hooksink raw_body (signature_valid true, 200) | `{"state":"succeeded","job_id":309,"result":{"z":{"a":2,"b":1},"big":9007199254740992,"dec":1.1,"exp":100,"html":"<&>"},"last_error":""}` | `{"state":"succeeded","job_id":311,"result":{"big":9007199254740993,"dec":1.10},"last_error":""}` |

Job 310: payload stored and returned as `{"d":2}`, result `{"d":2}`.

- H-E06-05 **confirmed**: the only lossy hop is inside the documented sample worker (client.go:57 float64 decode):
  `9007199254740993` → `9007199254740992`, `1.10` → `1.1`. #surprise (predicted from code) → #bug
  [E06-F03](../../defects/E06-F03-sample-worker-echo-loses-precision.md).
- H-E06-06 **confirmed**: API → storage → webhook keeps result digits exactly.
- H-E06-13 **confirmed**: jsonb normalises (`1e2`→`100`, key order a/b and z-first ordering by key length, duplicate `d` keeps
  last value 2) and Go encoders HTML-escape `<&>` in lease and webhook bytes. **Not a defect:** each value is equal as JSON
  (RFC 8259 §4 object members unordered; §6 numbers are values, `1e2` = `100`; §7 `<` is `<`; §4 duplicate names
  "SHOULD be unique", behaviour then unpredictable) and the SPEC promises only no precision loss and preserved JSON type.
  Signatures were valid over the escaped bytes actually sent.
- Ordering note (not a defect): outbox 70 (job 311) precedes 71 (job 309) because worker-b acked later.

## 15:35Z hypotheses before P5 (written by 15:35:59Z)

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-14 | Sibling of E06-F03: `qctl job get 309`-style output of job 311 shows `big` rounded (qctl decodes through internal/client float64) | client.go:57 | qctl prints 9007199254740993 |
| H-E06-15 | Second sibling: `qctl enqueue … --payload '{"big":9007199254740993}'` stores `9007199254740992` (qctl main.go:52-57 `jsonOpt` decodes into `any`); same for `--result` on ack | cmd/qctl/main.go:52-57 | psql payload keeps …993 |

## 15:36:13Z P5 qctl as a data path — #obs (`bash tools/p5-qctl.sh`, raw `results/e06/p5.txt`; qctl binary inside `queuelet-e06-api-1`)

```text
# qctl job get 311 (psql holds big 9007199254740993, dec 1.10)
    "big": 9007199254740992,   "dec": 1.1,     (payload and result both)
# qctl enqueue reports --kind manual --payload '{"big":9007199254740993,"dec":1.10}'  → id 312 queued
psql 312 | {"big": 9007199254740992, "dec": 1.1}
# curl control, same body → id 313 [201]; psql 313 | {"big": 9007199254740993, "dec": 1.10}
```
H-E06-14 and H-E06-15 **confirmed**. #surprise → #bug [E06-F04](../../defects/E06-F04-qctl-rounds-payload-numbers.md)
(README: "The CLI and the HTTP API expose the same operations"; SPEC:63 precision). Same cause family as E06-F03
(client-side float64 decode) but a distinct surface that **writes** corrupted data.

## 15:36:27Z P6 NUL characters — #obs (`bash tools/p6-nul.sh`, raw `results/e06/p6.txt`; my worker 4, kind manual)

```text
# enqueue payload {"s":"a\u0000b"}          {"code":"invalid_job","message":"pq: unsupported Unicode escape sequence"} [400]
# enqueue kind "man\u0000ual"               {"code":"invalid_job","message":"pq: invalid byte sequence for encoding \"UTF8\": 0x00"} [400]
lease (max 10) → 312, 313, 314 {"case":"ack-nul"}, 315 {"case":"nack-nul"}; 312/313 acked for cleanup
# ack 314 result {"s":"a\u0000b"}           {"code":"state_conflict","message":"pq: unsupported Unicode escape sequence"} [409]
psql 314 | leased | has_token t | attempts 0
# ack 314 again, clean result               {"ok":true} [200]
# nack 315 error "bad\u0000byte"            {"code":"state_conflict","message":"pq: invalid byte sequence for encoding \"UTF8\": 0x00"} [409]
psql 315 | leased | t | 0 | ''
# nack 315 again "bad byte", retry false    [200] → 315 dead, last_error "bad byte"; 314 succeeded
```
H-E06-07 **confirmed** (predicted from code + PostgreSQL knowledge). #surprise: a valid holder with a valid token gets
409 "state_conflict" for a value the store cannot hold; the lease is untouched. Oracle SPEC:16 (409 only for
"incompatible with the current state"; invalid → 400) and E03's resolved rule that storage-limit errors are not
conflicts → #bug [E06-F05](../../defects/E06-F05-nul-in-result-or-error-409-conflict.md). Enqueue's 400 is the right class
but leaks driver text (recorded inside E06-F05 as sibling, not counted separately).

## 15:37Z hypotheses before P7 (worker owner deleted under a running sample worker)

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-08 | (restated, refined) enqueue `sleep` 15 s on emails; while worker-a (row 2) holds it, `DELETE /v1/workers/2` → 200; job → `queued`, attempts 1, attempt row outcome `released` worker 2; after the sleep worker-a's ack → 409 (logged); then every lease → 404 worker_not_found logged ≈1/s; worker-a never re-registers or exits (container stays Up, RestartCount unchanged); the job is never leased again while only worker-a serves emails | worker.go:34-46; api/jobs.go:114-138; expire.go:43-80 | worker re-registers / exits / job processed |

## 15:37:00Z P7 delete a running sample worker's row — #obs (`bash tools/p7-worker-delete.sh`, raw `results/e06/p7.txt`)

```text
worker-a before: RestartCount 0 running ; job 316 (sleep 15 s, emails) leased by 2 at 15:37:00
DELETE /v1/workers/2 → {"ok":true} [200] at 15:37:00Z
psql 316 | queued | attempts 1 | leased_by NULL | no_token t ; attempts [1,2,"released","worker deleted"]
worker-a log: 15:37:00.314 processing 316 attempt 1 → 15:37:15.323 "completed","event":"ack","job_id":316,"status":409,"error":null
             then "HTTP 404: {\"error\":{\"code\":\"worker_not_found\"…}}" once per second (14 lines by 15:37:25)
15:37:25 psql 316 | queued | 1 | NULL ; worker-a after: RestartCount 0 running ; workers table: 1,3,4 (no new row)
emails stats: queued 1, oldest_queued_age 25.5 s, leased 0
```
H-E06-08 **confirmed**. The release side holds (attempt counted, `released` outcome, fields cleared — lesson-05 fix,
regression only). The dependent *process* is a zombie: it keeps polling with a deleted id forever, never exits
(`restart: on-failure` never fires) and never re-registers, so job 316 is stranded while the container looks healthy;
the executed sleep's ack is logged as `completed … status 409, error null`.
#surprise judged **not a defect**: SPEC defines `DELETE /v1/workers/{id}` only from the server side ("deregisters the
worker and releases its active leases"); it states no client obligation and the sample worker's behaviour after
deregistration is unspecified. Saved as question Q-E06-1 (should the sample worker exit non-zero or re-register on
`worker_not_found`?) and testability note (the `completed` log line reports `error:null` on a 409).
Recovery 15:37:53Z: `dc restart worker-a` → new row 5 `worker-a`; job 316 leased by 5 at 15:37:58 (attempt 2). Restored.

## 15:38Z hypotheses before P8 (queue PATCH / delete with dependents in flight)

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-09a | queue e6-patch max 3: job A queued; PATCH max_attempts 5; job B → A keeps 3, B 5; schedule run job 5 | SPEC:46; enqueue.go:43 | A changes to 5 |
| H-E06-09b | queue e6-hook webhook {hooksink, secret report-secret}; hooksink 500; cancel job C → outbox row fails; PATCH secret "changed" and url `/other`; hooksink ok → row for C delivered to `/hook` with signature_valid true (old snapshot); a later cancelled job D → `/other`, signature_valid false | SPEC snapshots; jobs.go:85-97 | C re-targeted or re-signed |
| H-E06-09c | PATCH webhook url "" → job E cancelled produces no outbox row; still-pending rows unaffected | SPEC "empty URL disables future snapshots" | outbox row for E |
| H-E06-09d | PATCH retention_hours 0 on e6-hook → its terminal jobs removed by the next sweep (≤2 s); attempts rows of removed jobs remain in psql; events `retained`; outbox rows remain | expire.go:90-92 | jobs remain |
| H-E06-02 | (restated) with hooksink 500 and a pending row for a job of e6-del, DELETE that job, then DELETE queue e6-del → 204; after hooksink ok the row is still delivered (body names deleted job and queue) | SPEC snapshot rule; queues.go:139 | 409 or row dropped |

## 15:38:20Z P8 queue PATCH / delete with dependents in flight — #obs (`bash tools/p8-patch-delete.sh`, raw `results/e06/p8.txt`)

```text
a) e6-patch: job 317 before PATCH max_attempts 5; 318 after; 319 via schedule run  → 317|3  318|5  319|5
b) hooksink mode 500; e6-hook {hook, report-secret}; cancel 320 → outbox 77 | url …/hook | report-secret | attempts 1 | 500
   PATCH webhook {url …/other, secret changed} → "webhook_secret":"changed","webhook_url":"http://hooksink:8201/other"
   cancel 321 → outbox 78 | …/other | changed
c) PATCH webhook {"url":""} → url "" (secret kept "changed"); cancel 322 → no outbox row for 322
e) e6-del {hook, report-secret}; cancel 323 → row 79 pending; DELETE /v1/jobs/323 [204]; DELETE /v1/queues/e6-del [204]
   15:38:23 mode ok → 15:38:31: 77 | …/hook | 2 | 200 | t ; 78 | …/other | 2 | 200 | t ; 79 | …/hook | 1 | 200 | t
   hooksink: true {"state":"cancelled","job_id":323,…}  true {…"job_id":320…}  false {…"job_id":321…}
d) e6-hook jobs 3; PATCH retention_hours 0; 15:38:34 jobs 0; job_events retained for 320, 321, 322; outbox rows 2 remain;
   GET /v1/jobs/320, 321 → job_not_found [404]; DELETE /v1/queues/e6-hook → [204]
```
H-E06-09a, b, c, d **confirmed**; H-E06-02 **confirmed** (row for deleted job 323 of deleted queue e6-del delivered).
Snapshot semantics hold exactly: the pending row keeps its URL and secret (signature valid with the old secret); the
row created after the PATCH carries the new secret (hooksink verifies with `report-secret`, hence `false` — expected).
Retention change applies retroactively to already-terminal jobs (`expire.go:90-92` uses the current setting) — SPEC:122
"Terminal jobs are deleted after retention_hours" names no snapshot, so **not a defect**; recorded as rule R-E06-4.

## 15:39Z hypotheses before P9

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-10 | (restated) 30 rounds × (16 parallel enqueues ‖ DELETE queue): zero orphan jobs in psql; outcomes per round are either delete 409 + all enqueues 201, or delete 204 + enqueues 404, with occasional enqueue 400/500 carrying FK text when the enqueue looked up the queue before the delete committed | queues.go:118 LOCK; enqueue.go:74; 002 FK | an orphan job row; or only clean 201/404 outcomes |
| H-E06-11 | (restated) a job with one failed attempt (retry_wait) purged → job gone, its attempts row remains in psql (unreachable through API: 404), event `purged` | queues.go:170 | attempts row removed |

## 15:39:03Z P9 race, run 1 — #obs (`bash tools/p9-race.sh`, raw `results/e06/p9.txt`, per-request files `results/e06/p9-race/`)

```text
# delete status counts     30 409
# enqueue status counts   480 201
orphan_jobs | orphan_schedules = 0 | 0
```
Dead end: enqueues always won (delete always saw a job). Vary one thing: issue the DELETE first and stagger the
enqueues by 0–9.9 ms (`tools/p9b-race.sh`). Prediction unchanged (H-E06-10).

## 15:39:55Z P9b race, DELETE first, enqueues staggered 0–9.9 ms — #obs (`bash tools/p9b-race.sh`, raw `results/e06/p9b.txt`, `results/e06/p9b-race/`)

```text
# delete status counts     26 204 ;  4 409
# enqueue status counts    64 201 ;  1 400 ; 415 404
# non-201/404 bodies       1 {"error":{"code":"invalid_job","message":"pq: insert or update on table \"jobs\" violates foreign key constraint \"jobs_queue_owner\""}}
# rounds with delete 204 and an enqueue 201: none
orphan_jobs | orphan_schedules = 0 | 0
```
H-E06-10 **confirmed** (both halves): ownership holds under the race — the FK plus the table lock never let an orphan
in — but a request that looked up the queue before the delete committed is answered **400 invalid_job** with driver
text, although the request was valid and the resource is now missing (SPEC:15 → 404). #surprise → #bug
[E06-F06](../../defects/E06-F06-enqueue-racing-queue-delete-400-fk-text.md).

## 15:40Z sibling predictions before P9c (E06-F06 siblings; stated before probing)

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-16 | `POST /v1/schedules` racing DELETE queue: some creates get **500 database_error** with `schedules_queue_owner` FK text; zero orphan schedules | schedules.go:42-60 | only 201/404 |
| H-E06-17 | `jobs:batch` racing DELETE queue: some batches 400 invalid_job with FK text, nothing partially stored | enqueue.go:92-124 | only 201/404 |

## 15:40:33Z P9c sibling races — #obs (`bash tools/p9c-race-siblings.sh`, raw `results/e06/p9c.txt`, `results/e06/p9c-race/`)

```text
# delete           39 204 ;   1 409
# schedule create   8 201 ; 307 404 ; 5 500
# batch             8 201 ; 312 404
# unusual bodies    5 500 {"error":{"code":"database_error","message":"pq: insert or update on table \"schedules\" violates foreign key constraint \"schedules_queue_owner\""}}
orphan_jobs | orphan_schedules = 0 | 0
```
H-E06-16 **confirmed** (sibling of E06-F06, added to that report). H-E06-17 **refuted by its stated condition** in this
sample: 320 batches saw only 201/404 (the window exists in code, enqueue.go:92-124 after queueVerb's lookup, but was
not hit in 40 rounds; recorded as sampling uncertainty, not as proof of absence).

## 15:41Z hypotheses before P10

| Id | Prediction | Source | Would be wrong if |
|---|---|---|---|
| H-E06-11 | (restated) purge of a retry_wait job with 1 failed attempt: purge `deleted:1`; GET job and attempts 404; psql attempts row remains; event `purged` | queues.go:170 | row removed |
| H-E06-18 | DELETE of a dead job with 2 attempts and retention removal of a succeeded job leave their attempts rows; the deleted job's id is not reused, so rows are unreachable via API | jobs.go:28-38; expire.go:90-92; bigserial | rows removed / id reused |
| H-E06-19 | Same idempotency key after the original job is deleted → new job id (key freed with the row); SPEC "returns the original job" cannot hold for a deleted original | enqueue.go:74; unique index | old id returned |

## 15:41:24Z P10 residue after purge / DELETE / retention — #obs (`bash tools/p10-residue.sh`, raw `results/e06/p10.txt`)

Queue e6-res (max 2, linear base 0), worker 4.
```text
## purge   885 | retry_wait | 1 → purge {"deleted":1} [200]; GET job 404; GET attempts 404
           psql attempts: 885 | 1 | failed | secret-ish detail ; events: created, leased, retry_wait, purged
## DELETE  886 | dead | 2 → DELETE [204]; psql attempts 886 n1 "fail 1", n2 "fail 2"
## key K1 again after delete → {"id":887,"state":"queued"} [201]
## retain  887 acked; PATCH retention_hours 0; 2.5 s later GET 887 → job_not_found; psql attempts 887 | 1 | succeeded
## totals  orphan_attempts 4 over 3 jobs ; outbox rows for missing jobs 3 (all delivered)
```
H-E06-11, H-E06-18 **confirmed**; H-E06-19 **confirmed** (new id 887 for the freed key — not a defect: the original
job no longer exists, SPEC:68 "returns the original job" has nothing to return; recorded as rule R-E06-6).
#surprise (charter prediction, now observed): job → attempt is the one owner/dependent pair that orphans; rows keep
worker error text after the job's retention → #bug [E06-F07](../../defects/E06-F07-attempts-outlive-their-job.md)
(approximation oracle, low).

## 15:41:43Z sweeper/event log path — #obs

```text
$ dc logs --since 15:41:20Z api | grep job_id 885|886|887   (trimmed to the event fields)
885 created → leased(worker 4, attempt 1) → retry_wait(error "secret-ish detail") → purged(worker null, attempt 1)
886 created → leased → retry_wait("fail 1") → queued → leased(attempt 2) → dead("fail 2") → deleted(attempt 2)
887 created → leased → succeeded → retained(15:41:28.115Z)
$ dc logs --since 15:31:30Z api | grep -v '"job event"'  → only "queuelet listening on :8200" (no background-task errors)
```
The fourth read path (API event console) agrees with `/v1/jobs/{id}/events` and psql for every removal kind; the loop
logged no error during any probe (races, NUL, poison payloads).

## Surprise ledger (every #surprise ends in a report or a reason)

| Surprise | Outcome |
|---|---|
| 002.sql adds FKs (contradicts charter premise) | not a defect — protective; recorded in entities model |
| DELETE schedule 204 on missing/foreign id | E06-F01 |
| non-object schedule payload fired into jobs; sample worker poisoned | E06-F02 |
| sample worker echo rounds numbers | E06-F03 |
| qctl rounds numbers on input and display | E06-F04 |
| NUL in ack result / nack error → 409 state_conflict | E06-F05 |
| race losers get 400/500 with FK text | E06-F06 |
| attempts rows outlive job | E06-F07 |
| jsonb normalisation, HTML escaping, duplicate keys | not a defect — value-equal JSON (RFC 8259 §4, §6, §7); see P4 |
| deleted worker's sample process becomes a zombie | not a defect — no client oracle in SPEC; Q-E06-1 |
| retention change applies to already-terminal jobs | not a defect — SPEC:122 has no snapshot rule; R-E06-4 |
| idempotency key reusable after original deleted | not a defect — original gone; R-E06-6 |
| webhook row for job 321 signature_valid false | expected — new secret "changed" vs hooksink's fixed `report-secret` |

## 15:45Z correction note (#read)

While writing reports I verified every `file:line` citation against the export and corrected the ones I had written
from memory during the run (e.g. `schedules.go:150-157` → `:136-143`, `SPEC:12` → `SPEC:15`, `fire.go:77` → `:76`,
`worker.go:19` → `:21`, `expire.go:140` → `:90-92`). Corrections were applied in place in this file and the E06 reports;
no observation, output or prediction text was changed.

## 15:45Z close of probing

Probing ended 15:41:43Z. No clock control was used (real UTC throughout). Stack left running until the record is written.

## 15:46:48Z final snapshot and stop — #obs

`results/e06/final-snapshot.txt`: queues 8, jobs 319, schedules 10, workers 4, attempts 174, outbox 79, events 2567;
orphan_jobs 0, orphan_schedules 0, orphan_attempts 4, outbox rows for missing jobs 3; non-object payload jobs 7; health ok.
15:47:26Z `dc stop` (project `queuelet-e06` only; volume kept for inspection). Other queuelet stacks (default, e01) were
never touched and still running.
