#!/usr/bin/env python3
import json,os,urllib.request,datetime
base=os.getenv('QUEUELET_URL','http://localhost:18200')
def call(path,body,key='acme-key'):
    req=urllib.request.Request(base+path,json.dumps(body).encode(),{'Authorization':'Bearer '+key,'Content-Type':'application/json'})
    with urllib.request.urlopen(req) as r:return json.load(r)
queues=[]
for name,lease,retries in [('emails',10,3),('reports',20,4),('imports',30,2)]:
    body={'name':name,'lease_seconds':lease,'max_attempts':retries,'backoff':{'kind':'exponential','base_seconds':2}}
    if name=='reports':body['webhook']={'url':'http://hooksink:8201/hook','secret':'report-secret'}
    queues.append(call('/v1/queues',body))
call('/v1/queues',{'name':'emails'},'globex-key')
ids=[]
for i in range(300):
    q=['emails','reports','imports'][i%3]
    body={'kind':'echo','payload':{'fixture':i},'priority':i%10}
    if i%4==0:body['delay_seconds']=3600
    ids.append(call('/v1/queues/'+q+'/jobs',body)['id'])
# Exercise the public worker protocol to leave representative terminal and retry states.
w=call('/v1/workers',{'name':'fixture-writer','kinds':['echo']})['worker_id']
for q in ['emails','reports','imports']:
    leased=call('/v1/queues/'+q+'/lease',{'worker_id':w,'max_jobs':25})['jobs']
    for i,j in enumerate(leased):
        if i%4==0:call('/v1/jobs/%s/ack'%j['id'],{'lease_token':j['lease_token'],'result':{'fixture':'done'}})
        elif i%4==1:call('/v1/jobs/%s/nack'%j['id'],{'lease_token':j['lease_token'],'error':'fixture retry','delay_seconds':3600})
        elif i%4==2:call('/v1/jobs/%s/nack'%j['id'],{'lease_token':j['lease_token'],'error':'fixture dead','retry':False})
        else:call('/v1/jobs/%s/cancel'%j['id'],{})
schedules=[]
for name,zone,cron in [('daily-report','America/New_York','0 9 * * *'),('import-pulse','UTC','*/15 * * * *')]:
    schedules.append(call('/v1/schedules',{'name':name,'queue':'imports','kind':'echo','payload':{'schedule':name},'cron':cron,'timezone':zone,'enabled':True}))
with open('results/manifest.json','w') as f:json.dump({'api':base,'keys':{'acme':'acme-key','globex':'globex-key','admin':'admin-key'},'queues':queues,'job_ids':ids,'schedules':schedules,'seed_worker':w},f,indent=2)
print('Seeded 300 jobs, 4 queues, 2 schedules; results/manifest.json')
