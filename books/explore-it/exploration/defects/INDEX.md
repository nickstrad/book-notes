# Reconciled findings

Original F reports preserve what the blind explorer wrote. Canonical D/U records below carry current fix status and link to that evidence. Counts are not one-to-one: several reports cover more than one fault.

| ID | Original | Current record |
|---|---|---|
| D01 | [F22](F22-qctl-exit-status-zero-on-error.md) | [fixed in lesson 01](D01-qctl-exit-status-zero-on-error.md) |
| D02 | [F15](F15-queue-delete-orphans.md) | [fixed in lesson 01](D02-queue-delete-orphans.md) |
| D03 | [F05](F05-expired-lease-not-counted.md) | [fixed in lesson 01](D03-expired-lease-not-counted.md) |
| D04 | [F11](F11-webhook-non-2xx-marked-delivered.md) | [fixed in lesson 01](D04-webhook-non-2xx-marked-delivered.md) |
| D05 | [F19](F19-range-and-charset-validation.md) | [fixed in lesson 01](D05-range-and-charset-validation.md) |
| D06 | [F16](F16-batch-bounds-and-atomicity.md) | [fixed in lesson 01](D06-batch-bounds-and-atomicity.md) |
| D07 | [F19](F19-range-and-charset-validation.md) | [fixed in lesson 01](D07-range-and-charset-validation.md) |
| D08 | [F17](F17-payload-numeric-precision.md) | [fixed in lesson 01](D08-payload-numeric-precision.md) |
| D09 | [F04](F04-max-attempts-off-by-one.md) | [fixed in lesson 01](D09-max-attempts-off-by-one.md) |
| D10 | [F18](F18-stats-queued-includes-other-states.md) | [fixed in lesson 01](D10-stats-queued-includes-other-states.md) |
| D11 | [F23](F23-fifo-ties-within-a-second.md) | [fixed in lesson 01](D11-fifo-ties-within-a-second.md) |
| D12 | [F25](F25-purge-leaves-scheduled-and-retry.md) | [fixed in lesson 01](D12-purge-leaves-scheduled-and-retry.md) |
| D13 | [F24](F24-resume-catch-up-burst.md) | [fixed in lesson 01](D13-resume-catch-up-burst.md) |
| D14 | [F06](F06-retry-accepts-non-dead-jobs.md) | [fixed in lesson 01](D14-retry-accepts-non-dead-jobs.md) |
| D15 | [F15](F15-queue-delete-orphans.md) | [fixed in lesson 01](D15-queue-delete-orphans.md) |
| D16 | [F07](F07-schedule-patch-not-reflected.md) | [fixed in lesson 01](D16-schedule-patch-not-reflected.md) |
| D17 | [F12](F12-webhook-result-stringified-truncated.md) | [fixed in lesson 01](D17-webhook-result-stringified-truncated.md) |
| D18 | [F01](F01-ack-ignores-lease-token.md) | [fixed in lesson 01](D18-ack-ignores-lease-token.md) |
| D19 | [F02](F02-nack-ignores-token-and-state.md) | [fixed in lesson 01](D19-nack-ignores-token-and-state.md) |
| D20 | [F05](F05-expired-lease-not-counted.md) | [fixed in lesson 01](D20-expired-lease-not-counted.md) |
| D21 | [F03](F03-job-endpoints-cross-namespace.md) | [fixed in lesson 01](D21-job-endpoints-cross-namespace.md) |
| D22 | [F08](F08-background-loop-stops-on-error.md) | [fixed in lesson 01](D22-background-loop-stops-on-error.md) |
| D23 | [F27](F27-schedule-zone-offset-frozen.md) | [fixed in lesson 01](D23-schedule-zone-offset-frozen.md) |
| D24 | [F10](F10-webhook-signature-wrong-bytes.md) | [fixed in lesson 01](D24-webhook-signature-wrong-bytes.md) |
| D25 | [F21](F21-lease-unknown-worker-200.md) | [fixed in lesson 01](D25-lease-unknown-worker-200.md) |
| D26 | [F20](F20-unknown-fields-and-enums-accepted.md) | [fixed in lesson 01](D26-unknown-fields-and-enums-accepted.md) |
| D27 | [F07](F07-schedule-patch-not-reflected.md) | [fixed in lesson 01](D27-schedule-patch-not-reflected.md) |
| U01 | [F09](F09-schedule-patch-accepts-invalid-cron.md) | [fixed in lesson 01](U01-schedule-patch-validation.md) |
| U02 | [F12](F12-webhook-result-stringified-truncated.md) | [fixed in lesson 01](U02-webhook-result-type.md) |
| U03 | [F13](F13-slow-webhook-stalls-background-loop.md) | [fixed in lesson 01](U03-webhook-blocks-core-progress.md) |
| U04 | [F14](F14-db-outage-reported-as-invalid-key.md) | [fixed in lesson 01](U04-database-fault-auth-health.md) |
| U05 | [F16](F16-batch-bounds-and-atomicity.md) | [fixed in lesson 01](U05-batch-partial-insertion.md) |
| U06 | [F19](F19-range-and-charset-validation.md) | [fixed in lesson 01](U06-queue-patch-limits.md) |
| U07 | [F20](F20-unknown-fields-and-enums-accepted.md) | [fixed in lesson 01](U07-backoff-kind-validation.md) |
| U08 | [F26](F26-stats-age-from-run-at.md) | [fixed in lesson 01](U08-queued-age-reference.md) |

## Experiment 02 additions

- [U09](U09-sample-worker-kind-registration.md): documented fail3/default-worker kind mismatch; fixed in lesson 02.
- [U10](U10-malformed-webhook-blocks-independent-delivery.md): malformed destination blocks later notifications; fixed in lesson 02.
- E02-F03: delivery-timeout semantics clarified in SPEC.md; retained snapshots and at-least-once duplicates; no extra bug counted.

Original E02-F reports and raw scores are immutable. Canonical statuses are maintained separately.

## Experiment 03 additions

All eight E03 reports are new unplanned faults, fixed in lesson 03. The planned E03 defects
(D05–D08, D26) were already fixed in lesson 01 and held as regressions.

| ID | Original | Current record |
|---|---|---|
| U11 | [E03-F01](E03-F01-malformed-json-text-accepted.md) | [fixed in lesson 03](U11-malformed-json-text-accepted.md) |
| U12 | [E03-F02](E03-F02-case-variant-member-names-accepted.md) | [fixed in lesson 03](U12-case-variant-member-names-accepted.md) |
| U13 | [E03-F03](E03-F03-duration-overflow-delay-and-extension.md) | [fixed in lesson 03](U13-duration-overflow-delay-and-extension.md) |
| U14 | [E03-F04](E03-F04-storage-limit-errors-reported-as-conflict.md) | [fixed in lesson 03](U14-storage-limit-errors-reported-as-conflict.md) |
| U15 | [E03-F05](E03-F05-large-retention-halts-global-retention.md) | [fixed in lesson 03](U15-large-retention-halts-global-retention.md) |
| U16 | [E03-F06](E03-F06-backoff-overflow-immediate-retry.md) | [fixed in lesson 03](U16-backoff-overflow-immediate-retry.md) |
| U17 | [E03-F07](E03-F07-backoff-max-seconds-ignored.md) | [fixed in lesson 03](U17-backoff-max-seconds-ignored.md) |
| U18 | [E03-F08](E03-F08-patch-webhook-partial-clears-other-member.md) | [fixed in lesson 03](U18-patch-webhook-partial-clears-other-member.md) |

Original E03-F reports and raw scores are immutable.

## Experiment 04 additions

All five E04 reports are new unplanned faults, fixed in lesson 04. The planned E04 defects
(D09–D11, D25) were already fixed in lesson 01; D09–D11 held as regressions here. The spec silence A1
(no default backoff cap) is resolved in lesson 04 by decision.

| ID | Original | Current record |
|---|---|---|
| U19 | [E04-F01](E04-F01-unmatched-route-errors-not-json.md) | [fixed in lesson 04](U19-unmatched-route-errors-not-json.md) |
| U20 | [E04-F02](E04-F02-attempt-history-after-retry-ambiguous.md) | [fixed in lesson 04](U20-attempt-history-after-retry.md) |
| U21 | [E04-F03](E04-F03-negative-delay-wraps-and-precedes-failure.md) | [fixed in lesson 04](U21-negative-delay-accepted.md) |
| U22 | [E04-F04](E04-F04-expired-lease-honoured-until-sweep.md) | [fixed in lesson 04](U22-expired-lease-honoured-until-sweep.md) |
| U23 | [E04-F05](E04-F05-queued-age-counts-leased-time.md) | [fixed in lesson 04](U23-queued-age-counts-leased-time.md) |

Original E04-F reports and raw scores are immutable.

## Experiment 05 additions

All three E05 reports are new unplanned faults, fixed in lesson 05. The planned E05 defects
(D12–D14) were already fixed in lesson 01; D12 and D14 held as regressions here and D13 was
re-probed as a pause, one-hour clock jump and resume. The spec silence A2 (a job leased when its
queue is purged) is resolved in lesson 05 by decision.

| ID | Original | Current record |
|---|---|---|
| U24 | [E05-F01](E05-F01-worker-delete-release-uncounted-attempt.md) | [fixed in lesson 05](U24-worker-delete-release-uncounted-attempt.md) |
| U25 | [E05-F02](E05-F02-terminal-rows-keep-lease-fields.md) | [fixed in lesson 05](U25-terminal-rows-keep-lease-fields.md) |
| U26 | [E05-F03](E05-F03-worker-release-logged-as-lease-expired.md) | [fixed in lesson 05](U26-release-logged-as-lease-expired.md) |

Original E05-F reports and raw scores are immutable.

## Experiment 06 additions

Seven original reports remain immutable; all seven canonical U records are fixed in
lesson 06. D15–D17 were found and repaired in E01, not rediscovered here. A4 is the
coordinator's dead-retention policy decision, not an eighth blind finding. Attempts
follow job deletion/purge/retention; events and pending snapshots stay independent.

| ID | Original | Current record |
|---|---|---|
| U27 | [E06-F01](E06-F01-schedule-delete-missing-204.md) | [fixed in lesson 06](U27-schedule-delete-missing-204.md) |
| U28 | [E06-F02](E06-F02-schedule-payload-not-validated.md) | [fixed in lesson 06](U28-schedule-payload-not-validated.md) |
| U29 | [E06-F03](E06-F03-sample-worker-echo-loses-precision.md) | [fixed in lesson 06](U29-sample-worker-echo-loses-precision.md) |
| U30 | [E06-F04](E06-F04-qctl-rounds-payload-numbers.md) | [fixed in lesson 06](U30-qctl-rounds-payload-numbers.md) |
| U31 | [E06-F05](E06-F05-nul-in-result-or-error-409-conflict.md) | [fixed in lesson 06](U31-nul-in-result-or-error-409-conflict.md) |
| U32 | [E06-F06](E06-F06-enqueue-racing-queue-delete-400-fk-text.md) | [fixed in lesson 06](U32-enqueue-racing-queue-delete-400-fk-text.md) |
| U33 | [E06-F07](E06-F07-attempts-outlive-their-job.md) | [fixed in lesson 06](U33-attempts-outlive-their-job.md) |

Original score: 20 stated / 18 confirmed / 2 refuted / 2 qualified reading contradictions / 30 cells.
