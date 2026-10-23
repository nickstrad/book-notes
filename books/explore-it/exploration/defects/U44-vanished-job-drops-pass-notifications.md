# U44 — vanished-job-drops-pass-notifications

Canonical status: **fixed in lesson 09**. Original blind report: [E09-F08](E09-F08-one-vanished-job-drops-other-dead-notifications.md).

Assigned in [E09 reconciliation](../../register/reconciliations/09.md).
Raw report remains immutable; repair status belongs here.

Repair: expiry and worker deletion notify each dead job inside their transaction; no post-commit loop. Evidence: `project/lab/repro/U44.sh`, `TestL09U44`, [lesson 09](../../lessons/09-explain-it.md) and `register/runs/09-opus-*.log`.
