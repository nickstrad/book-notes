# Book-notes diagram site — implementation proposal

Date: 2026-10-22. Planning seat: Astra. Scope: a static C4 catalog for litefleet and
queuelet, reusing the ts-db site. This document is the deliverable; no site, remote,
repository publication or deployment is authorized by this planning run.

## Current state and handoff

**Status:** survey complete; proposal awaits Nick's site-home decision and plan confirmation.
Goal: pick either book's system from a dropdown, then navigate its architecture and
read source-backed explanations. This plan is the implementation spec.

Planning checkout: `/root/Software/book-notes`, branch `main`, surveyed HEAD
`7092e39`. Initial `git remote -v` returned no remotes. At the final checkpoint,
origin had been added externally as `git@github.com:nickstrad/book-notes.git`;
`gh repo view nickstrad/book-notes --json visibility,url` reported PUBLIC. This
planning run did not create the remote or push. This plan-only commit is discoverable
with `git log -1 -- docs/diagram-site/PLAN.md`; no implementation is in progress and
there is no background run to resume. Per the brief, this checkpoint is inside the
plan: root and book STATE files are read-only for this task.

Next, in order: Nick chooses the home below and confirms the plan; coordinator records
that choice, checks usage pools and dispatches item 1; Sol executes items 1–5; Fable
reviews; Opus implements the review suggestions; Sol deploys and verifies item 8.
The coordinator maintains the destination repo's root STATE.md at each boundary,
including exact source revisions, commits, current runs, blockers and resume commands.

### Survey evidence and precedence

| Source | Surveyed revision / evidence | Consequence |
|---|---|---|
| `book-notes/STATE.md`, `books/release-it/STATE.md`, `PLAN.md`, `WALKTHROUGH.md` | main `7092e39`; hardened tag resolves to `936786399d7bab3ebbfa845b9a07df02b41ed363` | Use hardened code including forward review fixes, not the original plan's proposed implementation. Book review/history decisions remain Nick's existing business. |
| `/root/Software/booknotes-work/explore-it`, `books/explore-it/STATE.md`, `PLAN.md`, lesson index and code | committed snapshot `b6df7db044782dd3517d041a8aad95308a3a80c9`, branch `plan/explore-it` | STATE now reports build complete, checkpoint C reviewed, final review next, unmerged. This supersedes the brief's older “C wrap in flight” status. Only committed objects were used; leave that worktree untouched. |
| `/root/Software/ts-db/pages/`, `.github/workflows/pages.yml` | HEAD `fac3e3e5a971dcc81e62f454bb10a7643968b91e` | Working catalog, SDK adapter, hash routing, tests, inventories, Vite build and Pages workflow already exist. |
| `/root/Software/c4-sdk/STATE.md`, `package.json`, domain model, README | HEAD `bb12c7b3eb4a70152ed5ee13094da4bf7fa985be`; STATE records npm 0.1.0 published | Use the published dependency and public API. No SDK fork or renderer work. |

Recheck source SHAs at dispatch. Queuelet datasets may be drafted from the frozen
snapshot above, but final fidelity acceptance and publication wait for its final
review/fix pass and merge to book-notes main. Record the accepted merged revision;
review the delta rather than silently following a moving branch. Do not merge or
modify Explore It! as part of this site job. Owner of that prerequisite: its coordinator.

## 1. Site home — Nick's decision

**Updated recommendation: use the now-public `nickstrad/book-notes` repository (B),
with `pages/` in this checkout and URL `https://nickstrad.github.io/book-notes/`.**
This removes a new repository and the source-link workaround. Nick still owns the
hosting choice; the observed external publication is not this agent's authorization
to deploy. Verify which source commits have actually been pushed before linking them.

The original recommendation, based on the initial no-remote state, remains the concrete
alternative A: a new public repository `nickstrad/book-notes-diagrams`, local checkout
`/root/Software/book-notes-diagrams`, with the app under `pages/`. Expected Pages URL:
`https://nickstrad.github.io/book-notes-diagrams/`. A would isolate future site publishing, but cannot undo content already exposed by the
now-public book-notes remote. Use a fresh
Git history, not a clone or filtered push of book-notes. Only diagram JSON, original
short explanations, provenance inventories, site code, tests and operating docs go
into this repo. The architecture descriptions themselves become public.

| Option | Exposure and operational cost | Judgment |
|---|---|---|
| A. Standalone public diagram repo | Publishes site source, diagrams and curated explanations; no book text, lesson files, registers, project code or book-notes history. Separate source snapshots require provenance upkeep. | Alternative if Nick wants separate site ownership; cannot restore existing source privacy. |
| B. Publish book-notes and use `pages/` there | Publishes **all tracked book content and pushed history**, not just Pages output. Source links become straightforward and datasets live beside their source. Publishing only `pages/dist` does not make the rest of a public repo private. | Updated recommendation because the repo is already public; verify intended pushed scope before adding more content. |
| C. Private source repo with Pages, or a subdirectory of an existing public site | Private-repo Pages availability depends on account plan; the generated diagrams are still intended for public consumption. Reusing ts-db's deployment couples unrelated sites and requires changing its catalog/paths. | Adds account or ownership complexity; no advantage for this small standalone site. |

GitHub documents project URLs and private-repo plan eligibility in
[What is GitHub Pages?](https://docs.github.com/en/pages/getting-started-with-github-pages/what-is-github-pages).
No private-repo account entitlement or alternative repository-name availability was verified.
The existing book-notes remote visibility was checked at the final checkpoint.

Nick owns choosing A/B/C and authorizing any new creation or additional public exposure. A includes the
proposed repo name and curated publication boundary above; no separate naming decision
is needed unless he wants a different name. If B is chosen, the same `pages/` layout
works in book-notes with base `/book-notes/`. If C is chosen, Nick must identify the
existing destination or private repository before dispatch. No custom domain is needed.

### Source references without publishing the books

The SDK Source shape requires a URL, with optional `revision`, `path`, and `notes`.
For recommended B, use direct commit-pinned links to code that exists on the public
remote, and verify those URLs; retain inventories for fidelity auditing. For A, use public commit-pinned links to the new site's
`pages/content/<dataset>/inventory.md` with stable source anchors. Each entry records
the original book-notes SHA, repository-relative path and relevant symbol/line range,
and a short original description. Say explicitly that original source is local and
not publicly browsable. Do not fabricate GitHub book-notes URLs or expose `file://`
links. Do not copy source files or book passages just to make links work. Keep original
code revision in `revision` and original relative path in `path`; `notes` explains
that the public URL is the provenance record, not the code itself.

Before release, commit inventories, then pin dataset Source URLs to that inventory
commit. Validation checks anchor existence and pin format. CI validates self-contained
JSON and inventory coverage; Sol and Fable validate original paths against local
pinned git objects. If B is selected, direct commit-pinned code links can replace the
inventory indirection. Never copy ts-db's source URL assertions unchanged.

## 2. Reuse and new work

Copy these files from the surveyed ts-db revision, preserving attribution and any
applicable license notices; check source licensing before public release:

- `pages/package.json`, lockfile, TypeScript configs, `vite.config.ts`, `index.html`,
  `src/main.tsx`, `src/app.css`: retain tool versions initially, rename package/title,
  change base path. Use the lockfile with `npm ci`, not a broad dependency upgrade.
- `src/sdk.ts`: retain the single SDK boundary, `ArchitectureProvider`,
  `DiagramNavigation`, `DiagramView`, `DiagramDocumentation`, bounded fetch loader,
  catalog identity checks, CSS import and `layout-worker.js?url` worker packaging.
- `src/routes.ts`, `src/catalog.ts`, and their unit tests: preserve hash grammar,
  catalog `defaultId` lifting, invalid-route behavior and bounded reading.
- `src/App.tsx`: reuse the catalog dropdown and navigation shell. Extract loading,
  hash subscription and selection logic to hooks/pure functions; views receive props.
  Replace ts-db text/links and its incorrect-for-us “exist on main” status sentence
  with “snapshot <tag / SHA>”. Dataset-specific content stays in JSON, not JSX.
- `tests/fixtures.ts`, `bundles.test.ts`, `sources.test.ts`, `views.test.ts`,
  `smoke.spec.ts`, fixture-catalog pattern, `playwright.config.ts`: reuse mechanisms,
  replace project-specific IDs, expected views, source rules and copy. Do not keep
  ts-db's requirement to exhibit all seven diagram types.
- `.github/workflows/pages.yml`: retain `pages/` working directory, npm caching,
  build/artifact/deploy structure, environment URL and non-cancelling Pages concurrency.
  Add browser validation before deployment; adapt action versions to currently
  supported releases (the old file uses upload-pages-artifact v3; current docs use v4).

Depend on exact **`c4-sdk: 0.1.0` from npm**, with its React 19.3.0 peers. Do not vendor
`src/lib/c4`, use local symlinks, or depend on unpublished SDK changes. Preserve
transitive notices. The existing app already proves the relevant integration pattern.
New work is two datasets, their provenance/explanations, a small branding/configuration
adaptation, targeted assertions and the destination's operational documentation.

## 3. Datasets and diagram inventory

Catalog has exactly two initial entries, default `litefleet`:
`Litefleet — Release It! (v1-hardened)` and `Queuelet — Explore It! (v1-explored)`.
Each JSON bundle is `pages/public/architectures/<id>.json`; catalog is
`pages/public/catalog.json`. Each default diagram is `context`. One architecture is
visible at a time. Preserve SDK drill-down, Up/Back, documentation and deep links:
`#/architectures/litefleet/diagrams/context` (and queuelet equivalent).

Use C4 context (L1), container (L2), component (L3), dynamic and deployment views.
No landscape or code/L4 view: these add little for two independently selected systems.
No nine/twelve-way version selector in v1. Explain the lesson progression in documents
attached to the final architecture. A component is an in-process responsibility;
a deployment instance is a running copy. Never turn every Go package into a service.

### Litefleet: seven views

| Stable diagram ID / kind | Content and source under `books/release-it/` |
|---|---|
| `context` / context L1 | Tenant application/developer and operator; litefleet boundary; blob-store and metering partners, clearly marked as lab fakes. Sources: PLAN §2, WALKTHROUGH, `project/README.md`, `cmd/*`. |
| `containers` / container L2 | HAProxy ingress, gateway, control, PostgreSQL registry, node agent, tenant SQLite files and node-local outbox storage; backup and metering dependencies. Collapse gateway/node replicas here. Sources: `project/deploy/compose.yml`, `cmd/{gateway,control,node}`, registry and metering code. |
| `gateway-components` / component L3 | API handling, route client/cache/breaker, node client, retry and shed boundaries; streaming responses and deadlines. Sources: `internal/{api,routes,breaker,nodeclient,retry,shed,httpx}`. |
| `node-components` / component L3 | Admission/engine, handle management, tenant files, backup job/blob client, bounded usage queue/durable outbox, maintenance and drain. Sources: `cmd/node`, `internal/{engine,backup,blob,metering}`. |
| `query-flow` / dynamic | Tenant → HAProxy → gateway; cached resolve vs control lookup; node execution and streamed result; bounded failure. Sources: handlers, route cache and node client; lessons 02–08. No invented retry of an uncertain write. |
| `usage-flow` / dynamic | Node usage production, queue/outbox and delivery to metering; shutdown spill and pressure behavior. Sources: `internal/metering/*`, `cmd/node`, lesson 09. |
| `lab-deployment` / deployment | Actual Compose lab: three gateways, two nodes and distinct volumes, control/Postgres, HAProxy, Toxiproxy links, fake partners, Prometheus/exporter and profile-only migration/maintenance/lab tools. Source: compose, HAProxy/Prometheus/Toxiproxy configs and Makefile. Label lab-only tools and optional profiles. |

Inventory includes a 12-row lesson-to-diagram/source table: L01 metrics/health,
L02 deadlines, L03 resource bounds, L04 retry policy, L05 cache/breaker, L06 isolation,
L07 shedding, L08 streaming, L09 outbox, L10 steady-state maintenance, L11 external
fault harness, L12 migrations/drain/tolerant readers. This does not imply twelve
separate services. Use lesson notes plus hardened code to resolve plan deviations:
for example the gateway refuses during node restart; the planned retry of `draining`
is not implemented. Do not import ts-db's proposed moves, replication or fencing.

### Queuelet: five views

| Stable diagram ID / kind | Content and source under `books/explore-it/` |
|---|---|
| `context` / context L1 | Producer, worker/client and operator roles; queuelet API system; webhook receiver as external collaborator. Sources: `project/SPEC.md`, lesson 00, `exploration/models/{map,ecosystem}.md`. |
| `containers` / container L2 | Go API process, PostgreSQL, worker client, qctl CLI, webhook receiver (hooksink fake in lab). HTTP/JSON, SQL and webhook relationships explicitly named. Sources: `project/cmd/*`, `internal/client`, compose and SPEC. |
| `api-components` / component L3 | HTTP/auth handlers, jobs/transition logic, schedule firing, lifecycle sweep, webhook shipper, store, clock and event/stat observation. Scheduler and shipper are goroutines inside API, not separate containers. Sources: `internal/{api,jobs,schedule,webhook,store,clock,stats}` and `cmd/api`. |
| `job-flow` / dynamic | Enqueue → lease → execute → ack/nack, expiry/retry/cancel branches and webhook delivery, with causal steps checked against code. Cite `jobs/{ack,nack,expire,transition}.go`, worker, webhook and state/rule models. Job-state table belongs in documentation, not a pretend C4 state-chart type. |
| `lab-deployment` / deployment | API, Postgres persistent volume, worker-a/emails, worker-b/reports, hooksink; qctl/probe clients outside Compose services. Flag controllable clock as lab configuration. Source: committed compose and command entrypoints. |

The inventory maps all nine experiments to affected views and evidence: E01 map,
E02 observation/clock, E03 input variation, E04 invariants, E05 sequences, E06 entities
and boundaries, E07 states, E08 ecosystem, E09 structural transition rules. Read each
lesson's findings and the final models. These are exploration history, not nine
architectural stages; the inherited defects were largely fixed in E01. Do not claim
exactly-once execution or complete exploration. Retain the final STATE's unresolved
hypotheses and review limits as provenance qualifications where relevant.

### Fidelity acceptance

Sol owns the initial line-by-line inventory against pinned source objects: every
modeled element/relationship has an ID, source path + symbol, revision, protocol,
boundary and supporting lesson/model where applicable. All must occur in a view.
Every view has a key and short plain-language documentation; retain SDK-required
STE metadata as draft, with no claim of certified language compliance. Use existing
statuses for implemented architecture; explain historical faults in documents.

Fable independently checks **all twelve views**, topology, process boundaries, edge
directions, dynamic step ordering and critical behavioral claims against the code.
Opus fixes discrepancies and updates inventories/tests. Browser tests prove rendering
and navigation, not fidelity; passing them cannot replace this source review.

## 4. Execution items, ownership and acceptance

Sizes are focused agent-work estimates, excluding queue/usage waits: S ≈ 1–2 hours,
M ≈ 2–4 hours, L ≈ 4–6 hours. Each item is a fresh run in its own worktree with a
file brief, Owns/Reads claims, permission scope, acceptance commands, output/exit
records and checkpoint. No builder is dispatched by this planning task. Nick's cast
for this effort overrides default seat assignments: **Sol builds all items; Fable
performs the final independent review; Opus implements Fable's suggestions.**

| Item / owner / size | Deliverable and dependencies | Acceptance |
|---|---|---|
| 1. Destination and working shell / Sol / M | After Nick's decision, initialize the chosen local repo; copy the selected shell/config/tests and attribution; record source SHAs and plan pointer in root STATE. Owns scaffold, lockfile, routing/hooks, shared tests. Use a valid minimal test dataset until real bundles land, never ts-db content branded as book-notes. | `cd pages && npm ci && npm run check`; Playwright default rendering, dropdown and deep-link reload at the selected base; no ts-db branding. Props-only views, JSON-only architecture data, published SDK dependency. No remote creation before Nick's decision. |
| 2. Litefleet slice / Sol / L | After 1, seven views, inventory and L01–12 mapping. Owns litefleet JSON/content and `tests/litefleet.*`; reads shared shell. | SDK decoding has zero diagnostics; complete ID/source coverage; all seven routes render with correct headings; real context→container→component navigation, dynamic step controls, docs and Back; desktop/narrow screenshots inspected. |
| 3. Queuelet slice / Sol / L | After 1, five views, E01–09 mapping and source inventory. Owns queuelet JSON/content and `tests/queuelet.*`. Draft may precede merge; acceptance waits for reviewed merged snapshot and delta reconciliation. | Same schema/source gates; all five routes render; component and job-flow drill/step tests; inspected desktop/narrow screenshots; no scheduler/shipper drawn as separate services. |
| 4. Catalog integration and publication inputs / Sol / M | After 2+3, owns catalog, shared tests, header copy, provenance URL pins and README. Integrate slices serially, freeze inventory commit before final URL pinning. | `npm run check`; exactly two real dropdown entries, default litefleet; switching both directions, Back/Forward, reload, unknown IDs, failed/malformed fetch, third-entry fixture proving no hard-coded pair; source anchors and base-relative assets/worker valid. |
| 5. CI and release rehearsal / Sol / M | After 4; workflow, deployment instructions, source/publication manifest and evidence. Browser tests precede artifact upload. No deploy yet. | Clean-checkout `npm ci && npm run check`; inspect built artifact contains only intended site assets; test `/book-notes-diagrams/` (or chosen base), keyboard navigation and narrow layout; screenshots reviewed for overlap/clipping/docs. Workflow dry review includes permissions, artifact path, environment and fallback. |
| 6. Final independent review / Fable / M | Fresh clone of candidate; read full diff, inventories and original pinned source. Owns review report only. | PASS/NITS/FAIL with file pointers and evidence; independently run site checks and inspect screenshots; verify all views' fidelity and publication boundary. No need to rerun book chaos suites for a static visualization. |
| 7. Review suggestions / Opus / M allowance | Implement Fable's actionable suggestions, one logical change per commit, record each finding's disposition. Update site STATE and review response. | Affected checks plus final `npm run check`; fresh screenshots for changed visuals; source recheck for changed claims. Final review recheck by a fresh Fable run; unresolved FAIL is not shippable. Escalate standoff under global protocol, preserving Nick's human gates. |
| 8. Publish and live verify / Sol / S | After review closure, accepted queuelet snapshot and Nick's public-repo decision. Configure approved remote and Pages; deploy accepted commit. | Successful Actions run; live root, catalog, both JSONs, worker and all 12 deep links; real dropdown navigation and refresh; no console/page errors or failed asset requests. Record live URL, deployed SHA, run URL and screenshots in STATE. |

Items 2 and 3 can run concurrently after 1 with disjoint paths; the coordinator alone
changes catalog and shared files during integration. Serialize browser suites on this
shared VM. Item 5's workflow draft can overlap dataset authoring if it owns only
`.github/` and deployment docs, but release rehearsal waits for 4. Review, repair and
deploy are sequential. No overlap with edits or experiments in the Explore It! worktree.
Before dispatch, check all three subscription gauges; do not bypass limits with API keys.
Use remaining pools only within Nick's cast; if the required seat is unavailable, record
that blocker rather than silently replacing it.

Retain run evidence with item checkpoints; reusable discoveries go in the destination's
indexed `docs/knowledge/`, reusable scripts in indexed `work-scripts/`, separately
committed after assessing the finished work. Clean finished worktrees after preserving
commits and evidence. Estimates total roughly 20–32 focused hours including review/fixes;
parallel datasets shorten elapsed time, not the fidelity checks.

## 5. Deployment runbook

These are future implementation instructions, not actions performed by this plan.

1. After Nick chooses A, check `nickstrad/book-notes-diagrams` availability. Create that
   public repo with the accepted local content only, set origin and push main. If the
   name exists, inspect ownership/purpose and report a conflict; never overwrite it.
   For B, use the existing origin; verify the approved publication scope and source revisions
   before pushing additional content. Do not recreate the existing repository.
2. Set build and preview base to `/book-notes-diagrams/` (or the chosen repo slug).
   Preserve `import.meta.env.BASE_URL` for catalog loading and Vite's worker URL import.
   Use hash navigation so deep links do not need server-side routing.
3. In repository Settings → Pages select **GitHub Actions**. Configure the workflow
   with `contents: read`, `pages: write`, `id-token: write`, environment `github-pages`
   and deployment output URL. Use configure-pages, upload-pages-artifact and deploy-pages
   following [GitHub's custom workflow documentation](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages).
   Preserve the known `pages/` build structure; use current supported action majors.
4. Install dependencies and Chromium in CI, run typecheck/unit/schema tests, build,
   then Playwright against the built base-path preview before uploading `pages/dist`.
   Retain ts-db's `dist/index.html` → `dist/404.html` fallback. Test pull requests without
   deployment; deploy main pushes or workflow_dispatch. Upload only dist, never the repo.
5. Watch the Actions run to completion, obtain its emitted Pages URL, and perform
   item 8's live browser checks. An HTTP 200 alone is insufficient. Confirm asset and
   worker paths stay under the repository base and reload a non-default queuelet view.
   Verify public provenance links resolve to the recorded inventory commit/anchors.
6. Record result in STATE; if deployment fails, retain exact run/failure evidence and
   fix the configuration before reporting completion. Roll back by redeploying the
   last known-good site commit, without changing book tags.

Pages updates are **ungated under Nick's protocol** once the destination/publication
choice is authorized. Do not invent a second approval for each deploy. Public repository
creation and publishing book-notes itself remain Nick's explicit choices under §1.

## 6. Risks and open decisions

- **Nick: site home/public exposure and plan confirmation.** This is the only new
  product decision needed to start item 1. Updated recommendation: B, using the now-public existing remote; A remains fully specified.
- **Explore It! coordinator: final review and merge.** Current committed state is
  further along than the brief but still unmerged. Draft against a fixed SHA; reconcile
  final code before acceptance. Do not imply provisional diagrams describe main.
- **Source truth:** litefleet's plan and actual hardened behavior differ; queuelet's
  explored model has qualifications. Inventories and independent code review guard
  against converting intended behavior into factual arrows.
- **SDK/layout:** large deployment views can become crowded; retain focused views,
  inspect screenshots at 1600×1000 and 390×844, and split only a genuinely unreadable
  lab view with a recorded plan deviation. Existing SDK mobile documentation limitations
  are known from its STATE; verify our actual views and report any unresolved issue,
  rather than importing another project's waiver.
- **Public provenance:** A cannot offer public code links. The explicit inventory
  mechanism above is intentional; source publication is not an implicit prerequisite.
  Check copy permissions/attribution for reused ts-db shell before release; do not
  invent a license for source whose reuse terms are unclear.
- **Dependency/deploy drift:** the survey proves the local precedent, not a future
  install or Actions run. Item 1 tests the exact lockfile; item 5 verifies current Actions
  compatibility. Keep changes narrow and record version changes if needed.
- **Scope control:** no interactive lesson replay, running backend, book-content export,
  SDK rewrite, custom domain, or automatic extraction from Go is needed. Adding a third
  book should require JSON, inventory and tests, not another application component.

Planning validation: inspected the precursor shell, routes, SDK integration, tests and
workflow; read committed book states/plans and relevant topology/entrypoints. No runtime,
Playwright, live-site or full behavioral fidelity test was performed during this planning
pass. Those checks are acceptance work above, not claims about a site already built.
