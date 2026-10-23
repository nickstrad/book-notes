# R4 — rehearsal of the landing page's corrected run recipe

End-review finding R4 ([REVIEW.md](../../../REVIEW.md)): the recipe switched to lesson-07 without
rebuilding, so its verify and `make checks` ran against HEAD's binaries. Commit `1d6a64b` fixed the
recipe; `run.py` ran both of its blocks as written, on 2026-10-23, from a fresh clone at `1d6a64b`
(`/tmp/ei-fixes/r4/repo`), with an isolated stack (`queuelet-r4`, image tag `r4`, ports
18290/18291/15523, explicit `QUEUELET_DATABASE_URL`) and the worktree at `/tmp/ei-tag`.
After every step the ledger records the source commit, the `queuelet:r4` image ID and the running
API container's image ID. `acceptance.json` is the ledger; `logs/` holds each step's output.

| Step | Source | Image | Result | Seconds |
|---|---|---|---|---:|
| final `make up`, `make seed` | `1d6a64b` | `41723dbe2096` | exit 0 | 21.3 |
| `verify LESSON=00` | `1d6a64b` | `41723dbe2096` | `BASELINE HOLDS` | 16.8 |
| `verify LESSON=07 REF=explore-it/lesson-06` | `1d6a64b` | (ref stack) | `DEFECTS REPRODUCED: 3 present` | 79.9 |
| `verify LESSON=07` | `1d6a64b` | `41723dbe2096` | `FIXED AND CHECKED: 3 absent` | 260.1 |
| `make checks` | `1d6a64b` | `41723dbe2096` | **exit 2**: TestL06U29 and TestL07A3 PROBE ERROR timeouts (see below) | 1065.7 |
| worktree at lesson-07, `make up`, `make seed` | `8787e43` | `72a19547e8dc` | exit 0 | 81.7 |
| `verify LESSON=07` at its own tag | `8787e43` | `72a19547e8dc` | `FIXED AND CHECKED: 3 absent` | 337.7 |
| back to final: `make up`, `make seed` | `1d6a64b` | `2763222a7b31` | exit 0 (rebuilt) | 55.5 |
| `make checks` | `1d6a64b` | `2763222a7b31` | exit 0, all 94 pass | 716.7 |
| `make down`, worktree removed | — | — | exit 0 | — |

The image ID changes at each switch, which shows the recipe now tests the binaries it names. The
rebuild of the unchanged final source produced a new image ID (`2763222a…` vs `41723dbe…`); the
build is not byte-reproducible, so identity is the source commit plus a fresh build, not a fixed ID.

**The first `make checks` failure is retained, not a product finding.** It ran while another
session's video render held the 4-core host at load 12–14. Both failures were PROBE ERRORs from probe
time limits, not verdicts: U29's `docker stop -t 2` exceeded its 10 s timeout, and A3's HTTP worker
container took 24 s to start against a 10 s wait. A3's interrupted probe left one worker container
(`queuelet-e07-a677f4498854`), removed by hand. The rebuilt final stack then passed the complete
suite with unchanged code and instruments at load ≈4.
