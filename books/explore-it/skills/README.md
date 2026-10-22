# Exploration skills

Eight procedures for learning a codebase you did not write by experimenting on it. Each was
drafted from *Explore It!* before any experiment, executed blind by at least one of queuelet's
nine experiments, and revised from what that run taught. Every skill ends with a worked example
on queuelet (real files, scores and findings) and a "Pitfalls the experiments exposed" section
listing the traps that changed its procedure.

| Skill | Use when | Executed by |
|---|---|---|
| [map-codebase-from-behavior](map-codebase-from-behavior/SKILL.md) | inheriting a service; first day in a repo; "how does this actually work?"; later, grouping findings by code habit | E01; the cause model in E09 |
| [charter-as-probe](charter-as-probe/SKILL.md) | deciding what to try first or next; a subsystem to learn or a risk to settle | every charter, E01–E09 |
| [make-invisible-visible](make-invisible-visible/SKILL.md) | you cannot tell what a job, request or loop did; consoles disagree; behaviour depends on time | E02 |
| [sweep-variables](sweep-variables/SKILL.md) | an API body, config or CLI with options; "what values should I try?" | E03 |
| [derive-oracles](derive-oracles/SKILL.md) | no spec; disputed correctness; a number nobody checks | E04 |
| [probe-sequences-and-states](probe-sequences-and-states/SKILL.md) | a status column, lease, retry or cancel protocol; events from several actors in any order | E05, E07 |
| [probe-entities-and-boundaries](probe-entities-and-boundaries/SKILL.md) | entities with dependents; data that crosses clients and partners; trust boundaries; "what if X goes away?" | E06, E08 |
| [follow-the-surprise](follow-the-surprise/SKILL.md) | "that's odd"; "I saw it once"; a race; a defect to write up; a list of findings to read for causes | every report; E09 |

A suggested order on an unfamiliar system: map → charter → make visible → sweep → oracles →
sequences and states → entities and boundaries, with follow-the-surprise whenever something odd
appears and again at the end to read the findings for their causes.

## Measured against the answer key

PLAN section 7 sets the bar: run blind against `explore-it/v0-inherited`, the skills should reveal
at least 19 of the 27 planted defects, and the measured number is whatever the nine runs found
without disclosure.

**Measured: 26 of 27**, all in the first run. Experiment 01, a recon that read each verb's code and
probed what the reading predicted, demonstrated 25 planted defects live and D23 (DST) through an
executed harness plus stored service state. The 27th, D07 (queue names that cannot be addressed),
was found only in part: charset evidence, without the slash/space manifestation, which lesson 01
discloses. Of the four planted spec silences, A2 and A3 were chartered from the plan and A4 was a
policy disclosure, so they do not count as blind finds; A1 was measured in E04, but an exported
lesson note had already named it. Beyond the register, the runs found 45 unplanned faults
(U01–U45) and one new spec silence (A5). Details: [defect index](../exploration/defects/INDEX.md),
[lesson 01](../lessons/01-map-it.md), [lessons index](../lessons/README.md).

What that number does and does not show: the explorer read the source, which the plan allows, and
code reading found most of the register at once. The later experiments' value is the 37 faults
found after the planted ones were gone, each by the probe family its skill teaches.

## Using them

- In this repo's arc, follow [the protocol](../exploration/PROTOCOL.md) as well: charters, probe
  logs, `revealed.md` scores and the blind export.
- The worked examples name queuelet's findings. To explore queuelet blind, export the tree at
  `explore-it/v0-inherited`, where these skills are the pre-experiment drafts.
- Book citations are by chapter and section number, checked against the publisher's table of
  contents; full book-text validation has not been done. Procedure wording is this repo's.

## Second targets and companions

Each skill sketches how to run it against litefleet, Release It!'s SUT. PLAN section 7 names three
of these as planned second targets: litefleet's tenant API (map), its placement decisions (make
visible) and its database lifecycle (sequences and states). None has been run yet. For a
neighbour what-if at load, Release It!'s `design-chaos-experiment` skill is the companion of
`probe-entities-and-boundaries`: this book produces the what-if list, that one runs an experiment
under load. That skill lives on `main` and is linked once this branch merges.
