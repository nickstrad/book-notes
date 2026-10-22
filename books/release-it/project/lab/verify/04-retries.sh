#!/usr/bin/env bash
# Lesson 04 — retries. Steady at 60 % for 60 s. From t=10 s to t=30 s every
# answer from node-a is delayed by 1.5-3.5 s (latency 2500 ms, jitter 1000 ms,
# all connections): about half of all attempts outlive the gateway's 2.5 s
# budget, after the node has already run them. Throughout, a hot tenant on
# node-b sends 10 bulk writes/s (1 000 rows each) to one database, so its
# writers contend for the file's write lock. Then DuckDB audits every tenant file for writes that were
# applied twice.
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY * 6 / 10))
OUT=04-retries.csv

reset_stack
# The hot database is an `analytics` database, which the steady mix never
# touches, so every write it receives is the hot tenant's.
read -r hot hot_id < <(jq -r '[.tenants[] | select(.class == "normal") | . as $t | .dbs[] | select(.node == "node-b" and .name == "analytics") | "\($t.id)/\(.name) \(.id)"][0]' "$RESULTS/manifest.json")
log "steady at $RATE req/s for 60 s; hot database $hot on node-b, 10 bulk writes/s"
start_load -scenario steady -rate "$RATE" -duration 60s -focus "$hot" -focus-rate 10 -focus-rows 1000 -out "/results/$OUT"
at 10; toxic node-a slow latency downstream 1.0 '{"latency":2500,"jitter":1000}'
at 28
retries=$(prom 'sum(increase(retry_attempts_total{dep="node-a"}[14s])) or vector(0)')
at 30; untoxic_busy node-a slow
wait_load >&2
at 64
hot_locked=$(prom 'sum(increase(http_requests_total{instance="node-b",route="POST /internal/db/{id}/execute",status="500"}[62s])) or vector(0)')
hot_shed=$(prom 'sum(increase(http_requests_total{instance="node-b",route="POST /internal/db/{id}/execute",status="503"}[62s])) or vector(0)')
busy=$(prom 'sum(increase(sqlite_busy_total[62s])) or vector(0)')
exhausted=$(prom 'sum(increase(retry_budget_exhausted_total[62s])) or vector(0)')

report_table "$OUT"
# Requests node-a received, from its own request log, against what tenants
# sent to node-a in the same 14 s.
iso() { date -u -d "@$(awk -v t="$T0" -v s="$1" 'BEGIN { printf "%.3f", t + s }')" +%Y-%m-%dT%H:%M:%S.%3N; }
from=$(iso 14) to=$(iso 28)
node_recv=$(compose logs --no-log-prefix --no-color node-a 2>/dev/null |
  jq -R -r --arg from "$from" --arg to "$to" 'fromjson? | select(.msg == "request" and (.route | startswith("POST /internal/db/")) and .time[0:23] >= $from and .time[0:23] < $to) | 1' | wc -l)
tenant_reqs=$(q "$OUT" "SELECT count(*) FROM runs WHERE node = 'node-a' AND ts_ms BETWEEN 14000 AND 28000")
hot_recv=$(compose logs --no-log-prefix --no-color node-b 2>/dev/null |
  jq -R -r --arg db "$hot_id" 'fromjson? | select(.msg == "request" and .database_id == $db and (.route | endswith("/execute"))) | 1' | wc -l)
hot_sent=$(q "$OUT" "SELECT count(*) FROM runs WHERE tenant_class = 'target'")
hot_ratio=$(awk -v a="$hot_recv" -v b="$hot_sent" 'BEGIN { printf "%.2f", b ? a / b : 0 }')
ratio=$(awk -v a="$node_recv" -v b="$tenant_reqs" 'BEGIN { printf "%.2f", b ? a / b : 0 }')
recovery=$(q "$OUT" "WITH s AS (SELECT ts_ms // 1000 AS sec, avg(CASE WHEN status = 200 AND err_class IS NULL THEN 1 ELSE 0 END) AS ok
    FROM runs WHERE node = 'node-a' AND ts_ms >= 30000 GROUP BY 1)
  SELECT coalesce(min(sec) - 30 + 1, 99) FROM s WHERE ok >= 0.99 AND sec >= 30")
log "auditing tenant files for duplicate notes.seq"
dups=$(lab/audit.sh dups)

echo
echo "node-a slow t=10-30 s"
printf '  requests node-a received / tenant requests for node-a (t=14-28 s)   %s / %s = %s\n' "$node_recv" "$tenant_reqs" "$ratio"
printf '  retries the gateways made to node-a in that window                 %s\n' "$retries"
printf '  seconds after the fault until node-a answers 99 %% again            %s\n' "$recovery"
printf '  rows applied twice (duplicate notes.seq across all tenant files)   %s\n' "$dups"
echo "hot database on node-b, 10 bulk writes/s"
printf '  write attempts node-b received / writes the tenant sent           %s / %s = %s\n' "$hot_recv" "$hot_sent" "$hot_ratio"
printf '  writes node-b failed with 500 (database is locked)                  %s\n' "$hot_locked"
printf '  writes node-b refused with 503 (pool or busy timeout)               %s\n' "$hot_shed"
printf '  sqlite_busy_total (writers that gave up on the lock)                %s\n' "$busy"
printf '  retry_budget_exhausted_total                                        %s\n' "$exhausted"

check node_a_load_ratio "$ratio" "<=" 1.3
check recovery_s "$recovery" "<=" 3
check duplicate_rows "$dups" "==" 0
check hot_db_load_ratio "$hot_ratio" "<=" 1.05
check writes_failed_locked "$hot_locked" "==" 0
verdict
