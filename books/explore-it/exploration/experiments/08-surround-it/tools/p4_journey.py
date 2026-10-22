"""E08 P4: follow one result from ack → outbox row → partner bytes; verify HMAC by hand (H-E08-09)."""
import base64, hashlib, hmac, json, sys, time, q
CAP = sys.argv[1]  # capture.jsonl path
GW = "172.28.0.1:18292"
datum = {"trace": "E08-ü- -<&>", "n": 12345678901234567890, "f": 1.10, "z": [], "nested": {"b": 1, "a": 2}}
q.queue("e08cap", webhook={"url": f"http://{GW}/cap", "secret": "e08-secret"})
q.queue("e08hs", webhook={"url": "http://hooksink:8201/hook", "secret": "report-secret"})
w = q.worker("e08p4", ("echo",))
q.hook("/admin/records", "DELETE")
ids = {}
for qn in ("e08cap", "e08hs"):
    j = q.enqueue(qn)
    lj = q.ok(q.lease(qn, w), 200)["jobs"][0]
    # send the datum as raw JSON text so Python float/int formatting is ours, not json.dumps'
    raw = '{"lease_token":"%s","result":{"trace":"E08-ü-\\u2028-<&>","n":12345678901234567890,"f":1.10,"z":[],"nested":{"b":1,"a":2}}}' % lj["lease_token"]
    import urllib.request
    req = urllib.request.Request(q.URL + f"/v1/jobs/{j}/ack", raw.encode(), method="POST",
                                 headers={"Authorization": "Bearer acme-key", "Content-Type": "application/json"})
    print(qn, "ack", urllib.request.urlopen(req).status, "job", j, q.stamp())
    ids[qn] = j
time.sleep(4)
for qn, j in ids.items():
    print("--", qn, "job", j)
    print("job.result (API):", json.dumps(q.job(j)["result"], ensure_ascii=False))
    print("outbox row (SQL):", q.sql(f"select id, attempts, status, delivered_at is not null, body::text from webhook_outbox where job_id={j}"))
cap = [json.loads(l) for l in open(CAP)]
c = cap[-1]
raw = base64.b64decode(c["raw_b64"])
sig = c["headers"].get("X-Queuelet-Signature")
mine = hmac.new(b"e08-secret", raw, hashlib.sha256).hexdigest()
print("== capture listener: path", c["path"], "content-type", c["headers"].get("Content-Type"))
print("raw bytes:", raw.decode())
print("received signature:", sig)
print("hand HMAC-SHA256(e08-secret, raw):", mine, "equal:", hmac.compare_digest(sig, mine))
print("wrong-secret check:", hmac.new(b"report-secret", raw, hashlib.sha256).hexdigest() == sig)
rec = q.hook("/admin/records")["records"][-1]
print("== hooksink record: status", rec["status"], "signature_valid", rec["signature_valid"])
print("raw_body:", rec["raw_body"])
outbox = json.loads(q.sql(f"select body::text from webhook_outbox where job_id={ids['e08hs']}"))
print("semantic equal (hooksink body == outbox body):", json.loads(rec["raw_body"]) == outbox)
print("semantic equal (capture raw == outbox e08cap):", json.loads(raw) == json.loads(q.sql(f"select body::text from webhook_outbox where job_id={ids['e08cap']}")))
print("result n preserved:", json.loads(raw)["result"]["n"] == 12345678901234567890, "f text in raw:", '"f":1.10' in raw.decode())
