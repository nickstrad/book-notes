# Book-notes architecture site

Static React/Vite catalog for the [approved plan](../docs/diagram-site/PLAN.md).
The production dropdown contains **Litefleet — Release It!** (seven views, default)
and **Queuelet — Explore It!** (five views). Data and explanations live in JSON,
loaded through the SDK module boundary; adding a dataset requires no view code.
Queuelet remains pinned to `b6df7db`, reconciled against merged `7d26bb8`:
project code is identical; one supporting lesson heading moved by three lines.
See the [site state](../docs/diagram-site/STATE.md) for evidence and review status.

Use Node **22.23.2** and the committed lockfile:

```sh
cd pages
npm ci
npm run check
```

`check` runs TypeScript, Vitest schema/routing/loading checks, a production build,
and Chromium Playwright checks against `/book-notes/`. If Chromium is absent,
run `npx playwright install chromium`. `npm run dev` starts Vite at the development
root; `npm run build && npm run preview` serves the production `/book-notes/` base.
The [Pages workflow](../.github/workflows/pages.yml) uses the same `npm ci` and
`npm run check`, after `npx playwright install --with-deps chromium` on Linux.
Checkout must include full history (`fetch-depth: 0`): fidelity tests read pinned
source objects and the reconciled Explore It! merge.

## Deployment and rehearsal

Pull requests run the build and browser gate without deploying. Main pushes and
manual dispatches on main deploy only after that gate. Pages is configured with
GitHub Actions as its source. The build job has read-only repository access;
the deploy job alone has `pages: write` and `id-token: write`, targets the
`github-pages` environment and reports its deployment URL. Runs are serialized
per ref. Node is pinned to 22.23.2; npm uses the committed lockfile.

Rehearse from a fresh, full clone of origin/main, in `pages/`:

```sh
npm ci
npx playwright install --with-deps chromium
npm run check
cp dist/index.html dist/404.html
node scripts/verify-dist.mjs
```

Only `pages/dist` is uploaded. Its manifest permits HTML (index plus identical
404 fallback), catalog JSON, the two architecture JSON bundles, generated
JavaScript/CSS assets including the layout worker, and the three license notice
files under `licenses/`, which it requires. The manifest check rejects
extra files and checks both datasets' view counts and `/book-notes/` asset and
worker paths. No book source, inventories, test fixtures or test screenshots
are uploaded. Playwright checks the built preview at that same base before upload.

Watch a deployment with `gh run watch <run-id> --exit-status` and record its URL,
result and source SHA in the site STATE. On failure, inspect `gh run view <run-id>
--log-failed`, fix and push again. To roll back, rerun a known-good main deployment
with `gh run rerun <run-id>`; book tags do not move. Live content checks belong to
plan item 8, after independent review.

Example deep link:
`/book-notes/#/architectures/queuelet/diagrams/context`.
Architecture switches use the SDK route port, so Back/Forward and reload preserve
selection. Unknown IDs and malformed routes show errors rather than a default fallback.
An empty, invalid or unavailable catalog shows a loading error.

## Structure

- `public/catalog.json`: ordered entries and optional `defaultId`; URLs resolve relative to the catalog.
- `public/architectures/*.json`: all architecture data, explanations and snapshot revisions.
- `src/sdk.ts`: sole runtime SDK import boundary, bounded fetch loader, catalog identity check and worker URL.
- `src/catalog.ts`, `src/routes.ts`, `src/site-model.ts`: bounded catalog adapter and pure decisions.
- `src/useSite.ts`: loading, cancellation, hash subscription, selection and focus callbacks.
- `src/Site.tsx`: props-only shell and SDK composition; `src/App.tsx` connects the hook.
- `tests/`: adapted upstream mechanisms and catalog and dataset browser/schema/source checks.

Playwright uses port 4327 and one worker. Run browser suites serially on the shared VM.
Screenshots are written beneath `node_modules/.cache/playwright-results/`; inspect
both 1600×1000 and 390×844 captures when changing the shell. Build output is ignored
at `dist/`. Test fixtures stay outside `public/`.

## Reuse and attribution

Shell, styles, configuration, lockfile, SDK adapter, routing and test mechanisms were
adapted from Nick Strad's `nickstrad/ts-db`, revision
`fac3e3e5a971dcc81e62f454bb10a7643968b91e`, under `pages/`.
No license file or explicit shell reuse notice was present in that source checkout.
This records provenance without inventing a license grant; the coordinator must
resolve reuse terms before the public release check in plan item 5.

The renderer is the published **c4-sdk 0.1.0**, with React/React DOM **19.3.0**;
no SDK source is vendored. SDK reference checkout inspected at
`bb12c7b3eb4a70152ed5ee13094da4bf7fa985be`. Tool versions are inherited from the
precursor lockfile without a dependency upgrade.

The build publishes license notices with the site, linked as "Licenses" in the
header: `licenses/c4-sdk-LICENSE.txt` and `licenses/c4-sdk-THIRD_PARTY_NOTICES.txt`
are c4-sdk's own MIT license and notices for the 76 packages in its bundles and
worker (including elkjs, EPL-2.0 OR GPL-3.0-or-later), copied verbatim from the
installed package; `licenses/NOTICES.txt` indexes them and reproduces the React,
React DOM and scheduler licenses. `license-notices.ts` emits them and fails the
build if any other npm package enters the bundle without a notice.
