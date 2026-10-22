# U16 — backoff overflow retries immediately

Original blind report: [E03-F06](E03-F06-backoff-overflow-immediate-retry.md).
Seen at `explore-it/lesson-02`; E03 notes H08. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity medium.
Current status: fixed in lesson 03.

`jobs.Backoff` computes the wait in floating point and saturates at the queue cap or at
`MaxSeconds`, so it is nondecreasing in the attempt number and never wraps. Unit test
`TestBackoffSaturates`. A queue without `max_seconds` is still uncapped below that bound: A1 (a default
cap) stays with E04.

`lab/repro/U16.sh` (`probe03.py`) replays the report's steps through the public API only; `TestL03U16`
shares that oracle. See [reconciliation 03](../../register/reconciliations/03.md) for the spec decision.
