#!/usr/bin/env bash
# Lesson 07 — shed load. The gateways are made the fleet's bottleneck
# (GATEWAY_CPUS=0.15 and a matching in-flight limit, NODE_CPUS=0.75: Compose
# overrides for this run). Offered load is 80 %
# of three gateways' capacity. At t=15 s gateway-b is killed, so the two
# survivors are offered 120 % of what they can serve; at t=60 s it is started
# again. Does the fleet keep serving what it can, or does it collapse?
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
export GATEWAY_CPUS=$L07_GATEWAY_CPUS NODE_CPUS=$L07_NODE_CPUS GATEWAY_SHED_LIMIT=$L07_GATEWAY_SHED_LIMIT
RATE=$((L07_GATEWAY_CAPACITY * 3 * 8 / 10))
OUT=07-burst.csv
TL=$RESULTS/07-timeline.txt

reset_stack
rm -f "$TL"
log "steady at $RATE req/s for 75 s (gateways at $GATEWAY_CPUS CPU); kill gateway-b at t=15 s, start it at t=60 s"
start_load -scenario steady -rate "$RATE" -duration 75s -out "/results/$OUT"
( while true; do echo "$(awk -v a="$(date +%s.%N)" -v b="$T0" 'BEGIN { printf "%.1f", a - b }') $(haproxy_up)" >> "$TL"; sleep 1; done ) & P=$!
at 15; compose kill gateway-b >/dev/null 2>&1; log "killed gateway-b"
at 58; inflight=$(prom 'max(max_over_time(http_inflight_requests{job="gateway"}[40s]))')
goroutines=$(prom 'max(max_over_time(go_goroutines{job="gateway"}[40s]))')
at 60; compose start gateway-b >/dev/null 2>&1; log "started gateway-b"
wait_load >&2
kill "$P" 2>/dev/null || true
report_table "$OUT"

goodput() { q "$OUT" "SELECT round(count(*) FILTER (status BETWEEN 200 AND 299 AND err_class IS NULL AND latency_ms < 3000) / (($2 - $1) / 1000.0), 1) FROM runs WHERE ts_ms >= $1 AND ts_ms < $2"; }
g_before=$(goodput 5000 15000)
g_after=$(goodput 25000 60000)
g_rejoin=$(goodput 65000 75000)
share=$(awk -v a="$g_after" -v b="$g_before" 'BEGIN { printf "%.1f", b ? 100 * a / b : 0 }')
# p99 of successful requests per 2 s window, and the median of those: a
# stall of the shared VM spoils one or two windows, not the measure.
read -r p99_ok p99_worst < <(q "$OUT" "WITH w AS (SELECT quantile_cont(latency_ms, 0.99) AS p99 FROM runs
    WHERE ts_ms BETWEEN 20000 AND 60000 AND status BETWEEN 200 AND 299 AND err_class IS NULL GROUP BY ts_ms // 2000)
  SELECT round(median(p99)), round(max(p99)) FROM w" | tr ',' ' ')
read -r p50_shed n_shed < <(q "$OUT" "SELECT round(quantile_cont(latency_ms, 0.5) FILTER (status = 503), 1), count(*) FILTER (status = 503)
  FROM runs WHERE ts_ms BETWEEN 20000 AND 60000" | tr ',' ' ')
b_rejoin=$(q "$OUT" "SELECT count(*) FILTER (instance = 'gateway-b' AND status = 200) FROM runs WHERE ts_ms >= 65000")
shed=$(prom 'sum(increase(shed_total[80s])) or vector(0)')

echo
printf '  goodput (2xx within 3 s): before the kill %s/s, after %s/s (%s %%), after the rejoin %s/s\n' "$g_before" "$g_after" "$share" "$g_rejoin"
printf '  p99 of successful requests after the kill   %s ms (median of 2 s windows; worst window %s ms)\n' "$p99_ok" "$p99_worst"
printf '  503s after the kill %s, p50 %s ms; shed_total %s\n' "$n_shed" "$p50_shed" "$shed"
printf '  max gateway in-flight / goroutines after the kill   %s / %s\n' "$inflight" "$goroutines"
printf '  requests gateway-b served after it came back       %s\n' "$b_rejoin"

check goodput_after_kill_pct "$share" ">=" 55
check ok_p99_ms "$p99_ok" "<=" 300
check max_gateway_inflight "$inflight" "<=" 32
check rejoined_gateway_served "$b_rejoin" ">" 0
verdict
