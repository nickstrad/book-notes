# Release It! (2nd ed.) — topics and skills

> **STUB landing page** — the phase-2 builder expands this once topics run. Plan: [`PLAN.md`](PLAN.md).

*Release It! Design and Deploy Production-Ready Software*, 2nd edition, Michael T. Nygard (2018).

**The argument in one paragraph.** Software is done when it survives production, not when it passes
QA. Systems fail through their integration points and their coupling: a small fault becomes a crack
and propagates along whatever path the architecture leaves open. Learn the recurring ways cracks
start (stability antipatterns), build the structures that stop them spreading (stability patterns),
design processes so they can be seen and run (transparency, demand control), deploy without
downtime (phased changes, tolerant versions), and keep proving it with deliberate chaos.

## Topics

Run any topic with `./run.sh` from its folder; `run-all.sh` here runs them all.

| # | Folder | Idea | Language | Before/after |
|---|---|---|---|---|
| 01 | [`01-integration-points-timeouts`](topics/01-integration-points-timeouts/) | Remote calls hang, not fail; bound every call in time | Go | yes |
| 02 | [`02-blocked-threads-resource-pool`](topics/02-blocked-threads-resource-pool/) | A leaked or stuck resource freezes every thread; return on every path, check out with a timeout | Rust | yes |
| 03 | [`03-circuit-breaker`](topics/03-circuit-breaker/) | Stop calling a failing dependency; fail fast, probe for recovery | Rust | yes |
| 04 | [`04-bulkheads`](topics/04-bulkheads/) | Partition capacity so a slow path cannot starve a critical one | Go | yes |
| 05 | [`05-chain-reactions-shed-load`](topics/05-chain-reactions-shed-load/) | One instance dies, the rest follow; shed excess load to survive | Go | yes |
| 06 | [`06-retry-storms-dogpile`](topics/06-retry-storms-dogpile/) | Synchronized retries knock a recovering dependency over; backoff with jitter | Node | yes |
| 07 | [`07-unbounded-result-sets`](topics/07-unbounded-result-sets/) | Queries that return everything eventually return too much; bound and stream | Node | yes |
| 08 | [`08-back-pressure`](topics/08-back-pressure/) | Unbounded queues grow until death; slow the producer or shed at the boundary | Rust | yes |
| 09 | [`09-steady-state`](topics/09-steady-state/) | Everything that accumulates must be recycled: logs, sessions, caches | Bash | yes |
| 10 | [`10-test-harness`](topics/10-test-harness/) | A deliberately misbehaving fake dependency to test integration points | Go | yes (naive vs hardened client) |
| 11 | [`11-transparency-health-metrics`](topics/11-transparency-health-metrics/) | Health checks, metrics and structured logs make trouble visible before it is an outage | Go | yes |
| 12 | [`12-handling-versions`](topics/12-handling-versions/) | Old and new versions always coexist; tolerant readers and expand/contract rollouts | Node | yes |

## Skills

See [`skills/README.md`](skills/README.md). Planned: `test-integration-points`,
`hunt-stability-antipatterns`, `suggest-stability-patterns`, `budget-timeouts-and-retries`,
`review-capacity-headroom`, `review-operational-transparency`, `plan-safe-release`,
`design-chaos-experiment`.
