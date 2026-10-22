---
name: follow-the-surprise
description: Characterize unexpected behavior, isolate variables and predict sibling defects; use for odd observations, intermittent reports or defect writeups.
---

# Follow the surprise

Draft for execution against queuelet. In this arc follow [the protocol](../../exploration/PROTOCOL.md); operate within the export and assigned charter.

## When to use

“That’s odd,” a one-off observation, console disagreement or a repeated implementation habit.

## Inputs

Original commands/output, tag/build identity, time/state/fixtures, prediction, named oracle and observation tools.

## Procedure

1. Preserve the original command/output, timestamp, ids, actor, configuration, timing and surrounding events before simplifying. Link the probe log. Every surprise must end in a report or written non-defect explanation.
2. State expected behavior and oracle, actual behavior and competing explanations. Distinguish a disproved code prediction from an oracle violation; surprise alone does not prove a defect.
3. Reproduce from known state, remove steps/data and vary one factor at a time. Record variables that mattered and those that did not. Keep the smallest explanatory reproduction with exact setup and observable result.
4. If it will not reproduce, retain exact original state/log, vary one thing at a time and investigate timing/clock sources. Report attempts, frequency and uncertainty without inventing a reliable sequence.
5. Read for siblings: record where the same cause could occur with file:line or “none found.” Group multiple findings by code habit; write predicted siblings and confirming probes before running them, then retain hits and misses.
6. Write title, seen-at tag, minimal steps, named oracle, expected/observed, isolated variables, sibling check, severity and status. Queuelet’s coordinator reconciles ids and fixer supplies repro scripts/checks. Preserve the original experiment record.
7. Update models/backlog. In a fix stage keep reproduction strong enough to fail before and pass after; non-reproduction is inconclusive until setup and observation are shown to work.

## Output

Characterized report or explicit non-defect explanation, reproduction evidence, uncertainty, sibling-prediction table, model delta and later regression-check link.

## Source

Elisabeth Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (Pragmatic Bookshelf, 2013), sections 10.4, 11.5, 11.6 and 13.3. Section numbering checked against the publisher’s contents extract. This draft adapts the approved plan’s chapter map; detailed source-text review is pending. The procedure and local record formats are repo guidance, not quotations from the book.

## Worked example

Pending real experiments. Real defect reports will supply characterized reproductions; E09 will supply predictions written before probing. No defect id or anticipated sibling is supplied here. No run results or scores are claimed yet.
