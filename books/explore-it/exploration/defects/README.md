# Defect reports

Explorer reports from experiment 01 (blind run, local ids F01–F27; the coordinator reconciles ids). Reports are produced by real exploration, never copied from the register. Follow [the protocol](../PROTOCOL.md) and [follow-the-surprise](../../skills/follow-the-surprise/SKILL.md).

| Report | Seen at | Oracle | Severity | Status | Experiment / evidence |
|---|---|---|---|---|---|
| [F01](F01-ack-ignores-lease-token.md) Ack completes a leased job without checking the lease token | v0-inherited | SPEC | High | open | E01 notes |
| [F02](F02-nack-ignores-token-and-state.md) Nack checks neither lease token nor job state: revives queued, succeeded and cancelled jobs | v0-inherited | SPEC | Critical | open | E01 notes |
| [F03](F03-job-endpoints-cross-namespace.md) Job-id endpoints ignore the caller's namespace | v0-inherited | SPEC | Critical | open | E01 notes |
| [F04](F04-max-attempts-off-by-one.md) A job with max_attempts N can be executed N+1 times through nack | v0-inherited | SPEC | High | open | E01 notes |
| [F05](F05-expired-lease-not-counted.md) Expired leases neither count nor record an attempt; jobs re-execute without bound | v0-inherited | SPEC | High | open | E01 notes |
| [F06](F06-retry-accepts-non-dead-jobs.md) Operator retry is accepted for succeeded (non-dead) jobs | v0-inherited | SPEC | Medium–high | open | E01 notes |
| [F07](F07-schedule-patch-not-reflected.md) Schedule PATCH does not recompute the next fire; disabled schedules keep firing | v0-inherited | SPEC | High | open | E01 notes |
| [F08](F08-background-loop-stops-on-error.md) Background loop stops permanently after any step error (DB restart, bad schedule) | v0-inherited | SPEC | Critical | open | E01 notes |
| [F09](F09-schedule-patch-accepts-invalid-cron.md) Schedule PATCH accepts an unparseable cron (and an unknown time zone) | v0-inherited | SPEC | High | open | E01 notes |
| [F10](F10-webhook-signature-wrong-bytes.md) Webhook signature is computed over different bytes than the body sent | v0-inherited | SPEC + HMAC-over-sent-bytes practice | High | open | E01 notes |
| [F11](F11-webhook-non-2xx-marked-delivered.md) Webhook responses with non-2xx status are marked delivered and never retried | v0-inherited | SPEC | High | open | E01 notes |
| [F12](F12-webhook-result-stringified-truncated.md) Webhook result is a JSON string and silently truncated at 1024 bytes | v0-inherited | SPEC | Medium | open | E01 notes |
| [F13](F13-slow-webhook-stalls-background-loop.md) A slow webhook partner stalls promotion, expiry and scheduling | v0-inherited | approximation: promotion ≈1 tick (SPEC silent on bound) | Medium | open | E01 notes |
| [F14](F14-db-outage-reported-as-invalid-key.md) During a database outage valid keys get 401 "invalid API key"; /health stays ok | v0-inherited | RFC 9110 401 vs 5xx; SPEC error classes | Medium | open | E01 notes |
| [F15](F15-queue-delete-orphans.md) DELETE queue removes a nonempty queue, orphans its jobs and schedules, and returns 204 for missing queues | v0-inherited | SPEC | Medium | open | E01 notes |
| [F16](F16-batch-bounds-and-atomicity.md) Batch enqueue: oversized truncated, empty accepted, partial insertion on error | v0-inherited | SPEC | Medium | open | E01 notes |
| [F17](F17-payload-numeric-precision.md) Payload numbers lose precision (float64 round-trip) | v0-inherited | SPEC | Medium | open | E01 notes |
| [F18](F18-stats-queued-includes-other-states.md) Stats "queued" count includes scheduled and retry_wait jobs | v0-inherited | SPEC | Low–medium | open | E01 notes |
| [F19](F19-range-and-charset-validation.md) Range/charset rules not enforced: queue names, priority, PATCHed limits | v0-inherited | SPEC | Low–medium | open | E01 notes |
| [F20](F20-unknown-fields-and-enums-accepted.md) Unknown fields and invalid enum values are accepted instead of rejected | v0-inherited | SPEC | Low–medium | open | E01 notes |
| [F21](F21-lease-unknown-worker-200.md) Lease with an unknown worker returns HTTP 200 with an error body | v0-inherited | SPEC | Low–medium | open | E01 notes |
| [F22](F22-qctl-exit-status-zero-on-error.md) qctl exits 0 on every error | v0-inherited | CLI exit-status convention | Low | open | E01 notes |
| [F23](F23-fifo-ties-within-a-second.md) FIFO within a priority is not by creation order for jobs created in the same second | v0-inherited | SPEC | Medium | open | E01 notes |
| [F24](F24-resume-catch-up-burst.md) Resuming a paused queue enqueues every missed schedule tick at once | v0-inherited | SPEC | Medium | open | E01 notes |
| [F25](F25-purge-leaves-scheduled-and-retry.md) Purge removes only `queued` jobs; scheduled and retry_wait remain (spec silence) | v0-inherited | SPEC silent ("pending" undefined) | Low–medium | spec | E01 notes |
| [F26](F26-stats-age-from-run-at.md) oldest_queued_age is measured from run_at, so a backdated run_at reports years of queueing | v0-inherited | approximation: age ≤ job lifetime | Low | open | E01 notes |
| [F27](F27-schedule-zone-offset-frozen.md) Schedules freeze the UTC offset at creation; DST changes are ignored | v0-inherited | SPEC | Medium | open | E01 notes |

Each report contains title, seen-at tag/build, original-observation link, minimal setup/steps, expected behavior with named oracle, observed behavior/evidence, variables isolated (including irrelevant ones), sibling check with file:line or “none found,” severity/impact and status.

Use a provisional local id during the blind run if needed. The coordinator later reconciles matching planted findings to `D##-<slug>.md` and unplanned findings to `U##-<slug>.md`; do not look up the answer key to choose an id. Status values are `open`, `fixed in lesson NN`, `won't fix: <why>` and `spec`. Keep an unresolved intermittent observation honest about attempts and uncertainty.

## Experiment 02 (make-it-visible) — explorer reports at `explore-it/lesson-01`

Local ids E02-F01… are provisional; the coordinator reconciles them. E01 entries above are unchanged.

| Report | Seen at | Oracle | Severity | Status | Experiment / evidence |
|---|---|---|---|---|---|
| [E02-F01](E02-F01-sample-worker-ignores-failN-kinds.md) Documented `fail3` kind is never leased by the sample worker run with the documented kinds | lesson-01 | README worker mode; SPEC:73 | Low | open | E02 notes H10 |
| [E02-F02](E02-F02-unparsable-webhook-url-blocks-all-notifications.md) One unparsable webhook URL stops notification delivery for every queue and namespace | lesson-01 | SPEC:121, SPEC:10 (approx.), SPEC:17 | High | open | E02 notes H11–H11d |
| [E02-F03](E02-F03-timeout-shorter-than-partner-unbounded-duplicates.md) A partner slower than 2 s receives the same notification forever; queuelet and partner disagree | lesson-01 | SPEC:120-121 (silent on timeout/cap) | Medium | spec | E02 notes H12 |

## Experiment 03 (vary-it) — explorer reports at `explore-it/lesson-02`

Local ids E03-F01… are provisional; the coordinator reconciles them. E01/E02 entries above are unchanged.

| Report | Seen at | Oracle | Severity | Status | Experiment / evidence |
|---|---|---|---|---|---|
| [E03-F01](E03-F01-malformed-json-text-accepted.md) Malformed JSON bodies accepted: trailing data ignored, second object dropped, invalid UTF-8 rewritten | lesson-02 | SPEC:17; RFC 8259 §2, §8.1 | Low–medium | open | E03 notes H01, H16, H19 |
| [E03-F02](E03-F02-case-variant-member-names-accepted.md) Case-variant member names accepted as documented fields; last variant wins | lesson-02 | SPEC:17; RFC 8259 §8.3 | Low–medium | open | E03 notes H02, H20 |
| [E03-F03](E03-F03-duration-overflow-delay-and-extension.md) delay_seconds / nack delay / heartbeat extension beyond ≈292 y wrap: "later" becomes "now" | lesson-02 | SPEC:48-49; worker protocol | Medium | open | E03 notes H04, H17, H18 |
| [E03-F04](E03-F04-storage-limit-errors-reported-as-conflict.md) Out-of-range settings and over-long names → 409 conflict / 500 with driver text | lesson-02 | SPEC:13-16; RFC 9110 409 | Low | open | E03 notes H06, H11 |
| [E03-F05](E03-F05-large-retention-halts-global-retention.md) One queue's large retention_hours halts retention for all namespaces; health 503 | lesson-02 | SPEC:98, :25, :127-130; SPEC:10 approx. | High | open | E03 notes H09 |
| [E03-F06](E03-F06-backoff-overflow-immediate-retry.md) Backoff overflow: attempt ≥32 (default base) retried immediately | lesson-02 | SPEC:27-28 | Medium | open | E03 notes H08 |
| [E03-F07](E03-F07-backoff-max-seconds-ignored.md) Undocumented backoff.max_seconds accepted, stored, never applied | lesson-02 | SPEC:17, :26-28 | Low–medium | open | E03 notes H07 |
| [E03-F08](E03-F08-patch-webhook-partial-clears-other-member.md) PATCH partial webhook clears the omitted member (notifications off / unsigned) | lesson-02 | approximation: RFC 7396 + SPEC:33-34 | Medium | open | E03 notes H10 |

## Experiment 06 (relate-it) — explorer reports at `explore-it/lesson-05`

Local ids E06-F01… are provisional; the coordinator reconciles them. Earlier entries above are unchanged.

| Report | Seen at | Oracle | Severity | Status | Experiment / evidence |
|---|---|---|---|---|---|
| [E06-F01](E06-F01-schedule-delete-missing-204.md) DELETE schedule answers 204 for a missing id and another tenant's schedule | lesson-05 | SPEC:15, :10 | Low | open | E06 notes P2, H-E06-03 |
| [E06-F02](E06-F02-schedule-payload-not-validated.md) Schedules accept non-object payloads; cron fires them into jobs; sample worker cannot decode them | lesson-05 | SPEC:60; run's own check | Medium | open | E06 notes P2, P2b, P3 |
| [E06-F03](E06-F03-sample-worker-echo-loses-precision.md) Sample worker echo rounds payload numbers | lesson-05 | README echo; SPEC:63 | Low–medium | open | E06 notes P4 |
| [E06-F04](E06-F04-qctl-rounds-payload-numbers.md) qctl rounds numbers on input (stored) and display | lesson-05 | README CLI = API; SPEC:63 | Medium | open | E06 notes P5 |
| [E06-F05](E06-F05-nul-in-result-or-error-409-conflict.md) U+0000 in ack result / nack error → 409 state_conflict, lease kept | lesson-05 | SPEC:16, :19-22 | Low–medium | open | E06 notes P6 |
| [E06-F06](E06-F06-enqueue-racing-queue-delete-400-fk-text.md) Writes racing queue delete → 400/500 with FK text, not 404 | lesson-05 | SPEC:15 | Low | open | E06 notes P9b, P9c |
| [E06-F07](E06-F07-attempts-outlive-their-job.md) Attempt rows outlive their job after purge/DELETE/retention | lesson-05 | approximation SPEC:122 + :80-81 | Low | open | E06 notes P10 |

## Experiment 07 (state-it) — explorer reports at `explore-it/lesson-06`

Local ids E07-F01… are provisional; the coordinator reconciles them. Earlier entries above are unchanged.

| Report | Seen at | Oracle | Severity | Status | Experiment / evidence |
|---|---|---|---|---|---|
| [E07-F01](E07-F01-worker-delete-racing-lease-leaves-zombie-lease.md) Worker DELETE racing that worker's lease leaves the job leased by the deleted worker; it can still heartbeat/ack; second DELETE 404; (other order) lease 500 `sql: no rows` | lesson-06 | SPEC:93-95, :107, :15 | Medium–high | open | E07 notes P2, P3 |
| [E07-F02](E07-F02-purge-misses-rows-changing-state-concurrently.md) Purge racing lease / nack / worker release / expiry sweep leaves jobs pending or unflagged-leased | lesson-06 | SPEC:54-56 | Medium–high | open | E07 notes P2, P4, P5 |

## Experiment 08 (surround-it) — explorer reports at `explore-it/lesson-07`

Local ids E08-F01… are provisional; the coordinator reconciles them. Earlier entries above are unchanged.
Rows added by the coordinator from the explorer's reports (the explorer did not edit this index).

| Report | Seen at | Oracle | Severity | Status | Experiment / evidence |
|---|---|---|---|---|---|
| [E08-F01](E08-F01-api-does-not-recover-from-start-without-database.md) API started while PostgreSQL is unreachable exits for good; stays down after the database returns | lesson-07 | SPEC:186-189; approximation 30 s | Medium | open | E08 notes P3, P3b |
| [E08-F02](E08-F02-slow-destination-starves-all-later-notifications.md) Twenty slow or unreachable webhook rows starve every later notification, for every tenant | lesson-07 | SPEC:179, :197-198, :10 | High | open | E08 notes P5, P5b, P5c |
| [E08-F03](E08-F03-dst-gap-skips-and-overlap-double-fires.md) Fixed-time schedule in the DST gap skips that day; in the repeated hour runs twice | lesson-07 | SPEC:162 (silent); classic cron | Medium | spec | E08 notes P6 |
| [E08-F04](E08-F04-frozen-database-hangs-health-and-api.md) Frozen database hangs every request, `/health` included; no 503 | lesson-07 | SPEC:188-189; code's 1 s ping deadline | Medium–high | open | E08 notes P8, P8b |

## Experiment 09 (explain-it) — explorer reports at `explore-it/lesson-08`

Local ids E09-F01… are provisional; the coordinator reconciles them. Earlier entries above are unchanged. E09-F04 and
E09-F05 were drafted and withdrawn during the run (lab-only clock class, already judged under H-E04-08); the ids are
unused.

| Report | Seen at | Oracle | Severity | Status | Experiment / evidence |
|---|---|---|---|---|---|
| [E09-F01](E09-F01-one-unadvanceable-schedule-stops-all-later-schedules.md) One schedule that cannot advance stops every later schedule, in every namespace | lesson-08 | SPEC:166, :189, :10 (approx.) | High | open | E09 notes H-E09-02/02b/03/06 |
| [E09-F02](E09-F02-never-matching-cron-accepted-year-one.md) Never-matching cron accepted; next_run_at 0001-01-01 | lesson-08 | SPEC:19 (approx.), :170-171 | Medium | open | E09 notes H-E09-01/04/05 |
| [E09-F03](E09-F03-cron-zone-prefix-overrides-timezone.md) `CRON_TZ=`/`TZ=` prefix accepted and overrides `timezone` | lesson-08 | SPEC:160-162 | Medium | open | E09 notes H-E09-07/07b |
| [E09-F06](E09-F06-schedule-zone-offset-frozen-in-record.md) Schedule record reports a frozen `zone_offset` after DST | lesson-08 | SPEC:162 (approx.) | Low | open | E09 notes H-E09-13 |
| [E09-F07](E09-F07-ack-cancel-error-after-commit-and-lost-notification.md) Ack/cancel ‖ job DELETE: committed transition answered 409; notification lost (also with 200) | lesson-08 | SPEC:177, :203, :16 | Medium–high | open | E09 notes H-E09-14/16 |
| [E09-F08](E09-F08-one-vanished-job-drops-other-dead-notifications.md) Expiry/worker-delete pass: one vanished job drops later jobs' notifications; worker DELETE 500 | lesson-08 | SPEC:177, :204-205, :15 | Medium–high | open | E09 notes H-E09-15/17 |
| [E09-F09](E09-F09-non-iana-local-timezone-accepted.md) Schedule timezone `Local` accepted | lesson-08 | SPEC:161 | Low | open | E09 notes H-E09-08 |
