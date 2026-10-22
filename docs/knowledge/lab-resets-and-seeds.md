# Lab resets, seeds and fixtures

Learned building and reviewing Release It! (`books/release-it/project/lab/lib.sh`;
`books/release-it/REVIEW.md` H1 and M4).

- **Snapshot the seed.** Seeding 200 tenants and a 5 M-row ledger takes over a minute. After the
  first seed, `reset_stack` saves a Postgres template database (`CREATE DATABASE seed_<key>
  TEMPLATE litefleet`) and a copy of the tenant files in a volume, keyed by a hash of the seed
  arguments, the schema and the seed generator. Later resets restore it in seconds.
- **Postgres's first start.** On a fresh volume the image's initdb runs a temporary server, so a
  healthcheck can pass before the real server listens. Retry the first `psql` for up to 30 s.
- **Select fixtures by property, not by name.** Lesson 04's verify picked "the `analytics`
  database on node-b". The seed places databases by count parity, so when lesson 06 added three
  databases for one tenant, every `analytics` database moved to node-a and the selector came back
  empty. The script ran with `-focus null` and died without a verdict. Select by the property the
  experiment needs (`node == "node-b"`), and fail loudly when a selector is empty.
- **The harness comes from the working tree.** `make verify REF=<tag>` builds only the services
  from the tag, so a harness change made later can break the before-run of an earlier lesson.
  Re-run earlier lessons' verifies after changing the seed or the lab.
