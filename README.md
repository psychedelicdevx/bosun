# bosun

A terminal UI for Docker that stays out of your way. Bordered panels, one-key actions, live logs and stats. It is the tool I wanted while babysitting a dozen containers across side projects, so I built it.

![bosun in action](docs/demo.gif)

## Why

I spend half my day in `docker ps`, `docker logs -f`, `docker exec`, over and over. The web dashboards are heavy and the CLI makes me retype the same three commands. bosun keeps everything one keypress away and never leaves the terminal.

It is not trying to replace Docker Desktop. It is a fast keyboard cockpit for the containers you already have running.

## Install

You need a running Docker daemon. That is it for the Homebrew route.

Homebrew:

```bash
brew install psychedelicdevx/tap/bosun
```

With Go (1.24 or newer):

```bash
go install github.com/psychedelicdevx/bosun/cmd@latest
```

Or build it yourself:

```bash
git clone https://github.com/psychedelicdevx/bosun.git
cd bosun
go build -o bosun ./cmd
./bosun
```

## Usage

Run it:

```bash
bosun
```

It connects to your local daemon through the usual socket, lists every container, and hands you the keyboard.

You can point it at another daemon. bosun follows your active `docker context`, so if you have run `docker context use` it connects there without any extra setup. `DOCKER_HOST` overrides that, and a `--host` flag overrides everything:

```bash
bosun --host tcp://10.0.0.5:2375
```

`tcp://` and `unix://` endpoints work today. `ssh://` contexts are not wired up yet, so for an SSH-reachable daemon expose it over `tcp://` (or open a tunnel) and point `DOCKER_HOST` or `--host` at that.

### Podman

bosun talks to the Docker API directly, so it works with Podman too. Start Podman's API service and point bosun at its socket:

```bash
podman system service --time=0 &
bosun --host unix://$XDG_RUNTIME_DIR/podman/podman.sock
```

Listing containers and images, streaming and filtering logs, and the lifecycle actions all work against Podman. Live stats depend on your Podman host having cgroups set up the usual way; if your setup cannot report them, bosun says so instead of guessing. The list on the left is what you navigate. The panel on the right shows details, and switches to logs or stats when you ask for them. Whichever panel has focus gets a green border, same idea as lazygit.

If you run Compose stacks, the list groups containers by project so you can see each stack together and collapse the ones you are not touching right now. Containers that are not part of a project sit in their own group at the bottom. With the cursor on a project header, `s`, `x`, and `r` start, stop, or restart the whole stack at once. It operates the containers that already exist rather than recreating them from the compose file, so it stays fast and works the same over a remote daemon.

Press `1` through `4` to switch between four views. Containers is the default. Images shows every image with its size, lets you remove ones you do not need, and prunes the dangling `<none>` images that pile up after rebuilds. Volumes lists your named volumes with their driver and mountpoint. Networks lists your networks with their driver, scope, subnet, and gateway. In the volumes and networks views, `d` removes the selected one after asking.

If the daemon is down or the socket is not readable, bosun tells you in plain language instead of dumping a stack trace.

Want to try it without touching your real containers? Run `bosun --demo` for a self-contained sandbox with fake containers, logs, and stats. Handy for screenshots and for kicking the tires.

## Keys

| Key | What it does |
| --- | --- |
| `↑` `↓` or `k` `j` | move through the list |
| `1`–`4` | switch between containers, images, volumes, networks |
| `/` | filter the list by name |
| `space` | collapse or expand a compose group |
| `tab` | switch focus between the two panels |
| `enter` | stream logs for the selected container |
| `/` (in logs) | filter log lines by text as they stream |
| `y` (in logs) | copy the shown logs to your clipboard |
| `S` | live CPU and memory |
| `e` | drop into a shell inside a running container |
| `s` | start |
| `x` | stop |
| `r` | restart |
| `d` | remove container or image, asks first |
| `p` | prune dangling images (images view) |
| `esc` | back to the details view |
| `?` | show the keybindings |
| `q` | quit |

Logs follow along at the bottom while new lines arrive. Scroll up and it stops chasing so you can read, then it picks back up once you return to the bottom. Press `/` while viewing logs to filter to just the lines that match, and `y` to copy what you are looking at to your clipboard. The copy uses the terminal's own clipboard escape, so it works over SSH too. The list refreshes itself every couple of seconds, so anything you change from another terminal shows up on its own.

## Themes

Press `T` to cycle color themes while running, or pick one at launch:

```bash
bosun --theme dracula
```

Built in themes: `default`, `catppuccin`, `dracula`, `tokyonight`, `nord`, `gruvbox`, `rosepine`, `solarized`, `monokai`, `mono`. Run `bosun --themes` to list them, or set `BOSUN_THEME` in your shell to make one stick.

## How it works

The whole thing is one static binary. Under the hood:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) drives the event loop, with [Lip Gloss](https://github.com/charmbracelet/lipgloss) for the panels and colors.
- The official [Docker Go SDK](https://github.com/docker/docker) talks to the daemon.
- Logs and stats each run in their own goroutine and feed the UI through channels, so streaming never blocks a keypress. Closing a panel cancels its stream cleanly, no leaked goroutines.

## Status

Early but usable every day. Container list, logs, stats, the lifecycle actions, and shell exec all work. Next on the bench: compose project grouping, image and volume panels, and a name filter for when the list gets long.

Issues and pull requests are welcome.

## License

MIT
