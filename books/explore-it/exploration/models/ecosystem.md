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
