# Experiment 02: make-it-visible

Source: Ch. 3.1 (the moonwalking bear), 3.2 "Digging Deeper", 3.3 "Testability and Making the
  Invisible Visible", 3.4 "Consoles and Logs".

*Explore the job lifecycle with the api and worker logs, `psql`, the source of the
  background loop and the `crash` and `fail` worker kinds to discover what happens to a job that the
  API does not show.* Prediction (from reading `expire.go` and `ship.go`): "a crashed worker's job
  retries up to `max_attempts` and a webhook delivery is marked delivered only on 2xx". Refuted if
  the consoles disagree with the API or with each other. 60 minutes.

Probe family: run the `crash` kind and watch three consoles at once (api log, `attempts` table, the
  worker's stdout); run the webhook with hooksink in `500` mode and compare the api log with
  hooksink's record; write down every question the consoles could not answer, as a **testability
  request** (3.3): the experiment's first product.

Budget: at most 60 minutes wall clock. Record elapsed reading/probing/writing.
