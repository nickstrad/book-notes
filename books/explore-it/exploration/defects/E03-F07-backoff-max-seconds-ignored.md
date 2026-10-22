# E03-F07 — Undocumented `backoff.max_seconds` is accepted, stored and echoed (even negative) but never applied

- Seen at: `explore-it/lesson-02` plain export; stack `queuelet-e03`; real clock; fixture seed 04:16:24Z.
- Original observation: [E03 notes](../experiments/03-vary-it/notes.md) E03-H07 (stated 04:16Z; acceptance observed 04:17:59Z,
  effect 04:19:56Z and 04:20:13Z). Raw: `project/results/e03/p4-queuecfg.txt`, `p6-backoff.txt`, `p6b-backoff-controls.txt`.
  Prediction source: `internal/api/queues.go:20,46`; grep of `backoff_max|MaxSeconds|max_seconds` finds no reader.

## Minimal reproduction
1. `POST /v1/queues -d '{"name":"e03-maxs2","backoff":{"kind":"exponential","base_seconds":10,"max_seconds":1}}'` → 201,
   response `"backoff_max":1`.
2. Register an echo worker, enqueue, lease, `nack {"lease_token":…,"error":"x"}` (no delay override).
3. Read `run_at`.

## Expected (oracle)
SPEC:17 unknown fields should be rejected; SPEC:26-28 documents the backoff object as `kind` and `base_seconds` only — so
`max_seconds` is an unknown field → 400. If it is instead a supported field, a cap of 1 s must bound the retry wait, and a
negative cap must be rejected (SPEC:26-27 nonnegative seconds).

## Observed
```text
create e03-maxs  {base 4,  max_seconds -7} → 201 "backoff_max":-7
create e03-maxs2 {base 10, max_seconds 1}  → 201 "backoff_max":1
e03-maxs  job 842 nack at 04:19:56.169 → run_at 04:20:00.190 (+4 s  = base·2^0; cap ignored)
e03-maxs2 job 844 nack at 04:20:13.643 → run_at 04:20:23.621 (+10 s = base·2^0; cap of 1 s ignored)
max_seconds 2147483648 → 409 queue_conflict (int column; E03-F04)
```

## Variables isolated
- Positive and negative caps both have no effect; validation runs on none of them.
- Not PATCHable (`PATCH {"backoff":…}` → 400 unknown field), so the stored value cannot be corrected either.

## Sibling check
- Other accepted-but-unread inputs in the queue/job inputs: none found — every other `QueueInput`/`JobInput` field is used in
  the INSERT and read by lease/nack/expire/retain (`jobs.go`, `nack.go`, `expire.go`).
- Interaction: an honoured cap would also have bounded the overflow in [E03-F06](E03-F06-backoff-overflow-immediate-retry.md).

## Severity / status
Low–medium: operators who set a cap believe retries are bounded (e.g. "never wait more than 60 s") while exponential waits grow
without limit; the API confirms the setting back to them. Status: open (spec decision: reject the field or implement it).
