# E06-F04 — qctl rounds numbers both ways: `enqueue --payload` stores a different number, and `job get` displays one

Seen at: export of `explore-it/lesson-05` (stack `queuelet-e06`), experiment 06 (relate-it).
Original observation: [E06 notes](../experiments/06-relate-it/notes.md) P5, 15:36:13Z (H-E06-14, H-E06-15, predicted
from code; sibling of [E06-F03](E06-F03-sample-worker-echo-loses-precision.md)).

## Minimal reproduction

Actor: acme key; qctl binary from the e06 image (run inside the api container against `http://localhost:8200`).

```sh
source exploration/experiments/06-relate-it/tools/env.sh
q() { docker exec -e QUEUELET_URL=http://localhost:8200 -e QUEUELET_KEY=acme-key queuelet-e06-api-1 qctl "$@"; }
q enqueue reports --kind manual --payload '{"big":9007199254740993,"dec":1.10}'   # {"id": 312, "state": "queued"}
sql "select payload::text from jobs where id=312"     # {"big": 9007199254740992, "dec": 1.1}
k acme-key POST /v1/queues/reports/jobs '{"kind":"manual","payload":{"big":9007199254740993,"dec":1.10}}'   # control, id 313
sql "select payload::text from jobs where id=313"     # {"big": 9007199254740993, "dec": 1.10}
q job get 311      # stored 9007199254740993 / 1.10, printed "big": 9007199254740992, "dec": 1.1
```
Replay: `tools/p5-qctl.sh`.

## Expected (named oracle)

README: "The CLI and the HTTP API expose the same operations." SPEC:63 numeric precision. The same body sent by curl
is stored exactly (job 313), so the CLI must not change the data it sends, and its display must show the stored value.

## Observed

- Input: `qctl enqueue … --payload` stored `9007199254740992` / `1.1` (job 312).
- Output: `qctl job get 311` printed `9007199254740992` / `1.1` for a row whose payload and result hold
  `9007199254740993` / `1.10`.

## Variables isolated

- Matters: `cmd/qctl/main.go:52-57` `jsonOpt` decodes every `--payload`/`--result`/`--json` option into `any`
  (float64) before re-encoding; `main.go:277-280` decodes every response into `any` before `MarshalIndent`.
- Does not matter: endpoint, queue, kind. Values ≤ 2^53 without trailing zeros are unaffected.
- `jsonOpt` also ignores parse errors (an invalid `--payload` becomes `null`) — the API then rejects it (400), so no
  corrupt write results; noted, not counted.

## Sibling search

All qctl verbs share `jsonOpt` and the printer: `ack --result`, `batch --json`, `schedule create --json`,
`queue create/patch --json` (predicted same loss; not probed individually). Worker client: E06-F03.

## Severity and status

Severity: **medium** for the write path (an operator using the documented CLI silently enqueues different data than
typed), **low** for display. One report because one decoding habit causes both. Status: **open**.
