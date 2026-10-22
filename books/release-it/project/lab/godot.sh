#!/usr/bin/env bash
# make godot — "Waiting for Godot" (Release It! Ch. 12) on litefleet. The
# obvious fix for the 32-bit counters is one migration that changes their
# type in place, both tables in one transaction:
#
#   ALTER TABLE databases    ALTER COLUMN bytes_total TYPE bigint, ...;
#   ALTER TABLE usage_ledger ALTER COLUMN bytes TYPE bigint, ...;
#
# Applied while an analyst's 30-day report is running, the ALTER waits for its
# exclusive lock behind the report, every query on `databases` (resolve,
# create, heartbeat) queues behind the ALTER, and the queue drains only when
# the 5 M-row ledger has been rewritten. Runs on any tree; the gateways' route
# cache is cut to 5 s for the run so cache misses keep reaching control.
source "$(dirname "$0")/lib.sh"
export ROUTE_CACHE_TTL=5s
OUT=godot.csv
TL=$RESULTS/godot-timeline.txt
psql_db() { compose exec -T postgres psql -U litefleet -d litefleet -qAt "$@"; }

reset_stack
rm -f "$TL"
log "steady at 150 req/s + 2 creates/s for 60 s; an analyst's report loops from t=5 s; the in-place ALTER at t=15 s"
start_load -scenario create -rate 150 -create-rate 2 -cold-rate 5 -duration 60s -out "/results/$OUT"
at 5
( while true; do
    psql_db -c "SELECT d.tenant_id, sum(l.bytes) FROM usage_ledger l JOIN databases d ON d.id = l.database_id GROUP BY 1" >/dev/null 2>&1
  done ) & REPORTS=$!
( while true; do
    echo "$(awk -v a="$(date +%s.%N)" -v b="$T0" 'BEGIN { printf "%.0f", a - b }') $(psql_db -c "SELECT count(*) FROM pg_locks WHERE NOT granted" 2>/dev/null)" >> "$TL"
    sleep 1
  done ) & SAMPLER=$!
at 15
log "applying the in-place migration"
t_start=$(date +%s.%N)
psql_db -c "BEGIN;
  ALTER TABLE databases ALTER COLUMN bytes_total TYPE bigint, ALTER COLUMN rows_total TYPE bigint;
  ALTER TABLE usage_ledger ALTER COLUMN bytes TYPE bigint, ALTER COLUMN rows TYPE bigint;
  COMMIT;" >/dev/null
alter_s=$(awk -v a="$(date +%s.%N)" -v b="$t_start" 'BEGIN { printf "%.1f", a - b }')
log "migration finished after $alter_s s"
wait_load >&2
kill "$REPORTS" "$SAMPLER" 2>/dev/null || true
wait 2>/dev/null || true

report_table "$OUT"
echo
echo "  t s  lock waiters"
awk '{ printf "  %-4s %s\n", $1, $2 }' "$TL" | head -n 60
q "$OUT" "SELECT 'creates during the migration: ' || count(*) || ', failed ' || count(*) FILTER (status >= 500 OR err_class IN ('client_timeout','conn'))
  FROM runs WHERE route = 'create' AND ts_ms BETWEEN 15000 AND (15 + $alter_s) * 1000"
q "$OUT" "SELECT 'cold reads (cache misses) during the migration: ' || count(*) || ', failed ' || count(*) FILTER (status >= 500 OR err_class IN ('client_timeout','conn'))
  FROM runs WHERE tenant_class = 'cold' AND ts_ms BETWEEN 15000 AND (15 + $alter_s) * 1000"
echo "  the migration held every query on databases for about $alter_s s"
echo "  (lesson 12's way: expand with nullable columns, backfill in batches, contract later; and"
echo "   control migrate sets lock_timeout = 3s, so a migration that cannot get its lock gives up)"
