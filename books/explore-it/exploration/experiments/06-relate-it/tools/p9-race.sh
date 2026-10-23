#!/usr/bin/env bash
# P9 (H-E06-10): concurrent enqueue vs DELETE queue; counts outcomes and checks orphans.
source "$(dirname "$0")/env.sh"
out=$PROJ/results/e06/p9-race; rm -rf $out; mkdir -p $out
date -u +%T
for r in $(seq 1 ${ROUNDS:-30}); do
  curl -s -o /dev/null -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -d '{"name":"e6-race"}' $QUEUELET_URL/v1/queues
  for i in $(seq 1 16); do curl -s -o $out/r$r-e$i.body -w '%{http_code}\n' -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -d '{"kind":"x"}' $QUEUELET_URL/v1/queues/e6-race/jobs > $out/r$r-e$i.code & done
  curl -s -o $out/r$r-del.body -w '%{http_code}\n' -X DELETE -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e6-race > $out/r$r-del.code &
  wait
  # tidy for the next round: cancel+delete any jobs, then the queue
  for j in $(sql "select j.id from jobs j join queues q on q.id=j.queue_id where q.name='e6-race' and q.namespace_id=1" | tail -n +2); do curl -s -o /dev/null -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/jobs/$j/cancel; curl -s -o /dev/null -X DELETE -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/jobs/$j; done
  curl -s -o /dev/null -X DELETE -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e6-race
done
date -u +%T
echo '# delete status counts'; cat $out/r*-del.code | sort | uniq -c
echo '# enqueue status counts'; cat $out/r*-e*.code | sort | uniq -c
echo '# non-201/404 enqueue bodies'; for f in $out/r*-e*.code; do c=$(cat $f); [ "$c" != 201 ] && [ "$c" != 404 ] && cat ${f%.code}.body; done | sort | uniq -c | head
echo '# rounds where delete=204 and some enqueue=201'; for r in $(seq 1 ${ROUNDS:-30}); do [ "$(cat $out/r$r-del.code)" = 204 ] && grep -l 201 $out/r$r-e*.code >/dev/null 2>&1 && echo "round $r"; done
sql "select (select count(*) from jobs j where not exists(select 1 from queues q where q.id=j.queue_id)) orphan_jobs,(select count(*) from schedules s where not exists(select 1 from queues q where q.id=s.queue_id)) orphan_schedules"
