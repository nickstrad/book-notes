# Nouns and verbs

Legend: **O** walked in E01 with evidence, **I** read only, **—** not tried.

| Noun | Verbs (API / qctl) | Walked | Sequences tried | Untested sequences |
|---|---|---|---|---|
| namespace | (implicit via key); stats | O H30 | acme→globex stats 403, admin→globex 200 | admin key mutations (Q4) |
| API key | bearer auth | O H32, H42 | valid / empty / DB down | key rotation (no verb) |
| queue | create, ls, get, patch, pause, resume, purge, rm | O all (H01, H14, H18, H24, H40) | pause→schedule→resume (F24); rm nonempty (F15); patch limits (F19) | patch webhook→pending outbox |
| job | enqueue, batch, ls, get, attempts, ack, nack, heartbeat, cancel, retry, rm | O all | lease→ack; lease→nack(×2); lease→expire→lease (F05); cancel(leased)→nack (F02); ack→retry (F06); ack→nack (F02) | heartbeat after expiry; cancel scheduled; rm then attempts |
| lease | lease (worker_id, max_jobs) | O | concurrent? no; unknown worker (F21); same-second FIFO (F23) | max_jobs 100 under contention |
| attempt | (read) attempts | O | ack/nack produce rows; expiry produces none (F05) | worker delete release |
| worker | create, rm, run | O H01, H22b, H40, H41 | rm releases lease | rm while acking |
| schedule | create, ls, get, patch, run, rm | O H09, H10, H40 | patch cron/tz/enabled (F07); bad cron (F09→F08); run on deleted queue (F15) | DST end-to-end (TR2) |
| webhook outbox entry | (implicit on terminal) | O H11-H13, H34 | ack/cancel/dead → POST; 500; timeout | change secret with entries pending |
| hooksink record | /admin/records, /admin/mode | O | ok/500/timeout | slow (3 s) |
| stats | queue / namespace | O H01, H17, H30 | | throughput window edge |

## E02 revision — `explore-it/lesson-01`; ★ = sequence newly walked ([E02 notes](../experiments/02-make-it-visible/notes.md))

| Noun | Sequence walked | Result |
|---|---|---|
| ★ job | enqueue crash → lease → exit → expire → lease … → dead (emails, reports) | dead at max; notification for reports |
| ★ job | enqueue sleep → lease → freeze worker → expire → stale ack → re-lease → ack | 409 then succeeded attempts 2 |
| ★ webhook outbox entry | ack → POST 500 ×11 → mode ok → delivered | delivered once on 2xx |
| ★ webhook outbox entry | bad-URL queue → ack → PATCH URL → DELETE job | stays, blocks all (E02-F02) |
| ★ worker | crash → container restart → re-register | row per incarnation, stale rows remain |
