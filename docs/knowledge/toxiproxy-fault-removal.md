# Toxiproxy 2.12: removing a fault from a busy proxy

Learned building Release It! lessons 02 and 04 (`books/release-it/lessons/02-timeouts.md`,
`04-retries.md`, "Deviations").

- Removing a `latency` toxic, or a `timeout` toxic on the **downstream** stream, from a proxy with
  thousands of open links can wedge that proxy for 15 s or more, including for fresh connections.
  The toxic flushes held data to connections that are already gone.
- For "accepts and never answers", use the `timeout` toxic on the **upstream** stream: requests are
  swallowed before they reach the service, the client sees the same thing, and it removes cleanly.
- To remove any toxic under load, disable the proxy, delete the toxic, re-enable the proxy
  (`untoxic_busy` in `books/release-it/project/lab/lib.sh`). Every open link is reset once; traffic
  resumes on fresh connections.
- General rule: an injector has failure modes of its own. Check that removing the fault really
  removes it before trusting a recovery number.
