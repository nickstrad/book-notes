import json,hmac,hashlib,sys
for line in open(sys.argv[1]):
    r=json.loads(line)
    if not r["sig"]: continue
    raw=r["body"].encode(); obj=json.loads(raw)
    pretty=json.dumps(obj,indent=2,sort_keys=True,separators=(',',': ')).encode()
    h=lambda b: hmac.new(sys.argv[2].encode(),b,hashlib.sha256).hexdigest()
    print("raw_body:",r["body"]); print("sig received :",r["sig"]); print("hmac(raw)    :",h(raw), h(raw)==r["sig"]); print("hmac(pretty) :",h(pretty), h(pretty)==r["sig"])
