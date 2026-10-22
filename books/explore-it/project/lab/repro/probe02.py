#!/usr/bin/env python3
"""E02 public oracles. No fixture SQL writes or SUT source inspection.
Same probe and real-time bounds run against current and historical binaries.
"""
import json
import sys
import time
import subprocess
import probe as p

ID=p.ID

def complete(queue=None,key='acme-key'):
    queue=queue or p.QPATH
    j=p.ok('POST',queue+'/jobs',{'kind':'manual'},key=key)['id']
    w=p.ok('POST','/v1/workers',{'name':p.NAME,'kinds':['manual']},key=key)['worker_id']
    batch=p.ok('POST',queue+'/lease',{'worker_id':w},key=key)['jobs']
    p.require(len(batch)==1,('missing lease',batch))
    p.ok('POST',f'/v1/jobs/{j}/ack',{'lease_token':batch[0]['lease_token'],'result':{'ok':True}},key=key)
    return j

def events(j,key='acme-key'):
    return p.call('GET',f'/v1/jobs/{j}/events',key=key)

def queue(**kwargs):
    return p.ok('POST','/v1/queues',dict(name=p.NAME,**kwargs))

def delivered(j):
    return any(r['body'].get('job_id')==j for r in p.ok('GET','/admin/records',base=p.HOOK)['records'])

def run():
    if ID=='U09':
        # Actual Compose default worker, echo control on the same queue.
        jobs=[]
        for kind in ['echo','fail3']:
            jobs.append(p.ok('POST','/v1/queues/emails/jobs',{'kind':kind,'priority':9,'max_attempts':1})['id'])
        p.wait(lambda:p.ok('GET',f'/v1/jobs/{jobs[0]}')['state']=='succeeded',seconds=12)
        end=time.monotonic()+8
        while time.monotonic()<end:
            j=p.ok('GET',f'/v1/jobs/{jobs[1]}')
            if j['attempts']>0: break
            time.sleep(.2)
        p.require(j['state'] in ['queued','dead'],j)
        return j['attempts']==0,f'echo {jobs[0]} succeeded; default-worker fail3 {jobs[1]} state={j["state"]} attempts={j["attempts"]}; bound=8s after control'
    if ID=='U10':
        bad=p.NAME+'-bad'; statuses=[]
        for url in ['http://bad host/x','ftp://x/y','/relative']:
            st,_=p.checked('POST','/v1/queues',{'name':bad+str(len(statuses)),'webhook':{'url':url}},statuses=(201,400),key='globex-key')
            statuses.append(st)
        queue(webhook={'url':'http://hooksink:8201/hook','secret':'report-secret'})
        for url in ['http://bad host/x','ftp://x/y','/relative']:
            statuses.append(p.checked('PATCH',p.QPATH,{'webhook':{'url':url}},statuses=(200,400))[0])
        # Publicly create the exact blocking historical state if accepted.
        # On repaired versions bad destinations never enter persistent state.
        if statuses[0]==201: complete('/v1/queues/'+bad+'0',key='globex-key')
        p.ok('PATCH',p.QPATH,{'webhook':{'url':'http://hooksink:8201/hook','secret':'report-secret'}})
        j=complete()
        end=time.monotonic()+8
        while time.monotonic()<end and not delivered(j): time.sleep(.2)
        healthy=delivered(j)
        p.require(all(s==400 for s in statuses) or statuses[0]==201,('inconsistent validation',statuses))
        return statuses!=[400]*6 or not healthy,f'create/PATCH statuses={statuses}; independent job {j} partner_received={healthy}; 8s bound'
    if ID in ['T01','T02']:
        queue(lease_seconds=1,max_attempts=1,webhook={'url':'http://hooksink:8201/hook','secret':'report-secret'})
        p.ok('POST','/admin/mode',{'mode':'500'},base=p.HOOK)
        try:
            j=complete()
            # Public crash surrogate: lease once and abandon, waiting real expiry.
            c=p.ok('POST',p.QPATH+'/jobs',{'kind':'manual'})['id']
            w=p.ok('POST','/v1/workers',{'name':p.NAME,'kinds':['manual']})['worker_id']
            p.ok('POST',p.QPATH+'/lease',{'worker_id':w})
            p.wait(lambda:p.ok('GET',f'/v1/jobs/{c}')['state']=='dead',seconds=5)
            p.wait(lambda:delivered(j),seconds=5)
            time.sleep(.5) # allow console poll
            if ID=='T02':
                st,a=events(j); st2,b=events(c)
                p.require(st in (200,404) and st2==st,('events status',st,st2))
                if st==404: return True,f'job {j}/{c} events HTTP 404; crash state and partner failure have no history'
                a=a['events']; b=b['events']
                p.require(a and b,('empty history',a,b))
                failed=[r for r in a if r['event']=='notification_failed']
                p.require(failed and failed[0]['status']==500 and failed[0]['outbox_id'] and failed[0]['error'],a)
                p.require(any(r['event']=='lease_expired' for r in b),b)
                st,other=events(j,'globex-key'); p.require(st==200 and other['events']==[],('namespace leak',other))
                attempts=p.ok('GET',f'/v1/jobs/{c}/attempts')['attempts']
                p.require(attempts and attempts[0]['started_at']<attempts[0]['finished_at'],attempts)
                p.ok('DELETE',f'/v1/jobs/{j}')
                p.require(any(r['event']=='deleted' for r in events(j)[1]['events']),'deletion lost history')
                return False,f'job {j}: {[r["event"] for r in a]}; abandoned {c}: {[r["event"] for r in b]}; 500/outbox/error and true lease start retained; cross-tenant empty'
            raw=subprocess.check_output(p.COMPOSE+['logs','--no-color','api'],text=True,stdin=subprocess.DEVNULL)
            logs=[]
            for line in raw.splitlines():
                try: logs.append(json.loads(line[line.index('{'):]))
                except (ValueError,json.JSONDecodeError): pass
            rows=[r for r in logs if r.get('job_id') in (j,c) and r.get('queue')==p.NAME and r.get('event')]
            required={'job_id','queue','event','worker','attempt','outbox_id','status','error'}
            valid=all(required<=r.keys() for r in rows)
            names={r['event'] for r in rows}
            return not (valid and {'created','leased','succeeded','lease_expired','notification_failed'}<=names),f'jobs {j}/{c} JSON lifecycle events={sorted(names)}; correlated_fields={valid}; 500 and crash consoles compared'
        finally: p.ok('POST','/admin/mode',{'mode':'ok'},base=p.HOOK)
    if ID=='T03':
        # API controls are unavailable historically; that is this visibility gap.
        st,v=p.call('POST','/admin/clock',{'action':'set','at':'2030-01-01T00:00:00Z'},key='admin-key')
        p.require(st in (200,404),('clock setup',st,v))
        if st==404: return True,'authenticated lab clock endpoint HTTP 404; semantic-time control unavailable'
        try:
            p.require(p.call('POST','/admin/clock',{'action':'advance','seconds':1})[0]==403,'namespace key controls global clock')
            queue(lease_seconds=60,max_attempts=1)
            j=p.ok('POST',p.QPATH+'/jobs',{'kind':'manual','delay_seconds':3600})['id']
            p.require(p.ok('GET',f'/v1/jobs/{j}')['state']=='scheduled','not delayed')
            start=time.monotonic()
            p.ok('POST','/admin/clock',{'action':'advance','seconds':3601},key='admin-key')
            p.wait(lambda:p.ok('GET',f'/v1/jobs/{j}')['state']=='queued',seconds=4)
            w=p.ok('POST','/v1/workers',{'name':p.NAME,'kinds':['manual']})['worker_id']
            p.ok('POST',p.QPATH+'/lease',{'worker_id':w})
            p.ok('POST','/admin/clock',{'action':'advance','seconds':61},key='admin-key')
            p.wait(lambda:p.ok('GET',f'/v1/jobs/{j}')['state']=='dead',seconds=4)
            return False,f'job {j} scheduled→queued→leased→dead; advanced 3662 semantic seconds in {time.monotonic()-start:.2f}s real time; non-admin 403'
        finally: p.ok('POST','/admin/clock',{'action':'reset'},key='admin-key')
    raise RuntimeError('unknown id '+ID)

if __name__=='__main__':
    try:
        present,detail=run(); print(('PRESENT: ' if present else 'ABSENT: ')+ID+' — '+detail)
    except Exception as e:
        print('PROBE ERROR: '+ID+' — '+repr(e),file=sys.stderr); sys.exit(2)
