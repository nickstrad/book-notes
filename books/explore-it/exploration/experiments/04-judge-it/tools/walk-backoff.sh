# walk-backoff.sh QUEUE WORKER N — enqueue one job, then N times: lease, nack, print wait (run_at - failure),
# set the lab clock to run_at and wait for one Promote tick. Uses the frozen lab clock (admin key).
source /tmp/explore-it-experiment-04/books/explore-it/exploration/experiments/04-judge-it/tools/e04env.sh
T=$(dirname "$0")
api POST /v1/queues/$1/jobs -d '{"kind":"echo","payload":{"walk":1}}' | tr -d '\n'; echo
for k in $(seq 1 $3); do
  read id a tok < <(bash $T/lease1.sh $1 $2)
  [ "$id" = NONE ] && { echo "attempt $k: no job leasable"; break; }
  api POST /v1/jobs/$id/nack -d "{\"lease_token\":\"$tok\",\"error\":\"walk $k\"}" >/dev/null
  row=$(sql "select state||' '||attempts||' '||extract(epoch from run_at-event_at)::bigint||' '||to_char(run_at,'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"') from jobs where id=$id" | sed -n 2p)
  set -- "$1" "$2" "$3"; read st at w ra <<<"$row"
  echo "lease attempt $a → $st attempts=$at wait_s=$w ($(python3 -I -c "print(round($w/86400,2))") d) run_at=$ra"
  [ "$st" = dead ] && break
  adm POST /admin/clock -d "{\"action\":\"set\",\"at\":\"$ra\"}" >/dev/null; sleep 1.3
done
