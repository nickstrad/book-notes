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

## E05 revision — build `explore-it/lesson-04` ([E05 notes](../experiments/05-sequence-it/notes.md)); ★ = new observed cell

Sequence probes P1–P8, isolation probes and seeded random walks ([walk.sh](../experiments/05-sequence-it/walk.sh), seeds
11–13 uniform, 31–32 holder-biased; 21/22 stopped early, not counted). O = observed in E05 with the cited probe; I = inferred from code only.

```text
 queued ─lease(w)─► leased(w, attempt = attempts+1)
   ▲                  │ DELETE worker w ─► queued, attempts unchanged, no attempt row       (E05-F01: budget bypass)
   └──────────────────┘   …repeatable without bound; next lease reports the same attempt number
 leased ─ack─► succeeded   (lease_token / leased_by / lease_expires_at kept)               (E05-F02)
 leased ─cancel─► cancelled (lease fields kept; no attempt row for the delivery)            (E05-F02)
 leased ─nack / expiry / worker release─► lease fields cleared
 event console: any exit from leased after an earlier expiry is logged `lease_expired`      (E05-F03)
 pause: blocks lease only; Promote and Expire keep running; schedules advance without enqueuing
```

| State | Event (sequence) | Should (oracle) | Predicted (code) | Observed (E05) |
|---|---|---|---|---|
| ★ leased | DELETE holder worker, repeated ×4 (max 2) | ≤ max executions; attempt never reused (SPEC Worker/Jobs) | queued, attempts unchanged | O 4 deliveries all `attempt 1`, attempts `[]` (P6, E05-F01) |
| ★ leased (attempt 2) | DELETE worker → lease | next attempt 3 | attempt 2 again | O attempt 2 twice (H-E05-11) |
| ★ leased (max 1) | DELETE worker → lease → nack | dead after 1 execution | 2 deliveries | O 2 deliveries, then dead (H-E05-12) |
| ★ released by worker delete | old-token heartbeat | 409 | 409 | O 409 ×4 (P6) |
| ★ leased | cancel | cancelled, lease ended | lease fields kept | O token, leased_by, future expiry kept; attempts `[]` (P5, H-E05-14, E05-F02) |
| ★ cancelled (ex-leased) | DELETE that worker | unchanged | unchanged | O still `leased_by` deleted worker (H-E05-13) |
| ★ leased | ack | succeeded, lease ended | lease fields kept | O kept (H-E05-16, walks; E05-F02) |
| ★ leased | nack retry=false | dead, lease cleared | cleared | O cleared (H-E05-17) |
| ★ leased (after an earlier expiry) | ack / cancel / worker delete | event `succeeded` / `cancelled` / `queued` | `lease_expired` | O `lease_expired` ×3 (H-E05-15, H-E05-18, E05-F03) |
| ★ leased | purge | not pending → kept | kept | O `deleted 0`; nack 200 → retry_wait; next purge deletes it (P3) |
| ★ scheduled (idempotency K) | purge → enqueue K | new job | new id | O id 1 purged, id 2 scheduled (P1) |
| ★ cancelled | retry | 409 | 409 | O (P4) |
| ★ dead | cancel | 409 | 409 | O (P4); dead keeps future `run_at` (backoff computed) |
| ★ succeeded | retry / cancel | 409 / 409 | 409 | O (P4) |
| ★ scheduled, retry_wait | cancel → clock past run_at | stays cancelled | stays | O (P8) |
| ★ scheduled, queue paused | clock past run_at | (SPEC: pause blocks leases/scheduled deliveries) | promoted to queued | O queued while paused, lease `[]` (P2); judged not a delivery |
| ★ schedule `* * * * *`, queue paused | clock +3600, resume | bounded catch-up | 0 jobs, next advances | O 0 jobs, next_run_at 14:26 (P2) |
| ★ leased, queue paused | lease expires (max 1) | expiry still applies | dead | O dead + `expired` row (P7) |
| all | random walks (950 steps, 5 completed walks: seeds 11–13 uniform, 31–32 holder-biased) — I1 state set, I2 finished⇔terminal, I3 attempts≤max, I5 contiguous history, I6 cancelled never leased, I9 retry not before failure, I10 stats=count, I12 promote after tick, I13 sweep after tick, R2 stale token | hold | hold | O no violation (walk logs results/e05/walk-*.txt) |

Unobserved / untried here: schedules PATCH sequences, queue DELETE mid-lease, retention sweep interleavings, two API
replicas, webhooks during these sequences (`#offcharter`).

## E07 revision — build `explore-it/lesson-06` ([E07 notes](../experiments/07-state-it/notes.md)); ★ = new observed cell

Single-actor lifecycle at lesson-06 (P7, 49 cells): every verb outside its guard is 409 / `[]` with no side effect;
the table has no missing arrow. The extra arrows found are **two-actor overlaps**, drawn dashed (O = observed):

```text
 queued ─lease(W) ‖ DELETE worker W─┄┄► leased(by deleted W)   heartbeat 200 · ack 200 · DELETE W 404   (E07-F01)
 queued ─lease ‖ purge──────────────┄┄► leased, purge flag NOT set ─nack─► retry_wait (pending again)   (E07-F02)
 leased ─nack(retry) ‖ purge────────┄┄► retry_wait survives the purge                                   (E07-F02)
 leased ─worker release ‖ purge─────┄┄► queued survives the purge                                       (E07-F02)
 leased ─expiry sweep ‖ purge───────┄┄► queued survives the purge                                       (E07-F02)
 leased(A) ─exp─► [≤1 tick: leased, nobody may run it] ─sweep─► queued ─lease(B)─► leased(B); A's token 409
```

| State | Event (actors, clock) | Should (oracle) | Predicted (code) | Observed (E07) |
|---|---|---|---|---|
| leased(A)→expired→leased(B) | A ack | 409 (SPEC:107) | 409 | O 409 (P1); prior-like E02 row, not counted |
| ★ leased(A)→expired→leased(B) | A nack | 409 | 409 | O 409, B's lease intact (P1) |
| ★ leased(A)→expired→leased(B) | A heartbeat | 409 | 409 | O 409 (P1) |
| ★ leased(A, 2nd token) after expiry | ack with A's 1st token | 409 | 409 | O 409; 2nd token 200 (P1 H02) |
| ★ leased | heartbeat extend < remaining | spec silent (Q-E07-1) | expiry = now+extend | O shortened to now+1; later ack 409 (P1 H05) |
| ★ dead (purge-killed) | retry → lease → nack(retry) | retry_wait | retry_wait, flag cleared | O (P1 H06) |
| ★ queued | lease(W) ‖ DELETE W | released or 404 | leased by deleted W | O 16/20 (1 job), 28/30 (50 jobs) (P2, P3) — **E07-F01** |
| ★ leased by deleted W | W heartbeat / W ack | 409 (not a registered holder) | 200 | O 200 / 200 succeeded (P3 C3) — E07-F01 |
| ★ leased by deleted W | DELETE W again | release | 404 | O 404 (P3 C3) — E07-F01 |
| ★ queued | lease(W) ‖ DELETE W (delete first, after API check) | 404 | 500 `sql: no rows` | O once (P2 trial 5) — E07-F01 secondary |
| ★ queued | lease ‖ purge | deleted or leased+flag | leased, unflagged | O 10/10 trials, 77–170 jobs (P2) — **E07-F02** |
| ★ leased | nack(retry) ‖ purge | dead or deleted | retry_wait | O 10/10 (P2) — E07-F02 |
| ★ leased | worker release ‖ purge | dead or deleted | queued | O 9/10 (60 jobs), 6/20 (1 job) (P5) — E07-F02 |
| ★ leased | expiry sweep ‖ purge | dead or deleted | queued | O 7/8 (P5) — E07-F02 |
| ★ leased | ack ‖ ack (holder) | one 200 | one 200 | O 20/20 pairs (P2 dup) |
| ★ leased | ack ‖ nack (holder) | one 200, one attempt | same | O 20/20 pairs (P2 dup) |
| ★ leased(A), exp passed, pre-sweep | lease(B) | (spec: released by loop) | `[]`, then ≤1 tick | O `[]`, B leased attempt 2 after 0.27–1.5 s (P6) |
| ★ retry_wait, run_at passed, pre-promote | lease(B) | ≤1 tick | `[]`, then ≤1 tick | O 0.2–1.01 s quiet; **3.04 s+** behind a 1,110-row sweep (P6 r1, S6, H-E07-17) |
| ★ scheduled | lease(B), wrong-token heartbeat/ack/nack, retry | `[]` / 409 | same | O (P7) |
| ★ queued | wrong-token heartbeat/ack, retry | 409 | 409 | O (P7) |
| ★ leased | lease(B) | `[]` | `[]` | O (P7) |
| ★ retry_wait | lease(B), wrong-token heartbeat/ack/nack, retry | `[]` / 409 | same | O (P7) |
| ★ succeeded | lease(B), wrong-token heartbeat/ack | `[]` / 409 | same | O (P7) |
| ★ dead | wrong-token heartbeat/ack/nack | 409 | 409 | O (P7) |
| ★ cancelled | wrong-token heartbeat/ack, cancel | 409 | 409 | O (P7) |
| all 7 states | remaining matrix cells (cancel, delete, retry, lease where prior) | as E01–E05 | same | O (P7) — regression at lesson-06, not counted |

Unobserved / inferred here: cancel ‖ lease/nack/release, retry ‖ purge, promote ‖ purge (guard re-check predicts a
serial outcome); retention interleavings; clock set backwards (lab-only).
