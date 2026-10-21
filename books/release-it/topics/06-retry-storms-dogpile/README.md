# 06 · Dogpile, Self-Denial Attacks and jittered backoff

> **STUB** — teaching outline from `../../PLAN.md` (topic 06). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 4 "Dogpile", "Self-Denial Attacks", "Users", "Force
Multiplier" (mention); Ch. 5 "Governor", "Handshaking".
**Language:** Node — timers on one event loop make the scheduling story literal, and retry loops
are most often written in JavaScript HTTP clients.
**Run:** `./run.sh` (`export PATH="$HOME/.nvm/versions/node/v22.23.2/bin:$PATH"` if `node` is missing).
**Budget:** 25 s (two 10 s simulations).

## The idea

Synchronized behaviour by many clients produces pulses of demand: identical retry timers, caches
that expire together, cron at the top of the hour, a fleet rebooting after a deploy, a marketing
email at 9:00. The pulse knocks the dependency over the moment it recovers. Remedies: randomness
(jitter), exponential backoff, retry budgets, and governors on automation.

_Builder: 2–4 paragraphs; cite the book's dogpile causes (booting after an outage, cron, config
pushes) and the self-denial attack examples (deep links, promotions)._

## What you will see

200 simulated clients against a dependency with capacity 50 req/s that goes down for 2 s. An ASCII
histogram of requests per 100 ms bucket over 10 s, then totals: `succeeded`, `failed`,
`overload_ticks`.

### Before

Fixed 1 s retry delay. Every failure cohort retries in lockstep; on recovery the dependency gets
200 requests in one tick, exceeds capacity, fails again; repeating spikes.

```
DOGPILE: recovery failed 4 times under synchronized retries
```

### After

Exponential backoff with full jitter (`random(0, base * 2^n)`), capped delay, maximum 3 retries.
The histogram flattens under capacity and recovery sticks on the first try.

## How it works

_Builder: the `retryDelay(attempt)` function in both files is the entire difference; a seeded PRNG
so the output is reproducible; `before.mjs`, `after.mjs`, no dependencies._

## Try changing

_Builder: e.g. jitter without backoff; backoff without jitter (still pulses); a governor that caps
retries fleet-wide._

## Where to read more

Ch. 4 "Dogpile", "Self-Denial Attacks"; Ch. 5 "Governor"; the `budget-timeouts-and-retries` skill.

## Deviations from plan

None yet.
