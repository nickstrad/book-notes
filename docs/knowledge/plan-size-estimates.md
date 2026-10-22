# Plan size estimates against delivered lessons

Learned from Release It! (`books/release-it/REVIEW.md` L9). Use it when estimating the next book's
lesson sizes.

Code-only lines per lesson (tests, lab, load generator and seed generator excluded), delivered
against the plan's estimate:

| Lesson | 01 | 02 | 03 | 04 | 05 | 06 | 07 | 08 | 09 | 10 | 11 | 12 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Plan | 350 | 170 | 130 | 150 | 190 | 220 | 130 | 150 | 160 | 170 | 320 | 290 |
| Delivered | 679 | 207 | 115 | 175 | 390 | 263 | 119 | 306 | 265 | 273 | 194 | 507 |

- Lessons that add a new mechanism (instrumentation, a circuit breaker with a cache, streaming
  results, a migration job plus a drain) ran about 2× the estimate; single-rule fixes landed close.
- Every lesson over ~400 total lines was split into two commits, as the lesson contract says.
- "A reader can follow the diff in minutes" held for 02, 03, 04, 06, 07, 09, 10 and 11; 01, 05, 08
  and 12 take closer to twenty minutes.
