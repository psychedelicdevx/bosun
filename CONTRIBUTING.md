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

## Before you open a PR

Run the checks the CI runs:

```bash
go build ./...
go vet ./...
go test ./...
```

Keep changes focused. A small PR that does one thing is easier to review and much more likely to get merged than a large one that does five. If you are planning something big, open an issue first so we can talk it through before you write a lot of code.

## What would help most

- Bug reports, especially with the exact steps and your Docker or Podman version.
- Rough edges in the UI on unusual terminal sizes or themes.
- The items on the roadmap in the README under Status.

## Style

The code aims to read like the code already there: small functions, standard library first, no cleverness that needs decoding at 3am. Match the surrounding style and you are good.

Issues and pull requests are both welcome. Thanks again.
