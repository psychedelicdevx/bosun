# Product

## Register

product

## Users

Developers and operators who manage Docker or Podman containers from a terminal, often across several side projects or remote hosts. They value fast keyboard workflows, compact information, and tools that remain legible during focused troubleshooting.

## Product Purpose

bosun is a lightweight keyboard cockpit for containers that removes repetitive Docker CLI work without becoming a heavyweight desktop dashboard. Success means the common loop of finding a container, reading or filtering logs, checking stats, and taking lifecycle actions stays one keypress away.

## Brand Personality

Calm, sharp, and capable. bosun should feel modern and polished while remaining terminal-native, direct, and trustworthy.

The canonical visual north star is `docs/social-preview.png`, with its reproducible source in `docs/social-preview.html`. The approved original is also named `Docker Logs.png`. The application UI should stay visually aligned with that preview as features evolve.

## Anti-references

- Heavy desktop dashboards that hide routine actions behind navigation and chrome.
- Flat monochrome TUIs where active state, hierarchy, and keyboard affordances are hard to scan.
- Neon cyberpunk styling, decorative gradients, excessive borders, or color used without meaning.
- Interfaces that look polished in marketing media but noticeably rougher in the product itself.

## Design Principles

- Make the product earn the promise made by its media.
- Keep the active task and available keypresses instantly scannable.
- Use hierarchy and color semantically, never as decoration.
- Preserve terminal-native speed, density, and remote-friendly behavior.
- Prefer a small number of consistent visual roles across every view and theme.

## Accessibility & Inclusion

Maintain strong contrast on dark terminal backgrounds, preserve the monochrome theme, and never communicate container state or selection through color alone. Status dots pair with text, selection uses weight and background, and keyboard actions remain visible as text labels.
