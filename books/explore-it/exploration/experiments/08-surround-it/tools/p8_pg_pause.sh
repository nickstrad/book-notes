#!/usr/bin/env bash
# E08 P8: freeze postgres for 12 s; time authenticated requests and /health (H-E08-20)
set -u
cd "$(dirname "$0")/../../../../project"
. ../exploration/experiments/08-surround-it/tools/env.sh
C=(docker compose -f deploy/compose.yml)
ts() { date -u +%T.%N | cut -c1-10; }
req() { local t0=$(date +%s.%N); local s=$(curl -s -m 60 -o /tmp/explore-it-experiment-08/books/explore-it/project/results/e08/p8-body-$1 -w '%{http_code}' "$2" ${3:+-H "$3"}); echo "$(ts) req $1 → $s after $(echo "$(date +%s.%N)-$t0" | bc | cut -c1-5)s"; }
echo "$(ts) pause postgres"; "${C[@]}" pause postgres 2>&1 | tail -1
sleep 0.5
req auth1 "$QUEUELET_URL/v1/queues" "Authorization: Bearer acme-key" &
req auth2 "$QUEUELET_URL/v1/jobs/1" "Authorization: Bearer globex-key" &
req health "$QUEUELET_URL/health" &
req badkey "$QUEUELET_URL/v1/queues" "Authorization: Bearer nope" &
sleep 11.5
echo "$(ts) unpause postgres"; "${C[@]}" unpause postgres 2>&1 | tail -1
wait
req health-after "$QUEUELET_URL/health"
echo "auth1 body: $(head -c 80 results/e08/p8-body-auth1)"; echo "badkey body: $(cat results/e08/p8-body-badkey)"
