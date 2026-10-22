#!/usr/bin/env bash
# P9c (H-E06-16/17): schedule create and batch enqueue racing DELETE queue (DELETE first, requests staggered 0-9.9 ms).
source "$(dirname "$0")/env.sh"
out=$PROJ/results/e06/p9c-race; rm -rf $out; mkdir -p $out
date -u +%T
for r in $(seq 1 ${ROUNDS:-40}); do
  curl -s -o /dev/null -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -d '{"name":"e6-race2"}' $QUEUELET_URL/v1/queues
  curl -s -o $out/r$r-del.body -w "%{http_code}\n" -X DELETE -H "Authorization: Bearer acme-key" $QUEUELET_URL/v1/queues/e6-race2 > $out/r$r-del.code &
  for i in $(seq 1 8); do (sleep 0.00$((i%10))$((i*3%10)); curl -s -o $out/r$r-s$i.body -w '%{http_code}\n' -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -d '{"name":"x","queue":"e6-race2","kind":"x","payload":{},"cron":"0 0 1 1 *"}' $QUEUELET_URL/v1/schedules > $out/r$r-s$i.code) & done
  for i in $(seq 1 8); do (sleep 0.00$((i%10))$((i*7%10)); curl -s -o $out/r$r-b$i.body -w '%{http_code}\n' -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -d '{"jobs":[{"kind":"x"},{"kind":"y"}]}' $QUEUELET_URL/v1/queues/e6-race2/jobs:batch > $out/r$r-b$i.code) & done
  wait
  for s in $(sql "select s.id from schedules s join queues q on q.id=s.queue_id where q.name='e6-race2' and q.namespace_id=1" | tail -n +2); do curl -s -o /dev/null -X DELETE -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/schedules/$s; done
  for j in $(sql "select j.id from jobs j join queues q on q.id=j.queue_id where q.name='e6-race2' and q.namespace_id=1" | tail -n +2); do curl -s -o /dev/null -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/jobs/$j/cancel; curl -s -o /dev/null -X DELETE -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/jobs/$j; done
  curl -s -o /dev/null -X DELETE -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e6-race2
done
date -u +%T
echo '# delete'; cat $out/r*-del.code | sort | uniq -c
echo '# schedule create'; cat $out/r*-s*.code | sort | uniq -c
echo '# batch'; cat $out/r*-b*.code | sort | uniq -c
echo '# unusual bodies'; for f in $out/r*-[sb]*.code; do c=$(cat $f); [ "$c" != 201 ] && [ "$c" != 404 ] && { echo -n "$c "; cat ${f%.code}.body; }; done | sort | uniq -c | head
sql "select (select count(*) from jobs j where not exists(select 1 from queues q where q.id=j.queue_id)) orphan_jobs,(select count(*) from schedules s where not exists(select 1 from queues q where q.id=s.queue_id)) orphan_schedules"
