# 09 · Steady State

> **STUB** — teaching outline from `../../PLAN.md` (topic 09). The phase-2 builder fills each
> section, adds the code, and removes this banner. Keep the section order.

**Source:** *Release It!* 2nd ed., Ch. 5 "Steady State" (data purging; log files; in-memory
caching; avoid fiddling).
**Language:** Bash — real-world recycling is cron + `find -mtime` + rotation; Bash is the language
the remedy is actually written in.
**Run:** `./run.sh` (works in a `mktemp -d` scratch dir, removed on exit).
**Budget:** 10 s.

## The idea

For every mechanism that accumulates a resource, some other mechanism must recycle it. Logs that
grow forever, tables never purged, caches without bounds all kill a correct system, just later,
and usually at 5 a.m. Humans logging in to clean up ("fiddling") is a symptom, not a remedy.

_Builder: 2–4 paragraphs; cite the three sub-sections (data purging, log files, in-memory
caching) and the book's warning against fiddling._

## What you will see

`service.sh` simulates one request per tick, each writing a log line, a session file and a cache
entry; one tick is one simulated day. A fixed KB quota stands in for the disk. Per day: `log_kb`,
`sessions`, `cache_entries`, `total_kb`, `headroom_kb`.

### Before

No housekeeping. `total_kb` climbs daily; on the day it crosses the quota the service's write
fails and it exits 1.

```
DISK FULL on day 19: no mechanism ever recycled what the service accumulated
```

### After

`housekeeper.sh` runs alongside: rotates logs keeping 7, purges sessions older than 3 days, caps
the cache at 50 entries by age. `total_kb` plateaus by day 8 and stays flat through day 30.

## How it works

_Builder: the three housekeeping commands (`mv`/`gzip` rotation, `find -mtime +3 -delete`,
`ls -t | tail -n +51 | xargs rm`) and how "days" are simulated via `touch -d`; `before.sh`,
`after.sh`, `service.sh`, `housekeeper.sh`._

## Try changing

_Builder: e.g. keep 30 logs (plateau is higher but still a plateau); run the housekeeper every 5
days (sawtooth)._

## Where to read more

Ch. 5 "Steady State"; Ch. 8 "Transparency" → "Logging" for what the logs should contain.

## Deviations from plan

None yet.
