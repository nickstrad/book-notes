# E07 revealed — state-it (build `explore-it/lesson-06`)

Status: **done**. The state table was built before probing; every "should be rejected" and "undefined" cell in it was
driven, starting with the cells where the spec and the code disagreed or where the spec was silent.

Elapsed: about 20 minutes of wall clock, from 17:59:17Z to about 18:18Z, out of a 60-minute budget.
- Reading: about 4 minutes (17:59–18:02:30Z: brief, protocol, skill, models, SPEC, guards in `internal/jobs` and
  `internal/api`, plus the Compose build in the background).
- Probing: about 10 minutes (18:03:30–18:15Z, P1–P7, with log entries written between probes).
- Writing: about 6 minutes (state table and hypotheses at 18:02:30–18:03:30Z; reports, models and this file at
  18:15–18:18Z).

Records: [notes.md](notes.md), with raw output under [evidence/project-results/e07/](evidence/project-results/e07/) and
replay helpers under [tools/](tools/) (`env.sh`, `q.py`, `p1`–`p7`).

## What the system showed

- **Single actor: the table has no missing arrows.** I drove 7 states × 7 verbs with a foreign worker or a wrong token
  (P7). All 49 cells matched the `WHERE` guards. No rejected verb changed the state, the attempt count, the lease
  fields or the attempt rows. The charter's deep sequence also held: lease with A, expire, lease with B, then A's ack,
  nack and heartbeat. Each A verb got 409, B completed, and the attempt history read n1 `expired`(A), n2
  `succeeded`(B) (P1). A stale first token is also refused when the same worker leases the job again.
- **Two actors overlapping: there are arrows the table lacks.** Each guard only re-checks a row whose snapshot version
  already matched it. Two overlaps break this:
  - **Worker DELETE racing that worker's lease** ([E07-F01](../../defects/E07-F01-worker-delete-racing-lease-leaves-zombie-lease.md)).
    Both calls succeed, and the job stays `leased` by a worker that no longer exists. The deleted worker can still
    heartbeat (extending the lease without bound) and ack. A second DELETE returns 404, so the operator cannot release
    the job. This happened in 16/20 one-job trials. In the other order the lease answers 500 `sql: no rows`.
  - **Purge racing any transition into or out of `leased`** ([E07-F02](../../defects/E07-F02-purge-misses-rows-changing-state-concurrently.md)).
    The transitions were lease, nack, worker release and the expiry sweep. In every case jobs survived the purge as
    `queued`/`retry_wait`, or as leased without the purge flag, and they will run. The lesson-05 rule "never pending
    again" holds only when nothing else is happening; one leased job is enough to break it (6/20 trials).
- **The "expired but not yet swept" gap.** After `lease_expires_at`, the holder is refused at once. Other workers get
  the job one background tick later (0.2–1.5 s on a quiet stack). During that gap nobody may run the job, which is
  consistent with SPEC:108-109 and :116.
- **The background loop is sequential.** When a single sweep released 1,110 expiries (left over from my own race
  runs), promotion of an unrelated due job was delayed by more than 3 s (S6). I did not count this as a defect: the
  spec gives no latency bound and the load was my own. It is recorded as H-E07-17 and T-E07-1.
- **Not defects** (reasons are in the notes): heartbeat can shorten a lease (S2 → Q-E07-1, a spec silence), and a
  nack 409 carries driver text as its message (S1 → Q-E07-2, #offcharter). Two other surprises were folded into
  E07-F01 and E07-F02 as part of their characterization: S3, the race rate being much higher than expected, and S5,
  the purge window outlasting what the concurrency alone explains.

## Model delta

- [states.md](../../models/states.md), E07 revision: a two-actor diagram with dashed arrows, and 41 newly observed
  cells: 18 sequence/race cells plus 23 matrix cells never observed before. Matrix cells that re-confirm earlier
  observations are marked as regressions and not counted.
- [rules.md](../../models/rules.md), E07 revision: R-E07-1 (worker deletion releases its leases, violated),
  R-E07-2 (purge under concurrency, violated), R-E07-3 (rejected verbs have no side effects, holds), R-E07-4 (one
  completion wins, holds).
- [causes.md](../../models/causes.md), E07 additions: C-E07-1 (set-based statements select by snapshot while rows
  move across their predicate) and C-E07-2 (lease reads the worker without a lock; holder verbs check only the token).
- [hypotheses.md](../../hypotheses.md): H-E07-01…17, Q-E07-1/2, T-E07-1/2.
- [defects/README.md](../../defects/README.md): an E07 section with the two reports.

## Unresolved uncertainties and next hypotheses

- **H-E07-16:** a lease on a worker deleted between the API's existence check and the lease transaction returns 500.
  This was seen once by accident and not driven on purpose; it is a sibling of the queue variant H-E06-20.
- **H-E07-17:** a large expiry batch stalls Promote and cron. This needs sweep durations, which T-E07-1 asks for.
- **Not driven (inferred to serialise via guard re-checks):** cancel racing lease/nack/release, retry racing purge,
  promote racing purge, and retention interleavings.
- **T-E07-2:** the races are probabilistic. A transaction pause hook in the lab would make E07-F01 and E07-F02
  deterministic for the fixer's repro scripts.
- **Mechanism not confirmed at runtime:** the explanation of both reports comes from reading the guards and fits
  every observed outcome, including the serial ones. A pause hook would let it be confirmed at runtime.

## Log discipline note

Three predictions (H-E07-08, -09, -14) were reasoned before their probes but written to the log afterwards. They are
kept and marked post-hoc, and they are excluded from the counts below. Several estimated timestamp labels were
corrected to the command stamps at 18:14Z; the correction is itself logged.

## Score

| Measure | Count | Evidence / counting basis |
|---|---|---|
| Hypotheses stated | 13 | H-E07-01, 02, 03, 04, 04b, 05, 06, 07, 10, 11, 12, 13, 15, each written in notes before its probe. Post-hoc H-E07-08, 09, 14 and the unprobed backlog items 16 and 17 are excluded |
| Hypotheses confirmed | 12 | 01, 02, 03, 04, 04b, 05, 06, 07, 10, 11, 12, 15 (P1–P7) |
| Hypotheses refuted | 1 | H-E07-13: the stated condition ">1.5 s" was observed in P6 r1 (3.04 s behind a 1,110-row sweep). The narrower claim, on a quiet stack, held in r2 |
| Code-reading predictions contradicted by the running system | 1 | H-E07-13. Its source was the 1 s loop tick in `api.go:221`, which did not account for one unbounded Expire transaction. All defect-predicting readings were confirmed |
| Defects found | 2 | E07-F01, E07-F02. No passing regression counted. S1, S2 and S6 are written up as not-a-defect |
| Model cells filled | 41 | states.md E07 ★ rows: 17 sequence/race rows, one of which (the zombie heartbeat/ack row) holds 2 cells, giving 18 cells, plus 23 matrix cells never observed before (scheduled 5, queued 3, leased 1, retry_wait 5, succeeded 3, dead 3, cancelled 3). Rows that re-confirm earlier observations and inferred arrows are excluded |
