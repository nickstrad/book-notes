---
name: design-chaos-experiment
description: Design a chaos experiment as an executable card — prerequisites (monitoring, recovery path, steady-state definition, abort switch), a falsifiable hypothesis with thresholds, blast radius, the fault and how it is injected from outside the code, a timeline, what to observe, a before/after comparison, and how to automate and repeat it; plus a people-focused disaster-simulation variant. Use for "test our resilience", planning a game day, or proving a new stability pattern works.
---

# design-chaos-experiment

A chaos experiment is a controlled test of a hypothesis about how the system behaves when
something fails. It is not "break things and see": it starts from a defined steady state, injects
one fault from outside, measures against thresholds written down in advance, and ends with a
verdict. The best experiments are scripts that anyone can rerun before and after a change.

## When to use

- "How resilient are we?" or "would we survive X?"
- After adding a stability pattern (timeout, breaker, bulkhead, shed limit, drain): prove it works
  and keeps working.
- Planning a game day or a disaster-recovery exercise.
- Turning a postmortem's trigger into a regression test.

## Inputs

- A finding or a pattern to test (from `hunt-stability-antipatterns` or
  `suggest-stability-patterns`).
- Monitoring that can show the steady state and the fault's effect (if not:
  `review-operational-transparency` first).
- A way to inject faults from outside the service: a fault-injecting proxy, container or VM
  kill/pause/restart, a partner fake with admin modes, resource limits, SQL against the database.
- A load generator that can replay a realistic mix open-loop.

## Procedure

1. **Check the prerequisites; do not run without them.**
   - You can observe the steady state and the fault's effect in metrics and logs.
   - There is a recovery path, and you know how long it takes.
   - There is an abort switch (remove the fault, stop the load) and a person who will use it.
   - In production: start with opt-in or synthetic traffic and the smallest blast radius; never
     the first time.

2. **Define steady state in numbers.** Pick output metrics users feel, not internals: success rate,
   p99 latency, goodput, per class of user. Measure them before the fault in the same run. The
   steady-state window is the baseline the hypothesis compares against.

3. **Write the hypothesis with thresholds before running.** "When <fault> happens to <target>
   under <load>, <metric> stays <comparison> <number>, and <metric> recovers within <seconds>
   after the fault ends." Each threshold is a line the verdict will print. Include both a
   protection metric (others are unaffected) and a correctness metric where relevant (no duplicates,
   nothing lost).

4. **Bound the blast radius.** Which instances, tenants or users can be affected; what makes it
   impossible for the fault to spread beyond them; the maximum duration.

5. **Choose the injection and the target.** One fault at a time, from outside the code:

   | Fault | Injection |
   |---|---|
   | dependency slow | proxy `latency` (with jitter) |
   | dependency hangs | proxy `timeout` (accept, never answer) |
   | slow bytes / fragments / reset | `bandwidth`, `slicer`, `reset_peer` |
   | instance dies / freezes / restarts | `kill`, `pause`, `restart` (SIGTERM) |
   | partner misbehaves | fake's admin mode (5xx HTML, bad JSON, huge body, slow headers) |
   | data hazard | SQL that creates the condition (big table, lock holder, constraint violations) |
   | overload | offered rate as a fraction of the measured capacity |
   | resource wall | lower limits for the run (CPU, memory, `nofile`, disk size) |

   Never inject through a hidden code path in the service; then you are testing the test hook.

6. **Write the timeline.** Warm-up → baseline window → fault on at a fixed time → fault off →
   recovery window → end. Fixed times, shared clock between load generator and script.

7. **Plan the observations.** For each threshold: which query, over which window, read at which
   time. Prefer the load generator's own per-request record (status, latency, instance) for
   user-facing metrics; metrics and logs for internal ones. Print a comparable report table, then
   the verdict line.

8. **Run it twice: against the tree without the fix and with it.** The experiment is valid only if
   it **reproduces the problem** without the fix and **passes** with it. An experiment that passes
   both ways tests nothing.

9. **Automate and repeat.** Make it one command, idempotent (resets state first), a few minutes
   long, with a single verdict line. Run it before releases and after any change near the
   mechanism. Add it to a suite.

10. **Disaster-simulation variant (people).** For a game day: the same card, but the system under
    test includes the humans. Inject the fault without announcing which; measure time to detect,
    time to diagnose (did the dashboards and logs name the cause?), time to mitigate; record where
    the runbook or the tools failed them.

## Output

An experiment card:

```
Name / finding or pattern under test
Prerequisites: monitoring ✓  recovery path ✓  abort switch ✓
Steady state: <metrics, baseline window>
Hypothesis: <sentence>, thresholds: <metric op number> …
Blast radius: <targets, max duration>
Injection: <fault, injector, target>      Load: <mix, rate as % of capacity, open-loop>
Timeline: t=… baseline · t=… fault on · t=… fault off · t=… end
Observations: <metric ← query/window>
Verdict line: IMPROVED: … / FLAW REPRODUCED: …
Before/after: <command without the fix> / <command with it>
Repeat: <command, cadence>
```

## Source

*Release It!* 2nd ed.: Ch. 17 "Chaos Engineering" (prerequisites, designing the experiment,
injecting chaos, targeting chaos, automate and repeat, disaster simulations); Ch. 5 "Test
Harnesses" for the injectors; Ch. 3's chain of failure for choosing what to break. Every lesson in
this repo is one such experiment: `../../project/lab/verify/NN-*.sh`, run by
`make verify LESSON=NN [REF=<parent tag>]`.

## Worked example: litefleet's verify scripts

Every `project/lab/verify/NN-*.sh` is an experiment card made executable, built on
`project/lab/lib.sh`: `reset_stack` (restore the seeded snapshot: idempotent), `start_load`
(open-loop `cmd/loadgen`, one CSV row per request), `at <s>` (shared clock), `toxic`/`untoxic`
(Toxiproxy API), `compose kill/pause/restart`, `prom` (Prometheus), `q` (DuckDB over the CSV),
`report_table`, `check <name> <value> <op> <threshold>`, `verdict`. Step 8 is the runner's `REF=`:
it builds the parent tag's images and runs today's script against them.

The card for lesson 07 (`project/lab/verify/07-shed-load.sh`), transcribed:

```
Name: shed load — does the gateway tier survive losing one of three at 80 %?
Steady state: goodput (2xx within 3 s) per second, t=5–15 s
Hypothesis: after one gateway dies, survivors keep serving what two can; thresholds:
  goodput_after_kill_pct >= 55 · ok_p99_ms <= 300 · max_gateway_inflight <= 32 · rejoined_gateway_served > 0
Blast radius: lab stack; gateways only (GATEWAY_CPUS=0.15 makes them the bottleneck; NODE_CPUS=0.75)
Injection: docker compose kill gateway-b    Load: steady mix, 516 req/s = 80 % of 3 × 215, open-loop
Timeline: t=0 load · t=15 kill · t=58 read in-flight · t=60 start gateway-b · t=75 end
Observations: goodput windows from the CSV; p99 as the median of 2 s windows' p99; in-flight and
  goroutines from Prometheus; HAProxy UP count every 1 s
Before/after: make verify LESSON=07 REF=release-it/lesson-06 / make verify LESSON=07
Repeat: make verify LESSON=07 (≈2 min)
```

Result: `FLAW REPRODUCED: goodput_after_kill_pct 11.4 (want >= 55); ok_p99_ms 2784.0 (want <= 300);
max_gateway_inflight 997 (want <= 32)` → `IMPROVED: goodput_after_kill_pct 68.5 >= 55; ok_p99_ms
184.0 <= 300; max_gateway_inflight 13 <= 32; rejoined_gateway_served 1671 > 0`.

The other cards, by injection (thresholds and both verdicts are in each lesson note):

| Lesson | Injection | Key threshold, before → after |
|---|---|---|
| 01 | `latency 800ms` on node-a; `pause postgres` | control health fails ≤ 3 s: never → 0.82 s |
| 02 | `timeout` toxic on node-a; runaway `WITH RECURSIVE`; half-sent request | gateway in-flight ≤ 300: 1 130 → 149 |
| 03 | 30 % of writes violate `UNIQUE` | node fds ≤ 600: 1 024 → 471 |
| 04 | `latency 2500±1000ms` on node-a; hot database | duplicate rows == 0: 308 → 0 |
| 05 | `timeout` toxic on control for 30 s | known databases served ≥ 95 %: 0 % → 99.25 % |
| 06 | noisy tenant; 20 concurrent reports; blob store at 50 KB/s | other tenants' p99 ≤ 200 ms: 534 → 87 |
| 08 | whale `SELECT *` ten at a time | restarts == 0: 5 → 0 |
| 09 | metering `latency 1000ms`; `restart node-a` | events missing ≤ 5: 2 921 → 0 |
| 10 | 48 MB tmpfs volumes; 90 s of writes | disk-full errors == 0: 2 898 → 0 |
| 11 | six Toxiproxy faults; four fake modes (`make test-chaos`) | chaos tests failed == 0: 5 → 0 |
| 12 | migration + rolling restart + mixed wire versions | mixed-fleet query failures == 0 %: 99.9 % → 0 % |

What building them taught, as rules for step 5–8:

- **Injectors have their own failure modes.** Toxiproxy 2.12 wedged a proxy when a latency or
  downstream `timeout` toxic was removed under thousands of open links; the scripts use the
  upstream `timeout` toxic and remove faults by cycling the proxy (`untoxic_busy`). Validate that
  removing the fault really removes it.
- **Make the target the bottleneck, or the experiment proves nothing.** Lesson 07 at default CPU
  could not overload the gateways at all; the run sets `GATEWAY_CPUS=0.15` explicitly and records
  it in `lab/capacity.env`.
- **Choose metrics robust to a noisy host.** On a shared VM, one host stall spoiled a plain p99;
  the median of 2 s windows' p99 kept the verdict honest, with the worst window printed beside it.
- **Count from the source of truth.** Lesson 04 counts node-a's received requests from its own
  request log, because Prometheus counter increases over a 14 s window undercounted.
