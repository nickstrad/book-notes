# U24 — releasing a lease by deleting the worker does not count an attempt

Original blind report: [E05-F01](E05-F01-worker-delete-release-uncounted-attempt.md).
Seen at `explore-it/lesson-04`; E05 notes P6. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity medium–high.
Current status: fixed in lesson 05.

Worker deletion now ends its leases through the same release path as lease expiry (`internal/jobs/expire.go`, `release`): each counts an attempt, writes a `released` attempt row with error `worker deleted`, and makes the job `dead` once its attempts are spent. Dead jobs are notified after the commit.

Spec decision: a release by worker deregistration counts as an attempt, like an expiry (Q-E05-1).

`lab/repro/U24.sh` (`probe05.py`) leases and deletes the worker four times on `max_attempts: 2`;
`TestL05U24` shares that oracle. See [reconciliation 05](../../register/reconciliations/05.md).
