# Experiment 03: vary-it

Source: Ch. 4.1–4.4; A2.1 (Some/None/All, Zero/One/Many, Beginning/Middle/End, Goldilocks,
  Boundaries, Count, Position, Selection, Strings, Numbers, Paths); 10.1 "Exploring an API" (does it
  reject what it does not understand?).

*Explore enqueue, batch enqueue and queue configuration with the variables list,
  the heuristics cheat sheet and raw `curl` requests to discover inputs the system accepts but cannot
  honour.* Prediction (from reading the validators in `internal/api`): "every documented range is
  enforced and unknown fields are rejected". Refuted by any 2xx the spec says should be a 4xx.
  60 minutes.

Probe family: list the variables first (`models/variables.md`: not just fields: counts, sizes,
  positions, timing, order, config, and the **shape of the request itself**: unknown fields, unknown
  enum values, wrong types), apply the named heuristics to each, and record a variation table
  (variable × value × observed). Use `curl`, because `qctl` only sends what the API expects.

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
