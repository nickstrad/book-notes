# Release It! — breakdown plan

**Book:** *Release It! Design and Deploy Production-Ready Software*, 2nd edition, Michael T.
Nygard, Pragmatic Bookshelf, 2018. All chapter numbers below are from the 2nd edition (the 1st
edition, 2007, numbers chapters differently and lacks Dogpile, Force Multiplier, Interconnect,
Control Plane, Handling Versions and Chaos Engineering).

**Phase:** 1 of 4 (research & plan). This document is the brief for phase 2 (builder) and phase 3
(skills). Structure decisions here are final unless a builder records a deviation in the topic
README and in `STATE.md`.

---

## 1. The book's core argument

Software is not done when it passes QA; it is done when it survives production, and production
is hostile in ways development never shows. Real systems fail because of their integration
points and their coupling, not because of isolated bugs: a small fault (a slow reply, a thrown
exception in a `close()` call, a query that returns too much) becomes a crack, and the crack
propagates along whatever path the architecture has left open, until it is an outage. Nygard's
program has three movements:

1. **Create stability** by learning the recurring ways cracks start (stability antipatterns) and
   putting in the structures that stop them from spreading (stability patterns): timeouts,
   circuit breakers, bulkheads, fail fast, shedding load, back pressure, steady state.
2. **Design for production** so the system can be run: processes that are transparent (health,
   metrics, logs), interconnects that control demand, a control plane with mechanical advantage.
3. **Deliver and adapt**: deploy without downtime by sequencing changes and handling versions,
   then keep the system alive through adaptation and deliberate chaos.

The book's recurring unit of analysis is the **impulse** (a sudden shock) and the **stress**
(a sustained force) acting on the system, and the recurring question is "where does the strain
go?" (Ch. 3 "Defining Stability"). A transferable idea from this book is therefore one that
shows strain moving through a system and a structure that stops it.

## 2. Parts and chapters

| Part | Chapter | What it contributes to this breakdown |
|---|---|---|
| — | 1 Living in Production | Aiming for the right target: design for production, not for QA. The "a million dollars here" cost framing. Motivation only. |
| I Create Stability | 2 Case Study: The Exception That Grounded an Airline | An uncaught `SQLException` from a `close()` in a resource-pool path leaks connections, exhausts the pool, blocks every request thread; check-in kiosks go dark. Anchor story for topics 01 and 02. |
| | 3 Stabilize Your System | Defining Stability (impulse, stress, strain); Extending Your Life Span; Failure Modes; Stopping Crack Propagation; Chain of Failure. Conceptual frame for every topic. |
| | 4 Stability Antipatterns | Integration Points · Chain Reactions · Cascading Failures · Users · Blocked Threads · Self-Denial Attacks · Scaling Effects · Unbalanced Capacities · Dogpile · Force Multiplier · Slow Responses · Unbounded Result Sets. Source of every `before`. |
| | 5 Stability Patterns | Timeouts · Circuit Breaker · Bulkheads · Steady State · Fail Fast · Let It Crash · Handshaking · Test Harnesses · Decoupling Middleware · Shed Load · Create Back Pressure · Governor. Source of every `after`. |
| II Design for Production | 6 Case Study: Phenomenal Cosmic Powers, Itty-Bitty Living Space | A retailer's holiday season: diagnosis through vital signs and diagnostic tests when the system gives few. Motivates transparency (topic 11). |
| | 7 Foundations | Networking in the data center and cloud; hosts, VMs, containers. Infrastructure-level; not demoable in-process. Omitted as a topic. |
| | 8 Processes on Machines | Code; Configuration; **Transparency** (designing for transparency, logging, instance metrics, health checks). Source of topic 11. |
| | 9 Interconnect | Solutions at different scales; DNS; Load Balancing; **Demand Control** (how systems fail under load; shedding load and back pressure); routing; discovery. Source for topics 05 and 08. |
| | 10 Control Plane | Mechanical advantage; System-Wide Transparency (logs and stats, what to expose); configuration, provisioning, command and control. Feeds topic 11 and skill `review-operational-transparency`. |
| | 11 Security | OWASP Top 10, least privilege, configured passwords. Important but not this book's distinctive idea. Omitted as a topic. |
| III Deliver Your System | 12 Case Study: Waiting for Godot | A deployment whose long-running schema migration takes the site down for hours. Motivates topic 12. |
| | 13 Design for Deployment | The Fallacy of Planned Downtime; Automated and Continuous Deployment; **Phases of Deployment** (relax between phases; expansion; rollout; cleanup). Source of topic 12 and skill `plan-safe-release`. |
| | 14 Handling Versions | Help Others Handle Your Versions (nonbreaking vs breaking API changes); Handle Others' Versions. Source of topic 12. |
| IV Solve Systemic Problems | 15 Case Study: Trampled by Your Own Customers | A launch whose load tests modelled the wrong users (sessions, bots, scrapers); The Testing Gap. Feeds topics 06, 10 and skill `review-capacity-headroom`. |
| | 16 Adaptation | Convex returns; process and organization; evolutionary system and information architecture. Organizational; omitted as a topic, cited in skills where relevant. |
| | 17 Chaos Engineering | Breaking things to make them better; prerequisites; designing the experiment; injecting and targeting chaos; disaster simulations. Captured as skill `design-chaos-experiment`. |

## 3. Topics

Twelve topics, in reading order. Numbering follows the book's arc: the first nine are Part I
(cracks and the structures that stop them), then the test harness that proves them, then
transparency (Part II), then versions (Part III). Language counts: Go 5, Rust 3, Node 3, Bash 1.

Every topic: std-library only, no network beyond loopback, `./run.sh` finishes in under 30 s,
`before` bounded by a watchdog where it would otherwise hang. Output is a per-interval table or
ASCII timeline so `before` and `after` compare line by line.

### 01 · `01-integration-points-timeouts` — Integration Points and Timeouts

- **Source:** Ch. 4 "Integration Points" (socket-based protocols; the 5 a.m. problem; HTTP
  protocols; vendor API libraries); Ch. 5 "Timeouts"; Ch. 2 for the story.
- **Idea:** Integration points are the number-one killer of systems. Their worst failure is not a
  clean error but a hang: TCP accepts the connection and nothing ever comes back, or the
  listen queue is full and even the connect stalls. Without a timeout the caller waits forever,
  and each waiting caller is a thread the rest of the system no longer has.
- **Why core:** Nygard calls timeouts "the most basic" stability pattern and makes every other
  pattern (breaker, bulkhead, fail fast) depend on calls being bounded in time.
- **Teaching design:** A local `dependency` server exposes two misbehaviors: a port that
  accepts connections and never writes a byte, and a port that sends headers and then one byte
  every two seconds. A client with 10 worker goroutines issues requests; a status line prints
  once per second: in flight, completed, failed, hung.
- **Before:** default `http.Client{}` (no timeout). All 10 workers stick on the silent port
  within a second; completed stays at 0. A watchdog prints the state at 6 s and exits 1 with
  `HUNG: 10/10 workers blocked; production would still be waiting`.
- **After:** `net.Dialer{Timeout}`, `Transport.ResponseHeaderTimeout`, `http.Client{Timeout:
  500ms}` and a per-request `context.WithTimeout`. Each attempt fails within 500 ms with a named
  error; 100 attempts finish in about 5 s; the status line shows failures, zero hung.
- **Language:** Go. The std `net/http` client exposes exactly the three timeouts the book
  distinguishes (connect, response, total), and goroutines make "a hung thread per call"
  literal.
- **Before/after:** yes.
- **Run:** `./run.sh` (also `./run.sh before`, `./run.sh after`). Budget: 15 s.

### 02 · `02-blocked-threads-resource-pool` — Blocked Threads and a pool that fails fast

- **Source:** Ch. 4 "Blocked Threads" (spot the blocking; use caching carefully; libraries);
  Ch. 2 (the `close()` exception that drained the pool); Ch. 5 "Fail Fast", "Timeouts".
- **Idea:** Blocked threads are the proximate cause of most outages. Request threads block on a
  shared resource pool; the pool does not refill because a holder leaked its resource or is stuck
  on a slow call; soon every thread is blocked and even requests that never needed the resource
  are starved. Checking out with no timeout turns one leak into a frozen process.
- **Why core:** It is the mechanism of the book's opening case study and the reason the author
  insists on timeouts on every blocking call, not just network calls.
- **Teaching design:** A `Pool` of 4 connections guarded by `Mutex<Vec<Conn>>` + `Condvar`.
  16 worker threads serve a scripted stream of requests; 1 in 8 requests hits an error path.
  A monitor thread prints every 500 ms: available, checked out, threads waiting, requests served.
- **Before:** the error path returns early without returning the connection (the airline bug);
  `checkout()` uses `Condvar::wait` with no timeout. Available falls to 0 within ~2 s; waiting
  climbs to 16; served stops increasing. Watchdog at 5 s exits 1 with
  `FROZEN: 16 threads waiting on 0 connections; nothing was ever going to wake them`.
- **After:** two changes, shown separately in the output: (1) an RAII guard (`Drop` returns
  the connection) so the error path cannot leak; (2) `Condvar::wait_timeout(250 ms)` so
  `checkout()` returns `Err(PoolTimeout)` and the request fails fast instead of waiting. Served
  keeps increasing; a count of fast failures appears instead of waiting threads.
- **Language:** Rust. The remedy the book describes (return the resource on every path) IS the
  `Drop` guard; `wait_timeout` makes the bounded wait a different function name, not a flag.
- **Before/after:** yes.
- **Run:** `./run.sh` (cargo build is part of it; std-only so it is a few seconds). Budget: 15 s.

### 03 · `03-circuit-breaker` — Circuit Breaker

- **Source:** Ch. 5 "Circuit Breaker"; Ch. 4 "Integration Points", "Slow Responses",
  "Cascading Failures".
- **Idea:** When a dependency is failing, continuing to call it harms both sides: each call costs
  the caller a full timeout, and the dependency gets no room to recover. A breaker counts
  failures, trips open so calls fail immediately, and after a cooldown lets one trial call
  through (half-open) to test for recovery. State changes are logged and exposed to operations.
- **Why core:** It is the pattern that converts "timeouts" into "the system stays responsive
  during an outage", and the book's main answer to cascading failures.
- **Teaching design:** A simulated dependency scripted on a timeline: healthy (5 ms) for 2 s,
  down for 4 s (every call consumes the 300 ms timeout then fails), healthy again. A caller
  issues 20 req/s. Output: one row per second: breaker state, calls attempted, calls that
  reached the dependency, failures, mean latency.
- **Before:** no breaker. During the outage every call takes 300 ms; mean latency jumps 60×;
  the dependency absorbs ~80 calls while down; recovery is noticed only by luck.
- **After:** `enum State { Closed { failures }, Open { since }, HalfOpen }`; trips at 5
  consecutive failures; cooldown 1 s. During the outage calls fail in under 1 ms with
  `Err(CircuitOpen)`; the dependency receives ~4 probe calls instead of ~80; recovery is detected
  within one cooldown. Transition log lines `breaker: closed -> open` appear in the output.
- **Language:** Rust. An exhaustive `match` over the state enum is the whole pattern, and the
  compiler forces every transition to be handled; no dependencies.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 20 s (two 8 s timelines plus build).

### 04 · `04-bulkheads` — Bulkheads

- **Source:** Ch. 5 "Bulkheads"; Ch. 4 "Chain Reactions", "Cascading Failures", "Blocked Threads".
- **Idea:** Partition capacity so a failure in one compartment cannot sink the whole ship. The
  common sinking: one shared thread or connection pool serves both a critical path and a slow,
  noisy one; the slow one occupies every slot and the critical path dies of starvation.
- **Why core:** Bulkheads are the book's structural answer to coupling through shared
  resources; they decide the blast radius of every other failure.
- **Teaching design:** An in-process HTTP server with `/checkout` (5 ms) and `/report` (calls a
  dependency that is hanging to a 2 s timeout). A worker semaphore of 20 models the request
  thread pool. A generator sends 50 req/s to each path for 8 s. Output per second, per path:
  ok, rejected, p50 latency.
- **Before:** one shared semaphore. `/report` requests fill all 20 slots within a second;
  `/checkout` success rate falls toward 0 and its p50 climbs to the dependency timeout. Last
  line: `CHECKOUT STARVED: 20/20 workers held by /report`.
- **After:** separate semaphores per path (10 and 10), rejecting with 503 when a compartment is
  full. `/report` degrades (503s) while `/checkout` stays at 100% with 5 ms p50.
- **Language:** Go. Buffered channels as semaphores are the exact pool model, `net/http` gives a
  real server, and the change between `before` and `after` is a dozen lines.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 20 s.

### 05 · `05-chain-reactions-shed-load` — Chain Reactions and Shedding Load

- **Source:** Ch. 4 "Chain Reactions", "Cascading Failures", "Unbalanced Capacities", "Scaling
  Effects"; Ch. 5 "Shed Load", "Fail Fast", "Handshaking"; Ch. 9 "Demand Control" (how systems
  fail under load; preventing disaster).
- **Idea:** In a horizontally scaled layer, when one instance dies its share of load lands on
  the survivors. If they were near capacity, the next one falls, then the next: a chain
  reaction to zero. A layer that refuses work it cannot do (bounded queues, immediate 503s)
  survives at reduced capacity. Unbalanced capacities between layers make the first fall likely.
- **Why core:** It is the book's picture of how a healthy-looking cluster collapses, and
  shedding load is its answer to "too much traffic" at every scale.
- **Teaching design:** A discrete-time simulation, no network: 4 instances, each 100 req/s
  capacity, offered load 320 req/s; a round-robin balancer. At t = 3 s instance 1 is killed.
  Output: an ASCII timeline per second showing each instance's state (`ok`, `overloaded`,
  `dead`) and totals served/rejected/lost.
- **Before:** unbounded queues. Survivors take 107 req/s each; queues grow; an instance "dies"
  when its queue exceeds a memory threshold; load re-spreads; by about t = 8 s every instance
  is dead and served = 0. Last line: `CHAIN REACTION: 4 of 4 instances lost`.
- **After:** each instance bounds its queue and sheds the excess with an immediate rejection.
  Survivors serve 300 req/s and reject 20; no further loss. Last line shows capacity held.
- **Language:** Go. The simulation is small and the shedding loop is the same shape as the
  bulkhead rejection in topic 04, so a reader sees one idea in two places.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 10 s.

### 06 · `06-retry-storms-dogpile` — Dogpile, Self-Denial Attacks and Jittered Backoff

- **Source:** Ch. 4 "Dogpile", "Self-Denial Attacks", "Users", "Force Multiplier" (as a
  mention); Ch. 5 "Governor" (slow automated action down), "Handshaking".
- **Idea:** Synchronized behaviour by many clients produces pulses of demand: identical retry
  timers, caches that expire together, cron at the top of the hour, a fleet rebooting after a
  deploy, a marketing email that drops at 9:00. The pulse knocks the dependency over the moment
  it recovers. The remedies are randomness (jitter), exponential backoff, retry budgets, and
  governors on automation.
- **Why core:** It explains why systems fail *while* recovering, and it is the behaviour most
  developers unknowingly write into every retry loop.
- **Teaching design:** Node timers simulate 200 clients against a dependency with capacity 50
  req/s that goes down for 2 s. Output: an ASCII histogram of requests per 100 ms bucket over
  10 s plus totals: succeeded, failed, dependency overload ticks.
- **Before:** fixed 1 s retry delay. Every failure cohort retries in lockstep; on recovery the
  dependency gets 200 requests in one tick, exceeds capacity, fails again; the histogram shows
  repeating spikes. Last line: `DOGPILE: recovery failed N times under synchronized retries`.
- **After:** exponential backoff with full jitter (`random(0, base × 2^n)`), capped delay,
  maximum 3 retries. The histogram flattens under capacity and recovery sticks on the first try.
- **Language:** Node. Timers on a single event loop make the scheduling story literal, and retry
  loops are most often written in JavaScript HTTP clients.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 25 s (two 10 s simulations).

### 07 · `07-unbounded-result-sets` — Unbounded Result Sets and Slow Responses

- **Source:** Ch. 4 "Unbounded Result Sets" (black Monday; limits on other application-level
  protocols) and "Slow Responses".
- **Idea:** A query written for today's data returns "all matching rows". Production data grows;
  one day the result set is a million rows and the service stalls or dies of memory. Slow
  responses are worse than fast failures because they hold the caller's resources. The remedy is
  to bound what you ask for (`LIMIT`, pagination) and what you return, and to stream rather than
  materialize.
- **Why core:** It is the antipattern that hides longest (works for months) and the clearest
  example of the book's rule that the system must protect itself from its own data.
- **Teaching design:** An in-memory `orders` table whose size steps 1 k → 10 k → 100 k → 1 M
  rows, simulating months passing. A handler answers "all orders for customer X" each step. A
  100 ms heartbeat timer logs its drift (event-loop blocking). Output per step: rows, response
  time, heap MB, max heartbeat drift.
- **Before:** `query()` returns an array of every match; `JSON.stringify` the lot. Heap and
  response time scale with the table; heartbeat drift reaches hundreds of ms. Last line names
  the growth: `UNBOUNDED: response grew Nx with data; heap +M MB; event loop blocked X ms`.
  Builder may also offer `--max-old-space-size=64` in the README to make `before` crash for real.
- **After:** `query({ limit: 100, after: cursor })` returning an async iterator; responses stream
  pages. Heap and drift are flat across all four steps.
- **Language:** Node. `process.memoryUsage()` and a drifting `setInterval` make memory and
  blocking visible with no tooling; async iterators are the natural streaming form.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 20 s.

### 08 · `08-back-pressure` — Create Back Pressure (and when to shed instead)

- **Source:** Ch. 5 "Create Back Pressure", "Shed Load"; Ch. 9 "Demand Control";
  Ch. 4 "Unbalanced Capacities".
- **Idea:** When a producer outruns its consumer, something must give: the queue grows without
  bound (memory, then death), the producer is slowed (back pressure), or work is dropped (shed).
  Back pressure belongs inside a system boundary where the producer is yours to slow; at the
  boundary, shed load with a clear response instead of letting a stranger's demand fill your
  queue.
- **Why core:** It is the mechanism beneath "Unbalanced Capacities" and the book's rule that
  every queue must be bounded.
- **Teaching design:** A producer thread emits 1000 msg/s; a consumer handles 200 msg/s.
  Output every 500 ms: produced, consumed, queue depth, age of the oldest queued message, and
  (after) producer blocked time or messages shed.
- **Before:** `std::sync::mpsc::channel()` (unbounded). Depth and oldest-age grow linearly; at 5 s
  depth ≈ 4000 and age ≈ 4 s. Last line: `UNBOUNDED: queue grows 800/s; this runs until OOM`.
- **After:** two variants in one run: `sync_channel(100)` with blocking `send` (producer slows to
  200/s, depth ≤ 100, age ≤ 0.5 s), then `try_send` at the "boundary" where the overflow is
  counted and rejected (`TrySendError::Full`) with the consumer's throughput untouched.
- **Language:** Rust. Bounded versus unbounded is the difference between `sync_channel` and
  `channel`, and blocking versus shedding is `send` versus `try_send`: the lesson is two function
  names.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 15 s.

### 09 · `09-steady-state` — Steady State

- **Source:** Ch. 5 "Steady State" (data purging; log files; in-memory caching; avoid fiddling).
- **Idea:** For every mechanism that accumulates a resource, some other mechanism must recycle
  it. Logs that grow forever, tables never purged, caches without bounds all kill a correct
  system, just later, and usually at 5 a.m. Humans logging in to clean up ("fiddling") is a
  symptom, not a remedy.
- **Why core:** It is the book's answer to the slow-stress failures that are not about traffic
  at all, and the only pattern about time rather than load.
- **Teaching design:** `service.sh` simulates one request per tick, each writing a log line, a
  session file and a cache entry under a scratch dir; one tick is one simulated day. A disk
  budget (a fixed KB quota checked by the demo) stands in for the disk. Output: a per-day table
  of log KB, session count, cache entries, total KB, quota headroom.
- **Before:** no housekeeping. Total KB climbs every day; at the day it crosses the quota the
  service's write fails and it exits 1: `DISK FULL on day N: no mechanism ever recycled what the
  service accumulated`.
- **After:** `housekeeper.sh` runs alongside: rotates logs keeping 7, purges sessions older than
  3 days, caps the cache at 50 entries by age. Total KB plateaus by day 8 and stays flat through
  day 30.
- **Language:** Bash. Real-world recycling is cron + `find -mtime` + rotation; Bash is the
  language the remedy is actually written in.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 10 s (works in a `mktemp -d` dir, cleaned on exit).

### 10 · `10-test-harness` — Test Harnesses for integration points

- **Source:** Ch. 5 "Test Harnesses"; Ch. 4 "Integration Points" (socket-based protocols, HTTP
  protocols); Ch. 15 "The Testing Gap".
- **Idea:** Integration tests against a well-behaved stub prove only the happy path. A test
  harness is a deliberately misbehaving fake that reproduces the network- and protocol-level
  failures the real dependency will produce: connection refused, accept and never respond,
  slow bytes, garbage, oversized responses, resets mid-body, 5xx. One harness, one port per
  misbehaviour, reusable across projects.
- **Why core:** It is how the book says the stability patterns get verified, and it is the direct
  basis of the `test-integration-points` skill Nick asked for.
- **Teaching design:** `harness/` listens on 9 consecutive loopback ports, each with a named
  misbehaviour (refused — port deliberately not listening; accept-and-hang; read-then-silence;
  one byte per second headers; garbage bytes; 200 with invalid JSON body; 50 MB body; RST on
  first read via `SetLinger(0)`; HTTP 500). `probe/` calls every port and prints a table:
  port, misbehaviour, outcome, elapsed, error class.
- **Before:** naive probe: default `http.Client`, `io.ReadAll`, trusting `json.Unmarshal`. Table
  shows hangs (bounded to 3 s each by the runner's deadline so the demo ends), a panic-style
  parse failure, and a 50 MB allocation. Last line: `NAIVE CLIENT: N of 9 failure modes not
  handled`.
- **After:** hardened probe: dial/response/total timeouts, `io.LimitReader(1 MB)`, status-code
  check, strict `json.Decoder` with `DisallowUnknownFields`, error classification. Every port
  yields a fast, classified failure within 500 ms. Last line: `9 of 9 failure modes handled`.
- **Language:** Go. Raw `net` gives control over TCP behaviour (partial writes, linger/RST), and
  the hardened client shows the exact std knobs a reader will use at work.
- **Before/after:** yes (naive vs hardened probe against the same harness).
- **Run:** `./run.sh`. Budget: 30 s (the naive run is the long one; keep per-probe deadline 3 s).

### 11 · `11-transparency-health-metrics` — Transparency: health checks, metrics, structured logs

- **Source:** Ch. 8 "Transparency" (designing for transparency; logging; instance metrics;
  health checks); Ch. 10 "System-Wide Transparency" (logs and stats; what to expose); Ch. 6 case
  study (vital signs; diagnostic tests).
- **Idea:** A system without transparency cannot survive long in production: you cannot fix what
  you cannot see. Every instance should expose a health check the load balancer and control
  plane can act on (with a degraded state, not just up/down), instance metrics (traffic, errors,
  latency, pool and queue gauges, breaker states), and structured logs with request ids. The
  failures in topics 01 to 08 are then visible before they are outages.
- **Why core:** Part II's thesis; the book argues the operations view must be designed in, and
  every stability pattern should report its state.
- **Teaching design:** A service with a dependency that becomes slow at t = 5 s and fails at
  t = 8 s. A `watch` client prints a dashboard every second from what the service exposes.
- **Before:** black box: one `/` endpoint, unstructured `log.Printf("handled")`. The watch can
  only show the client-side latency climbing and, later, errors; nothing says why. Last line:
  `OPAQUE: the first signal of trouble was user-visible errors`.
- **After:** `/health` (checks the dependency with its own short timeout; reports `ok`,
  `degraded`, `failed` with reasons), `/metrics` via `expvar` (request count, error count, latency
  buckets, pool in-use, breaker state), `log/slog` JSON logs with request id and duration. The
  dashboard shows `degraded` at t = 5 s, before any error, and names the dependency.
- **Language:** Go. `expvar` and `log/slog` are std, so the whole operations surface is built
  with no dependencies.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 25 s.

### 12 · `12-handling-versions` — Handling Versions and the expand/contract rollout

- **Source:** Ch. 14 "Handling Versions" (help others handle your versions: nonbreaking vs
  breaking API changes; handle others' versions); Ch. 13 "Phases of Deployment" (relax between
  phases; expansion; rollout; cleanup) and "The Fallacy of Planned Downtime"; Ch. 12 case study.
- **Idea:** With many machines and no planned downtime, old and new versions run at the same
  time during every rollout. Make changes nonbreaking: add optional fields, never remove or
  redefine; be a tolerant reader (accept unknown fields, require only what you use). Sequence a
  breaking change as expansion → rollout → contraction, relaxing between phases. Treat others'
  versions defensively: validate, do not assume.
- **Why core:** It is Part III's main lesson and the practical consequence of "no planned
  downtime".
- **Teaching design:** A simulated rolling deploy: 3 server instances, a client fleet, a
  balancer picking a random instance per request. The change: v2 renames `total` to `amount` and
  adds `currency`. Output per rollout step: instances by version, requests ok/failed by cause.
- **Before:** strict v1 clients (`assert typeof body.total === 'number'`), strict v2 servers
  that reject v1 bodies. Error rate 33% → 66% → 100% as instances are upgraded one by one. Last
  line: `BREAKING ROLLOUT: N% of requests failed during mixed versions`.
- **After:** expansion phase (v2 sends both `total` and `amount` + `currency`, accepts both
  request shapes), clients made tolerant readers, rollout with zero errors, then contraction
  (drop `total`) only after every consumer reports v2. A `--skip-relax` flag in the README shows
  what happens if contraction runs before consumers moved.
- **Language:** Node. JSON APIs are where this bites, and strict versus tolerant parsing is a
  visible one-line difference in JavaScript.
- **Before/after:** yes.
- **Run:** `./run.sh`. Budget: 10 s.

### Deliberately omitted (decided, not forgotten)

| Idea | Chapter | Why not a topic |
|---|---|---|
| Let It Crash | 5 | Overlaps topic 02's fail-fast remedy; a supervisor demo earns a folder only after the twelve above exist. Candidate for a later addition. |
| Decoupling Middleware | 5 | Needs a broker to show honestly; violates the no-external-services rule. Cited in `suggest-stability-patterns`. |
| Governor | 5 | Mentioned in topic 06 and the capacity skill; a standalone demo would be contrived. |
| Foundations (networking, VMs, containers) | 7 | Infrastructure facts, not a demoable mechanism. |
| Code and Configuration | 8 | Build/config practices; covered by skill `review-operational-transparency` checklist items. |
| Security | 11 | Important but not this book's distinctive contribution. |
| Adaptation | 16 | Organizational and architectural guidance; cited in `plan-safe-release`. |
| Chaos Engineering | 17 | Captured as skill `design-chaos-experiment`; a cross-topic capstone demo is a candidate for later. |

## 4. Skills

Eight skills under `books/release-it/skills/<name>/SKILL.md`, each following the skills contract
in `AGENTS.md`. The three Nick named come first.

| # | Skill | What it does | Triggers | Source |
|---|---|---|---|---|
| 1 | `test-integration-points` | Enumerates every remote call in a codebase (HTTP clients, DB drivers, sockets, SDKs), lists the failure modes each must survive (refused, connect hang, response hang, slow bytes, garbage, oversized, reset, 5xx, protocol mismatch), checks which the existing tests cover, and produces a harness-based test matrix with stubs for the gaps and a check that timeouts actually fire. | Adding or reviewing a client to another service; "our integration tests pass but production fails"; before a launch. | Ch. 4 Integration Points; Ch. 5 Test Harnesses; Ch. 15 The Testing Gap; topics 01, 10. |
| 2 | `hunt-stability-antipatterns` | Walks the twelve Ch. 4 antipatterns over a codebase or design with concrete signals for each (calls without timeouts; pool checkouts without timeouts; `synchronized`/mutex around remote calls; queries without `LIMIT`; fixed retry delays; synchronized cron; automation acting on fleet-wide data without limits; front/back capacity ratios; deep-linkable expensive pages) and reports findings with evidence, blast radius and severity. | Design or code review; incident postmortem; "why did this fall over". | Ch. 3 Chain of Failure; Ch. 4; topics 01–08. |
| 3 | `suggest-stability-patterns` | Maps each found antipattern or symptom to candidate Ch. 5 patterns using the book's own cross-references, decides placement (caller, callee, boundary), names the parameters to set, and states the signal each pattern must expose to operations. | After `hunt-stability-antipatterns`; designing a new service's integration layer. | Ch. 5; Ch. 9 Demand Control; topics 01–09. |
| 4 | `budget-timeouts-and-retries` | Builds a nested deadline budget along a call chain from the edge SLO inward (each hop's total < parent's remaining), splits connect/response/total, and sets a retry policy: idempotent only, bounded count, exponential backoff with jitter, total retry time within budget, paired with a breaker, deadline propagated. | Setting any timeout; writing retry logic; tuning a client; a new latency SLO. | Ch. 5 Timeouts, Circuit Breaker; Ch. 4 Dogpile, Slow Responses; topics 01, 03, 06. |
| 5 | `review-capacity-headroom` | Lists layers and their capacities, checks front:back ratios and pool sizes × instances against downstream capacity, spots O(n²) point-to-point fan-out, estimates self-inflicted peaks (marketing, deploys, cache expiry), verifies shed-load thresholds and 503 behaviour the balancer honours, and recommends headroom and governors on autoscaling. | Scaling a layer; sizing pools; planning a launch or campaign; configuring autoscaling. | Ch. 4 Scaling Effects, Unbalanced Capacities, Users, Self-Denial Attacks; Ch. 5 Shed Load, Governor, Handshaking; Ch. 9 Load Balancing, Demand Control; Ch. 15; topics 04, 05, 08. |
| 6 | `review-operational-transparency` | Checklist review of a service's operations surface: health endpoint semantics (what it checks; degraded vs failed), metrics (traffic, errors, latency distribution, pool/queue/breaker gauges, business counters), logs (structured, request id, levels, no secrets, rotation), config and version exposure; ends with "would a slow dependency be visible before users see errors?" | New service before first deploy; "we had an outage and could not tell what was wrong"; adding a stability pattern (it must report state). | Ch. 8 Transparency; Ch. 10 System-Wide Transparency; Ch. 6; Ch. 5 Steady State (log rotation); topics 09, 11. |
| 7 | `plan-safe-release` | Classifies a change as additive or breaking; for breaking changes produces an expansion → rollout → contraction plan with tolerant-reader checks on consumers, the database sequence (add, dual-write, backfill, switch, drop later), rollout order (tolerant consumers first), relax-and-verify steps between phases, and rollback per phase. | Schema change; API change; "can we deploy this without downtime"; a rollout plan. | Ch. 13 Phases of Deployment, Fallacy of Planned Downtime; Ch. 14; Ch. 12; Ch. 16 (evolutionary architecture); topic 12. |
| 8 | `design-chaos-experiment` | Produces an experiment card: prerequisites (monitoring, recovery path, steady-state definition, kill switch), hypothesis, blast radius and opt-in, injection (kill instance, add latency, fail dependency), targeting, observation plan, automate-and-repeat cadence, and a people-focused disaster simulation variant. | "Test our resilience"; game day; after adding stability patterns to prove they work. | Ch. 17; topics 03, 04, 05, 10 as injection ideas. |

## 5. Build order and budgets (brief for phase 2)

Build in this order; each is independent, so the builder may parallelize by language:

1. Go: 01, 04, 05 (shared mental model: pools, semaphores, shedding), then 10, 11.
2. Rust: 02, 03, 08 (one cargo package each, `src/bin/before.rs` and `src/bin/after.rs`).
3. Node: 06, 07, 12 (`before.mjs`, `after.mjs`, shared nothing).
4. Bash: 09.
5. `books/release-it/run-all.sh` (runs every `run.sh`, prints a pass/fail table, exits non-zero
   on any failure) and `books/release-it/README.md` (landing page: argument in one paragraph,
   topic table, skills table).

Rules for the builder:

- Read `AGENTS.md`, then this plan, then the topic's stub README; fill the stub, keep its section
  order, and delete the "STUB" banner.
- `run.sh` is bash, `set -euo pipefail`, checks its toolchain (`go`, `cargo`, `node`), builds, runs
  `before` then `after`, and ends with one line summarising both.
- A `before` that would hang MUST be bounded by a watchdog and exit 1 with a last line that names
  the failure (the review bar checks this). `run.sh` treats that exit 1 as the expected result of
  `before` and still exits 0 overall if `after` passes.
- Keep every run under its budget on this VM. If a scripted timeline is too slow, shorten the
  timeline, do not raise the budget.
- Expected-output sketches in the stubs are targets, not contracts: match the shape and the
  named last line; exact numbers may differ.
- Record any deviation from this plan in the topic README under "Deviations from plan" and in
  `STATE.md`.

Phase 3 (skills) follows the table in section 4 and the skills contract. Phase 4 review runs
`run-all.sh` from a fresh clone and spot-checks one citation per topic and per skill.
