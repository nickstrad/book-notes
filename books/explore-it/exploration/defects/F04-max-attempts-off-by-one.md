# F04 — A job with max_attempts N can be executed N+1 times through nack

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H05 (01:42:34Z–01:42:37Z).

## Minimal reproduction
1. `POST /v1/queues {"name":"h05","max_attempts":1,"lease_seconds":2,"backoff":{"kind":"linear","base_seconds":0}}`.
2. Enqueue echo job 307; lease (attempt 1) with worker 4.
3. `POST /v1/jobs/307/nack {"lease_token":T,"error":"e1"}` (retry default true).
4. Lease again.

## Expected (oracle)
SPEC: "A failed job reaches dead when its maximum attempts are exhausted" and "No job may execute more than its maximum attempts." With max 1, the first failed attempt exhausts → `dead`.

## Observed
```
nack 200; 307|queued|1|1          (state|attempts|max_attempts after promotion)
lease again → {"attempt":2,"id":307,…}
attempts: 307|1|failed|e1
```

## Variables isolated
max_attempts=1 is the minimal case; the comparison in nack.go:15 is `attempts > max` where attempts is already the incremented count, so dead requires attempts = max+1. Backoff kind/base irrelevant (base 0 only made it immediate). Corroborating natural run (H41): fail2 on max 3 behaves correctly because it succeeds at attempt 3, so it does not exercise the boundary.

## Sibling check
Same boundary elsewhere: expire.go:6 has no attempt check at all (F05). Schedule-inserted jobs copy max_attempts (fire.go:44), no separate check. Retry resets attempts (operator.go:13) — by spec.

## Severity / status
High: one extra execution of non-idempotent work per job; violates an explicit Never rule. Status: open.
