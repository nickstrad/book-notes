# U11 — request body with trailing data or invalid UTF-8 accepted

Original blind report: [E03-F01](E03-F01-malformed-json-text-accepted.md).
Seen at `explore-it/lesson-02`; E03 notes H01, H16, H19. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low–medium.
Current status: fixed in lesson 03.

The shared decoder (`internal/api/api.go`) now reads the whole body, rejects invalid UTF-8, decodes one
object and requires end of input. Every command route uses it, so the unprobed siblings (ack, nack,
heartbeat, workers, lease, schedules, clock) are covered by construction, not counted as finds.

`lab/repro/U11.sh` (`probe03.py`) replays the report's steps through the public API only; `TestL03U11`
shares that oracle. See [reconciliation 03](../../register/reconciliations/03.md) for the spec decision.
