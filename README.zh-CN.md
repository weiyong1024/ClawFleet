# ClawSandbox

> 在单台机器上部署和管理多个隔离的 [OpenClaw](https://github.com/openclaw/openclaw) 实例。

[English](./README.md)

---

## 背景

LLM AI 的应用层发展分三个阶段：

1. **ChatBot** — 让人人获取知识
2. **Agent** — 让人人成为专业的人
3. **OpenClaw** — 让人人成为管理者

OpenClaw 是一个自托管的个人 AI 助手，能连接 WhatsApp、Telegram、Slack 等 20+ 消息平台。ClawSandbox 的目标是消除部署瓶颈——不应该让 OpenClaw 单体的获取成为卡点，而应该能轻松部署一个 OpenClaw 军团。

## ClawSandbox 能做什么

- **一条命令部署军团** — 给一个数字，就能得到对应数量的隔离 OpenClaw 实例
- **每个实例独立桌面** — 每只龙虾运行在独立的 Docker 容器中，内含 XFCE 桌面，通过 noVNC 在任意浏览器中访问
- **生命周期管理** — 用简单的 CLI 命令创建、启动、停止、销毁实例
- **数据持久化** — 每个实例的数据在容器重启后保留
- **资源隔离** — 实例之间以及与宿主系统之间完全隔离

## 前置要求

- macOS 或 Linux
- 已安装并运行 [Docker Desktop](https://www.docker.com/products/docker-desktop/)

## 快速开始

### 1. 安装

```bash
git clone https://github.com/weiyong1024/ClawSandbox.git
cd ClawSandbox
make build
# 可选：安装到系统 PATH
sudo make install
```

### 2. 构建 Docker 镜像

首次使用需要执行一次，下载约 1.4 GB，需要几分钟：

```bash
clawsandbox build
```

### 3. 部署龙虾军团

```bash
# 创建 3 个隔离的 OpenClaw 实例
clawsandbox create 3

# 查看状态
clawsandbox list
```

### 4. 配置每只龙虾

每只龙虾需要通过其桌面完成一次初始化：

```bash
# 在浏览器中打开 lobster-1 的 noVNC 桌面
clawsandbox desktop lobster-1
```

在桌面的终端中执行：

```bash
# 第一步：运行初始化向导（配置 LLM API Key、Telegram Bot 等）
openclaw onboard --flow quickstart

# 第二步：启动 Gateway
openclaw gateway --port 18789
```

Gateway 启动后，在桌面的 **Chromium 浏览器**中访问终端输出的地址（形如 `http://127.0.0.1:18789/#token=...`），即可打开 OpenClaw 控制台。

## CLI 命令

```bash
clawsandbox build                       # 构建 Docker 镜像
clawsandbox create <N>                  # 创建 N 个龙虾实例
clawsandbox list                        # 列出所有实例及状态
clawsandbox desktop <name>              # 在浏览器中打开实例桌面
clawsandbox start <name|all>            # 启动已停止的实例
clawsandbox stop <name|all>             # 停止运行中的实例
clawsandbox logs <name> [-f]            # 查看实例日志
clawsandbox destroy <name|all>          # 销毁实例（默认保留数据）
clawsandbox destroy --purge <name|all>  # 销毁实例并删除数据
```

## 资源占用参考

测试环境：M4 MacBook Air（16 GB 内存）

| 实例数 | 内存（空闲） | 内存（Chromium 活跃） |
|--------|-------------|----------------------|
| 1      | ~1.5 GB     | ~3 GB                |
| 3      | ~4.5 GB     | ~9 GB                |
| 5      | ~7.5 GB     | 不建议               |

## 项目状态

正在积极开发中。CLI 功能已可用，Web UI 计划在后续版本推出。

欢迎提 Issue 或 PR 参与贡献。

## License

MIT
