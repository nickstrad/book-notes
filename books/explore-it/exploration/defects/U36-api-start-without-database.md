# U36 — api-start-without-database

Original immutable report: [E08-F01](E08-F01-api-does-not-recover-from-start-without-database.md).

Status: **fixed in lesson 08**; canonical repair status, separate from the immutable original report.

Regression: `TestL08U36`; public repro: `project/lab/repro/U36.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-08`.

The API waits for an unreachable database at startup and serves once it answers (main
startup ping loop). The named siblings, API and PostgreSQL processes that die on their
own, are covered by `restart: unless-stopped` and a host-kill check in the same repro.
Assigned in [E08 reconciliation](../../register/reconciliations/08.md).
