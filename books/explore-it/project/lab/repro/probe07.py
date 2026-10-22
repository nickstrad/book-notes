#!/usr/bin/env python3
"""E07 public races and actual target-worker cooperation; no fixture SQL writes.
Bounded samples are evidence, not a guarantee against every possible schedule.
Every worker, HTTP exchange and trial is logged to stderr, including failures.
"""
import concurrent.futures as futures
import collections
import http.server
import json
import os
import pathlib
import subprocess
import sys
import tempfile
import threading
import time
import uuid
import probe as p
from probe04 import clock


def evidence(label, **values):
    print(json.dumps(dict(probe=p.ID, label=label, **values), sort_keys=True), file=sys.stderr, flush=True)


def race(*functions):
    gate = threading.Barrier(len(functions))
    def run(f):
        gate.wait(timeout=10)
        return f()
    with futures.ThreadPoolExecutor(max_workers=len(functions)) as pool:
        pending = [pool.submit(run, f) for f in functions]
        return [f.result(timeout=30) for f in pending]


def delayed(seconds, f):
    def run():
        time.sleep(seconds)
        return f()
    return run


def queue(suffix, n=1, **options):
    name = p.NAME + '-' + suffix
    q = p.ok('POST', '/v1/queues', dict(name=name, lease_seconds=300, max_attempts=3,
        backoff={'kind':'linear', 'base_seconds':3600}, **options))
    ids = p.ok('POST', f'/v1/queues/{name}/jobs:batch', {'jobs':[{'kind':'manual','payload':{}} for _ in range(n)]})['ids'] if n else []
    return name, q['id'], ids


def worker():
    return p.ok('POST', '/v1/workers', {'name':p.NAME,'kinds':['manual']})['worker_id']


def lease(name, w, count=100):
    return p.checked('POST', f'/v1/queues/{name}/lease', {'worker_id':w,'max_jobs':count}, statuses=(200,404,500))


def rows(qid):
    return json.loads(p.sql(f"SELECT COALESCE(json_agg(v ORDER BY id),'[]') FROM (SELECT j.*, (SELECT COALESCE(json_agg(a ORDER BY n),'[]') FROM attempts a WHERE a.job_id=j.id) AS attempt_history FROM jobs j WHERE queue_id={qid}) v"))


def attempts(j):
    return p.ok('GET', f'/v1/jobs/{j}/attempts')['attempts']


def clean(name, qid):
    for j in rows(qid):
        if j['state'] == 'leased': p.ok('POST', f'/v1/jobs/{j["id"]}/cancel', {})
    p.ok('POST', f'/v1/queues/{name}/purge', {})


def released(j, w, outcome, state):
    p.require(j['state'] == state and j['attempts'] == 1, ('release state/budget', j))
    p.require(all(j[k] is None for k in ('leased_by','lease_token','lease_expires_at')), ('release lease', j))
    a = j['attempt_history'] if 'attempt_history' in j else attempts(j['id'])
    p.require(len(a) == 1 and a[0]['n'] == 1 and a[0]['worker_id'] == w and a[0]['outcome'] == outcome, ('release history',a))


def worker_delete():
    # Independent positive sequential control, including the next attempt.
    name, qid, ids = queue('control')
    w = worker(); c, b = lease(name,w)
    p.require(c == 200 and b['jobs'][0]['id'] == ids[0], b)
    p.checked('DELETE', f'/v1/workers/{w}')
    released(p.job(ids[0]), w, 'released', 'queued')
    c, missing = lease(name,w)
    p.require(c == 404 and missing['error']['code'] == 'worker_not_found', missing)
    c, b = lease(name,worker()); item = b['jobs'][0]
    p.require(c == 200 and item['attempt'] == 2, b)
    p.require(p.ack(item['id'],item['lease_token'])[0] == 200, 'ack control')
    p.require([a['outcome'] for a in attempts(item['id'])] == ['released','succeeded'], 'control history')
    counts = collections.Counter(); zombies = 0; errors = 0
    # Several calls per worker widen the API-check/transaction window without hooks.
    for trial in range(32):
        name, qid, ids = queue(str(trial), 40)
        w = worker()
        funcs = [delayed((i % 4)*.0004, lambda:lease(name,w,10)) for i in range(12)]
        result = race(*funcs, delayed((trial % 5)*.0007, lambda:p.checked('DELETE',f'/v1/workers/{w}')))
        p.require(result[-1][0] == 200, result[-1])
        got = collections.defaultdict(list)
        live = rows(qid)
        # Preserve the response epochs and row/history evidence before any assertion.
        # One DELETE can release epoch 1 before an already-running lease takes
        # epoch 2. Repeated ids alone are not duplicate active ownership.
        evidence('worker-delete epochs', trial=trial, worker=w, responses=result,
                 rows=[{k:j[k] for k in ('id','state','attempts','leased_by','lease_token','lease_expires_at','attempt_history')} for j in live])
        p.require(p.sql(f'SELECT count(*) FROM workers WHERE id={w}') == '0', 'worker still exists')
        for c,b in result[:-1]:
            counts[c] += 1
            if c == 500:
                p.require(b['error'] == {'code':'database_error','message':'sql: no rows in result set'}, b)
                errors += 1
            elif c == 404: p.require(b['error']['code'] == 'worker_not_found', b)
            else:
                for item in b['jobs']:
                    p.require(item['id'] in ids and item['lease_token'], ('unknown/invalid lease',item))
                    got[item['id']].append(item)
        orphan = [j for j in live if j['state'] == 'leased']
        zombies += len(orphan)
        p.require(len(live) == 40, ('lost job',len(live)))
        for j in live:
            epochs = sorted(got.get(j['id'], []), key=lambda item:item['attempt'])
            p.require([item['attempt'] for item in epochs] == list(range(1,len(epochs)+1))
                      and len({item['lease_token'] for item in epochs}) == len(epochs),
                      ('duplicate ownership or invalid epoch sequence',epochs,j))
            if not epochs:
                p.require(j['state'] == 'queued' and j['attempts'] == 0 and not j['attempt_history']
                          and all(j[k] is None for k in ('leased_by','lease_token','lease_expires_at')), ('unleased mutated',j))
            elif j['state'] != 'leased':
                p.require(len(epochs) == 1, ('released multiple epochs from one deletion',epochs,j))
                released(j,w,'released','queued')
            else:
                p.require(len(epochs) in (1,2) and j['leased_by'] == w
                          and j['lease_token'] == epochs[-1]['lease_token']
                          and j['lease_expires_at'] is not None
                          and j['attempts'] == len(epochs)-1
                          and len(j['attempt_history']) == len(epochs)-1, ('zombie evidence',epochs,j))
                if len(epochs) == 2:
                    a=j['attempt_history'][0]
                    p.require(a['n'] == 1 and a['worker_id'] == w and a['outcome'] == 'released', ('epoch release evidence',epochs,j))
                    for verb, body in [('heartbeat',{'extend_seconds':300}),('ack',{'result':{}}),('nack',{'retry':True})]:
                        p.checked('POST',f'/v1/jobs/{j["id"]}/{verb}',dict(lease_token=epochs[0]['lease_token'],**body),statuses=(409,))
                    after=p.job(j['id'])
                    p.require(after['lease_token'] == j['lease_token'] and after['state'] == 'leased'
                              and after['attempts'] == 1 and len(attempts(j['id'])) == 1, ('old epoch changed current ownership',after))
        if orphan:
            item=max(got[orphan[0]['id']],key=lambda item:item['attempt'])
            c,_=p.checked('POST',f'/v1/jobs/{item["id"]}/heartbeat',{'lease_token':item['lease_token'],'extend_seconds':300})
            p.require(c == 200 and p.ack(item['id'],item['lease_token'])[0] == 200, 'historical zombie rights')
        evidence('worker-delete',trial=trial,statuses=[c for c,_ in result],zombie_ids=[j['id'] for j in orphan],errors=errors)
        clean(name,qid)
        if zombies and errors: break
    p.require(bool(zombies) == bool(errors), ('partial U34 evidence',zombies,errors,dict(counts)))
    return bool(zombies), f'zombie leases={zombies}; worker-disappearance 500={errors}; statuses={dict(counts)}; sequential release/404/next-attempt controls hold'


def purge_control():
    name,qid,ids=queue('sequential',6)
    w=worker(); c,b=lease(name,w,4)
    p.require(c == 200 and len(b['jobs']) == 4,b)
    p.require(p.ok('POST',f'/v1/queues/{name}/purge',{})['deleted'] == 2,'pending purge count')
    got=b['jobs']
    p.require(all(j['purge_requested'] for j in rows(qid)), 'leased not marked')
    p.require(p.ack(got[0]['id'],got[0]['lease_token'],{'ok':True})[0] == 200,'purged success')
    released(p.job(got[0]['id']),w,'succeeded','succeeded')
    p.require(p.nack(got[1]['id'],got[1]['lease_token'])[0] == 200,'purged nack')
    released(p.job(got[1]['id']),w,'failed','dead')
    p.checked('DELETE', f'/v1/workers/{w}')
    for j in got[2:]: released(p.job(j['id']),w,'released','dead')
    # Sequential expiry after marking, using only the documented clock.
    name,qid,ids=queue('expiry-control')
    w=worker(); c,b=lease(name,w)
    p.require(c == 200 and len(b['jobs']) == 1,b)
    p.ok('POST',f'/v1/queues/{name}/purge',{})
    clock('advance',seconds=301)
    p.wait(lambda:p.job(ids[0])['state']=='dead',8)
    released(p.job(ids[0]),w,'expired','dead')
    clock('reset')
    evidence('purge sequential controls',pending_deleted=2,success=1,failed=1,released=2,expired=1)


def purge_races():
    hits={mode:0 for mode in ('lease','nack','release','expiry')}
    purge_control()
    try:
        for mode in hits:
            for trial in range(8):
                n=300 if mode=='expiry' else 120 if mode=='lease' else 60
                name,qid,ids=queue(f'{mode}-{trial}',n)
                w=worker(); got=[]
                if mode!='lease':
                    while len(got)<n:
                        c,b=lease(name,w,min(100,n-len(got))); p.require(c==200 and b['jobs'],b); got+=b['jobs']
                purge=lambda:p.checked('POST',f'/v1/queues/{name}/purge',{})
                if mode=='lease':
                    ws=[w]+[worker() for _ in range(3)]
                    def acquire(w):
                        result=[]
                        for _ in range(6):
                            c,b=lease(name,w,10); p.require(c==200,b); result+=b['jobs']
                        return result
                    res=race(*[lambda w=w:acquire(w) for w in ws],delayed(.01+(trial%4)*.005,purge))
                    got=[j for batch in res[:-1] for j in batch]; pres=res[-1]
                elif mode=='nack':
                    res=race(*[delayed((i%10)*.003,lambda j=j:p.checked('POST',f'/v1/jobs/{j["id"]}/nack',{'lease_token':j['lease_token'],'retry':True,'error':'purge race'},statuses=(200,404,409))) for i,j in enumerate(got)], delayed(.008+(trial%4)*.003,purge))
                    p.require(all(c in (200,404,409) for c,_ in res[:-1]),res)
                    pres=res[-1]
                elif mode=='release':
                    res=race(lambda:p.checked('DELETE',f'/v1/workers/{w}'),delayed((trial%4)*.003,purge)); pres=res[-1]
                else:
                    clock('advance',seconds=301)
                    time.sleep((trial%6)*.12)
                    pres=purge()
                    p.wait(lambda:all(j['state']!='leased' for j in rows(qid)),10)
                live=rows(qid)
                bad=[j for j in live if j['state'] in ('queued','retry_wait','scheduled') or (j['state']=='leased' and not j['purge_requested'])]
                hits[mode]+=bool(bad)
                p.require(pres[0]==200 and len(live)+pres[1]['deleted']==n,('purge accounting',mode,pres,len(live),n))
                tokens={j['id']:j['lease_token'] for j in got}
                if mode=='lease':
                    p.require(all(j['state']=='leased' for j in live),('lease race state',live))
                    # All survivors must be attributable real leases. Nack proves the flag's effect.
                    holders={j['id']:j['leased_by'] for j in live}
                    for j in live:
                        p.require(j['id'] in tokens and j['attempts']==0 and not j['attempt_history'],j)
                    with futures.ThreadPoolExecutor(max_workers=12) as pool:
                        results=list(pool.map(lambda j:p.nack(j['id'],tokens[j['id']]),live))
                    p.require(all(code==200 for code,_ in results),'nack survivor')
                    after=rows(qid)
                    p.require({j['id'] for j in after}==set(holders),'survivor missing')
                    for j in after:
                        released(j,holders[j['id']],'failed','dead' if j['purge_requested'] else 'retry_wait')
                else:
                    outcome={'nack':'failed','release':'released','expiry':'expired'}[mode]
                    for j in live: released(j,w,outcome,'dead' if j['purge_requested'] else ('retry_wait' if mode=='nack' else 'queued'))
                evidence('purge race',mode=mode,trial=trial,deleted=pres[1]['deleted'],survivors=len(live),bad_ids=[j['id'] for j in bad],states=dict(collections.Counter(j['state'] for j in live)))
                clean(name,qid)
                if mode=='expiry': clock('reset')
                if hits[mode]>=2: break
        p.require(all(v>0 for v in hits.values()) or not any(hits.values()),('partial U35 evidence',hits))
        return any(hits.values()),f'purge race hits={hits}; all four paths sampled; deletion counts, attempt attribution, state/lease effects and sequential success/failure/expiry controls hold'
    finally: clock('reset')


class StreamingPartner(http.server.ThreadingHTTPServer):
    daemon_threads=True
    def __init__(self):
        super().__init__(('127.0.0.1',0),Stream)
        self.started=threading.Event(); self.ended=threading.Event(); self.ticks=[]; self.start_at=None; self.end_at=None


class Stream(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        s=self.server
        self.send_response(200); self.send_header('Content-Type','application/json'); self.end_headers()
        s.start_at=time.monotonic(); s.started.set()
        try:
            # These remote effects cannot be undone; socket closure stops future ticks.
            for _ in range(100):
                self.wfile.write(b' ' * 1024); self.wfile.flush(); s.ticks.append(time.monotonic()); time.sleep(.1)
        except (BrokenPipeError,ConnectionResetError): pass
        finally: s.end_at=time.monotonic(); s.ended.set()
    def log_message(self,*args): pass


def cooperation():
    clock('reset')
    config=json.loads(subprocess.check_output(p.COMPOSE+['config','--format','json'],text=True))
    image=config['services']['worker-a']['image']
    results={}
    for kind in ('sleep','http'):
        name,qid,_=queue('worker-'+kind,0)
        partner=StreamingPartner(); thread=threading.Thread(target=partner.serve_forever,daemon=True); thread.start()
        payload={'seconds':8} if kind=='sleep' else {'url':f'http://127.0.0.1:{partner.server_port}/stream'}
        jid=p.ok('POST',f'/v1/queues/{name}/jobs',{'kind':kind,'payload':payload})['id']
        container='queuelet-e07-'+uuid.uuid4().hex[:12]
        command=['docker','run','--rm','--network','host','--name',container,'-e','QUEUELET_URL='+p.BASE,'-e','QUEUELET_KEY=acme-key','-e','WORKER_QUEUE='+name,'-e','WORKER_NAME='+container,image,'worker']
        with tempfile.TemporaryFile(mode='w+') as log:
            process=subprocess.Popen(command,stdout=log,stderr=log,stdin=subprocess.DEVNULL)
            def events():
                log.seek(0); lines=log.read().splitlines(); result=[]
                for line in lines:
                    try: result.append(json.loads(line))
                    except json.JSONDecodeError: pass
                return result
            try:
                p.wait(lambda:any(e.get('event')=='processing' and e.get('job_id')==jid for e in events()),10)
                if kind=='http': p.require(partner.started.wait(3),'worker HTTP never started')
                before=p.job(jid); token=before['lease_token']; holder=before['leased_by']
                cancelled_at=time.monotonic(); p.ok('POST',f'/v1/jobs/{jid}/cancel',{})
                c,wire=p.checked('POST',f'/v1/jobs/{jid}/heartbeat',{'lease_token':token,'extend_seconds':30},statuses=(409,))
                for verb,body in [('heartbeat',{'extend_seconds':30}),('ack',{'result':{}}),('nack',{'retry':True})]:
                    for tok in (token,'wrong-token'):
                        _,reply=p.checked('POST',f'/v1/jobs/{jid}/{verb}',dict(lease_token=tok,**body),statuses=(409,))
                        if tok!='wrong-token' and verb=='heartbeat': continue
                        p.require(not reply.get('cancel_requested'),('leaked signal',reply))
                quiet=p.ok('POST',f'/v1/queues/{name}/jobs',{'kind':'echo','payload':{'quiet':kind}})['id']
                p.wait(lambda:p.job(quiet)['state']=='succeeded',12)
                quiet_elapsed=time.monotonic()-cancelled_at
                completion=p.wait(lambda:next((e for e in events() if e.get('job_id')==jid and e.get('event') in ('ack','nack')),None),3)
                final=p.job(jid)
                p.require(final['state']=='cancelled' and all(final[k] is None for k in ('lease_token','leased_by','lease_expires_at')),final)
                p.require(p.job(quiet)['result']=={'quiet':kind} and p.job(quiet)['attempts']==1,'quiet job failed')
                p.require(completion['status']==409,completion)
                observed=any(e.get('event')=='cancel_requested' and e.get('job_id')==jid for e in events())
                requested=wire.get('cancel_requested') is True
                p.require(observed==requested,('worker did not observe signal',wire,events()))
                a=attempts(jid)
                if requested:
                    released(final,holder,'cancelled','cancelled')
                    p.require(quiet_elapsed<4 and completion['event']=='nack',('work not interrupted',kind,quiet_elapsed,completion))
                else:
                    p.require(final['attempts']==0 and not a,('unexpected historical attempt',final,a))
                    p.require(quiet_elapsed>=(7 if kind=='sleep' else 4),('no continuing-work evidence',kind,quiet_elapsed))
                remote=None
                if kind=='http':
                    p.require(partner.ended.wait(3),'remote exchange never ended')
                    after_ticks=sum(t>cancelled_at+2.5 for t in partner.ticks)
                    p.require((after_ticks==0) if requested else (after_ticks>0),('remote continuation evidence',after_ticks))
                    remote={'ticks':len(partner.ticks),'ticks_after_cancel_2_5s':after_ticks,'end_after_cancel':round(partner.end_at-cancelled_at,3)}
                evidence('actual worker',kind=kind,image=image,command=command,heartbeat=wire,completion=completion,quiet_after_seconds=round(quiet_elapsed,3),remote=remote)
                results[kind]=requested
            finally:
                # Cleanup only, after evidence; never counted as cooperative stopping.
                subprocess.run(['docker','stop','-t','2',container],capture_output=True,timeout=10)
                try: process.wait(timeout=10)
                except subprocess.TimeoutExpired: process.kill(); process.wait()
                log.seek(0); print('WORKER LOG '+kind+'\n'+log.read(),file=sys.stderr,flush=True)
                partner.shutdown(); partner.server_close(); thread.join(timeout=2)
    p.require(len(set(results.values()))==1,('partial A3 evidence',results))
    return not all(results.values()),f'actual image={image}; cooperative sleep/HTTP={results}; revoked heartbeat/late completion 409; remote tick/worker completion evidence retained; unrelated echo jobs completed'


if __name__=='__main__':
    try:
        present,detail={'U34':worker_delete,'U35':purge_races,'A3':cooperation}[p.ID]()
        print(('PRESENT: ' if present else 'ABSENT: ')+p.ID+' — '+detail)
    except Exception as error:
        evidence('PROBE ERROR',error=repr(error))
        raise
