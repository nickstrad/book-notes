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
- **Phase 2 (build v0 + lessons):** done 2026-10-22, builder seat Claude Opus 5.5, linear history
  on `main`. Tags: `release-it/v0-flawed` (platform, lab, lesson 00 tour; `topics/` stubs removed),
  `release-it/lesson-01` (checkpoint A), `lesson-02` … `lesson-07` (checkpoint B), `lesson-08` …
  `lesson-12` (checkpoint C), `release-it/v1-hardened` (lessons index `lessons/README.md`, landing
  page). Every lesson's verify reproduces its flaw on the parent tag and is IMPROVED on its own.
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
  - Checkpoint C (details in each note): L08 whale threshold ≥ 99 % answered and p99 ≤ 1 s (plan
    100 ms; lesson 6's two read lanes queue the scans), gateway streams by transcoding. L09 spills
    the queue to the outbox on SIGTERM (the plan's check needs it), metering client 2 s. L10
    checkpoints any WAL past 64 KB, handle cache 128 (plan 64), 48 MB tmpfs volumes (plan 256 MB).
    L11 suite uses Toxiproxy's HTTP API (REF trees lack the Go client); 5 of 12 tests fail before
    (plan 4). L12: the gateway refuses during a node restart instead of retrying `draining` once
    (retries filled the shed limits and HAProxy), the gateways drain too, restart window ≤ 10 s with
    nothing cut off (plan ≤ 2 s; a restart is a new container), the contract is guarded and belongs
    to the next release, `make godot` stays outside the verify, and opening a database is a
    singleflight. Lessons 08, 11 and 12 are two commits each.
  - The verify runner restores a seeded snapshot (Postgres template database + tenant files) after
    the first seed, so a reset is ~25 s instead of ~1 min.
- **Phase 3 (skills):** done 2026-10-22, Claude Opus 5.5, commit `f27e8aa`. Eight
  `books/release-it/skills/<name>/SKILL.md` per PLAN.md section 5, each with a litefleet worked
  example (files, tags, measured before/after numbers); skills index and landing page updated.
- **Phase 4 (review):** Fable's final review done 2026-10-22: `books/release-it/REVIEW.md`.
  Verdict PASS with fixes recommended. Full verify matrix run at `main` and at every parent tag:
  every flaw reproduces on its parent tag; lessons 05–12 IMPROVED at `main`. Findings: 1 high
  (lesson 04's verify crashes at `main` and any tree from lesson 06 on, a seed-placement parity
  flip), 6 medium (lessons 01–03's verifies pass only at their own tags; lesson 12 mis-attributes
  the lab's Godot mechanism to the book; two before-symptoms need a loaded host; a cold-start
  Postgres race in `reset_stack`; no cross-family review; `v1-hardened` tag predates the skills),
  10 low. Nothing was fixed; Nick decides. The cross-family review question (AGENTS.md phase 4;
  PLAN.md section 6's Codex checkpoints) is still open.

## Next action

Nick reads `books/release-it/REVIEW.md` and decides which findings to fix (H1 first) and whether a
Codex/Muse review is added; then Twin verification and the report to Nick.
