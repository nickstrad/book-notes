#!/usr/bin/env bash
# P6 (H-E06-07): NUL (\u0000) in payload, result and error text on their way into storage.
source "$(dirname "$0")/env.sh"
date -u +%T
W=4
echo '# enqueue payload with \u0000'; k acme-key POST /v1/queues/reports/jobs '{"kind":"manual","payload":{"s":"a\u0000b"}}'
echo '# kind with \u0000'; k acme-key POST /v1/queues/reports/jobs '{"kind":"man\u0000ual","payload":{}}'
J1=$(k acme-key POST /v1/queues/reports/jobs '{"kind":"manual","payload":{"case":"ack-nul"}}' | head -1 | jq -r .id)
J2=$(k acme-key POST /v1/queues/reports/jobs '{"kind":"manual","payload":{"case":"nack-nul"}}' | head -1 | jq -r .id)
L=$(curl -s -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -d "{\"worker_id\":$W,\"max_jobs\":10}" $QUEUELET_URL/v1/queues/reports/lease)
echo "$L" | jq -c '.jobs[]|[.id,.payload]'
tok() { echo "$L" | jq -r ".jobs[]|select(.id==$1)|.lease_token"; }
for j in $(echo "$L" | jq -r '.jobs[].id'); do [ "$j" = "$J1" ] || [ "$j" = "$J2" ] || k acme-key POST /v1/jobs/$j/ack "{\"lease_token\":\"$(tok $j)\",\"result\":{\"cleanup\":true}}" >/dev/null; done
echo "# ack $J1 with result containing \\u0000"; k acme-key POST /v1/jobs/$J1/ack "{\"lease_token\":\"$(tok $J1)\",\"result\":{\"s\":\"a\\u0000b\"}}"
sql "select id,state,lease_token is not null has_token,attempts from jobs where id=$J1"
echo "# retry the same ack with a clean result"; k acme-key POST /v1/jobs/$J1/ack "{\"lease_token\":\"$(tok $J1)\",\"result\":{\"s\":\"ab\"}}"
echo "# nack $J2 with error containing \\u0000"; k acme-key POST /v1/jobs/$J2/nack "{\"lease_token\":\"$(tok $J2)\",\"error\":\"bad\\u0000byte\"}"
sql "select id,state,lease_token is not null has_token,attempts,last_error from jobs where id=$J2"
echo "# nack again with clean error"; k acme-key POST /v1/jobs/$J2/nack "{\"lease_token\":\"$(tok $J2)\",\"error\":\"bad byte\",\"retry\":false}"
sql "select id,state,attempts,last_error from jobs where id in ($J1,$J2)"
