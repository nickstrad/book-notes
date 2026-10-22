# Experiment 01: map-it

Source: Ch. 11.1 "Starting with a Recon Session", 11.4 "Selecting Techniques"; Ch. 2.1–2.4,
  2.6 (the template, good charters, generating and planning charters); 12.4 "Active Reading"
  applied to `SPEC.md`.

*Explore queuelet end to end with the source tree, `qctl`, the README and `SPEC.md`
  to discover what it does, what it promises, where the spec is silent, and what I predict the code
  does there.* Prediction: "the README's happy path works and the spec is complete". Refuted if a
  documented verb misbehaves or a spec sentence has no code behind it. 60 minutes, split roughly
  20 reading / 30 probing / 10 writing.

Probe family: walk every `qctl` verb once against the seed while reading the handler it lands in;
  list nouns and verbs; read the spec sentence by sentence asking "what does this not say?" (12.4)
  and, for each silence, read the code to predict the answer; rank the open questions by what would
  hurt most if wrong (the one line kept from 2.5) and tag each with the probe that fits (11.4).

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
