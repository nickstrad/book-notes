---
name: probe-entities-and-boundaries
description: Probe entity dependencies and ecosystem boundaries through CRUD, data tracing and neighbour failures; use for owners, service neighbours or trust boundaries.
---

# Probe entities and boundaries

Draft for execution against queuelet. In this arc follow [the protocol](../../exploration/PROTOCOL.md); operate within the export and assigned charter.

## When to use

An entity has dependents, data appears through several paths or behavior depends on a neighbour/tenant/clock.

## Inputs

Schema/source, entity/API descriptions, observation tools, fixtures, neighbour controls and authorized local failure scope.

## Procedure

1. Draw entities, attributes, owners and dependencies from schema/behavior. Label database constraints versus application enforcement. Build an entity × CRUD grid with untested cells.
2. Predict dependent behavior when an owner changes or disappears. Probe and inspect dependents through API and read-only SQL. Compare refused, cascaded, retained or orphaned outcomes with a named oracle.
3. Follow a recognizable datum through input, storage, worker/output and external copies. Compare values/bytes and documented transformations; transformations need an oracle rather than automatic byte-equality judgments.
4. Draw neighbours, datastore, workers, actors/tenants and clocks. Mark trust boundaries, what crosses each and source validation pointers. Record predictions before boundary probes.
5. Choose bounded what-ifs: absence, restart, slow/error response, changed identity or time. Record baseline/action/evidence/recovery criteria. Apply failures only inside the permitted test environment and restore it afterwards.
6. Update entity/ecosystem models and characterize surprises, including siblings. For experiments at load, hand the what-if to an available chaos-experiment workflow; this procedure produces its question and evidence.

## Output

Entity diagram/CRUD grid, dependency outcomes, data-journey comparison, context diagram and boundary table with validation pointer, prediction, what-if, observation and recovery.

## Source

Elisabeth Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (Pragmatic Bookshelf, 2013), sections 7.1–7.3 and 9.1–9.4. Section numbering checked against the publisher’s contents extract. This draft adapts the approved plan’s chapter map; detailed source-text review is pending. The procedure and local record formats are repo guidance, not quotations from the book.

## Worked example

Pending real experiments. E06 will supply entity/data results and E08 ecosystem evidence. No neighbour or boundary defect is claimed. No run results or scores are claimed yet.
