# F17 — Payload numbers lose precision (float64 round-trip)

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observation: [notes.md](../experiments/01-map-it/notes.md) H16 (01:43:46Z).

## Minimal reproduction
`POST /v1/queues/h15/jobs {"kind":"echo","payload":{"big":9007199254740993,"dec":0.1000000000000000055511151231257827,"neg":-12345678901234567890}}`; read DB and `GET /v1/jobs/{id}`.

## Expected (oracle)
SPEC: "Payload values are delivered to workers without loss of numeric precision."

## Observed
```
stored jsonb: {"big": 9007199254740992, "dec": 0.1, "neg": -12345678901234567000}
GET payload:  {"big":9007199254740992,"dec":0.1,"neg":-12345678901234567000}
```
Already lost at write time; lease delivers the stored value.

## Variables isolated
Integers above 2^53 and long decimals lose digits; small integers unaffected (fixture payloads intact).

## Sibling check
enqueue.go:14 (map[string]any), store.go:43-46 (read path also decodes to float64 for GET), schedule payloads (schedules.go:14, fire.go:42-43) and runSchedule (schedules.go:111-112) — same decoding, predicted lossy, not probed. Worker Job.Payload is map[string]any (worker.go:20) — client-side loss too. Lease path uses json.RawMessage (internal/jobs/jobs.go:65) — lossless itself, but data already lost.

## Severity / status
Medium: IDs/amounts above 2^53 are corrupted silently. Status: open.
