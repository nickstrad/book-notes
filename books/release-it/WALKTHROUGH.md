# Release It! — walkthrough

A guided tour of `books/release-it/` for someone arriving cold. Follow it top to bottom: orient,
read in order, then walk the tags with the stack running. Every path, tag and `make` target below
was checked against the repo on 2026-10-22 (see the end).

## 1. What this is

*Release It!* (Nygard, 2nd ed.) turned into one codebase. `litefleet` is a baby multitenant SQLite
database platform in Go: HAProxy in front of three gateways, a Postgres-backed control plane, two
node agents that each hold tenants' SQLite files, and two partner fakes (blob store, metering), all in
Docker Compose with Toxiproxy between the tiers and Prometheus scraping every service.

The commit history is the lesson. `release-it/v0-flawed` passes its tests and serves gentle load,
but carries eighteen deliberate flaws (F01–F18) of the kind the book warns about. Twelve lesson tags
each fix one flaw or one tight cluster, and `release-it/v1-hardened` is the end state. Eight skills
turn the book into procedures you can run on your own code, with litefleet as the worked example.

```
books/release-it/
  README.md     landing page: the argument, lesson table, skills
  PLAN.md       the approved plan (v3); PLAN-v1/-v2 are superseded drafts
  lessons/      00 (tour of v0 + the register) and 01–12, one note per tag; README.md is the index
  project/      litefleet: Makefile, cmd/, internal/, deploy/, lab/ (verify scripts, load, chaos)
  skills/       eight SKILL.md procedures; README.md is the index with triggers
  REVIEW.md     the final review (Fable) and its findings
  STATE.md      where the project stands now, and the open decisions
```

## 2. Reading order

1. **`README.md`** (five minutes): the book's argument in a paragraph, the platform, the lesson table.
2. **`PLAN.md` sections 2–4**: why this vehicle (2.0), the topology (2.5), the register (3) and the
   verification model (4.2). Skip the lesson sections for now; the notes supersede them.
3. **`lessons/00-the-flawed-start.md`**: the platform in one picture, the lab, and the register with
   `file:line` pointers into v0. This is the answer key for everything after.
4. **`lessons/01-…` to `12-…` in order.** Each has the same parts: the idea (with chapter citations),
   the flaw at the parent tag, the patch walkthrough, Verify (exact commands, before and after
   verdicts), Try changing, Deviations. `lessons/README.md` has every before/after on one page.
   Read lesson 01's "After the review" section last; it is a defect found after the review.
5. **`skills/README.md`**, then the skills in its suggested order: `review-operational-transparency`
   → `hunt-stability-antipatterns` → `suggest-stability-patterns` → `budget-timeouts-and-retries`,
   `review-capacity-headroom`, `plan-safe-release` → `test-integration-points`,
   `design-chaos-experiment`. Each ends with a worked example on litefleet.
6. **`REVIEW.md`**: the verdict (PASS with fixes recommended), the full verify matrix, and findings
   H1, M1–M6, L1–L10. Then **`STATE.md`** "Done": the commit that answered each finding, and which
   ones are still open.

## 3. Bring the stack up

Docker with Compose v2 and GNU make. All host ports are in the 1xxxx range (tenant API on 18080,
Prometheus on 19090; full table in `project/README.md`).

```
cd books/release-it/project
make up && make seed        # build images, start the stack; the first seed takes ~1 min
make verify LESSON=00       # baseline: BASELINE HOLDS on a quiet host
make down                   # remove everything, volumes included
```

`make verify LESSON=NN` resets the stack (it restores a seeded snapshot in seconds), runs that
lesson's load and fault timeline, prints a report table and Prometheus readings, and ends with a
verdict line: `IMPROVED:` or `FLAW REPRODUCED:` with the thresholds spelled out. Exit status 0 means
the thresholds held. `make verify LESSON=NN REF=<tag>` builds gateway, control and node from `<tag>`
(the harness still comes from your checkout), which is how a lesson reproduces its flaw on the
parent tag. Each run takes 2–4 minutes with warm images.

Other targets: `make test` (unit tests), `make test-chaos` (lesson 11's chaos matrix), `make migrate`,
`make rollout` and `make godot` (lesson 12), `make calibrate` (measure capacity into
`lab/capacity.env`), `make audit` (DuckDB read of every tenant file), `make logs`, `make ps`.

The lab runs on small containers on purpose and is sensitive to a busy host: latency lines (lesson
00's p99, lesson 08's report time) can miss at load average 8+. See
`docs/knowledge/lab-on-a-shared-vm.md` and STATE.md's open decisions.

## 4. The tag itinerary

Walk the tags in a separate worktree so the main checkout stays on `main`:

```
git worktree add /tmp/ri-tour release-it/v0-flawed
cd /tmp/ri-tour/books/release-it/project
make up && make seed && make verify LESSON=00
```

At each lesson stop, the same four steps (shown for lesson 05):

```
git switch --detach release-it/lesson-05
git diff release-it/lesson-04 release-it/lesson-05 -- .     # the patch, readable in minutes
make verify LESSON=05 REF=release-it/lesson-04              # FLAW REPRODUCED on the parent
make verify LESSON=05                                       # IMPROVED on this tag
```

`git log --oneline release-it/v0-flawed..release-it/v1-hardened` reads as the whole story; each
lesson commit's body has the flaw, the fix and both verdict lines. Every checkout shares the Compose
project name `litefleet`, so run one stack at a time. When you are done:
`make down`, then `git worktree remove /tmp/ri-tour` from the main checkout.

| Stop | Parent (`REF=`) | Watch for in the before run | What changes after |
|---|---|---|---|
| `release-it/v0-flawed` | — | `BASELINE HOLDS` at half capacity, yet `up = 0` for gateway, control and node: Prometheus has nothing to scrape. That is the only visible hint. | — |
| `lesson-01` transparency | `v0-flawed` | node-a slows and Postgres pauses, but every `/health` says `ok`, HAProxy shows 3 of 3 UP, logs are `handled POST /path` (F01) | metrics, JSON logs with request ids, health that fails within a second |
| `lesson-02` timeouts | `lesson-01` | ~1 100 requests in flight against a hung node; failures are clients giving up, not 504s; a runaway query keeps running (F02–F04) | ~150 in flight, 99 % answered 504 inside 2.9 s, statements cancel |
| `lesson-03` blocked threads | `lesson-02` | a `UNIQUE` violation leaks a connection and the file's write lock; node-a pinned at 1 024 fds (F05–F07) | rollback on every path, bounded pools, fds well under the wall |
| `lesson-04` retries | `lesson-03` | duplicate rows from retried writes; the tight `SQLITE_BUSY` loop (F08) | reads-only retries with a budget, `busy_timeout`, zero duplicates |
| `lesson-05` circuit breaker | `lesson-04` | control down means 0 % served and every gateway unhealthy (F09) | ~99 % served from the route cache; the breaker closes seconds after recovery |
| `lesson-06` bulkheads | `lesson-05` | a noisy tenant lifts neighbours' p99; control pool timeouts; a slow blob store holds the writer lock for the whole upload (F10–F11) | per-tenant admission, separate pools, lock held ~0.5 s |
| `lesson-07` shed load | `lesson-06` | kill a gateway under load: goodput collapses, p99 near 3 s, while every server is "up" (F12) | 503 fast at gateway, node and HAProxy; goodput holds |
| `lesson-08` unbounded results | `lesson-07` | the whale's `SELECT * FROM events` OOM-kills node-a repeatedly (F13) | capped, streamed answers; zero restarts; usage report aggregated in SQL |
| `lesson-09` back pressure | `lesson-08` | slow metering grows the usage queue to thousands; a node restart loses every unshipped event (F14) | bounded queue with a SQLite outbox; nothing lost |
| `lesson-10` steady state | `lesson-09` | WALs grow ~14 MB/min until the volume is full; handles and the ledger only grow (F15) | checkpointed WALs, purged ledger, evicted idle handles |
| `lesson-11` test harness | `lesson-10` | `make test-chaos` too: 5 of 12 chaos tests fail (F16) | 12 of 12 pass |
| `lesson-12` handling versions | `lesson-11` | add `REF_SERVICES=gateway` for old gateways with today's nodes: almost every query fails on the mixed fleet; a node restart cuts requests off (F17–F18). `make godot` replays the in-place `ALTER` stall | tolerant readers, expand/contract migration, node drain; zero errors through the rollout |
| `release-it/v1-hardened` | — | the end state; also carries every review fix and the post-review logging fix (`9367863`) | run any lesson's verify here, with any `REF=` |

Measured before/after numbers for every stop are in `lessons/README.md`; the exact verdict lines are
in each note's Verify section.

Two caveats, both in STATE.md's open decisions:

- **Lesson 04 from a later tag.** From a checkout of `lesson-06` through `lesson-12`,
  `make verify LESSON=04` dies without a verdict (REVIEW.md H1; the fix is a forward commit, not
  folded into history). Run it from `main` or `v1-hardened` with `REF=release-it/lesson-03`.
- **A tag carries its own scripts.** Lesson tags hold the verify scripts as they were then; the
  review fixes to lessons 01–04's scripts and lesson 00's 80 ms line exist only from `main`. To use
  today's harness against an old tree, stay on `main` and pass `REF=<tag>`.

## 5. Where things stand

- **`STATE.md`**: current status, every commit and tag, open decisions (cross-family review not run,
  H1 not folded into history, M3 thresholds, latency lines under host load, book summaries for
  Nick to check), and resume instructions.
- **`REVIEW.md`**: the verdict, the verify matrix as the reviewer ran it, and every finding.
- **`docs/knowledge/`** (repo root): what building the lab taught. Toxiproxy fault removal, running
  the lab on a shared VM, resets and seeds, plan size estimates, and how a verify workaround hid a
  defect. Start at its `README.md`.

## Checked for this walkthrough (2026-10-22)

All fourteen tags exist and `release-it/v1-hardened` (`9367863`) is an ancestor of `main`;
`git diff <parent> <tag> -- books/release-it/project/` is non-empty for every lesson. Every file,
directory and skill named above exists. `make -n` resolves `up`, `seed`, `down`, `test`,
`test-chaos`, `godot`, `rollout`, `migrate`, `calibrate` and `audit`; every `lab/*.sh` and
`lab/verify/*.sh` parses; `make test` passes at `main`; the section 4 worktree commands (add,
`switch --detach`, `diff -- .`, `log`, remove) work. The section 3 sequence was run on `main`: `make up`
and `make seed` succeeded (200 tenants, 243 databases), and `make verify LESSON=00` served every
request (no 4xx, 5xx or timeouts) with every service `up` but printed
`BASELINE BROKEN: p99_ms 97.5 (want <= 80)` at load average 16 on the shared VM, consistent with the host-load
latency item in STATE.md. The per-lesson verifies were not re-run for this
walkthrough; REVIEW.md and STATE.md record the last full runs.
