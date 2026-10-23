# Explore It! lessons

Run commands from `../project`. Each lesson tag points at a fix commit; the experiment commit
before it holds the blind explorer's record (charter, probe log, revealed, models, reports),
which no later commit rewrites. The canonical [defect index](../exploration/defects/INDEX.md)
carries repair status.

| Lesson | Tag | Skill run | Score | Findings and decisions | Verification |
|---|---|---|---|---|---|
| [00 — inherited system](00-the-inherited-system.md) | `explore-it/v0-inherited` | — | — | baseline | `make verify LESSON=00` |
| [01 — map it](01-map-it.md) | `explore-it/lesson-01` | [map-codebase-from-behavior](../skills/map-codebase-from-behavior/SKILL.md) | 48 / 44 / 2 / 1 / 27 / 110 | D01–D27 (D07 partly disclosed), U01–U08 | `make verify LESSON=01` |
| [02 — make it visible](02-make-it-visible.md) | `explore-it/lesson-02` | [make-invisible-visible](../skills/make-invisible-visible/SKILL.md) | 19 / 19 / 0 / 0 / 3 / 46 | U09–U10, T1–T3 | `make verify LESSON=02` |
| [03 — vary it](03-vary-it.md) | `explore-it/lesson-03` | [sweep-variables](../skills/sweep-variables/SKILL.md) | 23 / 21 / 2 / 2 / 8 / 28 | U11–U18 | `make verify LESSON=03` |
| [04 — judge it](04-judge-it.md) | `explore-it/lesson-04` | [derive-oracles](../skills/derive-oracles/SKILL.md) | 16 / 15 / 1 / 1 / 5 / 18 | U19–U23, A1 | `make verify LESSON=04` |
| [05 — sequence it](05-sequence-it.md) | `explore-it/lesson-05` | [probe-sequences-and-states](../skills/probe-sequences-and-states/SKILL.md) | 22 / 18 / 2 / 2 / 3 / 16 | U24–U26, A2 | `make verify LESSON=05` |
| [06 — relate it](06-relate-it.md) | `explore-it/lesson-06` | [probe-entities-and-boundaries](../skills/probe-entities-and-boundaries/SKILL.md) | 20 / 18 / 2 / 2 / 7 / 30 | U27–U33, A4 | `make verify LESSON=06` |
| [07 — state it](07-state-it.md) | `explore-it/lesson-07` | [probe-sequences-and-states](../skills/probe-sequences-and-states/SKILL.md) | 13 / 12 / 1 / 1 / 2 / 41 | U34–U35, A3 | `make verify LESSON=07` |
| [08 — surround it](08-surround-it.md) | `explore-it/lesson-08` | [probe-entities-and-boundaries](../skills/probe-entities-and-boundaries/SKILL.md) | 21 / 20 / 1 / 1 / 4 / 7 | U36–U38, A5 | `make verify LESSON=08` |
| [09 — explain it](09-explain-it.md) | `explore-it/lesson-09` | [follow-the-surprise](../skills/follow-the-surprise/SKILL.md) | 21 / 19 / 0 / 0 / 7 / 20 | U39–U45; transition table | `make verify LESSON=09` |
| final | `explore-it/v1-explored` | all eight, final form | — | landing page, final skills | `make checks` |

**Score** is the explorer's own count from each `revealed.md`, in its own units: hypotheses
stated / confirmed / refuted / code-reading predictions the running system contradicted /
reports filed / model cells filled. E01's two unresolved hypotheses ended *needs more*; E06's
two contradictions carry qualifications (see its note). Reports are not faults: one report can
cover several faults, and some reports are spec questions.

**What the numbers say.** The first experiment, a recon that read the code and probed what it
read, found 26 of the 27 planted defects; all 27 were fixed in lesson 01. Every later experiment
therefore found faults nobody planted (U09–U45) and decided five spec silences (A1–A5), while
the planted defects ran as early-fixed controls. New reports per experiment were
27 / 3 / 8 / 5 / 3 / 7 / 2 / 4 / 7: the stop signal did not fall. What changed is the route:
in E09, six of nine predictions written down before any probe produced reports.

Use `REF=<previous tag>` for the same public repros against historical binaries. See each
note for exact verdicts, measurement limits and deviations. The checkpoint reviews are under
[`../register/reviews/`](../register/reviews/).
