# Checkpoint A nit repair validation

2026-10-22 UTC, existing `plan/explore-it` worktree. Reviewed SUT target: `b839fdf`.
Only diagnostic control flow and documentation changed; no SUT/oracle edits.

Repair checks, all exit 0 (initial `git diff --check` excluded then-untracked review evidence):

- `bash -n books/explore-it/project/lab/verify.sh`
- `git diff --check`
- `git diff --cached --check -- books/explore-it/STATE.md books/explore-it/register/BUILD-STATE.md books/explore-it/project/lab/verify.sh books/explore-it/register/reviews/A/a-nits-validation.md`
- `python3 /tmp/explore-it-a-nits/check.py`

The scratch Python helper copied the edited verifier into a temporary project and
substituted git/archive, Compose and Go commands with controlled stand-ins.
No real containers or behavioral oracles ran. Each case asserted exact exit status:

| Case | Expected exit | Assertions |
| --- | --- | --- |
| REF up failure | 23 | ps/logs/metadata saved before down; scratch removed |
| REF wait failure | 29 | same diagnostics and cleanup |
| REF up failure plus ps/logs failures 7 and down failure 9 | 23 | original failure retained; diagnostic outputs saved |
| Current PRESENT | 1 | `VERIFICATION FAILED: 1 present; 0 absent` emitted |
| Current ABSENT | 0 | regression command reached, successful verdict retained |
| Historical PRESENT | 0 | historical successful verdict retained; no failure directory |

All six cases passed; the injected nonzero exits were expected assertions, not
unreported test failures. Evidence directories were under scratch, not real SUT
results. Diagnostics on real REF failures will be under the invoking original
project's gitignored `results/ref-failure.*`.

Full behavioral suites were deliberately not repeated for this narrow change.
The independent review's initial historical startup failure remains recorded in
`logs/historical01.log`, followed by the successful `logs/historical01-retry.log`.
Lost container logs cannot be recovered; the original transient's cause is unproven.
Reviewed exact verdicts:

```text
FIXED AND CHECKED: 35 absent; regression checks pass
DEFECTS REPRODUCED: 35 present
```

Deferred: note 3 direct Go lease/token/tenant assertions to E07/E09; note 4 final
skill worked examples to E09; note 5 intentionally no canonical filename rename.

After staging the original review evidence, unrestricted `git diff --cached
--check` exited 2 for pre-existing trailing whitespace in raw Compose logs.
Those evidence bytes are deliberately preserved. The scoped check above passes
for all authored repair files; this is the only validation-check failure.
