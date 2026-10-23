"""Sequential controls for H-E07-04/04b: lease 10 of 20, purge, nack all (no overlap). Usage: python3 -I p4_purge_control.py <tag>"""
import os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from q import *
tag = sys.argv[1]
qn = "e07pc" + tag
queue(qn, lease_seconds=300, max_attempts=3, backoff={"kind": "linear", "base_seconds": 3600})
w = worker("pc" + tag)
ok(call("POST", f"/v1/queues/{qn}/jobs:batch", {"jobs": [{"kind": "echo", "payload": {}} for _ in range(20)]}), 201, 200)
got = ok(lease(qn, w, 10), 200)["jobs"]
print(stamp(), "leased", len(got), "purge", call("POST", f"/v1/queues/{qn}/purge", {}))
qid = f"(SELECT id FROM queues WHERE name='{qn}')"
print("after purge", sql(f"SELECT state||':'||purge_requested||':'||count(*) FROM jobs WHERE queue_id={qid} GROUP BY state,purge_requested").split())
print("nack codes", sorted(verb(j["id"], "nack", lease_token=j["lease_token"], error="ctl")[0] for j in got))
print("after nacks", sql(f"SELECT state||':'||count(*) FROM jobs WHERE queue_id={qid} GROUP BY state").split())
