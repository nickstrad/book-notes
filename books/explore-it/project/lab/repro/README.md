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


E02 adds `probe02.py`, wrappers T01/T02/T03/U09/U10 and `lesson-02.ids`.
T01/T02/T03 are the machine names for plan T1/T2/T3; they count observation
gaps, separately from the two new U bugs. The E02 decoder preserves plain-text
historical 404 responses, so a missing clock/events endpoint is PRESENT for its
own visibility gap. The crash fixture abandons a public one-second lease and
uses identical real-time bounds on both versions. T01 inspects actual container
JSON logs; T02 checks retained history and namespace boundaries; T03 advances
semantic time and restores real time. U09 also runs the image's actual qctl
worker defaults with an eight-second container timeout; no current-source CLI
is substituted in historical runs. Run these instruments serially.

`python3 lab/observe02.py` replays an actual crash-worker exit, 500 notification,
slow-partner/sender deadline disagreement, recovery and container qctl
`job events`, `watch QUEUE`, `clock set|advance|reset`. It is an observation replay,
not a sixth counted bug. The direct legacy-row check uses a private database
schema; it is intentionally separate from the public U10 repro.

E03 adds `probe03.py`, wrappers U11–U18 and `lesson-03.ids`. U11/U12 send raw
bodies (trailing data, a second object, invalid UTF-8, case-variant names) so the
bytes on the wire are exactly the report's. U13 and U16 compare a stored time with
the moment of the request: a time in the past is the wrap. U15 leaves a terminal
job in a large-retention queue, waits six seconds for a zero-retention control to
be swept, then deletes the held job so a historical stack recovers. U13 acts on
the job the lease returned: its own wrapped enqueue is due first, so FIFO does not
give back the job just enqueued (the first REF attempt failed on exactly that).
U16 and U17 use a fresh queue holding one job. Run these instruments serially.

E04 adds `probe04.py`, wrappers U19–U23 and A1, `lesson-04.ids`, and
`invariants04.py`. A1 is the plan's spec-silence id: the black-box probe walks two
attempts at base 3000 s with no `max_seconds`, so a missing default cap shows as a
6000 s wait. U22 and U23 freeze and step the lab clock and always reset it. Every
probe leases with `lease_job`, which requires the lease to return the probe's own
job; U22 uses a fresh queue per round because the sweep requeues the last round's
job, and U21 enqueues its negative delay last because an accepted one is due at
once (the first REF attempt failed on exactly that). `invariants04.py I04 [seed]`
is the E04 rules list as a seeded random walk over 200 jobs (I1–I7, see its
docstring); `TestL04Invariants` runs it. It is a check, not a reproduction, so it
is not in `lesson-04.ids`.

E05 adds `probe05.py`, wrappers U24–U26 and A2, `lesson-05.ids`, and `walk05.py`.
U24 leases one job on `max_attempts: 2` and deletes the worker, up to four times;
an empty lease is the answer there, so it leases once per round instead of
waiting. U26 freezes the lab clock, lets two leases expire once, then acks one
job and cancels the other and reads each job's last event; it always resets the
clock. A2 leases, purges, then nacks. `walk05.py W05 [seed]` is the E05 sequence
charter as a seeded random walk of 100 verbs (W1–W7, see its docstring) on a
frozen clock; `TestL05Walk` runs it. Holder verbs are favoured while a lease is
held, because a uniform walk rarely acks. It is a check, not a reproduction, so
it is not in `lesson-05.ids`.

E06 adds `probe06.py`, U27–U33/A4 wrappers and `lesson-06.ids`; eight named
`TestL06*` checks run these same oracles. U29 executes a separate worker from the
target image on its unique queue and checks exact echo result and signed webhook,
plus sleep/fail numeric consumers. U30 executes the target API container's qctl:
input and display have independent API controls; ack/batch/schedule JSON siblings
are also checked. Numbers are observed with Decimal, not float parsing. U28
compares manual and cron jobs on an API clock and verifies PATCH rejects payload.
U31 requires unchanged valid leases after NUL rejection and exact clean retries.
U32 bounds public enqueue/schedule/batch races, checks accepted-row cardinality,
known statuses/codes, controls and zero orphans; historical PRESENT needs actual
errors on both reported paths, not merely a sampling inference. U33 creates real
attempts then removes their jobs three ways; SELECT-only observations distinguish
storage cleanup from API 404 and prove independent snapshots still deliver. A4
uses a frozen API clock below/above retention, checking dead plus succeeded and
cancelled controls in API and storage; it resets time in finally. Setup, tool,
parse, unexpected outcome or missing control always produces PROBE ERROR.

E07 adds `probe07.py`, U34/U35/A3 and `lesson-07.ids`. U34 samples up to 32
worker-delete trials with 12 concurrent lease requests each and requires actual
historical zombie leases **and** worker-disappearance 500s. U35 samples up to eight
trials per lease/nack/release/expiry path, requires all four historical hits, and
checks deletion counts, lease fields, attempts and attribution. Sequential purge
controls include pending deletion, successful ack, nack, worker release and expiry.
Both stop historical sampling once enough positive evidence exists; current runs
exhaust the bounds. Trial details go to stderr, including failure evidence. Bulk
SELECT-only snapshots keep the same per-job/attempt assertions without an HTTP
request for every field. No new runtime hook or fixture SQL write is used.

A3 starts the **built target image's** sample worker for both current and REF.
It observes an eight-second sleep, a local HTTP streaming partner's writes/socket
closure, completion logs and subsequent quiet echo completion. Docker stop is
cleanup after evidence, never proof of cooperation. Terminal state or process
survival alone does not satisfy the probe. Unit tests also exercise context
cancellation during sleep and HTTP body reads.

`TestL07StateMatrix` has 49 subtests (34 rejected, 6 empty leases, 9 accepted
transitions); `TestL07ExtraActors` adds 13 explicitly separate holder/expiry/
cancel-request cells. Wrong-token variants include empty, arbitrary text, backslash
and NUL. Direct checks additionally cover transactional missing-worker identity and
unrelated-queue progress while a purge waits. Checks may use internal database
locks; the public reproductions do not. D18–D20 stay unchanged passing controls in
both E07 verify directions and are excluded from its three-item verdict count.

E08 adds `probe08.py`, U36/U37/U38/A5 and `lesson-08.ids`. Each probe changes a
neighbour of the whole lab, so they run serially and restore it in `finally`.
U36 stops PostgreSQL, restarts only the API, starts PostgreSQL and requires `/health`
200 within 30 s with no operator start; on the current tree it then kills the API
and PostgreSQL processes from the host (an unexpected exit, unlike `docker kill`,
which Docker treats as an operator stop) and requires recovery again. U38 freezes
PostgreSQL with `docker pause` for 12 s on a timer and requires `/health` 503 within
3 s and an authenticated request 503 within 8 s, then healthy service after thaw.
A5 walks the API test clock across 2027-03-14 and 2027-11-07 in America/New_York
and requires 02:00/02:30 once on each civil day and 01:30 once in the repeated
hour; the 09:00 and hourly schedules are controls that must hold at every tag.
U37 starts a second hooksink container in `timeout` mode on the lab network, sends
22 globex notifications to it, then requires one acme notification to the healthy
hooksink within 20 s; the slow partner must have been attempted. Cleanup switches
that partner to `ok` and waits for this probe's rows to deliver before removing it,
because U03's control waits for the whole outbox to drain.
D21–D24 (fixed in E01) run as unchanged early-fixed controls in both directions
and are excluded from the four-item verdict count.
