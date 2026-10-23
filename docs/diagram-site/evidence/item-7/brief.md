You are Opus (Claude Opus 5.5), a fresh run on the book-notes diagrams site
on vm-1, working in the MAIN checkout /root/Software/book-notes (branch
main; commit and push when green — GitHub Pages updates are ungated).

TASK — plan item 7: implement Fable's review suggestions. Read
docs/diagram-site/REVIEW.md in full first, plus docs/diagram-site/STATE.md.
Fable's verdict came with 1 high, 4 medium, 7 low findings; nothing was
fixed in the review pass. Your job is the fix pass:

- One commit per finding (or per tightly-coupled group), each message
  naming the finding id (H1, M1, ...).
- H1 (license notices missing from dist while the README claims they are
  bundled) is the priority: publish the c4-sdk license and the third-party
  notices with the site and allow them in the dist manifest, so the claim
  and the artifact agree.
- M1: the Queuelet context explanation's stale "awaiting the book end
  review and merge" wording — update it to the resolved state.
- M2: the external-tag/boundary inconsistencies — make legend, document
  wording and parenting consistent across both datasets, following the
  review's recommended route; if the review offers two routes and one is
  clearly less invasive without lying about the model, take it and say why
  in the record.
- M3: add a sensible `paths` filter to the Pages workflow so book-content
  commits stop rebuilding/redeploying the site.
- Lows: fix what is fixable in scope (dead re-exports/CSS, snapshot labels,
  error-state h1, CI Playwright install time if a clean win exists, the
  test special case); record honest dispositions for anything you
  deliberately leave.
- M4 is NOT yours: the review closes the ts-db question on the facts and
  leaves a LICENSE decision for book-notes itself to Nick. Record it as
  open-for-Nick in STATE.md; do not add a license file.

GATES: re-run the full local gate after your fixes (unit/schema tests,
production build, Playwright catalog tests) — all green before pushing —
and verify the built dist now contains the license notices H1 requires.
Update docs/diagram-site/STATE.md with the fix record, commit and push.
End with a printed summary: per-finding disposition, gate results, push SHA.
