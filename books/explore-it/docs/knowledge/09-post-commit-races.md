# E09: post-commit races, frozen clocks and checks that test a table

Lessons from building lesson 09's probes and checks (`project/lab/repro/probe09.py`,
`test/checks/lesson09_test.go`). Evidence: `register/runs/09-opus-*`, including the
retained `attempt1` logs.

- **One delete spinner per job finds a post-commit tail fast.** A gap between a commit and
  a follow-up statement shows only when a second client acts on the same row in that gap.
  Start a thread per job that repeats `DELETE` until it stops answering 409, then issue the
  transition. With one spinner on the lowest id, the expiry race hit 2 times in 40 trials.
  With a spinner on every job of an eight-job pass, it hit on the first trial, because each
  later job's window includes every earlier notification.
- **Judge a race by what committed, not by the answer.** Read `job_events` (read-only) for
  the transition and `webhook_outbox` for its notification. A 409 for a committed ack and a
  200 with no outbox row are both defects, and only the database shows the second.
- **A far-future frozen clock dates everything the probe writes.** Outbox rows written at a
  frozen 2033 are due in 2033. After the clock resets they wait for years, and U03's
  whole-outbox drain control times out in `make checks`. Drain your own rows while the
  clock is still frozen, then reset it (E08's rule, met again through the clock).
- **Test the handlers against the table, and the table against the contract.** The model
  walk asks the code's own table (`jobs.Targets`) for each prediction, so it proves that
  the handlers agree with the table. A second, hand-written test pins the table to SPEC, so
  editing both together still fails. Prove the walk discriminates with a mutation: a
  `Retry` hard-coded to accept `cancelled` failed it at step 241.
- **zsh does not word-split a command held in a variable.** `R="python3 runner.py"; $R up`
  fails with "no such file or directory". Use a function: `R() { python3 runner.py "$@"; }`.
