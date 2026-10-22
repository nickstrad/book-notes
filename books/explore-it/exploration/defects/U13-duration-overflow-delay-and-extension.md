# U13 — delay, nack delay and heartbeat extension overflow into the past

Original blind report: [E03-F03](E03-F03-duration-overflow-delay-and-extension.md).
Seen at `explore-it/lesson-02`; E03 notes H04, H17, H18. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity medium.
Current status: fixed in lesson 03.

Enqueue `delay_seconds`, nack `delay_seconds` and heartbeat `extend_seconds` accept at most
2 147 483 647 (`jobs.MaxSeconds`); larger values are 400. Negative enqueue delay is unchanged
(spec silent, not claimed).

`lab/repro/U13.sh` (`probe03.py`) replays the report's steps through the public API only; `TestL03U13`
shares that oracle. See [reconciliation 03](../../register/reconciliations/03.md) for the spec decision.
