import os,sys,subprocess,time,json,pathlib,datetime
base=pathlib.Path('/tmp/explore-it-checkpoint-c'); logs=base/'logs'; ledger=base/'acceptance.json'
env=os.environ.copy(); env.update(COMPOSE_PROJECT_NAME='queuelet-review-c',QUEUELET_IMAGE_TAG='reviewc',QUEUELET_API_PORT='18270',QUEUELET_HOOK_PORT='18271',QUEUELET_PG_PORT='15503',QUEUELET_URL='http://localhost:18270',QUEUELET_HOOK_URL='http://localhost:18271',QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15503/queuelet?sslmode=disable')
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
for t in ['07','08','09']:
  run(f'tag{t}-up',f'tag{t}',['make','up'])
  run(f'tag{t}-verify',f'tag{t}',['make','verify',f'LESSON={t}'])
  run(f'tag{t}-down',f'tag{t}',['make','down'])
run('head-up','repo',['make','up'])
for l in ['00','01','02','03','04','05','06','07','08','09']:
  run(f'head{l}-verify','repo',['make','verify',f'LESSON={l}'])
for l,r in [('07','06'),('08','07'),('09','08')]:
  run(f'ref{l}-verify','repo',['make','verify',f'LESSON={l}',f'REF=explore-it/lesson-{r}'])
run('head-checks','repo',['make','checks'])
run('head-test','repo',['make','test'])
run('head-vet','repo',['go','vet','./...'])
run('head-down','repo',['make','down'])
