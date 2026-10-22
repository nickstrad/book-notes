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
