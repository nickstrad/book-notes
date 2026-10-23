# Entities

| Entity | Key attributes (schema) | Relationships | CRUD seen | Dependent outcomes (E01) |
|---|---|---|---|---|
| namespace | id, name | owns keys, queues, jobs, workers, schedules | R (stats) | admin key = namespace 1 + admin flag (I, migrations) |
| api_key | key_hash, namespace_id, admin | → namespace | R (auth) | DB outage misread as bad key (F14) |
| queue | name, lease_seconds, max_attempts, backoff_kind/base/max, retention_hours, webhook_url/secret, paused | 1→N jobs, schedules (no FK) | CRUD O | delete orphans jobs+schedules (F15); secret returned in GET (Q5) |
| job | kind, payload jsonb, priority, state, run_at, attempts, max_attempts, idempotency_key, lease_*, result, last_error, created_at(0), finished_at | → queue, namespace; 1→N attempts, outbox | CRUD O | precision (F17); stale result after retry (F06) |
| attempt | job_id, n, worker_id, started_at, finished_at, outcome, error | → job (no FK) | C via ack/nack, R | none on expiry (F05); duplicate n after retry (Q2); started_at=finished_at (Q1) |
| worker | name, kinds[], token, last_seen_at | → namespace; leases | C, D O | token returned but never used by API (I); last_seen_at never updated (I) |
| schedule | name, queue_id, kind, payload, cron, timezone, zone_offset, enabled, next/last_run_at | → queue | CRUD + run O | zone_offset frozen (F27); PATCH no recompute (F07) |
| webhook_outbox | job_id, url, secret, body, attempts, next_at, delivered_at | → job | C (Notify), U (Ship) | delivered on 500 (F11) |

## E02 revision — `explore-it/lesson-01` ([E02 notes](../experiments/02-make-it-visible/notes.md)); ★ = new observation

| Entity | Observation | Evidence |
|---|---|---|
| ★ worker | One row per process incarnation; old rows never removed; `last_seen_at` = registration time (heartbeats do not update it) | H03, H13, final snapshot (8 rows, 5 worker-b) |
| ★ attempt | `expired` rows carry the incarnation's worker id; `started_at` = `finished_at` for every outcome (expired, failed, succeeded) | H04, H09, H14 |
| ★ webhook_outbox | Stores only attempts count, next_at, delivered_at — no per-POST status; URL copied at Notify, so queue PATCH does not affect pending rows; survives job deletion (orphan) | H05, H11a, H11b |
| ★ job | `last_error` survives a later success; dead job's `run_at` = last nack + backoff | S9 (H14), S4 (H09) |

## E06 revision — `explore-it/lesson-05` ([E06 notes](../experiments/06-relate-it/notes.md)); ★ = new observation, (r) = regression of an earlier fix, (I) = inferred from code/schema only

### Constraints: database vs application (from migrations 001–004, read 15:28–15:31Z)

| Dependency | Database constraint | Application enforcement | Observed outcome when the owner goes |
|---|---|---|---|
| queue → job | ★ FK `jobs_queue_owner` NOT VALID, NO ACTION (002.sql:11-13) | deleteQueue LOCK + 409 if any job, any state (queues.go:109-151) | ★ 409 (queued and terminal-only); 204 after jobs deleted; 0 orphans even under a race (P1, P9b) |
| queue → schedule | ★ FK `schedules_queue_owner` NOT VALID (002.sql:15-17) | same check | ★ 409; 204 after schedule deleted (P1); race: 0 orphans, but 500 FK text (E06-F06) |
| job → attempt | none (001.sql:7) | none | ★ rows orphaned on purge, DELETE, retention; unreachable via API (E06-F07) |
| job → webhook_outbox | none | Notify snapshots url/secret/body | ★ pending row delivered after job and queue deleted (P8e); delivered rows never removed |
| job → job_events | none | events kept by design | ★ `purged`/`deleted`/`retained` readable via API and API log after removal (P1, P10) |
| worker → job.leased_by | none | ReleaseWorker in delete tx | ★ (r) released, attempt counted `released` (P7) |
| worker → attempt.worker_id | none | kept by SPEC | ★ attempt keeps deleted worker id 2 (P7) |
| worker → sample worker process | — | none (client) | ★ process polls with deleted id forever, 404 per second, never exits/re-registers (P7, Q-E06-1) |
| schedule → job (fire) | none (jobs have no schedule id) | fire copies kind/payload; **no payload validation** | ★ non-object payload jobs created (E06-F02); deleting a schedule leaves its jobs (I) |
| namespace → everything | none | no delete API (I) | not probeable |

### CRUD grid (entity × operation), observed in E06 unless marked

| Entity | Create | Read | Update | Delete |
|---|---|---|---|---|
| queue | ★ 201; race loser N/A | ★ GET after other tenant's DELETE 200 | ★ PATCH max_attempts: old jobs keep, new + run jobs get new; webhook PATCH: pending rows keep snapshot; url "" stops new rows; retention applies to existing terminal jobs | ★ empty 204; missing 404; other tenant 404; with job/schedule 409 |
| job | ★ enqueue exact digits; qctl rounds (E06-F04); NUL 400 driver text; race 400 FK text (E06-F06) | ★ GET/attempts 404 after delete/purge/retention; events still readable | ★ ack/nack with NUL → 409 (E06-F05) | ★ terminal 204; then queue deletable; key freed |
| schedule | ★ any payload type 201 (E06-F02); race 500 FK text | ★ GET 404 for missing | ★ PATCH enabled false (used) | ★ 204 for missing/foreign id (E06-F01); 204 real delete |
| worker | ★ 201 | — | — | ★ (r) release counted; process zombie (P7) |
| attempt | via ack/nack/expiry/release | ★ survives job removal in psql only | — | ★ never deleted (E06-F07) |
| webhook_outbox | ★ on terminal transition with URL; none with url "" | ★ hooksink raw_body = outbox body bytes (compact, HTML-escaped) | ★ attempts/status per POST | ★ never deleted (sibling of E06-F07) |
| job_events | trigger + purge/delete/retain inserts | ★ API and console agree | — | lab reset only (README) |

### Data journey (P4–P6): value of one datum per hop

| Hop | Exact? | Transformation (oracle) |
|---|---|---|
| curl enqueue → jobs.payload | ★ yes | jsonb: `1e2`→`100`, key order, dup key last wins (RFC 8259 value-equal) |
| qctl enqueue → jobs.payload | ★ **no** (E06-F04) | float64 rounding |
| schedule payload → fired job | ★ bytes copied; non-objects pass (E06-F02) | none |
| jobs.payload → lease response | ★ yes | HTML escape `<` (value-equal) |
| lease → sample worker → result | ★ **no** (E06-F03) | float64 rounding; non-object payload undecodable |
| manual ack result → jobs.result | ★ yes (UseNumber); NUL → 409 (E06-F05) | jsonb normalisation |
| jobs.result → outbox body → partner bytes | ★ yes; signature valid over sent bytes | compact + HTML escape |
| any job field → qctl display | ★ **no** (E06-F04) | float64 rounding |
