You are Sol (GPT-6.1 Sol), a fresh run on the book-notes diagrams site on vm-1,
working in the MAIN checkout /root/Software/book-notes (branch main; commit
and push when green — GitHub Pages updates are ungated for this project).

CONTEXT: read docs/diagram-site/PLAN.md (your item 4: Catalog integration)
and docs/diagram-site/STATE.md first. Items 1–3 are done and on main. Item 4
was held until Explore It! merged; that merge is now complete (main contains
books/explore-it/ through the end-review fix pass, record commit 1712ced;
explore-it/v1-explored still points at b6df7db — the fix pass changed docs
and the run recipe, not queuelet behavior).

YOUR TASK — item 4, catalog integration:
1. Verify the queuelet dataset's pinned source selectors still hold against
   the merged state (the dataset is pinned to landing commit b6df7db;
   confirm the pinned files/functions/line ranges are unchanged at the
   current tag and reconcile any delta you find, recording it in STATE.md).
2. Integrate BOTH real datasets (litefleet + queuelet) into the site
   catalog so the production build's dropdown lists them (replacing any
   shell-fixture-only catalog state) with working hash routing to every
   view of both datasets.
3. Keep the established architecture: JSON datasets decoupled from the app,
   data loaded as a module dependency, dumb views, logic in hooks/pure
   functions. Reuse what items 1–3 built; do not restructure the shell.
4. Run the full local gate: unit/schema tests, production build, Playwright
   catalog tests covering BOTH datasets end to end. All green before commit.
5. Update docs/diagram-site/STATE.md (item 4 done, evidence, the queuelet
   provisional→reconciled status) and the repo knowledge notes if the
   earlier items' pattern calls for it. Commit and push to main.

BOUNDARIES: do not modify books/ content (either book) or the datasets'
diagram content beyond selector reconciliation deltas you document. If the
queuelet reconciliation turns up a real fidelity problem (not a selector
drift), STOP, record it in STATE.md, and end your run with a clear summary —
do not improvise a dataset rewrite. End with a printed summary: catalog
contents, test counts, build result, push SHA.
