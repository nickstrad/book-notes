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

## E03 additions — `explore-it/lesson-02` (sibling predictions timestamped in [E03 notes](../experiments/03-vary-it/notes.md))

| Cause | Mechanism (source) | Findings | Sibling predictions (time, status) |
|---|---|---|---|
| C15 Lenient JSON decoder framing | single `Decoder.Decode`, no EOF check; Go v1 case-insensitive field match; U+FFFD substitution (api.go:49-58) | E03-F01, E03-F02 | PATCH trailing/case 04:24Z → confirmed (H19); ack case-variant 04:24Z → confirmed (H20); schedules/clock/worker create: predicted, not probed |
| C16 Unchecked time.Duration arithmetic on caller/config integers | `Duration(n)*Second` (enqueue.go:54, nack.go:24, jobs.go:84); `Backoff` shifts/multiplies (jobs.go:17-22) | E03-F03, E03-F06 | nack delay 04:24Z → confirmed (H17); heartbeat 04:24Z → confirmed (H18); lease_seconds bounded by int column: inferred safe |
| C17 Storage limits stand in for validation | upper bounds come from `int` columns / btree index size / timestamptz range (migrations/001.sql:3); errors mapped by call site (queues.go:47-49 → 409, :93-94 → 500) | E03-F04, E03-F05 | enqueue maps DB faults to 400 (enqueue.go:81-84): predicted, not probed |
| C18 Accepted input with no consumer | `backoff.max_seconds` stored, never read (queues.go:20,46) | E03-F07 | other unread fields: none found |
| C19 Nested PATCH object replaced wholesale | `Webhook *struct{URL,Secret string}` then both assigned (queues.go:72-75, 85-92) | E03-F08 | schedule PATCH nested objects: none found (pointers only), not probed |

## E04 additions — `explore-it/lesson-03` (sibling predictions timestamped in [E04 notes](../experiments/04-judge-it/notes.md))

| Cause | Mechanism (source) | Findings | Sibling predictions (time, status) |
|---|---|---|---|
| C20 Routing fallbacks bypass the error writer | method-qualified `ServeMux` patterns; mux 404/405 written by net/http, `fail()` only in handlers (api.go:119-172) | E04-F01 | OPTIONS/PATCH on known routes 12:49:30Z → confirmed; `{verb}` wildcard gives JSON 404 instead of 405: observed |
| C21 One-sided bounds on caller seconds | `> MaxSeconds` checks only (internal/api/jobs.go:67, enqueue.go:55); same `Duration(n)*Second` as C16 | E04-F03 | heartbeat lower bound guarded (jobs.go:77): inferred safe; clock advance both signs: inferred safe |
| C22 Lease validity decided by the sweep, not by the request | ack/nack/heartbeat guard state+token only (ack.go:10, nack.go:14, jobs.go:97); expiry evaluated once per tick (expire.go:12) | E04-F04 | nack before sweep 12:49:30Z → confirmed |
| C23 Release paths do not record readiness | expiry and worker delete set `queued` without touching `run_at` (expire.go:41, internal/api/jobs.go:126); stats derives readiness from `run_at` (stats.go:18) | E04-F05 | worker delete: predicted after P6 but logged only after probe P8 → confirmed (not counted as stated) |
| C24 Attempt counter reused as history key | Retry resets `attempts` (operator.go:13); history rows numbered from it, no incarnation key (001.sql:7) | E04-F02 | job_events order by id: inferred unaffected |

## E05 additions — `explore-it/lesson-04` (sibling predictions timestamped in [E05 notes](../experiments/05-sequence-it/notes.md))

| Cause | Mechanism (source) | Findings | Sibling predictions (time, status) |
|---|---|---|---|
| C25 Exits from `leased` written independently, each choosing which columns to reset | five UPDATEs leave `leased`: nack.go:60, expire.go:116, api/jobs.go:126 (clear lease, differ on attempts), ack.go:25, operator.go:6 (keep lease) — no shared "end lease" routine | E05-F01 (release skips attempt accounting), E05-F02 (ack/cancel keep lease fields) | release paths 13:24:30Z → confirmed (H-E05-13); nack retry=false clears 13:29:40Z → confirmed (H-E05-17). Related to C23 (release paths and readiness), fixed in lesson 04 for run_at only |
| C26 Observation keyed on sticky error text | 002.sql `observe_job()` labels `lease_expired` by `NEW.last_error='lease expired'`; `last_error` survives later transitions | E05-F03 | ack/cancel after expiry 13:34Z → confirmed (H-E05-18); nack with that literal text: predicted, not probed |
