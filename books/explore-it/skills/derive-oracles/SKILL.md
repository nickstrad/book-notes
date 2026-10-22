---
name: derive-oracles
description: Write down the Never/Always rules, comparable systems and approximations that decide "wrong", find the code line behind each, try to break each, and turn the survivors into invariant checks; use when there is no spec, correctness is disputed, or a number has no independent check.
---

# Derive oracles

An output is only wrong against something. This skill writes that something down before you probe: a
table of rules, each with its authority, the code that should enforce it, a prediction and the probe that
would break it. In this repo's arc you also follow [the protocol](../../exploration/PROTOCOL.md).

## When to use

- "How would I know it's wrong?"; the spec is thin, silent or disputed.
- A number nobody checks: stats, counts, ages, totals, usage reports, dashboards.
- Ordering, retries, leases, expiry, error responses, history or audit rows.
- A surprise in a probe log nobody has judged; or a fix to turn into a rule a check enforces.

Not for: choosing what to explore first (`charter-as-probe`), or enumerating input values (`sweep-variables`).

## Inputs

- The source, any spec, README or API docs, and the fields the API publishes (an `expires_at` is a promise).
- A running instance you may probe, and read-only access to its store (`psql`, `sqlite3`).
- A controllable clock if anything is time-based (else record a testability request); names of
  comparable systems or standards; somewhere to route owner questions.

## Procedure

1. **Harvest promises.** `grep -niE 'never|always|only|at most|at least|exact|atomic|in order|within'`
   over the spec and docs. Add every published field with a meaning (expiry times, counters, "numeric
   order"). Each becomes a candidate rule.
2. **Write the rules before any probe.** One row per rule, as Never/Always, with object, state, time
   window and exceptions. Add rules a user would assume but nobody wrote (an error is in the error form; a
   more negative delay never gives a later time; history order is execution order). Timestamp the table.
   (E04 wrote 13 rules for a 60-minute charter.)
3. **Name the authority for each rule.** `S` spec line (quote `file:line`); `A` approximation or
   consistency rule, with a numeric bound chosen before observing; `C` comparable system or standard,
   for calibration only, marked "recalled, not verified" unless you checked it; `—` spec silence, which is
   an owner question with options, not a bug. Another product's behaviour is never this product's
   requirement. (queuelet: SQS counts message age from the send, so E04 said its queued-age oracle
   depended on what "readiness" means — [E04-F05](../../exploration/defects/E04-F05-queued-age-counts-leased-time.md).)
4. **Scope each rule by every path that can reach it.** Grep every writer of the rule's fields
   (`grep -rn 'UPDATE jobs SET'`, `grep -rn "state *="`): user verbs, background sweeps, operator actions,
   deletes of owners. A rule probed on one path holds only for that path; the scope column says which.
   (queuelet: "attempts never exceed max" held on nack in E04; E05 broke it through worker deletion —
   [U24](../../exploration/defects/U24-worker-delete-release-uncounted-attempt.md).)
5. **Find the enforcing line and predict.** Write the guard that should enforce each rule, or "absent".
   Look for one-sided range checks (`> Max`, no `< 0`); guards on token and state but not the expiry time;
   router fallbacks and panics that bypass the error writer; a counter that is reset and also numbers
   history; derived values (age, totals) built from a field some path forgets to update. Write "holds" or
   "violated because …" before probing. A missing guard is still only a hypothesis.
6. **Choose an independent check.** Counts: the API against a direct query
   (`SELECT state, count(*) … GROUP BY state`), exact at rest. Order: inputs of known order plus a
   same-instant tie. Errors: `curl -i`, reading status, `Content-Type` and body together. Time: the
   boundary, one unit either side, and before and after one background tick. Timers: compute the expected
   wait by hand (`base·2^(n−1)`). Two views that share one derivation are not independent.
7. **Try to break each rule.** Smallest input first: zero, negative, the type's extreme (seconds held as
   nanoseconds: ±9 223 372 037), after expiry but before the sweep, after an operator reset, the wrong
   method and an unknown path, N parallel calls on one resource (E04 used 30). Keep an untouched control
   beside each probe, and require each step to act on its own item. (queuelet: FIFO handed lesson 04's
   first REF run the wrong job — [lesson 04](../../lessons/04-judge-it.md).)
8. **Judge every rule and every surprise.** Holds, violated, depends on an owner decision, or risk only
   (needs a lab-only control such as a backward clock step). Keep "my prediction was confirmed" apart from
   "the product broke the rule". Each surprise ends in a report quoting its oracle, or a written reason it
   is not one. A note without a named oracle is not a finding. (queuelet: `zone_offset` was read in E08 and
   judged only in E09 — [lesson 09](../../lessons/09-explain-it.md), [U42](../../exploration/defects/U42-schedule-zone-offset-frozen.md).)
9. **Turn rules into checks.** A violated rule gets a black-box repro printing `PRESENT`/`ABSENT` and a
   regression check sharing its oracle; an owner decision becomes one spec line. Survivors become
   invariants asserted after every round of a seeded random walk. For a lifecycle, put the state × event
   rules in one table, test the table against the contract written out by hand, and walk the running
   system against the table. Prove each check can fail, on the old tree or with one fault reintroduced.

## Output

- A rules section in `models/rules.md` (or beside your notes):
  `Rule | Authority / bound | Scope (paths) | Enforcing source | Predicted | Probe | Observed | Judgment`.
- One report per violated rule with the oracle under "Expected"; a `Q-…` with options for every `—` rule.
- Repros, regression checks and a walker printing one line: `INVARIANTS HOLD: …` or `INVARIANT BROKEN: …`.

## Pitfalls the experiments exposed

- **A rule held on one path is reported as held.** E04 probed R32 on nack (expiry had held in E02);
  worker deletion broke it in E05 ([rules.md](../../exploration/models/rules.md), E05 revision). Step 4.
- **An earlier fix asked half the question.** Lesson 03 bounded `delay_seconds` above only (U21); U08's
  readiness fix missed the release paths (U23) ([lesson 04](../../lessons/04-judge-it.md)). Step 7 probes
  both ends of every bound on every path.
- **Validity decided by a background sweep.** A heartbeat just past expiry revived the lease 5/5 times;
  one tick later the same request got 409
  ([E04-F04](../../exploration/defects/E04-F04-expired-lease-honoured-until-sweep.md)). Step 6.
- **The probe's own items interfere.** A negative-delay job became due at once and was leased ahead of
  the job the probe meant to nack ([lesson 04](../../lessons/04-judge-it.md), Verify). Step 7.
- **A check that never failed.** `TestL04Invariants` was not run against `lesson-03`; that it breaks
  there is a stated prediction only. E09 mutation-tested its model walk instead. Step 9.

## Source

Elisabeth Hendrickson, *Explore It!* (Pragmatic Bookshelf, 2013): Chapter 5 "Evaluate Results" — 5.1
"Never and Always", 5.2 "Alternative Resources", 5.3 "Approximations"; and 10.3 (status codes versus
bodies). Section numbers follow the plan's chapter map, checked against the publisher's table of
contents; full book-text validation has not been done. Procedure wording, table format and checks are
this repo's, not quotations. Lessons: [04](../../lessons/04-judge-it.md) (rules → repros → invariant
walk), [05](../../lessons/05-sequence-it.md) (rules as walk invariants), [09](../../lessons/09-explain-it.md)
(rules as a transition table and model walk).

## Worked example

**Experiment 04 against queuelet** ([charter](../../exploration/experiments/04-judge-it/charter.md),
[notes](../../exploration/experiments/04-judge-it/notes.md),
[revealed](../../exploration/experiments/04-judge-it/revealed.md)). Charter: explore retries, ordering,
stats and error responses with written Never/Always rules, the spec and SQS/pg-boss as comparable
systems. R31–R43 were written at 12:44Z, before any probe; the run took about 15 of 60 minutes.

Score (`revealed.md`): 16 hypotheses stated, 15 confirmed, 1 refuted, 1 code-reading prediction
contradicted by the running system, 5 defects found, 18 model cells filled. The refuted one was the
charter's "the code honours every rule on my list"; none of the explorer's own code-based predictions
was contradicted.

Found, all unplanned: [U19](../../exploration/defects/U19-unmatched-route-errors-not-json.md) (router
404/405 in plain text), [U20](../../exploration/defects/U20-attempt-history-after-retry.md) (attempt
numbers reused after retry), [U21](../../exploration/defects/U21-negative-delay-accepted.md) (negative
delay before the failure, or wrapped to 2322), [U22](../../exploration/defects/U22-expired-lease-honoured-until-sweep.md)
(expired lease honoured until the sweep), [U23](../../exploration/defects/U23-queued-age-counts-leased-time.md)
(queued age counts leased time). A1, no default backoff cap, was measured (30.34 days at attempt 20) and
routed as Q-E04-1; the exported lesson 03 note had named it, so it was not a blind find. The coordinator
decided a 3 600 s default cap ([reconciliation 04](../../register/reconciliations/04.md)). Missed or
left open: the worker-delete path of R32 (found in E05 as U24), stats under concurrency (H-E04-09), ack
atomicity (H-E04-10); FIFO inversion on a backward clock step was judged lab-only (H-E04-08).

Model delta: [rules.md](../../exploration/models/rules.md) R31–R43;
[states.md](../../exploration/models/states.md) 13 cells; [variables.md](../../exploration/models/variables.md)
8 rows; [causes.md](../../exploration/models/causes.md) C20–C24. Three rows as written before probing:

| Id | Rule | Authority | Predicted (code) | Would be wrong if |
|---|---|---|---|---|
| R35 | A retry always comes after the failure and within a bounded delay | S backoff; — default cap | override `delay_seconds` negative → run_at before failure; ≤ −9 223 372 037 → wraps far into the future | 400 for these |
| R39 | An expired lease is not current | S Worker; C SQS visibility timeout | ack/heartbeat after expiry but before the Expire tick succeed | 409 |
| R40 | Queued age measures time since the job became ready | S stats | a job requeued by lease expiry reports age from original run_at | age ≈ 0 |

Rules became checks: lesson 04 compressed R31–R43 into I1–I7 in
[`invariants04.py`](../../project/lab/repro/invariants04.py) (`TestL04Invariants`:
`INVARIANTS HOLD: I1–I7 over 200 jobs, 12 rounds, seed 4`). Lesson 09 put the lifecycle in one table,
[`transition.go`](../../project/internal/jobs/transition.go), checked by `TestTransitionContract`
([transition_test.go](../../project/internal/jobs/transition_test.go)) against a hand-written contract and
by `TestL09ModelWalk` ([lesson09_test.go](../../project/test/checks/lesson09_test.go)): 1 000 random API
events against the table, which failed at step 241 when D14 was reintroduced.

**Second target: litefleet (not yet run).** The plan names no second target for this skill; this is a
sketch. Harvest rules from the Release It! platform's API and registry: no two nodes serve one database at
once (`placement.owner_node`, `epoch`, each node's accepted rows in `_node.db`); placement is published only after the node confirms; a route grant is never
used past its TTL; `GET /v1/admin/usage` equals a direct `usage_ledger` sum for the same window, and the
metering partner's received-minus-duplicate count equals what the nodes sent (independent views). Probe
placement during a node drain and grants one unit either side of the TTL, then walk create, query, drain
and backup at random while asserting the survivors.
