---
name: charter-as-probe
description: Write one bounded exploration charter (target, instruments, information sought, falsifiable predictions with refutation conditions, probe family, budget) before probing a codebase; use when deciding what to try first or next on a subsystem, a risk or an inherited system.
---

# Charter as probe

A charter turns "go look at X" into an experiment: one target, named instruments, one kind of information,
and predictions written before the probe so the running system can prove them wrong. This skill writes the
charter and checks that it is neither a tour nor a test case. In this repo's arc, also follow
[the exploration protocol](../../exploration/PROTOCOL.md): the coordinator writes the charter, and the
explorer may narrow it or add backlog items but never widens it.

## When to use

- "What should I try first?" on a codebase, subsystem or feature you have not explored.
- A risk to settle ("can one tenant block another?", "does the loop survive the database going away?").
- A request too broad for one session ("explore the scheduler"), or picking the next probe from a backlog.

Not for: characterizing a surprise you already have (use `follow-the-surprise`), or a fixed expected
response (that is a regression check; write the check).

## Inputs

- The code at the exact version you will probe (tag or commit), and a way to run it.
- The question or risk, and who owns the charter (you, or a coordinator).
- The instruments that exist: source, spec/README, raw HTTP or CLI client, read access to the database,
  logs, partner records, a controllable clock, fault controls (container stop/pause, proxy toxics).
- The prior record, if any: hypothesis backlog, models, defect reports, and what was fixed since.
- The budget (wall clock) and the permission profile.

## Procedure

1. **One target, one kind of information.** Write *Explore `<target>` with `<instruments>` to discover
   `<information>`.* The target maps to files (`grep -rl` finds it); the information is a table or a yes/no
   per row. Too broad: you cannot name the first probe. Too narrow: the whole charter is one request with
   one expected response. A first charter on a small unknown codebase may be whole-system recon (queuelet:
   E01 covered about 2 kLOC in 21.8 minutes — [revealed](../../exploration/experiments/01-map-it/revealed.md)).
2. **Read the guard at the version under test.** Each prediction cites a source: spec line, `file:line` at
   this tag, model cell, earlier surprise or risk. If a fix touched the area since the source was written
   (`git log <old>..<tag> -- <path>`, lesson notes), predict what the fix left. (queuelet: E06 predicted
   orphans from migration 001, but lesson 01 had added foreign keys — [lesson 06](../../lessons/06-relate-it.md).)
3. **Split universal claims into per-item predictions, logged before the probe.** "The code honours every
   rule" falls to one counterexample and says nothing about which. List the items (rules, state × event
   cells, neighbours, variables); give each an id, *I predict …* and *I would be wrong if …*. The refutation
   is the observable negation of that prediction, not another event (queuelet: E02's "consoles disagree"
   fired while both halves of its prediction held — [lesson 02](../../lessons/02-make-it-visible.md)).
   Append each prediction with `date -u` before its probe, mid-run ones too; post-hoc ones are kept but not
   scored (E07 excluded three, E08 excluded P3b's siblings).
4. **List what the prediction does not name**, one what-if each: every process in the compose file or
   service list, every status value (the schema's `CHECK`/enum, `grep -rn "state = '"`), every caller that
   can act on the same row. (queuelet: E08's prediction held for the two things it named and failed at three
   it did not — [lesson 08](../../lessons/08-surround-it.md).)
5. **Probe first the predictions that depend on what the code does not show:** a driver honouring a
   deadline (`context.WithTimeout`, `PingContext`, client timeouts), a database limit, load (one unbounded
   transaction), two clients on one row, the clock. Of six refutations of the explorer's own predictions
   in E01–E09, four were of this kind (E03-H11 index size, H-E06-17 race window, H-E07-13 sweep under load,
   H-E08-20 driver ignoring a deadline); two misread code that was there (H10b loop step order, H-E05-10
   which verb clears lease fields) — [hypotheses](../../exploration/hypotheses.md). Re-read loop order and
   every writer of a field before predicting.
6. **Add a second actor** to any state or lifecycle charter: two workers, or an operator verb racing a
   worker verb on the same id. (queuelet: E07's 49 single-actor cells all held; two-actor overlaps gave U34
   and U35, and E09's post-commit tail, U43/U44, was invisible to one actor —
   [lesson 07](../../lessons/07-state-it.md), [lesson 09](../../lessons/09-explain-it.md).)
7. **Pick the probe family that fits the question** (11.4): unknown surface → recon walk of every verb;
   inputs → variation table with a raw client; correctness → Never/Always rules, then break each; order →
   unlikely verb pairs, then a random walk asserting the rules; lifecycle → state table with *should* and
   *predicted* per cell; owners/data → CRUD grid and one value followed through every hop; neighbours →
   context diagram, one what-if per boundary; a defect list → causes, then predicted sites with `file:line`.
8. **Set the budget and split.** State wall clock and a reading / pre-probe writing / probing / writing
   plan. On queuelet every run used 15–24 of its 60 minutes, and writing took nearly as long as probing
   (E01: 2.6 / 11.7 / 7.6 against a plan of 20 / 30 / 10). Stop rule: at half the budget, start no new
   probe family and start the record.
9. **Keep the answer out of the charter.** A charter written for someone else states the question and the
   instruments, never expected counts, groupings or defect names. (queuelet: E09's charter carried the plan's
   expected habit counts; the explorer matched them, which lesson 09 says is not independent evidence.)
10. **Backlog, do not widen.** Adjacent leads go to the backlog with id, source, refutation and status
    (`open`, `confirmed`, `refuted`, `needs more`). A prediction with no trigger stays `open`, trigger named
    (H-E08-23 saw the risk in `Fire` without a trigger; E09 found one: U39). A read observation without a
    named oracle is not a finding (U42's field was read in E08 and not judged).
11. **Close.** Verdict per prediction, charter status (`done`, `needs more`, `blocked`), actual split. A
    `needs more` charter gets at most one more run, narrowed to the open items.

## Output

`charter.md`, written before the first probe, and a prediction table scored at close:

```text
Explore <target> with <instruments> to discover <information>.
Version: <tag/commit>. Probe family: <family>. Budget: <N> min; plan <read>/<pre-write>/<probe>/<write>.
Not named here (one what-if each): <neighbours/states/actors>.
```

| Id | I predict | Source (`file:line` @ version) | Runtime factor | I would be wrong if | Logged at | Outcome |
|---|---|---|---|---|---|---|

Close-out: status, planned vs actual split, counts (stated / confirmed / refuted / needs more /
code-reading predictions contradicted / defects found), backlog items added.

## Pitfalls the experiments exposed

- **Universal or loosely tied charter sentences carry little information.** Of 11 refutations in E01–E09,
  five were one-line charter predictions (H00, E03-HC, H-E04-C, H-E05-C, H-E06-C), each felled by its first
  counterexample; E02's refutation condition did not test its prediction at all. Step 3.
- **Stale sources.** Charters written from the plan's v0 view re-tested early fixes (E06 orphans; E08's
  "loop survives a restart"). Step 2.
- **One actor, named neighbours only.** Complete single-actor tables passed while overlaps and unnamed
  boundaries failed (E07, E08, E09). Steps 4 and 6.
- **Leaked answers and post-hoc predictions** (E09 counts; E07/E08 post-hoc lines). Steps 3 and 9.
- **Refutation rate is not discovery rate.** E02 and E09 refuted nothing and filed 3 and 7 reports; a
  confirmed prediction of a defect is still a finding. Score finds separately.

## Source

Elisabeth Hendrickson, *Explore It!* (Pragmatic Bookshelf, 2013): Chapter 2 "Charter Your Explorations",
2.1–2.3 (the *Explore … with … to discover …* template; charters neither too broad nor too narrow) and 2.5
(here only a source of risk hypotheses); Chapter 1 "On Testing and Exploration", 1.3 (the time box as a
budget); Chapter 11 "Explore an Existing System", 11.4 "Selecting Techniques". Numbers follow
[PLAN 2.2](../../PLAN.md), checked against the publisher's contents; this repo records no titles for 1.3 and
2.1–2.5, and full book-text validation has not been done. The prediction/refutation lines, scoring and
procedure wording are this repo's. Demonstrated by every charter, e.g.
[E01](../../exploration/experiments/01-map-it/charter.md),
[E05](../../exploration/experiments/05-sequence-it/charter.md),
[E09](../../exploration/experiments/09-explain-it/charter.md); lessons [01](../../lessons/01-map-it.md),
[08](../../lessons/08-surround-it.md), [09](../../lessons/09-explain-it.md).

## Worked example

How the nine queuelet charters fared (from each `revealed.md`; all closed `done` after one run):

| Exp | Charter prediction | Verdict | Minutes of 60 |
|---|---|---|---|
| E01 | README happy path works and the spec is complete | refuted (happy path held) | 21.8 |
| E02 | crashed job retries to max; webhook delivered only on 2xx | confirmed | ≈22 |
| E03 | every documented range enforced, unknown fields rejected | refuted on the first probe | ≈21 |
| E04 | the code honours every rule on my list | refuted (5 rules broke) | ≈15 |
| E05 | every verb does what it says or is refused | refuted by the fourth probe | ≈24 |
| E06 | deleting an owner leaves orphans | refuted (holds only job → attempts) | ≈20 |
| E07 | per cell, from the guard | 49 single-actor cells held; 1 refuted | ≈20 |
| E08 | routes namespace-scoped; loop survives a DB restart | confirmed as named; refuted at 3 unnamed boundaries | ≈22 |
| E09 | the habits predict unfound defects | not refuted (6 of 9 sites gave reports) | ≈20 |

**E09.** Charter: *Explore the defect reports and models so far, and every `UPDATE` and firing query in
`internal/jobs` and `internal/schedule`, to discover the small number of habits that produced the 26
findings and to predict, before probing, where the same habits put a defect nobody has found*
([charter](../../exploration/experiments/09-explain-it/charter.md)). Four minutes in, before any probe, the
explorer wrote nine predicted sites PD1–PD9 with `file:line`. Score
([revealed](../../exploration/experiments/09-explain-it/revealed.md)): 21 stated, 19 confirmed, 0 refuted,
0 code-reading predictions contradicted, 7 defects found, 20 model cells filled. Found
[U39](../../exploration/defects/U39-one-schedule-stops-later-schedules.md),
[U40](../../exploration/defects/U40-never-matching-cron-accepted.md),
[U41](../../exploration/defects/U41-cron-zone-prefix-overrides-timezone.md),
[U42](../../exploration/defects/U42-schedule-zone-offset-frozen.md),
[U43](../../exploration/defects/U43-terminal-notification-after-commit.md),
[U44](../../exploration/defects/U44-vanished-job-drops-pass-notifications.md),
[U45](../../exploration/defects/U45-non-iana-local-timezone.md). Wrong or missed: PD5, PD8, PD9 behaved as
predicted but are not defects; drafts E09-F04/F05 were withdrawn; the worst outcome, an ack answered 200
whose notification was never written (U43), was not predicted; the charter's expected counts made the
grouping non-independent. Model delta: `causes.md` (habit table, new habit H-F), `states.md` 5 cells,
`rules.md`, `variables.md` 6 rows, `ecosystem.md` 4 rows. From [causes.md](../../exploration/models/causes.md)
(rows trimmed):

| Predicted site | Habit | Outcome | Report |
|---|---|---|---|
| PD1 never-matching cron accepted → Fire stops at it every tick | H-B + H-E | confirmed; cross-namespace | E09-F01, E09-F02 |
| PD6 error after commit answered 409 | H-D | confirmed; **plus** silent 200 with no notification — not predicted | E09-F07 |

**Second target: litefleet (not yet run).** PLAN 7 names no second target for this skill; this is a sketch.
A first charter on Release It!'s litefleet: *Explore the gateway's query path with the gateway and control
source at the tag, Toxiproxy on gateway → control and gateway → node, node `/metrics` and the registry tables,
to discover what a query does when the control plane is slow or gone.* Step 3 gives one prediction per
integration point (control, node, blob store, metering), each citing its timeout line and refuted by "the
query hangs past that deadline"; step 4 lists the unnamed neighbours (the second node, the usage report, the
load balancer); step 5 probes the deadline predictions first, since a driver or proxy may not honour them;
step 6 puts two tenants on one node. Budget 60 minutes, planned 5 / 2 / 15 / 8, at a lesson tag so the
predictions are about what the fixes left.
