# Seeder proof harness (answer key)

Run `make up && make seed` in ../../project, then `./run-all.sh` here.
All 27 wrappers dispatch to probe.py; each prints a measured PRESENT or ABSENT.
Exit codes: 0 PRESENT, 1 ABSENT, 2 probe error. An error is never absence.
A1.sh–A4.sh separately characterize the four unspecified behaviors.

Each probe uses a unique queue. The sweep preserves the fixture queues. D22
interrupts this Compose stack's database and restores/restarts its API afterwards.
Do not run exploration or checks concurrently with that probe. Reset after the
sweep to restore a clean explorer seed. No proof files enter the blind export.

Most probes use HTTP and CLI only. Because v0 intentionally has no test clock,
D13/D27 arrange due schedule times with external SQL, and D20 sets attempts to its
limit to isolate expiry's missing cap independently of D03's absent accounting.
A1 positions the completed-attempt counter at 19. These are seeder fixture
arrangements, not black-box lesson reproduction scripts; later lesson scripts must
use the public interfaces and their available time hooks. D11 uses only public
batch/lease/nack calls: identical creation times plus a real tuple-moving update
make PostgreSQL's unspecified tie ordering observable. It reports observed ids.

D23 inspects a future scheduled instant in the opposite DST season using Python's
IANA zoneinfo as the oracle. It avoids the nonexistent 02:00 spring-transition
instant: a correct IANA cron scheduler skips that date, whereas a fixed offset
produces the wrong local hour in the opposite season.

Evidence from the initial v0 run is in ../runs/00-proofs.log and
../runs/00-silences.log. The backoff example is measured, not rounded to the plan's
illustrative 60 days: at base=5 and attempt=20, base × 2^(attempt−1) is 30.34 days.
