# Diagram site state

Goal: static C4 catalog for the book projects, per [PLAN.md](PLAN.md).
Nick approved destination B (`pages/` in book-notes, `/book-notes/`) and item 1
through `/root/booknotes-sol-diag1.md`; this supersedes the proposal's pending gate.

Status: item 1 complete and verified; ready for Twin verification/push. Items 2–8 unstarted by this run.

## Done and in progress

- Approved plan: `e147656`.
- Shell build: `45795e7` on `main`: published SDK, JSON-only
  invented catalog, props-only shell, hooks/pure functions, routing and test scaffolding.
- Source revisions rechecked: ts-db `fac3e3e5a971dcc81e62f454bb10a7643968b91e`;
  c4-sdk `bb12c7b3eb4a70152ed5ee13094da4bf7fa985be`.
- No book datasets or book content changed. No remote or deployment changes made.

## Next steps

1. Twin verifies item 1 and pushes; subsequent fresh runs own items 2 and 3.
2. Continue the plan's integration, CI, independent Fable review, Opus fixes and publication.

## Open decisions and blockers

- Coordinator: resolve ts-db shell reuse terms before release; source checkout has
  no license file. Attribution retained in `pages/README.md`.
- Explore It! coordinator: provide reviewed merged snapshot for queuelet acceptance.
- Independent review remains a later item under the explicit brief; this run owns
  only item 1, not review dispatch.

## Resume

Checkout `/root/Software/book-notes`, branch `main`. No delegated agents or persistent
servers. Build/check evidence lives at `/tmp/booknotes-sol-diag1/` during this run.
From `pages/`, with Node 22.23.2: `npm ci && npm run check`.

## Acceptance and deviations

- Node 22.23.2: `cd pages && npm ci` exit 0; `npm run check` exit 0:
  TypeScript passes, 50 Vitest tests, production build at `/book-notes/`, 12 Chromium
  tests. All shell bundles decode without diagnostics. Includes default selection,
  dropdown both ways, Back/Forward, deep-link reload, third-entry catalog, invalid
  routes, bad/empty/unavailable catalog, keyboard focus and worker base paths.
- Inspected desktop 1600×1000, narrow 390×844 and narrow explanation-drawer captures:
  readable controls/content, no page overflow, no diagram/legend overlap. Navigation
  breadcrumbs scroll horizontally on narrow screens, as in the reused shell.
- [Machine result](evidence/item-1/result.json), [check output](evidence/item-1/check.txt),
  [install output](evidence/item-1/install.txt), [brief](evidence/item-1/brief.md), and
  screenshots are retained in `evidence/item-1/`.
- No functional plan deviation. Workflow override: the explicit brief requests commits
  on `main`, so this fresh item run used the main checkout rather than a new worktree.
- Knowledge assessment complete: [diagram-site-shell.md](../knowledge/diagram-site-shell.md)
  records catalog bounds/identity, route subscriptions and mobile document selection,
  separately committed and indexed. Existing tests provide reusable checks; no extra
  work script is needed.
