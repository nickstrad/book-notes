import os, subprocess, pathlib, time, json, datetime
base=pathlib.Path('/tmp/explore-it-checkpoint-b'); repo=base/'repo'; logs=base/'logs'; logs.mkdir(exist_ok=True)
source='/root/Software/booknotes-work/explore-it'

project=repo/'books/explore-it/project'
env=os.environ.copy(); env.update(COMPOSE_PROJECT_NAME='queuelet-review-b',QUEUELET_IMAGE_TAG='review-b',QUEUELET_API_PORT='18260',QUEUELET_HOOK_PORT='18261',QUEUELET_PG_PORT='15493',QUEUELET_URL='http://localhost:18260',QUEUELET_HOOK_URL='http://localhost:18261')
env['QUEUELET_DATABASE_URL']='postgres://queuelet:queuelet@localhost:15493/queuelet?sslmode=disable'
results=json.loads((base/'acceptance.json').read_text())
def run(label,args,required=True):
    start=time.monotonic(); stamp=datetime.datetime.now(datetime.timezone.utc).isoformat()
    print('START',label,flush=True)
    with (logs/(label+'.log')).open('w') as f:
        p=subprocess.run(args,cwd=project,env=env,stdout=f,stderr=subprocess.STDOUT)
    row=dict(label=label,command=args,started=stamp,elapsed_seconds=round(time.monotonic()-start,3),exit=p.returncode,log='logs/'+label+'.log')
    results.append(row); (base/'acceptance.json').write_text(json.dumps(results,indent=2)+'\n')
    print('END',label,p.returncode,row['elapsed_seconds'],flush=True)
    if required and p.returncode: raise RuntimeError(label)
try:
    run('head-up-retry',['make','up'])
    for n in ['02','03','04','05','06']: run('head'+n+'-verify-retry',['make','verify','LESSON='+n])
    for n,prev in [('02','01'),('03','02'),('04','03'),('05','04'),('06','05')]:
        run('ref'+n+'-verify',['make','verify','LESSON='+n,'REF=explore-it/lesson-'+prev])
    run('head-checks',['make','checks'])
    run('head-test',['make','test'])
    run('head-vet',['go','vet','./...'])
    run('head-A3-open',['bash','../register/proof/A3.sh'])
finally:
    run('final-down-retry',['make','down'],False)
