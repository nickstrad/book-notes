# lease1.sh QUEUE WORKER_ID [MAX] — lease and print compact job tuples (id attempt token)
source /tmp/explore-it-experiment-04/books/explore-it/exploration/experiments/04-judge-it/tools/e04env.sh
curl -sS -X POST -H 'Authorization: Bearer acme-key' -H 'Content-Type: application/json' \
  -d "{\"worker_id\":$2,\"max_jobs\":${3:-1}}" "$QUEUELET_URL/v1/queues/$1/lease" |
  python3 -I -c 'import json,sys; d=json.load(sys.stdin); [print(j["id"],j["attempt"],j["lease_token"]) for j in d.get("jobs",[])] or print("NONE", d)'
