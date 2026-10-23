# Experiment 05: sequence-it

Source: Ch. 6.1 "Nouns and Verbs", 6.2 "Random Navigation".

*Explore operator and worker verbs in unusual orders with the nouns-and-verbs
  list, the test clock and a random sequence generator to discover sequences that leave jobs in
  states the spec does not allow.* Prediction (from reading the handlers): "every verb either does
  what it says from any starting point or is refused". Refuted by a sequence that leaves a job the
  spec would call impossible. 60 minutes. The experiment commit includes the little generator the
  explorer writes (`exploration/experiments/05-sequence-it/walk.sh`: pick a noun, pick a verb, act,
  observe, check the rules from E04).

Probe family: pairs and triples of verbs chosen for being unlikely (purge then enqueue-with-delay;
  pause, advance the clock an hour, resume; lease, purge, nack; cancel, retry); then the random walk
  for the rest of the budget, asserting `rules.md` after every step.

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
