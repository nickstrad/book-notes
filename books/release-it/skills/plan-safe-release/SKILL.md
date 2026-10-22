---
name: plan-safe-release
description: Plan a schema, API or wire-format change so it deploys without downtime — classify it as additive or breaking, then for breaking changes produce an expand → roll out (drain, one at a time) → relax → contract plan with tolerant-reader checks on every consumer, a lock-safe database sequence (add, dual-write, backfill, switch, guarded drop), rollout order, and a rollback for each phase. Use for any schema or API change, "can we deploy this without downtime", or writing a rollout plan.
---

# plan-safe-release

A release is not an instant. For minutes or days, old and new code run side by side, old code reads
rows new code wrote, and consumers built last year read answers that gained a field this morning.
This skill turns one risky step into several safe ones, each of which can be stopped or undone.

## When to use

- A database migration, especially on a large or hot table.
- A change to an API, an event, a message, or an internal wire format.
- "Can we deploy this without a maintenance window?"
- Writing or reviewing a rollout plan or a deploy pipeline.

## Inputs

- The change: DDL, the API/schema diff, the code that reads and writes it.
- Every producer and consumer of the changed shape, including ones you do not own.
- Table sizes and the hottest queries on the affected tables; how migrations run today (at
  startup? as a job?); how instances are stopped (is there a drain?).
- The rollout mechanism: instance count, order, health gating.

## Procedure

1. **Classify the change.**
   - **Additive:** a new optional field, a nullable column with no default, a new endpoint, a new
     index built concurrently. Old code ignores it. Ship it alone.
   - **Breaking:** removing or renaming a field or column, changing a type, making something
     required, changing meaning. Old and new code cannot share it. Needs the phases below.
   - **Locking:** anything that rewrites a table or takes an exclusive lock for longer than your
     latency budget (`ALTER COLUMN ... TYPE`, adding a column with a volatile default on old
     engines, non-concurrent index builds). Treat as breaking even if the shape is compatible.

2. **Check every reader is tolerant.** For each consumer of the shape: does it ignore unknown
   fields (no `DisallowUnknownFields`, no strict schema rejection, no exact-field-set checks)? Does
   it validate only what it uses (types, ranges)? Does it accept the older request form? Strict
   readers turn an additive change into a breaking one. If a reader is strict, the **first**
   release makes it tolerant, and the change waits until that is everywhere.

3. **Expand.** Add the new shape next to the old one, additively:
   - Database: nullable new column/table; code writes both, reads new-or-old
     (`coalesce(new, old)`); the old field degrades safely (e.g. saturate instead of overflow).
   - API/wire: add new fields; keep every old one; version reported by each instance.
   - Run DDL as a **separate job**, not at application startup, with a **lock timeout** (seconds)
     so a migration that cannot get its lock fails and retries later, instead of queueing every
     other query behind it. Index builds `CONCURRENTLY`, outside a transaction.

4. **Roll out.** One instance at a time, gated on health:
   - Each instance **drains** on SIGTERM: fail its health check, stop admitting (`503` +
     `Retry-After`), finish in-flight work up to its deadline, flush state, exit. Wait for the
     balancer to notice before stopping.
   - Order: tolerant consumers before producers of the new shape; readers before writers.
   - Gate each step on the instance's health before touching the next; stop the rollout on errors.

5. **Backfill.** Copy old rows into the new shape in small batches (thousands of rows) with pauses,
   idempotent, resumable. Watch the critical path's p99 while it runs.

6. **Relax.** Let the mixed fleet run until nothing depends on the old shape. Verify with data, not
   time: no row unbackfilled, no instance reporting the old version within the window you care
   about, no reads of the old field in metrics/logs.

7. **Contract.** In a **later** release that uses only the final names: drop the old shape behind a
   **guard** that refuses while anything still needs it (unbackfilled rows, old versions seen
   recently). The guard is code, run by the migration, not a checklist item.

8. **Rollback per phase.** Write, for each phase, how to undo it and what is lost:
   expand → drop the new columns (nothing read them yet); rollout → roll instances back (old code
   ignores the new shape); backfill → stop it (resumable); contract → no rollback, which is why it
   is guarded and last.

9. **Rehearse the unsafe version.** If the obvious change is a locking one, run it once in a lab
   with a realistic long-running query in front of it, and measure the stall. That number is the
   argument for the phases.

## Output

A phase table:

```
| Phase | Change | Runs as | Lock / duration bound | Gate to next phase (check) | Rollback | Signal to watch |
```

plus the reader audit (`consumer | tolerant? | fix needed first`) and the rollout order.

## Source

*Release It!* 2nd ed.: Ch. 13 "Phases of Deployment", "The Fallacy of Planned Downtime"; Ch. 14
"Handling Versions" (help others handle your versions; handle others' versions); Ch. 12 case
study "Waiting for Godot" (the in-place change that stalled a site). Demonstrated in lesson 12
(`../../lessons/12-handling-versions.md`), with a first online migration in lesson 08.

## Worked example: litefleet lesson 12

**The change.** The whale's byte total passes 2³¹; the `int` counters on `databases` and the
5 M-row `usage_ledger` must become `bigint`. Separately, nodes want to add `column_types` and
`stats` to query answers (wire v2). Step 1: the obvious `ALTER COLUMN ... TYPE bigint` is
**locking** (full rewrite under `ACCESS EXCLUSIVE`); wire v2 is additive in shape but **breaking in
practice**, because step 2 finds strict readers: the gateway's node decoder
(`DisallowUnknownFields`, `internal/nodeclient/client.go`), its streaming row reader
(`internal/nodeclient/rows.go`), and the reference consumer (`cmd/loadgen`).

**Step 9 first** (`make godot`, `project/lab/godot.sh`): an analyst's report loops over the ledger
joined to `databases`; at t=15 s the in-place `ALTER` runs on both tables. Lock waiters climb to 8;
**every query on `databases` is held for ~35.9 s**; 68 of 72 creates and 163 of 184 cache-miss reads
fail, and with a 5 s route TTL the gateways' breaker opens and every tenant request is refused for
34 s. The earlier lessons made the failure fast and honest; only the phases make it not happen.

**The plan as built:**

| Phase | Change | Runs as | Gate | Rollback |
|---|---|---|---|---|
| reader fix | gateway and consumer ignore unknown fields, validate what they use | release before wire v2 | every gateway on it | redeploy old gateway |
| expand | `migrations/003_counters_expand.sql`: nullable `*_v2 bigint` columns (catalog only), `nodes.wire_version`; control writes both, reads `coalesce(v2, old)`, old columns saturate at 2³¹−1 | `control migrate` job (`cmd/control/jobs.go`), one connection, `lock_timeout = 3s` (`internal/registry/migrate.go`) | migration succeeded | drop the `_v2` columns |
| rollout | gateways a → b → c, then nodes a → b, each drained (`cmd/node/drain.go`, `cmd/gateway/main.go`) | `make rollout` (`lab/rollout.sh`), waits for `/health` | each instance healthy | roll back instances |
| backfill | old rows → `_v2` in id ranges of 5 000 with pauses | `control backfill` | no NULL `_v2` | stop; resumable |
| wire v2 | `NODE_WIRE_VERSION=2` adds `column_types`, `stats` | node restart | heartbeat reports 2 | set back to 1 |
| contract | `migrations/contract/004_counters_contract.sql`: drop old, rename, `NOT NULL` | `control migrate -contract`, **next release** | guard: refuses while any row is unbackfilled or any node reported wire v1 in the last day | none: guarded and last |

**Measured** (`make verify LESSON=12`, 100 s at 50 % load plus a create per second; migration at
t=3 s, rollout from t=8 s, node-b to wire v2 at t=62 s, contract attempted at the end):

| | old gateways + new nodes (`REF_SERVICES=gateway`) | all at `lesson-11` | `lesson-12` |
|---|---|---|---|
| node-b queries failed after it speaks wire v2 | **99.9 %** (3 210 of 3 212, `json: unknown field "column_types"`) | 0 % | **0 %** |
| failed requests during the gateway rollout | 5 | 3 | **0** |
| requests cut off by a node stopping | 0 | 5 | **0** |
| failures outside the node restarts | 4 282 | 3 | **0** |
| contract migration | — | n/a (no job) | **refused: rows still to backfill** |

A node restart still costs that node's tenants 3–6 s of fast `503 draining` / `node_unavailable`
refusals with `Retry-After`: a stateful tier has no instance to route to, so the drain makes the
gap honest, not zero. The contract's refusal is the expected end state of the run; a contract that
succeeds must ship with code that uses the final column names, which is the next release (recorded
in the lesson's deviations).
