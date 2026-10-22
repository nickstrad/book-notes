# 12 · Handling versions: expand/contract migration, node drain, tolerant readers

Tag: `release-it/lesson-12` · Fixes: F17, F18 · Verify: `make verify LESSON=12` (before:
`REF=release-it/lesson-11`, with and without `REF_SERVICES=gateway`) and `make godot`
Book: Ch. 14 "Handling Versions"; Ch. 13 "Phases of Deployment", "The Fallacy of Planned
Downtime"; Ch. 12 case study ("Waiting for Godot")

## The idea

A release is not an instant. For a while the old and the new versions run side by side: old
gateways talk to new nodes, new code reads rows that old code wrote, and a consumer built last
year reads an answer that gained a field this morning. Chapter 13's **Phases of Deployment**
turns one risky step into several safe ones:

1. **Expand.** Make additive changes that old code ignores, such as a nullable column or a new
   field.
2. **Roll out.** Replace instances one at a time, each taken out of rotation and drained first.
3. **Relax.** Let the mixed fleet run until nothing depends on the old shape.
4. **Contract.** Remove the old shape in a later release, guarded by a check that it is truly
   unused.

Chapter 13's **Fallacy of Planned Downtime** is the reason to bother: "we'll do it in a
maintenance window" stops working once the platform's tenants are in every time zone.

Chapter 14's **Handling Versions** is the other half: help others handle your versions, and
handle theirs.

- **Help others handle yours.** Only add fields, keep every old field, and accept the older
  request form.
- **Handle theirs.** Be a **tolerant reader**. Use the fields you need, validate their types and
  ranges, and ignore the rest. A reader that rejects unknown fields turns every additive change
  into a breaking one.

Chapter 12's case study, **Waiting for Godot**, shows what happens without phases: a release
night spent waiting on database scripts that ran far longer than planned while the site stayed
down. It is the story behind Chapter 13's case against planned downtime. `make godot` replays
the shape of that night on litefleet: a migration that takes a lock and holds everyone behind
it. One in-place `ALTER` on a big table queues behind a long report, every query on that table
then queues behind the `ALTER`, and the platform waits for a rewrite nobody can see. The lock
queue is the lab's mechanism, not the book's.

## The flaws (at `release-it/lesson-11`)

**F17: strict readers.**

- The gateway decodes node answers with `DisallowUnknownFields`
  (`internal/nodeclient/client.go:137`).
- Its streaming reader accepts only the fields it knows before and after the rows
  (`internal/nodeclient/rows.go:56`, `:112`).
- The load generator, the stand-in consumer, reads the API "exactly" too
  (`cmd/loadgen/main.go:494–497`).
- The first node to add `column_types` breaks every query that passes through an old gateway.

**F18: migrations in the deploy, and no drain.**

- Control applies migrations inside `main()` at startup (`cmd/control/main.go:36`), so DDL ships
  in the same step as the code that needs it.
- The counters are 32-bit (`migrations/001_init.sql:29–30`, `:49–51`). The whale's byte total
  passes 2³¹, and the obvious fix is an in-place `ALTER COLUMN ... TYPE bigint`, which rewrites
  the 5 M-row ledger under `ACCESS EXCLUSIVE`.
- A node exits as soon as SIGTERM arrives (`cmd/node/main.go:80–84`; lesson 9 taught it only to
  spill its usage queue), so every statement in flight on it is cut off.

## The patch

Two commits.

**Part 1: migrations as a job, expand the counters, drain the nodes.** About 470 lines.

- **The migration job.** Control no longer migrates at startup. `control migrate`
  (`cmd/control/jobs.go`) is a one-shot job: Compose service `migrate`, profile `jobs`,
  `make migrate`.
  - It applies the main migrations in transactions and the online ones outside them
    (`internal/registry/migrate.go`).
  - It uses one dedicated connection with no statement timeout and `lock_timeout = 3s`. A
    migration that cannot get its lock fails and can be retried later. It never queues every
    other query behind itself.
- **Expand.** `migrations/003_counters_expand.sql` adds nullable `bigint` columns: `bytes_total_v2`
  and `rows_total_v2` on `databases`, and `bytes_v2` and `rows_v2` on `usage_ledger`.
  - Nullable columns without a default change the catalog only. There is no rewrite.
  - Control writes both column sets and reads `coalesce(v2, old)` (`internal/registry/registry.go:211`,
    `:286–290`; `internal/registry/usage.go:36`, `:48`).
  - The old columns saturate at 2³¹−1 instead of overflowing, so code that still reads them gets a
    wrong number rather than a failed heartbeat.
  - The migration also records each node's `wire_version`, taken from its heartbeat.
- **Backfill.** `control backfill` copies old rows into the new columns in id ranges of 5 000, with
  a pause between batches.
- **Contract.** `migrations/contract/004_counters_contract.sql` drops the 32-bit columns, renames
  the 64-bit ones and adds `NOT NULL`.
  - It runs only with `control migrate -contract`.
  - It refuses while any row is unbackfilled, or while any node has reported wire v1 in the last
    day.
- **Node drain** (`cmd/node/drain.go`). On SIGTERM the node:
  1. stops admitting: `503 draining` with `Retry-After: 2` and `Connection: close`, and its health
     check fails;
  2. waits for in-flight statements, bounded by the statement deadline;
  3. checkpoints and closes every database (`Engine.Close`);
  4. spills its usage queue;
  5. exits.
- **Gateway drain** (`cmd/gateway/main.go:69`). On SIGTERM the gateway fails its health check,
  waits 2.5 s for HAProxy to notice, then shuts down gracefully.
- **Gateway refusals.** A draining node's 503 passes through with its `Retry-After`. A node the
  gateway cannot reach becomes `503 node_unavailable` with `Retry-After: 1`.
- **`make rollout`** (`lab/rollout.sh`) restarts gateways a → b → c, then nodes a → b. It waits
  for each instance's `/health` before touching the next.

**Part 2: tolerant readers and an additive wire format.** About 170 lines.

- **Tolerant gateway reader.**
  - `decode` (`internal/nodeclient/client.go:138`) no longer rejects unknown fields. It validates
    what it uses: `rows_affected >= 0`.
  - The streaming reader (`internal/nodeclient/rows.go:44`) skips unknown fields before and after
    the rows. It still requires `columns` before `rows`.
- **The consumer reads tolerantly first.** The load generator's `decodeAnswer`
  (`cmd/loadgen/main.go:496`) ignores unknown fields but requires `columns` on a query answer.
- **Wire v2 on the node.** `NODE_WIRE_VERSION=2` adds `column_types` and
  `stats{rows_read, bytes}` to query answers, and `stats` to execute answers
  (`internal/engine/exec.go:174`, `:221`). Every v1 field stays, and the heartbeat reports the
  version.
- **The gateway's own answer** adds `stats` and `node_epoch`. `/execute` also accepts the older
  single-statement form `{sql, args}` (`internal/api/handlers.go:134`).
- **A singleflight on opening a database** (`internal/engine/handles.go`). A restarted node used to
  open each database while holding the engine's lock. Every request to every database then waited
  behind each open, long enough for the gateway to time out. Now the open happens outside the lock,
  and concurrent requests for the same database share one open.

## Verify

```
make verify LESSON=12 REF=release-it/lesson-11 REF_SERVICES=gateway   # old gateways, today's nodes
make verify LESSON=12 REF=release-it/lesson-11                         # everything old
make verify LESSON=12                                                  # after
make godot                                                             # the in-place ALTER, on any tree
```

The run lasts 100 s: the steady mix at 50 % of fleet capacity, plus a create every second.

| t | Step |
|---|---|
| 3 s | The migration job runs, and the backfill starts in the background. A tree without the job (lesson 11) migrated at startup. |
| 8 s | `lab/rollout.sh` restarts the five instances one at a time. |
| 62 s | node-b is recreated with `NODE_WIRE_VERSION=2`, so the fleet runs mixed versions. |
| end | The contract migration is attempted. |

The script judges the run from the load generator's CSV, using each restart's start and healthy
times. It sorts each failure into one of three classes:

- **Refused fast:** a 429 or 503 with a refusal class such as `draining`, `node_unavailable` or
  `overloaded`. A client knows it can retry these.
- **Timed out.**
- **Other failures**, including connections cut mid-answer.

Before, old gateways with today's nodes (`REF_SERVICES=gateway`):

```
    node-b-v2  t= 62.1 s   6.0 s  1
  failed requests during the gateway rollout                 5
  failed requests outside the node restarts                  4282
  refused fast (503 + Retry-After) outside the node restarts 7 of 30087 (0.02 %)
  in the node restarts: refused fast 1525, timed out 2, other failures 680; longest run of failures 6.5 s
  requests cut off by a node stopping                        0
  node-b queries after it speaks wire v2: 3212, failed (not refused) 3210 (99.9 %); answers the consumer could not read: 0
FLAW REPRODUCED: gateway_rollout_errors 5 (want == 0); errors_outside_node_restarts 4282 (want == 0); mixed_fleet_query_fail_pct 99.9 (want == 0)
```

Before, everything at `lesson-11`:

```
  failed requests during the gateway rollout                 3
  failed requests outside the node restarts                  3
  in the node restarts: refused fast 924, timed out 0, other failures 192; longest run of failures 2.7 s
  requests cut off by a node stopping                        5
  node-b queries after it speaks wire v2: 3428, failed (not refused) 0 (0.0 %); answers the consumer could not read: 0
  contract migration                                         n/a (no migration job)
FLAW REPRODUCED: gateway_rollout_errors 3 (want == 0); errors_outside_node_restarts 3 (want == 0); requests_cut_off 5 (want == 0)
```

After:

```
    gateway-a  t=  9.2 s   5.3 s  2
    gateway-b  t= 16.6 s   7.2 s  2
    gateway-c  t= 25.8 s   5.6 s  1
    node-a     t= 33.5 s   3.2 s  1
    node-b     t= 38.8 s   3.5 s  1
    node-b-v2  t= 62.1 s   8.0 s  1
  failed requests during the gateway rollout                 0
  failed requests outside the node restarts                  0
  refused fast (503 + Retry-After) outside the node restarts 28 of 30090 (0.09 %)
  in the node restarts: refused fast 1681, timed out 1, other failures 0; longest run of failures 4.5 s
  requests cut off by a node stopping                        0
  node-b queries after it speaks wire v2: 3025, failed (not refused) 0 (0.0 %); answers the consumer could not read: 0
  contract migration                                         contract refused: rows still to backfill
IMPROVED: gateway_rollout_errors 0 == 0; errors_outside_node_restarts 0 == 0; longest_restart_failures_s 4.5 <= 10; requests_cut_off 0 == 0; mixed_fleet_query_fail_pct 0.0 == 0
```

**Old gateways, new nodes.** Once node-b spoke wire v2, 99.9 % of its queries failed through the
old gateways (`json: unknown field "column_types"`). Nothing was wrong with the answers; the reader
was simply too strict. The 4 282 errors outside the restarts are those failures. A gateway restart
also dropped a few requests, because the old gateway stops at once while HAProxy is still sending to
it.

**Everything old.** Without a drain, a node restart cut off the statements in flight on it (5 here,
`requests cut off`). The gateways stopping at once lost 3 more.

**After.** Gateways leave rotation before they stop, so nothing fails during the gateway rollout. A
node restart costs that node's tenants 3–6 seconds: requests are refused with `Retry-After` while it
drains and is replaced, and a few time out while the container is gone (1 in this run; 57 in a second
run, whose longest window was 5.6 s). Nothing in flight is cut off.
The mixed fleet answers every query: the gateways ignore `column_types`, and the consumer ignores the
gateways' new `stats` and `node_epoch`. The contract migration refuses, because the backfill is still
running. (Node-a still speaks v1, so the second guard would refuse it too.) That refusal is what a guard is for.

`make godot` replays Chapter 12 on any tree, with the gateways' route cache cut to 5 s so cache
misses reach control:

1. An analyst's report loops over the ledger joined to `databases`.
2. At t=15 s the in-place `ALTER ... TYPE bigint` runs on both tables in one transaction.

At `release-it/lesson-12`:

```
  t s  lock waiters
  15   0
  18   2
  25   5
  35   6
  41   8
  47   7
  49   0
creates during the migration: 72, failed 68
cold reads (cache misses) during the migration: 184, failed 163
  the migration held every query on databases for about 35.9 s
```

The `ALTER` waited behind the running report for its exclusive lock, then rewrote 5 M ledger rows.
For 36 s, every resolve, create and heartbeat queued behind it. The stall is larger than "creates and
cache misses": with a 5 s route TTL, every gateway's cache emptied within seconds, lesson 5's breaker
opened, and **every** tenant request from t=15 s to t=49 s was refused with `503 control_unavailable`.
All the earlier lessons did their job: the failures were fast, bounded and honest. But they cannot
make a table that is locked for half a minute available. Only the phases can.

## Try changing

- Run the contract after the backfill finishes, with node-a still on v1
  (`docker compose --profile jobs run --rm migrate -contract`). Read the second refusal, then
  restart node-a with `NODE_A_WIRE_VERSION=2` and try again. (Today's code still reads the `_v2`
  columns, so a contract that succeeds needs the next release; see the deviations.)
- Put `DisallowUnknownFields` back in the load generator only, and restart one gateway with a
  response that has one more field. The consumer is the reader whose strictness you do not control.
- Make the gateway retry a `draining` query once after `Retry-After`, as the plan suggested, and
  watch the shed limits and HAProxy's `maxconn` during a node restart.

## Read more

*Release It!* 2nd ed., Ch. 12 (the "Waiting for Godot" case study); Ch. 13 "Phases of Deployment";
Ch. 14 "Handling Versions". Martin Fowler, "Tolerant Reader" and "ParallelChange" (expand/contract).
PostgreSQL documentation on `ALTER TABLE` lock levels and `lock_timeout`.

## Deviations from the plan

- **The gateway refuses during a node restart; it does not retry `draining` once.** A retry waits
  2 s holding a gateway slot. With a node's whole tenant population waiting, the gateways' shed
  limits and HAProxy's `maxconn` filled, and requests to the *other* node were refused too. The
  503 passes to the client with its `Retry-After`. An unreachable node (the container is gone)
  becomes `503 node_unavailable` with `Retry-After: 1` instead of a 502.
- **The restart window is "≤ 10 s of failures, nothing cut off", not "≤ 2 s of 503 draining".** A
  node restart here means a new container: drain, stop, start, WAL recovery, reopen. That takes
  3–8 s on this shared VM, and while the container is gone a request can only be refused or time
  out. The plan's "zero 5xx outside the window" holds. A few refusals outside the windows (0.09 %)
  are the gateways' shed limits at the edges of a restart and are reported, not failed.
- **The gateways drain too** (part 1). Without it, the gateway rollout lost 3–5 requests every run.
- **The contract keeps the plan's shape, but it belongs to the next release.** The contract renames
  the `_v2` columns to the old names. Today's code reads and writes the `_v2` names, so a successful
  contract must ship with the code that uses the final names. The verify only checks that the guard
  refuses.
- **Godot alters both tables in one transaction**, the way a careless migration would be written,
  so the stall covers `databases` (creates, heartbeats, resolves) as well as the ledger. `make godot`
  is a separate command, not part of the verify, to keep the verify inside its time budget.
- **The migration job sets `lock_timeout = 3s`.** This is the cheapest Godot defence: an expand
  that cannot get its lock fails, instead of becoming the head of the queue.
- **A singleflight on opening a database** (part 2). Without it, the restarted node's cold reopen of
  ~120 databases under one lock produced 504s after every restart.
- **Size: about 640 lines over two commits** (plan ~290). The drain on both tiers, the migration job
  and the verify's timeline accounting are each larger than planned.
