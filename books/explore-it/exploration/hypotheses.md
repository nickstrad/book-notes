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
| H46 | negative retention_hours deletes terminal jobs on next tick | code expire.go:14 | jobs kept | variation / next | experiment 03: refuted | at lesson-02 `retention_hours:-1` create and `-5` PATCH → 400 (E03 notes 04:17:13Z), so the state cannot be reached; the v0 code it was read from changed in lesson 01. Large positive values break Retain instead (E03-F05) |
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
| E03-HC | Charter: every documented range is enforced and unknown fields are rejected | charter 03; validators internal/api | any 2xx where SPEC says 4xx | variation / E03 | experiment 03: refuted | E03-H01/H02 (malformed and case-variant bodies 201), E03-H07 (undocumented max_seconds 201) |
| E03-H01 | trailing garbage / two concatenated objects → 201, one job | code api.go:53 | 400 | request shape / E03 | experiment 03: confirmed | E03-F01 |
| E03-H02 | case-variant names bound to documented fields; last variant wins | code api.go:49-52 (Go json) | 400 unknown field | request shape / E03 | experiment 03: confirmed | E03-F02 |
| E03-H03 | lesson-01 range/enum/batch fixes hold | prior record D05/D06/D26/U05-U07 | any 2xx / partial insert | regression / E03 | experiment 03: confirmed | regression only, notes 04:17:13Z |
| E03-H04 | delay_seconds 9223372037 → queued (wrap); 9223372036 → scheduled | code enqueue.go:53-55 | 400 / scheduled | numbers / E03 | experiment 03: confirmed | E03-F03 |
| E03-H05 | negative delay → queued, run_at in past | code enqueue.go:53-59 | 400 | numbers / E03 | experiment 03: confirmed | not a bug (spec silent); Q-E03-2 |
| E03-H06 | int4 overflow → 409 on create, 500 on PATCH | code queues.go:46-49, :93-94 | 400 | boundaries / E03 | experiment 03: confirmed | E03-F04 |
| E03-H07 | backoff.max_seconds accepted (also negative), no effect | code queues.go:20,46; grep | 400 or capped | unknown field / E03 | experiment 03: confirmed | E03-F07 |
| E03-H08 | backoff overflow → run_at in the past, immediate retry | code jobs.go:17-22, nack.go:22-26 | run_at future | config × count / E03 | experiment 03: confirmed | E03-F06 |
| E03-H09 | retention_hours 1e8 + terminal job → Retain errors each tick, health 503, other queues not retained | code expire.go:66, api.go:148-154 | 400 or Retain works | config × loop / E03 | experiment 03: confirmed | E03-F05 |
| E03-H10 | PATCH webhook partial clears omitted member | code queues.go:85-92 | member kept | presence / E03 | experiment 03: confirmed | E03-F08 |
| E03-H11 | ~1.1 M-char name → 201 but unaddressable (431) | code queues.go:38 (no length bound) + Go header limit | 400 on create / GET 200 | string length / E03 | experiment 03: refuted | create → 409 at the unique index; all accepted names addressable (code-reading contradiction; folded into E03-F04) |
| E03-H12 | payload: \u0000 / 1e200000 → 400; jsonb normalises 1E2, -0, dup keys | code enqueue.go:60-70 + jsonb | stored exactly | numbers / E03 | experiment 03: confirmed | not bugs; Q-E03-3, Q-E03-5 |
| E03-H13 | >2 MiB batch → 400 request body too large | code api.go:50 | 201 | size / E03 | experiment 03: confirmed | not a bug (spec silent); Q-E03-4 |
| E03-H14 | duplicate idempotency key in one batch → ids [x,x], one job | code enqueue.go:70-73 | two jobs | selection / E03 | experiment 03: confirmed | holds |
| E03-H15 | per-job max_attempts 2147483648 → 400 with pq text | code enqueue.go:81-84 | 201 / 5xx | boundaries / E03 | experiment 03: confirmed | holds (message leak = E01 Q3) |
| E03-H16 | kind " " → 201; invalid UTF-8 kind → U+FFFD 201 | code enqueue.go:29; Go json | 400 | strings / E03 | experiment 03: confirmed | UTF-8 part folded into E03-F01; whitespace spec silent |
| E03-H17 | nack delay_seconds 9223372037 → run_at 1734, immediate retry | sibling of E03-H04; nack.go:23-24 | future / 400 | sibling / E03 | experiment 03: confirmed | E03-F03 sibling |
| E03-H18 | heartbeat extend 9223372037 → lease expires at once | sibling; jobs.go:84 | 400 / future | sibling / E03 | experiment 03: confirmed | E03-F03 sibling (attempt consumed) |
| E03-H19 | PATCH trailing garbage + MAX_ATTEMPTS → 200 applied | sibling; api.go:49 | 400 | sibling / E03 | experiment 03: confirmed | E03-F01/F02 siblings |
| E03-H20 | ack with LEASE_TOKEN/RESULT → 200 | sibling; api.go:49 | 400 | sibling / E03 | experiment 03: confirmed | E03-F02 sibling (first attempt was a FIFO dead end) |
| E03-H21 | raw lease bytes keep 1.50 and 20-digit decimals | code jobs.go:65 | rounded | regression / E03 | experiment 03: confirmed | D08 fix holds |
| E03-H22 | webhook ftp / no-host / space URL → 400; "" → 201 | lesson-02 U10 fix | any 2xx for invalid | regression / E03 | experiment 03: confirmed | holds |
| Q-E03-1 | should the API refuse non-JSON Content-Type? | E03 04:16:44Z | spec says clients only | spec decision | open |  |
| Q-E03-2 | should negative delay_seconds be rejected? | E03-H05 | spec defines sign | spec decision | open |  |
| Q-E03-3 | should valid-JSON payloads that jsonb cannot store (U+0000, lone surrogate, 1e200000) be documented limits with a service message? | E03-H12 | spec covers it | spec decision | open |  |
| Q-E03-4 | document the 2 MiB request cap; 413 / `payload_too_large` instead of `invalid_json`? | E03-H13 | spec covers it | spec decision | open |  |
| Q-E03-5 | storage amplification of exponent numbers (9-byte `1e400` → 401 digits) — bound? | E03-H12 | bounded by spec | risk / later | open | deliberately not stress-probed |
| H53 | schedule create/PATCH bodies share E03-F01/F02 leniency and payload path | E03 C15 sibling | 400 | variation / next (#offcharter) | open |  |
| H54 | lease `max_jobs` and list `limit`/`cursor` silently coerce out-of-range values (limit 1000 → 100) | code enqueue.go listJobs `limit` default branch | 400 or 500 cap | variation / next (#offcharter) | open |  |
| H55 | enqueue maps real DB faults (outage) to 400 invalid_job | E03 C17 sibling, enqueue.go:81-84 | 503 | ecosystem / next | open |  |
| H56 | queue lease_seconds 2147483647 lease never expires within any test horizon; heartbeat cannot shorten it | E03 variables row | expiry | config / next | open |  |

Allowed sources: recon, spec silence, code reading, model cell, surprise and risk headline. Status: `open` or `experiment NN: confirmed`, `experiment NN: refuted`, `experiment NN: needs more`.

## Experiment 04 (judge-it) — `explore-it/lesson-03`

Rules R31–R43 were stated as hypotheses with predictions and refutations at 12:44Z, before any probe
([E04 notes](experiments/04-judge-it/notes.md)); siblings S1/S2 at 12:49:30Z before their probe.

| Id | Statement / prediction | Source and evidence pointer | Would be wrong if | Probe family / experiment | Status | Outcome evidence |
|---|---|---|---|---|---|---|
| H-E04-C | Charter: "the code honours every rule on my list" | charter 04 (code reading of nack.go, stats, lease query) | any rule violated | oracle / E04 | experiment 04: refuted | R35, R36, R38, R39, R40 violated |
| R31 | No job leased by two workers at once | SPEC; jobs.go:57 | duplicate id across parallel leases | concurrency / E04 | experiment 04: confirmed | P4 |
| R32 | Attempts never exceed max_attempts | SPEC; nack.go:19 | lease after dead returns the job | sequence / E04 | experiment 04: confirmed | P2, P10 |
| R33 | FIFO within priority holds; an app-clock step back would invert it | SPEC; jobs.go:37 | batch/priority order wrong, or no inversion | ordering / E04 | experiment 04: confirmed | P9, P12; inversion judged lab-only |
| R34 | Stats equal a direct psql count at rest | SPEC; stats.go | mismatch | approximation / E04 | experiment 04: confirmed | P5 |
| R35 | Negative nack delay precedes failure; ≤ −9223372037 wraps to future | code nack.go:23-25, internal/api/jobs.go:67 | 400 or bounded | numbers / E04 | experiment 04: confirmed | E04-F03 |
| R36 | Mux-level 404/405 are plain text, not envelope | code api.go:119-172 | JSON envelope | errors / E04 | experiment 04: confirmed | E04-F01 |
| R37 | A cancelled job never runs | SPEC; operator.go:5-6 | leased or holder verb 200 | state / E04 | experiment 04: confirmed | P5 |
| R38 | After retry, attempts renumber from 1; history interleaves | code operator.go:13, 001.sql:7 | unique monotone n | sequence / E04 | experiment 04: confirmed | E04-F02 |
| R39 | Heartbeat/ack after expiry but before the sweep succeed | code jobs.go:97, ack.go:10 | 409 | timing / E04 | experiment 04: confirmed | E04-F04 |
| R40 | Requeued-by-expiry job reports age including leased time | code stats.go:18, expire.go:41 | age ≈ 0 | approximation / E04 | experiment 04: confirmed | E04-F05 |
| R41 | Retry of dead: immediately leasable, attempt 1 | SPEC; operator.go:12-14 | not leasable / attempt ≠ 1 | state / E04 | experiment 04: confirmed | P2 |
| R42 | Backoff exp base·2^(n−1), linear base·n | SPEC; jobs.go:25-35 | other waits | sequence / E04 | experiment 04: confirmed (exponential); linear needs more | P2, P10 |
| R43 | 404 missing, 409 conflict for job verbs | SPEC | other class | errors / E04 | experiment 04: confirmed | P1, P5 |
| H-E04-S1 | Nack after expiry before sweep → 200, recorded as failed | sibling of E04-F04 | 409 | timing / E04 | experiment 04: confirmed | P11 |
| H-E04-S2 | OPTIONS / wrong method on /v1/jobs/{id} also plain text | sibling of E04-F01 | JSON | errors / E04 | experiment 04: confirmed | P11 |
| H-E04-08 | Ordering by app-supplied `created_at` inverts FIFO on a real wall-clock step back or between skewed API replicas | P9 lab inversion; jobs.go:37 | id order used / no inversion with skewed replicas | ordering / later (risk) | open | lab-only evidence |
| H-E04-09 | Stats are three statements without a snapshot; under concurrent transitions sums can disagree with any single instant | code stats.go:9-23 | identical sums under load | approximation / later | open | not probed |
| H-E04-10 | Ack is not transactional (UPDATE, then INSERT attempts, then Notify): a fault after the UPDATE reports 409 for a succeeded job | code ack.go:10-18; internal/api/jobs.go:87 | 5xx or atomic | ecosystem / later | open | needs DB fault injection |
| H-E04-11 | `schedules.go:158 v["kind"].(string)` panics on a NULL kind | code read | no panic | #offcharter | open | |
| Q-E04-1 | Default backoff has no cap: 30.34 d at attempt 20 (lesson-03 A1) — should a default cap exist? | P10 | spec decision | spec decision | open | |
| Q-E04-2 | Does "readiness" in queued age mean the time a job last became queued? | E04-F05 | spec decision | spec decision | open | |
| Q-E04-3 | Is there a lease grace window after `lease_expires_at`? | E04-F04 | spec decision | spec decision | open | |
| Q-E04-4 | Should `throughput_last_hour` ("completed jobs") include dead/cancelled? Code counts succeeded only | stats.go:21 | spec decision | spec decision | open | |

## Experiment 05 (sequence-it) — `explore-it/lesson-04`

Stated before their probes in [E05 notes](experiments/05-sequence-it/notes.md) (13:22:19Z table; 13:24:30Z; 13:29:40Z;
13:34Z; 13:33Z walker v2). Invariants I1–I13 are listed in notes and [rules.md](models/rules.md) E05 revision.

| Id | Statement / prediction | Source and evidence pointer | Would be wrong if | Probe family / experiment | Status | Outcome evidence |
|---|---|---|---|---|---|---|
| H-E05-C | Charter: "every verb either does what it says from any starting point or is refused" | charter 05 (handler reading) | a sequence leaves a job in a state the spec calls impossible | sequence / E05 | experiment 05: refuted | P6: max 2 job delivered 4× as attempt 1 (E05-F01) |
| H-E05-01 | purge (scheduled, key K) → enqueue delay with K → new id, scheduled | queues.go:169 | old id / scheduled kept | sequence / E05 | experiment 05: confirmed | P1 |
| H-E05-02 | pause → clock +1 h → resume: lease [] while paused; delayed job promoted while paused; 0 schedule catch-up | fire.go:74, expire.go:137, jobs.go:53 | lease while paused or >1 catch-up | sequence+clock / E05 | experiment 05: confirmed | P2 |
| H-E05-03 | lease → purge → nack: purge deletes 0; nack 200 → retry_wait | queues.go:169 | leased purged / nack 409 | sequence / E05 | experiment 05: confirmed | P3 |
| H-E05-04 | cancel→retry 409; dead→cancel 409; succeeded→retry 409 | operator.go | any 200 | sequence / E05 | experiment 05: confirmed | P4 |
| H-E05-05 | lease → cancel keeps lease fields | operator.go:6 | cleared | sequence / E05 | experiment 05: confirmed | P5 (E05-F02) |
| H-E05-06 | (lease → DELETE worker) ×(max+2): attempt 1 each, attempts 0, never dead | api/jobs.go:126 | attempt rises / dead | sequence / E05 | experiment 05: confirmed | P6 (E05-F01) |
| H-E05-07 | lease expires while paused → dead (max 1) | expire.go:87 | stays leased | sequence / E05 | experiment 05: confirmed | P7 |
| H-E05-08 | cancelled scheduled / retry_wait stay cancelled past due | expire.go:137 | promoted | sequence+clock / E05 | experiment 05: confirmed | P8 |
| H-E05-09 | released holder's token → heartbeat 409 | jobs.go:100 | 200 | sequence / E05 | experiment 05: confirmed | P6 |
| H-E05-10 | walk finds only I4/I8 (worker delete) and I7c classes | above | other invariant fires | walk / E05 | experiment 05: refuted | I7 fired on **succeeded** rows (ack keeps lease fields), not predicted; no other class |
| H-E05-11 | after one counted failure, worker-delete release reuses attempt 2 | api/jobs.go:126 | attempt 3 | isolation / E05 | experiment 05: confirmed | isolate.txt |
| H-E05-12 | max 1: lease → delete worker → lease succeeds | same | empty | isolation / E05 | experiment 05: confirmed | isolate.txt |
| H-E05-13 | delete holder of a cancelled job → cancelled row unchanged | state filter | changed | sibling / E05 | experiment 05: confirmed | isolate.txt |
| H-E05-14 | cancelled in-flight delivery has no attempt row | operator.go:6 | row exists | sibling / E05 | experiment 05: confirmed | isolate.txt |
| H-E05-15 | after an expiry, worker-delete release is logged `lease_expired` | 002.sql trigger | `queued` | observation / E05 | experiment 05: confirmed | f02-f03.txt (E05-F03) |
| H-E05-16 | ack keeps lease fields on succeeded | ack.go:25 | null | sibling / E05 | experiment 05: confirmed | f02-f03.txt (E05-F02) |
| H-E05-17 | nack retry=false clears lease fields | nack.go:60 | kept | sibling / E05 | experiment 05: confirmed | f02-f03.txt |
| H-E05-18 | after an expiry, ack and cancel are logged `lease_expired` | 002.sql trigger | `succeeded`/`cancelled` | sibling / E05 | experiment 05: confirmed | f03-sibling.txt (E05-F03) |
| H-E05-19 | holder-biased walks: ≥10 successful holder verbs/seed; violations only E05-F01/F02 classes | walk v2 | other invariant | walk / E05 | experiment 05: confirmed (on its refutation condition; coverage sub-prediction ≥10 missed for seed 31: 7) | walk-b31/b32.txt |
| H-E05-20 | Worker nack with error text exactly `lease expired` is logged as an expiry | 002.sql | labelled by state | sibling / later | open | not probed |
| H-E05-21 | A third-party worker that deregisters on SIGTERM while executing turns a crash-looping poison job into unbounded redelivery (E05-F01 in the field) | E05-F01 | job reaches dead | ecosystem / later | open | sample worker does not deregister (code) |
| Q-E05-1 | Should worker deregistration count the released lease as an attempt (like expiry), or record a distinct non-counted "released" attempt with a cap? | E05-F01 | spec decision | spec decision | open | |
| Q-E05-2 | Should terminal rows retain `leased_by` (who completed) while clearing token/expiry? | E05-F02 | spec decision | spec decision | open | |
| T-E05-1 | Testability: `/admin/clock` has no "offset, keep ticking" mode; advance/set freeze semantic time until reset, so real-time lease expiry and clock jumps cannot be combined | clock.go:182-191 | — | testability request | open | minimal hook: `{"action":"offset","seconds":n}`; validation: lease_seconds 2 expires 2 s after an offset jump |
| #offcharter | `GET /v1/jobs/{id}` returns the live `lease_token` to any holder of the namespace key, so any client can ack another worker's lease | P5 GET output | token hidden | boundary / later | open | observed in output, not judged |

## Experiment 06 (relate-it) — `explore-it/lesson-05`

Stated before their probes in [E06 notes](experiments/06-relate-it/notes.md) (15:31Z table, 15:33Z, 15:35Z, 15:37Z, 15:38Z,
15:39Z, 15:40Z, 15:41Z).

| Id | Statement / prediction | Source and evidence pointer | Would be wrong if | Probe family / experiment | Status | Outcome evidence |
|---|---|---|---|---|---|---|
| H-E06-C | Charter: "deleting an owner leaves orphans" (no FK with ON DELETE) | charter 06 (migration reading) | code guards or cascades | entity / E06 | experiment 06: refuted for queue→job/schedule (409, FK); holds only for job→attempt | P1, P9b, P10 (E06-F07) |
| H-E06-01 | DELETE queue: any job or schedule → 409; empty 204; missing 404; foreign 404 | queues.go:109-151 | 204 with dependents | entity / E06 | experiment 06: confirmed | P1 |
| H-E06-02 | pending outbox row delivered after its job and queue are deleted | SPEC:163 | row dropped | entity / E06 | experiment 06: confirmed | P8e |
| H-E06-03 | DELETE schedule missing / foreign id → 204 | schedules.go:136-143 | 404 | entity / E06 | experiment 06: confirmed | P2 (E06-F01) |
| H-E06-04 | schedule accepts non-object payload; run 400; cron fire stores it | schedules.go:42-60, fire.go:76 | create 400 / fired `{}` | data / E06 | experiment 06: confirmed | P2b (E06-F02) |
| H-E06-05 | sample worker echo rounds `9007199254740993` | client.go:57 | exact | data / E06 | experiment 06: confirmed | P4 (E06-F03) |
| H-E06-06 | manual ack keeps digits to webhook | api.go:53-62 | digits lost | data / E06 | experiment 06: confirmed | P4 |
| H-E06-07 | NUL in ack result / nack error → 409; enqueue → 400 driver text | ack.go, api/jobs.go:88 | 2xx | data / E06 | experiment 06: confirmed | P6 (E06-F05) |
| H-E06-08 | deleting a running sample worker's row: release counted; process polls 404 forever | worker.go:34-46 | re-registers/exits | entity / E06 | experiment 06: confirmed | P7 |
| H-E06-09a–d | queue PATCH: max_attempts per job; webhook snapshot; url "" stops rows; retention retroactive | SPEC; Notify; Retain | re-targeted / changed | entity / E06 | experiment 06: confirmed | P8 |
| H-E06-10 | enqueue ‖ DELETE queue: no orphans; occasional 4xx/5xx with FK text | LOCK + FK | orphan / only 201,404 | race / E06 | experiment 06: confirmed | P9b (E06-F06) |
| H-E06-11 | purge leaves attempts rows | queues.go:170 | rows removed | entity / E06 | experiment 06: confirmed | P10 (E06-F07) |
| H-E06-12 | cron-fired `[1,2]` poisons sample worker: 3 expiries → dead "lease expired" | worker.go:21 | acked/nacked | data / E06 | experiment 06: confirmed | P3 (E06-F02) |
| H-E06-13 | jsonb normalises and Go HTML-escapes; values equal | jsonb; encoding/json | value changes | data / E06 | experiment 06: confirmed | P4 (not a defect) |
| H-E06-14 | qctl job get displays rounded number | qctl main.go:277-280 | exact | data / E06 | experiment 06: confirmed | P5 (E06-F04) |
| H-E06-15 | qctl enqueue stores rounded number | qctl main.go:52-57 | exact | data / E06 | experiment 06: confirmed | P5 (E06-F04) |
| H-E06-16 | schedule create ‖ DELETE queue → some 500 FK text | schedules.go:42-57 | only 201/404 | race / E06 | experiment 06: confirmed | P9c (E06-F06) |
| H-E06-17 | batch ‖ DELETE queue → some 400 FK text | enqueue.go:92-124 | only 201/404 | race / E06 | experiment 06: refuted (0 of 320; sampling) | P9c |
| H-E06-18 | DELETE and retention leave attempts rows; ids not reused | api/jobs.go:32, expire.go:90-92 | rows removed | entity / E06 | experiment 06: confirmed | P10 (E06-F07) |
| H-E06-19 | idempotency key freed after delete → new id | enqueue.go:74 | old id | entity / E06 | experiment 06: confirmed | P10 (not a defect) |
| H-E06-20 | Lease on a queue deleted between lookup and lease tx → 500 `sql: no rows` | jobs.go:42-50 | 404 | race / later | open | not probed |
| H-E06-21 | Sample worker acking a result with U+0000 (http kind) → 409 logged `completed`, job dies by expiry | E06-F05 + worker.go | job succeeds | data / later | open | not probed |
| H-E06-22 | `createWorker` with NUL in name/kinds → 500 database_error | api/jobs.go:107 | 400 | data / later | open | not probed |
| Q-E06-1 | Should the sample worker exit non-zero (so `restart: on-failure` re-registers) or re-register on `worker_not_found`? | P7 | spec decision | spec decision | open | |
| Q-E06-2 | Should attempts (and delivered outbox rows) follow job retention, or be documented history like `job_events`? | E06-F07 | spec decision | spec decision | open | |
| T-E06-1 | Testability: the worker's decode-failure log line has no job id; add the raw job ids to the error (or a `decode_failed` event) so a poison job can be identified | P3 | — | testability request | open | validation: fire `[1,2]` schedule → log names the job |
| T-E06-2 | Testability: no way to hold the queue-delete lock window deterministically; a lab hook (e.g. `QUEUELET_TEST_DELETE_PAUSE_MS`) would make E06-F06 reproducible on demand | P9 | — | testability request | open | validation: enqueue during pause → 404 |

## Experiment 07 (state-it) — `explore-it/lesson-06`

Stated before their probes in [E07 notes](experiments/07-state-it/notes.md) (≈18:03Z block, ≈18:08:3xZ, ≈18:09:2xZ,
≈18:13:2xZ). H-E07-08, -09 and -14 were written after their probes ran and are kept but not counted as stated.

| Id | Statement / prediction | Source and evidence pointer | Would be wrong if | Probe family / experiment | Status | Outcome evidence |
|---|---|---|---|---|---|---|
| H-E07-01 | Lease A → clock past expiry → sweep → lease B: A's ack, nack, heartbeat 409; B acks; attempts n1 expired(A), n2 succeeded(B) | ack.go:14, nack.go:13, jobs.go:100; SPEC:107 | any A verb 200 or changes the job | state / E07 | experiment 07: confirmed | P1 |
| H-E07-02 | Same worker re-leases after expiry: first token ack 409, second 200 | token guard | first token accepted | state / E07 | experiment 07: confirmed | P1 |
| H-E07-03 | DELETE worker ‖ lease by that worker leaves a job leased by the deleted worker | jobs.go:57, api/jobs.go:160, expire.go:20; SPEC:93 | 0 orphans in ≤40 trials | race / E07 | experiment 07: confirmed | P2, P3 (E07-F01) |
| H-E07-04 | Purge ‖ leases leaves pre-existing jobs leased without the purge flag; nack → retry_wait | queues.go:170; SPEC:54-56 | every surviving leased job flagged | race / E07 | experiment 07: confirmed | P2 purgelease (E07-F02) |
| H-E07-04b | Purge ‖ nacks(retry) leaves jobs retry_wait | queues.go:170, nack.go | all dead/deleted | race / E07 | experiment 07: confirmed | P2 purgenack (E07-F02) |
| H-E07-05 | Heartbeat extend 1 s on a 30 s lease shortens it to now+1 | jobs.go:100 | expiry unchanged | state / E07 | experiment 07: confirmed (spec silence → Q-E07-1) | P1 |
| H-E07-06 | Purge-killed dead job → retry → lease → nack(retry) → retry_wait | operator.go:13 | dead | state / E07 | experiment 07: confirmed | P1 |
| H-E07-07 | Holder's concurrent ack+ack / ack+nack: exactly one 200, one attempt row | guarded UPDATE / FOR UPDATE | two 200s or two rows | race / E07 | experiment 07: confirmed (40 pairs) | P2 dup |
| H-E07-08 | (post-hoc) sequential lease→delete releases; one-job race orphans; zombie heartbeats/acks; second DELETE 404 | P2 | — | race / E07 | not counted; observed as described | P3 |
| H-E07-09 | (post-hoc) sequential purge control flags and kills | A2 | — | state / E07 | not counted; observed as described | P4 |
| H-E07-10 | Purge ‖ worker-delete release leaves jobs queued | expire.go:20 + queues.go:170 | all dead/deleted in 10 | race / E07 | experiment 07: confirmed (9/10) | P5 (E07-F02) |
| H-E07-11 | Purge ‖ expiry sweep leaves jobs queued | expire.go:12 + queues.go:170 | all dead/deleted in 8 | race / E07 | experiment 07: confirmed (7/8) | P5 (E07-F02) |
| H-E07-12 | Expired, pre-sweep: B's lease [] then B gets it ≤1.5 s; A ack 409 | jobs.go:60 guard; loop tick api.go:221 | immediate or >1.5 s | state+clock / E07 | experiment 07: confirmed | P6 r1, r2 |
| H-E07-13 | retry_wait due, pre-promote: B's lease [] then ≤1.5 s | expire.go:69; api.go:221 | immediate or >1.5 s | state+clock / E07 | experiment 07: **refuted in r1** (3.04 s, 1,110-row sweep stalled the loop; S6), held in r2 on a quiet stack | P6 |
| H-E07-14 | (post-hoc) one leased job suffices for purge ‖ worker-delete | — | — | race / E07 | not counted; observed 6/20 | reduction |
| H-E07-15 | 7 states × 7 verbs matrix matches the guards; no side effect on rejection | guards table (notes 18:02Z) | any cell differs | state table / E07 | experiment 07: confirmed (49/49) | P7 |
| H-E07-16 | Lease on a worker deleted between the API's existence check and the lease tx → 500 `sql: no rows` (sibling of H-E06-20) | jobs.go:57 | 404 | race / later | open (seen once incidentally, P2 trial 5; not driven deliberately) | E07-F01 |
| H-E07-17 | A sweep of many expiries (≈1,000) delays Promote/cron for seconds because the loop is sequential and Expire is one unbounded transaction | P6 r1, api.go:230, expire.go:12 | promotion stays ≤1 tick under 1,000 expiries | load / later | open (#offcharter) | S6 |
| Q-E07-1 | Should heartbeat be allowed to shorten a lease (`extend_seconds` measured from now) or only lengthen it? | P1 H05 | spec decision | spec decision | open | |
| Q-E07-2 | Should a nack 409 say `state conflict` like ack/heartbeat rather than the driver text `sql: no rows in result set`? | P1 H01 | spec decision | #offcharter | open | |
| T-E07-1 | Testability: job_events carry only application time; with the frozen lab clock there is no real-time commit stamp or loop-step duration, so sweep length (S6, E07-F02 window) cannot be measured | P6 r1 | — | testability request | open | minimal hook: `committed_at clock_timestamp()` on job_events or a `/metrics` gauge per loop step; validation: replay P6 r1, see the 1,110-row sweep's duration |
| T-E07-2 | Testability: races in E07-F01/F02 are probabilistic; a lab pause hook inside the lease / purge / release transactions (e.g. `QUEUELET_TEST_TX_PAUSE_MS=lease:50`) would make the overlap deterministic (generalises T-E06-2) | P2–P5 | — | testability request | open | validation: one-job purge ‖ worker-delete → queued 20/20 |

## Experiment 08 (surround-it) — `explore-it/lesson-07`

Stated before their probes in [E08 notes](experiments/08-surround-it/notes.md) (hypothesis table written
19:21Z; H-E08-14 19:25:45Z; H-E08-15..17 19:29Z; H-E08-18 19:30:45Z; H-E08-19 ≈19:34:10Z; H-E08-20 ≈19:34:30Z;
H-E08-21 ≈19:35:05Z). Regression holds are not discoveries.

| Id | Statement / prediction | Source and evidence pointer | Would be wrong if | Probe family / experiment | Status | Outcome evidence |
|---|---|---|---|---|---|---|
| H-E08-C | Charter: "every route is namespace-scoped and the loop survives a database restart" | charter 08 (auth middleware and loop reading) | a route crosses tenants, or a DB event the loop/service does not survive | ecosystem / E08 | experiment 08: confirmed for routes and for a DB restart of a running API; refuted per neighbour for start-without-DB (E08-F01), frozen DB (E08-F04) and the shipper (E08-F02) | P1–P8 |
| H-E08-01 | globex-key on acme job ids → 404 on all nine job routes; acme job unchanged | api/jobs.go:200 ownJob | any 2xx/409 or change | tenant / E08 | experiment 08: confirmed (D21 regression) | P1 |
| H-E08-02 | foreign worker id in lease, foreign worker delete, foreign schedule routes → 404; foreign ns stats → 403 | jobs.go:174; queues.go:194 | any 2xx | tenant / E08 | experiment 08: confirmed | P1 |
| H-E08-03 | admin-key = acme tenant (200 on acme job, 404 on globex job) + stats anywhere | 001.sql:16 | otherwise | admin / E08 | experiment 08: confirmed | P1 |
| H-E08-04 | `compose restart postgres`: brief 503 auth_unavailable/health, delayed job promoted on time | api.go:149-245 | loop dead or 401 | DB / E08 | experiment 08: confirmed (regression) | P2 |
| H-E08-05 | API restart while DB down → exits and stays down after DB returns | main.go:19-22; compose (no restart policy) | api back within 30 s | DB+supervisor / E08 | experiment 08: confirmed | P3 (E08-F01) |
| H-E08-06 | partner 500 → retried every ~5 s, delivered on first 200 | ship.go:40-48 | delivered during 500 | partner / E08 | experiment 08: confirmed (D04 regression) | P5 |
| H-E08-07 | partner slow 3 s → failure at 2 s, partner records 200s; core promotion on time | ship.go:32 | delivered while slow / promotion late | partner / E08 | experiment 08: confirmed | P5 |
| H-E08-08 | ≥20 timed-out rows → later acme row never attempted for ≥60 s | ship.go:17, api.go:212 | acme attempted in 60 s | partner×tenant / E08 | experiment 08: confirmed | P5 (E08-F02) |
| H-E08-09 | hand HMAC over received bytes = header; bodies semantically equal to outbox | ship.go:22-30 | mismatch | data journey / E08 | experiment 08: confirmed (D24 regression) | P4 |
| H-E08-10 | DST gap: `30 2`/`0 2` NY skip 2027-03-14 | fire.go:20 | a run on 3/14 | clock / E08 | experiment 08: confirmed | P6 (E08-F03) |
| H-E08-11 | hourly NY across gap: 01:00 EST → 03:00 EDT | fire.go:20 | other | clock / E08 | experiment 08: confirmed | P6 |
| H-E08-12 | DST overlap: `30 1`/`0 1` NY fire twice on 2027-11-07 | fire.go:20 | once | clock / E08 | experiment 08: confirmed | P6 (E08-F03) |
| H-E08-13 | `0 9` NY: 14:00Z then 13:00Z across the gap | fire.go:20; SPEC.md:162 | 14:00Z on 3/14 | clock / E08 | experiment 08: confirmed (D23 regression) | P6 |
| H-E08-14 | sent bytes = jsonb reorder + Go HTML/U+2028 escaping; semantically equal; exact numbers; signature over sent bytes | store.go:44; ship.go:22 | semantic diff / rounding / bad signature | data journey / E08 | experiment 08: confirmed | P4 |
| H-E08-15 | 22 fast-failing (500) rows → acme delivered ≤3 s | ship.go:41 | >5 s | isolation / E08 | experiment 08: confirmed | P5b (1.4 s) |
| H-E08-16 | 5 timed-out rows → acme delivered after ≈10–12 s | ship.go:17 | never / <8 s | isolation / E08 | experiment 08: confirmed | P5b (11.1 s) |
| H-E08-17 | 20 rows to an unreachable tenant URL (partner healthy) → acme not attempted for 40 s | ship.go | delivered | sibling / E08 | experiment 08: confirmed | P5b (45.6 s, 0 attempts) |
| H-E08-18 | no tenant API action (PATCH url "", delete jobs, delete queue) ends the stall | SPEC.md:196-204 | acme attempted | recovery / E08 | experiment 08: confirmed | P5c |
| H-E08-19 | lease as another acme worker by id alone → 200, attributed to that worker | api/jobs.go:128-178 | refused | worker identity / E08 | experiment 08: confirmed | P7 (not a bug: spec silence, Q-E08-1) |
| H-E08-20 | frozen DB 12 s: auth requests hang then 200; `/health` 503 within ≈1 s | api.go:132,153 | prompt 503 on auth | DB / E08 | experiment 08: refuted (health clause: hung 11.8 s then 200) | P8 (E08-F04) |
| H-E08-21 | frozen DB 25 s: all DB-path requests incl. /health return only after unpause | P8 | any returns early | DB / E08 | experiment 08: confirmed | P8b (E08-F04) |
| H-E08-22 | A partner answering 302 (→ GET 200) is recorded as delivered without receiving the body (Go client follows redirects) | ship.go:31 default client | marked failed | partner / next | open | needs a redirect-capable partner (TR-E08-3) |
| H-E08-23 | One schedule that cannot advance starves later schedules every tick (Fire returns on first error) | fire.go:37-40 | later schedules fire | core loop / next | open | no external trigger found |
| H-E08-24 | More than 16 concurrent requests during a DB freeze exhaust the pool; requests queue behind it after unfreeze | store.go:16 | no queueing | DB / next | open | not probed |
| Q-E08-1 | What does the worker `token` (SPEC.md:91) authorise? Should lease/ack require it so attempt history attribution is trustworthy? | P7 | spec decision | spec decision | open | |
| Q-E08-2 | DST rule for non-existent and repeated local times (skip/once-after/twice)? | E08-F03 | spec decision | spec decision | open | |
| T-E08-1 | Testability: deterministic single-statement DB fault (for H-E04-10 / ack non-atomicity) | ack.go:13-26 | — | testability request | open | see notes "Not run / testability requests" |
| T-E08-2 | Testability: shipper backlog/lag metric and health signal | P5c (/health 200 during permanent stall) | — | testability request | open | validation: P5b blackhole flips it |
| T-E08-3 | Testability: hooksink records received signature (E01 TR1) and offers 302/307 modes | P4 needed a capture listener | — | testability request | open | |
