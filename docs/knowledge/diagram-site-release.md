# Diagram-site release checks need source history and an artifact boundary

Learned during diagram-site item 5, 2026-10-23; implementation `5cef3cf`.

The dataset tests are source-fidelity checks, not just JSON validation. They run
`git show` against immutable source revisions and compare Queuelet's pinned tree
with its accepted merge. Keep `fetch-depth: 0` in the Pages workflow. A shallow
checkout can build the UI while lacking the objects required to validate it.

A full clone still need not contain local book tags. At item 5, origin had no
tags, and a reconciliation assertion on `explore-it/v1-explored` failed despite
the exact pinned commit being present. Assert the known immutable SHA and its
commit object; compare source trees and selectors at those SHAs. Do not publish
or move book tags just to make site CI pass. A warm working checkout hides this
kind of release dependency, so rehearse from a fresh clone of origin.

Run browser checks against the production preview before uploading the Pages
artifact. The Node version, install, complete `npm run check`, fallback copy and
`node scripts/verify-dist.mjs` are shared by the workflow and deployment runbook.
The manifest script verifies the two intended datasets and view counts, exact
catalog membership, generated asset/worker base paths and identical index/404
HTML. Its allowlist rejects extra publication files. Upload only `pages/dist`,
never the repository, source inventories, tests or captures.

Keep pull-request checks separate from the main-only deploy job. Only deployment
needs Pages/OIDC write access and the `github-pages` environment. A successful
pipeline proves build/upload/deploy integration; live content and independent
diagram fidelity remain separate acceptance items. See
[item 5 evidence](../diagram-site/STATE.md) and [runbook](../../pages/README.md).
