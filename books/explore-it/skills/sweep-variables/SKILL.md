---
name: sweep-variables
description: Generate deliberate value, count, timing and request-shape variations; use for an API, configuration surface or command with options.
---

# Sweep variables

Draft for execution against queuelet. In this arc follow [the protocol](../../exploration/PROTOCOL.md); operate within the export and assigned charter.

## When to use

Happy-path samples leave uncertainty about accepted values, batch behavior, configuration or rejection.

## Inputs

Bounded surface, spec and validators, raw client, resettable fixtures, observation tools and named oracle or oracle question.

## Procedure

1. Inventory variables before values: fields, type, presence, counts, lengths, sizes, position, selection, timing, order, configuration and request shape. Include omitted/null/empty where meaningful and dependencies between fields.
2. Read validators and serialization paths. Predict acceptance and downstream effects before each probe with source pointers; keep expected behavior separate from code predictions.
3. Apply fitting heuristics: Some/None/All, Zero/One/Many, Beginning/Middle/End, Goldilocks, Boundaries, Count, Position, Selection, Strings, Numbers and Paths. Derive concrete boundaries from this contract and representation; local thresholds are not universal book rules.
4. Use a raw instrument when a CLI normalizes or forbids inputs. Vary unknown fields, enum values, wrong types and malformed bodies where relevant. Capture status/body, exit code and persistent effects.
5. Change one variable at a time to explain a surprise, then selected interactions suggested by code or model. Reset as needed; retain failed probes and explanations.
6. Mark attempted and untried rows separately. Characterize surprises to minimal reproductions and backlog additional families when budget expires.

## Output

Variation table: variable/family, heuristic, value, oracle, code prediction, response/effect, verdict and evidence; variables.md updates and report/backlog links.

## Source

Elisabeth Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (Pragmatic Bookshelf, 2013), sections 4.1–4.4, 10.1 and A2.1. Section numbering checked against the publisher’s contents extract. This draft adapts the approved plan’s chapter map; detailed source-text review is pending. The procedure and local record formats are repo guidance, not quotations from the book.

## Worked example

Pending real experiments. E03 will supply the inventory and actual variation table. No accepted-value findings or coverage counts exist yet. No run results or scores are claimed yet.
