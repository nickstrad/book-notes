# U42 — schedule-zone-offset-frozen

Canonical status: **fixed in lesson 09**. Original blind report: [E09-F06](E09-F06-schedule-zone-offset-frozen-in-record.md).

Assigned in [E09 reconciliation](../../register/reconciliations/09.md).
Raw report remains immutable; repair status belongs here.

Repair: `zone_offset` in schedule responses is computed for the time of the request. Evidence: `project/lab/repro/U42.sh`, `TestL09U42`, [lesson 09](../../lessons/09-explain-it.md) and `register/runs/09-opus-*.log`.
