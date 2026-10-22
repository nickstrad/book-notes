#!/usr/bin/env bash
# Lesson 06 — bulkheads. Three phases, each with steady load at 50 %:
#  a. a noisy tenant on node-a runs 20 concurrent ~1.5 s WITH RECURSIVE
#     queries in a loop from t=10 s to t=45 s;
#  b. the ops dashboard refreshes the usage report 20 at a time for 30 s,
#     with cold reads and creates (the paths that still need control);
#  c. the blob store's proxy drops to 50 KB/s and the node is told to back
#     up the 5 MB "bulky" database while its tenant keeps writing to it.
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY / 2))

reset_stack

log "phase a: noisy tenant on node-a, 20 concurrent recursive queries t=10-45 s"
loadgen -scenario noisy -rate "$RATE" -extra 20 -extra-from 10s -extra-for 35s -recursion 420000 \
  -duration 45s -out /results/06a-noisy.csv >&2
report_table 06a-noisy.csv
read -r base_p99 noisy_p99 other_5xx noisy_429 < <(q 06a-noisy.csv "SELECT
    round(quantile_cont(latency_ms, 0.99) FILTER (ts_ms BETWEEN 2000 AND 10000)),
    round(quantile_cont(latency_ms, 0.99) FILTER (ts_ms BETWEEN 15000 AND 45000)),
    count(*) FILTER (ts_ms BETWEEN 15000 AND 45000 AND (status >= 500 OR err_class IN ('client_timeout', 'conn'))),
    (SELECT count(*) FROM runs WHERE tenant_class = 'noisy' AND status = 429)
  FROM runs WHERE tenant_class = 'normal' AND node = 'node-a'" | tr ',' ' ')
ratio_a=$(awk -v a="$noisy_p99" -v b="$base_p99" 'BEGIN { printf "%.2f", b ? a / b : 0 }')

log "phase b: 20 concurrent usage-report refreshes for 30 s, with cold reads and creates"
rm -f "$RESULTS/06-health-control.txt"
start_load -scenario reports -rate "$RATE" -extra 20 -extra-from 5s -cold-rate 5 -create-rate 2 -duration 35s -out /results/06b-reports.csv
poll_health control http://127.0.0.1:18090/health "$RESULTS/06-health-control.txt" & P=$!
at 5; c0=$(prom 'sum(http_requests_total{job="control",route=~"POST /internal/(resolve|databases)",status="503"}) or vector(0)')
at 35; c1=$(prom 'sum(http_requests_total{job="control",route=~"POST /internal/(resolve|databases)",status="503"}) or vector(0)')
wait_load >&2
kill "$P" 2>/dev/null || true
report_table 06b-reports.csv
read -r crit_5xx crit_n < <(q 06b-reports.csv "SELECT count(*) FILTER (status >= 500 OR err_class IN ('client_timeout', 'conn')), count(*)
  FROM runs WHERE tenant_class <> 'ops' AND ts_ms BETWEEN 7000 AND 35000" | tr ',' ' ')
read -r rep_ok rep_busy rep_p99 < <(q 06b-reports.csv "SELECT count(*) FILTER (status = 200), count(*) FILTER (status = 503),
  round(quantile_cont(latency_ms, 0.99) FILTER (status = 200)) FROM runs WHERE tenant_class = 'ops'" | tr ',' ' ')
control_busy=$(awk -v a="$c1" -v b="$c0" 'BEGIN { print a - b }')
health_bad=$(awk '$1 >= 5 && $1 <= 35 && $2 != 200' "$RESULTS/06-health-control.txt" | wc -l)
health_n=$(awk '$1 >= 5 && $1 <= 35' "$RESULTS/06-health-control.txt" | wc -l)

log "phase c: blob store at 50 KB/s, backup of the 5 MB bulky database during its tenant's writes"
read -r bulky bulky_id < <(jq -r '.tenants[] | select(.class == "bulky") | "\(.id)/default \(.dbs[0].id)"' "$RESULTS/manifest.json")
toxic blobstore slow bandwidth upstream 1.0 '{"rate":50}'
start_load -scenario steady -rate "$RATE" -focus "$bulky" -focus-rate 10 -duration 30s -out /results/06c-backup.csv
at 3
curl -s -m 120 -o /dev/null -XPOST "http://127.0.0.1:18071/internal/backup/$bulky_id" &
wait_load >&2
lock_hold=$(prom 'histogram_quantile(1, sum by (le) (rate(writer_lock_hold_seconds_bucket{instance="node-a"}[40s])))')
untoxic_busy blobstore slow
report_table 06c-backup.csv
read -r tgt_n tgt_fail tgt_p99 < <(q 06c-backup.csv "SELECT count(*), count(*) FILTER (status <> 200 OR err_class IS NOT NULL),
  round(quantile_cont(latency_ms, 0.99)) FROM runs WHERE tenant_class = 'target' AND ts_ms >= 5000" | tr ',' ' ')

echo
echo "a. noisy tenant on node-a"
printf '  other node-a tenants p99: before the noisy tenant %s ms, during %s ms (%sx)\n' "$base_p99" "$noisy_p99" "$ratio_a"
printf '  other node-a tenants 5xx or timeouts during      %s\n' "$other_5xx"
printf '  noisy tenant answered 429 tenant_busy            %s\n' "$noisy_429"
echo "b. 20 concurrent usage reports"
printf '  control answered 503 pool_timeout to resolves or creates %s\n' "$control_busy"
printf '  tenant-visible 5xx or timeouts (steady, cold, creates)    %s of %s\n' "$crit_5xx" "$crit_n"
printf '  control /health not 200                                %s of %s checks\n' "$health_bad" "$health_n"
printf '  reports served %s (p99 %s ms), refused 503 report_busy %s\n' "$rep_ok" "$rep_p99" "$rep_busy"
echo "c. backup of the bulky database through a 50 KB/s blob store"
printf '  writes to the database being backed up: %s, failed %s, p99 %s ms\n' "$tgt_n" "$tgt_fail" "$tgt_p99"
printf '  longest write-lock hold by a backup                    %s s\n' "$lock_hold"

check noisy_other_p99_ms "$noisy_p99" "<=" 200
check noisy_other_5xx "$other_5xx" "==" 0
check reports_control_pool_timeouts "$control_busy" "==" 0
check reports_control_health_bad "$health_bad" "==" 0
check backup_target_failures "$tgt_fail" "==" 0
check backup_target_p99_ms "$tgt_p99" "<=" 250
check writer_lock_hold_s "$lock_hold" "<=" 1
verdict
