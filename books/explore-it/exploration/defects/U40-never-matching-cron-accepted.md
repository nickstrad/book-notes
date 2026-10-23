# U40 — never-matching-cron-accepted

Canonical status: **fixed in lesson 09**. Original blind report: [E09-F02](E09-F02-never-matching-cron-accepted-year-one.md).

Assigned in [E09 reconciliation](../../register/reconciliations/09.md).
Raw report remains immutable; repair status belongs here.

Repair: a never-matching cron is 400 `invalid_cron` on create and PATCH; disable-only PATCH always succeeds. Evidence: `project/lab/repro/U40.sh`, `TestL09U40`, [lesson 09](../../lessons/09-explain-it.md) and `register/runs/09-opus-*.log`.
