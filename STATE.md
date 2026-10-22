# STATE

Repo-level status for `book-notes`. Update in the same commit as each phase deliverable.

## Books

### release-it — Release It! 2nd ed. (Nygard)

- **Phase 1 (research & plan):** v1 done 2026-10-21 (twelve standalone topic projects); Nick
  reviewed and redirected to one flawed project improved lesson by lesson. v2 written 2026-10-21
  (a `checkout` shop service as the vehicle); Nick approved the structure and rejected the vehicle.
  v3 written 2026-10-21 by the planner seat (Claude Fable 5.1): `books/release-it/PLAN.md`,
  re-vehicled to `litefleet`, a baby multitenant SQLite platform (18 flaws, 12 lessons).
  **Nick approved v3**; v2 archived as `PLAN-v2-checkout.md`, v1 as `PLAN-v1-per-topic-projects.md`.
- **Phase 2 (build v0 + lessons):** in progress, builder seat Claude Opus 5.5, linear history on
  `main`. Done: `release-it/v0-flawed` (platform, lab, lesson 00 tour; `topics/` stubs removed),
  `release-it/lesson-01` (checkpoint A), `release-it/lesson-02` … `lesson-07` (checkpoint B).
  Next: checkpoint C, lessons L08–L12, then the landing page and `release-it/v1-hardened`.
  Deviations from the plan so far (each also in the lesson note it concerns):
  - v0 + lab are one commit (checkpoints A1 and A2 merged): the plan requires `make verify
    LESSON=00` to pass before the v0 tag, and that needs the lab.
  - Node `mem_limit` 512 MB (plan: 384 MB), so a leaking node reaches the fd wall before OOM.
  - Whale `events` is 15 000 rows × 4 KB (~60 MB) instead of 500 000 small rows, so F13 stays
    reproducible after lesson 2's 2 s statement deadline.
  - Tenant files carry a `notes` table without a unique constraint (steady writes, lesson 4 audit);
    the ledger carries `statements`; report classes add `cold`, `ops`, `target`.
  - v0 control keeps 32 idle Postgres connections (open connections still unbounded), so v0 serves
    gentle load; the backup job spreads snapshots over its interval.
  - Calibrated on vm-1: gateway 300 req/s, node 250 req/s, fleet 500 req/s (`lab/capacity.env`).
  - L01 is two commits (~650 lines of Go, plan ~350): node health reports physical facts and the
    gateway health needs a prober.
  - Checkpoint B (details in each lesson note's "Deviations"): recalibrated at its start (gateway
    350, node 300, fleet 600 req/s). Toxiproxy 2.12 wedges a proxy when a latency/timeout toxic is
    removed under thousands of links, so faults use upstream `timeout` toxics and removal cycles
    the proxy. Lab CPU weights on the shared VM: Postgres, control, Toxiproxy 4096 (Postgres capped
    at 1.5 CPUs); gateways, nodes, loadgen 2048. L02 phase 1 reads only (F05 would leak cancelled
    writes). L03 control pool 20 (plan 10; every request still resolved). L04 adds a hot-database
    phase (the slow-node storm was already capped by L02's deadlines; duplicates were the damage).
    L05 is two commits. L06 2 reads per database plus a per-tenant cap of 4 (plan: 4 reads per
    database), the noisy tenant has 4 databases (seed: 243), creates get a 2.5 s gateway budget.
    L07 makes the gateways the bottleneck for its run (0.15 CPU, nodes 0.75) and uses a
    window-median p99 (≤ 300 ms).
  - The verify runner restores a seeded snapshot (Postgres template database + tenant files) after
    the first seed, so a reset is ~25 s instead of ~1 min.
- **Phase 3 (skills):** not started; part of checkpoint C. Eight skills as in PLAN.md section 5.
- **Phase 4 (review):** not started. Reviewer from a different model family than the builder;
  final big review by the planner seat is also requested by Nick.

## Next action

Builder continues with checkpoint C (lesson 08, unbounded result sets). Cross-family review of
checkpoints A and B per the plan.
