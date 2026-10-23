# Astra end-review evidence

Review: [../../../REVIEW.md](../../../REVIEW.md). Source HEAD:
`e750542457e8b42fce28108714d72352ff19ae71`.

`run.py` ran the complete HEAD integration suite, uncached unit/application tests,
vet, and historical lesson08/09 verifies. Its initial suite omitted the required
seed: 93 functions passed and U09 failed on the missing `emails` fixture. That
failed run is retained. `seeded-run.py` passed all 94 checks after `make seed`
on a separate stack, with its own `seeded-acceptance.json` ledger. `v0-run.py` ran v0's baseline and HEAD's
D18/D21 public instruments against v0 binaries. The v0 probe log's `commit` field
identifies the **instrument checkout**; `v0-up.log` identifies the built v0 source.
The environments name separate images, Compose projects and ports. The only fixed
shared name is the verifier's historical `queuelet-ref`; historical runs are serial.

To reconstruct the layout before rerunning either script:

```sh
mkdir -p /tmp/explore-it-astra-review/logs
git clone --no-hardlinks <repository> /tmp/explore-it-astra-review/repo
git -C /tmp/explore-it-astra-review/repo checkout --detach e750542457e8b42fce28108714d72352ff19ae71
git -C /tmp/explore-it-astra-review/repo worktree add --detach /tmp/explore-it-astra-review/v0 explore-it/v0-inherited
```

Use a new evidence directory or preserve the previous logs before rerunning; the
runners use fixed paths. Check port/project availability first. All three runners tear
down their own stacks in `finally`; the REF verifier handles its temporary stack.
Each ledger retains command, source checkout, timestamps, duration, exit and log.
`SHA256SUMS` covers the review evidence files, excluding itself.

`integrity.json` records an independent SHA-256 comparison of each experiment's
`artifacts_copied_without_rewriting` manifest with that experiment commit. Entries
under `exploration/experiments/`, and original `F*`/`E*` reports under defects, were
also compared with HEAD. Mutable models/canonical status records were not required
to remain byte-identical to their first experiment.

`transcript-audit.json` records suspicious command/path inputs extracted directly
from all nine JSONL explorer transcripts, with transcript line numbers. It is a
candidate list for human interpretation, not an OS-isolation proof. The scan covered
`/root/`, bare-root `find`, runtime `/tmp/claude` paths, PLAN/register references and
git history/archive commands; original audits and exporter/charter content were read
separately. R2/R5 and the review's exception table give the judgments.
