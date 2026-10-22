# Queuelet service contract

Queuelet stores jobs durably in PostgreSQL and distributes them to workers.
A namespace contains queues, workers and schedules. A bearer API key selects a
namespace; an administrative key may inspect namespace statistics.

## Authentication and responses

Every `/v1/` request requires `Authorization: Bearer <key>`.
A namespace key accesses only its namespace's resources.
The health and metrics endpoints are public.
JSON commands use `Content-Type: application/json`.
Errors have the form `{"error":{"code":"…","message":"…"}}` and a 4xx or 5xx
HTTP status. Successful reads and mutations return 200 unless stated otherwise.
Creation returns 201. Deletion returns 204. A missing resource returns 404.
A command incompatible with the current state returns 409.
A path no route serves is 404 and a method a route does not accept is 405, both
in the same error form.
Unknown fields and invalid enum values should be rejected as invalid commands.
A command body is exactly one JSON object in UTF-8. Trailing data, invalid UTF-8
and member names that match a documented field only when case is ignored are
rejected with 400 `invalid_json`.
Every seconds-valued field and `max_attempts` accepts at most 2147483647.
A `delay_seconds` (enqueue or nack) is never negative; a negative value is 400.

## Queues

`POST /v1/queues` creates a queue with a unique name within the namespace.
Names use 1–128 ASCII letters, digits, underscores and hyphens. Only a taken name
is a 409 conflict; an invalid or out-of-range setting is 400.
Configuration defaults are `lease_seconds:30`, `max_attempts:3`,
`backoff:{kind:"exponential",base_seconds:5}`, `retention_hours:24`.
Lease duration and maximum attempts must be positive; retention hours range from
0 through 1000000.
Backoff kind is `linear` or `exponential`; the base is a nonnegative number
of seconds. Failed jobs retry with exponential backoff from `base_seconds`
or linear backoff when configured. An optional `backoff.max_seconds`, at least
`base_seconds`, caps each wait. Without it each wait is capped at 3600 seconds,
or at `base_seconds` when that is larger. A computed wait never overflows: it is capped
at 2147483647 seconds, so a retry is never scheduled before the failure.
An optional `webhook:{url,secret}` receives terminal notifications.

`GET /v1/queues` lists the namespace's queues.
`GET /v1/queues/{name}` reads configuration.
`PATCH /v1/queues/{name}` updates lease duration, maximum attempts, retention
or webhook settings. Existing jobs retain their maximum-attempt setting.
Omitted members keep their value, inside `webhook` as well: a secret-only PATCH
keeps the URL and a URL-only PATCH keeps the secret.
`DELETE /v1/queues/{name}` removes a queue; owned jobs and schedules must not
become orphaned. Callers must handle a conflict for a nonempty queue.
`POST /v1/queues/{name}/pause` prevents new leases and scheduled deliveries.
`POST /v1/queues/{name}/resume` allows work again. Missed periodic ticks should
not create an unbounded burst of catch-up jobs.
`POST /v1/queues/{name}/purge` removes all pending jobs: scheduled, queued and retry_wait.
A job leased when its queue is purged keeps its lease: an ack completes it, and a
nack, lease expiry or worker deletion makes it `dead`, never pending again.

## Jobs

`POST /v1/queues/{name}/jobs` accepts `kind`, a JSON object `payload`, and
optional `priority`, `delay_seconds`, `run_at`, `max_attempts`, `idempotency_key`.
Kind is required. Priority ranges from 0 through 9 and defaults to 5.
Payload values are delivered to workers without loss of numeric precision.
Choose either delay in seconds or an RFC3339 `run_at`, never both.
A job due now is `queued`; a future job is `scheduled`.
The service promotes due jobs to `queued`.
A queue's maximum-attempt setting is copied to each new job unless overridden.
Repeated use of an idempotency key within a queue returns the original job.

`POST /v1/queues/{name}/jobs:batch` accepts `{"jobs":[…]}` with 1–500 jobs.
An empty or oversized batch is invalid and returns 400 without insertion.
A batch is atomic: an invalid element leaves none of its new jobs stored.
It returns `{"ids":[…]}` after accepted jobs are stored.
`GET /v1/queues/{name}/jobs` supports state, kind, limit and cursor filters.
The limit defaults to 100 and is at most 500. The returned next cursor can
be used to continue through increasing job ids.
Valid state filters are `scheduled`, `queued`, `leased`, `retry_wait`,
`succeeded`, `dead` and `cancelled`.
`GET /v1/jobs/{id}` returns the current job, payload, result and last error.
`GET /v1/jobs/{id}/attempts` returns completed attempts in numeric order.
Attempt `n` numbers every execution of the job and is never reused, across
operator retries too; the lease's `attempt` counts toward `max_attempts`.
`DELETE /v1/jobs/{id}` removes a terminal job.
Attempt records belong to their job and are removed with it on deletion, purge and
automatic retention; durable job events and pending webhook snapshots keep their independent lifetime.
Only a `leased` job has a lease token, holder and expiry; every exit from
`leased` clears all three; attempt rows keep the worker that ran them.

## Worker protocol

`POST /v1/workers` takes a name and an array of supported kinds.
It returns a worker id and token.
`DELETE /v1/workers/{id}` deregisters the worker and releases its active leases.
Each release records an unsuccessful attempt, as an expired lease does; a job
whose maximum attempts are exhausted becomes `dead` instead of `queued`.
`POST /v1/queues/{name}/lease` takes `worker_id` and `max_jobs` (1–100).
Only jobs with a kind supported by that worker are eligible.
Leasing is atomic; concurrent workers cannot acquire the same lease.
Higher priorities run first. Within a priority, jobs run FIFO by creation order.
The response includes job id, kind, payload, attempt, lease token and expiry.

A worker sends `POST /v1/jobs/{id}/heartbeat` with the lease token and positive
`extend_seconds` to extend its lease.
`POST /v1/jobs/{id}/ack` supplies the lease token and result and completes work.
`POST /v1/jobs/{id}/nack` supplies the lease token, an error and optional retry
boolean (default true) or a `delay_seconds` override.
Only the current lease holder may complete or extend a job. A lease ends at its
`lease_expires_at`: from that instant heartbeat, ack and nack are 409, whether or
not the background loop has released the job yet. There is no grace period.
Ack moves a leased job to `succeeded`.
A result containing U+0000 (including object keys), or an error containing U+0000,
is 400 `invalid_result` or `invalid_error`; the valid lease remains usable.
Nack with retry moves a job to `retry_wait`; when due it becomes `queued`.
A failed job reaches `dead` when its maximum attempts are exhausted, or when
nacked with retry false. No job may execute more than its maximum attempts.
An expired lease records an unsuccessful attempt and releases the job for retry
unless the maximum attempts have been exhausted.

## Operator actions and retention

`POST /v1/jobs/{id}/cancel` cancels a nonterminal job.
Cancellation immediately makes a nonterminal job `cancelled` and revokes its lease.
If leased, cancellation records one `cancelled` attempt attributed to that holder
and consumes that delivery's attempt. Jobs cancelled before leasing consume none.
Cancelled jobs are never leased again; late ack/nack remain 409 and cannot revive them.
Already-started work stops cooperatively when the worker observes cancellation;
cancellation cannot roll back effects already performed by a remote service.

Heartbeat wire contract: a valid live lease returns 200
`{"ok":true,"cancel_requested":false}` and extends to now + `extend_seconds`.
After cancellation, the revoked holder's exact lease token returns 409
`{"error":{"code":"state_conflict","message":"job cancelled"},"cancel_requested":true}`.
This is a read-only observation: it neither extends nor recreates a lease, records
another attempt, nor authorizes ack/nack. A hash of the revoked token authenticates
this observation until job deletion/retention. Wrong/empty tokens, other nonleased
states, and cancelled jobs never leased (or cancelled before this feature) keep
409 `state_conflict` without the signal. Namespace ownership still applies (404
for foreign/missing jobs), and invalid extension values still return 400 first.
Repeated observations with the same revoked token are permitted and change nothing.

The sample worker checks heartbeat every 1 second while performing a job and
cancels its in-flight sleep or HTTP context on the authenticated 409 signal.
The heartbeat client has a 15-second request timeout; transport failure or an
ordinary 409 supplies no cancellation signal, so prompt stopping is not guaranteed
when the API is unreachable or the job has already been removed. Remote HTTP still
has its existing 5-second timeout. A job finishing before the next observation
may already have performed effects; its late completion is still rejected.
Operators should not interpret terminal cancellation as proof that remote work
has stopped. Other worker implementations must implement the observation contract.
`POST /v1/jobs/{id}/retry` permits an operator to retry a dead job, resets its
attempt count and makes it immediately available.
Succeeded and cancelled jobs cannot be retried.
Succeeded and cancelled jobs are deleted after `retention_hours`.
Dead jobs remain until an operator retries or explicitly deletes them; automatic retention exempts them.
The API returns complete stored results and errors.

## Schedules

`POST /v1/schedules` takes name, queue, kind, payload, cron, timezone and enabled.
Cron uses five fields: minute, hour, day of month, month and day of week.
Time zones use IANA identifiers, defaulting to UTC.
Schedule times follow the named zone's local civil time, including offset changes.
A schedule with a fixed hour runs once per matching local time: a time skipped by a
spring-forward transition runs once at the end of the gap, and a time that occurs twice
runs at its first occurrence only. Wildcard-hour schedules follow elapsed time.
The enabled flag defaults to true; disabled schedules do not enqueue jobs.
Schedule creation follows enqueue's payload rule: omitted means `{}`, explicit null
and non-object values are invalid. PATCH does not accept payload.
`GET /v1/schedules` lists schedules; `GET /v1/schedules/{id}` reads one.
`PATCH /v1/schedules/{id}` updates cron, timezone or enabled status; the next fire
must reflect the current configuration.
`DELETE /v1/schedules/{id}` removes it.
`POST /v1/schedules/{id}/run` enqueues one immediate execution.

## Notifications and operations

Terminal transitions produce a webhook outbox entry when the queue has a URL.
The entry includes job id, state, result and last error.
Notification results preserve their JSON type; results and errors are complete.
Each POST carries `X-Queuelet-Signature`, a hex HMAC-SHA256 over the body using
the configured secret. Delivery is successful only on a 2xx response.
Failed notifications are retried. Delivery is at least once.

`GET /v1/queues/{name}/stats` and `GET /v1/namespaces/{name}/stats` expose counts
by exact state, oldest queued age in seconds and completed jobs in the last hour.
Queued age uses the later of creation and readiness, over queued rows only.
A job released by lease expiry or worker deletion becomes ready at the release.
`GET /health` returns service health; `GET /metrics` returns minimal text metrics.
The background loop promotes scheduled jobs, expires leases, fires cron schedules,
sweeps retention and delivers webhooks. It should recover after a database restart. Webhook latency must not block core
lifecycle progress. Health reports database availability and loop degradation;
an unavailable authentication store returns 503 rather than invalid credentials.
`/health` answers within one second, degraded while the database does not answer;
authentication waits at most five seconds for the database before that 503.
An API that starts while its database is unreachable waits for it and then serves.


### Notification destinations, snapshots and acknowledgement

Queue create/PATCH rejects malformed or non-HTTP(S) webhook URLs with 400;
an empty URL disables future notification snapshots. Each completed job's outbox
row retains its destination, secret and body. Configuration edits apply to future
snapshots; deleting a job does not erase pending deliveries. A malformed legacy
row is counted as a failed attempt and rescheduled independently, so it does not
stop later rows.

"Delivered" means queuelet received a 2xx response before its two-second HTTP
client deadline. A timeout does not reveal whether the receiver completed its
side effect. Failed attempts retry after five seconds; duplicates are possible,
including when the receiver processed a timed-out request. Each pass attempts up to
20 due rows concurrently, least recently scheduled first, so slow or unreachable
destinations cannot keep other rows from being attempted. This lesson adds no
attempt cap, dead-letter state, timeout knob or backoff policy for notifications.
