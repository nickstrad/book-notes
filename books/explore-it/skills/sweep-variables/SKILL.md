---
name: sweep-variables
description: List every variable of an input surface (fields, counts, sizes, positions, timing, config, request shape, and each path a value travels), push each with named heuristics through the rawest instrument, and record a variation table; use for an API body, a config, a CLI with options, or "what values should I try".
---

# Sweep variables

A system that rejects the values you thought of can still accept the ones you did not. This skill finds those
values. You list the variables before choosing any value, predict each one from the code, vary them one at a time
through a raw client, and record every row, including the ones you did not try. In this repo's experiment arc,
also follow [the protocol](../../exploration/PROTOCOL.md).

## When to use

- A request body, query string, config or command with options that you have only seen on the happy path.
- Validators exist and you need to know what they miss: upper bounds, sign, case, framing, encoding.
- Numbers or durations that are multiplied, shifted or stored in fixed-width columns.
- A value that crosses several decoders (API, CLI, worker, webhook, database), or a fresh fix to input handling.

Not for: verb orders and lifecycles (`probe-sequences-and-states`), or neighbours that misbehave
(`probe-entities-and-boundaries`).

## Inputs

A bounded surface (named endpoints, commands or config keys) and a time budget; the spec, the validator and
decoder source and the storage migrations; a raw client (`curl`, a short script) as well as the product CLI; a
resettable lab with read-only database access, logs and partner records; an oracle per variable (spec line, RFC,
Never/Always rule, or a stated approximation).

## Procedure

1. **Pick the rawest instrument.** A product CLI sends only what the API expects and may change values on the way.
   Send bodies from files (`curl --data-binary @f.json`) and print status and body. Script in bash or Python, not
   shell `echo` (zsh expands `\u` escapes). Save each helper beside the notes so every logged command replays.
   (queuelet: qctl rounded payload numbers — [U30](../../exploration/defects/U30-qctl-rounds-payload-numbers.md).)
2. **Inventory variables before values.** Write a table: family, variable, heuristics, paths. Cover fields (type,
   presence, length, charset, sign), counts, sizes, positions, selection, timing, order and config × count
   interactions. Always add the four families a first recon tends to miss:
   - *Request shape:* trailing data, a second concatenated object, member-name case, exact duplicate keys,
     invalid UTF-8, Content-Type, empty, `null` and array bodies.
   - *Time arithmetic:* seconds multiplied into a fixed-width duration; backoff that shifts or multiplies by the
     attempt. Grep `time.Duration(`, `* time.Second`, `<<`, `interval`.
   - *Limits the store enforces:* `int` columns, index row size, timestamp range. Grep migrations for `int`,
     `unique`, `index`, then read how each store error maps to a status.
   - *Accepted input with no consumer:* grep each decoded field for a reader. Stored and echoed but never read is a finding.
   In the paths column, list every hop the value travels: server decoder, storage, read or lease path, CLI input,
   CLI display, worker, outbound webhook. Each path becomes its own row.
   (queuelet: all eight E03 faults sat in those four families — [lesson 03](../../lessons/03-vary-it.md).)
3. **Read the validators and predict.** Per variable, note which checks exist (lower bound, upper bound, enum,
   regex) and which are missing. Note decoder leniencies: one `Decode` call with no end-of-input check, member
   names matched without regard to case. Write the prediction, its source `file:line` and the refutation condition
   before the probe. Keep *should* (named oracle), *predicted* (code) and *observed* (ran) apart.
4. **Turn heuristics into concrete values,** with boundaries taken from the representation:
   - Boundaries, Numbers: 0, −1, the largest valid value and one more: 2^31−1/2^31 for an `int` column, 2^53/2^53+1
     for float64, ≈9 223 372 036 s for an int64-nanosecond duration, the timestamp minimum. Both signs, out to the
     limit. Also `5.0`, `"5"`, `1E2`, `-0`, `1.50`, `1e400`, a 20-digit decimal.
   - Strings: case variants of member names, whitespace only, NUL, a raw `0xff` byte, a lone surrogate, random
     versus repetitive long values (compression moves storage limits).
   - Zero/One/Many, Count: 0, 1, the documented maximum, maximum + 1. Goldilocks: body size around the cap.
   - Beginning/Middle/End, Position: the bad element first, middle and last; extra data at the end of the body.
   - Some/None/All, Selection: omitted, `null`, `{}`, a partial nested object, in create and in PATCH.
   - Paths: the same value sent once along each path from step 2.
5. **Vary one variable at a time, with a control** just inside the boundary (9 223 372 036 s beside 9 223 372 037 s).
   Use a fresh single-purpose resource per probe and act on the id the system returned. Read the persistent
   effect, not only the status: the stored row, computed times such as `run_at`, `/health`, logs, other tenants.
   (queuelet: FIFO handed back an older job twice in E03 and once broke a repro —
   [notes](../../exploration/experiments/03-vary-it/notes.md), [lesson 03](../../lessons/03-vary-it.md).)
6. **Signals:** a 2xx where the spec implies 4xx; a "later" that becomes "now" or another century; a value echoed
   that changes nothing; a range error reported as 409 or 500 with driver text (`pq:`); one tenant's setting that
   stops a shared background step; a different value at a later hop than at the first.
7. **Characterize each surprise:** bracket the threshold (5e7 h fine, 6e7 h fails), find what does not matter and
   how it recovers, and predict siblings from the code before probing them. Each surprise ends in a report (see
   `follow-the-surprise`) or a written reason it is not a bug.
8. **Re-sweep around existing fixes.** Probe what a fix did not ask: case after an unknown-field check, upper bound
   after a lower bound, sign after an upper bound, framing after a number-decoding change, other decoders after
   one is repaired. (queuelet: [U12](../../exploration/defects/U12-case-variant-member-names-accepted.md),
   [U14](../../exploration/defects/U14-storage-limit-errors-reported-as-conflict.md),
   [U21](../../exploration/defects/U21-negative-delay-accepted.md).)
9. **Stop and account.** At the budget, list untried rows and untried paths as backlog. "Spec silent" is a
   question with an id, not a verdict: if the computed effect breaks a rule (a retry before its failure), file it.

## Output

- Inventory table: family, variable, heuristics, paths.
- Variation table, one row per variable × value × path: family, variable, heuristic, values, path, oracle,
  predicted (source), observed, verdict/evidence. Mark new and regression rows. Only what ran counts as observed.
- An untried list, and non-bug judgments with reasons and spec-question ids.
- In this repo: `experiments/NN-*/notes.md`, the table appended to
  [models/variables.md](../../exploration/models/variables.md), reports under `exploration/defects/`.

## Pitfalls the experiments exposed

- **The first list was too narrow.** E01's recon recorded about sixteen variable rows; all eight E03 faults lived in
  families it never named ([lesson 03](../../lessons/03-vary-it.md)). Step 2 now names them.
- **Fixes closed a value, not the family.** The lesson-01 unknown-field fix ignored case
  ([U12](../../exploration/defects/U12-case-variant-member-names-accepted.md)); the lower-bound fix ignored upper
  bounds ([U14](../../exploration/defects/U14-storage-limit-errors-reported-as-conflict.md)). In E06 a
  number-precision fix to client decoders first reintroduced trailing-data acceptance
  ([E06 review](../../register/reviews/E06/review.md), finding 1). Step 8 re-sweeps the neighbours.
- **Sign was dismissed as a spec silence.** E03 judged negative `delay_seconds` "due now, not a bug" (E03-H05). E04
  pushed it to −9 223 372 037 s on nack too: a retry landed before its failure, and large negatives wrapped 292 years
  forward ([U21](../../exploration/defects/U21-negative-delay-accepted.md)). Steps 4 and 9 cover this.
- **The sweep stopped at one path.** E03 confirmed exact digits in raw lease bytes (E03-H21) and stopped. E06 found
  the sample worker and qctl rounding the same numbers
  ([U29](../../exploration/defects/U29-sample-worker-echo-loses-precision.md),
  [U30](../../exploration/defects/U30-qctl-rounds-payload-numbers.md)). The worker's http-kind response decoder
  still uses float64 and is unprobed through the actual worker, an open item in [STATE.md](../../STATE.md).
  Step 2 now gives each path its own row.
- **The instrument changed the value.** A Python printer showed `1.5` and `1.23e+19` for exact lease bytes
  ([E03 notes](../../exploration/experiments/03-vary-it/notes.md), 04:22:49Z); hooksink's parsed `body` field
  rounds too ([E06 knowledge](../../docs/knowledge/06-data-boundaries.md)). Read raw bytes; parse as Decimal.

## Source

Elisabeth Hendrickson, *Explore It!* (Pragmatic Bookshelf, 2013): Chapter 4, "Find Interesting Variations"
(4.1–4.4); Appendix A2, "Test Heuristics Cheat Sheet", A2.1 (Some/None/All, Zero/One/Many, Beginning/Middle/End,
Goldilocks, Boundaries, Count, Position, Selection, Strings, Numbers, Paths); 10.1, "Exploring an API". Section
numbers follow the [plan's chapter map](../../PLAN.md), checked against the publisher's table of contents; the
steps have not been validated against the full book text. Procedure wording, the four extra families and the path
rows are this repo's. Lessons: [03 Vary it](../../lessons/03-vary-it.md), [04 Judge it](../../lessons/04-judge-it.md)
(sign, U21), [06 Relate it](../../lessons/06-relate-it.md) (numbers along every path, U29/U30).

## Worked example

Experiment 03, blind, against `explore-it/lesson-02` ([charter](../../exploration/experiments/03-vary-it/charter.md),
[notes](../../exploration/experiments/03-vary-it/notes.md), [revealed](../../exploration/experiments/03-vary-it/revealed.md),
helper [namelen.py](../../exploration/experiments/03-vary-it/tools/namelen.py)). Charter: *explore enqueue, batch
enqueue and queue configuration with the variables list, the heuristics cheat sheet and raw `curl` requests to
discover inputs the system accepts but cannot honour.* About 21 of 60 minutes.

**Score:** 23 hypotheses stated, 21 confirmed, 2 refuted; 2 code-reading predictions contradicted by the running
system; 8 defects found; 28 model cells filled.

**Found** (all unplanned; the planned range and shape defects were fixed in lesson 01 and held as regressions):
[U11](../../exploration/defects/U11-malformed-json-text-accepted.md) trailing data / invalid UTF-8,
[U12](../../exploration/defects/U12-case-variant-member-names-accepted.md) case-variant names,
[U13](../../exploration/defects/U13-duration-overflow-delay-and-extension.md) delay/nack/heartbeat overflow,
[U14](../../exploration/defects/U14-storage-limit-errors-reported-as-conflict.md) range errors as 409/500,
[U15](../../exploration/defects/U15-large-retention-halts-global-retention.md) one queue's retention halts all tenants',
[U16](../../exploration/defects/U16-backoff-overflow-immediate-retry.md) backoff overflow,
[U17](../../exploration/defects/U17-backoff-max-seconds-ignored.md) `max_seconds` ignored,
[U18](../../exploration/defects/U18-patch-webhook-partial-clears-other-member.md) partial webhook PATCH.

**Wrong or missed.** The charter prediction ("every documented range is enforced and unknown fields are rejected")
fell on the first probe. E03-H11 predicted a huge queue name would be stored but unaddressable; the unique index
refused it with 409 at a compressibility-dependent length. Missed, found later: negative delays (U21, E04) and
client-side number decoding (U29/U30, E06).

**Model delta:** [variables.md](../../exploration/models/variables.md) gained the E03 variation table (19 new and 6
regression rows by the lesson's count) and an untried list; [rules.md](../../exploration/models/rules.md) R28–R30; [causes.md](../../exploration/models/causes.md) C15–C19.

**Excerpt** ([E03-F03](../../exploration/defects/E03-F03-duration-overflow-delay-and-extension.md)): a boundary
from the representation, with its control.

```text
delay 9223372036          → 201 scheduled, run_at 2319-02-01 04:04:49 (honoured)
delay 9223372037          → 201 queued,    run_at 1734-07-13 04:30:16 (wrapped ≈292 y into the past)
delay 9223372036854775807 → 201 queued,    run_at now − 1 s
```

**Second target: litefleet (not yet run).** PLAN section 7 names none for this skill; this is a sketch. On
litefleet's tenant API (`POST /v1/databases {name}`, `/query {sql, args}`, `/execute {statements}`,
`GET /v1/admin/usage?window=`), inventory each body's request shape; the database name (case, charset, length);
`args` types, with integers above 2^53 sent along every path (gateway → node → SQLite → JSON response); statement
count (0, 1, many) and the position of a failing statement in one transaction; result row counts; and `window`
(0, negative, huge, malformed). Oracles: the Release It! plan's API table and SQLite's type limits.
