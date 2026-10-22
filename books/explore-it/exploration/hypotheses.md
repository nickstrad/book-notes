# Hypothesis backlog

Experiment 01 (map-it) stated H00–H42 before their probes (see notes). Q*/H43+ are backlog. Preserve every proposed hypothesis, including off-charter questions and refuted predictions; update its outcome without rewriting its initial statement. See [the protocol](PROTOCOL.md).

| Id | Statement / prediction | Source and evidence pointer | Would be wrong if | Probe family / experiment | Status | Outcome evidence |
|---|---|---|---|---|---|---|
| H00 | README happy path works and the spec is complete | charter 01 | a documented verb misbehaves or a spec sentence has no code behind it | recon / E01 | experiment 01: refuted | happy path held (H01) but many documented verbs misbehave and several spec sentences have no code (loop recovery, unknown-field rejection, civil time): F01–F27 |
| H01 | qctl README sequence (queue ls, enqueue, jobs ls, worker create, lease, ack, get, attempts, stats) succeeds on the seed | recon README | any step errors | sequence / E01 | experiment 01: confirmed | experiments/01-map-it/notes.md 01:41:31Z–01:41:38Z |
| H02 | ack with a wrong lease_token on a leased job → 200, succeeded | code ack.go:10; SPEC worker | 409 | oracle / E01 | experiment 01: confirmed | F01 |
| H03 | nack with wrong token on a never-leased queued job → 200, retry_wait, attempts 1 | code nack.go:10-22 | 409 or no change | oracle/state / E01 | experiment 01: confirmed | F02 |
| H04 | globex-key can GET and cancel acme jobs by id | code internal/api/jobs.go:11; SPEC auth | 404/403 | boundary / E01 | experiment 01: confirmed | F03 |
| H05 | max_attempts=1: lease+nack(retry) → not dead; second lease possible | code nack.go:15 | first nack → dead | sequence / E01 | experiment 01: confirmed | F04 |
| H06 | after lease expiry: queued, attempts unchanged, no attempt row | code expire.go:6 | attempt row or attempts+1 | state / E01 | experiment 01: confirmed | F05 |
| H07 | cancel a leased job, then holder nack → 200 and job revived | code nack.go | 409 / stays cancelled | sequence / E01 | experiment 01: confirmed | F02 |
| H08 | retry on succeeded job → 200 queued | code operator.go:13 | 409 | state / E01 | experiment 01: confirmed | F06 |
| H09 | PATCH schedule cron/tz → next_run_at unchanged; PATCH enabled=false → unchanged and still fires | code schedules.go:84, fire.go:28 | next_run_at changes / no fire | state / E01 | experiment 01: confirmed | F07 |
| H09c | schedule created disabled then PATCH enabled=true → next_run_at null, never fires | code schedules.go:52-56,84 | fires in next minute | state / E01 | experiment 01: confirmed | F07 (one tick observed) |
| H10 | PATCH cron 'bogus' → 200; next Fire error stops the whole loop | code schedules.go:84, api.go:102-104 | 400 or promotion continues | ecosystem / E01 | experiment 01: confirmed | F08, F09 |
| H10b | restarting API does not recover from poisoned schedule: delayed job stays scheduled | code fire.go:44-48 | job promoted after restart | ecosystem / E01 | experiment 01: refuted | job 892 promoted on first tick (Promote precedes Fire, api.go:101); loop then died again — see F08 |
| H11 | hooksink records signature_valid=false for every delivery | code ship.go:22-25 | signature_valid=true | oracle / E01 | experiment 01: confirmed | F10 |
| H11b | received signature = HMAC(pretty JSON) ≠ HMAC(received bytes) | code ship.go:22-26 | matches received bytes | oracle / E01 | experiment 01: confirmed | F10; project/results/sigcap.jsonl |
| H12 | hooksink 500 → outbox delivered after 1 POST, no retry | code ship.go:40-43 | attempts>1 / undelivered | ecosystem / E01 | experiment 01: confirmed | F11 |
| H13 | webhook body.result is a JSON string; >1024 B truncated silently | code jobs.go:79-86 | object | oracle / E01 | experiment 01: confirmed | F12 |
| H14 | DELETE nonempty queue → 204; jobs still readable; DELETE missing → 204 | code queues.go:86 | 409/404 | entity / E01 | experiment 01: confirmed | F15 |
| H15 | batch 501 → 201 with 500; empty → 201; mid-batch invalid → 400 with partial insert | code enqueue.go:80-93 | 400 without insertion | variation / E01 | experiment 01: confirmed | F16 |
| H16 | payload 9007199254740993 reads back …992 | code enqueue.go:14 | exact | variation / E01 | experiment 01: confirmed | F17 |
| H17 | stats.queued = DB queued+scheduled+retry_wait | code stats.go:17 | equals DB queued | oracle / E01 | experiment 01: confirmed | F18 |
| H18 | unknown field → 201; backoff kind zigzag → 201; state=bogus → all jobs | code api.go:43 | 400 | variation / E01 | experiment 01: confirmed | F20 |
| H19 | queue name a.b → 201; priority 42 → 201 | code queues.go:35, enqueue.go:29 | 400 | variation / E01 | experiment 01: confirmed | F19 |
| H20 | lease unknown worker → HTTP 200 + error body | code internal/api/jobs.go:132 | 404 | oracle / E01 | experiment 01: confirmed | F21 |
| H21 | qctl prints HTTP 404 and exits 0 | code cmd/qctl/main.go:16-18 | exit≠0 | console / E01 | experiment 01: confirmed | F22 |
| H22 | same-second same-priority jobs may lease out of id order | code jobs.go:24 + timestamp(0) | (ties) — inversion not required | sequence / E01 | experiment 01: needs more | 40 untouched jobs leased in order (no inversion); see H22b |
| H22b | after release of #1 among same-second ties, #2..#10 lease before #1 | surprise H22 + code | #1 leases first | sequence / E01 | experiment 01: confirmed | F23 (inversion even before release) |
| H23 | pause, miss ≥2 ticks of * * * * *, resume → one job per missed tick at once | code fire.go:37-41 | exactly 1 job | sequence/clock / E01 | experiment 01: confirmed | F24 (4 jobs) |
| H24 | purge leaves scheduled and retry_wait jobs | code queues.go:109; spec silence | scheduled deleted | variation / E01 | experiment 01: confirmed | F25 (spec) |
| H25 | zone offset frozen at create; DST ignored — not probeable without clock | code fire.go:17 | — | testability / E01 | experiment 01: needs more | TR2; partial evidence H25b |
| H25b | schedule.Next with stored offset gives 13:00Z (08:00 EST) on 2026-11-02, not 14:00Z | code fire.go:17 + stored row | 14:00Z | harness (SUT function, not service) / E01 | experiment 01: confirmed | F27 |
| H26 | postgres restart → loop dead; delayed job never promoted | code api.go:102-104; SPEC ops | promotion resumes within 10 s | ecosystem / E01 | experiment 01: confirmed | F08 |
| H27 | nack on succeeded job → 200, retry_wait | surprise H03 + nack.go | 409 | state / E01 | experiment 01: confirmed | F02 |
| H28 | globex nack of acme job → 200 | sibling H04 + nack.go | 404/409 | boundary / E01 | experiment 01: confirmed | F02, F03 |
| H29 | idempotency key repeat → same id | code enqueue.go:55-58 | ids differ | variation / E01 | experiment 01: confirmed | holds |
| H30 | namespace stats: acme→globex 403; admin→globex 200 | code queues.go:128-141 | otherwise | boundary / E01 | experiment 01: confirmed | holds |
| H31 | heartbeat with wrong token → 409 | code jobs.go:91 | 200 | oracle / E01 | experiment 01: confirmed | holds |
| H32 | no auth → 401; /health 200; /metrics text | code api.go:56-57,79-90 | otherwise | recon / E01 | experiment 01: confirmed | holds |
| H33 | DELETE queued job → 409 | code internal/api/jobs.go:23 | 204 | state / E01 | experiment 01: confirmed | holds |
| H34 | with 6 timed-out webhooks, 1 s delayed job takes ≥10 s to promote | code ship.go:32, api.go:101 | promoted within 3 s | ecosystem / E01 | experiment 01: confirmed | F13 (13.3 s vs 2.1 s control) |
| H35 | /attempts shows duplicate n=1 after retry | code operator.go:13, ack.go:14 | unique n | state / E01 | experiment 01: confirmed | Q2 (not reported) |
| H36 | run_at + delay_seconds → 400 | code enqueue.go:26 | accepted | variation / E01 | experiment 01: confirmed | holds |
| H37 | run_at in the past → queued | code enqueue.go:47-50 | scheduled/400 | variation / E01 | experiment 01: confirmed | holds; led to F26 |
| H38 | create tz Mars/Base → 400; PATCH tz Mars/Base → 200 | code schedules.go:33,84 | patch 400 | variation / E01 | experiment 01: confirmed | F09 |
| H39 | PATCH queue max_attempts leaves existing jobs' value | code queues.go:82, enqueue.go:33-36 | existing changes | state / E01 | experiment 01: confirmed | holds |
| H40 | every remaining qctl verb succeeds on valid input; nack retry=false only via --json | recon README + qctl/main.go | a documented verb errors | recon / E01 | experiment 01: confirmed | project/results/h40-sweep.txt |
| H41 | qctl worker run: echo acked; fail2 fails 2× then succeeds at attempt 3 | recon README + worker.go | fail2 not succeeded by 22 s | sequence / E01 | experiment 01: confirmed | attempt rows n=1,2 failed, 3 succeeded |
| H42 | postgres stopped → valid key gets 401; /health 200 | surprise worker logs + api.go:86-88 | 5xx | ecosystem / E01 | experiment 01: confirmed | F14 |
| Q1 | attempts.started_at should be lease start, not finish | surprise 01:41:38Z | spec defines otherwise | oracle / spec decision | open |  |
| Q2 | attempt numbering after operator retry should stay unique/monotonic | H35 | spec says numbering restarts | spec decision | open |  |
| Q3 | 404 error messages should not expose driver text ("sql: no rows in result set") | #obs H04 controls | spec permits any message | console / spec | open |  |
| Q4 | admin key should be read-only beyond stats | spec silence | spec grants mutations | boundary / spec | open |  |
| Q5 | GET queue should not return webhook_secret | #obs 01:5xZ (report-secret returned) | spec requires echo | boundary / spec | open |  |
| H43 | worker deregistration release should count an attempt (like expiry) | code internal/api/jobs.go:110 | spec says release is free | state / next | open |  |
| H44 | webhook URL that fails http.NewRequest stops the loop | code ship.go:26-28 | loop continues | ecosystem / next | experiment 02: confirmed | at lesson-01 the shipper keeps ticking but delivers nothing after the bad row, all namespaces (E02-H11, E02-F02); core loop progress not observed during it |
| H45 | schedule catch-up after loop outage bursts like F24 | code fire.go:41 | bounded | sequence / next | open |  |
| H46 | negative retention_hours deletes terminal jobs on next tick | code expire.go:14 | jobs kept | variation / next | open |  |
| H47 | crash-kind job loops worker containers forever (F05 + crash) | code worker.go:84-85, expire.go:6 | job goes dead | sequence / next | experiment 02: refuted | at lesson-01 job 301 dead after 3 crashes, 304 after 4 (E02-H01, H13); the v0 code it was read from was changed by the lesson-01 fix |
| H48 | retry and ack across namespaces succeed (F03 siblings) | code operator.go:13, ack.go:10 | 404 | boundary / next | open |  |
| E02-H01 | crash job (emails, lease 10 s, max 3) dies by expiry after exactly 3 executions; attempts 3; 3 `expired` rows | charter; code expire.go:35-45 | >3 executions or not dead/3 rows | console comparison / E02 | experiment 02: confirmed | E02 notes 03:24:03Z |
| E02-H02 | API log shows no line for lease, expiry or dead | code (no lifecycle logging) | API logs a transition | console / E02 | experiment 02: confirmed | E02 notes 03:24:03Z |
| E02-H03 | worker log 3 × processing, 3 × registered; RestartCount +3; one worker row per incarnation | code worker.go:29-48; compose | counts differ | console / E02 | experiment 02: confirmed | E02 notes 03:24:03Z |
| E02-H04 | expired rows: distinct worker ids; started_at = finished_at = expiry tick | code expire.go:45 | started_at earlier / ids repeat | console / E02 | experiment 02: confirmed | E02 notes; TR-E02-5 |
| E02-H05 | partner 500: retried, hooksink count = outbox attempts, undelivered, API log silent | code ship.go:32-41 | delivered on 500 / counts differ / logged | console comparison / E02 | experiment 02: confirmed | E02 notes 03:26:10Z |
| E02-H06 | mode ok → delivered on next POST, one id-less log line | code ship.go:45-46 | not delivered / duplicate after delivery | console / E02 | experiment 02: confirmed | E02 notes 03:27:06Z |
| E02-H07 | retry interval fixed ≈5–6 s, no cap | code ship.go:37 | growth or give-up | variation / E02 | experiment 02: confirmed | 11 failures, not a bug (spec silent) |
| E02-H08 | /health ok during partner failures | code api.go:62-70 | degraded | console / E02 | experiment 02: confirmed | 52/52 samples |
| E02-H09 | `fail` n=5 on emails: 3 nacks → dead | code worker.go:86-95, nack.go:19 | 4th execution | sequence / E02 | experiment 02: confirmed | job 303 |
| E02-H10 | `fail3` never leased by documented worker kinds | code jobs.go:44 + cmd/worker/main.go:24 | processed within 10 s | variation / E02 | experiment 02: confirmed | E02-F01 |
| E02-H11 | unparsable webhook URL accepted; blocks every later outbox row, all tenants; health 503 | E01 H44; code ship.go:25-28 | acme delivered / health ok / 400 | ecosystem / E02 | experiment 02: confirmed | E02-F02 |
| E02-H11a | PATCH queue URL does not unblock pending row | code jobs.go:80 | delivered after PATCH | isolation / E02 | experiment 02: confirmed | E02-F02 |
| E02-H11b | DELETE job leaves orphan outbox row that keeps blocking | code migrations/001.sql:10 | row removed / delivered | isolation / E02 | experiment 02: confirmed | E02-F02 |
| E02-H11c | `ftp://` URL is a per-row failure, not blocking | code ship.go:31-41 | blocks | isolation / E02 | experiment 02: confirmed | E02-F02 |
| E02-H11d | PATCH accepts unparsable URL (200) | code queues.go:79-84 | 400 | sibling / E02 | experiment 02: confirmed | blocking via PATCH inferred, not re-triggered |
| E02-H12 | slow partner (3 s) gets duplicate 200s, never delivered while slow | code ship.go:31 | delivered while slow | ecosystem / E02 | experiment 02: confirmed | E02-F03 |
| E02-H13 | expiry-dead job on webhook queue notifies once (state dead, last_error) | code expire.go:53-56 | no outbox row | sequence / E02 | experiment 02: confirmed | outbox 73 |
| E02-H14 | frozen worker: expiry, stale ack 409, re-run, succeeded attempts 2 | code worker.go:49-71, ack.go | first ack accepted / no re-run | sequence / E02 | experiment 02: confirmed | job 305 (second fixture) |
| E02-H15 | DB restart: 503s and health degraded, then recovery and promotion | code api.go:62-70,129-145 | not recovered in 15 s | ecosystem / E02 (regression context) | experiment 02: confirmed | not a discovery |
| Q-E02-1 | which process exit ended an expired attempt should be observable | E02 S1/S2 | — | testability | open | TR-E02-1 |
| Q-E02-2 | should a later success clear or scope `last_error` (GET, webhook body)? | E02 S9 | spec defines it | spec decision | open |  |
| H49 | ship.go UPDATE errors (:37-40, :46-49) abort the whole batch like the parse error | E02 C11 sibling, code | other rows still shipped | ecosystem / next | open |  |
| H50 | a webhook body larger than the partner limit or a secret change with rows pending behaves per row (no global block) | E02 C11 sibling | global block | ecosystem / next | open |  |
| H51 | `qctl worker run --kinds echo,sleep,fail,crash,http` (README) also never leases `fail3` | E02-F01 sibling, code (same worker.Run) | fail3 leased | variation / next | open |  |
| H52 | after operator retry of a crash-dead job, attempts numbering restarts at 1 (duplicate n with expired rows) | E01 Q2 + E02 H01 | unique n | state / next | open |  |

Allowed sources: recon, spec silence, code reading, model cell, surprise and risk headline. Status: `open` or `experiment NN: confirmed`, `experiment NN: refuted`, `experiment NN: needs more`.
