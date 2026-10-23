# C4 datasets: keep fidelity evidence separate from rendering proof

Learned in diagram-site item 2, the Litefleet dataset at source snapshot
`936786399d7bab3ebbfa845b9a07df02b41ed363`. See
[the inventory](../../pages/content/litefleet/inventory.md) and
[the slice tests](../../pages/tests/litefleet.test.ts).

Model the committed implementation, then use plans and lesson notes to explain
its development. Litefleet's shipped behavior differs from the original plan:
control lifecycle calls bypass Toxiproxy, heartbeat activity populates the registry
ledger separately from partner delivery, and draining responses are not retried.
A durable overflow outbox is not proof that in-memory events survive a hard crash.
Record these distinctions before turning behavior into an arrow or guarantee.

Keep a machine-readable inventory for every element and relationship: stable ID,
boundary, protocol, source revision/path/symbol or config selector, exact line,
and containing views. Link supporting lessons separately. Test complete ID coverage
and verify selectors against pinned git objects. Browser tests prove navigation
and rendering, not architectural truth. Source-object checks need the historical
commit locally; a shallow CI checkout must fetch it or preserve history before
running the current Litefleet checks.

The published c4-sdk 0.1.0 requires the `external` tag for supporting containers in
component diagrams. That tag is global to the element: document that it means
outside the focused process in this dataset. Parent IDs and authored system/process
boundaries carry ownership. Deployment instances refer to logical containers;
volumes and lab tools can be infrastructure nodes without inventing more product
services. Dynamic scenarios have no choice node: label conditional/alternative
steps explicitly and explain skipped steps in the document.

Large SDK views deliberately open on the first node when fitting everything would
make labels too small. For Playwright, drill-down and Up are links; focus a canvas
link before clicking so the SDK reveals the offscreen target. Save both readable
native-scale captures and Fit overviews. Capture the canvas separately for dense
deployment views; full-page pictures shrink it beside the long physical inventory.
Reset page scroll before full-page captures to avoid sticky-navigation artifacts.

A narrow explanation drawer stays open across hash route changes and hides the
page's controls from the accessibility tree. Close it between screenshot-loop
iterations. A full SHA in the shell's snapshot label can cause horizontal overflow:
use tag + short SHA for display and full SHA in provenance. Keep mobile assertions
for page width and inspect the actual drawer, not just the hidden inline copy.

Existing slice tests are reusable verification tools; no separate work script was
needed. Dataset generation scratch files do not belong in the repo unless a future
workflow gives them an ongoing maintenance role.

When authoring against an unmerged book, freeze the architecture tag and record
later orientation documents separately. Queuelet item 3 pins `v1-explored` at
`b6df7db`; WALKTHROUGH exists only at later committed HEAD `e750542`, whose project
tree has no delta. Do not assign a later document a source URL at an older tag.
Use committed-object reads or an archive in `/tmp` while another agent reviews
the worktree; never inspect its uncommitted files as accepted source.

Keep local schema/browser acceptance distinct from final source/publication
acceptance. Record the reviewed merged SHA and reconcile the project delta before
publishing. Local `git show` checks establish source existence, not public URL
availability. Queuelet's inventory carries the pending review/merge and URL-check
gates explicitly. Its historical exploration models retain old faults; final code
and the lesson fix sections determine the architecture drawn.
