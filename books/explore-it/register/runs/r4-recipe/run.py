# Rehearse the landing page's corrected run recipe (end review R4), step by step as written,
# in a fresh clone, recording the source commit and the built image at every step.
import os,subprocess,time,json,pathlib,datetime
base=pathlib.Path('/tmp/ei-fixes/r4'); logs=base/'logs'; ledger=base/'acceptance.json'
final=base/'repo/books/explore-it/project'; tagtree=pathlib.Path('/tmp/ei-tag'); tagproj=tagtree/'books/explore-it/project'
env=os.environ.copy(); env.update(COMPOSE_PROJECT_NAME='queuelet-r4',QUEUELET_IMAGE_TAG='r4',QUEUELET_API_PORT='18290',QUEUELET_HOOK_PORT='18291',QUEUELET_PG_PORT='15523',QUEUELET_URL='http://localhost:18290',QUEUELET_HOOK_URL='http://localhost:18291',QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15523/queuelet?sslmode=disable')
def out(cmd,cwd=None): return subprocess.run(cmd,cwd=cwd,capture_output=True,text=True).stdout.strip()
def identity():
  img=out(['docker','image','inspect','queuelet:r4','--format','{{.Id}}'])
  api=out(['docker','ps','--filter','label=com.docker.compose.project=queuelet-r4','--filter','label=com.docker.compose.service=api','--format','{{.ID}}'])
  return {'image_queuelet_r4':img,'api_container_image':out(['docker','inspect',api,'--format','{{.Image}}']) if api else ''}
def run(name,cwd,cmd):
  log=logs/(name+'.log'); start=time.monotonic(); at=datetime.datetime.now(datetime.timezone.utc).isoformat()
  sha=out(['git','rev-parse','HEAD'],cwd)
  with log.open('w') as f:
    f.write(json.dumps({'command':cmd,'cwd':str(cwd),'commit':sha,'start':at})+'\n'); f.flush()
    p=subprocess.run(cmd,cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT)
    el=round(time.monotonic()-start,3); f.write(f'\nEXIT {p.returncode}; ELAPSED {el}s\n')
  verdict=''
  for line in log.read_text().splitlines():
    if line.startswith(('FIXED AND CHECKED','DEFECTS REPRODUCED','BASELINE HOLDS','VERIFICATION FAILED','PROBE ERROR','ok ','FAIL','--- FAIL')): verdict=line
  rec={'name':name,'cwd':str(cwd),'commit':sha,'command':cmd,'start':at,'exit':p.returncode,'elapsed_seconds':el,'last_verdict_line':verdict,'after':identity(),'log':'logs/'+log.name}
  data=json.loads(ledger.read_text()) if ledger.exists() else {'environment':{k:v for k,v in env.items() if k.startswith(('QUEUELET_','COMPOSE_'))},'runs':[]}
  data['runs'].append(rec); ledger.write_text(json.dumps(data,indent=2)+'\n')
  print(name,p.returncode,el,verdict,rec['after'],flush=True); return p.returncode
try:
  # Block 1: the final checkout.
  run('01-final-up',final,['make','up']); run('02-final-seed',final,['make','seed'])
  run('03-final-verify00',final,['make','verify','LESSON=00'])
  run('04-final-verify07-ref06',final,['make','verify','LESSON=07','REF=explore-it/lesson-06'])
  run('05-final-verify07',final,['make','verify','LESSON=07'])
  run('06-final-checks',final,['make','checks'])
  run('07-final-down',final,['make','down'])
  # Block 2: lesson-07 at its own tag, then back to a rebuilt final stack.
  run('08-worktree-add',final,['git','worktree','add','--detach',str(tagtree),'explore-it/lesson-07'])
  run('09-tag-up',tagproj,['make','up']); run('10-tag-seed',tagproj,['make','seed'])
  run('11-tag-verify07',tagproj,['make','verify','LESSON=07'])
  run('12-tag-down',tagproj,['make','down'])
  run('13-final-up',final,['make','up']); run('14-final-seed',final,['make','seed'])
  run('15-final-checks',final,['make','checks'])
  run('16-final-down',final,['make','down'])
  run('17-worktree-remove',final,['git','worktree','remove',str(tagtree)])
finally:
  subprocess.run(['docker','compose','-p','queuelet-r4','down','-v','--remove-orphans'],capture_output=True)
