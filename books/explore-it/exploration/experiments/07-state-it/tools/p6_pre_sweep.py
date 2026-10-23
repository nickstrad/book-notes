"""H-E07-12/13: availability of an expired lease / a due retry_wait job before the background tick.
Usage: python3 -I p6_pre_sweep.py <tag>"""
import os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from q import *
tag = sys.argv[1]
qn = "e07ps" + tag
queue(qn, lease_seconds=30, max_attempts=5, backoff={"kind": "linear", "base_seconds": 60})
A, B = worker("psA" + tag), worker("psB" + tag)
for trial in range(3):
    j = enqueue(qn)
    la = ok(lease(qn, A), 200)["jobs"][0]
    clock("advance", seconds=31)
    t0 = time.time()
    first = lease(qn, B)
    s0 = job(j)["state"]
    got = first[1]["jobs"] if first[0] == 200 else None
    while not got and time.time() - t0 < 3:
        time.sleep(0.05)
        got = ok(lease(qn, B), 200)["jobs"]
    print(f"H12 trial {trial} job {j}: immediate B lease {first[0]} {[x['id'] for x in first[1]['jobs']]} state then {s0}; "
          f"B got {[ (x['id'], x['attempt']) for x in got]} after {round(time.time()-t0,2)} s; A ack {verb(j, 'ack', lease_token=la['lease_token'], result={})[0]}")
    got and ok(verb(j, "ack", lease_token=got[0]["lease_token"], result={}), 200)
for trial in range(3):
    j = enqueue(qn)
    la = ok(lease(qn, A), 200)["jobs"][0]
    ok(verb(j, "nack", lease_token=la["lease_token"], error="r"), 200)
    print("H13 run_at", job(j)["run_at"], job(j)["state"])
    clock("advance", seconds=61)
    t0 = time.time()
    first = lease(qn, B)
    s0 = job(j)["state"]
    got = first[1]["jobs"]
    while not got and time.time() - t0 < 3:
        time.sleep(0.05)
        got = ok(lease(qn, B), 200)["jobs"]
    print(f"H13 trial {trial} job {j}: immediate B lease {[x['id'] for x in first[1]['jobs']]} state then {s0}; B got {[(x['id'], x['attempt']) for x in got]} after {round(time.time()-t0,2)} s")
    got and ok(verb(j, "ack", lease_token=got[0]["lease_token"], result={}), 200)
print("clock reset", clock("reset"))
