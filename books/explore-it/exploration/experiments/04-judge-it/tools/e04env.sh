# Adapted from experiments/03-vary-it/tools/e03env.sh (inherited copy unchanged).
# Source this file before every lab/client command in experiment 04.
export COMPOSE_PROJECT_NAME=queuelet-e04 QUEUELET_IMAGE_TAG=e04 QUEUELET_API_PORT=18250 QUEUELET_HOOK_PORT=18251 QUEUELET_PG_PORT=15483
export QUEUELET_URL=http://localhost:18250 QUEUELET_HOOK_URL=http://localhost:18251
export PROJ=/tmp/explore-it-experiment-04/books/explore-it/project
export RES=$PROJ/results/e04
dc() { docker compose -p queuelet-e04 -f "$PROJ/deploy/compose.yml" "$@"; }
# raw curl as acme: prints body then [HTTP status]
api() { local m=$1 p=$2; shift 2; curl -sS -X "$m" -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -w ' [%{http_code}]' "$QUEUELET_URL$p" "$@"; echo; }
adm() { local m=$1 p=$2; shift 2; curl -sS -X "$m" -H 'Authorization: Bearer admin-key' -H 'Content-Type: application/json' -w ' [%{http_code}]' "$QUEUELET_URL$p" "$@"; echo; }
sql() { PGPASSWORD=explorer psql -h localhost -p 15483 -U explorer -d queuelet -X -A -F' | ' -c "BEGIN READ ONLY; $1; COMMIT;" | grep -v -e '^BEGIN$' -e '^COMMIT$'; }
ts() { date -u +%Y-%m-%dT%H:%M:%S.%3NZ; }
