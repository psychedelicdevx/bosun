# Contributing to bosun

Thanks for taking a look. bosun is young and help is genuinely welcome, whether that is a bug report, a fix, or a feature you have wanted.

## Getting set up

You need Go 1.24 or newer and a running Docker or Podman daemon.

```bash
git clone https://github.com/psychedelicdevx/bosun.git
cd bosun
go build -o bosun ./cmd
./bosun
```

There is a self contained sandbox with fake containers, images, volumes, and networks, which is handy for poking at the UI without touching anything real:

```bash
./bosun --demo
```

## Project layout

- `cmd/` is the entry point and flag handling.
- `internal/docker/` talks to the daemon through the Docker SDK. No UI code here.
- `internal/ui/` is the Bubble Tea model, update loop, views, and themes.
- `internal/demo/` is the fake engine behind `--demo`.

Keep the boundary clean: `internal/docker` knows nothing about the UI, and `internal/ui` never imports the Docker SDK directly, it goes through the `Engine` interface.

## Before you open a PR

Run the same checks CI runs. All three must pass:

```bash
go build ./...
go vet ./...
go test ./...
```

Keep changes focused. A small PR that does one thing is easier to review and much more likely to get merged than a large one that does five. If you are planning something big, open an issue first so we can talk it through before you write a lot of code.

## Code style and conventions

The codebase is deliberately small and plain. Please match it.

- **No comments.** The code should read on its own. Reach for a clearer name or a smaller function before you reach for a comment. If a piece of code needs a comment to be understood, that is usually a sign it should be simplified instead.
- **Standard library first.** Do not add a dependency for something a few lines of Go can do. New dependencies need a real reason.
- **Small, focused functions** with clear names. Boring and obvious beats clever. Clever is what someone has to decode at 3am.
- **Match the surrounding style.** Read the file you are editing and write code that looks like it was already there.
- **Handle errors, do not swallow them.** Surface failures to the user in plain language.

## Tests

Anything with real logic gets a test: a branch, a parser, a state transition, a keybinding path. Look at the existing `_test.go` files for the style. They are plain Go tests, table driven where it helps, no extra frameworks and no fixtures. A trivial one line change does not need a test, but new behavior does.

## Commits

- Use short, present tense messages with a type prefix: `feat:`, `fix:`, `docs:`, `chore:`, `test:`, `refactor:`. Scope in parentheses is nice when it helps, for example `feat(logs): add filter`.
- Keep commits granular. One logical change per commit is much easier to read and revert than a single commit that does everything.

## What would help most

- Bug reports, especially with the exact steps and your Docker or Podman version.
- Rough edges in the UI on unusual terminal sizes or themes.
- The items on the roadmap in the README under Status.

Issues and pull requests are both welcome. Thanks again.
