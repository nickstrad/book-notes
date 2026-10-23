#!/usr/bin/env python3
"""E09 H-E09-15: per trial, lease K jobs (max_attempts 1) in QUEUE with WORKER, spin DELETE on the
lowest id X, advance the test clock past the lease so one Expire pass makes all K dead.
Usage: race_expire_delete.py QUEUE WORKER_ID TRIALS K  (env QUEUELET_URL). Prints per-trial ids."""
import json, os, sys, threading, time, urllib.request, urllib.error
base = os.environ["QUEUELET_URL"]
def call(method, path, body=None, key="acme-key"):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(base + path, data=data, method=method,
        headers={"Authorization": "Bearer " + key, "Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=10) as r:
            return r.status, r.read().decode()
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode()
queue, worker, trials, k = sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4])
for t in range(trials):
    for i in range(k):
        call("POST", f"/v1/queues/{queue}/jobs", {"kind": "e09d", "payload": {"trial": t, "i": i}})
    st, body = call("POST", f"/v1/queues/{queue}/lease", {"worker_id": worker, "max_jobs": k})
    ids = sorted(j["id"] for j in json.loads(body)["jobs"])
    x = ids[0]; res = {}
    stop = threading.Event()
    def spin():
        while not stop.is_set():
            s, _ = call("DELETE", f"/v1/jobs/{x}")
            if s == 204:
                res["deleted_at"] = time.time(); return
    th = threading.Thread(target=spin); th.start()
    call("POST", "/admin/clock", {"action": "advance", "seconds": 31}, key="admin-key")
    th.join(timeout=4); stop.set(); th.join()
    print(json.dumps({"trial": t, "x": x, "others": ids[1:], "x_deleted": "deleted_at" in res}), flush=True)
