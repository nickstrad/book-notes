import subprocess,pathlib,json,hashlib
root=pathlib.Path('/root/Software/booknotes-work/explore-it'); book=root/'books/explore-it'
def git(*a):return subprocess.check_output(['git',*a],cwd=root,text=True)
rows=[]
for n,exp in [(2,'e390c31'),(3,'09dd0a9'),(4,'dffb1ec'),(5,'01137b6'),(6,'b3b0559')]:
 tag=f'explore-it/lesson-{n:02}'; commit=git('rev-parse',tag).strip()
 files=git('diff','--name-only',exp,tag).splitlines()
 rows.append({'experiment':n,'exploration_commit':exp,'lesson_tag':tag,'fix_commit':commit,'outside_book':[x for x in files if not x.startswith('books/explore-it/')],'project_diff_stat':git('diff','--shortstat',exp,tag,'--','books/explore-it/project').strip(),'fix_commit_body':git('show','-s','--format=%B',tag)})
print(json.dumps({'head_at_start':'8aa2895','scope_and_history':rows,'merge_commits':git('rev-list','--merges','explore-it/lesson-01..HEAD').splitlines()},indent=2))
