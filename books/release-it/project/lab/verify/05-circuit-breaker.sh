#!/usr/bin/env bash
# Lesson 05 — circuit breaker. Steady at 50 % for 60 s, plus 5 req/s to
# databases nobody touched during the warm-up (cold) and 1 create/s. After a
# 10 s warm-up, control becomes unreachable (its proxy accepts and never
# answers) from t=10 s to t=40 s. Does the data plane keep serving known
# databases, do the rest fail fast, and does control get left alone?
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY / 2))
OUT=05-breaker.csv
TL=$RESULTS/05-timeline.txt

reset_stack
rm -f "$TL"
log "steady at $RATE req/s + cold 5/s + create 1/s for 60 s; control unreachable t=10-40 s"
start_load -scenario steady -rate "$RATE" -cold-rate 5 -create-rate 1 -duration 60s -out "/results/$OUT"
( while true; do
    echo "$(awk -v a="$(date +%s.%N)" -v b="$T0" 'BEGIN { printf "%.1f", a - b }') $(prom 'max(breaker_state{job="gateway"})') $(haproxy_up)" >> "$TL"
    sleep 0.5
  done ) & P=$!
at 10; toxic control outage timeout upstream 1.0 '{"timeout":0}'
at 15; a0=$(prom 'sum(dependency_requests_total{job="gateway",dep="control"}) or vector(0)')
at 40; a1=$(prom 'sum(dependency_requests_total{job="gateway",dep="control"}) or vector(0)')
untoxic_busy control outage
wait_load >&2
at 63; kill "$P" 2>/dev/null || true
report_table "$OUT"

read -r ok_pct p50 < <(q "$OUT" "SELECT round(100.0 * count(*) FILTER (status = 200 AND err_class IS NULL) / count(*), 2),
    round(quantile_cont(latency_ms, 0.5), 1) FROM runs WHERE tenant_class = 'normal' AND ts_ms BETWEEN 12000 AND 40000" | tr ',' ' ')
read -r miss_p50 miss_503 miss_n < <(q "$OUT" "SELECT round(quantile_cont(latency_ms, 0.5), 1), count(*) FILTER (status = 503), count(*)
  FROM runs WHERE (tenant_class = 'cold' OR route = 'create') AND ts_ms BETWEEN 14000 AND 40000" | tr ',' ' ')
attempts=$(awk -v a="$a1" -v b="$a0" 'BEGIN { print a - b }')
closed_after=$(awk '$1 >= 40 && $2 == 0 { printf "%.1f", $1 - 40; found = 1; exit } END { if (!found) print "NaN" }' "$TL")
up_min=$(awk '$1 >= 14 && $1 <= 40 { if (m == "" || $3 < m) m = $3 } END { print (m == "" ? "NaN" : m) }' "$TL")
t_open=$(prom 'sum(breaker_transitions_total{job="gateway",from="closed",to="open"}) or vector(0)')
t_half=$(prom 'sum(breaker_transitions_total{job="gateway",from="open",to="half-open"}) or vector(0)')
t_close=$(prom 'sum(breaker_transitions_total{job="gateway",from="half-open",to="closed"}) or vector(0)')
hits=$(prom 'sum(route_cache_hits_total) or vector(0)')

echo
echo "control unreachable t=10-40 s"
printf '  known databases served during the outage (t=12-40 s)   %s %% (p50 %s ms)\n' "$ok_pct" "$p50"
printf '  cold databases and creates during the outage           %s of %s answered 503, p50 %s ms\n' "$miss_503" "$miss_n" "$miss_p50"
printf '  calls the gateways made to control, t=15-40 s          %s (one trial per gateway per 5 s cool-down: 15)\n' "$attempts"
printf '  HAProxy gateways UP during the outage (min)            %s of 3\n' "$up_min"
printf '  breaker transitions closed>open %s, open>half-open %s, half-open>closed %s\n' "$t_open" "$t_half" "$t_close"
printf '  all breakers closed again after the outage ended      %s s\n' "$closed_after"
printf '  route cache hits                                       %s\n' "$hits"

check outage_success_pct "$ok_pct" ">=" 95
check outage_p50_ms "$p50" "<=" 10
check miss_p50_ms "$miss_p50" "<=" 10
check control_calls_during_outage "$attempts" "<=" 20
check breaker_closed_after_s "$closed_after" "<=" 8
check gateways_up_during_outage "$up_min" "==" 3
verdict
