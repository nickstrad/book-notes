# Disclosed-finding reproductions

Run `bash lab/repro/D18.sh` from the project, or use `make verify LESSON=01`.
The lesson's `.ids` file lists every assigned finding. E01 has 27 D cases and
8 U cases, all disclosed by that experiment/reconciliation. Each wrapper calls
`probe.py`; regression tests `TestL01<ID>` use these same behavior oracles.

These are integration instruments, requiring Python 3, Docker Compose, a running
queuelet stack and its seeded API keys. `QUEUELET_URL`, `QUEUELET_HOOK_URL` and
Compose's project/port variables select an isolated stack. Run serially: the
partner and database-outage cases temporarily affect that stack. Cleanup restores
partner mode and restarts the API after an outage.

Every completed case emits one `PRESENT: ID — detail` or `ABSENT: ID — detail`.
An unexpected status, invalid fixture or failed observation emits `PROBE ERROR`
and exits 2. Both present and absent are successful observations; the verifier
compares them with the required current/historical verdict. SQL is SELECT-only
and executed with PostgreSQL's read-only transaction default. Fixture changes
use public APIs and documented partner/Docker controls.

Historical runs build API, worker and qctl from REF, then copy current lab
instruments into the scratch tree. D01 executes that container's qctl. Do not
replace a historical command with current source or weaken a case for a missing
clock hook. D13 waits through two real minute ticks; D27 waits through one and
checks an enabled schedule control. The same waits run in the regression checks.
Consequently E01 verification takes several minutes, exceeding the original
plan's estimate for a two-finding recon. Preserve logs and actual timings.

The seeder's answer-key fixtures live outside this directory and are never
substitutes for these black-box reproductions. Earlier exploratory records stay
unchanged when repairs change the observed behavior.
