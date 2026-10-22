# E06-F05 — A result or error containing U+0000 is answered 409 "state_conflict" with driver text; the holder's lease is untouched

Seen at: export of `explore-it/lesson-05` (stack `queuelet-e06`), experiment 06 (relate-it).
Original observation: [E06 notes](../experiments/06-relate-it/notes.md) P6, 15:36:27Z (H-E06-07, predicted from code and
PostgreSQL behaviour). Related prior judgement: E03-H12 / Q-E03-3 (enqueue of `\u0000` → 400 judged a storage limit,
not a bug); this report is about the worker verbs, which E03 did not probe.

## Minimal reproduction

Actor: acme key, my worker 4 (`kinds:["manual"]`). Clock: real. Queue: reports.

```sh
source exploration/experiments/06-relate-it/tools/env.sh
k acme-key POST /v1/queues/reports/jobs '{"kind":"manual","payload":{"case":"ack-nul"}}'      # 314
# lease 314 with worker 4 → token T
k acme-key POST /v1/jobs/314/ack '{"lease_token":"T","result":{"s":"a\u0000b"}}'
#   {"error":{"code":"state_conflict","message":"pq: unsupported Unicode escape sequence"}} [409]
sql "select state,lease_token is not null,attempts from jobs where id=314"   # leased | t | 0
k acme-key POST /v1/jobs/315/nack '{"lease_token":"T2","error":"bad\u0000byte"}'
#   {"error":{"code":"state_conflict","message":"pq: invalid byte sequence for encoding \"UTF8\": 0x00"}} [409]
```
Replay: `tools/p6-nul.sh`.

## Expected (named oracle)

SPEC:16 "A command incompatible with the current state returns 409"; invalid commands are 400 (SPEC:19-22, and the lesson-03 resolution of E03-F04: a storage limit is not a conflict). Here the job *is* leased by the
caller with a valid, unexpired token, so 409 tells the worker "you lost the lease" when the real cause is a value the
store cannot hold. `\u0000` is valid JSON (RFC 8259 §7) and the body passes `decode`. Expected: 400 with a service
message (e.g. `invalid_result` / `invalid_error`), or acceptance with a documented transformation.

## Observed

- ack with `\u0000` in a result string → 409 `state_conflict`, message is the driver's text; job stays `leased`,
  token kept, attempts 0. A retry with a clean result → 200.
- nack with `\u0000` in error text → 409 `state_conflict` with driver text; job stays `leased`. Clean nack → 200, `dead`.
- enqueue payload/kind with `\u0000` → 400 `invalid_job` with driver text (class right, message leaks driver; E03 already
  recorded this).

## Variables isolated

- Matters: U+0000 in a JSON string reaching a `jsonb` (result) or `text` (last_error, attempts.error) column.
- Does not matter: retry flag, queue. A `\u0001` or other control character was not probed.
- Mapping: `ack.go:10-19` and `nack.go` return any database error; `internal/api/jobs.go:88` maps every non-nil error
  to 409.

## Sibling search

- `internal/api/jobs.go:88`: cancel/retry/heartbeat share the same "every error → 409" mapping; a DB outage during
  ack would also read as 409 (H-E04-10, still open).
- Field-level: `heartbeat` takes no strings; `createWorker` name/kinds with NUL → 500 `database_error` (predicted from
  `api/jobs.go:107`, not probed).
- Consequence for the sample worker: an `echo`/`http` result containing U+0000 is acked → 409 → logged as `completed`,
  the job then expires and is retried until `dead` "lease expired" (predicted, not probed).

## Severity and status

Severity: **low–medium**. Wrong error class on the worker protocol hides a data problem as a lease problem; a worker
following the protocol abandons work it still holds, and the job burns attempts by expiry. Status: **open**.
