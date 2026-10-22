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
