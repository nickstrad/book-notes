# Source this file before every lab/client command in experiment 02.
export COMPOSE_PROJECT_NAME=queuelet-e02 QUEUELET_IMAGE_TAG=e02 QUEUELET_API_PORT=18230 QUEUELET_HOOK_PORT=18231 QUEUELET_PG_PORT=15463
export QUEUELET_URL=http://localhost:18230 QUEUELET_HOOK_URL=http://localhost:18231
export PROJ=/tmp/explore-it-experiment-02/books/explore-it/project
export RES=$PROJ/results/e02
dc() { docker compose -p queuelet-e02 -f "$PROJ/deploy/compose.yml" "$@"; }
api() { local m=$1 p=$2; shift 2; curl -sS -X "$m" -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' "$QUEUELET_URL$p" "$@"; echo; }
sql() { PGPASSWORD=explorer psql -h localhost -p 15463 -U explorer -d queuelet -X -A -F' | ' -c "BEGIN READ ONLY; $1; COMMIT;" | grep -v -e '^BEGIN$' -e '^COMMIT$'; }
ts() { date -u +%Y-%m-%dT%H:%M:%S.%3NZ; }
# globex-key variant used in E02-H11*: prints body then [HTTP status]
g() { local m=$1 p=$2; shift 2; curl -sS -X "$m" -H 'Authorization: Bearer globex-key' -H 'Content-Type: application/json' -w ' [%{http_code}]' "$QUEUELET_URL$p" "$@"; echo; }
