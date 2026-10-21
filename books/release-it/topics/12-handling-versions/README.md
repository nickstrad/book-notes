# 12 · Handling Versions and the expand/contract rollout

> **STUB** — teaching outline from `../../PLAN.md` (topic 12). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 14 "Handling Versions" (help others handle your versions:
nonbreaking vs breaking API changes; handle others' versions); Ch. 13 "Phases of Deployment"
(relax between phases; expansion; rollout; cleanup), "The Fallacy of Planned Downtime"; Ch. 12
case study.
**Language:** Node — JSON APIs are where this bites, and strict vs tolerant parsing is a visible
one-line difference in JavaScript.
**Run:** `./run.sh`.
**Budget:** 10 s.

## The idea

With many machines and no planned downtime, old and new versions run at the same time during every
rollout. Make changes nonbreaking: add optional fields, never remove or redefine; be a tolerant
reader (accept unknown fields, require only what you use). Sequence a breaking change as expansion
→ rollout → contraction, relaxing between phases. Treat others' versions defensively.

_Builder: 2–4 paragraphs; cite Ch. 14's list of nonbreaking changes and Ch. 13's phase names;
mention the Ch. 12 schema-migration outage._

## What you will see

A simulated rolling deploy: 3 server instances, a client fleet, a balancer picking a random
instance per request. The change: v2 renames `total` → `amount` and adds `currency`. Per rollout
step: instances by version, requests `ok` / `failed` by cause.

### Before

Strict v1 clients (`assert typeof body.total === 'number'`), strict v2 servers rejecting v1
bodies. Error rate 33% → 66% → 100% as instances upgrade.

```
BREAKING ROLLOUT: 61% of requests failed during mixed versions
```

### After

Expansion (v2 sends `total`, `amount`, `currency`; accepts both request shapes), clients made
tolerant readers, rollout with zero errors, then contraction (drop `total`) only after every
consumer reports v2. A `--skip-relax` flag shows contraction before consumers moved.

```
SAFE ROLLOUT: 0 failures across expansion, rollout and contraction
```

## How it works

_Builder: `parseStrict` vs `parseTolerant`; the server's expansion-phase response builder;
`before.mjs`, `after.mjs`, seeded PRNG._

## Try changing

_Builder: e.g. change the type of `total` instead of renaming it (also breaking); add a required
request field in v2 (breaking for old clients)._

## Where to read more

Ch. 13 "Phases of Deployment"; Ch. 14; Ch. 12; the `plan-safe-release` skill.

## Deviations from plan

None yet.
