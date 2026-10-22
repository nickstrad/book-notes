# docs/knowledge — what building the books' labs taught

Durable findings about building and verifying this repo's lab projects, cut from the lesson notes'
"Deviations" sections and the final reviews. One topic per file; each names where it was learned.

- [toxiproxy-fault-removal.md](toxiproxy-fault-removal.md) — Toxiproxy 2.12 wedges a busy proxy
  when some toxics are removed; use upstream `timeout` and cycle the proxy.
- [lab-on-a-shared-vm.md](lab-on-a-shared-vm.md) — CPU weights, a bottleneck you choose, and
  thresholds that survive a noisy host.
- [lab-resets-and-seeds.md](lab-resets-and-seeds.md) — snapshot the seed, wait out Postgres's
  initdb, and select fixtures by property, not by name.
- [plan-size-estimates.md](plan-size-estimates.md) — Release It!'s planned against delivered
  lesson sizes, for the next plan's estimates.
