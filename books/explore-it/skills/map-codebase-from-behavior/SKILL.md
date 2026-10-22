---
name: map-codebase-from-behavior
description: Map an inherited codebase through source reading and behavioral recon; use on the first day in an unfamiliar repository.
---

# Map codebase from behavior

Draft for execution against queuelet. In this arc follow [the protocol](../../exploration/PROTOCOL.md); operate within the export and assigned charter.

## When to use

Inheriting a service, finding its entry points, or asking how it actually works.

## Inputs

A runnable local system, source, startup instructions, spec, tests, exploration record and time budget.

## Procedure

1. Set a recon boundary and budget. Read startup instructions, spec and test entry points; distinguish documented promises, assumptions and behavior demonstrated by existing checks.
2. Trace entry points into packages, persistence and background work. Record each noun and verb, its handler and source pointers in map.md and nouns-verbs.md. A source pointer predicts a path; it does not prove the path executes.
3. Predict a representative end-to-end path before running it, then walk documented verbs within the budget. Retain commands, responses, exit status and persistent effects beside source-based predictions.
4. Read spec sentences for silences and ways they could be false. Find where code decides the missing behavior. Write each prediction, its source and refutation condition before probing.
5. Rank remaining hypotheses by consequence and uncertainty. Name a fitting probe family: variation, oracle, sequence/state, entity/data journey or ecosystem/boundary. Backlog off-charter questions.
6. Keep models proportional to evidence: start with map and nouns/verbs, then add variables, rules, entities, states and ecosystem. Mark inferred and unprobed cells. Once findings exist, group causes and predict a sibling location before probing it.

## Output

A source-and-behavior map, nouns/verbs list and ranked hypothesis backlog with source, prediction, refutation condition, probe family, status and evidence links.

## Source

Elisabeth Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (Pragmatic Bookshelf, 2013), sections 11.1, 11.4, 12.4, 2.4, 2.6 and 13.3. Section numbering checked against the publisher’s contents extract. This draft adapts the approved plan’s chapter map; detailed source-text review is pending. The procedure and local record formats are repo guidance, not quotations from the book.

## Worked example

Pending real experiments. E01 will supply the recon and initial models; E09 the causal map. A second run against litefleet’s tenant API awaits both arcs. No run results or scores are claimed yet.
