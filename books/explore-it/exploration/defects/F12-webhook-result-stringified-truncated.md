# F12 — Webhook result is a JSON string and silently truncated at 1024 bytes

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H11/H13 (01:43:03Z–01:43:07Z).

## Minimal reproduction
Webhook queue h11. Ack job 308 with result `{"ok":true,"n":2}` and job 309 with a 1520-byte result `{"blob":"x"*1500,"n":1}`; read hooksink records.

## Expected (oracle)
SPEC: "The entry includes job id, state, result and last error. Notification data must agree with the stored data or explicitly indicate truncation."

## Observed
```
308: result_type "string", value "{\"n\": 2, \"ok\": true}"
309: result_type "string", length 1024, no truncation field; stored result length 1520
```

## Variables isolated
Result size (20 B vs 1520 B): type mismatch at any size; truncation above 1024 B. Code also truncates last_error to 256 bytes (jobs.go:83-85) — predicted same, not probed; byte slicing can split a UTF-8 rune (inferred).

## Sibling check
jobs/jobs.go:79-86 is the only body builder; used by ack, nack(dead) and cancel.

## Severity / status
Medium: receivers get a different type than the API returns and cannot detect truncation. Status: open.
