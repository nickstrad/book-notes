#!/usr/bin/env bash
# E03 payload/batch probe (bash; zsh echo expands \u escapes). Source env first.
source /tmp/explore-it-experiment-03/books/explore-it/exploration/experiments/03-vary-it/tools/e03env.sh
cd $PROJ
ts; echo "== H12 continued (surrogates, NUL, huge exponent)"
for p in '{"s":"\ud800"}' '{"s":"x\u0000y"}' '{"a":1e200000}'; do
  r=$(curl -sS -X POST -H 'Authorization: Bearer acme-key' -w ' [%{http_code}]' $QUEUELET_URL/v1/queues/e03-pay/jobs -d "{\"kind\":\"echo\",\"payload\":$p}"); printf 'sent %s → %s\n' "$p" "${r:0:200}"
done
printf '{"kind":"echo","payload":{"s":"a\xffb"}}' > $RES/h12-badutf8.json; printf 'raw invalid UTF-8 in payload → %s\n' "$(curl -sS -X POST -H 'Authorization: Bearer acme-key' -w ' [%{http_code}]' $QUEUELET_URL/v1/queues/e03-pay/jobs --data-binary @$RES/h12-badutf8.json)"
w=$(curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/workers -d '{"name":"e03-pay","kinds":["echo"]}' | python3 -c 'import json,sys;print(json.load(sys.stdin)["worker_id"])')
echo "lease payloads (worker view, worker $w):"; curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e03-pay/lease -d "{\"worker_id\":$w,\"max_jobs\":20}" | python3 -c 'import sys,json;[print("  ",j["id"],json.dumps(j["payload"])[:100]) for j in json.load(sys.stdin)["jobs"]]'
echo "== H13 batch 500 x ~4.3KB (>2MiB)"
python3 -c 'import json;print(json.dumps({"jobs":[{"kind":"echo","payload":{"t":"h13","pad":"x"*4300}}]*500}))' > $RES/h13.json; stat -c '%s bytes' $RES/h13.json
api POST /v1/queues/e03-pay/jobs:batch --data-binary @$RES/h13.json
python3 -c 'import json;print(json.dumps({"jobs":[{"kind":"echo","payload":{"t":"h13b","pad":"x"*4100}}]*500}))' > $RES/h13b.json; stat -c '%s bytes' $RES/h13b.json
api POST /v1/queues/e03-pay/jobs:batch --data-binary @$RES/h13b.json | cut -c1-60
echo "single job with 2.2MB payload"; python3 -c 'import json;print(json.dumps({"kind":"echo","payload":{"pad":"x"*2200000}}))' > $RES/h13c.json; api POST /v1/queues/e03-pay/jobs --data-binary @$RES/h13c.json
sql "SELECT payload->>'t',count(*) FROM jobs WHERE payload->>'t' IN ('h13','h13b') GROUP BY 1"
echo "== H14 batch duplicate idempotency key"
api POST /v1/queues/e03-pay/jobs:batch -d '{"jobs":[{"kind":"echo","idempotency_key":"k1","payload":{"v":1}},{"kind":"other","idempotency_key":"k1","payload":{"v":2}}]}'
api POST /v1/queues/e03-pay/jobs -d '{"kind":"third","idempotency_key":"k1","payload":{"v":3}}'
sql "SELECT id,kind,payload FROM jobs WHERE idempotency_key='k1'"
echo "== idempotency_key empty string"; api POST /v1/queues/e03-pay/jobs -d '{"kind":"echo","idempotency_key":"","payload":{"v":"e1"}}'; api POST /v1/queues/e03-pay/jobs -d '{"kind":"echo","idempotency_key":"","payload":{"v":"e2"}}'
echo "== batch shape siblings"; api POST /v1/queues/e03-pay/jobs:batch -d '{"JOBS":[{"Kind":"echo","payload":{"t":"bcase"}}]}'
api POST /v1/queues/e03-pay/jobs:batch -d '{"jobs":[{"kind":"echo","payload":{"t":"btrail"}}]} {"jobs":[]}'
sql "SELECT payload->>'t',count(*) FROM jobs WHERE payload->>'t' IN ('bcase','btrail') GROUP BY 1"
ts
