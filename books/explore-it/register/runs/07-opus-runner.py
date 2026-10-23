import os,sys,subprocess,time,json,pathlib,datetime
root=pathlib.Path('/root/Software/booknotes-work/explore-it/books/explore-it')
env=os.environ.copy(); env.update(COMPOSE_PROJECT_NAME='queuelet-fix07',QUEUELET_IMAGE_TAG='fix07',QUEUELET_API_PORT='18290',QUEUELET_HOOK_PORT='18291',QUEUELET_PG_PORT='15523',QUEUELET_URL='http://localhost:18290',QUEUELET_HOOK_URL='http://localhost:18291',QUEUELET_DATABASE_URL='postgres://queuelet:queuelet@localhost:15523/queuelet?sslmode=disable')
name=sys.argv[1]; cmd=sys.argv[2:]; log=root/'register/runs'/('07-opus-'+name+'.log'); ledger=root/'register/runs/07-opus-acceptance.json'
start=time.monotonic(); at=datetime.datetime.now(datetime.timezone.utc).isoformat()
with log.open('w') as f:
 f.write(json.dumps({'command':cmd,'cwd':str(root/'project'),'environment':{k:v for k,v in env.items() if k.startswith(('QUEUELET_','COMPOSE_'))},'start':at})+'\n'); f.flush()
 p=subprocess.run(cmd,cwd=root/'project',env=env,stdout=f,stderr=subprocess.STDOUT)
 elapsed=round(time.monotonic()-start,3); f.write(f'\nEXIT {p.returncode}; ELAPSED {elapsed}s\n')
data=json.loads(ledger.read_text()) if ledger.exists() else {'runs':[]}
data['runs'].append({'name':name,'command':cmd,'start':at,'exit':p.returncode,'elapsed_seconds':elapsed,'log':log.name}); ledger.write_text(json.dumps(data,indent=2)+'\n')
print(name,p.returncode,elapsed,log)
if p.returncode: print(log.read_text()[-6000:])
sys.exit(p.returncode)
