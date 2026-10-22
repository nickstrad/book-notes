# Experiment 08: surround-it

Source: Ch. 9.1 "Diagramming the Ecosystem", 9.2 "Trust Boundaries", 9.3 "What If?", 9.4
  "Follow the Data".

*Explore queuelet's neighbours (Postgres, the workers, the webhook partner, the
  clock, the second namespace) with a context diagram, the hooksink admin modes, `docker compose
  restart` and the test clock, to discover what crosses a boundary unchecked and what happens when
  a neighbour misbehaves.* Prediction (from reading the auth middleware and the loop): "every
  route is namespace-scoped and the loop survives a database restart". Refuted per neighbour.
  60 minutes.

Probe family: draw `ecosystem.md` with every boundary marked; for each boundary, what crosses it and
  where in the code it is checked (`#read`); for each neighbour, one what-if (restart Postgres;
  hooksink in `500`, `slow`, `timeout`; the other tenant's key on this tenant's job ids; the clock
  across 2027-03-14 02:00 in `America/New_York`); follow one webhook body from the outbox row to
  hooksink's record and verify the signature by hand.

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
