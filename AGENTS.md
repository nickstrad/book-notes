# AGENTS.md — how a book becomes software and skills in this repo

This file is the book-agnostic playbook. It applies to every book under `books/`. Nothing here is
specific to one book; the worked example is `books/release-it/` (start with its `PLAN.md`).
Repo-wide agent rules from `~/.claude/CLAUDE.md` (seats, fresh-agent lifecycle, cross-family
review, commit style) still apply; this file only adds what is particular to breaking down a book.

## Purpose

Nick reads a technical book. The repo turns it into two artifacts:

1. **One flawed archetype project per book, improved lesson by lesson.** A single realistic
   service or program, built with real technology and seeded with deliberate, realistic flaws that
   the book warns about. Each lesson of the book is a **patch commit** that fixes one flaw (or one
   tight cluster) in that same codebase. The commit history is the teaching artifact: the first tag
   is the "before", every lesson tag is a checkpoint a reader can run, and the final tag implements
   the book's lessons. `git log` reads as the codebase getting better.
2. **Skills** that CODIFY the book's main parts as reusable procedures an agent (or a person) can
   run on real code and real designs, each pointing at the project as its worked example.

The book itself is the source of truth. Every flaw, lesson and skill must trace back to specific
chapters.

## Phase flow and the confirmation gate

| Phase | Seat | Produces | Done when |
|---|---|---|---|
| 1. Research & plan | planner (top tier) | `books/<slug>/PLAN.md` (project design, deliberate-issues register, lesson sequence, skills, build order), `STATE.md` update | **Nick has confirmed the plan.** Nothing is delegated to a builder before that. |
| 2. Build v0 + lessons | builder (top tier; the work is subtle concurrency and failure behaviour) | `books/<slug>/project/` at the flawed tag, then one commit per lesson with its note and verify script, tags per plan | Every lesson's verify passes at its tag and reproduces the flaw at the previous tag; checkpoints reviewed |
| 3. Write skills | builder (may follow phase 2 directly) | `books/<slug>/skills/<skill>/SKILL.md` for each planned skill | Each skill meets the skills contract and cites its chapters and the project |
| 4. Review | reviewer from a DIFFERENT model family than the builder | Review notes in `STATE.md`; fixes folded into the lessons they belong to | Review bar below is met; verdict recorded |

**The gate.** The planner's plan is a proposal until Nick confirms it. Plan revisions (v1 → v2 → …)
are committed as documents; the previous plan is archived next to the new one when the approach
changes. Phase 2 may also be split into checkpoints (the plan says which), each confirmed by a
cross-family review before the next begins. Builders do not re-decide structure: the register and the
lesson order are fixed by the plan; thresholds and sizes may be tuned, with the deviation recorded in
the lesson note and in `STATE.md`.

Hand-off between phases is through files in the repo, never through chat.

## Choosing the archetype project

The plan picks one project per book. It earns the role if ALL of these hold:

- Every core, transferable idea of the book has a **natural home** in it: a place where the flaw the
  book describes would plausibly be written by a competent team, and where its fix is the book's
  remedy. If a core idea has no home, the archetype is wrong, not the idea.
- It is **modest**: a few thousand lines at the flawed tag, buildable by one builder pass.
- It is built on **real technology** for its domain (Docker Compose if it is a service; a real
  database; real instrumentation such as Prometheus; a real load balancer or fault injector where the
  book's lessons involve them), chosen and justified per component in the plan. One primary language,
  chosen by what makes the book's failures and remedies most visible.
- It **runs badly, not brokenly**: at the flawed tag it passes its own tests and serves gentle load.
  Flaws are behavioural (hangs, storms, unbounded growth, cascades, lost work), never compile errors
  or cartoon sabotage, and never marked in the code. The plan's register is the only list of them.

## The deliberate-issues register

The heart of the plan. One row per flaw: id, where it lives (file / area), how it shows under the
right load, the chapter and anti-pattern that names it, and the lesson that fixes it. Flaws must be
realistic (the kind real codebases have), individually demonstrable on the tree one lesson before
their fix, and fixable in one focused patch each. The register doubles as the answer key for the
book's "hunt for X" skill.

## Lesson contract

A lesson is one commit on the main line (two if the plan says the patch would exceed ~400 lines; the
lesson tag points at the last). It contains:

1. **The patch**: the fix for the register rows it names and nothing else. A reader must be able to
   `git diff <prev-tag> <tag> -- project/` and understand it in minutes.
2. **The note** `books/<slug>/lessons/NN-<slug>.md`: the idea in plain words with chapter citations;
   the flaw with `file:line` pointers into the previous tag; the patch walkthrough naming the lines
   that matter; the verification with the exact command and expected before/after output; "try
   changing"; where to read more. Written for someone who has not read the book yet.
3. **The verify script** `project/lab/verify/NN-<slug>.sh`, run as `make verify LESSON=NN`: resets
   state, applies the load and chaos timeline, prints a comparable report table and the named
   metrics, and ends with a verdict line starting `IMPROVED:` or `FLAW REPRODUCED:` against explicit
   thresholds. `make verify LESSON=NN REF=<previous tag>` runs the same script with the project built
   from the previous tag and must reproduce the flaw. Runs in a couple of minutes with warm images.
4. **The commit body**: flaw, fix, both verdict lines pasted in. It must stand alone in `git log`.
5. **A tag** per the plan's naming (`<slug>/v0-flawed`, `<slug>/lesson-NN`, `<slug>/v1-hardened`).

Chaos is injected only from outside the codebase (fault injector API, container kill/pause/restart,
partner-fake admin endpoints, SQL run by the script), never through hidden paths in the service.
History may be rewritten (fixups folded into their lesson, tags re-pointed) until Nick accepts the
final tag; after that it is frozen.

## Layout of a book

```
books/<slug>/
  PLAN.md                     # the confirmed plan; earlier versions archived alongside
  README.md                   # landing page: argument, lesson index with tags, skills index
  WALKTHROUGH.md              # guided tour for a cold reader (see below)
  lessons/
    README.md                 # index: lesson, tag, flaw ids, verify command
    00-<the-flawed-start>.md  # tour of the flawed tag and the register
    NN-<lesson>.md            # one note per lesson
  project/                    # the one codebase; `make up`, `make verify LESSON=NN`, `make down`
    lab/                      # load generator, verify runner and scripts, report queries
  skills/
    README.md                 # index of the book's skills with triggers
    <skill-name>/SKILL.md
```

Names are stable once committed: notes, skills and commit bodies link to them.

**Every book ships a `WALKTHROUGH.md`**, written as part of the book's landing/wrap (after the
review fixes, before the book is reported done). It is a tour a person follows, not a file listing,
and contains: an orientation (what the project is, the shape of `books/<slug>/`); the reading order
(plan, the flawed start, the lesson notes in order, the skills, the review and its fix pass); the
tag/commit itinerary, with the exact `git` commands for each stop and what flaw or change to observe
there; the commands to bring the project up and run a lesson's verify, copied from its own
Makefile/docs and verified against the repo (run the cheap ones; confirm every file, tag and target
named exists); and pointers to `STATE.md`, the review and `docs/knowledge/`. The worked example is
`books/release-it/WALKTHROUGH.md`.

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
- **Source** — the chapters the steps come from, and the lessons in this repo that demonstrate them.
- **Worked example** — what the skill produces when run against the book's project (for a "hunt"
  skill, the register is the answer key and the skill is expected to find most of it).

A skill may reference the book's vocabulary but must be usable by someone who has not read it.
Skill names are imperative verb phrases (`test-integration-points`, `plan-safe-release`).
Aim for 5 to 8 skills per book; each must earn its place by covering a part of the book the others
do not.

## Review bar

The phase-4 reviewer checks, from a FRESH clone on this VM:

1. At the flawed tag, the project builds, its tests pass and its baseline verify passes: the flaws
   are latent, not obvious.
2. For every lesson, `make verify LESSON=NN` passes at its tag and
   `make verify LESSON=NN REF=<previous tag>` reproduces the flaw, each within its time budget. The
   verdict lines match the commit body.
3. Each lesson's diff is confined to what its note claims; no flaw is fixed early by accident and no
   new one is introduced (spot-check two later lessons' verify at an earlier tag).
4. Every claim about the book in a note, the plan or a skill traces to a chapter or named section
   that actually says it. Chapter numbers match the edition named in the plan. Spot-check at least
   one citation per lesson and per skill.
5. Each skill is executable as a procedure and is not a chapter summary; its worked example matches
   the project.
6. History is linear, tags are where the plan says, and `git log` reads as the intended story.

Verdicts are PASS / NITS / FAIL per the repo-wide protocol; one repair round, then arbitration.

## Conventions

- Commit subjects `<area>: <imperative summary>`; `<area>` is `docs`, the book slug (`release-it`)
  for plan-level and landing-page commits, the project name (`litefleet` for Release It!) for lesson commits, `skills`,
  or `repo`. Lesson commits end their subject with `(lesson NN)`.
- `STATE.md` at the repo root records which book is in which phase and what is next; update it in
  the same commit as the phase's deliverable.
- Durable knowledge about this repo's process goes in `docs/knowledge/` (not the `kb` CLI, which is
  for VM/environment facts).
- Scratch files go in `/tmp/<task>/`, never in the repo. Load-test results under `project/results/`
  are gitignored.

## Starting a new book

1. Create `books/<slug>/` and write `PLAN.md` following `books/release-it/PLAN.md`: core argument →
   archetype project (what it does, components, partners, stack with rationale) → deliberate-issues
   register → lesson sequence (source, flaw, patch, commit subject, verify) → coverage honesty →
   skills → build order and budget → layout → decisions to confirm.
2. Add the book to the root `README.md` index and to `STATE.md`. Commit.
3. Wait for Nick's confirmation. Then dispatch phase 2 to a builder seat with the plan as the brief,
   one checkpoint at a time.
