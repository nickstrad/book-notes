# 03 — Vary it: the variables of a job, a queue and a request

A system that rejects the values you thought of can still accept the ones you did not. This
experiment listed the variables of enqueue, batch enqueue and queue configuration before trying any
value, then pushed each one with named heuristics through raw `curl`. It found eight new faults,
and none of them was in the place the plan expected.

The source is Hendrickson, *Explore It!* (2013), Chapter 4 (4.1–4.4, finding and varying variables),
Appendix A2.1 (the heuristics: Zero/One/Many, Goldilocks, Boundaries, Strings, Numbers and the rest)
and 10.1 "Exploring an API" (does it reject what it does not understand?). These citations follow the
approved [chapter map](../PLAN.md#22-chapter-map-kept-sections-only); the builder did not validate
them against the book text.

## What the blind experiment learned

[The original record](../exploration/experiments/03-vary-it/revealed.md) is unchanged from the
experiment commit. The explorer stated 23 hypotheses, confirmed 21 and refuted 2. Two predictions made
from reading the code were contradicted by the running system. It filed 8 reports and filled 28 model
cells, in about 21 minutes of a 60-minute budget.

The charter predicted, from reading the validators, that "every documented range is enforced and
unknown fields are rejected". The running system refuted it on the first probe: a body with trailing
garbage got a 201.

| Classification | Outcome |
|---|---|
| Expected here, found early | D05 priority range, D06 batch bounds, D07 name charset, D08 payload precision, D26 unknown fields: all found and fixed in E01; E03 re-probed them and they held |
| Found as new in E03 | U11–U18 (below) |
| Missed / disclosed | none: no planted defect is left for this experiment |
| Spec silence touched, not settled | A1 (a default backoff cap) stays with E04; U16 and U17 make the arithmetic safe and the explicit cap work, but leave a queue without a cap uncapped |

The E01 recon recorded about sixteen variable rows. This sweep listed dozens of variables and recorded
25 variation rows (19 new, 6 regressions) in
[`models/variables.md`](../exploration/models/variables.md). The difference is the lesson: the eight
faults sit in variables the recon never named. They are the **shape of the request** (framing, case),
**time arithmetic** (seconds that overflow a duration), **limits the database enforces for you**, and
**accepted input that changes nothing**.

## The eight faults, with pointers into `explore-it/lesson-02`

| ID | Report | Where it lived | What happened |
|---|---|---|---|
| U11 | [E03-F01](../exploration/defects/E03-F01-malformed-json-text-accepted.md) | `internal/api/api.go:53` decodes one value and stops | `{…} trailing-garbage` and `{…}{…}` got 201; the second object was silently dropped; invalid UTF-8 in `kind` was rewritten to U+FFFD |
| U12 | [E03-F02](../exploration/defects/E03-F02-case-variant-member-names-accepted.md) | same decoder; Go matches member names case-insensitively | `KIND`, `Priority`, `LEASE_TOKEN` counted as the documented fields; the last spelling won |
| U13 | [E03-F03](../exploration/defects/E03-F03-duration-overflow-delay-and-extension.md) | `enqueue.go:54`, `nack.go:24`, `jobs.go:84` multiply seconds into a `time.Duration` | `delay_seconds: 9223372037` wrapped to 1734 and ran at once; a long heartbeat extension ended its own lease |
| U14 | [E03-F04](../exploration/defects/E03-F04-storage-limit-errors-reported-as-conflict.md) | `queues.go:48` maps every insert error to 409; `:94` every update error to 500 | `lease_seconds: 2147483648` was a "conflict" on create and a server error on PATCH, with driver text |
| U15 | [E03-F05](../exploration/defects/E03-F05-large-retention-halts-global-retention.md) | `expire.go:66` subtracts `retention_hours` as an interval in one statement for all queues | one acme queue with `retention_hours: 1e8` and a cancelled job made the sweep fail every tick, for every namespace; `/health` stayed 503 |
| U16 | [E03-F06](../exploration/defects/E03-F06-backoff-overflow-immediate-retry.md) | `jobs.go:17-21` `Backoff` shifts and multiplies without a bound | at the default base the 32nd retry was scheduled for 1782, so it ran at once |
| U17 | [E03-F07](../exploration/defects/E03-F07-backoff-max-seconds-ignored.md) | `queues.go:20` stores `max_seconds`; `nack.go:14` never reads it | a cap of 1 s was echoed back and ignored |
| U18 | [E03-F08](../exploration/defects/E03-F08-patch-webhook-partial-clears-other-member.md) | `queues.go:90-91` replaces the whole `webhook` object | rotating only the secret turned notifications off; changing only the URL stopped signing |

## The patch

The fix commit's diff under `project/` is small; the lines that matter:

- **One strict decoder** (`internal/api/api.go`, `decode`). It reads the body, rejects invalid UTF-8,
  decodes exactly one object and requires end of input. Then `exactNames` walks the object against the
  target struct and rejects a member whose name matches a field only when case is ignored. `payload`
  and `result` are free-form and are not walked. Every command route goes through `decode`, so ack,
  nack, workers, schedules and the clock route are covered without separate patches.
- **One bound for seconds** (`jobs.MaxSeconds = 2147483647`, `internal/jobs/jobs.go`). Enqueue delay,
  nack delay and heartbeat extension above it are 400. Queue limits use the same bound, so nothing that
  passes validation can overflow a duration or an `int` column.
- **Saturating backoff** (`jobs.Backoff`). The wait is computed in floating point and clamped at the
  queue's `max_seconds` or at `MaxSeconds`. It can no longer go down as attempts go up.
  `TestBackoffSaturates` pins that down for 99 attempts.
- **Validation before the store** (`internal/api/queues.go`). Names are 1–128 characters; limits are
  in range on create and PATCH (`inRange`); `max_seconds` must be at least `base_seconds`. Only a
  unique violation is 409; any other store error is 500.
- **A sweep that cannot overflow** (`internal/jobs/expire.go`, `Retain`). It compares elapsed hours
  with the setting instead of subtracting an interval from now. `retention_hours` is capped at
  1 000 000, but the comparison also survives a legacy row stored before the cap.
- **Member-wise webhook PATCH** (`patchQueue`). `url` and `secret` are optional pointers; an omitted
  one keeps its value.

Spec lines added to `project/SPEC.md`, one decision each (from
[reconciliation 03](../register/reconciliations/03.md)): a command body is one UTF-8 JSON object with
exact member names; seconds fields and `max_attempts` accept at most 2 147 483 647; names are 1–128
characters and only a taken name is a conflict; retention is 0–1 000 000 hours; `backoff.max_seconds`
caps each wait and a computed wait never schedules a retry before the failure; PATCH merges `webhook`
member by member.

This patch tightens commands, not reads. Release It!'s tolerant-reader advice is about what you accept
from peers you consume; a service's own command API should refuse what it cannot honour. Both books are
right about their own side.

## Verify

From `books/explore-it/project`, sequentially:

```sh
make up
make verify LESSON=03
make verify LESSON=03 REF=explore-it/lesson-02
make checks
make test
```

`lesson-03.ids` lists U11–U18. Each wrapper runs `lab/repro/probe03.py`, which uses the public API only:
raw bodies for U11/U12, stored times compared with the request time for U13/U16, a zero-retention
control job for U15. `TestL03U11`…`TestL03U18` run the same oracles as regression checks.

Measured on 2026-10-22 (logs under `register/runs/03-*`):

| Command | Exit | Elapsed | Result |
|---|---:|---:|---|
| `make up` | 0 | 73 s | current stack built and healthy |
| `make verify LESSON=03` | 0 | 31 s | `FIXED AND CHECKED: 8 absent; regression checks pass` |
| `make verify LESSON=03 REF=explore-it/lesson-02` | 0 | 145 s | `DEFECTS REPRODUCED: 8 present` |
| `go test -tags checks ./test/checks -count=1 -v` (`make checks` plus `-v`) | 0 | 241 s | all 54 integration checks pass (46 earlier, 8 new) |
| `make test` | 0 | 1 s | unit suite passes, including `TestBackoffSaturates` |

The first REF attempt stopped with `PROBE ERROR: U13` after U11 and U12 had printed PRESENT. The probe
had enqueued a job with the wrapped delay, which the historical binary made due at once, then leased
"the next job" expecting the one it had just enqueued. FIFO handed back the wrapped job instead, and the
heartbeat on the other job's id got 409. That is the same FIFO interference the explorer logged twice.
The probe now acts on the job the lease returned. The failed log and its Compose diagnostics are kept
as `register/runs/03-historical-attempt1*`.

Paired results, the same oracle on both binaries:

| Fault | lesson-02 | lesson-03 |
|---|---|---|
| U11 | trailing text, second object, invalid UTF-8: 201, 201, 201 | 400, 400, 400 |
| U12 | `KIND`/`PAYLOAD` 201; `priority:1,PRIORITY:9` stored priority 9 | 400, 400 |
| U13 | enqueue, heartbeat, nack at 9 223 372 037 s: all landed in the past | all 400 |
| U14 | create 2^31 → 409, PATCH 2^31 → 500, 2 800-character name → 409 | 400, 400, 400; duplicate name still 409 |
| U15 | `retention_hours: 1e8` → 201; zero-retention control still there after 6 s; health 503 | 1e8 → 400; a 1 000 000-hour queue with a terminal job does not stop the sweep; health 200 |
| U16 | attempt 4 at base 2 147 483 647 s → run_at 1986 | run_at 2094 (saturated at 2^31 − 1 s) |
| U17 | `max_seconds: -7` → 201; attempt 2 at base 4, cap 5 waited 8 s | -7 → 400; waited 5 s |
| U18 | secret-only PATCH → URL `''`; URL-only PATCH → secret `''` | URL kept; secret `rotated` kept |

The answer-key controls for A1–A4 still print PRESENT on the fixed tree (`03-unresolved-controls.log`):
A1 schedules attempt 20 30.34 days out at base 5. Those silences belong to E04–E07.

**Dwell.** Measured only as the pair above (`register/runs/03-dwell.tsv`): all eight present at
lesson-02, absent at lesson-03. No `make dwell` sweep over v0 and lesson-01 was run, so this note does
not claim when each fault first appeared. E01's fixes passed right by several of them: it added
`DisallowUnknownFields` to the same decoder (D26) and lower bounds to the same queue validator (U06),
and neither fix asked about case, framing or upper bounds.

## Try changing one variable

Send `{"kind":"echo","payload":{}}` followed by a newline and a space: whitespace after the object is
still one JSON text and is accepted. Try `"Payload"` inside `payload` (allowed: payload is yours) and at
the top level (rejected). Nack with `delay_seconds: 2147483647` and with one more. Create a queue with
`base_seconds: 4, max_seconds: 5` and walk three attempts: 4, 5, 5. PATCH `{"webhook":{"url":""}}` and
confirm the secret stays while notifications stop.

For further reading: Chapter 4 and A2.1, the [sweep-variables skill](../skills/sweep-variables/SKILL.md),
the [probe notes](../exploration/experiments/03-vary-it/notes.md) (including the dead ends: FIFO
interference twice, zsh word splitting, a float-printing instrument artefact), and the
[reconciliation](../register/reconciliations/03.md).
