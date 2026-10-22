# U38 — frozen-database-hangs-requests

Original immutable report: [E08-F04](E08-F04-frozen-database-hangs-health-and-api.md).

Status: **fixed in lesson 08**; canonical repair status, separate from the immutable original report.

Regression: `TestL08U38`; public repro: `project/lab/repro/U38.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-08`.

`/health` (1 s) and authentication (5 s) are bounded independently of the driver and answer
503 when exceeded. Handlers past authentication and the loop are not bounded; a freeze that
starts mid-request still holds it (lesson 08, "What remains open").
Assigned in [E08 reconciliation](../../register/reconciliations/08.md).
