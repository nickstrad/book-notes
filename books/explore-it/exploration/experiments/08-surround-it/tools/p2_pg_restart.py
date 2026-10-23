"""E08 P2: docker compose restart postgres while a delayed job is due (H-E08-04)."""
import subprocess, threading, time, urllib.request, q
COMPOSE = ["docker", "compose", "-f", q.COMPOSE]
j = q.enqueue("e08p1", delay_seconds=6)
print(q.stamp(), "enqueued", j, q.job(j)["state"], "run_at", q.job(j)["run_at"])
log = []
stop = False
def poll():
    while not stop:
        try:
            with urllib.request.urlopen(q.URL + "/health", timeout=2) as r:
                h = r.status
        except urllib.error.HTTPError as e:
            h = e.code
        except Exception as e:
            h = type(e).__name__
        a = q.call("GET", f"/v1/jobs/{j}")
        a = (a[0], a[1]["state"] if a[0] == 200 else a[1]) if isinstance(a[1], dict) else a
        log.append((time.strftime("%H:%M:%S", time.gmtime()) + f".{int(time.time()*10)%10}", h, a))
        time.sleep(0.5)
t = threading.Thread(target=poll); t.start()
time.sleep(1)
print(q.stamp(), "restart postgres")
r = subprocess.run(COMPOSE + ["restart", "postgres"], capture_output=True, text=True)
print(q.stamp(), "restart returned", r.returncode)
time.sleep(15)
stop = True; t.join()
prev = None
for row in log:
    key = (row[1], str(row[2]))
    if key != prev:
        print(*row)
    prev = key
print("final", q.job(j)["state"], q.stamp())
