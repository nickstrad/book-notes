# U43 — terminal-notification-after-commit

Canonical status: **fixed in lesson 09**. Original blind report: [E09-F07](E09-F07-ack-cancel-error-after-commit-and-lost-notification.md).

Assigned in [E09 reconciliation](../../register/reconciliations/09.md).
Raw report remains immutable; repair status belongs here.

Repair: ack, cancel and nack → dead commit the transition, attempt row and outbox row together. Evidence: `project/lab/repro/U43.sh`, `TestL09U43`, [lesson 09](../../lessons/09-explain-it.md) and `register/runs/09-opus-*.log`.
