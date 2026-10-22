#!/usr/bin/env bash
# Lesson 02 — timeouts. Three phases:
#  1. steady at 50 % for 45 s; node-a's proxy accepts and never answers from
#     t=10 s to t=30 s (Toxiproxy `timeout` toxic on the upstream stream:
#     requests are swallowed before they reach the node. On the downstream
#     stream the toxic wedges Toxiproxy 2.12 when it is removed under
#     thousands of open links). Reads only: a write cancelled by the new
#     deadlines takes Execute's error path, which leaks its connection (and
#     the file's write lock) until lesson 03 fixes F05;
#  2. a tenant submits a runaway WITH RECURSIVE query (~30 s of node CPU)
#     through the public API with a 3 s client timeout;
#  3. a client opens a connection to a gateway and sends half a request.
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY / 2))
OUT=02-timeouts.csv

reset_stack
log "phase 1: steady reads at $RATE req/s for 45 s, node-a blackholed t=10-30 s"
start_load -scenario steady -write-frac 0 -rate "$RATE" -duration 45s -out "/results/$OUT"
at 10; toxic node-a blackhole timeout upstream 1.0 '{"timeout":0}'
at 29
inflight=$(prom 'max(max_over_time(http_inflight_requests{job="gateway"}[20s]))')
goroutines=$(prom 'max(max_over_time(go_goroutines{job="gateway"}[20s]))')
at 30; untoxic node-a blackhole
at 36
inflight_after=$(prom 'max(http_inflight_requests{job="gateway"})')
wait_load >&2
report_table "$OUT"
# A fast failure is the gateway's own answer: 504 when its deadline passes,
# or (from lesson 07) 503 `overloaded`, shed with Retry-After before the
# deadline is even reached.
read -r p99_out n_fail n_504 < <(q "$OUT" "SELECT round(quantile_cont(latency_ms, 0.99)) / 1000,
    count(*) FILTER (status <> 200 OR err_class IS NOT NULL),
    count(*) FILTER (status = 504 OR (status = 503 AND err_class = 'overloaded'))
  FROM runs WHERE node = 'node-a' AND ts_ms BETWEEN 12000 AND 29000" | tr ',' ' ')
share_504=$(awk -v a="$n_504" -v b="$n_fail" 'BEGIN { printf "%.1f", b ? 100 * a / b : 0 }')
recovered=$(q "$OUT" "SELECT round(100.0 * count(*) FILTER (status = 200 AND err_class IS NULL) / count(*), 1)
  FROM runs WHERE node = 'node-a' AND ts_ms BETWEEN 34000 AND 44000")

log "phase 2: a runaway query from the noisy tenant; phase 3: a stalled client"
key=$(jq -r '.tenants[] | select(.class == "noisy") | .key' "$RESULTS/manifest.json")
since=$(date -u +%Y-%m-%dT%H:%M:%S.%NZ)
curl -s -o /dev/null --max-time 3 "$EDGE/v1/databases/default/query" -H "Authorization: Bearer $key" \
  -d '{"sql": "WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 8000000) SELECT count(*) FROM c"}' || true &
stalled_start=$(date +%s.%N)
rm -f "$RESULTS/02-stalled.txt"
( exec 3<>/dev/tcp/127.0.0.1/18001; printf 'GET /health HTTP/1.1\r\nHost: gateway\r\n' >&3
  timeout 9 cat <&3 >/dev/null || true; date +%s.%N > "$RESULTS/02-stalled.txt" ) &
stalled_pid=$!
sleep 7
cpu=$(prom 'rate(process_cpu_seconds_total{instance="node-a"}[3s])')
cancelled=$(prom 'sum(statements_cancelled_total{instance="node-a",reason="deadline"})')
cancel_ms=$(compose logs --no-log-prefix --no-color --since "$since" node-a 2>/dev/null |
  jq -R -r -n 'first(inputs | fromjson? | select(.msg == "request" and .error_code == "statement_deadline") | .duration_ms) // "NaN"')
wait "$stalled_pid" || true
held=$(awk -v a="$(cat "$RESULTS/02-stalled.txt")" -v b="$stalled_start" 'BEGIN { printf "%.1f", a - b }')
late_ms=$(awk -v d="$cancel_ms" 'BEGIN { if (d == "NaN") print "NaN"; else printf "%.1f", d - 2000 }')

echo
echo "Phase 1: node-a accepts and never answers (t=10-30 s)"
printf '  max gateway in-flight during the outage     %s\n' "$inflight"
printf '  max gateway goroutines during the outage    %s\n' "$goroutines"
printf '  gateway in-flight 6 s after the toxic left  %s\n' "$inflight_after"
printf '  node-a p99 during the outage                %s s\n' "$p99_out"
printf '  node-a failures the gateway answered fast    %s of %s (%s %%) (504, or 503 overloaded)\n' "$n_504" "$n_fail" "$share_504"
printf '  node-a success rate t=34-44 s               %s %%\n' "$recovered"
echo "Phase 2: runaway WITH RECURSIVE (~30 s of work), client gives up at 3 s"
printf '  node-a CPU 4-7 s after submit               %s cores\n' "$cpu"
printf '  statements_cancelled_total{reason=deadline} %s\n' "$cancelled"
printf '  node request duration when cancelled        %s ms (%s ms past the 2 s deadline)\n' "$cancel_ms" "$late_ms"
echo "Phase 3: a client sends half a request and stalls"
printf '  gateway closed the stalled connection after %s s (gave up waiting at 9 s)\n' "$held"

check max_inflight "$inflight" "<=" 300
check max_goroutines "$goroutines" "<=" 1000
check outage_p99_s "$p99_out" "<" 3.0
check outage_504_pct "$share_504" ">=" 90
check runaway_cpu_cores "$cpu" "<=" 0.15
check statements_cancelled "$cancelled" ">=" 1
check cancel_late_ms "$late_ms" "<=" 50
check stalled_client_held_s "$held" "<=" 6
verdict
