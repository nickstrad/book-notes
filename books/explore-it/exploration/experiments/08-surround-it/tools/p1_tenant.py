"""E08 P1: other tenant's key / admin key against acme ids (H-E08-01..03)."""
import json, q
print("start", q.stamp())
# acme fixtures: a leased job (for ack/nack/heartbeat), a dead job (retry/delete), a queued one (cancel)
q.queue("e08p1")
w = q.worker("e08p1-w")
jl = q.enqueue("e08p1"); jd = q.enqueue("e08p1"); jq = q.enqueue("e08p1")
ls = q.ok(q.lease("e08p1", w, 2), 200)["jobs"]
tok = {j["id"]: j["lease_token"] for j in ls}
q.ok(q.verb(jd, "nack", lease_token=tok[jd], error="x", retry=False), 200)
before = {j: q.job(j) for j in (jl, jd, jq)}
G = "globex-key"
res = {}
for j in (jl, jd, jq):
    res[f"GET {j}"] = q.call("GET", f"/v1/jobs/{j}", key=G)[0]
    res[f"attempts {j}"] = q.call("GET", f"/v1/jobs/{j}/attempts", key=G)[0]
    res[f"events {j}"] = q.call("GET", f"/v1/jobs/{j}/events", key=G)
res["cancel jq"] = q.call("POST", f"/v1/jobs/{jq}/cancel", {}, key=G)[0]
res["retry jd"] = q.call("POST", f"/v1/jobs/{jd}/retry", {}, key=G)[0]
res["delete jd"] = q.call("DELETE", f"/v1/jobs/{jd}", key=G)[0]
res["ack jl (right token)"] = q.call("POST", f"/v1/jobs/{jl}/ack", {"lease_token": tok[jl], "result": {"g": 1}}, key=G)[0]
res["nack jl"] = q.call("POST", f"/v1/jobs/{jl}/nack", {"lease_token": tok[jl], "error": "g"}, key=G)[0]
res["heartbeat jl"] = q.call("POST", f"/v1/jobs/{jl}/heartbeat", {"lease_token": tok[jl], "extend_seconds": 5}, key=G)[0]
# H-E08-02
res["globex lease own emails w/ acme worker"] = q.call("POST", "/v1/queues/emails/lease", {"worker_id": w, "max_jobs": 1}, key=G)
res["globex DELETE acme worker"] = q.call("DELETE", f"/v1/workers/{w}", key=G)
sid = q.sql("select id from schedules where namespace_id=1 order by id limit 1")
for verb, m, b in (("GET", "GET", None), ("PATCH", "PATCH", {"enabled": False}), ("run", "POST", {}), ("DELETE", "DELETE", None)):
    path = f"/v1/schedules/{sid}" + ("/run" if verb == "run" else "")
    res[f"globex schedule {verb} {sid}"] = q.call(m, path, b, key=G)[0]
res["acme ns stats globex"] = q.call("GET", "/v1/namespaces/globex/stats")
res["globex ns stats acme"] = q.call("GET", "/v1/namespaces/acme/stats", key=G)[0]
res["globex list queues"] = [x["name"] for x in q.call("GET", "/v1/queues", key=G)[1]["queues"]]
# H-E08-03 admin key
gj = q.ok(q.call("POST", "/v1/queues/emails/jobs", {"kind": "echo", "payload": {}}, key=G), 201)["id"]
res["admin GET acme job"] = q.call("GET", f"/v1/jobs/{jq}", key="admin-key")[0]
res["admin GET globex job"] = q.call("GET", f"/v1/jobs/{gj}", key="admin-key")[0]
res["admin globex ns stats"] = q.call("GET", "/v1/namespaces/globex/stats", key="admin-key")[0]
res["admin cancel acme job"] = "skipped (would mutate)"
res["no key"] = q.call("GET", f"/v1/jobs/{jq}", key="")[0]
res["admin clock with acme-key"] = q.call("POST", "/admin/clock", {"action": "reset"})
after = {j: q.job(j) for j in (jl, jd, jq)}
for k, v in res.items():
    print(k, "→", json.dumps(v)[:200])
print("acme jobs unchanged:", before == after, {j: after[j]["state"] for j in after})
print("end", q.stamp())
