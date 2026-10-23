#!/usr/bin/env bash
# E03 rerun of H20/H21 on fresh single-purpose queues (FIFO interference in p10.sh). Source env first.
source /tmp/explore-it-experiment-03/books/explore-it/exploration/experiments/03-vary-it/tools/e03env.sh
cd $PROJ
J() { python3 -c 'import json,sys;d=json.load(sys.stdin);print(eval(sys.argv[1]))' "$1"; }
ts; w=6
for q in e03-h20 e03-h21; do api POST /v1/queues -d "{\"name\":\"$q\"}" | cut -c1-20; done
echo "== H20 ack case-variant"; id=$(curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e03-h20/jobs -d '{"kind":"echo","payload":{"t":"h20b"}}' | J 'd["id"]')
L=$(curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e03-h20/lease -d "{\"worker_id\":$w,\"max_jobs\":1}"); echo "leased id=$(echo "$L" | J 'd["jobs"][0]["id"]') (enqueued $id)"; tok=$(echo "$L" | J 'd["jobs"][0]["lease_token"]')
api POST /v1/jobs/$id/ack -d "{\"LEASE_TOKEN\":\"$tok\",\"RESULT\":{\"ok\":1}}"; sql "SELECT id,state,result FROM jobs WHERE id=$id" | sed -n 2p
echo "== H21 raw lease bytes"; id=$(curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e03-h21/jobs -d '{"kind":"echo","payload":{"a":1.50,"b":12345678901234567890.123456789,"c":9007199254740993}}' | J 'd["id"]'); echo "enqueued $id"
curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e03-h21/lease -d "{\"worker_id\":$w,\"max_jobs\":1}" | grep -o '"id":[0-9]*\|"payload":{[^}]*}'
ts
