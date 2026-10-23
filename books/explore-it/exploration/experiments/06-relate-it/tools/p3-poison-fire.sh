#!/usr/bin/env bash
# P3 (H-E06-12): cron-fired non-object payload reaching the sample worker on emails.
source "$(dirname "$0")/env.sh"
date -u +%T
A=$(k acme-key POST /v1/schedules '{"name":"arr","queue":"emails","kind":"echo","payload":[1,2],"cron":"* * * * *"}' | head -1 | jq -r .id)
N=$(k acme-key POST /v1/schedules '{"name":"nul","queue":"emails","kind":"echo","cron":"* * * * *"}' | head -1 | jq -r .id)
echo "schedules arr=$A nul=$N"
until [ "$(date -u +%S)" -ge 2 ] && [ "$(date -u +%S)" -lt 30 ]; do sleep 1; done
k acme-key PATCH /v1/schedules/$A '{"enabled":false}' >/dev/null; k acme-key PATCH /v1/schedules/$N '{"enabled":false}' >/dev/null
date -u +%T; echo "disabled $A $N"
IDS=$(sql "select string_agg(id::text,',') from jobs where queue_id=1 and (jsonb_typeof(payload)<>'object')" | tail -1); echo "fired jobs: $IDS"
for t in 1 2 3 4 5 6; do sleep 8; date -u +%T; sql "select id,state,attempts,payload,result,last_error from jobs where id in ($IDS) order by id" | tail -n +2; done
for j in ${IDS//,/ }; do echo "# attempts $j"; k acme-key GET /v1/jobs/$j/attempts | head -1 | jq -c '.attempts[]|[.n,.worker_id,.outcome,.error]'; done
