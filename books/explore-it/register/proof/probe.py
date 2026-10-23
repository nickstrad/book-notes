#!/usr/bin/env python3
"""Seeder/reviewer answer-key probes. Never export this directory to explorers.

Each probe creates isolated resources. SQL writes only arrange time/physical fixtures
where v0 has no clock; observations use API and SQL. Exit 0 means defect PRESENT,
1 means ABSENT, and 2 means the probe failed to run (never evidence of absence).
"""
import datetime as dt,json,os,pathlib,subprocess,sys,time,urllib.request,urllib.error,uuid,zoneinfo
ROOT=pathlib.Path(__file__).resolve().parents[2]/'project'
BASE=os.getenv('QUEUELET_URL','http://localhost:18200')
HOOK=os.getenv('QUEUELET_HOOK_URL','http://localhost:18201')
COMPOSE=['docker','compose','-f',str(ROOT/'deploy/compose.yml')]
D=sys.argv[1]
name='proof-'+D.lower()+'-'+uuid.uuid4().hex[:8]
def call(method,path,body=None,key='acme-key',base=BASE):
    req=urllib.request.Request(base+path,None if body is None else json.dumps(body).encode(),{'Authorization':'Bearer '+key,'Content-Type':'application/json'},method=method)
    try:
        with urllib.request.urlopen(req,timeout=15) as r:
            b=r.read();return r.status,json.loads(b) if b else None
    except urllib.error.HTTPError as e:
        b=e.read();return e.code,json.loads(b) if b else None
def ok(method,path,body=None,**kwargs):
    status,result=call(method,path,body,**kwargs)
    if status>=400:raise RuntimeError((method,path,status,result))
    return result
def sql(query):
    return subprocess.check_output(COMPOSE+['exec','-T','postgres','psql','-U','queuelet','-d','queuelet','-At','-v','ON_ERROR_STOP=1','-c',query],text=True).strip()
def queue(**kwargs):return ok('POST','/v1/queues',dict(name=name,**kwargs))
def enqueue(**kwargs):return ok('POST',f'/v1/queues/{name}/jobs',dict({'kind':'manual','payload':{}},**kwargs))['id']
def worker():return ok('POST','/v1/workers',{'name':name,'kinds':['manual']})['worker_id']
def lease(w=None,n=1):return ok('POST',f'/v1/queues/{name}/lease',{'worker_id':w or worker(),'max_jobs':n})['jobs']
def job(j):return ok('GET',f'/v1/jobs/{j}')
def ack(j,token='token',result=None):return call('POST',f'/v1/jobs/{j}/ack',{'lease_token':token,'result':result})
def nack(j,token='token',**kwargs):return ok('POST',f'/v1/jobs/{j}/nack',dict(lease_token=token,error='failure',**kwargs))
def schedule(**kwargs):return ok('POST','/v1/schedules',dict({'name':name,'queue':name,'kind':'manual','payload':{},'cron':'* * * * *','timezone':'UTC','enabled':True},**kwargs))
def wait(test,seconds=6):
    end=time.monotonic()+seconds
    while time.monotonic()<end:
        v=test()
        if v:return v
        time.sleep(.15)
    return False
def records(j):return [r for r in ok('GET','/admin/records',base=HOOK)['records'] if isinstance(r['body'],dict) and r['body'].get('job_id')==j]
def due_schedule(s,seconds=120):
    stamp=(dt.datetime.now(dt.timezone.utc)-dt.timedelta(seconds=seconds)).isoformat()
    sql(f"UPDATE schedules SET next_run_at='{stamp}' WHERE id={s['id']}")
def result(present,detail):
    print(('PRESENT' if present else 'ABSENT')+': '+D+' — '+detail)
    return 0 if present else 1
def run():
    if D=='D01':
        p=subprocess.run(['go','run','./cmd/qctl','enqueue','missing-'+name],cwd=ROOT,env=dict(os.environ,QUEUELET_URL=BASE),capture_output=True,text=True)
        return result(p.returncode==0 and '404' in p.stderr,f'CLI exit={p.returncode}, stderr={p.stderr.strip()}')
    if D=='D02':
        status,_=call('DELETE','/v1/queues/'+name)
        return result(status==204,f'delete absent queue returned {status}')
    if D in ('D03','D18','D20'):queue(lease_seconds=1,max_attempts=2)
    elif D in ('D04','D17','D24'):queue(webhook={'url':'http://hooksink:8201/hook','secret':'report-secret'})
    else:queue()
    if D=='A1':
        j=enqueue(max_attempts=30);l=lease()[0];sql(f'UPDATE jobs SET attempts=19 WHERE id={j}');nack(j,l['lease_token']);v=job(j)
        delay=(dt.datetime.fromisoformat(v['run_at'].replace('Z','+00:00'))-dt.datetime.now(dt.timezone.utc)).total_seconds()
        return result(delay>86400,f'attempt 20 schedules retry in {delay/86400:.2f} days with base 5 seconds')
    if D=='A2':
        j=enqueue();l=lease()[0];ok('POST',f'/v1/queues/{name}/purge',{});nack(j,l['lease_token'],delay_seconds=3600);state=job(j)['state']
        return result(state=='retry_wait',f'lease → purge → nack leaves {state}')
    if D=='A3':
        j=enqueue();l=lease()[0];ok('POST',f'/v1/jobs/{j}/cancel',{});status,_=ack(j,l['lease_token'],{'work':'completed'})
        return result(status==409 and job(j)['state']=='cancelled',f'cancelled running work; completion ack={status}, no cancellation signal in worker protocol')
    if D=='A4':
        ok('PATCH',f'/v1/queues/{name}',{'retention_hours':0});j=enqueue();l=lease()[0];nack(j,l['lease_token'],retry=False);removed=wait(lambda:call('GET',f'/v1/jobs/{j}')[0]==404)
        return result(bool(removed),'dead job deleted by retention sweep at zero-hour retention')
    if D=='D03':
        j=enqueue();w=worker();seen=[]
        for _ in range(3):
            batch=lease(w);assert len(batch)==1;seen.append(batch[0]['attempt'])
            assert wait(lambda:job(j)['state']=='queued',4)
        count=len(ok('GET',f'/v1/jobs/{j}/attempts')['attempts'])
        return result(seen==[1,1,1] and count==0,f'lease attempts={seen}; attempt rows={count}')
    if D in ('D04','D17','D24'):
        if D=='D04':ok('POST','/admin/mode',{'mode':'500'},base=HOOK)
        try:
            j=enqueue();l=lease()[0];ack(j,l['lease_token'],{'text':'x'*1500} if D=='D17' else {'ok':True})
            assert wait(lambda:records(j),8);record=records(j)[0]
            if D=='D04':
                marked=sql(f'SELECT delivered_at IS NOT NULL FROM webhook_outbox WHERE job_id={j}')
                return result(marked=='t' and record['status']==500,f"sink status={record['status']}; delivered={marked}")
            if D=='D17':
                full=json.dumps(job(j)['result']);sent=record['body']['result']
                return result(len(sent)==1024 and len(full)>1024 and 'result_truncated' not in record['body'],f'stored result={len(full)} chars, webhook result={len(sent)}, no flag')
            return result(not record['signature_valid'],f"signature_valid={record['signature_valid']}")
        finally:ok('POST','/admin/mode',{'mode':'ok'},base=HOOK)
    if D=='D05':
        j=enqueue(priority=100);v=job(j)
        return result(v['priority']==100,f"accepted priority={v['priority']}")
    if D=='D06':
        a=call('POST',f'/v1/queues/{name}/jobs:batch',{'jobs':[]});b=call('POST',f'/v1/queues/{name}/jobs:batch',{'jobs':[{'kind':'manual','payload':{}}]*501})
        return result(a[0]==201 and b[0]==201 and len(b[1]['ids'])==500,f'zero status={a[0]}, 501 status={b[0]}, inserted={len(b[1].get("ids",[]))}')
    if D=='D07':
        bad=name+'/archive';status,_=call('POST','/v1/queues',{'name':bad});read,_=call('GET','/v1/queues/'+bad.replace('/','%2F'))
        return result(status==201 and read==404,f'create slash-name={status}; address it={read}')
    if D=='D08':
        j=enqueue(payload={'id':9007199254740993});got=job(j)['payload']['id']
        return result(got!=9007199254740993,f'input=9007199254740993 stored={got}')
    if D=='D09':
        j=enqueue(max_attempts=3);w=worker();states=[]
        for _ in range(4):
            l=lease(w)[0];nack(j,l['lease_token'],delay_seconds=0);states.append(job(j)['state']);wait(lambda:job(j)['state']!='retry_wait')
        return result(states==['retry_wait']*3+['dead'],f'failure states={states}; attempts={job(j)["attempts"]}')
    if D=='D10':
        enqueue(delay_seconds=3600);v=ok('GET',f'/v1/queues/{name}/stats');exact=sql(f"SELECT count(*) FROM jobs j JOIN queues q ON q.id=j.queue_id WHERE q.name='{name}' AND j.state='queued'")
        return result(v['queued']!=int(exact),f'stats queued={v["queued"]}; SQL queued={exact}')
    if D=='D11':
        # Batch gives identical app time; tuple-moving state transitions expose missing tie-break.
        v=ok('POST',f'/v1/queues/{name}/jobs:batch',{'jobs':[{'kind':'manual','payload':{}} for _ in range(12)]});ids=v['ids'];w=worker();first=lease(w)[0];nack(first['id'],first['lease_token'],delay_seconds=0);assert wait(lambda:job(first['id'])['state']=='queued');nextjob=lease(w)[0]
        return result(nextjob['id']!=ids[0],f'oldest={ids[0]}, after release next leased={nextjob["id"]}; tied creation timestamps')
    if D=='D12':
        j=enqueue(delay_seconds=3600);ok('POST',f'/v1/queues/{name}/purge',{});status,v=call('GET',f'/v1/jobs/{j}')
        return result(status==200 and v['state']=='scheduled',f'purge left delayed job status={status} state={v.get("state")}')
    if D=='D13':
        s=schedule();ok('POST',f'/v1/queues/{name}/pause',{});due_schedule(s,181);time.sleep(1.2);ok('POST',f'/v1/queues/{name}/resume',{});assert wait(lambda:len(ok('GET',f'/v1/queues/{name}/jobs')['jobs'])>0);count=len(ok('GET',f'/v1/queues/{name}/jobs')['jobs'])
        return result(count>=3,f'resume enqueued {count} missed ticks')
    if D=='D14':
        j=enqueue();ok('POST',f'/v1/jobs/{j}/cancel',{});status,_=call('POST',f'/v1/jobs/{j}/retry',{});state=job(j)['state']
        return result(status==200 and state=='queued',f'retry cancelled status={status}, state={state}')
    if D=='D15':
        j=enqueue();s=schedule();ok('DELETE',f'/v1/queues/{name}');status,v=call('GET',f'/v1/jobs/{j}');schedule_status,_=call('GET',f'/v1/schedules/{s["id"]}')
        return result(status==200 and schedule_status==200,f'deleted owner; job={status}, schedule={schedule_status}')
    if D=='D16':
        s=schedule();v=ok('PATCH',f'/v1/schedules/{s["id"]}',{'cron':'0 0 1 1 *'})
        return result(v['next_run_at']==s['next_run_at'],f'cron changed to January 1, next_run_at remains {v["next_run_at"]}')
    if D=='D18':
        j=enqueue();a=lease()[0];assert wait(lambda:job(j)['state']=='queued',4);b=lease()[0];status,_=ack(j,a['lease_token'],{'owner':'A'});later,_=ack(j,b['lease_token'],{'owner':'B'})
        return result(status==200 and later==409,f'stale A ack={status}; current B ack={later}; result={job(j)["result"]}')
    if D=='D19':
        j=enqueue();l=lease()[0];ok('POST',f'/v1/jobs/{j}/cancel',{});nack(j,l['lease_token'],delay_seconds=3600);state=job(j)['state']
        return result(state=='retry_wait',f'cancel then nack leaves {state}')
    if D=='D20':
        j=enqueue();l=lease()[0];sql(f'UPDATE jobs SET attempts=max_attempts WHERE id={j}');assert wait(lambda:job(j)['state']!='leased',4);state=job(j)['state']
        return result(state=='queued',f'lease expired at attempts=max_attempts; state={state}')
    if D=='D21':
        j=enqueue();status,v=call('GET',f'/v1/jobs/{j}',key='globex-key');cancel,_=call('POST',f'/v1/jobs/{j}/cancel',{},key='globex-key')
        return result(status==200 and cancel==200,f'globex read acme={status}, cancel={cancel}')
    if D=='D22':
        # Stop only this project's database; restore service even when the observation fails.
        try:
            subprocess.run(COMPOSE+['stop','postgres'],check=True,stdout=subprocess.DEVNULL);time.sleep(2)
            subprocess.run(COMPOSE+['start','postgres'],check=True,stdout=subprocess.DEVNULL)
            for _ in range(40):
                try:sql('SELECT 1');break
                except subprocess.CalledProcessError:time.sleep(.25)
            j=enqueue(delay_seconds=1);time.sleep(3);state=job(j)['state'];health=ok('GET','/health')
            return result(state=='scheduled' and health['status']=='ok',f'after DB outage due job state={state}, health={health["status"]}')
        finally:
            subprocess.run(COMPOSE+['start','postgres'],stdout=subprocess.DEVNULL)
            subprocess.run(COMPOSE+['restart','api'],check=True,stdout=subprocess.DEVNULL);wait(lambda:health_ready(),10)
    if D=='D23':
        now=dt.datetime.now(dt.timezone.utc);zone=zoneinfo.ZoneInfo('America/New_York');current=now.astimezone(zone).utcoffset();month=1 if current==dt.timedelta(hours=-4) else 7
        s=schedule(cron=f'0 2 1 {month} *',timezone='America/New_York');next_at=dt.datetime.fromisoformat(s['next_run_at'].replace('Z','+00:00'));local=next_at.astimezone(zone)
        return result(local.hour!=2,f'cron requests 02:00 America/New_York; next fire local={local.isoformat()}')
    if D=='D25':
        status,v=call('POST',f'/v1/queues/{name}/lease',{'worker_id':999999999,'max_jobs':1})
        return result(status==200 and 'error' in v,f'unknown worker status={status}, body={v}')
    if D=='D26':
        status,v=call('PATCH',f'/v1/queues/{name}',{'lease_secs':99});enqueue();filter_status,filtered=call('GET',f'/v1/queues/{name}/jobs?state=queded')
        return result(status==200 and v['lease_seconds']==30 and filter_status==200 and len(filtered['jobs'])==1,f'typo patch={status} unchanged lease={v["lease_seconds"]}; invalid state={filter_status} returns job')
    if D=='D27':
        s=schedule();ok('PATCH',f'/v1/schedules/{s["id"]}',{'enabled':False});due_schedule(s,1);seen=wait(lambda:ok('GET',f'/v1/queues/{name}/jobs')['jobs'])
        return result(bool(seen),f'disabled schedule produced {len(seen) if seen else 0} jobs')
    raise RuntimeError('unknown proof '+D)
def health_ready():
    try:return call('GET','/health')[0]==200
    except Exception:return False
try:sys.exit(run())
except Exception as e:
    print('PROBE ERROR: '+D+' — '+repr(e),file=sys.stderr);sys.exit(2)
