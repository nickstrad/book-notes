# U37 — webhook-shipper-starvation

Original immutable report: [E08-F02](E08-F02-slow-destination-starves-all-later-notifications.md).

Status: **fixed in lesson 08**; canonical repair status, separate from the immutable original report.

Regression: `TestL08U37`; public repro: `project/lab/repro/U37.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-08`.

Each pass attempts up to 20 due rows concurrently, ordered by `next_at, id`, so a pass costs
at most one client deadline and failing rows move behind untried rows. Deadline, retry
interval, at-least-once and the absence of an attempt cap are unchanged. Sibling
H-E08-23 (`schedule.Fire` returns on the first failing schedule) remains a forward question.
Assigned in [E08 reconciliation](../../register/reconciliations/08.md).
