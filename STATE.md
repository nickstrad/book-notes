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
  `main`. Done: `release-it/v0-flawed` (platform, lab, lesson 00 tour; `topics/` stubs removed).
  Next: lessons L01–L12 in plan order, one tag each.
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
  - The verify runner restores a seeded snapshot (Postgres template database + tenant files) after
    the first seed, so a reset is ~25 s instead of ~1 min.
- **Phase 3 (skills):** not started; part of checkpoint C. Eight skills as in PLAN.md section 5.
- **Phase 4 (review):** not started. Reviewer from a different model family than the builder;
  final big review by the planner seat is also requested by Nick.

## Next action

Builder continues with lesson 01 (transparency). Cross-family review of each checkpoint per the plan.
