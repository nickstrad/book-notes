# U14 — out-of-range settings and long names reported as 409 / 500

Original blind report: [E03-F04](E03-F04-storage-limit-errors-reported-as-conflict.md).
Seen at `explore-it/lesson-02`; E03 notes H06, H11. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low.
Current status: fixed in lesson 03.

Queue create and PATCH check `lease_seconds`, `max_attempts`, `backoff.base_seconds` and
`backoff.max_seconds` against 2 147 483 647 and names against 1–128 characters before the store sees
them (400 `invalid_queue`). Only a unique violation is 409 `queue_conflict`, with a service message;
other store errors are 500.

`lab/repro/U14.sh` (`probe03.py`) replays the report's steps through the public API only; `TestL03U14`
shares that oracle. See [reconciliation 03](../../register/reconciliations/03.md) for the spec decision.
