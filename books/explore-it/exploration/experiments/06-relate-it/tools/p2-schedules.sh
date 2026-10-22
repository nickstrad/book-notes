#!/usr/bin/env bash
# P2 (H-E06-03, H-E06-04): schedule delete by missing/foreign id; schedule payload shapes.
source "$(dirname "$0")/env.sh"
date -u +%T
echo '# DELETE missing schedule id 99999'; k acme-key DELETE /v1/schedules/99999
echo '# GET / PATCH / run missing id for comparison'; k acme-key GET /v1/schedules/99999; k acme-key PATCH /v1/schedules/99999 '{"enabled":false}'; k acme-key POST /v1/schedules/99999/run
echo '# globex DELETE acme schedule 1 (seeded daily-report)'; k globex-key DELETE /v1/schedules/1; sql "select id,namespace_id,name from schedules where id=1"
echo '# repeated DELETE of an already-deleted id 3 (deleted in P1)'; k acme-key DELETE /v1/schedules/3
k acme-key POST /v1/queues '{"name":"e6-pay"}' >/dev/null
for P in 'omitted' '"payload":[1,2]' '"payload":"x"' '"payload":null' '"payload":7'; do
  B='{"name":"p","queue":"e6-pay","kind":"echo","cron":"* * * * *"'; [ "$P" != omitted ] && B="$B,$P"; B="$B}"
  echo "# create schedule with $P"; R=$(k acme-key POST /v1/schedules "$B"); echo "$R" | cut -c1-200
done
sql "select id,payload,enabled,next_run_at from schedules where queue_id=(select id from queues where name='e6-pay') order by id"
