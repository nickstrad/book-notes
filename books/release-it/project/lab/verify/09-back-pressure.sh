#!/usr/bin/env bash
# Lesson 09 — back pressure. Steady at 50 % for 60 s. The metering partner
# answers in 1 s (latency toxic) from t=1 s to t=30 s, and node-a is
# restarted (docker compose restart: SIGTERM, then a fresh process) at
# t=20 s. Afterwards, did metering receive one event for every statement the
# nodes ran, exactly once?
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY / 2))
OUT=09-back-pressure.csv
METERING=http://127.0.0.1:18082

reset_stack
log "steady at $RATE req/s for 60 s; metering slow t=1-30 s; restart node-a at t=20 s"
start_load -scenario steady -rate "$RATE" -duration 60s -out "/results/$OUT"
at 1; toxic metering slow latency downstream 1.0 '{"latency":1000}'
at 19; depth_a=$(prom 'max_over_time(usage_queue_depth{instance="node-a"}[18s])')
age_a=$(prom 'max_over_time(usage_queue_oldest_age_seconds{instance="node-a"}[18s])')
at 20; compose restart node-a >/dev/null 2>&1; log "restarted node-a"
at 30
depth=$(prom 'max(max_over_time(usage_queue_depth[29s]))')
age=$(prom 'max(max_over_time(usage_queue_oldest_age_seconds[29s]))')
untoxic_busy metering slow
wait_load >&2
report_table "$OUT"
read -r p99_before p99_during < <(q "$OUT" "SELECT round(quantile_cont(latency_ms, 0.99) FILTER (ts_ms >= 32000)),
  round(quantile_cont(latency_ms, 0.99) FILTER (ts_ms BETWEEN 2000 AND 18000)) FROM runs WHERE status = 200" | tr ',' ' ')

# Every statement a node answered 200 is one usage event (the load sends one
# statement per execute). Count them from the nodes' own request logs, across
# node-a's restart, and wait up to 35 s for metering to catch up.
executed() {
  compose logs --no-log-prefix --no-color node-a node-b 2>/dev/null |
    jq -R -r 'fromjson? | select(.msg == "request" and .status == 200 and (.route | test("/(query|execute)$"))) | 1' | wc -l
}
want=$(executed)
for _ in $(seq 35); do
  got=$(curl -s "$METERING/stats" | jq .received)
  ((got >= want)) && break
  sleep 1
done
dups=$(curl -s "$METERING/stats" | jq .duplicates)
missing=$((want - got))
outbox=$(lab/audit.sh outbox)

echo
printf '  usage queue: node-a max depth before the restart %s, oldest event %s s\n' "$depth_a" "$age_a"
printf '  usage queue, both nodes, t=1-30 s: max depth %s, oldest event %s s\n' "$depth" "$age"
printf '  p99 of successful statements: metering slow %s ms, metering healthy again %s ms\n' "$p99_during" "$p99_before"
printf '  statements the nodes ran %s; usage events metering received %s (missing %s); duplicates dropped %s\n' "$want" "$got" "$missing" "$dups"
printf '  events left in the outboxes                       %s\n' "$outbox"

check max_queue_depth "$depth" "<=" 1024
check oldest_event_age_s "$age" "<=" 10
check events_missing "${missing#-}" "<=" 5
check outbox_rows_left "$outbox" "==" 0
verdict
