# Experiment 06: relate-it

Source: Ch. 7.1 "Recognizing Entities, Attributes, and Dependencies", 7.2 "CRUD", 7.3
  "Following the Data".

*Explore queues, schedules, workers and jobs with an entity diagram drawn from the
  migrations and a CRUD grid to discover what happens to dependents when their owner changes or
  disappears, and where a payload, result or error changes on its way through the system.*
  Prediction (from reading the migrations: no foreign keys with `ON DELETE`): "deleting an owner
  leaves orphans". Refuted if the code guards or cascades. 60 minutes.

Probe family: draw `entities.md` from the schema; fill the CRUD grid (entity × operation) with the
  observed outcome for each cell; for each dependency, delete or change the owner and read the
  dependent through every path (API, `psql`, the webhook body, the sweeper's log).

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
