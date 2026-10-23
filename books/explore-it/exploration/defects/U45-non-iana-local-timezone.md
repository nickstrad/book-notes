# U45 — non-iana-local-timezone

Canonical status: **fixed in lesson 09**. Original blind report: [E09-F09](E09-F09-non-iana-local-timezone-accepted.md).

Assigned in [E09 reconciliation](../../register/reconciliations/09.md).
Raw report remains immutable; repair status belongs here.

Repair: `timezone` accepts IANA identifiers only; `Local` and an empty PATCH value are 400. Evidence: `project/lab/repro/U45.sh`, `TestL09U45`, [lesson 09](../../lessons/09-explain-it.md) and `register/runs/09-opus-*.log`.
