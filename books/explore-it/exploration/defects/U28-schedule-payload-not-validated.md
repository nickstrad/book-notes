# U28 — E06 finding 2

Original immutable report: [E06-F02](E06-F02-schedule-payload-not-validated.md).

Status: **fixed in lesson 06**; canonical repair status, separate from the immutable original report.

Regression: `TestL06U28`; public repro: `project/lab/repro/U28.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-06`.

Review observation: this fixes new schedule creation. Existing invalid templates are
neither migrated nor revalidated; cron still copies their stored payload while manual
run applies enqueue validation. Legacy-template upgrade policy is a checkpoint-B
question for coordinator/Nick, not a claimed cleanup or another blind finding.

