# Experiment 03 — vary-it: notes

Explorer: fresh headless agent (Claude Opus 5.5), blind export `/tmp/explore-it-experiment-03` of `explore-it/lesson-02`.
Stack: own Compose project `queuelet-e03` (image tag e03; API 18240, hook 18241, PG 15473). Helper:
[tools/e03env.sh](tools/e03env.sh) (adapted copy of E02's `e02env.sh`; inherited copy unchanged). Raw outputs:
`project/results/e03/`. All timestamps UTC. `api M PATH ...` = raw curl as acme-key, prints body then `[HTTP status]`.

## 2026-10-22T04:12:00Z — start; reading

#read STATE.md, charter, PROTOCOL.md, skills/sweep-variables/SKILL.md, SPEC.md, defects INDEX, lesson 02,
models/variables.md, hypotheses.md.

Prior record relevant to this charter (all reported fixed in lesson 01): D05/D07/U06 range + charset (F19),
D06/U05 batch bounds + atomicity (F16), D08 payload precision (F17), D26/U07 unknown fields/enums (F20).
Passing re-checks of those are regressions, not discoveries.

## 2026-10-22T04:15Z — #read validators and serialization paths (lesson-02 source)

- `internal/api/api.go:49-58` `decode`: `json.NewDecoder` + `DisallowUnknownFields` + `UseNumber`, 2 MiB
  `MaxBytesReader`, **one** `Decode` call. Interpretation: trailing bytes after the first JSON value are never
  read, so `{...}garbage` or two concatenated objects should be accepted. Go `encoding/json` v1 matches field
  names case-insensitively, so `"KIND"` / `"Priority"` would not count as unknown. Content-Type not checked.
- `internal/api/enqueue.go:13-21` `JobInput`; `:28-75` `insertJob`: kind non-empty; run_at XOR delay; priority
  0–9; max_attempts ≥1 (no upper bound; DB `int` column); `delay_seconds` → `now.Add(time.Duration(d)*time.Second)`
  with no range check — int64 nanosecond overflow above 9 223 372 036 s; negative delay unchecked; payload must
  start with `{` and be `json.Valid`, stored as text into `jsonb` (migrations/001.sql:4).
- `enqueue.go:76-86`: any `insertJob` error (including DB errors) → 400 `invalid_job` with raw error text.
- `enqueue.go:88-120` batch: 1–500 elements, one transaction, any element error → 400 + rollback.
- `internal/api/queues.go:12-26` `QueueInput` includes `backoff.max_seconds` (`*int`) — **not in SPEC**.
  `:38` validation: name regex `^[A-Za-z0-9_-]+$` (no length bound), lease ≥1, max ≥1, retention ≥0 (float64,
  no upper bound), base ≥0 (no upper bound), kind linear|exponential; no check on max_seconds.
  `:46-49` any INSERT error → **409 queue_conflict** (also for numeric out-of-range on `int` columns).
- `queues.go:63-95` PATCH: lease/max/retention pointers with lower bounds; `webhook` is a struct pointer, and when
  present **both** url and secret are replaced (`:90-91`) — an omitted member becomes "". `:93-94` any UPDATE
  error → 500 via `output`.
- `grep backoff_max|MaxSeconds|max_seconds`: only queues.go:20, :46 and migrations/001.sql:3 — stored, never read.
- `internal/jobs/jobs.go:17-22` `Backoff`: linear `Duration(base*attempt)*Second`; exponential
  `Duration(base)*Second*Duration(1<<(attempt-1))` — unchecked int64 overflow; `internal/jobs/nack.go:22-26` uses
  it as `run_at = now + wait`.
- `internal/jobs/expire.go` `Retain`: SQL `finished_at < $1 - q.retention_hours * interval '1 hour'` for all
  queues in one statement; `internal/api/api.go:148-154` any step error sets coreFailed (health 503) and is
  logged; loop continues. Postgres timestamptz minimum is 4713 BC, so a large retention_hours could make the
  expression error for every tick.

## Variable inventory (before values) — charter surface: enqueue, batch, queue config

| Family | Variable | Heuristics to apply |
|---|---|---|
| request shape | trailing data, concatenated objects, duplicate keys, key case, Content-Type, empty body, `null` body, array body | Strings, Selection |
| enqueue fields | kind (empty, whitespace, long, NUL, case), payload (absent, null, array, nested, NUL escape, huge numbers, -0, 1E2, duplicate keys), priority (−1,0,9,10, 5.0, "5"), delay_seconds (−1, 0, huge, overflow), run_at (past, far future, no zone), max_attempts (0,1,int4 max, int4+1), idempotency_key ("" , repeat) | Boundaries, Numbers, Zero/One/Many, Goldilocks |
| batch | count (0,1,500,501), position of invalid element (first/middle/last), body size vs 2 MiB, duplicate idem keys | Count, Position, Beginning/Middle/End |
| queue create | name (charset, length 1 / huge), lease_seconds (0,1,int4 max, int4+1), max_attempts, retention_hours (0, fractional, huge), backoff.kind/base/max_seconds | Boundaries, Goldilocks, Some/None/All |
| queue patch | partial webhook object (url only / secret only), null members, out-of-range numbers | Some/None/All, Selection |
| timing/config interaction | backoff base × attempt number (overflow), retention × clock | Numbers, Boundaries |

## 2026-10-22T04:16Z — #hypothesis (all stated before any probe; written while the stack built 04:15:13–04:16:17)

Charter prediction (H-C): "every documented range is enforced and unknown fields are rejected." Source: charter,
validators in internal/api. Wrong if any 2xx occurs where SPEC says 4xx. My own code-reading expectation is that
H-C will be refuted (E03-H01, -H02).

| Id | Prediction (code route) | Oracle (should) | Would be wrong if |
|---|---|---|---|
| E03-H01 | enqueue body `{"kind":"echo"} trailing-garbage` and two concatenated objects → 201, one job (api.go:53 single Decode) | SPEC:17 reject what it does not understand; RFC 8259 a JSON text is one value | 400 |
| E03-H02 | `{"KIND":"echo","Priority":3}` → 201 with kind echo, priority 3; `{"priority":1,"PRIORITY":9}` → priority 9 (case-insensitive match, last wins) | SPEC:17 unknown fields rejected; JSON member names are case-sensitive | 400 `unknown field` |
| E03-H03 | regressions: unknown field `x`, backoff.kind `zigzag`, priority 10/−1, name `a.b`, batch 0/501, mid-batch invalid → all 4xx, nothing inserted | SPEC | any 2xx / partial insert |
| E03-H04 | `delay_seconds: 9223372037` → 201 `queued` (overflow wraps to the past); `9223372036` → `scheduled` year ~2318 | SPEC:48-49 a future job is scheduled | 400 or `scheduled` for the overflow value |
| E03-H05 | `delay_seconds: -60` → 201 queued, run_at 60 s in past | SPEC silent on sign | 400 (would be fine) |
| E03-H06 | queue create `lease_seconds: 2147483648` → **409** queue_conflict with pq "out of range" text; PATCH same value → **500** database_error | SPEC:13-16 invalid command → 4xx (400); 409 is for state conflicts | 400 |
| E03-H07 | `backoff.max_seconds` (undocumented) accepted 201, stored as `backoff_max`, negative accepted, no effect on retry delay | SPEC:17 unknown fields rejected / SPEC:26-28 lists backoff fields kind+base only | 400, or retry delay capped |
| E03-H08 | queue backoff exponential base 2147483647: after 3 nacks with delay 0, a 4th default nack sets run_at ≈ 40 years in the **past** (overflow) and the job is re-queued immediately | SPEC:27-28 retries back off exponentially; wait must be ≥ previous | run_at in future |
| E03-H09 | queue `retention_hours: 1e8` accepted 201; once that queue has a terminal job, every Retain tick errors (timestamp out of range), /health 503, and terminal jobs of **other** queues past retention are not deleted | SPEC:98 terminal jobs deleted after retention; SPEC:127-130 loop health | 400 on create, or Retain keeps working |
| E03-H10 | PATCH `{"webhook":{"secret":"s2"}}` sets webhook_url "" (notifications silently disabled); `{"webhook":{"url":"…"}}` sets secret "" | PATCH partial-update convention (RFC 7396 merge patch: omitted members unchanged) — approximation; SPEC:33-34 | url retained |
| E03-H11 | queue name of ~1.1 M ASCII letters → 201; GET/DELETE by that name → 431 (Go 1 MiB header limit) so the queue cannot be addressed or removed | SPEC:22,35-36 a created name addresses the queue | 400 on create, or GET 200 |
| E03-H12 | payload with `"\u0000"` string or number `1e200000` → 400 with pq/jsonb text; `1E2`→100, `-0`→0, `1.50` kept, duplicate keys collapse to last | SPEC:44 payload JSON object, :46 no numeric precision loss | stored exactly as sent |
| E03-H13 | batch of 500 jobs with ~4.3 KB payloads (>2 MiB) → 400 invalid_json "request body too large", nothing stored | SPEC silent on size | 201 |
| E03-H14 | batch with the same idempotency key twice → 201, ids `[x,x]`, one job | SPEC:52 repeat returns original | two jobs |
| E03-H15 | per-job `max_attempts: 2147483648` → 400 invalid_job (pq out-of-range text) | SPEC 4xx | 201 / 5xx |
| E03-H16 | `kind: " "` (whitespace) → 201; `kind` with invalid UTF-8 byte → 201 with U+FFFD substituted | SPEC:46 kind required (silent on charset) | 400 |

Planned split: reading ~10 min (done), probing ~30 min, writing ~15 min.

## 2026-10-22T04:16:24Z — fixture

Stack up 04:16:17 (`dc up -d --build` exit 0, raw `results/e03/up.log`). `python3 lab/seed/seed.py` → "Seeded 300
jobs, 4 queues, 2 schedules"; manifest moved to `results/e03/manifest.json`. Real clock (no /admin/clock change).
Queue `imports` has no worker in Compose, so jobs enqueued there stay queued.

## 2026-10-22T04:16:44Z — #obs request shape (E03-H01, E03-H02) — raw `results/e03/p1-shape.txt`

```text
$ api POST /v1/queues/imports/jobs -d '{"kind":"echo","payload":{"t":"h01a"}} trailing-garbage'
{"id":301,"state":"queued"} [201]
$ api POST /v1/queues/imports/jobs -d '{"kind":"echo","payload":{"t":"h01b"}}{"kind":"second","payload":{"t":"h01b2"}}'
{"id":302,"state":"queued"} [201]
$ api POST ... -d '{"kind":"echo",}'                       # control: malformed first value
{"error":{"code":"invalid_json","message":"invalid character '}' looking for beginning of object key string"}} [400]
$ api POST ... -d '{"KIND":"echo","Priority":3,"PAYLOAD":{"t":"h02a"}}'
{"id":303,"state":"queued"} [201]
$ api POST ... -d '{"kind":"echo","priority":1,"PRIORITY":9,"payload":{"t":"h02b"}}'
{"id":304,"state":"queued"} [201]
$ api POST ... -d '{"kind":"echo","PRIORITY":9,"priority":1,"payload":{"t":"h02c"}}'
{"id":305,"state":"queued"} [201]
$ api POST ... -d '{"kind":"echo","kind":"other","payload":{"t":"h02d"}}'
{"id":306,"state":"queued"} [201]
$ curl ... -H 'Content-Type: text/plain' ... -d '{"kind":"echo","payload":{"t":"ctype"}}'
{"id":307,"state":"queued"} [201]
$ api POST ... --data-binary ''      → {"error":{"code":"invalid_json","message":"EOF"}} [400]
$ api POST ... -d 'null'             → {"error":{"code":"invalid_job","message":"kind required"}} [400]
$ api POST ... -d '[{"kind":"echo"}]' → {"error":{"code":"invalid_json","message":"json: cannot unmarshal array into Go value of type api.JobInput"}} [400]
$ api POST /v1/queues -d '{"name":"e03-trail"}{"name":"other"}'   → {... "id":5 ... "name":"e03-trail" ...} [201]
$ api POST /v1/queues -d '{"NAME":"e03-case","Lease_Seconds":7}'   → {... "id":6,"lease_seconds":7,... "name":"e03-case" ...} [201]
$ sql "SELECT id,kind,priority,payload FROM jobs WHERE ..."
301 | echo | 5 | {"t": "h01a"}
302 | echo | 5 | {"t": "h01b"}          # second object silently dropped: no 'second' kind row
303 | echo | 3 | {"t": "h02a"}
304 | echo | 9 | {"t": "h02b"}          # later case-variant key wins
305 | echo | 1 | {"t": "h02c"}
306 | other | 5 | {"t": "h02d"}         # exact duplicate key: last wins
307 | echo | 5 | {"t": "ctype"}
queues: e03-trail lease 30 (no 'other' queue); e03-case lease 7
```

E03-H01 confirmed (trailing data and a second object accepted; second object discarded). E03-H02 confirmed
(`KIND`, `Priority`, `PAYLOAD`, `NAME`, `Lease_Seconds` all bound; last case-variant wins).
#surprise S1 — trailing data / concatenated JSON accepted with 201 and silently partial → #bug E03-F01.
#surprise S2 — case-variant member names accepted as the documented fields → #bug E03-F02.
Not a bug (judgment): exact duplicate keys (last wins) — RFC 8259 §4 says names SHOULD be unique and leaves
behaviour implementation-defined; SPEC silent. Content-Type text/plain accepted — SPEC:12 states what clients
use, not that the server must refuse other types; recorded as #question Q-E03-1, not a defect.

## 2026-10-22T04:17:13Z — #obs documented ranges (E03-H03, regression only) — raw `results/e03/p2-ranges.txt`

```text
unknown field x → 400 invalid_json `json: unknown field "x"`; backoff.kind zigzag → 400 invalid_queue
priority -1 → 400; 0 → 201 (308); 9 → 201 (309); 10 → 400; 5.0 → 400 "cannot unmarshal number 5.0"; "5" → 400; null → 201 (310, default 5)
queue names a.b / ü-q / "" → 400 each; lease_seconds 0,-1 / max_attempts 0,-1 / retention -1 / base_seconds -1 → 400 each
PATCH imports max_attempts 0 / retention -5 / lease 0 → 400 (message always "lease_seconds must be positive")
PATCH {"backoff":...} → 400 unknown field "backoff"; ?state=bogus → 400 invalid_state
batch [] → 400; {} → 400; 501 → 400; 500 → 201 (ids 311..810); invalid element first/middle/last (priority 42) → 400 each
sql: b500 500 rows; h03p 3 rows; bpos-* 0 rows; b501 0 rows; no e03-r*/a.b/… queues; imports unchanged (2|24|30|exponential)
```

E03-H03 confirmed: every documented range/enum/batch bound probed is enforced and atomic (regression of lesson-01
fixes — not counted as discoveries). Minor: PATCH's message names lease_seconds for a max_attempts or retention
violation (`queues.go:81`); status/code are right, message is misleading — noted, not reported (cosmetic).

## 2026-10-22T04:17:32Z — #obs enqueue numbers and strings (E03-H04, -H05, -H15, -H16) — raw `results/e03/p3-numbers.txt`

```text
delay_seconds → response; sql id | state | run_at | created_at
-60          → 201 queued    817 | queued    | 2026-10-22 04:16:32.760301+00 | 04:17:32.760301
0            → 201 queued    818 | queued    | 2026-10-22 04:17:32.798329+00 | 04:17:32.798329
1            → 201 scheduled 819 | scheduled | 2026-10-22 04:17:33.829699+00 | 04:17:32.829699
9223372036   → 201 scheduled 820 | scheduled | 2319-02-01 04:04:49.07548+00  | 04:17:33.07548
9223372037   → 201 queued    821 | queued    | 1734-07-13 04:30:16.418908+00 | 04:17:33.128459
100000000000 → 201 scheduled 822 | scheduled | 2272-11-28 16:11:24.655136+00 | 04:17:33.202894   (asked ≈3169 years)
9223372036854775807 → 201 queued 823 | queued | 2026-10-22 04:17:32.26392+00 | 04:17:33.26392   (now − 1 s)
9223372036854775808 → 400 invalid_json "cannot unmarshal number ... of type int"
run_at 0001-01-01T00:00:00Z → 201 queued (824); 9999-12-31T23:59:59Z → 201 scheduled (825);
  2030-01-01T00:00:00 (no zone) → 400; +14:00 → 201 scheduled 2029-12-31 10:00:00+00 (826);
  .123456789Z → 201 queued, stored .123457 (827)
max_attempts 1 → 201 (828); 2147483647 → 201 (829); 2147483648 → 400 invalid_job `pq: value "2147483648" is out of range for type integer`
kind " " → 201 (830, hex 20); kind "ech\xffo" (raw invalid UTF-8 byte) → 201 (831) stored "ech�o" hex 656368efbfbd6f;
kind "a\u0000b" → 400 invalid_job `pq: invalid byte sequence for encoding "UTF8": 0x00`
```

E03-H04 confirmed: the overflow value wraps to 1734 and the job is immediately `queued`; additionally 1e11 s wraps
to a *different future* (2272) and int64-max lands one second in the past. Values ≤ 9 223 372 036 are honoured.
#surprise S3 → #bug E03-F03 (delay_seconds beyond Go Duration range accepted and wrapped).
E03-H05 confirmed: negative delay → queued with run_at in the past. Not a bug (judgment): SPEC silent on sign; the
job becomes due now, which is what a non-positive delay plausibly means; ordering uses created_at
(`jobs.go:24`), so a negative delay cannot jump the FIFO. Backlog Q-E03-2 (spec: reject negative delay?).
E03-H15 confirmed (400 with driver text — status right; message leaks `pq:` text, sibling of E01 Q3; not new).
E03-H16 confirmed: whitespace kind accepted (SPEC silent on kind charset, not a bug); invalid UTF-8 byte in the
request is silently rewritten to U+FFFD — RFC 8259 §8.1 requires UTF-8, so this is malformed JSON accepted with
altered data → folded into E03-F01 as a request-shape sibling (same "malformed JSON text accepted" oracle).
run_at microsecond rounding: Postgres timestamptz precision; SPEC silent; not a bug.

## 2026-10-22T04:17:59Z — #obs queue-config numbers (E03-H06, -H07 acceptance) — raw `results/e03/p4-queuecfg.txt`

```text
create lease_seconds 2147483647 → 201 (id 7); 2147483648 → 409 {"code":"queue_conflict","message":"pq: value \"2147483648\" is out of range for type integer"}
create max_attempts 2147483647 → 201 (id 8); 2147483648 → 409 queue_conflict (same pq text)
create backoff.base_seconds 2147483648 → 409; backoff.max_seconds 2147483648 → 409 (same pq text)
control: duplicate name imports → 409 queue_conflict `pq: duplicate key value violates unique constraint "queues_namespace_id_name_key"`
PATCH e03-case lease_seconds 2147483648 → 500 {"code":"database_error","message":"pq: value \"2147483648\" is out of range for type integer"}
PATCH e03-case max_attempts 2147483648 → 500 database_error (same)
create e03-maxs backoff {kind exponential, base 4, max_seconds -7} → 201, response "backoff_max":-7
retention_hours 0 → 201; 0.0001 → 201; 1e8 → 201 (stored 100000000); 1e308 → 201 (stored 1e+308)
lease_seconds 1e3 → 400 "cannot unmarshal number 1e3 ... of type int"
```

E03-H06 confirmed: out-of-range integers are reported as a state conflict (409) on create and as a server fault
(500) on PATCH, both with driver text. #surprise S4 → #bug E03-F04.
E03-H07 (acceptance half) confirmed: an undocumented `backoff.max_seconds` is accepted and a negative value stored;
the effect half is probed with E03-H08 below.
Not a bug: `1e3` rejected for an integer field — rejection is the safe side; SPEC does not require exponent forms.

## 2026-10-22T04:18:27Z — #obs retention_hours × Retain (E03-H09) — raw `results/e03/p5-retention.txt`, `p5b-retention-isolate.txt`

Real clock. Queues from 04:17:59: e03-ret-0 (retention 0), e03-ret-1e8, e03-ret-1e308 (all acme).

```text
04:18:27.563 health {"status":"ok"} [200]
control: job 832 in e03-ret-0, cancel 04:18:27.686 → row gone by 04:18:28.404; events created/cancelled/retained(04:18:28.010)
poison: job 833 in e03-ret-1e8, cancel 04:18:28.582; job 834 in e03-ret-0 cancelled right after
04:18:29.700 … 04:18:35.224 (6 samples, 1 s apart): health {"status":"degraded"} [503], job 834 rows=1 (never retained)
delayed job 835 (imports, delay 1 s) → queued after 3 s (Promote still runs; Retain is the failing step)
api log: {"level":"INFO","msg":"background task: pq: timestamp out of range"} every second (04:18:36.03, :37.03, :38.02)
04:18:57 globex queue gx-ret0 (retention 0), job 836 cancelled; 04:19:01.859 globex job rows=1, acme 834 rows=1
PATCH e03-ret-1e8 retention 24 → 200; 04:19:04.580 health ok [200], globex 836 rows=0, acme 834 rows=0, poison 833 rows=1
PATCH e03-ret-1e8 retention 1e8 → 200; 04:19:06.893 degraded [503]; DELETE /v1/jobs/833 → 204; 04:19:08.947 ok [200]
e03-ret-1e308: job cancelled → 04:19:11.604 degraded [503]; log "background task: pq: interval out of range"; DELETE job → 04:19:13.916 ok
threshold: retention 5e7 h + terminal job → ok [200]; 6e7 h + terminal job → degraded [503]; restored 24 → ok
```

E03-H09 confirmed in full. #surprise S5 → #bug E03-F05: one tenant's accepted `retention_hours` (≥ ~6e7 h, i.e.
older than PostgreSQL's 4713 BC minimum; or 1e308 → interval overflow) halts retention for **every queue in every
namespace** and holds /health at 503 while that queue owns any terminal job. Variables isolated: create vs PATCH
(both), timestamp vs interval overflow (both), namespace (globex also blocked), Promote unaffected, recovery by
lowering retention or deleting the terminal job. Threshold between 5e7 and 6e7 hours matches 2026 − (−4713) ≈
6739 years ≈ 5.9e7 h (source inference about why, the bracket itself is observed).

## 2026-10-22T04:19:44Z — #obs backoff arithmetic (E03-H08) and max_seconds effect (E03-H07) — raw `results/e03/p6-backoff.txt`, `p6b-backoff-controls.txt`

Worker 4 (`e03-prober`, kinds echo) registered via raw API; lease → nack with `delay_seconds:0` for early attempts
(so promotion is immediate), then a nack **without** delay so the queue backoff applies.

```text
e03-boexp {exponential, base 2147483647, max 10}: job 840
  attempts 1–3 (delay 0) run_at ≈ now; attempt 4 default nack 04:19:47.181 → retry_wait | 4 | 1986-08-30 06:38:09.501733+00
  04:19:49.331 → queued | 4 | 1986-08-30 …   (re-queued immediately instead of waiting)
e03-bolin {linear, base 2147483647, max 10}: job 841
  attempts 1–4 (delay 0); attempt 5 default nack 04:19:53.690 → retry_wait | 5 | 1782-07-04 20:55:55.003334+00 → queued 2 s later
e03-bodef {defaults: exponential base 5, max_attempts 40}:
  job 845: attempts 1–30 delay 0; attempt 31 default nack 04:20:43.434 → retry_wait | 31 | 2196-12-07 00:26:03 (5·2^30 s ≈ 170 y, honoured)
  job 846: attempts 1–31 delay 0; attempt 32 default nack 04:21:14.447 → retry_wait | 32 | 1782-07-04 20:57:20.723047+00
e03-maxs {base 4, max_seconds -7}: job 842 attempt 1 default nack 04:19:56.169 → run_at 04:20:00.190 (+4 s; cap ignored)
e03-maxs2 {base 10, max_seconds 1}: job 844 attempt 1 default nack 04:20:13.643 → run_at 04:20:23.621 (+10 s; cap of 1 s ignored)
```

Dead end (kept): "control" job 843 on e03-boexp at 04:20:13 — the lease returned the older re-queued job 840
(FIFO), so the nack for 843 got `409 state_conflict "sql: no rows in result set"`; 840 was left leased by worker 4
and will expire (one extra `expired` attempt on 840, not a finding). The job-845 attempt-31 row replaced it as the
non-overflow control.

E03-H08 confirmed (exponential and linear; extreme base and **default base 5** at attempt 32). #surprise S6 → #bug
E03-F06. E03-H07 confirmed: `max_seconds` accepted, stored, echoed and never applied (positive or negative).
#surprise S7 → #bug E03-F07.

## 2026-10-22T04:21:40Z — #obs PATCH webhook partial objects (E03-H10) — raw `results/e03/p7-patch-webhook.txt`, `p8-misc.txt`

```text
create e03-hook {url http://hooksink:8201/hook, secret report-secret} → 201
PATCH {"webhook":{"secret":"rotated"}}                → "webhook_secret":"rotated","webhook_url":""
PATCH {"webhook":{"url":"http://hooksink:8201/hook2"}} → "webhook_secret":"","webhook_url":"http://hooksink:8201/hook2"
PATCH {"webhook":{}}                                  → "webhook_secret":"","webhook_url":""
PATCH {"webhook":null}                                → unchanged (report-secret / …/hook)
PATCH {"max_attempts":4}                              → max 4, webhook unchanged (top-level omission = keep)
after secret-only PATCH: job 847 cancelled → webhook_outbox rows for 847: 0 (notifications silently off)
control (url+secret restored): job 848 cancelled → outbox row url …/hook secret report-secret delivered t
04:22:00 after url-only PATCH (secret now ""): job 849 cancelled → hooksink record status 200 "signature_valid":false;
  outbox 849 secret '' status 200 delivered t
```

E03-H10 confirmed: the `webhook` member is replaced wholesale, unlike top-level members, so rotating only the secret
disables notifications and changing only the URL silently drops the signing secret (partner sees invalid
signatures but queuelet records delivery). #surprise S8 → #bug E03-F08 (oracle is an approximation: SPEC:33-34
and RFC 7396 merge-patch convention; SPEC does not say how a partial webhook object is merged — flagged).

## 2026-10-22T04:22:00Z — #obs queue name length (E03-H11) — raw `p8-misc.txt`, `p8b-namelen.txt`, helper `tools/namelen.py`

```text
POST name "q"×1 100 000 (1 100 013-byte body) → 409 {"code":"queue_conflict","message":"pq: index row requires 12624 bytes, …"}
(dead end: zsh `curl "$URL/$N"` with a 1.1 M-char argv → "argument list too long"; GET/DELETE never sent)
(dead end: zsh did not word-split "$a" in a for-loop; rerun with explicit args 04:22:27)
namelen.py rand 2600 → POST 201, GET 200, DELETE 204
namelen.py rand 2800 → POST 409 "pq: index row size 2824 exceeds btree version 4 maximum 2704 for index \"qu…"
namelen.py rand 10000 → POST 409 "pq: index row requires 10024 bytes, maximum size is 8191"
namelen.py rep 100000 → POST 201, GET 200, DELETE 204
namelen.py rep 900000 → POST 409 "pq: index row requires 10336 bytes, maximum size is 8191"
```

E03-H11 **refuted** (code-reading prediction contradicted): I predicted the 1 MiB request-header limit would make a
huge name unaddressable; the running system instead refuses the name at the unique index, so every accepted name
was addressable. New observation: the effective length limit is an undocumented, *content-dependent* storage
artefact (≈2 700 bytes for incompressible names, ≥100 000 for repetitive ones) and is reported as 409
queue_conflict with driver text → folded into E03-F04 (storage-limit errors misreported as conflict).

## 2026-10-22T04:22:49Z — #obs payload fidelity (E03-H12), batch size (E03-H13), batch idempotency (E03-H14) — raw `p9-payload-batch.txt`, `p9b-payload-batch.txt`, script `tools/p9.sh`

```text
payload sent → GET /v1/jobs/{id} payload
{"a":1E2} → {"a":100}            {"a":-0} → {"a":0}            {"a":1.50} → {"a":1.50}
{"a":1e400} → {"a":1000…000} (401-digit integer literal)
{"a":12345678901234567890.123456789} → identical digits
{"a":1,"a":2} → {"a":2}          {"b":1,"a":2} → {"a":2,"b":1}  {"s":"😀"} → 😀
(04:22:49 run aborted: zsh `echo` expanded "\ud800" → "character not in range"; rerun under bash tools/p9.sh 04:23:09)
{"s":"\ud800"} → 400 invalid_job "pq: invalid input syntax for type json"
{"s":"x\u0000y"} → 400 invalid_job "pq: unsupported Unicode escape sequence"
{"a":1e200000} → 400 invalid_job "pq: value overflows numeric format"
raw byte 0xff inside a payload string → 400 invalid_job `pq: invalid byte sequence for encoding "UTF8": 0xff`
lease (worker 5, 04:23) printed through Python json.load: 852 {"a": 1.5}, 854 {"a": 1.2345678901234567e+19}
  ← INSTRUMENT ARTEFACT: my printer parsed floats; not evidence about the SUT. Raw re-check below.
batch 500 × 4300-byte pad (2 177 011 bytes) → 400 {"code":"invalid_json","message":"http: request body too large"}
batch 500 × 4100-byte pad (2 077 511 bytes) → 201 (500 ids 858…); single job with 2.2 MB payload → 400 same message
batch [{echo,k1,v1},{other,k1,v2}] → 201 {"ids":[1358,1358]}; single enqueue {third,k1} → 201 {"id":1358}; one row (echo,{"v":1})
idempotency_key "" twice → both {"id":1361} (empty string is a real key)
batch {"JOBS":[{"Kind":"echo",…}]} → 201 {"ids":[1363]}; batch body + trailing ` {"jobs":[]}` → 201 {"ids":[1364]}
```

E03-H12 confirmed. Judgments: 1E2→100, -0→0, 1e400 expansion, key reordering and duplicate collapse are jsonb
normalisations that keep numeric *value*; SPEC:47 promises no precision loss, which held — not a bug. The 400s for
lone surrogates / U+0000 / 1e200000 / raw invalid UTF-8 reject inputs at the jsonb boundary with a 4xx (no silent
acceptance), messages leak driver text (E01 Q3 sibling) — spec question Q-E03-3, not reported. Note: a 9-byte
`1e400` became 401 stored digits — storage amplification risk recorded (#question Q-E03-5), deliberately not
pushed further (would be a resource-exhaustion probe).
E03-H13 confirmed: the 2 MiB body cap means a documented 500-job batch can be refused; SPEC silent on body size,
the refusal is a clean 4xx without insertion — not a bug; code `invalid_json` for valid JSON is misleading
(spec/usability question Q-E03-4).
E03-H14 confirmed (SPEC:52 behaviour; the differing second element is silently dropped, consistent with
"returns the original job").
Batch inherits E03-F01/E03-F02 (sibling confirmed at runtime).

## 2026-10-22T04:24Z — #hypothesis sibling predictions (before sibling probes)

| Id | Prediction (code route) | Would be wrong if |
|---|---|---|
| E03-H17 | nack `delay_seconds: 9223372037` (jobs/nack.go:23-24, same Duration overflow) → run_at ≈1734, job re-queued at once | run_at in future / 400 |
| E03-H18 | heartbeat `extend_seconds: 9223372037` (jobs.go:84) → 200 and lease_expires_at in the past → expiry on next tick | 400 or future expiry |
| E03-H19 | PATCH queue body with trailing garbage and `{"MAX_ATTEMPTS":5}` → 200 and applied (same decode) | 400 |
| E03-H20 | ack body `{"LEASE_TOKEN":…,"RESULT":…}` → 200 (case-insensitive) | 400 |
| E03-H21 | raw lease response bytes keep `1.50` and `12345678901234567890.123456789` (jobs.go:65 RawMessage) | rounded digits |

## 2026-10-22T04:24:21Z — #obs siblings (E03-H17…H21) — raw `results/e03/p10-siblings.txt`, `p10b-siblings.txt`, scripts `tools/p10.sh`, `tools/p10b.sh`

```text
H17 job 1365 nack delay_seconds 9223372037 → 200; retry_wait | 1734-07-13 04:37:04.639086+00; 1.5 s later queued
H18 job 1366 leased, lease_expires_at 04:24:53; heartbeat extend_seconds 9223372037 → 200 {"ok":true};
    lease_expires_at 1734-07-13 04:37:06; 1.5 s later: queued | attempts 1 | last_error "lease expired"
H19 PATCH e03-sib '{"MAX_ATTEMPTS":5} garbage' → 200, "max_attempts":5
H20 first try: lease returned the re-queued job 1366 (FIFO), ack for 1367 → 409 state_conflict — dead end, kept.
    rerun on fresh queue e03-h20: job 1369 leased; ack '{"LEASE_TOKEN":…,"RESULT":{"ok":1}}' → 200; succeeded | {"ok": 1}
H21 first try: leased job was h20's (same FIFO interference). rerun e03-h21 job 1370 raw lease bytes:
    "payload":{"a":1.50,"b":12345678901234567890.123456789,"c":9007199254740993}
```

E03-H17, -H18 confirmed → same overflow family as E03-F03 (added to its sibling section; the heartbeat case turns a
request for a *longer* lease into an immediate expiry that consumes an attempt). E03-H19, -H20 confirmed → E03-F01
/ E03-F02 reach PATCH and the worker protocol. E03-H21 confirmed: the worker receives exact digits (D08 fix holds —
regression, not a discovery); the 04:23 float-looking output was my printer.

## 2026-10-22T04:25Z — #hypothesis E03-H22 (regression): webhook URL `ftp://x/h`, `http://` (no host), `http://bad host/x` → 400 on create; `""` → 201. Wrong if any 2xx for the first three.

## 2026-10-22T04:25:00Z — #obs regression E03-H22 and close — raw `results/e03/p11-webhook-regress.txt`, `all-logs-at-stop.txt`

```text
webhook ftp://x/h → 400 invalid_webhook; http:// → 400; "http://bad host/x" → 400; "" → 201   (E03-H22 confirmed, U10 fix holds)
final /health {"status":"ok"} [200]; jobs by state: cancelled 21, dead 18, leased 1, queued 1093, retry_wait 19, scheduled 79, succeeded 123
api+worker logs saved (2587 lines; 40 "background task" lines, all from the E03-H09 poison windows)
04:25:04 `dc stop` (project queuelet-e03 only): api, hooksink, postgres, worker-a, worker-b Exited. Volume kept (not `down -v`).
Other Docker projects were not touched: at 04:30:46 `docker ps` still listed 10 running queuelet/queuelet-e01 containers, 13 litefleet, 2 tsdb; queuelet-e03 containers all Exited.
```

#question / testability (no new instrument needed for this charter; all observations came from the API, read-only SQL,
hooksink records and API logs):
- TR-E03-1: `/health` says only `degraded`; the cause of E03-F05 was visible only in the container log. Minimal hook: health
  body names the failing step and last error (`{"status":"degraded","failing":["retain"],"error":"…"}`). Validation probe:
  repeat E03-H09 and read the cause from `/health` alone.
- TR-E03-2: no way to see the computed retry wait without a nack; a dry-run / `next_retry_at` per attempt in GET job would
  have made E03-F06 visible without 32 cycles.

#offcharter backlog (not probed): schedule create/PATCH body shape (H53), lease `max_jobs` and list `limit` coercion (H54),
enqueue DB-outage classification (H55), lease_seconds 2147483647 expiry (H56).

Surprise disposition: S1→E03-F01, S2→E03-F02, S3→E03-F03, S4→E03-F04, S5→E03-F05, S6→E03-F06, S7→E03-F07, S8→E03-F08.
Non-bug judgments with reasons are recorded inline (duplicate keys, Content-Type, negative delay, whitespace kind, µs
rounding, jsonb normalisation, jsonb-unstorable payloads, 2 MiB cap, `1e3` rejected, PATCH message wording, E03-H11 refuted
because the system refuses rather than strands the name).
