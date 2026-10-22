# E03-F02 — Case-variant member names (`KIND`, `Priority`, `LEASE_TOKEN`) are accepted as the documented fields; the last variant wins

- Seen at: `explore-it/lesson-02` plain export; stack `queuelet-e03`; fixture seed 04:16:24Z.
- Original observation: [E03 notes](../experiments/03-vary-it/notes.md) E03-H02 (stated 04:16Z, observed 04:16:44Z); siblings
  batch (04:23:09Z), E03-H19/H20 (04:24:21Z, rerun 04:24:37Z). Raw: `project/results/e03/p1-shape.txt`, `p9b-payload-batch.txt`,
  `p10-siblings.txt`, `p10b-siblings.txt`. Prediction source: `internal/api/api.go:49-52` (Go `encoding/json` v1 field matching).

## Minimal reproduction
1. `POST /v1/queues/imports/jobs -d '{"KIND":"echo","Priority":3,"PAYLOAD":{"t":"h02a"}}'` (acme-key).
2. `POST /v1/queues/imports/jobs -d '{"kind":"echo","priority":1,"PRIORITY":9,"payload":{"t":"h02b"}}'`.
3. `POST /v1/queues/imports/jobs -d '{"kind":"echo","PRIORITY":9,"priority":1,"payload":{"t":"h02c"}}'`.
4. Read the stored kind/priority.

## Expected (oracle)
SPEC:17 "Unknown fields … should be rejected as invalid commands." JSON member names are strings and are compared exactly
(RFC 8259 §8.3); `KIND` is not the documented `kind`. Expected 400 `invalid_json` "unknown field".

## Observed
```text
1 → 201 id 303: kind echo, priority 3, payload {"t": "h02a"}
2 → 201 id 304: priority 9   (later case-variant overrides the exact name)
3 → 201 id 305: priority 1
queue create {"NAME":"e03-case","Lease_Seconds":7} → 201, lease_seconds 7
batch {"JOBS":[{"Kind":"echo",…}]} → 201 {"ids":[1363]}
PATCH {"MAX_ATTEMPTS":5} → 200 max_attempts 5
ack {"LEASE_TOKEN":"<token>","RESULT":{"ok":1}} → 200; job 1369 succeeded, result {"ok": 1}
control: {"kind":"echo","x":1} → 400 `json: unknown field "x"` (exact unknown names are rejected)
```

## Variables isolated
- Key order: whichever spelling appears last wins (2 vs 3), so an exact-name value can be overridden by a variant later in the body.
- Applies to top-level and nested members (`PAYLOAD`, `JOBS[].Kind`) and to worker-protocol fields (`LEASE_TOKEN`).
- Exact duplicate keys (`"kind":"echo","kind":"other"` → kind other) behave the same way but RFC 8259 leaves duplicate-name
  behaviour implementation-defined, so that part alone is not claimed as a defect.

## Sibling check
All `decode` users (`internal/api/api.go:49`, list in [E03-F01](E03-F01-malformed-json-text-accepted.md)); the worker/qctl clients
send exact names so they are unaffected. `DisallowUnknownFields` (`api.go:51`) is the lesson-01 fix for D26; this is a gap in
that fix rather than a regression of it.

## Severity / status
Low–medium: a gateway or validator that inspects `priority`/`lease_token` case-sensitively can be bypassed by a variant
spelling that queuelet honours (parameter smuggling); clients with typos in casing get silent success. Status: open.
