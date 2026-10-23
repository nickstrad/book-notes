#!/usr/bin/env bash
# E08 P3: stop postgres, restart api while DB is down, start postgres; does api come back? (H-E08-05)
set -u
cd "$(dirname "$0")/../../../../project"
. ../exploration/experiments/08-surround-it/tools/env.sh
C=(docker compose -f deploy/compose.yml)
ts() { date -u +%T.%N | cut -c1-10; }
echo "$(ts) stop postgres"; "${C[@]}" stop postgres 2>&1 | tail -1
echo "$(ts) restart api"; "${C[@]}" restart api 2>&1 | tail -1
sleep 3
echo "$(ts) ps:"; "${C[@]}" ps -a --format '{{.Service}} {{.State}} {{.Status}}'
docker inspect queuelet-e08-api-1 --format 'api exit={{.State.ExitCode}} restarts={{.RestartCount}} policy={{.HostConfig.RestartPolicy.Name}}'
echo "$(ts) api log tail:"; "${C[@]}" logs --since 10s api 2>&1 | tail -2
echo "$(ts) start postgres"; "${C[@]}" start postgres 2>&1 | tail -1
for i in $(seq 1 ${WAIT:-30}); do
  s=$(curl -s -o /dev/null -w '%{http_code}' "$QUEUELET_URL/health")
  [ "$s" = 200 ] && { echo "$(ts) health 200 after ${i}s"; break; }
  sleep 1
done
echo "$(ts) last health=$s"
"${C[@]}" ps -a --format '{{.Service}} {{.State}} {{.Status}}'
echo "worker-a log tail:"; "${C[@]}" logs --since 40s worker-a 2>&1 | tail -2
