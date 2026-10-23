# U15 — one queue's large retention halts retention everywhere

Original blind report: [E03-F05](E03-F05-large-retention-halts-global-retention.md).
Seen at `explore-it/lesson-02`; E03 notes H09. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity high.
Current status: fixed in lesson 03.

`retention_hours` is 0–1 000 000 on create and PATCH. The sweep (`internal/jobs/expire.go` `Retain`)
compares elapsed hours with the setting instead of subtracting an interval from now, so no stored
value, including a legacy one, can overflow the statement and stop the sweep for other tenants.

`lab/repro/U15.sh` (`probe03.py`) replays the report's steps through the public API only; `TestL03U15`
shares that oracle. See [reconciliation 03](../../register/reconciliations/03.md) for the spec decision.
