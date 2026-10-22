# U20 — attempt numbers reused after an operator retry

Original blind report: [E04-F02](E04-F02-attempt-history-after-retry-ambiguous.md).
Seen at `explore-it/lesson-03`; E04 notes. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low–medium.
Current status: fixed in lesson 04.

Ack, nack and the expiry sweep number a new attempt row as the count of the job's rows plus one, so `n` runs 1, 2, 3 … across an operator retry and is never reused. The job's `attempts` and the lease's `attempt` still reset on retry and count toward `max_attempts`.

No unique key on `(job_id, n)` was added: every insert follows a guarded transition on the job row, so two inserts for one job cannot race.

`lab/repro/U20.sh` (`probe04.py`) replays the report's steps through the public API and the lab clock;
`TestL04U20` shares that oracle. See [reconciliation 04](../../register/reconciliations/04.md) for the
spec decision.
