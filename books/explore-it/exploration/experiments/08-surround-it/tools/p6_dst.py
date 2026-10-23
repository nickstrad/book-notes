"""E08 P6: test clock across 2027 DST transitions in America/New_York (H-E08-10..13)."""
import time, q
NY = "America/New_York"
def sched(name, cron):
    return q.ok(q.call("POST", "/v1/schedules", {"name": name, "queue": "e08dst", "kind": name, "cron": cron, "timezone": NY}), 201)
def snap(label):
    s = q.sql("select kind||'='||to_char(next_run_at at time zone 'UTC','MM-DD\"T\"HH24:MI\"Z\"') from schedules where queue_id=(select id from queues where name='e08dst' and namespace_id=1) order by id")
    j = q.sql("select kind||'@'||to_char(run_at at time zone 'UTC','MM-DD\"T\"HH24:MI\"Z\"') from jobs where queue_id=(select id from queues where name='e08dst' and namespace_id=1) order by id")
    print(f"[{label}] next: {' '.join(s.split())}\n    jobs: {' '.join(j.split())}")
def step(at):
    print(q.stamp(), "clock", q.clock("set", at=at)["now"]); time.sleep(2.5); snap(at)
q.queue("e08dst")
print(q.stamp(), "clock", q.clock("set", at="2027-03-13T12:00:00Z")["now"])
for n, c in (("dst-0230", "30 2 * * *"), ("dst-hourly", "0 * * * *"), ("dst-0200", "0 2 * * *"), ("dst-0900", "0 9 * * *")):
    v = sched(n, c); print("created", n, "next_run_at", v["next_run_at"], "zone_offset", v.get("zone_offset"))
snap("created")
for at in ("2027-03-13T14:00:30Z", "2027-03-14T06:00:30Z", "2027-03-14T07:00:30Z", "2027-03-14T13:00:30Z", "2027-03-15T06:00:30Z", "2027-03-15T06:30:30Z"):
    step(at)
print(q.stamp(), "clock", q.clock("set", at="2027-11-07T04:00:00Z")["now"])
for n, c in (("fb-0130", "30 1 * * *"), ("fb-0100", "0 1 * * *")):
    v = sched(n, c); print("created", n, "next_run_at", v["next_run_at"], "zone_offset", v.get("zone_offset"))
snap("fb created")
for at in ("2027-11-07T05:00:30Z", "2027-11-07T05:30:30Z", "2027-11-07T06:00:30Z", "2027-11-07T06:30:30Z", "2027-11-07T07:00:30Z"):
    step(at)
