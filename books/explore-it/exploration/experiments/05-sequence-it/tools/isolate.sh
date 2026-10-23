#!/usr/bin/env bash
# E05 isolation probes for E05-F01 / E05-F02. Clock is frozen (fixed) at this point; advances are explicit.
source "$(dirname "$0")/e05env.sh"
source <(sed -n '/^run()/,/^lease1()/p' "$(dirname "$0")/pairs.sh")
id_of() { sed 's/ \[.*//' | jq -r "$1"; }
echo "== H-E05-11 $(ts)"
run mkq i1 '"max_attempts":3,"backoff":{"kind":"linear","base_seconds":5}'
J=$(api POST /v1/queues/i1/jobs -d '{"kind":"k"}' | id_of .id); echo job $J
W=$(wk i1a); L=$(lease1 i1 $W); echo "lease by $W: $L"; T=$(echo "$L" | id_of '.jobs[0].lease_token')
run api POST /v1/jobs/$J/nack -d "{\"lease_token\":\"$T\",\"error\":\"fail1\"}"
run adm POST /admin/clock -d '{"action":"advance","seconds":10}'; sleep 2.5
run jget $J
W=$(wk i1b); echo "lease by $W: $(lease1 i1 $W)"
run api DELETE /v1/workers/$W
W=$(wk i1c); L=$(lease1 i1 $W); echo "lease by $W: $L"
run jget $J
run api GET /v1/jobs/$J/attempts
echo "== H-E05-12 $(ts)"
run mkq i2 '"max_attempts":1'
J=$(api POST /v1/queues/i2/jobs -d '{"kind":"k"}' | id_of .id); echo job $J
W=$(wk i2a); echo "lease by $W: $(lease1 i2 $W)"
run api DELETE /v1/workers/$W
W=$(wk i2b); L=$(lease1 i2 $W); echo "lease by $W: $L"; T=$(echo "$L" | id_of '.jobs[0].lease_token')
run api POST /v1/jobs/$J/nack -d "{\"lease_token\":\"$T\",\"error\":\"fail\"}"
run jget $J
run api GET /v1/jobs/$J/attempts
echo "== H-E05-13/14 $(ts)"
run api GET /v1/jobs/7/attempts
run api DELETE /v1/workers/5
run jget 7
