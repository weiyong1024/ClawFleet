# ClawSandbox System Design Document

> 版本: v0.1 (Draft)
> 日期: 2026-03-07

## 1. 概述

ClawSandbox 是一个 CLI 工具，在单台宿主机上批量创建和管理多个隔离的 OpenClaw 实例。每个实例运行在独立的 Docker 容器中，拥有完整的 Linux 桌面环境，用户可通过浏览器访问桌面、通过 Telegram 与各实例独立对话。

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      宿主机 (macOS / Linux)                  │
│                                                             │
│  ┌───────────────────────────────────┐                      │
│  │        ClawSandbox CLI (Go)       │                      │
│  │  ┌─────────┬──────────┬────────┐  │                      │
│  │  │ create  │  list    │ desktop│  │                      │
│  │  │ start   │  stop    │destroy │  │                      │
│  │  └─────────┴──────────┴────────┘  │                      │
│  │  ┌─────────────────────────────┐  │                      │
│  │  │   Docker Engine SDK (Go)    │  │                      │
│  │  └──────────────┬──────────────┘  │                      │
│  │                 │                 │                      │
│  │  ┌──────────────┴──────────────┐  │                      │
│  │  │     Port Allocator          │  │                      │
│  │  │  noVNC: 6901, 6902, ...     │  │                      │
│  │  │  Gateway: 18789, 18790, ... │  │                      │
│  │  └─────────────────────────────┘  │                      │
│  │  ┌─────────────────────────────┐  │                      │
│  │  │     Instance State Store    │  │                      │
│  │  │  ~/.clawsandbox/state.json  │  │                      │
│  │  └─────────────────────────────┘  │                      │
│  └───────────────────────────────────┘                      │
│                     │ Docker API                            │
│    ┌────────────────┼────────────────────────┐              │
│    │                │                        │              │
│    ▼                ▼                        ▼              │
│  ┌──────────┐  ┌──────────┐           ┌──────────┐         │
│  │lobster-1 │  │lobster-2 │    ...    │lobster-N │         │
│  │          │  │          │           │          │         │
│  │ XFCE     │  │ XFCE     │           │ XFCE     │         │
│  │ noVNC    │  │ noVNC    │           │ noVNC    │         │
│  │ Node≥22  │  │ Node≥22  │           │ Node≥22  │         │
│  │ Chromium │  │ Chromium │           │ Chromium │         │
│  │ OpenClaw │  │ OpenClaw │           │ OpenClaw │         │
│  │ Gateway  │  │ Gateway  │           │ Gateway  │         │
│  └──────────┘  └──────────┘           └──────────┘         │
│   :6901/:18789  :6902/:18790           :690N/:1878(8+N)    │
└─────────────────────────────────────────────────────────────┘
```

## 3. 组件详细设计

### 3.1 ClawSandbox CLI

**技术栈：**
- 语言：Go 1.22+
- CLI 框架：[cobra](https://github.com/spf13/cobra)
- Docker 交互：[Docker Engine SDK for Go](https://pkg.go.dev/github.com/docker/docker/client)
- 编译产物：单个静态链接二进制文件（darwin/arm64, darwin/amd64, linux/amd64, linux/arm64）

**命令设计：**

```
clawsandbox create <N>          # 创建 N 个龙虾实例
clawsandbox list                # 列出所有实例及状态
clawsandbox start <name|all>    # 启动实例
clawsandbox stop <name|all>     # 停止实例
clawsandbox destroy <name|all>  # 销毁实例（容器+数据）
clawsandbox desktop <name>      # 在浏览器中打开该实例的 noVNC 桌面
clawsandbox logs <name>         # 查看实例日志
clawsandbox status              # 显示军团概览（实例数、资源占用等）
```

**配置文件：** `~/.clawsandbox/config.yaml`

```yaml
# 基础镜像配置
image:
  name: "clawsandbox/openclaw"
  tag: "latest"

# 端口范围
ports:
  novnc_start: 6901      # noVNC 起始端口，递增分配
  gateway_start: 18789   # Gateway 起始端口，递增分配

# 容器资源限制
resources:
  memory_limit: "4g"     # 每个容器内存上限
  cpu_limit: "2.0"       # 每个容器 CPU 上限（核数）

# 实例命名前缀
naming:
  prefix: "lobster"      # 实例名: lobster-1, lobster-2, ...
```

### 3.2 Docker 镜像 (clawsandbox/openclaw)

**基础镜像选择：** `node:22-bookworm`（与 OpenClaw 官方 Docker 方案一致）

**镜像分层设计：**

```dockerfile
# Layer 1: 基础系统 + 桌面环境
FROM node:22-bookworm

# 系统依赖 + XFCE 桌面 + VNC/noVNC
RUN apt-get update && apt-get install -y \
    xfce4 xfce4-terminal \
    tigervnc-standalone-server \
    novnc websockify \
    dbus-x11 \
    # Chromium 及 Playwright 依赖
    chromium \
    fonts-noto-cjk \
    # 工具
    supervisor curl wget \
    && rm -rf /var/lib/apt/lists/*

# Layer 2: OpenClaw 安装
RUN npm install -g openclaw@latest

# Layer 3: Playwright Chromium（构建时安装，避免运行时下载）
ENV PLAYWRIGHT_BROWSERS_PATH=/home/node/.cache/ms-playwright
RUN npx playwright install chromium

# Layer 4: 启动配置
COPY supervisord.conf /etc/supervisor/conf.d/
COPY entrypoint.sh /entrypoint.sh

USER node
EXPOSE 6901 18789

ENTRYPOINT ["/entrypoint.sh"]
```

**进程管理 (supervisord)：**

容器内需要同时运行多个进程，使用 supervisord 统一管理：

| 进程 | 作用 | 端口 |
|------|------|------|
| Xvfb | 虚拟帧缓冲（虚拟显示器） | DISPLAY :1 |
| XFCE4 | 桌面环境 | — |
| VNC Server | VNC 协议服务 | 5901 (容器内) |
| noVNC + websockify | VNC → WebSocket 代理 | 6901 (映射到宿主) |
| OpenClaw Gateway | OpenClaw 核心服务 | 18789 (映射到宿主) |

**entrypoint.sh 职责：**
1. 初始化 VNC 密码（可由环境变量注入或使用默认值）
2. 如果 `~/.openclaw` 不存在，执行首次初始化
3. 启动 supervisord

**预估镜像大小：**
- 基础系统 + XFCE + VNC/noVNC: ~800MB
- Node.js 22 + OpenClaw: ~300MB
- Chromium (Playwright): ~300MB
- **总计: ~1.4GB**

### 3.3 端口分配策略

CLI 维护一个简单的端口分配器，避免冲突：

```
实例       noVNC 端口     Gateway 端口
lobster-1  6901          18789
lobster-2  6902          18790
lobster-3  6903          18791
...
lobster-N  6900+N        18788+N
```

分配逻辑：
- 从配置的起始端口开始递增
- 创建前检测端口是否被占用（`net.Listen` 探测）
- 分配结果记录在状态文件中

### 3.4 状态管理

**状态文件：** `~/.clawsandbox/state.json`

```json
{
  "instances": [
    {
      "name": "lobster-1",
      "container_id": "abc123...",
      "status": "running",
      "ports": {
        "novnc": 6901,
        "gateway": 18789
      },
      "created_at": "2026-03-07T10:00:00Z"
    }
  ]
}
```

状态文件是 CLI 的元数据缓存，真实状态以 Docker 容器状态为准。CLI 每次操作时与 Docker API 对账，修正不一致。

### 3.5 数据卷设计

每个容器挂载独立的宿主目录，实现数据持久化和隔离：

```
~/.clawsandbox/data/
├── lobster-1/
│   └── openclaw/     → 容器内 /home/node/.openclaw
├── lobster-2/
│   └── openclaw/     → 容器内 /home/node/.openclaw
└── lobster-3/
    └── openclaw/     → 容器内 /home/node/.openclaw
```

- 容器销毁后数据默认保留在宿主机上
- `clawsandbox destroy --purge` 可同时删除数据
- 目录权限设为 uid 1000（node 用户），与容器内用户一致

### 3.6 网络设计

**Docker 网络：**
- 创建专用 bridge 网络 `clawsandbox-net`
- 各容器通过容器名互相可达（为后续龙虾间通信预留）
- 每个容器仅暴露 noVNC 和 Gateway 端口到宿主机

```
clawsandbox-net (bridge)
├── lobster-1 (172.20.0.2)
├── lobster-2 (172.20.0.3)
└── lobster-3 (172.20.0.4)
```

**宿主机访问：**
- noVNC: `http://localhost:6901` → lobster-1 桌面
- Gateway: `ws://localhost:18789` → lobster-1 Gateway（内部用，用户一般不直接访问）

## 4. 用户操作流程

### 4.1 首次使用

```
# 1. 安装 ClawSandbox（下载单个二进制文件）
curl -fsSL https://get.clawsandbox.dev | sh

# 2. 创建 3 个龙虾
clawsandbox create 3
# → 拉取镜像（首次，约 1.4GB）
# → 创建 Docker 网络
# → 启动 3 个容器
# → 输出:
#   lobster-1  ✓  desktop: http://localhost:6901
#   lobster-2  ✓  desktop: http://localhost:6902
#   lobster-3  ✓  desktop: http://localhost:6903

# 3. 打开龙虾桌面，进入 OpenClaw onboard 配置 Telegram 等
clawsandbox desktop lobster-1
# → 自动在浏览器中打开 http://localhost:6901
```

### 4.2 日常使用

```
# 查看军团状态
clawsandbox list
# NAME        STATUS    DESKTOP              UPTIME
# lobster-1   running   http://localhost:6901 2d 3h
# lobster-2   running   http://localhost:6902 2d 3h
# lobster-3   stopped   -                    -

# 启动/停止
clawsandbox start lobster-3
clawsandbox stop lobster-1

# 查看日志
clawsandbox logs lobster-1
```

### 4.3 OpenClaw Onboard 流程（容器内）

用户通过 noVNC 进入桌面后，在终端中执行：

```bash
openclaw onboard --install-daemon
openclaw gateway --port 18789 --verbose
```

onboard 向导引导用户完成：
1. 选择 LLM Provider（OpenAI/Anthropic 等）并输入 API Key
2. 配置 Telegram Bot（输入从 @BotFather 获取的 bot token）
3. 选择默认模型
4. 启动 Gateway

**后续优化方向（非第一版）：** 将 onboard 流程通过 ClawSandbox CLI 自动化，用户只需提供 API key 和 Telegram bot token，无需手动进入桌面操作。

## 5. 资源预算

基于 M4 MacBook Air (16GB RAM, 512GB SSD) 的测试环境：

| 资源 | 单个实例 | 3 个实例 | 5 个实例 |
|------|---------|---------|---------|
| 内存（idle） | ~1.5GB | ~4.5GB | ~7.5GB |
| 内存（Chromium 活跃） | ~3GB | ~9GB | 不建议 |
| 磁盘（镜像） | 1.4GB（共享） | 1.4GB | 1.4GB |
| 磁盘（数据卷/实例） | ~200MB | ~600MB | ~1GB |
| CPU（idle） | <0.5 核 | <1.5 核 | <2.5 核 |

**建议：**
- 16GB 宿主机：最多同时运行 3 个活跃实例（含 Chromium），或 5 个轻载实例
- 默认每容器 memory_limit=4GB，防止单实例内存泄漏影响宿主
- 用户可通过配置调整资源限制

## 6. 项目目录结构

```
ClawSandbox/
├── cmd/                        # CLI 入口
│   └── clawsandbox/
│       └── main.go
├── internal/
│   ├── cli/                    # cobra 命令定义
│   │   ├── root.go
│   │   ├── create.go
│   │   ├── list.go
│   │   ├── start.go
│   │   ├── stop.go
│   │   ├── destroy.go
│   │   ├── desktop.go
│   │   ├── logs.go
│   │   └── status.go
│   ├── container/              # Docker 容器生命周期管理
│   │   ├── manager.go          # 创建/启动/停止/删除容器
│   │   ├── image.go            # 镜像构建/拉取
│   │   └── network.go          # Docker 网络管理
│   ├── port/                   # 端口分配
│   │   └── allocator.go
│   ├── state/                  # 实例状态管理
│   │   └── store.go
│   └── config/                 # 配置文件解析
│       └── config.go
├── docker/                     # Docker 镜像相关文件
│   ├── Dockerfile
│   ├── supervisord.conf
│   └── entrypoint.sh
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── PLAN.md
└── CLAUDE.md
```

## 7. 技术依赖总览

### 宿主机（ClawSandbox CLI）
| 依赖 | 用途 |
|------|------|
| Go 1.22+ | 编译 CLI |
| Docker Engine | 容器运行时（Docker Desktop for Mac 或 colima） |

### 容器内
| 依赖 | 版本 | 用途 |
|------|------|------|
| Debian Bookworm | 12 | 基础 OS |
| Node.js | ≥22 | OpenClaw 运行时 |
| OpenClaw | latest | AI 助手核心 |
| Chromium (Playwright) | — | 浏览器自动化 |
| XFCE4 | 4.18 | 轻量桌面环境 |
| TigerVNC | — | VNC 服务端 |
| noVNC + websockify | — | Web VNC 客户端 |
| supervisord | — | 容器内多进程管理 |

### Go 模块依赖
| 模块 | 用途 |
|------|------|
| `github.com/spf13/cobra` | CLI 框架 |
| `github.com/docker/docker/client` | Docker Engine API |
| `github.com/docker/go-connections` | Docker 辅助工具 |
| `gopkg.in/yaml.v3` | 配置文件解析 |

## 8. 交互界面演进策略

项目采用分阶段交付策略，先跑通核心再打磨体验：

### 阶段一：CLI（开发调试期）
所有操作通过命令行完成，快速验证核心功能闭环。这一阶段的产出是一个功能完整、可流畅使用的 CLI 工具。

### 阶段二：Web 配置页面（开源发布前）
在 CLI 跑通后，为非技术用户实现 Web UI：

- **军团管理面板**：ClawSandbox 在宿主机启动一个 Web 服务，提供统一的军团总览页面（创建/启停/销毁实例、查看状态）
- **实例管理后台**：每个龙虾实例有独立的配置页面（LLM Provider、Telegram Bot、模型选择等），用户无需进入桌面终端即可完成 OpenClaw 配置
- **桌面入口**：页面内嵌或链接到各实例的 noVNC 桌面

CLI 始终保留作为高级用户和自动化场景的入口。

## 9. 实现范围

### 阶段一 MVP（CLI）

**包含：**
- [x] Docker 镜像构建（Dockerfile + supervisord + entrypoint）
- [x] `create N` — 创建 N 个容器实例
- [x] `list` — 列出所有实例及状态
- [x] `start/stop` — 启停实例
- [x] `destroy` — 销毁实例
- [x] `desktop` — 浏览器打开 noVNC
- [x] 端口自动分配和冲突检测
- [x] 实例数据卷持久化

### 阶段二（Web UI）
- [ ] 军团管理面板（Web 服务）
- [ ] 实例配置页面（LLM/Telegram 等）
- [ ] OpenClaw onboard 自动化

**不包含（更远期）：**
- 龙虾间群聊和任务分派
- 实例模板/预设配置
- 远程宿主机支持

## 10. 验证方案

### 构建验证
```bash
# 编译 CLI
make build

# 构建 Docker 镜像
docker build -t clawsandbox/openclaw:latest -f docker/Dockerfile .
```

### 端到端验证
```bash
# 1. 创建 2 个实例
clawsandbox create 2

# 2. 确认容器运行中
clawsandbox list
docker ps | grep lobster

# 3. 访问桌面
clawsandbox desktop lobster-1
# → 浏览器打开 http://localhost:6901，看到 XFCE 桌面

# 4. 在桌面终端中验证 OpenClaw
#    打开终端 → 执行 openclaw --version → 确认可用

# 5. 停止/启动
clawsandbox stop lobster-1
clawsandbox start lobster-1
# → 确认数据（~/.openclaw）在容器重启后保留

# 6. 销毁
clawsandbox destroy lobster-2
clawsandbox list
# → 确认 lobster-2 已移除
```

### 资源验证
```bash
# 监控容器资源使用
docker stats lobster-1 lobster-2
# → 确认内存在 memory_limit 之内
```
