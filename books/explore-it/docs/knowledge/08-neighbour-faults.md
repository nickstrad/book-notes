# E08: injecting neighbour faults from outside the service

Lessons from building lesson 08's black-box probes (`project/lab/repro/probe08.py`).
Evidence: `register/runs/08-opus-*`, including the retained `attempt1`/`attempt2` logs.

- **`docker kill` is not a crash.** Docker treats `docker kill` (any signal) like an
  operator stop, so `restart: unless-stopped` does not bring the container back. To test a
  restart policy, kill the container's process from the host:
  `kill -9 $(docker inspect -f '{{.State.Pid}}' <container>)`. Docker sees that as an
  unexpected exit and restarts it (RestartCount goes to 1).
- **A frozen database is not a stopped one.** `docker compose pause postgres` keeps TCP
  sockets open while the server never answers, which is how a hung database or a dropped
  network path looks. A stopped database gives connection-refused errors and fast 503s.
  The two need separate probes.
- **lib/pq does not enforce context deadlines against an unresponsive server.** On cancel
  it sends a cancel request and keeps waiting for the server, so `PingContext` with a
  one-second context waited out a 12-second freeze. Bounding a call needs a goroutine and a
  `select` on a timer (`bounded` in `internal/api/api.go`). The mechanism is inferred from
  the behaviour; the driver source was not read.
- **Probes that use the whole outbox must leave it drained.** U03's control waits until
  no outbox row is undelivered. A probe that leaves rows to a removed destination makes
  every later run of U03 time out. Restore the partner and wait for your own rows.
- **A second partner on the lab network.** Run the target image's `hooksink` with
  `docker run --network <project>_default --network-alias <alias> -p 127.0.0.1::8201`.
  The API reaches it as `http://<alias>:8201`, and the probe sets its mode through the
  published port.
- **The checks suite exceeds Go's default 10-minute test timeout** from lesson 08 on;
  `make checks` passes `-timeout 30m`.
