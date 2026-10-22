---
name: make-invisible-visible
description: Read every console a system already has side by side for one object, list what none of them answers, and request the smallest hooks (correlated events, a retained history, a controllable clock); use when you cannot tell what a job, request or background loop actually did, or the behaviour is time-based.
---

# Make invisible visible

Use this skill when the API shows only the current state and nobody can say how an object got there. You compare
the consoles the system already has for one object over one time window. You write down every question none of them
answers. Then you ask for the smallest hook that would answer each one, and you check the hook once it exists. In
this repo's queuelet arc, also follow [the protocol](../../exploration/PROTOCOL.md): the explorer proposes hooks and
the fixer builds them.

## When to use

- "I can't tell what it did": a job, delivery or request ended in a state no log line explains.
- Thin logs: unstructured `printf` lines with no object ids, or logs that are silent on state changes.
- Anything time-based: leases, retries and backoff, retention, cron, heartbeats, liveness windows, DST.
- Two observers disagree, e.g. a partner recorded success but the sender recorded a failure.
- Background loops, outboxes or sweeps whose progress only shows as rows changing in the database.

Not for: choosing input values (`sweep-variables`) or deciding what "wrong" means (`derive-oracles`).

## Inputs

- One lifecycle question about one object (a job id, a delivery, a database) and a time window.
- The source, a running local stack you may reset, and read-only database access.
- The process outputs (container logs, stdout) and any partner or receiver records (a webhook sink, a mock).
- Permission to inject faults through lab controls only: worker kinds, partner modes, container actions.

## Procedure

1. **Bound the question and predict.** Name one object, one lifecycle and one window. From the code, write the
   expected sequence of states and side effects, and what would refute it, before you drive anything.
2. **Inventory the consoles and their producers.** Find every place the system records something:
   `grep -rnE 'log\.Print|slog\.|logger\.|fmt\.Fprint'` for logs; `grep -rniE '_events|history|audit|attempts|outbox'`
   in migrations for history tables; `/health`, `/metrics`, `/debug` routes; the partner's own records. For each console,
   note three things: which ids it carries, which clock stamps it (`time.Now`, SQL `now()`, `clock_timestamp()`, an
   app clock), and whether it survives a restart or a delete.
3. **Drive the lifecycle with failures and watch all consoles at once.** Use a crashing worker, a failing handler, a
   partner returning 500, a slow partner and a malformed configuration value. Sample every console about once per second
   with one script keyed by the object id. Read the database in a single snapshot. Sequential queries can straddle a
   background tick and invent a disagreement. (queuelet: the E02 watcher showed `attempts 3` beside 4 attempt rows —
   [revealed](../../exploration/experiments/02-make-it-visible/revealed.md).)
4. **Build the console comparison.** Write one row per object and window, with one column per console. Signals that
   matter: counts that differ between consoles; a log that is silent while state changed; a partner that recorded 2xx
   while the sender recorded a failure; `/health` OK while work is stuck; an error field that belongs to an earlier attempt.
   Name the oracle for each disagreement. Some are legitimate, such as a sender timeout against a partner that finished late.
5. **List the questions no console answers.** For each, decide between "did not happen" and "cannot observe". One
   current-state row never proves the history: a job that ran twice can look like a job that ran once. Stop rule: every
   question ends up either answered by an existing console or written as a request.
6. **Write a testability request per question.** Each request has four fields: the question, the missing signal, the
   minimal hook and a validation probe (re-run step 3 and say what the hook must show). Prefer these, smallest first:
   ids and the event name added to existing log lines; structured lifecycle events; a per-object event history written in
   the same transaction as the change and kept after deletes; true start times; a controllable clock for waits that are
   too long to sit through; a deterministic pause or fault hook for races; a backlog or lag gauge with a health cause.
   Do not change the system under test inside an exploration run. (queuelet: TR-E02-1…6 in
   [notes](../../exploration/experiments/02-make-it-visible/notes.md).)
7. **When a hook lands, test the hook.** Replay the same probe and fill the comparison again. Check the new console
   against an independent one (the partner, the database), because the hook can lie. Check that events survive deletes,
   respect tenant scope and can be deduplicated by id. (queuelet: event labels came from a sticky error text —
   [U26](../../exploration/defects/U26-release-logged-as-lease-expired.md).)
8. **Map the clocks before you trust a clock hook.** List which clocks it moves and which stay real: workers, network
   deadlines, partners, heartbeats, and SQL `now()`. Note whether `set`/`advance` freezes time or keeps it ticking. A
   frozen clock stops real-time expiry and dates every row it writes. Drain your own rows before you reset it. A backward
   step can create behaviour that production cannot reach, so classify it as a risk, not a defect. Use real time when
   duration or races are the question. (queuelet: [clock.go](../../project/internal/clock/clock.go) `Change` freezes on
   set/advance, T-E05-1; [09-post-commit-races](../../docs/knowledge/09-post-commit-races.md).)
9. **Track deferred requests.** Keep every unbuilt request in the backlog and re-raise it when a later finding depends
   on it. (queuelet: TR-E02-3 backlog/health was deferred; E08 found [U37](../../exploration/defects/U37-webhook-shipper-starvation.md)
   with `/health` at 200 during the stall and asked again as T-E08-2 — [hypotheses](../../exploration/hypotheses.md).)

## Output

- `notes.md` section "Console comparison": | Object / interval | Prediction | API | DB | Process logs | Partner | Disagreement |
- Unanswered questions, numbered, each marked *did not happen* or *cannot observe*.
- Testability requests: | Id | Question | Missing signal | Minimal hook | Validation probe | Status (open / built in lesson NN / deferred: why) |
- After hooks land: | Request | Probe replayed | What the hook showed | Independent console agrees? | Remaining blind spot |
- Model delta: lifecycle cells in `models/states.md`, console producers in `models/map.md`, clock and partner cells in `models/ecosystem.md`.

## Pitfalls the experiments exposed

- **Reading the current row as history.** GET showed `last_error:"lease expired"` while the next attempt was running.
  The error belonged to the previous attempt (E02 S1, [notes](../../exploration/experiments/02-make-it-visible/notes.md)).
  Step 5 forbids history claims from one row.
- **Consoles that do not persist.** Docker kept only the current container's exit code, and `RestartCount` was 0 again after
  a reset, so "crashed" was an inference (Q-E02-1). Step 2 records persistence per console. `docker kill` is not a crash
  either ([08-neighbour-faults](../../docs/knowledge/08-neighbour-faults.md)).
- **Trusting the new hook.** After one lease expiry, ack, cancel and worker release were all logged `lease_expired`
  ([U26](../../exploration/defects/U26-release-logged-as-lease-expired.md), found in E05). Step 7 cross-checks hooks.
- **Clock artifacts.** Backward clock steps produced FIFO inversion and early leases; E04 judged them lab-only (H-E04-08) and
  E09 withdrew two drafts on that ground ([revealed](../../exploration/experiments/09-explain-it/revealed.md)). A frozen
  clock left no real-time stamp to measure sweep length (T-E07-1). Step 8 handles both.
- **Deferring the health hook.** With no backlog signal, a stalled shipper looked healthy until E08
  ([revealed](../../exploration/experiments/08-surround-it/revealed.md)). Step 9 keeps the request alive.
- **Probabilistic races with no pause hook.** An E06 race window was not hit in 320 tries, and the E07 and E09 races were
  probabilistic; the pause and fault hooks they asked for (T-E06-2, T-E07-2, TR-E09-2) were not built. Judge a race by what
  committed, read from the event table
  ([09-post-commit-races](../../docs/knowledge/09-post-commit-races.md)).

## Source

Elisabeth Hendrickson, *Explore It!* (Pragmatic Bookshelf, 2013), Chapter 3 "Observe the Details": 3.1 (the moonwalking
bear), 3.2 "Digging Deeper", 3.3 "Testability and Making the Invisible Visible", 3.4 "Consoles and Logs". Section
numbers follow the plan's chapter map, checked against the publisher's table of contents; the full book text has not
been validated against these steps. The procedure wording is this repo's, not the book's. Demonstrated in
[lesson 02](../../lessons/02-make-it-visible.md); the hooks are used in [lesson 05](../../lessons/05-sequence-it.md),
[lesson 08](../../lessons/08-surround-it.md) and [lesson 09](../../lessons/09-explain-it.md).

## Worked example

**Experiment 02, queuelet at `explore-it/lesson-01`.** Charter: explore the job lifecycle with the API and worker logs,
`psql`, the background-loop source and the `crash`/`fail` worker kinds to discover what happens to a job that the API
does not show ([charter](../../exploration/experiments/02-make-it-visible/charter.md)). Score
([revealed](../../exploration/experiments/02-make-it-visible/revealed.md)): 19 hypotheses stated, 19 confirmed, 0 refuted,
0 code-reading predictions contradicted, 3 defects found, 46 model cells filled. Finds:
[U09](../../exploration/defects/U09-sample-worker-kind-registration.md) (a documented `fail3` kind never leased),
[U10](../../exploration/defects/U10-malformed-webhook-blocks-independent-delivery.md) (one unparsable webhook URL
stopped every tenant's notifications) and E02-F03, a spec clarification, not a bug (a slow partner recorded 200 five times
while queuelet counted five failures). D03 and D04, which the plan aimed at E02, had been found and fixed in lesson 01;
here they held as regressions. No prediction was refuted, but the charter's own refutation condition (consoles disagree)
was met in the slow-partner variation. Whether the core loop kept running while the shipper was blocked was inferred, not
observed. Model delta: `states.md` 13 cells, `rules.md` 7, `ecosystem.md` 7, `map.md` 5, `variables.md` 5,
`nouns-verbs.md` 5, `entities.md` 4 ([models](../../exploration/models/README.md)).

Excerpt of the console comparison ([notes](../../exploration/experiments/02-make-it-visible/notes.md), trimmed):

| Object / interval | API | DB | Process logs | Partner | Disagreement |
|---|---|---|---|---|---|
| job 301 crash | dead, attempts 3 | 3 `expired` rows | worker: 3 × processing; API: nothing | — | start times and exit cause invisible |
| job 302, partner in 500 mode | succeeded | outbox attempts 12 | API: 1 line, no id | 12 records | API log hides 11 failures |
| outbox 71 bad URL | — | attempts 0 forever | parse error each second, no id | 0 records | health "degraded" without cause |

The six requests became three hooks in lesson 02 ([reconciliation](../../register/reconciliations/02.md)): JSON
lifecycle events with ids (T01), `job_events` plus `GET /v1/jobs/{id}/events` with the true lease start (T02), and a
lab-only API clock (T03). TR-E02-3 and TR-E02-4 were deferred. Replaying the crash probe, crash job 395 now showed
`created, leased, lease_expired, notification_failed` with outbox 128 and status 500; T03 advanced 3,662 semantic seconds
in 1.95 real seconds ([lesson 02](../../lessons/02-make-it-visible.md)). Later runs used them: the E08 DST probes ran on
the clock ([A5](../../exploration/defects/A5-dst-gap-and-overlap.md)), and the events exposed U26 in E05.

**Second target (not yet run): litefleet's placement decisions.** Question: why did a new database land on a node?
Consoles: the `POST /v1/databases` response (`node`), the `placement` and `nodes` tables, each node's `_node.db`, and the
control-plane log. [`createDatabase`](../../../release-it/project/cmd/control/main.go) (line 91) calls
`NodeLoad` and [`placement.Pick`](../../../release-it/project/internal/placement/placement.go). Only the winner is
stored; the candidate loads, the nodes dropped as stale and the tie-break are logged nowhere. The likely minimal hook
is one `placement_decided` event carrying those fields. Mind the clock: liveness is SQL
`now() - interval '30 seconds'` ([registry.go](../../../release-it/project/internal/registry/registry.go) line 94), so an
app-level test clock would not move it. Probe staleness by pausing a node's heartbeats instead.
