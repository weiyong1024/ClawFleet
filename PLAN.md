# ClawSandbox 规划讨论

## 项目目标

在一台用户机器上，清晰简洁地构建出多个 OpenClaw（龙虾）实例组成的军团。

用户体验对标 LM Studio：通过一套 CLI（或 UI），不需要关心底层技术细节，即可完成龙虾军团的创建和管理。

追求高质量开源项目标准：代码设计遵循良好的工程实践，保持高度的一致性和可读性，目标成为高 star 数的社区项目。

最大化 token 利用率：让用户在可控的开销下，最大化连续使用 LLM 服务商的 token——人休息，AI 员工不休息。这是 ClawSandbox 的核心技术价值。

## 已确定的设计决策

### 1. 军团构建
- 第一版：用户给一个数字 N，系统自动创建 N 个相互隔离的 OpenClaw 实例
- 每个龙虾在用户的 Telegram 通讯录中显示为独立联系人，用户可以分别对话

### 2. 通信模式（第一版）
- 用户在 Telegram 中分别与每个龙虾实例独立对话
- 龙虾之间的协作/群聊作为后续版本的能力

### 3. 虚拟化方案：Docker + 桌面环境 + noVNC
每个龙虾是一个 Docker 容器，内部包含：
- 轻量 Linux 桌面（XFCE）
- noVNC（用户通过浏览器访问该龙虾的桌面）
- OpenClaw 完整运行时（Node.js ≥22、Chromium 等）
- 独立的 OpenClaw Gateway 进程

选择理由：
- **独立桌面**：用户可以登录每个龙虾的桌面去管理 OpenClaw，满足可视化管理需求
- **隔离性好**：网络/文件系统/进程完全隔离，不污染宿主系统
- **维护简单**：Dockerfile 声明式定义，镜像可复用，升级方便
- **用户无感**：ClawSandbox 封装所有 Docker 操作，用户不需要懂 Docker

### 4. 交互界面：分阶段演进

**阶段一（开发调试期）：CLI 优先**
- 所有操作通过 CLI 完成，快速验证核心功能
- 面向开发者和技术用户

**阶段二（开源发布前）：Web 配置页面**
- 在 CLI 跑通后，实现 Web UI 配置页面
- 每个龙虾实例有自己的管理后台（类似 OpenClaw 自带的 Control UI）
- ClawSandbox 提供一个统一的军团管理面板，可以总览和操作所有实例
- 面向所有人，非技术用户无需接触终端即可完成全部配置

CLI 始终保留作为高级用户和自动化场景的入口。

### 5. 用户交互流程（阶段一 CLI 设想）

```
# 创建 3 个龙虾
clawsandbox create 3

# 查看军团状态
clawsandbox list

# 打开某个龙虾的桌面（自动在浏览器中打开 noVNC）
clawsandbox desktop lobster-1

# 停止 / 启动 / 销毁
clawsandbox stop lobster-2
clawsandbox start lobster-2
clawsandbox destroy lobster-1
```

## 架构概览

```
用户机器（宿主）
├── ClawSandbox CLI / UI
│   ├── 管理 Docker 容器生命周期
│   ├── 分配端口（noVNC、Gateway）
│   └── 统一配置分发（API key 等）
│
├── 容器: lobster-1
│   ├── XFCE Desktop + noVNC (:6901)
│   ├── OpenClaw Gateway (:18789)
│   ├── Chromium
│   └── Telegram channel 连接
│
├── 容器: lobster-2
│   ├── XFCE Desktop + noVNC (:6902)
│   ├── OpenClaw Gateway (:18790)
│   ├── Chromium
│   └── Telegram channel 连接
│
└── 容器: lobster-3
    └── ...
```

### 6. 开发语言：Go
- 编译为单个二进制，用户零依赖安装
- Docker 生态原生语言，SDK 一等公民
- CLI 工具事实标准（cobra 框架）

### 7. 目标平台
- 宿主机：macOS（优先，M4 Mac 开发测试）、Linux
- 容器内：Linux（OpenClaw 运行环境）
- 开发机参考配置：M4 MacBook Air, 16GB RAM, 512GB SSD

## 待讨论（遇到时再细化）

- API key 等凭证如何统一管理？共享还是每个实例独立配置？
- Telegram bot 的创建是自动化的还是需要用户手动创建？
