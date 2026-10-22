# U33 — E06 finding 7

Original immutable report: [E06-F07](E06-F07-attempts-outlive-their-job.md).

Status: **fixed in lesson 06**; canonical repair status, separate from the immutable original report.

Regression: `TestL06U33`; public repro: `project/lab/repro/U33.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-06`.

Coordinator decision: attempt records belong to their job and are removed with it on explicit deletion, purge and retention; job_events and pending webhook snapshots keep their documented independent lifetime.
