# Experiment 09: explain-it

Source: Ch. 13.3 "Discovering Systemic Sources of Problems"; 8.4 (the table as code); 13.5 in
  one line (the stop signal: findings per experiment).

*Explore the defect reports and models so far, and every `UPDATE` and firing
  query in `internal/jobs` and `internal/schedule`, to discover the small number of habits that
  produced the 26 findings and to predict, before probing, where the same habits put a defect
  nobody has found.* Prediction: the explorer writes `models/causes.md` (findings grouped by cause,
  with the code habit each group reveals) and a list of **predicted defects** with `file:line` and
  the probe that would confirm each, then spends the rest of the budget running those probes.
  Refuted if the predictions find nothing. 60 minutes.

Probe family: group the defects (expected: eight missing guards on an `UPDATE` or a query, five
  silent acceptances, three clock/time mistakes, two consoles that lie, the rest singletons); list
  every state-changing statement in the two packages with the guard it carries; predict the gaps;
  drive them.

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
