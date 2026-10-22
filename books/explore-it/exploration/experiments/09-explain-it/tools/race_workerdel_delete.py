#!/usr/bin/env python3
"""E09 H-E09-17: per trial create a worker [e09d], lease K jobs (max_attempts 1) from QUEUE,
spin DELETE on the lowest job id X, then DELETE the worker (releases all K as dead).
Usage: race_workerdel_delete.py QUEUE TRIALS K  (env QUEUELET_URL)."""
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
queue, trials, k = sys.argv[1], int(sys.argv[2]), int(sys.argv[3])
summary = {}
for t in range(trials):
    w = json.loads(call("POST", "/v1/workers", {"name": f"e09wd{t}", "kinds": ["e09d"]})[1])["worker_id"]
    for i in range(k):
        call("POST", f"/v1/queues/{queue}/jobs", {"kind": "e09d", "payload": {"wtrial": t, "i": i}})
    ids = sorted(j["id"] for j in json.loads(call("POST", f"/v1/queues/{queue}/lease", {"worker_id": w, "max_jobs": k})[1])["jobs"])
    x = ids[0]; stop = threading.Event()
    def spin():
        while not stop.is_set():
            if call("DELETE", f"/v1/jobs/{x}")[0] == 204:
                return
    th = threading.Thread(target=spin); th.start()
    st, body = call("DELETE", f"/v1/workers/{w}")
    th.join(timeout=4); stop.set(); th.join()
    summary[st] = summary.get(st, 0) + 1
    print(json.dumps({"trial": t, "worker": w, "x": x, "others": ids[1:], "worker_delete": st, "body": body.strip()[:160]}), flush=True)
print(json.dumps({"summary": summary}))
