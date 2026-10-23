# F16 — Batch enqueue: oversized truncated, empty accepted, partial insertion on error

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H15 (01:43:44Z–01:43:46Z).

## Minimal reproduction
Fresh queue h15.
A. `POST /v1/queues/h15/jobs:batch` with 501 jobs `{"kind":"x"}` (results/h15-501.json).
B. `{"jobs":[]}`.
C. `{"jobs":[{"kind":"ok1"},{"kind":"ok2"},{"payload":{}},{"kind":"ok4"}]}`.

## Expected (oracle)
SPEC: "accepts {jobs:[…]} with 1–500 jobs. An empty or oversized batch is invalid and returns 400 without insertion. It returns {ids:[…]} after accepted jobs are stored."

## Observed
```
A: 201, 500 ids; 500 rows inserted (job 501 silently dropped)
B: {"ids":[]} [HTTP 201]
C: {"error":{"code":"invalid_job","message":"kind required"}} [HTTP 400]; ok1 and ok2 were inserted
```

## Variables isolated
Size (0, 4, 501) and validity of one element. C shows non-atomic behaviour independent of size.

## Sibling check
enqueue.go:80-93 only; insertJob per element without a transaction. Idempotency inside a batch not probed.

## Severity / status
Medium: producers lose jobs silently (A) or get duplicates on retry after a 400 (C). Status: open.
