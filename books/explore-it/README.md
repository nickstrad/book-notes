# Explore It! — one inherited job queue, learned through nine experiments

> **Built.** `explore-it/v0-inherited`, nine lesson tags and `explore-it/v1-explored` are on branch
> `plan/explore-it`. Start with [`lessons/00-the-inherited-system.md`](lessons/00-the-inherited-system.md);
> the index with every experiment's score is [`lessons/README.md`](lessons/README.md). Plan:
> [`PLAN.md`](PLAN.md). Where things stand: [`STATE.md`](STATE.md).

*Explore It! Reduce Risk and Increase Confidence with Exploratory Testing*, Elisabeth Hendrickson
(Pragmatic Bookshelf, 2013). This breakdown keeps the book's method for probing and learning a
system and leaves out its people-and-process layer; [PLAN section 0](PLAN.md#0-what-this-plan-deliberately-omits-and-why-the-rest-is-the-books-real-core)
lists what was cut and why.

**The argument in one paragraph.** Testing is two activities. Checking confirms what you already
decided the software should do. Exploring designs an experiment, runs it and learns from the result
in one motion, so each observation shapes the next probe. When you inherit a codebase, exploring is
the only activity available on day one: there is nothing to check against yet. Give each probe a
charter with a prediction and a way to be wrong; watch everything the system reveals, and add hooks
where it hides too much; name the variables and vary them on purpose; write down what "wrong" means
before judging; build models of sequences, states, entities and neighbours, because their empty
cells are the next probes; follow every surprise to a minimal reproduction and its siblings; and
read the findings for the habits that produced them.

## The project: `queuelet`, and the record beside it

`queuelet` is a small Go job queue over Postgres 17: producers enqueue JSON jobs into named queues
inside a namespace, workers lease batches with a timeout, heartbeat, ack or nack; the service
retries with backoff, dead-letters, fires cron schedules in IANA time zones, purges, and notifies a
webhook partner (`hooksink`). There is an HTTP API and a CLI, `qctl`, and no UI. At
`explore-it/v0-inherited` it has a spec, thin tests and a README, passes its own baseline, and
carries 27 planted defects, four spec silences and three testability gaps that nobody on the
"team" has looked for.

Beside it grows the **exploration record** ([`exploration/`](exploration/)): the protocol, the
hypothesis backlog, eight living models (map, nouns and verbs, variables, rules, entities, states,
ecosystem, causes) and every defect report as the explorer wrote it. Each experiment was run by a
fresh agent from a plain export of the tree with the source but without the plan, the register or
git history. That boundary was a prompt and export boundary, not OS isolation, and it leaked in
documented places: an exported lesson note named A1 to E04, E09's charter supplied the expected
cause groups, and E04, E07 and E09 read their own runtime output outside the export. The
[exceptions table](register/blindness-exceptions.md) lists each and what it does and does not
change; E01's 26/27 below precedes all of them.

## What happened

The planned arc was a learning curve: shallow defects early, deep ones late. The record shows
something else, and the lessons say so instead of staging the plan.

- **Experiment 01's recon found 26 of the 27 planted defects** (D07 only partly) by reading the code
  for each verb and probing what it predicted. All 27 were fixed in lesson 01. The skills index
  records this as the measured blind find rate against the register: 26 of 27, against the plan's
  bar of 19.
- **Experiments 02–09 then found 37 faults nobody planted** (U09–U45) and settled five spec
  silences (A1–A5), each with the probe family its chapter teaches: hooks, variable sweeps,
  oracles, sequences and states, entities, neighbours, causes. The planted defects ran as
  early-fixed controls, ABSENT at every later tag.
- **The stop signal did not fall.** New reports per experiment: 27 / 3 / 8 / 5 / 3 / 7 / 2 / 4 / 7.
  What E09 measured was prediction yield: six of nine sites predicted before probing produced
  reports, plus a sibling finding. No comparable series exists for E01–E08, so the record does not
  show whether the predicted share of findings rose; E01 already wrote predictions before its probes.
- **The structural fix landed.** The missing-guard habit that produced eight planted defects is now
  one transition table (`project/internal/jobs/transition.go`) that every state-changing statement
  takes its guard from, held by a contract test and a 1 000-step model-based walk.

## Run it

```
cd books/explore-it/project
make up && make seed
make verify LESSON=00                                    # BASELINE HOLDS
git checkout explore-it/lesson-07
make verify LESSON=07 REF=explore-it/lesson-06           # DEFECTS REPRODUCED (historical binaries)
make verify LESSON=07                                    # FIXED AND CHECKED
make checks                                              # every regression check, about 12 minutes
make down
```

A verify resets the fixtures, runs the lesson's black-box repro scripts (API, `qctl`, read-only
SQL, the lab clock) and then its regression checks. Most take one to five minutes; race campaigns
and real minute ticks dominate. To run beside another stack, set `COMPOSE_PROJECT_NAME`, the
`QUEUELET_*_PORT` variables and `QUEUELET_DATABASE_URL` as the [project README](project/README.md)
describes.

## Lessons

| # | Tag | Lesson | Findings | Book |
|---|---|---|---|---|
| 00 | `explore-it/v0-inherited` | [The inherited system](lessons/00-the-inherited-system.md) | — | 1.1–1.2, 11 |
| 01 | `explore-it/lesson-01` | [Map it: recon of an inherited codebase](lessons/01-map-it.md) | D01–D27, U01–U08 | 11.1, 11.4, 12.4, 2 |
| 02 | `explore-it/lesson-02` | [Make it visible: consoles, logs and hooks](lessons/02-make-it-visible.md) | U09–U10, T1–T3 | 3 |
| 03 | `explore-it/lesson-03` | [Vary it: inputs and the request shape](lessons/03-vary-it.md) | U11–U18 | 4, A2.1, 10.1 |
| 04 | `explore-it/lesson-04` | [Judge it: write the oracles down, then break them](lessons/04-judge-it.md) | U19–U23, A1 | 5, 10.3 |
| 05 | `explore-it/lesson-05` | [Sequence it: verbs in orders nobody intended](lessons/05-sequence-it.md) | U24–U26, A2 | 6.1, 6.2 |
| 06 | `explore-it/lesson-06` | [Relate it: entities, dependents and the data's journey](lessons/06-relate-it.md) | U27–U33, A4 | 7 |
| 07 | `explore-it/lesson-07` | [State it: the state table and its sneak paths](lessons/07-state-it.md) | U34–U35, A3 | 8 |
| 08 | `explore-it/lesson-08` | [Surround it: neighbours, boundaries and the clock](lessons/08-surround-it.md) | U36–U38, A5 | 9 |
| 09 | `explore-it/lesson-09` | [Explain it: the pattern, the prediction and the cause](lessons/09-explain-it.md) | U39–U45, transition table | 13.3, 8.4 |
| — | `explore-it/v1-explored` | Final skills, this page and the lessons index | | |

The register (the answer key) is [PLAN section 4](PLAN.md#4-the-deliberate-defect-register); the
canonical status of every finding is the [defect index](exploration/defects/INDEX.md). Each note
ends with its deviations from the plan.

## Skills

Eight procedures for exploring a codebase you did not write, each executed by an experiment here
and revised from what it taught. Index with triggers: [`skills/README.md`](skills/README.md).

| Skill | Use it when |
|---|---|
| [map-codebase-from-behavior](skills/map-codebase-from-behavior/SKILL.md) | you inherited a service and need to know how it actually works |
| [charter-as-probe](skills/charter-as-probe/SKILL.md) | you need to decide what to try first, and how you would know you were wrong |
| [make-invisible-visible](skills/make-invisible-visible/SKILL.md) | you cannot tell what the system did, or time-based behaviour can only be waited for |
| [sweep-variables](skills/sweep-variables/SKILL.md) | an input surface, request body or config needs more than a few happy values |
| [derive-oracles](skills/derive-oracles/SKILL.md) | nobody wrote down what "wrong" means |
| [probe-sequences-and-states](skills/probe-sequences-and-states/SKILL.md) | something has a status column, a lifecycle or verbs that can arrive in any order |
| [probe-entities-and-boundaries](skills/probe-entities-and-boundaries/SKILL.md) | entities own dependents, data travels through several paths, or a neighbour can misbehave |
| [follow-the-surprise](skills/follow-the-surprise/SKILL.md) | "that's odd", a flaky report, or a list of findings to read for their causes |

The final skills' worked examples name queuelet's findings. To explore blind, export the tree at
`explore-it/v0-inherited`, where the skills are the pre-experiment drafts.

## Reviews

Checkpoint reviews are under [`register/reviews/`](register/reviews/): A (v0 and E01), B (E02–E06)
and C (E07–E09 and the full register). Lessons 07–09 were built by Claude (Opus) after the Codex
seat was interrupted, so their independent cross-family review is the end review, still to run.
