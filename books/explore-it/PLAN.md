# Explore It! — breakdown plan v1: one inherited job queue, explored session by session

**Book:** *Explore It! Reduce Risk and Increase Confidence with Exploratory Testing*, Elisabeth
Hendrickson, Pragmatic Bookshelf, 2013 (first edition, the only one; 186 pages, ISBN
978-1-937785-02-4). Chapter and section numbers below are from that edition's table of contents
(verified against the publisher's preview).

**Phase:** plan (v1), for Nick's confirmation. Design only: no SUT code, no sessions, no skills exist
yet. Written on branch `plan/explore-it` in its own worktree while the Release It! build runs on
`main`; this pass adds files under `books/explore-it/` only. Suggested edits to `AGENTS.md`, the root
`README.md` and `STATE.md` are listed in section 9 as notes, to be applied when the branch merges.

---

## 0. The shape: how a book about a practice fits a repo built for books about systems

`AGENTS.md` was written for *Release It!*: the book names flaws in a system, so the archetype is a
flawed system and each lesson is a patch that fixes a flaw, with a verify script that reproduces the
flaw on the previous tag. *Explore It!* names no flaws in systems. It names skills in people: how to
charter, observe, vary, model, evaluate and report so that a system's surprises come out before
production does. The thing that gets better lesson by lesson is therefore **the exploration**, not the
system. The honest adaptation keeps every structural element of the playbook and changes what each
one holds:

| Playbook element | Release It! | Explore It! (this plan) |
|---|---|---|
| The archetype | one flawed service, improved by patches | one **inherited** service (the system under test, SUT) seeded with realistic defects, plus the **exploration record** that grows next to it (`exploration/`: charters, sessions, models, defect reports) |
| What a lesson is | one technique of the book applied as a fix | one technique of the book applied as a **time-boxed exploratory session** against the SUT, charter first; then the defects it found are **fixed and turned into checks** |
| The lesson commit | patch + note + verify script | **two commits**: the session (charter, notes, debrief, models, defect reports, reproduction scripts) and the fix (SUT patch + regression checks + note); the lesson tag points at the fix |
| The register | flaws → chapter → lesson | defects → **technique that finds them** → chapter → lesson; plus a register of spec ambiguities (Ch. 12) and testability gaps (Ch. 3) |
| The verify | load + chaos + thresholds; `IMPROVED:` / `FLAW REPRODUCED:` | the session's black-box **reproduction scripts** plus the regression checks; `FIXED AND CHECKED:` / `DEFECTS REPRODUCED:`. `REF=<earlier tag>` shows the defect was present all along: the **dwell table** (which tags the defect sat in while earlier sessions ran past it) is the payoff |
| "Skill compounding" in `git log` | the system hardens | early sessions find shallow bugs; later, better-charted and better-modelled sessions find the deep ones (stale acks, lost cancels, DST drift, a scheduler that silently dies) that the earlier sessions walked past |
| Skills | written after the lessons, from the lessons | written **before** the sessions, from the book, and **executed** by the sessions; each session is the skill's worked example and sends corrections back (Ch. 13.7, capturing nuggets) |

Two rules make the record honest rather than theatre:

1. **The explorer is blind.** The agent that runs a session never sees this plan's register, the
   proof scripts, or the SUT's git history. It works from a plain export of the SUT tree, the
   exploration record so far, the skill for the lesson and its charter. Defects it finds are real
   finds; defects it misses are recorded as misses in the debrief and disclosed in the lesson note.
   The seeder (who plants the register) and the explorer are agents from **different model
   families**, which the repo's cross-family rule already requires for review.
2. **The session artifacts are committed as produced.** Notes keep their timestamps and their dead
   ends. The coordinator formats, never rewrites. A reader should be able to see a session go wrong.

Everything else in `AGENTS.md` (one archetype, real technology, `make up / make verify LESSON=NN
[REF=tag]`, lesson notes outside `project/`, tags, the review bar, commit style) applies unchanged.

## 1. The book

### 1.1 Core argument

Testing is two activities, and most teams only do one. **Checking** confirms the software does what
you already decided it should; it is designed in advance and is what test cases and automated suites
hold. **Exploring** is designing experiments, running them and learning from the result in the same
motion, so that each observation shapes the next experiment (Ch. 1.1 "Two Sides of Testing", 1.2
"Essential Elements of Exploratory Testing"). Checks catch regressions in the risks you named;
exploration finds the risks you did not name. *Tested = checked + explored* (Ch. 1.1). The book's
program: give exploration a **charter** so it is not wandering (Ch. 2); learn to **observe** what is
actually happening, including what the interface hides (Ch. 3); find the **variables** and vary
them deliberately (Ch. 4); decide what "correct" means when no one wrote it down, with Never/Always
rules, alternative resources and approximations (Ch. 5); use **models** of sequences, entities,
states and the surrounding ecosystem as generators of experiments (Ch. 6–9); apply all of it to
APIs, legacy systems and requirements (Ch. 10–12); and make exploration a planned, estimated, debriefed
part of the team's strategy, feeding what it finds back into checks (Ch. 13).

### 1.2 Part and chapter map

| Part / chapter | Sections | What it gives the practitioner |
|---|---|---|
| **I. Establishing Foundations** | | |
| 1. On Testing and Exploration | 1.1 Two Sides of Testing; 1.2 Essential Elements of Exploratory Testing; 1.3 Working in Time-Boxed Sessions; 1.4 Practice Sessions | checked vs explored; exploration as simultaneous learning, design and execution; the session as the unit of work |
| 2. Charter Your Explorations | 2.1 Chartering the Expedition; 2.2 A Simple Charter Template; 2.3 Good Charters; 2.4 Generating Charters; 2.5 The Nightmare Headline Game; 2.6 Planning Charters | the template *Explore `<target>` with `<resources>` to discover `<information>`*; what makes a charter too broad or too narrow; sources of charters; the headline game for risk-driven charters; a backlog of charters |
| 3. Observe the Details | 3.1 But Did You See the Moonwalking Bear?; 3.2 Digging Deeper; 3.3 Testability and Making the Invisible Visible; 3.4 Consoles and Logs | inattentional blindness; looking past the interface; asking for testability hooks; reading the consoles the system already has |
| 4. Find Interesting Variations | 4.1 Variables Are Things That Vary; 4.2 Subtle Variables, Big Disasters; 4.3 Identifying Variables; 4.4 Variables! They're Everywhere! | variables beyond input fields: counts, sizes, positions, timing, sequences, configuration; the named heuristics (Some/None/All; Zero/One/Many; Beginning/Middle/End; Goldilocks; and the rest of Appendix 2) |
| 5. Evaluate Results | 5.1 Never and Always; 5.2 Alternative Resources; 5.3 Approximations | oracles when there is no spec: rules the system must never/always break, comparable systems and standards, "is it roughly right" |
| **II. Adding Dimensions** | | |
| 6. Vary Sequences and Interactions | 6.1 Nouns and Verbs; 6.2 Random Navigation; 6.3 Personas | listing the system's nouns and verbs and combining them at random; walking randomly; exploring as a particular kind of user |
| 7. Explore Entities and Their Relationships | 7.1 Recognizing Entities, Attributes, and Dependencies; 7.2 CRUD; 7.3 Following the Data | the entity model as a generator: every entity × create/read/update/delete, every dependency × "what happens to the dependent", data followed through every place it appears |
| 8. Discover States and Transitions | 8.1 Recognizing States and Events; 8.2 Diagramming a State Model; 8.3 Exploring with State Models; 8.4 Changing the Representation: State Tables | the state model as a generator: every state × every event, including the events the diagram has no arrow for (the sneak paths) |
| 9. Explore the Ecosystem | 9.1 Diagramming the Ecosystem; 9.2 Trust Boundaries; 9.3 What If?; 9.4 Follow the Data | the context diagram; what crosses a trust boundary and whether it is checked; "what if this neighbour misbehaves"; data across the boundary |
| **III. Putting It in Context** | | |
| 10. Explore When There Is No User Interface | 10.1 Exploring an API; 10.2 Exploring a Programming Language; 10.3 Exploring a Web Service; 10.4 Characterizing Bugs | exploring through tools instead of screens; isolating a bug to its minimal reproduction and the variables that matter |
| 11. Explore an Existing System | 11.1 Starting with a Recon Session; 11.2 Sharing Observations; 11.3 Interviewing Stakeholders to Gather Questions; 11.4 Selecting Techniques; 11.5 Recording Your Findings; 11.6 What About the Dreaded Unreproducible Bugs? | the recon session for a system you inherited; questions as the first product; which technique fits which question; how findings are recorded |
| 12. Explore Requirements | 12.1 Getting Into the Requirements Meeting; 12.2 In the Requirements Meeting; 12.3 Charter During Requirements Discussions; 12.4 Active Reading | exploring the requirements before (and instead of) the code; charters from ambiguities; reading a spec adversarially |
| 13. Integrate Exploration Throughout | 13.1 Exploring as Part of Your Test Strategy; 13.2 Paired Exploration; 13.3 Discovering Systemic Sources of Problems; 13.4 Estimating Exploration; 13.5 How to Tell When You Have Explored Enough; 13.6 Debriefing Stakeholders; 13.7 Capturing Useful Nuggets of Wisdom | the checked+explored strategy; pairing; reading the defect list for causes, not symptoms; estimating and stopping; the debrief; the team's growing heuristic list |
| A1. Interviewing for Exploratory Testing Skills | A1.1, A1.2 | out of scope here (hiring) |
| A2. Test Heuristics Cheat Sheet | A2.1 General Heuristics; A2.2 Web Heuristics | the named heuristics collected; A2.1 is the kit the sessions draw from |

### 1.3 What a practitioner does differently after reading it

- Writes **charters**, not test cases, and keeps a backlog of them; plans by charter and session.
- Works in **time-boxed sessions** with notes and a **debrief** (charter status, bugs, questions, new
  charters), not in open-ended "testing".
- Treats every observation as data: reads logs and consoles alongside the interface and **asks for
  testability** hooks when the system hides what it is doing.
- Names the **variables** of a feature before touching it and varies them with named heuristics
  instead of trying "a few values".
- Makes **oracles explicit**: for each feature, the Never/Always rules, the comparable system, the
  approximation used to judge the result.
- Draws **models** (nouns and verbs, entities and CRUD, states and events, the ecosystem with its trust
  boundaries) because the model's blank cells are where the bugs are.
- Reports bugs **characterized**: minimal reproduction, the variables that matter, the oracle violated.
- Explores the **requirements** before the code exists, and turns ambiguities into charters.
- Plans exploration as part of the strategy, **estimates** it, knows when to stop, and feeds findings
  back into **checks** and into the team's heuristics.

The vocabulary used throughout this plan is Hendrickson's (charters, sessions, heuristics, models).
"Tours" are Whittaker's (*Exploratory Software Testing*, 2009) and are not used.

## 2. The vehicle: `queuelet`, an inherited baby job queue, and the exploration record beside it

### 2.0 Why this vehicle

What an exploratory-testing vehicle needs is different from what a stability vehicle needs. It must
have **many states and events** (Ch. 8), **entities with dependencies** (Ch. 7), **many nouns and
verbs** that combine (Ch. 6), **variables of every kind** including time (Ch. 4), an **ecosystem with
a trust boundary** (Ch. 9), **no UI** (Ch. 10) and a spec worth reading adversarially (Ch. 12). It must
be small enough that a session can cover a charter in an hour, and seeded defects must be
**semantic** (wrong state, lost cancel, off-by-one, precision loss, wrong timezone), the kind a
person finds by thinking, not by load.

| Candidate | Fit | Why it did or did not win |
|---|---|---|
| **Reuse `litefleet`** (the Release It! SUT) | medium | Real, in Nick's domain, already built. But its register is stability-shaped: its defects are found by load and fault injection, which is Release It!'s harness, not by hands-on exploration. Its tenant-facing surface is a SQL passthrough with few verbs and almost no state of its own (the interesting state is inside SQLite). Seeding it with semantic defects would mean editing the Release It! codebase during or after that arc, and Explore It! tags would depend on `release-it/v1-hardened`, which is still being rewritten. Litefleet stays available as a **second worked example** for two skills once both arcs land (section 6). |
| **A baby durable job queue** | near 1:1 | Nick's second domain. A job has eight states and a dozen events, so sneak paths are everywhere (ack after lease expiry, cancel while running, retry a succeeded job). Entities with real dependencies: queues own jobs and schedules, workers hold leases, schedules enqueue jobs. Verbs galore for Nouns and Verbs. Variables of every kind: payload size and numeric precision, priority, delay, batch count, cron expressions with time zones (the time/date heuristics finally bite). An ecosystem: Postgres, worker processes, a webhook partner across a trust boundary, and the clock. No UI: an HTTP API and a CLI, which is Ch. 10 verbatim. And a queue is the substrate of the workflow platform and of a deploy pipeline, so it is the smallest thing that still looks like what Nick builds. |
| Workflow platform | good but large | Everything fits, but the minimum (definitions, steps, timers, history) is 2–3× the budget before a single defect is planted; the queue is its kernel. |
| Baby Vercel product (deployments) | medium | Deployments have states, but a build pipeline needs real work to run and most of the richness (builders, artifacts, domains) is cost without exploration payoff. A deploy queue is, again, a job queue. |

### 2.1 What it does

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

### 2.2 API (v0)

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
is the systemic source lesson 11 finds (Ch. 13.3).

A background loop in the service, once per second: promote `scheduled` → `queued`; expire leases;
fire due schedules; sweep retention; ship the webhook outbox.

**CLI** `qctl` (`queue create|ls|get|pause|resume|purge|rm`, `enqueue`, `jobs ls`, `job get|attempts|
cancel|retry|rm`, `worker run <q> --kinds …` (the sample worker), `schedule …`, `stats`; from lesson 02
also `watch` and `clock`). It is the explorer's main instrument (Ch. 10.1, 10.3).

### 2.3 Data model

Postgres 17, hand-written SQL: `namespaces(id, name)`, `api_keys(key_hash, namespace_id, admin)`,
`queues(id, namespace_id, name, lease_seconds, max_attempts, backoff_kind, backoff_base, backoff_max,
retention_hours, webhook_url, webhook_secret, paused, unique(namespace_id, name))`, `jobs(id bigint,
queue_id, kind, payload jsonb, priority, state, run_at, attempts, max_attempts, idempotency_key,
leased_by, lease_token, lease_expires_at, result jsonb, last_error, created_at timestamp(0),
finished_at)`, `attempts(job_id, n, worker_id, started_at, finished_at, outcome, error)`,
`workers(id, namespace_id, name, kinds, token, last_seen_at)`, `schedules(id, namespace_id, queue_id,
name, kind, payload, cron, timezone, enabled, next_run_at, last_run_at)`, `webhook_outbox(id, job_id,
url, body, attempts, next_at, delivered_at)`, `job_events(job_id, at, event, detail)` (added in
lesson 02). Several of these columns are where the register lives (section 3).

**Seed** (`make seed`, seconds): namespaces `acme` and `globex` with keys, queues `emails`,
`reports`, `imports` with different lease/retry settings, one webhook on `reports` pointing at the
hook sink, 300 jobs in mixed states, two schedules (one in `America/New_York`), and workers
`worker-a` and `worker-b` running. `results/manifest.json` lists keys and ids.

### 2.4 Topology and stack (Docker Compose)

```
 qctl / curl / explorer ──► api :18200 ──► postgres :15433
 worker-a, worker-b (cmd/worker: kinds echo, sleep, fail{n}, crash, http) ──► api  (lease/heartbeat/ack/nack)
 api ──► hooksink :18201 (webhook partner fake: records deliveries, verifies HMAC, /admin/mode: ok|500|slow|timeout)
 worker 'http' kind ──► hooksink (an outbound call across the boundary)
 test clock (from lesson 02): api and workers read `clock.Now()`; `qctl clock set|advance|reset` under QUEUELET_TEST_CLOCK=1
```

| Choice | Pick | Why |
|---|---|---|
| Language | **Go 1.26**, one module (`cmd/api`, `cmd/worker`, `cmd/qctl`, `cmd/hooksink`) | Same toolchain as litefleet, so the builder's and reviewer's setup carries over; static binaries, fast Compose builds; `encoding/json`'s `float64` default is itself one of the planted defects (numeric precision), which is realistic Go. |
| Store | **Postgres 17** (`postgres:17-alpine`, cached on this VM), `FOR UPDATE SKIP LOCKED` leasing | The canonical real-stack job queue (River, Graphile Worker, pg-boss, Oban all do this). `timestamp` vs `timestamptz`, `timestamp(0)` precision and `jsonb` normalization are real places for defects. |
| Cron | `robfig/cron/v3` parser with IANA locations via `time.LoadLocation` | Real cron semantics; the DST defect is in how v0 *uses* it, not in the library. |
| Partner fake | own `hooksink` (~150 lines) | Records every delivery with body and signature verdict; `/admin/mode` makes it misbehave (500, slow, timeout) for the Ch. 9 what-ifs. No Toxiproxy: the what-ifs this book asks are "the partner answers badly" and "Postgres restarts", which `/admin/mode` and `docker compose restart` cover. |
| Clock | an app-level `clock` package; `QUEUELET_TEST_CLOCK=1` enables `POST /admin/clock` | Lesson 02's testability hook (Ch. 3.3), used by lessons 04, 05, 06, 08 and the checks. Every SQL statement takes `$now` from the app; nothing uses Postgres `now()`, so the test clock is the only clock. |
| Observability | `log/slog` JSON from lesson 02 (v0: sparse `log.Printf`); `GET /v1/jobs/{id}/events` from lesson 02; a minimal `/metrics` | The book's "consoles and logs" (3.4); not Prometheus-scale: there is no load story here. |
| Explorer tooling | `qctl`, `curl`, `jq`, `psql` (read-only role `explorer`), `docker compose logs` | The book's point in Ch. 10: no UI means the tools are the interface. The read-only `psql` role is a resource the charters name (Ch. 2.2). |
| Checks | Go tests under `project/test/checks/` (build tag `checks`, run against the running stack with the test clock) and unit tests | Exploration finds, checks keep (Ch. 13.1). Reproduction scripts (`lab/repro/D##.sh`) are black-box and separate from checks. |

Host ports in the 1xxxx range (18200 api, 18201 hooksink, 15433 Postgres) because 5432 and 8080 are
taken on this VM.

### 2.5 What v0 does right (so the defects are believable)

Clean package layout; `SKIP LOCKED` leasing that works; retries with backoff; dead-lettering on
nack; a webhook outbox with retries; cron schedules; keyset paging; a CLI that covers the API; unit
tests for the lease query and the backoff math; a README; a spec (`project/SPEC.md`, ~120 lines,
written as a competent team writes one: mostly right, silent in four places). The happy path
(enqueue → lease → ack → webhook) works and `make verify LESSON=00` proves it.

### 2.6 Testability gaps at v0 (the Ch. 3 register)

| ID | Gap | Why it matters | Closed in |
|---|---|---|---|
| T1 | Logs are sparse `log.Printf` lines without job ids or state transitions | you cannot tell from the console what a job did | L02 |
| T2 | No per-job event history; `GET /v1/jobs/{id}` shows only the current state | a job that ran twice looks like a job that ran once | L02 |
| T3 | No clock control; lease expiry, backoff, retention and cron can only be explored by waiting | a one-hour backoff or a DST transition cannot be explored in a session | L02 |

Closing these is a legitimate SUT patch (Ch. 3.3 "Testability and Making the Invisible Visible" is
about asking for exactly this), and the hooks are the resources later charters name.

## 3. The deliberate-defect register

Every defect planted in v0, where it lives, how it shows, which technique and chapter finds it, and
the lesson expected to find it. This is the answer key for the `charter-exploration`,
`sweep-variations` and `model-to-explore` skills (section 6) and the only list of the defects. The
explorer never sees it (section 4.2). "Expected in" is a prediction, not a rule: a defect found early
is fixed early and the later lesson's charter aims at what is left; a defect missed is disclosed in
the lesson note (section 4.2, rule 4). Defects marked *opt.* may be dropped by the builder if
checkpoint A runs long.

| ID | Defect in v0 (file / area) | How it shows | Technique that finds it (chapter) | Expected in |
|---|---|---|---|---|
| D01 | `qctl` prints API errors but exits 0 (`cmd/qctl`) | a script cannot tell a failed enqueue from a successful one | recon: first hour with the CLI (11.1) | L01 |
| D02 | `DELETE /v1/queues/{q}` returns 204 when no such queue exists (rows-affected unchecked) | typos in a delete script "succeed" | recon; CRUD on the obvious nouns (11.1, 7.2) | L01 |
| D03 | lease expiry re-queues the job but neither increments `attempts` nor writes an attempt row (`internal/jobs/expire.go`) | a worker that crashes retries forever; `max_attempts` never reached; the job looks like it ran once | consoles and the event history once T1/T2 exist; the `crash` worker kind (3.3, 3.4) | L02 |
| D04 | the webhook shipper marks a delivery `delivered` on any HTTP response; status is logged, not checked (`internal/webhook/ship.go`) | logs say delivered, hooksink says 500 | two consoles disagree (3.4) | L02 |
| D05 | `priority` outside 0–9 is accepted and sorted (`-5`, `100`) | one job with priority 100 pre-empts every queue forever | Goldilocks / boundaries on a numeric variable (4.3, A2) | L03 |
| D06 | `jobs:batch` with 501+ jobs silently inserts the first 500 and returns 201; a batch of 0 returns 201 `{ids: []}` (spec: 400) | producers lose jobs without an error | Zero, One, Many / Too Many on a count (4.3, A2) | L03 |
| D07 *opt.* | queue names with `/` or spaces are accepted at create but cannot be addressed by any later route | a queue exists that nothing can reach | strings variation: special characters (4.3, A2) | L03 |
| D08 | payload is decoded into `map[string]any` for validation and re-encoded, so integers above 2^53 lose precision (`9007199254740993` → `…992`, `2^63-1` → `9.223372036854776e18`) (`internal/api/enqueue.go`) | a worker receives a different id than the producer sent | numbers at 2^53 / 2^63 boundaries (4.3); or Follow the Data (7.3) if missed | L03 (or L06) |
| D09 | nack's retry decision uses `attempts > max_attempts` after incrementing (`internal/jobs/nack.go`) | `max_attempts: 3` runs four times | Never/Always: "attempts never exceed max_attempts" (5.1) | L04 |
| D10 | `stats.queued` counts `scheduled` and `retry_wait` rows too; `oldest_queued_age` goes negative for future `run_at` and is clamped to 0 (`internal/stats`) | the dashboard number disagrees with `SELECT count(*)` | approximations: "is this number roughly right?" checked against a direct count (5.3) | L04 |
| D11 | `created_at` is `timestamp(0)`; the lease query orders by `priority DESC, created_at` | jobs enqueued within the same second lease in arbitrary order; spec promises FIFO within a priority | Never/Always: "always FIFO within a priority"; alternative resource: the spec and SQS FIFO docs (5.1, 5.2) | L04 |
| D12 | `purge` deletes only `queued` rows; `scheduled` and `retry_wait` survive (spec: "all pending jobs") | enqueue-with-delay → purge → the job still runs | Nouns and Verbs in odd orders (6.1) | L05 |
| D13 | while a queue is paused the scheduler skips cron ticks; on resume it fires every missed tick at once | a one-hour pause of a `*/5` schedule enqueues twelve jobs on resume | persona: the operator who pauses for maintenance (6.3) | L05 |
| D14 | `POST /v1/jobs/{id}/retry` has no state guard: allowed from `cancelled` and `succeeded` (spec: from `dead` only) | a cancelled or finished job runs again | random sequences (6.1, 6.2); or the state table (8.4) if missed | L05 (or L07) |
| D15 | `DELETE /v1/queues/{q}` deletes the queue row only: jobs and schedules stay (no cascade, no guard) | orphan jobs remain leased or queued; namespace stats count them; the schedule keeps failing every tick | dependencies: "what happens to the dependents" (7.1), CRUD delete (7.2) | L06 |
| D16 *opt.* | `PATCH /v1/schedules/{id}` changing `cron` or `timezone` does not recompute `next_run_at` | the old time fires once more | CRUD update on the schedule entity (7.2) | L06 |
| D17 *opt.* | the webhook body truncates `result` to 1 KB and the dead-letter `last_error` to 256 chars with no truncation flag; `GET /v1/jobs/{id}` returns them whole | the same datum differs by the path you read it through | Follow the Data (7.3) | L06 |
| D18 | `ack` checks `state = 'leased'` but not `lease_token` (`internal/jobs/ack.go`) | worker A's lease expires, B leases, A's late ack succeeds, B's ack gets 409 and B's result is lost: the job "succeeded" with A's stale result | sneak path: ack × leased-by-another (8.3) | L07 |
| D19 | `nack` has no state guard | cancel a leased job; the worker's later nack moves it to `retry_wait` and it runs again: the cancel is lost | state table cell nack × cancelled (8.4) | L07 |
| D20 | the expiry sweep never dead-letters: only nack checks `max_attempts` (`internal/jobs/expire.go`) | after D03 is fixed, a crashing worker still re-queues past `max_attempts` (attempts 4, 5, 6 …) | state table cell lease-expired × (attempts = max) (8.4); or Never/Always in L04 if the explorer uses the `crash` kind there | L07 (or L04) |
| D21 | job routes look up by id only, not by namespace (`internal/api/jobs.go`) | a `globex` key reads, cancels and acks `acme` jobs; ids are sequential and guessable | trust boundaries (9.2) | L08 |
| D22 | the background loop `return`s on the first database error instead of continuing | after `docker compose restart postgres`, no lease expires, no schedule fires, no webhook ships, until the api restarts; `/health` stays 200 | What If: the database goes away (9.3) | L08 |
| D23 | `next_run_at` for a schedule is computed in a `time.FixedZone` captured at creation, not the IANA location | after a DST transition every fire is an hour off; `0 2 * * *` in `America/New_York` fires at 03:00 | What If: the clock (9.3); time/date heuristics (A2); needs the L02 test clock | L08 |
| D24 *opt.* | the HMAC is computed over the indented JSON while the compact body is sent | every signature fails verification; hooksink only warns at v0 | Follow the Data across the boundary (9.4) | L08 |
| D25 | `POST /v1/queues/{q}/lease` with an unknown worker returns HTTP 200 `{"error": …}` | clients treat the error as an empty lease | exploring a web service: status codes vs bodies (10.3) | L09 |
| D26 | `PATCH /v1/queues/{q}` silently ignores unknown fields (typo `lease_secs` → 200, no change); `?state=` with an unknown value silently drops the filter | the API never tells you you were wrong | exploring an API: "does it reject what it does not understand?" (10.1); characterizing the bug to the field level (10.4) | L09 |

Spec ambiguities planted in `project/SPEC.md` (the Ch. 12 register; resolved by a spec change and the
code change it implies):

| ID | The spec says | It is silent on | v0 does | Expected in |
|---|---|---|---|---|
| A1 | "failed jobs retry with exponential backoff from `base_seconds`" | a cap | `base·2^n` uncapped: attempt 20 waits 60 days; overflow past 2^63 ns | L10 |
| A2 | "purge removes all pending jobs" | leased jobs that later fail | re-queues them (nack → `retry_wait`) into a queue the operator believes empty | L10 |
| A3 | "a cancelled job is never run" | a job already running when cancelled | marks `cancelled`; the worker runs on, blind; its ack 409s | L10 (with D19) |
| A4 | "terminal jobs are deleted after `retention_hours`" | whether `dead` is terminal for retention | deletes dead jobs too: the dead letter evaporates | L10 |

Twenty-six defects, four ambiguities, three testability gaps, eleven lessons. The clustering is by
technique, not by code area: that is the point. Several defects share one code smell (a missing state
guard on an `UPDATE`: D12, D14, D18, D19, D20, A2, A3), which lesson 11 names as the systemic source
and fixes structurally (Ch. 13.3).

Unplanned defects (real bugs the builder did not intend) will be found; they are recorded as `U##`
in `exploration/defects/`, fixed in the lesson that found them, and counted in that lesson's payoff.

## 4. The lesson sequence

### 4.1 Shape of the history

```
explore-it/v0-inherited   explore-it: add queuelet, the inherited job queue, its spec and the exploration protocol (v0)
explore-it/lesson-01      queuelet: session 01, recon of the inherited system (lesson 01)
                          queuelet: fix D01, D02 found in session 01; add checks (lesson 01)
explore-it/lesson-02      queuelet: session 02, observe the details (lesson 02)
                          queuelet: close testability gaps T1–T3; fix D03, D04; add checks (lesson 02)
explore-it/lesson-03      … session 03, find interesting variations / fix D05–D08 …
explore-it/lesson-04      … session 04, evaluate results with explicit oracles / fix D09–D11 …
explore-it/lesson-05      … session 05, vary sequences and interactions / fix D12–D14 …
explore-it/lesson-06      … session 06, entities, CRUD and following the data / fix D15–D17 …
explore-it/lesson-07      … session 07, the state model and its sneak paths / fix D18–D20 …
explore-it/lesson-08      … session 08, the ecosystem, trust boundaries and what-ifs / fix D21–D24 …
explore-it/lesson-09      … session 09, the API as the interface; characterizing bugs / fix D25–D26 …
explore-it/lesson-10      … session 10, active reading of the spec / resolve A1–A4 in SPEC.md and code …
explore-it/lesson-11      … session 11, paired; strategy and systemic sources / central transition table + model-based check …
explore-it/v1-explored    explore-it: lessons index, landing page, final skills
```

**Order.** Recon first (Ch. 11), because that is how anyone meets an inherited system and because it
produces the charter backlog every later session draws from (Ch. 2 is taught inside L01: the recon
ends with the Nightmare Headline Game). Observation second (Ch. 3), because its testability hooks are
what make the later sessions possible in an hour instead of a day. Then the book's own order:
variations, oracles (Part I); sequences, entities, states, ecosystem (Part II); API, requirements,
integration (Part III). Two deliberate departures: Ch. 10 (no UI) is practised in every session,
since the SUT has no UI, and gets its own lesson late for 10.4 (characterizing bugs) once there are
enough bugs to characterize; Ch. 12 (requirements) comes after the code sessions rather than before,
because the reader is inheriting a system, and the spec reads differently once you know where the
code is wrong. The note for L10 says what the book says: on a new feature, this session comes first.

**Tags** as above; history rewritable until Nick accepts `v1-explored`, then frozen.

### 4.2 The session protocol (`exploration/PROTOCOL.md`, written in checkpoint A)

1. **Charter** in the book's template (2.2): *Explore `<target>` with `<resources>` to discover
   `<information>`*. One target, named resources (tools, data, hooks, models, the spec), one kind of
   information. The lesson's charter is written by the coordinator from this plan; the explorer may
   add charters to the backlog, never widen the one it holds (2.3 "Good Charters").
2. **Time box** (1.3): a session is one headless agent run with a stated budget, 60 minutes wall
   clock (recon 45; the paired session 90), the budget written in the charter and the elapsed time in
   the debrief.
3. **Blind explorer.** The explorer runs from an export, not a checkout: `git archive <tag>
   books/explore-it/project books/explore-it/exploration books/explore-it/skills
   books/explore-it/lessons` into an empty directory with no `.git`, no `PLAN.md`, no `register/`.
   Its brief names the charter, the skill to follow, the protocol, and nothing about what is planted.
   The explorer and the seeder are different model families (section 7). The debrief records how the
   session went, including blind alleys.
4. **Notes** (11.5): `notes.md`, timestamped lines tagged `#obs`, `#bug`, `#question`, `#idea`,
   `#risk`, `#offcharter`, with the exact commands run. The debrief (`debrief.md`, 13.6): charter
   status (*done* / *needs more* / *blocked*), bugs with ids, questions, risks, new charters, time
   split (setup / on-charter / off-charter), what the explorer would do next. A session may end
   *needs more*; the lesson then gets a second session with a narrowed charter. A lesson closes after
   at most two sessions; planted defects still unfound are **disclosed** in the lesson note under
   "What the sessions missed, and why" and fixed in the same fix commit, flagged `disclosed`.
5. **Defect reports** (10.4, 11.5): `defects/D##-<slug>.md` (or `U##`): title; tag seen at;
   minimal steps; expected, with the **oracle named** (the Never/Always rule, the alternative
   resource, the approximation, or the spec line); observed; variables isolated (what did not
   matter); severity; status (`open` → `fixed in lesson NN` or `won't fix: <why>` or `spec`).
   The explorer writes the report; the coordinator reconciles it with the register (same id if
   planted, a `U` id if not) and the fixer writes `lab/repro/D##.sh` from the steps.
6. **The charter backlog** (`exploration/charters.md`, 2.6): every charter ever proposed, with status
   (`planned`, `session NN: done|needs more`), source (recon, headline game, model, requirement,
   debrief), and the lesson it belongs to. Session-based management as a file in git.
7. **The fix** is a separate agent (the fixer, same family as the seeder) with the defect reports
   as its brief: the smallest patch per defect, a regression check per defect under
   `project/test/checks/`, the repro script, the lesson note. It does not touch anything the reports
   do not name.
8. **Models** produced by a session (`exploration/models/`: `nouns-verbs.md`, `variables.md`,
   `entities.md`, `states.md`, `ecosystem.md`) are living files: later sessions revise them; the
   diff of a model between tags is part of the story.

### 4.3 Verification model (same for every lesson)

- `make up` starts the stack; `make seed` loads the fixtures; `make reset` restores the seed in
  seconds; `make down` removes everything.
- `make verify LESSON=NN` runs `lab/verify.sh NN`: reset; run every `lab/repro/D##.sh` the lesson
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
- **The dwell table.** `make dwell LESSON=NN` runs the lesson's repro scripts against every earlier
  tag and prints `defect × tag → PRESENT/ABSENT`. For L07 it reads: D18 present at v0, 01, 02, 03,
  04, 05, 06; absent at 07. The lesson note pastes it under "What the earlier sessions walked past"
  and says why (the charters did not aim there; the state model did not exist; the session ran out
  of time box). That table is this book's equivalent of `FLAW REPRODUCED` and the evidence for the
  claim that technique, not luck, found the defect. Optional and slow (~1 minute per tag); the
  verdict never depends on it.
- `make checks` runs every regression check at any tag (no regressions across lessons);
  `make test` runs the unit tests. Lesson 00's verdict is `BASELINE HOLDS:` (smoke path works, unit
  tests pass, the thin v0 checks pass).
- Thresholds are binary (present/absent) and the fixer may not weaken a repro script to make a
  defect "absent"; the reviewer runs the previous tag's repro to confirm.

### 4.4 Lessons

Each lesson: source, technique and charter, the planted defects it is expected to find, the two
commits, the verify, the payoff (what earlier sessions missed), size.

#### L00 · The inherited system: a tour of v0 (no session)

- **Source:** Ch. 1 (two sides of testing; the session); Ch. 11 opening (inheriting a system).
- **Contents:** `lessons/00-the-inherited-system.md`: what queuelet is, how to run it, how to read
  its spec and its thin tests; the session protocol in a page; **no register**. Unlike Release It!'s
  lesson 00, this tour does not list the defects: the reader is invited to explore blind before
  reading lessons 01–11, and the note says the answer key is section 3 of this plan.
- **Verify:** `make verify LESSON=00` → `BASELINE HOLDS: enqueue→lease→ack→webhook in 2.1 s; unit
  tests 14/14; checks 3/3`.

#### L01 · Recon: meet the system, inventory its nouns and verbs, build the charter backlog

- **Source:** Ch. 11.1 "Starting with a Recon Session", 11.2 "Sharing Observations", 11.3
  "Interviewing Stakeholders to Gather Questions", 11.4 "Selecting Techniques"; Ch. 2.1–2.6 (the
  template, good charters, generating charters, the Nightmare Headline Game, planning); Ch. 1.3.
- **Charter:** *Explore queuelet end to end with `qctl`, the README and `SPEC.md` to discover what it
  does, what it promises, and where its risks are.* 45 minutes. The stakeholder interview (11.3) is
  a written Q&A with the coordinator playing the product owner, limited to ten questions, committed
  as `sessions/01-recon/interview.md`.
- **Expected finds:** D01, D02 (shallow by design; recon finds what any first hour finds).
- **Session commit:** `exploration/sessions/01-recon/{charter,notes,debrief,interview}.md`;
  `exploration/models/nouns-verbs.md` (first version: every noun and verb the explorer met);
  `exploration/charters.md` seeded with ~15 charters from recon, the interview, and the headline game
  (2.5: "Jobs silently lost after a deploy", "Customer billed twice", "Cron fired at the wrong hour",
  "Another tenant cancelled our jobs"), each tagged with the technique the explorer thinks fits
  (11.4); `exploration/defects/D01…`, `D02…`.
- **Fix commit:** the two patches (an `os.Exit(1)` path; a rows-affected check), two checks, two
  repro scripts, `lessons/01-recon.md`.
- **Verify:** `make verify LESSON=01` → `FIXED AND CHECKED: D01, D02 absent`; `REF=v0-inherited` →
  `DEFECTS REPRODUCED`.
- **Payoff:** not the two bugs: the backlog. The note shows how the headline game turned "what could
  be in the news" into charters that L05, L07 and L08 later execute, and how the recon's questions
  ("what happens to a running job when I cancel it?") are the seeds of A3 and D19.
- **Size:** session ~250 lines of prose; fix ~40 lines.

#### L02 · Observe the details: consoles, logs and making the invisible visible

- **Source:** Ch. 3.1 (the moonwalking bear: you miss what you are not looking for), 3.2 "Digging
  Deeper", 3.3 "Testability and Making the Invisible Visible", 3.4 "Consoles and Logs".
- **Charter:** *Explore the job lifecycle with the api and worker logs, `psql` and the `crash` and
  `fail` worker kinds to discover what the API does not show.* 60 minutes. The explorer is told that
  the stack's logs and database are its consoles and that it may **ask for testability hooks** in the
  debrief (3.3): that request is the lesson's first product.
- **Expected finds:** D03 (a job that "ran once" four times; visible in `attempts` vs the worker's
  log), D04 (api log says delivered; hooksink says 500).
- **Session commit:** `sessions/02-observe/`; `defects/D03`, `D04`; the debrief's testability
  requests (T1–T3 in the explorer's words, plus whatever else it asks for).
- **Fix commit:** T1 `log/slog` JSON with `job_id`, `queue`, `event`, `worker` on every
  transition; T2 `job_events` table and `GET /v1/jobs/{id}/events`, `qctl job events`, `qctl watch
  <q>`; T3 the `clock` package and `qctl clock set|advance|reset` under `QUEUELET_TEST_CLOCK=1` (a
  Compose profile sets it in the lab); D03 (count the attempt on expiry), D04 (check the status);
  checks; `lessons/02-observe.md`.
- **Verify:** `FIXED AND CHECKED: D03, D04 absent; checks 2/2; events endpoint shows
  lease→expired→queued→lease→ack for job 1017`. The REF run reproduces both and the note shows the
  v0 console next to the L02 console for the same job.
- **Payoff:** L01 ran the `crash` worker and saw nothing wrong, because nothing was visible. The note
  is honest that the hooks are a patch to the SUT and argues from 3.3 that asking for them is the
  tester's job. Every later lesson's charter lists `qctl clock` or `job events` under resources.
- **Size:** ~300 lines (slog ~60, events ~90, clock ~70, fixes ~30, checks ~50); the largest patch
  after v0; split into two commits (hooks; fixes) if it passes ~400.

#### L03 · Find interesting variations: the variables of a job and a queue

- **Source:** Ch. 4.1–4.4; A2.1 (Some/None/All, Zero/One/Many, Beginning/Middle/End, Goldilocks,
  Boundaries, Count, Position, Selection, Strings, Numbers, Paths; names as in the cheat sheet).
- **Charter:** *Explore enqueue, batch enqueue and queue configuration with the variables list and
  the heuristics cheat sheet to discover inputs the system accepts but cannot honour.* 60 minutes.
  The skill `sweep-variations` is the procedure: list variables first (`models/variables.md`), then
  sweep.
- **Expected finds:** D05 (priority 100), D06 (batch 0 / 501), D07 (names), D08 (2^53).
- **Session commit:** `sessions/03-variations/`; `models/variables.md` (every variable of the
  enqueue and queue surfaces with its type, range, and the heuristic applied); the variation table
  (variable × value × observed) in the notes; defects.
- **Fix commit:** validation (priority range, batch 1–500 with 400 beyond, name charset), payload
  kept as `json.RawMessage` and validated without decoding to `any` (D08); checks; note.
- **Verify:** `FIXED AND CHECKED: D05–D08 absent`; the note's table shows how many variables the
  recon session touched (a handful) versus the sweep (dozens).
- **Payoff:** L01 enqueued a job with priority 5 and a batch of 3. The dwell table shows D05–D08
  present at v0, 01, 02: not because they were hidden but because no one varied the variables.
- **Size:** session ~300 lines; fix ~90 lines.

#### L04 · Evaluate results: Never/Always rules, alternative resources, approximations

- **Source:** Ch. 5.1 "Never and Always", 5.2 "Alternative Resources", 5.3 "Approximations".
- **Charter:** *Explore retries, ordering and stats with a written list of Never/Always rules for a
  job queue, the spec, and SQS/pg-boss documentation as comparable systems, to discover where
  queuelet breaks its own rules.* 60 minutes, with `qctl clock` to walk backoff windows. The explorer
  writes the rules **before** touching the system (`models/rules.md`: "a job is never leased by two
  workers at once", "attempts never exceed max_attempts", "within a priority, always FIFO", "stats
  always equal a direct count", "a cancelled job never runs" …) and then tries to break each.
- **Expected finds:** D09 (off-by-one), D10 (stats), D11 (same-second ordering). D20 is possible
  here if the explorer breaks "attempts never exceed max" with the `crash` kind.
- **Fix commit:** `>=`; stats by exact state with age over `queued` only; `created_at timestamptz(6)`
  plus `id` as the tiebreaker in the lease order (a migration); checks that encode each rule as an
  invariant; note.
- **Verify:** `FIXED AND CHECKED: D09–D11 absent; invariants 5/5 hold over 200 random jobs`.
- **Payoff:** `models/rules.md` becomes the invariant list the checks enforce from here on; the
  note shows rule → check line by line (Ch. 13.1 in miniature). The dwell table shows D09 present
  through L03: sessions that varied inputs did not question outputs.
- **Size:** session ~250 lines; fix ~110 lines incl. migration.

#### L05 · Vary sequences and interactions: nouns and verbs, random navigation, personas

- **Source:** Ch. 6.1 "Nouns and Verbs", 6.2 "Random Navigation", 6.3 "Personas".
- **Charter:** *Explore operator and worker verbs in unusual orders with the nouns-and-verbs list, a
  random sequence generator and three personas (the maintenance operator, the impatient producer,
  the flaky worker) to discover sequences that leave jobs in states the spec does not allow.*
  60 minutes. The session commit includes the little generator the explorer writes
  (`exploration/sessions/05-sequences/walk.sh`: pick a noun, pick a verb, act, observe).
- **Expected finds:** D12 (purge misses delayed and retrying jobs), D13 (resume fires missed ticks),
  D14 (retry from cancelled/succeeded).
- **Fix commit:** purge covers all non-leased pending states; the scheduler advances `next_run_at`
  while paused (at most one catch-up); retry guarded to `dead`; checks (including a scripted random
  walk of 100 steps that asserts the rules from L04 after each step); note.
- **Verify:** `FIXED AND CHECKED: D12–D14 absent; random walk 100/100 steps invariant`.
- **Payoff:** every earlier session used the verbs in the order the README lists them. The note
  reprints the random walk's log where step 37 (`purge`) followed by step 38 (`enqueue --delay`) and
  step 52 (`jobs ls`) shows the surviving job.
- **Size:** session ~300 lines incl. the walker; fix ~100 lines.

#### L06 · Entities and their relationships: CRUD, dependencies, following the data

- **Source:** Ch. 7.1 "Recognizing Entities, Attributes, and Dependencies", 7.2 "CRUD", 7.3
  "Following the Data".
- **Charter:** *Explore queues, schedules, workers and jobs with an entity diagram and a CRUD grid to
  discover what happens to dependents when their owner changes or disappears, and where a payload or
  result changes on its way through the system.* 60 minutes.
- **Expected finds:** D15 (delete queue orphans), D16 (update schedule), D17 (truncation); D08 if
  L03 missed it.
- **Session commit:** `models/entities.md` (diagram, attributes, dependencies, the CRUD grid with
  one cell per entity × operation and the observed outcome), defects.
- **Fix commit:** delete refuses while jobs or schedules exist (409 `queue_not_empty`) or cascades
  with `?force=true`; recompute `next_run_at` on update; truncation flags (`result_truncated`,
  `error_truncated`) or no truncation; checks; note.
- **Verify:** `FIXED AND CHECKED: D15–D17 absent`.
- **Payoff:** the CRUD grid has 16 cells; L01 touched 6. The note shows the grid before and after
  the session with each cell's verdict.
- **Size:** session ~250 lines; fix ~90 lines.

#### L07 · States and transitions: the state model and its sneak paths

- **Source:** Ch. 8.1 "Recognizing States and Events", 8.2 "Diagramming a State Model", 8.3
  "Exploring with State Models", 8.4 "Changing the Representation: State Tables".
- **Charter:** *Explore the job lifecycle with a state model drawn from the spec and the event
  history, two workers and the test clock, to discover transitions the model has no arrow for.*
  60 minutes. The explorer draws the model first (`models/states.md`: diagram, then the **state
  table**: every state × every event, with "expected" filled from the spec), then drives every
  cell that the table marks "should be rejected" or "undefined".
- **Expected finds:** D18 (stale ack), D19 (nack after cancel), D20 (expiry past max). These are the
  deep ones: they need two workers, a clock, and a table that says what "ack in state leased by
  someone else" should do.
- **Fix commit:** token check on ack/nack/heartbeat (409 `stale_lease`); state guard on nack; expiry
  path dead-letters at max; checks driven from the state table (one test per rejected cell); note.
- **Verify:** `FIXED AND CHECKED: D18–D20 absent; state table: 44 cells, 44 as expected`.
- **Payoff:** the dwell table for D18 reads present at every tag from v0 to 06. Sessions 01–06 all
  acked jobs; none acked a job with a lease that had expired, because nothing had asked "what if
  this event arrives in that state". The note puts the filled state table next to the diagram and
  shows which cells the earlier sessions had actually exercised (about a third).
- **Size:** session ~350 lines (the table is big); fix ~120 lines.

#### L08 · The ecosystem: trust boundaries, what-ifs, data across the boundary

- **Source:** Ch. 9.1 "Diagramming the Ecosystem", 9.2 "Trust Boundaries", 9.3 "What If?", 9.4
  "Follow the Data".
- **Charter:** *Explore queuelet's neighbours (Postgres, the workers, the webhook partner, the clock,
  the second namespace) with a context diagram, the hooksink admin modes, `docker compose restart`
  and the test clock, to discover what crosses a boundary unchecked and what happens when a
  neighbour misbehaves.* 60 minutes.
- **Expected finds:** D21 (namespace leak), D22 (loop dies), D23 (DST), D24 (signature).
- **Session commit:** `models/ecosystem.md` (context diagram with trust boundaries marked; the
  what-if list with each neighbour's failure modes and the observed outcome), defects.
- **Fix commit:** namespace scoping on every job route; the loop logs and continues, `/health`
  reports `degraded` when a tick has failed; `next_run_at` through the IANA location each time;
  sign the bytes you send; checks (including a DST check that advances the clock across 2027-03-14
  in `America/New_York`); note.
- **Verify:** `FIXED AND CHECKED: D21–D24 absent; cron 0 2 * * * fires at 02:00 EST and 02:00 EDT
  across the transition`.
- **Payoff:** the context diagram shows six boundaries; the earlier sessions had explored inside one
  box. The note links D21 to the L01 headline "another tenant cancelled our jobs" and shows the
  headline game paying off seven lessons later. This lesson is also where the repo's two books
  touch: the `what-if` list is a chaos-experiment list, and the note points at Release It!'s
  `design-chaos-experiment` skill for how to run them at load.
- **Size:** session ~300 lines; fix ~120 lines.

#### L09 · No user interface: the API as the interface; characterizing bugs

- **Source:** Ch. 10.1 "Exploring an API", 10.3 "Exploring a Web Service", 10.4 "Characterizing
  Bugs" (10.2, exploring a programming language, is noted and skipped).
- **Charter:** *Explore the HTTP surface with `curl`, malformed and near-valid requests, and the
  method/status/content-type matrix to discover where the API accepts what it should refuse or
  answers in the wrong shape.* 60 minutes. Second part of the session: take the three least
  characterized defect reports so far and **re-characterize** them (10.4): minimal steps, variables
  that matter, variables that do not.
- **Expected finds:** D25 (200 with an error body), D26 (silent unknown fields / filter values).
- **Session commit:** the request matrix (route × method × body shape × observed status), the
  three re-characterized reports (diffs of `defects/D##.md`), defects.
- **Fix commit:** proper status on lease errors; strict decoding (`DisallowUnknownFields`) on writes
  with a 400 naming the field; unknown enum values 400; checks; note. (The note contrasts this with
  Release It! L12's tolerant *reader*: a server should be strict about what it accepts as commands
  and tolerant about what it reads from peers; both books are right about their own side.)
- **Verify:** `FIXED AND CHECKED: D25, D26 absent; matrix 38 cells as expected`.
- **Payoff:** every earlier session used `qctl`, which only ever sends well-formed requests; the
  CLI was the moonwalking bear. The note shows the matrix.
- **Size:** session ~250 lines; fix ~70 lines.

#### L10 · Explore requirements: active reading of the spec, charters from ambiguities

- **Source:** Ch. 12.3 "Charter During Requirements Discussions", 12.4 "Active Reading"; 12.1–12.2
  adapted (the "meeting" is a written exchange with the coordinator as product owner).
- **Charter:** *Explore `SPEC.md` with active reading (every sentence: what does it not say? what
  would make it false? what word is doing too much work?) and the defect list so far, to discover
  requirements that are ambiguous, missing or already contradicted by the code.* 60 minutes.
  Output: an ambiguity list with a proposed resolution per item and a charter per resolution.
- **Expected finds:** A1–A4, and at least two ambiguities the plan did not plant (every spec has
  them). Resolution of each is a **product decision**: the explorer proposes, the coordinator (as
  product owner) decides in `sessions/10-requirements/decisions.md`.
- **Fix commit:** `SPEC.md` revised (the diff is the lesson), code follows: backoff cap
  (`max_seconds` default 3600), purge semantics for leased jobs (`nack` after purge → `dead`),
  `cancel_requested` surfaced in the heartbeat response and honoured on ack, dead jobs exempt from
  retention; checks; note.
- **Verify:** `FIXED AND CHECKED: A1–A4 resolved; spec diff 31 lines; checks 4/4`.
- **Payoff:** the note shows that four of the earlier defects (D12, D14, D19, D20) were arguments
  about what the spec meant, settled by code written before anyone asked. The book's claim (12.3)
  that the cheapest exploration is of the requirements is demonstrated backwards: here it came
  last and cost eight sessions. On a new feature it comes first.
- **Size:** session ~250 lines; spec ~40 lines changed; fix ~80 lines.

#### L11 · Integrate exploration throughout: strategy, pairing, systemic sources, stopping

- **Source:** Ch. 13.1 "Exploring as Part of Your Test Strategy", 13.2 "Paired Exploration", 13.3
  "Discovering Systemic Sources of Problems", 13.4 "Estimating Exploration", 13.5 "How to Tell When
  You Have Explored Enough", 13.6 "Debriefing Stakeholders", 13.7 "Capturing Useful Nuggets of
  Wisdom".
- **Charter (paired, 90 minutes):** two explorers from different families, one driving and one
  navigating by turns, with the whole exploration record, to discover *whether the system is explored
  enough*: re-run the ten highest-risk charters from the backlog against `lesson-10` and look for
  anything the record does not already know.
- **Session commit:** `sessions/11-paired/` (both agents' notes, interleaved by timestamp);
  `exploration/STRATEGY.md`: the checked+explored map (which risks the checks hold, which the
  sessions covered, which neither), the estimate (sessions spent per lesson against the plan's
  budget), the stop rule applied (13.5: new findings per session fell from 3–4 to 0–1 over L08–L11),
  the stakeholder debrief written for Nick (13.6), and the **systemic analysis** (13.3): the defect
  list grouped by cause: seven missing state guards, four silent acceptances, three clock/time
  bugs, two console lies, and what each group says about how the code was written.
- **Fix commit:** the systemic fix for the biggest group: `internal/jobs/transition.go`, one
  transition table (the L07 state table as code) that every handler goes through, with a **model-based
  check** that drives random event sequences against the table (the L05 walker, now a test);
  `skills/*/SKILL.md` final revisions from the sessions' nuggets (13.7); the lessons index and
  landing page; `v1-explored`.
- **Verify:** `FIXED AND CHECKED: all 26 defects absent at v1-explored; checks 31/31; model-based
  walk 1 000 steps clean`, plus the paired session's own finds, if any, fixed.
- **Payoff:** the strategy document is the artifact a team would keep; the systemic fix shows
  that eleven sessions of findings reduce to a handful of causes, which is the book's last argument.
- **Size:** session ~400 lines; fix ~150 lines plus skills edits.

### 4.5 Expected find rates, honestly

A blind explorer will not find every planted defect in its lesson. The plan expects 60–80 % on the
first session and most of the rest on a second, narrowed session; the remainder are disclosed. The
plan also expects out-of-order finds (D08 in L06, D14 or D20 in L04) and unplanned finds. All three
outcomes are recorded, not hidden, and the lesson notes carry a "Found / Missed / Found early /
Unplanned" table. The dwell table makes the teaching claim without needing a perfect script: a
defect that sat through six sessions and fell to the seventh's technique is the lesson, however many
others fell early.

## 5. Coverage honesty

| Chapter | Lesson | Note |
|---|---|---|
| 1 On Testing and Exploration | L00, L01, protocol | checked vs explored framed in L00; the session as unit of work in the protocol |
| 2 Charter Your Explorations | L01 (+ every charter) | template, good charters, generating, headline game, planning: all exercised in L01 and reused in every lesson |
| 3 Observe the Details | L02 | testability hooks as a SUT patch, argued from 3.3 |
| 4 Find Interesting Variations | L03 | with A2.1's heuristics as the kit |
| 5 Evaluate Results | L04 | rules → invariant checks |
| 6 Vary Sequences and Interactions | L05 | nouns/verbs, random navigation (as a script), personas |
| 7 Entities and Relationships | L06 | |
| 8 States and Transitions | L07 (+ L11's transition table) | |
| 9 Explore the Ecosystem | L08 | touches Release It!'s chaos skill |
| 10 No User Interface | every session; L09 for 10.4 | 10.2 (a programming language) skipped |
| 11 Explore an Existing System | L01; 11.5–11.6 in the protocol | unreproducible bugs (11.6): the protocol has a procedure; no planted defect is intermittent on purpose, but D11 (same-second ordering) behaves like one and the L04 note says so |
| 12 Explore Requirements | L10 | the meeting is a written exchange; 12.1–12.2's social advice summarized, not enacted |
| 13 Integrate Exploration Throughout | L11; 13.6 in every debrief | |
| A1 Interviewing | omitted | hiring, out of scope; mentioned in the landing page |
| A2 Cheat Sheet | `sweep-variations` skill; L03 | A2.2 (web heuristics) does not apply to an API-only SUT |

Nothing of Part I or II is dropped. Omitted on purpose: A1, 10.2, A2.2.

## 6. Skills

Eight skills, written in checkpoint A from the book (drafts), executed by the sessions, and revised in
L11 from what the sessions taught. Each follows the skills contract in `AGENTS.md` (when to use,
inputs, procedure, output, source, worked example). The worked example of each is a session in this
repo; two also name litefleet as a second target to run against once both arcs land.

| Skill | Triggers | Procedure in one line | Source | Worked example |
|---|---|---|---|---|
| `recon-existing-system` | "we inherited this"; a new codebase or service; first day on a project | run a recon session (map, nouns and verbs, questions), interview the stakeholder for questions, select techniques per question, seed the charter backlog | Ch. 11.1–11.4, 2.4 | session 01; second target: litefleet's tenant API |
| `charter-exploration` | "what should we test"; a new feature; a risk review; a backlog that is empty or vague | write charters in the template, check each against "good charter" rules, generate more with the Nightmare Headline Game and from the models and requirements, plan sessions | Ch. 2 | `exploration/charters.md` across all lessons |
| `run-exploration-session` | a charter to execute; "go find out whether…" | set up, time-box, observe through every console, take tagged notes, record defects as reports, debrief in the template, update the backlog | Ch. 1.3, 3, 11.5, 13.6 | any `sessions/NN/`; session 02 for the consoles step |
| `sweep-variations` | an input surface, a config, a command with options; "what values should we try" | list the variables (not just fields: counts, sizes, positions, timing, order, config), apply the named heuristics to each, record a variation table | Ch. 4, 6.1–6.3, A2.1 | session 03; session 05's walker |
| `model-to-explore` | a feature with state; entities with relationships; a service with neighbours; "draw me the …" | build the entity/CRUD grid, the state table, and the context diagram with trust boundaries; turn every empty or "undefined" cell into a charter; run the what-ifs | Ch. 7, 8, 9 | sessions 06, 07, 08; second target: litefleet's database lifecycle (create → provision → placement → drop) and its topology |
| `derive-oracles` | "how do we know it's right"; no spec; a stats or report feature; a number nobody checks | write Never/Always rules, name comparable systems and standards, choose approximations, turn rules into invariant checks | Ch. 5 | session 04 → `models/rules.md` → checks |
| `report-defect` | a bug to write up; a flaky or unreproducible report; "I saw it once" | characterize: minimal reproduction, isolate variables, name the oracle, severity, environment; the unreproducible-bug procedure | Ch. 10.4, 11.5, 11.6 | `defects/D18`, the three re-characterized in session 09 |
| `plan-exploration-strategy` | sprint or release planning; "how much testing is enough"; a growing defect list; a postmortem | map checked vs explored risks, estimate sessions, set the stop rule, run the systemic-sources analysis on the defect list, write the stakeholder debrief, capture nuggets | Ch. 13, 12.3 | `exploration/STRATEGY.md`, session 11 |

Answer-key exercise: running `recon-existing-system` then `charter-exploration`, `sweep-variations`,
`model-to-explore` and `derive-oracles` blind against `explore-it/v0-inherited` should find most of
section 3 (the plan says 18 of 26 is the bar, measured by the actual sessions: whatever the eleven
sessions found without disclosure is the measured number and goes in the skills index).

Overlap with the Release It! skills: `model-to-explore`'s what-if step and `design-chaos-experiment`
meet; the Explore It! skill produces the list, the Release It! skill runs one at load. Both skill
indexes cross-link.

## 7. Build order, seats and budget

Three checkpoints with cross-family review between them, as for Release It!. The difference is the
cast: three roles instead of one builder.

| Role | Does | Seat (proposal) | Why |
|---|---|---|---|
| Coordinator | writes charters from this plan, briefs the explorer and the fixer, reconciles defect reports with the register, commits session artifacts as produced, plays product owner in L01 and L10 | the planner seat, interactive (Claude Code) | accumulated context is the value; it never explores and never fixes |
| Seeder / fixer | builds v0 with the register planted, writes the proof scripts, later writes each lesson's fix, checks, repro scripts and note | Codex Astra (`gpt-6-astra`), headless, fresh agent per unit | volume building; a different family from the explorer is the blindness guarantee; the Claude pool is on the Release It! build |
| Explorer | runs each session blind from the export | Claude Opus 5.5 headless (Muse Wonder Full for the second seat in L11's pair and as the fallback when the Claude pool is limited) | the skills are Claude Code / Codex format and are exercised by running them; the planning/research lean suits exploration |
| Reviewer | per checkpoint, from a fresh clone: the review bar in section 7.2 | Muse Wonder Full (`--model muse-wonder-full-1.0`, `--contributor-mode off`) | the third family; also the arbiter |

If Nick prefers Claude as the seeder (as for litefleet), swap seeder and explorer; the rule is only
that they differ.

**Checkpoint A — v0, the lab, the protocol, the skills drafts, L00 and L01 (~2 800–3 300 lines;
Release It! A was 3 500–4 000).**
- A1 *the system runs*: `project/` (api ~1 300 lines, worker ~250, qctl ~500, hooksink ~150,
  migrations, Compose, Makefile `up/down/seed/reset/verify/dwell/checks/test/logs`), `SPEC.md` with
  A1–A4 silent, unit tests, README. Plant all 26 defects where section 3 says; natural code, no
  markers. Proof scripts `books/explore-it/register/proof/D##.sh` (outside `project/`, never
  exported) that each demonstrate one defect; the reviewer runs all 26.
- A2 *the record and the kit*: `exploration/PROTOCOL.md`, empty `charters.md`, `models/README.md`,
  `defects/README.md`; `lab/verify.sh` with `REF` builds, `lab/repro/` framework, `lab/dwell.sh`;
  eight `SKILL.md` drafts from the book (coordinator-reviewed for citations); `lessons/00`. Prove
  `make up && make seed && make verify LESSON=00`. Tag `v0-inherited`.
- Then L01 end to end (recon session by the explorer, fix by the fixer) to prove the whole loop,
  including the export mechanism and the blind brief, before B starts. Tag `lesson-01`.

**Checkpoint B — L02 through L07 (six sessions, up to three second sessions; ~600 lines of SUT
patches and checks, ~1 800 lines of record; Release It! B was ~1 000 lines of patches).** Per lesson:
coordinator writes the charter → explorer session → coordinator reconciles → (second session if
*needs more* and a planted defect is unfound) → fixer → reviewer spot-checks the previous tag's
repro. L02 is the largest patch and may be two commits.

**Checkpoint C — L08 through L11, skills final, landing page (~450 lines of patches, ~1 600 of
record; Release It! C was ~1 100 lines of patches).** L11's paired session needs two explorer seats
at once.

**Budget, compared to Release It!.** Code: about 65 % (a smaller v0, small patches, no load lab,
no fault injector). Prose and record: about 150 % (eleven sessions of notes and debriefs, five
models, thirty defect reports, a spec, a strategy). Agent runs: Release It! is one builder pass in
three checkpoints plus three reviews; this is ~14 explorer runs (11 + up to 3 repeats), ~12 fixer
runs, 3 reviews, plus the coordinator's standing session. Wall clock is dominated by sessions
(60–90 minutes each by the time box) rather than by builds; verify runs are seconds, not minutes.
Net: roughly the same total agent-hours as Release It!, spread across all three pools instead of one,
with far less VM load. Pool check before each checkpoint per the repo protocol; if the Claude pool is
limited, Muse explores; if two pools are limited, finish the in-flight lesson and stop.

**Rules for the seeder and the fixer.**
- Read `AGENTS.md`, this plan, then `PROTOCOL.md`. The register, the lesson order and the protocol
  are fixed; the fix for a defect is whatever the defect report and the spec say, not what the plan
  sketches, if they differ (record the deviation in the note).
- v0 passes its own tests and `make verify LESSON=00`. A planted defect that breaks the happy path
  is planted wrong.
- Every SQL statement takes `$now` from the app clock; nothing uses Postgres `now()`.
- A fix commit touches only what its lesson's defect reports name. No defect is fixed early by
  accident (the reviewer runs later lessons' proof scripts at earlier tags).
- Repro scripts are black-box (API, CLI, read-only `psql`, test clock, `docker compose`); checks may
  use internals. A repro script is never weakened to turn PRESENT into ABSENT.
- The register is never written anywhere under `project/`, `exploration/`, `skills/` or `lessons/`
  until the lesson that discloses it.
- Commit subjects as in 4.1; the fix commit's body carries the defect ids, both verdict lines and the
  dwell summary. Trailer per repo rules.

**Rules for the coordinator.**
- The explorer's brief contains: the export path, the charter, the skill name, `PROTOCOL.md`, the
  time box, and the instruction that it may read anything in the export and nothing outside it. It
  never contains hints about planted defects. The brief is committed with the session.
- Session artifacts are committed as produced; formatting only.
- A lesson closes after at most two sessions. Disclosure is written in the lesson note, never into
  the session's own files.

## 8. Layout

```
books/explore-it/
  PLAN.md                           # this document
  README.md                         # landing page: argument, lesson index with tags, skills index
  register/                         # answer-key material outside the explorer's export
    proof/D01.sh … D26.sh           # the seeder's proof that each planted defect is live at v0
  lessons/
    README.md                       # index: lesson, tag, defects found/disclosed, verify command
    00-the-inherited-system.md      # tour of v0 without the register
    01-recon.md … 11-integrate.md
  exploration/                      # the record that grows: the thing the lessons improve
    PROTOCOL.md                     # charter template, session rules, note/debrief/defect templates
    charters.md                     # the living charter backlog
    STRATEGY.md                     # from L11
    models/                         # nouns-verbs.md variables.md rules.md entities.md states.md ecosystem.md
    sessions/NN-<slug>/             # brief.md charter.md notes.md debrief.md (+ interview.md, walk.sh, decisions.md)
    defects/                        # README.md (index with status) D##-<slug>.md U##-<slug>.md
  project/                          # queuelet: the inherited SUT
    README.md  SPEC.md  Makefile  go.mod  go.sum
    cmd/api/  cmd/worker/  cmd/qctl/  cmd/hooksink/
    internal/api/  internal/jobs/ (lease, ack, nack, expire, cancel, retry; transition.go from L11)
    internal/schedule/  internal/webhook/  internal/stats/  internal/store/  internal/clock/ (L02)
    migrations/  deploy/ (compose.yml, Dockerfiles)
    lab/ (verify.sh, repro/D##.sh, dwell.sh, reset.sh, seed/)
    test/checks/ (L01 onward)       results/ (gitignored)
  skills/
    README.md
    recon-existing-system/ charter-exploration/ run-exploration-session/ sweep-variations/
    model-to-explore/ derive-oracles/ report-defect/ plan-exploration-strategy/
```

The explorer's export is `project/`, `exploration/`, `skills/`, `lessons/` at the previous tag; never
`PLAN.md`, `register/`, or `.git`.

## 9. Suggested changes outside `books/explore-it/` (notes, not edits in this pass)

- **`AGENTS.md`**, "Purpose": add that for a book about a practice the archetype is a SUT plus a
  growing exploration record, and that a lesson's "patch" is a session plus the fix of what it
  found. "Lesson contract": allow the two-commit lesson (session, fix) with the tag on the fix, and
  the verdict words `FIXED AND CHECKED:` / `DEFECTS REPRODUCED:` beside `IMPROVED:` / `FLAW
  REPRODUCED:`. "Phase flow": allow skills to be drafted in phase 2's first checkpoint when the
  lessons execute them, with the final revision still at the end. Add the **blind-explorer rule**
  under the deliberate-issues register: when the lessons are sessions, the seeder and the explorer
  are different families and the explorer works from an export without the plan. "Layout of a
  book": add the optional `exploration/` and `register/` folders. "Review bar": add "every session's
  brief contains no hint of the register; every disclosed miss is in the lesson note".
- **Root `README.md`**: a row in the books table (*Explore It!*, Hendrickson, 2013; status: plan
  v1 awaiting confirmation; link to this plan); a sentence in the intro that for practice books the
  history shows an exploration getting better rather than a system; the reader's path gains
  `make dwell LESSON=NN`.
- **`STATE.md`**: an `explore-it` section: phase 1 plan v1 written 2026-10-22 by the planner seat
  (Claude Fable 5.1) on branch `plan/explore-it`, awaiting Nick; phases 2–4 not started; seats as in
  section 7; "merge after the Release It! arc lands; adds-only".
- **`docs/knowledge/`** (does not exist yet): the export mechanism for blind agents and the
  two-commit lesson are process knowledge worth a page once they have run once.

## 10. Decisions Nick is asked to confirm

1. **The shape** (section 0): sessions as lessons, two commits per lesson, the SUT fixed as it goes,
   the dwell table as the payoff, skills drafted first and executed by the sessions.
2. **The blind-explorer rule** and the three-role cast (coordinator / seeder-fixer / explorer) with
   the proposed seats (Codex seeds and fixes, Claude explores, Muse reviews), or the swap.
3. **The vehicle**: `queuelet`, a fresh baby job queue (Go, Postgres 17, a webhook partner fake, a
   test clock), over reusing litefleet (kept as a second worked example for two skills), the
   workflow platform and the deploy product. The name is a placeholder.
4. **The register** at 26 defects (4 optional), 4 spec ambiguities, 3 testability gaps, and the
   eleven lessons in the stated order, with Ch. 12 late and Ch. 10 spread.
5. **The protocol**: 60-minute sessions (45 recon, 90 paired), at most two sessions per lesson,
   disclosure of misses in the note, artifacts committed as produced, the written stakeholder
   interview with the coordinator as product owner.
6. **Checkpoints** A (v0 + record + skills drafts + L00 + L01), B (L02–L07), C (L08–L11) with
   cross-family review; the budget framing in section 7.
7. **Isolation**: the branch merges after the Release It! arc lands; the build may start earlier in
   its own worktree since everything is under `books/explore-it/`; only the section 9 edits wait.
