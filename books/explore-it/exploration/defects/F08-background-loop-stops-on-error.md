# F08 — Background loop stops permanently after any step error (DB restart, bad schedule)

- Seen at: `explore-it/v0-inherited` (blind export); isolated stack `queuelet-e01`, image `queuelet:e01` built from the export 2026-10-22T01:39–01:40Z. Fixture: `lab/reset.sh` seed at 01:41:03Z.
- Original observations: [notes.md](../experiments/01-map-it/notes.md) H10 (01:48:56Z–01:49:11Z), H10b (01:49:18Z–01:49:42Z), H26 (01:49:46Z–01:50:24Z).

## Minimal reproduction
Trigger 1 — database restart (lab container action on the isolated project):
1. `docker compose -f deploy/compose.yml restart postgres` (ready after ~2 s).
2. Enqueue `{"kind":"echo","delay_seconds":1}` (returns 201 id 896); wait 10 s.
Trigger 2 — a schedule whose cron cannot be parsed (see F09 for how it gets in): PATCH a due `* * * * *` schedule to `"cron":"bogus"`, wait for the minute, enqueue a delay_seconds=1 job, wait 8 s.

## Expected (oracle)
SPEC: "The background loop promotes scheduled jobs, expires leases, fires cron schedules, sweeps retention and delivers webhooks. It should recover after a database restart." + "The service promotes due jobs to queued."

## Observed
```
T1: 896|scheduled (10 s after due); api log: "background task: dial tcp 172.22.0.2:5432: connect: connection refused" and nothing after
T2: api log 01:49:00 "background task: expected exactly 5 fields, found 1: [bogus]"; 892|scheduled|due=t 8 s later
    API request path keeps serving (GET queue 200, /health ok)
Restart api: first tick promotes 892, then Fire inserts one more job and logs the same error again (01:49:21); h23 job count 5→6
Recovery only after DELETE /v1/schedules/4 + api restart (control: delay job 895 queued within 4 s, no errors)
```

## Variables isolated
- Cause is shared: api.go:102-104 `return`s from Loop on any step error. Two independent triggers reproduce it.
- Step order matters (api.go:101: Promote, Expire, Fire, Retain, Ship): a Fire error still lets the first post-restart tick promote/expire once — this refuted my H10b prediction.
- Poisoned schedule inserts a duplicate job each API start before failing (fire.go:44 insert precedes Next at :48).
- /health stays 200 throughout (no liveness signal; see F14 sibling).

## Sibling check
Every Loop step returns DB errors (expire.go, fire.go, ship.go:19,36,45). webhook.Ship returns on `http.NewRequest` error (ship.go:27-28) — a queue webhook URL that fails request construction (e.g. containing a control character) would also stop the loop: predicted, not probed. Retain with a NULL/odd retention — not probed.

## Severity / status
Critical: after a routine DB restart nothing is promoted, expired, fired, retained or notified until a manual API restart, while health reports ok. Status: open.
