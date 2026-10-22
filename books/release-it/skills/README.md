# Release It! skills — index

One `SKILL.md` per folder below, in the format Claude Code and Codex load, following the skills
contract in the repo's `AGENTS.md`. Each is a procedure you can run on your own code or design, and
each ends with a worked example on litefleet: real files, lesson tags, and the measured before and
after numbers. The set and its sources: `../PLAN.md` section 5 (original descriptions in the
archived `../PLAN-v1-per-topic-projects.md` section 4).

A natural order: `review-operational-transparency` (can you see it?) → `hunt-stability-antipatterns`
→ `suggest-stability-patterns` → the focused reviews (`budget-timeouts-and-retries`,
`review-capacity-headroom`, `plan-safe-release`) → prove the result with `test-integration-points`
and `design-chaos-experiment`.

| Skill | Use it when | Procedure in one line | Worked example on litefleet |
|---|---|---|---|
| [`test-integration-points`](test-integration-points/SKILL.md) | adding or reviewing a client to another service; "tests pass, production fails" | enumerate remote calls → list failure modes each must survive → check coverage → produce a harness-based test matrix and stubs | lesson 11's harness matrix: 5 of 12 chaos tests fail before, 0 after |
| [`hunt-stability-antipatterns`](hunt-stability-antipatterns/SKILL.md) | design or code review; postmortem; "why did this fall over" | walk the twelve Ch. 4 antipatterns with concrete signals → findings with evidence, blast radius, severity | run on `release-it/v0-flawed`; the register F01–F18 is the answer key (≥ 13 of 18) |
| [`suggest-stability-patterns`](suggest-stability-patterns/SKILL.md) | after a hunt; designing an integration layer | map each finding to Ch. 5 patterns → placement → parameters → signal to expose | the register mapped to lessons 02–12 with placement, parameters, fallback, signal |
| [`budget-timeouts-and-retries`](budget-timeouts-and-retries/SKILL.md) | setting a timeout; writing retries; a new latency SLO | nested deadline budget from the edge inward → connect/response/total → idempotent, bounded, jittered retries paired with a breaker | lessons 02 and 04: 2.9 s ⊃ 2.5 s ⊃ 2 s; in-flight 1 130 → 149; duplicate rows 308 → 0 |
| [`review-capacity-headroom`](review-capacity-headroom/SKILL.md) | scaling a layer; sizing pools; planning a launch; autoscaling | layers and capacities → ratios and pool math → self-inflicted peaks → shed thresholds → headroom and governors | `make calibrate`; lesson 07's gateway kill (goodput 11.4 % → 68.5 %); lesson 10's growth table |
| [`review-operational-transparency`](review-operational-transparency/SKILL.md) | new service before first deploy; "we could not tell what was wrong" | health semantics → metrics → logs → config/version exposure → "would a slow dependency be visible before errors?" | lesson 01's metrics, logs and three health states; the health rules lessons 05 and 06 changed |
| [`plan-safe-release`](plan-safe-release/SKILL.md) | schema or API change; "can we deploy without downtime" | classify additive/breaking → expansion → rollout → contraction with tolerant-reader checks, DB sequence, rollback per phase | lesson 12: expand → rollout with drain → guarded contract; `make godot` stalls control for 36 s |
| [`design-chaos-experiment`](design-chaos-experiment/SKILL.md) | "test our resilience"; game day; proving new stability patterns | prerequisites → hypothesis → blast radius → injection and targeting → observe → automate and repeat | every `project/lab/verify/NN-*.sh`; lesson 07's script transcribed as a card |
