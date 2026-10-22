# Rules (named oracles)

Authority: SPEC.md unless stated. Scope: v0-inherited. Observation column cites E01 notes/defects.

| Rule | Authority | Scope | E01 observation |
|---|---|---|---|
| R1 Namespace key accesses only its namespace's resources | SPEC Auth | all /v1 | Violated for job-id endpoints (F03); holds for queues, schedules, workers, lease worker check |
| R2 Only the current lease holder may complete or extend | SPEC Worker | ack/nack/heartbeat | Violated ack (F01), nack (F02); holds heartbeat (H31) |
| R3 No job executes more than max_attempts | SPEC Worker (Never rule) | nack, expiry | Violated (F04, F05) |
| R4 Expired lease records an unsuccessful attempt | SPEC Worker | Loop/Expire | Violated (F05) |
| R5 Cancelled job is never run | SPEC Operator (Never) | cancel + any later verb | Violated via nack (F02) |
| R6 Succeeded/cancelled cannot be retried | SPEC Operator | retry | Violated (F06) |
| R7 Missing resource → 404; state conflict → 409; errors 4xx/5xx | SPEC Responses | all | Violated: lease unknown worker 200 (F21); delete missing queue 204 (F15); DB outage → 401 (F14) |
| R8 Unknown fields / invalid enums rejected | SPEC Responses | all bodies, filters | Violated (F20) |
| R9 Names ASCII [A-Za-z0-9_-]; priority 0–9; positive limits | SPEC Queues/Jobs | create, patch | Violated (F19) |
| R10 Payload numeric precision preserved | SPEC Jobs | enqueue → lease/GET | Violated (F17) |
| R11 Batch 1–500, else 400 without insertion | SPEC Jobs | batch | Violated (F16) |
| R12 FIFO by creation order within priority | SPEC Worker | lease | Violated on same-second ties (F23); held for 40 untouched jobs (H22) |
| R13 Delete queue: no orphans, conflict if nonempty | SPEC Queues | DELETE queue | Violated (F15) |
| R14 Pause stops scheduled deliveries; no unbounded catch-up | SPEC Queues | Fire | Pause holds; catch-up violated (F24) |
| R15 PATCH schedule: next fire reflects config; disabled never enqueues | SPEC Schedules | PATCH, Fire | Violated (F07, F09) |
| R16 Civil time incl. offset changes | SPEC Schedules | Next | Violated via harness (F27) |
| R17 Webhook HMAC over body; success only 2xx; retried; data agrees or marks truncation | SPEC Notifications | Ship, Notify | Violated (F10, F11, F12); transport-error retry holds (H34) |
| R18 Stats by exact state | SPEC Notifications/ops | stats | Violated (F18); age reference silent (F26) |
| R19 Loop recovers after DB restart | SPEC ops | Loop | Violated (F08) |
| R20 Idempotency key returns original job | SPEC Jobs | enqueue | Holds (H29) |
| R21 run_at XOR delay_seconds | SPEC Jobs | enqueue | Holds (H36) |
| R22 Existing jobs retain max_attempts on queue PATCH | SPEC Queues | PATCH | Holds (H39) |
| R23 CLI exits non-zero on failure | convention (not in SPEC) | qctl | Violated (F22) |
| R24 Partner latency should not gate core progress | approximation (no SPEC bound) | Loop | Violated (F13) |

## E02 revision — `explore-it/lesson-01` ([E02 notes](../experiments/02-make-it-visible/notes.md)); ★ = new observation

| Rule | Authority | Scope | E02 observation |
|---|---|---|---|
| ★ R3 No job executes more than max_attempts | SPEC Worker | expiry via crash | Holds for crash (3/3, 4/4) and fail (3/3). Freeze case: 2 executions within max 3 (H14) |
| ★ R4 Expired lease records an unsuccessful attempt | SPEC Worker | Expire | Holds (H01, H13); started_at = finished_at = expiry tick (Q1 persists) |
| ★ R17 Webhook success only on 2xx; retried; at least once | SPEC Notifications | Ship | 2xx rule holds (H05/H06); retry unbounded (H07); partner 2xx after 2 s counted as failure (E02-F03, spec) |
| ★ R19 Loop recovers after DB restart; health reports degradation | SPEC ops | Loop, /health | Holds (H15); health lags one tick |
| ★ R25 One tenant's configuration must not stop another tenant's processing | approximation of SPEC:10 tenant isolation | shared shipper | Violated (E02-F02) |
| ★ R26 Invalid command values rejected | SPEC:17 | queue webhook URL | Violated for unparsable URL on create and PATCH (E02-F02, H11d) |
| ★ R27 README sample worker runs documented kinds | README "Worker mode" | sample worker | Violated for `fail3` (E02-F01) |

## E03 revision — `explore-it/lesson-02` ([E03 notes](../experiments/03-vary-it/notes.md)); ★ = new observation

| Rule | Authority | Scope | E03 observation |
|---|---|---|---|
| ★ R8 Unknown fields / invalid enums rejected | SPEC:17 | all bodies | Exact unknown names rejected (holds); case-variant names accepted (E03-F02); malformed JSON text with trailing data accepted (E03-F01); undocumented `backoff.max_seconds` accepted and ignored (E03-F07) |
| ★ R9 Documented ranges | SPEC:25-26, :46 | create, PATCH, enqueue | Lower bounds and priority hold; upper int limits misreported (E03-F04); delay beyond Duration range wraps (E03-F03); large retention accepted but breaks Retain (E03-F05) |
| ★ R10 Payload numeric precision | SPEC:47 | enqueue → GET/lease | Holds: exact digits in raw lease bytes (H21); jsonb normalises representation (1E2→100, −0→0) without value loss |
| ★ R11 Batch 1–500 atomic | SPEC:54-56 | batch | Holds for count/position; additional undocumented 2 MiB body cap |
| ★ R7 Error classes | SPEC:13-16 | queue create/PATCH | 409 used for out-of-range and index-size errors; 500 on PATCH (E03-F04) |
| ★ R28 Retry wait follows configured backoff (nonnegative, nondecreasing) | SPEC:27-28 | nack | Violated at attempt ≥32 with defaults (E03-F06); cap field ignored (E03-F07) |
| ★ R29 A future delay/extension is never honoured as "now" | SPEC:48-49; worker heartbeat | enqueue, nack, heartbeat | Violated above 9 223 372 036 s (E03-F03) |
| ★ R25 One tenant's configuration must not stop another tenant's processing | approximation of SPEC:10 | Retain | Violated: one acme queue blocks globex retention (E03-F05) |
| ★ R30 PATCH omissions keep current values | approximation (RFC 7396) + SPEC:33-34 | queue PATCH | Holds top-level; violated inside `webhook` (E03-F08) |

## E04 revision — `explore-it/lesson-03` ([E04 notes](../experiments/04-judge-it/notes.md)); ★ = new observation

Written as Never/Always rules before probing (notes 12:44Z). Oracle table: rule / authority / bound / enforcing source /
prediction / observation / judgment. C = comparable system, as recalled and not verified in this run.

| Rule | Authority / bound | Enforcing source | Predicted | Observed (E04) | Judgment / uncertainty |
|---|---|---|---|---|---|
| ★ R31 A job is never leased by two workers at once | SPEC "Leasing is atomic" | jobs.go:57 `FOR UPDATE SKIP LOCKED` | holds | 30 parallel leases × 2 workers on 40 jobs: 40 distinct, 0 duplicates (P4) | holds |
| R32 Attempts never exceed max_attempts | SPEC Worker | nack.go:19, expire.go:37 | holds | max 2 → dead after 2; max 21 → dead at 21; lease after dead empty (P2, P10) | holds (regression) |
| ★ R33 Within a priority, FIFO by creation order | SPEC Worker; C SQS FIFO (sequence numbers per group) | `LeaseOrder` jobs.go:37 | holds; app-clock step back inverts | batch order and priority 9,9,5,5,0 hold (P9, P12); clock −60 s → later job leased first (P9) | holds in real time; inversion needs the lab clock to step back — not a defect here; risk H-E04-08 |
| ★ R34 Stats always equal a direct count | SPEC "counts by exact state"; psql count, exact at rest | stats.go:8-15 | holds | queue and namespace views equal psql counts across 6 states; throughput = psql (P5) | holds at rest; concurrent consistency untested (three statements, no snapshot) |
| ★ R35 A retry comes after the failure, within a bounded delay | SPEC backoff line ("never scheduled before the failure", cap 2^31−1 s) | nack.go:22-25 computed wait bounded; override unbounded below | override can precede failure / wrap | −5 s → before failure; −9223372037 → +292 y; enqueue same (P3) | **violated**: [E04-F03](../defects/E04-F03-negative-delay-wraps-and-precedes-failure.md) |
| ★ R35b Default retry delay is bounded by something operational | — (spec silence; lesson 03 A1); C SQS 12 h visibility max | jobs.go:25-35 | 30 d at attempt 20 | 5·2^(n−1): 30.34 d at attempt 20, 60.68 d at 21 (P10) | spec-compliant (bound 68 y); open decision Q-E04-1, not counted |
| ★ R36 An error is always 4xx/5xx with the JSON envelope | SPEC Responses | api.go:50 `fail()` only in handlers | mux fallbacks plain text | 404/405 text/plain for wrong method/unknown path (P1, P11) | **violated**: [E04-F01](../defects/E04-F01-unmatched-route-errors-not-json.md) |
| ★ R37 A cancelled job never runs | SPEC Operator | operator.go:5-6; ack/nack/heartbeat `state='leased'` | holds | cancel leased → heartbeat/ack/nack/retry 409; cancelled queued job skipped by lease (P5) | holds |
| ★ R38 Attempt history: one numbered row per execution, numeric order = execution order | SPEC "completed attempts in numeric order"; consistency approx. | 001.sql:7 no key; operator.go:13 attempts=0 | duplicates after retry | n=1 failed, n=1 succeeded, n=2 failed (P2) | **violated** (spec silent on retry history): [E04-F02](../defects/E04-F02-attempt-history-after-retry-ambiguous.md) |
| ★ R39 An expired lease is not current: no extend/complete after `lease_expires_at` | SPEC Worker ("current lease holder"; expired lease records an attempt) | only expire.go:12 checks expiry | honoured until sweep | heartbeat/ack/nack 200 before the tick, 409 after (P6, P7, P11) | **violated** unless owner specifies a grace: [E04-F04](../defects/E04-F04-expired-lease-honoured-until-sweep.md) |
| ★ R40 Queued age measures waiting since the job became ready | SPEC "later of creation and readiness"; bound ≤ time since release | stats.go:18 uses run_at; expiry/release keep run_at | counts leased time | age 600 / 100 for a job queued 0 s (P6, P8); control correct | **violated** under the reading "readiness = became queued": [E04-F05](../defects/E04-F05-queued-age-counts-leased-time.md) |
| ★ R41 Operator retry makes a dead job immediately available with attempts 0 | SPEC Operator | operator.go:12-14 | holds | retry → next lease attempt 1 (P2) | holds |
| ★ R42 Backoff follows kind/base | SPEC Queues | jobs.go:25-35 | holds | exp base 2: 2 s; exp base 5: exact doubling to 2^20·5 (P2, P10) | holds (linear not re-probed) |
| ★ R43 404 missing / 409 state conflict / 400 invalid | SPEC Responses | handlers | holds for job verbs | 404 missing job and unknown verb; 409 for verbs on cancelled job (P1, P5) | holds; message text leaks driver error (`sql: no rows in result set`, E01 Q3) |
