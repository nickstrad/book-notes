import os,sys,subprocess,time,json,pathlib,datetime
base=pathlib.Path('/tmp/explore-it-astra-review'); logs=base/'logs'; ledger=base/'v0-acceptance.json'
env=os.environ.copy(); env.update(COMPOSE_PROJECT_NAME='queuelet-astra-v0',QUEUELET_IMAGE_TAG='astrav0',QUEUELET_API_PORT='18282',QUEUELET_HOOK_PORT='18283',QUEUELET_PG_PORT='15514',QUEUELET_URL='http://localhost:18282',QUEUELET_HOOK_URL='http://localhost:18283',QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15514/queuelet?sslmode=disable')
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
  if run('v0-up','v0',['make','up']) != 0: raise RuntimeError('startup failed')
  run('v0-baseline','v0',['make','verify','LESSON=00'])
  for d in ['D18','D21']:
    run('v0-'+d,'repo',['bash','lab/repro/'+d+'.sh'])
finally:
  run('v0-down','v0',['make','down'])
