# U33 — E06 finding 7

Original immutable report: [E06-F07](E06-F07-attempts-outlive-their-job.md).

Status: **open**; assigned to the experiment 06 fix.

Coordinator decision: attempt records belong to their job and are removed with it on explicit deletion, purge and retention; job_events and pending webhook snapshots keep their documented independent lifetime.
