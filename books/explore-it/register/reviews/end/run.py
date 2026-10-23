import os,sys,subprocess,time,json,pathlib,datetime
base=pathlib.Path('/tmp/explore-it-astra-review'); logs=base/'logs'; ledger=base/'acceptance.json'
env=os.environ.copy(); env.update(COMPOSE_PROJECT_NAME='queuelet-astra',QUEUELET_IMAGE_TAG='astra',QUEUELET_API_PORT='18280',QUEUELET_HOOK_PORT='18281',QUEUELET_PG_PORT='15513',QUEUELET_URL='http://localhost:18280',QUEUELET_HOOK_URL='http://localhost:18281',QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15513/queuelet?sslmode=disable')
def run(name,tree,cmd):
  cwd=base/tree/'books/explore-it/project'; log=logs/(name+'.log')
  start=time.monotonic(); at=datetime.datetime.now(datetime.timezone.utc).isoformat()
  sha=subprocess.run(['git','rev-parse','HEAD'],cwd=cwd,capture_output=True,text=True).stdout.strip()
  with log.open('w') as f:
    f.write(json.dumps({'command':cmd,'tree':tree,'commit':sha,'environment':{k:v for k,v in env.items() if k.startswith(('QUEUELET_','COMPOSE_'))},'start':at})+'\n'); f.flush()
    p=subprocess.run(cmd,cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT)
    el=round(time.monotonic()-start,3); f.write(f'\nEXIT {p.returncode}; ELAPSED {el}s\n')
  verdict=''
  for line in log.read_text().splitlines():
    if line.startswith(('FIXED AND CHECKED','DEFECTS REPRODUCED','BASELINE HOLDS','VERIFICATION FAILED','PROBE ERROR','ok ','FAIL','--- FAIL')): verdict=line
  data=json.loads(ledger.read_text()) if ledger.exists() else {'runs':[]}
  data['runs'].append({'name':name,'tree':tree,'commit':sha,'command':cmd,'start':at,'exit':p.returncode,'elapsed_seconds':el,'last_verdict_line':verdict,'log':'logs/'+log.name}); ledger.write_text(json.dumps(data,indent=2)+'\n')
  print(name,p.returncode,el,verdict,flush=True); return p.returncode

try:
  if run('head-up','repo',['make','up']) != 0: raise RuntimeError('startup failed')
  run('head-checks','repo',['make','checks'])
  run('head-test','repo',['go','test','-count=1','./...'])
  run('head-vet','repo',['go','vet','./...'])
  for l,r in [('08','07'),('09','08')]:
    run('ref'+l,'repo',['make','verify','LESSON='+l,'REF=explore-it/lesson-'+r])
finally:
  run('head-down','repo',['make','down'])
