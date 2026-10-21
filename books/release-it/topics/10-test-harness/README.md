# 10 · Test Harnesses for integration points

> **STUB** — teaching outline from `../../PLAN.md` (topic 10). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 5 "Test Harnesses"; Ch. 4 "Integration Points"
(socket-based protocols, HTTP protocols); Ch. 15 "The Testing Gap".
**Language:** Go — raw `net` gives control over TCP behaviour (partial writes, linger/RST), and
the hardened client shows the exact std knobs a reader will use at work.
**Run:** `./run.sh`.
**Budget:** 30 s (the naive run is the long one; per-probe deadline 3 s).

## The idea

Integration tests against a well-behaved stub prove only the happy path. A test harness is a
deliberately misbehaving fake that reproduces the network- and protocol-level failures the real
dependency will produce: connection refused, accept and never respond, slow bytes, garbage,
oversized responses, resets mid-body, 5xx. One harness, one port per misbehaviour, reusable across
projects. This is how the stability patterns get verified.

_Builder: 2–4 paragraphs; reproduce the book's list of harness behaviours and its point that the
harness is not a mock: it is a stand-in for the network itself._

## What you will see

`harness/` listens on 9 consecutive loopback ports, each with a named misbehaviour:

| # | Misbehaviour | How the harness does it |
|---|---|---|
| 1 | connection refused | port deliberately not listening |
| 2 | accept, never respond | accept, never read or write |
| 3 | read request, then silence | read until headers end, then nothing |
| 4 | slow headers | one byte per second |
| 5 | garbage | random bytes instead of HTTP |
| 6 | 200 with invalid body | valid headers, `Content-Type: application/json`, body `{"oops` |
| 7 | oversized body | 50 MB body |
| 8 | reset mid-response | `SetLinger(0)` then close after partial write |
| 9 | HTTP 500 | plain error status |

`probe/` calls every port and prints: `port`, `misbehaviour`, `outcome`, `elapsed`, `error_class`.

### Before

Naive probe: default `http.Client`, `io.ReadAll`, trusting `json.Unmarshal`. Hangs (bounded to 3 s
each by the runner's deadline), a parse failure treated as success, a 50 MB allocation.

```
NAIVE CLIENT: 6 of 9 failure modes not handled
```

### After

Hardened probe: dial/response/total timeouts, `io.LimitReader(1 MB)`, status-code check, strict
`json.Decoder` with `DisallowUnknownFields`, error classification. Every port yields a fast,
classified failure.

```
HARDENED CLIENT: 9 of 9 failure modes handled; slowest 503 ms
```

## How it works

_Builder: `harness/main.go` (one goroutine per port), `probe/main.go` with a `--mode naive|hardened`
flag or two files; `run.sh` starts the harness, runs both probes, stops it. Before/after here is
the same harness with a naive vs a hardened client._

## Try changing

_Builder: e.g. add port 10 "slow body" (headers fast, body 1 byte/s) and see which timeout catches it._

## Where to read more

Ch. 5 "Test Harnesses"; Ch. 4 "Integration Points"; the `test-integration-points` skill.

## Deviations from plan

None yet.
