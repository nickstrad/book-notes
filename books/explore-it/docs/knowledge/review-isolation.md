# Isolated review stacks need an explicit direct-check database URL

Compose's `QUEUELET_PG_PORT` only changes the published PostgreSQL port. The direct
Go integration checks in `test/checks/lesson02_direct_test.go` use
`QUEUELET_DATABASE_URL`, defaulting to localhost:15433. Setting the Compose port,
API URL and hooksink URL does not redirect those checks.

For a review stack publishing PostgreSQL on 15493, export:

```sh
export QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15493/queuelet?sslmode=disable'
```

Keep this aligned with `QUEUELET_PG_PORT`, alongside the separate Compose project,
image tag and API/hook addresses. Otherwise public repros can pass while direct
checks connect to another stack or fail to connect. `make verify LESSON=02` selects
those direct checks, and `make checks` includes them too.

Checkpoint B's first HEAD E02 attempt demonstrated this configuration error; the
corrected run and full 74-check suite passed. Original failure, retry, exact runner
environment and cleanup evidence: [checkpoint B review](../../register/reviews/B/review-b.md).
The project README now documents the requirement. No runtime change was needed.
