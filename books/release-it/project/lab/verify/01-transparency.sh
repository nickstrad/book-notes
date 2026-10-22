#!/usr/bin/env bash
# Lesson 01 — transparency. Steady load at 50 % for 60 s; node-a gets 800 ms
# of latency from t=15 s to t=33 s; Postgres is paused from t=35 s to t=45 s.
# Can the operator see which dependency is slow, does health say so, and do
# the logs tie a slow request to its cause?
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY / 2))
OUT=01-transparency.csv
TL=$RESULTS/01-timeline.txt

reset_stack
rm -f "$RESULTS"/01-health-*.txt "$TL"
since=$(date -u +%Y-%m-%dT%H:%M:%SZ)
log "steady at $RATE req/s for 60 s"
start_load -scenario steady -rate "$RATE" -duration 60s -out "/results/$OUT"
poll_health control http://127.0.0.1:18090/health "$RESULTS/01-health-control.txt" & P1=$!
poll_health gateway http://127.0.0.1:18001/health "$RESULTS/01-health-gateway.txt" & P2=$!
( while true; do echo "$(awk -v a="$(date +%s.%N)" -v b="$T0" 'BEGIN { printf "%.1f", a - b }') $(haproxy_up)" >> "$TL"; sleep 1; done ) & P3=$!

at 15; toxic node-a slow latency downstream 1.0 '{"latency":800}'
at 32
node_a_p99=$(prom 'histogram_quantile(0.99, sum by (le) (rate(dependency_duration_seconds_bucket{job="gateway",dep="node-a"}[10s])))')
node_b_p99=$(prom 'histogram_quantile(0.99, sum by (le) (rate(dependency_duration_seconds_bucket{job="gateway",dep="node-b"}[10s])))')
edge_p99=$(prom 'histogram_quantile(0.99, sum by (le) (rate(http_request_duration_seconds_bucket{job="gateway",route=~"POST /v1/databases/.*"}[10s])))')
at 33; untoxic node-a slow
at 35; compose pause postgres >/dev/null 2>&1; log "paused postgres"
at 45; compose unpause postgres >/dev/null 2>&1; log "unpaused postgres"
wait_load >&2
kill "$P1" "$P2" "$P3" 2>/dev/null; wait 2>/dev/null || true

report_table "$OUT"

# Health: seconds from the pause (t=35) to the first failed answer.
first_failed() { awk '$1 >= 35 && $1 <= 47 && ($2 == 503 || $0 ~ /failed/) { printf "%.2f", $1 - 35; found = 1; exit } END { if (!found) print "NaN" }' "$1"; }
# The gateway's health counts as reporting the outage when it says failed
# or, from lesson 05 (it keeps serving from cached routes), degraded.
first_reported() { awk '$1 >= 35 && $1 <= 47 && ($2 == 503 || $0 ~ /failed|degraded/) { printf "%.2f", $1 - 35; found = 1; exit } END { if (!found) print "NaN" }' "$1"; }
ctl_failed=$(first_failed "$RESULTS/01-health-control.txt")
gw_failed=$(first_reported "$RESULTS/01-health-gateway.txt")
ctl_detail=$(awk '$1 >= 35 && $1 <= 47 && $2 == 503 { $1 = ""; $2 = ""; print; exit }' "$RESULTS/01-health-control.txt")
gw_detail=$(awk '$1 >= 35 && $1 <= 47 && ($2 == 503 || $0 ~ /failed|degraded/) { $1 = ""; $2 = ""; print; exit }' "$RESULTS/01-health-gateway.txt")
up_during_pause=$(awk '$1 >= 38 && $1 <= 45 { if (min == "" || $2 < min) min = $2 } END { print (min == "" ? "NaN" : min) }' "$TL")

# Logs: every line JSON; every request line carries a request id; a slow
# node-a call can be found by its dependency line.
logs=$RESULTS/01-logs.txt
compose logs --no-log-prefix --no-color --since "$since" gateway-a gateway-b gateway-c control node-a node-b > "$logs" 2>/dev/null
total=$(wc -l < "$logs")
json=$(jq -R 'fromjson? | select(type == "object")' -c < "$logs" | wc -l)
reqs=$(jq -R 'fromjson? | select(.msg == "request")' -c < "$logs" | wc -l)
reqs_with_id=$(jq -R 'fromjson? | select(.msg == "request" and (.request_id // "") != "")' -c < "$logs" | wc -l)
json_pct=$(awk -v a="$json" -v b="$total" 'BEGIN { printf "%.1f", b ? 100 * a / b : 0 }')
id_pct=$(awk -v a="$reqs_with_id" -v b="$reqs" 'BEGIN { printf "%.1f", b ? 100 * a / b : 0 }')
example=$(jq -R -c -n 'first(inputs | fromjson? | select(.msg == "dependency" and .dep == "node-a" and .dur_ms > 700) | {request_id, dep, dur_ms, outcome})' < "$logs")

echo
echo "Prometheus and health during the run"
printf '  up{job="gateway"}                                 %s\n' "$(prom 'sum(up{job="gateway"})')"
printf '  gateway → node-a p99 at t=32 s (latency on node-a) %s s\n' "$node_a_p99"
printf '  gateway → node-b p99 at t=32 s                     %s s\n' "$node_b_p99"
printf '  gateway request p99 at t=32 s                      %s s\n' "$edge_p99"
printf '  control /health first failed after pause           %s s  %s\n' "$ctl_failed" "$ctl_detail"
printf '  gateway-a /health first failed/degraded after pause %s s  %s\n' "$gw_failed" "$gw_detail"
printf '  HAProxy gateways UP during pause (min)             %s of 3\n' "$up_during_pause"
printf '  log lines that are JSON                            %s of %s (%s %%)\n' "$json" "$total" "$json_pct"
printf '  request log lines with a request_id                %s of %s (%s %%)\n' "$reqs_with_id" "$reqs" "$id_pct"
printf '  a slow node-a call, found by its log line          %s\n' "${example:-none}"

check up_gateway "$(prom 'sum(up{job="gateway"})')" "==" 3
check node_a_dep_p99_s "$node_a_p99" ">=" 0.7
check node_b_dep_p99_s "$node_b_p99" "<=" 0.2
check control_health_failed_after_s "$ctl_failed" "<=" 3
check gateway_health_failed_after_s "$gw_failed" "<=" 5
check log_json_pct "$json_pct" "==" 100
check request_id_pct "$id_pct" "==" 100
verdict
