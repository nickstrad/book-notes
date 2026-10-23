# E08-F04 — An unresponsive (frozen) database hangs every request, including /health, with no deadline; no 503 is ever returned

Seen at: export of `explore-it/lesson-07` (stack `queuelet-e08`), experiment 08 (surround-it).
Original observations: [E08 notes](../experiments/08-surround-it/notes.md) P8 (19:34:38Z, H-E08-20) and P8b
(19:35:12Z, H-E08-21).

## Minimal reproduction

Actor: lab operator (docker) plus any client. Clock: real. Fixture: any.

```sh
cd books/explore-it/exploration/experiments/08-surround-it/tools
./p8b_pg_pause_long.sh     # docker compose pause postgres for 25 s; GET /health at +1 s and +10 s, acme GET /v1/queues at +5 s
```

`docker compose pause` freezes the PostgreSQL processes while their TCP sockets stay open — the shape of a hung
database, a lock pile-up or a network path that drops packets without resets.

## Expected (named oracle)

- SPEC.md:188 "Health reports database availability" — `/health` should answer promptly and report degraded while the
  database does not answer. The code states the intended bound itself: a 1 s `PingContext` deadline
  (`internal/api/api.go:131-133`).
- SPEC.md:189 "an unavailable authentication store returns 503 rather than invalid credentials" — authenticated
  requests should fail with 503 `auth_unavailable` rather than wait indefinitely (approximation: within a few seconds).

## Observed

```text
P8  (12 s freeze) health → 200 after 11.79s; acme GET /v1/queues → 200 after 11.80s;
                  globex GET /v1/jobs/1 → 404 after 11.80s; bad key → 401 after 11.79s
P8b (25 s freeze) health+1 → 200 after 24.32s; auth+5 → 200 after 20.33s; health+10 → 200 after 15.33s
```

Every request on the database path, including `/health`, blocks for exactly the remaining freeze and then succeeds.
During the freeze nothing reports a problem.

## Variables isolated

- Matters: the database being *connected but unresponsive*. A stopped/crashed database (connection refused) gives a
  prompt 503 and `/health` 503 (P2, P3b), so the outage classification works only for refusing databases.
- Does not matter: endpoint or key (health, valid key, wrong key, other tenant all hang the same).
- Does not matter: freeze length (12 s and 25 s; latency = remaining freeze).
- Mechanism: auth `QueryRow` has no context (`internal/api/api.go:153`); store queries use no context anywhere
  (`internal/store/store.go`); the HTTP server has no write timeout (`cmd/api/main.go:41`). Why the 1 s ping deadline
  does not bound `/health` is an **inference**: lib/pq honours cancellation by sending a cancel request and still
  waiting for the server's reply; the driver source is outside the export and was not read.

## Sibling check

- Background loop and shipper use the same context-free store calls (`internal/api/api.go:205-245`): predicted to stall
  for the freeze and resume — consistent with health/requests resuming at unpause; loop stall not separately measured.
- Pool limit `SetMaxOpenConns(16)` (`internal/store/store.go:16`): predicted that more than 16 concurrent hung requests
  queue for a connection as well; not probed.
- Workers: their HTTP client has its own timeout (not read in this run); they would see slow responses rather than 503.

## Severity and status

Medium–high: a hung database turns into a hung API whose health endpoint cannot say so; load balancers and monitors
see timeouts instead of a degraded answer, and requests and goroutines accumulate for the whole incident.

Status: open.
