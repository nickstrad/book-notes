#!/usr/bin/env bash
# Hand-driven chaos through the Toxiproxy API (the verify scripts use the same
# calls through lib.sh).
#   lab/toxics.sh list
#   lab/toxics.sh latency  <proxy> <ms> [toxicity]     e.g. latency node-a 800
#   lab/toxics.sh timeout  <proxy>                     accept, never answer
#   lab/toxics.sh bandwidth <proxy> <KB/s> [upstream|downstream]
#   lab/toxics.sh down|up  <proxy>
#   lab/toxics.sh reset                                remove every toxic, enable every proxy
# Proxies: control, node-a, node-b, blobstore, metering.
source "$(dirname "$0")/lib.sh"
cmd=${1:-list}; shift || true
case $cmd in
  list) curl -fs "$TOXI/proxies" | jq -r 'to_entries[] | "\(.key)\tenabled=\(.value.enabled)\ttoxics=\([.value.toxics[] | "\(.name)(\(.type) \(.attributes|tostring))"] | join(","))"' ;;
  latency) toxic "$1" "latency" latency downstream "${3:-1.0}" "{\"latency\":$2}" ;;
  timeout) toxic "$1" "timeout" timeout downstream 1.0 '{"timeout":0}' ;;
  bandwidth) toxic "$1" "bandwidth" bandwidth "${3:-upstream}" 1.0 "{\"rate\":$2}" ;;
  down) curl -fs -XPOST "$TOXI/proxies/$1" -d '{"enabled":false}' >/dev/null ;;
  up) curl -fs -XPOST "$TOXI/proxies/$1" -d '{"enabled":true}' >/dev/null ;;
  reset) curl -fs -XPOST "$TOXI/reset" >/dev/null ;;
  *) sed -n '2,10p' "$0"; exit 2 ;;
esac
