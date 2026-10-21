# AGENTS.md — how a book becomes software and skills in this repo

This file is the book-agnostic playbook. It applies to every book under `books/`. Nothing here is
specific to one book; the first worked example is `books/release-it/` (start with its `PLAN.md`).
Repo-wide agent rules from `~/.claude/CLAUDE.md` (seats, fresh-agent lifecycle, cross-family
review, commit style) still apply; this file only adds what is particular to breaking down a book.

## Purpose

Nick reads a technical book. The repo turns it into two artifacts:

1. **Topic projects** that TEACH the book's core ideas by making them visible: small, runnable
   programs where the reader watches the failure the book warns about, then watches the remedy.
2. **Skills** that CODIFY the book's main parts as reusable procedures an agent (or a person)
   can run on real code and real designs: "test the integration points", "hunt for X", "plan Y".

The book itself is the source of truth. Every topic and skill must trace back to specific chapters.

## Phase flow

Each book goes through four phases. Each phase is one unit of work for a fresh agent (per the
repo-wide fresh-agent lifecycle). Hand-off is through files in the repo, never through chat.

| Phase | Seat | Produces | Done when |
|---|---|---|---|
| 1. Research & plan | planner (top tier) | `books/<slug>/PLAN.md`, scaffold of topic folders with stub READMEs, `books/<slug>/skills/README.md` index, root `README.md` index entry, `STATE.md` update | Plan names every topic and skill with sources, teaching design, language and run command; stubs exist; committed |
| 2. Implement topics | builder (top tier for subtle concurrency/failure demos, middle tier for the rest) | Code, `run.sh`, finished README per topic, `books/<slug>/run-all.sh` | Every topic's `./run.sh` passes from a fresh checkout within the time budget |
| 3. Write skills | builder (may follow phase 2 directly) | `books/<slug>/skills/<skill>/SKILL.md` for each planned skill | Each skill meets the skills contract below and cites its chapters |
| 4. Review | reviewer from a DIFFERENT model family than the builder | Review notes in the PR / `STATE.md`; fixes land in a repair round | Review bar below is met; verdict recorded |

Phases 2 and 3 may be split across several agents by topic; the plan fixes the structure so
builders do not re-decide it. If a builder finds the plan wrong for a topic, they record the
deviation and its reason in that topic's README under "Deviations from plan" and in `STATE.md`.

## Topic selection bar

A topic earns a folder only if ALL of these hold:

- It carries a **core, transferable idea** of the book: something the author returns to, builds
  on, or names as a cause of real failures. Not a passing mention, not a tooling survey.
- It can be **shown**, not just told: a reader can run something and see the behavior change.
- It is **self-contained**: runnable with no external services, network access, or accounts.
- It is **not a duplicate**: if two chapters teach one mechanism, make one topic and cite both.

Aim for 8 to 12 topics per book. If the count climbs, merge; if a chapter does not pass the bar,
list it in the plan under "Deliberately omitted" with the reason, so the omission is a decision.

## Layout of a book

```
books/<slug>/
  PLAN.md               # phase-1 research product: argument, parts, topics, skills, omissions
  README.md             # landing page: what the book argues, topic index, skills index
  run-all.sh            # runs every topic's run.sh; the reviewer's first command
  topics/
    NN-<topic>/         # NN = reading order; <topic> = kebab-case idea name
      README.md         # the teaching document (see contract)
      run.sh            # the single documented run command
      before/ after/    # or before.* / after.* — language-appropriate (see contract)
  skills/
    README.md           # index of the book's skills with triggers
    <skill-name>/
      SKILL.md          # the procedure (see contract)
```

Names are stable once committed: skills and READMEs link to them.

## Per-topic folder contract

Every topic folder provides:

1. **A teaching README** with, in this order: the idea in plain words (2 to 4 paragraphs, with
   chapter citations), what the reader will see (before and after), how the demo works (code
   walkthrough naming the few lines that matter), how to run it, "try changing" suggestions, and
   where to read more in the book. Written for someone who has not read the book yet.
2. **A runnable demo** in exactly one of Rust, Go, Node, or Bash, using only that toolchain's
   standard library (no crates.io, no npm packages, no Go modules outside std). Deterministic
   enough to read: fixed seeds, scripted timelines, printed tables or ASCII timelines rather than
   raw log spew.
3. **Before/after** whenever the idea is a failure mode and its remedy. `before` shows the failure
   happening and names it in its final line; `after` applies the remedy and shows the difference
   in the same output format so the two runs can be compared line for line. A `before` that
   would hang or run forever must be bounded by a watchdog that prints the state and exits
   non-zero with a message saying production would not have stopped. When the idea is a tool or
   a practice rather than a remedy, the README says "before/after: not applicable" and explains
   what the single demo shows instead.
4. **A single documented run command**: `./run.sh` at the topic root runs `before` then `after`
   (or the single demo) and exits 0 on a clean run. `./run.sh before` and `./run.sh after` run one
   side. `run.sh` performs any build step itself. The whole run finishes in under about 30 seconds
   on this VM; a topic that needs longer is split or simplified.
5. **No shared code between topics.** Each topic is a complete, copyable example. Duplication
   across topics is accepted; coupling is not.

## Language rubric

Choose per topic by what teaches the idea best. The question to answer in the plan is "which
toolchain makes the failure and the remedy most visible in the fewest lines?" Defaults:

- **Rust** when the lesson is about who owns a resource, a state machine that must be exhaustive,
  or a bounded-vs-unbounded primitive, and the type system or RAII IS the remedy (`Drop` guards,
  `sync_channel` vs `channel`, `wait_timeout`). Std-only; `cargo run` from a package with
  `src/bin/before.rs` and `src/bin/after.rs`.
- **Go** when the lesson is about concurrent servers, worker pools, sockets, HTTP client/server
  knobs, or many goroutines doing simple things; the std `net`, `net/http`, `context`, `expvar`
  and `log/slog` packages cover production-shaped demos with no dependencies. One `go.mod` per
  topic with `before/main.go` and `after/main.go`.
- **Node** when the lesson is about event-loop scheduling and timers, JSON/API contracts, retry
  logic as commonly written in web clients, or memory growth visible via `process.memoryUsage()`.
  ESM `.mjs`, no `package.json` dependencies.
- **Bash** when the lesson is about files, disks, processes, cron-like housekeeping, or
  orchestrating other programs; bash 5 plus coreutils only.

Rules: one language per topic. Never choose by habit or by "the rest are in Go". If more than
half of a book's topics land in one language, re-check each rationale. Record the rationale in
the plan and in the topic README in one sentence.

## Skills contract

A skill is an **actionable procedure**, not a summary of the book. It lives at
`books/<slug>/skills/<skill-name>/SKILL.md` in the format that Claude Code and Codex load:

```
---
name: <skill-name>
description: <one line: what it does and when to use it — this is what triggers it>
---
```

followed by these sections:

- **When to use** — concrete triggers (the kind of code, question, or event that should invoke it).
- **Inputs** — what the agent needs at hand (a repo, a design doc, an incident timeline).
- **Procedure** — numbered steps an agent can execute, each with what to look for or do, and
  concrete signals (grep patterns, questions, thresholds) where the book gives them.
- **Output** — the shape of the result (a table, a checklist, a plan) so results are comparable.
- **Source** — the chapters the steps come from, and the topics in this repo that demonstrate them.

A skill may reference the book's vocabulary but must be usable by someone who has not read it.
Skill names are imperative verb phrases (`test-integration-points`, `plan-safe-release`).
Aim for 5 to 8 skills per book; each must earn its place by covering a part of the book the
others do not.

## Review bar

The phase-4 reviewer checks, from a FRESH clone on this VM:

1. `books/<slug>/run-all.sh` exits 0 and every topic finished within its budget. Any `before`
   that is bounded by a watchdog exits as its README says it will.
2. For every before/after topic, the two outputs differ in the way the README promises and the
   `before` names its failure in its last line.
3. Every claim about the book in a README, PLAN or skill traces to a chapter or named section
   that actually says it. Chapter numbers match the edition named in the plan. Spot-check at
   least one citation per topic and per skill.
4. Each topic's language choice matches the plan or records a justified deviation.
5. Each skill is executable as a procedure (a reader could follow the steps on a real repo) and
   is not a chapter summary.
6. No external services, network calls, or non-std dependencies anywhere in `books/`.

Verdicts are PASS / NITS / FAIL per the repo-wide protocol; one repair round, then arbitration.

## Conventions

- Commit subjects `<area>: <imperative summary>`; `<area>` is `docs`, `release-it` (or the book
  slug), `skills`, or `repo`.
- `STATE.md` at the repo root records which book is in which phase and what is next; update it in
  the same commit as the phase's deliverable.
- Durable knowledge about this repo's process goes in `docs/knowledge/` (not the `kb` CLI, which
  is for VM/environment facts).
- Scratch files go in `/tmp/<task>/`, never in the repo.

## Starting a new book

1. Create `books/<slug>/` and write `PLAN.md` following the phase-1 row above and the
   `books/release-it/PLAN.md` structure: core argument → parts with chapter refs → topics → skills
   → deliberately omitted → build order and budgets.
2. Scaffold `topics/NN-<topic>/README.md` stubs carrying the teaching outline from the plan, and
   `skills/README.md` listing planned skills.
3. Add the book to the root `README.md` index and to `STATE.md`. Commit.
4. Dispatch phase 2 to a builder seat with the plan as the brief.
