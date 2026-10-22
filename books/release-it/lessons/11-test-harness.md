# 11 · Test harness: chaos tests against Toxiproxy and a misbehaving partner

Tag: `release-it/lesson-11` · Fixes: F16 · Verify: `make verify LESSON=11` (before: `REF=release-it/lesson-10`)
Book: Ch. 5 "Test Harnesses"; Ch. 4 "Integration Points"; Ch. 15 "The Testing Gap"; Ch. 17 (the
verify scripts are chaos experiments; this lesson automates the integration-point ones)

## The idea

Unit tests against a stub that always answers in a millisecond prove that the happy path works.
They say nothing about the failures Chapter 4 lists for every integration point: a connection that
is refused, reset or never answered; an answer that arrives a byte at a time; headers that never
come; a body that is HTML instead of JSON, or ten megabytes long, or cut off halfway. Chapter 15's
**Testing Gap** is the distance between what the tests exercise and what production does. The first
time such a failure happens should not be in production.

Chapter 5's **Test Harness** closes the gap. It is a stand-in for the remote system that can be
made to misbehave on purpose, in every way the real one might, so a client can be tested against
each failure mode. Some failures belong to the network: latency, timeouts, resets, slow bytes, a
partner that is down. Some belong to the application: the wrong status, a bad body, a huge body,
slow headers. A good harness has both, and a test for each that asserts the client failed *the
right way*, with an error the caller can act on, inside its budget.

## The flaw (F16, at `release-it/lesson-10`)

- The tests (`internal/nodeclient/client_test.go` and friends) cover the happy path against
  `httptest` stubs that answer in a millisecond.
- The node client (`internal/nodeclient/client.go:116–129`) read every error answer whole with
  `io.ReadAll` and decoded it strictly, so an HTML 500 from a proxy came back as "bad rows". A reset
  or refused connection was just an error string.
- The blob-store client (`internal/blob/client.go:61–69`) read the whole body and decoded it
  *before* it looked at the status. An HTML 500 came back as "bad response: invalid character '<'",
  and a 10 MB error page was read into memory in full.

## The patch

Two commits.

**Part 1: hardened clients** (~130 lines).

- `internal/httpx/errors.go` defines the failure classes a caller acts on. A `*CallError` carries a
  class (`ErrTimeout`, `ErrUnreachable`, `ErrBadPayload`) and the underlying error, and
  `errors.Is`/`errors.As` see both. `Kind()` (`:34`) names the class. The existing `*RemoteError`
  gains `Kind()` too (`:45`): `stale_epoch`, `remote_5xx` or `remote_4xx`. `Transport` (`:57`)
  classifies a failed exchange as out of time or unreachable. `ReadError` (`:77`) reads at most
  `MaxErrorBody` = 1 MB of an error answer and takes a typed error only from a JSON body
  (`IsJSON`, `:94`).
- Both clients check the status first, then the `Content-Type`, then decode
  (`internal/blob/client.go:57–70`, `internal/nodeclient/client.go:119–128`). A bad payload is not
  retried, since it would be bad again.

**Part 2: the harness** (~350 lines, mostly the test file).

- **The misbehaving partner.** The blob-store fake gains `POST /admin/mode`
  (`cmd/blobstore/main.go`): `500` (HTML error page), `malformed-json` (a 200 cut off), `huge-body`
  (10 MB of HTML at 5 MB/s), and `slow-headers` (3 s before any header).
- **The suite** (`test/chaos/chaos_test.go`, build tag `chaos`). Each test drives a real client
  from inside the Compose network against the running stack and asserts the failure class and the
  elapsed time. The gateway's node client, with the gateway's 2.9 s deadline, is tested under
  Toxiproxy `latency`, `timeout`, `slicer`, `bandwidth 1 KB/s`, `reset_peer` and `down` on node-a's
  proxy, plus a stale epoch. The node's blob client is tested against each fake mode. The control
  client's breaker is tested opening while control is down and closing after it returns. Each
  test logs one row of the harness matrix. The tests use only the clients' public calls, and
  classify errors by `Kind()` when the error has one, so the same file compiles and runs against
  an older tree.
- `make test-chaos` runs the suite (`lab/chaos.sh`) in about 30 s.

## Verify

```
make verify LESSON=11 REF=release-it/lesson-10      # before: today's tests, lesson 10's clients
make verify LESSON=11                               # after
```

Before (`lesson-10`):

```
  client      misbehaviour     outcome        elapsed
  node       latency 3s       timeout          2900 ms
  node       accept, no answer timeout          2900 ms
  node       slicer           ok               1160 ms
  node       bandwidth 1KB/s  timeout          2501 ms
  node       reset_peer       unclassified      156 ms
  node       down             unclassified      112 ms
  node       stale epoch      stale_epoch         4 ms
  blobstore  500              unclassified        4 ms
  blobstore  malformed-json   unclassified        4 ms
  blobstore  huge-body        unclassified     2046 ms
  blobstore  slow-headers     timeout          1004 ms
  control    down             breaker_open        0 ms
  control    back up          ok                  6 ms
      reset_peer: outcome "unclassified" (... read: connection reset by peer); want "unreachable"
      500: outcome "unclassified" (blobstore: bad response: invalid character '<' looking for beginning of value); want "remote_5xx"
      huge-body: took 2.046173366s; budget 1s
FLAW REPRODUCED: chaos_tests_failed 5 (want == 0)
```

After:

```
  node       reset_peer       unreachable       153 ms
  node       down             unreachable        38 ms
  blobstore  500              remote_5xx          5 ms
  blobstore  malformed-json   bad_payload         4 ms
  blobstore  huge-body        remote_5xx        139 ms
  blobstore  slow-headers     timeout          1002 ms
  ... (the other seven rows as before)
IMPROVED: chaos_tests_run 12 == 12; chaos_tests_failed 0 == 0
```

The rows that already passed before are the failures earlier lessons fixed. Timeouts (lesson 2)
end every hang inside its budget; the slicer test shows the streaming reader (lesson 8)
reassembles an answer that arrives 64 bytes at a time; the breaker (lesson 5) opens and recovers.
The five that failed are this lesson's: failures the clients noticed but could not name, and an
error page the blob client read all ten megabytes of, taking two seconds to say "error". After the
patch, each failure has a class a caller can act on (retry an unreachable node, do not retry a bad
payload, give a slow partner time, page a human for a 5xx), and no error answer costs more than
1 MB.

## Try changing

- Add a mode to the fake that answers 200 with `Content-Type: text/html` and a JSON body, and
  decide what the client should call it.
- Point the node-client tests at node-b and toxic node-a. Every test should pass in milliseconds,
  which is a check that the harness is testing what you think it is.
- Add the matrix for the control client (resolve under `latency`, `reset_peer`, `down`) and for
  the metering client, then run `make test-chaos` before every release.

## Read more

*Release It!* 2nd ed., Ch. 5 "Test Harnesses"; Ch. 15 "The Testing Gap"; Ch. 17 "Chaos Engineering".
The Toxiproxy README for the full list of toxics.

## Deviations from the plan

- **The suite talks to Toxiproxy's HTTP API instead of using its Go client.** The REF run compiles
  today's tests against an older tree's `go.mod`, which does not have that dependency.
- **5 of 12 tests fail before, not 4.** `slow-headers` already passes, because lesson 2's
  response-header timeout ends it at 1 s, and the standard error types call that a timeout. The two
  connection-level faults (`reset_peer`, `down`) were unclassified, which the plan did not list.
- **Two commits** (about 500 lines; plan ~320): the clients' classification lives in a shared
  `httpx` file both clients use, and the test file is about 280 lines.
- The tests classify through a `Kind()` method when the error has one. That is how the same test
  file can judge an older client that has no classes at all.
