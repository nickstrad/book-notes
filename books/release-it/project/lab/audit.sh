#!/usr/bin/env bash
# make audit — DuckDB over every tenant file on both node volumes (mounted
# read-only), plus each node's own _node.db.
#   audit.sh            print the audit tables
#   audit.sh dups       print one number: rows in tenants' notes tables whose
#                       seq appears more than once (a write applied twice)
#   audit.sh outbox     print one number: unshipped usage events in outboxes
#   audit.sh wal        print one number: total WAL bytes on both nodes
source "$(dirname "$0")/lib.sh"
compose run --rm --no-deps -T --entrypoint bash report /lab/audit-inner.sh "${1:-all}" 2>/dev/null
