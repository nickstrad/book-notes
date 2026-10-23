# F13 — A slow webhook partner stalls promotion, expiry and scheduling

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H34 (01:48:18Z–01:48:45Z).

## Minimal reproduction
1. `POST :18221/admin/mode {"mode":"timeout"}` (hooksink sleeps 10 s).
2. Ack 6 jobs in webhook queue h11 (6 outbox rows).
3. Enqueue `{"kind":"echo","delay_seconds":1}` in a queue without webhook; poll state.
Control: same step 3 with mode ok and empty outbox.

## Expected (oracle)
SPEC "The service promotes due jobs to queued" (no explicit bound). Approximation oracle: promotion within about one loop tick after due (control measured 2.1 s for a 1 s delay); a partner's latency should not gate core queue progress.

## Observed
```
mode timeout, 6 rows: promotion latency 13.3 s (≈ 6 × 2 s client timeout + tick)
control mode ok:      promotion latency 2.1 s
```
Inferred bound: up to 20 rows × 2 s = 40 s per tick (ship.go:17 LIMIT 20), and rows re-qualify every 5 s → sustained stall while the partner is down.

## Variables isolated
Partner mode (timeout vs ok) with otherwise identical steps. Number of outbox rows scales the delay (6 rows → ~12 s extra).

## Sibling check
Loop is single-threaded (api.go:93-108); any slow step (Fire catch-up loops, F24) would also delay promotion — predicted. Lease expiry is in the same loop → leases outlive their expiry by the stall (inferred, not probed).

## Severity / status
Medium: a partner outage degrades delayed/retry promotion and lease expiry for every namespace. Oracle is an approximation (spec gives no latency bound). Status: open.
