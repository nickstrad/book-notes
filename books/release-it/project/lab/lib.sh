# Shared helpers for the lab scripts. Source it; do not run it.
# Everything here drives the stack from outside: compose, the Toxiproxy API,
# Prometheus, the public API, DuckDB over the results.
set -euo pipefail

LAB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$LAB_DIR")"
RESULTS="$PROJECT_DIR/results"
# The tree whose code is under test (REF runs point it at the extracted REF).
LF_TREE="${LF_TREE:-$PROJECT_DIR}"
export LF_CONF="${LF_CONF:-$PROJECT_DIR/deploy}"

TOXI=http://127.0.0.1:18474
PROM=http://127.0.0.1:19090
EDGE=http://127.0.0.1:18080

compose() { docker compose -f "$PROJECT_DIR/deploy/compose.yml" "$@"; }

log() { printf '[%s] %s\n' "$(date +%H:%M:%S)" "$*" >&2; }

# wait_http URL [tries] — wait until URL answers 2xx.
wait_http() {
  local url=$1 tries=${2:-120}
  for _ in $(seq "$tries"); do
    curl -fs -o /dev/null --max-time 1 "$url" && return 0
    sleep 0.5
  done
  log "timed out waiting for $url"; return 1
}

psql_() { compose exec -T postgres psql -U litefleet -d postgres -v ON_ERROR_STOP=1 -qAt -c "$1"; }

# seed_key — identifies a seeded state: seed arguments, the registry schema of
# the tree under test, and the seed generator.
seed_key() {
  { echo "${SEED_ARGS[*]:-}"; cat "$LF_TREE"/migrations/*.sql; cat "$PROJECT_DIR/cmd/seedgen/main.go"; } |
    sha1sum | cut -c1-10
}

# reset_stack [compose -f overrides...] — every service recreated, tenant
# volumes and the registry database fresh, data seeded. The first reset for a
# given seed key runs seedgen (~1 min) and snapshots the result (a Postgres
# template database plus a copy of the tenant files); later resets restore the
# snapshot in seconds.
reset_stack() {
  local extra=("$@") key hit=0
  key=$(seed_key)
  log "resetting stack (seed $key)"
  compose "${extra[@]}" --profile lab down --remove-orphans >/dev/null 2>&1 || true
  docker volume rm -f litefleet_node-a-data litefleet_node-b-data litefleet_blobs >/dev/null
  docker volume create litefleet-seedcache >/dev/null
  compose "${extra[@]}" up -d --no-build --wait postgres >/dev/null 2>&1
  psql_ "DROP DATABASE IF EXISTS litefleet WITH (FORCE)"
  if [[ $(psql_ "SELECT count(*) FROM pg_database WHERE datname = 'seed_$key'") == 1 ]] &&
     docker run --rm -v litefleet-seedcache:/cache alpine test -f "/cache/$key/manifest.json"; then
    psql_ "CREATE DATABASE litefleet TEMPLATE seed_$key"
    hit=1
  else
    psql_ "CREATE DATABASE litefleet"
  fi
  compose "${extra[@]}" up -d --no-build >/dev/null 2>&1
  wait_http http://127.0.0.1:18090/health
  wait_http http://127.0.0.1:18071/health
  wait_http http://127.0.0.1:18072/health
  for p in 18001 18002 18003; do wait_http "http://127.0.0.1:$p/health"; done
  wait_http "$TOXI/proxies"
  curl -fs -XPOST "$TOXI/reset" >/dev/null
  if ((hit)); then
    docker run --rm -v litefleet-seedcache:/cache:ro -v litefleet_node-a-data:/a -v litefleet_node-b-data:/b \
      -v "$RESULTS":/results alpine sh -c "
        cp -a /cache/$key/node-a/. /a/tenants/ && cp -a /cache/$key/node-b/. /b/tenants/ &&
        cp /cache/$key/manifest.json /results/manifest.json"
    log "restored seed snapshot $key"
  else
    log "seeding (first run for seed $key)"
    compose run --rm --no-deps seedgen "${SEED_ARGS[@]}" 2>&1 | grep -E 'seeded|wrote' >&2
    compose stop control postgres-exporter >/dev/null 2>&1
    psql_ "DROP DATABASE IF EXISTS seed_$key"
    psql_ "CREATE DATABASE seed_$key TEMPLATE litefleet"
    compose start control postgres-exporter >/dev/null 2>&1
    docker run --rm -v litefleet-seedcache:/cache -v litefleet_node-a-data:/a:ro -v litefleet_node-b-data:/b:ro \
      -v "$RESULTS":/results:ro alpine sh -c "
        rm -rf /cache/$key && mkdir -p /cache/$key/node-a /cache/$key/node-b &&
        cp -a /a/tenants/. /cache/$key/node-a/ && cp -a /b/tenants/. /cache/$key/node-b/ &&
        cp /results/manifest.json /cache/$key/"
    wait_http http://127.0.0.1:18090/health
    log "snapshotted seed $key"
  fi
  wait_http "$PROM/-/ready"
  warmup
}

# warmup — gentle traffic so pools, handles and caches are warm before a
# measurement (a cold Postgres backend plans every query from scratch, a node
# opens each tenant file on first use).
warmup() {
  log "warming up"
  loadgen -scenario steady -rate "${WARM_RATE:-100}" -duration "${WARM_FOR:-10s}" -out /results/warmup.csv >/dev/null
  sleep 2
}

# migrate_online — apply the tree under test's online migrations, if it has
# any (migrations/online appears in lesson 08).
migrate_online() {
  local f
  for f in "$LF_TREE"/migrations/online/*.sql; do
    [[ -e $f ]] || continue
    log "migration $(basename "$f")"
    compose exec -T postgres psql -U litefleet -d litefleet -v ON_ERROR_STOP=1 -q < "$f"
  done
}

# toxic PROXY NAME TYPE STREAM TOXICITY ATTRS_JSON
toxic() {
  curl -fs -XPOST "$TOXI/proxies/$1/toxics" -H 'Content-Type: application/json' \
    -d "{\"name\":\"$2\",\"type\":\"$3\",\"stream\":\"$4\",\"toxicity\":$5,\"attributes\":$6}" >/dev/null
  log "toxic $2 ($3 $6) on $1"
}
untoxic() { curl -fs -XDELETE "$TOXI/proxies/$1/toxics/$2" >/dev/null; log "removed toxic $2 from $1"; }

# untoxic_busy PROXY NAME — remove a toxic from a proxy with many open links.
# Toxiproxy 2.12 can wedge a proxy for many seconds when a latency or timeout
# toxic is removed while thousands of links are open (the toxic flushes held
# data to connections that are already gone). Disabling the proxy closes
# every link first; the toxic is then removed from an idle proxy and traffic
# resumes on fresh connections, at the cost of one reset per open link.
untoxic_busy() {
  curl -fs -XPOST "$TOXI/proxies/$1" -d '{"enabled":false}' >/dev/null
  curl -fs -XDELETE "$TOXI/proxies/$1/toxics/$2" >/dev/null
  curl -fs -XPOST "$TOXI/proxies/$1" -d '{"enabled":true}' >/dev/null
  log "removed toxic $2 from $1 (proxy cycled)"
}

# prom EXPR — first value of an instant query, or "NaN".
prom() {
  curl -fsG "$PROM/api/v1/query" --data-urlencode "query=$1" |
    jq -r '.data.result[0].value[1] // "NaN"' |
    awk '/^-?[0-9.e+-]+$/ { printf "%.4g\n", $1; next } { print }'
}

# loadgen ARGS... — run the load generator; prints its JSON summary.
loadgen() { compose run --rm --no-deps loadgen "$@" 2>/dev/null | tail -n 1; }

# duck SQL — run DuckDB in the report container with the results mounted.
duck() { compose run --rm --no-deps -T report -csv -noheader -c "$1" 2>/dev/null; }

# runs_view CSV — SQL prelude defining view `runs` over a loadgen CSV.
runs_view() {
  echo "CREATE OR REPLACE VIEW runs AS SELECT * FROM read_csv('/results/$1', header=true, nullstr='',
    columns={'ts_ms':'BIGINT','scenario':'VARCHAR','route':'VARCHAR','tenant_class':'VARCHAR','node':'VARCHAR',
             'status':'INTEGER','latency_ms':'DOUBLE','err_class':'VARCHAR','instance':'VARCHAR'});"
}

# q CSV SQL — scalar/CSV query over a run.
q() { duck "$(runs_view "$1") $2"; }

# report_table CSV — the standard per-window table.
report_table() {
  compose run --rm --no-deps -T report -box -c "$(runs_view "$1") $(cat "$LAB_DIR/report.sql")" 2>/dev/null
}

# cmp A OP B — float comparison through awk; exit status is the result.
cmp() { awk -v a="$1" -v b="$3" "BEGIN { exit !(a $2 b) }"; }

# Verdict bookkeeping: check NAME "a OP b" description.
FAILS=()
PASSES=()
check() {
  local name=$1 a=$2 op=$3 b=$4
  if [[ $a != NaN ]] && cmp "$a" "$op" "$b"; then PASSES+=("$name $a $op $b"); else FAILS+=("$name $a (want $op $b)"); fi
}

# verdict — print the verdict line; exit 0 if every check held, 1 otherwise.
verdict() {
  local joined
  echo
  if ((${#FAILS[@]} == 0)); then
    joined=$(IFS=';'; echo "${PASSES[*]}")
    echo "${VERDICT_PASS:-IMPROVED:} ${joined//;/; }"
    exit 0
  fi
  joined=$(IFS=';'; echo "${FAILS[*]}")
  echo "${VERDICT_FAIL:-FLAW REPRODUCED:} ${joined//;/; }"
  exit 1
}

SEED_ARGS=()

# start_load ARGS... — run loadgen in the background, its first request at
# T0 (epoch seconds, set here, 3 s from now so the container has started), so
# the CSV's ts_ms and the script's `at` share one clock. Sets LOAD_PID.
start_load() {
  local t0_ms=$(( $(date +%s%3N) + 3000 ))
  T0=$(awk -v m="$t0_ms" 'BEGIN { printf "%.3f", m / 1000 }')
  compose run --rm --no-deps loadgen -start-at "$t0_ms" "$@" 2>/dev/null | tail -n 1 > "$RESULTS/.summary.json" &
  LOAD_PID=$!
}

# at SECONDS — sleep until T0 + SECONDS.
at() {
  local wait
  wait=$(awk -v t0="$T0" -v s="$1" -v now="$(date +%s.%N)" 'BEGIN { w = t0 + s - now; print (w > 0 ? w : 0) }')
  sleep "$wait"
}

# wait_load — wait for the background loadgen; prints its summary.
wait_load() { wait "$LOAD_PID" || true; cat "$RESULTS/.summary.json"; }

# poll_health NAME URL OUT — append "t_seconds http_code body" every 0.25 s
# until killed. Used to time health transitions against T0.
poll_health() {
  local url=$2 out=$3
  while true; do
    local now code body
    now=$(date +%s.%N)
    body=$(curl -s --max-time 2 -w ' %{http_code}' "$url" || echo ' 000')
    code=${body##* }
    printf '%s %s %s\n' "$(awk -v a="$now" -v b="$T0" 'BEGIN { printf "%.2f", a - b }')" "$code" "${body% *}" >> "$out"
    sleep 0.25
  done
}

# haproxy_up — gateways HAProxy currently considers UP.
haproxy_up() { curl -s "http://127.0.0.1:18404/stats;csv" | awk -F, '$1 == "gateways" && $2 ~ /^gateway/ && $18 ~ /^UP/' | wc -l; }
