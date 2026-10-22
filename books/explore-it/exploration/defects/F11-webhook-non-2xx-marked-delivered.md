# F11 — Webhook responses with non-2xx status are marked delivered and never retried

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H12 (01:43:19Z–01:43:28Z).

## Minimal reproduction
1. `POST :18221/admin/mode {"mode":"500"}`.
2. Ack a job in a webhook queue (h11, job 310). Wait 8 s.

## Expected (oracle)
SPEC: "Delivery is successful only on a 2xx response. Failed notifications are retried."

## Observed
```
hooksink records: {"status":500,"job":310,…} (exactly one)
outbox: 72|310|attempts=1|delivered=t
api log: "webhook delivered status=500"
```

## Variables isolated
Only the partner mode changed (ok → 500). Transport errors (mode timeout, H34) *are* retried every 5 s — the status code is simply not inspected (ship.go:40-43).

## Sibling check
Only one delivery path (webhook/ship.go). Retry cadence for transport errors is fixed 5 s with no cap/backoff (read, not a spec violation).

## Severity / status
High: notifications are silently lost whenever the partner returns 4xx/5xx; at-least-once broken. Status: open.
