# Experiment 05 — sequence-it: revealed

Build: plain export of `explore-it/lesson-04`, isolated stack `queuelet-e05` (API :18260, PG :15493, test clock on).
Explorer: fresh headless Claude Opus 5.5 session. Record: [notes.md](notes.md); walker: [walk.sh](walk.sh); helpers in
[tools/](tools/); raw output in `project/results/e05/`.

## Status

**done.** The charter's prediction ("every verb either does what it says from any starting point or is refused") was
refuted by the fourth probe: a job with `max_attempts` 2 was handed to four workers, each told it was attempt 1.

## Elapsed

13:17:51Z → 13:42Z (final `date -u` 13:41:39Z before closing notes), about 24 minutes of the 60-minute budget.

| Activity | Minutes (approx.) | Notes |
|---|---|---|
| Reading | 5 | STATE, charter, protocol, skill, SPEC, models, lesson 04, the lifecycle source (13:17:51–13:21:30, plus short reads during the run) |
| Probing | 12 | directed pairs and triples 13:22:22–13:24:41; walks and minimal reproductions 13:26:45–13:40:11 (about 5 min of that was spent waiting on background walks while I wrote reports) |
| Writing | 7 | predictions before each probe, three defect reports, models, this file |

## What the system showed

- **Worker deregistration bypasses the retry budget.** `DELETE /v1/workers/{id}` puts the worker's leased jobs back in
  `queued` without counting an attempt and without writing to the attempts history. The next lease reports the same
  attempt number. This can repeat without limit: 4 deliveries on max 2 (P6), 2 deliveries on max 1 (H-E05-12), and attempt 2
  handed out twice after a real failure (H-E05-11). In walk seed 32, job 212 ran four times before it went `dead` on
  max 2. Lease expiry and nack, which the code treats as the same "holder gone" situation, both count the attempt.
  → [E05-F01](../../defects/E05-F01-worker-delete-release-uncounted-attempt.md)
- **Ack and cancel leave a live-looking lease on terminal jobs.** `succeeded` and `cancelled` rows keep `lease_token`,
  `leased_by` and an unexpired `lease_expires_at`, and `GET /v1/jobs/{id}` returns them. A job cancelled while leased
  has no attempt row for that delivery, and it keeps naming its worker after that worker is deleted. Nack, expiry and
  worker release all clear these fields.
  → [E05-F02](../../defects/E05-F02-terminal-rows-keep-lease-fields.md)
- **After one lease expiry, the event history mislabels later exits from `leased`.** Ack (to `succeeded`), cancel and
  worker release are all logged as `lease_expired`, in both `/v1/jobs/{id}/events` and the API log. The trigger decides
  on the sticky `last_error` text.
  → [E05-F03](../../defects/E05-F03-worker-release-logged-as-lease-expired.md)
- **What held:**
  - purge → enqueue with the same idempotency key (P1);
  - purge skips a leased job, and a later nack sends it to `retry_wait` (P3);
  - cancel/retry/terminal combinations return 409 (P4);
  - pause blocks leases, does not stop expiry or promotion, and produces no catch-up burst after +1 h and resume (P2, P7);
  - cancelled scheduled and retry_wait jobs stay cancelled past their due time (P8);
  - a released holder's token gets 409 (P6);
  - across 950 random-walk steps (5 completed seeds), no violation of I1, I2, I3, I5, I6, I9, I10, I12, I13 or the
    stale-token rule.
- **Surprises that are not defects:**
  - Every clock `advance` or `set` freezes semantic time (clock.go:182-191). This is documented lab behaviour, recorded
    as testability request T-E05-1.
  - A dead job keeps a future `run_at`. No reader uses it.
  - A scheduled job is promoted to `queued` while its queue is paused. It cannot be leased, so nothing is delivered.
  - A leased job survives purge. It is not "pending" in the SPEC's sense.
  - `GET /v1/jobs/{id}` exposes the live lease token to the whole namespace. Saved as `#offcharter`, not judged.

## Model delta

- [states.md](../../models/states.md), E05 revision: lifecycle diagram for the five exits from `leased` and the pause
  arrows, plus 18 starred state/event cells (should, predicted, observed).
- [rules.md](../../models/rules.md), E05 revision: invariants I1–I13 as walk checks, new rules R44–R48, and outcomes for
  R2, R3, R5 and R6.
- [nouns-verbs.md](../../models/nouns-verbs.md), E05 revision: the sequence grid (directed pairs/triples and walks) and
  the sequences still untested.
- [causes.md](../../models/causes.md): C25 (each exit from `leased` decides on its own which columns to reset) and C26
  (observation keyed on sticky error text), with timestamped sibling predictions.
- [hypotheses.md](../../hypotheses.md): H-E05-C, H-E05-01 to H-E05-21, Q-E05-1/2, T-E05-1 and one `#offcharter` item.

## Unresolved and next hypotheses

- **Q-E05-1 (owner):** should a release by worker deregistration count as an attempt, or be recorded as a distinct
  release with its own cap?
- **Q-E05-2 (owner):** should a terminal row keep `leased_by` (who completed it) while clearing the token and expiry?
- **H-E05-20:** a nack whose error text is exactly `lease expired` will be logged as an expiry (not probed).
- **H-E05-21:** a worker that deregisters on SIGTERM while executing turns a crash-looping poison job into unbounded
  redelivery (ecosystem).
- **T-E05-1:** add an "offset and keep ticking" clock mode so clock jumps and real-time lease expiry can be combined.
- **Walker coverage limits:**
  - Successful holder verbs were rare in the uniform walks (0–4 acks per seed). Holder-biased v2 raised this to 7 and 12.
  - Operator retry succeeded only 6 times in all walks.
  - Schedules, queue DELETE, retention and webhooks were not part of the walk.
  - One rejected or skipped operation is not coverage of the success arrow.
- **Dead ends kept:**
  - The seed-1 smoke run crashed on `set -u`.
  - Seed 21 starved: no workers left and the queue paused about half the time.
  - My broad `pkill -f walk.sh` also killed my own shell and stopped seed 22.

## Score

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 22 | H-E05-C, H-E05-01 … H-E05-21, each written in notes before its probe (13:22:19Z, 13:24:30Z, 13:29:40Z, 13:34Z, 13:33Z). H-E05-20/21 are backlog, not probed. Invariants I1–I13 are rules, not counted here |
| Hypotheses confirmed | 18 | H-E05-01…09, 11…19 (P1–P8, isolate.txt, f02-f03.txt, f03-sibling.txt, walk-b31/b32.txt). H-E05-19 is confirmed on its stated refutation condition; its coverage sub-prediction missed for seed 31 |
| Hypotheses refuted | 2 | H-E05-C (P6, E05-F01); H-E05-10 (I7 fired on succeeded rows, walk-11 step 24) |
| Code-reading predictions contradicted by the running system | 2 | H-E05-C (handler reading); H-E05-10 (I had read ack.go:25 and predicted I7 for cancel only) |
| Defects found | 3 | [E05-F01](../../defects/E05-F01-worker-delete-release-uncounted-attempt.md), [E05-F02](../../defects/E05-F02-terminal-rows-keep-lease-fields.md), [E05-F03](../../defects/E05-F03-worker-release-logged-as-lease-expired.md). Repeated walk hits of the same causes are counted once; regressions that held (F06/R6, F25 purge, E04-F04 token-after-expiry) are not counted |
| Model cells filled | 16 | States E05 table: 18 starred cells, minus 2 revisits already observed earlier (cancelled→retry in E04 P5; succeeded→retry in E01 F06) |
| Open | 2 | H-E05-20, H-E05-21 (not probed) |
