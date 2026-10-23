# U10 — malformed destination stops independent notifications

Original blind report: [E02-F02](E02-F02-unparsable-webhook-url-blocks-all-notifications.md).
Seen at `explore-it/lesson-01`; E02 notes H11–H11d. Raw report is unchanged.
Classification: unplanned, live black-box observation; severity high.
Current status: fixed in lesson 02.

The original report supplies steps, SPEC invalid-command and delivery oracles,
cross-namespace evidence, URL-class isolation, and sibling checks. Create and
PATCH now reject malformed, relative and non-HTTP(S) destinations with 400.
An empty destination disables future snapshots. The shipper counts and reschedules
request-construction failures per row, then attempts subsequent independent rows.

`lab/repro/U10.sh` uses public malformed destinations and an independent healthy
notification. It never writes fixtures through SQL. `TestL02U10` shares that
oracle; `TestL02LegacyMalformedOutboxIsolation` separately seeds a persisted bad
row in a private PostgreSQL schema and asserts its failed-attempt/error/retry
state and the following healthy delivery. The latter is a direct integration
check, not historical black-box evidence.

Outbox destination/body/secret snapshots intentionally survive queue configuration
changes and job deletion. Those operations are not orphan repairs. E02-F03 is a
spec clarification: delivered means a 2xx received before the sender's deadline;
timeouts may follow successful receiver side effects and retries may duplicate.
No new retry cap, timeout configuration or dead-letter policy is decided here.
