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
