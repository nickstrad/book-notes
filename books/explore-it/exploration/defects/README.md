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
