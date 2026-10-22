import pathlib,subprocess,json,tarfile,io,hashlib
root=pathlib.Path('/root/Software/booknotes-work/explore-it'); book=root/'books/explore-it'
commits={2:'e390c31',3:'09dd0a9',4:'dffb1ec',5:'01137b6',6:'b3b0559'}
report={}
for n,ref in commits.items():
    man=json.loads((book/f'register/runs/{n:02}-artifact-manifest.json').read_text())['artifacts_copied_without_rewriting']
    raw=subprocess.check_output(['git','archive',ref,'books/explore-it/exploration'],cwd=root)
    with tarfile.open(fileobj=io.BytesIO(raw)) as tar:
        data={m.name.removeprefix('books/explore-it/'):tar.extractfile(m).read() for m in tar if m.isfile()}
    bad=[x['file'] for x in man if x['file'] not in data or hashlib.sha256(data[x['file']]).hexdigest()!=x['sha256']]
    # Living models/backlog/index may evolve. Original individual reports and experiment artifacts may not.
    stable=[x for x in man if x['file'].startswith('exploration/experiments/') or (x['file'].startswith('exploration/defects/') and not x['file'].endswith(('README.md','INDEX.md')))]
    headbad=[x['file'] for x in stable if not (book/x['file']).exists() or hashlib.sha256((book/x['file']).read_bytes()).hexdigest()!=x['sha256']]
    report[f'E{n:02}']={'experiment_commit':ref,'manifest_entries':len(man),'manifest_mismatches_at_experiment':bad,'stable_entries_checked_at_HEAD':len(stable),'stable_HEAD_mismatches':headbad}
# Verify frozen public repros against HEAD; earlier unit/check files legitimately grow.
for n in [1,2,4,5,6]:
    p=book/f'register/runs/{n:02}-instruments.sha256'; changed=[]; checked=0
    for line in p.read_text().splitlines():
        h,path=line.split(maxsplit=1); path=path.lstrip('*'); path=path.removeprefix('books/explore-it/project/'); target=book/'project'/path
        if not path.startswith('lab/repro/'):continue
        checked+=1
        if not target.exists() or hashlib.sha256(target.read_bytes()).hexdigest()!=h:changed.append(path)
    report[f'public_instruments_{n:02}']={'checked':checked,'mismatches':changed}
print(json.dumps(report,indent=2))
