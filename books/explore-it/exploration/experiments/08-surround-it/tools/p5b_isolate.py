"""E08 P5b: isolate E08-F02. args: <mode> <N> <globex-url> <observe-seconds>"""
import sys, time, q
mode, N, gurl, obs = sys.argv[1], int(sys.argv[2]), sys.argv[3], float(sys.argv[4])
tag = f"e08b{int(time.time())%100000}"
def ob(j):
    return q.sql(f"select id,attempts,coalesce(status,-1),delivered_at is not null,left(error,70) from webhook_outbox where job_id={j}")
def terminal(qn, w, key):
    j = q.ok(q.call("POST", f"/v1/queues/{qn}/jobs", {"kind": "echo", "payload": {}}, key=key), 201)["id"]
    lj = q.ok(q.call("POST", f"/v1/queues/{qn}/lease", {"worker_id": w, "max_jobs": 1}, key=key), 200)["jobs"][0]
    q.ok(q.call("POST", f"/v1/jobs/{j}/ack", {"lease_token": lj["lease_token"], "result": {}}, key=key), 200)
    return j
print("pending before:", q.sql("select count(*) from webhook_outbox where delivered_at is null"))
q.ok(q.call("POST", "/v1/queues", {"name": tag + "g", "webhook": {"url": gurl, "secret": "g"}}, key="globex-key"), 201)
q.queue(tag + "a", webhook={"url": "http://hooksink:8201/admin/records", "secret": "a"})
gw = q.ok(q.call("POST", "/v1/workers", {"name": "g", "kinds": ["echo"]}, key="globex-key"), 201)["worker_id"]
aw = q.worker("a")
q.hook("/admin/mode", "POST", {"mode": mode})
g = [terminal(tag + "g", gw, "globex-key") for _ in range(N)]
a = terminal(tag + "a", aw, "acme-key"); t0 = time.time()
print(q.stamp(), f"mode {mode}; {N} globex rows → {gurl}; acme job {a}", ob(a))
while time.time() - t0 < obs:
    r = ob(a)
    if r.split("|")[3] == "t":
        break
    time.sleep(0.5)
print(q.stamp(), f"after {round(time.time()-t0,1)} s acme:", ob(a))
print("globex attempts:", q.sql(f"select string_agg(attempts::text,',' order by id)||' err='||max(left(error,60)) from webhook_outbox where job_id between {g[0]} and {g[-1]}"))
