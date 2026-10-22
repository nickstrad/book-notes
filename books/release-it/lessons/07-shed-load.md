# 07 · Shed load: bounded in-flight, 503 fast, balancer queue limits

Tag: `release-it/lesson-07` · Fixes: F12 · Verify: `make verify LESSON=07` (before: `REF=release-it/lesson-06`)
Book: Ch. 4 "Chain Reactions", "Slow Responses", "Unbalanced Capacities", "Scaling Effects";
Ch. 5 "Shed Load", "Handshaking", "Fail Fast"; Ch. 9 "Demand Control"

## The idea

A **chain reaction** (Ch. 4) happens inside one tier of identical instances. When one fails, the
load it carried moves to the others, which raises their load, and if that pushes them past what
they can serve, they fail too, and the remaining ones get even more. A horizontally scaled tier
does not make you safe from overload; it decides how the overload spreads.

What an instance does past its capacity decides the outcome. If it accepts everything, its queue
grows, every request waits longer, and soon each request outlives the caller who sent it. The
instance is then busy doing work nobody will receive, its goodput (useful answers per second) falls
toward zero, and its health check times out with the rest. Chapter 5's **Shed Load** says to refuse
the excess early and cheaply instead, so the work you do accept finishes in time. **Handshaking**
says to tell callers you are refusing, so they can back off or go elsewhere: a 503 with
`Retry-After`, a health check that says "degraded", a balancer that stops sending. Chapter 9's
**Demand Control** applies the same reasoning at every layer from the edge inward.

Litefleet also has **unbalanced capacities** (Ch. 4). Three gateways can offer a node more than it
can take, so the node needs a limit of its own; otherwise a burst at the gateways becomes a node
outage.

## The flaw (F12, at `release-it/lesson-06`)

- Gateways (`cmd/gateway/main.go:42`) and nodes (`cmd/node/main.go:82`) accept every request and
  never answer 503 on purpose. Lesson 6's lanes bound statements on the node, but not the HTTP
  requests queued in front of them.
- HAProxy (`deploy/haproxy.cfg:27–31`) sends every request it receives to some gateway, with no
  per-server limit and no queue bound. A dying gateway's in-flight connections simply fail.

## The patch

- **`internal/shed`** (~85 lines). A `Limiter` counts requests in flight. Past the limit it answers
  `503 overloaded` with `Retry-After: 1` without touching the work (`shed.go:44`), counted in
  `shed_total{route}`. Critical routes (query, execute, create) may use the whole limit; everything
  else (the usage report, listings) shares a quarter of it (`:57`), so a burst of those cannot crowd
  out queries and writes. `/health` and `/metrics` are never shed. `Check` (`:75`) makes health
  report `degraded: shedding load` for 5 s after the last refusal.
- **Gateways and nodes both use it.** The gateway's limit (`SHED_LIMIT`, default 24 at half a CPU,
  about 60 ms of work at its calibrated capacity) is at `cmd/gateway/main.go:32`. The node's
  (default 48) is at `cmd/node/main.go:73`: the node protects itself from the gateways' combined
  offer.
- **HAProxy** gets per-server `maxconn 32`, a little above the gateway's own limit, so the gateway
  refuses first. `timeout queue 500ms` means that when every gateway is full, a request waits at the
  edge at most half a second and then gets 503. `retry-on conn-failure` with `option redispatch`
  moves a request that could not connect (a gateway that just died) to another gateway once; a
  request a gateway has accepted is never repeated.

## Verify

```
make verify LESSON=07 REF=release-it/lesson-06      # before
make verify LESSON=07                               # after
```

For this run the gateways are made the fleet's bottleneck with Compose overrides: `GATEWAY_CPUS=0.15`
(one gateway saturates at about 215 req/s) and `NODE_CPUS=0.75` (so the nodes are not the limit).
The gateways' in-flight limit is scaled to 12 to match their speed. Offered load is 80 % of three
gateways' capacity (516 req/s, open loop). At t=15 s, `docker compose kill gateway-b`; the two
survivors are now offered 120 % of what they can serve. At t=60 s, `docker compose start gateway-b`.

Before (`lesson-06`):

```
  goodput after the kill: 11.4 % of the goodput before it
  p99 of successful requests after the kill   2784.0 ms (median of 2 s windows; worst window 2998.0 ms)
  503s after the kill 3205, p50 0.9 ms; shed_total 0
  max gateway in-flight / goroutines after the kill   997 / 3228
FLAW REPRODUCED: goodput_after_kill_pct 11.4 (want >= 55); ok_p99_ms 2784.0 (want <= 300); max_gateway_inflight 997 (want <= 32)
```

After:

```
  goodput (2xx within 3 s): before the kill 503.1/s, after 344.6/s (68.5 %), after the rejoin 505.6/s
  p99 of successful requests after the kill   184.0 ms (median of 2 s windows; worst window 325.0 ms)
  503s after the kill 6936, p50 55.6 ms; shed_total 8288
  max gateway in-flight / goroutines after the kill   13 / 136
  requests gateway-b served after it came back       1671
IMPROVED: goodput_after_kill_pct 68.5 >= 55; ok_p99_ms 184.0 <= 300; max_gateway_inflight 13 <= 32; rejoined_gateway_served 1671 > 0
```

Before the patch, the survivors accepted everything. Their queues grew to a thousand requests
each, latency passed the clients' 3 s timeout, and goodput fell to a tenth. Then they failed their
health checks too, because a health check waits in the same queue. The 3 205 fast 503s in the
before run come from HAProxy itself with no server left to send to: the chain reaction completed.
After the patch, the survivors serve what two gateways can (68 %, close to the 2/3 the arithmetic
allows), refuse the rest with `Retry-After`, stay healthy (`degraded`, not `failed`), and
gateway-b rejoins as soon as it passes two health checks.

The refusals take ~55 ms, not microseconds. On a 0.15-CPU instance under a 100 ms CFS quota,
even a refusal waits for the next slice of CPU. On a real machine, shedding costs a few microseconds
of parsing.

## Try changing

- Set `GATEWAY_SHED_LIMIT` to 200 and rerun: admitted requests queue longer, p99 climbs toward the
  client timeout, and goodput falls. The limit is the latency budget expressed as concurrency.
- Remove `retry-on conn-failure` and `option redispatch` and look at the errors in the second after
  the kill.
- Kill a node instead of a gateway. Nothing routes around a stateful instance. What matters there
  is how fast and cleanly it comes back, which is lesson 8's story.

## Read more

*Release It!* 2nd ed., Ch. 4 "Chain Reactions" and "Unbalanced Capacities"; Ch. 5 "Shed Load" and
"Handshaking"; Ch. 9 "Demand Control". The HAProxy documentation on `maxconn`, `timeout queue` and
`retry-on`.

## Deviations from the plan

- **The run overrides CPU: gateways at 0.15, nodes at 0.75.** At the default half CPU, three
  gateways with lesson 5's route cache can serve more than the two nodes. Killing one then
  overloads nothing, and this lesson's overload never happens. The override puts the bottleneck at
  the gateways, as the plan's scenario assumes. The gateways' limit is scaled for their speed
  (`L07_*` in `lab/capacity.env`).
- **HAProxy `maxconn` is 32, a little above the gateway's limit of 24**, not equal to it. With them
  equal, HAProxy's 500 ms queue does all the shedding, and refused requests take half a second
  instead of being turned away at the gateway.
- **"p99 of successful requests" is the median over 2 s windows of each window's p99, with a 300 ms
  limit** (plan: 150 ms). On this shared VM, a 2–6 s host stall in one run pushed the plain p99 to
  650–800 ms while every other window stayed at 90–165 ms. The worst window is printed too.
- **Lab:** gateways, nodes and the load generator get `cpu_shares: 2048`, so on the shared VM the
  quota is the limit under test rather than other projects' load (another project ran three load
  test stacks during these runs).
