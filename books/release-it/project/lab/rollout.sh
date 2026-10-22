#!/usr/bin/env bash
# make rollout — restart the fleet one instance at a time, the way a deploy
# would: gateways a -> b -> c, then nodes a -> b, each time waiting until the
# restarted instance's /health answers 200 before touching the next.
source "$(dirname "$0")/lib.sh"
declare -A port=([gateway-a]=18001 [gateway-b]=18002 [gateway-c]=18003 [node-a]=18071 [node-b]=18072)
for svc in gateway-a gateway-b gateway-c node-a node-b; do
  t0=$(date +%s.%N)
  compose restart -t 15 "$svc" >/dev/null 2>&1
  wait_http "http://127.0.0.1:${port[$svc]}/health" 120
  log "rollout: $svc back after $(awk -v a="$(date +%s.%N)" -v b="$t0" 'BEGIN { printf "%.1f", a - b }') s"
  sleep 2
done
