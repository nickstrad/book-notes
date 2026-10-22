# Diagram site: keep the catalog adapter at the SDK boundary

Learned in diagram-site plan item 1; implementation `45795e7`, source precedent
`ts-db` at `fac3e3e5a971dcc81e62f454bb10a7643968b91e`.

The published c4-sdk 0.1.0 catalog schema accepts only `entries`. The host site's
optional `defaultId` must be lifted before SDK validation. Keep the bounded stream
adapter in `pages/src/catalog.ts`: it buffers at most 256 KiB + 1 byte and cancels
the source, including an endless body or an understated Content-Length. Stripping
keys with an unbounded `response.json()` would remove that protection. Lift only
`defaultId`; unknown fields must still reach the SDK and fail validation.

Catalog entries and bundles must agree on `id`, or a URL can name an architecture
other than the one shown. `pages/src/sdk.ts` checks that identity after decoding.
The whole catalog fails if an entry cannot load; empty catalogs, invalid defaults,
malformed routes and unknown architectures must not silently fall back. Preserve
`tests/catalog.test.ts`, `tests/routes.test.ts` and `tests/views.test.ts` while adding
real datasets. The third-entry fixture verifies that no application code assumes a pair.

The provider owns route details (selected elements and dynamic steps). The site
only resolves which architecture owns a route. Subscribe through the SDK route port,
which reports pushState writes as well as browser navigation; hashchange alone misses
SDK navigation. Keep data and snapshot revisions in JSON, decisions in hooks/pure
functions, and rendering in props-only views. Avoid "exists on main" copy: a diagram
models a source snapshot.

On a narrow screen the SDK exposes an explanation drawer while retaining an inline
document copy. Scope assertions to the `Explanation` region when opening the drawer;
an unscoped text assertion can match both copies. Inspect its screenshot as well as
the default shell. Playwright clears its output directory on the next run, so copy
accepted captures to the item checkpoint after the final browser run.

No additional work script was needed: `cd pages && npm run check` already provides
reusable type, schema, route, build and browser checks. Use Node 22.23.2 and the
committed lockfile; item 1 deliberately retained the source shell's versions.
