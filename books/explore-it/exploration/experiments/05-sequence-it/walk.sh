#!/usr/bin/env bash
# E05 random walk: pick a noun, pick a verb, act, observe, check the rules (rules.md R3/R5/R34/R35/R37/R38 as I1-I13).
# Usage: [BIAS=1] bash walk.sh SEED STEPS   (needs the e05 stack; source of env: tools/e05env.sh)
# Replay: same SEED and STEPS against a stack where the queue name w<SEED> does not exist yet; bash $RANDOM is
# deterministic for a given seed, but responses (ids, tokens) depend on stack state, so the log records every op.
# Output: one line per step "step|op|args|http|body(trimmed)" plus "VIOLATION ..." lines on first sight of (inv,id).
# (no set -u: empty associative arrays in the summary)
source "$(dirname "$0")/tools/e05env.sh"
SEED=${1:?seed}; STEPS=${2:-150}; RANDOM=$SEED
Q=w$SEED; MAX=2
declare -A TOK GRANTS LASTATT CANCELLED SEEN
JOBS=(); WORKERS=(); STALE=()
call() { local out; out=$(api "$@"); CODE=${out##*[}; CODE=${CODE%]}; BODY=${out% \[*}; }
pick() { local -n arr=$1; [ ${#arr[@]} -gt 0 ] || return 1; echo "${arr[$((RANDOM % ${#arr[@]}))]}"; }
# v2 (BIAS=1): holder verbs pick, 3 times in 4, among jobs the walker has held a lease token for.
pickh() { if [ "${BIAS:-0}" = 1 ] && [ ${#TOK[@]} -gt 0 ] && [ $((RANDOM % 4)) -ne 0 ]; then local ks=("${!TOK[@]}"); echo "${ks[$((RANDOM % ${#ks[@]}))]}"; else pick JOBS; fi; }
log() { echo "$STEP|$1|$2|$CODE|$(echo "$BODY" | tr -d '\n' | cut -c1-160)"; }
violation() { local key="$1:$2"; [ -n "${SEEN[$key]:-}" ] && return; SEEN[$key]=1; echo "VIOLATION step=$STEP inv=$1 job=$2 $3"; }
check() { # $1 = "tick" when the loop has had time to run after a clock change
  local now; now=$(adm POST /admin/clock -d '{"action":"advance","seconds":0}' | sed 's/ \[.*//' | jq -r .now)
  local extra=""
  [ "${1:-}" = tick ] && extra="UNION ALL SELECT 'I12',id,state FROM j WHERE state IN ('scheduled','retry_wait') AND run_at <= '$now'
    UNION ALL SELECT 'I13',id,state FROM j WHERE state='leased' AND lease_expires_at <= '$now'"
  while IFS='|' read -r inv id st; do
    inv=${inv// /}; id=${id// /}; st=${st// /}
    [ -n "$inv" ] && violation "$inv" "$id" "state=$st"
  done < <(PGPASSWORD=explorer psql -h localhost -p 15493 -U explorer -d queuelet -X -A -t -F'|' -c "
    WITH j AS (SELECT * FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='$Q' AND namespace_id=1))
    SELECT 'I1',id,state FROM j WHERE state NOT IN ('scheduled','queued','leased','retry_wait','succeeded','dead','cancelled')
    UNION ALL SELECT 'I2',id,state FROM j WHERE (finished_at IS NOT NULL) <> (state IN ('succeeded','dead','cancelled'))
    UNION ALL SELECT 'I3',id,state FROM j WHERE attempts > max_attempts
    UNION ALL SELECT 'I5',a.job_id,'' FROM attempts a JOIN j ON j.id=a.job_id GROUP BY a.job_id HAVING count(*)<>max(a.n) OR count(DISTINCT a.n)<>count(*)
    UNION ALL SELECT CASE WHEN state='cancelled' THEN 'I7c' ELSE 'I7' END,id,state FROM j WHERE (state='leased') <> (lease_token IS NOT NULL AND leased_by IS NOT NULL AND lease_expires_at IS NOT NULL)
    UNION ALL SELECT 'I9',id,state FROM j WHERE state IN ('scheduled','retry_wait') AND run_at < event_at
    $extra")
  # I10 stats equal a direct count
  local s c; s=$(api GET /v1/queues/$Q/stats | sed 's/ \[.*//' | jq -c '{scheduled,queued,leased,retry_wait,succeeded,dead,cancelled}')
  c=$(PGPASSWORD=explorer psql -h localhost -p 15493 -U explorer -d queuelet -X -A -t -c "SELECT json_build_object('scheduled',count(*) FILTER (WHERE state='scheduled'),'queued',count(*) FILTER (WHERE state='queued'),'leased',count(*) FILTER (WHERE state='leased'),'retry_wait',count(*) FILTER (WHERE state='retry_wait'),'succeeded',count(*) FILTER (WHERE state='succeeded'),'dead',count(*) FILTER (WHERE state='dead'),'cancelled',count(*) FILTER (WHERE state='cancelled')) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='$Q' AND namespace_id=1)" | jq -c .)
  [ "$s" = "$c" ] || violation I10 q "stats=$s count=$c"
}
STEP=0; CODE=-; BODY=""
call POST /v1/queues -d "{\"name\":\"$Q\",\"lease_seconds\":20,\"max_attempts\":$MAX,\"backoff\":{\"kind\":\"linear\",\"base_seconds\":5}}"; log setup "queue $Q"
for i in 1 2; do call POST /v1/workers -d "{\"name\":\"$Q-w$i\",\"kinds\":[\"k\"]}"; WORKERS+=("$(echo "$BODY" | jq -r .worker_id)"); log setup worker; done
for STEP in $(seq 1 "$STEPS"); do
  r=$((RANDOM % 100)); tick=""
  if   [ $r -lt 12 ]; then call POST /v1/queues/$Q/jobs -d '{"kind":"k"}'; JOBS+=("$(echo "$BODY" | jq -r .id)"); log job.enqueue now
  elif [ $r -lt 17 ]; then call POST /v1/queues/$Q/jobs -d '{"kind":"k","delay_seconds":15}'; JOBS+=("$(echo "$BODY" | jq -r .id)"); log job.enqueue delay15
  elif [ $r -lt 32 ]; then
    if [ "${BIAS:-0}" = 1 ] && [ ${#WORKERS[@]} -eq 0 ]; then # v2: never starve of workers
      call POST /v1/workers -d "{\"name\":\"$Q-a$STEP\",\"kinds\":[\"k\"]}"; WORKERS+=("$(echo "$BODY" | jq -r .worker_id)"); log worker.create auto; fi
    w=$(pick WORKERS) || { w=999999; }; n=$((1 + RANDOM % 3))
    call POST /v1/queues/$Q/lease -d "{\"worker_id\":$w,\"max_jobs\":$n}"; log lease.lease "w=$w n=$n"
    if [ "$CODE" = 200 ]; then
      while read -r id att tok; do [ -z "$id" ] && continue
        [ -n "${TOK[$id]:-}" ] && STALE+=("$id:${TOK[$id]}"); TOK[$id]=$tok
        GRANTS[$id]=$(( ${GRANTS[$id]:-0} + 1 ))
        [ -n "${CANCELLED[$id]:-}" ] && violation I6 "$id" "cancelled job leased"
        [ "${GRANTS[$id]}" -gt $MAX ] && violation I4 "$id" "grant ${GRANTS[$id]} > max $MAX (attempt=$att)"
        [ "$att" -le "${LASTATT[$id]:-0}" ] && violation I8 "$id" "attempt $att <= previous ${LASTATT[$id]}"
        LASTATT[$id]=$att
      done < <(echo "$BODY" | jq -r '.jobs[] | "\(.id) \(.attempt) \(.lease_token)"')
    fi
  elif [ $r -lt 42 ]; then j=$(pickh) || continue; call POST /v1/jobs/$j/ack -d "{\"lease_token\":\"${TOK[$j]:-none}\",\"result\":{\"s\":$STEP}}"; log job.ack "j=$j"
  elif [ $r -lt 50 ]; then j=$(pickh) || continue; call POST /v1/jobs/$j/nack -d "{\"lease_token\":\"${TOK[$j]:-none}\",\"error\":\"e$STEP\"}"; log job.nack "j=$j"
  elif [ $r -lt 53 ]; then j=$(pickh) || continue; call POST /v1/jobs/$j/nack -d "{\"lease_token\":\"${TOK[$j]:-none}\",\"error\":\"e$STEP\",\"retry\":false}"; log job.nack "j=$j retry=false"
  elif [ $r -lt 56 ]; then j=$(pickh) || continue; call POST /v1/jobs/$j/heartbeat -d "{\"lease_token\":\"${TOK[$j]:-none}\",\"extend_seconds\":20}"; log job.heartbeat "j=$j"
  elif [ $r -lt 59 ]; then s=$(pick STALE) || continue; j=${s%%:*}; call POST /v1/jobs/$j/ack -d "{\"lease_token\":\"${s#*:}\",\"result\":0}"; log job.ack "j=$j STALE token"
    [ "$CODE" = 200 ] && violation R2 "$j" "stale token ack accepted"
  elif [ $r -lt 65 ]; then j=$(pick JOBS) || continue; call POST /v1/jobs/$j/cancel; log job.cancel "j=$j"; [ "$CODE" = 200 ] && CANCELLED[$j]=1
  elif [ $r -lt 71 ]; then j=$(pick JOBS) || continue; call POST /v1/jobs/$j/retry; log job.retry "j=$j"
    [ "$CODE" = 200 ] && { GRANTS[$j]=0; LASTATT[$j]=0; }
  elif [ $r -lt 74 ]; then j=$(pick JOBS) || continue; call DELETE /v1/jobs/$j; log job.delete "j=$j"
  elif [ $r -lt 77 ]; then call POST /v1/queues/$Q/pause; log queue.pause ""
  elif [ $r -lt 81 ]; then call POST /v1/queues/$Q/resume; log queue.resume ""
  elif [ $r -lt 83 ]; then call POST /v1/queues/$Q/purge; log queue.purge ""
  elif [ $r -lt 86 ]; then call POST /v1/workers -d "{\"name\":\"$Q-s$STEP\",\"kinds\":[\"k\"]}"; WORKERS+=("$(echo "$BODY" | jq -r .worker_id)"); log worker.create ""
  elif [ $r -lt 90 ]; then [ ${#WORKERS[@]} -gt 0 ] || continue; k=$((RANDOM % ${#WORKERS[@]})); w=${WORKERS[$k]}
    call DELETE /v1/workers/$w; log worker.delete "w=$w"; [ "$CODE" = 200 ] && { unset 'WORKERS[k]'; WORKERS=("${WORKERS[@]}"); }
  else secs=$(( (RANDOM % 3 == 0) ? 25 : 6 )); out=$(adm POST /admin/clock -d "{\"action\":\"advance\",\"seconds\":$secs}"); CODE=${out##*[}; CODE=${CODE%]}; BODY=${out% \[*}
    log clock.advance "+$secs"; sleep 1.6; tick=tick
  fi
  check $tick
done
echo "SUMMARY bias=${BIAS:-0} seed=$SEED steps=$STEPS jobs=${#JOBS[@]} violations=${#SEEN[@]} keys=$(printf '%s ' "${!SEEN[@]}" | tr ' ' '\n' | cut -d: -f1 | sort | uniq -c | tr '\n' ' ')"
