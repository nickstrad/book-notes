# Queuelet

A small durable job queue: Go HTTP API, PostgreSQL storage, workers, cron schedules
and a webhook outbox. The CLI and the HTTP API expose the same operations.

Requirements: Go 1.26, Docker with Compose, Python 3, curl. PostgreSQL tools are
available inside the database container; `jq` is useful for manual exploration.

```sh
make up
make seed
make verify LESSON=00
go run ./cmd/qctl queue ls
go run ./cmd/qctl enqueue imports --kind echo --payload '{"message":"hello"}'
go run ./cmd/qctl jobs ls imports
make logs
make down
```

The API listens at `http://localhost:18200`, the webhook sink at port 18201 and
PostgreSQL at port 15433. Seed keys are `acme-key`, `globex-key`, and `admin-key`.
Set `QUEUELET_URL` and `QUEUELET_KEY` to select a service and namespace.
`make seed` / `make reset` discard this stack's data and restore 300 mixed-state
jobs, queues emails/reports/imports, two schedules and two running workers.
`results/manifest.json` contains the fixture ids and keys. Workers consume emails
and reports; imports is available for manual work. The local fixture credentials
are not intended for deployment.

```sh
go run ./cmd/qctl worker create manual --kinds echo
go run ./cmd/qctl lease imports --worker-id 4 --max-jobs 1
go run ./cmd/qctl job ack 123 --lease-token TOKEN --result '{"done":true}'
go run ./cmd/qctl job get 123
go run ./cmd/qctl job attempts 123
go run ./cmd/qctl stats imports
```

Use the returned worker/job ids and lease token in those examples. Worker mode:
`go run ./cmd/qctl worker run imports --kinds echo,sleep,fail,crash,http`.
Echo returns its payload; sleep reads payload.seconds; fail reads payload.n (or
use kind fail3); crash exits the worker; http GETs payload.url with a timeout.

Queue verbs: create, ls, get, patch, pause, resume, purge, rm.
Job verbs: get, attempts, ack, nack, heartbeat, cancel, retry, rm.
`batch <queue> --json '{"jobs":[…]}'` enqueues a batch.
`schedule create --json '{"name":"hourly","queue":"imports","kind":"echo","payload":{},"cron":"0 * * * *","timezone":"UTC"}'`
creates a schedule; schedule ls/get/patch/run/rm manage it.
Commands accepting arbitrary configuration use `--json`. A raw HTTP request is
also useful when exploring request structure:

```sh
curl -s -H 'Authorization: Bearer acme-key' http://localhost:18200/v1/queues
PGPASSWORD=explorer psql -h localhost -p 15433 -U explorer -d queuelet
curl -s http://localhost:18201/admin/records
curl -s -X POST -H 'Content-Type: application/json' \
  -d '{"mode":"500"}' http://localhost:18201/admin/mode
```

The read-only explorer role can inspect every table. Hooksink modes are ok, 500,
slow and timeout. DELETE /admin/records clears its delivery log.

Read [SPEC.md](SPEC.md) for intended behavior. Source starts in cmd/api and
internal/api; jobs owns lease transitions, schedule owns cron firing, webhook owns
outbox delivery, stats owns aggregate reads, and store wraps database access.
The schema is in migrations. The API applies migrations when starting.

`make test` runs unit tests; `make checks` runs integration checks against the
running service. `make verify LESSON=NN REF=explore-it/lesson-MM` builds historical
binaries in a temporary independent Compose stack and runs current reproduction
scripts against them. `make dwell LESSON=NN` reports those probes across tags.
Test output belongs in ignored results/. `make down` deletes this stack's volume.

### Observe jobs and control lab time

`qctl job events ID` returns retained namespace-scoped history, including after
job deletion; `qctl watch QUEUE` polls histories for up to 500 jobs once per real second. Rows identify the
queue, worker, attempt, event time, and notification outbox/status/error. New
attempts retain their lease start separately from completion. Pre-hook leases
have unknown (`null`) starts; no timing history is invented. `event_at` uses
application time. API lifecycle JSON logs mirror committed events asynchronously;
a crash between logging and marking can repeat a log event (deduplicate event_id).
Worker JSON logs correlate processing/completion with the same job id.

The Compose **lab** enables `QUEUELET_TEST_CLOCK=1`. Production defaults to no
clock endpoint. With `QUEUELET_KEY=admin-key`, use `qctl clock set
2030-01-01T00:00:00Z`, `qctl clock advance 3600`, and `qctl clock reset`.
Only admin keys can change this global lab clock; namespace keys receive 403.
Set/advance freeze API semantic time; reset and API restart return to real UTC.
Enqueue, leases, retries, sweeps, cron, stats and event timestamps share it.
Background polling, worker execution/heartbeats, process sleeps, HTTP deadlines
and hooksink timestamps stay real. Advancing the API clock cannot accelerate a
worker's sleep or a partner's response. Avoid mixing independent clock experiments
on the same stack. SQL receives application timestamps; observational triggers
record them without choosing transitions. Event history is retained until lab
reset, independently of job retention, for exploration (production retention
policy for this diagnostic table is beyond this lab).
