# F10 — Webhook signature is computed over different bytes than the body sent

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observations: [notes.md](../experiments/01-map-it/notes.md) H11 (01:43:03Z–01:43:07Z), H11b (01:46:01Z–01:46:05Z). Raw captures: project/results/sigcap.jsonl.

## Minimal reproduction
1. Queue with `"webhook":{"url":"http://hooksink:8201/hook","secret":"report-secret"}` (hooksink secret matches). Enqueue, lease, ack.
2. `curl :18221/admin/records` → signature_valid.
3. Byte-level: capture listener on the docker gateway (results/tools/sigcap.py), queue webhook secret `s3cret`, ack one job; compare the received `X-Queuelet-Signature` with HMAC-SHA256 over (a) received bytes, (b) 2-space indented JSON of the same object (results/tools/sigcheck.py).

## Expected (oracle)
SPEC: "Each POST carries X-Queuelet-Signature, a hex HMAC-SHA256 over the body using the configured secret." Standard webhook practice: HMAC over the exact transmitted bytes.

## Observed
```
hooksink: 50/50 seed-era records and both h11 records signature_valid=false (status 200)
raw_body: {"job_id":873,"last_error":"","result":"{\"k\": 1}","state":"succeeded"}
sig received : a793b82d…c685
hmac(raw)    : 09cea41e…8f66 False
hmac(pretty) : a793b82d…c685 True
```

## Variables isolated
Secret matches (hooksink and queue both report-secret); in the capture probe secret was known. Body content irrelevant: every delivery fails. Signature = HMAC(MarshalIndent(body)) (ship.go:23-25) while body = Marshal(body) (ship.go:22, 26).

## Sibling check
No other signer. Hooksink verifies over raw bytes (cmd/hooksink/main.go:52-56) — correct receiver.

## Severity / status
High: every partner that verifies signatures must reject every notification (or disable verification). Status: open.
