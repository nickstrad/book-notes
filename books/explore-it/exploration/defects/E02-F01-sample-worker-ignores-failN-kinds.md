# E02-F01 — Documented `fail3` kind is never leased by the sample worker run with the documented kinds

- Seen at: `explore-it/lesson-01` plain export; isolated stack `queuelet-e02`, image `queuelet:e02` built 2026-10-22T03:21–03:23Z;
  fixture `lab/reset.sh` 03:23:12Z.
- Original observation: [E02 notes](../experiments/02-make-it-visible/notes.md) E02-H10 (stated 03:24Z, observed 03:26:10Z–03:27:30Z).

## Minimal reproduction
1. Stack with compose `worker-a` (default `WORKER_KINDS=echo,sleep,fail,crash,http`, `cmd/worker/main.go:24`) consuming `emails`.
2. `POST /v1/queues/emails/jobs {"kind":"fail3","payload":{},"priority":8}` (acme key) → `{"id":304,"state":"queued"}` at 03:26:10Z.
3. Wait ≥ 60 s; read `GET /v1/jobs/304`, `attempts`, worker stdout.

## Expected (oracle)
README "Worker mode": `qctl worker run imports --kinds echo,sleep,fail,crash,http` … "fail reads payload.n (or use kind fail3)".
The sample worker *implements* `fail<N>` (`internal/worker/worker.go:86-95`), so a reader of README expects a `fail3` job to be
run (and fail 3 times) by a worker started with the documented kind list. SPEC:73 "Only jobs with a kind supported by that
worker are eligible" — the worker supports the kind but does not advertise it.

## Observed
```text
03:27:30.969 jobs 304 | fail3 | queued | attempts 0 | last_error ''      (80 s after enqueue)
attempts for 304: none; worker-a stdout: no line for 304 (it processed job 303 kind=fail at the same time)
```
Lease eligibility is exact string match `kind=ANY($2)` (`internal/jobs/jobs.go:44`) against the registered kinds; `fail3` ≠ `fail`.
No console says why the job waits: no API field, no log, stats only show queued age growing.

## Variables isolated
- Same worker, same queue, same moment: kind `fail` with payload `{"n":5}` (job 303) ran 3 times and died — the worker and queue work.
- Priority irrelevant (304 at 8 below 303 at 9; 303 finished at 03:26:18, 304 still queued 70 s later).
- Not probed: `qctl worker run` path (same `worker.Run`, inference only); a worker registered with `--kinds fail3` (E01 H41
  ran `fail2` that way — consistent with exact matching).

## Sibling check
`fail<N>` is the only prefix-matched kind in `perform` (`worker.go:86`); `echo/sleep/crash/http` are exact. None found elsewhere.

## Severity / status
Low: sample-worker/README mismatch; a tester following the README gets a job that silently never runs and may misread it as
a lease defect. Testability: TR-E02-4 (why is a queued job not eligible). Status: open.
