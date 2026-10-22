# Exploration skills — drafts

These eight procedures follow the plan’s chapter map. Section numbering has been checked against the publisher’s table of contents; detailed source-text review and real execution remain pending. Worked examples will be filled from actual records, including misses and refuted predictions.

| Skill | Use when |
|---|---|
| [map-codebase-from-behavior](map-codebase-from-behavior/SKILL.md) | Inheriting a service, finding its entry points, or asking how it actually works. |
| [charter-as-probe](charter-as-probe/SKILL.md) | A subsystem to learn, a risk to settle or an exploration request too broad for one session. |
| [make-invisible-visible](make-invisible-visible/SKILL.md) | An operation’s effects are unclear, lifecycle timing cannot be controlled or observers disagree. |
| [sweep-variables](sweep-variables/SKILL.md) | Happy-path samples leave uncertainty about accepted values, batch behavior, configuration or rejection. |
| [derive-oracles](derive-oracles/SKILL.md) | “Wrong” is unclear, a lifecycle needs invariants or an aggregate lacks a trusted expected value. |
| [probe-sequences-and-states](probe-sequences-and-states/SKILL.md) | A status-bearing object, retry/cancel/timeout protocol or operator/worker interaction. |
| [probe-entities-and-boundaries](probe-entities-and-boundaries/SKILL.md) | An entity has dependents, data appears through several paths or behavior depends on a neighbour/tenant/clock. |
| [follow-the-surprise](follow-the-surprise/SKILL.md) | “That’s odd,” a one-off observation, console disagreement or a repeated implementation habit. |

Read [the protocol](../exploration/PROTOCOL.md) before a queuelet experiment. Measured blind discovery count: **not yet available**. The final count comes from actual runs, excluding disclosures. No expected-result list belongs in these skills.

Second targets after both arcs land: litefleet’s tenant API, placement decisions and database lifecycle. For a neighbour what-if at load, the planned companion is Release It!’s `design-chaos-experiment`; its cross-link awaits that skill landing. The boundary skill here produces the what-if question.
