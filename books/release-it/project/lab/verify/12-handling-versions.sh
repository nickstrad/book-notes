#!/usr/bin/env bash
# Lesson 12 — handling versions. Steady at 50 % with a create every second
# for 100 s while a release goes out the way it should: the migration job
# (and the backfill) runs first, then a rolling restart of gateways a, b, c
# and nodes a, b, then node-b is restarted speaking wire v2 so the fleet runs
# mixed versions. Finally the contract migration is attempted; it must
# refuse while anything still depends on the old shape.
#
# An in-place node restart costs that node's tenants a few seconds: requests
# are refused fast (503 draining, then node_unavailable) or, while the
# container is gone, time out at the gateway's budget. What must hold is that
# nothing in flight is cut off and that the window is short.
#
# REF=release-it/lesson-11 REF_SERVICES=gateway: older gateways with today's
# nodes (the wire-v2 step). REF=release-it/lesson-11: everything older (nodes
# that exit on SIGTERM). make godot replays the in-place ALTER separately.
source "$(dirname "$0")/../lib.sh"
source "$LAB_DIR/capacity.env"
RATE=$((FLEET_CAPACITY / 2))
OUT=12-rollout.csv
ROLL=$RESULTS/12-roll.txt
export ROLL_LOG=$ROLL

reset_stack
rm -f "$ROLL"
since=$(date -u +%Y-%m-%dT%H:%M:%S.%NZ)
log "steady at $RATE req/s + a create per second for 100 s, during a release"
start_load -scenario steady -rate "$RATE" -create-rate 1 -duration 100s -out "/results/$OUT"
at 3
if migrations_job; then
  compose --profile jobs run --rm --no-deps migrate 2>&1 | grep -E 'applied|up to date' | jq -r .msg >&2 || true
  compose --profile jobs run --rm --no-deps migrate backfill >/dev/null 2>&1 & BACKFILL=$!
else
  log "this tree migrates inside control's startup: no migration job to run"
fi
at 8; lab/rollout.sh
at 62
inflight=$(prom 'http_inflight_requests{instance="node-b"}')
t0=$(date +%s.%N)
NODE_B_WIRE_VERSION=2 compose $(profiles) up -d --no-deps --no-build node-b >/dev/null 2>&1
wait_http http://127.0.0.1:18072/health 120
echo "node-b-v2 $t0 $(date +%s.%N) $inflight" >> "$ROLL"
log "node-b restarted speaking wire v2"
wait_load >&2
contract="n/a (no migration job)"
if migrations_job; then
  [[ -n ${BACKFILL:-} ]] && { kill "$BACKFILL" 2>/dev/null || true; }
  contract=$(compose --profile jobs run --rm --no-deps migrate migrate -contract 2>&1 | grep -o 'contract refused[^"(]*\|migrations up to date' | head -n 1 || true)
fi
report_table "$OUT" | tail -n 25

# A failure is any 4xx/5xx or client-side error; a fast refusal is a 429 or
# 503 with Retry-After (draining, node unavailable, shed, pool or lane full). Windows, in the CSV's
# milliseconds: each node restart from its start to
# 3 s after it is healthy; the gateway rollout from its first restart to the
# last gateway being healthy.
read -r gw_from gw_to < <(awk -v t0="$T0" '$1 ~ /^gateway/ { s = ($2 - t0) * 1000; e = ($3 - t0) * 1000; if (min == "" || s < min) min = s; if (e > max) max = e }
  END { printf "%d %d\n", min, max + 1000 }' "$ROLL")
windows=$(awk -v t0="$T0" '$1 ~ /^node/ { printf "(ts_ms BETWEEN %d AND %d) OR ", ($2 - t0) * 1000, ($3 - t0) * 1000 + 3000 }' "$ROLL")
windows="(${windows% OR })"
v2_end=$(awk -v t0="$T0" '$1 == "node-b-v2" { printf "%d", ($3 - t0) * 1000 + 3000 }' "$ROLL")
fail="(status >= 400 OR err_class IS NOT NULL)"
refusal="(status IN (429, 503) AND err_class IN ('draining', 'node_unavailable', 'overloaded', 'pool_timeout', 'node_busy', 'tenant_busy', 'busy', 'control_unavailable'))"
read -r total gw_errors outside outside_ref in_ref in_timeout hard mixed_n mixed_fail decode < <(q "$OUT" "SELECT count(*),
    count(*) FILTER (ts_ms BETWEEN $gw_from AND $gw_to AND $fail AND NOT $refusal),
    count(*) FILTER (NOT $windows AND $fail AND NOT $refusal),
    count(*) FILTER (NOT $windows AND $refusal),
    count(*) FILTER ($windows AND $refusal),
    count(*) FILTER ($windows AND $fail AND NOT $refusal AND (status = 504 OR err_class = 'client_timeout')),
    count(*) FILTER ($windows AND $fail AND NOT $refusal AND NOT (status = 504 OR err_class = 'client_timeout')),
    count(*) FILTER (node = 'node-b' AND route = 'query' AND ts_ms > $v2_end),
    count(*) FILTER (node = 'node-b' AND route = 'query' AND ts_ms > $v2_end AND $fail AND NOT $refusal),
    count(*) FILTER (err_class = 'decode')
  FROM runs WHERE route IN ('query', 'execute')" | tr ',' ' ')
longest=$(for w in $(awk -v t0="$T0" '$1 ~ /^node/ { printf "%d:%d\n", ($2 - t0) * 1000, ($3 - t0) * 1000 + 3000 }' "$ROLL"); do
  q "$OUT" "SELECT coalesce(round((max(ts_ms) - min(ts_ms)) / 1000.0, 1), 0) FROM runs WHERE ts_ms BETWEEN ${w%:*} AND ${w#*:} AND $fail AND route IN ('query', 'execute')"
done | sort -n | tail -n 1)
drains=$(compose logs --no-log-prefix --no-color --since "$since" node-a node-b 2>/dev/null | jq -R -r 'fromjson? | select(.msg | startswith("drained")) | .msg')
if [[ -n $drains ]]; then
  cut_off=$(grep -o '([0-9]* requests still running' <<<"$drains" | grep -o '[0-9]*' | paste -sd+ | bc)
else
  cut_off=$(awk '$1 ~ /^node/ { s += $4 } END { print s + 0 }' "$ROLL") # no drain: whatever was in flight was cut off
fi
refused_pct=$(awk -v a="$outside_ref" -v b="$total" 'BEGIN { printf "%.2f", b ? 100 * a / b : 0 }')
mixed_pct=$(awk -v a="$mixed_fail" -v b="$mixed_n" 'BEGIN { printf "%.1f", b ? 100 * a / b : 0 }')

echo
echo "  restarts (service, seconds from start to healthy, requests in flight before)"
awk -v t0="$T0" '{ printf "    %-10s t=%5.1f s  %4.1f s  %s\n", $1, $2 - t0, $3 - $2, $4 }' "$ROLL"
printf '  failed requests during the gateway rollout                 %s\n' "$gw_errors"
printf '  failed requests outside the node restarts                  %s\n' "$outside"
printf '  refused fast (503 + Retry-After) outside the node restarts %s of %s (%s %%)\n' "$outside_ref" "$total" "$refused_pct"
printf '  in the node restarts: refused fast %s, timed out %s, other failures %s; longest run of failures %s s\n' "$in_ref" "$in_timeout" "$hard" "$longest"
printf '  requests cut off by a node stopping                        %s\n' "$cut_off"
printf '  node-b queries after it speaks wire v2: %s, failed (not refused) %s (%s %%); answers the consumer could not read: %s\n' "$mixed_n" "$mixed_fail" "$mixed_pct" "$decode"
printf '  contract migration                                         %s\n' "$contract"

check gateway_rollout_errors "$gw_errors" "==" 0
check errors_outside_node_restarts "$outside" "==" 0
check longest_restart_failures_s "$longest" "<=" 10
check requests_cut_off "$cut_off" "==" 0
check mixed_fleet_query_fail_pct "$mixed_pct" "==" 0
verdict
