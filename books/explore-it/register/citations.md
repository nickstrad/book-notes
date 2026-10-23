# Citation ledger — what has been checked against the book

The review bar ([AGENTS.md](../../../AGENTS.md), item 4) asks that every claim about the book
trace to a section that actually says it. That check needs the book's full text, which this
repo does not hold. This ledger records exactly what was checked, against which source, and
what remains. It answers end-review finding R1 ([REVIEW.md](../REVIEW.md)); it does not close it.

Edition: Hendrickson, *Explore It!*, Pragmatic Bookshelf, P1.0 (February 2013),
ISBN 978-1-937785-02-4.

## Sources checked (2026-10-23)

| Source | Covers | What it can confirm |
|---|---|---|
| Publisher contents, [toc.pdf](https://media.pragprog.com/titles/ehxta/toc.pdf) | Every chapter and section title, Ch. 1–13, A1, A2 | That a cited number exists and its title names the claimed topic. No page numbers (the extract prints `?`) |
| Excerpt, [charters.pdf](https://media.pragprog.com/titles/ehxta/charters.pdf) | 2.3 in full, 2.4 most of it | Charters neither too specific nor too broad; a charter as a prompt; charters from requirements, implicit expectations, stakeholder questions, existing artifacts (source, bug database) and new discoveries during sessions |
| Excerpt, [states.pdf](https://media.pragprog.com/titles/ehxta/states.pdf) | Ch. 8 intro, 8.1, 8.2 (part) | Windows of vulnerability; model from **observable behavior, setting internal knowledge aside**; states as behavioral modes ("while…"); external, system-generated and time events; narrow focus, perspective, level of abstraction |
| Excerpt, [explore.pdf](https://media.pragprog.com/titles/ehxta/explore.pdf) | Ch. 10 intro, 10.1, 10.2 | Techniques apply without a GUI; vary data, size and count; Zero/One/Many as interface-independent; 10.1's worked API session (empty strings, nulls, long inputs, a hang) |

Not available to this repo: the body text of Chapters 1, 3–7, 9, 11–13, sections 2.1, 2.2,
2.5, 2.6, 8.3, 8.4, 10.3, 10.4, and Appendix A2.

Status key: **excerpt** — the excerpt says it; **title** — the contents title names the
topic, the claimed detail is unchecked; **unchecked** — the title does not establish the
claim; **adaptation** — this repo's method, cited for inspiration, not a book claim.

## Lessons

| Note | Citation | Claim attributed | Status |
|---|---|---|---|
| [00](../lessons/00-the-inherited-system.md) | 1.1 "Two Sides of Testing", 1.2 "Essential Elements of Exploratory Testing" | Checked tests vs. exploration; a passing happy path is a starting point | title |
| [00](../lessons/00-the-inherited-system.md) | 11.1, 11.4 | Start an inherited system with a recon | title |
| [01](../lessons/01-map-it.md) | 11.1 "Starting with a Recon Session", 11.4 "Selecting Techniques" | Recon first, then choose probe families | title |
| [01](../lessons/01-map-it.md) | 12.4 "Active Reading" | Read the spec actively for silences | title (12.4 sits in "Explore Requirements"; applying it to an inherited spec is an adaptation) |
| [01](../lessons/01-map-it.md) | 2.1–2.4, 2.6 | Charters and a backlog of them | excerpt for 2.3–2.4; title for 2.1, 2.2, 2.6 |
| [02](../lessons/02-make-it-visible.md) | 3.1–3.4 | Unseen behavior; testability hooks; consoles and logs | title (3.1's title is "But Did You See the Moonwalking Bear?") |
| [03](../lessons/03-vary-it.md) | 4.1–4.4 | Find and vary variables, including subtle ones | title |
| [03](../lessons/03-vary-it.md) | A2.1 | The named heuristics (Zero/One/Many, Goldilocks, Boundaries, Strings, Numbers…) | Zero/One/Many: excerpt (Ch. 10 intro). A2.1's title is "General Heuristics"; the other names are unchecked |
| [03](../lessons/03-vary-it.md) | 10.1 "Exploring an API" | Vary inputs at an API without a UI | excerpt. "Does it reject what it does not understand?" is this repo's framing, not 10.1's wording |
| [04](../lessons/04-judge-it.md) | 5.1–5.3 | Never/Always rules, alternative resources, approximations as oracles | title |
| [04](../lessons/04-judge-it.md) | 10.3 | Status codes versus bodies | **unchecked**: 10.3 is "Exploring a Web Service"; the status-code point is not established |
| [05](../lessons/05-sequence-it.md) | 6.1 "Nouns and Verbs", 6.2 "Random Navigation" | Directed sequences from nouns × verbs; random walks | title |
| [06](../lessons/06-relate-it.md) | 7.1–7.3 | Entities, CRUD grid, following data | title (7.2 is "CRUD: Create, Read, Update, Delete") |
| [07](../lessons/07-state-it.md) | 8.1–8.2 | States, events, a state model | excerpt |
| [07](../lessons/07-state-it.md) | 8.3, 8.4 | Exploring with the model; state tables and sneak paths | title. ("Sneak paths" appears only in the plan's E07 heading, unchecked) |
| [08](../lessons/08-surround-it.md) | 9.1–9.4 | Ecosystem diagram, trust boundaries, what-if, following data out | title |
| [09](../lessons/09-explain-it.md) | 13.3 "Discovering Systemic Sources of Problems" | Group findings by cause | title |
| [09](../lessons/09-explain-it.md) | 13.3 | A habit that caused one bug predicts siblings; probe predictions | **adaptation** (corrected in the note; was worded as Hendrickson's) |
| [09](../lessons/09-explain-it.md) | 8.4 | The state table written as code | adaptation of a title-level citation |
| [09](../lessons/09-explain-it.md) | 13.5 "How to Tell When You Have Explored Enough" | When to stop | title. The findings-per-experiment signal is the plan's, as the note says |

## Skills

| Skill | Citation | Claim attributed | Status |
|---|---|---|---|
| [charter-as-probe](../skills/charter-as-probe/SKILL.md) | 2.2 "A Simple Charter Template" | *Explore … with … to discover …* | title; the template wording itself is unchecked |
| | 2.3 "Good Charters" | Neither too broad nor too narrow | excerpt |
| | 2.5 "The Nightmare Headline Game" | A source of risk hypotheses | title |
| | 1.3 "Working in Time-Boxed Sessions" | The time box as a budget | title |
| | 11.4 | Selecting techniques | title |
| [map-codebase-from-behavior](../skills/map-codebase-from-behavior/SKILL.md) | 11.1, 11.4, 12.4 | Recon steps, probe family, active reading | title |
| | 2.4, 2.6 | Generating charters (incl. from source and bug history); a backlog | excerpt for 2.4; title for 2.6 |
| | 13.3 | Cause model (step 10) | title; prediction step is adaptation |
| | 13.5 | Stop signal | title |
| [make-invisible-visible](../skills/make-invisible-visible/SKILL.md) | 3.1–3.4 | As lesson 02 | title |
| [sweep-variables](../skills/sweep-variables/SKILL.md) | 4.1–4.4, A2.1, 10.1 | As lesson 03 | as lesson 03 |
| [derive-oracles](../skills/derive-oracles/SKILL.md) | 5.1–5.3 | As lesson 04 | title |
| | 10.3 | Status codes versus bodies | **unchecked**, as lesson 04 |
| [probe-sequences-and-states](../skills/probe-sequences-and-states/SKILL.md) | 6.1, 6.2, 8.1–8.4 | As lessons 05, 07 | 8.1–8.2 excerpt; rest title |
| | 8 intro | (not cited before) Model from observable behavior | excerpt. Step 1 starts from the code's writers instead: an **adaptation**, now explained in the skill |
| | 10.4 "Characterizing Bugs" | Isolate a surprise to a minimal repro | title |
| [probe-entities-and-boundaries](../skills/probe-entities-and-boundaries/SKILL.md) | 7.1–7.3, 9.1–9.4 | As lessons 06, 08 | title |
| [follow-the-surprise](../skills/follow-the-surprise/SKILL.md) | 10.4, 11.5 "Recording Your Findings", 11.6 "What About the Dreaded Unreproducible Bugs?" | Characterize, record, handle unreproducible bugs | title |
| | 13.3 | Habits behind findings | title; habit signals are adaptation |

## What closing R1 takes (Nick)

With an authorized full copy of the P1.0 edition, for every row marked **title** or
**unchecked**: find the passage, add its page or e-book location, and mark it supported,
adaptation, or unsupported. Priority: 10.3 (lesson 04, derive-oracles), 13.3 and 13.5
(lesson 09, map, follow-the-surprise), 8.3–8.4 (lesson 07), A2.1's heuristic
names (lesson 03, sweep-variables). Correct any unsupported attribution in the authored
note or skill; leave raw explorer records as they are and add an erratum here. Until then
the review bar's item 4 is met only at section-title level.
