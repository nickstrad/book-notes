# U25 — ack and cancel leave the lease token, holder and expiry on terminal jobs

Original blind report: [E05-F02](E05-F02-terminal-rows-keep-lease-fields.md).
Seen at `explore-it/lesson-04`; E05 notes P5 and walk seed 11. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low.
Current status: fixed in lesson 05 (lease fields). The report's second observation, that a delivery cancelled in flight has no attempt row, belongs to the open question of what cancelling a running job means and is not changed here.

Ack and cancel clear `lease_token`, `leased_by` and `lease_expires_at`, as nack, expiry and worker release already did. Ack takes the holder for its attempt row from the pre-update row (`UPDATE … FROM jobs o … RETURNING o.leased_by`).

Spec decision: only a `leased` job has a lease token, holder and expiry; attempt rows keep the worker (Q-E05-2).

`lab/repro/U25.sh` (`probe05.py`) acks one job and cancels another while leased, then reads both;
`TestL05U25` shares that oracle.
