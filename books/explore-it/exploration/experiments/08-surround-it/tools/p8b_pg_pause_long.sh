#!/usr/bin/env bash
# E08 P8b: 25 s freeze; staggered /health and authenticated request latencies (H-E08-21)
set -u
cd "$(dirname "$0")/../../../../project"
. ../exploration/experiments/08-surround-it/tools/env.sh
C=(docker compose -f deploy/compose.yml)
ts() { date -u +%T.%N | cut -c1-10; }
req() { local t0=$(date +%s.%N); local s=$(curl -s -m 60 -o /dev/null -w '%{http_code}' "$2" ${3:+-H "$3"}); echo "$(ts) req $1 → $s after $(echo "$(date +%s.%N)-$t0" | bc | cut -c1-5)s"; }
echo "$(ts) pause postgres"; "${C[@]}" pause postgres 2>&1 | tail -1
sleep 1;  req health+1 "$QUEUELET_URL/health" &
sleep 4;  req auth+5 "$QUEUELET_URL/v1/queues" "Authorization: Bearer acme-key" &
sleep 5;  req health+10 "$QUEUELET_URL/health" &
sleep 15; echo "$(ts) unpause postgres"; "${C[@]}" unpause postgres 2>&1 | tail -1
wait
"${C[@]}" logs --since 40s api 2>&1 | grep -v '"job event"' | tail -4
