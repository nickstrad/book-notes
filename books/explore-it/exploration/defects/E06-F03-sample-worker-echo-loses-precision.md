# E06-F03 — The sample worker rounds payload numbers: `echo` returns 9007199254740992 for 9007199254740993 and 1.1 for 1.10

Seen at: export of `explore-it/lesson-05` (stack `queuelet-e06`), experiment 06 (relate-it).
Original observation: [E06 notes](../experiments/06-relate-it/notes.md) P4, 15:35:24Z (H-E06-05, predicted from code).
Prior record: D08 (lesson 01) fixed API-side precision and left "client-side qctl/worker numeric loss" as an unprobed
prediction; this is its first observation.

## Minimal reproduction

Actor: acme key; worker-b (sample worker, documented kinds) serves `reports`, which has a webhook to hooksink.
Clock: real.

```sh
source exploration/experiments/06-relate-it/tools/env.sh
k acme-key POST /v1/queues/reports/jobs '{"kind":"echo","priority":9,"payload":{"big":9007199254740993,"dec":1.10}}'
sleep 6
sql "select payload::text, result::text from jobs where id=<id>"
curl -s $QUEUELET_HOOK_URL/admin/records | jq -r '.records[].raw_body'
```
Replay: `tools/p4-journey.sh` (job A).

## Expected (named oracle)

README "Echo returns its payload"; SPEC:63 "Payload values are delivered to workers without loss of numeric
precision". The worker shipped with queuelet (`cmd/worker`, `qctl worker run`) is the documented consumer; echo's
result should be value-equal to the payload. A control job acked by hand with the same numbers (job 311) shows the
API, storage and webhook keep the digits, so the loss is attributable to the worker alone.

## Observed

| Hop | Value of `big` / `dec` |
|---|---|
| request | 9007199254740993 / 1.10 |
| psql `jobs.payload`, `GET /v1/jobs`, lease bytes | 9007199254740993 / 1.10 |
| psql `jobs.result` (after worker-b's echo) | **9007199254740992 / 1.1** |
| outbox body, hooksink raw_body (signature valid) | **9007199254740992 / 1.1** |
| control job 311, manual ack | 9007199254740993 / 1.10 everywhere |

## Variables isolated

- Matters: the job passes through `internal/client.Client.JSON` (`client.go:57`, `json.Unmarshal` without `UseNumber`)
  into `worker.Job.Payload map[string]any` (`worker.go:21`), i.e. float64.
- Does not matter: queue, webhook, priority. Integers ≤ 2^53 and decimals without trailing zeros survive.

## Sibling search

- `worker.go:82, :88` (`sleep`, `fail`) read float64 from the same map — only affects values beyond 2^53, harmless.
- `worker.go:98-110` `http` kind: response decoded with `json.Unmarshal` into `any` → same loss for remote results.
- `cmd/qctl` uses the same pattern for input and output → separate report [E06-F04](E06-F04-qctl-rounds-payload-numbers.md).
- Server side: lease (`jobs.go:81` RawMessage), GET (`store.go` RawMessage), ack (`api.go:53-62` UseNumber), Notify and
  Ship (RawMessage) — exact (P4 control).

## Severity and status

Severity: **low–medium**. Large ids (e.g. 64-bit database keys or snowflake ids) in payloads come back altered from the
reference worker and are notified to partners altered, with a valid signature. Status: **open**.
