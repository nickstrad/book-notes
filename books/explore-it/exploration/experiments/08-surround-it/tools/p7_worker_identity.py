"""E08 P7: lease as another acme worker by id only (H-E08-19)."""
import q
wa = q.sql("select id from workers where name='worker-a' order by id desc limit 1")
q.queue("e08p7"); j = q.enqueue("e08p7")
r = q.lease("e08p7", int(wa))
print(q.stamp(), "worker-a id", wa, "lease as worker-a →", r[0], [(x["id"], x["attempt"]) for x in r[1]["jobs"]])
print("job leased_by:", q.job(j)["leased_by"], "state:", q.job(j)["state"])
print("nack cleanup:", q.verb(j, "nack", lease_token=r[1]["jobs"][0]["lease_token"], error="e08 cleanup", retry=False)[0])
print("attempt record holder:", [(a["n"], a.get("worker_id"), a.get("outcome")) for a in q.attempts(j)["attempts"]])
