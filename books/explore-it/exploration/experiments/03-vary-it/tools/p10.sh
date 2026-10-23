#!/usr/bin/env bash
# E03 sibling probes H17–H21 (bash). Source env first.
source /tmp/explore-it-experiment-03/books/explore-it/exploration/experiments/03-vary-it/tools/e03env.sh
cd $PROJ
J() { python3 -c 'import json,sys;d=json.load(sys.stdin);print(eval(sys.argv[1]))' "$1"; }
ts; api POST /v1/queues -d '{"name":"e03-sib","lease_seconds":30}' | cut -c1-40
w=$(curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/workers -d '{"name":"e03-sib","kinds":["echo"]}' | J 'd["worker_id"]'); echo worker=$w
lease() { curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e03-sib/lease -d "{\"worker_id\":$w,\"max_jobs\":1}"; }
enq() { curl -sS -X POST -H 'Authorization: Bearer acme-key' $QUEUELET_URL/v1/queues/e03-sib/jobs -d "$1" | J 'd["id"]'; }
echo "== H17 nack delay overflow"; id=$(enq '{"kind":"echo","payload":{"t":"h17"}}'); tok=$(lease | J 'd["jobs"][0]["lease_token"]')
api POST /v1/jobs/$id/nack -d "{\"lease_token\":\"$tok\",\"error\":\"h17\",\"delay_seconds\":9223372037}"; sql "SELECT id,state,run_at FROM jobs WHERE id=$id" | sed -n 2p; sleep 1.5; sql "SELECT id,state,run_at FROM jobs WHERE id=$id" | sed -n 2p
tok=$(lease | J 'd["jobs"][0]["lease_token"]'); api POST /v1/jobs/$id/ack -d "{\"lease_token\":\"$tok\",\"result\":{}}" >/dev/null
echo "== H18 heartbeat extend overflow"; id=$(enq '{"kind":"echo","payload":{"t":"h18"}}'); tok=$(lease | J 'd["jobs"][0]["lease_token"]')
sql "SELECT id,state,lease_expires_at FROM jobs WHERE id=$id" | sed -n 2p
api POST /v1/jobs/$id/heartbeat -d "{\"lease_token\":\"$tok\",\"extend_seconds\":9223372037}"; sql "SELECT id,state,lease_expires_at,attempts FROM jobs WHERE id=$id" | sed -n 2p; sleep 1.5
sql "SELECT id,state,lease_expires_at,attempts,last_error FROM jobs WHERE id=$id" | sed -n 2p
echo "== H19 PATCH shape"; api PATCH /v1/queues/e03-sib -d '{"MAX_ATTEMPTS":5} garbage' | grep -o '"max_attempts":[0-9]*\|\[[0-9]*\]'
echo "== H20 ack case-variant"; id=$(enq '{"kind":"echo","payload":{"t":"h20"}}'); tok=$(lease | J 'd["jobs"][0]["lease_token"]')
api POST /v1/jobs/$id/ack -d "{\"LEASE_TOKEN\":\"$tok\",\"RESULT\":{\"ok\":1}}"; sql "SELECT id,state,result FROM jobs WHERE id=$id" | sed -n 2p
echo "== H21 raw lease bytes"; id=$(enq '{"kind":"echo","payload":{"a":1.50,"b":12345678901234567890.123456789,"c":9007199254740993}}')
lease | grep -o '"payload":{[^}]*}'
ts
