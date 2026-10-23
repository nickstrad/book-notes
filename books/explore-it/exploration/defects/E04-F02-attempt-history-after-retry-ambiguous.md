# E04-F02 — After an operator retry, attempt numbers restart and the attempt history is listed out of execution order

- Seen at: `explore-it/lesson-03` plain export; stack `queuelet-e04`; lab clock frozen at 2030-01-01T00:00:00Z; raw API.
- Original observation: [E04 notes](../experiments/04-judge-it/notes.md) P2 (12:45:26Z), hypothesis R38 stated 12:44Z.
  Raw: `project/results/e04/p2-retry-history.txt`. Prediction source (code): `migrations/001.sql:7` (`attempts` has no key),
  `internal/jobs/operator.go:12-14` (Retry sets `attempts=0`), `internal/jobs/ack.go:14` / `nack.go:30` insert the new
  attempt number, `internal/api/jobs.go:25` lists `ORDER BY n` only.

## Minimal reproduction
1. Queue `{"name":"e04a","max_attempts":2,"backoff":{"kind":"exponential","base_seconds":2}}`; worker with kind echo.
2. Enqueue an echo job; lease, nack; (advance clock 2 s) lease, nack → `dead`.
3. `POST /v1/jobs/{id}/retry`; lease (attempt 1), ack.
4. `GET /v1/jobs/{id}/attempts`.

## Expected (oracle)
SPEC Jobs: "`GET /v1/jobs/{id}/attempts` returns completed attempts in numeric order." SPEC Operator: retry "resets its
attempt count". Consistency rule R38 (approximation): the attempt history has one row per execution and its documented order
is the order executions happened; two different executions are distinguishable. Spec is silent on whether history survives a
retry — the decision for the owner is either to number across incarnations or to expose an incarnation/order key.

## Observed
```text
n=1 failed     "f1"  started 00:00:00 finished 00:00:00     (incarnation 1)
n=1 succeeded        started 00:00:02 finished 00:00:07     (incarnation 2 — the last execution)
n=2 failed     "f2"  started 00:00:02 finished 00:00:02     (incarnation 1)
```
Two executions share `n=1`; the final success sits between the earlier incarnation's failures. Order among equal `n` is not
defined by the query, so it can also vary.

## Variables isolated
- Requires an operator retry of a dead job (the only path that lowers `attempts`). Without retry, numbering was 1..N and
  monotone in every earlier probe (e.g. P10, 21 attempts).
- Did not matter: outcome kind (failed/succeeded both renumber), backoff settings.
- Not tested: retry after death by lease expiry (same code path by reading; expected identical).

## Sibling check
- `job_events.attempt` (observation trigger, `migrations/002.sql`) also restarts, but events are ordered by `id`, so the
  event history stays chronological — no defect there.
- Webhook bodies carry no attempt number — none found.

## Severity / status
Low–medium: operators reading attempts after a retry see a misleading story (a "first attempt succeeded" alongside a "first
attempt failed"); tooling that keys attempts by `(job_id, n)` collapses two executions. Status: open (needs a spec line on
retry history).
