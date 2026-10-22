---
name: probe-sequences-and-states
description: Explore lifecycles with unusual verb sequences, replayable random walks and state/event tables; use when events can arrive in different orders.
---

# Probe sequences and states

Draft for execution against queuelet. In this arc follow [the protocol](../../exploration/PROTOCOL.md); operate within the export and assigned charter.

## When to use

A status-bearing object, retry/cancel/timeout protocol or operator/worker interaction.

## Inputs

Nouns/verbs, spec lifecycle and rules, source guards, fixtures, relevant actors/clocks and charter.

## Procedure

1. List nouns and verbs. Select unlikely pairs/triples, repetitions and interruptions. Predict effects before driving each sequence; record starting state and actor identity.
2. Check agreed rules after every step. For random walks retain seed, operation log, preconditions and observations for replay. A skipped operation is not coverage.
3. Derive states/events from spec and observation, including background actions and timeouts. Diagram them and label inferred states and unobserved arrows.
4. Expand into state × event cells with should (named oracle), predicted (source/guards), observed and evidence. Keep unknown distinct from forbidden. Include ownership/actor context when it changes behavior.
5. Drive rejected and undefined cells, prioritizing disagreement between should and predicted. Use multiple actors or controlled timing where needed; demonstrate prerequisite state instead of assuming a sleep reached it.
6. Reduce failing walks to shortest explanatory sequences, preserving needed identities/timing. Update the model and reports. The fixer turns agreed cells into model-based checks; the explorer does not silently repair the SUT.

## Output

Sequence grid, replayable walk log, state/event table with should/predicted/observed; separate filled, untried and blocked cells; minimal failures and model delta.

## Source

Elisabeth Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (Pragmatic Bookshelf, 2013), sections 6.1, 6.2 and 8.1–8.4. Section numbering checked against the publisher’s contents extract. This draft adapts the approved plan’s chapter map; detailed source-text review is pending. The procedure and local record formats are repo guidance, not quotations from the book.

## Worked example

Pending real experiments. E05 will provide a walker and E07 a state table. Litefleet’s database lifecycle is a second target after both arcs. No transitions are pre-filled. No run results or scores are claimed yet.
