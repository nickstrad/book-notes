# Diagram site state

Goal: static C4 catalog for the book projects, per [PLAN.md](PLAN.md).
Nick approved destination B (`pages/` in book-notes, `/book-notes/`) and the plan.
The explicit item briefs supersede the proposal's pending gate and worktree default.

Status (2026-10-23): item 7 fix pass complete (see "Item 7" below): H1, M1–M3 and
L1–L3, L6, L7 fixed; L4, L5 deliberately left with recorded reasons; M4 is Nick's
LICENSE decision. Local gate green (66 unit, 35 Chromium, build, manifest with license
notices). Next: a fresh Fable recheck of the fix pass, then item 8. Item 6's review is
in [REVIEW.md](REVIEW.md); items 1–5 remain as recorded below.

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
  short SHA + "local tag" label and a pinned-commit link (item 7, L2); all source
  references retain the full SHA.
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

1. Fresh Fable recheck of the item 7 fix pass (diff `7f5d316..` the item 7 record).
2. Sol: item 8 live verification, including the new `licenses/` files and Licenses link.

## Open decisions and blockers

- **Nick (open):** whether `book-notes` gets a LICENSE (currently `license: null`, so
  visitors have no reuse rights beyond viewing). No license file added (review M4).
  The ts-db reuse question is closed on the facts: ts-db and book-notes share one owner
  (nickstrad, public repo, sole author of ts-db `pages/`); attribution stays in
  `pages/README.md`.
- Upstream request (not filed; Nick/c4-sdk owner): a per-view `external` override in
  c4-sdk, so a container can be grey only in the component views that need the tag (M2).
- CI fetches full history, including reconciled merge `7d26bb8`. Origin has no
  book tags; fidelity checks use immutable source SHAs rather than requiring a tag.
- Queuelet merge prerequisite resolved; independent all-view fidelity review remains item 6.

## Resume

Checkout `/root/Software/book-notes`, branch `main`. Item 5 has no background run or
server to resume; scratch is `/tmp/booknotes-sol-diag5/`. Node 22.23.2:
`cd pages && npm ci && npm run check`. Catalog and dataset slices now exercise real
production data without overrides; only third-entry and error tests intercept loading.
No worktree cleanup is required for this explicitly authorized main-checkout run.

## Item 5 — CI/release rehearsal complete

- Explicit [brief](evidence/item-5/brief.md) authorizes main-checkout commits,
  push and a real Pages deployment, superseding item 5's older no-deploy wording.
- No workflow existed on main. Added `.github/workflows/pages.yml`: full-history
  checkout, Node 22.23.2, `npm ci`, Chromium plus OS dependencies, `npm run check`,
  identical 404 fallback, publication manifest validation, upload `pages/dist`.
  PRs build only; main pushes/manual main runs deploy to `github-pages`.
  Only deploy has Pages/OIDC write permissions. Existing Pages configuration uses
  workflow builds; environment permits main. Current upstream action majors checked
  via GitHub release metadata: checkout/setup-node 7, configure-pages 6,
  upload-pages-artifact/deploy-pages 5. No dependency versions changed.
- Fresh clone of origin/main `c9696eea92af8591573386d2661f89b623388763` at
  `/tmp/booknotes-sol-diag5/rehearsal`, with only the candidate Queuelet test fix
  and manifest script overlaid. Node 22.23.2: install, Chromium install with OS
  dependencies and complete check exit 0: 66 unit/schema/source tests, production
  build, 34 Chromium tests. Initial attempt used system Node 18 and stopped before
  testing; rerun explicitly selected the pinned nvm toolchain.
- Clean clone exposed the unpushed local `explore-it/v1-explored` tag assumption.
  Reconciliation now verifies the immutable pinned commit object directly; the
  exact SHA assertion, merged-project diff and all selector/line checks remain.
  No book tags were published or changed. Initial failing check is retained.
- Manifest passes: Litefleet seven and Queuelet five views; default Litefleet;
  both JSONs, catalog, generated JS/CSS/worker and identical index/404 HTML only.
  HTML and worker references resolve under `/book-notes/`. Browser checks exercise
  that base, keyboard, narrow layouts and dataset routes. No live-site content
  verification performed; item 8 remains responsible after review.
- [Check](evidence/item-5/check.txt), [install](evidence/item-5/install.txt),
  [manifest](evidence/item-5/manifest.txt), [initial missing-tag failure](evidence/item-5/initial-missing-tag-check.txt).
  Deployment/rehearsal commands and rollback are in `pages/README.md`.
- Implementation pushed: `5cef3cfead2c8c1c0463196860f72cb94950cef0`.
  Real [Actions run 39887098435](https://github.com/nickstrad/book-notes/actions/runs/39887098435)
  **SUCCESS**; watched with `gh run watch 39887098435 --exit-status` (exit 0).
  Hosted fresh checkout passed install, Chromium, full check, fallback, manifest,
  artifact upload, configure-pages and deploy-pages. Deploy completed in 14 seconds.
  [Run/job/step result](evidence/item-5/actions-run.json) and
  [log excerpt](evidence/item-5/actions-summary.txt) retain evidence. No site URLs
  were fetched for live content verification. Items 6–8 remain outstanding.
- Knowledge/script assessment: reusable `pages/scripts/verify-dist.mjs` is part of
  the workflow/runbook. Indexed [release guidance](../knowledge/diagram-site-release.md)
  is saved in a separate knowledge commit. Final evidence/knowledge commits are
  docs-only and use `[skip ci]` to avoid rebuilding identical site output.
  Scratch `/tmp/booknotes-sol-diag5/`; no local server or background run remains.

## Item 7 — review fix pass (Opus)

Main checkout per explicit [brief](evidence/item-7/brief.md); commit and push authorized.
One commit per finding, base `7f5d316`:

| Finding | Disposition | Commit |
|---|---|---|
| H1 license notices | **Fixed.** `pages/license-notices.ts` (Vite plugin) emits `licenses/c4-sdk-LICENSE.txt` and `licenses/c4-sdk-THIRD_PARTY_NOTICES.txt` verbatim from the installed package, plus `licenses/NOTICES.txt` (index + React, React DOM, scheduler 0.28.0 MIT texts, the only other bundled packages). Build fails if any other npm package enters the bundle. Header "Licenses" link; `verify-dist.mjs` requires the three files; README corrected; browser test fetches them. | `2d0437a` |
| M1 stale Queuelet wording | **Fixed** with the review's sentence; inventory.md past tense. Also the same stale "Provisional … precede publication" note on all 67 Queuelet sources. | `9e6c058`, `12ed879` |
| M2 external tag / boundary | **Fixed, wording route (no model change).** Container views get their own grey key entry; the shared key no longer calls in-system containers partner fakes; both `doc-containers` (plain + STE) explain grey and the boundary. Queuelet worker/qctl/hooksink stay parented to the system: they are Queuelet-built clients and a lab fake, so re-parenting would misstate the model, and the service box is an authored group of what runs as the service (API + PostgreSQL). Re-parenting would also change counts, audits and tests. SDK request recorded under Open decisions. | `6108988` |
| M3 Pages path filter | **Fixed.** `paths: [pages/**, .github/workflows/pages.yml]` on push and pull_request; `workflow_dispatch` kept; README updated. | `a756e93` |
| M4 ts-db / LICENSE | **Not this item** (brief). Recorded as Nick's open decision above; README's stale "resolve before item 5" sentence now points there. | item 7 record |
| L1 dead exports/CSS | **Fixed.** Dropped `ArchitectureViewer` and unused type re-exports, `.dataset-note-planned` and its colour variable. | `2b694f8` |
| L2 snapshot labels | **Fixed without publishing tags.** Revisions read "9367863 (local tag release-it/v1-hardened)" / "b6df7db (local tag explore-it/v1-explored)"; the shell links the full pinned commit derived from the bundle's own source URLs (`snapshotOf`, unit + browser tested). Pushing tags remains Nick's call. | `e471924` |
| L3 heading structure | **Fixed.** Site title renders as `h1` when no diagram is shown; exactly one `h1` in every state (browser tested). | `2124211` |
| L4 CI time | **Left, by evidence.** Run 39887098435's 6m18s was apt (`Fetched 25.9 MB in 5min 57s (72.7 kB/s)`, a slow mirror); browser downloads took ~6 s, so a browser cache saves almost nothing and `--with-deps` stays for a reliable fresh runner. Action SHA pinning (optional) not done. | — |
| L5 deployment Fit readability | **Left.** Borderline; plan permits a split only if genuinely unreadable; native scale + physical tree carry the text. | — |
| L6 tmpfs overlay | **Fixed.** One clause in Litefleet `doc-lab-deployment`, checked against `deploy/compose.tmpfs.yml` at `9367863`. | `415200f` |
| L7 test special case | **Fixed.** Inventory carries `mergedRevision` and `mergedLine` (src-67 → 68); test reads them (verified it fails if the data is wrong). Stale `publicationStatus` updated. | `d7f006f` |

Gate (Node 22.23.2, after all fixes): `npm ci` exit 0; `npm run check` exit 0: TypeScript,
**66** Vitest, production build, **35** Chromium tests (34 + Licenses). Then
`cp dist/index.html dist/404.html && node scripts/verify-dist.mjs` exit 0; dist contains
`licenses/NOTICES.txt`, `licenses/c4-sdk-LICENSE.txt` and
`licenses/c4-sdk-THIRD_PARTY_NOTICES.txt`, the SDK files byte-identical (`cmp`) to the
package's. Inspected 1600 and 390 captures: header link, snapshot note and keys render
without overflow. Evidence: [check](evidence/item-7/check.txt),
[install](evidence/item-7/install.txt), [manifest](evidence/item-7/manifest.txt).
Scratch `/tmp/booknotes-opus-diag7/`; no server or background run remains.
