# Explore It! — breakdown plan v2: one inherited job queue, learned through nine experiments

**Book:** *Explore It! Reduce Risk and Increase Confidence with Exploratory Testing*, Elisabeth
Hendrickson, Pragmatic Bookshelf, 2013 (first edition, the only one; 186 pages, ISBN
978-1-937785-02-4). Chapter and section numbers below are from that edition's table of contents
(verified against the publisher's preview).

**Phase:** plan (v2), for Nick's confirmation. Design only: no SUT code, no experiments, no skills
exist yet. Written on branch `plan/explore-it` in its own worktree while the Release It! build runs on
`main`; this pass adds files under `books/explore-it/` only. Suggested edits to `AGENTS.md`, the root
`README.md` and `STATE.md` are listed in section 10 as notes, to be applied when the branch merges.

**What changed from v1.** Nick scoped the book: *"That book had a lot of fluff so feel free to skip
shallow concepts or netting with people soft skills things. I'm interested in the probing, testing,
learning of new codebase part."* And the shape he wants: *"a good system in place with interesting
experiments set of to learn about it and cool skills to explore codebases."* v2 is v1 cut to that
centre. Section 0 says what was cut and why what remains is the book's core. The vehicle is unchanged.
Eleven lessons became **nine experiments**; eight people-and-process skills became **eight
codebase-exploration skills** (different eight); the register gained one defect and lost one lesson's
worth of theatre. The v1 → v2 changelog is at the end.

---

## 0. What this plan deliberately omits, and why the rest is the book's real core

Hendrickson's book has two layers. One is a **method for learning a system you did not write by
experimenting on it**: state what you think it does, poke it in a way that would prove you wrong,
watch everything it reveals (not just the response), and fold what you learned into a model that
tells you where to poke next. That layer is Chapters 2–9 and the parts of 10–13 that are about the
system. The other layer is about **people and teams**: how to talk to stakeholders, pair, estimate,
run requirements meetings, debrief managers and build a team's heuristics list. Nick asked for the
first layer only. This plan takes him at his word.

| Omitted | Book sections | What v1 had for it | Why it goes |
|---|---|---|---|
| Stakeholder interviews and "sharing observations" | 11.2, 11.3 | a written Q&A with the coordinator playing product owner in L01 | theatre about communication; the questions it produced come just as well from reading the spec and the code (12.4, folded into E01) |
| Requirements meetings and chartering during them | 12.1, 12.2, 12.3 | L10, a whole lesson with "decisions.md" from a role-played product owner | a meeting technique; the one durable idea (the spec is silent, so the code decided: find out what) is now a hypothesis source in every experiment, and the four planted ambiguities are re-homed in the experiments that probe them |
| Personas | 6.3 | the maintenance operator, the impatient producer, the flaky worker in L05 | a shallow generator; the sequences it produced come from nouns × verbs and the state table directly |
| Paired exploration, estimating, "explored enough", stakeholder debriefs, the team's nuggets list, exploration in the test strategy | 13.1, 13.2, 13.4, 13.5, 13.6, 13.7 | L11, a 90-minute paired session with `STRATEGY.md` and a debrief written for Nick | team process; the one systemic idea in Ch. 13 (13.3, read the defect list for causes) is now the last experiment, E09, and "findings become regression checks" is the fixer's standing rule rather than a strategy document |
| The Nightmare Headline Game as a group activity | 2.5 | the risk-brainstorm ritual closing the recon | the output (risk hypotheses) is kept as one of the hypothesis sources in `charter-as-probe`; the game is not |
| Practice sessions; interviewing testers; web heuristics; exploring a programming language | 1.4, A1, A2.2, 10.2 | omitted in v1 too | out of scope for an API-only SUT and a reader who is an engineer, not a hiring manager |
| Session-based test management as a team ritual, the debrief as a report | 1.3 (the management side), 13.6 | `debrief.md` in the stakeholder template | the time box stays as a budget on an experiment; the "debrief" is replaced by `revealed.md`: what the system showed and how the model changed, written for the next experiment, not for a manager |

What survives is **exploration as probing an unfamiliar codebase**: charters as hypotheses (Ch. 2),
observing everything the system reveals and adding hooks when it reveals too little (Ch. 3), varying
inputs, data, state and sequences on purpose (Ch. 4, 6), knowing what "wrong" means when nobody wrote
it down (Ch. 5), building models (entities, states, ecosystem) whose empty cells are the next probes
(Ch. 7–9), doing it through tools because there is no UI (Ch. 10), starting with a recon of a system
you inherited (11.1, 11.4), characterizing a surprise down to its variables (10.4, 11.5–11.6), and
reading the defect list for what it says about how the code was written (13.3). That is the method,
and it is what an engineer dropped into a new codebase actually uses. The step count fell from eleven
to nine because the two lessons built around people content had nothing left once the people content
went; the surviving steps are not thinner, they are the same probes with the ritual stripped off.

## 1. The shape: experiments against an inherited codebase

`AGENTS.md` was written for *Release It!*: the book names flaws in a system, so the archetype is a
flawed system and each lesson is a patch that fixes a flaw, with a verify script that reproduces the
flaw on the previous tag. *Explore It!* names no flaws in systems; it teaches a way of finding them.
The thing that gets better step by step is the **explorer's model of the system**, and the system
improves as a consequence. The adaptation keeps every structural element of the playbook and
changes what each one holds:

| Playbook element | Release It! | Explore It! (this plan) |
|---|---|---|
| The archetype | one flawed service, improved by patches | one **inherited** service (the system under test, SUT) seeded with realistic defects, plus the **exploration record** beside it (`exploration/`: hypotheses, probe logs, models, defect reports) |
| What a lesson is | one technique of the book applied as a fix | one **experiment**: a hypothesis about the system, a technique of the book used as the probe, a time-boxed run against the SUT, what the system revealed, how the model changed; then the defects it found are **fixed and turned into checks** |
| The lesson commit | patch + note + verify script | **two commits**: the experiment (charter, probe log, revealed, models, defect reports) and the fix (SUT patch + regression checks + note); the lesson tag points at the fix |
| The register | flaws → chapter → lesson | defects → **the probe that reveals them** → chapter → experiment; plus four spec silences (where the code decided) and three testability gaps |
| The verify | load + chaos + thresholds; `IMPROVED:` / `FLAW REPRODUCED:` | the experiment's black-box **reproduction scripts** plus the regression checks; `FIXED AND CHECKED:` / `DEFECTS REPRODUCED:`. `REF=<earlier tag>` shows the defect was present all along: the **dwell table** (which tags the defect sat in while earlier experiments ran past it) is the payoff |
| "Skill compounding" in `git log` | the system hardens | early experiments reveal shallow defects; later, better-modelled ones reveal the deep ones (stale acks, lost cancels, DST drift, a loop that silently dies) that the earlier ones walked past, and the last experiment predicts a defect from the pattern before probing for it |
| Skills | written after the lessons, from the lessons | written **before** the experiments, from the book, and **executed** by them; each experiment is a skill's worked example and sends corrections back |

Three rules keep the record honest rather than theatre:

1. **The explorer is blind to the register, not to the code.** The agent that runs an experiment
   never sees this plan's register, the proof scripts or the SUT's git history. It works from a
   plain export of the SUT tree (source included), the exploration record so far, the skill for the
   experiment and its charter. It **may read the code**: this is a book about learning a codebase,
   and reading is one of the two ways to form a hypothesis (the other is the running system). Every
   hypothesis records its source (spec, code, model, or a surprise) and whether the probe confirmed
   it. Defects it finds are real finds; defects it misses are recorded as misses and disclosed in the
   lesson note. The seeder (who plants the register) and the explorer are agents from **different
   model families**, which the repo's cross-family rule already requires for review.
2. **The record is committed as produced.** Probe logs keep their timestamps and their dead ends.
   The coordinator formats, never rewrites. A reader should be able to see an experiment go wrong.
3. **Every experiment is scored.** `revealed.md` ends with a table: hypotheses stated, confirmed,
   refuted; predictions made from reading the code that the running system contradicted; defects
   found; model cells filled. The score is the learning curve the lesson notes draw.

Everything else in `AGENTS.md` (one archetype, real technology, `make up / make verify LESSON=NN
[REF=tag]`, lesson notes outside `project/`, tags, the review bar, commit style) applies unchanged.
"Lesson NN" in the repo's vocabulary is "experiment NN" in this plan's; they are the same thing.

## 2. The book's core, as this plan uses it

### 2.1 Core argument

Testing is two activities. **Checking** confirms the software does what you already decided it
should; it is designed in advance and is what test cases and automated suites hold. **Exploring** is
designing experiments, running them and learning from the result in the same motion, so each
observation shapes the next experiment (Ch. 1.1, 1.2). Checks catch regressions in the risks you
named; exploration finds the risks you did not name. For an engineer inheriting a codebase the second
activity is the only one available on day one: there is nothing to check against yet, only a system
to learn. The book's program, stripped to that use: give each probe a **charter** so it has a
hypothesis and a budget (Ch. 2); **observe** what is actually happening, including what the interface
hides, and ask for hooks when it hides too much (Ch. 3); find the **variables** and vary them on
purpose (Ch. 4); decide what "wrong" means when no one wrote it down (Ch. 5); use **models** of
sequences, entities, states and the surrounding ecosystem as generators of probes (Ch. 6–9); do it
all through tools, because there is no UI (Ch. 10); start with a recon of the system you inherited
(11.1, 11.4); characterize every surprise to its minimal reproduction (10.4, 11.5–11.6); and read the
defect list for what it says about how the code was written (13.3).

### 2.2 Chapter map (kept sections only)

| Chapter | Sections kept | What it gives the explorer of a new codebase |
|---|---|---|
| 1. On Testing and Exploration | 1.1, 1.2, 1.3 (the time box as a budget) | exploration as simultaneous learning, design and execution; the experiment as the unit |
| 2. Charter Your Explorations | 2.1, 2.2, 2.3, 2.4, 2.6 (2.5 as a one-line hypothesis source) | the template *Explore `<target>` with `<resources>` to discover `<information>`*, read here as a hypothesis with instruments; too broad / too narrow; where hypotheses come from; a backlog of them |
| 3. Observe the Details | 3.1–3.4 | inattentional blindness; looking past the interface; asking for testability hooks; reading the consoles the system already has |
| 4. Find Interesting Variations | 4.1–4.4 | variables beyond input fields: counts, sizes, positions, timing, sequences, configuration; the named heuristics of A2.1 |
| 5. Evaluate Results | 5.1–5.3 | oracles when there is no spec: Never/Always rules, comparable systems and standards, approximations |
| 6. Vary Sequences and Interactions | 6.1, 6.2 | nouns and verbs combined in odd orders; random walks |
| 7. Explore Entities and Their Relationships | 7.1–7.3 | the entity model as a generator: entity × CRUD, dependency × "what happens to the dependent", data followed through every place it appears |
| 8. Discover States and Transitions | 8.1–8.4 | the state model as a generator: every state × every event, especially the cells the diagram has no arrow for |
| 9. Explore the Ecosystem | 9.1–9.4 | the context diagram; what crosses a trust boundary unchecked; "what if this neighbour misbehaves"; data across the boundary |
| 10. Explore When There Is No User Interface | 10.1, 10.3, 10.4 | tools as the interface; isolating a surprise to its minimal reproduction and the variables that matter |
| 11. Explore an Existing System | 11.1, 11.4, 11.5, 11.6 | the recon of an inherited system; which probe fits which question; recording findings; unreproducible bugs |
| 12. Explore Requirements | 12.4 only (active reading), folded into E01 and the hypothesis sources | reading a spec adversarially: what does it not say, what would make it false |
| 13. Integrate Exploration Throughout | 13.3 only | reading the defect list for causes, not symptoms |
| A2. Test Heuristics Cheat Sheet | A2.1 | the kit `sweep-variables` draws from |

### 2.3 What an engineer does differently after this arc

- Meets a new codebase with a **recon**: entry points, nouns and verbs, what the spec is silent on,
  what the code decided there, and a backlog of hypotheses ranked by what would hurt most.
- Writes each probe as a **charter with a hypothesis** and a budget, not as "poke around".
- Treats every observation as data: reads logs, the database and event histories alongside the
  response, and **adds hooks** (structured events, a controllable clock) when the system hides what
  it is doing.
- Names the **variables** of a surface before touching it and varies them with named heuristics
  instead of trying "a few values".
- Writes the **oracle** down before judging an output: the Never/Always rule, the comparable system,
  the approximation.
- Draws **models** (nouns and verbs, entities and CRUD, states and events, the ecosystem with its
  trust boundaries) because the model's blank cells are where the bugs are, and keeps them as living
  files.
- **Follows surprises** to a minimal reproduction with the variables isolated, then looks for the
  sibling: the same cause elsewhere in the code.
- Reads the defect list for **how the code was written** (a missing guard pattern, a silent-accept
  habit) and fixes the cause structurally, with a model-based check that keeps it fixed.

The vocabulary is Hendrickson's (charters, sessions, heuristics, models). "Tours" are Whittaker's
(*Exploratory Software Testing*, 2009) and are not used.

## 3. The vehicle: `queuelet`, an inherited baby job queue, and the exploration record beside it

### 3.0 Why this vehicle (unchanged from v1)

What this arc needs is a system with **many states and events** (Ch. 8), **entities with
dependencies** (Ch. 7), **many nouns and verbs** that combine (Ch. 6), **variables of every kind**
including time (Ch. 4), an **ecosystem with a trust boundary** (Ch. 9), **no UI** (Ch. 10), a spec
that is silent in places (12.4) and a codebase small enough to read in an afternoon yet large enough
that reading alone does not reveal everything. Seeded defects must be **semantic** (wrong state, lost
cancel, off-by-one, precision loss, wrong timezone), the kind an engineer finds by thinking and
probing, not by load.

| Candidate | Fit | Why it did or did not win |
|---|---|---|
| **Reuse `litefleet`** (the Release It! SUT) | medium | Real, in Nick's domain, already built. But its register is stability-shaped (load and fault injection find its flaws), its tenant surface is a SQL passthrough with few verbs and little state of its own, and seeding it would mean editing the Release It! codebase during that arc. Litefleet stays a **second worked example** for two skills once both arcs land (section 7). |
| **A baby durable job queue** | near 1:1 | Nick's second domain. A job has eight states and a dozen events, so sneak paths are everywhere. Entities with real dependencies: queues own jobs and schedules, workers hold leases, schedules enqueue jobs. Verbs galore. Variables of every kind: payload size and numeric precision, priority, delay, batch count, cron with time zones. An ecosystem: Postgres, workers, a webhook partner across a trust boundary, the clock. No UI: an HTTP API and a CLI. A queue is the substrate of the workflow platform and of a deploy pipeline, so it is the smallest thing that still looks like what Nick builds. |
| Workflow platform | good but large | Everything fits, but the minimum (definitions, steps, timers, history) is 2–3× the budget before a single defect is planted; the queue is its kernel. |
| Baby Vercel product (deployments) | medium | Deployments have states, but a build pipeline needs real work to run and most of the richness is cost without exploration payoff. A deploy queue is, again, a job queue. |

### 3.1 What it does

`queuelet` is a single-binary job queue service over Postgres, the kind a small platform team writes
for itself and later wishes it had tested. Producers **enqueue** jobs (a kind, a JSON payload, a
priority, an optional delay) into named **queues** inside a **namespace** (the tenant). **Workers**
register, **lease** batches of jobs with a lease timeout, **heartbeat** to extend, and **ack** or
**nack** (with or without retry). The service **expires** leases of crashed workers, **retries** failed
jobs with backoff up to `max_attempts`, **dead-letters** the rest, fires **schedules** (cron
expressions in an IANA time zone), **purges** terminal jobs after a retention window, and **notifies** a
per-queue **webhook** (HMAC-signed) when a job reaches a terminal state. Operators pause, resume and
purge queues, cancel and retry jobs, and read stats. There is a CLI, `qctl`, for all of it. The
service is "inherited": it has a spec, thin tests, a README, and nobody on the team has explored it.
It is about 2 500 lines of Go: readable in an afternoon, which is the point.

### 3.2 API (v0)

All routes take `Authorization: Bearer <key>`; a key belongs to one namespace. One admin key sees
every namespace. Errors are `{"error": {"code", "message"}}` with 4xx/5xx status (with the exceptions
the register plants). Host port 18200.

| Route | Purpose |
|---|---|
| `POST /v1/queues` `{name, lease_seconds=30, max_attempts=3, backoff:{kind: exponential\|linear, base_seconds=5, max_seconds?}, retention_hours=24, webhook:{url, secret}?}` | create a queue |
| `GET /v1/queues`, `GET /v1/queues/{q}`, `PATCH /v1/queues/{q}`, `DELETE /v1/queues/{q}` | CRUD |
| `POST /v1/queues/{q}/pause`, `/resume`, `/purge` | operator verbs |
| `POST /v1/queues/{q}/jobs` `{kind, payload, priority=5 (0–9), run_at? \| delay_seconds?, max_attempts?, idempotency_key?}` → `201 {id, state}` | enqueue |
| `POST /v1/queues/{q}/jobs:batch` `{jobs:[…]}` (1–500) → `201 {ids}` | batch enqueue |
| `GET /v1/queues/{q}/jobs?state=&kind=&limit=&cursor=` | list (keyset paging) |
| `GET /v1/jobs/{id}`, `GET /v1/jobs/{id}/attempts`, `POST /v1/jobs/{id}/cancel`, `POST /v1/jobs/{id}/retry`, `DELETE /v1/jobs/{id}` | job verbs |
| `POST /v1/workers` `{name, kinds}` → `{worker_id, token}`; `DELETE /v1/workers/{id}` (releases its leases) | worker registry |
| `POST /v1/queues/{q}/lease` `{worker_id, max_jobs (1–100)}` → `{jobs:[{id, kind, payload, attempt, lease_token, lease_expires_at}]}` | lease a batch (`FOR UPDATE SKIP LOCKED`) |
| `POST /v1/jobs/{id}/heartbeat` `{lease_token, extend_seconds}`; `/ack` `{lease_token, result}`; `/nack` `{lease_token, error, retry=true, delay_seconds?}` | the worker protocol |
| `POST /v1/schedules` `{name, queue, kind, payload, cron, timezone, enabled}`; `GET/PATCH/DELETE /v1/schedules/{id}`; `POST /v1/schedules/{id}/run` | cron schedules |
| `GET /v1/queues/{q}/stats`, `GET /v1/namespaces/{ns}/stats` | counts by state, oldest queued age, throughput |
| `GET /health`, `GET /metrics` (minimal) | ops |

Job states: `scheduled` (run_at in the future) → `queued` → `leased` → `succeeded` | `retry_wait` →
`queued` … | `dead`; `cancelled` from any non-terminal state. Events: enqueue, run_at reached,
lease, heartbeat, ack, nack(retry), nack(no retry), lease expired, cancel, retry, delete, purge,
retention sweep, worker deregistered, queue paused/resumed/deleted. The v0 code has no central
transition table; each handler's `UPDATE … WHERE` carries whatever guard its author remembered. That
is the pattern E09 names and fixes (Ch. 13.3).

A background loop in the service, once per second: promote `scheduled` → `queued`; expire leases;
fire due schedules; sweep retention; ship the webhook outbox.

**CLI** `qctl` (`queue create|ls|get|pause|resume|purge|rm`, `enqueue`, `jobs ls`, `job get|attempts|
cancel|retry|rm`, `worker run <q> --kinds …` (the sample worker), `schedule …`, `stats`; from E02
also `watch`, `job events` and `clock`). It is the explorer's main instrument (Ch. 10.1, 10.3), and
also its blind spot: it only ever sends well-formed requests (E03).

### 3.3 Data model

Postgres 17, hand-written SQL: `namespaces(id, name)`, `api_keys(key_hash, namespace_id, admin)`,
`queues(id, namespace_id, name, lease_seconds, max_attempts, backoff_kind, backoff_base, backoff_max,
retention_hours, webhook_url, webhook_secret, paused, unique(namespace_id, name))`, `jobs(id bigint,
queue_id, kind, payload jsonb, priority, state, run_at, attempts, max_attempts, idempotency_key,
leased_by, lease_token, lease_expires_at, result jsonb, last_error, created_at timestamp(0),
finished_at)`, `attempts(job_id, n, worker_id, started_at, finished_at, outcome, error)`,
`workers(id, namespace_id, name, kinds, token, last_seen_at)`, `schedules(id, namespace_id, queue_id,
name, kind, payload, cron, timezone, enabled, next_run_at, last_run_at)`, `webhook_outbox(id, job_id,
url, body, attempts, next_at, delivered_at)`, `job_events(job_id, at, event, detail)` (added in E02).
Several of these columns are where the register lives (section 4).

**Seed** (`make seed`, seconds): namespaces `acme` and `globex` with keys, queues `emails`,
`reports`, `imports` with different lease/retry settings, one webhook on `reports` pointing at the
hook sink, 300 jobs in mixed states, two schedules (one in `America/New_York`), and workers
`worker-a` and `worker-b` running. `results/manifest.json` lists keys and ids.

### 3.4 Topology and stack (Docker Compose)

```
 qctl / curl / explorer ──► api :18200 ──► postgres :15433
 worker-a, worker-b (cmd/worker: kinds echo, sleep, fail{n}, crash, http) ──► api  (lease/heartbeat/ack/nack)
 api ──► hooksink :18201 (webhook partner fake: records deliveries, verifies HMAC, /admin/mode: ok|500|slow|timeout)
 worker 'http' kind ──► hooksink (an outbound call across the boundary)
 test clock (from E02): api and workers read `clock.Now()`; `qctl clock set|advance|reset` under QUEUELET_TEST_CLOCK=1
```

| Choice | Pick | Why |
|---|---|---|
| Language | **Go 1.26**, one module (`cmd/api`, `cmd/worker`, `cmd/qctl`, `cmd/hooksink`) | Same toolchain as litefleet, so the builder's and reviewer's setup carries over; static binaries, fast Compose builds; small enough to read whole. `encoding/json`'s `float64` default is itself one of the planted defects (numeric precision), which is realistic Go. |
| Store | **Postgres 17** (`postgres:17-alpine`, cached on this VM), `FOR UPDATE SKIP LOCKED` leasing | The canonical real-stack job queue (River, Graphile Worker, pg-boss, Oban all do this). `timestamp` vs `timestamptz`, `timestamp(0)` precision and `jsonb` normalization are real places for defects. |
| Cron | `robfig/cron/v3` parser with IANA locations via `time.LoadLocation` | Real cron semantics; the DST defect is in how v0 *uses* it, not in the library. |
| Partner fake | own `hooksink` (~150 lines) | Records every delivery with body and signature verdict; `/admin/mode` makes it misbehave (500, slow, timeout) for the Ch. 9 what-ifs. No Toxiproxy: the what-ifs here are "the partner answers badly" and "Postgres restarts", which `/admin/mode` and `docker compose restart` cover. |
| Clock | an app-level `clock` package; `QUEUELET_TEST_CLOCK=1` enables `POST /admin/clock` | E02's testability hook (Ch. 3.3), used by E04, E05, E07, E08 and the checks. Every SQL statement takes `$now` from the app; nothing uses Postgres `now()`, so the test clock is the only clock. |
| Observability | `log/slog` JSON from E02 (v0: sparse `log.Printf`); `GET /v1/jobs/{id}/events` from E02; a minimal `/metrics` | The book's "consoles and logs" (3.4); not Prometheus-scale: there is no load story here. |
| Explorer tooling | the source tree, `qctl`, `curl`, `jq`, `psql` (read-only role `explorer`), `docker compose logs` | Ch. 10: no UI means the tools are the interface. The read-only `psql` role and the source are resources the charters name (Ch. 2.2). |
| Checks | Go tests under `project/test/checks/` (build tag `checks`, run against the running stack with the test clock) and unit tests | Exploration finds, checks keep. Reproduction scripts (`lab/repro/D##.sh`) are black-box and separate from checks. |

Host ports in the 1xxxx range (18200 api, 18201 hooksink, 15433 Postgres) because 5432 and 8080 are
taken on this VM.

### 3.5 What v0 does right (so the defects are believable)

Clean package layout; `SKIP LOCKED` leasing that works; retries with backoff; dead-lettering on
nack; a webhook outbox with retries; cron schedules; keyset paging; a CLI that covers the API; unit
tests for the lease query and the backoff math; a README; a spec (`project/SPEC.md`, ~120 lines,
written as a competent team writes one: mostly right, silent in four places). The happy path
(enqueue → lease → ack → webhook) works and `make verify LESSON=00` proves it.

### 3.6 Testability gaps at v0 (the Ch. 3 register)

| ID | Gap | Why it matters | Closed in |
|---|---|---|---|
| T1 | Logs are sparse `log.Printf` lines without job ids or state transitions | you cannot tell from the console what a job did | E02 |
| T2 | No per-job event history; `GET /v1/jobs/{id}` shows only the current state | a job that ran twice looks like a job that ran once | E02 |
| T3 | No clock control; lease expiry, backoff, retention and cron can only be explored by waiting | a one-hour backoff or a DST transition cannot be probed in an hour | E02 |

Closing these is a legitimate SUT patch (Ch. 3.3 is about asking for exactly this), and the hooks are
the instruments later charters name.

## 4. The deliberate-defect register

Every defect planted in v0, where it lives, how it shows, which probe reveals it, and the experiment
expected to reveal it. This is the answer key for the skills (section 7) and the only list of the
defects. The explorer never sees it (section 5.2). "Expected in" is a prediction, not a rule: a defect
found early is fixed early and the later experiment's charter aims at what is left; a defect missed
is disclosed in the lesson note (section 5.2, rule 4). Defects marked *opt.* may be dropped by the
builder if checkpoint A runs long. The **Read?** column says whether a careful reader of the code can
spot the defect without running anything (R), needs the running system (P), or needs both (RP): the
experiments are designed so that each kind appears, and `revealed.md` scores which route found it.

| ID | Defect in v0 (file / area) | How it shows | Probe that reveals it (chapter) | Read? | Expected in |
|---|---|---|---|---|---|
| D01 | `qctl` prints API errors but exits 0 (`cmd/qctl`) | a script cannot tell a failed enqueue from a successful one | recon: first hour with the CLI and its source (11.1) | R | E01 |
| D02 | `DELETE /v1/queues/{q}` returns 204 when no such queue exists (rows-affected unchecked) | typos in a delete script "succeed" | recon; CRUD on the obvious nouns (11.1, 7.2) | RP | E01 |
| D03 | lease expiry re-queues the job but neither increments `attempts` nor writes an attempt row (`internal/jobs/expire.go`) | a worker that crashes retries forever; `max_attempts` never reached; the job looks like it ran once | consoles and the event history once T1/T2 exist; the `crash` worker kind (3.3, 3.4) | P | E02 |
| D04 | the webhook shipper marks a delivery `delivered` on any HTTP response; status is logged, not checked (`internal/webhook/ship.go`) | logs say delivered, hooksink says 500 | two consoles disagree (3.4) | RP | E02 |
| D05 | `priority` outside 0–9 is accepted and sorted (`-5`, `100`) | one job with priority 100 pre-empts every queue forever | Goldilocks / boundaries on a numeric variable (4.3, A2.1) | RP | E03 |
| D06 | `jobs:batch` with 501+ jobs silently inserts the first 500 and returns 201; a batch of 0 returns 201 `{ids: []}` (spec: 400) | producers lose jobs without an error | Zero, One, Many / Too Many on a count (4.3, A2.1) | R | E03 |
| D07 *opt.* | queue names with `/` or spaces are accepted at create but cannot be addressed by any later route | a queue exists that nothing can reach | strings variation: special characters (4.3, A2.1) | P | E03 |
| D08 | payload is decoded into `map[string]any` for validation and re-encoded, so integers above 2^53 lose precision (`9007199254740993` → `…992`) (`internal/api/enqueue.go`) | a worker receives a different id than the producer sent | numbers at 2^53 / 2^63 boundaries (4.3); or Follow the Data (7.3) if missed | RP | E03 (or E06) |
| D26 | `PATCH /v1/queues/{q}` silently ignores unknown fields (typo `lease_secs` → 200, no change); `?state=` with an unknown value silently drops the filter | the API never tells you you were wrong | request-shape variation with raw `curl`, not `qctl`: unknown fields, unknown enum values (4.3, 10.1) | RP | E03 |
| D09 | nack's retry decision uses `attempts > max_attempts` after incrementing (`internal/jobs/nack.go`) | `max_attempts: 3` runs four times | Never/Always: "attempts never exceed max_attempts" (5.1) | R | E04 |
| D10 | `stats.queued` counts `scheduled` and `retry_wait` rows too; `oldest_queued_age` goes negative for future `run_at` and is clamped to 0 (`internal/stats`) | the dashboard number disagrees with `SELECT count(*)` | approximations: "is this number roughly right?" against a direct count (5.3) | RP | E04 |
| D11 | `created_at` is `timestamp(0)`; the lease query orders by `priority DESC, created_at` | jobs enqueued within the same second lease in arbitrary order; spec promises FIFO within a priority | Never/Always: "always FIFO within a priority"; alternative resource: the spec and SQS FIFO docs (5.1, 5.2) | RP | E04 |
| D25 | `POST /v1/queues/{q}/lease` with an unknown worker returns HTTP 200 `{"error": …}` | clients treat the error as an empty lease | Never/Always: "an error is always a 4xx/5xx" checked against every route (5.1, 10.3) | RP | E04 |
| D12 | `purge` deletes only `queued` rows; `scheduled` and `retry_wait` survive (spec: "all pending jobs") | enqueue-with-delay → purge → the job still runs | nouns and verbs in odd orders (6.1) | RP | E05 |
| D13 | while a queue is paused the scheduler skips cron ticks; on resume it fires every missed tick at once | a one-hour pause of a `*/5` schedule enqueues twelve jobs on resume | pause × schedule tick × resume, with the clock (6.1, 6.2) | RP | E05 |
| D14 | `POST /v1/jobs/{id}/retry` has no state guard: allowed from `cancelled` and `succeeded` (spec: from `dead` only) | a cancelled or finished job runs again | random sequences (6.1, 6.2); or the state table (8.4) if missed | R | E05 (or E07) |
| D15 | `DELETE /v1/queues/{q}` deletes the queue row only: jobs and schedules stay (no cascade, no guard) | orphan jobs remain leased or queued; namespace stats count them; the schedule keeps failing every tick | dependencies: "what happens to the dependents" (7.1), CRUD delete (7.2) | RP | E06 |
| D16 *opt.* | `PATCH /v1/schedules/{id}` changing `cron` or `timezone` does not recompute `next_run_at` | the old time fires once more | CRUD update on the schedule entity (7.2) | R | E06 |
| D17 *opt.* | the webhook body truncates `result` to 1 KB and the dead-letter `last_error` to 256 chars with no truncation flag; `GET /v1/jobs/{id}` returns them whole | the same datum differs by the path you read it through | Follow the Data (7.3) | RP | E06 |
| D18 | `ack` checks `state = 'leased'` but not `lease_token` (`internal/jobs/ack.go`) | worker A's lease expires, B leases, A's late ack succeeds, B's ack gets 409 and B's result is lost | sneak path: ack × leased-by-another (8.3); readable in the `WHERE` clause once the state table asks the question | RP | E07 |
| D19 | `nack` has no state guard | cancel a leased job; the worker's later nack moves it to `retry_wait` and it runs again: the cancel is lost | state table cell nack × cancelled (8.4) | RP | E07 |
| D20 | the expiry sweep never dead-letters: only nack checks `max_attempts` (`internal/jobs/expire.go`) | after D03 is fixed, a crashing worker still re-queues past `max_attempts` | state table cell lease-expired × (attempts = max) (8.4); or Never/Always in E04 with the `crash` kind | RP | E07 (or E04) |
| D21 | job routes look up by id only, not by namespace (`internal/api/jobs.go`) | a `globex` key reads, cancels and acks `acme` jobs; ids are sequential and guessable | trust boundaries (9.2) | R | E08 |
| D22 | the background loop `return`s on the first database error instead of continuing | after `docker compose restart postgres`, no lease expires, no schedule fires, no webhook ships, until the api restarts; `/health` stays 200 | What If: the database goes away (9.3) | RP | E08 |
| D23 | `next_run_at` for a schedule is computed in a `time.FixedZone` captured at creation, not the IANA location | after a DST transition every fire is an hour off; `0 2 * * *` in `America/New_York` fires at 03:00 | What If: the clock (9.3); time/date heuristics (A2.1); needs the E02 test clock | RP | E08 |
| D24 *opt.* | the HMAC is computed over the indented JSON while the compact body is sent | every signature fails verification; hooksink only warns at v0 | Follow the Data across the boundary (9.4) | RP | E08 |
| D27 | the schedule firing query filters on `next_run_at <= $now` only; `PATCH … {enabled:false}` does not clear `next_run_at` (`internal/schedule/fire.go`) | a disabled schedule keeps enqueuing every tick | **predicted** from the pattern (a guard missing from a query, like D12/D14/D19) before being probed (13.3); or CRUD update in E06 if found early | R | E09 (or E06) |

Spec silences planted in `project/SPEC.md` (the 12.4 register): places where the spec does not say
and the code decided. Each is a hypothesis of the form "the spec is silent on X; I predict the code
does Y"; the probe finds out what it does; the fix is a spec line and the code change it implies,
decided by the coordinator in the fix commit's body (one line each, no meeting).

| ID | The spec says | It is silent on | v0 does | Probe | Expected in |
|---|---|---|---|---|---|
| A1 | "failed jobs retry with exponential backoff from `base_seconds`" | a cap | `base·2^n` uncapped: attempt 20 waits 60 days; overflow past 2^63 ns | Never/Always: "a retry always comes within some bound"; the clock walks attempts (5.1) | E04 |
| A2 | "purge removes all pending jobs" | leased jobs that later fail | re-queues them (nack → `retry_wait`) into a queue the operator believes empty | sequence: lease → purge → nack (6.1) | E05 |
| A4 | "terminal jobs are deleted after `retention_hours`" | whether `dead` is terminal for retention | deletes dead jobs too: the dead letter evaporates | Follow the Data: where does a dead job go; CRUD delete by the sweeper (7.3, 7.2) | E06 |
| A3 | "a cancelled job is never run" | a job already running when cancelled | marks `cancelled`; the worker runs on, blind; its ack 409s | state table cell cancel × leased (8.4) | E07 (with D19) |

Twenty-seven defects (four optional), four silences, three testability gaps, nine experiments. The
clustering is by probe, not by code area: that is the point. Eight rows share one code smell (a
missing state guard on an `UPDATE` or a query: D12, D14, D18, D19, D20, D27, A2, A3), which E09
names as the pattern and fixes structurally (Ch. 13.3).

Unplanned defects (real bugs the builder did not intend) will be found; they are recorded as `U##`
in `exploration/defects/`, fixed in the experiment that found them, and counted in its score.

## 5. The experiment arc

### 5.1 Shape of the history

```
explore-it/v0-inherited   explore-it: add queuelet, the inherited job queue, its spec and the exploration protocol (v0)
explore-it/lesson-01      queuelet: experiment 01, recon of the inherited codebase (lesson 01)
                          queuelet: fix D01, D02 revealed by experiment 01; add checks (lesson 01)
explore-it/lesson-02      queuelet: experiment 02, make the invisible visible (lesson 02)
                          queuelet: close T1–T3; fix D03, D04; add checks (lesson 02)
explore-it/lesson-03      … experiment 03, vary the inputs and the request shape / fix D05–D08, D26 …
explore-it/lesson-04      … experiment 04, write the oracles down, then break them / fix D09–D11, D25, A1 …
explore-it/lesson-05      … experiment 05, verbs in orders nobody intended / fix D12–D14, A2 …
explore-it/lesson-06      … experiment 06, entities, dependents and the data's journey / fix D15–D17, A4 …
explore-it/lesson-07      … experiment 07, the state table and its sneak paths / fix D18–D20, A3 …
explore-it/lesson-08      … experiment 08, the neighbours and the clock / fix D21–D24 …
explore-it/lesson-09      … experiment 09, the pattern: predict the next defect, then fix the cause / fix D27 + transition table …
explore-it/v1-explored    explore-it: lessons index, landing page, final skills
```

**Order.** Recon first (11.1, 12.4), because that is how anyone meets an inherited codebase and
because it produces the hypothesis backlog every later experiment draws from. Observation second
(Ch. 3), because its hooks are what make the later experiments possible in an hour instead of a day.
Then the book's own order: variations, oracles (Part I); sequences, entities, states, ecosystem
(Part II). Ch. 10 is practised in every experiment, since the SUT has no UI, and 10.4
(characterizing) is a rule of the protocol rather than a step. Last, the one Ch. 13 idea that is about
the code: read the findings for the pattern, predict from it, probe the prediction, fix the cause.

**The leaner arc in one line:** *map it → make it visible → vary it → judge it → sequence it → relate
it → state it → surround it → explain it.* Each verb is one experiment and one probe family.

**Tags** as above; history rewritable until Nick accepts `v1-explored`, then frozen.

### 5.2 The experiment protocol (`exploration/PROTOCOL.md`, written in checkpoint A)

1. **Charter as hypothesis** (2.2, 2.3). The book's template *Explore `<target>` with `<resources>`
   to discover `<information>`* plus two lines the book implies and this arc makes explicit:
   *I predict …* (what the explorer expects, and from what: spec, code, model, or a previous
   surprise) and *I would be wrong if …* (what observation refutes it). One target, named
   instruments (tools, data, hooks, models, the spec, the source), one kind of information. The
   experiment's charter is written by the coordinator from this plan; the explorer may add
   hypotheses to the backlog, never widen the one it holds.
2. **Budget** (1.3). An experiment is one headless agent run with a stated budget of 60 minutes wall
   clock, written in the charter; the elapsed time and its split (reading / probing / writing) goes
   in `revealed.md`.
3. **Blind to the register, not to the code.** The explorer runs from an export, not a checkout:
   `git archive <tag> books/explore-it/project books/explore-it/exploration books/explore-it/skills
   books/explore-it/lessons` into an empty directory with no `.git`, no `PLAN.md`, no `register/`.
   Its brief names the charter, the skill to follow, the protocol, and nothing about what is
   planted. It reads any file in the export and runs anything in the lab. The explorer and the
   seeder are different model families (section 8).
4. **Probe log** (11.5): `notes.md`, timestamped lines tagged `#read` (something learned from the
   source, with `file:line`), `#obs`, `#surprise`, `#bug`, `#question`, `#hypothesis`,
   `#offcharter`, with the exact commands run and their output trimmed. **Revealed**
   (`revealed.md`): charter status (*done* / *needs more* / *blocked*); what the system showed, in
   plain words; the model delta (which model files changed and how); the score table (hypotheses
   stated / confirmed / refuted; code-reading predictions contradicted by the running system;
   defects found; model cells filled); the next hypotheses. An experiment may end *needs more*; it
   then gets a second run with a narrowed charter. An experiment closes after at most two runs;
   planted defects still unfound are **disclosed** in the lesson note under "What the experiment
   missed, and why" and fixed in the same fix commit, flagged `disclosed`.
5. **Follow the surprise** (10.4, 11.5, 11.6). Every `#surprise` either becomes a defect report or a
   written reason it is not one. A defect report (`defects/D##-<slug>.md` or `U##`): title; tag seen
   at; minimal steps; expected, with the **oracle named** (the Never/Always rule, the comparable
   system, the approximation, or the spec line); observed; variables isolated (what did not
   matter); the **sibling check** (where else in the code the same cause could live, with a
   `file:line` or "none found"); severity; status (`open` → `fixed in lesson NN` or `won't fix:
   <why>` or `spec`). The 11.6 procedure for a surprise that will not reproduce: record the exact
   state, vary one thing at a time, suspect timing and the clock, keep the log. The explorer writes
   the report; the coordinator reconciles it with the register (same id if planted, a `U` id if
   not) and the fixer writes `lab/repro/D##.sh` from the steps.
6. **The hypothesis backlog** (`exploration/hypotheses.md`, 2.4, 2.6): every hypothesis ever
   proposed, with status (`open`, `experiment NN: confirmed|refuted|needs more`), source (recon,
   spec silence, code reading, model cell, surprise, risk headline), and the experiment it belongs
   to.
7. **The fix** is a separate agent (the fixer, same family as the seeder) with the defect reports
   as its brief: the smallest patch per defect, a regression check per defect under
   `project/test/checks/`, the repro script, the lesson note. It does not touch anything the reports
   do not name. For a spec silence it also writes the spec line and records the decision in the
   commit body.
8. **Living models** (`exploration/models/`: `map.md` (entry points, packages, where each verb
   lands), `nouns-verbs.md`, `variables.md`, `rules.md`, `entities.md`, `states.md`,
   `ecosystem.md`, and from E09 `causes.md`): later experiments revise them; the diff of a model
   between tags is part of the story.

### 5.3 Verification model (same for every experiment)

- `make up` starts the stack; `make seed` loads the fixtures; `make reset` restores the seed in
  seconds; `make down` removes everything.
- `make verify LESSON=NN` runs `lab/verify.sh NN`: reset; run every `lab/repro/D##.sh` the experiment
  names, each a black-box script through `qctl`/`curl`/read-only `psql` (and the test clock) that
  prints `PRESENT: D18 — stale ack accepted; job 1042 succeeded with worker-a's result while worker-b
  held the lease` or `ABSENT: D18 — stale ack rejected 409 stale_lease`; then `go test -tags checks
  ./test/checks -run 'L0NN'`. Verdict `FIXED AND CHECKED: D18, D19, D20 absent; checks 3/3` or
  `DEFECTS REPRODUCED: D18 present, D19 present, D20 present`. Exit 0 / 1. Runs in well under a
  minute: there is no load to apply.
- `make verify LESSON=NN REF=<tag>` runs the same scripts with the api, worker and CLI images built
  from `git archive <tag> project/` while the scripts come from the working tree. On `lesson-07`,
  `REF=explore-it/lesson-06` reproduces D18–D20. `REF=explore-it/v0-inherited` shows they were there
  from the start.
- **The dwell table.** `make dwell LESSON=NN` runs the experiment's repro scripts against every
  earlier tag and prints `defect × tag → PRESENT/ABSENT`. For E07 it reads: D18 present at v0, 01,
  02, 03, 04, 05, 06; absent at 07. The lesson note pastes it under "What the earlier experiments
  walked past" and says why (no hypothesis aimed there; the state model did not exist; the budget
  ran out). That table is this book's equivalent of `FLAW REPRODUCED` and the evidence that the
  probe, not luck, revealed the defect. Optional and slow (~1 minute per tag); the verdict never
  depends on it.
- `make checks` runs every regression check at any tag (no regressions across experiments);
  `make test` runs the unit tests. Lesson 00's verdict is `BASELINE HOLDS:` (smoke path works, unit
  tests pass, the thin v0 checks pass).
- Thresholds are binary (present/absent) and the fixer may not weaken a repro script to make a
  defect "absent"; the reviewer runs the previous tag's repro to confirm.

### 5.4 The experiments

Each experiment: source, hypothesis and probe, the planted defects it is expected to reveal, the two
commits, the verify, what it reveals about the earlier experiments, size. The hypothesis is the
charter in the protocol's form; the probe is the technique; "revealed" is what the record should
show when the run goes as planned; "model delta" is the learning.

#### E00 · The inherited system: a tour of v0 (no experiment)

- **Source:** Ch. 1.1–1.2; Ch. 11 opening (inheriting a system).
- **Contents:** `lessons/00-the-inherited-system.md`: what queuelet is, how to run it, how to read
  its spec, its tests and its source in an afternoon; the protocol in a page; **no register**. The
  reader is invited to explore blind before reading lessons 01–09; the note says the answer key is
  section 4 of this plan.
- **Verify:** `make verify LESSON=00` → `BASELINE HOLDS: enqueue→lease→ack→webhook in 2.1 s; unit
  tests 14/14; checks 3/3`.

#### E01 · Map it: recon of an inherited codebase

- **Source:** Ch. 11.1 "Starting with a Recon Session", 11.4 "Selecting Techniques"; Ch. 2.1–2.4,
  2.6 (the template, good charters, generating and planning charters); 12.4 "Active Reading"
  applied to `SPEC.md`.
- **Hypothesis:** *Explore queuelet end to end with the source tree, `qctl`, the README and `SPEC.md`
  to discover what it does, what it promises, where the spec is silent, and what I predict the code
  does there.* Prediction: "the README's happy path works and the spec is complete". Refuted if a
  documented verb misbehaves or a spec sentence has no code behind it. 60 minutes, split roughly
  20 reading / 30 probing / 10 writing.
- **Probe:** walk every `qctl` verb once against the seed while reading the handler it lands in;
  list nouns and verbs; read the spec sentence by sentence asking "what does this not say?" (12.4)
  and, for each silence, read the code to predict the answer; rank the open questions by what would
  hurt most if wrong (the one line kept from 2.5) and tag each with the probe that fits (11.4).
- **Expected finds:** D01, D02 (shallow by design: recon reveals what any first hour reveals).
- **Revealed:** the system in a page; a map of the code (`models/map.md`: which package owns which
  verb, where the background loop lives, where each `UPDATE` is); the spec's four silences as open
  hypotheses with the code's predicted answer (A1–A4 in the explorer's words, unknowingly); ~15
  hypotheses in the backlog, each with a probe family; two defects.
- **Model delta:** from nothing to `map.md`, `nouns-verbs.md` v1, `hypotheses.md` v1.
- **Fix commit:** the two patches (an `os.Exit(1)` path; a rows-affected check), two checks, two
  repro scripts, `lessons/01-map-it.md`.
- **Verify:** `make verify LESSON=01` → `FIXED AND CHECKED: D01, D02 absent`; `REF=v0-inherited` →
  `DEFECTS REPRODUCED`.
- **What it teaches:** not the two bugs: the backlog and the map. The note shows how "what happens
  to a running job when I cancel it?" (a spec silence found by reading) becomes E07's charter, and
  how the explorer's code-reading predictions for the four silences fare when E04–E07 probe them.
- **Size:** experiment ~250 lines of prose; fix ~40 lines.

#### E02 · Make it visible: consoles, logs, and the hooks the codebase lacks

- **Source:** Ch. 3.1 (the moonwalking bear), 3.2 "Digging Deeper", 3.3 "Testability and Making the
  Invisible Visible", 3.4 "Consoles and Logs".
- **Hypothesis:** *Explore the job lifecycle with the api and worker logs, `psql`, the source of the
  background loop and the `crash` and `fail` worker kinds to discover what happens to a job that the
  API does not show.* Prediction (from reading `expire.go` and `ship.go`): "a crashed worker's job
  retries up to `max_attempts` and a webhook delivery is marked delivered only on 2xx". Refuted if
  the consoles disagree with the API or with each other. 60 minutes.
- **Probe:** run the `crash` kind and watch three consoles at once (api log, `attempts` table, the
  worker's stdout); run the webhook with hooksink in `500` mode and compare the api log with
  hooksink's record; write down every question the consoles could not answer, as a **testability
  request** (3.3): the experiment's first product.
- **Expected finds:** D03 (a job that "ran once" four times; visible in `attempts` vs the worker's
  log), D04 (api log says delivered; hooksink says 500).
- **Revealed:** the lifecycle as the consoles show it, with the gaps marked; T1–T3 in the explorer's
  words plus whatever else it asks for.
- **Model delta:** `states.md` v0 (a sketch of the lifecycle as observed, incomplete on purpose);
  `map.md` gains the loop's five sweeps.
- **Fix commit:** T1 `log/slog` JSON with `job_id`, `queue`, `event`, `worker` on every
  transition; T2 `job_events` table and `GET /v1/jobs/{id}/events`, `qctl job events`, `qctl watch
  <q>`; T3 the `clock` package and `qctl clock set|advance|reset` under `QUEUELET_TEST_CLOCK=1`
  (a Compose profile sets it in the lab); D03 (count the attempt on expiry), D04 (check the status);
  checks; `lessons/02-make-it-visible.md`.
- **Verify:** `FIXED AND CHECKED: D03, D04 absent; checks 2/2; events endpoint shows
  lease→expired→queued→lease→ack for job 1017`. The REF run reproduces both and the note shows the
  v0 console next to the E02 console for the same job.
- **What it teaches:** E01 ran the `crash` worker and saw nothing wrong, because nothing was
  visible. Every later charter lists `qctl clock` or `job events` under instruments; the note argues
  from 3.3 that adding hooks to a codebase you are learning is part of learning it.
- **Size:** ~300 lines (slog ~60, events ~90, clock ~70, fixes ~30, checks ~50); the largest patch
  after v0; split into two commits (hooks; fixes) if it passes ~400.

#### E03 · Vary it: the variables of a job, a queue, and a request

- **Source:** Ch. 4.1–4.4; A2.1 (Some/None/All, Zero/One/Many, Beginning/Middle/End, Goldilocks,
  Boundaries, Count, Position, Selection, Strings, Numbers, Paths); 10.1 "Exploring an API" (does it
  reject what it does not understand?).
- **Hypothesis:** *Explore enqueue, batch enqueue and queue configuration with the variables list,
  the heuristics cheat sheet and raw `curl` requests to discover inputs the system accepts but cannot
  honour.* Prediction (from reading the validators in `internal/api`): "every documented range is
  enforced and unknown fields are rejected". Refuted by any 2xx the spec says should be a 4xx.
  60 minutes.
- **Probe:** list the variables first (`models/variables.md`: not just fields: counts, sizes,
  positions, timing, order, config, and the **shape of the request itself**: unknown fields, unknown
  enum values, wrong types), apply the named heuristics to each, and record a variation table
  (variable × value × observed). Use `curl`, because `qctl` only sends what the API expects.
- **Expected finds:** D05 (priority 100), D06 (batch 0 / 501), D07 (names), D08 (2^53), D26 (silent
  unknown fields and filter values).
- **Revealed:** the variation table; the places where the validator in the code is narrower than the
  spec and the places where there is none.
- **Model delta:** `variables.md` v1; `map.md` marks which handlers validate and which do not.
- **Fix commit:** validation (priority range, batch 1–500 with 400 beyond, name charset), payload
  kept as `json.RawMessage` and validated without decoding to `any` (D08), strict decoding
  (`DisallowUnknownFields`) on writes with a 400 naming the field, unknown enum values 400; checks;
  note. (The note contrasts this with Release It! L12's tolerant *reader*: strict about commands,
  tolerant about what you read from peers; both books are right about their own side.)
- **Verify:** `FIXED AND CHECKED: D05–D08, D26 absent`; the note's table shows how many variables the
  recon touched (a handful) versus the sweep (dozens).
- **What it teaches:** E01 enqueued a job with priority 5 and a batch of 3 through a CLI that only
  sends well-formed requests. The dwell table shows D05–D08 and D26 present at v0, 01, 02: not
  hidden, just never varied.
- **Size:** experiment ~300 lines; fix ~120 lines.

#### E04 · Judge it: write the oracles down, then break them

- **Source:** Ch. 5.1 "Never and Always", 5.2 "Alternative Resources", 5.3 "Approximations"; 10.3
  (status codes vs bodies).
- **Hypothesis:** *Explore retries, ordering, stats and error responses with a written list of
  Never/Always rules for a job queue, the spec, and SQS/pg-boss documentation as comparable
  systems, to discover where queuelet breaks rules nobody wrote down.* Prediction (from reading
  `nack.go`, `stats`, the lease query): "the code honours every rule on my list". Refuted by a
  single violated rule. 60 minutes, with `qctl clock` to walk backoff windows.
- **Probe:** write the rules **before** touching the system (`models/rules.md`: "a job is never
  leased by two workers at once", "attempts never exceed max_attempts", "within a priority, always
  FIFO", "stats always equal a direct count", "a retry always comes within a bounded delay", "an
  error is always a 4xx/5xx", "a cancelled job never runs" …), then try to break each: by reading the
  code for the line that enforces it, and by running the probe that would violate it. For numbers
  nobody checks (stats), compare with a direct `psql` count (5.3). For ordering, compare with what
  SQS FIFO promises (5.2).
- **Expected finds:** D09 (off-by-one), D10 (stats), D11 (same-second ordering), D25 (200 with an
  error body), A1 (uncapped backoff: the rule "a retry always comes within a bound" has no code and
  no spec line behind it). D20 is possible here if the explorer breaks "attempts never exceed max"
  with the `crash` kind.
- **Revealed:** the rules list with a verdict per rule and the code line (or absence) behind each.
- **Model delta:** `rules.md` v1, which becomes the invariant list the checks enforce from here on.
- **Fix commit:** `>=`; stats by exact state with age over `queued` only; `created_at
  timestamptz(6)` plus `id` as the tiebreaker in the lease order (a migration); proper status on
  lease errors; backoff cap `max_seconds` default 3600 with the spec line; checks that encode each
  rule as an invariant; note.
- **Verify:** `FIXED AND CHECKED: D09–D11, D25 absent; A1 resolved; invariants 6/6 hold over 200
  random jobs`.
- **What it teaches:** rule → check line by line. The dwell table shows D09 present through E03:
  experiments that varied inputs did not question outputs. And D09 is readable in one line of
  `nack.go`: the note shows that the explorer only saw it once a rule told it what to look for.
- **Size:** experiment ~250 lines; fix ~130 lines incl. migration.

#### E05 · Sequence it: verbs in orders nobody intended

- **Source:** Ch. 6.1 "Nouns and Verbs", 6.2 "Random Navigation".
- **Hypothesis:** *Explore operator and worker verbs in unusual orders with the nouns-and-verbs
  list, the test clock and a random sequence generator to discover sequences that leave jobs in
  states the spec does not allow.* Prediction (from reading the handlers): "every verb either does
  what it says from any starting point or is refused". Refuted by a sequence that leaves a job the
  spec would call impossible. 60 minutes. The experiment commit includes the little generator the
  explorer writes (`exploration/experiments/05-sequence-it/walk.sh`: pick a noun, pick a verb, act,
  observe, check the rules from E04).
- **Probe:** pairs and triples of verbs chosen for being unlikely (purge then enqueue-with-delay;
  pause, advance the clock an hour, resume; lease, purge, nack; cancel, retry); then the random walk
  for the rest of the budget, asserting `rules.md` after every step.
- **Expected finds:** D12 (purge misses delayed and retrying jobs), D13 (resume fires missed ticks),
  D14 (retry from cancelled/succeeded), A2 (nack after purge re-queues into an "empty" queue).
- **Revealed:** the sequences that broke a rule, each reduced to its shortest form (10.4).
- **Model delta:** `nouns-verbs.md` gains a verb × verb "tried / broke" grid; `states.md` gains its
  first sneak-path arrows.
- **Fix commit:** purge covers all non-leased pending states and leased jobs are marked
  `purge_requested` so a later nack dead-letters (spec line for A2); the scheduler advances
  `next_run_at` while paused (at most one catch-up); retry guarded to `dead`; checks (including a
  scripted random walk of 100 steps that asserts the rules after each step); note.
- **Verify:** `FIXED AND CHECKED: D12–D14 absent; A2 resolved; random walk 100/100 steps invariant`.
- **What it teaches:** every earlier experiment used the verbs in the order the README lists them.
  The note reprints the walk's log where step 37 (`purge`) followed by step 38 (`enqueue --delay`)
  and step 52 (`jobs ls`) shows the surviving job.
- **Size:** experiment ~300 lines incl. the walker; fix ~110 lines.

#### E06 · Relate it: entities, their dependents, and the data's journey

- **Source:** Ch. 7.1 "Recognizing Entities, Attributes, and Dependencies", 7.2 "CRUD", 7.3
  "Following the Data".
- **Hypothesis:** *Explore queues, schedules, workers and jobs with an entity diagram drawn from the
  migrations and a CRUD grid to discover what happens to dependents when their owner changes or
  disappears, and where a payload, result or error changes on its way through the system.*
  Prediction (from reading the migrations: no foreign keys with `ON DELETE`): "deleting an owner
  leaves orphans". Refuted if the code guards or cascades. 60 minutes.
- **Probe:** draw `entities.md` from the schema; fill the CRUD grid (entity × operation) with the
  observed outcome for each cell; for each dependency, delete or change the owner and read the
  dependent through every path (API, `psql`, the webhook body, the sweeper's log).
- **Expected finds:** D15 (delete queue orphans jobs and schedules), D16 (update schedule keeps the
  old `next_run_at`), D17 (truncation without a flag), A4 (the sweeper deletes `dead` jobs); D08 if
  E03 missed it; D27 is possible here under CRUD update on a schedule.
- **Revealed:** the CRUD grid with a verdict per cell; the journey of one payload and one error
  through every place they appear.
- **Model delta:** `entities.md` v1; `map.md` gains the "where a datum is copied or truncated"
  annotations.
- **Fix commit:** delete refuses while jobs or schedules exist (409 `queue_not_empty`) or cascades
  with `?force=true`; recompute `next_run_at` on update; truncation flags (`result_truncated`,
  `error_truncated`) or no truncation; `dead` exempt from retention with the spec line; checks;
  note.
- **Verify:** `FIXED AND CHECKED: D15–D17 absent; A4 resolved`.
- **What it teaches:** the CRUD grid has 16 cells; E01 touched 6. The note shows the grid before
  and after with each cell's verdict, and the prediction from the schema being right for jobs and
  wrong (in a different way) for schedules.
- **Size:** experiment ~250 lines; fix ~100 lines.

#### E07 · State it: the state table and its sneak paths

- **Source:** Ch. 8.1 "Recognizing States and Events", 8.2 "Diagramming a State Model", 8.3
  "Exploring with State Models", 8.4 "Changing the Representation: State Tables".
- **Hypothesis:** *Explore the job lifecycle with a state table built from the spec, the event
  history and the `WHERE` clauses in `internal/jobs`, two workers and the test clock, to discover
  transitions the table has no arrow for that the code allows anyway.* Prediction: for every cell
  the explorer writes what the code will do, from reading the guard (or its absence), before driving
  the cell. Refuted per cell. 60 minutes.
- **Probe:** draw the diagram, then the **state table** (every state × every event) with two columns
  per cell: "should" (from the spec) and "predicted" (from the code); then drive every cell marked
  "should be rejected" or "undefined", starting with those where the two columns disagree. The
  deep ones need two workers and the clock: lease with A, expire the lease, lease with B, ack with
  A's token.
- **Expected finds:** D18 (stale ack), D19 (nack after cancel), D20 (expiry past max), A3 (cancel
  while leased: the worker runs on, blind). D14 if E05 missed it.
- **Revealed:** the filled table, with the cells where the code contradicted the spec, the cells
  where the running system contradicted the code reading, and the cells nobody had driven before.
- **Model delta:** `states.md` v2: the table, now the authoritative model of the lifecycle.
- **Fix commit:** token check on ack/nack/heartbeat (409 `stale_lease`); state guard on nack; expiry
  path dead-letters at max; `cancel_requested` surfaced in the heartbeat response and honoured on
  ack (spec line for A3); checks driven from the state table (one test per rejected cell); note.
- **Verify:** `FIXED AND CHECKED: D18–D20 absent; A3 resolved; state table: 44 cells, 44 as
  expected`.
- **What it teaches:** the dwell table for D18 reads present at every tag from v0 to 06.
  Experiments 01–06 all acked jobs; none acked with an expired lease, because nothing had asked
  "what if this event arrives in that state". The note puts the table next to the diagram and shows
  which cells the earlier experiments had actually exercised (about a third), and that D18 was one
  `AND lease_token = $2` away from visible the whole time: reading finds it only once the table
  asks.
- **Size:** experiment ~350 lines (the table is big); fix ~130 lines.

#### E08 · Surround it: the neighbours, the boundary, and the clock

- **Source:** Ch. 9.1 "Diagramming the Ecosystem", 9.2 "Trust Boundaries", 9.3 "What If?", 9.4
  "Follow the Data".
- **Hypothesis:** *Explore queuelet's neighbours (Postgres, the workers, the webhook partner, the
  clock, the second namespace) with a context diagram, the hooksink admin modes, `docker compose
  restart` and the test clock, to discover what crosses a boundary unchecked and what happens when
  a neighbour misbehaves.* Prediction (from reading the auth middleware and the loop): "every
  route is namespace-scoped and the loop survives a database restart". Refuted per neighbour.
  60 minutes.
- **Probe:** draw `ecosystem.md` with every boundary marked; for each boundary, what crosses it and
  where in the code it is checked (`#read`); for each neighbour, one what-if (restart Postgres;
  hooksink in `500`, `slow`, `timeout`; the other tenant's key on this tenant's job ids; the clock
  across 2027-03-14 02:00 in `America/New_York`); follow one webhook body from the outbox row to
  hooksink's record and verify the signature by hand.
- **Expected finds:** D21 (namespace leak), D22 (loop dies), D23 (DST), D24 (signature).
- **Revealed:** the context diagram with a verdict per boundary and per what-if.
- **Model delta:** `ecosystem.md` v1; `rules.md` gains cross-boundary rules ("a key never sees
  another namespace's rows").
- **Fix commit:** namespace scoping on every job route; the loop logs and continues, `/health`
  reports `degraded` when a tick has failed; `next_run_at` through the IANA location each time;
  sign the bytes you send; checks (including a DST check that advances the clock across 2027-03-14
  in `America/New_York`); note.
- **Verify:** `FIXED AND CHECKED: D21–D24 absent; cron 0 2 * * * fires at 02:00 EST and 02:00 EDT
  across the transition`.
- **What it teaches:** the context diagram shows six boundaries; the earlier experiments had probed
  inside one box. The note links D21 to E01's top-ranked risk hypothesis ("another tenant acts on
  our jobs"), which sat open for seven experiments because no probe had aimed across the boundary.
  This is also where the repo's two books touch: the what-if list is a chaos-experiment list, and
  the note points at Release It!'s `design-chaos-experiment` skill for how to run them at load.
- **Size:** experiment ~300 lines; fix ~120 lines.

#### E09 · Explain it: the pattern, the prediction, and the cause

- **Source:** Ch. 13.3 "Discovering Systemic Sources of Problems"; 8.4 (the table as code); 13.5 in
  one line (the stop signal: findings per experiment).
- **Hypothesis:** *Explore the defect reports and models so far, and every `UPDATE` and firing
  query in `internal/jobs` and `internal/schedule`, to discover the small number of habits that
  produced the 26 findings and to predict, before probing, where the same habits put a defect
  nobody has found.* Prediction: the explorer writes `models/causes.md` (findings grouped by cause,
  with the code habit each group reveals) and a list of **predicted defects** with `file:line` and
  the probe that would confirm each, then spends the rest of the budget running those probes.
  Refuted if the predictions find nothing. 60 minutes.
- **Probe:** group the defects (expected: eight missing guards on an `UPDATE` or a query, five
  silent acceptances, three clock/time mistakes, two consoles that lie, the rest singletons); list
  every state-changing statement in the two packages with the guard it carries; predict the gaps;
  drive them.
- **Expected finds:** D27 (a disabled schedule keeps firing: the guard missing from the firing
  query) plus any `U##` the pattern predicts that the seeder did not plan; the note is honest about
  which.
- **Revealed:** `causes.md`; the predicted-versus-found table; the stop signal (new findings per
  experiment over E01–E09, expected to fall from 3–5 to 1–2 as the model fills in).
- **Model delta:** `causes.md` v1: the map of how the code was written, which is what an engineer
  who has learned a codebase actually carries in their head.
- **Fix commit:** the structural fix for the biggest group: `internal/jobs/transition.go`, one
  transition table (the E07 state table as code) that every handler goes through, with a
  **model-based check** that drives random event sequences against the table (the E05 walker, now
  a test); the schedule firing query guarded (D27); `skills/*/SKILL.md` final revisions from what
  the experiments taught; the lessons index and landing page; `v1-explored`.
- **Verify:** `FIXED AND CHECKED: all 27 defects absent at v1-explored; checks 30/30; model-based
  walk 1 000 steps clean`.
- **What it teaches:** nine experiments of findings reduce to a handful of causes, and the causes
  predict. The note shows the prediction being made from the pattern before the probe, which is the
  moment the explorer stopped exploring a black box and started reading a codebase it understood.
- **Size:** experiment ~300 lines; fix ~150 lines plus skills edits.

### 5.5 Expected find rates, honestly

A blind explorer will not reveal every planted defect in its experiment. The plan expects 60–80 % on
the first run and most of the rest on a second, narrowed run; the remainder are disclosed. The plan
also expects out-of-order finds (D08 in E06, D14 or D20 in E04, D27 in E06) and unplanned finds. All
three outcomes are recorded, not hidden, and the lesson notes carry a "Found / Missed / Found early /
Unplanned" table plus the score from `revealed.md`. The dwell table makes the teaching claim without
needing a perfect script: a defect that sat through six experiments and fell to the seventh's probe is
the lesson, however many others fell early. The one number the arc is expected to show rising is the
share of defects found by a **prediction from reading the code that the probe then confirmed**: low
in E01 (the reader does not yet know what to look for), high in E07 and E09.

## 6. Coverage honesty

| Chapter / section | Where | Note |
|---|---|---|
| 1.1–1.3 | E00, protocol | checked vs explored framed in E00; the budget in the protocol; the management side of 1.3 omitted |
| 1.4 Practice Sessions | omitted | |
| 2.1–2.4, 2.6 | E01 and every charter; `charter-as-probe` | the template read as a hypothesis with instruments |
| 2.5 Nightmare Headline Game | one line in E01 and the skill | risk hypotheses as a source; the group game omitted |
| 3 | E02 | hooks as a SUT patch, argued from 3.3 |
| 4, A2.1 | E03 | request shape added as a variable family (10.1) |
| 5 | E04 | rules → invariant checks; 10.3's status-vs-body as a rule |
| 6.1, 6.2 | E05 | random navigation as a script |
| 6.3 Personas | omitted | |
| 7 | E06 | |
| 8 | E07 (+ E09's transition table) | |
| 9 | E08 | touches Release It!'s chaos skill |
| 10.1, 10.3 | every experiment; E03, E04 | the SUT has no UI |
| 10.2 | omitted | |
| 10.4 Characterizing Bugs | protocol rule 5; `follow-the-surprise` | every report is characterized; no separate step |
| 11.1, 11.4 | E01 | |
| 11.2, 11.3 | omitted | |
| 11.5, 11.6 | protocol rules 4–5 | no planted defect is intermittent on purpose, but D11 (same-second ordering) behaves like one and the E04 note says so |
| 12.4 Active Reading | E01; a hypothesis source in every charter | the four spec silences are probed in E04–E07 |
| 12.1–12.3 | omitted | |
| 13.3 | E09 | |
| 13.1, 13.2, 13.4, 13.6, 13.7 | omitted | 13.5 is one line in E09 (the stop signal) |
| A1 | omitted | |
| A2.2 | omitted | web heuristics; API-only SUT |

Nothing of Part II is dropped. Part I loses only 1.4 and the game. Part III keeps what is about the
system and drops what is about the room.

## 7. Skills

Eight skills, each about exploring a codebase, written in checkpoint A from the book (drafts),
executed by the experiments, and revised in E09 from what the experiments taught. Each follows the
skills contract in `AGENTS.md` (when to use, inputs, procedure, output, source, worked example). The
worked example of each is an experiment in this repo; three also name litefleet as a second target to
run against once both arcs land. v1's `recon-existing-system`, `run-exploration-session`,
`report-defect` and `plan-exploration-strategy` are gone or absorbed: the first into the map skill,
the second into the protocol, the third into `follow-the-surprise`, the fourth entirely.

| Skill | Triggers | Procedure in one line | Source | Worked example |
|---|---|---|---|---|
| `map-codebase-from-behavior` | "we inherited this"; first day in a repo; "how does this thing actually work" | recon in a budget: entry points and packages, nouns and verbs, the spec read for silences with the code's answer predicted, a ranked hypothesis backlog; then keep the living models (`map`, `nouns-verbs`, `variables`, `rules`, `entities`, `states`, `ecosystem`) and, once there are findings, `causes` | 11.1, 11.4, 12.4, 2.4, 2.6, 13.3 | E01, `models/`, E09; second target: litefleet's tenant API |
| `charter-as-probe` | "what should I try first"; a feature or subsystem to learn; a risk to settle | write one experiment: target, instruments, information sought, *I predict*, *I'd be wrong if*, budget; check it is neither a tour nor a test case; pick the probe family that fits the question | 2.1–2.3, 2.5 (as a source), 1.3, 11.4 | every `experiments/NN/charter.md` |
| `make-invisible-visible` | "I can't tell what it did"; a system with thin logs; anything time-based | find the consoles the system already has (logs, the database, partner records, process output), read them side by side, list what none of them answers, add the smallest hooks (structured events, an event history, a controllable clock) | 3 | E02; second target: litefleet's placement decisions |
| `sweep-variables` | an input surface, a config, a command with options, a request body; "what values should I try" | list the variables (fields, counts, sizes, positions, timing, order, config, request shape), apply the named heuristics to each, record a variation table, use the rawest instrument available | 4, 10.1, A2.1 | E03 |
| `derive-oracles` | "how would I know it's wrong"; no spec; a number nobody checks | write Never/Always rules before probing, name comparable systems and standards, choose approximations, find the code line behind each rule, try to break each, turn the survivors into invariant checks | 5, 10.3 | E04 → `rules.md` → checks |
| `probe-sequences-and-states` | a lifecycle; verbs that can arrive in any order; anything with a status column | nouns × verbs in unlikely orders, a random walk asserting the rules, the state table with "should" and "predicted" per cell, drive every rejected or undefined cell | 6.1, 6.2, 8 | E05's walker, E07's table; second target: litefleet's database lifecycle |
| `probe-entities-and-boundaries` | entities with owners; a service with neighbours; a trust boundary; "what if X goes away" | the entity diagram and CRUD grid, dependents after owner changes, data followed through every place it appears, the context diagram with boundaries, what crosses each and where it is checked, one what-if per neighbour | 7, 9 | E06, E08 |
| `follow-the-surprise` | "that's odd"; a flaky report; "I saw it once"; a defect to write up | characterize: minimal reproduction, variables isolated, oracle named, the unreproducible procedure; then the sibling check (same cause elsewhere, by reading); then the durable forms: defect report, regression check, model delta | 10.4, 11.5, 11.6, 13.3 | `defects/D18`, E09's predictions |

Answer-key exercise: running `map-codebase-from-behavior` then the six probe skills blind against
`explore-it/v0-inherited` should reveal most of section 4 (the plan says 19 of 27 is the bar,
measured by the actual experiments: whatever the nine runs found without disclosure is the measured
number and goes in the skills index).

Overlap with the Release It! skills: `probe-entities-and-boundaries`'s what-if step and
`design-chaos-experiment` meet; the Explore It! skill produces the list, the Release It! skill runs
one at load. Both skill indexes cross-link.

## 8. Build order, seats and budget

Three checkpoints with cross-family review between them, as for Release It!. The difference is the
cast: three roles instead of one builder.

| Role | Does | Seat (proposal) | Why |
|---|---|---|---|
| Coordinator | writes charters from this plan, briefs the explorer and the fixer, reconciles defect reports with the register, commits experiment artifacts as produced, decides spec silences in one line each | the planner seat, interactive (Claude Code) | accumulated context is the value; it never explores and never fixes |
| Seeder / fixer | builds v0 with the register planted, writes the proof scripts, later writes each experiment's fix, checks, repro scripts and note | Codex Astra (`gpt-6-astra`), headless, fresh agent per unit | volume building; a different family from the explorer is the blindness guarantee; the Claude pool is on the Release It! build |
| Explorer | runs each experiment blind to the register, from the export | Claude Opus 5.5 headless (Muse Wonder Full as the fallback when the Claude pool is limited) | the skills are Claude Code / Codex format and are exercised by running them; the planning/research lean suits exploration |
| Reviewer | per checkpoint, from a fresh clone: the review bar in `AGENTS.md` plus the blindness check | Muse Wonder Full (`--model muse-wonder-full-1.0`, `--contributor-mode off`) | the third family; also the arbiter |

If Nick prefers Claude as the seeder (as for litefleet), swap seeder and explorer; the rule is only
that they differ.

**Checkpoint A — v0, the lab, the protocol, the skills drafts, E00 and E01 (~2 700–3 200 lines;
Release It! A was 3 500–4 000).**
- A1 *the system runs*: `project/` (api ~1 300 lines, worker ~250, qctl ~500, hooksink ~150,
  migrations, Compose, Makefile `up/down/seed/reset/verify/dwell/checks/test/logs`), `SPEC.md` with
  A1–A4 silent, unit tests, README. Plant all 27 defects where section 4 says; natural code, no
  markers. Proof scripts `books/explore-it/register/proof/D##.sh` (outside `project/`, never
  exported) that each demonstrate one defect; the reviewer runs all 27.
- A2 *the record and the kit*: `exploration/PROTOCOL.md`, empty `hypotheses.md`,
  `models/README.md`, `defects/README.md`; `lab/verify.sh` with `REF` builds, `lab/repro/`
  framework, `lab/dwell.sh`; eight `SKILL.md` drafts from the book (coordinator-reviewed for
  citations); `lessons/00`. Prove `make up && make seed && make verify LESSON=00`. Tag
  `v0-inherited`.
- Then E01 end to end (recon by the explorer, fix by the fixer) to prove the whole loop, including
  the export mechanism and the blind brief, before B starts. Tag `lesson-01`.

**Checkpoint B — E02 through E06 (five experiments, up to two second runs; ~560 lines of SUT
patches and checks, ~1 400 lines of record).** Per experiment: coordinator writes the charter →
explorer run → coordinator reconciles → (second run if *needs more* and a planted defect is unfound)
→ fixer → reviewer spot-checks the previous tag's repro. E02 is the largest patch and may be two
commits.

**Checkpoint C — E07 through E09, skills final, landing page (~400 lines of patches, ~950 of
record).**

**Budget, compared to v1 and to Release It!.** Code: about the same as v1 (the SUT is unchanged;
one more small defect; the fixes that moved from L09–L11 are now spread over E03, E04 and E09).
Prose and record: about 70 % of v1 (no interview, no decisions file, no strategy document, no
stakeholder debrief, no paired session; `revealed.md` is shorter than `debrief.md` and more useful).
Agent runs: ~11 explorer runs (9 + up to 2 repeats, down from ~14), ~10 fixer runs (down from ~12),
3 reviews, plus the coordinator's standing session. Against Release It!: code about 65 %, prose about
110 %, agent-hours a little under, spread across all three pools instead of one, with far less VM
load. Pool check before each checkpoint per the repo protocol; if the Claude pool is limited, Muse
explores; if two pools are limited, finish the in-flight experiment and stop.

**Rules for the seeder and the fixer.**
- Read `AGENTS.md`, this plan, then `PROTOCOL.md`. The register, the experiment order and the
  protocol are fixed; the fix for a defect is whatever the defect report and the spec say, not what
  the plan sketches, if they differ (record the deviation in the note).
- v0 passes its own tests and `make verify LESSON=00`. A planted defect that breaks the happy path
  is planted wrong. A planted defect must survive a careful read without a comment, a TODO or an
  odd name pointing at it: the explorer reads the code.
- Every SQL statement takes `$now` from the app clock; nothing uses Postgres `now()`.
- A fix commit touches only what its experiment's defect reports name. No defect is fixed early by
  accident (the reviewer runs later experiments' proof scripts at earlier tags).
- Repro scripts are black-box (API, CLI, read-only `psql`, test clock, `docker compose`); checks may
  use internals. A repro script is never weakened to turn PRESENT into ABSENT.
- The register is never written anywhere under `project/`, `exploration/`, `skills/` or `lessons/`
  until the experiment that discloses it.
- Commit subjects as in 5.1; the fix commit's body carries the defect ids, both verdict lines, the
  dwell summary and, for a spec silence, the one-line decision. Trailer per repo rules.

**Rules for the coordinator.**
- The explorer's brief contains: the export path, the charter, the skill name, `PROTOCOL.md`, the
  budget, and the instruction that it may read anything in the export and nothing outside it. It
  never contains hints about planted defects. The brief is committed with the experiment.
- Experiment artifacts are committed as produced; formatting only.
- An experiment closes after at most two runs. Disclosure is written in the lesson note, never into
  the experiment's own files.

## 9. Layout

```
books/explore-it/
  PLAN.md                           # this document (v2); PLAN-v1-sessions.md archived alongside if Nick wants the comparison
  README.md                         # landing page: argument, experiment index with tags, skills index
  register/                         # answer-key material outside the explorer's export
    proof/D01.sh … D27.sh           # the seeder's proof that each planted defect is live at v0
  lessons/
    README.md                       # index: experiment, tag, defects found/disclosed, score, verify command
    00-the-inherited-system.md      # tour of v0 without the register
    01-map-it.md … 09-explain-it.md
  exploration/                      # the record that grows: the explorer's model of the system
    PROTOCOL.md                     # charter-as-hypothesis template, budget, probe log, revealed, defect report
    hypotheses.md                   # the living hypothesis backlog
    models/                         # map.md nouns-verbs.md variables.md rules.md entities.md states.md ecosystem.md causes.md
    experiments/NN-<slug>/          # brief.md charter.md notes.md revealed.md (+ walk.sh)
    defects/                        # README.md (index with status) D##-<slug>.md U##-<slug>.md
  project/                          # queuelet: the inherited SUT
    README.md  SPEC.md  Makefile  go.mod  go.sum
    cmd/api/  cmd/worker/  cmd/qctl/  cmd/hooksink/
    internal/api/  internal/jobs/ (lease, ack, nack, expire, cancel, retry; transition.go from E09)
    internal/schedule/  internal/webhook/  internal/stats/  internal/store/  internal/clock/ (E02)
    migrations/  deploy/ (compose.yml, Dockerfiles)
    lab/ (verify.sh, repro/D##.sh, dwell.sh, reset.sh, seed/)
    test/checks/ (E01 onward)       results/ (gitignored)
  skills/
    README.md
    map-codebase-from-behavior/ charter-as-probe/ make-invisible-visible/ sweep-variables/
    derive-oracles/ probe-sequences-and-states/ probe-entities-and-boundaries/ follow-the-surprise/
```

The explorer's export is `project/`, `exploration/`, `skills/`, `lessons/` at the previous tag; never
`PLAN.md`, `register/`, or `.git`.

## 10. Suggested changes outside `books/explore-it/` (notes, not edits in this pass)

- **`AGENTS.md`**, "Purpose": add that for a book about a practice the archetype is a SUT plus a
  growing exploration record, and that a lesson's "patch" is an experiment plus the fix of what it
  revealed. "Lesson contract": allow the two-commit lesson (experiment, fix) with the tag on the
  fix, and the verdict words `FIXED AND CHECKED:` / `DEFECTS REPRODUCED:` beside `IMPROVED:` /
  `FLAW REPRODUCED:`. "Phase flow": allow skills to be drafted in phase 2's first checkpoint when
  the lessons execute them, with the final revision still at the end. Add the **blind-explorer
  rule** under the deliberate-issues register: when the lessons are experiments, the seeder and the
  explorer are different families and the explorer works from an export without the plan (and may
  read the code). "Layout of a book": add the optional `exploration/` and `register/` folders.
  "Review bar": add "every experiment's brief contains no hint of the register; every disclosed
  miss is in the lesson note; every planted defect survives a careful read without a marker".
- **Root `README.md`**: a row in the books table (*Explore It!*, Hendrickson, 2013; status: plan
  v2 awaiting confirmation; link to this plan); a sentence in the intro that for practice books the
  history shows an explorer's model of a system getting better, and the system with it; the
  reader's path gains `make dwell LESSON=NN`.
- **`STATE.md`**: an `explore-it` section: phase 1 plan v1 written 2026-10-22 by the planner seat
  (Claude Fable 5.1) on branch `plan/explore-it`; Nick scoped the book to probing, testing and
  learning a codebase and v2 was written the same day; awaiting Nick; phases 2–4 not started; seats
  as in section 8; "merge after the Release It! arc lands; adds-only".
- **`docs/knowledge/`** (does not exist yet): the export mechanism for blind agents and the
  two-commit lesson are process knowledge worth a page once they have run once.

## 11. Decisions Nick is asked to confirm

1. **The cut** (section 0): the people and process layer of the book is omitted (11.2–11.3,
   12.1–12.3, 6.3, 13.1–13.2 and 13.4–13.7, 2.5 as a game, 1.4, A1, A2.2, 10.2); what remains is the
   probing, testing and learning method, and the arc is nine experiments instead of eleven lessons.
2. **The shape** (section 1): steps as experiments (hypothesis → probe → revealed → model delta,
   scored), two commits per experiment, the SUT fixed as it goes, the dwell table as the payoff,
   skills drafted first and executed by the experiments.
3. **The explorer reads the code.** Blind to the register and the history, not to the source;
   every hypothesis names its source and the record scores reading-based predictions against the
   running system. (v1 was silent on this; the export already contained the source.)
4. **The cast and seats**: coordinator / seeder-fixer / explorer / reviewer with Codex seeding and
   fixing, Claude exploring, Muse reviewing, or the swap.
5. **The vehicle**: `queuelet`, a fresh baby job queue (Go, Postgres 17, a webhook partner fake, a
   test clock), over reusing litefleet (kept as a second worked example for three skills), the
   workflow platform and the deploy product. The name is a placeholder.
6. **The register** at 27 defects (4 optional), 4 spec silences re-homed into E04–E07, 3
   testability gaps, and the nine experiments in the stated order, with E09 as a prediction
   experiment rather than a paired session.
7. **The skills**: the eight in section 7, all about exploring a codebase; the four people-process
   skills of v1 dropped or absorbed.
8. **Checkpoints** A (v0 + record + skills drafts + E00 + E01), B (E02–E06), C (E07–E09) with
   cross-family review; the budget framing in section 8.
9. **Isolation**: the branch merges after the Release It! arc lands; the build may start earlier in
   its own worktree since everything is under `books/explore-it/`; only the section 10 edits wait.

---

## Changelog v1 → v2 (2026-10-22, Nick's scoping)

- **Cut** L10 (requirements meeting) and L11 (paired session, strategy, estimate, stakeholder
  debrief); the four spec silences moved into E04 (A1), E05 (A2), E06 (A4), E07 (A3); the systemic
  analysis (13.3) became E09, now a prediction experiment with a planted target (D27).
- **Cut** L09 as a step: D25 moved to E04 (an oracle: errors are always 4xx/5xx), D26 to E03
  (request shape as a variable family); 10.4 characterizing became a protocol rule.
- **Cut** from L01 the stakeholder interview and the headline game as a ritual; 12.4 active reading
  of the spec moved in.
- **Cut** personas from E05.
- **Recast** every lesson as an experiment with a hypothesis, a prediction, a refutation condition,
  "revealed" and a model delta; `debrief.md` became `revealed.md` with a score table; `charters.md`
  became `hypotheses.md`; `models/` gained `map.md` and `causes.md`.
- **Made explicit** that the explorer reads the code; added the **Read?** column to the register and
  the "reading predictions contradicted by the running system" score.
- **Added** D27 (disabled schedule keeps firing) as E09's predicted target.
- **Re-cut** the skills from eight people-and-process-leaning skills to eight codebase-exploration
  skills; four of v1's gone or absorbed.
- **Recomputed** the budget: 9 experiments, ~11 explorer runs, ~10 fixer runs, record ~70 % of v1.
