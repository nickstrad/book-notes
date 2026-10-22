# F18 — Stats "queued" count includes scheduled and retry_wait jobs

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H01 stats (01:41:38Z) vs H17 DB counts (01:41:46Z).

## Minimal reproduction
Seed fixture; `qctl stats imports` (or `GET /v1/queues/imports/stats`); compare `SELECT state,count(*) FROM jobs WHERE queue_id=3 GROUP BY state`.

## Expected (oracle)
SPEC: "expose counts by exact state".

## Observed
```
stats: queued 80, scheduled 25, retry_wait 6 …
DB:    queued 49, scheduled 25, retry_wait 6
```
80 = 49 + 25 + 6: those jobs are counted twice across keys; the sum of the stats keys exceeds the number of jobs.

## Variables isolated
Only states scheduled/retry_wait are folded; other keys match DB. Same code serves namespace stats.

## Sibling check
stats/stats.go:17 (fold). Also oldest_queued_age uses the same widened state set and run_at (F26).

## Severity / status
Low–medium: dashboards/alerts on queue depth are inflated by future work. Status: open.
