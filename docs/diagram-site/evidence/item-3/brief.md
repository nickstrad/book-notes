You are Sol (GPT-6.1 Sol), a fresh build run working in the book-notes repo on
vm-1: /root/Software/book-notes (branch main; origin github.com/nickstrad/book-notes).

Read first, in this order: the repo AGENTS.md, docs/diagram-site/PLAN.md in
full, docs/diagram-site/STATE.md (items 1–2 complete and pushed), then your
source material for this item: the Explore It! project on branch
plan/explore-it — read it in its worktree /root/Software/booknotes-work/explore-it
(books/explore-it/: PLAN.md, WALKTHROUGH.md, STATE.md, the lessons, the defect
register, and the queuelet project code at tag explore-it/v1-explored). That
worktree is READ-ONLY for you: an Astra review run is active there. Read
committed state only; write nothing there.

YOUR CHUNK — PLAN ITEM 3 ONLY: the queuelet C4 diagram dataset, exactly as the
plan specifies (diagram types/levels, content, fidelity verification against
the v1-explored code, acceptance checks including Playwright). Model the
architecture AS EXPLORED AND FIXED at v1-explored: the queue service/API,
Postgres, workers, the webhook delivery path and its fake in the lab, the
scheduler/clock, and the lab/deployment surroundings the plan includes. Note in
STATE.md that publication of this dataset follows the plan's sequencing
(Explore It! end review + merge) even though authoring happens now. Items 4–8
are separate fresh runs — do not start them.

Rules:
- Dataset as decoupled JSON in the location/format the shell and item 2
  established, consistent with the litefleet dataset's conventions.
- Follow the plan's fidelity-audit method (source selectors pinned to the
  v1-explored commit) so a reviewer can check every element against the code.
- Run the plan's acceptance checks for item 3 and record results; update
  docs/diagram-site/STATE.md at the end of your run.
- Commit on main in the repo's message style; do NOT push (Twin verifies and
  pushes). Do not modify books/release-it/, books/explore-it/ content on any
  branch, or the explore-it worktree.
- Leave the tree clean. Print a summary: diagrams authored (types, element and
  relationship counts), fidelity evidence, acceptance results, deviations,
  commit SHAs.
