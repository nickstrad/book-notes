# E06: checking data boundaries and dependent lifetimes

Project-specific findings from experiment 06 and its repair; VM facts belong elsewhere.

- Exact storage does not imply exact clients. Check CLI input against storage and CLI output
  against a separately API-written value. A single CLI round trip can hide symmetric corruption.
  Use a large integer, a long fractional decimal and a nested value; parse observations without
  binary floating-point conversion. `lab/repro/probe06.py` supplies the public E06 checks.
- REF must run target-image qctl and worker binaries. Current instrument code is appropriate;
  current application decoding is not. Starting a target-image worker on a unique queue makes
  its data path observable without changing the service or colliding with seeded workers.
- A decoder change to json.Number also changes numeric consumers: sleep/failure controls need
  to keep working, including omitted-field defaults. Preserve both the data and the behavior. Tests consuming
  the shared client must parse IDs as json.Number/Int64 rather than assuming float64; retain
  their behavioral assertions. E06 preserves the failed baseline run and repeats the entire
  paired sequence after this test-consumer adaptation.
- Deleting an owner has several independent effects. Check actual storage, API reachability,
  durable events and pending delivery completion. API 404 alone cannot establish storage cleanup.
  Attempts now follow their job; events and notification snapshots retain independent lifetimes.
- Retention needs both a preserved class and deleted controls. Freeze API time, observe below
  cutoff, advance past it, bound real waits for sweeps, then inspect real rows. Do not substitute
  a projected count or SQL-mutated fixture for storage evidence.
- The API reruns migration SQL on every startup. Constraint additions need the same catalog guard
  as existing migrations. The first E06 rebuild exposed a missing guard; its log and diagnostics
  are preserved, and repeated startups during reset/acceptance exercise the repair.
- A sampled race non-hit is not proof that its window is absent. Require the actual reported
  historical errors for the paired claim. Current checks also need successful/missing-owner
  controls, exact accepted-row cardinality, known error codes and no orphaned writes.

No reusable VM helper or knowledge-store write was warranted. The E06 oracle is indexed in
`project/lab/repro/README.md`; the acceptance launcher remained scratch under `/tmp`.
- Hooksink's `body` convenience field is itself parsed through float64. Precision assertions
  must decode `raw_body`, the original received bytes, using Decimal/integer parsing. E06's
  first current attempt failed explicitly on this measurement error; the corrected instrument
  is frozen identically for both final verification directions. The sink is unchanged.

## Whole-document JSON parsing after UseNumber

The E06 review of superseded fix `7a38788` exposed a second decoder obligation:
`json.Decoder.Decode` reads one value from a stream. Success alone allows trailing
garbage or another value, whereas `json.Unmarshal` validates the entire document.
Changing number representation must preserve that completeness contract as well.

For these already-buffered client/CLI inputs, amended fix `755d21c` shares
`internal/client.DecodeJSON`: validate all bytes with `json.Valid`, then decode with
`UseNumber`. This also avoids populating output from a syntactically invalid prefix.
For a streaming reader, the existing API pattern instead requires a second Decode to
return `io.EOF`. Trailing whitespace is valid in both approaches.

Propagate malformed CLI-option errors before transport. Preserve the printer's existing
full raw-response fallback when a response is not a complete JSON document; the shared
JSON client returns an error. Exact numbers alone cannot prove either property.

Exercise application code: `cmd/qctl/main_test.go` builds the real executable and checks
its exit code, request count, wire body and display against an HTTP stub. Client tests
serve actual response bytes; worker tests pass actual client-decoded payloads to perform.
Keep whitespace, big integers, long decimals and nested values alongside malformed input,
trailing garbage, concatenated values and omitted/default numeric consumers. A test that
copies decoder logic can miss a forgotten application call site.

The worker http-result decoder is a separate pre-existing sibling, explicitly deferred;
the echo repair does not establish its precision. Likewise new schedule validation does
not establish legacy-template cleanup. Scope repair claims to the paths actually exercised.

Evidence: [repair acceptance](../../register/runs/06-repair-acceptance.json), all five
required commands passed; original acceptance and raw records remain unchanged. Independent
validation is pending. No additional reusable work script was warranted: the small shared
decoder belongs to the application, and the acceptance launcher remains scratch under `/tmp`.
