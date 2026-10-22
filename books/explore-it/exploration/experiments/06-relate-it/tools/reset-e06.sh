#!/usr/bin/env bash
# Adapted copy of project/lab/reset.sh (inherited copy unchanged): pins the compose project
# to queuelet-e06 and the e06 ports so it can never touch another stack.
set -euo pipefail
source "$(dirname "$0")/env.sh"
cd "$PROJ"
mkdir -p results
compose=(docker compose -p queuelet-e06 -f deploy/compose.yml)
"${compose[@]}" stop worker-a worker-b >/dev/null
"${compose[@]}" stop api >/dev/null
"${compose[@]}" exec -T postgres psql -U queuelet -d queuelet -v ON_ERROR_STOP=1 -c "DO \$\$ BEGIN IF to_regclass('job_events') IS NOT NULL THEN TRUNCATE job_events RESTART IDENTITY; END IF; END \$\$; TRUNCATE queues,jobs,attempts,workers,schedules,webhook_outbox RESTART IDENTITY;" >/dev/null
"${compose[@]}" start api >/dev/null
bash lab/wait.sh
python3 lab/seed/seed.py
curl -fsS -X POST -H 'Content-Type: application/json' -d '{"mode":"ok"}' "$QUEUELET_HOOK_URL/admin/mode" >/dev/null
curl -fsS -X DELETE "$QUEUELET_HOOK_URL/admin/records" >/dev/null
"${compose[@]}" start worker-a worker-b >/dev/null
