# Source before every lab/client command in experiment 06 (isolated stack queuelet-e06).
export COMPOSE_PROJECT_NAME=queuelet-e06 QUEUELET_IMAGE_TAG=e06
export QUEUELET_API_PORT=18270 QUEUELET_HOOK_PORT=18271 QUEUELET_PG_PORT=15503
export QUEUELET_URL=http://localhost:18270 QUEUELET_HOOK_URL=http://localhost:18271
export PROJ=/tmp/explore-it-experiment-06/books/explore-it/project
export E06=/tmp/explore-it-experiment-06/books/explore-it/exploration/experiments/06-relate-it
export PGPASSWORD=explorer
dc() { docker compose -p queuelet-e06 -f "$PROJ/deploy/compose.yml" "$@"; }
# read-only SQL as the explorer role
sql() { psql -h localhost -p 15503 -U explorer -d queuelet -X -A -F' | ' -P footer=off -c "$1"; }
# curl wrapper: k KEY METHOD PATH [BODY] -> prints body and status
k() { local key=$1 m=$2 p=$3 b=${4-}; if [ -n "$b" ]; then curl -s -X "$m" -H "Authorization: Bearer $key" -H 'Content-Type: application/json' -d "$b" -w ' [%{http_code}]\n' "$QUEUELET_URL$p"; else curl -s -X "$m" -H "Authorization: Bearer $key" -w ' [%{http_code}]\n' "$QUEUELET_URL$p"; fi; }
