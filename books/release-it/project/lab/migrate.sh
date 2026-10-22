#!/usr/bin/env bash
# make migrate — apply the online migrations (migrations/online/*.sql, in
# order) to the running registry. Each must be safe to run while the platform
# serves (CONCURRENTLY, IF NOT EXISTS) and safe to run twice.
source "$(dirname "$0")/lib.sh"
migrate_online
