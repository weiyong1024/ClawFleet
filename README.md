# ClawSandbox

> Deploy and manage a fleet of isolated [OpenClaw](https://github.com/openclaw/openclaw) instances on a single machine.

[中文文档](./README.zh-CN.md)

---

## Background

LLM AI applications are evolving through three stages:

1. **ChatBot** — helps everyone access knowledge
2. **Agent** — makes everyone a professional
3. **OpenClaw** — makes everyone a manager

OpenClaw is a self-hosted personal AI assistant that connects to 20+ messaging platforms including WhatsApp, Telegram, and Slack. ClawSandbox removes the deployment bottleneck — instead of struggling to run a single instance, you can spin up an entire fleet with one command.

## What ClawSandbox Does

- **One-command fleet deployment** — give it a number, get that many isolated OpenClaw instances
- **Full desktop per instance** — each lobster runs in its own Docker container with an XFCE desktop, accessible from any browser via noVNC
- **Lifecycle management** — create, start, stop, and destroy instances with simple CLI commands
- **Data persistence** — each instance's data survives container restarts
- **Resource isolation** — instances are isolated from your host system and from each other

## Requirements

- macOS or Linux
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed and running

## Quick Start

### 1. Install

```bash
# Build from source
git clone https://github.com/weiyong1024/ClawSandbox.git
cd ClawSandbox
make build
# Optionally add to PATH:
sudo make install
```

### 2. Build the Docker image

Required once. Downloads ~1.4 GB and takes a few minutes.

```bash
clawsandbox build
```

### 3. Deploy your fleet

```bash
# Create 3 isolated OpenClaw instances
clawsandbox create 3

# Check status
clawsandbox list
```

### 4. Set up each lobster

Each lobster needs a one-time configuration via its desktop:

```bash
# Open lobster-1's desktop in your browser (noVNC)
clawsandbox desktop lobster-1
```

Inside the desktop terminal:

```bash
# Step 1: Run the setup wizard (configure LLM API key, Telegram bot, etc.)
openclaw onboard --flow quickstart

# Step 2: Start the Gateway
openclaw gateway --port 18789
```

Once the Gateway is running, open **Chromium** on the desktop and navigate to the URL shown in the terminal (e.g. `http://127.0.0.1:18789/#token=...`) to access the OpenClaw Control UI.

## CLI Reference

```bash
clawsandbox build                       # Build the Docker image
clawsandbox create <N>                  # Create N lobster instances
clawsandbox list                        # List all instances and their status
clawsandbox desktop <name>              # Open an instance's desktop in the browser
clawsandbox start <name|all>            # Start a stopped instance
clawsandbox stop <name|all>             # Stop a running instance
clawsandbox logs <name> [-f]            # View instance logs
clawsandbox destroy <name|all>          # Destroy instance (data kept by default)
clawsandbox destroy --purge <name|all>  # Destroy instance and delete its data
```

## Resource Usage

Tested on M4 MacBook Air (16 GB RAM):

| Instances | RAM (idle) | RAM (Chromium active) |
|-----------|------------|-----------------------|
| 1         | ~1.5 GB    | ~3 GB                 |
| 3         | ~4.5 GB    | ~9 GB                 |
| 5         | ~7.5 GB    | not recommended       |

## Project Status

Actively developed. CLI is functional. Web UI planned for a future release.

Contributions and feedback welcome — please open an issue or PR.

## License

MIT
