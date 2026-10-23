#!/usr/bin/env python3
"""E09 H-E09-16: lease jobs in QUEUE with WORKER (100 per lease), then for each race
POST /v1/jobs/{id}/ack against a DELETE spinner. Usage: race_ack_delete.py QUEUE WORKER_ID N"""
import json, os, sys, threading, urllib.request, urllib.error
base = os.environ["QUEUELET_URL"]; key = "acme-key"
def call(method, path, body=None):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(base + path, data=data, method=method,
        headers={"Authorization": "Bearer " + key, "Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=10) as r:
            return r.status, r.read().decode()
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode()
queue, worker, n = sys.argv[1], int(sys.argv[2]), int(sys.argv[3])
summary, odd, first, last = {}, [], None, None
done = 0
while done < n:
    m = min(100, n - done)
    for i in range(m):
        call("POST", f"/v1/queues/{queue}/jobs", {"kind": "e09a", "payload": {"i": done + i}})
    _, body = call("POST", f"/v1/queues/{queue}/lease", {"worker_id": worker, "max_jobs": m})
    for j in json.loads(body)["jobs"]:
        jid, tok = j["id"], j["lease_token"]
        first = first or jid; last = jid
        res = {}; stop = threading.Event()
        def spin():
            while not stop.is_set():
                s, _ = call("DELETE", f"/v1/jobs/{jid}")
                if s in (204, 404):
                    res["delete"] = s; return
        th = threading.Thread(target=spin); th.start()
        st, b = call("POST", f"/v1/jobs/{jid}/ack", {"lease_token": tok, "result": {"ok": True}})
        th.join(timeout=5); stop.set(); th.join()
        k = f"ack={st} delete={res.get('delete')}"
        summary[k] = summary.get(k, 0) + 1
        if st != 200:
            odd.append({"job": jid, "ack": st, "body": b.strip()})
    done += m
print(json.dumps({"jobs": [first, last], "summary": summary}))
for o in odd[:20]:
    print(json.dumps(o))
