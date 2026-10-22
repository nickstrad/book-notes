# 06 — Relate it: owners, dependents and the data's journey

An entity diagram suggests two useful questions: what happens when an owner disappears,
and does a value survive every path that carries it? Here, queue deletion already refused
while jobs or schedules existed. Following payloads beyond the API found different answers:
the API preserved numbers, but queuelet's CLI and sample worker rounded them. Following a
job through deletion found attempt rows that no API could reach afterwards.

The sources are Hendrickson, *Explore It!* (2013), Chapter 7.1 “Recognizing Entities,
Attributes, and Dependencies”, 7.2 “CRUD”, and 7.3 “Following the Data”. These are
[approved-plan citations](../PLAN.md#22-chapter-map-kept-sections-only), checked against
the chapter map, not a claim of full-text validation by this fixer.

## What the blind experiment learned

The [original record](../exploration/experiments/06-relate-it/revealed.md) belongs to
exploration commit `b3b0559`. It is preserved byte-for-byte, including its reports, tools,
models and score. The fresh Claude explorer used about 20 minutes of its 60-minute budget:
roughly 3 reading, 10 probing and 7 writing. This fresh Codex Sol fix follows the coordinator's
[assignment](../register/reconciliations/06.md); it is not another exploration run.

| Original measure | Actual score |
|---|---:|
| Hypotheses stated | 20 |
| Confirmed | 18 |
| Refuted | 2 |
| Code-reading predictions contradicted, as recorded | 2 |
| Reports | 7 |
| Observed model cells | 30 |

The two contradictions need their original qualifications. H-E06-C predicted owner
orphans from the original schema; reading migration 002 had already counter-predicted the
queue guard before the probe confirmed it. H-E06-17 predicted a batch race error that did
not occur in 320 tries. That sampling non-hit does not prove the window absent. The
explorer explicitly reported no new contradiction of its own deterministic code predictions.

| Classification | Actual outcome |
|---|---|
| Expected here, found early | D15 queue deletion orphans dependents, D16 schedule edits keep an old next fire, D17 webhook truncation: found and fixed in E01. D08 numeric loss at the API and D27 disabled firing were also repaired early. They are not rediscoveries here |
| Found in E06 | Seven unplanned reports, reconciled to U27–U33 below; all assigned repairs implemented; deferred sibling questions below |
| Missed / disclosed planted defects | None remain to disclose |
| Spec silence, resolved here | A4 dead-job retention, a coordinator policy decision. The explorer observed succeeded/cancelled retention, did not probe dead retention and filed no A4 report |

The plan's teaching prediction changed. E06 did not newly expose D15–D17 or show different
queue/schedule orphan behavior. The repaired queue guards and snapshot rules held. The
useful distinction became **which dependencies should share a lifetime**, and **which
client decoder lies between two otherwise exact storage boundaries**.

## The seven faults, at `explore-it/lesson-05`

Paths below are relative to `project/` in that tag, not line numbers in the repaired tree.

| ID | Original report | Previous-tag source | Observable failure |
|---|---|---|---|
| U27 | [E06-F01](../exploration/defects/E06-F01-schedule-delete-missing-204.md) | `internal/api/schedules.go:136` | DELETE missing/repeated/foreign schedule says 204 even when no row changed |
| U28 | [E06-F02](../exploration/defects/E06-F02-schedule-payload-not-validated.md) | `internal/api/schedules.go:55`, `internal/schedule/fire.go:76`; compare `internal/api/enqueue.go:64` | template accepts null/non-objects, cron copies them into jobs, manual run rejects them; omitted payload becomes null |
| U29 | [E06-F03](../exploration/defects/E06-F03-sample-worker-echo-loses-precision.md) | `internal/client/client.go:57`, `internal/worker/worker.go:79` | worker decodes numbers as float64; echo result and webhook contain the rounded value |
| U30 | [E06-F04](../exploration/defects/E06-F04-qctl-rounds-payload-numbers.md) | `cmd/qctl/main.go:52`, `cmd/qctl/main.go:277` | enqueue input is rounded before storage; job get rounds exact stored payload and result for display |
| U31 | [E06-F05](../exploration/defects/E06-F05-nul-in-result-or-error-409-conflict.md) | `internal/api/jobs.go:65`, `internal/api/jobs.go:88` | NUL in result/error causes PostgreSQL rejection reported as 409 state_conflict with driver text, while lease remains valid |
| U32 | [E06-F06](../exploration/defects/E06-F06-enqueue-racing-queue-delete-400-fk-text.md) | `internal/api/enqueue.go:88`, `internal/api/schedules.go:57` | owner disappears between lookup and insert: 400 invalid_job or 500 database_error with constraint text instead of 404 |
| U33 | [E06-F07](../exploration/defects/E06-F07-attempts-outlive-their-job.md) | `migrations/001.sql:7`, `internal/api/jobs.go:32`, `internal/api/queues.go:170`, `internal/jobs/expire.go:91` | explicit DELETE, purge and retention remove jobs but leave attempts, including error text |

A4 lived in `internal/jobs/expire.go:91`: the sweep included dead alongside succeeded and
cancelled. That was consistent with the then-silent policy, not an eighth blind defect.

## CRUD before and after

“Before” is the explorer's observed lesson-05 grid in
[entities.md](../exploration/models/entities.md#crud-grid-entity--operation-observed-in-e06-unless-marked).
“After” describes this repair's direct checks and the unchanged observations explicitly
labelled below. Worker read/update cells without an API remain unobserved. The pre-E06 model was
also narrower: its original entity rows still said queue→job/schedule had no FK,
and attempt deletion had no observed outcome. E06 adds the repaired FK/guard table,
observes each job-removal path and separates the outbox/event lifetime. The model
file retains those dated revisions rather than overwriting what was believed earlier.

| Entity | Create before → after | Read before → after | Update before → after | Delete before → after |
|---|---|---|---|---|
| Queue | race loser leaks FK → 404 queue_not_found | namespace ownership holds → unchanged | max-attempt and notification snapshots hold → unchanged | dependent guard holds → race controls still refuse 409, empty delete 204 |
| Job | qctl rounds → exact payload; race 400 → 404 | qctl rounds payload/result → exact display | NUL ack/nack 409 → service-owned 400, valid lease untouched and clean retry succeeds | terminal delete removes job only → job and attempts leave; events and pending notifications remain |
| Schedule | non-object/null accepted, omission null → non-object/null rejected, omission {}; race 500 → 404 | missing GET 404 → still 404 | cron/zone/enabled rules already repaired → unchanged; PATCH still refuses payload | missing/foreign/repeated 204 → 404, real delete 204 |
| Worker | registers → isolated worker still registers and processes | no new read observation | no new update observation | release counted; process may keep polling deleted id → unchanged, recovery question deferred |
| Attempt | completion records it → one real row before each deletion test | orphan visible only in storage → live job's attempt remains reachable | no update API | never removed → removed on job DELETE, purge and retention |
| Outbox | terminal snapshot → exact worker echo result snapshot | signed partner bytes preserve API values → worker echo values now exact too | delivery retries → pending snapshots still deliver after job deletion | independent lifetime → unchanged; delivered-row cleanup remains open |
| Job event | transitions recorded → unchanged | survives job removal → checked through API and real storage | no mutation API | independent lifetime → unchanged |

Automatic job deletion adds a policy distinction: dead persists; succeeded/cancelled age
out. It is tested with a frozen API clock below cutoff and beyond two sweep opportunities,
with API status and SELECT counts agreeing. Retained dead work can still be retried.

## Follow payload, result and error

| Datum / hop | Before | After / evidence |
|---|---|---|
| Payload: HTTP enqueue → jsonb → API/lease | exact values; jsonb normalizes form | exact API control still holds; normalization of key order, whitespace and `1e2` is not value loss |
| Payload: qctl --payload → API → jsonb | float64 rounding | direct write/read check preserves `9007199254740993`, a long fractional decimal and a nested 64-bit integer |
| Payload: schedule create → stored template → manual/cron job | incompatible validation; cron admitted poison payloads | same object/default contract; both paths produce equivalent valid values; PATCH gains no field |
| Payload → worker decode → echo result | numbers rounded inside client | target container's worker produces exact result; sleep and fail numeric consumers still work |
| Result: ack → jsonb → outbox → partner | API exact; worker already rounded it upstream | exact worker result reaches partner with valid signature; API and webhook values agree |
| Result: qctl --result/--json and job get | shared float decoder | ack, batch and schedule JSON sibling paths checked directly; exact API-written payload/result displayed exactly |
| Error/result: NUL → storage | PG rejects it; service calls it state_conflict | reject before mutation, including nested strings and object keys; literal backslash-u text stays valid |
| Error: nack → attempt/history → job removal | attempt error text outlives inaccessible job | attempts share job lifetime; events and pending terminal snapshots keep independent lifetime |

## The focused patch and decisions

- `objectPayload` in `internal/api/enqueue.go` is the existing enqueue rule factored out for
  schedule creation. Omission becomes `{}`; explicit null and other JSON types are rejected.
  No payload PATCH field is added. Manual/cron execution of newly created templates consumes
  validated payloads; legacy invalid templates are not migrated or revalidated.
- `deleteSchedule` checks RowsAffected and returns service-owned schedule_not_found for zero.
- Client JSON decoding and qctl option/output decoding use `UseNumber`. The worker's existing
  sleep/failure numeric consumers accept json.Number, retaining their omitted-value defaults.
  The baseline lease check parses its returned ID as exact int64 and retains its priority and
  concurrent-disjoint-lease assertions.
- Ack recursively checks result strings and object keys for NUL; nack checks error text.
  Both return 400 invalid_result/invalid_error before touching lease, attempts or result.
- `missingQueue` recognizes only the assigned owner foreign-key violations, returning 404
  queue_not_found without schema text. Enqueue, batch, manual schedule run and schedule create
  use that mapping. No speculative lease-race repair is included.
- `migrations/005.sql` adds `attempts_job_owner ON DELETE CASCADE` and a job-id index. It is
  NOT VALID so existing legacy orphans remain inspectable; it enforces new writes and removes
  attempts when their existing owner subsequently leaves. This fix does not bulk-delete old
  orphan evidence or alter independent outbox/events lifetimes.
- Retention selects succeeded/cancelled, exempting dead.

The decided [SPEC lines](../project/SPEC.md) are: **dead jobs remain until an operator
retries or explicitly deletes them; automatic retention applies only to succeeded/cancelled**
(A4); **attempt records belong to their job and leave with it on deletion, purge and retention**
(U33 / Q-E06-2 attempts half). Schedule payloads use the existing enqueue contract; NUL result/error
is a 400 with a valid lease left usable. Delivered-outbox cleanup stays undecided.

## Verify

From `books/explore-it/project`, sequentially:

```sh
make up
make verify LESSON=06
make verify LESSON=06 REF=explore-it/lesson-05
make checks
make test
```

`lesson-06.ids` names U27–U33 and A4. Their wrappers use
[`probe06.py`](../project/lab/repro/probe06.py); `TestL06U27` through `TestL06A4` execute
these same public oracles. Historical qctl executes inside the historical API container;
the sample-worker probe starts an isolated container from that target's image. Current
client source is never used to decode historical input. Public repros perform no SQL writes.
Unexpected setup, transport, parse, status, control or storage observations are PROBE ERROR.

Measured 2026-10-22; exact commands, exits, elapsed times and output are in
[`06-acceptance.json`](../register/runs/06-acceptance.json) and its linked logs.

| Command | Exit | Elapsed | Exact result |
|---|---:|---:|---|
| `make up` (guarded migration) | 0 | 41.282 s | current stack ready |
| `make verify LESSON=06` | 0 | 151.177 s | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| `make verify LESSON=06 REF=explore-it/lesson-05` | 0 | 57.671 s | `DEFECTS REPRODUCED: 8 present` |
| `make checks` | 0 | 398.913 s | all 74 integration checks pass (66 earlier, 8 E06) |
| `make test` | 0 | 1.239 s | unit suite passes |

| Surface | lesson-05, measured | Repaired tree, measured |
|---|---|---|
| U27 schedule DELETE foreign/repeated/missing | 204/204/204 | 404/404/404; owner delete 204, foreign resource intact |
| U28 template create | all five invalid shapes 201; omission null; manual run 400; cron fires 7 templates | invalid shapes 400; omission {}; manual run 201; cron fires 2 valid templates |
| U29 worker echo + webhook | big integer …992, fraction 0.12345678901234568 | …993, fraction 0.12345678901234567890123456789; raw webhook equals stored result, valid signature |
| U30 CLI input/display and ack/batch/schedule siblings | all six exact-value comparisons false | all six true; independent API controls exact |
| U31 NUL string/nested/key result and error | all four 409 | all four 400; lease unchanged and clean retry succeeds |
| U32 enqueue/schedule owner races | 2 actual 400 FK errors / 4 actual 500 FK errors; batch sibling also hit once | 800 enqueue + 800 schedule + 128 batch writes: only 201/404; accepted writes match stored rows, zero orphans |
| U33 attempts after purge/delete/retention | 1/1/1 orphan rows | 0/0/0; live attempt, events and pending delivery controls hold |
| A4 after cutoff | dead API 404, storage 0; succeeded/cancelled removed | dead API 200, storage 1; succeeded/cancelled API 404, storage 0; below-cutoff controls remain and retained dead retry works |

**Dwell is paired only.** [`06-dwell.tsv`](../register/runs/06-dwell.tsv) derives from
these two runs: all eight PRESENT at lesson-05, ABSENT at lesson-06. No full sweep
or `make dwell` run was performed; this does not date their introduction. The
batch race was observed in fixer regression evidence, not retroactively counted
as an explorer finding or changed hypothesis score.

**Instrument and record integrity.** The final pair uses the same frozen
[`06-instruments.sha256`](../register/runs/06-instruments.sha256), rechecked after
the pair. All 3,447 original explorer artifacts match their original manifest;
only canonical U repair statuses and their index change in exploration/. The sink's
parsed convenience field rounds numbers, so U29 decodes its preserved `raw_body`
for wire evidence; hooksink itself is unchanged.

**Retained failures and deviations.** The first rebuilt startup failed because the
new constraint lacked a restart guard (`06-up-final.log`, diagnostics retained).
The guard was added and reset/restarts now work. The first current verification
stopped with PROBE ERROR on the sink's parsed convenience field (`06-current.log`,
diagnostics retained), before any historical verdict. That measurement error was
corrected to raw bytes; the failed instrument manifest is kept separately, and the
complete final pair ran with the corrected immutable instrument. An initial successful
build was repeated after fixing omitted-number defaults in worker consumers. No
fault, control or verdict was weakened. The patch follows seven actual unplanned
findings instead of the already-repaired D15–D17 sketch; A4 is policy disclosure.
The first full suite then exposed a baseline check's float64 type assumption after the shared
client change (`06-checks.log`, diagnostics retained). That test now parses an exact int64
without changing the behavior asserted. The entire required sequence was rerun with a newly
frozen manifest; public repro code did not change between the two successful pairs.
Legacy preexisting attempt orphans remain inspectable, without bulk cleanup.

**Review repair.** Muse Wonder Full's [E06 NITS review](../register/reviews/E06/review.md)
of superseded fix `7a38788` found that single streaming `Decode` calls accepted trailing
garbage or a second value. The amended fix validates the complete document before decoding
with `UseNumber`. Invalid `--payload`, `--result` and `--json` now return a clear CLI error
before any request; incomplete response printing preserves the full raw response, and
the shared client returns an error. Actual-binary CLI/stub-server tests, client HTTP tests
and worker-default tests exercise these paths. Repros, numeric PRESENT/ABSENT oracles,
original artifacts and score are unchanged. Repair acceptance is recorded separately in
[`06-repair-acceptance.json`](../register/runs/06-repair-acceptance.json); the earlier pair
and failed attempts above remain evidence. Independent validation of this repair is pending.

The repair reran the same five commands sequentially, all exit 0: `make up` 50.593 s;
current verify 121.898 s (`FIXED AND CHECKED: 8 absent; regression checks pass`);
unchanged lesson-05 REF 49.876 s (`DEFECTS REPRODUCED: 8 present`); `make checks`
399.806 s (74/74); `make test` 5.668 s (including the new actual-code tests).
`make down` stopped only the repair lab. Final rechecks matched all 3,447 original
artifact hashes, 95 frozen instrument hashes and the unchanged review report.


## Try changing one variable

- Replace the integer with `9007199254740992` and compare with `9007199254740993`: the
  float64 boundary can hide the decoder defect for one value and expose it for its neighbor.
- Put NUL in a nested result array or an object key; compare with literal `\\u0000` text.
- Omit schedule payload, send `{}`, then send null. The first two mean an empty object;
  the last is invalid. Compare manual run and cron firing.
- Remove an attempted job by each of purge, explicit DELETE and retention. Check job,
  attempt, event and pending notification independently rather than assuming all share a lifetime.
- Change retention to two hours and advance the API clock to either side of that cutoff;
  keep dead, succeeded and cancelled controls together.

For further reading: Chapter 7.1–7.3, the
[probe-entities-and-boundaries skill](../skills/probe-entities-and-boundaries/SKILL.md),
the [original notes](../exploration/experiments/06-relate-it/notes.md) and
[reconciliation](../register/reconciliations/06.md). Forward questions remain the sample
worker's recovery after deregistration (Q-E06-1), missing job ids in decode logs (T-E06-1),
delivered-outbox lifetime (Q-E06-2 remainder), deterministic race instrumentation (T-E06-2),
the explorer's batch non-hit (H-E06-17; the fixer subsequently hit and repaired its shared
error mapping), deleted-queue lease race (H-E06-20), worker NUL result and
registration consequences (H-E06-21/22), and the previously handed-forward questions including A3.
Review finding 2's pre-existing worker http-response number decoder remains explicitly
deferred: the actual http-worker path was not probed or changed here, so the echo repair
does not establish its precision. Legacy invalid schedule templates are also unchanged;
creation is repaired, but cron/manual consistency for pre-upgrade templates needs an
explicit upgrade-policy decision at checkpoint B. Neither question adds a discovery or score.
