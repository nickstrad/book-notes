Completed the single E06 repair round. JSON numbers remain exact; malformed, trailing, or concatenated CLI JSON now fails before any request. The shared client also enforces complete-document parsing.

Commits, in order:

1. `755d21c` — amended E06 fix, exact subject retained; parent `b3b0559`. Supersedes `7a38788`; `explore-it/lesson-06` points here.
2. `7148dee` — separate project-knowledge documentation commit.

All required commands passed sequentially:

| Command | Result |
|---|---|
| `make up` | Pass |
| `make verify LESSON=06` | 8 absent; regression checks pass |
| `make verify LESSON=06 REF=explore-it/lesson-05` | 8 present |
| `make checks` | 74 checks pass |
| `make test` | Pass, including actual-code strictness/default tests |

Evidence: [06-repair-acceptance.json](/root/Software/booknotes-work/explore-it/books/explore-it/register/runs/06-repair-acceptance.json).

All 3,447 artifact hashes and 95 frozen instruments match. Original evidence, repros, oracles, proofs and discovery score remain unchanged. The owned lab is stopped.

The worker HTTP precision sibling and legacy-template upgrade policy are explicitly deferred. Independent validation remains pending; no checkpoint B review or E07 started.

Tracked files and index are clean. Only coordinator-owned files remain untracked: the repair brief, `reviews/E06/`, `06-fixer-*`, `06-review-pools.json`, and `06-repair-pools.json`.