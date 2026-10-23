# Queuelet — explored snapshot fidelity inventory

Pinned source: `explore-it/v1-explored` → `b6df7db044782dd3517d041a8aad95308a3a80c9`. Machine audit: [inventory.json](inventory.json). Bundle: [queuelet.json](../../public/architectures/queuelet.json).

Item 4 reconciliation: the accepted Explore It! fix-pass record is `1712ced`, merged
on main at `7d26bb82db862b2fd483589f6b903d8dae26ae33`. The complete project tree is
identical to this pinned snapshot. All 67 selectors remain valid at the pinned SHA;
66 retain their merged line positions. `src-67` (`09-explain-it.md`, structural-fix
heading) moved from line 65 to 68 after citation edits; the section is unchanged.
Pinned links deliberately retain line 65 at `b6df7db`. No diagram rewrite or repin
is needed. See [reconciliation evidence](../../../docs/diagram-site/evidence/item-4/reconciliation.json).
The production catalog now includes Queuelet; independent all-view review remains item 6.

## Method and boundaries

Read committed sources only. The readonly Explore It! worktree HEAD was `e750542457e8b42fce28108714d72352ff19ae71` for the later WALKTHROUGH; its project tree has no delta from the pinned tag. All architecture selectors below refer to the tag commit, never a moving branch. The walkthrough does not exist at the tag, so it is orientation rather than pinned architecture evidence. Each selector records its exact source line and trimmed content in the machine inventory; tests compare it to `git show <revision>:<path>`. A matching selector is evidence to inspect, not an automated proof of the surrounding behavior. Independent all-view review is item 6.

A component means an in-process responsibility. All eight components belong to API. The scheduler is called from the lifecycle loop; the shipper is a separate API goroutine. The store represents a shared pool/helpers, not a mandatory wrapper for every SQL call. PostgreSQL remains inside the logical service boundary; external tags mean outside the focused API process. The external receiver contract and hooksink lab realization are distinguished, not two simultaneous deliveries. qctl and probes belong outside the Compose service list. Deployment host grouping includes local CLI tools; the child Compose group contains exactly five service instances and the data volume.

## View inventory

| View | Kind | Elements | Relationships |
|---|---|---:|---:|
| `context` | context | 5 | 4 |
| `containers` | container | 9 | 8 |
| `api-components` | component | 12 | 17 |
| `job-flow` | dynamic | 6 | 13 |
| `lab-deployment` | deployment | 9 | 8 |

## Exploration mapping

These are experiments that improved the model and code, not nine separate architectures. Most inherited defects were found and fixed in E01. Historical model tables retain previous observations; use final code and lesson fix sections to interpret them.

| Experiment | Affected views | Final behavior and evidence |
|---|---|---|
| E01 | `context`, `containers`, `api-components`, `job-flow` | [01-map-it.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/01-map-it.md): Recon, namespace/auth and early fixes (all 27 planted defects found in E01; D07 initially partial). Source paths: `books/explore-it/project/internal/api/api.go`, `books/explore-it/project/internal/jobs/jobs.go`. |
| E02 | `api-components`, `lab-deployment` | [02-make-it-visible.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/02-make-it-visible.md): Committed observation and semantic clock; independent webhook loop. Source paths: `books/explore-it/project/internal/api/observe.go`, `books/explore-it/project/internal/clock/clock.go`. |
| E03 | `api-components`, `job-flow` | [03-vary-it.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/03-vary-it.md): Strict input and JSON variation; bounds and bounded backoff. Source paths: `books/explore-it/project/internal/api/enqueue.go`, `books/explore-it/project/internal/jobs/jobs.go`. |
| E04 | `api-components`, `job-flow` | [04-judge-it.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/04-judge-it.md): State/expiry invariants, attempt history and default backoff cap. Source paths: `books/explore-it/project/internal/jobs/ack.go`, `books/explore-it/project/internal/jobs/expire.go`. |
| E05 | `job-flow`, `api-components` | [05-sequence-it.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/05-sequence-it.md): Unusual sequences, release counting, purge and terminal cleanup. Source paths: `books/explore-it/project/internal/jobs/expire.go`, `books/explore-it/project/internal/jobs/purge.go`. |
| E06 | `containers`, `job-flow` | [06-relate-it.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/06-relate-it.md): Entity lifetimes and exact numeric data across CLI, worker and webhook. Source paths: `books/explore-it/project/internal/client/client.go`, `books/explore-it/project/internal/store/store.go`. |
| E07 | `api-components`, `job-flow` | [07-state-it.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/07-state-it.md): State table, concurrent worker deletion/purge and immediate cancellation. Source paths: `books/explore-it/project/internal/jobs/operator.go`, `books/explore-it/project/internal/jobs/purge.go`. |
| E08 | `context`, `api-components`, `lab-deployment` | [08-surround-it.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/08-surround-it.md): Neighbour faults, fair concurrent webhook attempts and civil-time schedules. Source paths: `books/explore-it/project/internal/webhook/ship.go`, `books/explore-it/project/internal/schedule/fire.go`. |
| E09 | `api-components`, `job-flow` | [09-explain-it.md](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/09-explain-it.md): Structural transition guards, isolated schedule errors and transactional outbox. Source paths: `books/explore-it/project/internal/jobs/transition.go`, `books/explore-it/project/internal/jobs/jobs.go`. |

## ID audit

| ID / kind | Boundary | Protocol / technology | Views | Sources |
|---|---|---|---|---|
| `producer` / element | external role or system | human role / instance | context, containers, job-flow | [src-1](#src-1) |
| `worker-role` / element | external role or system | human role / instance | context, containers | [src-2](#src-2) |
| `operator` / element | external role or system | human role / instance | context, containers | [src-3](#src-3) |
| `queuelet` / element | external role or system | Go / PostgreSQL | context | [src-4](#src-4) |
| `receiver` / element | external role or system | HTTP | context, containers, api-components, job-flow | [src-5](#src-5) |
| `api` / element | queuelet | Go | containers, job-flow | [src-6](#src-6) |
| `postgres` / element | queuelet | PostgreSQL 17 | containers, api-components, job-flow | [src-7](#src-7) |
| `worker` / element | queuelet | Go | containers, api-components, job-flow | [src-2](#src-2) |
| `qctl` / element | queuelet | Go | containers, api-components, job-flow | [src-3](#src-3) |
| `handlers` / element | api | Go | api-components | [src-14](#src-14) |
| `jobs` / element | api | Go | api-components | [src-15](#src-15) |
| `scheduler` / element | api | Go | api-components | [src-16](#src-16) |
| `sweep` / element | api | Go | api-components | [src-17](#src-17) |
| `shipper` / element | api | Go | api-components | [src-18](#src-18) |
| `store` / element | api | Go | api-components | [src-7](#src-7) |
| `clock` / element | api | Go | api-components | [src-19](#src-19) |
| `observation` / element | api | Go | api-components | [src-20](#src-20) |
| `host` / element | external role or system | Local shell / Docker | lab-deployment | [src-48](#src-48) |
| `lab` / element | host | Docker Compose | lab-deployment | [src-49](#src-49) |
| `api-run` / element | lab | human role / instance | lab-deployment | [src-50](#src-50) |
| `pg-run` / element | lab | human role / instance | lab-deployment | [src-51](#src-51) |
| `worker-a` / element | lab | human role / instance | lab-deployment | [src-52](#src-52) |
| `worker-b` / element | lab | human role / instance | lab-deployment | [src-53](#src-53) |
| `hooksink` / element | queuelet | Go | containers | [src-5](#src-5) |
| `hook-run` / element | lab | human role / instance | lab-deployment | [src-54](#src-54) |
| `data` / element | lab | Docker named volume | lab-deployment | [src-55](#src-55) |
| `host-tools` / element | host | CLI / Python / curl | lab-deployment | [src-56](#src-56) |
| `ctx-enqueue` / relationship | client → system | HTTP/JSON + bearer key | context | [src-1](#src-1) |
| `ctx-work` / relationship | client → system | HTTP/JSON + lease token | context | [src-2](#src-2) |
| `ctx-operate` / relationship | operator → system | HTTP/JSON + bearer key | context | [src-3](#src-3) |
| `ctx-webhook` / relationship | system → receiver | HTTP POST / HMAC-SHA256 | context | [src-8](#src-8) |
| `producer-api` / relationship | client → API | HTTP/JSON + bearer key | containers | [src-9](#src-9) |
| `operator-cli` / relationship | operator → CLI | CLI arguments / JSON | containers | [src-10](#src-10) |
| `developer-worker` / relationship | role → client | process invocation | containers | [src-11](#src-11) |
| `cli-api` / relationship | CLI → API | HTTP/JSON + bearer key | containers | [src-12](#src-12) |
| `worker-api` / relationship | worker → API | HTTP/JSON + bearer key / lease token | containers | [src-2](#src-2) |
| `api-db` / relationship | API → database | PostgreSQL wire / SQL | containers | [src-13](#src-13) |
| `api-hook` / relationship | API → external receiver | HTTP/JSON + HMAC-SHA256 | containers, job-flow | [src-8](#src-8) |
| `c-worker` / relationship | client → API | HTTP/JSON | api-components | [src-31](#src-31) |
| `c-cli` / relationship | CLI → API | HTTP/JSON | api-components | [src-14](#src-14) |
| `c-auth` / relationship | API process | Go call | api-components | [src-32](#src-32) |
| `c-jobs` / relationship | API process | Go call | api-components | [src-33](#src-33) |
| `c-handler-clock` / relationship | API process | Go call | api-components | [src-34](#src-34) |
| `c-handler-observe` / relationship | API process | Go call | api-components | [src-29](#src-29) |
| `c-handler-schedule` / relationship | API process | Go call | api-components | [src-35](#src-35) |
| `c-sweep-clock` / relationship | API process | Go call | api-components | [src-36](#src-36) |
| `c-sweep-jobs` / relationship | API process | Go call | api-components | [src-37](#src-37) |
| `c-sweep-schedule` / relationship | API process | Go call | api-components | [src-38](#src-38) |
| `c-schedule-store` / relationship | API process | Go call | api-components | [src-39](#src-39) |
| `c-jobs-store` / relationship | API process | Go call | api-components | [src-26](#src-26) |
| `c-ship-store` / relationship | API process | Go call | api-components | [src-40](#src-40) |
| `c-ship-clock` / relationship | API process | Go call | api-components | [src-41](#src-41) |
| `c-ship-hook` / relationship | API → receiver | HTTP POST / HMAC-SHA256 | api-components | [src-42](#src-42) |
| `c-observe-store` / relationship | API process | Go call | api-components | [src-43](#src-43) |
| `c-store-db` / relationship | API → database | PostgreSQL wire / SQL | api-components | [src-13](#src-13) |
| `f-enqueue` / relationship | producer → API | HTTP/JSON | job-flow | [src-1](#src-1) |
| `f-insert` / relationship | API → database | SQL | job-flow | [src-44](#src-44) |
| `f-lease` / relationship | worker → API | HTTP/JSON | job-flow | [src-27](#src-27) |
| `f-claim` / relationship | API → database | SQL transaction | job-flow | [src-45](#src-45) |
| `f-heartbeat` / relationship | worker → API | HTTP/JSON + lease token | job-flow | [src-46](#src-46) |
| `f-ack` / relationship | worker → API | HTTP/JSON + lease token | job-flow | [src-21](#src-21) |
| `f-nack` / relationship | worker → API | HTTP/JSON + lease token | job-flow | [src-22](#src-22) |
| `f-expire` / relationship | API → database | SQL transaction | job-flow | [src-23](#src-23) |
| `f-operator` / relationship | CLI → API | HTTP/JSON + bearer key | job-flow | [src-24](#src-24) |
| `f-terminal` / relationship | API → database | SQL transaction | job-flow | [src-26](#src-26) |
| `f-promote` / relationship | API → database | SQL | job-flow | [src-47](#src-47) |
| `f-outbox` / relationship | API → database | SQL | job-flow | [src-18](#src-18) |
| `lab-api-db` / relationship | lab network | SQL / PostgreSQL | lab-deployment | [src-57](#src-57) |
| `lab-pg-data` / relationship | container → volume | filesystem mount | lab-deployment | [src-55](#src-55) |
| `lab-worker-a` / relationship | lab network | HTTP/JSON + bearer key | lab-deployment | [src-58](#src-58) |
| `lab-worker-b` / relationship | lab network | HTTP/JSON + bearer key | lab-deployment | [src-58](#src-58) |
| `lab-hook` / relationship | lab network | HTTP POST / HMAC-SHA256 | lab-deployment | [src-59](#src-59) |
| `lab-tools-api` / relationship | host → lab | HTTP/JSON | lab-deployment | [src-60](#src-60) |
| `lab-tools-hook` / relationship | host → lab fake | HTTP/JSON (lab admin) | lab-deployment | [src-61](#src-61) |
| `lab-tools-db` / relationship | host → database | psql / PostgreSQL | lab-deployment | [src-62](#src-62) |
| `receiver-fake` / relationship | API → lab fake | HTTP POST / HMAC-SHA256 | containers | [src-59](#src-59) |

## Pinned selectors

<a id="src-1"></a>
### src-1

[`books/explore-it/project/internal/api/enqueue.go:78`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/enqueue.go#L78) — selector `func (s *Server) enqueue`.

<a id="src-2"></a>
### src-2

[`books/explore-it/project/internal/worker/worker.go:26`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/worker/worker.go#L26) — selector `func Run(`.

<a id="src-3"></a>
### src-3

[`books/explore-it/project/cmd/qctl/main.go:24`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/cmd/qctl/main.go#L24) — selector `func run(`.

<a id="src-4"></a>
### src-4

[`books/explore-it/project/cmd/api/main.go:17`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/cmd/api/main.go#L17) — selector `func main()`.

<a id="src-5"></a>
### src-5

[`books/explore-it/project/cmd/hooksink/main.go:16`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/cmd/hooksink/main.go#L16) — selector `func main()`.

<a id="src-6"></a>
### src-6

[`books/explore-it/project/cmd/api/main.go:49`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/cmd/api/main.go#L49) — selector `app := api.New(s)`.

<a id="src-7"></a>
### src-7

[`books/explore-it/project/internal/store/store.go:11`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/store/store.go#L11) — selector `func Open(`.

<a id="src-8"></a>
### src-8

[`books/explore-it/project/internal/webhook/ship.go:38`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/webhook/ship.go#L38) — selector `func deliver(`.

<a id="src-9"></a>
### src-9

[`books/explore-it/project/internal/api/enqueue.go:90`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/enqueue.go#L90) — selector `func (s *Server) batch`.

<a id="src-10"></a>
### src-10

[`books/explore-it/project/cmd/qctl/main.go:17`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/cmd/qctl/main.go#L17) — selector `func main()`.

<a id="src-11"></a>
### src-11

[`books/explore-it/project/cmd/worker/main.go:28`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/cmd/worker/main.go#L28) — selector `worker.Run`.

<a id="src-12"></a>
### src-12

[`books/explore-it/project/internal/client/client.go:26`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/client/client.go#L26) — selector `func (c *Client) Do`.

<a id="src-13"></a>
### src-13

[`books/explore-it/project/internal/store/store.go:12`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/store/store.go#L12) — selector `sql.Open`.

<a id="src-14"></a>
### src-14

[`books/explore-it/project/internal/api/api.go:120`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/api.go#L120) — selector `func (s *Server) Handler()`.

<a id="src-15"></a>
### src-15

[`books/explore-it/project/internal/jobs/transition.go:46`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/transition.go#L46) — selector `var table =`.

<a id="src-16"></a>
### src-16

[`books/explore-it/project/internal/schedule/fire.go:91`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/schedule/fire.go#L91) — selector `func Fire(`.

<a id="src-17"></a>
### src-17

[`books/explore-it/project/internal/api/api.go:219`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/api.go#L219) — selector `func (s *Server) Loop(`.

<a id="src-18"></a>
### src-18

[`books/explore-it/project/internal/webhook/ship.go:21`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/webhook/ship.go#L21) — selector `func Ship(`.

<a id="src-19"></a>
### src-19

[`books/explore-it/project/internal/clock/clock.go:15`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/clock/clock.go#L15) — selector `func (c *Clock) Now()`.

<a id="src-20"></a>
### src-20

[`books/explore-it/project/internal/api/observe.go:35`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/observe.go#L35) — selector `func (s *Server) logEvents(`.

<a id="src-21"></a>
### src-21

[`books/explore-it/project/internal/jobs/ack.go:10`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/ack.go#L10) — selector `func (j *Service) Ack(`.

<a id="src-22"></a>
### src-22

[`books/explore-it/project/internal/jobs/nack.go:5`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/nack.go#L5) — selector `func (j *Service) Nack(`.

<a id="src-23"></a>
### src-23

[`books/explore-it/project/internal/jobs/expire.go:8`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/expire.go#L8) — selector `func (j *Service) Expire(`.

<a id="src-24"></a>
### src-24

[`books/explore-it/project/internal/jobs/operator.go:11`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/operator.go#L11) — selector `func (j *Service) Cancel(`.

<a id="src-25"></a>
### src-25

[`books/explore-it/project/internal/jobs/operator.go:49`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/operator.go#L49) — selector `func (j *Service) Retry(`.

<a id="src-26"></a>
### src-26

[`books/explore-it/project/internal/jobs/jobs.go:95`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/jobs.go#L95) — selector `func notify(`.

<a id="src-27"></a>
### src-27

[`books/explore-it/project/internal/jobs/jobs.go:45`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/jobs.go#L45) — selector `func (j *Service) Lease(`.

<a id="src-28"></a>
### src-28

[`books/explore-it/project/internal/stats/stats.go:8`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/stats/stats.go#L8) — selector `func Read(`.

<a id="src-29"></a>
### src-29

[`books/explore-it/project/internal/api/queues.go:184`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/queues.go#L184) — selector `stats.Read`.

<a id="src-30"></a>
### src-30

[`books/explore-it/project/migrations/003.sql:4`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/migrations/003.sql#L4) — selector `CREATE`.

<a id="src-31"></a>
### src-31

[`books/explore-it/project/internal/api/jobs.go:40`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/jobs.go#L40) — selector `func (s *Server) jobVerb`.

<a id="src-32"></a>
### src-32

[`books/explore-it/project/internal/api/api.go:165`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/api.go#L165) — selector `SELECT namespace_id,admin`.

<a id="src-33"></a>
### src-33

[`books/explore-it/project/internal/api/jobs.go:70`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/jobs.go#L70) — selector `s.Jobs.Ack`.

<a id="src-34"></a>
### src-34

[`books/explore-it/project/internal/api/observe.go:28`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/observe.go#L28) — selector `s.Clock.Change`.

<a id="src-35"></a>
### src-35

[`books/explore-it/project/internal/api/schedules.go:53`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/schedules.go#L53) — selector `schedule.Next`.

<a id="src-36"></a>
### src-36

[`books/explore-it/project/internal/api/api.go:231`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/api.go#L231) — selector `now := s.Clock.Now()`.

<a id="src-37"></a>
### src-37

[`books/explore-it/project/internal/api/api.go:249`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/api.go#L249) — selector `s.Jobs.Promote, s.Jobs.Expire`.

<a id="src-38"></a>
### src-38

[`books/explore-it/project/internal/api/api.go:249`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/api.go#L249) — selector `return schedule.Fire`.

<a id="src-39"></a>
### src-39

[`books/explore-it/project/internal/schedule/fire.go:104`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/schedule/fire.go#L104) — selector `func fireOne(`.

<a id="src-40"></a>
### src-40

[`books/explore-it/project/internal/webhook/ship.go:22`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/webhook/ship.go#L22) — selector `SELECT * FROM webhook_outbox`.

<a id="src-41"></a>
### src-41

[`books/explore-it/project/internal/api/api.go:232`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/api.go#L232) — selector `e := webhook.Ship(s.S, now)`.

<a id="src-42"></a>
### src-42

[`books/explore-it/project/internal/webhook/ship.go:41`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/webhook/ship.go#L41) — selector `mac.Write(body)`.

<a id="src-43"></a>
### src-43

[`books/explore-it/project/internal/api/observe.go:12`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/observe.go#L12) — selector `SELECT * FROM job_events`.

<a id="src-44"></a>
### src-44

[`books/explore-it/project/internal/api/enqueue.go:72`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/api/enqueue.go#L72) — selector `INSERT INTO jobs`.

<a id="src-45"></a>
### src-45

[`books/explore-it/project/internal/jobs/jobs.go:66`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/jobs.go#L66) — selector `FOR UPDATE SKIP LOCKED`.

<a id="src-46"></a>
### src-46

[`books/explore-it/project/internal/worker/worker.go:64`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/worker/worker.go#L64) — selector `CancelRequested bool`.

<a id="src-47"></a>
### src-47

[`books/explore-it/project/internal/jobs/expire.go:80`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/jobs/expire.go#L80) — selector `func (j *Service) Promote(`.

<a id="src-48"></a>
### src-48

[`books/explore-it/project/Makefile:2`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/Makefile#L2) — selector `COMPOSE =`.

<a id="src-49"></a>
### src-49

[`books/explore-it/project/deploy/compose.yml:1`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L1) — selector `name: queuelet`.

<a id="src-50"></a>
### src-50

[`books/explore-it/project/deploy/compose.yml:17`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L17) — selector `  api:`.

<a id="src-51"></a>
### src-51

[`books/explore-it/project/deploy/compose.yml:3`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L3) — selector `  postgres:`.

<a id="src-52"></a>
### src-52

[`books/explore-it/project/deploy/compose.yml:34`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L34) — selector `  worker-a:`.

<a id="src-53"></a>
### src-53

[`books/explore-it/project/deploy/compose.yml:44`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L44) — selector `  worker-b:`.

<a id="src-54"></a>
### src-54

[`books/explore-it/project/deploy/compose.yml:29`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L29) — selector `  hooksink:`.

<a id="src-55"></a>
### src-55

[`books/explore-it/project/deploy/compose.yml:16`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L16) — selector `volumes: ["data:`.

<a id="src-56"></a>
### src-56

[`books/explore-it/project/Makefile:22`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/Makefile#L22) — selector `qctl:`.

<a id="src-57"></a>
### src-57

[`books/explore-it/project/deploy/compose.yml:24`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L24) — selector `DATABASE_URL:`.

<a id="src-58"></a>
### src-58

[`books/explore-it/project/deploy/compose.yml:39`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/deploy/compose.yml#L39) — selector `QUEUELET_URL:`.

<a id="src-59"></a>
### src-59

[`books/explore-it/project/lab/seed/seed.py:10`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/lab/seed/seed.py#L10) — selector `hooksink:8201`.

<a id="src-60"></a>
### src-60

[`books/explore-it/project/internal/client/client.go:18`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/internal/client/client.go#L18) — selector `http://localhost:18200`.

<a id="src-61"></a>
### src-61

[`books/explore-it/project/cmd/hooksink/main.go:25`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/cmd/hooksink/main.go#L25) — selector `/admin/mode`.

<a id="src-62"></a>
### src-62

[`books/explore-it/project/lab/reset.sh:8`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/project/lab/reset.sh#L8) — selector `psql`.

<a id="src-63"></a>
### src-63

[`books/explore-it/exploration/models/map.md:1`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/exploration/models/map.md#L1) — selector `#`. Supporting exploration record; historical observations must be interpreted through the final fix.

<a id="src-64"></a>
### src-64

[`books/explore-it/exploration/models/ecosystem.md:1`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/exploration/models/ecosystem.md#L1) — selector `#`. Supporting exploration record; historical observations must be interpreted through the final fix.

<a id="src-65"></a>
### src-65

[`books/explore-it/exploration/models/rules.md:1`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/exploration/models/rules.md#L1) — selector `#`. Supporting exploration record; historical observations must be interpreted through the final fix.

<a id="src-66"></a>
### src-66

[`books/explore-it/exploration/models/states.md:1`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/exploration/models/states.md#L1) — selector `#`. Supporting exploration record; historical observations must be interpreted through the final fix.

<a id="src-67"></a>
### src-67

[`books/explore-it/lessons/09-explain-it.md:65`](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/lessons/09-explain-it.md#L65) — selector `## The structural fix`. Supporting exploration record; historical observations must be interpreted through the final fix.

## Job-state reference

This table is documentation, not a sixth diagram type. Source: `internal/jobs/transition.go`, plus operation-specific token, expiry, attempt and purge checks.

| Event | From | To / condition |
|---|---|---|
| promote | scheduled, retry_wait | queued when due |
| lease | queued | leased, matching kind / unpaused queue |
| heartbeat | leased | leased, current unexpired token |
| ack | leased | succeeded, current unexpired token |
| nack | leased | retry_wait or dead (refused, spent or purged) |
| expire / release worker | leased | queued or dead (spent or purged); consumes attempt |
| cancel | scheduled, queued, leased, retry_wait | cancelled immediately; cooperative stop at worker |
| operator retry | dead | queued; resets attempts, retains previous last_error |
| purge | scheduled, queued, retry_wait / leased | deletes pending; marks leased, never requeues after failure |
| delete / retain | succeeded, dead, cancelled / succeeded, cancelled | removes row; retention excludes dead |

Terminal transition, attempt and notification insertion share the transaction where an attempt applies. Cancel of unleased work has no attempt row. HTTP delivery is separate and can repeat. A worker can execute twice after lease expiry. Do not infer exactly-once execution, one notification side effect, or complete exploration.

## Qualifications carried forward

- The final book STATE at this snapshot awaited cross-family end review; that review's fix pass (record `1712ced`) merged at `7d26bb8` with an identical project tree (see reconciliation). Checkpoint C passed with a U32 sampling flake and nits; full book-text citations were not validated. Diagram browser success does not close book review.
- Q-E09-1 (wildcard-hour spelling) and Q-E09-2 (last_error clearing) remain Nick decisions. Code treats `0 0-23 * * *` as fixed-hour; success and retry retain last_error. Legacy unadvanceable schedules leave health degraded until disabled/deleted; legacy invalid templates are not revalidated.
- Backward lab-clock movement can invert FIFO and lease before run_at. Semantic time controls do not freeze real network/process timers. This is not a general real-clock correctness proof.
- Worker identity within a tenant (Q-E08-1), redirect handling (H-E08-22), database pool exhaustion during freeze (H-E08-24), worker http-result float64 decoding, and unprobed cancel/retry/promote/retention interleavings remain open.
- Health/auth have bounds against a frozen database; a freeze after authentication can still hold handlers and lifecycle SQL. Health is not a backlog or lag gauge. The constant `queuelet_up 1` metric is not proof of end-to-end progress.
- H-E09-18/19 (nack-to-dead concurrent delete and real-clock retention) are structurally covered by the transaction change, not separately probed. Post-commit fault injection and other named testability requests remain deferred. No chaos suite was rerun for static authoring.

See pinned [STATE](https://github.com/nickstrad/book-notes/blob/b6df7db044782dd3517d041a8aad95308a3a80c9/books/explore-it/STATE.md), lessons 06–09 and `exploration/defects/INDEX.md` for the complete owner-assigned limits. All STE text is an unreviewed draft; no certified language-compliance claim.
