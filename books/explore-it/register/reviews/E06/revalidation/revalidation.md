# E06 review-repair revalidation (fresh Claude Opus 5.5, cross-company fallback)

**Verdict: PASS** — 0 findings. Both NITS from the prior Muse review of superseded `7a38788` are resolved; no unplanned SUT change.

## Seat and scope

- Seat deviation: the protocol's third-family seat (Muse Wonder Full) returned the original NITS. The fresh Muse revalidation hit repeated provider stream failures before running any test, so it is **not counted as a verdict**. This is the fresh cross-company fallback (Anthropic; builder was Codex Sol / OpenAI).
- Scope: only the repair of the two NITS in `prior-review.md`. Not a full E06 or checkpoint B review; earlier fixes, citations and migrations not re-reviewed.
- Clone `/tmp/explore-it-revalidate06-claude/source`, branch `plan/explore-it`. Tag `explore-it/lesson-06` = `755d21c` (parent `b3b0559`), knowledge-only HEAD `7148dee`, superseded `7a38788` present in the clone.
- `git diff 7a38788..explore-it/lesson-06 -- books/explore-it/project/` is byte-identical to the saved `repair.diff` (`cmp` equal). Project changes are limited to five files: `cmd/qctl/main.go`, `internal/client/client.go`, and three new test files (`cmd/qctl/main_test.go`, `internal/client/client_test.go`, `internal/worker/worker_test.go`). `lab/` and `test/checks/` (repros, probe06, 74 checks) are unchanged since `7a38788`.
- `7148dee` touches only `docs/knowledge/` (2 files).

## Finding 1 (trailing data / prefix acceptance): resolved

| Check | Evidence |
|---|---|
| UseNumber still exact | `client.DecodeJSON` (`internal/client/client.go:60-68`) runs `json.Valid` first, then `Decoder.UseNumber().Decode`. Tests show `9007199254740993` and a 29-digit fraction survive on the wire (all four option sites), in the display, and through `Client.JSON`. |
| Whole-document completeness at every new decode site | All three sites from the review now go through `DecodeJSON`: qctl options `--json/--payload/--result` (`main.go:52-61`), the response printer (`main.go:281`) and `Client.JSON` (`client.go:57`). No streaming `Decode` without validation remains. The only other decoders are `main.go:104` (pre-existing `json.Unmarshal`, already whole-document) and `worker.go:111` (deferred, see Finding 2). |
| Malformed, trailing or second-value options fail before HTTP | Actual-binary test `TestQctlJSON`: for each of enqueue `--payload`, job ack `--result`, batch `--json` and schedule create `--json`, the inputs `""`, `" \t"`, `{"big":}`, `…xyz`, `…{}` and `…\ntrue` each give exit 1, stderr `--<opt>: invalid JSON`, empty stdout, and **zero** requests to the stub server. |
| Printer keeps the raw fallback | Responses `{"id":1}xyz`, `{"id":1}\n{"id":2}` and `{"id":}` are printed byte-for-byte raw with exit 0, which matches the lesson-05 behaviour (`Unmarshal` failure led to a raw print). |
| Whitespace and numeric controls accepted | Leading and trailing whitespace on options and responses is accepted. `--result` values `null`, `true`, `"text"`, a big integer, a long fraction and `[1,null]` are sent unchanged. With options omitted the body is still `{"kind":"echo","priority":5}`. |
| Client rejects without decoding a prefix | `TestJSONResponseCompleteness` covers malformed, garbage, second object, second scalar and empty input: each returns an error and `out` stays nil. |
| Worker defaults retained | `TestDecodedWorkerDefaults` sends jobs through the real `Client.JSON` and calls `perform`. It covers sleep omitted, sleep wrong type, `seconds: 0`, fail3 omitted, exhausted, wrong type and `n: 0`, fail omitted, numeric and exhausted, and an exact echo round-trip. All pass. |
| U30 paired oracle semantics preserved | probe06 and the checks are unchanged. The new strictness is tested as a unit and current-only control, so the REF judgement for U30 is still about numeric precision. |

## Finding 2 (worker http-kind precision sibling): resolved by explicit deferral

- `worker.go:111` (`json.Unmarshal` into `any`) is unchanged. No http behaviour was changed.
- The deferral is stated explicitly in all three required places:
  - `STATE.md` open questions (the "E06 review finding 2 … Deferred … Checkpoint B must retain this follow-up" entry)
  - the U29 canonical record ("Scope / deferred sibling (review finding 2)")
  - the lesson note `lessons/06-relate-it.md` ("remains explicitly deferred … unprobed")
- The Outbox row of the lesson's journey table was narrowed to "worker *echo* values", so it no longer overstates the fix.

## Observation carried from the prior review: legacy invalid schedule templates

Explicitly open, owned by the coordinator and Nick for checkpoint B. It is recorded in `STATE.md`, the U28 canonical record and the lesson note. It is not claimed as fixed and not counted as a finding.

## Integrity and evidence truthfulness

- Raw explorer artifacts: recomputed from `register/runs/06-artifact-manifest.json`, 3447 entries, 0 missing, 0 mismatches. `exploration/` outside `defects/` has no diff `b3b0559..7148dee`. The only canonical record edits are appended notes in U28, U29 and U30. The score and the seven-report count are unchanged.
- Frozen instruments: `sha256sum -c ../register/runs/06-instruments.sha256` gives 95 OK and 0 failures. Manifest sha256 `379fef80…` matches the acceptance JSON.
- Code and test sha256 values recorded in `06-repair-acceptance.json` match the five files in the clone.
- `review_report_sha256` `7dbd40f9…` in the acceptance JSON equals the sha256 of `prior-review.md`.
- Repair acceptance (`06-repair-acceptance.json`): every run is exit 0.

| Run | Time | Result |
|---|---|---|
| focused tests ×2 | — | pass |
| `make up` | 50.593 s | exit 0 |
| `sha256sum -c` (instruments) | — | exit 0 |
| `make verify LESSON=06` | 121.898 s | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| `make verify LESSON=06 REF=explore-it/lesson-05` | 49.876 s | `DEFECTS REPRODUCED: 8 present` |
| `make checks` | 399.806 s | pass (`ok queuelet/test/checks 398.731s`) |
| `make test` | 5.668 s | pass (includes qctl, client and worker) |
| `make down` | — | exit 0 |
| integrity script | — | exit 0 |

- The verdict lines in `06-repair-current.log` and `06-repair-historical.log` match `STATE.md`, the lesson note and `BUILD-STATE`.
- The REF log contains real U29 rounding (`…992`, `0.12345678901234568`) and real U32 FK hits.
- `test/checks` contains 74 `func Test`, consistent with the "74/74" claim. The checks log is non-verbose, so the per-test count comes from source, not from the log.

## Checks actually run by me (logs in this scratch root)

| Command | Result |
|---|---|
| `go test ./cmd/qctl ./internal/client ./internal/worker -count=1 -v` (temp GOCACHE/GOPATH, `GOFLAGS=-buildvcs=false`) | exit 0; 34 `--- PASS`, 0 FAIL (`go-test.log`) |
| `go vet ./...` | exit 0 (`go-vet.log`) |
| `gofmt -l cmd internal` | clean |
| `cmp` git diff vs saved `repair.diff` | identical |
| Raw artifact hash recompute (3447) | 0 missing, 0 mismatch |
| `sha256sum -c 06-instruments.sha256` | 95/95 OK |

No separate actual-binary stub probe was needed: `TestQctlJSON` builds and runs the real qctl binary against an `httptest` server and counts requests, so the evidence does not rest on a copy of the decoder.

## Observations (not findings)

1. **Empty string is now rejected.** `--payload ""` / `--json ""` / `--result ""` used to be treated as "option omitted" (the `raw != ""` gate). They now exit 1 with `invalid JSON`. This is a slightly stricter CLI behaviour that the tests assert on purpose, and it is consistent with the requested "surface a parse error". It is not a regression on any documented path. The flag form `--json` with no value still parses as `true`.
2. **Dangling review link.** The new links to `register/reviews/E06/review.md` (STATE, lesson, U29 and U30 records) point to a file not present in this clone at `7148dee`. The coordinator is expected to copy the review artifacts into `books/explore-it`; until then the links dangle.

## Material limits

- I did not run Docker myself (none needed, per the brief). The current/REF/74-check verdicts rest on the builder's recorded logs, which I inspected and found internally consistent and checksum-backed.
- This is a bounded repair validation only; the full checkpoint B review (E02–E06) remains a separate future unit.
