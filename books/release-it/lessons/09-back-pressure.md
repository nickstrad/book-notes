# 09 · Back pressure: a bounded usage queue that cannot lose work

Tag: `release-it/lesson-09` · Fixes: F14 · Verify: `make verify LESSON=09` (before: `REF=release-it/lesson-08`)
Book: Ch. 5 "Create Back Pressure", "Shed Load"; Ch. 4 "Unbalanced Capacities"; Ch. 5
"Decoupling Middleware" as the next step

## The idea

Between a producer and a slower consumer, something has to give. An unbounded queue looks like it
absorbs the difference, but it only postpones it: the queue grows, memory grows, the oldest items
age into minutes, and a crash loses everything that was waiting. Chapter 5's **Create Back
Pressure** says to bound every queue and decide what happens when it is full: make the producer
wait (back pressure), refuse new work (shedding, lesson 7), or put the work somewhere durable. The
bound turns an invisible, growing liability into a visible, fixed one.

Usage events are billable. Losing one is unbilled usage, so "refuse" is not an option, and making a
tenant's statement wait on the billing partner would put a third party on the critical path. The
answer here is a small bounded queue in memory, a short wait for a producer when it is full, and an
**outbox**: a table in the node's own SQLite database where events the queue cannot take wait,
durably, for a sweeper to deliver. The database is the queue of record. That is a tiny version of
Chapter 5's **Decoupling Middleware**: once work has to survive a restart, it belongs in something
that persists.

## The flaw (F14, at `release-it/lesson-08`)

- `internal/metering/queue.go:18–29`: every event is appended to an unbounded slice.
- `queue.go:55–67`: one goroutine ships 50 at a time and, on any error, sleeps a second and retries
  the same batch. Against a partner that answers in a second, it ships 50 events a second while a
  node produces about 150.
- Nothing is persisted. SIGTERM ends the process at once (`cmd/node/main.go`), and every event
  still in the slice is gone.

## The patch

- **A bounded queue** (`internal/metering/queue.go`): `Capacity` = 1 024 events (`:21`). `Record`
  (`:60`) puts an event in a free slot. If the queue is full, it waits at most `EnqueueBudget` =
  20 ms for one. That is the back pressure: a statement can be delayed a little, never blocked.
  After that, the event goes to the outbox (`usage_enqueue_timeouts_total`).
- **Four shippers** (`ship`, `:129`) send batches of up to 100, each with its own timeout. A batch
  that fails goes to the outbox rather than being retried in memory. Event ids make redelivery
  idempotent, because metering drops repeats.
- **The outbox** (`internal/metering/outbox.go`): `usage_outbox(event_id primary key, body)` in
  `/data/_node.db`, written with `INSERT OR IGNORE` (`:36`). The **sweeper** (`sweep`, `:155`) runs
  every 5 s and delivers up to 500 rows per pass, deleting each batch once metering has
  acknowledged it.
- **Nothing lost on restart.** On SIGTERM the node spills the queue to the outbox before it exits
  (`Spill`, `:179`, wired at `cmd/node/main.go:80`). It writes what is queued, gives batches being
  sent up to 3 s to finish, and writes what is queued again, because statements finishing in that
  time still produce events. Anything still in flight is written too; if it was delivered after
  all, metering drops the duplicate. (The node still abandons its in-flight *requests* on SIGTERM;
  draining them is lesson 12.)
- `_node.db` gets a busy timeout and a single connection (`cmd/node/main.go:52`), because the
  shippers and the spill now write to it concurrently.
- The metering client gets 2 s for the whole exchange, headers included (lesson 2 gave it a 1 s
  header timeout). Against a partner that answers in a second, a 1 s header timeout counted every
  delivered batch as failed and sent it again.
- Metrics: `usage_queue_depth` (now ≤ 1 024), `usage_queue_oldest_age_seconds`,
  `usage_enqueue_timeouts_total`, `usage_outbox_rows`, `usage_sweeper_shipped_total`.

## Verify

```
make verify LESSON=09 REF=release-it/lesson-08      # before
make verify LESSON=09                               # after
```

Steady mix at 50 % for 60 s. The metering partner answers in 1 s (Toxiproxy `latency` 1 000 ms)
from t=1 s to t=30 s, and `docker compose restart node-a` (SIGTERM, then a new process) runs at
t=20 s. The script reads queue depth and oldest-event age from Prometheus. Afterwards it counts the
statements the nodes answered 200 from their own request logs (across node-a's restart), waits up
to 35 s for metering's distinct-event count to catch up, and checks the outboxes with DuckDB.

Before (`lesson-08`):

```
  usage queue: node-a max depth before the restart 2618, oldest event 17.22 s
  usage queue, both nodes, t=1-30 s: max depth 4076, oldest event 28.46 s
  p99 of successful statements: metering slow 70.0 ms, metering healthy again 12.0 ms
  statements the nodes ran 18762; usage events metering received 15841 (missing 2921); duplicates dropped 1346
FLAW REPRODUCED: max_queue_depth 4076 (want <= 1024); oldest_event_age_s 28.46 (want <= 10); events_missing 2921 (want <= 5)
```

After:

```
  usage queue: node-a max depth before the restart 155, oldest event 0.9769 s
  usage queue, both nodes, t=1-30 s: max depth 155, oldest event 0.9769 s
  p99 of successful statements: metering slow 12.0 ms, metering healthy again 9.0 ms
  statements the nodes ran 18772; usage events metering received 18772 (missing 0); duplicates dropped 566
  events left in the outboxes                       0
IMPROVED: max_queue_depth 155 <= 1024; oldest_event_age_s 0.9769 <= 10; events_missing 0 <= 5; outbox_rows_left 0 == 0
```

Before the patch, a slow partner grew the queue by about 100 events a second, its oldest event was
half a minute old, and the restart threw away 2 921 events: usage nobody will bill. After the
patch, four shippers keep up with a slow partner, so the queue never gets near its bound. The
restart spilled what was queued, the sweeper delivered it, and metering counted exactly one event
per statement. The tenants' statements did not notice the slow partner in either version (12 ms
p99), because the queue sits between them.

## Try changing

- Set `Shippers` to 1 and the toxic to 2 000 ms. The queue fills, statements start waiting their
  20 ms, and `usage_enqueue_timeouts_total` and `usage_outbox_rows` climb. Then remove the toxic and
  watch the sweeper drain it at its bounded rate.
- `docker compose kill node-a` (SIGKILL) instead of `restart`. Events in memory are lost again,
  because a spill needs a chance to run. Persisting on every `Record` would close that gap at the
  cost of a write per statement; the outbox exists so that you can choose.
- Make the outbox a Postgres table and the sweeper a separate service: that is Chapter 5's
  Decoupling Middleware at the next size up.

## Read more

*Release It!* 2nd ed., Ch. 5 "Create Back Pressure" and "Decoupling Middleware". The
transactional-outbox pattern (Chris Richardson, *Microservices Patterns*).

## Deviations from the plan

- **The node spills its queue to the outbox on SIGTERM.** With only the plan's mechanism (outbox
  on a full queue), every event waiting in the bounded queue at a restart is still lost, and the
  plan's own check ("metering has received every event exactly once, including the ones queued
  before the restart") cannot pass. The drain of in-flight requests stays in lesson 12.
- **Duplicates are reported, not required to be 0.** Delivery is at least once and idempotent by
  event id. Metering counts each event once (`missing 0`), and the 566 repeats it dropped are
  batches whose answers were lost when node-a restarted or the proxy was cycled to remove the
  toxic.
- **The metering client's header timeout is 2 s, not lesson 2's 1 s**, and `_node.db` gets
  `busy_timeout` and one connection. Both surfaced in this lesson's runs: the first duplicated
  every batch against a 1 s partner, and the second dropped events when four shippers wrote the
  outbox at once.
