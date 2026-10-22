# 02 · Timeouts on every integration point and every statement

Tag: `release-it/lesson-02` · Fixes: F02, F03, F04 · Verify: `make verify LESSON=02` (before: `REF=release-it/lesson-01`)
Book: Ch. 4 "Integration Points" (socket-based protocols, HTTP protocols), "Slow Responses",
"Blocked Threads"; Ch. 5 "Timeouts"

## The idea

Nygard calls integration points "the number-one killer of systems" (Ch. 4). Every socket, HTTP
call and database connection is a place where the other side can fail in a way your code was not
written for. The most dangerous way is not refusing; it is accepting the connection and then never
answering, or answering slowly. A caller with no time limit waits forever. While it waits it holds a
thread (here, a goroutine), a socket and whatever else it had checked out, and so does every caller
after it. Chapter 4's "Slow Responses" makes the related point that a slow answer is worse than a
fast failure, because it ties up resources on both sides.

The remedy in Ch. 5, "Timeouts", is simple to state: never wait without a limit. Put a timeout on
every call that crosses the network, on every resource checkout, and on work that can run away.
Choose the limits together, so that an inner wait always ends before the outer one gives up. A
timeout turns a hang into an error, and an error is something the rest of the system can handle,
retry, report or shed.

## The flaw (at `release-it/lesson-01`)

- **F02.** Every client is `&http.Client{}`: no connect timeout, no response-header timeout, no
  total. Control (`internal/routes/client.go:28`), nodes (`internal/nodeclient/client.go:26`), blob
  store (`internal/blob/client.go:25`), metering (`internal/metering/client.go:25`), control's
  lifecycle commands to nodes (`cmd/control/nodes.go:25`) and the node's heartbeat
  (`cmd/node/main.go:179`) all wait as long as the other side cares to take. Lesson 1 gave the
  clients a context, but it carried only the request id; `http.NewRequest` (`routes/client.go:74`,
  `nodeclient/client.go:77`) never bound the request to it.
- **F03.** The node runs statements with no deadline: `h.db.BeginTx(context.Background(), ...)`
  (`internal/engine/exec.go:73`), `tx.Exec(q, ...)` (`:164`), `tx.Query(q, ...)` (`:176`). Control's
  registry calls ran without a context and Postgres had no `statement_timeout`
  (`internal/registry/registry.go:34`).
- **F04.** `http.Server{Addr, Handler}` (`internal/httpx/httpx.go:24`) has no read, write or idle
  timeouts, so a client that sends half a request holds a goroutine and a socket indefinitely.
  HAProxy waits five minutes on both sides (`deploy/haproxy.cfg:12–13`).

## The patch

`internal/httpx/client.go:18` builds every client from three numbers: dial, response headers, and
total. The budget runs from the edge inward:

| Wait | Limit | Where |
|---|---|---|
| tenant request at the gateway (`/query`, `/execute`) | 2.9 s | `internal/api/handlers.go:23` (`withDeadline`) |
| gateway → node, per attempt | dial 200 ms, total 2.5 s | `internal/nodeclient/client.go:27` |
| statement on the node | 2 s | `internal/engine/exec.go:33` |
| gateway → control | dial 200 ms, total 500 ms (usage report: 10 s) | `internal/routes/client.go:33–34` |
| registry statement in Postgres | 2 s (`statement_timeout`) | `internal/registry/registry.go:36–44` |
| node → metering | 2 s; node → blob store 5 min (a floor of 200 KB/s for the 60 MB whale) | blob and metering clients |
| control → node, node → control | 2 s | `cmd/control/nodes.go:25`, `cmd/node/main.go:179` |
| servers | read headers 5 s, read 10 s, write 15 s, idle 60 s | `internal/httpx/httpx.go:29` |
| HAProxy | client and server 10 s, `http-request` 5 s | `deploy/haproxy.cfg` |

Each limit sits inside the one around it. The node stops a statement at 2 s, so its 504 reaches the
gateway inside the gateway's 2.5 s per-attempt limit, and the gateway answers inside its 2.9 s
deadline, which sits inside the 3 s our clients wait. A deadline equal to the caller's own timeout
races it, and the answer arrives as the caller hangs up.

Every outgoing request is now bound to its context (`http.NewRequestWithContext`). The retry
loop's sleep stops when the context ends (`nodeclient/client.go:66`), because a sleep is also a
wait. The gateway turns a dependency that ran out of time into `504 timeout` with the request id
(`handlers.go:212`).

On the node, every statement runs under a 2 s deadline derived from the request's context
(`exec.go:96`, `:189`, `:201`). When the deadline passes, the `ncruces/go-sqlite3` driver calls
`sqlite3_interrupt`, which stops the statement wherever it is. On this VM it lands within a
millisecond (kb entry `go-sqlite-drivers-cancellation-and-threads`). `classify` (`exec.go:42`)
reports it as `504 statement_deadline` and counts it in
`statements_cancelled_total{reason="deadline"}`. If the caller hangs up first, the statement is
interrupted the same way and counted with `reason="caller_gone"`.

Note what the patch does not do. `Execute` still begins its transaction with `h.db.Begin()`, which
is not bound to the context, and only the statements are. Binding the transaction would have
`database/sql` roll it back when the request ends. That would hide the leak lesson 3 is about, so
it is lesson 3's fix.

## Verify

```
make verify LESSON=02 REF=release-it/lesson-01      # before
make verify LESSON=02                               # after
```

Three phases.

1. **Steady reads** at 50 % of fleet capacity (300 req/s) for 45 s. From t=10 s to t=30 s, node-a's
   proxy accepts every connection and never answers (Toxiproxy `timeout` toxic, timeout 0, on the
   upstream stream).
2. **A runaway query.** The noisy tenant sends `WITH RECURSIVE ... x < 8000000` through the public
   API, about 30 s of node CPU, with a 3 s client timeout. Seven seconds later the script reads
   node-a's CPU and the cancellation counter.
3. **A stalled client** opens a connection to gateway-a, sends half a request, and waits up to 9 s
   for the server to hang up.

Before (`lesson-01`):

```
  max gateway in-flight during the outage     1130
  max gateway goroutines during the outage    4650
  gateway in-flight 6 s after the toxic left  1200
  node-a failures answered 504 by the gateway 0 of 2469 (0.0 %)
  node-a CPU 4-7 s after submit               0.495 cores
  gateway closed the stalled connection after 9.0 s (gave up waiting at 9 s)
FLAW REPRODUCED: max_inflight 1130 (want <= 300); max_goroutines 4650 (want <= 1000); outage_p99_s 3.021 (want < 3.0); outage_504_pct 0.0 (want >= 90); runaway_cpu_cores 0.495 (want <= 0.15); statements_cancelled NaN (want >= 1); cancel_late_ms NaN (want <= 50); stalled_client_held_s 9.0 (want <= 6)
```

After:

```
  max gateway in-flight during the outage     149
  max gateway goroutines during the outage    632
  gateway in-flight 6 s after the toxic left  2
  node-a p99 during the outage                2.914 s
  node-a failures answered 504 by the gateway 99.4 %
  node-a success rate t=34-44 s               100.0 %
  node-a CPU 4-7 s after submit               0.03 cores
  statements_cancelled_total{reason=deadline} 1
  node request duration when cancelled        2000.5 ms (0.5 ms past the 2 s deadline)
  gateway closed the stalled connection after 5.0 s
IMPROVED: max_inflight 149 <= 300; max_goroutines 632 <= 1000; outage_p99_s 2.914 < 3.0; outage_504_pct 99.4 >= 90; runaway_cpu_cores 0.03 <= 0.15; statements_cancelled 1 >= 1; cancel_late_ms 0.5 <= 50; stalled_client_held_s 5.0 <= 6
```

Before the patch, every node-a request became a goroutine waiting on a socket that would never
answer. In-flight work grew with the arrival rate for as long as the outage lasted, and 1 200
requests were still stuck after the fault had gone. Those requests were swallowed and will never
be answered, so the gateway would wait for them forever. After the patch, in-flight work is
bounded by rate × deadline. Every failure is a fast, attributable 504, and the moment node-a
answers again, so does everything else. The runaway query used to keep half a core busy for 30 s
for a client that had left; now it stops 0.5 ms after its 2 s deadline.

One thing in the after numbers is worth noticing. A node-a request takes 2.9 s to fail, not
2.5 s: the first attempt times out at 2.5 s, and the retry loop (F08) spends the last 0.4 s on a
second attempt. That is lesson 4.

## Try changing

- Run phase 1 with writes (`-write-frac 0.25` in the script). Writes the new deadlines cancel take
  `Execute`'s error path, which returns before `tx.Rollback()` (F05). Each one leaks a connection
  and that file's write lock, and node-a can reach its fd wall during the outage. Timeouts create
  new error paths, and lesson 3 is about what error paths leak.
- Set `RequestDeadline` to 3 s, the same as the client's timeout, and look at the share of 504s:
  the gateway's answer and the client's give-up race.
- Raise the node's `StatementDeadline` above the gateway's per-attempt 2.5 s. The node keeps
  computing for a gateway that has already left.

## Read more

*Release It!* 2nd ed., Ch. 4 "Integration Points" and "Slow Responses", Ch. 5 "Timeouts". Go's
`net/http` documentation on `Transport.ResponseHeaderTimeout` and the `http.Server` timeouts. The
SQLite documentation for `sqlite3_interrupt`.

## Deviations from the plan

- **The blackhole is the `timeout` toxic on the upstream stream, not the downstream one.** With the
  downstream toxic, Toxiproxy 2.12 wedged the node-a proxy for 15+ s after the toxic was removed
  while thousands of links were open: even fresh connections timed out. The upstream toxic drops
  requests before they reach the node, which the gateway sees the same way (accepted, never
  answered), and it is removed cleanly.
- **Phase 1 sends reads only.** With writes, the run sometimes reached node-a's fd wall through F05,
  as "Try changing" describes. That failure belongs to lesson 3 and made lesson 2's run
  nondeterministic.
- **The gateway's request deadline is 2.9 s, not 3 s**, so it sits inside the client's 3 s timeout.
  The node client's response-header timeout equals its 2.5 s total (the plan said 1 s for every
  client): a node sends headers only once its statement has finished, and a statement may
  legitimately take up to 2 s. The usage report gets its own 10 s client, because it is a heavy read
  by nature (lesson 8 makes it light).
- **Thresholds.** Outage p99 is `< 3.0 s` (before the client gives up), not `≤ 2.7 s`, because F08's
  retry runs to the 2.9 s deadline until lesson 4. A new threshold requires ≥ 90 % of node-a
  failures to be 504s from the gateway rather than client timeouts. The plan also expected
  in-flight work "not to fall when the toxic is removed". With the upstream toxic, the swallowed
  requests are never answered, and the before run indeed had 1 200 still in flight after removal.
- **Lab:** `make calibrate` was re-run at the start of checkpoint B and forgives one noisy step (the
  VM is shared and shows steal time): gateway 350, node 300, fleet 600 req/s. loadgen gained
  `-start-at` so the CSV's `ts_ms` and the script's timeline share one clock.
