# F26 — oldest_queued_age is measured from run_at, so a backdated run_at reports years of queueing

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) #surprise 01:51:07Z (namespace stats), isolation 01:51:30Z.

## Minimal reproduction
1. `POST /v1/queues/h34/jobs {"kind":"x","run_at":"2020-01-01T00:00:00Z"}` (accepted, state queued — H37).
2. `GET /v1/queues/h34/stats` and `GET /v1/namespaces/acme/stats`.

## Expected (oracle)
SPEC: "oldest queued age in seconds" (reference point unstated). Consistency bound: a job cannot have been queued longer than it has existed, so age ≤ now − created_at of the oldest queued job (here ≈ 60 s).

## Observed
```
h34 "oldest_queued_age":214797090.198135     (≈ 6.8 years)
imports 622.4, emails 0
898|queued|run_at 2020-01-01 00:00:00+00|created_at 2026-10-22 01:50:30
```

## Variables isolated
Only job 898 differs between h34 and other queues. stats/stats.go:19 uses min(run_at) over queued/scheduled/retry_wait.

## Sibling check
Schedule catch-up jobs (F24) carry past run_at values too → inflate age by up to the pause length (inferred).

## Severity / status
Low: one client-supplied timestamp makes a queue-age alert fire permanently. Spec reference point is silent; the bound above is an approximation oracle. Status: open.
