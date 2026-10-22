#!/usr/bin/env bash
# E05 H-E05-18: event label for ack / cancel after an earlier lease expiry
source "$(dirname "$0")/e05env.sh"
source <(sed -n '/^run()/,/^lease1()/p' "$(dirname "$0")/pairs.sh")
id_of() { sed 's/ \[.*//' | jq -r "$1"; }
echo "== H-E05-18 $(ts)"
run mkq f3s '"lease_seconds":2,"max_attempts":5'
A=$(api POST /v1/queues/f3s/jobs -d '{"kind":"k"}' | id_of .id); B=$(api POST /v1/queues/f3s/jobs -d '{"kind":"k"}' | id_of .id); echo "jobs $A $B"
W=$(wk f3s1); run api POST /v1/queues/f3s/lease -d "{\"worker_id\":$W,\"max_jobs\":2}"
run adm POST /admin/clock -d '{"action":"advance","seconds":3}'; sleep 1.6
run sql "select id,state,attempts,last_error from jobs where id in ($A,$B)"
L=$(api POST /v1/queues/f3s/lease -d "{\"worker_id\":$W,\"max_jobs\":2}"); echo "$L"
TA=$(echo "$L" | sed 's/ \[.*//' | jq -r ".jobs[] | select(.id==$A) | .lease_token")
run api POST /v1/jobs/$A/ack -d "{\"lease_token\":\"$TA\",\"result\":{\"ok\":true}}"
run api POST /v1/jobs/$B/cancel
for j in $A $B; do api GET /v1/jobs/$j | sed 's/ \[.*//' | jq -c '{id,state,last_error}'; api GET /v1/jobs/$j/events | sed 's/ \[.*//' | jq -c '.events[] | select(.outbox_id==null) | {id,event,worker,attempt}' | tail -2; done
