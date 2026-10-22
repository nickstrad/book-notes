"""E08 P5c: can either tenant end the stall through the API? (H-E08-18) arg: globex queue name, acme outbox id"""
import sys, time, q
gq, aid = sys.argv[1], sys.argv[2]
print(q.stamp(), "full error:", q.sql(f"select distinct error from webhook_outbox where queue='{gq}'"))
print("PATCH webhook url ''", q.call("PATCH", f"/v1/queues/{gq}", {"webhook": {"url": ""}}, key="globex-key")[0])
ids = q.sql(f"select job_id from webhook_outbox where queue='{gq}'").split()
print("DELETE jobs", {q.call("DELETE", f"/v1/jobs/{j}", key="globex-key")[0] for j in ids})
print("DELETE queue", q.call("DELETE", f"/v1/queues/{gq}", key="globex-key"))
t0 = time.time()
while time.time() - t0 < 45:
    time.sleep(5)
    print(q.stamp(), "acme", q.sql(f"select id,attempts,delivered_at is not null from webhook_outbox where id={aid}"),
          "| globex pending", q.sql(f"select count(*)||' max attempts '||max(attempts) from webhook_outbox where queue='{gq}' and delivered_at is null"),
          "| health", q.call("GET", "/health")[0])
