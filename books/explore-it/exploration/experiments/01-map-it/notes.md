# Experiment 01 — map-it — notes

Build: export of `explore-it/v0-inherited` at /tmp/explore-it-experiment-01 (no git metadata).
Explorer: Claude Opus 5.5, headless, single agent. Start of run: 2026-10-22T01:38:30Z (first `date -u`).

Charter (copied): Explore queuelet end to end with the source tree, qctl, README and SPEC.md to discover
what it does, what it promises, where the spec is silent, and what I predict the code does there.
Charter prediction **H00**: "the README's happy path works and the spec is complete".
Refuted if a documented verb misbehaves or a spec sentence has no code behind it.
Planned split: 20 reading / 30 probing / 10 writing.

## 2026-10-22T01:38:42Z #read — phase start: reading
Read charter, PROTOCOL.md, SKILL map-codebase-from-behavior, README.md, SPEC.md, lesson 00,
exploration scaffolding (models/README, defects/README, hypotheses.md, brief.md). No probes yet.

## 2026-10-22T01:40:26Z #read — source walk (all of cmd/, internal/, migrations, lab/, deploy/, test/checks)
Source pointers below are code-reading *predictions* of paths, not observations.
- cmd/api/main.go:21-31 applies migrations/001.sql then starts `app.Loop` + HTTP on :8200.
- internal/api/api.go:92-129 router; auth = sha256(bearer) lookup in api_keys; /health,/metrics bypass auth.
- api.go:131-146 `Loop`: 1 s ticker running Promote, Expire, schedule.Fire, Retain, webhook.Ship in sequence;
  **on any step error it logs and `return`s — loop ends for the life of the process** (SPEC: "should recover after a DB restart").
- api.go:80-87 `decode` uses plain json.Decoder: no DisallowUnknownFields (SPEC: unknown fields rejected).
- queues.go:273 create validates name non-empty, lease>=1, max>=1, base>=0 only; no name charset check, no backoff.kind enum check.
- queues.go:299-320 PATCH: no validation of max_attempts / retention; unknown fields ignored.
- queues.go:323-330 DELETE queue: unconditional delete by name, always 204 (no 404, no nonempty 409, jobs/schedules left orphaned).
- queues.go:346-353 purge deletes only state='queued' (not scheduled/retry_wait). SPEC "pending" is undefined → silence.
- enqueue.go:12-20 payload decoded into map[string]any → float64 numbers (precision risk >2^53). priority not range-checked (0–9).
- enqueue.go:80-82 batch >500 silently truncated to 500; empty batch → 201 {"ids":[]}; insertJob per job without a transaction → partial insert on a mid-batch error.
- enqueue.go:100-105 invalid state filter silently ignored; limit>500 → 100.
- jobs.go(api):127-146 getJob / attempts / deleteJob query by id **without namespace filter**; jobVerb cancel/retry likewise (jobs/operator.go).
- jobs.go(api):248-250 lease with unknown worker → `fail(w, 200, "worker_not_found")` — error body with HTTP 200.
- jobs/ack.go:10 Ack: WHERE id AND state='leased' — **lease_token not checked**.
- jobs/nack.go:176-188 Nack: **neither token nor state checked**; dead iff `attempts+1 > max` (so max=N allows N+1 executions?).
- jobs/expire.go:4-6 Expire: requeues expired leases without incrementing attempts or writing an attempts row.
- jobs/operator.go:12-14 Retry: no state filter (succeeded/cancelled/queued all "retryable").
- jobs/jobs.go:59 LeaseOrder "priority DESC, created_at"; migrations: created_at is `timestamp(0)` (1-second resolution) → FIFO ties within a second.
- jobs/jobs.go:104-123 Notify: result serialized as a JSON *string*, truncated to 1024 bytes; last_error truncated to 256 bytes; no truncation flag.
- webhook/ship.go:236-257 Ship signs `MarshalIndent` (pretty) bytes but sends compact `Marshal` bytes; marks delivered on *any* HTTP response (500 included); retries only transport errors (+5 s).
- schedule/fire.go:137 Next uses a FixedZone with the offset captured at create (schedules.zone_offset) → DST changes ignored.
- schedule/fire.go:148 Fire selects by next_run_at only — no `enabled` filter; paused queue: `continue` without advancing → catch-up loop at fire.go:161 on resume.
- schedules.go(api):75-90 PATCH updates cron/timezone/enabled columns only — next_run_at and zone_offset not recomputed, cron/timezone not validated.
- stats/stats.go:232 folds scheduled+retry_wait into "queued" (SPEC: counts by exact state).
- cmd/qctl/main.go:16-18 errors printed to stderr but process never exits non-zero.
- cmd/hooksink/main.go: verifies HMAC over raw body, records signature_valid; modes ok/500/slow/timeout.
- deploy/compose.yml: fixed project name `queuelet`, image tag `queuelet:${QUEUELET_IMAGE_TAG:-local}`.

## 2026-10-22T01:40:26Z #read — lab decision (environment)
`docker compose ls` shows a project named `queuelet` already running from a compose file **outside this export**
(path not read). Its build identity vs. this export is unknown, and `make up` here would rebuild the shared
`queuelet:local` image under that project name. Decision: run an isolated stack from this export, same mechanism
lab/verify.sh uses for REF builds: COMPOSE_PROJECT_NAME=queuelet-e01, API 18220, hooksink 18221, PG 15453,
QUEUELET_IMAGE_TAG=e01. The pre-existing `queuelet` project is not touched. Build log: project/results/e01-up.log.

## 2026-10-22T01:40:50Z #hypothesis — stated before any probe (source: R=recon/README, S=spec silence, C=code reading)
Ranked by consequence × uncertainty; probe family in [brackets]. Predictions are what I think the *code* does.
| Id | Rank | Source | Prediction | Refuted if | Family |
|---|---|---|---|---|---|
| H00 | – | charter | README happy path works and the spec is complete | a documented verb misbehaves or a spec sentence has no code behind it | recon |
| H01 | 3 | R, qctl README block | queue ls / enqueue / jobs ls / worker create / lease / job ack / get / attempts / stats all succeed on seed | any step returns error | sequence |
| H02 | 1 | SPEC "Only the current lease holder may complete", C ack.go:10 | ack with a wrong lease_token on a leased job → 200 and job succeeded | 409 state_conflict | oracle |
| H03 | 1 | same, C nack.go:176 | nack with wrong token on a *queued (never leased)* job → 200, job moves to retry_wait, attempts=1 | 409 or no state change | oracle/state |
| H04 | 1 | SPEC "namespace key accesses only its namespace", C api/jobs.go:128 | globex-key GET /v1/jobs/{acme id} → 200 with payload; globex-key cancel acme job → 200 | 404/403 | boundary |
| H05 | 2 | SPEC "No job may execute more than its maximum attempts", C nack.go:181 | queue max_attempts=1: lease+nack(retry) → retry_wait (not dead); a second lease is possible | first nack gives dead | sequence |
| H06 | 2 | SPEC "expired lease records an unsuccessful attempt", C expire.go:5 | after lease expiry the job is queued with attempts=0 and no attempts row | attempts row exists or attempts=1 | state |
| H07 | 2 | SPEC "cancelled job is never run", C nack.go | cancel a leased job, then nack it with its token → 200, job back in retry_wait (revived) | nack → 409 / stays cancelled | sequence |
| H08 | 2 | SPEC "Succeeded and cancelled jobs cannot be retried", C operator.go:13 | retry on succeeded job → 200, state queued | 409 | state |
| H09 | 2 | SPEC "next fire must reflect current configuration", C api/schedules.go:84 | PATCH schedule cron → next_run_at unchanged; PATCH enabled=false → next_run_at unchanged (and still fires) | next_run_at changes | state |
| H10 | 2 | SPEC "should recover after a database restart"/loop, C api.go:140-142 + schedules PATCH no cron validation | PATCH cron to "bogus" on a due schedule → 200; next Fire errors and the whole background loop stops (a delay_seconds job never promotes) | 400 on patch, or promotion continues | ecosystem |
| H11 | 2 | SPEC webhook signature, C ship.go:237-245 | hooksink records signature_valid=false for every delivery | signature_valid=true | oracle |
| H12 | 2 | SPEC "successful only on 2xx", C ship.go:257 | hooksink mode 500 → outbox row delivered_at set after 1 POST, no retry | attempts>1 / delivered_at null | ecosystem |
| H13 | 3 | SPEC "Notification data must agree with stored data", C jobs.go:114-121 | webhook body.result is a JSON string, not the stored object; >1024 B result truncated silently | body.result is an object | oracle |
| H14 | 2 | SPEC "Callers must handle a conflict for a nonempty queue", C queues.go:324 | DELETE nonempty queue → 204; its jobs still readable via GET /v1/jobs/{id}; DELETE missing queue → 204 | 409 / 404 | entity |
| H15 | 3 | SPEC batch "empty or oversized … 400 without insertion", C enqueue.go:80 | 501-job batch → 201 with 500 ids; empty → 201 ids [] ; mid-batch invalid → 400 but earlier jobs inserted | 400 and no insertion | variation |
| H16 | 3 | SPEC "without loss of numeric precision", C enqueue.go:14 | payload {"n":9007199254740993} reads back 9007199254740992 | exact readback | variation |
| H17 | 3 | SPEC "counts by exact state", C stats.go:232 | stats.queued = DB queued+scheduled+retry_wait | equals DB queued | oracle |
| H18 | 3 | SPEC "Unknown fields and invalid enum values rejected", C api.go:81 | queue create with unknown field → 201; backoff.kind "zigzag" → 201; jobs?state=bogus → 200 all jobs | 400 | variation |
| H19 | 4 | SPEC name charset / priority 0-9, C queues.go:273, enqueue.go:30 | queue name "a.b" → 201; priority 42 → 201 | 400 | variation |
| H20 | 3 | SPEC "missing resource returns 404", C api/jobs.go:249 | lease with unknown worker_id → HTTP 200 + error body | 404 | oracle |
| H21 | 4 | R qctl, C qctl/main.go:16 | qctl on HTTP 404 prints "HTTP 404" on stderr and exits 0 | exit≠0 | console |
| H22 | 3 | SPEC "jobs run FIFO by creation order", C jobs.go:59 + created_at timestamp(0) | 10 same-priority jobs enqueued in one second may lease out of id order | needs many trials; lease order==id order would not refute (ties) | sequence |
| H23 | 3 | SPEC "Missed periodic ticks should not create an unbounded burst", C fire.go:157-161 | schedule "* * * * *" on paused queue, wait ≥2 min, resume → ≥2 jobs inserted in one tick | exactly 1 job | sequence/clock |
| H24 | 4 | SPEC purge "removes all pending jobs" (silence: what is pending?), C queues.go:347 | purge leaves scheduled and retry_wait jobs | scheduled jobs deleted | variation |
| H25 | 4 | SPEC "DST offset changes", C fire.go:137 | zone offset frozen at create (schedules.zone_offset) — not probeable without clock control | — | testability |
| H26 | 3 | SPEC loop "recover after DB restart", C api.go:142 | restart postgres → background loop dead; delayed job never promoted until API restart | promotion resumes | ecosystem |

### 2026-10-22T01:41:03Z #obs fixture: make seed equivalent (lab/reset.sh with queuelet-e01 env); actors acme-key; clock = host UTC

### 2026-10-22T01:41:03Z #obs seed
```
$ bash lab/reset.sh
 Container queuelet-e01-worker-b-1 Stopping 
 Container queuelet-e01-worker-a-1 Stopping 
 Container queuelet-e01-worker-b-1 Stopped 
 Container queuelet-e01-worker-a-1 Stopped 
 Container queuelet-e01-api-1 Stopping 
 Container queuelet-e01-api-1 Stopped 
 Container queuelet-e01-postgres-1 Waiting 
 Container queuelet-e01-postgres-1 Healthy 
 Container queuelet-e01-api-1 Starting 
 Container queuelet-e01-api-1 Started 
Seeded 300 jobs, 4 queues, 2 schedules; results/manifest.json
 Container queuelet-e01-postgres-1 Waiting 
 Container queuelet-e01-postgres-1 Healthy 
 Container queuelet-e01-worker-b-1 Starting 
 Container queuelet-e01-worker-a-1 Starting 
 Container queuelet-e01-worker-b-1 Started 
 Container queuelet-e01-worker-a-1 Started 
[exit 0]
```

### 2026-10-22T01:41:17Z #obs manifest
```
$ python3 -c "import json;m=json.load(open('results/manifest.json'));print([ (q['id'],q['name'],q['max_attempts'],q['lease_seconds'],q['webhook_url']) for q in m['queues']], 'jobs',len(m['job_ids']),m['job_ids'][:3],m['job_ids'][-1],'worker',m['seed_worker'],'sched',[(s['id'],s['name'],s['next_run_at'],s['zone_offset']) for s in m['schedules']])"
[(1, 'emails', 3, 10, ''), (2, 'reports', 4, 20, 'http://hooksink:8201/hook'), (3, 'imports', 2, 30, '')] jobs 300 [1, 2, 3] 300 worker 1 sched [(1, 'daily-report', '2026-10-22T13:00:00Z', -14400), (2, 'import-pulse', '2026-10-22T01:45:00Z', 0)]
[exit 0]
```

### 2026-10-22T01:41:28Z #obs dead end: `go build -o results/qctl ./cmd/qctl` failed: 'error obtaining VCS status: exit status 128 / Use -buildvcs=false'. Retrying with -buildvcs=false (same source).

### 2026-10-22T01:41:31Z #obs H01 README happy path via qctl (results/qctl built from export with -buildvcs=false)

### 2026-10-22T01:41:31Z #obs queue ls
```
$ results/qctl queue ls | jq -c '.queues[]|{id,name,max_attempts,paused}'
{"id":1,"name":"emails","max_attempts":3,"paused":false}
{"id":2,"name":"reports","max_attempts":4,"paused":false}
{"id":3,"name":"imports","max_attempts":2,"paused":false}
[exit 0]
```

### 2026-10-22T01:41:31Z #obs enqueue
```
$ results/qctl enqueue imports --kind echo --payload '{"message":"hello"}'
{
  "id": 301,
  "state": "queued"
}
[exit 0]
```

### 2026-10-22T01:41:32Z #obs jobs ls
```
$ results/qctl jobs ls imports --state queued --limit 3 | jq -c '.jobs[]|{id,state,priority,created_at}', '.next_cursor'
jq: error: syntax error, unexpected end of file (Unix shell quoting issues?) at <top-level>, line 1:
.jobs[]|{id,state,priority,created_at},                                      
jq: 1 compile error
[exit 3]
```

### 2026-10-22T01:41:32Z #obs worker create
```
$ results/qctl worker create manual --kinds echo
{
  "token": "406245a4d197c628de4a575c54bf45cc",
  "worker_id": 4
}
[exit 0]
```

### 2026-10-22T01:41:38Z #obs jobs ls (retry; jq quoting fixed)
```
$ results/qctl jobs ls imports --state queued --limit 3 | jq -c '(.jobs[]|{id,state,priority,created_at}), {next_cursor}'
{"id":3,"state":"queued","priority":2,"created_at":"2026-10-22T01:41:08Z"}
{"id":6,"state":"queued","priority":5,"created_at":"2026-10-22T01:41:08Z"}
{"id":12,"state":"queued","priority":1,"created_at":"2026-10-22T01:41:08Z"}
{"next_cursor":12}
[exit 0]
```

### 2026-10-22T01:41:38Z #obs lease
```
$ results/qctl lease imports --worker-id 4 --max-jobs 1
{
  "jobs": [
    {
      "attempt": 1,
      "id": 27,
      "kind": "echo",
      "lease_expires_at": "2026-10-22T01:42:08.290984257Z",
      "lease_token": "aa62446ef937e46b87deb20a26185122",
      "payload": {
        "fixture": 26
      }
    }
  ]
}
[exit 0]
```

### 2026-10-22T01:41:38Z #obs second lease captured to results/h01-lease2.json: id=87

### 2026-10-22T01:41:38Z #obs ack
```
$ results/qctl job ack 87 --lease-token 8037d23c3ea5cd0a2a1b7745d774f441 --result '{"done":true}'
{
  "ok": true
}
[exit 0]
```

### 2026-10-22T01:41:38Z #obs get
```
$ results/qctl job get 87 | jq -c '{id,state,attempts,result,finished_at}'
{"id":87,"state":"succeeded","attempts":1,"result":{"done":true},"finished_at":"2026-10-22T01:41:38.481753Z"}
[exit 0]
```

### 2026-10-22T01:41:38Z #obs attempts
```
$ results/qctl job attempts 87 | jq -c .
{"attempts":[{"error":"","finished_at":"2026-10-22T01:41:38.481753Z","job_id":87,"n":1,"outcome":"succeeded","started_at":"2026-10-22T01:41:38.481753Z","worker_id":4}]}
[exit 0]
```

### 2026-10-22T01:41:38Z #obs stats
```
$ results/qctl stats imports | jq -c .
{"cancelled":6,"dead":6,"leased":1,"oldest_queued_age":30.896924,"queued":80,"retry_wait":6,"scheduled":25,"succeeded":8,"throughput_last_hour":8}
[exit 0]
```

### 2026-10-22T01:41:46Z #obs H17 DB exact counts for imports (queue_id=3)
```
$ db "SELECT state,count(*) FROM jobs WHERE queue_id=3 GROUP BY state ORDER BY 1"
cancelled|6
dead|6
leased|1
queued|49
retry_wait|6
scheduled|25
succeeded|8
[exit 0]
```

### 2026-10-22T01:41:46Z #obs H01 side-check: queued imports with priority > 6 at time of lease (should be none if priority order held)
```
$ db "SELECT id,priority,state FROM jobs WHERE queue_id=3 AND state='queued' AND priority>6 ORDER BY priority DESC,id LIMIT 5"

[exit 0]
```

### 2026-10-22T01:41:54Z #obs H02 — ack with wrong lease token. Setup: fresh queue h02 (acme), worker 4 (kinds echo)

### 2026-10-22T01:41:54Z #obs create queue
```
$ c acme-key -X POST http://localhost:18220/v1/queues -d '{"name":"h02"}' | jq -c '{id,name,max_attempts}' 2>/dev/null || true
{"id":5,"name":"h02","max_attempts":3}
[exit 0]
```

### 2026-10-22T01:41:54Z #obs enqueue
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h02/jobs -d '{"kind":"echo","payload":{"a":1}}'
{"id":302,"state":"queued"}
 [HTTP 201]
[exit 0]
```

### 2026-10-22T01:41:54Z #obs lease
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h02/lease -d '{"worker_id":4,"max_jobs":1}'
{"jobs":[{"attempt":1,"id":302,"kind":"echo","lease_expires_at":"2026-10-22T01:42:24.914717093Z","lease_token":"75de57dbdfaf6346f912050ae63bcea5","payload":{"a":1}}]}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:41:55Z #obs ack with WRONG token
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/302/ack -d '{"lease_token":"not-the-token","result":{"by":"impostor"}}'
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:41:55Z #obs job state after
```
$ db "SELECT id,state,attempts,result FROM jobs WHERE id=302"
302|succeeded|1|{"by": "impostor"}
[exit 0]
```

### 2026-10-22T01:42:02Z #obs H03 — nack with wrong token on a queued (never leased) job, queue h02 (max_attempts 3)

### 2026-10-22T01:42:02Z enqueued kind=other id=303 (kind not in worker 4's kinds, so nobody leases it)

### 2026-10-22T01:42:02Z #obs before
```
$ db "SELECT id,state,attempts FROM jobs WHERE id=303"
303|queued|0
[exit 0]
```

### 2026-10-22T01:42:02Z #obs nack wrong token, never leased
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/303/nack -d '{"lease_token":"x","error":"spoof"}'
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:02Z #obs after
```
$ db "SELECT id,state,attempts,last_error,run_at>now() AS future FROM jobs WHERE id=303; SELECT job_id,n,worker_id,outcome,error FROM attempts WHERE job_id=303"
303|retry_wait|1|spoof|t
303|1||failed|spoof
[exit 0]
```

### 2026-10-22T01:42:03Z #obs H08 — retry a succeeded job (302 from H02)

### 2026-10-22T01:42:03Z #obs retry succeeded 302
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/302/retry
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:03Z #obs after
```
$ db "SELECT id,state,attempts,result,finished_at FROM jobs WHERE id=302"
302|queued|0|{"by": "impostor"}|
[exit 0]
```

### 2026-10-22T01:42:13Z #hypothesis H27 (source: surprise from H03 + nack.go:176-188). Sibling of H03: nack on a *succeeded* job → 200, state retry_wait, attempts 2, result kept; then promoted to queued (run again). Refuted if 409.

### 2026-10-22T01:42:13Z #obs H07 — cancel a leased job, then holder nacks with its real token (queue h02)

### 2026-10-22T01:42:13Z enqueued 304; lease returned id=302 token=9f29f22a3ea82498ceb8a7bf7537ddd0

### 2026-10-22T01:42:13Z #obs cancel leased
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/302/cancel
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:13Z #obs state
```
$ db "SELECT id,state,attempts FROM jobs WHERE id=302"
302|cancelled|0
[exit 0]
```

### 2026-10-22T01:42:13Z #obs holder nacks (retry default)
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/302/nack -d '{"lease_token":"9f29f22a3ea82498ceb8a7bf7537ddd0","error":"worker noticed failure","delay_seconds":1}'
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:16Z #obs state 2.5s later (Promote runs every 1s)
```
$ db "SELECT id,state,attempts,finished_at FROM jobs WHERE id=302"
302|queued|1|
[exit 0]
```

### 2026-10-22T01:42:16Z #obs can it be leased again?
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h02/lease -d '{"worker_id":4,"max_jobs":5}' | head -c 300
{"jobs":[{"attempt":2,"id":302,"kind":"echo","lease_expires_at":"2026-10-22T01:42:46.669001221Z","lease_token":"bf35ddfabe8cb41523a256a7b4edf5f0","payload":{"a":1}},{"attempt":1,"id":304,"kind":"echo","lease_expires_at":"2026-10-22T01:42:46.669001221Z","lease_token":"ebfe267fdba67ccc65caeed8d40023ec
[exit 0]
```

### 2026-10-22T01:42:16Z #obs H27 — nack on succeeded job 87 (acked in H01) with bogus token

### 2026-10-22T01:42:16Z #obs before
```
$ db "SELECT id,state,attempts FROM jobs WHERE id=87"
87|succeeded|1
[exit 0]
```

### 2026-10-22T01:42:16Z #obs nack succeeded job
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/87/nack -d '{"lease_token":"x","error":"late"}'
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:17Z #obs after
```
$ db "SELECT id,state,attempts,result,last_error FROM jobs WHERE id=87"
87|retry_wait|2|{"done": true}|late
[exit 0]
```

### 2026-10-22T01:42:24Z #obs H04 — globex-key against acme job ids (acme job 1 = emails fixture 0, payload {fixture:0})

### 2026-10-22T01:42:24Z #obs globex GET acme job 1
```
$ c globex-key http://localhost:18220/v1/jobs/1 | head -c 400
{"attempts":0,"created_at":"2026-10-22T01:41:08Z","finished_at":null,"id":1,"idempotency_key":null,"kind":"echo","last_error":"","lease_expires_at":null,"lease_token":null,"leased_by":null,"max_attempts":3,"namespace_id":1,"payload":{"fixture":0},"priority":0,"queue_id":1,"result":null,"run_at":"2026-10-22T02:41:07.800247Z","state":"scheduled"}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:24Z #obs globex attempts acme job 87
```
$ c globex-key http://localhost:18220/v1/jobs/87/attempts
{"attempts":[{"error":"","finished_at":"2026-10-22T01:41:38.481753Z","job_id":87,"n":1,"outcome":"succeeded","started_at":"2026-10-22T01:41:38.481753Z","worker_id":4},{"error":"late","finished_at":"2026-10-22T01:42:16.9861Z","job_id":87,"n":2,"outcome":"failed","started_at":"2026-10-22T01:42:16.9861Z","worker_id":4}]}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:24Z acme enqueued 305 in h02

### 2026-10-22T01:42:24Z #obs globex cancel acme job
```
$ c globex-key -X POST http://localhost:18220/v1/jobs/305/cancel
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:25Z #obs globex delete acme (now cancelled) job
```
$ c globex-key -X DELETE http://localhost:18220/v1/jobs/305
 [HTTP 204]
[exit 0]
```

### 2026-10-22T01:42:25Z #obs acme reads it
```
$ c acme-key http://localhost:18220/v1/jobs/305
{"error":{"code":"job_not_found","message":"sql: no rows in result set"}}
 [HTTP 404]
[exit 0]
```

### 2026-10-22T01:42:25Z #obs control: globex GET acme queue
```
$ c globex-key http://localhost:18220/v1/queues/h02
{"error":{"code":"queue_not_found","message":"sql: no rows in result set"}}
 [HTTP 404]
[exit 0]
```

### 2026-10-22T01:42:25Z #obs control: globex GET acme schedule 1
```
$ c globex-key http://localhost:18220/v1/schedules/1
{"error":{"code":"schedule_not_found","message":"sql: no rows in result set"}}
 [HTTP 404]
[exit 0]
```

### 2026-10-22T01:42:34Z #hypothesis H28 (sibling of H04, source nack.go has no namespace or token filter): globex-key nack of an acme queued job → 200 and state retry_wait. Refuted if 404/409.

### 2026-10-22T01:42:34Z #obs H28 globex nack acme job 306
```
$ c globex-key -X POST http://localhost:18220/v1/jobs/306/nack -d '{"error":"from globex","retry":false}'; echo; db "SELECT id,namespace_id,state,last_error FROM jobs WHERE id=306"
{"ok":true}
 [HTTP 200]
306|1|dead|from globex
[exit 0]
```

### 2026-10-22T01:42:34Z #obs H05 — max_attempts=1 queue, nack with retry true via the real token. Queue h05 lease_seconds=2 backoff base 0

### 2026-10-22T01:42:34Z #obs create h05
```
$ c acme-key -X POST http://localhost:18220/v1/queues -d '{"name":"h05","max_attempts":1,"lease_seconds":2,"backoff":{"kind":"linear","base_seconds":0}}' | head -c 200
{"backoff_base":0,"backoff_kind":"linear","backoff_max":null,"id":6,"lease_seconds":2,"max_attempts":1,"name":"h05","namespace_id":1,"paused":false,"retention_hours":24,"webhook_secret":"","webhook_ur
[exit 0]
```

### 2026-10-22T01:42:35Z job 307 leased: {"jobs":[{"attempt":1,"id":307,"kind":"echo","lease_expires_at":"2026-10-22T01:42:37.068318551Z","lease_token":"daa1ba9907e3d669eb00591281bf88d3","payload":{}}]

### 2026-10-22T01:42:35Z #obs nack #1 (max_attempts 1)
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/307/nack -d '{"lease_token":"daa1ba9907e3d669eb00591281bf88d3","error":"e1"}'; echo; db "SELECT id,state,attempts,max_attempts FROM jobs WHERE id=307"
{"ok":true}
 [HTTP 200]
307|queued|1|1
[exit 0]
```

### 2026-10-22T01:42:37Z #obs lease again
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h05/lease -d '{"worker_id":4}' | head -c 200
{"jobs":[{"attempt":2,"id":307,"kind":"echo","lease_expires_at":"2026-10-22T01:42:39.227871179Z","lease_token":"d85b5644f160665715bab0b0723235ae","payload":{}}]}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:42:37Z #obs attempts table
```
$ db "SELECT job_id,n,outcome,error FROM attempts WHERE job_id=307"
307|1|failed|e1
[exit 0]
```

### 2026-10-22T01:42:44Z #obs H06 — job 307 (h05, max_attempts 1, attempts 1) was leased as attempt 2 with lease_seconds=2; do nothing, wait 4 s

### 2026-10-22T01:42:48Z #obs after expiry
```
$ db "SELECT id,state,attempts,lease_token IS NULL AS no_token FROM jobs WHERE id=307; SELECT count(*) AS attempt_rows FROM attempts WHERE job_id=307"
307|queued|1|t
1
[exit 0]
```

### 2026-10-22T01:42:49Z #obs lease a third time
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h05/lease -d '{"worker_id":4}' | head -c 120
{"jobs":[{"attempt":2,"id":307,"kind":"echo","lease_expires_at":"2026-10-22T01:42:51.122482912Z","lease_token":"4f95b928
[exit 0]
```

### 2026-10-22T01:42:53Z #obs expire again; lease a 4th time
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h05/lease -d '{"worker_id":4}' | head -c 120; echo; db "SELECT id,state,attempts FROM jobs WHERE id=307; SELECT count(*) FROM attempts WHERE job_id=307"
{"jobs":[{"attempt":2,"id":307,"kind":"echo","lease_expires_at":"2026-10-22T01:42:55.19615937Z","lease_token":"bbf9b829a
307|leased|1
1
[exit 0]
```

### 2026-10-22T01:43:03Z #obs H11/H13 — webhook signature and body. Queue h11 with webhook hooksink (secret report-secret = hooksink WEBHOOK_SECRET). Records cleared first.

### 2026-10-22T01:43:03Z #obs existing hooksink records from seed/worker-b (before clearing)
```
$ curl -s http://localhost:18221/admin/records | jq -c '[.records[]|.signature_valid]|group_by(.)|map({v:.[0],n:length})'
[{"v":false,"n":50}]
[exit 0]
```

### 2026-10-22T01:43:04Z #obs ack 308 small result
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/308/ack -d '{"lease_token":"881f8a362dfd7f89c6fbce5986124e17","result":{"ok":true,"n":2}}'
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:43:04Z #obs ack 309 1.5KB result
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/309/ack -d '{"lease_token":"5a6b33d4fb1bf184d1c8ec4553d74245","result":{"blob": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", "n": 1}}'
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:43:07Z #obs hooksink records
```
$ curl -s http://localhost:18221/admin/records | jq -c '.records[]|{sig:.signature_valid,status,job:.body.job_id,state:.body.state,result_type:(.body.result|type),result_len:(.body.result|tostring|length),result_head:(.body.result|tostring|.[0:40]),err:.body.last_error}'
{"sig":false,"status":200,"job":308,"state":"succeeded","result_type":"string","result_len":20,"result_head":"{\"n\": 2, \"ok\": true}","err":""}
{"sig":false,"status":200,"job":309,"state":"succeeded","result_type":"string","result_len":1024,"result_head":"{\"n\": 1, \"blob\": \"xxxxxxxxxxxxxxxxxxxxxx","err":""}
[exit 0]
```

### 2026-10-22T01:43:07Z #obs stored result length for 309
```
$ db "SELECT id,length(result::text) FROM jobs WHERE id=309"
309|1520
[exit 0]
```

### 2026-10-22T01:43:19Z #question hooksink records signature_valid but not the received X-Queuelet-Signature header, so I cannot tell *which* bytes were signed from hooksink alone. Testability request TR1 below.

### 2026-10-22T01:43:19Z #obs H12 — hooksink mode 500; ack one h11 job; watch outbox

### 2026-10-22T01:43:19Z #obs set mode 500
```
$ curl -s -X POST -H 'Content-Type: application/json' -d '{"mode":"500"}' http://localhost:18221/admin/mode
{"mode":"500"}
[exit 0]
```

### 2026-10-22T01:43:20Z #obs ack 310
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/310/ack -d '{"lease_token":"2a4b5fe7b5de32db0e0cd3fd1d50a836","result":{"r":5}}'
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:43:28Z #obs hooksink records after 8 s
```
$ curl -s http://localhost:18221/admin/records | jq -c '.records[]|{status,job:.body.job_id,at}'
{"status":500,"job":310,"at":"2026-10-22T01:43:20.499537012Z"}
[exit 0]
```

### 2026-10-22T01:43:28Z #obs outbox row
```
$ db "SELECT id,job_id,attempts,delivered_at IS NOT NULL AS delivered FROM webhook_outbox WHERE job_id=310"
72|310|1|t
[exit 0]
```

### 2026-10-22T01:43:28Z #obs api log line
```
$ docker logs queuelet-e01-api-1 2>&1 | grep -c 'webhook delivered status=500'
1
[exit 0]
```

### 2026-10-22T01:43:28Z #obs restore mode ok
```
$ curl -s -X POST -H 'Content-Type: application/json' -d '{"mode":"ok"}' http://localhost:18221/admin/mode
{"mode":"ok"}
[exit 0]
```

### 2026-10-22T01:43:35Z #obs H14 — DELETE nonempty queue h02 (has jobs 302-306) and a missing queue

### 2026-10-22T01:43:35Z #obs h02 job count
```
$ db "SELECT count(*) FROM jobs WHERE queue_id=5"
4
[exit 0]
```

### 2026-10-22T01:43:35Z #obs DELETE h02
```
$ c acme-key -X DELETE http://localhost:18220/v1/queues/h02
 [HTTP 204]
[exit 0]
```

### 2026-10-22T01:43:35Z #obs orphans
```
$ db "SELECT count(*) FROM jobs j LEFT JOIN queues q ON q.id=j.queue_id WHERE q.id IS NULL"
4
[exit 0]
```

### 2026-10-22T01:43:35Z #obs orphan still readable
```
$ c acme-key http://localhost:18220/v1/jobs/304 | jq -c '{id,queue_id,state}' 2>/dev/null; c acme-key http://localhost:18220/v1/jobs/304 | tail -c 12
{"id":304,"queue_id":5,"state":"queued"}

 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:43:36Z #obs DELETE missing queue
```
$ c acme-key -X DELETE http://localhost:18220/v1/queues/never-existed
 [HTTP 204]
[exit 0]
```

### 2026-10-22T01:43:36Z #obs H14b — schedule on deleted queue: create schedule on queue h14s, delete queue, run schedule

### 2026-10-22T01:43:36Z #obs delete queue h14s with schedule 3
```
$ c acme-key -X DELETE http://localhost:18220/v1/queues/h14s; echo; c acme-key http://localhost:18220/v1/schedules/3 | jq -c '{id,queue_id,next_run_at}' 2>/dev/null; c acme-key -X POST http://localhost:18220/v1/schedules/3/run
 [HTTP 204]
{"id":3,"queue_id":8,"next_run_at":"2027-01-01T00:00:00Z"}
{"error":{"code":"queue_not_found","message":"sql: no rows in result set"}}
 [HTTP 404]
[exit 0]
```

### 2026-10-22T01:43:44Z #obs H15 — batch variations on fresh queue h15

### 2026-10-22T01:43:44Z #obs 501-job batch
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h15/jobs:batch --data @results/h15-501.json | jq -c '{n:(.ids|length)}' 2>/dev/null; db "SELECT count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h15')"
{"n":500}
500
[exit 0]
```

### 2026-10-22T01:43:46Z #obs empty batch
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h15/jobs:batch -d '{"jobs":[]}'
{"ids":[]}
 [HTTP 201]
[exit 0]
```

### 2026-10-22T01:43:46Z #obs batch with invalid 3rd job
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h15/jobs:batch -d '{"jobs":[{"kind":"ok1"},{"kind":"ok2"},{"payload":{}},{"kind":"ok4"}]}'; echo; db "SELECT kind,count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h15') AND kind<>'x' GROUP BY kind ORDER BY 1"
{"error":{"code":"invalid_job","message":"kind required"}}
 [HTTP 400]
ok1|1
ok2|1
[exit 0]
```

### 2026-10-22T01:43:46Z #obs H16 — numeric precision

### 2026-10-22T01:43:46Z #obs stored jsonb
```
$ db "SELECT payload FROM jobs WHERE id=813"
{"big": 9007199254740992, "dec": 0.1, "neg": -12345678901234567000}
[exit 0]
```

### 2026-10-22T01:43:46Z #obs GET job payload (raw)
```
$ curl -s -H 'Authorization: Bearer acme-key' http://localhost:18220/v1/jobs/813 | grep -o '"payload":{[^}]*}'
"payload":{"big":9007199254740992,"dec":0.1,"neg":-12345678901234567000}
[exit 0]
```

### 2026-10-22T01:43:55Z #obs H18/H19 — validation variations

### 2026-10-22T01:43:55Z #obs unknown field on queue create
```
$ c acme-key -X POST http://localhost:18220/v1/queues -d '{"name":"h18a","bogus":1}' | tail -c 14
"}
 [HTTP 201]
[exit 0]
```

### 2026-10-22T01:43:55Z #obs invalid backoff.kind enum
```
$ c acme-key -X POST http://localhost:18220/v1/queues -d '{"name":"h18b","backoff":{"kind":"zigzag","base_seconds":1}}' | grep -o '"backoff_kind":"[a-z]*"\|HTTP [0-9]*'
"backoff_kind":"zigzag"
HTTP 201
[exit 0]
```

### 2026-10-22T01:43:55Z #obs unknown field on enqueue
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h15/jobs -d '{"kind":"x","priorty":9}'
{"id":814,"state":"queued"}
 [HTTP 201]
[exit 0]
```

### 2026-10-22T01:43:55Z #obs invalid state filter
```
$ c acme-key 'http://localhost:18220/v1/queues/imports/jobs?state=bogus&limit=500' | jq -c '{n:(.jobs|length),states:([.jobs[].state]|unique)}' 2>/dev/null; c acme-key 'http://localhost:18220/v1/queues/imports/jobs?state=bogus' -o /dev/null
{"n":101,"states":["cancelled","dead","queued","retry_wait","scheduled","succeeded"]}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:43:55Z #obs limit 1000
```
$ c acme-key 'http://localhost:18220/v1/queues/imports/jobs?limit=1000' | jq -c '{n:(.jobs|length)}' 2>/dev/null
{"n":100}
[exit 5]
```

### 2026-10-22T01:43:55Z #obs queue name a.b and ü
```
$ c acme-key -X POST http://localhost:18220/v1/queues -d '{"name":"a.b"}' | tail -c 12; c acme-key -X POST http://localhost:18220/v1/queues -d '{"name":"ü-q"}' | tail -c 12

 [HTTP 201]
 [HTTP 201]
[exit 0]
```

### 2026-10-22T01:43:55Z #obs priority 42 and -3
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h15/jobs -d '{"kind":"x","priority":42}'; c acme-key -X POST http://localhost:18220/v1/queues/h15/jobs -d '{"kind":"x","priority":-3}'
{"id":815,"state":"queued"}
 [HTTP 201]{"id":816,"state":"queued"}
 [HTTP 201]
[exit 0]
```

### 2026-10-22T01:43:55Z #obs PATCH queue max_attempts 0 / -1, retention -5
```
$ c acme-key -X PATCH http://localhost:18220/v1/queues/h15 -d '{"max_attempts":-1,"retention_hours":-5}' | jq -c '{max_attempts,retention_hours}' 2>/dev/null; c acme-key -X POST http://localhost:18220/v1/queues/h15/jobs -d '{"kind":"x"}'
{"max_attempts":-1,"retention_hours":-5}
{"error":{"code":"invalid_job","message":"max_attempts must be positive"}}
 [HTTP 400]
[exit 0]
```

### 2026-10-22T01:44:07Z note: previous '[exit 5]' on limit=1000 is jq failing on the ' [HTTP 200]' suffix my curl -w adds after it printed {"n":100}; not a SUT signal.

### 2026-10-22T01:44:07Z #obs H20/H21 — lease with unknown worker; qctl exit status

### 2026-10-22T01:44:07Z #obs lease unknown worker 9999 (curl)
```
$ c acme-key -X POST http://localhost:18220/v1/queues/imports/lease -d '{"worker_id":9999}'
{"error":{"code":"worker_not_found","message":"unknown worker"}}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:44:07Z #obs qctl lease unknown worker; exit status of qctl
```
$ results/qctl lease imports --worker-id 9999; echo "qctl_exit=$?"
{
  "error": {
    "code": "worker_not_found",
    "message": "unknown worker"
  }
}
qctl_exit=0
[exit 0]
```

### 2026-10-22T01:44:07Z #obs qctl job get missing; exit status
```
$ results/qctl job get 99999999; echo "qctl_exit=$?"
HTTP 404: {"error":{"code":"job_not_found","message":"sql: no rows in result set"}}

qctl_exit=0
[exit 0]
```

### 2026-10-22T01:44:07Z #obs qctl unknown verb; exit status
```
$ results/qctl queue frob x; echo "qctl_exit=$?"
unknown queue verb
qctl_exit=0
[exit 0]
```

### 2026-10-22T01:44:07Z #obs globex worker id used on acme queue (worker 4 is acme) via globex key
```
$ c globex-key -X POST http://localhost:18220/v1/queues/emails/lease -d '{"worker_id":4}'
{"error":{"code":"worker_not_found","message":"unknown worker"}}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:44:08Z #obs H24 — purge semantic: queue h24 with 1 queued, 1 scheduled (delay 3600), 1 retry_wait

### 2026-10-22T01:44:08Z #obs before purge
```
$ db "SELECT kind,state FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h24') ORDER BY id"
q|queued
s|scheduled
r|retry_wait
[exit 0]
```

### 2026-10-22T01:44:08Z #obs purge
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h24/purge; echo; db "SELECT kind,state FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h24') ORDER BY id"
{"deleted":1}
 [HTTP 200]
s|scheduled
r|retry_wait
[exit 0]
```

### 2026-10-22T01:44:20Z #obs H23 setup — queue h23 paused, schedule cron '* * * * *' UTC on it. Clock: host UTC 01:44:20.

### 2026-10-22T01:44:21Z #obs pause h23
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h23/pause
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:44:21Z #obs create schedule
```
$ c acme-key -X POST http://localhost:18220/v1/schedules -d '{"name":"h23-every-minute","queue":"h23","kind":"echo","cron":"* * * * *"}' | jq -c '{id,next_run_at,enabled}' 2>/dev/null
{"id":4,"next_run_at":"2026-10-22T01:45:00Z","enabled":true}
[exit 5]
```

### 2026-10-22T01:44:21Z #obs H09 — PATCH schedule effects. Schedule yearly '0 0 1 1 *' on imports, then PATCH cron to '* * * * *'

### 2026-10-22T01:44:21Z #obs before
```
$ db "SELECT id,cron,timezone,zone_offset,enabled,next_run_at FROM schedules WHERE id=5"
5|0 0 1 1 *|UTC|0|t|2027-01-01 00:00:00+00
[exit 0]
```

### 2026-10-22T01:44:21Z #obs PATCH cron * * * * * and timezone Asia/Tokyo
```
$ c acme-key -X PATCH http://localhost:18220/v1/schedules/5 -d '{"cron":"* * * * *","timezone":"Asia/Tokyo"}' | jq -c '{cron,timezone,zone_offset,next_run_at}' 2>/dev/null
{"cron":"* * * * *","timezone":"Asia/Tokyo","zone_offset":0,"next_run_at":"2027-01-01T00:00:00Z"}
[exit 5]
```

### 2026-10-22T01:44:21Z #obs H09b — disabled-by-PATCH schedule: create '* * * * *' on imports kind h09b, PATCH enabled=false, check if it still fires in next minute

### 2026-10-22T01:44:21Z #obs PATCH enabled=false
```
$ c acme-key -X PATCH http://localhost:18220/v1/schedules/6 -d '{"enabled":false}' | jq -c '{id,enabled,next_run_at}' 2>/dev/null
{"id":6,"enabled":false,"next_run_at":"2026-10-22T01:45:00Z"}
[exit 5]
```

### 2026-10-22T01:44:21Z #obs H09c — created disabled, then PATCH enabled=true: predict next_run_at stays null → never fires

### 2026-10-22T01:44:21Z #obs PATCH enabled=true
```
$ c acme-key -X PATCH http://localhost:18220/v1/schedules/7 -d '{"enabled":true}' | jq -c '{id,enabled,next_run_at}' 2>/dev/null
{"id":7,"enabled":true,"next_run_at":null}
[exit 5]
```

### 2026-10-22T01:45:05Z #obs after the 01:45 tick: schedules state
```
$ db "SELECT id,name,enabled,next_run_at,last_run_at FROM schedules ORDER BY id"
1|daily-report|t|2026-10-22 13:00:00+00|
2|import-pulse|t|2026-10-22 02:00:00+00|2026-10-22 01:45:00+00
3|orph|t|2027-01-01 00:00:00+00|
4|h23-every-minute|t|2026-10-22 01:45:00+00|
5|h09|t|2027-01-01 00:00:00+00|
6|h09b|f|2026-10-22 01:46:00+00|2026-10-22 01:45:00+00
7|h09c|t||
[exit 0]
```

### 2026-10-22T01:45:05Z #obs jobs created by schedules (kinds h09b/h09c, queue h23)
```
$ db "SELECT j.kind,q.name,j.state,count(*) FROM jobs j JOIN queues q ON q.id=j.queue_id WHERE j.kind IN ('h09b','h09c','h09kind') OR q.name='h23' GROUP BY 1,2,3"
h09b|imports|queued|1
[exit 0]
```

### 2026-10-22T01:45:17Z #obs H09b/H09c outcome is in the previous block: disabled schedule 6 (h09b) fired at 01:45 and advanced; schedule 7 (h09c, created disabled then enabled) has next_run_at NULL. H23: schedule 4 next_run_at frozen at 01:45 while h23 is paused.

### 2026-10-22T01:45:17Z #obs H22 — FIFO within a priority. Queue h22; enqueue 40 echo jobs priority 5 back-to-back; lease one at a time; compare lease order to id order

### 2026-10-22T01:45:18Z #obs created_at distribution
```
$ db "SELECT created_at,count(*),min(id),max(id) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h22') GROUP BY 1 ORDER BY 1"
2026-10-22 01:45:17|8|822|829
2026-10-22 01:45:18|32|830|861
[exit 0]
```

### 2026-10-22T01:45:20Z lease order of payload.i (enqueue order 1..40): 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40

### 2026-10-22T01:45:20Z #obs inversions count
```
$ python3 -c "o=[int(x) for x in ' 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40'.split()];print('inversions',sum(1 for a in range(len(o)) for b in range(a+1,len(o)) if o[a]>o[b]))"
inversions 0
[exit 0]
```

### 2026-10-22T01:45:35Z #hypothesis H22b (source: H22 inconclusive + created_at timestamp(0), heap-order tie-break inference). Variation: enqueue 10 same-priority jobs within one second; lease job #1 with a temp worker, DELETE that worker (releases lease → UPDATE moves tuple). Predict: next leases return #2..#10 before #1. Refuted if #1 leases first.

### 2026-10-22T01:45:35Z #obs created_at groups
```
$ db "SELECT created_at,count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h22b') GROUP BY 1 ORDER BY 1"
2026-10-22 01:45:35|5
2026-10-22 01:45:36|5
[exit 0]
```

### 2026-10-22T01:45:35Z #obs temp worker 5 leases first job
```
$ curl -s -H 'Authorization: Bearer acme-key' -X POST http://localhost:18220/v1/queues/h22b/lease -d '{"worker_id":5}' | jq -c '.jobs[]|{id,i:.payload.i}'
{"id":863,"i":2}
[exit 0]
```

### 2026-10-22T01:45:35Z #obs delete worker 5 (release)
```
$ c acme-key -X DELETE http://localhost:18220/v1/workers/5; echo; db "SELECT id,state,payload->>'i' FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h22b') AND payload->>'i'='1'"
{"ok":true}
 [HTTP 200]
862|queued|1
[exit 0]
```

### 2026-10-22T01:45:36Z H22b lease order of payload.i after release: 3 4 5 2 1 6 7 8 9 10

### 2026-10-22T01:45:48Z #obs H22b outcome: temp worker's first lease returned i=2 (id 863), not i=1 (id 862), both created_at 01:45:35; after release, order '3 4 5 2 1 6 7 8 9 10' — FIFO-by-creation violated among same-second ties even before the release. Inversion is from tie-breaking, not the release alone.

### 2026-10-22T01:46:01Z #hypothesis H11b (source: ship.go:236-245, H11 observation). With a capture listener (results/tools/sigcap.py on docker gateway 172.22.0.1:18229, reachability checked via wget from hooksink container), the received X-Queuelet-Signature will equal HMAC-SHA256(secret, json.MarshalIndent(body,'','  ')) and NOT HMAC over the received bytes. Refuted if it matches the received bytes.

### 2026-10-22T01:46:05Z #obs H11b compare
```
$ python3 -I results/tools/sigcheck.py results/sigcap.jsonl s3cret
raw_body: {"job_id":873,"last_error":"","result":"{\"k\": 1}","state":"succeeded"}
sig received : a793b82d82c48b3d26281be3c77f91db8595f80e99cfe49e8de6d137b25bc685
hmac(raw)    : 09cea41e33eac6606b16950657014e2e540fcbe203406f0d4d64c32110948f66 False
hmac(pretty) : a793b82d82c48b3d26281be3c77f91db8595f80e99cfe49e8de6d137b25bc685 True
[exit 0]
```

### 2026-10-22T01:46:08Z #obs H23 pre-resume: schedule 4 and h23 jobs
```
$ db "SELECT next_run_at,last_run_at FROM schedules WHERE id=4; SELECT count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h23')"
2026-10-22 01:45:00+00|
0
[exit 0]
```

### 2026-10-22T01:46:19Z #hypothesis H29-H33 (code reading, expected-correct paths, walked while H23 waits): H29 idempotency_key repeat → same id both times (refuted if ids differ); H30 /v1/namespaces/globex/stats: acme-key 403, admin-key 200 (refuted otherwise); H31 heartbeat with wrong token → 409 (jobs.go:126 checks token; refuted if 200); H32 no auth → 401, /health 200 JSON, /metrics text 'queuelet_up 1'; H33 DELETE queued job → 409.

### 2026-10-22T01:46:19Z #obs H29 idempotency x2
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h29/jobs -d '{"kind":"echo","idempotency_key":"k1"}'; c acme-key -X POST http://localhost:18220/v1/queues/h29/jobs -d '{"kind":"other","idempotency_key":"k1"}'
{"id":874,"state":"queued"}
 [HTTP 201]{"id":874,"state":"queued"}
 [HTTP 201]
[exit 0]
```

### 2026-10-22T01:46:19Z #obs H30 ns stats
```
$ c acme-key http://localhost:18220/v1/namespaces/globex/stats; echo; c admin-key http://localhost:18220/v1/namespaces/globex/stats; echo; c globex-key http://localhost:18220/v1/namespaces/nope/stats
{"error":{"code":"forbidden","message":"namespace mismatch"}}
 [HTTP 403]
{"cancelled":0,"dead":0,"leased":0,"oldest_queued_age":0,"queued":0,"retry_wait":0,"scheduled":0,"succeeded":0,"throughput_last_hour":0}
 [HTTP 200]
{"error":{"code":"namespace_not_found","message":"sql: no rows in result set"}}
 [HTTP 404]
[exit 0]
```

### 2026-10-22T01:46:19Z #obs H31 heartbeat wrong token on leased 874
```
$ c acme-key -X POST http://localhost:18220/v1/jobs/874/heartbeat -d '{"lease_token":"nope","extend_seconds":60}'
{"error":{"code":"state_conflict","message":"state conflict"}}
 [HTTP 409]
[exit 0]
```

### 2026-10-22T01:46:19Z #obs H32 auth/health/metrics
```
$ c '' http://localhost:18220/v1/queues; echo; curl -s -w ' [HTTP %{http_code}]' http://localhost:18220/health; curl -s -w ' [HTTP %{http_code}]' http://localhost:18220/metrics
{"error":{"code":"unauthorized","message":"invalid API key"}}
 [HTTP 401]
{"status":"ok"}
 [HTTP 200]queuelet_up 1
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:46:19Z #obs H33 DELETE queued job 876
```
$ c acme-key -X DELETE http://localhost:18220/v1/jobs/876
{"error":{"code":"state_conflict","message":"state conflict"}}
 [HTTP 409]
[exit 0]
```

### 2026-10-22T01:48:04Z #obs H23 — resuming h23 at 01:48:04 after ticks 01:45, 01:46, 01:47, 01:48 were missed while paused. Prediction: 4 jobs appear within ~1 s (one per missed tick).

### 2026-10-22T01:48:04Z #obs resume h23
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h23/resume
{"ok":true}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:48:06Z #obs h23 jobs after resume
```
$ db "SELECT id,run_at,created_at,state FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h23') ORDER BY id; SELECT next_run_at,last_run_at FROM schedules WHERE id=4"
879|2026-10-22 01:45:00+00|2026-10-22 01:48:05|queued
880|2026-10-22 01:46:00+00|2026-10-22 01:48:05|queued
881|2026-10-22 01:47:00+00|2026-10-22 01:48:05|queued
882|2026-10-22 01:48:00+00|2026-10-22 01:48:05|queued
2026-10-22 01:49:00+00|2026-10-22 01:48:00+00
[exit 0]
```

### 2026-10-22T01:48:18Z #hypothesis H34 (source: api.go:139 Loop runs Ship inline; ship.go:246 2 s client timeout per row, up to 20 rows per tick). With hooksink mode 'timeout' and 6 undelivered outbox rows, a delay_seconds=1 job will take ≥10 s to be promoted (loop stalled by webhook I/O). Refuted if it is promoted within 3 s.

### 2026-10-22T01:48:32Z #obs H34: hooksink mode=timeout, 6 acks on h11 → 6 outbox rows; delayed job 889 (delay 1 s) became queued after 13.3 s

### 2026-10-22T01:48:32Z #obs outbox for h11 timeout rows
```
$ db "SELECT count(*),max(attempts) FROM webhook_outbox WHERE delivered_at IS NULL"
6|1
[exit 0]
```

### 2026-10-22T01:48:43Z #obs outbox drained after mode ok?
```
$ db "SELECT count(*) FROM webhook_outbox WHERE delivered_at IS NULL"
0
[exit 0]
```

### 2026-10-22T01:48:45Z #obs H34 control: hooksink mode=ok, outbox empty → delayed job 890 (delay 1 s) queued after 2.1 s

### 2026-10-22T01:48:56Z #obs H10 — PATCH schedule 4 (h23, '* * * * *') cron to 'bogus'. Prediction: PATCH 200; at the next minute the background loop logs an error and stops; a delay_seconds=1 job then never promotes.

### 2026-10-22T01:48:56Z #obs PATCH cron bogus
```
$ c acme-key -X PATCH http://localhost:18220/v1/schedules/4 -d '{"cron":"bogus"}' | jq -c '{id,cron,next_run_at}' 2>/dev/null
{"id":4,"cron":"bogus","next_run_at":"2026-10-22T01:49:00Z"}
[exit 5]
```

### 2026-10-22T01:49:03Z #obs api log after the minute tick
```
$ docker logs --since 2m queuelet-e01-api-1 2>&1 | grep -v 'webhook delivered' | tail -5
2026/10/22 01:49:00 background task: expected exactly 5 fields, found 1: [bogus]
[exit 0]
```

### 2026-10-22T01:49:11Z #obs delayed job 892 8 s later; schedule 4
```
$ db "SELECT id,state,run_at<now() AS due FROM jobs WHERE id=892; SELECT id,cron,next_run_at,last_run_at FROM schedules WHERE id=4"
892|scheduled|t
4|bogus|2026-10-22 01:49:00+00|2026-10-22 01:48:00+00
[exit 0]
```

### 2026-10-22T01:49:11Z #obs API request path still serving?
```
$ c acme-key http://localhost:18220/v1/queues/h34 | tail -c 14; echo; curl -s http://localhost:18220/health
"}
 [HTTP 200]
{"status":"ok"}
[exit 0]
```

### 2026-10-22T01:49:18Z #hypothesis H10b (source: fire.go:164-168 insert precedes Next; H10 observation). Restarting the API container does not recover: first tick re-inserts an h23 job and the loop dies again; job 892 stays scheduled. Refuted if 892 becomes queued after restart.

### 2026-10-22T01:49:18Z #obs h23 job count before restart
```
$ db "SELECT count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h23')"
5
[exit 0]
```

### 2026-10-22T01:49:18Z #obs restart api (isolated project)
```
$ docker compose -f deploy/compose.yml restart api 2>&1 | tail -2; QUEUELET_URL=http://localhost:18220 bash lab/wait.sh; echo wait=$?
 Container queuelet-e01-api-1 Restarting 
 Container queuelet-e01-api-1 Started 
curl: (56) Recv failure: Connection reset by peer
wait=0
[exit 0]
```

### 2026-10-22T01:49:26Z #obs after restart
```
$ db "SELECT id,state FROM jobs WHERE id=892; SELECT count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h23')"; docker logs --since 15s queuelet-e01-api-1 2>&1 | tail -3
892|queued
6
2026/10/22 01:49:20 queuelet listening on :8200
2026/10/22 01:49:21 background task: expected exactly 5 fields, found 1: [bogus]
[exit 0]
```

### 2026-10-22T01:49:26Z #obs recover: DELETE schedule 4, restart api again
```
$ c acme-key -X DELETE http://localhost:18220/v1/schedules/4; docker compose -f deploy/compose.yml restart api 2>&1 | tail -1; QUEUELET_URL=http://localhost:18220 bash lab/wait.sh; sleep 4; db "SELECT id,state FROM jobs WHERE id=892"
 [HTTP 204] Container queuelet-e01-api-1 Started 
curl: (56) Recv failure: Connection reset by peer
892|queued
[exit 0]
```

### 2026-10-22T01:49:42Z #surprise H10b refuted as stated: job 892 WAS promoted after the restart. Reason found in api.go:139 step order: Promote and Expire run before Fire in each tick, so the first tick after a restart promotes, then Fire inserts one more h23 job (count 5→6) and the loop dies again (log 01:49:21). Recovery requires deleting the poisoned schedule. My code-reading prediction ignored step order.

### 2026-10-22T01:49:46Z #obs post-recovery control: delayed job 895 after 4 s; recent api log
```
$ db "SELECT id,state FROM jobs WHERE id=895"; docker logs --since 10s queuelet-e01-api-1 2>&1 | grep -c 'background task'
895|queued
0
[exit 1]
```

### 2026-10-22T01:49:46Z #hypothesis H26 (source: api.go:140-142; SPEC 'should recover after a database restart'). Restart postgres container; while DB is down a Loop step errors → loop returns. After DB is back, request path works (database/sql reconnects) but a delayed job is never promoted. Refuted if promotion resumes within 10 s of DB return.

### 2026-10-22T01:49:46Z #obs restart postgres (isolated project)
```
$ docker compose -f deploy/compose.yml restart postgres 2>&1 | tail -1; for i in $(seq 1 30); do pg_isready -h localhost -p 15453 -q && break; sleep 1; done; echo pg_ready_after=$i
 Container queuelet-e01-postgres-1 Started 
pg_ready_after=2
[exit 0]
```

### 2026-10-22T01:50:01Z #obs after DB restart
```
$ db "SELECT id,state FROM jobs WHERE id=896"; docker logs --since 40s queuelet-e01-api-1 2>&1 | grep 'background task' | tail -2
896|scheduled
2026/10/22 01:49:21 background task: expected exactly 5 fields, found 1: [bogus]
2026/10/22 01:49:47 background task: dial tcp 172.22.0.2:5432: connect: connection refused
[exit 0]
```

### 2026-10-22T01:50:24Z #obs H26 outcome above: enqueue (request path) worked after DB restart (id 896 returned), but 896 still 'scheduled' 10 s after due; log 'background task: dial tcp 172.22.0.2:5432: connect: connection refused' then nothing. Loop dead. Recovering lab: restart api container of queuelet-e01.

### 2026-10-22T01:50:24Z #obs lab recovery: restart api
```
$ docker compose -f deploy/compose.yml restart api 2>&1 | tail -1; QUEUELET_URL=http://localhost:18220 bash lab/wait.sh 2>/dev/null; sleep 3; db "SELECT id,state FROM jobs WHERE id=896"
 Container queuelet-e01-api-1 Started 
896|queued
[exit 0]
```

### 2026-10-22T01:50:29Z #hypothesis H35 (source: operator.go:13 retry sets attempts=0; ack.go:14 inserts n=attempts). Job 302 history (ack→retry→cancel→nack→leased) gives /attempts with duplicate n=1 rows. Refuted if n values are unique.

### 2026-10-22T01:50:29Z #obs H35 attempts for 302
```
$ c acme-key http://localhost:18220/v1/jobs/302/attempts | head -c 600
{"attempts":[{"error":"","finished_at":"2026-10-22T01:41:55.209608Z","job_id":302,"n":1,"outcome":"succeeded","started_at":"2026-10-22T01:41:55.209608Z","worker_id":4},{"error":"worker noticed failure","finished_at":"2026-10-22T01:42:13.811739Z","job_id":302,"n":1,"outcome":"failed","started_at":"2026-10-22T01:42:13.811739Z","worker_id":4}]}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:50:29Z #hypothesis H36-H39 (code reading; expected-correct unless noted): H36 enqueue with run_at AND delay_seconds → 400; H37 run_at in the past → state queued; H38 schedule create with timezone 'Mars/Base' → 400 invalid_timezone, but PATCH timezone 'Mars/Base' → 200 (no validation, schedules.go:84); H39 PATCH queue max_attempts leaves existing job max_attempts unchanged.

### 2026-10-22T01:50:29Z #obs H36
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h34/jobs -d '{"kind":"x","delay_seconds":5,"run_at":"2030-01-01T00:00:00Z"}'
{"error":{"code":"invalid_job","message":"run_at and delay_seconds are exclusive"}}
 [HTTP 400]
[exit 0]
```

### 2026-10-22T01:50:29Z #obs H37
```
$ c acme-key -X POST http://localhost:18220/v1/queues/h34/jobs -d '{"kind":"x","run_at":"2020-01-01T00:00:00Z"}'
{"id":898,"state":"queued"}
 [HTTP 201]
[exit 0]
```

### 2026-10-22T01:50:29Z #obs H38 create
```
$ c acme-key -X POST http://localhost:18220/v1/schedules -d '{"name":"tz","queue":"h34","kind":"x","cron":"0 0 1 1 *","timezone":"Mars/Base"}'
{"error":{"code":"invalid_timezone","message":"unknown time zone Mars/Base"}}
 [HTTP 400]
[exit 0]
```

### 2026-10-22T01:50:29Z #obs H38 patch schedule 5
```
$ c acme-key -X PATCH http://localhost:18220/v1/schedules/5 -d '{"timezone":"Mars/Base"}' | grep -o '"timezone":"[^"]*"\|HTTP [0-9]*'
"timezone":"Mars/Base"
HTTP 200
[exit 0]
```

### 2026-10-22T01:50:30Z #obs H39 patch h34 max_attempts 7; job 899 keeps 3
```
$ c acme-key -X PATCH http://localhost:18220/v1/queues/h34 -d '{"max_attempts":7}' >/dev/null; db "SELECT id,max_attempts FROM jobs WHERE id=899"; echo new=$(enq h34 '{"kind":"x"}'); db "SELECT max(id),max_attempts FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='h34') GROUP BY 2 ORDER BY 1 DESC LIMIT 1"
899|3
new=900
900|7
[exit 0]
```

### 2026-10-22T01:50:49Z #hypothesis H40 (source: README verb list + qctl/main.go). Every remaining qctl verb against the seed returns the API's 2xx body (printed JSON) with exit 0; 'job nack' cannot express retry=false/delay except via --json (no flags, main.go:149-160). Refuted if a documented verb errors on valid input.

### 2026-10-22T01:50:49Z #obs H40 qctl verb sweep (one line per verb: verb → exit, first 110 chars of output)
results/qctl queue create qv --max-attempts 2 --lease-seconds 5 exit=0 { "backoff_base": 5, "backoff_kind": "exponential", "backoff_max": null, "id": 21, "lease_seconds": 5, "max_at
results/qctl queue get qv                                    exit=0 { "backoff_base": 5, "backoff_kind": "exponential", "backoff_max": null, "id": 21, "lease_seconds": 5, "max_at
results/qctl queue patch qv --json {"retention_hours":1}     exit=0 { "backoff_base": 5, "backoff_kind": "exponential", "backoff_max": null, "id": 21, "lease_seconds": 5, "max_at
results/qctl queue pause qv                                  exit=0 { "ok": true }
results/qctl queue resume qv                                 exit=0 { "ok": true }
results/qctl enqueue qv --kind echo --priority 9 --idempotency-key a exit=0 { "id": 901, "state": "queued" }
results/qctl batch qv --json {"jobs":[{"kind":"echo"},{"kind":"fail3"}]} exit=0 { "ids": [ 902, 903 ] }
results/qctl jobs ls qv --kind echo                          exit=0 { "jobs": [ { "attempts": 0, "created_at": "2026-10-22T01:50:50Z", "finished_at": null, "id": 901, "idempotenc
results/qctl queue purge qv                                  exit=0 { "deleted": 3 }
results/qctl enqueue qv --kind echo --delay 600              exit=0 { "id": 904, "state": "scheduled" }
results/qctl job heartbeat 905 --lease-token cf59561f982a11ae0e634b3573ff2aba --extend-seconds 30 exit=0 { "ok": true }
results/qctl job nack 905 --lease-token cf59561f982a11ae0e634b3573ff2aba --error boom --json {"retry":false} exit=0 { "ok": true }
results/qctl job get 905                                     exit=0 { "attempts": 1, "created_at": "2026-10-22T01:50:50Z", "finished_at": "2026-10-22T01:50:50.625336Z", "id": 905
results/qctl job retry 905                                   exit=0 { "ok": true }
results/qctl job cancel 905                                  exit=0 { "ok": true }
results/qctl job rm 905                                      exit=0 
results/qctl worker rm 6                                     exit=0 { "ok": true }
results/qctl schedule create --json {"name":"hourly","queue":"qv","kind":"echo","payload":{},"cron":"0 * * * *","timezone":"UTC"} exit=0 { "cron": "0 Makefile README.md SPEC.md cmd deploy go.mod go.sum internal lab migrations results test Makefile
results/qctl schedule get 8                                  exit=0 { "cron": "0 Makefile README.md SPEC.md cmd deploy go.mod go.sum internal lab migrations results test Makefile
results/qctl schedule patch 8 --json {"enabled":false}       exit=0 { "cron": "0 Makefile README.md SPEC.md cmd deploy go.mod go.sum internal lab migrations results test Makefile
results/qctl schedule run 8                                  exit=0 { "id": 906, "state": "queued" }
results/qctl schedule rm 8                                   exit=0 
results/qctl stats qv                                        exit=0 { "cancelled": 0, "dead": 0, "leased": 0, "oldest_queued_age": 0.092387, "queued": 2, "retry_wait": 0, "schedu
results/qctl stats x --namespace acme                        exit=0 { "cancelled": 18, "dead": 19, "leased": 0, "oldest_queued_age": 214797051.171649, "queued": 731, "retry_wait"
results/qctl queue rm qv                                     exit=0 

### 2026-10-22T01:51:07Z harness artifact: in the sweep above, '"cron": "0 Makefile README.md …' is my own unquoted `echo $out` glob-expanding '*' in the cron string, not SUT output. The schedule was created correctly (exit 0, later get/patch/run/rm succeeded).

### 2026-10-22T01:51:07Z #surprise namespace stats oldest_queued_age = 214797051 s (~6.8 y) on acme. Likely cause (to test): job 898 enqueued in H37 with run_at 2020-01-01 is queued; stats.go:234 measures age from min(run_at), not from when it became queued/created.

### 2026-10-22T01:51:07Z #hypothesis H41 (source: README worker mode; worker.go). `qctl worker run h41 --kinds echo,fail` (timeout 22 s) on queue h41 (max_attempts 3, exponential base 5) with one echo job and one fail2 job: echo acked; fail2 fails attempt 1 and 2 (retry_wait 5 s then 10 s), succeeds attempt 3; attempts rows n=1 failed, n=2 failed, n=3 succeeded. Refuted if fail2 not succeeded by 22 s or wrong attempts.

### 2026-10-22T01:51:07Z #obs worker run (timeout 22s)
```
$ timeout 22 results/qctl worker run h41 --kinds echo,fail2 2>&1 | grep -v '^$' | tail -8
2026/10/22 01:51:08 processing kind=echo attempt=1
2026/10/22 01:51:08 completed status=200 error=<nil>
2026/10/22 01:51:08 processing kind=fail2 attempt=1
2026/10/22 01:51:08 completed status=200 error=<nil>
2026/10/22 01:51:13 processing kind=fail2 attempt=2
2026/10/22 01:51:13 completed status=200 error=<nil>
2026/10/22 01:51:24 processing kind=fail2 attempt=3
2026/10/22 01:51:24 completed status=200 error=<nil>
[exit 0]
```

### 2026-10-22T01:51:30Z #obs result
```
$ db "SELECT id,kind,state,attempts,result FROM jobs WHERE id IN (908,909); SELECT job_id,n,outcome,error FROM attempts WHERE job_id IN (908,909) ORDER BY job_id,n"
908|echo|succeeded|1|{"x": 1}
909|fail2|succeeded|3|{}
908|1|succeeded|
909|1|failed|requested failure 1
909|2|failed|requested failure 2
909|3|succeeded|
[exit 0]
```

### 2026-10-22T01:51:30Z #obs F-candidate stats age isolation: compare namespace age with/without job 898

### 2026-10-22T01:51:30Z #obs age per acme queue (h34 holds 898)
```
$ for q in h34 imports emails; do echo $q $(c acme-key http://localhost:18220/v1/queues/$q/stats | grep -o '"oldest_queued_age":[0-9.]*'); done; db "SELECT id,state,run_at,created_at FROM jobs WHERE id=898"
h34 "oldest_queued_age":214797090.198135
imports "oldest_queued_age":622.409063
emails "oldest_queued_age":0
898|queued|2020-01-01 00:00:00+00|2026-10-22 01:50:30
[exit 0]
```

### 2026-10-22T01:52:02Z #hypothesis H25b (source: fire.go:137 FixedZone(offset captured at create); stored seed row daily-report zone_offset=-14400). Route: NOT the running service — a scratch harness results/dst/main.go calls the SUT's schedule.Next/Offset with the stored row values (no service clock control exists; see TR2). Prediction: Next('0 9 * * *','America/New_York',-14400, after 2026-11-02T00:00Z) = 2026-11-02T13:00Z (08:00 EST) whereas civil 09:00 EST = 14:00Z. Refuted if 14:00Z.

### 2026-10-22T01:52:02Z #obs H25b scratch harness
```
$ go run -buildvcs=false ./results/dst
stored-offset next: 2026-11-02 13:00:00 +0000 UTC = 08:00 EST | offset now would be -18000
[exit 0]
```

### 2026-10-22T01:52:04Z #obs running-service corroboration: stored offset rows
```
$ db "SELECT id,name,timezone,zone_offset,next_run_at FROM schedules WHERE id=1"
1|daily-report|America/New_York|-14400|2026-10-22 13:00:00+00
[exit 0]
```

### 2026-10-22T01:52:16Z #obs dead end: `pkill -f 'results/tools/sigcap.py'` also matched and killed my own shell (exit 144) before logging; pgrep afterwards shows the listener is gone too. Raw captures remain in project/results/sigcap.jsonl.

### 2026-10-22T01:52:16Z #obs lab container state at end of probing
```
$ docker compose -f deploy/compose.yml ps --format '{{.Service}} {{.Status}}'
api Up About a minute
hooksink Up 11 minutes
postgres Up 2 minutes (healthy)
worker-a Up 10 minutes
worker-b Up 10 minutes
[exit 0]
```

### 2026-10-22T01:52:16Z #obs worker-a/b last log lines
```
$ docker logs --tail 2 queuelet-e01-worker-a-1 2>&1; docker logs --tail 2 queuelet-e01-worker-b-1 2>&1
2026/10/22 01:49:47 HTTP 401: {"error":{"code":"unauthorized","message":"invalid API key"}}
2026/10/22 01:50:24 Post "http://api:8200/v1/queues/emails/lease": dial tcp 172.22.0.4:8200: connect: connection refused
2026/10/22 01:49:47 HTTP 401: {"error":{"code":"unauthorized","message":"invalid API key"}}
2026/10/22 01:50:24 Post "http://api:8200/v1/queues/reports/lease": dial tcp 172.22.0.4:8200: connect: connection refused
[exit 0]
```

### 2026-10-22T01:52:16Z #read — phase change: probing ends, writing begins (reports, models, hypotheses, revealed).

### 2026-10-22T01:52:25Z #surprise worker-a/b logs at 01:49:47 (the H26 postgres restart) show 'HTTP 401 … invalid API key' for the valid acme-key.

### 2026-10-22T01:52:25Z #hypothesis H42 (source: surprise above + api.go:124 maps any QueryRow error to 401). With postgres stopped, a request with valid acme-key returns 401 unauthorized 'invalid API key' (not 5xx); /health stays 200 ok. Refuted if 5xx.

### 2026-10-22T01:52:25Z #obs stop postgres (isolated project)
```
$ docker compose -f deploy/compose.yml stop postgres 2>&1 | tail -1
 Container queuelet-e01-postgres-1 Stopped 
[exit 0]
```

### 2026-10-22T01:52:26Z #obs valid key while DB down; health
```
$ c acme-key http://localhost:18220/v1/queues; echo; curl -s -w ' [HTTP %{http_code}]' http://localhost:18220/health
{"error":{"code":"unauthorized","message":"invalid API key"}}
 [HTTP 401]
{"status":"ok"}
 [HTTP 200]
[exit 0]
```

### 2026-10-22T01:52:26Z #obs start postgres; restart api (lab recovery from F08)
```
$ docker compose -f deploy/compose.yml start postgres 2>&1 | tail -1; for i in $(seq 1 30); do pg_isready -h localhost -p 15453 -q && break; sleep 1; done; docker compose -f deploy/compose.yml restart api 2>&1 | tail -1; QUEUELET_URL=http://localhost:18220 bash lab/wait.sh 2>/dev/null; c acme-key http://localhost:18220/v1/queues/h34 | tail -c 12
 Container queuelet-e01-postgres-1 Started 
 Container queuelet-e01-api-1 Started 

 [HTTP 200]
[exit 0]
```

## 2026-10-22T01:57:10Z #read — ERRATUM: file:line citations in the 01:38:42Z source-walk entry and in hypothesis rows
While writing reports I found my early citations came from `cat -n` over several concatenated files, so for every file
that was not first in its batch the number is cumulative. The log above is kept as written; corrected per-file numbers
(verified with grep at ~01:55Z) are: api.go Loop 93-108 (return on error 102-104, step list 101), decode 42-49 (decoder 43),
auth 86-88, /health 56; queues.go create validation 35, PATCH 56-83 (lease check 73, UPDATE 82), DELETE 85-92 (stmt 86),
purge 109; internal/api/jobs.go getJob 10-16 (query 11), attempts 19, deleteJob 23, deleteWorker release 110, lease worker
check 130, `fail(w, 200, …)` 132; internal/jobs/expire.go Expire 6, Promote 10, Retain 14; internal/jobs/jobs.go Backoff 17-22,
LeaseOrder 24, Lease 26-67, Notify 69-88 (truncation 81, 84), Heartbeat 91; nack.go read 10, dead test 15, update 22,
attempt insert 26; operator.go Cancel 6, Retry 13; stats.go fold 17, age 19; schedule/fire.go FixedZone 17, select 28,
paused skip 37, catch-up loop 41, insert 44, Next 48; webhook/ship.go select 17, Marshal 22, MarshalIndent 23, mac.Write 25,
NewRequest 26, client timeout 32, delivered update 43; hooksink verify 52-56; store.go JSON decode 43-46.
Files cited correctly as first-in-batch: cmd/api/main.go, enqueue.go, ack.go, schedules.go, cmd/qctl/main.go, worker.go.
Defect reports and models use the corrected numbers. No prediction or outcome changes.

## 2026-10-22T01:57:10Z #surprise dispositions (every surprise → report or reason)
| Surprise | Disposition |
|---|---|
| stats queued 80 vs DB 49 (01:41:46Z) | F18 |
| impostor ack accepted (H02) | F01 |
| nack revives queued/succeeded/cancelled (H03/H07/H27) | F02 |
| cross-namespace job access (H04/H28) | F03 |
| attempt 2 on max_attempts 1 (H05) | F04 |
| expiry leaves attempts untouched (H06) | F05 |
| retry of succeeded accepted (H08) | F06 |
| schedule PATCH no effect / disabled fires (H09) | F07 |
| loop dies (H10, H26) | F08 (+F09 trigger) |
| H10b refuted: restart promotes once | explained by step order api.go:101; folded into F08 "variables isolated"; not a separate bug |
| signature_valid false everywhere (H11/H11b) | F10 |
| 500 counted delivered (H12) | F11 |
| result stringified/truncated (H13) | F12 |
| 13.3 s promotion under webhook timeout (H34) | F13 |
| 401 for valid key during DB outage; /health ok (H42) | F14 |
| orphaned jobs/schedules on queue delete (H14) | F15 |
| batch truncation/empty/partial (H15) | F16 |
| precision loss (H16) | F17 |
| validation gaps (H18/H19) | F19, F20 |
| lease unknown worker → 200 (H20) | F21 |
| qctl exit 0 (H21) | F22 |
| FIFO inversion (H22b) | F23 |
| catch-up burst (H23) | F24 |
| purge leaves scheduled/retry_wait (H24) | F25 (status spec) |
| oldest_queued_age ≈ 6.8 years | F26 |
| DST offset frozen (H25b, harness route) | F27 |
| attempts.started_at == finished_at for every row (01:41:38Z onward) | **Not reported**: SPEC never defines attempt timestamps; no oracle says started_at must be lease time. Backlogged as Q1 (the column name suggests otherwise). |
| /attempts shows two rows with n=1 for job 302 (H35) | **Not reported separately**: arises from retry resetting attempts (SPEC requires the reset). Here it followed F06's illegitimate retry; for a legit dead→retry the spec is silent on numbering. Backlogged Q2. |
| 404 messages expose "sql: no rows in result set" | **Not a bug**: error envelope and status match SPEC; information leak is minor and has no oracle. Q3. |
| lease picked job 302 before 304 in H07 | **Not a bug**: 302 had older created_at (retried), FIFO by creation order. |
| admin-key acts as an ordinary acme key for mutations | **Not reported**: SPEC says only that an admin key "may inspect namespace statistics"; silent on its other rights. Q4. |
| `webhook_secret` returned in plain text by GET /v1/queues/{q} (seen in h05 create output) | **Not reported**: SPEC silent; noted as risk Q5 #offcharter. |

## 2026-10-22T01:57:10Z #question — testability requests
- **TR1** Question: which bytes did the API sign? Missing signal: hooksink records `signature_valid` but not the received
  `X-Queuelet-Signature`. Minimal hook: add `"signature"` (received header) to each /admin/records entry. Validation probe:
  ack a job in a webhook queue and recompute HMAC over `raw_body`. (Worked around with a host capture listener, F10.)
- **TR2** Question: does a schedule follow civil time across a DST change end-to-end? Missing signal: no clock control for the
  API loop. Minimal hook: a documented lab clock offset (e.g. env `QUEUELET_CLOCK_OFFSET` read by Loop/handlers) or an admin
  "tick at time T" endpoint in the lab build. Validation probe: set clock to 2026-11-01T12:00Z, observe daily-report fire at 14:00Z.
- **TR3** Question: is the background loop alive? Missing signal: /health and /metrics do not reflect the loop. Minimal hook:
  `queuelet_loop_last_tick_seconds` in /metrics. Validation probe: restart postgres, watch the gauge stop advancing (F08).

## 2026-10-22T01:57:10Z #offcharter backlog
- Retention sweep with negative/zero retention (F19 PATCH accepts -5) — predicted immediate deletion, not probed.
- Worker `crash` kind with F05: predicted infinite crash loop of worker containers (restart: on-failure).
- Concurrent lease atomicity under load (baseline check covers disjointness; not re-probed).
- Webhook URL that fails http.NewRequest → predicted to stop the loop (F08 sibling).
- Idempotency key inside a batch / concurrent duplicate enqueue.

## 2026-10-22T01:57:10Z #read — spec silences and my code-based predictions (charter: "where the spec is silent")
| Spec sentence / gap | What it does not say | Code predicts | Status |
|---|---|---|---|
| purge "removes all pending jobs" | which states are pending | queued only (queues.go:109) | observed H24 → F25 |
| "oldest queued age" | measured from created_at, run_at or promotion | min(run_at) (stats.go:19) | observed → F26 |
| "returns completed attempts in numeric order" | numbering after operator retry | restarts at 1 | observed H35 → Q2 |
| attempts fields | meaning of started_at | equals finish time (ack.go:14, nack.go:26) | observed → Q1 |
| "DELETE /v1/workers/{id} releases its active leases" | whether release counts an attempt | no attempt (internal/api/jobs.go:110) | read only |
| admin key | rights beyond namespace stats | full rights on namespace 1 (migrations seed) | read only → Q4 |
| "Failed notifications are retried" | retry cadence/cap | every 5 s forever on transport errors (ship.go:34) | observed H34 (retries happened) |
| idempotency "returns the original job" | status code on replay; differing body | 201 with original {id,state}, body ignored | observed H29 |
| lease "response includes … expiry" | lease_expires_at precision/time zone | Go time in UTC with ns | observed H01 |
| heartbeat "positive extend_seconds" | upper bound | none | read only |
| retention_hours | non-integer/negative values | float accepted, negative accepted (F19) | observed H18/H19 |
| pause | effect on already-leased jobs | none (lease only) | read only |

### 2026-10-22T01:57:14Z #obs Q5 check (written disposition above cited only an empty secret field; verifying): GET reports queue as acme

### 2026-10-22T01:57:14Z #obs GET /v1/queues/reports webhook fields
```
$ c acme-key http://localhost:18220/v1/queues/reports | grep -o '"webhook_[a-z]*":"[^"]*"'
"webhook_secret":"report-secret"
"webhook_url":"http://hooksink:8201/hook"
[exit 0]
```

## 2026-10-22T02:00:30Z #read — run closed
Artifacts: revealed.md, ../../hypotheses.md, ../../models/*.md, ../../defects/F01–F27. Isolated stack queuelet-e01 left running (not torn down).
