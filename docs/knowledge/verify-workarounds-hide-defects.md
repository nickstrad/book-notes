# A verify workaround can hide a defect

Learned after Release It!'s final review (`books/release-it/lessons/01-transparency.md`, "After the
review").

- Lesson 02's verify read a node request log line that, from lesson 08 on, was missing. The
  review-fix pass taught the script to bound the number from a histogram instead. The check passed
  and the cause, a request logger that ran after the handler and so was skipped by its abort panic,
  stayed in the platform until a later pass found it.
- When a signal the platform promises is missing, first ask whether the platform broke that
  promise. Change the script only when the platform's behaviour changed on purpose, and say which
  lesson changed it.
- Code that records an outcome (request logs, counters, spans) runs in a `defer`, so a panic or an
  early return cannot skip it. In Go, `recover` there and panic again with the same value, so
  `http.ErrAbortHandler` still cuts the connection.
