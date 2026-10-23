# U34 — worker-delete-racing-lease

Original immutable report: [E07-F01](E07-F01-worker-delete-racing-lease-leaves-zombie-lease.md).

Status: **fixed in lesson 07**; canonical repair status, separate from the immutable original report.

Regressions: `TestL07U34`, `TestL07WorkerMissingInsideLease`; public repro: `project/lab/repro/U34.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-07`.

Both symptoms are covered: no lease survives for a deleted worker (lease first → released
with one attributable attempt; delete first → 404 `worker_not_found` without mutation), and
a worker disappearing inside the lease transaction is 404, not 500 `sql: no rows`.
Assigned in [E07 reconciliation](../../register/reconciliations/07.md).
