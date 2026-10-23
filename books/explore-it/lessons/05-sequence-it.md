# 05 — Sequence it: verbs in orders nobody intended

Every earlier experiment used the verbs roughly in the order the README lists them: create a queue,
enqueue, lease, ack or nack. This one used them in orders nobody planned for. First came a handful of
pairs and triples chosen because they are unlikely: purge and then enqueue with a delay; pause, jump
the clock an hour, resume; lease, purge, nack; cancel, then retry. Then a random walk that picked a noun
and a verb, acted, and checked the rules after every step. Most of the sequences held. Three broke,
and all three were exits from the `leased` state that the code handled differently from each other.

The source is Hendrickson, *Explore It!* (2013), Chapter 6: 6.1 "Nouns and Verbs" and 6.2 "Random
Navigation". These citations follow the approved
[chapter map](../PLAN.md#22-chapter-map-kept-sections-only); the builder did not check them against
the book text.

## What the blind experiment learned

[The original record](../exploration/experiments/05-sequence-it/revealed.md) is unchanged from the
experiment commit. The explorer stated 22 hypotheses, confirmed 18 and refuted 2. One refuted
hypothesis was the charter's prediction, "every verb either does what it says from any starting
point or is refused". The other was its own prediction of which rules the walk would break. Both were
code-reading predictions that the running system contradicted. It filed 3 reports and filled 16 model
cells in about 24 minutes of a 60-minute budget, including 950 random-walk steps over five seeds.

The walker is [`walk.sh`](../exploration/experiments/05-sequence-it/walk.sh). It does what the
charter asks: pick a verb, act, observe, then check thirteen invariants built from the E04 rules. Seed
11 printed, among others:

```
VIOLATION step=128 inv=I4 job=21 grant 3 > max 2 (attempt=2)
VIOLATION step=146 inv=I4 job=24 grant 3 > max 2 (attempt=2)
VIOLATION step=146 inv=I8 job=24 attempt 2 <= previous 2
```

The violations were not the lesson; the reduction was. The explorer traced every I4/I8 hit through
the event history and found a `DELETE /v1/workers/{id}` between the two leases every time, and never
without one. It then reduced the sequence to four repetitions of three verbs (E05-F01 below). This
is 10.4's "reduce to the shortest form", applied to a random walk's output.

| Classification | Outcome |
|---|---|
| Expected here, found early | D12 purge leaves delayed and retrying jobs, D13 resume fires every missed tick, D14 retry from cancelled or succeeded: all found and fixed in E01. E05 re-ran each as a sequence (P1; P2 and P7 with a one-hour clock jump; P4), and all three held |
| Spec silence, resolved here | A2, a job that is leased when its queue is purged: the explorer ran the chartered sequence (P3), measured it correctly and judged it consistent with SPEC, so it filed no report. Decided below |
| Found as new in E05 | U24–U26 (below) |
| Missed / disclosed | none: no planted defect is left for this experiment |

**A2 was probed but not judged a defect, and that judgement was fair.** The SPEC says purge "removes
all pending jobs". A leased job is not pending, so its surviving the purge breaks no written rule. The
explorer wrote that down instead of calling it a bug. What the SPEC does not say is what happens when
that job fails afterwards: at lesson-04 a nack sends it to `retry_wait`, which puts work back into a
queue the operator has just emptied. That is a question for the owner, and the fix commit answers it.
The exported lesson 01 note names A2. The explorer did not open that note: the only lesson note it
read was lesson 04's.

## The three faults, with pointers into `explore-it/lesson-04`

| ID | Report | Where it lived | What happened | Rule |
|---|---|---|---|---|
| U24 | [E05-F01](../exploration/defects/E05-F01-worker-delete-release-uncounted-attempt.md) | `internal/api/jobs.go:126` re-queues a deleted worker's leases with `attempts` unchanged and no attempt row; `jobs.go:77` reports `attempt = attempts+1` | `max_attempts: 2`, lease then delete the worker four times: 4 deliveries, each `attempt: 1`, history `[]` | R3/R32 no job executes more than `max_attempts` |
| U25 | [E05-F02](../exploration/defects/E05-F02-terminal-rows-keep-lease-fields.md) | `ack.go:10` and `operator.go:6` change the state but leave the lease columns | a `succeeded` or `cancelled` job still shows a token, a holder and an unexpired `lease_expires_at` | consistency: every other exit from `leased` clears them |
| U26 | [E05-F03](../exploration/defects/E05-F03-worker-release-logged-as-lease-expired.md) | `migrations/003.sql:21` labels an exit from `leased` `lease_expired` whenever `last_error = 'lease expired'`, and `last_error` persists | after one expiry, the job's later ack, cancel or worker release is logged as another expiry | an event names the transition that happened |

The report for E05-F03 cites `migrations/002.sql` for the trigger. The trigger is in `003.sql`; the
raw report is left as written.

**One cause behind all three.** The explorer's causes model
([C25](../exploration/models/causes.md)) says it: *each exit from `leased` decides on its own which
columns to reset*. There are five exits: ack, nack, expiry, worker release and cancel. Each was
written separately, and each made its own choice about `attempts`, the attempt row, the three lease
columns and `last_error`. The new
[state model](../exploration/models/states.md) puts the five side by side, and the gaps show up in
that table before they show up at runtime. This is the pattern this book keeps returning to: the same
rule (here, "what leaving `leased` means") enforced in several places by several hands.

## The fix

Commit `queuelet: fix experiment 05 findings; resolve A2; add checks (lesson 05)`. The diff to
`project/` outside `lab/` and `test/` is 88 lines added and 29 removed, SPEC and README included.

- **U24.** Expiry and worker deletion now share one release path, `release` in
  `internal/jobs/expire.go`. Each release counts an attempt, writes an attempt row (`expired` /
  `lease expired`, or `released` / `worker deleted`) and makes the job `dead` once its attempts run
  out. `deleteWorker` calls `Jobs.ReleaseWorker` inside its transaction and notifies dead jobs after
  the commit.
- **U25.** Ack and cancel set `lease_token`, `leased_by` and `lease_expires_at` to NULL. Ack used to
  read the holder for its attempt row back from the job, so it now takes it from the pre-update row
  (`UPDATE jobs j … FROM jobs o … RETURNING o.leased_by`).
- **U26.** The trigger no longer reads error text. The expiry sweep marks its own transaction with
  `SET LOCAL queuelet.release='lease_expired'`, and the trigger labels a release `lease_expired`
  only when it sees that mark. This also covers the explorer's unprobed prediction H-E05-20: a nack
  whose error text happens to be `lease expired`.
- **A2.** `migrations/004.sql` adds `jobs.purge_requested`. Purge sets it on the queue's leased jobs
  in the same statement that deletes the pending ones. Nack and `release` make a marked job `dead`.
  Ack still completes it, and an operator retry clears the mark.

Spec decisions (the commit body and [reconciliation 05](../register/reconciliations/05.md)):

1. A release by worker deregistration records an unsuccessful attempt, as an expiry does (Q-E05-1).
2. Only a `leased` job has a lease token, holder and expiry. Every exit clears all three, and the
   attempt rows keep the worker that ran them (Q-E05-2).
3. An event is labelled with the job's new state, except a release by lease expiry, which is
   `lease_expired`. Earlier error text never changes the label.
4. A job leased when its queue is purged keeps its lease. An ack completes it; a nack, an expiry or a
   worker deletion makes it `dead`, never pending again.

**Left open.** E05-F02 also noticed that a delivery cancelled while in flight leaves no attempt row.
That belongs to a larger question the SPEC does not answer: what cancelling a job that is already
running should mean for the worker running it. This fix does not decide that question.

## The walk as a check

`lab/repro/walk05.py` is the charter turned into a regression check. It takes 100 random verbs on a
frozen clock: enqueue, enqueue with a delay, new worker, delete a worker, lease, ack, nack, nack
without retry, heartbeat, cancel, retry, purge, pause, resume and advance the clock. After every step
it asserts:

| Invariant | From |
|---|---|
| W1 no job is leased more than `max_attempts` times since its last retry | R32, E05-F01 |
| W2 a job has a lease token, holder and expiry exactly when it is `leased` | E05-F02 |
| W3 every `lease_expired` event matches an `expired` attempt | E05-F03 |
| W4 a job leased when its queue was purged never becomes pending again | A2 |
| W5 a cancelled job is never leased again | R37 |
| W6 attempt `n` is 1..k, never reused | R38 |
| W7 a verb the walker's model allows is not refused | the charter's prediction |

While a lease is held, the walk favours holder verbs. The explorer found that a uniform walk rarely
acks (0–4 acks per seed), and the first version of this check had the same weakness.

Measured on 2026-10-22 (`register/runs/05-walk-*.log`):

- Fixed tree: `WALK HOLDS: 100/100 steps, W1–W7 after each` for seeds 5–9.
- lesson-04 binaries (a scratch stack): every seed 5–9 breaks, W2 on an ack in four seeds and W3 at
  step 82 in seed 5. With W2 and W4 masked, W3 breaks in three of five seeds. With W2 masked
  under the earlier weights, W4 breaks in four of six.
- **W1 never fired at lesson-04 in any measured seed.** It needs three leases of one job with a worker
  deletion between two of them, and 100 random steps rarely produce that. U24 is held by its own
  reproduction and `TestL05U24`; the walk does not guard it.

## Verify

From `books/explore-it/project`, sequentially:

```sh
make up
make verify LESSON=05
make verify LESSON=05 REF=explore-it/lesson-04
make checks
make test
```

`lesson-05.ids` lists U24, U25, U26 and A2. Each wrapper runs `lab/repro/probe05.py` through the
public API, the lab clock and the event history. `TestL05U24` … `TestL05A2` run the same oracles, and
`TestL05Walk` runs the walk with seed 5.

Measured on 2026-10-22 (logs under `register/runs/05-*`):

| Command | Exit | Elapsed | Result |
|---|---:|---:|---|
| `make up` | 0 | 40 s | current stack built and healthy |
| `make verify LESSON=05 REF=explore-it/lesson-04` | 0 | 60 s | `DEFECTS REPRODUCED: 4 present` |
| `make verify LESSON=05` | 0 | 115 s | `FIXED AND CHECKED: 4 absent; regression checks pass` |
| `go test -tags checks ./test/checks -count=1 -v` (`make checks` plus `-v`) | 0 | 335 s | all 66 integration checks pass (61 earlier, 5 new incl. `TestL05Walk`) |
| `make test` | 0 | <1 s | unit suite passes |

Paired results, the same oracle on both binaries:

| Fault | lesson-04 | lesson-05 |
|---|---|---|
| U24 | 4 deliveries on `max_attempts: 2`, attempt `[1, 1, 1, 1]`, history `[]`, still `queued` | 2 deliveries, attempt `[1, 2]`, history `released, released`, `dead` |
| U25 | `succeeded` and `cancelled` rows keep token, holder and expiry | none kept |
| U26 | after one expiry: ack logged `lease_expired`, cancel logged `lease_expired` | `succeeded`, `cancelled` |
| A2 | lease → purge → nack leaves `retry_wait` | `dead` |

The seeder's answer-key control for A2 now prints ABSENT. The other two open silences still print
PRESENT (`05-unresolved-controls.log`).

**Dwell.** Measured only as the pair above (`register/runs/05-dwell.tsv`): all four present at
lesson-04 and absent at lesson-05. No `make dwell` sweep was run, so this note does not claim when each
fault first appeared. Earlier experiments came close. E01 saw a worker deletion release a job and
wrote down whether a release counts an attempt, but as a reading question; it never drove the release
twice. E04 checked "attempts never exceed max" against nack, and E02 had checked it against expiry.
Those were the two exits the earlier models listed. It took a walker that deletes workers at random
to try a third.

## Try changing one variable

- Heartbeat a purge-marked job: the lease extends as usual, because the mark only changes what a
  failure does.
- Purge twice: once while a job is leased, then again after its nack has made it `dead`. The second
  purge deletes nothing, because dead is terminal.
- Ack a purge-marked job: it succeeds. Then retry a purge-marked job that went `dead`: it runs again.
- Run `python3 lab/repro/walk05.py W05 11` with other seeds, and change the `HOLDER` weights to see how
  quickly a uniform walk stops reaching acks.
- Set `max_attempts: 1` and delete the worker once: the job goes straight to `dead`, notified like an
  expiry.

For further reading: Chapter 6, the
[probe-sequences-and-states skill](../skills/probe-sequences-and-states/SKILL.md), the
[probe notes](../exploration/experiments/05-sequence-it/notes.md) (including the dead ends: the
`set -u` crash on the first smoke run, the seed that starved with no workers and a paused queue, and the
`pkill -f walk.sh` that killed the explorer's own shell), and the
[reconciliation](../register/reconciliations/05.md).
