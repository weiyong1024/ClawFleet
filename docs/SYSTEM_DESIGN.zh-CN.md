# ClawFleet 系统设计文档

> 版本: v0.4 | 日期: 2026-03-30

[English](./SYSTEM_DESIGN.md)

---

## 1. 概述

ClawFleet 在单台机器上部署和管理多个隔离的 OpenClaw 实例，通过浏览器 Dashboard 完成全部操作。每个实例运行在独立的 Docker 容器中，拥有完整的 Linux 桌面（XFCE4 + TigerVNC + noVNC），用户可通过浏览器访问。用户可在 Dashboard 或 CLI 中管理军团——创建实例、配置 LLM 供应商、分配消息渠道、定义角色人设、监控资源。

## 2. 架构分层

ClawFleet 构建于 ClawSandbox 之上，后者是专为容器编排和实例生命周期管理打造的基础设施层。

```
┌─────────────────────────────────────────────────────────────┐
│                   浏览器 (Dashboard UI)                       │
│              Preact SPA @ http://localhost:8080              │
└──────────────────────────┬──────────────────────────────────┘
                           │ REST API + WebSocket
┌──────────────────────────▼──────────────────────────────────┐
│                  ClawFleet（产品层）                           │
│  internal/web/ + internal/cli/                               │
│  REST API、WebSocket 流、资产管理、技能、                       │
│  i18n、花名册、快照、daemon 管理                                │
├─────────────────────────────────────────────────────────────┤
│                ClawSandbox（基础设施层）                       │
│  internal/container/、/state/、/port/、/config/、             │
│  /assets/、/snapshot/、/version/                              │
│  Docker 编排、状态持久化、端口分配                               │
└──────────────────────────┬──────────────────────────────────┘
                           │ Docker API (go-dockerclient)
┌──────────────────────────▼──────────────────────────────────┐
│                      Docker Engine                           │
│  ┌──────────┐  ┌──────────┐           ┌──────────┐          │
│  │ claw-1   │  │ claw-2   │    ...    │ claw-N   │          │
│  │ XFCE4    │  │ XFCE4    │           │ XFCE4    │          │
│  │ noVNC    │  │ noVNC    │           │ noVNC    │          │
│  │ OpenClaw │  │ OpenClaw │           │ OpenClaw │          │
│  │ Gateway  │  │ Gateway  │           │ Gateway  │          │
│  └──────────┘  └──────────┘           └──────────┘          │
│   :6901/:18789  :6902/:18790           :690N/:1878(8+N)     │
└─────────────────────────────────────────────────────────────┘
```

**依赖规则：** ClawFleet → ClawSandbox（严格单向，不可反向依赖）。

## 3. 组件设计

### 3.1 CLI

**技术栈：** Go 1.25+、Cobra、go-dockerclient。单个静态链接二进制文件，支持 darwin/linux × amd64/arm64。

**命令：**

| 分组 | 命令 | 说明 |
|------|------|------|
| 军团 | `create <N>` | 创建 N 个实例 |
| | `list` | 列出所有实例及状态 |
| | `start <name\|all>` | 启动实例 |
| | `stop <name\|all>` | 停止实例 |
| | `restart <name\|all>` | 重启实例 |
| | `destroy <name\|all> [--purge]` | 销毁实例，可选删除数据 |
| | `desktop <name>` | 在浏览器中打开 noVNC 桌面 |
| | `logs <name> [-f]` | 查看/跟踪容器日志 |
| | `configure <name>` | 交互式配置向导 |
| Dashboard | `dashboard serve` | 启动 Web 服务（前台） |
| | `dashboard start [--host --port]` | 以后台 daemon 启动 |
| | `dashboard stop` | 停止 daemon |
| | `dashboard restart` | 重启 daemon |
| | `dashboard status` | 查看 daemon 状态 |
| 镜像 | `build` | 在本地构建 Docker 镜像 |
| 快照 | `snapshot save <name>` | 归档实例灵魂 |
| | `snapshot list` | 列出已保存的灵魂 |
| | `snapshot delete <name>` | 删除已保存的灵魂 |
| 系统 | `config` | 显示配置文件 |
| | `version` | 输出版本信息 |

### 3.2 Web Dashboard

内嵌的 Preact SPA，由 Go HTTP 服务在端口 8080 上提供。

**REST API（25+ 端点）：**

| 类别 | 端点 | 用途 |
|------|------|------|
| 实例 | `GET/POST /instances`、`POST /{name}/start\|stop`、`DELETE /{name}`、`POST /batch-destroy`、`POST /{name}/configure`、`GET /{name}/configure/status`、`POST /{name}/restart-bot`、`POST /{name}/reset` | 完整实例生命周期 |
| 资产 | `GET/POST/PUT/DELETE /assets/models`、`/assets/channels`、`/assets/characters`、`POST /assets/models/test`、`POST /assets/channels/test` | 模型、渠道、角色 CRUD + 验证 |
| 技能 | `GET /{name}/skills`、`POST /{name}/skills/install`、`DELETE /{name}/skills/{slug}`、`GET /skills/search` | 通过 ClawHub 管理技能 |
| 快照 | `GET/POST/DELETE /snapshots` | 灵魂归档 |
| 镜像 | `GET /image/status`、`POST /image/build`、`POST /image/pull`、`GET /image/openclaw-versions` | 镜像生命周期 + OpenClaw 版本选择器 |

**WebSocket 流：**

| 端点 | 用途 |
|------|------|
| `/ws/stats` | 实时 CPU/内存监控 |
| `/ws/logs/{name}` | 实时容器日志 |
| `/ws/events` | 生命周期事件（创建、启动、停止等） |

**控制台代理：** `/console/{name}` 反向代理到实例的 noVNC 桌面，在 Dashboard 内嵌入桌面访问。Gateway LAN bridge（`0.0.0.0:18790`）使代理能够访问 Gateway UI。

**前端组件（21 个）：** toolbar、sidebar、dashboard、instance-card、instance-desktop、create-dialog、configure-dialog、image-page、logs-viewer、model/channel/character 资产页面和对话框、skills、skill-manager-dialog、snapshots、snapshot-dialog、stats-chart、connection-status、toast。

**国际化：** 支持英文和中文，可在工具栏切换。

### 3.3 Docker 镜像

**仓库地址：** `ghcr.io/clawfleet/clawfleet`

**基础镜像：** `node:22-bookworm`

**分层设计：**

| 层 | 内容 | 大小 |
|---|------|------|
| 1 | 系统包：XFCE4、TigerVNC、noVNC、Chromium、CJK 字体、supervisord | ~800 MB |
| 2 | OpenClaw：`npm install -g openclaw@${OPENCLAW_VERSION}` + 飞书扩展 | ~300 MB |
| 3 | Playwright Chromium：预装在 `/ms-playwright` | ~300 MB |
| 4 | 启动配置：supervisord.conf + entrypoint.sh | <1 MB |
| **总计** | | **~1.4 GB** |

**进程管理（supervisord）：**

| 进程 | 作用 | 用户 | 端口 | 自启动 |
|------|------|------|------|--------|
| xvnc | VNC 服务 + X11 帧缓冲 | node | 5901（内部） | 是 |
| xfce4 | 桌面环境 | node | — | 是 |
| novnc | VNC → WebSocket 代理 | node | 6901（映射到宿主） | 是 |
| openclaw | Gateway（配置后启动） | node | 18789（内部） | 条件启动 |
| gateway-bridge | TCP 代理 18789 → 18790（0.0.0.0） | node | 18790（映射到宿主） | 条件启动 |

**entrypoint.sh：** 创建 `.vnc` 和 `.openclaw` 目录，设置 VNC 密码（如提供 `$VNC_PASSWORD`），若存在 `.configured` 标记则自动启动 OpenClaw，然后启动 supervisord。

### 3.4 资产管理

资产是可分配给实例的共享资源。

**模型资产：** LLM 供应商配置（供应商、API Key、模型名）。支持 Anthropic、OpenAI、Google、DeepSeek 预设模型。保存前通过供应商特定的测试端点验证 API Key。**模型是共享的** — 多个实例可同时使用同一模型。

**渠道资产：** 消息平台配置（Telegram bot token、Discord bot token、Slack webhook、Lark App ID + Secret）。保存前验证凭证。**渠道是独占的** — 每个渠道只能分配给一个实例。

**角色资产：** 人设定义（名称、角色、性格、背景、特点、约束）。渲染为 `SOUL.md` Markdown 并写入实例的 `~/.openclaw/SOUL.md`。Gateway 会在文件变更时热加载。

### 3.5 实例配置

用户在 Dashboard 中点击"配置"时，系统通过 `docker exec` 执行多步配置序列：

1. 设置模型供应商和 API Key（`openclaw config set`）
2. 设置模型名称
3. 设置 DM 和群组策略为 "open"，允许所有发送者
4. 写入渠道配置
5. 渲染并写入 `SOUL.md`（角色 + 花名册）
6. 写入 `.configured` 标记
7. 启动/重启 OpenClaw Gateway 进程

配置状态实时跟踪并报告给前端。

### 3.6 花名册系统

花名册通过向每个实例的 `SOUL.md` 注入团队元数据来实现 bot 间协作。每个 bot 知道团队里有谁、角色是什么、什么时候应该 @对方。

**渲染：** 配置实例时，ClawFleet 收集所有已配置实例的角色数据，构建 `## Your Team` 段落，包含每个队友的名称、角色、渠道和一句话描述，然后追加到 SOUL.md。

**设计原则（提示词即代码）：**
- 明确的判断标准：何时 @队友
- 否定约束：何时不应该提及（如不要提及自己）
- 信息密集、易于扫描：每个队友一行，不做长篇叙述

### 3.7 技能管理

- **内置技能（52 个）：** 随 OpenClaw 一起发布。状态取决于二进制/环境依赖。
- **托管技能：** 通过 `npx clawhub` 安装到 `~/.openclaw/skills/`。
- Dashboard 提供搜索（通过 ClawHub API）、安装和卸载操作。
- ClawHub 有速率限制（~20 次/分钟）— 错误会被优雅处理。

### 3.8 快照系统（灵魂归档）

快照捕获实例的 OpenClaw 数据目录以供后续复用：

- **保存：** 将 `~/.clawfleet/data/<name>/openclaw/` 复制到 `~/.clawfleet/snapshots/<id>/`，剥离 `channels/` 和 `sessions/`（敏感/临时数据）。
- **加载：** 快照可恢复到新实例。
- **元数据：** 名称、来源实例、创建时间戳存储在 `state.json` 中。

### 3.9 端口分配

从配置的基础端口顺序分配：

```
实例       noVNC    Gateway（内部）   Gateway LAN bridge
claw-1     6901     18789            18790（→ 0.0.0.0）
claw-2     6902     18790            18791
claw-N     6900+N   18788+N          18789+N
```

分配前通过 `net.Listen` 探测端口可用性，避免冲突。

### 3.10 状态管理

**状态文件：** `~/.clawfleet/state.json` — 实例、资产和快照的元数据缓存。容器实际状态以 Docker 为准；CLI 每次操作时与 Docker API 对账。

```json
{
  "instances": [{
    "name": "claw-1",
    "container_id": "abc123...",
    "status": "running",
    "ports": { "novnc": 6901, "gateway": 18789 },
    "created_at": "2026-03-30T10:00:00Z",
    "model_asset_id": "anthropic-1",
    "channel_asset_id": "telegram-1",
    "character_asset_id": "alice-1"
  }],
  "model_assets": [...],
  "channel_assets": [...],
  "character_assets": [...],
  "snapshots": [...]
}
```

### 3.11 数据卷

```
~/.clawfleet/
├── config.yaml              # 用户配置
├── state.json               # 实例 + 资产元数据
├── serve.pid                # Dashboard daemon PID
├── logs/                    # Dashboard 日志
├── data/                    # 实例数据（按实例隔离）
│   ├── claw-1/
│   │   └── openclaw/        → 容器内 /home/node/.openclaw
│   │       ├── SOUL.md      # 角色提示词
│   │       ├── openclaw.json
│   │       ├── skills/
│   │       ├── knowledge/
│   │       └── sessions/
│   └── claw-N/
└── snapshots/               # 已保存的灵魂
    └── <id>/
        └── openclaw/
```

容器重启后数据保留。`clawfleet destroy --purge` 可同时删除。

### 3.12 网络设计

- Bridge 网络 `clawfleet-net` 在首次使用时创建
- 容器可通过容器名互相访问（用于实例间通信）
- noVNC 端口绑定到宿主机，用于桌面访问
- Gateway LAN bridge 端口（`18790`）绑定到 `0.0.0.0`，用于控制台代理访问

## 4. 安装与部署

### 4.1 一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/clawfleet/ClawFleet/main/scripts/install.sh | sh
```

**流程：**
1. 检测 OS（macOS/Linux）和架构（amd64/arm64）
2. 确保 Docker 已安装（macOS 用 Colima，Linux 用 Docker Engine）
3. 从 GitHub Releases 下载最新 CLI 二进制文件（含 checksum 校验）
4. 从 GHCR 拉取预构建 Docker 镜像（~1.4 GB）
5. 以后台 daemon 启动 Dashboard
6. 打开浏览器访问 `http://localhost:8080`

**选项：** `--version <tag>`、`--skip-pull`、`--no-daemon`

### 4.2 Daemon 管理

Dashboard 以后台 daemon 运行，按平台管理：

| 平台 | 管理器 | 机制 |
|------|--------|------|
| macOS | launchd | `~/Library/LaunchAgents/com.clawfleet.dashboard.plist` |
| Linux（非 root） | systemd 用户服务 | `~/.config/systemd/user/clawfleet-dashboard.service` |
| Linux（root） | systemd 系统服务 | `/etc/systemd/system/clawfleet-dashboard.service` |
| 回退 | PID 文件 | `~/.clawfleet/serve.pid` |

**默认绑定地址：** macOS 为 `127.0.0.1`（仅本地），Linux 为 `0.0.0.0`（远程访问）。

## 5. 版本管理

### 5.1 ClawFleet 版本

一次 `git tag` 同时锁定 CLI 和 Docker 镜像版本。

```
git tag v0.4.0 && git push origin v0.4.0
        │
        ▼
   GitHub Actions (release.yml)
   ┌──────────────────────┬────────────────────────────────┐
   │  release job          │  docker job                     │
   │  GoReleaser           │  docker/build-push-action       │
   │  CLI 二进制 × 4       │  ghcr.io 镜像（多架构）          │
   │  (darwin/linux         │  :v0.4.0 + :latest             │
   │   × amd64/arm64)      │                                 │
   └──────────┬────────────┴───────────────┬────────────────┘
              ▼                            ▼
       GitHub Release              ghcr.io/clawfleet/clawfleet
```

**版本包（`internal/version/`）：** `Version`、`GitCommit`、`BuildDate` 通过 ldflags 注入。`ImageTag()` 推导 Docker 镜像标签——release 构建（如 `v0.4.0`）使用版本标签，dev 构建回退到 `latest`。

### 5.2 OpenClaw 版本锁定

Docker 镜像内的 OpenClaw 版本是受控的，不依赖构建时 npm 的 `@latest`。

**单一真相源：** `internal/version/version.go`

```go
const RecommendedOpenClawVersion = "2026.3.23-2"
```

**版本流转：**

```
version.go: RecommendedOpenClawVersion = "2026.3.23-2"
        │
        ├──→ CI (release.yml)
        │    提取方式：grep 'RecommendedOpenClawVersion =' version.go
        │    传递方式：OPENCLAW_VERSION build-arg → Docker 构建
        │    结果：预构建 GHCR 镜像包含 openclaw@2026.3.23-2
        │
        ├──→ Dashboard → Build（本地构建）
        │    版本选择器默认值 = RecommendedOpenClawVersion
        │    用户可覆盖为 npm registry 中的任意版本
        │
        └──→ Dashboard → Pull
             拉取预构建 GHCR 镜像（版本已由 CI 内置）
```

**各路径的用户体验：**

| 路径 | OpenClaw 版本 | 由谁决定 |
|------|-------------|---------|
| `install.sh`（一键安装） | `RecommendedOpenClawVersion` | CI build-arg ← `version.go` |
| Dashboard → Pull | 同上 | 同一个预构建镜像 |
| Dashboard → Build（本地） | 用户选择（默认推荐版本） | UI 版本选择器 |

**升级流程：** 当新的 OpenClaw 版本经过测试验证后，更新 `version.go` 中的 `RecommendedOpenClawVersion`，发布新的 ClawFleet release。下次 `install.sh` 或 Dashboard Pull 即可获得新版本。

### 5.3 镜像命名与标签

- **仓库地址：** `ghcr.io/clawfleet/clawfleet`
- **标签：** `:<version>`（如 `:v0.4.0`）+ `:latest`
- **运行时默认标签：** 由 `version.ImageTag()` 决定——release 构建使用版本标签，dev 构建使用 `latest`

### 5.4 自动拉取镜像

当 `clawfleet create` 或 Dashboard 的创建 API 发现本地缺少镜像时，自动尝试从 GHCR 拉取。

## 6. 资源预算

基于 M4 MacBook Air（16 GB RAM，512 GB SSD）测试：

| 资源 | 单个实例 | 3 个实例 | 5 个实例 |
|------|---------|---------|---------|
| 内存（idle） | ~1.5 GB | ~4.5 GB | ~7.5 GB |
| 内存（Chromium 活跃） | ~3 GB | ~9 GB | 不建议 |
| 磁盘（镜像，共享） | 1.4 GB | 1.4 GB | 1.4 GB |
| 磁盘（数据卷/实例） | ~200 MB | ~600 MB | ~1 GB |
| CPU（idle） | <0.5 核 | <1.5 核 | <2.5 核 |

**建议：**
- 16 GB 宿主机：最多 3 个活跃实例（含 Chromium），或 5 个轻载实例
- 默认每容器 `memory_limit=4g`，防止单实例影响宿主
- 可通过 `~/.clawfleet/config.yaml` 调整

## 7. 项目目录结构

```
ClawFleet/
├── cmd/clawfleet/              # 二进制入口
│   └── main.go
├── internal/
│   ├── cli/                    # Cobra 命令（24 个文件）
│   │   ├── root.go             # 根命令，注册子命令
│   │   ├── create.go           # 实例创建
│   │   ├── list.go             # 军团列表
│   │   ├── start/stop/restart/destroy.go
│   │   ├── desktop.go          # 打开 noVNC 桌面
│   │   ├── logs.go             # 容器日志
│   │   ├── configure.go        # 交互式配置向导
│   │   ├── dashboard*.go       # Dashboard serve/start/stop/restart/status
│   │   ├── daemon*.go          # 平台特定 daemon 管理
│   │   ├── snapshot*.go        # 快照 save/list/delete
│   │   ├── build.go            # 镜像构建命令
│   │   ├── config_show.go      # 显示配置文件
│   │   └── version.go          # 版本显示
│   ├── container/              # Docker 编排（8 个文件）
│   │   ├── client.go           # Docker 客户端初始化
│   │   ├── manager.go          # 容器生命周期
│   │   ├── image.go            # 镜像构建/拉取/检查/标签
│   │   ├── configure.go        # 多步 OpenClaw 配置
│   │   ├── network.go          # Docker 网络管理
│   │   ├── skills.go           # 技能安装/卸载
│   │   └── stats.go            # 资源统计采集
│   ├── port/                   # 端口分配器
│   │   └── allocator.go
│   ├── state/                  # JSON 状态持久化
│   │   ├── store.go            # 实例元数据
│   │   ├── assets.go           # 模型/渠道/角色资产
│   │   └── snapshots.go        # 快照元数据
│   ├── config/                 # YAML 配置加载
│   │   └── config.go
│   ├── assets/                 # 内嵌 Docker 构建上下文
│   │   ├── embed.go
│   │   └── docker/
│   │       ├── Dockerfile
│   │       ├── supervisord.conf
│   │       └── entrypoint.sh
│   ├── snapshot/               # 灵魂归档逻辑
│   │   └── snapshot.go
│   ├── version/                # 构建版本信息
│   │   └── version.go          # Version + RecommendedOpenClawVersion
│   └── web/                    # Web Dashboard（15+ 文件）
│       ├── server.go           # HTTP 服务 + PID 管理
│       ├── routes.go           # 路由注册
│       ├── embed.go            # 前端资源内嵌
│       ├── handlers.go         # 实例生命周期处理器
│       ├── handlers_assets.go  # 资产 CRUD
│       ├── handlers_configure.go  # 配置端点
│       ├── handlers_image.go   # 镜像构建/拉取/版本
│       ├── handlers_skills.go  # 技能管理
│       ├── handlers_snapshots.go  # 快照 CRUD
│       ├── handlers_console.go # 控制台代理（反向代理到 noVNC）
│       ├── events.go           # 实时更新事件总线
│       ├── ws_stats.go         # WebSocket：资源统计
│       ├── ws_logs.go          # WebSocket：容器日志
│       ├── ws_events.go        # WebSocket：生命周期事件
│       ├── validate.go         # LLM/渠道凭证验证
│       └── static/             # 内嵌前端
│           ├── index.html
│           ├── css/style.css
│           └── js/
│               ├── app.js      # Preact 主应用
│               ├── api.js      # REST 客户端
│               ├── ws.js       # WebSocket 管理器
│               ├── i18n.js     # 国际化
│               └── components/ # 21 个 Preact 组件
├── scripts/
│   ├── install.sh              # 一键安装脚本
│   └── ensure-go.sh            # Go 版本引导
├── docs/
│   ├── SYSTEM_DESIGN.md
│   ├── SYSTEM_DESIGN.zh-CN.md
│   └── images/                 # 截图
├── growth/                     # 营销物料
├── .github/workflows/
│   └── release.yml             # CI/CD 流水线
├── .goreleaser.yml             # 多平台发布配置
├── Makefile                    # 构建目标
├── CLAUDE.md                   # AI 助手指南
├── ROADMAP.md                  # 产品路线图
├── README.md / README.zh-CN.md
└── LICENSE                     # MIT
```

## 8. 技术依赖

### 宿主机
| 依赖 | 用途 |
|------|------|
| Go 1.25+ | 编译 CLI |
| Docker Engine | 容器运行时（macOS 用 Colima，Linux 用 Docker Engine） |

### 容器内
| 依赖 | 版本 | 用途 |
|------|------|------|
| Debian Bookworm | 12 | 基础 OS |
| Node.js | 22 | OpenClaw 运行时 |
| OpenClaw | 按 release 锁定 | AI 助手核心 |
| Chromium (Playwright) | — | 浏览器自动化 |
| XFCE4 | 4.18 | 轻量桌面环境 |
| TigerVNC | — | VNC 服务端 |
| noVNC + websockify | — | Web VNC 客户端 |
| supervisord | — | 容器内多进程管理 |

### Go 模块
| 模块 | 用途 |
|------|------|
| `github.com/spf13/cobra` | CLI 框架 |
| `github.com/fsouza/go-dockerclient` | Docker Engine API |
| `github.com/gorilla/websocket` | WebSocket 支持 |
| `gopkg.in/yaml.v3` | 配置文件解析 |

## 9. CI/CD 流水线

**触发条件：** 推送匹配 `v*` 的标签（如 `v0.4.0`）

**任务（并行）：**

| 任务 | 工具 | 产出 |
|------|------|------|
| `release` | GoReleaser | 4 个平台的 CLI 二进制（darwin/linux × amd64/arm64）→ GitHub Release + checksums |
| `docker` | docker/build-push-action | 多架构镜像（linux/amd64 + linux/arm64）→ GHCR，版本标签 + `:latest` |

`docker` 任务从 `internal/version/version.go` 中提取 `RecommendedOpenClawVersion`，作为 `OPENCLAW_VERSION` build-arg 传入，确保预构建镜像包含经过测试的 OpenClaw 版本。

**发版流程：**

```bash
# 1. 如需更新 OpenClaw 版本，修改 RecommendedOpenClawVersion
# 2. 打标签并推送
git tag v0.5.0
git push origin v0.5.0
# CI 自动完成：二进制构建、镜像构建推送、GitHub Release 创建
```

## 10. 配置

**文件：** `~/.clawfleet/config.yaml`

```yaml
image:
  name: "ghcr.io/clawfleet/clawfleet"
  tag: "v0.4.0"             # 由 version.ImageTag() 决定

ports:
  novnc_start: 6901         # 顺序分配：6901, 6902, ...
  gateway_start: 18789      # 顺序分配：18789, 18790, ...

resources:
  memory_limit: "4g"        # 每容器内存上限
  cpu_limit: 2.0            # 每容器 CPU 上限（核数）

naming:
  prefix: "claw"            # 实例名：claw-1, claw-2, ...
```

## 11. 验证方案

### 端到端（一键安装）
```bash
# 全新机器
curl -fsSL https://raw.githubusercontent.com/clawfleet/ClawFleet/main/scripts/install.sh | sh
# → Docker 已安装、CLI 已下载、镜像已拉取、Dashboard 运行在 :8080

# 验证镜像内 OpenClaw 版本
docker exec claw-1 npm list -g openclaw
# → 应显示 RecommendedOpenClawVersion
```

### 构建验证
```bash
make build && make test
```

### 手动生命周期
```bash
clawfleet create 2
clawfleet list
clawfleet stop claw-1
clawfleet start claw-1
# → 重启后数据保留
clawfleet destroy claw-2
```

### 资源验证
```bash
docker stats claw-1 claw-2
# → 内存在 memory_limit 限制之内
```
