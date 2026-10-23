# 07 — State it: a complete table can still miss an overlap

A state table asks what each event may do in every state, including the cells with no
arrow in the diagram. Queuelet passed all 49 cells the explorer drove. The surprises
appeared when two actors overlapped: deleting a worker could leave its newly acquired
lease alive, and purging a queue could miss jobs crossing the leased/pending boundary.
The state model needed actors and timing as well as states.

The sources are Hendrickson, *Explore It!* (2013), Chapter 8.1 “Recognizing States and
Events”, 8.2 “Diagramming a State Model”, 8.3 “Exploring with State Models”, and 8.4
“Changing the Representation: State Tables”. These are **approved-plan references**
([E07](../PLAN.md#e07--state-it-the-state-table-and-its-sneak-paths)); full book-text
validation is still pending. The concurrency conclusions below come from this project's
evidence, not an assertion that the book describes PostgreSQL locking.

The blind [exploration record](../exploration/experiments/07-state-it/revealed.md),
committed as `49d8d67`, remains unchanged. It contains 13 stated hypotheses, 12 confirmed,
1 refuted, 1 contradicted code-reading prediction, 2 reports and 41 newly observed model
cells. The complete matrix contains 49 cells, including prior observations. Three post-hoc
predictions were excluded from the score. Process timestamps give about 18 minutes;
the explorer estimated about 20. The reconciliation preserves both the timing correction
and the runtime-task-output boundary exception; this was not an OS-isolated export.

H-E07-13 predicted availability within 1.5 seconds from reading the one-second loop tick.
A sweep of 1,110 expired jobs delayed an unrelated promotion beyond three seconds. The
quiet rerun held, but did not erase that contradiction. No latency limit is added here.
The forward question remains how to observe sweep duration and its effect on other work.

D18 (stale ack), D19 (nack in the wrong state), D20 (expiry at the attempt limit), and
D14 (retry from terminal states) were repaired in E01. They were **early fixes**, not
E07 discoveries or missed planted defects. U34 and U35 are the two new findings. A3 is
a separate coordinator policy decision: this blind run did not probe cancellation with
a running sample worker. The fixer supplied that evidence without backfilling the record.

## The faults at lesson-06

All source pointers in this table refer to `explore-it/lesson-06`, relative to `project/`.

| Item | Previous-tag location | Before | Repair |
|---|---|---|---|
| [U34](../exploration/defects/U34-worker-delete-racing-lease.md) | `internal/jobs/jobs.go:57`; `internal/api/jobs.go:160`; `internal/jobs/expire.go:37` | Lease reads a worker without holding its identity stable. Concurrent deletion can miss its leases; deletion between the API lookup and transaction instead produces 500 `sql: no rows` | Hold a shared worker-row lock through lease commit; map the transaction's missing worker to 404 |
| [U35](../exploration/defects/U35-purge-transition-race.md) | `internal/api/queues.go:170`; `internal/jobs/expire.go:44` | Two state-filtered writes see one snapshot. A row that changes between leased and pending can fail both predicates | Lock the queue's pending and leased rows first, then classify their current states in subsequent statements |
| A3 policy | `internal/jobs/operator.go:6`; `internal/worker/worker.go:52`, `:59`, `:63`, `:84`, `:101`, `:106` | Cancel revokes the lease immediately, but records no cancelled attempt; worker ignores heartbeat answers and keeps sleeping/reading HTTP | Record the cancelled delivery atomically; authenticate a read-only cancellation observation; propagate it into the running operation's context |

## Before and after the state table

This is the explorer's 7 × 7 matrix, preserved in size and event selection. `409` means
unchanged state, lease, result, attempt count and attempt history; `[]` is an empty lease.
The same transition/status table holds before and after. The marked cell changes its
attempt accounting, as required by the A3 decision.

| State | Lease by B | Wrong heartbeat | Wrong ack | Wrong nack | Cancel | Retry | Delete |
|---|---|---|---|---|---|---|---|
| scheduled | [] | 409 | 409 | 409 | cancelled | 409 | 409 |
| queued | leased(B) | 409 | 409 | 409 | cancelled | 409 | 409 |
| leased(A) | [] | 409 | 409 | 409 | cancelled* | 409 | 409 |
| retry_wait | [] | 409 | 409 | 409 | cancelled | 409 | 409 |
| succeeded | [] | 409 | 409 | 409 | 409 | 409 | deleted |
| dead | [] | 409 | 409 | 409 | 409 | queued | deleted |
| cancelled | [] | 409 | 409 | 409 | 409 | 409 | deleted |

`*` Before: cleared lease, no attempt for the interrupted delivery. After: cleared lease,
exactly one `cancelled` attempt with the holder and lease start, and attempts incremented
once. Cancelling an unleased job does not consume an attempt.

The executable table uses an independent expected-status grid, fresh fixtures for each
cell, and full job/history comparisons for every rejection or empty lease. It runs **49
subtests: 34 rejected cells, 6 empty-lease cells, 9 accepted transitions**. It also asserts
the accepted states, holder, lease fields, attempt budget, and preserved retry history.
The plan's illustrative 44 cells are not a measurement and are not manufactured here.

| Additional actor/timing cell | Lesson-06 | Lesson-07 |
|---|---|---|
| lease(W) overlapping DELETE W | zombie lease or worker-disappearance 500 possible | released with one attributable attempt, or 404 without mutation |
| purge overlapping lease | unflagged leased survivors | survivors marked; subsequent nack is dead |
| purge overlapping nack | retry_wait survives | dead or deleted |
| purge overlapping worker release | queued survives | dead or deleted |
| purge overlapping expiry | queued survives | dead or deleted |
| cancelled holder heartbeat | ordinary 409, no observation | 409 plus authenticated `cancel_requested:true`, no mutation |
| cancelled holder's late ack/nack | 409 | still 409, no revival or second attempt |
| cancellation during sample sleep/HTTP | operation continues | operation observes cancellation and stops cooperatively |

The Go checks name **13 extra cells**, outside the 49: three current-holder verbs,
three expired-holder verbs, and seven cancellation observation/repeat/late-completion
cases. Separate regressions exercise U34/U35/A3 and the transaction's missing-worker
sentinel. The old D18–D20 instruments run unchanged at both tags as passing controls;
they never enter E07's historical new-finding count.

## Focused implementation

`Lease` (`internal/jobs/jobs.go:57`) takes `FOR SHARE` on the worker before locking jobs. Concurrent leases may share
that lock; worker deletion waits for them to commit, then releases their leases. A delete
that wins first leaves no worker row and returns the explicit missing-worker error,
which the API translates to 404. This does not add a global or queue-wide lease mutex.

`Purge` (`internal/jobs/purge.go:18`) first locks the queue's pending and leased rows in id order. A row
that moves between pending and leased while purge waits still matches that set, so it is
locked rather than skipped; terminal rows are not locked. After waiting, its separate
UPDATE/DELETE statements see the latest committed states. Both writes (`purge.go:38` and
`:41`) are restricted to the locked ids; a concurrently enqueued job may belong after
the purge. Pending jobs and their attempts are deleted; leased jobs are marked, may
still succeed, and become dead on an unsuccessful exit. Worker release uses the same id
lock order. Expiry orders by id and keeps `SKIP LOCKED`; leasing keeps its existing
priority/FIFO order and `SKIP LOCKED`. Purge never waits for a worker lock while holding
job locks. This avoids the opposite lock order without serializing unrelated queues.

Cancellation (`internal/jobs/operator.go:18`, `:26`, `:31`) locks the job, records a leased delivery's attempt, stores a hash of its
revoked token, and makes the job terminal in one transaction. The three active lease
fields remain NULL. The hash supports an observation; it grants no completion rights.
Existing cancelled jobs have no new token history invented for them.

**A3 decision: Cancellation immediately makes a nonterminal job cancelled and revokes its lease; the worker cooperatively stops in-flight sleep/HTTP on heartbeat cancel_requested, and late ack/nack cannot succeed or revive it; already-performed remote effects cannot be rolled back.**

The exact [SPEC](../project/SPEC.md#operator-actions-and-retention) wire contract is:

- A current live holder receives 200 `{"ok":true,"cancel_requested":false}` and an extension.
- The cancelled delivery's exact revoked token receives 409 with `error.code=state_conflict`,
  `error.message="job cancelled"`, and top-level `cancel_requested:true`.
- Wrong/empty tokens and other nonleased jobs keep ordinary 409 without that signal.
  Missing/foreign jobs remain 404; extension validation still precedes observation.
- Observing repeatedly changes no lease, job or attempt. The signal lasts only until
  the job is removed; legacy cancellations and jobs cancelled before leasing have none.

The sample worker (`internal/worker/worker.go:55`, `:68`, `:100`, `:120`) polls every second while performing work and cancels the sleep timer
or HTTP request/body context on that signal. Heartbeat's existing request timeout is
15 seconds; remote HTTP retains its five-second timeout. API failures, deletion before
observation, or effects performed before observation prevent a promise of immediate
remote stopping. The ordinary late-completion guard still applies. HTTP numeric decoding
is unchanged and remains the separately deferred E06 review issue.

## Reproduce and check

From `books/explore-it/project`, set the isolated environment and run sequentially:

```sh
export COMPOSE_PROJECT_NAME=queuelet-fix07 QUEUELET_IMAGE_TAG=fix07
export QUEUELET_API_PORT=18290 QUEUELET_HOOK_PORT=18291 QUEUELET_PG_PORT=15523
export QUEUELET_URL=http://localhost:18290 QUEUELET_HOOK_URL=http://localhost:18291
export QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15523/queuelet?sslmode=disable'
make up
make verify LESSON=07
make verify LESSON=07 REF=explore-it/lesson-06
make checks
make test
go vet ./...
make down
```

The public [probe](../project/lab/repro/probe07.py) samples ordinary API races with bounded
jitter. It uses no fixture SQL writes or hidden synchronization hooks. Read-only SQL
checks stored state and attempt attribution. U34 requires both real historical symptoms;
U35 requires hits in all four reported paths. Missing controls, unexpected statuses,
transport/runtime failures and partial historical evidence are PROBE ERROR, not ABSENT.
No bounded sample proves the absence of every possible interleaving.

The REF fixture correction is described under deviations below. Every unexpected
outcome is still a PROBE ERROR, and no oracle was relaxed.

A3 runs the worker binary from the target's built Docker image, including the historical
image in REF. It observes processing and completion logs for an eight-second sleep and
a streaming HTTP response. The partner records continued writes after cancellation and
connection closure. Each worker must also complete a quiet echo job before cleanup.
Process survival and terminal API state alone do not count as stopping evidence. Container
stop is cleanup after the observations, never the cancellation mechanism.

### Results (final run, sequential, isolated stack `queuelet-fix07`)

Ledger: [07-opus-acceptance.json](../register/runs/07-opus-acceptance.json); one
`07-opus-<step>.log` per step. Instruments were frozen in
[07-instrument-hashes-opus.json](../register/runs/07-instrument-hashes-opus.json) before
the pair and rechecked after the suite (`07-opus-integrity.log`: no hash, raw-record,
tag or inherited-instrument mismatch).

| Step | Exit | Seconds |
|---|---|---|
| `make up` | 0 | 58.4 |
| `make verify LESSON=07` | 0 | 251.3 |
| `make verify LESSON=07 REF=explore-it/lesson-06` | 0 | 86.2 |
| `make checks` (81 check functions) | 0 | 514.4 |
| `make test` | 0 | 3.2 |
| `go vet ./...` | 0 | 0.3 |

At lesson-07:

```text
ABSENT: U34 — zombie leases=0; worker-disappearance 500=0; statuses={404: 78, 200: 306}; sequential release/404/next-attempt controls hold
ABSENT: U35 — purge race hits={'lease': 0, 'nack': 0, 'release': 0, 'expiry': 0}; all four paths sampled; …
ABSENT: A3 — actual image=queuelet:fix07; cooperative sleep/HTTP={'sleep': True, 'http': True}; revoked heartbeat/late completion 409; …
ABSENT: D18 / D19 / D20 (unchanged early-fixed controls)
FIXED AND CHECKED: 3 absent; regression checks pass
```

At lesson-06 (historical images, current scripts):

```text
PRESENT: U34 — zombie leases=80; worker-disappearance 500=3; statuses={200: 15, 404: 6, 500: 3}; …
PRESENT: U35 — purge race hits={'lease': 2, 'nack': 2, 'release': 2, 'expiry': 2}; all four paths sampled; …
PRESENT: A3 — actual image=queuelet:ref; cooperative sleep/HTTP={'sleep': False, 'http': False}; …
ABSENT: D18 / D19 / D20 (unchanged early-fixed controls)
DEFECTS REPRODUCED: 3 present
```

The verify run is about four minutes, not "well under a minute": bounded race sampling
and two real eight-second worker jobs dominate. This is a measured departure from the
plan's runtime target, kept rather than shortening the samples.

## Dwell and the departure from the proposal

Dwell is paired lesson-06/current evidence only; no full historical sweep was run. U34,
U35 and policy A3 are tested at those two points, without dating their introduction.
D18–D20 are ABSENT at both. The proposal's story of D18 dwelling through lessons 01–06
is counterfactual: E01 had already asked the stale-token question and repaired it.
E07 instead shows that all the individual guard cells can hold while concurrent
statements still violate the lifecycle. A3's running-worker evidence is a fixer policy
probe, not a third blind discovery.

## Try changing and read more

In a disposable copy, remove the shared worker lock and rerun U34; distinguish a returned
404 from an orphaned lease and from a 500. Move purge's classification back into its first
snapshot and rerun each of U35's four paths. Keep sequential success/release controls in
both experiments so that “nothing happened” cannot count as a fix. Lengthen heartbeat's
interval beyond the HTTP timeout: the terminal job still looks cancelled, but remote work
can continue until its own timeout. Observe that difference before choosing an oracle.

Read Chapter 8.1–8.4 through the approved-plan references above, the original
[state model](../exploration/models/states.md#e07-revision--build-explore-itlesson-06-e07-notes--new-observed-cell),
[notes](../exploration/experiments/07-state-it/notes.md), and
[reconciliation](../register/reconciliations/07.md). Forward questions remain heartbeat
shortening, error consistency, sweep observability, and unprobed cancel/retry/promote/
retention interleavings. E06 HTTP number decoding, broad error work, performance bounds,
E08 and E09 are outside this fix. Independent review is pending.

## Deviations

- **Provider interruption and seat change.** The first fixer (Codex Astra, under a Codex
  Sol coordinator) ended with exit 1 at 18:46:54Z on the provider's `possible cybersecurity
  risk` error ([run record](../register/runs/07-fixer-initial-run.jsonl)). A second Astra
  repair run started at 18:48:06Z. It passed a current verify and reached REF before Twin
  stopped it for the handoff under Nick's rule: when that filter fires, Codex stops and Opus
  takes over, with no retry around the filter. No terminal provider event was retained for
  this second run, so whether the filter fired on it is unknown. `07-ref-repair.log` keeps its verdict
  lines but has no exit record, so it is not counted as acceptance. Claude Opus 5.5 then
  reviewed the uncommitted work as another author's diff, finished it and committed it.
  Its run is the `07-opus-*` acceptance set above. The fixer seat therefore moved from
  the seeder family (Codex) to Claude for this lesson. The independent review should
  come from a family other than Claude.
- **REF fixture correction.** The first REF attempt (`07-ref-acceptance.log`) stopped
  with a PROBE ERROR. The probe asserted that no job id appears twice among lease
  responses. In the retained lesson-06 diagnostics, job 502 was leased by worker 11,
  released by that worker's deletion, then leased again by the same deleted worker as
  delivery 2 ([evidence](../register/runs/07-opus-fixture-evidence.txt)). That is a U34
  zombie on a second delivery, not two holders. The probe now groups responses by job,
  requires contiguous attempt numbers with distinct tokens, and requires the stored lease
  to equal the last delivery and the earlier release to be attributed. It also requires
  old-delivery tokens to be rejected without changing the current lease. It snapshots raw
  responses and rows before asserting. Only the fixture changed; no oracle was weakened,
  and the SUT is untouched by this correction. Neither later REF run happened to sample a
  second delivery; their PRESENT verdicts rest on first-delivery zombies and 500s.
- **Opus review changes to the inherited diff.** Cancel on a terminal job returns the
  previous `state conflict` message again, instead of leaking `sql: no rows in result set`.
  Purge locks only pending and leased rows, not the queue's terminal history. `lab/verify.sh`
  lost a coordinator-specific `/tmp` scratch path and a lesson-07-only `-v` flag. Hashes
  were re-frozen after these changes; the probe scripts and checks are byte-identical to
  the repair freeze.
- **Sample worker heartbeat interval** dropped from 5 s to 1 s so that cancellation is
  observed promptly. That is five times as many heartbeat requests per running job.
- **State-table size:** 49 driven cells plus 13 named extras, not the plan's illustrative 44.
- **Dwell:** paired lesson-06/lesson-07 only; no full sweep.
