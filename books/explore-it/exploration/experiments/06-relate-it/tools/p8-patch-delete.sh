#!/usr/bin/env bash
# P8 (H-E06-09a-d, H-E06-02): owner (queue) changes while dependents are in flight.
source "$(dirname "$0")/env.sh"
date -u +%T
mode() { curl -s -X POST -H 'Content-Type: application/json' -d "{\"mode\":\"$1\"}" $QUEUELET_HOOK_URL/admin/mode; }
id() { head -1 | jq -r .id; }
echo '## a) max_attempts'
k acme-key POST /v1/queues '{"name":"e6-patch"}' >/dev/null
A=$(k acme-key POST /v1/queues/e6-patch/jobs '{"kind":"x"}' | id)
k acme-key PATCH /v1/queues/e6-patch '{"max_attempts":5}' | cut -c1-120
B=$(k acme-key POST /v1/queues/e6-patch/jobs '{"kind":"x"}' | id)
S=$(k acme-key POST /v1/schedules '{"name":"r","queue":"e6-patch","kind":"x","payload":{},"cron":"0 0 1 1 *"}' | id); R=$(k acme-key POST /v1/schedules/$S/run | id)
sql "select id,max_attempts from jobs where id in ($A,$B,$R) order by id"
echo '## b) webhook snapshot'
curl -s -X DELETE $QUEUELET_HOOK_URL/admin/records >/dev/null; mode 500
k acme-key POST /v1/queues '{"name":"e6-hook","webhook":{"url":"http://hooksink:8201/hook","secret":"report-secret"}}' >/dev/null
C=$(k acme-key POST /v1/queues/e6-hook/jobs '{"kind":"x","payload":{"tag":"C"}}' | id); k acme-key POST /v1/jobs/$C/cancel >/dev/null
sleep 2; sql "select id,job_id,url,secret,attempts,status,error from webhook_outbox where job_id=$C"
k acme-key PATCH /v1/queues/e6-hook '{"webhook":{"url":"http://hooksink:8201/other","secret":"changed"}}' | grep -o '"webhook_secret":"[^"]*","webhook_url":"[^"]*"'
D=$(k acme-key POST /v1/queues/e6-hook/jobs '{"kind":"x","payload":{"tag":"D"}}' | id); k acme-key POST /v1/jobs/$D/cancel >/dev/null
echo '## c) disable webhook'
k acme-key PATCH /v1/queues/e6-hook '{"webhook":{"url":""}}' | grep -o '"webhook_secret":"[^"]*","webhook_url":"[^"]*"'
E=$(k acme-key POST /v1/queues/e6-hook/jobs '{"kind":"x","payload":{"tag":"E"}}' | id); k acme-key POST /v1/jobs/$E/cancel >/dev/null
echo '## e) H-E06-02 delete job and queue with pending row'
k acme-key POST /v1/queues '{"name":"e6-del","webhook":{"url":"http://hooksink:8201/hook","secret":"report-secret"}}' >/dev/null
F=$(k acme-key POST /v1/queues/e6-del/jobs '{"kind":"x","payload":{"tag":"F"}}' | id); k acme-key POST /v1/jobs/$F/cancel >/dev/null
k acme-key DELETE /v1/jobs/$F; k acme-key DELETE /v1/queues/e6-del
sql "select id,job_id,queue,url,secret,attempts,status,delivered_at is not null dlv from webhook_outbox where job_id in ($C,$D,$E,$F) order by id"
echo "C=$C D=$D E=$E F=$F"; date -u +%T; mode ok; echo; sleep 8; date -u +%T
sql "select id,job_id,url,attempts,status,delivered_at is not null dlv from webhook_outbox where job_id in ($C,$D,$E,$F) order by id"
curl -s $QUEUELET_HOOK_URL/admin/records | jq -r '.records[]|select(.status==200)|"\(.signature_valid) \(.raw_body)"' | grep -E "\"job_id\":($C|$D|$E|$F)[,}]"
curl -s $QUEUELET_HOOK_URL/admin/records | jq -r '.records|length'
echo '## d) retention 0'
sql "select count(*) jobs from jobs where queue_id=(select id from queues where name='e6-hook')"
k acme-key PATCH /v1/queues/e6-hook '{"retention_hours":0}' | grep -o '"retention_hours":[0-9.]*'
sleep 3; date -u +%T
sql "select count(*) jobs from jobs where queue_id=(select id from queues where name='e6-hook')"
sql "select job_id,event from job_events where job_id in ($C,$D,$E) and event='retained'"
sql "select count(*) outbox_rows from webhook_outbox where job_id in ($C,$D,$E)"
for j in $C $D; do k acme-key GET /v1/jobs/$j; done
k acme-key DELETE /v1/queues/e6-hook
