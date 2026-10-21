# Release It! skills — index

> **Planned, not yet written.** Phase 3 adds one `SKILL.md` per folder below following the skills
> contract in the repo's `AGENTS.md`. Full descriptions and sources: `../PLAN.md` section 4.

| Skill | Use it when | Procedure in one line |
|---|---|---|
| `test-integration-points/` | adding or reviewing a client to another service; "tests pass, production fails" | enumerate remote calls → list failure modes each must survive → check coverage → produce a harness-based test matrix and stubs |
| `hunt-stability-antipatterns/` | design or code review; postmortem; "why did this fall over" | walk the twelve Ch. 4 antipatterns with concrete signals → findings with evidence, blast radius, severity |
| `suggest-stability-patterns/` | after a hunt; designing an integration layer | map each finding to Ch. 5 patterns → placement → parameters → signal to expose |
| `budget-timeouts-and-retries/` | setting a timeout; writing retries; a new latency SLO | nested deadline budget from the edge inward → connect/response/total → idempotent, bounded, jittered retries paired with a breaker |
| `review-capacity-headroom/` | scaling a layer; sizing pools; planning a launch; autoscaling | layers and capacities → ratios and pool math → self-inflicted peaks → shed thresholds → headroom and governors |
| `review-operational-transparency/` | new service before first deploy; "we could not tell what was wrong" | health semantics → metrics → logs → config/version exposure → "would a slow dependency be visible before errors?" |
| `plan-safe-release/` | schema or API change; "can we deploy without downtime" | classify additive/breaking → expansion → rollout → contraction with tolerant-reader checks, DB sequence, rollback per phase |
| `design-chaos-experiment/` | "test our resilience"; game day; proving new stability patterns | prerequisites → hypothesis → blast radius → injection and targeting → observe → automate and repeat |
