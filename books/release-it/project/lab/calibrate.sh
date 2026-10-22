#!/usr/bin/env bash
# make calibrate — measure one gateway's and one node's capacity and write
# lab/capacity.env. Capacity is the highest offered rate (open loop, steady
# mix) at which >= 99.5 % of requests succeed and p99 <= 100 ms. Scenario
# rates in the verify scripts are fractions of FLEET_CAPACITY.
source "$(dirname "$0")/lib.sh"

reset_stack

ramp() { # label loadgen-args... ; prints the capacity
  local label=$1 best=0; shift
  for r in $RATES; do
    local s ok sent p99
    s=$(loadgen "$@" -scenario steady -rate "$r" -duration 8s -out /results/calibrate.csv)
    ok=$(jq .ok <<<"$s"); sent=$(jq .sent <<<"$s"); p99=$(jq .p99_ms <<<"$s")
    log "$label rate $r: ok $ok/$sent p99 ${p99} ms"
    if cmp "$ok" ">=" "$(awk -v s="$sent" 'BEGIN{print s*0.995}')" && cmp "$p99" "<=" 100; then best=$r; else break; fi
    sleep 3
  done
  echo "$best"
}

RATES="100 150 200 250 300 350 400 450 500 550 600 700 800"
gw=$(ramp gateway-a -target http://gateway-a:8080)
reset_stack
node=$(ramp node-a -node node-a)
fleet=$(( 3 * gw < 2 * node ? 3 * gw : 2 * node ))
cat > "$LAB_DIR/capacity.env" <<EOT
# Written by make calibrate on $(date -u +%Y-%m-%d) at $(git -C "$PROJECT_DIR" describe --tags --always 2>/dev/null || echo untracked).
# Offered rate (req/s, steady mix) one instance sustains with >= 99.5 % success and p99 <= 100 ms.
GATEWAY_CAPACITY=$gw
NODE_CAPACITY=$node
# min(3 gateways, 2 nodes): the rate the fleet as a whole sustains.
FLEET_CAPACITY=$fleet
EOT
cat "$LAB_DIR/capacity.env"
