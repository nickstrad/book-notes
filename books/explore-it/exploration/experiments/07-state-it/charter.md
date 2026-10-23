# Experiment 07: state-it

Source: Ch. 8.1 "Recognizing States and Events", 8.2 "Diagramming a State Model", 8.3
  "Exploring with State Models", 8.4 "Changing the Representation: State Tables".

*Explore the job lifecycle with a state table built from the spec, the event
  history and the `WHERE` clauses in `internal/jobs`, two workers and the test clock, to discover
  transitions the table has no arrow for that the code allows anyway.* Prediction: for every cell
  the explorer writes what the code will do, from reading the guard (or its absence), before driving
  the cell. Refuted per cell. 60 minutes.

Probe family: draw the diagram, then the **state table** (every state × every event) with two columns
  per cell: "should" (from the spec) and "predicted" (from the code); then drive every cell marked
  "should be rejected" or "undefined", starting with those where the two columns disagree. The
  deep ones need two workers and the clock: lease with A, expire the lease, lease with B, ack with
  A's token.

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
