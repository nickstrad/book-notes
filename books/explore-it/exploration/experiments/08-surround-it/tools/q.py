"""E08 minimal (copied from inherited 07-state-it/tools/q.py; defaults changed to the e08 stack) queuelet HTTP client (stdlib only). Reads QUEUELET_URL from env (tools/env.sh)."""
import json, os, subprocess, time, urllib.request, urllib.error

URL = os.environ.get("QUEUELET_URL", "http://localhost:18290")
COMPOSE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "../../../../project/deploy/compose.yml")
PG = ["docker", "compose", "-f", COMPOSE, "-p", os.environ.get("COMPOSE_PROJECT_NAME", "queuelet-e08"),
      "exec", "-T", "postgres", "psql", "-U", "queuelet", "-d", "queuelet", "-At", "-c"]


def call(method, path, body=None, key="acme-key"):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(URL + path, data=data, method=method)
    req.add_header("Authorization", "Bearer " + key)
    if data is not None:
        req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req, timeout=30) as r:
            raw = r.read()
            return r.status, (json.loads(raw) if raw else None)
    except urllib.error.HTTPError as e:
        raw = e.read()
        try:
            return e.code, json.loads(raw)
        except Exception:
            return e.code, raw.decode(errors="replace")


def ok(res, *codes):
    st, body = res
    assert st in codes, (st, body)
    return body


def queue(name, **cfg):
    cfg = {"name": name, **cfg}
    return ok(call("POST", "/v1/queues", cfg), 201)


def worker(name, kinds=("echo",)):
    return ok(call("POST", "/v1/workers", {"name": name, "kinds": list(kinds)}), 201)["worker_id"]


def enqueue(q, kind="echo", **kw):
    return ok(call("POST", f"/v1/queues/{q}/jobs", {"kind": kind, "payload": {}, **kw}), 201, 200)["id"]


def lease(q, w, n=1):
    return call("POST", f"/v1/queues/{q}/lease", {"worker_id": w, "max_jobs": n})


def verb(j, v, **body):
    return call("POST", f"/v1/jobs/{j}/{v}", body)


def job(j):
    return call("GET", f"/v1/jobs/{j}")[1]


def attempts(j):
    return call("GET", f"/v1/jobs/{j}/attempts")[1]


def events(j):
    return call("GET", f"/v1/jobs/{j}/events")[1]


def clock(action, **kw):
    return ok(call("POST", "/admin/clock", {"action": action, **kw}, key="admin-key"), 200)


def sql(q):
    return subprocess.run(PG + [q], capture_output=True, text=True).stdout.strip()


def wait_state(j, states, timeout=10):
    t = time.time()
    while time.time() - t < timeout:
        s = job(j)["state"]
        if s in states:
            return s, round(time.time() - t, 2)
        time.sleep(0.1)
    return job(j)["state"], None


def stamp():
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def hook(path, method="GET", body=None):
    """hooksink admin (mode/records) on the e08 stack."""
    base = os.environ.get("QUEUELET_HOOK_URL", "http://localhost:18291")
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(base + path, data=data, method=method)
    if data is not None:
        req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req, timeout=10) as r:
        return json.loads(r.read())
