#!/usr/bin/env bash
# E08 P3b: recover api; control = restart api with DB up; sibling = SIGKILL api / postgres with no manual action.
set -u
cd "$(dirname "$0")/../../../../project"
. ../exploration/experiments/08-surround-it/tools/env.sh
C=(docker compose -f deploy/compose.yml)
ts() { date -u +%T.%N | cut -c1-10; }
health() { curl -s -o /dev/null -w '%{http_code}' "$QUEUELET_URL/health"; }
waith() { for i in $(seq 1 $1); do [ "$(health)" = 200 ] && { echo "$(ts) health 200 after ${i}s"; return; }; sleep 1; done; echo "$(ts) health still $(health) after $1 s"; }
echo "$(ts) recover: start api"; "${C[@]}" start api 2>&1 | tail -1; waith 20
echo "$(ts) control: restart api with postgres up"; "${C[@]}" restart api 2>&1 | tail -1; waith 20
echo "$(ts) sibling 1: SIGKILL api (simulated crash)"; docker kill -s KILL queuelet-e08-api-1 >/dev/null; waith 15
docker inspect queuelet-e08-api-1 --format 'api state={{.State.Status}} exit={{.State.ExitCode}} restarts={{.RestartCount}}'
echo "$(ts) recover: start api"; "${C[@]}" start api 2>&1 | tail -1; waith 20
echo "$(ts) sibling 2: SIGKILL postgres (simulated crash)"; docker kill -s KILL queuelet-e08-postgres-1 >/dev/null; waith 15
docker inspect queuelet-e08-postgres-1 --format 'postgres state={{.State.Status}} exit={{.State.ExitCode}} restarts={{.RestartCount}} policy={{.HostConfig.RestartPolicy.Name}}'
curl -s -H 'Authorization: Bearer acme-key' "$QUEUELET_URL/v1/queues" | head -c 150; echo
echo "$(ts) recover: start postgres"; "${C[@]}" start postgres 2>&1 | tail -1; waith 20
"${C[@]}" ps -a --format '{{.Service}} {{.State}} {{.Status}}'
