#!/usr/bin/env bash
# E05 minimal probes H-E05-15 (event label), H-E05-16 (ack keeps lease fields), H-E05-17 (nack retry=false clears)
source "$(dirname "$0")/e05env.sh"
source <(sed -n '/^run()/,/^lease1()/p' "$(dirname "$0")/pairs.sh")
id_of() { sed 's/ \[.*//' | jq -r "$1"; }
echo "== H-E05-15 $(ts)"
run mkq f3 '"lease_seconds":2,"max_attempts":5'
J=$(api POST /v1/queues/f3/jobs -d '{"kind":"k"}' | id_of .id); echo job $J
W=$(wk f3a); run lease1 f3 $W
run adm POST /admin/clock -d '{"action":"advance","seconds":3}'; sleep 1.6
run jget $J
W=$(wk f3b); run lease1 f3 $W
run adm POST /admin/clock -d '{"action":"advance","seconds":0}'
run api DELETE /v1/workers/$W
run jget $J
api GET /v1/jobs/$J/events | sed 's/ \[.*//' | jq -c '.events[] | {id,at,event,worker,attempt,error}'
echo "== H-E05-16 $(ts)"
run mkq f2
J=$(api POST /v1/queues/f2/jobs -d '{"kind":"k"}' | id_of .id); echo job $J
W=$(wk f2a); T=$(lease1 f2 $W | id_of '.jobs[0].lease_token')
run api POST /v1/jobs/$J/ack -d "{\"lease_token\":\"$T\",\"result\":{\"ok\":1}}"
api GET /v1/jobs/$J | sed 's/ \[.*//' | jq -c '{state,leased_by,lease_token,lease_expires_at,finished_at}'
echo "== H-E05-17 $(ts)"
J=$(api POST /v1/queues/f2/jobs -d '{"kind":"k"}' | id_of .id); echo job $J
T=$(lease1 f2 $W | id_of '.jobs[0].lease_token')
run api POST /v1/jobs/$J/nack -d "{\"lease_token\":\"$T\",\"error\":\"no\",\"retry\":false}"
api GET /v1/jobs/$J | sed 's/ \[.*//' | jq -c '{state,leased_by,lease_token,lease_expires_at,finished_at}'
