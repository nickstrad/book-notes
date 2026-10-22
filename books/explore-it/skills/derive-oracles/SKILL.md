---
name: derive-oracles
description: Make explicit the rules and comparisons used to judge behavior; use when specs are missing, correctness is disputed or numbers lack an independent check.
---

# Derive oracles

Draft for execution against queuelet. In this arc follow [the protocol](../../exploration/PROTOCOL.md); operate within the export and assigned charter.

## When to use

“Wrong” is unclear, a lifecycle needs invariants or an aggregate lacks a trusted expected value.

## Inputs

Spec/source, outputs, independently available data, relevant standards or comparable systems and question scope.

## Procedure

1. Write candidate Never/Always rules before probing. Define object, state, time window, exceptions and authority. Mark spec silences unresolved rather than silently deciding product behavior.
2. Name each oracle: a spec line, consistency rule, alternative resource or approximation. Cite the actual standard/comparable system and why it applies; another product’s behavior alone is not this product’s requirement.
3. For approximations choose a defensible bound before observing. State timing/sampling assumptions and what mismatch establishes. Prefer an independent calculation; two views sharing the same derivation may share its defect.
4. Find code enforcing each rule and write its runtime prediction before probing. A missing guard is a hypothesis until evidence demonstrates its effect.
5. Try to violate each rule with small inputs, sequences or states. Inspect status/body, exit code and effects together. A refuted prediction is not necessarily a product defect if the oracle remains uncertain.
6. Preserve agreed rules in rules.md. Route unresolved decisions to the owner; queuelet’s coordinator records spec decisions in fix commits. In the fix stage convert agreed rules into invariant checks, retaining old-failure and repaired-result evidence.

## Output

Oracle table: rule/scope, authority/reference, bounds, enforcing source, prediction, probe, observation, judgment and uncertainty; links to rules, decisions and checks.

## Source

Elisabeth Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (Pragmatic Bookshelf, 2013), sections 5.1–5.3 and 10.3. Section numbering checked against the publisher’s contents extract. This draft adapts the approved plan’s chapter map; detailed source-text review is pending. The procedure and local record formats are repo guidance, not quotations from the book.

## Worked example

Pending real experiments. E04 will connect written rules to runtime evidence and checks. This draft proves no invariants. No run results or scores are claimed yet.
