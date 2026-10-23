# U26 — after one lease expiry, later exits from `leased` are logged as `lease_expired`

Original blind report: [E05-F03](E05-F03-worker-release-logged-as-lease-expired.md).
Seen at `explore-it/lesson-04`; E05 notes, walk seed 11 and H-E05-15/18. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity medium (observability).
Current status: fixed in lesson 05.

The event trigger (`migrations/003.sql`) no longer reads the sticky `last_error`. The expiry sweep marks its own transaction (`SET LOCAL queuelet.release='lease_expired'`) and the trigger labels a release `lease_expired` only inside it. This also covers the report's unprobed prediction H-E05-20 (a nack whose error text is `lease expired`).

`lab/repro/U26.sh` (`probe05.py`) lets two leases expire once, leases again, acks one and cancels the other, and reads the last event of each;
`TestL05U26` shares that oracle.
