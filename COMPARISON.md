# bosun vs other Docker TUIs

An honest comparison. bosun is young and does less than some of these tools, and this page says so where it is true. The goal is to help you pick the right one, not to win an argument.

If you want the shortest version: bosun is a fast, keyboard first TUI that covers the daily container loop, gets logs right, and works with Docker and Podman. If you need the deepest feature set and the biggest community, lazydocker is still the safer pick today.

## At a glance

| | bosun | lazydocker | ctop | oxker | dozzle |
| --- | --- | --- | --- | --- | --- |
| Interface | TUI | TUI | TUI | TUI | Web |
| Containers | manage | manage | view | manage | view |
| Images | yes | yes | no | no | no |
| Volumes | yes | yes | no | no | no |
| Networks | yes | yes | no | no | no |
| Compose grouping | yes | yes | no | no | no |
| Stack start/stop/restart | yes | yes | no | no | no |
| Logs: follow | yes | yes | yes | yes | yes |
| Logs: filter / search | yes | no | no | no | yes |
| Logs: copy to clipboard | yes | no | no | no | yes |
| Live CPU / memory | yes | yes, with graphs | yes | yes | no |
| Memory in MB/GB | yes | percent only | yes | yes | n/a |
| Shell into container | yes | yes | yes | no | no |
| Podman | yes | no | no | no | yes |
| Remote over tcp / unix | yes | yes | no | limited | yes |
| Remote over ssh | not yet | yes | no | no | yes |
| Follows `docker context` | yes | yes | no | no | partial |
| Themes | 10 | limited | no | limited | yes |
| Single static binary | yes | yes | yes | yes | no, runs a server |
| Actively maintained | yes | slowing | no, last release 2024 | yes | yes |

## The short version per tool

**lazydocker** is the mature, full featured veteran with a large community. It covers everything bosun does and more: volume sizes, ASCII stats graphs, custom commands, and ssh contexts. If you want the most complete, most battle tested Docker TUI, use lazydocker. Its weak spots are logs, where users have long asked for text copy, search, and a fix for the recurring empty logs bug, and Podman, which has been the single most requested feature for years and is still not supported. It is also slowing down, with releases now months apart.

**ctop** is a focused container metrics viewer, like `top` for containers. It is great for a quick "what is eating my CPU" glance. It does not manage containers, has no compose awareness, and its log view is known to break on scroll. It has not had a release since 2024.

**oxker** is a clean, lightweight Rust TUI that shows containers, logs, and stats on one screen. It is a lovely simple viewer. It is not compose first and has a smaller feature surface than bosun or lazydocker.

**dozzle** is a real time log viewer with a web UI, not a terminal app. It is excellent at logs across many containers and supports Docker, Swarm, and Kubernetes. It is a different shape of tool: it runs a server and you open a browser, where bosun stays in your terminal and also manages containers.

## Where bosun is genuinely better

- **Logs.** This is bosun's focus, and it is the weakest area across most of the field. bosun streams logs with follow, pauses follow when you scroll up and resumes at the bottom, filters lines live with `/`, and copies what you are viewing to the clipboard with `y`. The copy uses the terminal's own clipboard escape, so it works over SSH.
- **Podman.** bosun talks to the Docker API directly, so it works against Podman's socket with no extra setup. Listing, logs, and lifecycle actions are verified working. This is the feature lazydocker users have asked for the most and never got.
- **It does not break when Docker updates.** bosun negotiates the API version with the daemon, so it avoids the "client version is too old" errors that lazydocker users hit after Docker upgrades.
- **Memory in real units.** Stats show usage in MB and GB, not just a percent.
- **Actively developed and small.** bosun is a lean codebase shipping regularly, versus larger tools that are slowing down or already stalled.

## Where lazydocker still wins

- Volume sizes, ASCII stats graphs over time, and user defined custom commands.
- `ssh://` remote contexts out of the box. bosun handles tcp and unix endpoints today; for an ssh reachable daemon you expose it over tcp or open a tunnel.
- Six years of maturity, a huge community, and packaging everywhere.

## When to pick which

- You live in the terminal, want logs done right, and use Docker or Podman: **bosun**.
- You want the deepest feature set and the most proven tool: **lazydocker**.
- You only want a resource monitor: **ctop**, though it is unmaintained.
- You want the leanest possible viewer: **oxker**.
- You want logs in a browser across many hosts: **dozzle**.

Notes: claims about bosun are from the current release. Podman was verified for container listing, images, and log streaming; live stats depend on the Podman host's cgroup setup. Details about other tools are from their public repositories and issue trackers as of August 2026 and may change.
