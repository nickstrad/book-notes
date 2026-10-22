# U17 — backoff.max_seconds stored but never applied

Original blind report: [E03-F07](E03-F07-backoff-max-seconds-ignored.md).
Seen at `explore-it/lesson-02`; E03 notes H07. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low–medium.
Current status: fixed in lesson 03.

Nack reads `backoff_max` and passes it to `Backoff` as the cap. Create rejects a cap below 1, below
`base_seconds` or above `MaxSeconds`. The field is now in SPEC.md. It is still not PATCHable; the
report did not claim that.

`lab/repro/U17.sh` (`probe03.py`) replays the report's steps through the public API only; `TestL03U17`
shares that oracle. See [reconciliation 03](../../register/reconciliations/03.md) for the spec decision.
