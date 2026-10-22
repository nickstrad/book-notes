#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p results
compose=(docker compose -f deploy/compose.yml)
"${compose[@]}" stop worker-a worker-b >/dev/null
"${compose[@]}" stop api >/dev/null
"${compose[@]}" exec -T postgres psql -U queuelet -d queuelet -v ON_ERROR_STOP=1 -c 'TRUNCATE queues,jobs,attempts,workers,schedules,webhook_outbox RESTART IDENTITY;' >/dev/null
"${compose[@]}" start api >/dev/null
bash lab/wait.sh
python3 lab/seed/seed.py
curl -fsS -X POST -H 'Content-Type: application/json' -d '{"mode":"ok"}' "${QUEUELET_HOOK_URL:-http://localhost:18201}/admin/mode" >/dev/null
curl -fsS -X DELETE "${QUEUELET_HOOK_URL:-http://localhost:18201}/admin/records" >/dev/null
"${compose[@]}" start worker-a worker-b >/dev/null
