#!/usr/bin/env bash
# usage: watch-outbox.sh JOB_ID SECONDS OUTFILE — samples outbox rows for a job, hooksink records for it, and /health.
source "$(dirname "$0")/e02env.sh"
id=$1; n=$2; out=$3; : > "$out"
for ((i=0;i<n;i++)); do
  {
    echo "=== $(ts)"
    echo "OUTBOX: $(sql "select id||' att='||attempts||' next='||next_at||' deliv='||coalesce(delivered_at::text,'-') from webhook_outbox where job_id=$id" | sed -n '2,$p' | grep -v rows | tr '\n' ' ')"
    echo "HOOK:   $(curl -sS $QUEUELET_HOOK_URL/admin/records | jq -c --argjson id $id '[.records[]|select(.body.job_id==$id)|{status,signature_valid}]|{n:length,last:.[-1]}')"
    echo "HEALTH: $(curl -sS -w ' %{http_code}' $QUEUELET_URL/health)"
  } >> "$out"
  sleep 1
done
