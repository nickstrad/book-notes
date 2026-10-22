# Release It! (2nd ed.) — one flawed SQLite platform, improved lesson by lesson

> **Built.** `release-it/v0-flawed`, twelve lesson tags and `release-it/v1-hardened` are on `main`.
> Start with [`lessons/00-the-flawed-start.md`](lessons/00-the-flawed-start.md); the index with every
> lesson's before and after is [`lessons/README.md`](lessons/README.md). Plan: [`PLAN.md`](PLAN.md).

*Release It! Design and Deploy Production-Ready Software*, 2nd edition, Michael T. Nygard (2018).

**The argument in one paragraph.** Software is done when it survives production, not when it passes
QA. Systems fail through their integration points and their coupling: a small fault becomes a crack
and propagates along whatever path the architecture leaves open. Learn the recurring ways cracks
start (stability antipatterns), build the structures that stop them spreading (stability patterns),
design processes so they can be seen and run (transparency, demand control), deploy without
downtime (phased changes, tolerant versions), and keep proving it with deliberate chaos.

## The project: `litefleet`

A baby multitenant SQLite database platform. Tenants authenticate with an API key, create databases
and run SQL against them. Each database is one SQLite file in WAL mode on one of two **node agents**.
Three stateless **gateways** behind HAProxy authenticate, ask the **control plane** (a Postgres
registry) which node owns the database, forward the statement and stream the rows back. Nodes back
up to a **blob store** and emit usage events to a **metering** partner; an ops dashboard polls a
fleet-wide usage report. Everything runs in Docker Compose with Toxiproxy between the tiers,
Prometheus scraping every service, an open-loop load generator, and DuckDB for the run reports and
for auditing tenant files. It is the control-plane / data-plane shape of the ts-db planes design at
baby size; what the baby leaves out (moves, fencing, replication) is named in each lesson's note.

`v0` runs and passes its tests, and it ships with eighteen deliberate flaws of the kind real
codebases have: no timeouts, a transaction not rolled back on an error path, retries on writes, no
breaker around the control plane, no per-tenant limits, a backup that holds the writer lock during
upload, no row cap, an unbounded usage queue, WALs that never checkpoint, strict JSON readers,
migrations at startup. Each lesson is one patch commit that fixes one flaw or one tight cluster, with
a note and a verify script that reproduces the flaw on the previous tag.

## Run it

```
cd books/release-it/project
make up && make seed                                   # the first seed takes ~1 min; resets reuse a snapshot
make verify LESSON=00                                  # v0 serves gentle load, flaws and all
git checkout release-it/lesson-05
make verify LESSON=05 REF=release-it/lesson-04         # FLAW REPRODUCED (builds the parent's images)
make verify LESSON=05                                  # IMPROVED
make down
```

Each verify runs in 1–3 minutes with warm images. It resets the stack, runs one scenario with its
fault timeline through Toxiproxy or Docker, and prints a table and a verdict line with explicit
thresholds. Exit status 0 means the lesson's thresholds hold.

## Lessons

Before-and-after headlines for every lesson are in [`lessons/README.md`](lessons/README.md).

| # | Tag | Lesson | Flaws | Book |
|---|---|---|---|---|
| 00 | `release-it/v0-flawed` | [The flawed start: tour of v0 and the register](lessons/00-the-flawed-start.md) | — | Ch. 1–3 |
| 01 | `release-it/lesson-01` | [Transparency: health, metrics, structured logs](lessons/01-transparency.md) | F01 | Ch. 8, 10, 6 |
| 02 | `release-it/lesson-02` | [Timeouts on every integration point and every statement](lessons/02-timeouts.md) | F02–F04 | Ch. 4, 5 |
| 03 | `release-it/lesson-03` | [Blocked threads: roll back on every path, bound the handle pools, fail fast](lessons/03-blocked-threads.md) | F05–F07 | Ch. 2, 4, 5 |
| 04 | `release-it/lesson-04` | [Retries: reads only, jittered backoff, a budget, `busy_timeout`](lessons/04-retries.md) | F08 | Ch. 4, 5 |
| 05 | `release-it/lesson-05` | [Circuit breaker around control, TTL route cache as the degraded mode](lessons/05-circuit-breaker.md) | F09 | Ch. 5, 4 |
| 06 | `release-it/lesson-06` | [Bulkheads: per-tenant admission, separate pools, backups off the writer lock](lessons/06-bulkheads.md) | F10–F11 | Ch. 5, 4 |
| 07 | `release-it/lesson-07` | [Shed load at the gateway, the node and the balancer](lessons/07-shed-load.md) | F12 | Ch. 4, 5, 9 |
| 08 | `release-it/lesson-08` | [Unbounded result sets: cap and stream rows, aggregate usage in SQL](lessons/08-unbounded-results.md) | F13 | Ch. 4, 5 |
| 09 | `release-it/lesson-09` | [Back pressure: bounded usage queue with a SQLite outbox](lessons/09-back-pressure.md) | F14 | Ch. 5, 4 |
| 10 | `release-it/lesson-10` | [Steady state: checkpoint WALs, purge the ledger, evict idle handles](lessons/10-steady-state.md) | F15 | Ch. 5 |
| 11 | `release-it/lesson-11` | [Test harness: chaos tests against Toxiproxy and a misbehaving blob store](lessons/11-test-harness.md) | F16 | Ch. 5, 15, 17 |
| 12 | `release-it/lesson-12` | [Handling versions: expand/contract, node drain, tolerant readers](lessons/12-handling-versions.md) | F17–F18 | Ch. 12, 13, 14 |
| — | `release-it/v1-hardened` | Final state: lessons index and this page | | |

The register (what each flaw is, where it lives, how it shows) is `PLAN.md` section 3, and with
`file:line` pointers into v0 in lesson 00. Each note ends with its deviations from the plan.

## Skills

Not written yet; they are the next phase. See [`skills/README.md`](skills/README.md). Planned: `test-integration-points`,
`hunt-stability-antipatterns`, `suggest-stability-patterns`, `budget-timeouts-and-retries`,
`review-capacity-headroom`, `review-operational-transparency`, `plan-safe-release`,
`design-chaos-experiment`. Each cites the project as its worked example; the register is the answer
key for the hunt.

## Earlier plans

`PLAN-v2-checkout.md` (same structure, a web-shop checkout service as the vehicle) and
`PLAN-v1-per-topic-projects.md` (twelve standalone mini-programs; its `topics/` stubs were removed
when v0 was built).
