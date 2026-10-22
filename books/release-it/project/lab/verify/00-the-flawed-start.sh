#!/usr/bin/env bash
# Lesson 00 — the baseline. v0 passed QA: under gentle load (the steady mix at
# half the fleet's calibrated capacity) it serves every request quickly. The
# flaws are latent; nothing here triggers them.
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY / 2))
OUT=00-steady.csv

reset_stack
log "steady at $RATE req/s (50 % of fleet capacity $FLEET_CAPACITY) for 60 s"
loadgen -scenario steady -rate "$RATE" -duration 60s -out "/results/$OUT" >&2

report_table "$OUT"
read -r total ok five p99 < <(q "$OUT" "SELECT count(*), count(*) FILTER (status BETWEEN 200 AND 299 AND err_class IS NULL),
  count(*) FILTER (status >= 500 OR err_class IN ('client_timeout','conn')), round(quantile_cont(latency_ms, 0.99), 1) FROM runs" | tr ',' ' ')
success=$(awk -v a="$ok" -v b="$total" 'BEGIN { printf "%.2f", 100 * a / b }')

echo
echo "Prometheus: services reporting metrics"
for job in gateway control node haproxy blobstore metering postgres; do
  printf '  up{job="%s"} = %s\n' "$job" "$(prom "sum(up{job=\"$job\"})")"
done

check "success_pct" "$success" ">=" 99.5
check "p99_ms" "$p99" "<=" 50
check "errors_5xx_or_timeout" "$five" "==" 0
VERDICT_PASS="BASELINE HOLDS:" VERDICT_FAIL="BASELINE BROKEN:" verdict
