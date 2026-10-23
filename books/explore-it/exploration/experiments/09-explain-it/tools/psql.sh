#!/usr/bin/env bash
# Read-only SQL against the e09 database, as the explorer role.
source "$(dirname "$0")/env.sh"
docker compose -f "$(dirname "$0")/../../../../project/deploy/compose.yml" exec -T postgres psql -U explorer -d queuelet -P pager=off -c "$1" 2>&1 || \
  docker compose -f "$(dirname "$0")/../../../../project/deploy/compose.yml" exec -T -e PGPASSWORD=explorer postgres psql -h localhost -U explorer -d queuelet -P pager=off -c "$1"
