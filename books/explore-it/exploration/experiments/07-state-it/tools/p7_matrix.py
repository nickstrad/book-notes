"""H-E07-15: state x verb matrix. Each cell builds a fresh job in the target state. Usage: python3 -I p7_matrix.py <tag>"""
import os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from q import *
tag = sys.argv[1]
qn = "e07mx" + tag
queue(qn, lease_seconds=300, max_attempts=3, backoff={"kind": "linear", "base_seconds": 3600})
A, B = worker("mxA" + tag), worker("mxB" + tag)


def make(state):
    """Return (job id, holder token or None) with the job in `state`; isolated by draining queued jobs first."""
    if state == "scheduled":
        return enqueue(qn, delay_seconds=3600), None
    j = enqueue(qn)
    if state == "queued":
        return j, None
    l = ok(lease(qn, A), 200)["jobs"][0]
    assert l["id"] == j, (l, j)
    tok = l["lease_token"]
    if state == "leased":
        return j, tok
    if state == "retry_wait":
        ok(verb(j, "nack", lease_token=tok, error="mk"), 200)
    elif state == "succeeded":
        ok(verb(j, "ack", lease_token=tok, result={}), 200)
    elif state == "dead":
        ok(verb(j, "nack", lease_token=tok, error="mk", retry=False), 200)
    elif state == "cancelled":
        ok(verb(j, "cancel"), 200)
    return j, None


def snap(j):
    s = job(j)
    if s is None or "state" not in s:
        return "gone"
    return (s["state"], s["attempts"], s["leased_by"], bool(s["lease_token"]), len(attempts(j)["attempts"]))


verbs = ["lease_B", "heartbeat_wrong", "ack_wrong", "nack_wrong", "cancel", "retry", "delete"]
print(stamp(), "queue", qn, "A", A, "B", B)
print("state | " + " | ".join(verbs))
for st in ["scheduled", "queued", "leased", "retry_wait", "succeeded", "dead", "cancelled"]:
    row = []
    for v in verbs:
        # keep the queue free of other queued jobs so lease_B targets only this job
        for r in sql(f"SELECT id FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') AND state='queued'").split():
            verb(int(r), "cancel")
        j, tok = make(st)
        before = snap(j)
        if v == "lease_B":
            code, body = lease(qn, B)
            code = f"{code}:{[x['id'] for x in body['jobs']]}" if code == 200 else code
        elif v.endswith("_wrong"):
            name = v.split("_")[0]
            extra = {"heartbeat": {"extend_seconds": 60}, "ack": {"result": {}}, "nack": {"error": "w"}}[name]
            code = verb(j, name, lease_token="0" * 32, **extra)[0]
        elif v == "delete":
            code = call("DELETE", f"/v1/jobs/{j}")[0]
        else:
            code = verb(j, v)[0]
        after = snap(j)
        row.append(f"{code} {'=' if after == before else before[0] + '→' + (after if after == 'gone' else after[0])}")
    print(f"{st} | " + " | ".join(row))
print(stamp(), "done")
