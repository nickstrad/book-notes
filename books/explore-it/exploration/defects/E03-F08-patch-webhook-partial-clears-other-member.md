# E03-F08 — PATCH with a partial `webhook` object clears the omitted member: secret-only disables notifications, URL-only drops the signing secret

- Seen at: `explore-it/lesson-02` plain export; stack `queuelet-e03`; hooksink mode ok; fixture seed 04:16:24Z.
- Original observation: [E03 notes](../experiments/03-vary-it/notes.md) E03-H10 (stated 04:16Z, observed 04:21:40Z–04:22:03Z).
  Raw: `project/results/e03/p7-patch-webhook.txt`, `p8-misc.txt`. Prediction source: `internal/api/queues.go:72-75, 85-92`.

## Minimal reproduction
1. `POST /v1/queues -d '{"name":"e03-hook","webhook":{"url":"http://hooksink:8201/hook","secret":"report-secret"}}'`.
2. `PATCH /v1/queues/e03-hook -d '{"webhook":{"secret":"rotated"}}'`; enqueue and cancel a job; inspect `webhook_outbox`.
3. Restore, then `PATCH … -d '{"webhook":{"url":"http://hooksink:8201/hook"}}'`; cancel a job; read hooksink `/admin/records`.

## Expected (oracle)
Approximation, because SPEC does not define merge depth: SPEC:33-34 "PATCH … updates … webhook settings", and the same
endpoint treats an omitted top-level member as "keep" (observed: `{"max_attempts":4}` keeps the webhook). RFC 7396 (JSON
Merge Patch) applies the same rule recursively: omitted members are unchanged. SPEC (notification destinations section) says
only an *empty URL* disables future notifications — not an omitted one. Expected: secret-only PATCH keeps the URL;
URL-only PATCH keeps the secret.

## Observed
```text
PATCH {"webhook":{"secret":"rotated"}}  → 200 "webhook_secret":"rotated","webhook_url":""
  job 847 cancelled → 0 outbox rows (notification silently not produced)
  control: url+secret restored, job 848 cancelled → outbox row delivered
PATCH {"webhook":{"url":"…/hook2"}}     → 200 "webhook_secret":"","webhook_url":"…/hook2"
  (url-only on …/hook) job 849 cancelled → hooksink record status 200, "signature_valid":false; outbox secret '' delivered t
PATCH {"webhook":{}} → both cleared;  PATCH {"webhook":null} → both unchanged
```

## Variables isolated
- Only the nested object is replaced wholesale; top-level members merge (COALESCE, `queues.go:93`).
- `null` vs `{}` differ: null keeps, empty object clears both.
- Partner view: queuelet records the unsigned notification as delivered (2xx), so neither side surfaces the lost secret.

## Sibling check
- Schedule PATCH (`internal/api/schedules.go`) uses pointer fields for cron/timezone/enabled (no nested object) — not probed
  (#offcharter). Queue create has no merge question. No other nested objects in PATCH bodies found.

## Severity / status
Medium: a routine secret rotation silently turns off all notifications for the queue, and a URL migration silently stops
signing them; both return 200 with the changed values echoed only in the response body. Status: open (needs a spec line on
nested merge semantics).
