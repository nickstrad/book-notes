# Item 2 visual inspection — 2026-10-23

All seven views were inspected at 1600×1000 and 390×844. Files are in
`screenshots/`, named by stable diagram ID and viewport width:

- `<id>-<width>.png`: page at the top, with the SDK's default viewport.
- `<id>-<width>-canvas.png`: Fit View of the actual interactive canvas.
- `<id>-<width>-overview.png`: full page after Fit, including relationship text,
  key, sources, dynamic scenario text or the deployment inventory as applicable.
- `<id>-390-docs.png`: open mobile explanation drawer.

Context has a clear system/role/partner layout. Containers and gateway components
are wide; Fit shows topology, while native scale and keyboard focus/pan expose
readable labels. Node components have distinct tenant storage and node-local
outbox storage. Query and usage views provide both graph and numbered scenario
text. The lab deployment shows all replicas, distinct volumes and optional tools;
its 25 elements and 48 relationships need zoom/pan for detailed reading. Node boxes
and labels have separate layout space; the key stays below the canvas. No extra
view or renderer change was needed.

Narrow controls, wrapped snapshot label, headings and explanation drawers are
readable without page overflow. The picker truncates its selected label on narrow
screens; the full dataset title is directly below it. Breadcrumbs scroll sideways.
A phone user scrolls down to the canvas and uses the drawer for the explanation.
Fit overview labels on a phone are intentionally small; it is not a substitute for
native-scale inspection or the accessible relationship/scenario lists.

During verification, the full SHA in the snapshot label produced 430px page width
at a 390px viewport. The label now uses tag + short SHA; sources retain full SHAs.
Tests initially treated drill-down and Up as buttons; they are links, and keyboard
focus reveals offscreen canvas targets before clicking. The capture loop now closes
the modal drawer between routes and resets page scroll to avoid sticky-navigation
artifacts in full-page screenshots. Final checks pass. The shared shell and architecture claims were unchanged by
those test corrections.

This records the builder's inspection. Independent architecture/visual review
remains plan item 6.
