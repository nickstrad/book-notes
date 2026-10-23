# U09 — documented sample kind is not registered

Original blind report: [E02-F01](E02-F01-sample-worker-ignores-failN-kinds.md).
Seen at `explore-it/lesson-01`; E02 notes H10. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low.
Current status: fixed in lesson 02.

The original report supplies the minimal steps, README/SPEC oracle, observed
80-second nonexecution, isolated variables and sibling prediction. The repair
advertises `fail3` alongside `fail` in the default standalone worker and qctl
sample-worker mode, and makes the documented explicit list match. Other numbered
kinds require explicit registration; queue-service kind matching stays exact.

`lab/repro/U09.sh` compares fail3 with a successful echo on the same default
Compose worker, then separately runs the actual container qctl worker defaults
for eight seconds on an isolated queue. This turns the original CLI sibling
inference into a measured regression, without changing the original explorer's
find count. `TestL02U09` repeats the same oracle. Acceptance and measured evidence:
[lesson 02](../../lessons/02-make-it-visible.md), `register/runs/02-*.log`.
