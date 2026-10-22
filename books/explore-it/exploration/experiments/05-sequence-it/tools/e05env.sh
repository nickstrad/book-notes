# Adapted from experiments/04-judge-it/tools/e04env.sh (inherited copy unchanged).
# Source this file before every lab/client command in experiment 05.
export COMPOSE_PROJECT_NAME=queuelet-e05 QUEUELET_IMAGE_TAG=e05 QUEUELET_API_PORT=18260 QUEUELET_HOOK_PORT=18261 QUEUELET_PG_PORT=15493
export QUEUELET_URL=http://localhost:18260 QUEUELET_HOOK_URL=http://localhost:18261
export PROJ=/tmp/explore-it-experiment-05/books/explore-it/project
export RES=$PROJ/results/e05
dc() { docker compose -p queuelet-e05 -f "$PROJ/deploy/compose.yml" "$@"; }
# raw curl as acme: prints body then [HTTP status]
api() { local m=$1 p=$2; shift 2; curl -sS -X "$m" -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -w ' [%{http_code}]' "$QUEUELET_URL$p" "$@"; echo; }
adm() { local m=$1 p=$2; shift 2; curl -sS -X "$m" -H 'Authorization: Bearer admin-key' -H 'Content-Type: application/json' -w ' [%{http_code}]' "$QUEUELET_URL$p" "$@"; echo; }
sql() { PGPASSWORD=explorer psql -h localhost -p 15493 -U explorer -d queuelet -X -A -F' | ' -c "BEGIN READ ONLY; $1; COMMIT;" | grep -v -e '^BEGIN$' -e '^COMMIT$'; }
ts() { date -u +%Y-%m-%dT%H:%M:%S.%3NZ; }
