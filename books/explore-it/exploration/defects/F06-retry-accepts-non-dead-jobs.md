# F06 — Operator retry is accepted for succeeded (non-dead) jobs

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H08 (01:42:03Z); qctl sweep H40 (retry of a dead job, 01:50:49Z) as control.

## Minimal reproduction
Job 302 in state succeeded (from F01 repro). `POST /v1/jobs/302/retry` (acme-key).

## Expected (oracle)
SPEC Operator actions: "Succeeded and cancelled jobs cannot be retried." + "A command incompatible with the current state returns 409."

## Observed
```
{"ok":true} [HTTP 200]
302|queued|0|{"by": "impostor"}|(finished_at NULL)
```
The job was then leased and run again (H07, attempt 2). Stale result kept on a queued job.

## Variables isolated
State succeeded probed. Cancelled not separately probed via retry (predicted same: operator.go:13 has no state filter). Retry of a dead job (legitimate) → 200 (control).

## Sibling check
Cancel filters terminal states correctly (operator.go:6). DELETE job filters terminal states (H33 → 409 for queued). Retry of leased/queued jobs: same unfiltered UPDATE — predicted, not probed.

## Severity / status
Medium–high: completed work can be re-run by an operator mistake; violates explicit spec line. Status: open.
