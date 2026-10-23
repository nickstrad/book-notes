# Brief: finalize one Explore It! skill (checkpoint C)

Repo worktree: /root/Software/booknotes-work/explore-it (branch plan/explore-it). Book dir:
`books/explore-it/`. You finalize exactly ONE skill file: `books/explore-it/skills/<SKILL>/SKILL.md`
(named in your prompt). Write nothing else. Do not commit, do not run git commands that change
state, do not start Docker stacks, do not edit any other file. Scratch notes, if any, go in
/tmp/explore-it-checkpoint-c/scratch-<SKILL>/.

## Read first

1. `AGENTS.md` (repo root), section "Skills contract" — the required format and sections.
2. `books/explore-it/PLAN.md` section 7 (the skills table: triggers, one-line procedure, source
   chapters, worked example) and section 2.2 (chapter map with section titles).
3. The current draft `skills/<SKILL>/SKILL.md` (a checkpoint-A draft written before any experiment).
4. `exploration/PROTOCOL.md`.
5. The experiment records your prompt names: `exploration/experiments/NN-*/{charter,notes,revealed}.md`
   and `brief.md`, the lesson notes `lessons/NN-*.md`, the models in `exploration/models/`,
   the defect reports in `exploration/defects/` (raw `F##`/`E##-F##` reports are the explorer's own
   words; `D##`/`U##` files are canonical status; `INDEX.md` maps them), and the reconciliations in
   `register/reconciliations/NN.md`.

## Facts that hold for the whole arc (do not contradict them)

- Experiment 01's blind recon (Opus, reading the code and probing) found and reported 26 of the 27
  planted defects (D07 only partially: charset evidence; the slash/space manifestation was disclosed
  in lesson 01), plus U01–U08; all 27 were fixed in lesson 01. Later experiments therefore found NEW,
  unplanned faults (U09–U45) and spec decisions (A1–A5), and re-ran early-fixed D-defects as controls.
  Never write as though a later experiment discovered a planted D-defect.
- Raw blind reports per experiment E01–E09: 27/3/8/5/3/7/2/4/7. The stop signal did not fall.
- Scores per experiment live at the end of each `revealed.md` and in each lesson note. Quote them
  exactly from there.
- Book citations: Elisabeth Hendrickson, *Explore It!* (Pragmatic Bookshelf, 2013). Section numbers and
  titles come from PLAN 2.2 / the lesson notes, checked against the publisher's table of contents.
  Full book-text validation has NOT been done: keep one honest sentence saying so in Source. Never
  invent a quotation from the book. Procedure wording is this repo's, not the book's.

## What the final skill must be

An actionable procedure an agent can run on a codebase it has never seen, not a chapter summary and
not a queuelet walkthrough. Generic procedure first; queuelet only in the worked example and in short
"queuelet:" evidence pointers where a step was corrected by the experiments.

Sections, in this order (frontmatter first):

```
---
name: <SKILL>
description: <one line: what it does and when to use it; trigger-focused>
---
# <Title>
(one short paragraph: what the skill is for, and that in this repo's arc you also follow PROTOCOL.md)
## When to use      — concrete triggers (kinds of code, questions, events); also one line "Not for:"
## Inputs           — what must be at hand
## Procedure        — numbered steps. Each step: what to do, what to look for, and CONCRETE SIGNALS
                      (grep patterns, SQL/HTTP probes, questions to ask, thresholds, stop rules).
                      Fold in what the experiments taught (see below); where a step exists because of
                      a queuelet lesson, end it with a short "(queuelet: … — link)" pointer.
## Output           — the shape of the result: table columns / file names, so results are comparable
## Pitfalls the experiments exposed — 3–6 bullets, each a concrete mistake or trap observed in the
                      record (with link), and what the procedure now does about it
## Source           — book sections with titles; the honest validation sentence; the lessons in this
                      repo that demonstrate the skill (links)
## Worked example   — what the skill produced when run against queuelet: which experiment(s), the
                      charter in one line, the actual score (exact numbers), what it found (U/A ids
                      with links to canonical records), what it predicted wrongly or missed, the model
                      delta (which model files), and one short excerpt or table from the record that
                      shows the procedure working. Then "Second target" (litefleet, from PLAN 7) as a
                      one-paragraph sketch of how to run it there, clearly marked as not yet run.
```

Length: roughly 90–150 lines. Plain, precise English (short sentences, no hype, no filler).
Use relative links from the skill's location (e.g. `../../exploration/experiments/05-sequence-it/walk.sh`,
`../../lessons/07-state-it.md`). Every link you write must point to an existing file: check each with
`ls`. Every number, id and `file:line` you cite must be verified in the record. If records disagree,
prefer canonical D/U records and lesson notes, and say so.

## When done

Reply with: the file path, line count, the list of links you verified, and anything you found in the
records that looked inconsistent (do not fix it; just report it).
