# Explore It! — walkthrough

A guided tour of `books/explore-it/` for someone arriving cold. Follow it top to bottom: orient,
read in order, bring the lab up, then walk the tags. Every path, tag and `make` target below was
checked against the repo on 2026-10-22 (see the end).

## 1. What this is

*Explore It!* (Hendrickson, 2013) turned into one codebase and the record of learning it.
`queuelet` is a small Go job queue over Postgres 17: producers enqueue JSON jobs into named queues
inside a namespace, workers lease, heartbeat, ack or nack, and the service retries, dead-letters,
fires cron schedules in IANA time zones and notifies a webhook partner (`hooksink`). There is an
HTTP API and a CLI, `qctl`, all in Docker Compose.

The project is a deliberately defect-seeded system, explored blind. `explore-it/v0-inherited` passes
its own baseline but carries 27 planted defects (D01–D27), four spec silences (A1–A4) and three
testability gaps (T1–T3). Nine experiments were each run by a fresh agent from an export of the
previous tag (source, exploration record, skills, lessons) without the plan, the register or git
history. The boundary was the prompt and the export, not OS isolation, and it has documented
exceptions: A1 reached E04 through an exported lesson note, E09's charter supplied the expected
cause groups, and three runs read their own runtime output outside the export. See the
[exceptions table](register/blindness-exceptions.md); E01's result precedes all of them. Each experiment is two commits: the explorer's record, then a fixer's patch with
repro scripts and checks, tagged `explore-it/lesson-NN`. `explore-it/v1-explored` adds the final
skills and landing page.

The surprise is the story: experiment 01's recon found 26 of the 27 planted defects, and all 27 were
fixed in lesson 01. Experiments 02–09 then found 37 more faults nobody planted (U09–U45) and settled
the spec silences, plus a fifth (A5), with the probe family each chapter teaches.

```
books/explore-it/
  README.md       landing page: the argument, what happened, lesson and skill tables
  PLAN.md         the confirmed plan (v2); section 4 is the planted-defect register
  STATE.md        status, every commit and tag, open items with owners
  lessons/        00 (the inherited system) and 01–09, one note per tag; README.md has the scores
  exploration/    PROTOCOL.md, hypotheses.md, models/, experiments/NN-*/, defects/ (INDEX.md)
  project/        queuelet: Makefile, SPEC.md, cmd/, internal/, deploy/, lab/ (verify, repro), test/
  skills/         eight SKILL.md procedures; README.md is the index with triggers
  register/       build records: proof/ (v0 answer-key probes), reconciliations/, runs/, reviews/
  docs/knowledge/ what building the lab taught (fault injection, post-commit races, isolation)
```

## 2. Reading order

1. **`README.md`** (five minutes): the argument in a paragraph, what happened, the two tables.
2. **`PLAN.md`** sections 0–1 (what was cut and why; the experiments-against-an-inherited-codebase
   shape), 3 (the vehicle) and 5.2–5.3 (the experiment protocol and verification model). Leave
   section 4, the register, for later if you want to try exploring blind first (section 6).
3. **The v0-inherited system**: `lessons/00-the-inherited-system.md`, then the project as the
   explorers received it: `project/README.md`, `project/SPEC.md` and `exploration/PROTOCOL.md`.
4. **The nine experiments, in order.** For each, read the lesson note, then the explorer's own
   record under `exploration/experiments/NN-<slug>/` (`charter.md`, `notes.md`, `revealed.md`,
   `evidence/`):

   | # | Note | Record | Found and fixed |
   |---|---|---|---|
   | 01 | `lessons/01-map-it.md` | `experiments/01-map-it/` | D01–D27, U01–U08 |
   | 02 | `lessons/02-make-it-visible.md` | `experiments/02-make-it-visible/` | U09–U10, T1–T3 |
   | 03 | `lessons/03-vary-it.md` | `experiments/03-vary-it/` | U11–U18 |
   | 04 | `lessons/04-judge-it.md` | `experiments/04-judge-it/` | U19–U23, A1 |
   | 05 | `lessons/05-sequence-it.md` | `experiments/05-sequence-it/` | U24–U26, A2 |
   | 06 | `lessons/06-relate-it.md` | `experiments/06-relate-it/` | U27–U33, A4 |
   | 07 | `lessons/07-state-it.md` | `experiments/07-state-it/` | U34–U35, A3 |
   | 08 | `lessons/08-surround-it.md` | `experiments/08-surround-it/` | U36–U38, A5 |
   | 09 | `lessons/09-explain-it.md` | `experiments/09-explain-it/` | U39–U45, transition table |

   `lessons/README.md` puts every experiment's score on one page. Explorers' records are never
   rewritten after the fact; the notes say where later work corrected them.
5. **The defect register.** The answer key is `PLAN.md` section 4. The canonical status of every
   finding is `exploration/defects/INDEX.md`: each D/U row links the explorer's original report
   (`F*`, `E0N-F*`) to the current fixed record.
6. **The checkpoint reviews**, under `register/reviews/`: `A/review-a.md` (v0 and E01, Muse,
   NITS), `B/review-b.md` (E02–E06, Sol, PASS) and `C/review-c.md` (E07–E09 and the full register,
   Opus, PASS with one instrument flake and nits).
7. **The final skills**: `skills/README.md`, then the skills in its suggested order:
   `map-codebase-from-behavior` → `charter-as-probe` → `make-invisible-visible` → `sweep-variables`
   → `derive-oracles` → `probe-sequences-and-states` → `probe-entities-and-boundaries`, with
   `follow-the-surprise` whenever something odd appears. Each ends with a queuelet worked example
   and the pitfalls the experiments exposed.

## 3. Bring the lab up

Go 1.26, Docker with Compose, Python 3 and curl. The API listens on 18200, hooksink on 18201,
Postgres on 15433; seed keys are `acme-key`, `globex-key` and `admin-key` (`project/README.md`).

```
cd books/explore-it/project
make up && make seed        # build images, start the stack, load 300 mixed-state jobs
make verify LESSON=00       # BASELINE HOLDS: enqueue→lease→ack→webhook; unit and happy-path checks pass
make down                   # remove the stack and its volume
```

`make verify LESSON=NN` resets the fixtures, runs every black-box repro script listed in
`lab/repro/lesson-NN.ids` (API, `qctl`, read-only SQL, the lab clock), then that lesson's
regression checks. Each repro prints `PRESENT: ID — …` or `ABSENT: ID — …`; the run ends
`FIXED AND CHECKED: N absent; regression checks pass`. Lessons 07–09 also run early-fixed planted
defects as controls, which must stay `ABSENT`.

`make verify LESSON=NN REF=<tag>` archives `<tag>`'s project into a temporary stack
(`queuelet-ref`, ports 18210/18211/15443), copies your checkout's current instruments into it and
runs them against the historical binaries; it ends `DEFECTS REPRODUCED: N present`. Most verifies
take one to five minutes; lesson 01's waits through real minute ticks and takes about six.

Other targets: `make test` (unit tests), `make checks` (all 94 regression checks, about 12
minutes), `make reset` (same as `make seed`), `make dwell LESSON=NN` (that lesson's repros at
every earlier tag, each built fresh: slow), `make logs`, `make qctl ARGS='queue ls'`. To run beside
another stack, set `COMPOSE_PROJECT_NAME`, the `QUEUELET_*_PORT`/`_URL` variables and
`QUEUELET_DATABASE_URL` as `project/README.md` describes; lesson 07's note has a full example.

## 4. The tag itinerary

Walk the tags in a separate worktree, starting at the inherited system:

```
git worktree add --detach /tmp/ei-tour explore-it/v0-inherited
cd /tmp/ei-tour/books/explore-it/project
make up && make seed && make verify LESSON=00
```

At v0 the baseline holds, the exploration record is empty, and every planted defect is live. The
seeder's answer-key probes show them directly: `bash ../register/proof/run-all.sh` prints a
`PRESENT` line for each of D01–D27 (it interrupts the database for D22; reset afterwards).

At each lesson stop, the same steps (shown for lesson 07):

```
git switch --detach explore-it/lesson-07
git log --oneline explore-it/lesson-06..explore-it/lesson-07   # this stop's commits, ending experiment → fix
git diff explore-it/lesson-06 explore-it/lesson-07 -- .        # the patch and its instruments
make up && make seed                                            # rebuild the stack from this tag
make verify LESSON=07 REF=explore-it/lesson-06                  # DEFECTS REPRODUCED: 3 present
make verify LESSON=07                                           # FIXED AND CHECKED: 3 absent
```

`git log --oneline explore-it/v0-inherited..explore-it/v1-explored` reads as the whole story; each
fix commit's body carries both verdict lines. Every checkout shares the Compose project name
`queuelet`, so run one stack at a time. When done: `make down`, then
`git worktree remove /tmp/ei-tour` from the main checkout.

| Stop | `REF=` parent | Reproduces on the parent (`N present`) | Fixed at this tag |
|---|---|---|---|
| `explore-it/v0-inherited` | — | `BASELINE HOLDS`; D01–D27 live but latent | — |
| `lesson-01` map it | `v0-inherited` | 35: all 27 planted defects (ack ignores the lease token, job endpoints cross namespaces, max-attempts off by one, non-2xx webhooks marked delivered, background loop dies on one error, …) and U01–U08 (invalid cron on patch, slow webhook stalls the loop, a DB outage reported as a bad key, …) | all 35 |
| `lesson-02` make it visible | `lesson-01` | 5: T1–T3 (no structured lifecycle logs, no retained job history, no lab clock), U09 (`fail3` not registered by the default worker), U10 (one malformed webhook URL blocks every later notification) | 5; adds `qctl job events`, `qctl watch`, `qctl clock` |
| `lesson-03` vary it | `lesson-02` | 8: U11–U18 (trailing JSON and case-variant names accepted, duration and backoff overflow, storage errors as 409, one large retention halts all retention, `max_seconds` ignored, partial webhook patch clears the other member) | 8 |
| `lesson-04` judge it | `lesson-03` | 6: U19–U23 (unmatched routes answered outside the JSON error form, ambiguous attempt history, negative delay, expired lease honoured until the sweep, queued age counts leased time) and A1 (no default retry cap) | 6; A1 decided at 3600 s |
| `lesson-05` sequence it | `lesson-04` | 4: U24–U26 (worker delete releases without counting the attempt, terminal rows keep lease fields, release logged as `lease_expired`) and A2 (leased job at purge) | 4; A2 decided |
| `lesson-06` relate it | `lesson-05` | 8: U27–U33 (schedule delete not 204, schedule payload unvalidated, echo worker and `qctl` lose number precision, NUL text as 409, enqueue racing queue delete, attempts outlive their job) and A4 (dead-job retention) | 8; A4 decided |
| `lesson-07` state it | `lesson-06` | 3: U34 (worker delete racing a lease leaves a zombie lease), U35 (purge misses rows changing state) and A3 (cancel while running); D18–D20 `ABSENT` as controls | 3; A3 decided |
| `lesson-08` surround it | `lesson-07` | 4: U36 (API never recovers from starting without its DB), U37 (one slow destination starves later notifications), U38 (a frozen DB hangs health and the API) and A5 (DST gap and overlap); D21–D24 as controls | 4; A5 decided |
| `lesson-09` explain it | `lesson-08` | 7: U39–U45 (one unadvanceable schedule stops all later ones, never-matching cron accepted, a `CRON_TZ=`/`TZ=` cron prefix overrides the time zone, zone offset frozen in the record, error after a committed ack/cancel and a lost notification, one vanished job drops later dead jobs' notifications, non-IANA zones accepted); D12, D14, D18–D20, D27 as controls | 7; every state change now guarded by `internal/jobs/transition.go` |
| `explore-it/v1-explored` | — | no project change from `lesson-09`: final skills, landing page, lessons index | run any verify here with any `REF=`, then `make checks` |

Three things to know on the way:

- **A tag carries its own instruments.** `lab/repro/lesson-NN.ids` first exists at
  `explore-it/lesson-NN`, so `make verify LESSON=09` from an earlier checkout exits with "No
  reproductions registered". To see a later lesson's faults on an older tree, stay on the later
  checkout and pass `REF=<older tag>`. `make dwell LESSON=NN` does that for every earlier tag.
- **The stack is whatever `make up` built.** The non-`REF` verify runs against the running stack,
  so rebuild after every switch. Walk forward; `make down` before going back to an earlier tag.
- **Exact verdicts and timings** are in each note's verify section, and the last full run of all of
  them is `register/runs/final/acceptance.json`.

## 5. Explore it yourself first (optional)

To try the method before reading the answers, export the inherited tag the way the explorers got
it (`exploration/PROTOCOL.md`, "Roles and export"):

```
mkdir /tmp/ei-blind
git archive explore-it/v0-inherited books/explore-it/project books/explore-it/exploration \
  books/explore-it/skills books/explore-it/lessons | tar -x -C /tmp/ei-blind
```

That export has the draft skills and lesson 00 but no plan, register or history. Start with
`map-codebase-from-behavior` and `charter-as-probe`, write a prediction before each probe, and
compare with lesson 01 afterwards.

## 6. Where things stand

- **`STATE.md`**: build complete, every commit and tag, and the open items with owners. Astra's
  (Codex) cross-family end review ran on 2026-10-23 ([REVIEW.md](REVIEW.md): PASS WITH FIXES);
  STATE.md records its findings R1–R6 and what the fix pass did with each. Also open: spec questions for Nick (Q-E09-1, Q-E09-2), upgrade policy for
  legacy schedules and payloads, the U32 sampler flake, testability requests and forward hypotheses.
- **`exploration/defects/INDEX.md`**: the status of every D, U, A and T finding.
- **`register/reviews/`**: the three checkpoint reviews with their runners and logs; review C's
  "Open items carried forward" lists what the next reviewer inherits.
- **`docs/knowledge/`** (this book's): neighbour fault injection, data boundaries, post-commit
  races and running reviews in isolation. Start at its `README.md`.

## Checked for this walkthrough (2026-10-22)

All eleven `explore-it/*` tags exist and `explore-it/v1-explored` (`b6df7db`) is an ancestor of the
branch; `lesson-09` → `v1-explored` changes nothing under `project/`. Every file, directory, experiment
record (`charter.md`, `notes.md`, `revealed.md`, `evidence/`), review and skill named above exists;
`register/proof/run-all.sh` exists at `v0-inherited`. The `.ids` counts are 35/5/8/6/4/8/3/4/7, matching
the table and `register/runs/final/acceptance.json`. `make -n` resolves `up`, `down`, `seed`, `reset`,
`test`, `checks`, `logs`, `verify`, `dwell` and `qctl`; every `lab/*.sh`, `lab/repro/*.sh` and
`register/proof/*.sh` parses; `make test` passes. The section 5 `git archive` export was made and holds
`project`, `exploration`, `skills` (eight drafts) and `lessons` only. Section 4 was run as written
in a scratch worktree on a quiet host: at `v0-inherited`, `make up`, `make seed` and
`make verify LESSON=00` (`BASELINE HOLDS`, 15 s); then `git switch --detach explore-it/lesson-07`,
`make up && make seed`, `make verify LESSON=07 REF=explore-it/lesson-06`
(`DEFECTS REPRODUCED: 3 present`, controls D18–D20 `ABSENT`, 133 s) and `make verify LESSON=07`
(`FIXED AND CHECKED: 3 absent; regression checks pass`, 300 s); `make down` and the worktree removal
succeeded. Not re-run for this walkthrough: the other lessons' verifies, `make checks`, `make dwell`
and the v0 proof sweep; STATE.md and `register/runs/final/` record the last full runs.
