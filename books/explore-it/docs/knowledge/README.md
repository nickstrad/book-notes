# Explore It! project knowledge

- [E06: data boundaries and lifetime checks](06-data-boundaries.md): replay client binaries
  from the target image, preserve exact numbers and whole-document JSON parsing, separate dependent lifetimes, and
  interpret owner-race sampling honestly. Runtime evidence: `register/runs/06-*`.

- [Isolated review database configuration](review-isolation.md): Compose port overrides
  and direct Go integration-check connections are separate; set the explicit database URL.
  Evidence: checkpoint B review and retained failed/corrected acceptance.

- [E08: injecting neighbour faults from outside the service](08-neighbour-faults.md):
  `docker kill` is an operator stop, a frozen database is not a stopped one, lib/pq ignores
  context deadlines against a frozen server, and probes must leave shared queues drained.
  Evidence: `register/runs/08-opus-*`.
