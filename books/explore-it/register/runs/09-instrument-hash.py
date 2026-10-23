import hashlib,json,pathlib,sys
root=pathlib.Path('/root/Software/booknotes-work/explore-it/books/explore-it')
files=['project/Makefile','project/lab/verify.sh','project/lab/repro/probe.py','project/lab/repro/probe04.py','project/lab/repro/probe09.py','project/lab/repro/lesson-09.ids','project/test/checks/lesson09_test.go','project/internal/jobs/transition_test.go']+[f'project/lab/repro/{i}.sh' for i in 'U39 U40 U41 U42 U43 U44 U45 D12 D14 D18 D19 D20 D27'.split()]
h={f:hashlib.sha256((root/f).read_bytes()).hexdigest() for f in sorted(files)}
out=root/'register/runs/09-instrument-hashes.json'
if sys.argv[1]=='freeze': out.write_text(json.dumps(h,indent=2)+'\n'); print(len(h),'frozen')
else:
    old=json.loads(out.read_text()); print(json.dumps({'instrument_hash_mismatches':[f for f in old if old[f]!=h.get(f)]}))
