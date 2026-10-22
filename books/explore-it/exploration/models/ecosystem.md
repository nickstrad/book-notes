# Ecosystem

| Neighbour | Trust boundary | Data path | What-if tried (E01) | Result |
|---|---|---|---|---|
| PostgreSQL | API ↔ DB (internal network) | all state; explorer RO role | restart (H26); stop (H42) | request path reconnects; loop dead (F08); valid keys → 401 (F14); /health ok throughout |
| Webhook partner (hooksink) | outbound HTTP, HMAC-signed | outbox → POST | mode 500 (H12); timeout (H34); capture listener (H11b) | 500 marked delivered (F11); timeout stalls loop (F13); signature over wrong bytes (F10) |
| Workers (containers) | bearer key, lease token | lease/heartbeat/ack/nack | API restart; DB outage | O: workers log 401 (01:49:47) and connection refused (01:50:24); I: they keep polling (internal/worker/worker.go:42-46) |
| Producers / operators (qctl, curl) | bearer key | HTTP JSON | invalid input sweep | lax validation (F19, F20); qctl exit 0 (F22) |
| Other namespace (tenant) | key → namespace | job ids are global sequential ints | globex on acme jobs | read/cancel/delete/kill (F03) |
| Clock | host UTC; Go time.Now in API | Loop ticks, cron | no control; DST via harness | F27; TR2 |

Testability requests: TR1 (hooksink records received signature), TR2 (lab clock control), TR3 (loop liveness metric) — see E01 notes.

## E02 revision — `explore-it/lesson-01` ([E02 notes](../experiments/02-make-it-visible/notes.md)); ★ = new observed cell

| Neighbour | Trust boundary | Data path | What-if tried (E02) | Result |
|---|---|---|---|---|
| ★ Docker restart policy (`restart: on-failure`) | process supervisor | worker exit → new container incarnation → new `POST /v1/workers` | crash kind on emails and reports | RestartCount +3 / +4; each incarnation a new worker row; exit code of the crashing run not retained (H03, H13) |
| ★ Frozen worker (`docker pause`) | process scheduling | lease held without heartbeat | pause 14 s on 10 s lease | expiry, stale ack 409 visible only in worker stdout, work executed twice (H14) |
| ★ Webhook partner, 500 | outbound HTTP | outbox → POST | mode 500 for 58 s | retried every 5–6 s, delivered on first 200; API log silent for failures (H05–H07) |
| ★ Webhook partner, slow 3 s | outbound HTTP, 2 s client timeout | outbox → POST | mode slow 24 s | partner records 5 × 200, queuelet 5 failures (E02-F03) |
| ★ Webhook URL from tenant configuration | tenant input → shared shipper | queue.webhook_url copied into outbox | `http://bad host/x` (globex) | all tenants' notifications stop; health degraded (E02-F02) |
| ★ Webhook URL, unsupported scheme | same | same | `ftp://x/y` | isolated per-row failure, retried forever (H11c) |
| ★ PostgreSQL restart | API ↔ DB | all state | `dc restart postgres` | 503 auth_unavailable and health degraded ≈4–6 s, loop recovers, delayed job promoted (H15; F08/F14 regressions hold) |

Testability requests from this run: TR-E02-1…6 in the [E02 notes](../experiments/02-make-it-visible/notes.md) (correlated
lifecycle events; per-POST outbox history; backlog/liveness metrics; eligibility reason; lease start; controllable clock).
