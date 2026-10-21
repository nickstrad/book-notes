# book-notes

Technical books, turned into small programs that teach their core ideas and into skills that
codify their main parts as reusable procedures.

Each book gets a folder under `books/`. Inside, every core idea of the book becomes a **topic**:
a runnable demo (Rust, Go, Node, or Bash, chosen per topic by what teaches best) with a teaching
README, and, when the idea is a failure mode and its remedy, a `before` run that shows the failure
and an `after` run that shows the fix. At the end of each book sits a `skills/` folder whose
`SKILL.md` files turn the book's main parts into procedures an agent can run on real code.

The process for breaking down any book is in [`AGENTS.md`](AGENTS.md). It is book-agnostic; the
first worked example is Release It!.

## How a book is laid out

```
books/<slug>/
  PLAN.md             research product: the book's argument, chosen topics, chosen skills, omissions
  README.md           landing page and index for the book
  run-all.sh          runs every topic demo (the reviewer's first command)
  topics/NN-<topic>/  README.md + run.sh + before/after demo, one language per topic
  skills/<skill>/     SKILL.md: when to use, inputs, procedure, output, source chapters
```

Every topic runs with one command from its folder:

```
./run.sh          # before, then after
./run.sh before   # just the failure
./run.sh after    # just the remedy
```

Demos use only the standard library of their language and need no network or services.

## Books

| Book | Status | Plan |
|---|---|---|
| Release It! Design and Deploy Production-Ready Software, 2nd ed. (Michael T. Nygard, 2018) | phase 1 done: planned and scaffolded; topics and skills not yet built | [`books/release-it/PLAN.md`](books/release-it/PLAN.md) |

## Toolchains

Go 1.26, Node 22 (`export PATH="$HOME/.nvm/versions/node/v22.23.2/bin:$PATH"`), Rust stable via
rustup (`~/.cargo/bin`), bash 5. Each `run.sh` checks for the toolchain it needs.
