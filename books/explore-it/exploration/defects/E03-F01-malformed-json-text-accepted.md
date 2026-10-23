# E03-F01 — Malformed JSON request bodies are accepted: trailing data ignored, a second object silently dropped, invalid UTF-8 rewritten

- Seen at: `explore-it/lesson-02` plain export; stack `queuelet-e03`, image `queuelet:e03`; fixture `lab/seed/seed.py` 04:16:24Z.
- Original observation: [E03 notes](../experiments/03-vary-it/notes.md) E03-H01 (stated 04:16Z, observed 04:16:44Z),
  E03-H16 (04:17:32Z), batch sibling (04:23:09Z), PATCH sibling E03-H19 (04:24:21Z). Raw: `project/results/e03/p1-shape.txt`,
  `p3-numbers.txt`, `p9b-payload-batch.txt`, `p10-siblings.txt`. Prediction source: `internal/api/api.go:49-58`.

## Minimal reproduction
Any namespace key, any existing queue (here acme `imports`, which has no Compose worker):
1. `curl -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/imports/jobs -d '{"kind":"echo","payload":{"t":"h01a"}} trailing-garbage'`
2. `... -d '{"kind":"echo","payload":{"t":"h01b"}}{"kind":"second","payload":{"t":"h01b2"}}'`
3. `printf '{"kind":"ech\xffo","payload":{"t":"h16-utf8"}}' > f.json; curl ... --data-binary @f.json`
4. Read back the rows (`GET /v1/jobs/{id}` or read-only SQL).

## Expected (oracle)
SPEC:12-17 — commands are JSON; "Unknown fields and invalid enum values should be rejected as invalid commands" (the
service should reject what it does not understand). RFC 8259 §2: a JSON text is exactly one value surrounded by optional
whitespace, so `{…} trailing-garbage` and `{…}{…}` are not JSON texts; RFC 8259 §8.1: JSON exchanged between systems MUST
be UTF-8. Expected 400 `invalid_json` and nothing stored.

## Observed
```text
1 → {"id":301,"state":"queued"} [201]   stored {"t": "h01a"}
2 → {"id":302,"state":"queued"} [201]   stored kind echo {"t": "h01b"}; no job of kind "second" exists
3 → {"id":831,"state":"queued"} [201]   stored kind "ech�o" (hex 656368efbfbd6f — U+FFFD substituted)
control: '{"kind":"echo",}' → 400 invalid_json (a malformed *first* value is rejected)
```
Cause (code route, consistent with the observations): `decode` calls `json.Decoder.Decode` once and never checks for
EOF afterwards (`api.go:50-53`); Go's decoder converts invalid UTF-8 inside strings to U+FFFD.

## Variables isolated
- Trailing content: arbitrary text and a complete second JSON object both ignored; the second object's command is lost
  with a 201 for the request as a whole.
- Endpoint: job enqueue (H01), queue create (`{"name":"e03-trail"}{"name":"other"}` → 201, no `other` queue), batch
  (`{"jobs":[…]} {"jobs":[]}` → 201 one id), queue PATCH (`{"MAX_ATTEMPTS":5} garbage` → 200 applied).
- Invalid UTF-8 placement: inside a `string` field (kind) → silently rewritten; inside the raw `payload` → rejected 400 by
  PostgreSQL (`pq: invalid byte sequence for encoding "UTF8": 0xff`), so the payload path is not affected.
- Did not matter: Content-Type (text/plain also accepted, see notes Q-E03-1), namespace.

## Sibling check
Every handler that uses `decode` (`internal/api/api.go:49`): queue create/PATCH (`queues.go:35`, `:77`; observed), enqueue and
batch (`enqueue.go:78`, `:92`; observed), ack/nack/heartbeat (`jobs.go:60`; not probed with trailing data), worker create,
lease, schedules create/PATCH, `/admin/clock` (`observe.go:25`) — predicted identical by code, not probed.
Case-insensitive member matching is a separate decoder leniency: [E03-F02](E03-F02-case-variant-member-names-accepted.md).

## Severity / status
Low–medium: a client or proxy that concatenates or mis-frames requests gets a 2xx while part of its command is discarded;
corrupted kind strings are stored as if valid. Status: open.
