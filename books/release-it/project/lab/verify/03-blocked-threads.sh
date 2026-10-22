#!/usr/bin/env bash
# Lesson 03 — blocked threads. First, fail fast: invalid requests should be
# rejected before they cost a control round trip or open a tenant file.
# Then the `errs` mix (30 % of writes violate items.seq UNIQUE) at 50 % for
# 60 s: does a failed write leak its connection and its file's write lock?
# The verdict reads the nodes' own answers: until lessons 05-06 a blip in
# control still reaches tenants through every request's resolve, which this
# lesson does not change. A few "database is locked" from writes that
# collide with a backup's lock are F08's tight BUSY loop (lesson 04).
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY / 2))
OUT=03-errs.csv

reset_stack

log "fail fast: 50 empty executes, then 40 writes sent to /query on databases nobody has opened"
key=$(jq -r '[.tenants[] | select(.class == "normal")][10].key' "$RESULTS/manifest.json")
resolves() { prom 'sum(http_requests_total{job="control",route="POST /internal/resolve"})'; }
r0=$(resolves)
for _ in $(seq 50); do
  curl -s -o /dev/null "$EDGE/v1/databases/default/execute" -H "Authorization: Bearer $key" -d '{"statements": []}'
done
sleep 2.5
resolve_cost=$(awk -v a="$(resolves)" -v b="$r0" 'BEGIN { print a - b }')
o0=$(prom 'sum(sqlite_open_databases)')
jq -r '.tenants[] | . as $t | .dbs[] | select(.name != "default") | "\($t.key) \(.name)"' "$RESULTS/manifest.json" |
  while read -r k db; do
    curl -s -o /dev/null "$EDGE/v1/databases/$db/query" -H "Authorization: Bearer $k" -d '{"sql": "DELETE FROM items"}'
  done
sleep 2.5
open_cost=$(awk -v a="$(prom 'sum(sqlite_open_databases)')" -v b="$o0" 'BEGIN { print a - b }')

log "errs mix at $RATE req/s for 60 s"
since=$(date -u +%Y-%m-%dT%H:%M:%S.%NZ)
start_load -scenario errs -rate "$RATE" -duration 60s -out "/results/$OUT"
for t in 15 30 45 60; do at "$t"; log "t=$t s fds node-a $(prom 'process_open_fds{instance="node-a"}') node-b $(prom 'process_open_fds{instance="node-b"}')"; done
wait_load >&2
max_fds=$(prom 'max(max_over_time(process_open_fds{job="node"}[75s]))')
node_5xx=$(prom 'sum(increase(http_requests_total{job="node",route=~"POST /internal/db/.*",status=~"5.."}[75s]))')
at 68
health_a=$(curl -s -m 2 -o /dev/null -w '%{http_code}' http://127.0.0.1:18071/health || true)
health_b=$(curl -s -m 2 -o /dev/null -w '%{http_code}' http://127.0.0.1:18072/health || true)
fds_after=$(prom 'max(process_open_fds{job="node"})')

report_table "$OUT"
read -r fivexx valid_pct p50_409 < <(q "$OUT" "SELECT
    count(*) FILTER (status >= 500 OR err_class IN ('client_timeout', 'conn')),
    round(100.0 * count(*) FILTER (tenant_class = 'normal' AND route = 'execute' AND status = 200 AND err_class IS NULL)
          / count(*) FILTER (tenant_class = 'normal' AND route = 'execute'), 2),
    round(quantile_cont(latency_ms, 0.5) FILTER (status = 409), 1)
  FROM runs" | tr ',' ' ')
locked=$(compose logs --no-log-prefix --no-color --since "$since" node-a node-b 2>/dev/null | grep -c 'database is locked' || true)

echo
echo "Fail fast"
printf '  control resolves spent on 50 empty executes       %s\n' "$resolve_cost"
printf '  tenant files opened by 40 writes sent to /query   %s\n' "$open_cost"
echo "errs mix, 60 s"
printf '  max process_open_fds on a node                    %s (limit 1024)\n' "$max_fds"
printf '  5xx answered by the nodes                         %s\n' "$node_5xx"
printf '  5xx or timeouts seen by tenants (all causes)      %s\n' "$fivexx"
printf '  valid writes that succeeded                       %s %%\n' "$valid_pct"
printf '  409 constraint p50                                %s ms\n' "$p50_409"
printf '  "database is locked" in node logs                 %s\n' "$locked"
printf '  8 s after the load: node-a /health %s, node-b /health %s, max fds %s\n' "$health_a" "$health_b" "$fds_after"

check resolves_for_invalid "$resolve_cost" "==" 0
check files_opened_for_invalid "$open_cost" "==" 0
check max_node_fds "$max_fds" "<=" 600
check node_5xx "$node_5xx" "==" 0
check valid_write_pct "$valid_pct" ">=" 95
check constraint_p50_ms "$p50_409" "<=" 20
check locked_errors "$locked" "<=" 5
check node_a_healthy_after "$health_a" "==" 200
verdict
