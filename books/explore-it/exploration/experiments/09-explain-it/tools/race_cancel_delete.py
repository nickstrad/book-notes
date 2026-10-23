#!/usr/bin/env python3
"""E09 H-E09-14: race POST /v1/jobs/{id}/cancel against a DELETE spinner on the same job.
Usage: race_cancel_delete.py QUEUE N  (env QUEUELET_URL; acme-key). Prints a JSON summary and
per-trial lines where cancel did not return 200."""
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
queue, n = sys.argv[1], int(sys.argv[2])
ids = [json.loads(call("POST", f"/v1/queues/{queue}/jobs", {"kind": "e09none", "payload": {"t": i}})[1])["id"] for i in range(n)]
summary = {}
odd = []
for jid in ids:
    res = {}
    stop = threading.Event()
    def spin():
        while not stop.is_set():
            st, body = call("DELETE", f"/v1/jobs/{jid}")
            if st in (204, 404):
                res["delete"] = st; return
    t = threading.Thread(target=spin); t.start()
    st, body = call("POST", f"/v1/jobs/{jid}/cancel")
    t.join(timeout=5); stop.set()
    k = f"cancel={st} delete={res.get('delete')}"
    summary[k] = summary.get(k, 0) + 1
    if st != 200:
        odd.append({"job": jid, "cancel": st, "body": body.strip(), "delete": res.get("delete")})
print(json.dumps({"jobs": [ids[0], ids[-1]], "summary": summary}))
for o in odd[:20]:
    print(json.dumps(o))
