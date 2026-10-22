# E04-F01 — Wrong-method and unknown-path requests get Go's plain-text 404/405, not the error envelope

- Seen at: `explore-it/lesson-03` plain export; stack `queuelet-e04`; acme key; no fixture seed.
- Original observation: [E04 notes](../experiments/04-judge-it/notes.md) P1 (12:45:13Z) and P11 (12:49:59Z).
  Raw: `project/results/e04/p1-errors.txt`, `p11-siblings.txt`. Prediction source (code): `internal/api/api.go:119-172`
  registers method-qualified patterns on a `http.ServeMux`; only handlers call `fail()` (`api.go:50-52`), so a request the
  mux itself rejects never reaches it.

## Minimal reproduction
```sh
curl -i -X PUT -H 'Authorization: Bearer acme-key' http://localhost:18250/v1/queues
curl -i       -H 'Authorization: Bearer acme-key' http://localhost:18250/v1/nope
```

## Expected (oracle)
SPEC "Authentication and responses": "Errors have the form `{"error":{"code":"…","message":"…"}}` and a 4xx or 5xx HTTP
status." Never/Always rule R36: *every* error response from `/v1/` is a 4xx/5xx **with** the JSON envelope, so clients can
parse one error shape.

## Observed
```text
PUT    /v1/queues        → 405, Content-Type: text/plain; charset=utf-8, body "Method Not Allowed"   (Allow: GET, HEAD, POST)
GET    /v1/nope          → 404, text/plain, "404 page not found"
DELETE /v1/schedules     → 405 text/plain
GET    /v1/jobs/1/ack    → 405 text/plain (Allow: POST)
OPTIONS /v1/queues       → 405 text/plain
PATCH  /v1/jobs/66       → 405 text/plain (Allow: DELETE, GET, HEAD)
```
The status class is right (4xx); the body form is not.

## Variables isolated
- Authentication runs first: the same `PUT /v1/queues` without a key gets a JSON 401 envelope. Only authenticated requests
  reach the mux fallback.
- Method vs path: both unknown paths (404) and known paths with an unregistered method (405) are affected.
- Did not matter: queue/job existence (`/v1/jobs/1/ack` with no job 1 still 405 before any handler).

## Sibling check
- `POST /v1/jobs/{id}/{verb}` and `POST /v1/queues/{q}/{verb}` wildcards: an unknown verb reaches the handler and gets a JSON
  404 `not_found` "unknown operation" (`internal/api/jobs.go:83-85`, `queues.go` queueVerb default). So `POST
  /v1/jobs/66/events` (a GET-only route) is a JSON **404**, while `PATCH /v1/jobs/66` is a text **405** — same mistake, two
  classes. Recorded here, not counted separately.
- `/health`, `/metrics` are public and outside `/v1/`; not probed for wrong methods.

## Severity / status
Low: no data is wrong, but a client that parses the documented envelope fails on these responses (and a proxy/log pipeline
keyed on `error.code` sees none). Status: open.
