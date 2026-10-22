---
name: test-integration-points
description: Enumerate every remote call a service makes, list the failure modes each must survive (refused, hang, slow bytes, reset, 5xx, garbage, huge or cut-off bodies, stale state), check which the tests exercise, and produce a fault-injection test matrix with stubs for the gaps that assert the right failure class inside the time budget. Use when adding or reviewing a client to another service, before a launch, or when "tests pass but production fails".
---

# test-integration-points

Unit tests against a stub that answers in a millisecond prove the happy path. Integration points
fail in other ways: a connection accepted and never answered, bytes that trickle, a reset halfway,
an HTML error page where JSON was promised, ten megabytes where one kilobyte was expected. This
skill finds every integration point, decides how each must fail, and builds tests that make the
real client meet each failure on purpose.

## When to use

- Adding or reviewing a client for another service, database, queue or blob store.
- "Our integration tests pass but production fails."
- After adding timeouts, breakers or classification: prove they fire.
- Before a launch, and as a recurring suite before every release.

## Inputs

- The repo, and a way to run its clients against real network stand-ins: a fault-injecting proxy
  (Toxiproxy or similar) between the client and the dependency, and a fake partner with an admin
  endpoint that switches its behaviour.
- Each client's budget: the deadline its caller gives it.
- The failure classes callers act on (if none exist yet, step 4 defines them).

## Procedure

1. **Enumerate the integration points.** Search for outbound calls:
   `http.Client`, `NewRequest`, `grpc.Dial`, `sql.Open`, `redis.`, `amqp.`, SDK constructors,
   `net.Dial`, `fetch(`, `requests.`, `HttpClient`. Include background callers (heartbeats,
   shippers, backup uploads), not just request paths. One row per (caller, dependency, operation).

2. **List the failure modes each must survive.** For each row, the standard set:

   | Layer | Failure mode | Injector |
   |---|---|---|
   | network | refused / down | proxy disabled, container stopped |
   | network | accepted, never answered | proxy `timeout` toxic (timeout 0) |
   | network | slow answer (latency) | `latency` toxic |
   | network | slow bytes | `bandwidth` toxic (e.g. 1 KB/s) |
   | network | fragmented answer | `slicer` toxic |
   | network | reset mid-exchange | `reset_peer` toxic |
   | application | 5xx with a non-JSON body | fake mode |
   | application | 200 with malformed / truncated body | fake mode |
   | application | huge error body | fake mode |
   | application | headers never come | fake mode (slow headers) |
   | protocol | stale state / version mismatch | a real conflict (moved resource, new field) |

3. **Check coverage.** For each row × mode, does a test exist that (a) injects that failure against
   the *real* client, not a mock of it, and (b) asserts both the outcome and the elapsed time? Mark
   covered / partially (asserts error, not class or time) / missing.

4. **Define what "failing correctly" means.** A client fails correctly when the caller gets an
   error **class** it can act on, inside the **budget**, having used **bounded resources**:
   - Classes: `timeout`, `unreachable`, `bad_payload`, `remote_5xx`, `remote_4xx`, domain classes
     (`stale_epoch`, `breaker_open`). An unclassified error string fails the test.
   - Elapsed ≤ the client's budget (plus a small slack).
   - Bodies of error answers read up to a cap; status checked before decoding; `Content-Type`
     checked before parsing JSON.
   - What the caller does with each class is part of the contract: retry `unreachable`, never
     retry `bad_payload`, back off on `remote_5xx`.

5. **Write the matrix and the stubs.** One table-driven test per (client, mode), each:
   inject → call through the client's public API → classify → assert class and elapsed → remove the
   fault (in a `defer`/cleanup so a failed test does not poison the next). Run against the
   running stack, not `httptest`, so real sockets, real timeouts and real proxies are in play. Keep
   the suite behind a build tag or separate target so it runs on demand and before releases.

6. **Prove the harness tests what you think.** Point a test at a healthy instance while the fault
   is on another: it must pass fast. A suite that passes against the wrong target proves nothing.

## Output

The harness matrix, one row per (client, mode), and the list of gaps with stub names:

```
| client | misbehaviour | expected class | budget | test | status (covered / partial / missing) |
```

## Source

*Release It!* 2nd ed.: Ch. 4 "Integration Points" (the failure modes of socket-based and HTTP
protocols); Ch. 5 "Test Harnesses"; Ch. 15 "The Testing Gap"; Ch. 17 as the larger method.
Demonstrated in lesson 11 (`../../lessons/11-test-harness.md`), resting on lessons 02, 05 and 08.

## Worked example: litefleet lesson 11

The template output is `project/test/chaos/chaos_test.go` (build tag `chaos`, run with
`make test-chaos`, about 30 s) plus the blob-store fake's `POST /admin/mode` (`cmd/blobstore/main.go`)
with modes `500`, `malformed-json`, `huge-body`, `slow-headers`. Failure classes live in
`internal/httpx/errors.go` (`*CallError` with `Kind()`; `ReadError` caps error bodies at 1 MB).

Step 1 on litefleet finds the gateway's node and control clients, the node's blob and metering
clients, control → node lifecycle calls and the node → control heartbeat. Lesson 11 tests the
first three; its "Try changing" names the metering and control-resolve rows as the gaps. Step 3 at
`release-it/lesson-10`: every test was happy-path against `httptest` stubs (F16).

The matrix, run with today's test file against lesson 10's clients
(`make verify LESSON=11 REF=release-it/lesson-10`), then at lesson 11:

| client | misbehaviour | before (lesson-10) | after (lesson-11) |
|---|---|---|---|
| node | latency 3 s | timeout 2 900 ms | same |
| node | accept, no answer | timeout 2 900 ms | same |
| node | slicer | ok 1 160 ms | same |
| node | bandwidth 1 KB/s | timeout 2 501 ms | same |
| node | reset_peer | **unclassified** 156 ms | unreachable 153 ms |
| node | down | **unclassified** 112 ms | unreachable 38 ms |
| node | stale epoch | stale_epoch 4 ms | same |
| blobstore | 500 (HTML) | **unclassified** 4 ms | remote_5xx 5 ms |
| blobstore | malformed-json | **unclassified** 4 ms | bad_payload 4 ms |
| blobstore | huge-body (10 MB) | **unclassified, 2 046 ms (budget 1 s)** | remote_5xx 139 ms |
| blobstore | slow-headers | timeout 1 004 ms | timeout 1 002 ms |
| control | down → back up | breaker_open 0 ms → ok 6 ms | same |

Verdicts: `FLAW REPRODUCED: chaos_tests_failed 5` before, `IMPROVED: chaos_tests_run 12 == 12;
chaos_tests_failed 0 == 0` after. What the matrix shows that a list of tests does not: the rows
that already passed are the earlier lessons' work (timeouts from lesson 02 end every hang at the
2.9 s deadline; lesson 08's streaming reader reassembles the slicer's 64-byte chunks; lesson 05's
breaker opens and recovers), and the five that failed were failures the clients *noticed but could
not name*, plus an error page read in full. The tests classify through `Kind()` when the error has
one, which is how one test file can judge an older tree.
