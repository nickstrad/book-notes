# U41 — cron-zone-prefix-overrides-timezone

Canonical status: **fixed in lesson 09**. Original blind report: [E09-F03](E09-F03-cron-zone-prefix-overrides-timezone.md).

Assigned in [E09 reconciliation](../../register/reconciliations/09.md).
Raw report remains immutable; repair status belongs here.

Repair: cron is exactly five fields; a `CRON_TZ=`/`TZ=` prefix is 400 `invalid_cron`. Evidence: `project/lab/repro/U41.sh`, `TestL09U41`, [lesson 09](../../lessons/09-explain-it.md) and `register/runs/09-opus-*.log`.
