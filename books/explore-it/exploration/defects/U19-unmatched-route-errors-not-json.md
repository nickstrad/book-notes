# U19 — unmatched paths and methods answered outside the error form

Original blind report: [E04-F01](E04-F01-unmatched-route-errors-not-json.md).
Seen at `explore-it/lesson-03`; E04 notes. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low.
Current status: fixed in lesson 04.

The authenticated handler asks the mux which pattern serves the request. When none does, `unmatched` runs the mux's own 404/405 handler only to learn its status (and keep the `Allow` header), then answers with the documented envelope: `not_found` or `method_not_allowed`.

Wrong method on a job route still differs from an unknown verb (`POST /v1/jobs/{id}/events` is a JSON 404 from the wildcard, `PATCH /v1/jobs/{id}` a JSON 405); both are now in the error form, which is what the report claimed.

`lab/repro/U19.sh` (`probe04.py`) replays the report's steps through the public API and the lab clock;
`TestL04U19` shares that oracle. See [reconciliation 04](../../register/reconciliations/04.md) for the
spec decision.
