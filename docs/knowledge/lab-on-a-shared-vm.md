# Running a load-and-chaos lab on a shared VM

Learned building and reviewing Release It! (`books/release-it/lessons/04`, `06`, `07`, `08`
"Deviations"; `books/release-it/REVIEW.md` M3 and L7).

- **CPU weights.** Other projects' load tests share the VM. Give the components that stand for
  "someone else's machine" (Postgres, control plane, fault injector) `cpu_shares: 4096`, the
  services under test 2048, and cap Postgres (`cpus: 1.5`) so its scans cannot starve the rest.
- **Make the target the bottleneck.** If the component the experiment is about is not the fleet's
  bottleneck, overloading it proves nothing. Lesson 07 sets `GATEWAY_CPUS=0.15` and
  `NODE_CPUS=0.75` for its run and records them in `lab/capacity.env`.
- **Thresholds that survive noise.** One host stall spoils a plain p99; the median of short
  windows' p99 (lesson 07) keeps a verdict honest, with the worst window printed beside it.
  Absolute lines set from one run leave thin margins (lesson 00's p99, lesson 08's report time).
- **Before-symptoms can need a loaded host.** A flaw that depends on contention (lesson 04's
  busy-loop on the write lock, lesson 06's noisy neighbour in absolute terms) can vanish on a quiet
  VM. Record the host's load next to the numbers, and prefer checks that show on both.
