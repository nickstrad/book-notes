from pathlib import Path
import subprocess,tarfile,io,re,sys,json
repo=Path(subprocess.check_output(['git','rev-parse','--show-toplevel'],text=True).strip())
n=int(sys.argv[1]); tag=sys.argv[2]
slugs=['','map-it','make-it-visible','vary-it','judge-it','sequence-it','relate-it','state-it','surround-it','explain-it']
skills=['','map-codebase-from-behavior','make-invisible-visible','sweep-variables','derive-oracles','probe-sequences-and-states','probe-entities-and-boundaries','probe-entities-and-boundaries','probe-entities-and-boundaries','follow-the-surprise']
# E07 uses the state/sequence procedure.
skills[7]='probe-sequences-and-states'
base=Path(f'/tmp/explore-it-experiment-{n:02d}')
base.mkdir(exist_ok=False)
paths=[f'books/explore-it/{p}' for p in ('project','exploration','skills','lessons')]
data=subprocess.check_output(['git','archive',tag,*paths],cwd=repo)
with tarfile.open(fileobj=io.BytesIO(data)) as t:t.extractall(base,filter='data')
root=base/'books/explore-it'
plan=(repo/'books/explore-it/PLAN.md').read_text()
section=plan.split(f'#### E{n:02d} · ',1)[1].split('\n#### ',1)[0]
hyp=section.split('- **Hypothesis:** ',1)[1].split('\n- **Probe:**',1)[0]
probe=section.split('- **Probe:** ',1)[1].split('\n- **Expected finds:**',1)[0]
source=section.split('- **Source:** ',1)[1].split('\n- **Hypothesis:**',1)[0]
exp=root/f'exploration/experiments/{n:02d}-{slugs[n]}'
exp.mkdir(parents=True,exist_ok=True)
charter=f'# Experiment {n:02d}: {slugs[n]}\n\nSource: {source}\n\n{hyp}\n\nProbe family: {probe}\n\nBudget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.\n'
(exp/'charter.md').write_text(charter)
brief=f'''# Blind exploration brief — experiment {n:02d}

Your entire permitted working area is `{base}`. It is a plain source export of checkpoint `{tag}` with no Git metadata. Read and use any files inside this export; read no files outside it. Do not search for or consult any external answer key or history. In particular, do not access any other repository checkout or agent session. Use only inherited instruments and information in this export and the running queuelet lab. Do not create agents.

Permission profile: local export artifact writes and queuelet lab operations are authorized; no Git writes, SUT source writes, external messages, paid API keys, or changes to unrelated services. Acceptance: a timestamped real probe log, characterized reports, living model updates and scored revealed record. The coordinator has already checked environment orientation and subscription pools; use the environment documented inside the export.

Read `{exp}/charter.md`, `{root}/exploration/PROTOCOL.md`, and `{root}/skills/{skills[n]}/SKILL.md`. Follow the charter and skill to execute the experiment, without changing the SUT. You may inspect source and run the real stack; record predictions and refutation conditions BEFORE each probe. Create hypotheses/models before probes when the charter calls for them. The budget is 60 minutes, not a request to wait for the full hour.

Write timestamped commands and actual trimmed outputs to `{exp}/notes.md`. Keep unsuccessful probes and dead ends. Write `{exp}/revealed.md` with status, elapsed reading/probing/writing, observations, model delta, next hypotheses, and score table: stated/confirmed/refuted hypotheses; code-reading predictions contradicted by execution; defects found; model cells filled. Update models and hypotheses under exploration/. Write characterized defect reports under exploration/defects using local identifiers F01, F02, ...; the coordinator will reconcile identifiers later. Each needs minimal reproduction, named oracle, observed output, variables isolated, sibling search, severity, and status. Every surprise needs either a report or written reason it is not a bug. Do not fix source or read future lessons. If tools are inadequate, record a testability request and continue useful probes. Do not invent outputs or replace runtime probes with source inferences; distinguish these routes explicitly.

No commits or tags; the coordinator will copy and commit your artifacts exactly as produced. Preserve any helper scripts needed to replay a logged command under the experiment directory; raw run output may be in results/. Keep generated scratch/results in this export's project/results/ or /tmp paths WITHIN this permitted directory. Do not reset or restart any unrelated Docker project. Before your first probe, record the current UTC timestamp and your predictions in the log. At the end return a short result and file list.
'''
(exp/'brief.md').write_text(brief)
(base/'input.md').write_text(brief)
meta={'base':str(base),'root':str(root),'experiment':str(exp),'tag':tag,'number':n,'slug':slugs[n]}
Path('/tmp/explore-it-coordinator/current.json').write_text(json.dumps(meta))
print(json.dumps(meta))
