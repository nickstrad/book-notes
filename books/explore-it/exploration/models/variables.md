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
