# Diagram site state

Goal: static C4 catalog for the book projects, per [PLAN.md](PLAN.md).
Nick approved destination B (`pages/` in book-notes, `/book-notes/`) and the plan.
The explicit item briefs supersede the proposal's pending gate and worktree default.

Status (2026-10-23): item 2 complete, verified locally and committed on main for Twin
verification/push. Items 3–8 remain separate runs; none was started here.

## Done and in progress

- Approved plan: `e147656`.
- Item 1 shell: `45795e7`, knowledge/handoff `c5173a6`, pushed before this run.
  Reuse sources: ts-db `fac3e3e5a971dcc81e62f454bb10a7643968b91e`;
  c4-sdk `bb12c7b3eb4a70152ed5ee13094da4bf7fa985be`, published dependency 0.1.0.
- Item 2: `pages/public/architectures/litefleet.json`, seven views, plain explanations
  and STE drafts, [complete fidelity inventory](../../pages/content/litefleet/inventory.md),
  [machine inventory](../../pages/content/litefleet/inventory.json), L01–L12 mapping,
  and `pages/tests/litefleet.{test,spec}.ts`. Dataset commit: `2047046` on main.
- Source snapshot rechecked: `release-it/v1-hardened` resolves to
  `936786399d7bab3ebbfa845b9a07df02b41ed363`. Working project code matches that tree.
- No changes to book content, the Explore It! worktree, shared shell, catalog,
  remote, hosting or deployment. No delegated agents or background servers remain.

## Item 2 acceptance

- Node 22.23.2: `cd pages && npm ci` exit 0; `npm run check` exit 0:
  TypeScript, 58 unit tests, production build at `/book-notes/`, 22 Chromium tests.
- After refining screenshot scroll positions, the two affected browser cases were
  rerun with `npx playwright test tests/litefleet.spec.ts -g 'all view screenshots'`:
  exit 0. The dataset and application code did not change after the full check.
- SDK decode: zero diagnostics. All 49 elements and 84 relationships appear in a
  view and in the 133-row audit. All 102 source selectors/lines match pinned git
  objects; all 31 unique public source paths returned HTTP 200. All twelve lesson
  notes, mapped source paths and view IDs resolve.
- All seven routes render and reload with correct headings and documents. Real
  context → containers → gateway/node component links work, as do Up, Back,
  plain/STE explanation switching and both dynamic views' step controls. Browser
  checks reject page/console errors, failed requests and wrong-base worker URLs.
- Inspected 1600×1000 and 390×844 captures for all seven views, with native-scale
  views, Fit overviews and narrow explanation drawers. Controls and explanations
  are readable; no horizontal page overflow. Dense diagrams require pan/zoom to
  read all labels; Fit is a topology overview. The SDK deliberately starts large
  views on the first node at readable scale. Keys and text alternatives remain
  available below the canvas; the deployment view also exposes its physical tree.
- Evidence: [result](evidence/item-2/result.json), [full check](evidence/item-2/check.txt),
  [install](evidence/item-2/install.txt), [capture rerun](evidence/item-2/screenshots-check.txt),
  [public URL checks](evidence/item-2/source-urls.json), [brief](evidence/item-2/brief.md),
  [inspection notes](evidence/item-2/inspection.md), screenshots under
  `evidence/item-2/screenshots/`.

## Fidelity findings and deviations

- As-built behavior wins over the original plan: no uncertain-write retry or
  `draining` retry; cached routes expire; gateway streams by transcoding; registry
  usage comes from heartbeats independently of partner delivery. The outbox protects
  overflow/failed sends/graceful spill, with no crash-lossless or exactly-once claim.
  Node-local state holds `node_info`/`usage_outbox`; placement metadata is in tenant
  files. Registry pools are 8/2/1; handle LRU is 128. See inventory for source proof.
- No functional scope deviation: exactly seven views. The displayed snapshot uses
  tag + short SHA to fit narrow screens; all source references retain the full SHA.
  STE metadata remains draft, with no certification claim.
- Production catalog integration is item 4. The slice tests supply a test-only
  Litefleet catalog while loading the real built JSON and worker. Default shell
  fixtures and their tests remain intact for the next slice.
- Workflow override: `/root/booknotes-sol-diag2.md` explicitly requested main and
  commits without push; this fresh run used the main checkout, not a worktree.
- Independent fidelity review belongs to item 6 under the explicit brief. No book
  chaos suites were run for this static dataset.
- Knowledge/script assessment complete: [diagram-dataset-fidelity.md](../knowledge/diagram-dataset-fidelity.md)
  records pinned-object audits and SDK capture/navigation behavior, in a separate
  indexed knowledge commit. Existing slice tests supply reusable checks; no extra
  work script is needed.

## Next steps and owners

1. Twin verifies item 2 and pushes the local commits.
2. Fresh Sol run: item 3, Queuelet slice, with acceptance against its reviewed merged
   snapshot. Its coordinator owns that source prerequisite; leave the worktree alone.
3. Follow items 4–8 in order: catalog integration, CI/rehearsal, Fable review,
   Opus fixes/recheck, publication. No publication was performed by this run.

## Open decisions and blockers

- Coordinator: resolve ts-db shell reuse terms before release; source checkout has
  no license file. Attribution remains in `pages/README.md`.
- Explore It! coordinator: provide reviewed merged snapshot for Queuelet acceptance.
- Item 5 coordinator: local fidelity tests use the pinned source git objects. Preserve
  source history in the rehearsal checkout (for example `fetch-depth: 0`).

## Resume

Checkout `/root/Software/book-notes`, branch `main`. No live run to resume. Retained
run evidence is above; scratch authoring files are in `/tmp/booknotes-sol-diag2/`.
Use Node 22.23.2 and `cd pages && npm ci && npm run check`. Item 2's Playwright
catalog override is in `tests/litefleet.spec.ts`; do not mistake the unchanged shell
catalog for missing dataset work. No worktree cleanup is required for this explicit
main-checkout run.
