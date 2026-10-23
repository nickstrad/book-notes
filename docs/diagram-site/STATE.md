# Diagram site state

Goal: static C4 catalog for the book projects, per [PLAN.md](PLAN.md).
Nick approved destination B (`pages/` in book-notes, `/book-notes/`) and the plan.
The explicit item briefs supersede the proposal's pending gate and worktree default.

Status (2026-10-23): item 4 complete. Production catalog contains Litefleet (default,
seven views) and Queuelet (five views). Queuelet is reconciled against the merged
Explore It! fix pass. Full local gate passes: 66 unit tests, 34 Chromium tests,
TypeScript and production build. Items 5–8 remain separate runs.

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

## Item 3 — Queuelet authored snapshot

Dataset/evidence commit: `4765098` on main (not pushed by this run).

- Five views in `pages/public/architectures/queuelet.json`: context L1 (5 elements /
  4 relationships), containers L2 (9/8), API components L3 (12/17), job-flow dynamic
  (6/13, 16 conditional/success scenario steps), lab deployment (9/8).
  Model totals: **27 unique elements, 49 unique relationships**.
- Pinned tag `explore-it/v1-explored` resolves to
  `b6df7db044782dd3517d041a8aad95308a3a80c9`. Its 76-row ID audit, 67 exact
  source selectors and E01–E09 mapping are in the [inventory](../../pages/content/queuelet/inventory.md)
  and [machine audit](../../pages/content/queuelet/inventory.json).
- Read committed objects only from Explore It!. Later committed walkthrough at
  `e750542457e8b42fce28108714d72352ff19ae71` supplies orientation; no project-code
  delta from the pinned tag. WALKTHROUGH does not exist at the tag and is not
  presented as tag evidence. The active review worktree was not modified.
- Source truth: lifecycle sweep, scheduler, shipper, clock and observations live
  inside API; worker execution is outside it. Terminal changes and notification
  snapshots share a SQL transaction. Delivery can repeat; health/auth bounds do
  not cover every database operation. Historical models are not final-code claims.
  Open hypotheses, legacy policy questions and review limits remain in inventory.
- Scope/workflow: explicit `/root/booknotes-sol-diag3.md` requests main and a commit
  without push; this fresh run uses the main checkout. Catalog and shared shell
  remain for item 4. Browser tests override only catalog loading and exercise the
  real JSON, production assets and worker. No book content, book tags, remote or
  hosting changes. No delegation or book chaos suites.
- Sequencing: local authoring and checks do not close the plan's reviewed-merged
  snapshot gate. The Explore It! coordinator must provide the accepted merged SHA;
  reconcile its delta before final fidelity acceptance/publication. Public source
  URL verification is deferred until that prerequisite and item 4. No push or deploy.

## Item 3 local acceptance

- Node 22.23.2: `cd pages && npm ci` exit 0; `npm run check` exit 0.
  TypeScript, all 65 unit tests, production build at `/book-notes/`, and all 30
  Chromium tests pass (8 Queuelet browser cases). SDK decoding: zero diagnostics.
- All 76 modeled IDs are drawn and audited; all 67 source selectors/lines match
  pinned git objects. Nine experiment notes, mapped paths and view IDs resolve.
- Five routes render and reload. Context → containers → API components drill-down,
  Up, Back, plain/STE documents and dynamic Next/Previous controls work. Tests
  reject page/console errors, failed requests and wrong-base worker URLs.
- Inspected desktop 1600×1000 and narrow 390×844 screenshots for all five views,
  with native-scale, Fit overview and narrow explanation captures. No horizontal
  page overflow. Large views require pan/zoom for readable labels; Fit is a
  topology overview. Keys, text alternatives and deployment physical tree remain
  available. No new rendering exception or functional plan deviation.
- Evidence: [result](evidence/item-3/result.json), [full check](evidence/item-3/check.txt),
  [install](evidence/item-3/install.txt), [source audit](evidence/item-3/source-audit.json),
  [inspection](evidence/item-3/inspection.md), [brief](evidence/item-3/brief.md), and
  35 captures under `evidence/item-3/screenshots/`.
- No live background run/server remains. Reusable slice tests supply verification;
  no new work script is needed. The separate knowledge checkpoint updates indexed
  [dataset fidelity guidance](../knowledge/diagram-dataset-fidelity.md) with the
  provisional-snapshot handoff finding.

## Item 4 — production catalog and source reconciliation

Catalog/evidence commit: `6243bc3` on main.

- Main checkout, per explicit [brief](evidence/item-4/brief.md); commit and push authorized.
  `pages/public/catalog.json` now contains exactly Litefleet and Queuelet, default
  Litefleet. Removed invented shell datasets from production assets; third-entry
  fixture stays test-only. No changes to `pages/src/`, either real dataset JSON or books.
- Queuelet provisional → reconciled: reviewed fix-pass record `1712ced` is merged
  at `7d26bb82db862b2fd483589f6b903d8dae26ae33`. Tag `explore-it/v1-explored`
  still resolves to `b6df7db044782dd3517d041a8aad95308a3a80c9`.
  The entire project tree is identical. All 67 original pinned selectors pass;
  66 remain at their original lines in the merge. `src-67` in lesson 09 moved
  from line 65 to 68 after citation edits; its structural-fix section is identical.
  The valid old-SHA URL retains line 65. No pin/content changes or fidelity problem.
  Inventory records the reconciliation; tests cover merged selectors and section.
- All 64 unique commit-pinned source page URLs across both datasets returned HTTP 200.
  Every source fragment is checked against its pinned git object by dataset tests.
- Node 22.23.2: `npm ci` and final `npm run check` exit 0: TypeScript, 66 unit/schema
  tests, production build at `/book-notes/`, 34 Chromium tests. Slice browser tests
  now use the unmodified production catalog for all twelve view/reload pairs,
  drill-down, Up/Back, documentation and scenario controls. Shared checks cover
  exactly two dropdown entries, default, switches both ways, Forward, keyboard,
  unknown/malformed routes, catalog errors, each bundle's malformed/failed fetch,
  and a third-entry fixture proving catalog extensibility. Base-relative assets
  and worker paths are checked, with runtime/request errors rejected.
- Inspected desktop 1600×1000 and narrow 390×844 context captures for both datasets.
  No page overflow/new overlap. Long select text truncates at narrow width while
  the complete dataset title wraps below; existing pan/zoom behavior remains.
- Evidence: [result](evidence/item-4/result.json), [final check](evidence/item-4/check.txt),
  [install](evidence/item-4/install.txt), [reconciliation](evidence/item-4/reconciliation.json),
  [public URL checks](evidence/item-4/source-urls.json), [inspection](evidence/item-4/inspection.md),
  four screenshots in `evidence/item-4/screenshots/`. Initial check passed 30 browser
  tests; two new HTTP-error assertions were corrected to the SDK's generic HTTP 503
  contract before the final 34-test gate (initial output retained).
- Authored JSON explanation text keeps its snapshot-era review wording under the
  explicit content boundary. Current inventory/STATE records the resolved merge gate;
  independent all-view fidelity review remains item 6.
- Knowledge assessment complete: updated indexed [dataset-fidelity guidance](../knowledge/diagram-dataset-fidelity.md)
  in the separate knowledge checkpoint with merge/line reconciliation and production-catalog test guidance.
  Existing tests are reusable; no additional work script needed.

## Next steps and owners

1. Fresh Sol handles item 5 (CI and clean-checkout release rehearsal).
2. Fable reviews all twelve views in item 6; Opus handles fixes/recheck, then Sol
   publication/live verification in items 7–8.

## Open decisions and blockers

- Coordinator: resolve ts-db shell reuse terms before release; source checkout has
  no license file. Attribution remains in `pages/README.md`.
- Item 5 coordinator: source fidelity tests require historical git objects,
  including reconciled merge `7d26bb8`; preserve source history (`fetch-depth: 0`).
- Queuelet merge prerequisite resolved; independent all-view fidelity review remains item 6.

## Resume

Checkout `/root/Software/book-notes`, branch `main`. Item 4 has no background run or
server to resume; scratch is `/tmp/booknotes-sol-diag4/`. Node 22.23.2:
`cd pages && npm ci && npm run check`. Catalog and dataset slices now exercise real
production data without overrides; only third-entry and error tests intercept loading.
No worktree cleanup is required for this explicitly authorized main-checkout run.
