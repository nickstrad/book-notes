#!/usr/bin/env bash
# Lesson 08 — unbounded result sets. Steady at 30 % for 45 s while the whale
# tenant runs `SELECT * FROM events` (15 000 rows, 60 MB) ten at a time,
# nodes at 512 MB and gateways at 256 MB. Then ten refreshes of the usage
# report over the 5 M-row ledger.
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY * 3 / 10))
OUT=08-whale.csv
SERVICES=(node-a node-b gateway-a gateway-b gateway-c)

restarts() { local s; for s in "${SERVICES[@]}"; do docker inspect -f '{{.RestartCount}}' "litefleet-$s-1"; done | paste -sd+ | bc; }

reset_stack
migrate_online
r0=$(restarts)
log "steady at $RATE req/s for 45 s + 10 concurrent whale scans"
loadgen -scenario whale -rate "$RATE" -extra 10 -duration 45s -out "/results/$OUT" >&2
r1=$(restarts)
rss_node=$(prom 'max(max_over_time(process_resident_memory_bytes{job="node"}[50s])) / 1048576')
rss_gw=$(prom 'max(max_over_time(process_resident_memory_bytes{job="gateway"}[50s])) / 1048576')
report_table "$OUT"
read -r whale_ok whale_n whale_p99 other_5xx node_a_ok < <(q "$OUT" "SELECT
    count(*) FILTER (tenant_class = 'whale' AND status = 200 AND err_class IS NULL), count(*) FILTER (tenant_class = 'whale'),
    round(quantile_cont(latency_ms, 0.99) FILTER (tenant_class = 'whale' AND status = 200 AND err_class IS NULL)),
    count(*) FILTER (tenant_class = 'normal' AND (status >= 500 OR err_class IN ('client_timeout', 'conn'))),
    round(100.0 * count(*) FILTER (tenant_class = 'normal' AND node = 'node-a' AND status = 200) / count(*) FILTER (tenant_class = 'normal' AND node = 'node-a'), 2)
  FROM runs" | tr ',' ' ')

log "ten usage-report refreshes"
worst=0
for _ in $(seq 10); do
  t=$(curl -s -o /dev/null -m 10 -w '%{time_total}' "$EDGE/v1/admin/usage?window=24h" -H "Authorization: Bearer lfk_admin_ops")
  worst=$(awk -v a="$t" -v b="$worst" 'BEGIN { print (a > b ? a : b) }')
done

echo
printf '  container restarts during the run (nodes, gateways)   %s\n' "$((r1 - r0))"
printf '  peak RSS: node %s MB (limit 512), gateway %s MB (limit 256)\n' "$rss_node" "$rss_gw"
printf '  whale scans answered %s of %s, p99 %s ms\n' "$whale_ok" "$whale_n" "$whale_p99"
printf '  other tenants: 5xx or timeouts %s; node-a tenants served %s %%\n' "$other_5xx" "$node_a_ok"
printf '  slowest of ten usage reports                           %s s\n' "$worst"

whale_pct=$(awk -v a="$whale_ok" -v b="$whale_n" 'BEGIN { printf "%.1f", b ? 100 * a / b : 0 }')
check restarts "$((r1 - r0))" "==" 0
check whale_answered_pct "$whale_pct" ">=" 99
# Ten concurrent scans share the database's two read lanes (lesson 06), so
# most of a scan's latency is waiting for a lane.
check whale_p99_ms "${whale_p99/NULL/NaN}" "<=" 1000
check other_tenant_5xx "$other_5xx" "==" 0
check report_max_s "$worst" "<=" 0.5
verdict
