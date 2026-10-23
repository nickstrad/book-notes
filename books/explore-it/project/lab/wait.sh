#!/usr/bin/env bash
set -euo pipefail
for i in {1..60}; do
  if curl -fsS "${QUEUELET_URL:-http://localhost:18200}/health" >/dev/null; then exit 0; fi
  sleep 1
done
exit 1
