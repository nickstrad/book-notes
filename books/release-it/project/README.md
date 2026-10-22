# litefleet

A baby multitenant SQLite database platform. Tenants authenticate with an API key, create databases
and run SQL against them. Each database is one SQLite file in WAL mode on one of two node agents.
Three stateless gateways behind HAProxy authenticate the key, ask the control plane (a Postgres
registry) which node owns the database, forward the statement and return the rows. Nodes snapshot
databases to a blob store, send one usage event per statement to a metering partner, and heartbeat
to control. An ops dashboard polls a fleet-wide usage report.

This codebase is the worked example for *Release It!* (2nd ed.) in `books/release-it/`. The commit
history is the point: `release-it/v0-flawed` is the platform that passed QA, and each
`release-it/lesson-NN` tag fixes what the book says will go wrong. See `../lessons/`.

## Run it

Docker with Compose v2+ and GNU make. Host ports are all in the 1xxxx range.

```
make up        # build images, start the stack
make seed      # 200 tenants, 240 databases, a 5 M-row usage ledger (~1 min)
make verify LESSON=00
make migrate   # online migrations (migrations/online/*.sql); safe while serving, safe twice
make down      # remove everything, volumes included
```

| Port | What |
|---|---|
| 18080 | the tenant API (HAProxy) |
| 18404 | HAProxy stats (`/stats`) and metrics (`/metrics`) |
| 18001–18003 | gateway-a, gateway-b, gateway-c directly |
| 18090 | control |
| 18071, 18072 | node-a, node-b |
| 15432 | Postgres (`litefleet` / `litefleet`) |
| 18081, 18082 | blobstore, metering (partner fakes) |
| 18474 | Toxiproxy API |
| 19090 | Prometheus |

After `make seed`, `results/manifest.json` lists every tenant's API key and databases.

## The tenant API

All routes take `Authorization: Bearer <api-key>`. Errors are
`{"error": {"code", "message", "request_id"}}`.

| Route | Body | Answer |
|---|---|---|
| `POST /v1/databases` | `{"name": "app"}` | `201 {id, name, node, ...}` |
| `GET /v1/databases` | | `{"databases": [...]}` |
| `DELETE /v1/databases/{name}` | | `204` |
| `GET /v1/databases/{name}/stats` | | size, row and byte totals, last backup |
| `POST /v1/databases/{name}/query` | `{"sql": "SELECT ...", "args": [...]}` | `{"columns": [...], "rows": [[...]]}` |
| `POST /v1/databases/{name}/execute` | `{"statements": [{"sql", "args"}, ...]}` | `{"rows_affected", "last_insert_rowid"}` |
| `GET /v1/admin/usage?window=24h` | (admin key) | per-tenant statements, rows, bytes, storage |

`query` takes one read statement. `execute` runs its statements in one transaction.

A query's answer is capped at **1 000 rows or 1 MB**, whichever comes first, and says
`"truncated": true` when it was cut. The cap is the platform's bulkhead for its nodes' memory, not a
suggestion: to read a large table, page through it with a keyset query.

```
{"sql": "SELECT id, body FROM events WHERE id > ? ORDER BY id LIMIT 500", "args": [last_id_seen]}
```

```
KEY=$(jq -r '.tenants[10].key' results/manifest.json)
curl -s localhost:18080/v1/databases/default/query -H "Authorization: Bearer $KEY" \
  -d '{"sql": "SELECT id, body FROM items WHERE id > ? ORDER BY id LIMIT 5", "args": [100]}'
```

## Layout

```
cmd/gateway  cmd/control  cmd/node        the platform
cmd/blobstore  cmd/metering               partner fakes (honest protocol, realistic limits)
cmd/seedgen  cmd/loadgen                  lab tools
internal/api        gateway handlers        internal/routes      gateway → control client
internal/nodeclient gateway → node client   internal/registry    control's Postgres registry
internal/placement  where new databases go  internal/engine      node: handles, execution, admission
internal/backup     snapshots to blobstore  internal/blob        node → blobstore client
internal/metering   usage queue and client  internal/httpx       server, logging, JSON, errors
internal/wire       shared JSON shapes      migrations/          registry schema
deploy/   compose.yml, Dockerfiles, haproxy.cfg, prometheus.yml, toxiproxy.json
lab/      verify runner and scripts, report and audit SQL, calibration, toxics helper
```

## The lab

- `make verify LESSON=NN` resets the stack (a seeded snapshot is restored in seconds after the
  first seed), runs the lesson's load and chaos timeline, prints the per-window report and the
  Prometheus readings, and ends with a verdict line.
- `make verify LESSON=NN REF=<tag>` runs the same script with gateway, control and node (or the
  services in `REF_SERVICES`) built from another commit; the harness comes from the working tree.
- `make calibrate` measures one gateway's and one node's capacity into `lab/capacity.env`.
- `make audit` reads every tenant file with DuckDB (duplicate rows, WAL sizes, outboxes).
- `lab/toxics.sh` adds and removes Toxiproxy toxics by hand.
- `make test` runs the unit tests.
