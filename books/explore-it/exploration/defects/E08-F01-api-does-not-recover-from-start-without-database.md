# E08-F01 — An API start while PostgreSQL is unreachable exits for good: the service stays down after the database returns

Seen at: export of `explore-it/lesson-07` (stack `queuelet-e08`), experiment 08 (surround-it).
Original observations: [E08 notes](../experiments/08-surround-it/notes.md) P3 (19:23:13Z, H-E08-05) and P3b
(19:24:03Z).

## Minimal reproduction

Actor: lab operator (docker compose). Clock: real. Fixture: any (seeded lab).

```sh
cd books/explore-it/exploration/experiments/08-surround-it/tools
./p3_api_restart_db_down.sh     # stop postgres → restart api → start postgres → poll /health 30 s
```

Steps: `docker compose stop postgres`; `docker compose restart api`; `docker compose start postgres`; poll
`GET /health`.

## Expected (named oracle)

SPEC.md:186-187 "The background loop … should recover after a database restart"; SPEC.md:189 "an unavailable
authentication store returns 503". Together they say a database outage is a recoverable, reportable condition,
not a fatal one. Comparable practice: the same compose file already gives the worker containers
`restart: on-failure` so they survive their neighbour's absence. Approximation: once PostgreSQL is healthy again, the
API answers `/health` 200 within a few seconds without an operator action (bound: 30 s).

## Observed

```text
19:23:13.6 stop postgres → Stopped
19:23:14.2 restart api → Started
19:23:18.2 api exited  Exited (1) 3 seconds ago   (exit=1 restarts=0 policy=no)
api log: "dial tcp: lookup postgres on 127.0.0.11:53: server misbehaving"   ← log.Fatal
19:23:18.8 start postgres → Started
19:23:50.0 last health=000; api exited (1) 35 s ago; postgres running (healthy) 31 s
worker-a: Post "http://api:8200/v1/queues/emails/lease": dial tcp: lookup api … (keeps polling)
```

The API never came back; every client (workers included) sees connection failures until someone runs
`docker compose start api`.

## Variables isolated

- Matters: the API *process start* happening while the DB is unreachable (`cmd/api/main.go:19-22`: `store.Open`
  pings and `log.Fatal`s; migrations would also Fatal). Control: `restart api` with the DB up → health 200 after 2 s
  (P3b 19:24:05).
- Matters for persistence: no restart policy on `api` in `deploy/compose.yml` (`policy=no`). `depends_on:
  service_healthy` only orders `compose up`; it does not apply to `restart`/`start` or a crash.
- Does not matter: how the DB is down (stopped here; DNS lookup failure in the log), worker activity.
- A running API survives a DB crash (P2, P3b SIGKILL postgres: 503 during, 200 within 2 s of a manual start).

## Sibling check (predictions preceded the P3b probe, logged post-hoc — see notes)

- API crash with the DB up: `docker kill -s KILL api` → exited 137, health 000 for 15 s, `restarts=0` (P3b
  19:24:07). Same missing supervision.
- PostgreSQL crash: `docker kill -s KILL postgres` → exited 137, `policy=no`, API 503 for 15 s until manual start
  (P3b 19:24:25). Same missing supervision for the datastore.
- hooksink: `compose.yml` gives it no restart policy either (read only; lab partner, not SUT behaviour).
- No retry loop around `store.Open` or the migration `Exec`s (`cmd/api/main.go:19-35`).

## Severity and status

Medium: after any host reboot, redeploy or crash that orders the API before its database (or a DB blip during an
API restart), the whole service stays down with no self-healing and no health endpoint answering, even though the
running process otherwise tolerates DB restarts. Lab deployment scope; real deployments may add their own supervisor.

Status: open.
