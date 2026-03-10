# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

ClawSandbox — a minimal, easy-to-use virtualized sandbox tool for OpenClaw. Solves the problem of rapidly deploying multiple isolated OpenClaw instances on a single machine, enabling users to run an entire OpenClaw fleet. Open-sourced on GitHub.

## Build/Test/Lint Commands

```bash
# Download dependencies
go mod tidy

# Build CLI binary → bin/clawsandbox
make build

# Run tests
make test

# Build Docker image (first time, ~1.4GB, takes several minutes)
make docker-build
# or via the CLI itself (uses embedded Dockerfile):
./bin/clawsandbox build

# Cross-compile for all platforms
make build-all
```

## Architecture

Go CLI tool (cobra) that manages Docker containers with an embedded Web Dashboard. Key packages:

- `cmd/clawsandbox/` — binary entry point
- `internal/cli/` — cobra commands (build, create, list, start, stop, restart, destroy, desktop, logs, dashboard, config, version)
- `internal/container/` — Docker SDK wrappers (client, image build/check/pull, network, container lifecycle, stats)
- `internal/port/` — TCP port availability checker and allocator
- `internal/state/` — instance metadata store (`~/.clawsandbox/state.json`), mutex-protected
- `internal/config/` — config file loader (`~/.clawsandbox/config.yaml`)
- `internal/assets/` — embedded Docker build context (Dockerfile, supervisord.conf, entrypoint.sh)
- `internal/web/` — Web Dashboard: HTTP server, REST API handlers, WebSocket endpoints (stats/logs/events), embedded frontend
- `internal/version/` — build version info (injected via ldflags)

Each claw instance is a Docker container running: XFCE4 desktop + TigerVNC + noVNC (browser access on port 690N) + OpenClaw Gateway (port 1878N).

Container data is persisted at `~/.clawsandbox/data/<name>/openclaw/` → `/home/node/.openclaw` inside the container.

## Engineering Principles

All design decisions, project structure, and code implementation must follow best engineering practices. Specifically:

### Security & Least Privilege
- Never use overly permissive settings (e.g. `chmod 0777`) as shortcuts. Solve permission problems by understanding the ownership model and applying minimal necessary access.
- Container processes must run with the correct user for the operation: system management commands (e.g. `supervisorctl`) run as `root`, application commands (e.g. `openclaw`) run as the unprivileged `node` user.
- Never embed secrets or credentials in code, images, or config files.

### Correctness Over Convenience
- Understand the tools you're automating. Read `--help`, check actual behavior, and verify assumptions (e.g. model name formats, plugin enablement, API readiness) before writing integration code.
- Prefer explicit configuration over implicit defaults. If a third-party tool has a default that doesn't suit our use case (e.g. `dmPolicy: "pairing"`), set the desired value explicitly rather than hoping users will figure it out.
- When orchestrating multi-step processes, respect ordering dependencies and readiness checks (e.g. wait for a service to be healthy before issuing commands against it).

### Simplicity & Minimal Surface
- Don't add abstractions, flags, or config options for hypothetical future needs. Solve the current problem directly.
- Prefer calling existing CLI tools (`docker exec` + `openclaw` CLI) over writing config files directly — this keeps the integration resilient to upstream format changes.
