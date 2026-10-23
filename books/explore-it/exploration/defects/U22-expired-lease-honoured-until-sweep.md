# U22 — expired lease honoured until the sweep runs

Original blind report: [E04-F04](E04-F04-expired-lease-honoured-until-sweep.md).
Seen at `explore-it/lesson-03`; E04 notes. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low–medium.
Current status: fixed in lesson 04.

Heartbeat, ack and nack add `lease_expires_at > now` to their guards. A lease ends at its published expiry by the service clock, whether or not the sweep has run; a late request gets 409 and the sweep records the expired attempt.

Spec decision: no grace period (Q-E04-3).

`lab/repro/U22.sh` (`probe04.py`) replays the report's steps through the public API and the lab clock;
`TestL04U22` shares that oracle. See [reconciliation 04](../../register/reconciliations/04.md) for the
spec decision.
