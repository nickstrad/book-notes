# U23 — queued age counts time spent leased after a release

Original blind report: [E04-F05](E04-F05-queued-age-counts-leased-time.md).
Seen at `explore-it/lesson-03`; E04 notes. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity low.
Current status: fixed in lesson 04.

Release by lease expiry (`Expire`) and by worker deletion (`deleteWorker`) set `run_at` to the release time, as operator retry already did. Stats keep using the later of creation and `run_at`, which now means "since the job last became ready".

Spec decision: readiness is the moment a job last became `queued` (Q-E04-2).

`lab/repro/U23.sh` (`probe04.py`) replays the report's steps through the public API and the lab clock;
`TestL04U23` shares that oracle. See [reconciliation 04](../../register/reconciliations/04.md) for the
spec decision.
