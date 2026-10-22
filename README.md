# book-notes

Technical books, turned into one deliberately flawed project per book that the book's lessons
improve commit by commit, and into skills that codify the book's main parts as reusable procedures.

Each book gets a folder under `books/`. Inside, a single realistic project built on real technology
(Docker Compose, a real database, real instrumentation) starts at a tagged **flawed** state seeded
with the mistakes the book warns about. Every core lesson of the book is then a **patch commit**
with its own tag, a teaching note, and a verify script that reproduces the flaw on the previous tag
and shows the improvement on this one. Reading `git log` between the first and last tag is reading
the book applied. A `skills/` folder turns the book's main parts into procedures an agent can run on
real code, each pointing at the project as its worked example.

The process for breaking down any book is in [`AGENTS.md`](AGENTS.md). It is book-agnostic; the
worked example is Release It!.

## How a book is laid out

```
books/<slug>/
  PLAN.md             the confirmed plan: project design, flaw register, lesson sequence, skills
  README.md           landing page and index for the book
  lessons/NN-*.md     one teaching note per lesson (00 is the tour of the flawed start)
  project/            the one codebase: make up / make verify LESSON=NN / make down
  skills/<skill>/     SKILL.md: when to use, inputs, procedure, output, source, worked example
```

Reader's path, for any book:

```
git checkout <slug>/v0-flawed            # the before
make -C books/<slug>/project up verify LESSON=00
git checkout <slug>/lesson-05
make -C books/<slug>/project verify LESSON=05 REF=<slug>/lesson-04   # FLAW REPRODUCED
make -C books/<slug>/project verify LESSON=05                        # IMPROVED
```

## Books

| Book | Status | Plan |
|---|---|---|
| Release It! Design and Deploy Production-Ready Software, 2nd ed. (Michael T. Nygard, 2018) | building: `release-it/v0-flawed` tagged (the flawed `litefleet` platform, 18 flaws); lessons landing one tag at a time | [`books/release-it/PLAN.md`](books/release-it/PLAN.md) |

Earlier plans are archived next to the current one: v2 (same structure, a web-shop `checkout`
service as the vehicle) at `books/release-it/PLAN-v2-checkout.md`, and v1 (twelve standalone
mini-projects) at `books/release-it/PLAN-v1-per-topic-projects.md`; v1's stub folders under
`books/release-it/topics/` were removed when v0 was built.

## Toolchains

Docker 29 with Compose v5, Go 1.26, Node 22 (`export PATH="$HOME/.nvm/versions/node/v22.23.2/bin:$PATH"`),
Rust stable via rustup (`~/.cargo/bin`), bash 5. Host ports 5432 and 8080 are taken on this VM by
another project; book projects bind host ports in the 1xxxx range.
