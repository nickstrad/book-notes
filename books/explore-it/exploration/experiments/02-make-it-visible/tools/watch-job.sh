#!/usr/bin/env bash
# usage: watch-job.sh JOB_ID SECONDS OUTFILE — samples API GET, DB jobs/attempts/workers row and worker restart count once per second.
source "$(dirname "$0")/e02env.sh"
id=$1; n=$2; out=$3; : > "$out"
for ((i=0;i<n;i++)); do
  {
    echo "=== $(ts)"
    echo "API: $(api GET /v1/jobs/$id | jq -c '{state,attempts,leased_by,lease_expires_at,last_error}')"
    echo "DB:  $(sql "select state,attempts,leased_by,lease_expires_at from jobs where id=$id" | sed -n 2p)"
    echo "ATT: $(sql "select n||':'||worker_id||':'||outcome from attempts where job_id=$id order by n" | sed -n '2,$p' | grep -v rows | tr '\n' ' ')"
    echo "WRK: $(sql "select string_agg(id::text,',') from workers where name='worker-a'" | sed -n 2p) restarts=$(docker inspect -f '{{.RestartCount}} {{.State.Status}}' queuelet-e02-worker-a-1)"
  } >> "$out"
  sleep 1
done
