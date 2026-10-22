# Job lifecycle

```text
 enqueue(future) ─► scheduled ─Promote─► queued ◄─Promote── retry_wait
 enqueue(now)    ───────────────────────► queued                ▲
                                   lease │  ▲ Expire (no attempt!)│ nack(retry, attempts+1 ≤ max+…)
                                         ▼  │                    │
                                        leased ──────────────────┘
                                         │ ack → succeeded     │ nack(retry=false | attempts+1 > max) → dead
   cancel (any non-terminal) → cancelled; retry (ANY state, F06) → queued; nack (ANY state, F02) → retry_wait|dead
```

| State | Event | Should (SPEC) | Predicted (code) | Observed (E01) |
|---|---|---|---|---|
| queued | lease | leased, kind-filtered, priority→FIFO | same; ties arbitrary | O leased; FIFO tie inversion (F23) |
| queued | nack (no lease) | 409 | retry_wait | O retry_wait (F02) |
| queued | purge | deleted | deleted | O (H24) |
| scheduled/retry_wait | purge | "pending" ? | kept | O kept (F25 spec) |
| scheduled | due | queued | queued within ~1 s | O ~2 s; 13 s under webhook timeout (F13); never if loop dead (F08) |
| leased | ack (holder) | succeeded | succeeded | O (H01) |
| leased | ack (wrong token) | 409 | succeeded | O succeeded (F01) |
| leased | nack, attempts < max | retry_wait | retry_wait | O (H41) |
| leased | nack, attempts = max | dead | retry_wait | O retry_wait → re-run (F04) |
| leased | lease expiry | attempt recorded; queued or dead | queued, no attempt | O (F05) |
| leased | heartbeat wrong token | 409 | 409 | O (H31) |
| leased | cancel | cancelled | cancelled | O (H07) |
| leased | worker deleted | released | queued, no attempt | O queued (H22b) |
| succeeded | retry | 409 | queued | O queued (F06) |
| succeeded | nack | 409 | retry_wait | O (F02) |
| cancelled | nack | never run | retry_wait → queued | O re-leased (F02) |
| dead | retry | queued, attempts 0 | same | O (H40) |
| dead (via nack retry=false) | — | dead + notify | same | O (H40, H28) |
| terminal | delete | 204 | 204 | O (H40; cross-namespace F03) |
| non-terminal | delete | 409 | 409 | O (H33) |
| terminal | retention_hours pass | deleted | deleted | I (not observed) |

## Schedule lifecycle (abbrev.)
| State | Event | Should | Observed |
|---|---|---|---|
| enabled | PATCH cron/tz | next reflects config | unchanged (F07) |
| enabled | PATCH enabled=false | stops | keeps firing (F07) |
| created disabled | PATCH enabled=true | fires | next null, never fires (F07) |
| enabled, queue paused | ticks pass, resume | bounded catch-up | one job per tick (F24) |
| any | PATCH bad cron | 400 | 200; loop dies (F09/F08) |
| queue deleted | run | ? (orphan forbidden) | 404 queue_not_found (F15) |

## E02 revision — build `explore-it/lesson-01` (observed in [E02 notes](../experiments/02-make-it-visible/notes.md))

The E01 rows above describe v0. At lesson-01 the attempt-accounting arrows changed; each row below is a runtime observation
in E02 (O) unless marked I (inferred). Counted new cells are marked ★.

```text
 queued ─lease─► leased ─(worker crash / freeze; no ack)─► Expire tick: attempts+1, row 'expired'
                   │                                       ├─ n < max ─► queued ─lease─► leased (next incarnation)
                   │                                       └─ n ≥ max ─► dead ─Notify─► outbox entry (if queue has URL)
                   └─ stale ack after expiry ─► 409 (job already queued/re-leased)
 outbox: pending ─POST 2xx within 2 s─► delivered
         pending ─POST non-2xx | timeout | dial error─► pending (attempts+1, next_at = tick+5 s, forever, silent)
         pending ─http.NewRequest error─► pending, attempts 0, and **every later row is never attempted** (E02-F02)
```

| State | Event | Should (SPEC) | Predicted (code, lesson-01) | Observed (E02) |
|---|---|---|---|---|
| ★ leased | worker process exits (crash), n < max | expired attempt; released for retry | queued, attempts+1, `expired` row | O job 301 n1,n2; 304 n1–n3 (H01, H13) |
| ★ leased | worker crash, n = max | dead | dead, finished_at, Notify | O 301 dead at n3 (emails, max 3); 304 dead at n4 (reports, max 4) |
| ★ dead (via expiry) | Notify | outbox entry with state/last_error | one entry | O outbox 73 `{"state":"dead","last_error":"lease expired","result":null}`, delivered once (H13) |
| ★ leased | worker frozen past lease, then resumes and acks | only current holder completes | expired; ack 409; re-lease | O job 305: ack 409, attempt 2 succeeded, work executed twice (H14) |
| ★ succeeded (after earlier failure) | read | "last error" | last_error kept | O `last_error:"lease expired"` on succeeded job (S9, Q-E02-2) |
| ★ leased → retry_wait → queued (fail kind) | nack ×3, max 3 | dead after 3 | dead after 3 | O job 303, backoff gaps 3.2 s, 4.9 s (H09) |
| ★ queued | no registered worker advertises the kind | (not specified) | stays queued | O `fail3` job 304 stays queued 80 s (H10, E02-F01) |
| ★ outbox pending | partner 500 | retried; delivered only on 2xx | +5 s, forever | O 11 × 500 at 5–6 s, then 200 → delivered (H05–H07) |
| ★ outbox pending | partner 2xx after 3 s (> 2 s client timeout) | (spec silent) | failure, retry | O 5 partner-200 receipts, 0 delivered (H12, E02-F03) |
| ★ outbox pending | URL unparsable | (retry) | blocks shipper | O blocks all later rows across namespaces (H11, E02-F02) |
| ★ outbox pending | URL parses, scheme unsupported | retried | per-row failure | O attempts+1 per 5 s, others delivered (H11c) |
| ★ outbox pending | owning job deleted | (no orphans, by analogy with queues) | row remains | O orphan row 71 remains and keeps blocking (H11b) |
| ★ scheduled (delay) | DB restart during wait | promoted after recovery | promoted next good tick | O job 306 promoted+acked 0.5 s after due (H15; regression context) |
| leased | lease expiry | attempt recorded; queued or dead | (E01 row: no attempt) | **revised**: O recorded at lesson-01 (F05 fix holds) |
| leased | nack, attempts = max | dead | dead | O at lesson-01 (job 303) — E01 F04 row superseded |

## E04 revision — build `explore-it/lesson-03` ([E04 notes](../experiments/04-judge-it/notes.md)); ★ = new observed cell

| State | Event | Should (SPEC / rule) | Predicted (code) | Observed (E04) |
|---|---|---|---|---|
| ★ dead (via nack, max 2) | operator retry, lease, ack | succeeded; history distinguishes runs | attempts restart at 1 | O succeeded; attempts list n=1,1,2 (E04-F02) |
| ★ dead | lease | not leased | not leased | O `{"jobs":[]}` (P2) |
| ★ leased, `lease_expires_at` passed, before sweep | heartbeat | 409 (not current holder) | 200, lease extended | O 200 ×6; no expired attempt (E04-F04) |
| ★ leased, expired, before sweep | ack | 409 / or stated grace | 200 succeeded | O succeeded attempts 1 (E04-F04) |
| ★ leased, expired, before sweep | nack | expired attempt | 200 failed attempt | O attempt "failed: late nack" (E04-F04) |
| ★ leased, expired, after sweep | heartbeat | 409 | 409 | O 409; queued, attempts+1 (P7) |
| ★ leased | nack delay −5 | retry not before failure | run_at failure−5 s | O (E04-F03) |
| ★ leased | nack delay −9223372037 | ≤ bounded, monotone | run_at +292 y | O 2322-04-12 (E04-F03) |
| ★ leased | cancel, then holder heartbeat/ack/nack/retry | 409 each | 409 | O 409 ×4 (P5) |
| ★ queued (cancelled) | lease | skipped | skipped | O next job leased (P5) |
| ★ leased → queued by expiry / worker delete | stats age | ≈0 (readiness = release) | from created/run_at | O 600 / 100 (E04-F05) |
| ★ queued ×40 | 30 concurrent leases | each once | each once | O 40 distinct (P4) |
| ★ retry_wait, attempts 20 (defaults) | nack | bounded? (spec silent) | 30.34 d | O run_at +2 621 440 s (P10; A1 spec silence) |
