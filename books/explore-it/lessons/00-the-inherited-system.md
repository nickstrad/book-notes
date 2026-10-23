# 00 — The inherited system

Queuelet is a small job queue inherited by a platform team. Producers submit JSON jobs to named queues in a namespace; workers lease them and acknowledge success or failure. Jobs can wait, retry and finish; schedules submit jobs over time. Operators manage queues and jobs. Postgres holds state, and a webhook partner records terminal notifications. An HTTP API and `qctl` are the interfaces.

The project has source, a spec, startup instructions and tests. A passing happy path establishes a starting point; it cannot establish all the behavior you have not yet asked about. The learning artifact here is your evolving model, followed by fixes and checks for demonstrated defects.

## Start the local system

From the exported `books/explore-it/project/` directory:

```sh
make up
make seed
make verify LESSON=00
```

The baseline runner exercises enqueue → lease → ack → webhook and runs the unit and thin baseline checks. A successful run ends with `BASELINE HOLDS:` and its actual measurements. The initial live-stack run passed the unit suite and four integration checks: queue CRUD, enqueue/lease/ack/webhook, authentication, and priority with concurrent disjoint leasing. Its verdict was:

```text
BASELINE HOLDS: enqueue→lease→ack→webhook; unit and happy-path checks pass
```

This establishes those paths only. The baseline does not assert unprobed behavior. Use `project/README.md` for exact CLI syntax, credentials and seed manifest locations in this build. Reset fixtures with `make reset`; stop/remove the lab with `make down` when finished.

## Read and probe in an afternoon

Start with `project/README.md`, `project/SPEC.md` and the tests. List the promises each makes and the questions each leaves open. Follow the API entry point into handlers and persistence; trace a `qctl` command to its request and the corresponding effect. Read worker and partner-fake code as part of the system, and identify background actions alongside request handlers.

Before executing a path, write what you think it will do and why, with a source pointer. Then run it and observe responses, command exit status, logs, database state and partner records where available. Reading supplies useful predictions; the running system can still contradict them. Keep a small map of where verbs land, then a backlog of questions ranked by consequence and uncertainty. You need not finish reading every line before making a useful probe.

The initial exploration record is intentionally empty. Its models and hypotheses will grow from real observations. Features and hooks must be discovered in the exported build rather than assumed from a later lesson.

## The experiment in a page

Read [the protocol](../exploration/PROTOCOL.md). Every experiment has one target, instruments, information sought, “I predict,” “I would be wrong if,” and a 60-minute budget. The coordinator writes its charter; the explorer works within it. Read source freely inside the export and state predictions before probing.

Keep timestamped notes with exact commands and trimmed evidence, including dead ends. Follow every surprise to a minimal reproduction or explain why it is not a defect; name the oracle that makes the result wrong. Record variables isolated and look for sibling causes. Put adjacent questions in the backlog. When the system hides an answer, request the smallest useful observation hook.

Close with `revealed.md`: done/needs more/blocked, actual time split, observations, model delta and next hypotheses. End with the score: hypotheses stated/confirmed/refuted, code-reading predictions contradicted by execution, defects found and model cells filled. Inference is not observed coverage. At most two runs close an experiment; misses remain visible. A separate fixer turns reports into patches, reproducible checks and a lesson note. Experiment records retain their original observations.

## Explore before reading the answers

For a blind attempt, use a plain export of project, exploration, skills and lessons at the inherited checkpoint, without git history or the plan. The full export boundary is in the protocol. Start with [map-codebase-from-behavior](../skills/map-codebase-from-behavior/SKILL.md) and [charter-as-probe](../skills/charter-as-probe/SKILL.md). Try changing the chosen verb, its starting state, one input variable or the console you observe; write your prediction first.

Read lessons 01–09 after your own attempt. The answer key is section 4 of the plan, outside the blind export. This lesson deliberately supplies no register, anticipated findings or filled models.

## Where to read more

Hendrickson, [*Explore It!*](https://pragprog.com/titles/ehxta/explore-it/) (2013), 1.1 “Two Sides of Testing,” 1.2 “Essential Elements of Exploratory Testing,” and Chapter 11 “Explore an Existing System,” especially 11.1 and 11.4. Section numbering checked against the publisher’s contents extract. This lesson is a local introduction based on the plan’s mapping; detailed source-text review remains pending. The blind-export mechanism and explicit prediction/scoring record are repo adaptations.
