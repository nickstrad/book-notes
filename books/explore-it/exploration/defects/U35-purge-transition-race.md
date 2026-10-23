# U35 — purge-transition-race

Original immutable report: [E07-F02](E07-F02-purge-misses-rows-changing-state-concurrently.md).

Status: **fixed in lesson 07**; canonical repair status, separate from the immutable original report.

Regressions: `TestL07U35`, `TestL07PurgeIndependentQueue`; public repro: `project/lab/repro/U35.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-07`.

Purge racing lease, nack, worker release and expiry ends every job deleted or marked
(dead on unsuccessful exit), never pending again. Sequential controls keep pending deletion
and leased-success semantics; purges of unrelated queues do not serialize.
Assigned in [E07 reconciliation](../../register/reconciliations/07.md).
