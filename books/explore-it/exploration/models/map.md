# Map — entry points, packages, verbs, background work

Build: `explore-it/v0-inherited` export. Legend: **O** observed in E01 (link), **I** inferred from source only, **?** unknown.
Source pointers are per-file line numbers (see erratum in [E01 notes](../experiments/01-map-it/notes.md)).

## Processes (deploy/compose.yml)
| Process | Entry | Role | Evidence |
|---|---|---|---|
| api | cmd/api/main.go:15-38 | applies migrations/001.sql, starts HTTP :8200 and `Loop` | O: stack up 01:40Z, logs "queuelet listening" |
| worker-a / worker-b | cmd/worker/main.go → internal/worker/worker.go:25-75 | register, poll lease every 200 ms, heartbeat 5 s, ack/nack | O: worker logs; H41 (`qctl worker run`) |
| hooksink | cmd/hooksink/main.go | webhook partner fake; /admin/mode, /admin/records | O: H11, H12, H34 |
| postgres | postgres:17 | all state | O: explorer role reads |
| qctl | cmd/qctl/main.go | CLI → HTTP via internal/client | O: H01, H40 sweep |

## HTTP routes → handler → domain (internal/api/api.go:54-91)
| Route | Handler | Domain / SQL | Walked |
|---|---|---|---|
| auth middleware | api.go:79-90 | sha256(key) → api_keys | O H32, H42 (DB down → 401, F14) |
| GET /health, /metrics | api.go:56-57 | static | O H32 |
| POST/GET /v1/queues, GET/PATCH/DELETE /v1/queues/{q} | queues.go:25-92 | queues table | O H01, H14, H18, H40 |
| POST /v1/queues/{q}/{jobs,jobs:batch,lease,pause,resume,purge} | queues.go:93-119 → enqueue.go / jobs.go:116 / jobs.Lease | jobs | O all six |
| GET /v1/queues/{q}/jobs | enqueue.go:95-117 | jobs ORDER BY id | O H01, H18 |
| GET /v1/queues/{q}/stats, /v1/namespaces/{ns}/stats | queues.go:120-141 → stats.Read | aggregates | O H01, H17, H30 |
| GET/DELETE /v1/jobs/{id}, GET …/attempts | internal/api/jobs.go:10-29 | no namespace filter | O H04 (F03) |
| POST /v1/jobs/{id}/{cancel,retry,ack,nack,heartbeat} | internal/api/jobs.go:30-76 → internal/jobs/{operator,ack,nack,jobs}.go | jobs, attempts, outbox | O all five |
| POST /v1/workers, DELETE /v1/workers/{id} | internal/api/jobs.go:77-115 | workers; release leases | O H22b, H40 |
| /v1/schedules CRUD + run | internal/api/schedules.go | schedules | O H09, H10, H14b, H40 |

## Background loop (internal/api/api.go:93-108), every 1 s, in order
| Step | Code | Observed |
|---|---|---|
| Promote scheduled/retry_wait → queued | internal/jobs/expire.go:9-11 | O (delay jobs promote ~2 s, H34 control) |
| Expire leases → queued (no attempt) | expire.go:5-7 | O H06 (F05) |
| Fire schedules (catch-up loop) | internal/schedule/fire.go:27-63 | O H09b, H23 (F24) |
| Retain terminal jobs | expire.go:13-15 | I (not observed; 24 h default) |
| Ship webhooks (inline HTTP, 2 s timeout, 20 rows) | internal/webhook/ship.go:16-49 | O H11, H12, H34 (F13) |
| Any step error → loop returns forever | api.go:102-104 | O H10, H26 (F08) |

## Persistence (migrations/001.sql)
Tables namespaces, api_keys, queues, jobs, attempts, workers, schedules, webhook_outbox; no foreign keys (O: orphans F15);
`jobs.created_at timestamp(0)` (O: second buckets H22); unique (queue_id, idempotency_key) (O: H29).

## E02 revision — `explore-it/lesson-01` background work and consoles ([E02 notes](../experiments/02-make-it-visible/notes.md))

| Step | Code (lesson-01, per-file lines) | Observed in E02 |
|---|---|---|
| ★ Core loop: Promote, Expire, Fire, Retain each tick; error logged as `background task: %v`, loop continues | internal/api/api.go:129-145 | O during DB restart: 24 error lines, then recovery (H15) |
| ★ Separate shipper goroutine, 1 s tick, `webhook task: %v` on error | api.go:113-128; internal/webhook/ship.go:17-49 | O: success log only, no failure log (H05); parse error each second (H11) |
| ★ Expire writes `expired` attempt rows and dead-letters at max | internal/jobs/expire.go:12-56 | O H01, H13 |
| ★ /health = DB ping ∧ last core tick ok ∧ last ship tick ok | api.go:62-70 | O 503 on bad URL (H11) and DB restart (H15); 200 during partner 500/slow (H08, H12) |
| /metrics constant `queuelet_up 1` | api.go:71 | O (H08) |

Console producers: API process log has no request or lifecycle logging and no ids; worker stdout has no job ids
(internal/worker/worker.go:48,71); hooksink keeps per-POST records in memory (cmd/hooksink/main.go:67). ★ console inventory O.
