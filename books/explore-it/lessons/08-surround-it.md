# 08 — Surround it: the neighbours fail in ways the code never asked about

An ecosystem model draws the system as one box among its neighbours: the database, the
workers, the webhook partner, the clock and the other tenant. Each line between boxes is a
boundary. For each neighbour you ask what crosses the line, where the code checks it, and
what happens when the neighbour misbehaves. Queuelet's tenant boundary held. The failures
were in the neighbours the code took for granted: a database that is absent when the API
starts, a database that is connected but frozen, a partner that answers slowly, and local
times that a daylight-saving change skips or repeats.

The sources are Hendrickson, *Explore It!* (2013), Chapter 9.1 “Diagramming the
Ecosystem”, 9.2 “Trust Boundaries”, 9.3 “What If?” and 9.4 “Follow the Data”. These are
**approved-plan references**
([E08](../PLAN.md#e08--surround-it-the-neighbours-the-boundary-and-the-clock)); full
book-text validation is still pending. The database, partner and DST conclusions below come
from this project's evidence, not from claims about what the book says of PostgreSQL or cron.

The blind [exploration record](../exploration/experiments/08-surround-it/revealed.md),
committed in the exploration commit before this one, is unchanged. It contains
21 stated hypotheses, 20 confirmed, 1 refuted, 1 contradicted code-reading prediction,
4 reports and 7 newly observed ecosystem cells. Process timestamps give about 22 minutes
of the 60-minute budget. The explorer excluded P3b's sibling predictions from its score
because they were logged after the run.

The refuted hypothesis is the lesson's sharpest moment. Reading `internal/api/api.go:128`
at lesson-07, the explorer predicted that `/health` would answer 503 within about a second
while PostgreSQL was frozen: the code passes a one-second context to `PingContext`. The
running system held the request for the whole 11.8-second freeze and then answered 200.
The deadline was in the code, but the driver did not honour it.

The plan expected four planted defects here: D21 (namespace leak), D22 (the loop dies on a
database error), D23 (DST offset captured at creation) and D24 (signature over different
bytes). All four were repaired in E01. The explorer re-probed each and recorded it as
holding. They are **early fixes**, neither E08 discoveries nor missed planted defects, and
they run here as unchanged controls. U36–U38 and A5 are the new findings.

## The faults at lesson-07

All source pointers refer to `explore-it/lesson-07`, relative to `project/`.

| Item | Report | Previous-tag location | Before | Repair |
|---|---|---|---|---|
| U36 | [E08-F01](../exploration/defects/E08-F01-api-does-not-recover-from-start-without-database.md) | `cmd/api/main.go:19-22`; `deploy/compose.yml` (`api`, `postgres`: no restart policy) | an API started while PostgreSQL is unreachable calls `log.Fatal` and stays down after the database returns; crashed API or database processes stay down | wait for the database at startup; `restart: unless-stopped` for both |
| U37 | [E08-F02](../exploration/defects/E08-F02-slow-destination-starves-all-later-notifications.md) | `internal/webhook/ship.go:17`, `:21` | 20 rows to a slow destination fill every pass (20 × 2 s) in id order, so later rows from every tenant are never attempted | attempt each pass's rows concurrently, least recently scheduled first |
| U38 | [E08-F04](../exploration/defects/E08-F04-frozen-database-hangs-health-and-api.md) | `internal/api/api.go:128`, `:163` | a frozen database hangs `/health` and authentication for the whole freeze, then they answer 200 | bound health (1 s) and authentication (5 s) independently of the driver; 503 when exceeded |
| A5 | [E08-F03](../exploration/defects/E08-F03-dst-gap-skips-and-overlap-double-fires.md) | `internal/schedule/fire.go:22` | spec silent; a fixed-time schedule in the spring-forward gap does not run that day, and one in the repeated hour runs twice | classic cron rule (decision below) |

## Before and after the ecosystem model

The [ecosystem model](../exploration/models/ecosystem.md) was not new in E08. E01 created it
and already probed neighbours there: a PostgreSQL restart and stop, webhook partner failures,
workers during an outage, another tenant's key, and the missing clock control; E02 revised it.
E08 revisited that model and extended it with neighbour states the earlier runs had not tried:
a database absent at startup, a frozen database, a shared shipper under one slow destination,
and the DST transitions. Every E08 finding sits on one of its lines:

| Boundary | What-if | Before (lesson-07) | After (lesson-08) |
|---|---|---|---|
| tenant ↔ tenant | the other tenant's key on our ids | 404 everywhere (D21 control) | unchanged |
| API ↔ PostgreSQL, running | restart the database | ≈1 s of 503, loop continues (D22 control) | unchanged |
| API ↔ PostgreSQL, at start | the database is missing when the API starts | API exits 1 and stays down | API waits, serves 1–2 s after the database returns |
| API ↔ PostgreSQL, frozen | the database stops answering | every request hangs, then 200 | `/health` 503 at 1 s, authentication 503 at 5 s |
| API ↔ partner | 22 rows to a destination that times out | later notifications never attempted | another tenant's notification delivered in ≈2 s |
| API ↔ clock | 2027-03-14 and 2027-11-07 in New York | 02:xx skipped that day; 01:xx twice | once per civil day; once in the repeated hour |
| outbox → partner | signature over the bytes received | holds (D24 control) | unchanged |

The charter's prediction, "every route is namespace-scoped and the loop survives a database
restart", was confirmed for exactly what it said. It was refuted at three boundaries the
sentence did not name: absent at start, frozen, and the tenant boundary inside the shipper.
That is the ecosystem model's argument (9.3): the bugs live in the neighbour's state you
did not list.

## Focused implementation

**U36.** `main` (`cmd/api/main.go:27`) treats a failed startup ping as an outage to wait
out. It logs the error once a second and pings again until the database answers or the
process is told to stop. Migration failures after a successful ping are still fatal.
The Compose file gives `api` and `postgres` `restart: unless-stopped`, so a process that
dies on its own comes back. An operator `stop` is respected.

**U37.** `Ship` (`internal/webhook/ship.go:21`) selects up to 20 due rows ordered by
`next_at, id` and delivers them concurrently, one goroutine per row (`deliver`, `:38`).
A pass now costs at most one two-second client deadline. A failed row's `next_at` moves
five seconds past the pass, behind rows that have not been tried. The two-second deadline,
the five-second retry, at-least-once delivery and the absence of an attempt cap are
unchanged (SPEC already excluded a cap or backoff policy). The loop comment that claimed
the single shipper preserved delivery order was removed: SPEC never promised an order,
and keeping it is what let one destination stop all the others.

**U38.** `bounded` (`internal/api/api.go:209`) runs a database call in a goroutine and
stops waiting after its limit. lib/pq keeps waiting for an unresponsive server after its
context is cancelled, so the context alone could not bound the call. That mechanism is the
explorer's inference, and the after-run evidence is consistent with it. `/health` uses a
one-second bound (`:129`), and authentication uses five seconds (`:164`) before its
existing 503 `auth_unavailable`. An abandoned call finishes in the background when the
server answers; the pool of 16 bounds the connections they hold, not the goroutines. Handlers past
authentication are not bounded. A freeze that begins mid-request still holds that request
(see below).

**A5 decision: a schedule with a fixed hour runs once per matching local time; a time
skipped by a spring-forward transition runs once at the end of the gap, and a time that
occurs twice runs at its first occurrence only; wildcard-hour schedules follow elapsed time.**

`Next` (`internal/schedule/fire.go:23`) keeps robfig's elapsed-time behaviour when the hour
field starts with `*`. For fixed hours it walks wall-clock times in a zone-free calendar
and maps each one to an instant with `civil` (`:44`). `civil` picks the earliest instant
that shows that wall time, or the end of the gap (`ZoneBounds`) when none does. A candidate
at or before the previous fire is skipped, which removes the second 01:30. This is the
classic cron convention named in the report; robfig's documented "skipped interval is not
run" is the rejected alternative. The D23 behaviour (09:00 and 02:00 follow the offset
change outside the transition hour) is a control at both tags.

## Reproduce and check

From `books/explore-it/project`, set the isolated environment and run sequentially:

```sh
export COMPOSE_PROJECT_NAME=queuelet-fix08 QUEUELET_IMAGE_TAG=fix08
export QUEUELET_API_PORT=18390 QUEUELET_HOOK_PORT=18391 QUEUELET_PG_PORT=15623
export QUEUELET_URL=http://localhost:18390 QUEUELET_HOOK_URL=http://localhost:18391
export QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15623/queuelet?sslmode=disable'
make up
make verify LESSON=08
make verify LESSON=08 REF=explore-it/lesson-07
make checks
make test
go vet ./...
make down
```

The public [probe](../project/lab/repro/probe08.py) uses public routes, Docker lifecycle
(`stop`, `start`, `restart`, `pause`), a second hooksink container on the lab network,
a host `kill -9` for the crash siblings and read-only SQL. It uses no fixture SQL writes
and no hooks inside the service. `docker kill` would not do for a crash: Docker treats it
as an operator stop and does not apply the restart policy (observed while writing U36).
Missing controls, unexpected statuses and transport failures are PROBE ERROR, not ABSENT.
See the [repro README](../project/lab/repro/README.md) for each probe's bounds.

### Results (final run, sequential, isolated stack `queuelet-fix08`)

Ledger: [08-opus-acceptance.json](../register/runs/08-opus-acceptance.json); one
`08-opus-<step>.log` per step. Instruments were frozen in
[08-instrument-hashes.json](../register/runs/08-instrument-hashes.json) before the final
pair and rechecked after the suite (`08-opus-integrity.log`).

| Step | Exit | Seconds |
|---|---|---|
| `make up` | 0 | 43.8 |
| `make verify LESSON=08` | 0 | 141.2 |
| `make verify LESSON=08 REF=explore-it/lesson-07` | 0 | 134.9 |
| `make checks` (85 check functions) | 0 | 573.9 |
| `make test` | 0 | 0.3 |
| `go vet ./...` | 0 | 0.3 |

At lesson-08:

```text
ABSENT: U36 — api during outage='running Up 4 seconds'; health 200 within 30 s of database start=True (0.6s); after process crash recovered={'api': True, 'postgres': True}
ABSENT: U38 — during freeze /health=503 after 1.0s (bound 503 ≤3s); authenticated GET=503 after 5.0s (bound 503 ≤8s); recovered after unpause
ABSENT: A5 — spring gap days={'gap-0230': ['2027-03-14', '2027-03-15'], 'gap-0200': ['2027-03-14', '2027-03-15']}; fall-back 01:30 runs=['2027-11-07T05:30:00Z']; 09:00 and hourly controls hold
ABSENT: U37 — acme notification delivered after=2.1s (bound 20 s) behind 22 globex rows to a timing-out destination; slow destination requests=22; acme outbox attempts|delivered=1|t
ABSENT: D21 / D22 / D23 / D24 (unchanged early-fixed controls)
FIXED AND CHECKED: 4 absent; regression checks pass
```

At lesson-07 (historical images and Compose file, current scripts):

```text
PRESENT: U36 — api during outage='exited Exited (1) 4 seconds ago'; health 200 within 30 s of database start=False; api after='exited Exited (1) 35 seconds ago'
PRESENT: U38 — during freeze /health=200 after 11.31s (bound 503 ≤3s); authenticated GET=200 after 11.32s (bound 503 ≤8s); recovered after unpause
PRESENT: A5 — spring gap days={'gap-0230': ['2027-03-15'], 'gap-0200': ['2027-03-15']}; fall-back 01:30 runs=['2027-11-07T05:30:00Z', '2027-11-07T06:30:00Z']; 09:00 and hourly controls hold
PRESENT: U37 — acme notification delivered after=never (bound 20 s) behind 22 globex rows to a timing-out destination; slow destination requests=11; acme outbox attempts|delivered=0|f
ABSENT: D21 / D22 / D23 / D24 (unchanged early-fixed controls)
DEFECTS REPRODUCED: 4 present
```

At lesson-07 the slow destination saw 11 requests in 20 s, one every two seconds, which is
the serial pass. At lesson-08 it saw all 22 at once, and the healthy partner's row was
delivered on the next tick.

## Dwell and the departure from the proposal

Dwell is paired lesson-07/current evidence only, as in E07; no full historical sweep was
run. U36–U38 and A5 are tested at those two points, without dating their introduction.
The proposal's story, that D21 sat open for seven experiments because no probe had aimed
across the tenant boundary, is counterfactual: E01 asked that question and repaired it.
E08 instead shows a tenant boundary that holds at the API while one tenant could still
deny another service through a shared background resource (U37). Its trust boundary is
the outbox, not a route.

## What remains open

- A freeze that starts after authentication still holds that request; only `/health` and
  authentication are bounded. A per-request database deadline needs a driver that honours
  cancellation against an unresponsive server, or a bound on every store call. The loop
  also stalls for the freeze and resumes afterwards (not separately measured).
- The explorer's forward questions are not part of this fix: Q-E08-1 (any key holder in a
  tenant can lease as any worker id), H-E08-22 (redirects), H-E08-23 (one schedule that
  cannot advance stops later schedules in `Fire`), H-E08-24 (pool exhaustion during a
  freeze), T-E08-1…3 (fault hook, shipper lag metric, partner signature/redirect modes).
  `/health` still cannot report a shipper backlog (T-E08-2).

## Try changing and read more

In a disposable copy, remove `ORDER BY next_at` but keep the goroutines and rerun U37 with
40 slow rows instead of 22. Then do the reverse: keep the order and deliver serially. Each
half fixes a different count. Replace `bounded` in `/health` with the original
`PingContext` and watch U38 reproduce, even though the context deadline is still in the
code. Change A5's rule to "run at the shifted time" (02:30 → 03:30 EDT) and decide which
oracle you would defend to someone whose backup runs at 02:30.

Read Chapter 9.1–9.4 through the approved-plan references above, the
[ecosystem model](../exploration/models/ecosystem.md), the
[notes](../exploration/experiments/08-surround-it/notes.md) and the
[reconciliation](../register/reconciliations/08.md). For running these what-ifs under
load, see Release It!'s `design-chaos-experiment` skill once that arc lands.

## Deviations

- **Fixer seat.** The plan names the seeder's family (Codex) as fixer. This chunk's brief
  assigned the whole experiment, fix included, to Claude Opus 5.5, as in E07's takeover.
  The explorer was a separate, fresh blind Opus process, and the fix was written after its
  record was committed. Independent review should come from a non-Claude family.
- **Runtime.** Verify takes about two and a half minutes, not "well under a minute": real outages,
  a 12-second freeze, clock steps and a 20-second delivery window dominate.
- **Instrument revision.** The first current/REF pair (`08-opus-*-attempt1.log`) passed
  with the same verdicts. The U37 message printed `after=Nones`; the message was fixed,
  hashes were re-frozen and both directions were rerun. No oracle or bound changed.
- **Instrument interaction and suite timeout.** The second pair passed, but `make checks`
  then failed `TestL01U03` and hit Go's default 10-minute test timeout
  (`08-opus-*-attempt2.log`). U37 had removed its slow partner and left 22 undeliverable
  rows; U03's control waits for the whole outbox to drain. U37 now switches its partner
  to `ok` and waits for its own rows to deliver before removing it. The suite (E07 alone
  took 514 s) now exceeds ten minutes, so `make checks` passes `-timeout 30m`. Neither
  change touches an oracle; hashes were re-frozen and the whole sequence was rerun.
- **The D23 control.** The plan's verify line ("cron 0 2 * * * fires at 02:00 EST and
  02:00 EDT across the transition") cannot hold literally: 02:00 does not exist on
  2027-03-14. A5's decision runs it at 03:00 EDT that day and at 02:00 on every other day.
