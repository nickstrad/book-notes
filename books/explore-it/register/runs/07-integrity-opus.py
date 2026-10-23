import pathlib,subprocess,hashlib,json
root=pathlib.Path('/root/Software/booknotes-work/explore-it')
b=root/'books/explore-it'
def digest(p): return hashlib.sha256(p.read_bytes()).hexdigest()
frozen=json.loads((b/'register/runs/07-instrument-hashes-opus.json').read_text())
mismatch=[s for s,h in frozen.items() if digest(b/s)!=h]
raw=[p for p in (b/'exploration/experiments/07-state-it').rglob('*') if p.is_file()]
raw+=list((b/'exploration/defects').glob('E07-*.md'))
raw += [b/'exploration/models/states.md',b/'exploration/hypotheses.md']
raw_mismatch=[str(p.relative_to(b)) for p in raw if subprocess.check_output(['git','show','49d8d67:'+str(p.relative_to(root))],cwd=root)!=p.read_bytes()]
old=json.loads((b/'register/runs/07-integrity-before.json').read_text())['old_tags']
tag_mismatch=[t for t,h in old.items() if subprocess.check_output(['git','rev-parse',t],cwd=root,text=True).strip()!=h]
tracked=subprocess.check_output(['git','ls-tree','-r','--name-only','49d8d67','books/explore-it/project/lab/repro','books/explore-it/project/test/checks'],cwd=root,text=True).splitlines()
inherited=[s for s in tracked if not s.endswith('.md')]
inherited_mismatch=[s for s in inherited if subprocess.check_output(['git','show','49d8d67:'+s],cwd=root)!=(root/s).read_bytes()]
r={'frozen_files':len(frozen),'hash_mismatches':mismatch,'raw_files':len(raw),'raw_mismatches':raw_mismatch,'old_tags':old,'tag_mismatches':tag_mismatch,'inherited_instruments':len(inherited),'inherited_mismatches':inherited_mismatch}
print(json.dumps(r,indent=2))
assert not any([mismatch,raw_mismatch,tag_mismatch,inherited_mismatch])
