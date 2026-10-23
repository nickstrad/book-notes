#!/usr/bin/env bash
# P4 (H-E06-05/06/07/13): follow one datum enqueue → storage → lease → worker → result → webhook.
source "$(dirname "$0")/env.sh"
date -u +%T
curl -s -X DELETE $QUEUELET_HOOK_URL/admin/records >/dev/null
P='{"big":9007199254740993,"dec":1.10,"exp":1e2,"html":"<&>","z":{"b":1,"a":2}}'
echo "# A: echo via worker-b, payload $P"
A=$(k acme-key POST /v1/queues/reports/jobs "{\"kind\":\"echo\",\"priority\":9,\"payload\":$P}" | head -1 | jq -r .id); echo "job $A"
echo '# B: duplicate key payload via worker-b'; B=$(k acme-key POST /v1/queues/reports/jobs '{"kind":"echo","priority":9,"payload":{"d":1,"d":2}}' | head -1 | jq -r .id); echo "job $B"
echo '# C: manual kind, my worker, ack with exact big result'
W=$(k acme-key POST /v1/workers '{"name":"e6-manual","kinds":["manual"]}' | head -1 | jq -r .worker_id); echo "worker $W"
C=$(k acme-key POST /v1/queues/reports/jobs "{\"kind\":\"manual\",\"payload\":$P}" | head -1 | jq -r .id); echo "job $C"
L=$(curl -s -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -d "{\"worker_id\":$W,\"max_jobs\":1}" $QUEUELET_URL/v1/queues/reports/lease); echo "lease raw: $L"
T=$(echo "$L" | jq -r '.jobs[0].lease_token')
k acme-key POST /v1/jobs/$C/ack "{\"lease_token\":\"$T\",\"result\":{\"big\":9007199254740993,\"dec\":1.10}}"
sleep 6
for j in $A $B $C; do echo "# GET job $j (raw bytes)"; curl -s -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/jobs/$j | jq -c '{state,payload,result,last_error}'; curl -s -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/jobs/$j | grep -o '"payload":{[^}]*}[^,]*' | head -1; done
echo '# psql'; sql "select id,payload::text,result::text from jobs where id in ($A,$B,$C) order by id"
echo '# hooksink raw bodies'; curl -s $QUEUELET_HOOK_URL/admin/records | jq -r '.records[]|"\(.signature_valid) \(.status) \(.raw_body)"'
echo '# outbox'; sql "select id,job_id,attempts,status,delivered_at is not null delivered,body::text from webhook_outbox where job_id in ($A,$B,$C) order by id"
