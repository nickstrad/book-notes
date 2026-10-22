# E05-F02 — Ack and cancel leave a live-looking lease on terminal jobs; a cancelled in-flight delivery has no attempt record

Seen at: export of `explore-it/lesson-04` (stack `queuelet-e05`), experiment 05 (sequence-it).
Original observations: [E05 notes](../experiments/05-sequence-it/notes.md) P5 (13:22:54Z, cancel), walk seed 11 I7 on
succeeded rows (13:26:45Z), minimal ack probe H-E05-16, sibling H-E05-17.

## Minimal reproduction

Actor: acme key. Clock: real (P5) or frozen test clock (walk); no difference.

```sh
source exploration/experiments/05-sequence-it/tools/e05env.sh
api POST /v1/queues -d '{"name":"p5"}'
api POST /v1/queues/p5/jobs -d '{"kind":"k"}'                                      # job 7
W=$(api POST /v1/workers -d '{"name":"p5w","kinds":["k"]}' | sed 's/ \[.*//' | jq -r .worker_id)
api POST /v1/queues/p5/lease -d "{\"worker_id\":$W,\"max_jobs\":1}"
api POST /v1/jobs/7/cancel                                                         # 200
api GET /v1/jobs/7            # "state":"cancelled", "lease_token":"edbb…", "leased_by":5, "lease_expires_at": +30 s
api GET /v1/jobs/7/attempts   # {"attempts":[]}
```

Ack variant: lease → `ack` with the token → `GET /v1/jobs/{id}` shows `"state":"succeeded"` with the token, worker and a
future `lease_expires_at` (walk seed 11 job 22; minimal probe in notes H-E05-16). Replay: `tools/pairs.sh P5`, `tools/f02.sh`.

## Expected (named oracle)

Approximation, stated with its bounds: **internal consistency of the product's own transitions**. Every other exit from
`leased` clears the lease (`nack.go:60`, `expire.go:116`, worker release `internal/api/jobs.go:126` all set
`lease_token=NULL, leased_by=NULL, lease_expires_at=NULL`). SPEC defines a lease as something a job in `leased` has
("A lease ends at its `lease_expires_at`"; "Only the current lease holder may complete or extend a job"); a terminal
job has no current holder, so it should not display one. For the cancel case, SPEC "Attempt `n` numbers every execution
of the job" — the delivery that was in progress when the operator cancelled is an execution that started.

The SPEC does not list job fields, so this is a consistency oracle, not a spec line; the owner could decide that
`leased_by` on a succeeded job means "completed by". Even then a non-null `lease_token` and an unexpired
`lease_expires_at` on a terminal row are not meaningful.

## Observed

| Exit | state | leased_by | lease_token | lease_expires_at | attempts row for the delivery |
|---|---|---|---|---|---|
| cancel while leased (job 7) | cancelled | 5 | kept | 13:23:24.81 (30 s after cancel) | **none** (`[]`) |
| ack (job 22) | succeeded | 27 | kept | 14:28:44 (20 s after finish) | n1 succeeded |
| nack retry=false (H-E05-17) | dead | NULL | NULL | NULL | n1 failed |
| expiry (P7) | dead | NULL | NULL | NULL | n1 expired |
| worker delete (P6) | queued | NULL | NULL | NULL | none (E05-F01) |

After `DELETE /v1/workers/5` (H-E05-13) the cancelled job 7 still says `leased_by 5`, a worker that no longer exists.
Walk seeds 11 and 12 flagged I7 on succeeded rows (seed 13 had no successful ack of a current lease), and I7c appeared in seeds 12 and 13 when a leased job was cancelled.

No behavioural effect found: every verb guards on `state='leased'` (ack/nack/heartbeat 409 on these rows in the walks;
worker deletion and expiry filter on state). The effect is on what API readers, dashboards and the event console see.

## Variables isolated

- Matters: the exit transition (ack or cancel from `leased`). Nack, expiry and worker release clear the fields.
- Does not matter: clock mode; whether the worker still exists; queue settings.
- Cancel of a queued/scheduled/retry_wait job: no lease fields to keep (P8 rows 10/11 show NULLs).

## Sibling search

`operator.go:6` (Cancel) and `ack.go:25` (Ack) are the two UPDATEs leaving `leased` without lease columns; the other
writers listed above clear them. `operator.go:13` Retry (dead → queued) inherits NULLs from nack/expiry. Retain copies
`leased_by` into the `retained` event (`expire.go:141`), so the stale worker id propagates to history for both exits.

## Severity and status

Severity: **low**. Misleading reads (a terminal job appears to hold a live lease and names a possibly deleted worker; the
operator-cancelled execution is absent from the attempts history), no state-machine effect observed. Needs an owner
decision on field semantics.

Status: `open`.
