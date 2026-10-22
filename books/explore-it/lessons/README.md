# Explore It! lessons

Run commands from `../project`. Each lesson tag includes its separate exploration
commit and fix commit. Original explorer reports remain immutable; the canonical
[defect index](../exploration/defects/INDEX.md) carries repair status.

| Lesson | Tag | Findings / decisions | Verification |
|---|---|---|---|
| [00 — inherited system](00-the-inherited-system.md) | `explore-it/v0-inherited` | baseline | `make verify LESSON=00` |
| [01 — map it](01-map-it.md) | `explore-it/lesson-01` | D01–D27, U01–U08 | `make verify LESSON=01` |
| [02 — make it visible](02-make-it-visible.md) | `explore-it/lesson-02` | U09–U10, T1–T3 | `make verify LESSON=02` |
| [03 — vary it](03-vary-it.md) | `explore-it/lesson-03` | U11–U18 | `make verify LESSON=03` |
| [04 — judge it](04-judge-it.md) | `explore-it/lesson-04` | U19–U23, A1 | `make verify LESSON=04` |
| [05 — sequence it](05-sequence-it.md) | `explore-it/lesson-05` | U24–U26, A2 | `make verify LESSON=05` |
| [06 — relate it](06-relate-it.md) | `explore-it/lesson-06` | U27–U33, A4 | `make verify LESSON=06` |
| [07 — state it](07-state-it.md) | `explore-it/lesson-07` | U34–U35, A3 | `make verify LESSON=07` |
| [08 — surround it](08-surround-it.md) | `explore-it/lesson-08` | U36–U38, A5 | `make verify LESSON=08` |

Use `REF=<previous tag>` for the same public repros against historical binaries.
See each note for exact verdicts and measurement limits. This index records
completed lessons only.
