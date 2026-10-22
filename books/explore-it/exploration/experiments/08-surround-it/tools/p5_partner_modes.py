"""E08 P5: hooksink modes 500 / slow / timeout; cross-tenant head-of-line in the shared shipper (H-E08-06..08)."""
import sys, time, q
phase = sys.argv[1]
def ob(j):
    return q.sql(f"select id,attempts,coalesce(status,-1),delivered_at is not null,left(error,60),to_char(event_at,'HH24:MI:SS.MS') from webhook_outbox where job_id={j} order by id")
def recs(j):
    return [(r["status"], r["at"][11:23]) for r in q.hook("/admin/records")["records"] if isinstance(r["body"], dict) and r["body"].get("job_id") == j]
def terminal(qn, w, key="acme-key"):
    j = q.ok(q.call("POST", f"/v1/queues/{qn}/jobs", {"kind": "echo", "payload": {}}, key=key), 201)["id"]
    lj = q.ok(q.call("POST", f"/v1/queues/{qn}/lease", {"worker_id": w, "max_jobs": 1}, key=key), 200)["jobs"][0]
    q.ok(q.call("POST", f"/v1/jobs/{j}/ack", {"lease_token": lj["lease_token"], "result": {"p": phase}}, key=key), 200)
    return j
w = q.worker("e08p5")
if phase == "500":
    q.hook("/admin/mode", "POST", {"mode": "500"})
    j = terminal("e08hs", w); print(q.stamp(), "mode 500; acked job", j)
    for _ in range(4):
        time.sleep(4); print(q.stamp(), "outbox", ob(j))
    print("hooksink records for job:", recs(j))
    q.hook("/admin/mode", "POST", {"mode": "ok"}); print(q.stamp(), "mode ok")
    time.sleep(7); print(q.stamp(), "outbox", ob(j)); print("records:", recs(j))
elif phase == "slow":
    q.hook("/admin/mode", "POST", {"mode": "slow"})
    j = terminal("e08hs", w); d = q.enqueue("e08p1", delay_seconds=4)
    print(q.stamp(), "mode slow; acked job", j, "delayed job", d, "run_at", q.job(d)["run_at"])
    st, dt = q.wait_state(d, {"queued"}, 10); print(q.stamp(), "delayed job ->", st, "after", dt, "s (run_at+~4s expected)")
    time.sleep(6); print(q.stamp(), "outbox", ob(j)); print("records:", recs(j))
    q.hook("/admin/mode", "POST", {"mode": "ok"}); print(q.stamp(), "mode ok")
    time.sleep(8); print(q.stamp(), "outbox", ob(j)); print("records (count = partner side-effects):", recs(j))
elif phase == "timeout":
    N = int(sys.argv[2])
    q.ok(q.call("POST", "/v1/queues", {"name": "e08slow", "webhook": {"url": "http://hooksink:8201/hook", "secret": "g"}}, key="globex-key"), 201)
    q.queue("e08fast", webhook={"url": "http://hooksink:8201/admin/records", "secret": "a"})
    gw = q.ok(q.call("POST", "/v1/workers", {"name": "g", "kinds": ["echo"]}, key="globex-key"), 201)["worker_id"]
    q.hook("/admin/mode", "POST", {"mode": "timeout"}); print(q.stamp(), "mode timeout")
    g = [terminal("e08slow", gw, key="globex-key") for _ in range(N)]
    a = terminal("e08fast", w)
    print(q.stamp(), f"globex terminal jobs {g[0]}..{g[-1]} ({N}); acme job {a}")
    print("outbox ids: globex", q.sql(f"select min(id)||'..'||max(id) from webhook_outbox where job_id between {g[0]} and {g[-1]}"), "acme", ob(a))
    d = q.enqueue("e08p1", delay_seconds=3); print("core check: delayed job", d)
    t0 = time.time()
    while time.time() - t0 < float(sys.argv[3]):
        time.sleep(10)
        print(q.stamp(), "acme", ob(a), "| globex attempts", q.sql(f"select string_agg(attempts::text,',' order by id) from webhook_outbox where job_id between {g[0]} and {g[-1]}"), "| delayed", q.job(d)["state"])
    q.hook("/admin/mode", "POST", {"mode": "ok"}); print(q.stamp(), "mode ok")
    t1 = time.time()
    while time.time() - t1 < 90:
        r = ob(a)
        if r.split("|")[3] == "t":
            break
        time.sleep(1)
    print(q.stamp(), f"acme delivered {round(time.time()-t1,1)} s after mode ok:", ob(a))
    print("globex pending:", q.sql(f"select count(*) from webhook_outbox where delivered_at is null and job_id between {g[0]} and {g[-1]}"))
