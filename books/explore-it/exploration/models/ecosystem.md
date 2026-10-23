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

## E06 revision — `explore-it/lesson-05` ([E06 notes](../experiments/06-relate-it/notes.md)); ★ = new observed cell

| Neighbour | Trust boundary | Data path | What-if tried (E06) | Result |
|---|---|---|---|---|
| ★ Sample worker (`internal/worker`) | client decode of lease JSON | lease bytes → `map[string]any` → result | payload > 2^53, `1.10`; array payload from a schedule; its own row deleted | rounds numbers (E06-F03); array undecodable → job dies by expiry, log names no job (E06-F02); deleted row → zombie polling, recovered only by container restart (P7) |
| ★ qctl | operator input → HTTP | `--payload` → `any` → JSON | big number | stores a different number; display rounds (E06-F04) |
| ★ PostgreSQL as validator | API → jsonb/text columns | result/error/payload | `\u0000` | jsonb/text reject; mapped to 409 on worker verbs (E06-F05), 400 on enqueue |
| ★ PostgreSQL FKs (002) | owner deletion vs concurrent writes | queue lookup → insert | DELETE racing enqueue/schedule create | no orphans; losers get 400/500 with constraint names (E06-F06) |
| ★ Webhook partner | snapshot at terminal time | outbox url/secret/body | PATCH url+secret while row pending; delete job and queue while pending | pending row keeps old url/secret (valid signature); delivered after owner deletion (P8) |

## E08 revision — `explore-it/lesson-07` ([E08 notes](../experiments/08-surround-it/notes.md)); ★ = new observed cell

Context diagram (observed boundaries marked ★; drawn before probes, confirmed by them):

```text
              producers/operators (curl, qctl)            other tenant (globex-key)
                         │ ★ bearer key → namespace             │ ★ key → ns 2; all job/schedule/worker ids 404
                         ▼                                       ▼
 admin-key ★──────▶  queuelet API (:8200) ◀──── ★ worker containers (acme-key; worker_id only; lease_token)
 (= acme + stats)    │ auth mw │ handlers │ test clock ★ (semantic time only)
                     │ core loop (promote/expire/fire/retain) ║ shipper loop (separate goroutine)
                     ▼ ★ SQL: no deadlines                      ║ ★ outbound HTTP, HMAC over sent bytes,
                  PostgreSQL 17                                  ▼   2 s deadline, ONE global serial queue
                                                        tenant-chosen URLs (hooksink / anything reachable)
 docker compose ★ supervisor: workers on-failure; api, postgres, hooksink: none
 IANA tz db ★ via Go LoadLocation; cron = robfig/cron v3 local civil time
```

Boundary table (what crosses → where checked → what-if → observed → recovery):

| Boundary | What crosses | Checked at | What-if (E08) | Observed | Recovery |
|---|---|---|---|---|---|
| ★ Tenant (key → namespace) | job/schedule/worker ids, queue names | api.go:149-170 auth; jobs.go:200 ownJob; queues/schedules `namespace_id` filters; jobs.go:174 worker | globex key on acme ids, all routes | 404 everywhere; events `[]` 200 (scoped by event ownership; also for nonexistent ids); stats 403 | n/a (D21 holds) |
| ★ Admin key | stats for any namespace, clock | 001.sql:16 (admin = ns 1); observe.go:16 | admin on acme/globex jobs | admin is a full acme key; 404 on globex jobs; 200 globex stats; acme-key → /admin/clock 403 | n/a |
| ★ Worker identity within a tenant | `worker_id` (token issued, never checked) | jobs.go:173-178 (namespace only) | acme client leases as worker-a | 200; attempt history blames worker-a (Q-E08-1) | n/a |
| PostgreSQL restart (running API) | all state | per-request auth error → 503; loop logs and continues | `compose restart postgres` | ≈1 s of 503 `auth_unavailable`; promotion on time (H-E08-04) | automatic |
| ★ PostgreSQL absent at API start | DSN ping, migrations | main.go:19-35 `log.Fatal` | stop DB → restart api → start DB | api exit 1, stays down (E08-F01) | manual `compose start api` |
| ★ PostgreSQL crash / API crash | process | compose: no restart policy | SIGKILL each | stays down until manual start; running API 503 meanwhile (E08-F01 siblings) | manual |
| ★ PostgreSQL frozen (connected, silent) | every query | no context/deadline anywhere; /health 1 s ping ineffective | `compose pause` 12 s / 25 s | every request incl. /health hangs for the freeze, then 200 (E08-F04) | automatic at unfreeze; nothing signals it |
| Webhook partner 500 | outbox body + signature | ship.go:40-48 | mode 500 | retried every 6 s; delivered on first 200 (D04 holds) | automatic |
| Webhook partner slow 3 s | same | 2 s client deadline | mode slow | failure each time, partner records 200 each time (3 side effects for 1 notification; documented) | automatic |
| ★ Webhook partner timeout / unreachable URL × other tenants | the shared outbox order | ship.go:17 `ORDER BY id LIMIT 20`; next_at from pre-pass tick | 22 timed-out globex rows; 20 rows to an unreachable globex URL | later rows of every tenant never attempted (E08-F02); /health 200; tenant PATCH/delete cannot end it | only when the partner recovers; unreachable URL: never (no cap) |
| ★ Data journey ack → outbox → partner | result JSON | store.go:44 RawMessage; ship.go:22 | traced datum with `<&>`, U+2028, 20-digit int, `1.10` | jsonb key reorder, compact, Go HTML/U+2028 escapes; semantically equal; HMAC by hand over received bytes matches (D24 holds) | n/a |
| ★ Clock (test clock) × IANA DST | semantic now → cron Next | fire.go:13-23 robfig v3 (zone_offset ignored) | 2027-03-14 gap, 2027-11-07 overlap, NY | gap fixed times skipped that day; overlap fixed times fire twice; hourly and 09:00 correct (E08-F03, spec) | n/a |

Uncertainty: lib/pq cancellation behaviour behind E08-F04 is inferred, not read. Untested cells: redirecting partner
(H-E08-22), pool exhaustion during freeze (H-E08-24), DB fault between ack statements (T-E08-1), workers' own
behaviour against a frozen API (their HTTP timeout not read this run).

## E09 revision — `explore-it/lesson-08` ([E09 notes](../experiments/09-explain-it/notes.md)); ★ = new observed cell

| Neighbour / path | What if | Observed |
|---|---|---|
| ★ robfig/cron parser (library) | caller passes a `CRON_TZ=`/`TZ=` prefix | library zone wins over the API's `timezone` (E09-F03) |
| ★ robfig/cron `Next` | expression never matches | zero time, no error; API stores year 1 (E09-F02) |
| ★ Go `time.LoadLocation` | `Local` | server process zone accepted (E09-F09) |
| ★ operator / cleanup client | deletes terminal jobs as they appear | notifications lost, success reported as 409/500 (E09-F07, F08) |
