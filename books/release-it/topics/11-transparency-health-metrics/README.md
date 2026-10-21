# 11 · Transparency: health checks, metrics, structured logs

> **STUB** — teaching outline from `../../PLAN.md` (topic 11). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 8 "Transparency" (designing for transparency; logging;
instance metrics; health checks); Ch. 10 "System-Wide Transparency" (logs and stats; what to
expose); Ch. 6 case study (vital signs; diagnostic tests).
**Language:** Go — `expvar` and `log/slog` are std, so the whole operations surface needs no
dependencies.
**Run:** `./run.sh`.
**Budget:** 25 s.

## The idea

A system without transparency cannot survive long in production: you cannot fix what you cannot
see. Every instance should expose a health check the balancer and control plane can act on (with
a degraded state, not just up/down), instance metrics (traffic, errors, latency, pool and queue
gauges, breaker states) and structured logs with request ids. The failures in topics 01–08 are
then visible before they are outages.

_Builder: 2–4 paragraphs; cite Ch. 8's list of what instance metrics to expose and Ch. 10's
"what to expose" guidance; mention that stability patterns must report their state._

## What you will see

A service whose dependency becomes slow at t = 5 s and fails at t = 8 s. A `watch` client prints a
dashboard every second from whatever the service exposes.

### Before

Black box: one `/` endpoint, `log.Printf("handled")`. The dashboard shows only client-side latency
rising, then errors. Nothing says why.

```
OPAQUE: the first signal of trouble was user-visible errors at t=8s
```

### After

`/health` (checks the dependency with its own short timeout; `ok` / `degraded` / `failed` with
reasons), `/metrics` via `expvar` (requests, errors, latency buckets, pool in-use, breaker state),
`log/slog` JSON logs with request id and duration. The dashboard shows `degraded: dependency p50
400ms` at t = 5 s, three seconds before any error.

```
VISIBLE: health went degraded at t=5s naming the dependency; first error at t=8s
```

## How it works

_Builder: the health handler's own timeout (never let the health check hang); the `expvar`
registrations; `slog` with a request-id attribute; `before/main.go`, `after/main.go`, `watch` may be
a flag in the same binary._

## Try changing

_Builder: e.g. make the health check call the dependency without a timeout and watch `/health`
itself hang (a classic)._

## Where to read more

Ch. 8 "Transparency"; Ch. 10 "System-Wide Transparency"; Ch. 6; the
`review-operational-transparency` skill.

## Deviations from plan

None yet.
