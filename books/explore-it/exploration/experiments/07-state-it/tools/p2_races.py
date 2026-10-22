"""Two-actor race probes. Usage: python3 -I p2_races.py <probe> <tag> [trials]
probe: wdel (H-E07-03), purgelease (H-E07-04), purgenack (H-E07-04b), dup (H-E07-07)."""
import os, sys, threading, time, random
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from q import *

probe, tag = sys.argv[1], sys.argv[2]
trials = int(sys.argv[3]) if len(sys.argv) > 3 else 20
random.seed(int.from_bytes(tag.encode(), "big") % 2**31)


def race(*fns):
    """Start fns together behind a barrier; return their results in order."""
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


def delayed(f, lo, hi):
    def g():
        time.sleep(random.uniform(lo, hi))
        return f()
    return g


print(stamp(), "probe", probe, "tag", tag, "trials", trials)
hits = 0
if probe == "wdel":
    qn = "e07wd" + tag
    queue(qn, lease_seconds=300, max_attempts=3)
    for t in range(trials):
        w = worker(f"wd{tag}-{t}")
        batch(qn, 50)
        lres, dres = race(lambda: lease(qn, w, 100), delayed(lambda: call("DELETE", f"/v1/workers/{w}"), 0, 0.02))
        n_leased = len(lres[1]["jobs"]) if lres[0] == 200 else None
        orphan = sql(f"SELECT count(*) FROM jobs WHERE leased_by={w} AND state='leased' AND NOT EXISTS(SELECT 1 FROM workers WHERE id={w})")
        print(f"trial {t} worker {w} lease {lres[0]} jobs={n_leased} err={lres[1] if lres[0]!=200 else ''} delete {dres[0]} orphan_leased={orphan}")
        if orphan != "0":
            hits += 1
        # clean: cancel everything left in the queue so the next trial starts empty
        sql(f"SELECT 1")
        for row in sql(f"SELECT id FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') AND state NOT IN ('succeeded','dead','cancelled')").split():
            verb(int(row), "cancel")
elif probe == "purgelease":
    for t in range(trials):
        qn = f"e07pl{tag}-{t}"
        queue(qn, lease_seconds=300, max_attempts=3)
        ws = [worker(f"pl{tag}-{t}-{i}") for i in range(4)]
        batch(qn, 200)
        def leaser(w):
            def f():
                got = []
                for _ in range(8):
                    st, b = lease(qn, w, 10)
                    if st == 200:
                        got += b["jobs"]
                return got
            return f
        res = race(*[leaser(w) for w in ws], delayed(lambda: call("POST", f"/v1/queues/{qn}/purge", {}), 0.01, 0.08))
        pres = res[-1]
        leased = [j for r in res[:-1] for j in r]
        unflag = sql(f"SELECT id FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') AND state='leased' AND NOT purge_requested ORDER BY id").split()
        flag = sql(f"SELECT count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') AND state='leased' AND purge_requested")
        print(f"trial {t} queue {qn} purge {pres} leased_total={len(leased)} leased_flagged={flag} leased_unflagged={len(unflag)} ids={unflag[:5]}")
        if unflag:
            hits += 1
            tok = {j['id']: j['lease_token'] for j in leased}
            jid = int(unflag[0])
            print("   nack(retry) on unflagged", jid, verb(jid, "nack", lease_token=tok[jid], error="after purge"), job(jid)["state"])
elif probe == "purgenack":
    for t in range(trials):
        qn = f"e07pn{tag}-{t}"
        queue(qn, lease_seconds=300, max_attempts=3, backoff={"kind": "linear", "base_seconds": 3600})
        w = worker(f"pn{tag}-{t}")
        batch(qn, 60)
        got = ok(lease(qn, w, 100), 200)["jobs"]
        nackers = [(lambda j=j: verb(j["id"], "nack", lease_token=j["lease_token"], error="race")) for j in got]
        res = race(*[delayed(f, 0, 0.05) for f in nackers], delayed(lambda: call("POST", f"/v1/queues/{qn}/purge", {}), 0.01, 0.04))
        pres = res[-1]
        codes = {}
        for st, _ in res[:-1]:
            codes[st] = codes.get(st, 0) + 1
        dist = sql(f"SELECT state||':'||count(*) FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') GROUP BY state ORDER BY 1").split()
        pending = sql(f"SELECT id FROM jobs WHERE queue_id=(SELECT id FROM queues WHERE name='{qn}') AND state IN ('retry_wait','queued','scheduled') ORDER BY id").split()
        print(f"trial {t} queue {qn} purge {pres} nack codes {codes} states {dist} pending_ids={pending[:5]}")
        if pending:
            hits += 1
elif probe == "dup":
    qn = "e07dup" + tag
    queue(qn, lease_seconds=300, max_attempts=3)
    w = worker("dup" + tag)
    batch(qn, 2 * trials)
    got = ok(lease(qn, w, 100), 200)["jobs"]
    for i, j in enumerate(got):
        tok = j["lease_token"]
        if i % 2 == 0:
            label, fns = "ack+ack", [lambda: verb(j["id"], "ack", lease_token=tok, result={"x": 1}), lambda: verb(j["id"], "ack", lease_token=tok, result={"x": 2})]
        else:
            label, fns = "ack+nack", [lambda: verb(j["id"], "ack", lease_token=tok, result={"x": 1}), lambda: verb(j["id"], "nack", lease_token=tok, error="dup")]
        res = race(*fns)
        n_att = len(attempts(j["id"])["attempts"])
        s = job(j["id"])
        bad = sorted(r[0] for r in res) != [200, 409] or n_att != 1
        hits += bad
        print(f"{label} job {j['id']} codes {[r[0] for r in res]} state {s['state']} attempts_field {s['attempts']} attempt_rows {n_att}{'  <-- VIOLATION' if bad else ''}")
print(stamp(), "hits", hits, "of", trials)
