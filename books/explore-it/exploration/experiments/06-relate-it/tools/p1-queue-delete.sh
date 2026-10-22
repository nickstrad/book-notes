#!/usr/bin/env bash
# P1 (H-E06-01, H-E06-C): queue deletion with each kind of dependent.
source "$(dirname "$0")/env.sh"
date -u +%T
echo '# empty queue → 204'; k acme-key POST /v1/queues '{"name":"e6-empty"}' >/dev/null; k acme-key DELETE /v1/queues/e6-empty
echo '# missing queue → 404?'; k acme-key DELETE /v1/queues/e6-missing
echo '# globex deletes acme-only name e6-own'; k acme-key POST /v1/queues '{"name":"e6-own"}' >/dev/null; k globex-key DELETE /v1/queues/e6-own; k acme-key GET /v1/queues/e6-own | cut -c1-80
echo '# queue with one queued job'; k acme-key POST /v1/queues '{"name":"e6-q1"}' >/dev/null; k acme-key POST /v1/queues/e6-q1/jobs '{"kind":"echo","payload":{}}'; k acme-key DELETE /v1/queues/e6-q1
echo '# cancel it (terminal) then delete queue'; J=$(sql "select id from jobs where queue_id=(select id from queues where name='e6-q1')" | tail -1); k acme-key POST /v1/jobs/$J/cancel; k acme-key DELETE /v1/queues/e6-q1
echo '# delete the terminal job, then queue'; k acme-key DELETE /v1/jobs/$J; k acme-key DELETE /v1/queues/e6-q1
echo '# queue with only a schedule'; k acme-key POST /v1/queues '{"name":"e6-s1"}' >/dev/null; S=$(k acme-key POST /v1/schedules '{"name":"s","queue":"e6-s1","kind":"echo","payload":{},"cron":"0 0 1 1 *"}' | jq -r .id 2>/dev/null); echo "schedule $S"; k acme-key DELETE /v1/queues/e6-s1
echo '# delete schedule then queue'; k acme-key DELETE /v1/schedules/$S; k acme-key DELETE /v1/queues/e6-s1
echo '# psql orphan check'; sql "select (select count(*) from jobs j where not exists(select 1 from queues q where q.id=j.queue_id)) orphan_jobs,(select count(*) from schedules s where not exists(select 1 from queues q where q.id=s.queue_id)) orphan_schedules,(select count(*) from attempts a where not exists(select 1 from jobs j where j.id=a.job_id)) orphan_attempts"
echo "# events for job $J"; k acme-key GET /v1/jobs/$J/events | jq -c '.events[]|[.event,.queue]' 2>/dev/null || k acme-key GET /v1/jobs/$J/events
