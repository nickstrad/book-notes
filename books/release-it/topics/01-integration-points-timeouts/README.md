# 01 · Integration Points and Timeouts

> **STUB** — teaching outline from `../../PLAN.md` (topic 01). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 4 "Integration Points" (socket-based protocols; the 5 a.m.
problem; HTTP protocols; vendor API libraries); Ch. 5 "Timeouts"; Ch. 2 case study for the story.
**Language:** Go — the std `net/http` client exposes the three timeouts the book distinguishes
(connect, response, total), and goroutines make "one hung thread per call" literal.
**Run:** `./run.sh` (runs `before` then `after`; `./run.sh before`, `./run.sh after` run one side).
**Budget:** 15 s.

## The idea

Integration points are the number-one killer of systems. Their worst failure is not a clean error
but a hang: TCP accepts the connection and nothing ever comes back, or the listen queue is full
and even the connect stalls. Without a timeout the caller waits forever, and each waiting caller
is a thread the rest of the system no longer has. Timeouts are the most basic stability pattern;
every other pattern (breaker, bulkhead, fail fast) depends on calls being bounded in time.

_Builder: 2–4 paragraphs, cite the sections above; include the book's point that a socket
connection can be accepted by the kernel while the application never reads it._

## What you will see

A local `dependency` server with two misbehaving ports: one accepts connections and never writes a
byte; one sends headers then one byte every two seconds. A client with 10 worker goroutines issues
requests. A status line prints once per second: `in-flight`, `completed`, `failed`, `hung`.

### Before

Default `http.Client{}` (no timeout). All 10 workers stick on the silent port within a second;
`completed` stays at 0. A watchdog prints the state at 6 s and exits 1:

```
t=6s  in-flight=10 completed=0 failed=0 hung=10
HUNG: 10/10 workers blocked; production would still be waiting
```

### After

`net.Dialer{Timeout}`, `Transport.ResponseHeaderTimeout`, `http.Client{Timeout: 500ms}` and a
per-request `context.WithTimeout`. Each attempt fails within 500 ms with a named error; 100 attempts
finish in about 5 s; the status line shows failures, zero hung.

## How it works

_Builder: name the few lines that matter — where each of the three timeouts is set and which
failure mode each one catches. Layout: `dependency/main.go`, `before/main.go`, `after/main.go`,
one `go.mod`._

## Try changing

_Builder: e.g. remove only the dial timeout and see which port still hangs; lower the worker count._

## Where to read more

Ch. 4 "Integration Points" → "Socket-Based Protocols" and "HTTP Protocols"; Ch. 5 "Timeouts";
then topic 03 (what to do once calls time out repeatedly).

## Deviations from plan

None yet.
