#!/usr/bin/env bash
# E05 pair/triple probes P1-P8. Usage: bash pairs.sh P1 [P2 ...]  (prints commands and outputs; tee into results)
source "$(dirname "$0")/e05env.sh"
run() { echo "\$ $*"; "$@"; }
jget() { sql "select id,state,attempts,max_attempts,run_at,created_at,finished_at,leased_by,lease_token is not null as tok,lease_expires_at,last_error from jobs where id=$1"; }
mkq() { api POST /v1/queues -d "{\"name\":\"$1\"${2:+,$2}}"; }
wk() { api POST /v1/workers -d "{\"name\":\"$1\",\"kinds\":[\"k\"]}" | sed 's/ \[.*//' | jq -r .worker_id; }
lease1() { api POST /v1/queues/$1/lease -d "{\"worker_id\":$2,\"max_jobs\":1}"; }
P1() { echo "== P1 purge then enqueue-with-delay (idempotency key) $(ts)"
  run mkq p1
  run api POST /v1/queues/p1/jobs -d '{"kind":"k","delay_seconds":60,"idempotency_key":"K1"}'
  run api POST /v1/queues/p1/purge
  run api POST /v1/queues/p1/jobs -d '{"kind":"k","delay_seconds":60,"idempotency_key":"K1"}'
  run sql "select id,state,idempotency_key from jobs where queue_id=(select id from queues where name='p1')"; }
P2() { echo "== P2 pause, advance clock 1h, resume $(ts)"
  run mkq p2
  W=$(wk p2w); echo "worker $W"
  run api POST /v1/queues/p2/jobs -d '{"kind":"k","delay_seconds":60}'
  run api POST /v1/schedules -d '{"name":"p2s","queue":"p2","kind":"k","payload":{},"cron":"* * * * *"}'
  run api POST /v1/queues/p2/pause
  run adm POST /admin/clock -d '{"action":"advance","seconds":3600}'
  sleep 3
  run sql "select id,state,run_at,created_at from jobs where queue_id=(select id from queues where name='p2') order by id"
  run sql "select id,next_run_at,last_run_at from schedules where name='p2s'"
  run lease1 p2 $W
  sleep 2
  run api POST /v1/queues/p2/resume
  sleep 3
  run sql "select id,state,run_at,created_at from jobs where queue_id=(select id from queues where name='p2') order by id"
  run api POST /v1/queues/p2/lease -d "{\"worker_id\":$W,\"max_jobs\":10}"
  run sql "select id,next_run_at,last_run_at from schedules where name='p2s'"; }
P3() { echo "== P3 lease, purge, nack $(ts)"
  run mkq p3
  W=$(wk p3w)
  J=$(api POST /v1/queues/p3/jobs -d '{"kind":"k"}' | sed 's/ \[.*//' | jq -r .id); echo "job $J"
  L=$(lease1 p3 $W); echo "$L"; T=$(echo "$L" | sed 's/ \[.*//' | jq -r '.jobs[0].lease_token')
  run api POST /v1/queues/p3/purge
  run api POST /v1/jobs/$J/nack -d "{\"lease_token\":\"$T\",\"error\":\"after purge\"}"
  run jget $J
  run api POST /v1/queues/p3/purge
  run jget $J; }
P4() { echo "== P4 cancel->retry, dead->cancel, succeeded->retry $(ts)"
  run mkq p4 '"max_attempts":1'
  W=$(wk p4w)
  A=$(api POST /v1/queues/p4/jobs -d '{"kind":"k"}' | sed 's/ \[.*//' | jq -r .id)
  run api POST /v1/jobs/$A/cancel
  run api POST /v1/jobs/$A/retry
  B=$(api POST /v1/queues/p4/jobs -d '{"kind":"k"}' | sed 's/ \[.*//' | jq -r .id)
  T=$(lease1 p4 $W | sed 's/ \[.*//' | jq -r '.jobs[0].lease_token')
  run api POST /v1/jobs/$B/nack -d "{\"lease_token\":\"$T\",\"error\":\"x\"}"
  run jget $B
  run api POST /v1/jobs/$B/cancel
  C=$(api POST /v1/queues/p4/jobs -d '{"kind":"k"}' | sed 's/ \[.*//' | jq -r .id)
  T=$(lease1 p4 $W | sed 's/ \[.*//' | jq -r '.jobs[0].lease_token')
  run api POST /v1/jobs/$C/ack -d "{\"lease_token\":\"$T\",\"result\":1}"
  run api POST /v1/jobs/$C/retry
  run api POST /v1/jobs/$C/cancel; }
P5() { echo "== P5 lease, cancel: lease fields $(ts)"
  run mkq p5
  W=$(wk p5w)
  J=$(api POST /v1/queues/p5/jobs -d '{"kind":"k"}' | sed 's/ \[.*//' | jq -r .id)
  run lease1 p5 $W
  run api POST /v1/jobs/$J/cancel
  run jget $J
  run api GET /v1/jobs/$J; }
P6() { echo "== P6 lease, delete worker, lease ... (max_attempts 2) $(ts)"
  run mkq p6 '"max_attempts":2'
  J=$(api POST /v1/queues/p6/jobs -d '{"kind":"k"}' | sed 's/ \[.*//' | jq -r .id); echo "job $J"
  for i in 1 2 3 4; do
    W=$(wk p6w$i)
    L=$(lease1 p6 $W); echo "lease by $W: $L"
    T=$(echo "$L" | sed 's/ \[.*//' | jq -r '.jobs[0].lease_token')
    run api DELETE /v1/workers/$W
    run api POST /v1/jobs/$J/heartbeat -d "{\"lease_token\":\"$T\",\"extend_seconds\":30}"
    run jget $J
  done
  run api GET /v1/jobs/$J/attempts
  run sql "select event,worker,attempt from job_events where job_id=$J order by id"; }
P7() { echo "== P7 lease (lease_seconds 2, max 1), pause, expiry while paused $(ts)"
  run mkq p7 '"lease_seconds":2,"max_attempts":1'
  W=$(wk p7w)
  J=$(api POST /v1/queues/p7/jobs -d '{"kind":"k"}' | sed 's/ \[.*//' | jq -r .id)
  run lease1 p7 $W
  run api POST /v1/queues/p7/pause
  sleep 4
  run jget $J
  run api GET /v1/jobs/$J/attempts; }
P8() { echo "== P8 cancel scheduled / cancel retry_wait, then time passes $(ts)"
  run mkq p8 '"backoff":{"kind":"linear","base_seconds":30}'
  W=$(wk p8w)
  A=$(api POST /v1/queues/p8/jobs -d '{"kind":"k","delay_seconds":30}' | sed 's/ \[.*//' | jq -r .id)
  B=$(api POST /v1/queues/p8/jobs -d '{"kind":"k"}' | sed 's/ \[.*//' | jq -r .id)
  T=$(lease1 p8 $W | sed 's/ \[.*//' | jq -r '.jobs[0].lease_token')
  run api POST /v1/jobs/$B/nack -d "{\"lease_token\":\"$T\",\"error\":\"x\"}"
  run api POST /v1/jobs/$A/cancel
  run api POST /v1/jobs/$B/cancel
  run adm POST /admin/clock -d '{"action":"advance","seconds":120}'
  sleep 3
  run jget $A; run jget $B
  run lease1 p8 $W; }
for p in "$@"; do $p; done
