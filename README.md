# ClawFleet

[![GitHub release](https://img.shields.io/github/v/release/clawfleet/ClawFleet)](https://github.com/clawfleet/ClawFleet/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/clawfleet/ClawFleet/blob/main/LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-required-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux-lightgrey)](https://github.com/clawfleet/ClawFleet)
[![Wiki](https://img.shields.io/badge/Docs-Wiki-blue)](https://github.com/clawfleet/ClawFleet/wiki)

🌐 **Website:** [clawfleet.io](https://clawfleet.io) · 💬 **Community:** [Discord](https://discord.gg/b5ZSRyrqbt) · 📝 **Blog:** [Dev.to](https://dev.to/weiyong1024/i-built-an-open-source-tool-to-run-ai-agents-on-my-laptop-they-collaborate-in-discord-managed-1c42)

> Run a fleet of AI agents on your laptop in 5 minutes — OpenClaw and Hermes in isolated containers, managed from one browser dashboard. Use your ChatGPT subscription, no cloud bills.

[中文文档](./README.zh-CN.md)

**Imagine buying N dedicated Mac Minis**, each running its own AI agent — some OpenClaw, some Hermes, all collaborating in Discord. Your own AI company — data stays on your hardware, no SaaS subscription.

**ClawFleet makes that free.** Each agent runs in its own Docker container with isolated filesystem and networking. On your existing Mac or Linux box. ~500 MB RAM per OpenClaw instance, ~150 MB per Hermes instance.

![Dashboard](docs/images/fleet.png)

## Get Started

```bash
curl -fsSL https://clawfleet.io/install.sh | sh
```

5 minutes: Docker installed, image pulled, dashboard running at `http://localhost:8080`. Connect any model provider with a single API key — every instance, OpenClaw or Hermes, runs in its own Docker container with full isolation.

[![Install Demo](https://img.youtube.com/vi/jE5ZR8g477s/maxresdefault.jpg)](https://youtu.be/jE5ZR8g477s)
[![▶ Watch Install Demo (30s)](https://img.shields.io/badge/▶_Watch_Install_Demo-30s-red?style=for-the-badge&logo=youtube)](https://youtu.be/jE5ZR8g477s)

---

## What ClawFleet Does

- **Two runtimes, one Dashboard** — OpenClaw and Hermes, both first-class ([more ↓](#supported-runtimes))
- **Sandboxed instances** — each agent in its own Docker container, isolated from host and peers
- **Any LLM provider** — OpenAI, Anthropic, Google, DeepSeek, or your ChatGPT subscription _(OpenClaw)_
- **`clawfleet shell`** — drop into any instance's terminal: Hermes TUI chat or OpenClaw bash shell
- **Version pinning** — lock tested runtime versions so upstream breaking changes don't touch you
- **Character system** — reusable personas (bio, backstory, style, traits) _(OpenClaw)_
- **Skill management** — 52 built-in + 13,000+ community skills via ClawHub _(OpenClaw)_
- **Soul Archive** — snapshot persona + memory + config, clone into new hires _(OpenClaw)_

## Requirements

- macOS or Linux
- **Mac users:** strongly recommended to install [Docker Desktop](https://www.docker.com/products/docker-desktop/) first for the best experience  
  <sub>Otherwise ClawFleet will automatically install Colima as an alternative Docker runtime.</sub>

## Install Details

The install command above will:
1. Install Docker if needed (Colima on macOS, Docker Engine on Linux)
2. Download and install the `clawfleet` CLI
3. Pull the pre-built sandbox image (~1.4 GB)
4. Start the Dashboard as a background daemon
5. Open http://localhost:8080 in your browser

<details>
<summary><strong>Linux server deployment notes</strong></summary>

The Dashboard listens on `0.0.0.0:8080` by default on Linux. Restrict to localhost:

```bash
clawfleet dashboard stop
clawfleet dashboard start --host 127.0.0.1
```

SSH tunnel from your laptop:

```bash
ssh -fNL 8081:127.0.0.1:8080 user@your-server  # then http://localhost:8081
```

The **Control Panel** (OpenClaw's built-in web UI) requires a [secure context](https://developer.mozilla.org/en-US/docs/Web/Security/Secure_Contexts) — the SSH tunnel provides this. Hermes's native Dashboard does not require a secure context. All other Dashboard features work without a tunnel.

</details>

> **Manual install?** See the [Getting Started](https://github.com/clawfleet/ClawFleet/wiki/Getting-Started) wiki page.

## Run Your Agent Company

Think of ClawFleet as **your AI company**. Assets are the tools your company owns; Fleet is your team of AI employees. You assign tools to employees, then put your AI workforce into production.

### Stock your toolbox

**Assets → Models** — register LLM API keys. The "brains" your employees think with. Each model is validated before saving. Models are shared across all runtimes.

![Models](docs/images/assets-models.png)

**Assets → Characters** — reusable personas. Think of them as "job descriptions" — a CTO, a CPO, a CMO. Bio, backstory, communication style, personality traits. _Characters today are surfaced for OpenClaw instances; Hermes uses its own personality system in its native Dashboard._

![Characters](docs/images/assets-characters.png)

**Assets → Channels** — connect messaging platforms. The "workstations" where your employees serve customers. Validated before saving. _OpenClaw supports 24+ channels; Hermes today supports Discord / Telegram / Slack._

![Channels](docs/images/assets-channels.png)

### Hire & equip your team

**Fleet → Create** — spin up an instance, OpenClaw or Hermes. Each one is a new employee joining your company.

**Fleet → Configure** — assign a model, character, and channel. Give your CTO a Claude brain and a Discord workstation. Give your CMO a GPT brain and a Slack feed. Different employees, different tools, different personalities.

![Fleet](docs/images/fleet.png)

### Teach them new skills

**Fleet → Skills** — each instance has access to 52 built-in skills (weather, GitHub, coding, and more). Want more? Search 13,000+ community skills on [ClawHub](https://clawhub.com) and install them with one click. Different employees can learn different skills. _Skill Manager today is OpenClaw-specific — Hermes manages skills via its own Dashboard._

![Skills](docs/images/skills.png)

### Save & clone employee souls

Once an employee is trained and performing well, save their soul — personality, memory, model config, and conversation history — so you can clone them instantly. _Soul Archive today is OpenClaw-specific._

**Fleet → Save Soul** — click on any configured instance to save its soul.

![Save Soul](docs/images/soul-save-dialog.png)

**Fleet → Soul Archive** — browse all saved souls, ready to be loaded into new hires.

![Soul Archive](docs/images/soul-archive.png)

**Fleet → Create → Load Soul** — when creating new instances, pick a soul. The new employee starts with all the knowledge and personality of the original.

![Load Soul](docs/images/soul-create.png)

### Watch your team collaborate

Connect your fleet to messaging platforms and watch your AI employees work together. Here, an engineer, product manager, and marketer welcome a new teammate — all running autonomously in a Discord group chat.

![Bot Collaboration](docs/images/welcome-on-board-for-bot.jpeg)

## Supported Runtimes

ClawFleet manages all agent runtimes as first-class citizens in one unified Dashboard — with shared asset pool (Models, Channels), per-instance container isolation, live stats, logs, event streams, and `clawfleet shell` access.

Today ClawFleet ships with two runtimes: **OpenClaw** and **Hermes**. OpenClaw has been supported longer, so more of its features are currently surfaced in the Dashboard UI. Hermes support is newer; its current Dashboard scope is container lifecycle and Configure (model + channel), with deeper features accessible via Hermes's own native Dashboard. Bringing both runtimes to full Dashboard parity is an active direction — and it's how future runtimes will be added too.

### OpenClaw — channel-native agent ensemble

Upstream: [github.com/openclaw/openclaw](https://github.com/openclaw/openclaw)

OpenClaw is a local-first assistant gateway that makes a single agent addressable from 24+ messaging platforms — Telegram, Discord, Slack, Lark, WhatsApp, Signal, Matrix, and more. The bot lives *inside* the conversation as a participant, with pairing codes, @-mentions, and roster-aware coordination with other bots.

Pick OpenClaw when you want:
- A bot that lives as a participant inside a group chat (not a DM assistant you have to call)
- A fleet of distinct personas that collaborate (CEO + CTO + PM in one Discord) via Roster-aware @mentions
- Installable skills from a curated community catalog (ClawHub, 13,000+)

In ClawFleet today:
- **Configure** — assign Model + Character + Channel from the Dashboard
- **Characters** as reusable personas, hot-reloaded via SOUL.md
- **Skills** — 52 bundled + ClawHub install from the Skill Manager
- **Save Soul** — snapshot persona + config into the archive, clone into new hires

### Hermes — single-user learning agent

Upstream: [github.com/NousResearch/hermes-agent](https://github.com/NousResearch/hermes-agent)

Hermes is built around a learning loop: the agent writes new skills from experience, improves them while using them, and maintains memory that compounds across sessions (FTS5 cross-session search + Honcho user modeling). The primary interface is a TUI; messaging platforms are secondary remote controls.

Pick Hermes when you want:
- One personal agent that learns you over time (not a group-chat participant, but a "my assistant")
- First-class cron / scheduled automations delivered to messaging
- Remote execution on SSH / Daytona / Modal (serverless) backends
- Long-tail LLM providers (Nous Portal, GLM, Kimi, MiMo, MiniMax, OpenRouter's 200+)
- A TUI-first workflow

In ClawFleet today:
- **Configure** — assign Model + Channel (Discord / Telegram / Slack) from the Dashboard, using the same Asset pool OpenClaw uses
- **Hermes Dashboard** — one-click link to Hermes's native Dashboard for credential pool, cron, personality, terminal backends
- **`clawfleet shell hermes-1`** — drop into Hermes's interactive TUI

![OpenClaw Control Panel and Hermes Dashboard side by side](docs/images/runtime_native_dashboard.png)

On top of either runtime, ClawFleet also exposes the instance's graphical desktop in the browser via noVNC — useful for watching what the agent does, manually steering a workflow, or demoing. Available today on OpenClaw images (which ship an XFCE desktop); bringing equivalent visibility to Hermes is on the roadmap.

![Per-instance browser desktop](docs/images/instance-desktop.jpeg)

## Documentation

See the **[Wiki](https://github.com/clawfleet/ClawFleet/wiki)** for full documentation:
- [Getting Started](https://github.com/clawfleet/ClawFleet/wiki/Getting-Started) — prerequisites, install, first instance
- [Dashboard Guide](https://github.com/clawfleet/ClawFleet/wiki/Dashboard-Guide) — sidebar, asset management, fleet management
- LLM Provider guides — [Anthropic](https://github.com/clawfleet/ClawFleet/wiki/Provider-Anthropic) | [OpenAI](https://github.com/clawfleet/ClawFleet/wiki/Provider-OpenAI) | [Google](https://github.com/clawfleet/ClawFleet/wiki/Provider-Google) | [DeepSeek](https://github.com/clawfleet/ClawFleet/wiki/Provider-DeepSeek)
- Channel guides — [Telegram](https://github.com/clawfleet/ClawFleet/wiki/Channel-Telegram) | [Discord](https://github.com/clawfleet/ClawFleet/wiki/Channel-Discord) | [Slack](https://github.com/clawfleet/ClawFleet/wiki/Channel-Slack) | [Lark](https://github.com/clawfleet/ClawFleet/wiki/Channel-Lark)
- [CLI Reference](https://github.com/clawfleet/ClawFleet/wiki/CLI-Reference) | [FAQ](https://github.com/clawfleet/ClawFleet/wiki/FAQ)

## CLI Reference

Every command supports `--help`:

```bash
clawfleet --help              # All commands
clawfleet dashboard --help    # Dashboard subcommands
```

Quick reference:

```bash
clawfleet create <N> [--runtime openclaw|hermes]   # Create N instances (default: openclaw)
clawfleet create <N> --pull                         # Force re-pull image from registry
clawfleet create 1 --from-snapshot <soul>           # Clone from a saved soul (OpenClaw)
clawfleet configure <name>                          # Configure model + channel (OpenClaw via CLI; Hermes via Dashboard)
clawfleet list                                      # List instances and status
clawfleet shell <name>                              # Drop into terminal: Hermes TUI / OpenClaw bash
clawfleet desktop <name>                            # Open noVNC desktop (OpenClaw)
clawfleet start|stop|restart <name|all>             # Lifecycle
clawfleet logs <name> [-f]                          # View logs
clawfleet destroy <name|all> [--purge]              # Destroy (--purge also deletes data)
clawfleet snapshot save|list|delete <name>          # Soul archive (OpenClaw)
clawfleet dashboard serve|stop|restart|open         # Dashboard daemon
clawfleet build                                     # Build image locally
clawfleet config | version                          # Config / version info
```

## Reset

Destroy all instances (including data), stop the Dashboard, remove all build artifacts:

```bash
make reset
```

## Resource Usage

Idle memory, measured on M4 MacBook Air (16 GB RAM):

| Instances | OpenClaw RAM | Hermes RAM |
|-----------|--------------|------------|
| 1         | ~700 MB      | ~140 MB    |
| 3         | ~2.1 GB      | ~400 MB    |
| 5         | ~3.5 GB      | ~700 MB    |

<sub>OpenClaw memory rises ~3× when the agent is actively browsing (Chromium loaded). Hermes stays roughly flat.</sub>

## License

MIT · Contributions welcome — open an issue or PR. Reach out: weiyong1024@gmail.com
