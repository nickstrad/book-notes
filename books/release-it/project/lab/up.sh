#!/usr/bin/env bash
# Start the stack (images already built) and wait until every service answers.
source "$(dirname "$0")/lib.sh"
if migrations_job; then
  compose up -d --no-build --wait postgres
  compose --profile jobs run --rm --no-deps migrate
fi
compose $(profiles) up -d --no-build
wait_http http://127.0.0.1:18090/health
for p in 18071 18072 18001 18002 18003; do wait_http "http://127.0.0.1:$p/health"; done
log "stack is up: tenants on $EDGE, prometheus on $PROM, toxiproxy on $TOXI"
