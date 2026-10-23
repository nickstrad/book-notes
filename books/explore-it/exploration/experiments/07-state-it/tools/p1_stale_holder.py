"""H-E07-01/02/05/06/07: two-worker stale-token cells, heartbeat shortening, purge-killed retry, duplicate verbs.
Usage: python3 -I p1_stale_holder.py  (after sourcing tools/env.sh)."""
import json, sys, os, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from q import *

tag = sys.argv[1] if len(sys.argv) > 1 else str(int(time.time()))
qn = "e07p1" + tag
print(stamp(), "queue", qn, queue(qn, lease_seconds=30, max_attempts=3)["name"])
A, B = worker("A" + tag), worker("B" + tag)
print("workers A", A, "B", B)

# H-E07-01: lease A -> clock +40 -> sweep -> lease B -> A verbs with A's token -> B ack
j = enqueue(qn)
la = ok(lease(qn, A), 200)["jobs"][0]
print("H01 lease A", j, "attempt", la["attempt"], "exp", la["lease_expires_at"])
print("H01 clock", clock("advance", seconds=40))
print("H01 after sweep:", wait_state(j, {"queued", "dead"}))
lb = ok(lease(qn, B), 200)["jobs"]
print("H01 lease B ->", [(x["id"], x["attempt"]) for x in lb])
for v, body in [("ack", {"result": {"by": "A"}}), ("nack", {"error": "A late"}), ("heartbeat", {"extend_seconds": 60})]:
    print("H01 A", v, verb(j, v, lease_token=la["lease_token"], **body))
jj = job(j)
print("H01 job after A verbs:", jj["state"], "leased_by", jj["leased_by"], "token==B", jj["lease_token"] == lb[0]["lease_token"])
print("H01 B ack", verb(j, "ack", lease_token=lb[0]["lease_token"], result={"by": "B"}))
print("H01 final", {k: job(j)[k] for k in ("state", "attempts", "result", "leased_by", "lease_token")})
print("H01 attempts", [(a["n"], a["worker_id"], a["outcome"]) for a in attempts(j)["attempts"]])
print("H01 events", [(e["event"], e["worker"], e["attempt"]) for e in events(j)["events"]])

# H-E07-02: lease A -> expire -> lease A again; first token ack
j2 = enqueue(qn)
t1 = ok(lease(qn, A), 200)["jobs"][0]["lease_token"]
clock("advance", seconds=40)
print("H02 after sweep:", wait_state(j2, {"queued", "dead"}))
second = ok(lease(qn, A), 200)["jobs"][0]
print("H02 relet attempt", second["attempt"])
print("H02 ack first token", verb(j2, "ack", lease_token=t1, result={}))
print("H02 ack second token", verb(j2, "ack", lease_token=second["lease_token"], result={}))
print("H02 attempts", [(a["n"], a["worker_id"], a["outcome"]) for a in attempts(j2)["attempts"]])

# H-E07-05: heartbeat shortening
j5 = enqueue(qn)
l5 = ok(lease(qn, A), 200)["jobs"][0]
print("H05 lease exp", l5["lease_expires_at"], "now", clock("advance", seconds=0)["now"])
print("H05 heartbeat 1", verb(j5, "heartbeat", lease_token=l5["lease_token"], extend_seconds=1))
print("H05 job exp", job(j5)["lease_expires_at"])
clock("advance", seconds=2)
print("H05 after +2s: ack", verb(j5, "ack", lease_token=l5["lease_token"], result={}), "state", wait_state(j5, {"queued"}, 5))

# H-E07-06: purge-killed dead -> retry -> lease -> nack(retry)
qn6 = qn + "x"
queue(qn6, lease_seconds=30, max_attempts=3)
j6 = enqueue(qn6)
l6 = ok(lease(qn6, A), 200)["jobs"][0]
print("H06 purge", call("POST", f"/v1/queues/{qn6}/purge", {}))
print("H06 flag", sql(f"SELECT state,purge_requested FROM jobs WHERE id={j6}"))
print("H06 nack(retry)", verb(j6, "nack", lease_token=l6["lease_token"], error="x"), job(j6)["state"])
print("H06 retry", verb(j6, "retry"), sql(f"SELECT state,purge_requested,attempts FROM jobs WHERE id={j6}"))
l6b = ok(lease(qn6, A), 200)["jobs"][0]
print("H06 nack(retry) after retry", verb(j6, "nack", lease_token=l6b["lease_token"], error="y"), job(j6)["state"])
print("H06 attempts", [(a["n"], a["outcome"]) for a in attempts(j6)["attempts"]])
print("clock reset", clock("reset"))
