"""Siblings of E07-F02. Usage: python3 -I p5_purge_siblings.py wdel|expire <tag> [trials]"""
import os, sys, threading, time, random
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from q import *

probe, tag = sys.argv[1], sys.argv[2]
trials = int(sys.argv[3]) if len(sys.argv) > 3 else 8
random.seed(17)


def race(*fns):
    out = [None] * len(fns)
    bar = threading.Barrier(len(fns))

    def run(i, f):
        bar.wait()
        out[i] = f()
    ts = [threading.Thread(target=run, args=(i, f)) for i, f in enumerate(fns)]
    [t.start() for t in ts]
    [t.join() for t in ts]
    return out


def fill(qn, n, lease_s):
    queue(qn, lease_seconds=lease_s, max_attempts=3)
    w = worker("s" + qn)
    for k in range(0, n, 500):
        ok(call("POST", f"/v1/queues/{qn}/jobs:batch", {"jobs": [{"kind": "echo", "payload": {}} for _ in range(min(500, n - k))]}), 201, 200)
    got = []
    while len(got) < n:
        got += ok(lease(qn, w, 100), 200)["jobs"]
    return w, got


def dist(qn):
    return sql(f"SELECT state||':'||purge_requested||':'||count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') GROUP BY state,purge_requested ORDER BY 1").split()


print(stamp(), probe, tag, trials)
hits = 0
for t in range(trials):
    qn = f"e07s{probe[:2]}{tag}-{t}"
    if probe == "wdel":
        w, got = fill(qn, int(os.environ.get("N", "60")), 300)
        def p():
            time.sleep(random.uniform(0, 0.02))
            return call("POST", f"/v1/queues/{qn}/purge", {})
        d, pr = race(lambda: call("DELETE", f"/v1/workers/{w}"), p)
        res = dist(qn)
        print(f"trial {t} delete {d[0]} purge {pr} -> {res}")
    else:
        w, got = fill(qn, 300, 5)
        clock("advance", seconds=10)
        time.sleep(random.uniform(0, 1.2))
        pr = call("POST", f"/v1/queues/{qn}/purge", {})
        time.sleep(1.5)  # let the sweep finish
        res = dist(qn)
        print(f"trial {t} purge {pr} -> {res}")
    if any(r.startswith(("queued", "retry_wait", "scheduled")) for r in res):
        hits += 1
if probe == "expire":
    print("clock reset", clock("reset"))
print(stamp(), "hits", hits, "of", trials)
