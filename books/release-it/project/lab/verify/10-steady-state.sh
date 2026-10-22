#!/usr/bin/env bash
# Lesson 10 — steady state. The node data volumes are 48 MB tmpfs (a Compose
# override) and the whale is left out of the seed. Steady load at 60 % with
# half the requests writes, plus cold reads and a new database (created, then
# read once) every second, for 90 s. Every 30 s the script samples what grows:
# WAL bytes, open handles, the usage ledger, the volumes. Does anything grow
# without bound?
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
export LF_COMPOSE_EXTRA="-f $PROJECT_DIR/deploy/compose.tmpfs.yml"
SEED_ARGS=(-whale=false)
RATE=$((FLEET_CAPACITY * 6 / 10))
OUT=10-steady.csv
S=$RESULTS/10-samples.txt

reset_stack
migrate_online
used_mb() { compose exec -T "$1" df -k /data 2>/dev/null | awk 'NR == 2 { printf "%.1f", $3 / 1024 }'; }
sample() {
  printf '%s %s %s %s %s %s %s\n' "$1" \
    "$(prom 'sum(sqlite_wal_bytes_total{instance="node-a"}) / 1048576')" \
    "$(prom 'sum(sqlite_wal_bytes_total{instance="node-b"}) / 1048576')" \
    "$(prom 'sum(sqlite_open_databases{instance="node-a"})')" \
    "$(prom 'sum(sqlite_open_databases{instance="node-b"})')" \
    "$(compose exec -T postgres psql -U litefleet -d litefleet -qAt -c 'SELECT count(*) FROM usage_ledger')" \
    "$(used_mb node-a)" >> "$S"
}
rm -f "$S"
log "steady at $RATE req/s (half writes) + cold reads + a new database per second, 90 s; volumes 48 MB"
start_load -scenario steady -rate "$RATE" -write-frac 0.5 -cold-rate 5 -create-rate 1 -duration 90s -out "/results/$OUT"
for t in 1 30 60 90; do at "$t"; sample "$t"; done
wait_load >&2
report_table "$OUT" | tail -n 12
# A full volume surfaces as SQLITE_FULL or, on a WAL append, as a disk I/O
# error (ENOSPC).
full=$(compose logs --no-log-prefix --no-color node-a node-b 2>/dev/null | grep -cE 'disk is full|disk I/O error' || true)
write_fail=$(q "$OUT" "SELECT count(*) FROM runs WHERE route = 'execute' AND (status >= 500 OR err_class IN ('client_timeout', 'conn'))")
node_full=$(q "$OUT" "SELECT count(*) FROM runs WHERE err_class = 'node_full'")

echo
echo "  t s  WAL node-a MB  WAL node-b MB  handles node-a  handles node-b  ledger rows  node-a volume used MB (of 48)"
awk '{ printf "  %-4s %-14s %-14s %-15s %-15s %-12s %s\n", $1, $2, $3, $4, $5, $6, $7 }' "$S"
read -r w30 w90 h60 h90 l60 l90 v30 v90 < <(awk '$1 == 30 { w30 = $2; v30 = $7 } $1 == 60 { h60 = $4; l60 = $6 } $1 == 90 { w90 = $2; h90 = $4; l90 = $6; v90 = $7 }
  END { print w30, w90, h60, h90, l60, l90, v30, v90 }' "$S")
echo
echo "  growth per minute (last minute) and time to exhaustion at that rate:"
awk -v w30="$w30" -v w90="$w90" -v h60="$h60" -v h90="$h90" -v l60="$l60" -v l90="$l90" -v v30="$v30" -v v90="$v90" 'function eta(left, perMin) {
    if (perMin <= 0) return "never"; m = left / perMin
    if (m < 120) return sprintf("%.0f minutes", m); if (m < 2880) return sprintf("%.1f hours", m / 60); return sprintf("%.1f days", m / 1440) }
  BEGIN {
    printf "    node-a volume  %+8.1f MB/min    full (48 MB) in %s\n", (v90 - v30) / 1, eta(48 - v90, (v90 - v30) / 1)
    printf "    WAL node-a     %+8.1f MB/min\n", (w90 - w30) / 1
    printf "    handles node-a %+8.1f /min      handle budget (150) in %s; nofile (1024, 3 fds each) in %s\n", (h90 - h60) * 2, eta(150 - h90, (h90 - h60) * 2), eta((1024 - 3 * h90) / 3, (h90 - h60) * 2)
    printf "    ledger         %+8.0f rows/min  (%.0f rows now)\n", (l90 - l60) * 2, l90
  }'
printf '  writes failed: %s (node log lines saying the disk is full or failing: %s); requests refused node_full: %s\n' "$write_fail" "$full" "$node_full"

check wal_node_a_mb "$w90" "<=" 8
check handles_growth_last_30s "$(awk -v a="$h90" -v b="$h60" 'BEGIN { print a - b }')" "<=" 0
check ledger_rows_growth_last_30s "$(awk -v a="$l90" -v b="$l60" 'BEGIN { print a - b }')" "<=" 0
check disk_full_errors "$full" "==" 0
verdict
