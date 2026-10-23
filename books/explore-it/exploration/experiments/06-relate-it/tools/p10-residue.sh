#!/usr/bin/env bash
# P10 (H-E06-11/18/19): what remains of a job after purge, DELETE and retention.
source "$(dirname "$0")/env.sh"
date -u +%T
W=4; id() { head -1 | jq -r .id; }
lease() { curl -s -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' -d "{\"worker_id\":$W,\"max_jobs\":1}" $QUEUELET_URL/v1/queues/e6-res/lease | jq -r '.jobs[0]|"\(.id) \(.lease_token)"'; }
k acme-key POST /v1/queues '{"name":"e6-res","max_attempts":2,"backoff":{"kind":"linear","base_seconds":0}}' >/dev/null
echo '## purge'; P=$(k acme-key POST /v1/queues/e6-res/jobs '{"kind":"manual","payload":{"case":"purge"}}' | id)
read J T < <(lease); k acme-key POST /v1/jobs/$J/nack "{\"lease_token\":\"$T\",\"error\":\"secret-ish detail\",\"delay_seconds\":3600}" >/dev/null
sql "select id,state,attempts from jobs where id=$P"; k acme-key POST /v1/queues/e6-res/purge
k acme-key GET /v1/jobs/$P; k acme-key GET /v1/jobs/$P/attempts; sql "select job_id,n,outcome,error from attempts where job_id=$P"; sql "select event from job_events where job_id=$P order by id"
echo '## DELETE dead job with 2 attempts, idempotency key K1'
D=$(k acme-key POST /v1/queues/e6-res/jobs '{"kind":"manual","payload":{"case":"del"},"idempotency_key":"K1"}' | id)
for n in 1 2; do read J T < <(lease); k acme-key POST /v1/jobs/$J/nack "{\"lease_token\":\"$T\",\"error\":\"fail $n\",\"delay_seconds\":0}" >/dev/null; sleep 1.2; done
sql "select id,state,attempts from jobs where id=$D"; k acme-key DELETE /v1/jobs/$D
sql "select job_id,n,outcome,error from attempts where job_id=$D order by n"
echo '## same idempotency key after delete'; k acme-key POST /v1/queues/e6-res/jobs '{"kind":"manual","payload":{"case":"again"},"idempotency_key":"K1"}'
echo '## retention removal of a succeeded job'
read J T < <(lease); k acme-key POST /v1/jobs/$J/ack "{\"lease_token\":\"$T\",\"result\":{\"ok\":1}}" >/dev/null; echo "acked $J"
k acme-key PATCH /v1/queues/e6-res '{"retention_hours":0}' >/dev/null; sleep 2.5
k acme-key GET /v1/jobs/$J | head -1; sql "select job_id,n,outcome from attempts where job_id=$J"
echo '## residue totals'; sql "select count(*) orphan_attempts, count(distinct job_id) jobs from attempts a where not exists(select 1 from jobs j where j.id=a.job_id)"
sql "select count(*) outbox_rows_for_missing_jobs, count(*) filter (where delivered_at is not null) delivered from webhook_outbox o where not exists(select 1 from jobs j where j.id=o.job_id)"
