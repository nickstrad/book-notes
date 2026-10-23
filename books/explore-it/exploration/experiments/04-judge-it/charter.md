# Experiment 04: judge-it

Source: Ch. 5.1 "Never and Always", 5.2 "Alternative Resources", 5.3 "Approximations"; 10.3
  (status codes vs bodies).

*Explore retries, ordering, stats and error responses with a written list of
  Never/Always rules for a job queue, the spec, and SQS/pg-boss documentation as comparable
  systems, to discover where queuelet breaks rules nobody wrote down.* Prediction (from reading
  `nack.go`, `stats`, the lease query): "the code honours every rule on my list". Refuted by a
  single violated rule. 60 minutes, with `qctl clock` to walk backoff windows.

Probe family: write the rules **before** touching the system (`models/rules.md`: "a job is never
  leased by two workers at once", "attempts never exceed max_attempts", "within a priority, always
  FIFO", "stats always equal a direct count", "a retry always comes within a bounded delay", "an
  error is always a 4xx/5xx", "a cancelled job never runs" …), then try to break each: by reading the
  code for the line that enforces it, and by running the probe that would violate it. For numbers
  nobody checks (stats), compare with a direct `psql` count (5.3). For ordering, compare with what
  SQS FIFO promises (5.2).

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
