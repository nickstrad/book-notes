# Causes (E01, v0-inherited)

Grouping is by mechanism read in source after the behaviour was observed. Sibling predictions with timestamps are in
[E01 notes](../experiments/01-map-it/notes.md); "predicted" means not yet probed.

| Cause | Mechanism (source) | Findings | Sibling predictions (time, status) |
|---|---|---|---|
| C1 Job mutations keyed by id only | UPDATE … WHERE id=$1 with missing token/state/namespace predicates (ack.go:10, nack.go:10-22, operator.go:6,13, internal/api/jobs.go:11-23) | F01, F02, F03, F06 | H27 nack-on-succeeded 01:42:13Z → confirmed; H28 cross-ns nack 01:42:34Z → confirmed; retry cross-ns predicted |
| C2 Attempt accounting split across paths | dead test `attempts+1 > max` (nack.go:15); Expire does not count (expire.go:6); worker release does not count (internal/api/jobs.go:110) | F04, F05 | worker-delete release uncounted: predicted |
| C3 Background loop is a single fragile sequence | `return` on any step error (api.go:102-104); inline webhook I/O (ship.go:32); Fire catch-up loop (fire.go:41) | F08, F13, F24 | NewRequest error stops loop: predicted; catch-up after outage: predicted |
| C4 Schedule state not derived from config | next_run_at/zone_offset computed only at create (schedules.go:33-56); PATCH writes columns (schedules.go:84); Fire ignores enabled (fire.go:28) | F07, F09, F27 | — |
| C5 Lax input decoding/validation | json.Decoder without DisallowUnknownFields (api.go:43); per-field checks only on create; map[string]any payloads | F16, F17, F19, F20 | schedule payload precision: predicted |
| C6 Error mapping collapses causes | any DB error → 401 (api.go:86-88), → 404 (api.go:117-123, internal/api/jobs.go:12-14); fail(w,200) (internal/api/jobs.go:132); qctl never exits non-zero | F14, F15 (204 for missing), F21, F22 | outage → 404 on queue lookups: predicted (masked by F14) |
| C7 Webhook payload construction | signs MarshalIndent, sends Marshal (ship.go:22-25); status unchecked (ship.go:40-43); stringify+truncate (jobs.go:79-86) | F10, F11, F12 | last_error truncation: predicted |
| C8 Ordering/reporting semantics | created_at timestamp(0) + no id tiebreak (jobs.go:24); stats fold & run_at age (stats.go:17-19) | F18, F23, F26 | catch-up jobs share created_at: observed in F24 data (01:48:05 ×4) |
| C9 Spec silence | "pending" undefined | F25 | — |
| C10 No referential integrity | no FKs in migrations/001.sql; unconditional DELETE (queues.go:86) | F15 | orphan outbox rows after job delete: predicted |

## E02 additions — `explore-it/lesson-01` (sibling predictions timestamped in [E02 notes](../experiments/02-make-it-visible/notes.md))

| Cause | Mechanism (source) | Findings | Sibling predictions (time, status) |
|---|---|---|---|
| C11 Shared shipper aborts its batch on a per-row configuration error | `return e` on `http.NewRequest` failure before updating the row, `ORDER BY id LIMIT 20` (ship.go:17-28); no URL validation (queues.go:21-41, :79-84) | E02-F02 | PATCH path accepts bad URL 03:30Z → confirmed (H11d); ftp URL does not block 03:30Z → confirmed (H11c); UPDATE-error aborts (ship.go:37-40, :46-49): predicted, not probed |
| C12 Delivery policy fixed in code, not in spec | 2 s client timeout (ship.go:31); fixed +5 s, no cap (ship.go:37) | E02-F03 (spec) | 500 and ftp rows retried forever: observed (H07, H11c) |
| C13 Consoles lack correlation ids | log lines without job/outbox/worker ids (ship.go:45; worker.go:48,71); no request log | testability TR-E02-1…5 (not defects) | — |
| C14 Sample worker advertises kinds by family name | default kinds list (cmd/worker/main.go:24) vs prefix handling (worker.go:86) vs exact match (jobs.go:44) | E02-F01 | none found |
