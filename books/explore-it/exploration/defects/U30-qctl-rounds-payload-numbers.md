# U30 — E06 finding 4

Original immutable report: [E06-F04](E06-F04-qctl-rounds-payload-numbers.md).

Status: **fixed in lesson 06**; canonical repair status, separate from the immutable original report.

Regression: `TestL06U30`; public repro: `project/lab/repro/U30.sh`.
Repair commit: the commit containing this record, tagged `explore-it/lesson-06`.

Review repair: finding 1 in the [E06 review](../../register/reviews/E06/review.md)
of superseded fix `7a38788` identified a new prefix-acceptance regression. CLI JSON
options now require a complete document and report a nonzero error before any request;
output keeps the prior full raw-response fallback when JSON is incomplete. Actual-binary
stub-server unit tests cover exact numbers, whitespace, malformed input, trailing garbage,
multiple values and zero requests on invalid options. The numeric public repro/oracle
and original seven-report score are unchanged; independent repair validation is pending.

