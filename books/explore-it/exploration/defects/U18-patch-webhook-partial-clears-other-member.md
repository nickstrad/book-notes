# U18 — partial webhook PATCH clears the omitted member

Original blind report: [E03-F08](E03-F08-patch-webhook-partial-clears-other-member.md).
Seen at `explore-it/lesson-02`; E03 notes H10. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity medium.
Current status: fixed in lesson 03.

PATCH reads `webhook.url` and `webhook.secret` as optional members and keeps whichever is omitted, like
RFC 7396. `"url": ""` still disables notifications; `null` keeps both. SPEC.md states the rule.

`lab/repro/U18.sh` (`probe03.py`) replays the report's steps through the public API only; `TestL03U18`
shares that oracle. See [reconciliation 03](../../register/reconciliations/03.md) for the spec decision.
