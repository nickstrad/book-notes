#!/usr/bin/env bash
# P7 (H-E06-08): delete the row of a running sample worker (worker-a = id 2) while it holds a lease.
source "$(dirname "$0")/env.sh"
date -u +%T
RC0=$(docker inspect -f '{{.RestartCount}} {{.State.Status}}' queuelet-e06-worker-a-1); echo "worker-a before: $RC0"
J=$(k acme-key POST /v1/queues/emails/jobs '{"kind":"sleep","priority":9,"payload":{"seconds":15}}' | head -1 | jq -r .id); echo "job $J"
for i in $(seq 1 10); do S=$(sql "select state||' by '||coalesce(leased_by::text,'-') from jobs where id=$J" | tail -1); [ "${S%% *}" = leased ] && break; sleep 1; done; date -u +%T; echo "job $J: $S"
k acme-key DELETE /v1/workers/2; T0=$(date -u +%FT%TZ); echo "deleted at $T0"
sql "select id,state,attempts,leased_by,lease_token is null no_token from jobs where id=$J"
k acme-key GET /v1/jobs/$J/attempts | head -1 | jq -c '.attempts[]|[.n,.worker_id,.outcome,.error]'
sleep 25; date -u +%T
sql "select id,state,attempts,leased_by from jobs where id=$J"
echo "worker-a after: $(docker inspect -f '{{.RestartCount}} {{.State.Status}}' queuelet-e06-worker-a-1)"
dc logs --since "$T0" worker-a 2>&1 | cut -c1-200 | sort | uniq -c | sort -rn | head -8
sql "select id,name from workers order by id"
k acme-key GET /v1/queues/emails/stats
