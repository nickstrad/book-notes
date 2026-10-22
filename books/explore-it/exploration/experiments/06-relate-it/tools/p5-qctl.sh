#!/usr/bin/env bash
# P5 (H-E06-14/15): qctl (binary inside the e06 api container) as a data path.
source "$(dirname "$0")/env.sh"
date -u +%T
q() { docker exec -e QUEUELET_URL=http://localhost:8200 -e QUEUELET_KEY=acme-key queuelet-e06-api-1 qctl "$@"; }
echo '# qctl job get 311 (stored big 9007199254740993)'; q job get 311 | grep -E '"big"|"dec"'
echo '# qctl enqueue with big payload'; q enqueue reports --kind manual --payload '{"big":9007199254740993,"dec":1.10}'
ID=$(sql "select max(id) from jobs where kind='manual'" | tail -1); sql "select id,payload::text from jobs where id=$ID"
echo '# same through curl for control'; k acme-key POST /v1/queues/reports/jobs '{"kind":"manual","payload":{"big":9007199254740993,"dec":1.10}}'
ID2=$(sql "select max(id) from jobs where kind='manual'" | tail -1); sql "select id,payload::text from jobs where id=$ID2"
