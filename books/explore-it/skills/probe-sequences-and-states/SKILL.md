---
name: probe-sequences-and-states
description: Explore a lifecycle (a status column, a lease, a retry or cancel protocol) by driving verbs in unlikely orders, running a seeded random walk that checks invariants after every step, and filling a state × event table with a second actor per cell; use when events can arrive in different orders or from different actors.
---

# Probe sequences and states

Use this skill to find out what a stateful object really does when verbs arrive in orders, from actors and at
times nobody planned for. It produces a sequence grid, a replayable walk, a state table with *should*,
*predicted* and *observed* per cell, and minimal reproductions. In this repo's arc you also follow
[the protocol](../../exploration/PROTOCOL.md): predictions before probes, the record in `notes.md`, no fixes
by the explorer.

## When to use

- A `state`/`status` column; leases, claims, heartbeats, retries, cancellation, pause/resume, purge, soft delete.
- Background actors that change state on a timer: sweeps, promotion, expiry, retention, cron.
- Questions like "what if the holder disappears?", "can a cancelled thing come back?", "what does purge do
  to work in flight?", or a bug that shows only after a particular history.

Not for: single-request input validation (`sweep-variables`); ownership without a lifecycle
(`probe-entities-and-boundaries`).

## Inputs

- The running system with a reset path, an API client, read-only access to the store, and at least two actor
  identities (two workers, two clients).
- A controllable clock if any transition is time-based; without one, expiry and promotion cells are blocked.
- The spec's lifecycle text, the source of every statement that writes the state, and any written rules.

## Procedure

1. **List states, events and writers.** States: `CHECK` constraints, enums,
   `grep -rnE "state *(=|IN|<>|NOT IN)|SET +state"` over source and migrations. Events: every route, CLI verb
   and background loop that runs one of those statements, with the actor that owns it. For each state, put
   every exit side by side with the columns it resets (holder, token, expiry, counters, history row, error
   text). Signal: two exits from one state that reset different columns. (queuelet: five exits from `leased`
   each made their own choice; three differences were U24–U26 — [lesson 05](../../lessons/05-sequence-it.md).)
2. **Write the invariants as queries before probing.** One SQL query per rule, returning violating rows:
   state in the allowed set; finished timestamp set exactly when terminal; counters within limits; holder
   fields set exactly when held; history numbered 1..k without gaps; stats equal a direct count; nothing due
   left after one tick. Track client-side facts the store cannot show: hand-outs since the last retry, attempt
   number strictly increasing.
3. **Drive directed sequences.** Choose pairs and triples for being unlikely: undo after done (cancel →
   retry); empty the container while a member is held (lease → purge → fail); remove the actor while it holds
   something (lease → delete worker), then repeat it; pause → clock jump → resume; a verb from the previous
   holder after hand-over. Write *I predict* and *I'd be wrong if* first. After each sequence run every
   invariant query and read the event history, not only the HTTP status.
4. **Run a seeded random walk.** Pick a noun and a verb, act, log `step|verb|args|status|body`, run the
   invariant queries. Keep the seed and the full log (ids and tokens depend on stack state, so the log is the
   replay). Print each violation once per (invariant, row). Count *successful* calls per verb per seed: a
   refused or skipped call is not coverage. If a success arrow has fewer than about five hits per seed, bias
   the walk toward rows the walker holds, never let it run out of actors, and rerun. (queuelet: a uniform
   walk made 0–4 acks per seed — [E05 notes](../../exploration/experiments/05-sequence-it/notes.md).)
5. **Reduce every violation.** Pull the affected rows' event history; find the verb that appears between the
   two bad observations every time and never otherwise; rebuild the shortest sequence, keeping the actor
   identities and clock steps it needs. Count repeated hits of one cause once. (queuelet: every I4/I8 hit had
   a worker deletion between two leases —
   [U24](../../exploration/defects/U24-worker-delete-release-uncounted-attempt.md).)
6. **Build the state table.** Every state × every event, including background events and wrong-actor
   variants (foreign worker, stale token). Columns: *should* (spec line or named rule), *predicted* (guard
   text with `file:line`, or "no guard"), *observed*, evidence. Mark ✱ where should and predicted disagree or
   the spec is silent; drive those first, then every should-be-rejected cell. A rejection passes only if state,
   counters, holder fields and history are unchanged: diff the full row. Keep *unknown* apart from
   *forbidden*; a silence is a question for the owner. (queuelet: a leased job surviving purge broke no written
   rule and became decision A2 — [reconciliation 05](../../register/reconciliations/05.md).)
7. **Add a second actor to each accepted cell.** Ask what happens if another actor touches the same row
   mid-transition. Start with set-based statements (purge, sweeps, bulk release) against per-row transitions,
   deleting an actor against its own call, and work done after the commit (notifications, history rows). Start
   both calls from a barrier with 0–50 ms jitter, run at least 20 trials, report k/N, and read the store to
   confirm each trial's starting state instead of trusting a sleep. 0/N is not absence; ask for a pause hook.
   (queuelet: 49 single-actor cells held while overlaps gave U34 and U35 —
   [lesson 07](../../lessons/07-state-it.md); post-commit races U43/U44 stayed hidden until E09 because every
   E07 cell had one actor — [lesson 09](../../lessons/09-explain-it.md).)
8. **Update the models and hand off.** Mark new observed cells ★, draw two-actor arrows dashed, list what was
   not tried. Propose the durable form to the fixer: one transition table that every state guard is built
   from, a contract test written by hand from the spec, and the walker as a model-based check that asks the
   table what each event may do; mutation-test it by reintroducing a known bad arrow. (queuelet:
   [transition.go](../../project/internal/jobs/transition.go),
   [transition_test.go](../../project/internal/jobs/transition_test.go),
   [lesson09_test.go](../../project/test/checks/lesson09_test.go).)

Stop when every ✱ and should-be-rejected cell has an observation, each success arrow has walk coverage, and
a fifth of the budget is left for writing.

## Output

- `notes.md`: per sequence, the prediction, command, trimmed output and invariant results.
- The walker and its logs: one line per step, `VIOLATION step= inv= row=` lines, per-verb success counts.
- `models/nouns-verbs.md`: `| Sequence (start state) | Result | Evidence |`, plus untested sequences.
- `models/states.md`: diagram, then `| State | Event (actors, clock) | Should (oracle) | Predicted (code) |
  Observed |`, ★ on new cells, with untried and blocked cells listed. Defect reports for minimal reproductions.

## Pitfalls the experiments exposed

- **Reading the handlers predicts the lifecycle wrongly.** E05's charter prediction ("every verb either does
  what it says from any starting point or is refused") and the explorer's walk prediction were both refuted
  ([E05 revealed](../../exploration/experiments/05-sequence-it/revealed.md)). Steps 3–6 judge only what ran.
- **A uniform walk rarely reaches the success arrows.** The 100-step `walk05.py` check never fired its
  attempts invariant at lesson-04: U24 needs three leases with a deletion between two. Step 4 counts
  successes; each finding keeps its own directed reproduction ([lesson 05](../../lessons/05-sequence-it.md)).
- **A complete single-actor table can pass and still be wrong.** E07 drove all 49 cells and every guard held;
  the faults were overlaps. Step 7 is part of the table
  ([E07 revealed](../../exploration/experiments/07-state-it/revealed.md)).
- **The event history can mislabel transitions.** After one expiry, later acks, cancels and releases were
  logged `lease_expired`, because the label read sticky error text
  ([U26](../../exploration/defects/U26-release-logged-as-lease-expired.md)). Step 3 reads history after
  sequences that pass through an earlier exit.
- **A race fixture can share the bug's assumption.** The lesson-07 reference probe assumed a job id appears
  once among lease responses; a deleted worker had leased one job twice. Assert per-epoch facts (contiguous
  attempts, distinct tokens) ([lesson 07, deviations](../../lessons/07-state-it.md)).

## Source

Elisabeth Hendrickson, *Explore It!* (Pragmatic Bookshelf, 2013): 6.1 "Nouns and Verbs", 6.2 "Random
Navigation", 8.1 "Recognizing States and Events", 8.2 "Diagramming a State Model", 8.3 "Exploring with State
Models", 8.4 "Changing the Representation: State Tables"; step 5 borrows 10.4, isolating a surprise to its
minimal reproduction. Section numbers and titles follow this repo's chapter map, checked against the
publisher's table of contents; full validation against the book text has not been done. The procedure, walker
and table columns are this repo's wording, not the book's. Lessons: [05](../../lessons/05-sequence-it.md),
[07](../../lessons/07-state-it.md), and the structural fix in [09](../../lessons/09-explain-it.md).

## Worked example

**E05** ([charter](../../exploration/experiments/05-sequence-it/charter.md)): operator and worker verbs in
unusual orders with the nouns-and-verbs list, the test clock and a random generator, to find sequences that
leave jobs in states the spec does not allow. Score: 22 stated, 18 confirmed, 2 refuted, 2 code-reading
predictions contradicted, 3 defects, 16 model cells; about 24 of 60 minutes; 950 walk steps over 5 seeds.
Found [U24](../../exploration/defects/U24-worker-delete-release-uncounted-attempt.md) (4 deliveries on
`max_attempts: 2`), [U25](../../exploration/defects/U25-terminal-rows-keep-lease-fields.md) and
[U26](../../exploration/defects/U26-release-logged-as-lease-expired.md); ran the A2 sequence and fairly judged
it a spec silence. [walk.sh](../../exploration/experiments/05-sequence-it/walk.sh), seed 11:
`VIOLATION step=128 inv=I4 job=21 grant 3 > max 2 (attempt=2)`, reduced to (lease → delete worker) × 4.

**E07** ([charter](../../exploration/experiments/07-state-it/charter.md)): the job lifecycle with a state table
built from the spec, the event history and the `WHERE` clauses, two workers and the test clock. Score: 13
stated, 12 confirmed, 1 refuted, 1 code-reading prediction contradicted (availability within 1.5 s; a 1,110-row
sweep delayed promotion past 3 s), 2 defects, 41 new cells in a 49-cell matrix. Two-actor cells from
[states.md](../../exploration/models/states.md):

| State | Event (actors) | Should | Predicted | Observed (E07) |
|---|---|---|---|---|
| queued | lease(W) ‖ DELETE W | released or 404 | leased by deleted W | 16/20 (1 job), 28/30 (50 jobs) — [U34](../../exploration/defects/U34-worker-delete-racing-lease.md) |
| leased | worker release ‖ purge | dead or deleted | queued | 9/10 (60 jobs), 6/20 (1 job) — [U35](../../exploration/defects/U35-purge-transition-race.md) |
| leased | ack ‖ ack (holder) | one 200 | one 200 | 20/20 pairs held |

Not probed: cancel-while-running (A3, a coordinator decision — [reconciliation 07](../../register/reconciliations/07.md)).
Missed: the post-commit tail, found in E09 as [U43](../../exploration/defects/U43-terminal-notification-after-commit.md)
and [U44](../../exploration/defects/U44-vanished-job-drops-pass-notifications.md). Model delta across both:
[states.md](../../exploration/models/states.md), [nouns-verbs.md](../../exploration/models/nouns-verbs.md),
[rules.md](../../exploration/models/rules.md), [causes.md](../../exploration/models/causes.md).

**Structural outcome (lesson 09).** The E07 table became `transition.go`; fourteen call sites take their state
guards from it. `TestL09ModelWalk` is the E05 walker as a model-based check: 1,000 random events over six
one-job queues, each outcome checked against `jobs.Targets`. With D14 reintroduced (a `Retry` that accepts
`cancelled`), it failed at step 241.

**Second target, not yet run: litefleet's database lifecycle.** A database row moves `provisioning` → `active`
→ `dropped` through `BeginCreate`, `FinishCreate`, `AbortCreate` and `MarkDropped` in
[registry.go](../../../release-it/project/internal/registry/registry.go), with a placement epoch the node
checks ([Release It! plan, 2.2–2.3](../../../release-it/PLAN.md)). Step 1 records which of those statements
carry a state guard. Sequences: drop before provisioning confirms; drop then re-create one name; query with
the old epoch afterwards. Walk create/drop/query/execute/heartbeat/backup asserting "an active database has
one placement" and "a dropped database has no file on any node". Second-actor cells: drop ‖ execute, drop ‖
snapshot upload, two creates of one name, a heartbeat reporting a just-dropped database.
