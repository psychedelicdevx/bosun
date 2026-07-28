# bosun

A fast, clean terminal UI for managing Docker containers. Think `lazydocker`, but snappier, with better keybindings and a cleaner visual hierarchy.

<!-- TODO: demo.gif here -->

## Features

- Live container list, color coded by status (running / stopped / errored)
- Stream logs for any container with follow, scroll back to pause
- Live CPU and memory stats
- Start, stop, restart, remove containers (remove asks to confirm)
- Exec a shell straight into a running container
- Arrow keys or `j`/`k`, whichever you prefer

## Install

Requires Go 1.24+ and a running Docker daemon.

```bash
go install github.com/psychedelicdevx/bosun/cmd@latest
```

Or build from source:

```bash
git clone https://github.com/psychedelicdevx/bosun.git
cd bosun
go build -o bosun ./cmd
./bosun
```

## Usage

Just run it:

```bash
bosun
```

It connects to your local Docker daemon (via the standard `DOCKER_HOST` / socket), lists every container, and you drive it from the keyboard.

## Keybindings

| Key | Action |
|---|---|
| `↑` / `↓` (or `k` / `j`) | move selection |
| `enter` | stream logs for the selected container |
| `S` | live CPU / memory stats |
| `e` | exec a shell into a running container |
| `s` | start container |
| `x` | stop container |
| `r` | restart container |
| `d` | remove container (asks to confirm) |
| `esc` | back to the list |
| `?` | toggle help |
| `q` | quit |

## Built with

- [bubbletea](https://github.com/charmbracelet/bubbletea) + [lipgloss](https://github.com/charmbracelet/lipgloss) for the TUI
- The official [Docker Go SDK](https://github.com/docker/docker)
