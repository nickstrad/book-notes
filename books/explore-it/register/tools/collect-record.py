from pathlib import Path
import shutil,json,hashlib,sys,subprocess
m=json.loads(Path('/tmp/explore-it-coordinator/current.json').read_text())
root=Path(m['root']); repo=Path(subprocess.check_output(['git','rev-parse','--show-toplevel'],text=True).strip())/'books/explore-it'
exp=Path(m['experiment'])
for required in ['charter.md','brief.md','notes.md','revealed.md']:
    if not (exp/required).is_file():raise SystemExit(f'missing {required}')
# Reject edits to inherited raw records rather than silently replacing history.
import tarfile,io
snapshot=subprocess.check_output(['git','archive',m.get('source_commit',m['tag']),'books/explore-it/exploration','books/explore-it/project'])
violations=[]
with tarfile.open(fileobj=io.BytesIO(snapshot)) as archive:
    for member in archive.getmembers():
        if not member.isfile():continue
        rel=Path(member.name).relative_to('books/explore-it')
        if rel.parts[:2]==('exploration','defects') and rel.name in ('README.md','INDEX.md'):
            path=root/rel
            if not path.is_file() or not path.read_bytes().startswith(archive.extractfile(member).read()):violations.append(str(rel)+' (index must retain inherited prefix)')
            continue
        immutable=(rel.parts[0]=='project' or rel.parts[:2] in [('exploration','experiments'),('exploration','defects')])
        if not immutable:continue
        path=root/rel
        if not path.is_file() or path.read_bytes()!=archive.extractfile(member).read():violations.append(str(rel))
if violations:raise SystemExit('Inherited artifacts changed; preserve originals and reconcile explicitly: '+', '.join(violations))
if (root/'STATE.md').is_file():
    (exp/'evidence').mkdir(exist_ok=True)
    shutil.copyfile(root/'STATE.md',exp/'evidence/export-state.md')
# Retain the explorer's text scripts and raw evidence used by recorded commands.
# Binary qctl builds are reproducible from the exported source and are omitted.
for p in (root/'project/results').rglob('*'):
    if not p.is_file() or p.stat().st_size > 2_000_000: continue
    try: p.read_text()
    except (UnicodeDecodeError, OSError): continue
    target=exp/'evidence/project-results'/p.relative_to(root/'project/results')
    target.parent.mkdir(parents=True,exist_ok=True)
    shutil.copyfile(p,target)
files=[]
for p in (root/'exploration').rglob('*'):
    if not p.is_file():continue
    rel=p.relative_to(root)
    target=repo/rel
    if target.is_file() and target.read_bytes()==p.read_bytes():continue
    target.parent.mkdir(parents=True,exist_ok=True)
    shutil.copyfile(p,target)
    files.append({'file':str(rel),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()})
manifest={'experiment':m['number'],'source_tag':m['tag'],'export_path':m['base'],'artifacts_copied_without_rewriting':files}
p=repo/'register/runs'/f'{m["number"]:02d}-artifact-manifest.json'
p.parent.mkdir(parents=True,exist_ok=True);p.write_text(json.dumps(manifest,indent=2)+'\n')
for name in ['run-start.txt','run-end.txt','exit-status.txt','run.jsonl','run.stderr','launch-failure-run.stderr','launch-failure-exit-status.txt']:
    source=Path(m['base'])/name
    if source.is_file():shutil.copyfile(source,p.parent/f'{m["number"]:02d}-explorer-{name}')
print(json.dumps(manifest,indent=2))
