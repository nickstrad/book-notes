# Experiment 03 — vary-it: revealed

Status: **done** (one run). Export `explore-it/lesson-02`, own stack `queuelet-e03`; [notes](notes.md) hold every command,
trimmed output, dead end and timestamp. Raw output is under `project/results/e03/`; replay helpers are in [tools/](tools/).

## Elapsed

Run started 04:11:52Z (`run-start.txt`); record closed ≈04:33Z — about 21 minutes of a 60-minute budget.
Split (from note timestamps; boundaries overlap because notes were written between probes):

| Phase | Window | Approx. |
|---|---|---|
| Reading (state, charter, protocol, skill, SPEC, prior record, validators) + hypotheses while the stack built | 04:11:52–04:16:44 | 5 min |
| Probing (11 probe batches, inline note writing) | 04:16:44–04:25:04 | 8 min |
| Writing (8 reports, models, backlog, this file, citation checks) | 04:25:04–≈04:33 | 8 min |

## What the system showed

The lesson-01 validation fixes hold: unknown member names, bad enums, priority bounds, name charset, positive limits, batch
count and batch atomicity were all rejected correctly (E03-H03, regression only). The charter prediction — "every documented
range is enforced and unknown fields are rejected" — was nevertheless refuted, because the remaining gaps are in
the parts that nothing validates:

- **The request shape.** The decoder reads one JSON value and stops, so trailing junk or a second object gets a 201 and the
  rest of the body is lost (E03-F01). It also matches member names regardless of case, so `KIND`, `Priority` and
  `LEASE_TOKEN` count as the documented fields, and whichever spelling comes last wins (E03-F02).
- **Time arithmetic.** `delay_seconds`, nack `delay_seconds` and heartbeat `extend_seconds` above ≈9.22e9 s wrap
  around in Go's `time.Duration`, so "much later" becomes "now" (a heartbeat even ends its own lease and uses up an attempt)
  (E03-F03). Queue backoff overflows in the same way. With the default base of 5 s, the 32nd attempt is scheduled
  for 1782 and retried at once (E03-F06).
- **Database limits standing in for validation.** Upper bounds come from `int` columns, the btree index size and the
  timestamp range. They show up as 409 "conflict" on create, 500 on PATCH, or driver text (E03-F04). The worst case is
  `retention_hours`: a valid nonnegative value ≥ ~6e7 makes the single global Retain statement fail on every tick once
  that queue holds a terminal job. Retention then stops for every namespace and /health stays 503 with no cause given
  (E03-F05).
- **Accepted input that does nothing.** The undocumented `backoff.max_seconds` is stored and echoed back but never applied
  (E03-F07). A PATCH with a partial `webhook` object clears whichever member it leaves out. Rotating only the secret turns
  notifications off, and changing only the URL drops the signing secret (E03-F08).

Payload fidelity held at the worker: the raw lease bytes were exact (`1.50`, 20-digit decimals, 2^53+1). jsonb changes the
representation but not the value (1E2→100, key order, duplicate keys collapsed).

## Surprises and their disposition

| Surprise | Disposition |
|---|---|
| S1 trailing data / second object accepted | [E03-F01](../../defects/E03-F01-malformed-json-text-accepted.md) |
| S2 case-variant member names | [E03-F02](../../defects/E03-F02-case-variant-member-names-accepted.md) |
| S3 delay overflow | [E03-F03](../../defects/E03-F03-duration-overflow-delay-and-extension.md) (+ nack delay, heartbeat siblings) |
| S4 409/500 for out-of-range numbers | [E03-F04](../../defects/E03-F04-storage-limit-errors-reported-as-conflict.md) (+ name length) |
| S5 retention poison | [E03-F05](../../defects/E03-F05-large-retention-halts-global-retention.md) |
| S6 backoff overflow | [E03-F06](../../defects/E03-F06-backoff-overflow-immediate-retry.md) |
| S7 max_seconds ignored | [E03-F07](../../defects/E03-F07-backoff-max-seconds-ignored.md) |
| S8 partial webhook PATCH | [E03-F08](../../defects/E03-F08-patch-webhook-partial-clears-other-member.md) |
| Name-length limit depends on compressibility (H11 refuted) | folded into E03-F04 (it is the same misreported storage limit, not a separate fault) |
| Duplicate exact keys, Content-Type ignored, negative delay, whitespace kind, µs rounding, jsonb normalisation, jsonb-unstorable payloads, 2 MiB cap, `1e3` rejected, PATCH message names lease_seconds | not bugs: reasons inline in notes (spec silence or safe rejection); spec questions Q-E03-1…5 |
| Lease view showed `1.5` / `1.23e+19` | instrument artefact (my Python printer parsed floats); raw-bytes rerun H21 showed exact digits |

Dead ends kept in the notes: FIFO interference (twice the lease returned an older re-queued job and the intended nack/ack got
409), zsh argv/word-splitting and `echo` escape expansion (reran under bash/Python), shell `curl` with a 1.1 M-character URL.

## Model delta

- [models/variables.md](../../models/variables.md) — E03 variation table: 19 new (★) variation rows with oracle, observation
  and verdict; 6 regression rows (r); untried rows listed.
- [models/rules.md](../../models/rules.md) — E03 revision: new R28 (backoff monotone), R29 (future never "now"), R30 (PATCH
  omissions keep), new scopes for R7, R8, R9, R10, R11, R25.
- [models/causes.md](../../models/causes.md) — C15 lenient decoder framing, C16 unchecked Duration arithmetic, C17 storage
  limits as validation, C18 accepted input with no consumer, C19 nested PATCH replacement; sibling predictions timestamped.
- [hypotheses.md](../../hypotheses.md) — E03-HC, E03-H01…H22, Q-E03-1…5, backlog H53…H56; H46 outcome updated (refuted:
  negative retention is now rejected).
- Testability requests TR-E03-1 (health names the failing step) and TR-E03-2 (expose computed retry wait) in notes.

## Unresolved uncertainty and next hypotheses

- Schedule create/PATCH bodies are predicted (from code) to share E03-F01/F02 and might carry their own payload or overflow
  paths (H53, #offcharter). The clock route and worker create were not probed for F01/F02.
- Lease `max_jobs` / list `limit` / `cursor` coercion (H54); enqueue maps a real DB outage to 400 (H55, inferred);
  `lease_seconds` 2147483647 expiry (H56).
- Exponential backoff at attempt 33 (wraps to the future) and ≥65 (shift gives 0) is arithmetic inference only.
- E03-F08's oracle is an approximation; it needs a spec line on nested PATCH semantics. E03-F07 needs a decision between
  rejecting the field and implementing it.
- Storage amplification from exponent-form numbers (Q-E03-5) was deliberately not stress-tested.

## Score

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 23 | E03-HC (charter prediction) + E03-H01…H16 stated 04:16Z before any probe; H17…H21 stated 04:24Z and H22 at 04:25Z, each before its probe |
| Hypotheses confirmed | 21 | H01–H10, H12–H22 (notes 04:16:44Z–04:25:00Z). H03, H21, H22 are regression confirmations, not discoveries |
| Hypotheses refuted | 2 | E03-HC (2xx for malformed / case-variant / undocumented-field bodies, notes 04:16:44Z, 04:17:59Z); E03-H11 (409 at the index instead of an unaddressable queue, 04:22:00Z–04:22:28Z) |
| Code-reading predictions contradicted by the running system | 2 | E03-H11 (source: queues.go:38 + Go header limit) and E03-HC (charter source: reading the validators in internal/api) |
| Defects found | 8 | E03-F01…E03-F08 (distinct reports; sibling cases folded into their parent reports, not counted separately) |
| Model cells filled | 28 | 19 ★ rows in variables.md E03 table + 9 ★ rows in rules.md E03 revision, each backed by a runtime observation; cause rows (mechanism inference) and regression rows not counted |

Hypotheses left `needs more`: none of E03-H01…H22. Backlog H53–H56 and Q-E03-1…5 remain open.
