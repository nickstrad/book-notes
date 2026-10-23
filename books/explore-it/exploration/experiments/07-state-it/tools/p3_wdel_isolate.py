"""Isolate H-E07-03: sequential control, max_jobs=1 race, and consequences for an orphaned lease.
Usage: python3 -I p3_wdel_isolate.py <tag>"""
import os, sys, threading, time, random
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from q import *

tag = sys.argv[1]
random.seed(7)


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


def batch(qn, n):
    return ok(call("POST", f"/v1/queues/{qn}/jobs:batch", {"jobs": [{"kind": "echo", "payload": {"i": i}} for i in range(n)]}), 201, 200)["ids"]


def orphans(w):
    return sql(f"SELECT count(*) FROM jobs WHERE leased_by={w} AND state='leased'")


qn = "e07wi" + tag
queue(qn, lease_seconds=300, max_attempts=3)
print(stamp(), "queue", qn)

# C1 sequential control: lease 50, then delete (no overlap)
w = worker("ctl" + tag)
batch(qn, 50)
st, b = lease(qn, w, 100)
print("C1 lease", st, len(b["jobs"]), "then delete", call("DELETE", f"/v1/workers/{w}")[0], "orphan_leased", orphans(w),
      "states", sql(f"SELECT state||':'||count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') GROUP BY 1"))
for r in sql(f"SELECT id FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') AND state NOT IN ('succeeded','dead','cancelled')").split():
    verb(int(r), "cancel")

# C2 race with max_jobs=1 (one queued job), delete delayed 0-10 ms, 20 trials
hits = 0
for t in range(20):
    w = worker(f"one{tag}-{t}")
    batch(qn, 1)
    def d():
        time.sleep(random.uniform(0, 0.01))
        return call("DELETE", f"/v1/workers/{w}")
    lres, dres = race(lambda: lease(qn, w, 1), d)
    o = orphans(w)
    hits += o != "0"
    print(f"C2 trial {t} lease {lres[0]} n={len(lres[1]['jobs']) if lres[0]==200 else lres[1]} delete {dres[0]} orphan_leased={o}")
    for r in sql(f"SELECT id FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') AND state NOT IN ('succeeded','dead','cancelled')").split():
        verb(int(r), "cancel")
print("C2 hits", hits, "of 20")

# C3 consequences: create orphan via race (retry up to 10x), then deleted worker heartbeats and acks
for t in range(10):
    w = worker(f"cons{tag}-{t}")
    batch(qn, 50)
    lres, dres = race(lambda: lease(qn, w, 100), lambda: (time.sleep(0.005), call("DELETE", f"/v1/workers/{w}"))[1])
    if lres[0] == 200 and orphans(w) != "0":
        held = [j for j in lres[1]["jobs"] if sql(f"SELECT state||'|'||leased_by FROM jobs WHERE id={j['id']}") == f"leased|{w}"]
        j0, j1 = held[0], held[1]
        print("C3 worker", w, "exists?", sql(f"SELECT count(*) FROM workers WHERE id={w}"), "orphan jobs", len(held))
        print("C3 lease by deleted worker id ->", lease(qn, w, 1))
        print("C3 heartbeat by deleted worker", verb(j0["id"], "heartbeat", lease_token=j0["lease_token"], extend_seconds=3600),
              sql(f"SELECT lease_expires_at FROM jobs WHERE id={j0['id']}"))
        print("C3 ack by deleted worker", verb(j1["id"], "ack", lease_token=j1["lease_token"], result={"by": "deleted"}),
              job(j1["id"])["state"], [(a["n"], a["worker_id"], a["outcome"]) for a in attempts(j1["id"])["attempts"]])
        print("C3 events j1", [(e["event"], e["worker"]) for e in events(j1["id"])["events"]])
        print("C3 second DELETE of same worker", call("DELETE", f"/v1/workers/{w}"))
        break
    for r in sql(f"SELECT id FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') AND state NOT IN ('succeeded','dead','cancelled')").split():
        verb(int(r), "cancel")
print(stamp(), "done")
