# A5 — dst-gap-and-overlap

Original immutable report: [E08-F03](E08-F03-dst-gap-skips-and-overlap-double-fires.md).

Status: **resolved in lesson 08 (spec decision)**; canonical repair status, separate from the immutable original report.

Regression: `TestL08A5`, unit `TestNextAcrossDST`; public repro: `project/lab/repro/A5.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-08`.

Spec silence resolved: a fixed-hour schedule runs once per matching local time; a skipped
time runs once at the end of the gap; a repeated time runs at its first occurrence only;
wildcard-hour schedules follow elapsed time (SPEC, Schedules).
Assigned in [E08 reconciliation](../../register/reconciliations/08.md).
