#!/usr/bin/env python3
"""Replay actual crash-worker/500/slow-partner consoles after E02 hooks.
Use: python3 lab/observe02.py (running seeded lab; serial execution).
"""
import json
import pathlib
import subprocess
import sys
import time
sys.path.insert(0,str(pathlib.Path(__file__).resolve().parent/'repro'))
sys.argv=['observe02','T02']
import probe as p
import probe02 as e

e.queue(lease_seconds=1,max_attempts=1,webhook={'url':'http://hooksink:8201/hook','secret':'report-secret'})
p.ok('POST','/admin/mode',{'mode':'500'},base=p.HOOK)
try:
    j=p.ok('POST',p.QPATH+'/jobs',{'kind':'crash'})['id']
    result=subprocess.run(p.COMPOSE+['run','--rm','--no-deps','-e','WORKER_QUEUE='+p.NAME,'-e','WORKER_NAME=e02-crash','worker-a'],stdin=subprocess.DEVNULL,capture_output=True,text=True,timeout=20)
    print('actual crash worker exit='+str(result.returncode)); print(result.stdout)
    p.require(result.returncode==3,('worker did not crash',result.stdout,result.stderr))
    p.wait(lambda:p.job(j)['state']=='dead',seconds=5)
    p.wait(lambda:any(v['event']=='notification_failed' for v in e.events(j)[1]['events']),seconds=5)
    print('crash/500 history '+json.dumps(e.events(j)[1]))
    p.ok('POST','/admin/mode',{'mode':'slow'},base=p.HOOK)
    slow=e.complete()
    p.wait(lambda:any(v['event']=='notification_failed' and 'deadline' in v['error'] for v in e.events(slow)[1]['events']),seconds=12)
    rows=e.events(slow)[1]['events']; records=[r for r in p.ok('GET','/admin/records',base=p.HOOK)['records'] if r['body']['job_id']==slow]
    p.require(records and records[0]['status']==200,records)
    print('slow partner records '+json.dumps(records)); print('sender timeout history '+json.dumps(rows))
    p.ok('POST','/admin/mode',{'mode':'ok'},base=p.HOOK)
    p.wait(lambda:any(v['event']=='notification_delivered' for v in e.events(slow)[1]['events']),seconds=12)
    print('recovery history '+json.dumps(e.events(slow)[1]))
    # Exercise the archived image's actual qctl (the source CLI is not substituted).
    watched=subprocess.run(p.COMPOSE+['exec','-T','-e','QUEUELET_URL=http://localhost:8200','api','timeout','2','qctl','watch',p.NAME],stdin=subprocess.DEVNULL,capture_output=True,text=True)
    p.require(watched.returncode in (124,143) and 'lease_expired' in watched.stdout,('qctl watch failed',watched.returncode,watched.stdout,watched.stderr))
    print('qctl watch '+p.NAME+' '+watched.stdout)
    for args in [('job','events',str(j)),('clock','set','2030-01-01T00:00:00Z'),('clock','advance','5'),('clock','reset')]:
        output=subprocess.check_output(p.COMPOSE+['exec','-T','-e','QUEUELET_URL=http://localhost:8200','-e','QUEUELET_KEY=admin-key','api','qctl',*args],text=True,stdin=subprocess.DEVNULL)
        print('qctl '+' '.join(args)+' '+output.strip())
    print('OBSERVATION REPLAY HOLDS: actual worker exit 3; crash expiry; HTTP 500 failure; partner 200 with sender deadline error; recovery 2xx; qctl events/clock commands')
finally:
    p.ok('POST','/admin/mode',{'mode':'ok'},base=p.HOOK)
    p.ok('POST','/admin/clock',{'action':'reset'},key='admin-key')
