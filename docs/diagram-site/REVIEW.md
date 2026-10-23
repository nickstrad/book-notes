# Diagram site — independent review (plan item 6)

Reviewer: Claude Fable 5.1. Date: 2026-10-23. Reviewed revision: `ea1056c` on `main`
(main checkout `/root/Software/book-notes`, as the item brief directs; the plan's
"fresh clone" wording was superseded by the brief). Brief: `/root/booknotes-fable-diag-review.md`.
Scope: the site built in items 1–5 per [PLAN.md](PLAN.md) and [STATE.md](STATE.md).

## Verdict: PASS WITH FIXES

The site does what the plan asked: two JSON datasets, a props-only shell over the
published `c4-sdk@0.1.0`, hash routing, a checked Pages pipeline, and a live deployment
whose asset hashes match a local build. Every gate re-ran green. The source-fidelity
claims I sampled all held. The fixes below are about publication hygiene (third-party
notices), one stale public sentence, how the SDK's `external` tag reads in container
views, and CI cost. None changes the architecture drawn.

Finding counts: **1 high, 4 medium, 7 low**. Item 7 (Opus) implements; nothing was
fixed in this pass.

## 1. Gates re-run (this review, Node 22.23.2)

Commands: `cd pages && npm ci && npm run check` (output in `/tmp/fable-diag-review/`,
not committed; counts below are from that run).

| Gate | Result |
|---|---|
| `npm ci` | exit 0, 0 vulnerabilities |
| `tsc -b` | pass |
| Vitest unit/schema/source | 7 files, **66 passed** |
| `vite build --base /book-notes/` | pass; `index-DeMpY6Is.js` 800.9 kB, `layout-worker-BeIcR9gj.js` 2107.6 kB, CSS 33.3 kB |
| Playwright (Chromium, 1 worker) | **34 passed** in 1.9 min |
| `npm run check` overall | exit 0 |

These match STATE's claim (66 unit, 34 browser). The live site
`https://nickstrad.github.io/book-notes/` returns 200 for `/`, `catalog.json`,
`architectures/queuelet.json` and `404.html`, and its `index.html` references the same
`index-DeMpY6Is.js` / `index-ChqLlcAG.css` hashes my local build produced, so the
deployed artifact is the rehearsed build. Actions run 39887098435 is the only run and
concluded `success` (build 7m20s, deploy 14s). Pages is configured `build_type: workflow`,
`https_enforced: true`.

Working tree was clean before and after the run (`git status --short` empty).

## 2. Dataset fidelity (sampled against pinned sources)

Pinned objects verified locally: `936786399d7bab3ebbfa845b9a07df02b41ed363`
(`release-it/v1-hardened`) and `b6df7db044782dd3517d041a8aad95308a3a80c9`
(`explore-it/v1-explored`) are both commits and both ancestors of `main` and of
`origin/main` (`git merge-base --is-ancestor`), so CI's `fetch-depth: 0` checkout can
always reach them even if `plan/explore-it` is deleted. Origin has **0 tags**, as item 5
recorded. 8 of 8 randomly sampled commit-pinned source URLs (4 per dataset) returned
HTTP 200.

### Litefleet (7 views, 49 elements, 84 relationships, 102 sources)

Every numeric or behavioral claim I pulled from the explanations matched the pinned code:

| Claim in dataset | Source at `9367863` |
|---|---|
| 24-request shed limit; noncritical routes share 6 slots; health/metrics never shed | `cmd/gateway/main.go:39-43` (`SHED_LIMIT` default 24); `internal/shed/shed.go:8-9,43,57` (`limit/4`) |
| 2.9 s request deadline; `Retry-After` on `draining`; no draining retry | `internal/api/handlers.go:24,250-253`; `internal/nodeclient/client.go:86` (`transient` excludes `draining`) |
| Route cache 60 s TTL, refresh after 10 s | `internal/routes/cache.go:20-21` |
| Node client 2.5 s per attempt; bounded read retries | `internal/nodeclient/client.go:38-40` |
| Breaker guards control calls except the usage report | `internal/routes/client.go:27-28,143-147` |
| 2 reads / 1 writer per DB, 4 per tenant, 16 per node, 250 ms wait | `internal/engine/admission.go:35-40` |
| 2 s statement deadline; 1000 rows / 1 MiB cap with `truncated` | `internal/engine/exec.go:35,121-122` |
| Handle LRU 128; checkpoint above 64 KiB | `internal/engine/steady.go:22,24` |
| Queue 1024 slots, 20 ms enqueue budget, 4 shippers, batch 100, sweep every 5 s up to 500 rows | `internal/metering/queue.go:21-26` |
| SIGTERM spill waits 3 s | `cmd/node/drain.go:34` |
| Heartbeats every 5 s, direct to control; provision/drop direct to nodes | `cmd/node/main.go:80,226`; `deploy/compose.yml:72` (`NODES: node-a=http://node-a:8070`) |
| Registry pools 8 / 2 / 1 | `internal/registry/registry.go:49-57` |
| Metering fake deduplicates by event ID in memory | `cmd/metering/main.go:33,59-64` |
| Gateway↔control and gateway↔node via Toxiproxy; node↔partners via Toxiproxy | `deploy/compose.yml:103-106,124` |
| HAProxy health-checks `/health`; three gateway servers | `deploy/haproxy.cfg:29-43` |
| Prometheus scrape targets (gateway ×3, control, node ×2, haproxy, blobstore, metering, toxiproxy, postgres-exporter) | `deploy/prometheus.yml:8-35` |
| Purger enabled by `lab/up.sh` via `lab/lib.sh` | `lab/up.sh:8`, `lab/lib.sh:22` |

Edge directions and process boundaries checked: `heartbeat` node→control, `provision`
control→node, `n-sweep-meter` outbox→meter, no meter→registry edge (the ledger comes from
heartbeats: `registry.go` `recordUsage(ctx, tx, hb.Databases)`). The query-flow scenario's
cache-miss branch and reverse replies are consistent with `handlers.go` / `routes/client.go`.
The L01–L12 mapping names real pinned notes and paths (test-enforced).

### Queuelet (5 views, 27 elements, 49 relationships, 67 sources)

| Claim in dataset | Source at `b6df7db` |
|---|---|
| One-second core tick: Promote → Expire → Fire → Retain; separate shipper goroutine on its own ticker | `internal/api/api.go:223-251` |
| Shipper takes at most 20 due rows concurrently, 2 s HTTP timeout, retries after 5 semantic seconds, HMAC-SHA256 over the actual bytes | `internal/webhook/ship.go:18-22,40-48,65` |
| Terminal transition and notification snapshot share one transaction | `internal/jobs/ack.go` (`notify(tx, …)` before `tx.Commit()`); `jobs.go:92-108`; `nack.go` notifies only when `dead` |
| Nack → `retry_wait` or `dead` when refused/exhausted/purged | `internal/jobs/nack.go:19-21` |
| Expiry/worker release counts an attempt, returns to `queued` or `dead` | `internal/jobs/expire.go:31-71` |
| Cancel is terminal; only `dead` accepts operator retry, which resets attempts | `internal/jobs/transition.go:54-55`; `operator.go:49-50` (`attempts=0`) |
| Retention deletes succeeded/cancelled; dead letters persist | `transition.go:62` (`EvRetain: Succeeded, Cancelled → Gone`) |
| Auth hashes bearer keys; admin flag | `internal/api/api.go:162-165` |
| qctl's `worker run` uses the same worker implementation | `cmd/qctl/main.go:215-221` |
| Compose: api :8200, hooksink :8201, pg :5432 → host 18200/18201/15433; workers on emails/reports; restart policies (api/pg `unless-stopped`, workers `on-failure`, hooksink none); `QUEUELET_TEST_CLOCK=1`; `data` volume | `deploy/compose.yml` (whole file) |
| Seeded `reports` queue points at `hooksink:8201`; lab tools reset via `psql` exec | `lab/seed/seed.py:10`; `lab/reset.sh:8` |

Scheduler, sweeper, shipper, clock and observation are all inside the API process, as
the plan required (test-enforced: 8 components, all `parentId: api`). The job-flow
scenario's 16 steps are ordered consistently with the code paths above.

### Fidelity verdict

No incorrect arrow, boundary, step order or numeric claim found in the sample. The
inventories' "behavior checked against code" sections are accurate. The remaining
fidelity findings (M1, M2, L6) are about presentation and one stale sentence, not
about what the code does.

## 3. Architecture rules

- **Data decoupled from the app:** all dataset content, titles and snapshot labels are
  in `pages/public/catalog.json` and `pages/public/architectures/*.json`. Only the
  repository link and the site's own copy are in `Site.tsx`. ✔
- **Dumb views:** `Site.tsx` receives `{loaded, selection, choose, skipToMain}` and
  renders; no fetch, no routing decisions. ✔
- **Logic in hooks/pure functions:** `useSite.ts` (loading, abort, hash subscription),
  `site-model.ts`, `routes.ts`, `catalog.ts` are pure and unit-tested. ✔
- **SDK as a module dependency:** `c4-sdk: "0.1.0"` exact in `package.json`; single
  boundary in `src/sdk.ts`; no vendored SDK. ✔
- **Third-entry extensibility:** proven by `tests/fixtures/catalog-three.json` in both
  Vitest and Playwright. ✔

Violations: none. Nits are L1.

## 4. The source-license question (item 1 limitation)

STATE.md "Open decisions" line 183 and `pages/README.md:52-54` carry item 1's note:
the ts-db shell had no license file; "the coordinator must resolve reuse terms before
the public release check in plan item 5". Item 5 then deployed publicly with the
question still open.

Assessment, with evidence:

- `gh repo view nickstrad/ts-db` → `visibility: PUBLIC`, owner `nickstrad`,
  `licenseInfo: null`. `git log --format=%an -- pages` in ts-db → 13 commits, all
  `nickstrad`. The copied shell is Nick's own work; reusing it in Nick's own repo needs
  no third-party grant. **That part of the question is closed: no license issue.**
- What is actually unresolved is different: (a) `nickstrad/book-notes` itself has
  `license: null`, so visitors have no reuse rights beyond viewing; that is Nick's call
  (human gate), not a defect; (b) the deployed bundle contains c4-sdk and 76
  third-party packages (65 MIT, 9 ISC, 1 BSD-3, elkjs `EPL-2.0 OR GPL-3.0-or-later`)
  and ships **no notices at all** — see H1.

What should happen: Opus records the closure of the ts-db question in STATE/README
(same-owner reuse, attribution retained), ships the notices (H1), and STATE lists one
remaining human decision for Nick: whether to add a LICENSE to book-notes.

## 5. CI / deploy

`.github/workflows/pages.yml` is sound: top-level `contents: read`; only `deploy` has
`pages: write` + `id-token: write`; `github-pages` environment with URL output;
non-cancelling concurrency; PRs build only; `fetch-depth: 0` (required by the
`git show` fidelity tests and documented); browser tests and `verify-dist.mjs` run
before upload; `404.html` is byte-identical to `index.html` (checked live). The
rehearsal evidence (`evidence/item-5/actions-run.json`) is real and complete.

Fragile or costly points: M3 (no path filter), L4 (6-minute Playwright install, no
browser cache; action majors not SHA-pinned). Nothing blocks item 8.

## 6. Findings

### High

**H1 — Deployed site ships third-party code with no license notices; README overstates.**
`pages/dist/assets/*.js` after a fresh build contain 0 occurrences of `MIT License`,
`Copyright (c)` or `@license` (`grep -c` over both chunks = 0; Vite/esbuild strip legal
comments). `node_modules/c4-sdk/THIRD_PARTY_NOTICES.md` lists 76 packages bundled into
the SDK and its ELK worker, including `elkjs@0.12.0` (EPL-2.0 OR GPL-3.0-or-later), and
`node_modules/c4-sdk/LICENSE` (MIT) is not in `dist`. `pages/README.md:58-59` says the
notices "remain in the installed npm package and bundled dependency notices"; the second
half is not true of the published artifact. `scripts/verify-dist.mjs:6-9` would currently
reject a notices file. Fix: copy `c4-sdk/LICENSE` and `THIRD_PARTY_NOTICES.md` into
`pages/public/` at build time (or commit them), link them from the site header/footer
("Licenses"), extend the `verify-dist.mjs` allowlist to require them, and correct the
README sentence. Keep c4-sdk's own first-party MIT text intact.

### Medium

**M1 — Public Queuelet context explanation is stale.** `public/architectures/queuelet.json`
line 1899 (`doc-context.plain`) and its STE twin say the snapshot is "still awaiting the
book end review and merge". Item 4 recorded the merge at `7d26bb8` (record `1712ced`),
and `pages/README.md:7-8` and the inventory say so. Item 4's inspection note kept the
wording under its content boundary, but it is now the sentence every visitor reads
first. Fix: reword to "This is the explored and fixed b6df7db snapshot; its project
tree is identical to the merged end-review fix pass (7d26bb8). It is not a claim of
complete exploration or exactly-once execution." Also soften
`content/queuelet/inventory.md:485` ("awaits cross-family end review") to past tense
with the merge SHA.

**M2 — The SDK's global `external` tag mis-styles container views.** The published SDK
rejects a component view whose supporting containers lack the `external` tag
(validator message "Supporting systems and containers must be tagged external in a component view" in `c4-sdk/dist-sdk/shared-CLmsJL9r.cjs`: removing the tag yields 4 `invalid-member`
errors in litefleet and 3 in queuelet; verified by decoding modified copies). Because the
tag is per element, in `litefleet` `containers` the Control plane, Node agent, Tenant
SQLite files and Node-local outbox storage are drawn in the grey "external" tone inside
the Litefleet boundary, next to white Gateway/HAProxy/Registry, while the key says grey
means "Outside the focused process — a collaborator, store or partner fake" (there is no
focused process in a container view). In `queuelet` `containers`, PostgreSQL is grey
inside the boundary, and Worker client, qctl CLI and hooksink fake are rendered
**outside** the "Queuelet service boundary" box although the model says
`parentId: queuelet` for all three (screenshots
`queuelet-containers-1600-canvas.png`, `litefleet-containers-1600-canvas.png` from
this run). Readers cannot tell lab fakes from first-party containers by colour. Fix
(no SDK change): give the two container views their own key entry text, e.g.
"Grey: also a supporting element in a component view; still inside the system", and
reword the shared `external` key description so it no longer says "partner fake" where
it covers Control plane and Node agent; state in `doc-containers` (queuelet) that worker,
qctl and hooksink are drawn outside the service box because they are clients and a lab
fake, or re-parent them if that is the intended reading. Record an upstream request to
c4-sdk for a per-view `external` override. Decide one way and make both datasets
consistent.

**M3 — Every push to `main` rebuilds and redeploys the site.** `pages.yml` has no
`paths` filter, so book-only commits trigger a ~7.5-minute run and a redeploy of
identical output; items 4–5 worked around it with manual `[skip ci]` on docs commits
(`ba1acaf`, `ea1056c`), which is easy to forget and skips PR checks too. Fix: add
`paths: [pages/**, .github/workflows/pages.yml]` to both `push` and `pull_request`
(keep `workflow_dispatch`). Safe because the fidelity tests read pinned commits by SHA,
not working-tree book files.

**M4 — Close the ts-db reuse question in the records and surface the real decision.**
See §4. `pages/README.md:52-54` still instructs a resolution "before the public release
check in plan item 5", which has passed; STATE "Open decisions" still lists it. Fix: in
README and STATE, record that ts-db and book-notes share one owner (nickstrad, public
repo, sole author of `pages/`), keep the attribution paragraph, and replace the open
item with the one human decision that remains: whether `book-notes` gets a LICENSE
(currently `license: null`). Do not add a license without Nick.

### Low

**L1 — Dead exports and CSS.** `src/sdk.ts:30` re-exports `ArchitectureViewer` and
line 37 re-exports types `CatalogEntry`, `Diagram`, `Route` that nothing outside `sdk.ts`
imports; `src/app.css:180` `.dataset-note-planned` has no user. Fix: drop them (keeps the
"only what the site uses" boundary honest).

**L2 — Snapshot labels name tags that do not exist publicly.** The shell shows
"Snapshot release-it/v1-hardened · 9367863" and "explore-it/v1-explored · b6df7db"
(`bundle.revision`), but `git ls-remote --tags origin` returns 0 tags, so a visitor can
resolve only the SHA. Fix: either push the two immutable tags (Nick's call; item 5
chose not to) or render the SHA as a link to the commit and present the tag as
"local tag". Tests (`views.test.ts:23`, `smoke.spec.ts:30`) pin the current string.

**L3 — Heading structure in non-diagram states.** The site title is a `<p>`
(`Site.tsx:23`); the only `<h1>` comes from the SDK diagram. In loading, failed and
Not Found states the page has an `<h2>` with no `<h1>` (`Site.tsx:38,57`). Fix: use
`<h1>` for the error/Not Found panels, or make the site title an `<h1>` only when no
diagram is shown. Keyboard, skip link, labels, focus ring and contrast are fine.

**L4 — CI time and pinning.** `npx playwright install --with-deps chromium` took 6m18s
of the 7m20s build (`evidence/item-5/actions-run.json`). Fix: cache
`~/.cache/ms-playwright` keyed on the `@playwright/test` version and skip
`--with-deps` on cache hit. Optionally pin actions to commit SHAs.

**L5 — Deployment views are overviews only at Fit.** Litefleet `lab-deployment`
(25 elements, 48 edges) is unreadable at Fit (`litefleet-lab-deployment-1600-canvas.png`);
readable at native scale with pan, and the physical tree below the canvas carries the
text. The plan allows a split only if genuinely unreadable; this is borderline. Optional:
split the lab tools/profiles into a second deployment view with a recorded plan deviation.

**L6 — Node volumes are tmpfs under lesson 10's overlay.** `node-a-data`/`node-b-data`
are modeled as persistent "Docker named volume" (correct for `deploy/compose.yml:199-203`),
but `deploy/compose.tmpfs.yml` redefines them as 48 MB tmpfs for lesson 10's verify.
Fix: one clause in `doc-lab-deployment` ("lesson 10's verify overlays them with 48 MB tmpfs").

**L7 — Reconciliation special case lives in test code.** `tests/queuelet.test.ts:43`
hard-codes `src-67 → line 68` at the merge commit. Fix: carry an optional `mergedLine`
in `content/queuelet/inventory.json` and have the test read it, so the next
reconciliation edits data, not the test.

## 7. What was not done

- No fresh clone (brief chose the main checkout); no book chaos suites (plan: not needed).
- Live browser verification (console errors, worker request under the base path on the
  live host) remains item 8; I only checked HTTP status, sizes and asset hashes.
- Fidelity was sampled, not exhaustive: ~30 behavioral/numeric claims and the key edge
  directions across all twelve views, plus the test-enforced selector and coverage audits
  (102 + 67 selectors, 133 + 76 IDs) that I re-ran rather than re-derived.

## Recheck of the item 7 fix pass

Reviewer: Claude Fable 5.1. Date: 2026-10-23. Rechecked revision: `c4a023c` on `main`
(fix commits `2d0437a`..`d7f006f`, records `f542c5f`..`c4a023c`). Brief:
`/root/booknotes-fable-diag-recheck.md`. Method: each disposition in STATE's item 7
table was checked against the diff, the current files, a fresh production build, the
CI logs and the live site; nothing was fixed in this pass.

### Verdict: PASS WITH NITS

Every fixed finding is fixed as described; the two LEFT findings' reasons hold; M4's
open record is accurate. No disposition is disputed. Two record nits, no code change
needed (below).

### Gates re-run (Node 22.23.2, `cd pages && npm ci && npm run check`, then
`cp dist/index.html dist/404.html && node scripts/verify-dist.mjs`; output in
`/tmp/fable-diag-recheck/`, not committed)

| Gate | Result |
|---|---|
| `npm ci` | exit 0, 0 vulnerabilities |
| `tsc -b` | pass |
| Vitest unit/schema/source | 7 files, **66 passed** |
| `vite build --base /book-notes/` | pass; `index-B_L0nYzB.js` 801.5 kB, `index-DqMESnE-.css` 33.2 kB, `layout-worker-BeIcR9gj.js` 2107.6 kB |
| Playwright (Chromium, 1 worker) | **35 passed** in 2.1 min (34 + "the Licenses link serves the bundled packages' notices") |
| `npm run check` | exit 0 |
| `verify-dist.mjs` | PASS; 11 files, the three `licenses/` files required and present |

Working tree clean before and after. Live `https://nickstrad.github.io/book-notes/`
`index.html` references the same `index-B_L0nYzB.js` / `index-DqMESnE-.css` as this
build, so the deployment (run 39890114511, `f542c5f`) is the rechecked source.

### Per-finding verification

| Finding | Claimed | Verified |
|---|---|---|
| H1 | Fixed, `2d0437a` | **Yes.** `pages/license-notices.ts` emits the three files at `generateBundle` and errors on any bundled `node_modules` package outside c4-sdk/react/react-dom/scheduler (build passed, so the coverage claim is enforced, not asserted). Fresh `dist/licenses/c4-sdk-LICENSE.txt` (1067 B) and `c4-sdk-THIRD_PARTY_NOTICES.txt` (108203 B) are `cmp`-identical to `node_modules/c4-sdk/LICENSE` and `THIRD_PARTY_NOTICES.md`; the notices list 76 packages with `elkjs@0.12.0 — EPL-2.0 OR GPL-3.0-or-later` at line 697. `NOTICES.txt` indexes c4-sdk and carries three `## <pkg>@<version>` MIT sections. The live site serves all three with HTTP 200, byte-identical to the package files / local build. Header "Licenses" link present (`Site.tsx`), `verify-dist.mjs` requires the files, README's "remain in the installed npm package" sentence is gone, and the Playwright test exists and passed. |
| M1 | Fixed, `9e6c058` + `12ed879` | **Yes.** `doc-context.plain` carries the review's sentence; the inventory bullet is past tense with `1712ced`/`7d26bb8`; all 67 source notes now read "reconciled against the merged end-review fix pass 7d26bb8; public URL verified". `grep` for the old wording in the JSON, inventory.md and inventory.json returns 0. |
| M2 | Fixed by wording, no model change, `6108988` | **Yes.** Both container views have their own grey key entry (Litefleet: "Supporting element (grey)" distinguishing inside/outside the boundary; Queuelet: naming PostgreSQL inside the service box and worker/qctl/hooksink outside it). The shared entry on the other 10 views no longer says "partner fake" for in-system containers. Litefleet `doc-containers` (plain + STE) and Queuelet `doc-containers` (plain + STE) explain grey and the box. The model confirms the explanation: Queuelet's box is an authored group `containers-boundary` with members `api`, `postgres` only, while `worker`, `qctl`, `hooksink` keep `parentId: queuelet`; Litefleet's group holds all seven containers, with `control`, `node`, `tenant-files`, `node-state` tagged `external`. Consistent across datasets. SDK per-view override request is under Open decisions. |
| M3 | Fixed, `a756e93` | **Yes.** `paths: [pages/**, .github/workflows/pages.yml]` on both `push` and `pull_request`; `workflow_dispatch` kept; README explains. Consistent with the fidelity tests reading pinned SHAs. |
| M4 | Not a fix; Nick's decision | **Accurate.** `gh api repos/nickstrad/book-notes` → `license: null`, `visibility: public`. STATE's open decision records the same-owner closure of the ts-db question and the one remaining decision; README's "before item 5" sentence is replaced with a pointer to STATE. No license file was added. |
| L1 | Fixed, `2b694f8` | **Yes.** `ArchitectureViewer`, `CatalogEntry`, `Diagram`, `Route`, `RoutePort` re-exports and `.dataset-note-planned` / `--site-planned` removed; no remaining references in `src/`, `tests/` or `index.html`; typecheck and build pass. |
| L2 | Fixed without tags, `e471924` | **Yes.** Both bundle revisions read `<short SHA> (local tag <tag>)`; `snapshotOf` derives the full-SHA commit link from the bundle's own GitHub source URLs; unit test (`views.test.ts`) and browser test (`smoke.spec.ts`) pin label and href. Origin still has no tags, as intended. |
| L3 | Fixed, `2124211` | **Yes.** `Site.tsx` renders the title as `h1` when nothing is shown and `p` otherwise; browser tests assert exactly one `h1` in the diagram state and the title as the only `h1` in Not Found and failed-load states. |
| L4 | Left, by evidence | **Reason holds.** Run 39887098435 log: apt `Fetched 25.9 MB in 5min 57s (72.7 kB/s)` (02:00:12→02:06:09); browser download+install 02:06:09→02:06:21. Run 39890114511 (same step, same packages): `Fetched 25.9 MB in 1s (37.8 MB/s)`, whole step 21 s. A browser cache would not have helped; the slow run was the mirror. Action SHA pinning remains optional and undone, as recorded. |
| L5 | Left | **Reason holds.** PLAN.md line 274 permits a split "only [for] a genuinely unreadable lab view with a recorded plan deviation"; my review rated the view borderline and the split optional, and the view is readable at native scale with the physical tree. Leaving it is within the plan. |
| L6 | Fixed, `415200f` | **Yes.** The clause names `deploy/compose.tmpfs.yml` and 48 MB; at `9367863` that file sets `node-a-data`/`node-b-data` to `type: tmpfs … size=48m` for lesson 10. |
| L7 | Fixed, `d7f006f` | **Yes.** `inventory.json` carries `mergedRevision` and `mergedLine: 68` on `src-67`; the test reads both (`source.mergedLine ?? source.line`) and asserts the merge SHA; `publicationStatus` updated. The 66-test unit run includes it. |

### Regression spot-check

The fix commits touched both dataset JSONs (key text, three documents, 67 source
notes), `Site.tsx`, `site-model.ts`, `sdk.ts`, `app.css`, the workflow and one test
each for L2/L3/L7. All twelve routes, drill-down, Up/Back, documents, scenario controls,
catalog errors and the third-entry fixture passed in the 35-test browser run; view
counts (7 + 5), element/relationship audits and all 102 + 67 source selectors passed
in the 66-test unit run; both datasets still carry exactly 7 and 5 `external` key
entries. `catalog.json` was not touched. No regression found.

### Nits (record only; no repair round needed)

1. STATE's L4 row says browser downloads "took ~6 s" in run 39887098435; the log shows
   the download-and-install phase after apt was about 12 s (02:06:09→02:06:21). The
   conclusion (apt mirror, not browser download) is unaffected.
2. The item 7 record says the three `licenses/` URLs return 200 but not that they are
   byte-identical live; they are (checked here against the package and local build).
   Worth stating in item 8's live verification.
