# Book-notes architecture site

Static React/Vite shell for the [approved plan](../docs/diagram-site/PLAN.md), item 1.
The catalog currently contains two **invented shell tests**, not book architectures.
Litefleet and Queuelet datasets belong to separate items; replace these fixtures during
catalog integration. The third-entry fixture is test-only and is not copied to dist.

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
No deployment workflow is introduced by item 1.

Example deep link:
`/book-notes/#/architectures/shell-alternate/diagrams/context`.
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
- `tests/`: adapted upstream mechanisms and shell-specific browser/schema/source checks.

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
`bb12c7b3eb4a70152ed5ee13094da4bf7fa985be`. Its MIT license and third-party notices
remain in the installed npm package and bundled dependency notices. Tool versions
are inherited from the precursor lockfile without a dependency upgrade.
