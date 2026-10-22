---
name: probe-entities-and-boundaries
description: Probe what happens to dependents when an owner changes or disappears, follow one datum through every place it is stored or decoded, and ask "what if this neighbour misbehaves" at each boundary. Use for entities with owners, data that crosses clients and partners, trust boundaries, or "what if X goes away".
---

# Probe entities and boundaries

Two models drive this skill. The entity model asks what happens to each dependent when its owner changes, and whether a
value survives every path that carries it. The ecosystem model draws the system as one box among its neighbours and
asks, per boundary, what crosses it, where it is checked, and what happens when the neighbour misbehaves. In this repo's
arc you also follow [the protocol](../../exploration/PROTOCOL.md): predictions before probes, a timestamped log, and
every surprise ends in a report or a written reason.

## When to use

- A schema or API has owners and dependents (tenant → project → item; job → attempts, events, notifications).
- A value travels through more than one decoder: API, storage, a CLI, a worker, a webhook, an export.
- The service has neighbours: a database, workers, an outbound partner, a clock, another tenant.
- Someone asks "what if X goes away", "is tenant data isolated" or "what happens on a database restart".

Not for: lifecycle orders and status columns (use `probe-sequences-and-states`) or single-field input ranges (use
`sweep-variables`).

## Inputs

- Source and every migration or schema file, in order; the spec or API description.
- A disposable lab you may reset, with container controls (stop, start, pause, kill) and, for time-based systems, a
  controllable clock.
- Two actors (keys or tenants), and fixtures with at least one owner that has every kind of dependent.
- Read-only database access, and a partner stub whose mode you can set (ok, error, slow, timeout) and that keeps the raw
  request bytes.

## Procedure

1. **Read the constraints as they stand now.** Read every migration, not only the first. For each owner → dependent
   pair, record the database constraint (FK, `ON DELETE`, none) and the application enforcement (`file:line`)
   separately. Signals: `grep -n "REFERENCES\|FOREIGN KEY\|ON DELETE\|NOT VALID"` over migrations; delete handlers that
   lock, count dependents or return 409; pairs with neither are orphan candidates. (queuelet: the charter predicted
   orphans from migration 001, but 002 had added FKs and a 409 guard; orphans appeared only for job → attempt, the pair
   never linked — [E06 notes](../../exploration/experiments/06-relate-it/notes.md).)
2. **Draw the entity diagram and an empty CRUD grid** (entity × create/read/update/delete) before probing. Mark each
   cell blank, inferred (I) or observed. Only an observation fills a cell.
3. **Remove or change each owner, then read each dependent through every path**: the API, read-only SQL, the event log,
   any outbound copy. Write the expected outcome first: refused, cascaded, retained by design or orphaned. Signals:
   `SELECT count(*) FROM child c LEFT JOIN owner o ON o.id = c.owner_id WHERE o.id IS NULL`; an API 404 next to a
   surviving storage row; a delete that answers 204 for a missing or foreign id (no `RowsAffected` check). Repeat for
   every removal path: explicit delete, purge, retention. (queuelet:
   [U27](../../exploration/defects/U27-schedule-delete-missing-204.md),
   [U33](../../exploration/defects/U33-attempts-outlive-their-job.md).)
4. **Race the owner's removal against dependent creation.** Run N parallel creates against one delete, count status
   codes per round, then count orphans in storage. If one side always wins, vary one thing: send the delete first and
   stagger the creates by milliseconds. Signal: a loser answered 400 or 500 with a constraint name instead of 404. A
   sampled non-hit is not proof of absence; record the tries and leave it open. (queuelet: 30 rounds were a dead end,
   the staggered run hit it — [p9b-race.sh](../../exploration/experiments/06-relate-it/tools/p9b-race.sh),
   [U32](../../exploration/defects/U32-enqueue-racing-queue-delete-400-fk-text.md).)
5. **Compare every creation path for one entity.** If the API, a batch, a manual trigger and a timer can all create it,
   list the validation each applies (`grep` for the shared validator and the paths that bypass it), then push one
   invalid value through all of them. (queuelet: manual run rejected a non-object payload that the cron fire stored —
   [U28](../../exploration/defects/U28-schedule-payload-not-validated.md).)
6. **Follow one recognizable datum through every hop**: input → storage → read API → worker or CLI decode → result →
   outbound bytes → partner. Use values that expose lossy decoders: an integer above 2^53 (`9007199254740993`), a long
   decimal, `1.10`, nested values, `<&>`, U+2028, a NUL character. Record per hop: exact, value-equal (key order,
   whitespace, escaping, `1e2` → `100`; oracle RFC 8259) or changed. Signals: `json.Unmarshal` without `UseNumber`,
   `float64`, `map[string]any`, `parseFloat`. Check CLI input against storage and CLI display against an API-written
   value: one CLI round trip hides symmetric rounding. Read instruments' raw bytes, not their parsed fields. (queuelet:
   the API kept every digit, its own worker and CLI did not —
   [U29](../../exploration/defects/U29-sample-worker-echo-loses-precision.md),
   [U30](../../exploration/defects/U30-qctl-rounds-payload-numbers.md),
   [U31](../../exploration/defects/U31-nul-in-result-or-error-409-conflict.md),
   [knowledge](../../docs/knowledge/06-data-boundaries.md).)
7. **Draw the context diagram.** One box for the system; around it every neighbour: datastore, workers, partners,
   operators, other tenants, clock and time-zone data, the process supervisor. Each line is a boundary: record what
   crosses it and where the code checks it (`file:line`), or "nowhere". Signals: one background resource (outbound
   queue, pool, loop) shared by every tenant; `log.Fatal` or `panic` on a startup dependency; no `restart:` policy; a
   deadline handed to a driver call.
8. **Write one what-if per neighbour state, not per neighbour**, each with a prediction and refutation before it runs. A
   database has at least four states: restarted under a running API, absent when the API starts, frozen (connected but
   silent), crashed. A partner has error, slow, timeout and unreachable; repeat slow with more rows than one pass
   handles and another tenant's row behind them. The clock has the DST gap and overlap dates of each zone in use. Record
   baseline, action, observation and recovery (automatic, manual, never). (queuelet: the charter sentence held for what
   it named and failed at three states it did not — [E08
   revealed](../../exploration/experiments/08-surround-it/revealed.md).)
9. **Inject faults from outside, the way the neighbour fails.** `docker compose pause` gives a frozen database (sockets
   stay open); `stop` gives an absent one; probe both. `docker kill` counts as an operator stop and skips the restart
   policy; kill the process from the host to simulate a crash. Time each request during the fault: a deadline in the
   code is not a deadline in effect. Afterwards restore the neighbour and drain what you queued. (queuelet:
   [p8_pg_pause.sh](../../exploration/experiments/08-surround-it/tools/p8_pg_pause.sh),
   [U38](../../exploration/defects/U38-frozen-database-hangs-requests.md),
   [knowledge](../../docs/knowledge/08-neighbour-faults.md).)
10. **Probe across the tenant boundary.** Use the other tenant's key on this tenant's ids on every route, and compare
    with a nonexistent id. Then ask which shared resource one tenant could fill, by configuration alone, so that another
    waits. (queuelet: routes held, but 20 slow rows from one tenant starved every later notification —
    [U37](../../exploration/defects/U37-webhook-shipper-starvation.md).)
11. **Close.** Fill the CRUD grid and boundary table from observations only. Check each `file:line` against the source
    before writing reports. Characterize each surprise with a sibling check; file a spec silence (what a schedule in a
    DST gap should do) as a question with candidate oracles. For a what-if worth running under load, hand its baseline,
    fault and recovery criteria to Release It!'s `design-chaos-experiment` skill (it reaches this branch with the
    merge): this skill produces the list and the evidence; that one runs one fault at load.

## Output

- `models/entities.md`: entity diagram; constraint table (dependency | DB constraint | application enforcement |
  observed when the owner goes); CRUD grid with observed, (I) inferred, (r) regression and blank cells; data-journey
  table (hop | exact? | transformation and oracle).
- `models/ecosystem.md`: context diagram; boundary table (boundary | what crosses | checked at | what-if | observed |
  recovery).
- Hypotheses with prediction and refutation, defect reports and spec questions, as the protocol specifies.

## Pitfalls the experiments exposed

- **Premise from an old schema.** E06's charter read migration 001 and predicted orphans; 002 had added FKs. Step 1
  reads all migrations and separates database from application enforcement ([E06
  notes](../../exploration/experiments/06-relate-it/notes.md)).
- **Exact storage, lossy clients.** Every boundary the API owned was exact; the losses were in queuelet's own worker and
  CLI, and the partner stub's parsed `body` field rounds too. Step 6 follows the datum into clients and reads raw bytes
  ([lesson 06](../../lessons/06-relate-it.md)).
- **API 404 taken as cleanup.** A job gone from the API still had attempt rows, error text included, in storage. Step 3
  reads each dependent in storage for every removal path
  ([U33](../../exploration/defects/U33-attempts-outlive-their-job.md)).
- **Race non-hits.** The first race design never hit the window, and the batch sibling was not hit in 320 tries (the
  fixer later hit it). Step 4 varies the ordering and keeps non-hits open ([reconciliation
  06](../../register/reconciliations/06.md)).
- **Deadline in code, none in effect.** From a one-second `PingContext`, the explorer predicted a 503 within a second;
  `/health` hung for the 11.8-second freeze and answered 200. Step 9 times requests under a frozen neighbour ([E08
  notes](../../exploration/experiments/08-surround-it/notes.md)).
- **Charter named only the obvious states.** "The loop survives a database restart" held; absent-at-start, frozen and
  the shared shipper failed. Step 8 enumerates neighbour states ([lesson 08](../../lessons/08-surround-it.md)).

## Source

Elisabeth Hendrickson, *Explore It!* (Pragmatic Bookshelf, 2013). Chapter 7 "Explore Entities and Their Relationships":
7.1 "Recognizing Entities, Attributes, and Dependencies", 7.2 "CRUD", 7.3 "Following the Data". Chapter 9 "Explore the
Ecosystem": 9.1 "Diagramming the Ecosystem", 9.2 "Trust Boundaries", 9.3 "What If?", 9.4 "Follow the Data". Section
numbers and titles were checked against the publisher's table of contents; this procedure has not been validated against
the book's full text. Procedure wording, signals and record formats are this repo's, not quotations from the book.

Lessons that demonstrate it: [06 — Relate it](../../lessons/06-relate-it.md) and [08 — Surround
it](../../lessons/08-surround-it.md).

## Worked example

**E06, relate-it** ([charter](../../exploration/experiments/06-relate-it/charter.md),
[revealed](../../exploration/experiments/06-relate-it/revealed.md)), at `explore-it/lesson-05`. Charter: explore queues,
schedules, workers and jobs with an entity diagram and a CRUD grid to discover what happens to dependents when their
owner changes or disappears, and where a payload, result or error changes on its way. About 20 of 60 minutes. Score: 20
hypotheses stated, 18 confirmed, 2 refuted, 2 code-reading predictions contradicted, 7 defects, 30 model cells. Found
U27–U33 ([U27](../../exploration/defects/U27-schedule-delete-missing-204.md),
[U28](../../exploration/defects/U28-schedule-payload-not-validated.md),
[U29](../../exploration/defects/U29-sample-worker-echo-loses-precision.md),
[U30](../../exploration/defects/U30-qctl-rounds-payload-numbers.md),
[U31](../../exploration/defects/U31-nul-in-result-or-error-409-conflict.md),
[U32](../../exploration/defects/U32-enqueue-racing-queue-delete-400-fk-text.md),
[U33](../../exploration/defects/U33-attempts-outlive-their-job.md)). Wrong: the charter's orphan prediction (refuted for
queue → job and queue → schedule) and the batch-race prediction (not hit). Missed: dead-job retention was never probed;
the coordinator settled it as spec silence A4, not counted as a find ([lesson 06](../../lessons/06-relate-it.md)). Model
delta: [entities.md](../../exploration/models/entities.md) (constraint table, CRUD grid, data journey) and
[ecosystem.md](../../exploration/models/ecosystem.md).

**E08, surround-it** ([charter](../../exploration/experiments/08-surround-it/charter.md),
[revealed](../../exploration/experiments/08-surround-it/revealed.md)), at `explore-it/lesson-07`. Charter: explore
queuelet's neighbours (Postgres, workers, webhook partner, clock, second namespace) with a context diagram, partner
modes, container restarts and the test clock, to discover what crosses a boundary unchecked and what happens when a
neighbour misbehaves. About 22 of 60 minutes. Score: 21 stated, 20 confirmed (six are regression holds), 1 refuted, 1
code-reading prediction contradicted, 4 defects, 7 model cells. Found
[U36](../../exploration/defects/U36-api-start-without-database.md),
[U37](../../exploration/defects/U37-webhook-shipper-starvation.md),
[U38](../../exploration/defects/U38-frozen-database-hangs-requests.md) and spec decision
[A5](../../exploration/defects/A5-dst-gap-and-overlap.md). Wrong: H-E08-20, `/health` under a frozen database. Left
open: worker identity within a tenant (Q-E08-1), redirecting partners, pool exhaustion during a freeze. Model delta:
[ecosystem.md](../../exploration/models/ecosystem.md) context diagram and boundary table.

The step 9 measurement that refuted the code reading ([E08
notes](../../exploration/experiments/08-surround-it/notes.md), P8):

```text
19:34:38.5 pause postgres
19:34:50.8 unpause postgres
19:34:51.0 req health → 200 after 11.79s
19:34:51.0 req auth1 (acme GET /v1/queues) → 200 after 11.80s
```

**Second target (not yet run): litefleet**, the Release It! arc's multitenant SQLite platform. Entities: tenant → API
key; database → node placement → SQLite file → blob-store snapshots and metering events. Drop a database and check the
registry row, the node's file, its snapshots and its usage ledger separately; race `DELETE /v1/databases/{name}` against
`/execute`. Follow one large integer and one NUL through `/execute` → SQLite → `/query` → metering event. Context
diagram: balancer, gateways, control plane and its Postgres, two node agents, blob store, metering; for each, what-ifs
for absent, frozen, slow and crashed, plus one tenant's heavy query against another tenant on the same node. Running
those faults under load is `design-chaos-experiment`'s job.
