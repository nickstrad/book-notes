import json, hashlib, pathlib, subprocess, sys
root = pathlib.Path('/tmp/explore-it-review06/source')
man = json.loads((root/'books/explore-it/register/runs/06-artifact-manifest.json').read_text())
arts = man['artifacts_copied_without_rewriting']
print('manifest entries:', len(arts))
bad=[]; missing=[]
for a in arts:
    p = root/'books/explore-it'/a['file']
    if not p.exists():
        missing.append(a['file']); continue
    h=hashlib.sha256(p.read_bytes()).hexdigest()
    if h!=a['sha256']: bad.append((a['file'],h,a['sha256']))
print('missing:',missing)
print('mismatches:',len(bad))
for b in bad[:20]: print(b)
# instruments
for name in ['06-instruments.sha256','06-instruments-pair1.sha256','06-instruments-attempt1.sha256']:
    p=root/'books/explore-it/register/runs'/name
    lines=p.read_text().strip().splitlines()
    print(name, len(lines), 'sha256 of file:', hashlib.sha256(p.read_bytes()).hexdigest())
# verify final instruments against project cwd semantics: paths in manifest?
print(open(root/'books/explore-it/register/runs/06-instruments.sha256').readlines()[:5])
# immutability: explorer artifacts should be unchanged b3b0559..lesson-06 except allowed U*/INDEX
out=subprocess.check_output(['git','diff','--name-only','b3b0559..explore-it/lesson-06','--','books/explore-it/exploration/'],cwd=root,text=True)
print('exploration diff b3b0559..lesson06:\n',out)
# exploration commit parent check
print(subprocess.check_output(['git','rev-list','--parents','-n','1','explore-it/lesson-06'],cwd=root,text=True))
print(subprocess.check_output(['git','rev-list','--parents','-n','1','b3b0559'],cwd=root,text=True))
