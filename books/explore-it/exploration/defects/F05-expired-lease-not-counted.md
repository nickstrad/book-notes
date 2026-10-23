# F05 — Expired leases neither count nor record an attempt; jobs re-execute without bound

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H06 (01:42:44Z–01:42:53Z).

## Minimal reproduction
Queue h05 (max_attempts 1, lease_seconds 2). Job 307 has attempts=1 (one failed nack, F04) and is leased again.
1. Do nothing for 4 s (lease expires; background loop runs every 1 s).
2. Inspect job and attempts; lease again; wait 4 s; lease again.

## Expected (oracle)
SPEC: "An expired lease records an unsuccessful attempt and releases the job for retry unless the maximum attempts have been exhausted." and "No job may execute more than its maximum attempts."

## Observed
```
after expiry: 307|queued|1|t   attempt_rows=1 (only the earlier nack)
3rd lease → {"attempt":2,"id":307,…}
4th lease → {"attempt":2,…}; 307|leased|1 ; attempt rows still 1
```
A max_attempts=1 job was leased 4 times; the lease response keeps reporting attempt 2.

## Variables isolated
Lease duration (2 s) only sets timing. Starting attempts value irrelevant: expire.go:6 resets to queued without touching attempts. Worker never acked/nacked — the crash/timeout path. Worker kind `crash` (README) would loop forever across worker restarts (inferred, not probed).

## Sibling check
Worker deregistration (internal/api/jobs.go:110) also requeues leases without an attempt row — read, not probed; spec says it "releases its active leases", silent on counting. No other expiry path found.

## Severity / status
High: poison jobs (crash/timeout) are retried forever; attempts history and max_attempts are not enforced; never reaches dead or notifies. Status: open.
