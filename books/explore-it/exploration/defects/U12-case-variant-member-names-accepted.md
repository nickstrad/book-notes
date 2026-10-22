# U12 — case-variant member names matched as documented fields

Original blind report: [E03-F02](E03-F02-case-variant-member-names-accepted.md).
Seen at `explore-it/lesson-02`; E03 notes H02, H20. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low–medium.
Current status: fixed in lesson 03.

After the strict decode, the decoder walks the request object against the target struct and rejects
any member whose name is not an exact json tag (`exactNames`). Free-form `payload` and `result` are
not walked. Exact duplicate names stay implementation-defined, as the report left them.

`lab/repro/U12.sh` (`probe03.py`) replays the report's steps through the public API only; `TestL03U12`
shares that oracle. See [reconciliation 03](../../register/reconciliations/03.md) for the spec decision.
