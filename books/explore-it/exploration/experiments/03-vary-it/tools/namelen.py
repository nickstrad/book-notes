#!/usr/bin/env python3
"""E03 helper: create a queue with a long name, then GET and DELETE it by that name.
Usage: namelen.py <kind:rep|rand> <length>. Uses QUEUELET_URL and acme-key."""
import json, os, random, string, sys, urllib.request, urllib.error
base = os.environ["QUEUELET_URL"]
kind, n = sys.argv[1], int(sys.argv[2])
random.seed(n)
name = "q" * n if kind == "rep" else "".join(random.choice(string.ascii_letters + string.digits) for _ in range(n))
def call(method, path, body=None):
    req = urllib.request.Request(base + path, None if body is None else json.dumps(body).encode(),
                                 {"Authorization": "Bearer acme-key", "Content-Type": "application/json"}, method=method)
    try:
        with urllib.request.urlopen(req) as r:
            return r.status, r.read()[:120]
    except urllib.error.HTTPError as e:
        return e.code, e.read()[:120]
print(kind, n, "POST", *call("POST", "/v1/queues", {"name": name}))
print(kind, n, "GET", call("GET", "/v1/queues/" + name)[0])
print(kind, n, "DELETE", call("DELETE", "/v1/queues/" + name)[0])
