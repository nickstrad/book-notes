# U21 — negative delay_seconds accepted on enqueue and nack

Original blind report: [E04-F03](E04-F03-negative-delay-wraps-and-precedes-failure.md).
Seen at `explore-it/lesson-03`; E04 notes. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity medium.
Current status: fixed in lesson 04.

Enqueue and nack reject a negative `delay_seconds` with 400, next to the existing upper bound. That removes both the "retry before the failure" case and the wrap below −9 223 372 036 s.

Reconciliation 03 had left negative delays accepted as "due now" because nobody had claimed them; E04 claimed them. `run_at` in the past remains accepted (no arithmetic, not claimed).

`lab/repro/U21.sh` (`probe04.py`) replays the report's steps through the public API and the lab clock;
`TestL04U21` shares that oracle. See [reconciliation 04](../../register/reconciliations/04.md) for the
spec decision.
