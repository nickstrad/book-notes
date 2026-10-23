# Variables

| Family | Variable | Variations tried (E01) | Effect observed |
|---|---|---|---|
| identity | API key | acme, globex, admin, empty, valid-with-DB-down | job endpoints ignore namespace (F03); DB down → 401 (F14) |
| lease | token | correct, wrong, absent | ack/nack ignore it (F01/F02); heartbeat checks (H31) |
| job state | state at command time | queued, leased, succeeded, cancelled, dead | nack/retry accept any (F02/F06); cancel/delete filter correctly |
| attempts | max_attempts | 1, 2, 3, 7 | N+1 executions (F04); expiry uncounted (F05) |
| time | delay_seconds / run_at | 1, 600, 3600, past 2020, both | both → 400 (H36); past accepted (H37) → age years (F26) |
| time | created_at second bucket | same second ×10/×40 | tie order not FIFO (F23) |
| time | pause duration vs cron | 3 min on `* * * * *` | 4-job burst (F24) |
| time | zone offset | America/New_York across 2026-11-01 (harness) | frozen (F27) |
| size | batch length | 0, 4, 501 | F16 |
| size | result size | 20 B, 1520 B | truncated at 1024 (F12) |
| number | payload numbers | 2^53+1, long decimal, 20-digit | float64 rounding (F17) |
| enum | backoff.kind, state filter | zigzag, bogus | accepted (F20) |
| range | priority, name charset, PATCH max/retention | 42, -3, "a.b", "ü-q", -1, -5 | accepted (F19) |
| partner | hooksink mode | ok, 500, timeout | 500 delivered (F11); timeout stalls loop (F13) |
| infra | postgres availability | restart, stop | loop dead (F08); 401 (F14) |
| schedule | cron / timezone / enabled via PATCH | each | no recompute (F07); invalid accepted (F09) |

## E02 revision — `explore-it/lesson-01` ([E02 notes](../experiments/02-make-it-visible/notes.md)); ★ = new variation observed

| Family | Variable | Variations tried (E02) | Effect observed |
|---|---|---|---|
| ★ process fate | worker during lease | normal, crash (exit 3), frozen 14 s | crash → expiry chain to dead; frozen → 409 ack, re-run |
| ★ partner | hooksink latency vs 2 s timeout | 0 s, 3 s (slow) | 3 s → success counted as failure, duplicates (E02-F03) |
| ★ partner | URL class | valid, unsupported scheme `ftp://`, unparsable `http://bad host/x` | per-row retry vs global block (E02-F02) |
| ★ kind | job kind string vs worker kinds | `fail` + payload n, `fail3` | `fail3` never leased (E02-F01) |
| ★ queue config | lease_seconds/max_attempts at crash | 10 s/3 (emails), 20 s/4 (reports) | dead after 3 / 4 expiries, 32.5 s / 84 s |

## E03 revision — `explore-it/lesson-02` ([E03 notes](../experiments/03-vary-it/notes.md)); ★ = new variation observed, (r) = regression of an earlier fix

Variation table for the E03 charter (enqueue, batch, queue configuration). "Observed" means a runtime probe on stack
`queuelet-e03`; nothing in this table is a source inference unless marked *inferred*.

| Family | Variable | Heuristic | Values tried | Oracle | Observed | Verdict / evidence |
|---|---|---|---|---|---|---|
| ★ request shape | trailing data after first JSON value | Strings, Position(End) | ` trailing-garbage`, second object, ` {"jobs":[]}` | SPEC:17, RFC 8259 §2 | 201/200; extra ignored, second object dropped | bug [E03-F01](../defects/E03-F01-malformed-json-text-accepted.md) |
| ★ request shape | member-name case | Strings, Selection | `KIND`, `Priority`, `PAYLOAD`, `NAME`, `JOBS`/`Kind`, `MAX_ATTEMPTS`, `LEASE_TOKEN` | SPEC:17 | accepted as documented fields; last variant wins | bug [E03-F02](../defects/E03-F02-case-variant-member-names-accepted.md) |
| ★ request shape | encoding | Strings | raw 0xff in kind / in payload | RFC 8259 §8.1 | kind → U+FFFD 201; payload → 400 | kind part in E03-F01 |
| ★ request shape | duplicate exact key; Content-Type text/plain; empty / null / array body | Some/None/All | each | RFC 8259 (dup implementation-defined); SPEC:12 | last wins; accepted; 400 EOF / 400 kind required / 400 | not bugs (notes 04:16:44Z) |
| unknown field / enum (r) | `x`, backoff `zigzag`, `?state=bogus`, PATCH `backoff` | — | each | SPEC:17 | 400 each | holds (D26/U07 fixed) |
| priority (r) | −1, 0, 9, 10, 5.0, "5", null | Boundaries, Numbers | each | SPEC:46 | 400,201,201,400,400,400,201(default) | holds |
| ★ time | delay_seconds | Boundaries, Numbers | −60, 0, 1, 9223372036, 9223372037, 1e11, 2^63−1, 2^63 | SPEC:48-49 | ≤9223372036 honoured; above wraps (1734 queued / 2272 / now−1 s); 2^63 → 400 | bug [E03-F03](../defects/E03-F03-duration-overflow-delay-and-extension.md); negative → queued (spec silent) |
| ★ time | run_at | Boundaries | 0001-01-01Z, 9999-12-31Z, no zone, +14:00, 9 fractional digits | SPEC:48 | queued, scheduled, 400, scheduled (UTC 10:00 prev day), µs rounding | holds |
| ★ attempts | per-job max_attempts | Boundaries | 1, 2147483647, 2147483648 | SPEC | 201, 201, 400 (pq text) | holds (message leak only) |
| ★ kind | kind string | Strings | " ", "a\u0000b" | SPEC:46 | 201, 400 | spec silent on charset |
| ★ payload | number forms | Numbers | 1E2, -0, 1.50, 1e400, 20-digit decimal, 2^53+1, 1e200000 | SPEC:47 | 100, 0, 1.50, 401-digit integer, exact, exact (lease raw bytes), 400 | value preserved; not bugs; 1e400 amplification Q-E03-5 |
| ★ payload | object shape / strings | Strings | dup keys, key order, emoji, lone surrogate, U+0000 | SPEC:44 | collapsed (last), sorted, kept, 400, 400 | jsonb normalisation; not bugs (Q-E03-3) |
| size (r) | batch count | Count, Zero/One/Many | 0, missing, 500, 501 | SPEC:54-55 | 400, 400, 201, 400 | holds (D06) |
| ★ size | request body bytes | Goldilocks | batch 2 077 511 B, 2 177 011 B; single 2.2 MB payload | SPEC silent | 201; 400 "request body too large"; 400 | limit 2 MiB (`api.go:50`), Q-E03-4 |
| position (r) | invalid batch element | Beginning/Middle/End | first, middle, last | SPEC:56 | 400, nothing stored | holds (U05) |
| ★ selection | idempotency key in batch | Zero/One/Many | same key twice in one batch; then single; "" key | SPEC:52 | ids [x,x], one row; repeat returns x; "" is a real key | holds |
| range (r) | queue lower bounds | Boundaries | lease/max 0,−1; retention −1; base −1; PATCH 0/−5 | SPEC:25-26 | 400 each | holds (D05/U06); PATCH message always names lease_seconds |
| ★ range | queue int upper bounds | Boundaries | lease/max/base/max_seconds 2147483647, 2147483648 (create, PATCH) | SPEC:13-16 | 201; 409 create / 500 PATCH with pq text | bug [E03-F04](../defects/E03-F04-storage-limit-errors-reported-as-conflict.md) |
| ★ string length | queue name | Boundaries, Goldilocks | random 2600/2800/10000; "q"×100000/900000/1100000 | SPEC:22 + addressability | 201/409/409; 201/409/409 (index-row limit, compressibility-dependent) | E03-F04; H11 refuted |
| ★ range | retention_hours | Numbers, Boundaries | 0, 0.0001, 5e7, 6e7, 1e8, 1e308 | SPEC:25, :85 | all 201; ≥6e7 with a terminal job → Retain fails globally, health 503 | bug [E03-F05](../defects/E03-F05-large-retention-halts-global-retention.md) |
| ★ config × count | backoff base × attempt | Numbers | exp base 2147483647 attempt 4; linear attempt 5; default base 5 attempt 31/32 | SPEC:27-28 | wraps to 1986 / 1782 / 2196 ok / 1782 | bug [E03-F06](../defects/E03-F06-backoff-overflow-immediate-retry.md) |
| ★ undocumented field | backoff.max_seconds | Some/None/All | −7, 1 | SPEC:17, :24-27 | accepted, stored, ignored (+4 s, +10 s waits) | bug [E03-F07](../defects/E03-F07-backoff-max-seconds-ignored.md) |
| ★ presence | PATCH webhook members | Some/None/All | {secret}, {url}, {}, null, top-level only | RFC 7396 approx. | omitted member cleared; null keeps; top-level merges | bug [E03-F08](../defects/E03-F08-patch-webhook-partial-clears-other-member.md) |
| ★ time (sibling) | nack delay_seconds, heartbeat extend_seconds | Boundaries | 9223372037 | SPEC worker | wraps to 1734: immediate retry / immediate expiry (+1 attempt) | E03-F03 siblings |
| webhook URL (r) | ftp, no host, space, "" | — | each | SPEC destinations | 400,400,400,201 | holds (U10) |

Untried rows (backlog): schedule create/PATCH body shape and payload (#offcharter), lease `max_jobs` / list `limit`
coercion (#offcharter), worker `kinds` array shapes, concurrent idempotent enqueue race, `lease_seconds` 2147483647 expiry
behaviour (accepted; expiry not observed), very long `kind` / `idempotency_key`.

## E04 additions — `explore-it/lesson-03` ([E04 notes](../experiments/04-judge-it/notes.md))

| Family | Variable | Heuristic | Values | Oracle | Observed | Judgment |
|---|---|---|---|---|---|---|
| ★ sign/range | nack `delay_seconds` | Numbers, Boundaries | −5, −9223372036, −9223372037, −18446744073 | SPEC backoff "never before the failure"; monotone | past / 1737 / 2322 / +0.71 s | bug E04-F03 |
| ★ sign/range | enqueue `delay_seconds` | Numbers | −5, −9223372037 | monotone | queued past / scheduled 2322 | E04-F03 |
| ★ timing | request vs sweep after expiry | Timing | immediate vs +1.5 s | SPEC current holder | 200 vs 409 | bug E04-F04 |
| ★ method/path | HTTP method on /v1 routes | — | PUT, DELETE, GET, OPTIONS, PATCH, unknown path | SPEC error envelope | text/plain 404/405 | bug E04-F01 |
| ★ count | attempt number across retry | Count | 2 failures, retry, 1 success | numeric = execution order | 1,1,2 | bug E04-F02 |
| ★ count | default backoff attempt | Count | 1…21 | spec bound 2^31−1 s | 5·2^(n−1) exact | holds; A1 spec silence |
| ★ concurrency | parallel leases | Many | 30 calls, 2 workers, 40 jobs | atomic lease | 40 distinct | holds |
| ★ time source | app clock step back between enqueues | Timing | −60 s | FIFO by creation | inverted | lab-only; risk H-E04-08 |

## E09 additions — `explore-it/lesson-08` ([E09 notes](../experiments/09-explain-it/notes.md))

| Family | Variable | Kind | Variations | Oracle | Observed | Outcome |
|---|---|---|---|---|---|---|
| ★ content | cron day/month combination | Selection | Feb 30, Apr 31 (fixed and wildcard hour) | no impossible next fire | next_run_at 0001-01-01 | E09-F02 |
| ★ content | cron prefix token | Selection | `CRON_TZ=Asia/Tokyo`, `TZ=Asia/Tokyo` | 5 fields | accepted, zone overrides | E09-F03 |
| ★ content | timezone alias | Selection | `Local` | IANA only | accepted | E09-F09 |
| ★ position | schedule id order relative to a failing schedule | Position | lower, higher | independent | higher ids blocked | E09-F01 |
| ★ concurrency | DELETE of a just-terminal job | Timing | ‖ cancel (300), ‖ ack (300), ‖ expiry pass (40), ‖ worker delete (40) | outbox + true status | 2/300, 125/300, 2/40, 1/40 hits | E09-F07, F08 |
| ★ time source | app clock step back after promotion | Timing | −11 min 35 s | due-only lease | future job leased | risk H-E04-08 |
