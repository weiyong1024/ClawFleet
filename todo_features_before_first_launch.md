# ClawFleet — 首次宣发前缺口功能清单

> **给 PSE Session 的上下文**：ClawFleet 是一个自托管的 AI 军团 WebUI 管理工具，基于 Docker 容器隔离运行多个 OpenClaw 实例。核心功能（实例生命周期、资产管理、Skills、快照、Dashboard、CLI）已完备。本文档列出 Phase 2 全渠道首发前必须补齐的功能缺口。
>
> **代码库**：`/Users/wy1024/claude_code_project/ClawFleet`（Go 后端 + Preact 前端）
>
> **仓库**：`github.com/clawfleet/ClawFleet`（已从 `weiyong1024/ClawFleet` 迁移至 `clawfleet` 组织）
>
> **关键参考文件**：
> - `CLAUDE.md` — 项目架构和开发规范
> - `ROADMAP.md` — 产品路线图和增长策略
> - `docs/SYSTEM_DESIGN.md` — 系统设计文档
> - `internal/web/static/js/i18n.js` — 国际化（EN + ZH，所有面向用户文案必须双语）
>
> **设计原则**（摘自 memory/design-philosophy.md）：
> - 面向用户的文案一律用结果描述，隐藏技术术语
> - 以乔布斯/乔尼·艾佛为标杆，极简、直觉
> - 每个按钮都必须有存在的理由

---

## 完成进度总览

| # | 功能 | 状态 | 版本 | 完成日期 |
|---|------|------|------|---------|
| 0 | 一键安装脚本 install.sh | ✅ 已完成 | v0.2.0 → v0.3.0 | 2026-03-23 |
| 1 | Roster 最小版 — SOUL.md 团队信息注入 | ✅ 已完成 | v0.4.0 | 2026-03-28 |
| 2 | OpenClaw 版本选择器 | ✅ 已完成 | v0.3.0 | 2026-03-27 |
| 3 | 控制台直通 + 一键重启 bot | ✅ 已完成 | v0.3.0 | 2026-03-25 |
| 4 | 首次使用引导 | ⏸ 暂缓 | — | — |
| 5 | README 补全 | 🔲 待做（部分已完成） | — | — |

---

## 决策记录（已与产品负责人确认）

| 议题 | 决策 | 理由 |
|------|------|------|
| Roster 系统 | 最小版：SOUL.md 注入 | 3-5 天即可支撑杀手级 demo |
| 热升级 | 降级到 Phase 2.5 | 首发用户全新安装，快照系统已能保存进度 |
| UX 改善 | 首次引导优先于创建流程一步化 | "不知道该干嘛"比"步骤太多"更紧急 |
| OpenClaw 版本 | 轻量版本选择器 | 默认推荐版本 + latest + 版本列表，做得简单 |
| 实例管理 | 控制台直通 + 一键重启 bot | 消除最频繁的 noVNC 跳转场景 |
| Ubuntu 镜像 | 首发不做 | Skills 依赖 npm+apt，与桌面环境无关 |
| Docker 自动安装 | macOS 用 Colima，Linux 用 Docker Engine | 避免 Docker Desktop EULA 弹窗，全自动零手动 |
| 镜像仓库 | `ghcr.io/clawfleet/clawfleet` | 仓库已迁移至 clawfleet org |
| 默认 onboard 链路 | Pull 预构建镜像（非 Build） | 更快、确定性更高 |

---

## 必做清单

### 0. ✅ 一键安装脚本 install.sh

**已完成**（PR #41, #42, #43）。最新版本：v0.3.0。

**实现内容**：
- `scripts/install.sh` — 全自动一键部署脚本：自动安装 Docker（macOS: Homebrew + Colima / Linux: get.docker.com）、下载 CLI 二进制、拉取预构建镜像、启动 Dashboard daemon、打开浏览器
- Daemon 管理（`dashboard start/stop/restart/status`）：macOS 用 LaunchAgent（KeepAlive 自动重启）、Linux 用 systemd user service、fallback PID 模式
- Dashboard Image Pull（`POST /api/v1/image/pull`）：SSE 流式进度，Pull 按钮为推荐操作
- CI 多架构 Docker 构建（`linux/amd64` + `linux/arm64`）
- 仓库从 `weiyong1024/ClawFleet` 迁移至 `clawfleet/ClawFleet`，Go module path 更新
- README Quick Start 更新为一行命令
- Wiki Getting Started / CLI Reference / Home 更新

**已知注意事项**：
- Linux 服务器上 Dashboard 默认绑 `0.0.0.0:8080`，macOS 绑 `127.0.0.1`
- 当 GoReleaser 制品版本与二进制内嵌版本不一致时（如复用 rc 制品），`ImageTag()` 返回的 tag 可能与 pull 的镜像 tag 不匹配。正式 release 流程下不存在此问题
- `openclaw@latest` 的 `extensions/feishu` 目录可能不存在，Dockerfile 已做条件处理

---

### 1. ✅ Roster 最小版 — SOUL.md 团队信息注入

**已完成**（2026-03-28，commit `df7e996`）。

**实现内容**：
- `buildRoster()` 函数：Configure 时自动收集其他已配置实例的名字+角色+频道，注入当前实例 SOUL.md 的 "Your Team" 段落
- `refreshTeammateRosters()` 函数：Configure/Destroy/Reset/Start 时同步刷新所有运行中实例的 SOUL.md
- SOUL.md 新增 "How to collaborate" 规则（bounded roundtable pattern）
- Character 热加载：SOUL.md 变更后 OpenClaw Gateway 自动感知，无需重启
- 测试覆盖：3 个单元测试（排除自身、排除未配置实例、包含已停止但有 character 的实例）

**原始目标**：Configure 实例时自动将队友信息注入 SOUL.md，让 bot 知道该 @mention 谁。这是 bot 群内协作 demo 的前置条件（Phase 2 全渠道首发的核心卖点）。

**背景**：当前每个实例完全独立，不知道 fleet 中还有其他 bot。最小版方案是在 Configure 时把其他实例的名字+角色+监听频道写入当前实例 SOUL.md 的 lore 段，让 OpenClaw 自然地在对话中 @mention 队友。Character 系统支持热加载（修改 SOUL.md 后无需重启 gateway）。

**关键文件**：
- `internal/container/configure.go` — SOUL.md 写入点（使用 heredoc 写入容器内文件）
- `internal/state/store.go` — 获取所有实例配置
- `internal/state/assets.go` — 读取 Character 资产（bio/lore/style/topics/adjectives）
- `internal/web/handlers_configure.go` — Configure API handler

**改动**：
1. `configure.go` — 新增 `buildRosterLore()` 函数：遍历其他已配置实例，收集名字+角色+监听频道，生成文本追加到 SOUL.md 的 lore 段
2. `store.go` — 新增 `ListConfiguredInstances()` 返回已配置实例的 model/channel/character 摘要
3. `handlers_configure.go` — Configure 成功后触发其他实例 SOUL.md 刷新（Character 支持热加载，无需重启 gateway）

**验证**：创建 3 实例 → 配置不同 Character + Discord Channel → 检查每个 SOUL.md 包含另外两人信息 → Discord 中测试 @mention 协作

---

### 2. ✅ OpenClaw 版本选择器

**已完成**（2026-03-27 确认）。

**实现内容**：
- `Dockerfile` — `ARG OPENCLAW_VERSION=latest`，Build 时通过 build-arg 传入
- `version.go` — `RecommendedOpenClawVersion = "2026.3.24"` 硬编码推荐版本
- `handlers_image.go` — `GET /api/v1/image/openclaw-versions` 查询 npm registry，返回推荐版本 + 版本列表；Build API 接受 `openclaw_version` 参数（默认推荐版本）
- `image-page.js` — Dashboard 版本下拉框（推荐版本 + latest + npm 版本列表）+ i18n (EN/ZH)
- `build.go` — CLI `--openclaw-version` flag（默认推荐版本）
- `api.js` — 前端 `openclawVersions()` API 调用

---

### 3. ✅ 控制台直通 + 一键重启 bot

**已完成**（PR #44）。最新版本：v0.3.0。

**实现内容**：
- **控制台直通**：Dashboard 反向代理 `/console/{name}/*` 转发到容器内 Gateway Web UI
  - Gateway 保持 loopback 模式（无需 auth/pairing）
  - Node.js TCP bridge（`gateway-bridge`）在 `0.0.0.0:18790` 桥接到 `127.0.0.1:18789`
  - Docker 端口映射从 18790（bridge）而非 18789（gateway），解决 loopback 不可达问题
  - Configure 时设置 `gateway.auth.mode=none`（覆盖 onboard 自动生成的 token）+ `gateway.controlUi.allowedOrigins=["*"]`
  - 实例卡片新增 "⚙ Control Panel" / "⚙ 控制面板" 按钮
- **一键重启 bot**：`POST /api/v1/instances/{name}/restart-bot` 调用 `supervisorctl restart openclaw`
  - 实例卡片新增 "🔄 Restart Bot" / "🔄 重启龙虾" 按钮，带确认弹框
- **i18n**：EN + ZH 全覆盖

**已知注意事项**：
- 远程 Linux 服务器上，Control Panel 的 WebSocket 需要浏览器安全上下文（HTTPS 或 localhost）。通过 HTTP 访问远程 IP 时需要 SSH 隧道：`ssh -L 8080:127.0.0.1:8080 user@server`。其他 Dashboard 功能不受影响。已在 README 中说明。
- 已有实例需要重新 Configure 才会设置 `auth.mode=none` + `allowedOrigins`。新建实例自动配置。

---

### 4. ⏸ 首次使用引导 — 暂缓

**决策**：产品负责人决定暂缓此项。理由：参考乔布斯设计理念——产品应该简洁优雅到用户拿到就知道怎么用，需要说明书的产品体验一定不完美。应通过优化产品本身的直觉性来解决上手问题，而非添加引导层。

如果后续种子用户反馈确实存在上手困难，再重新评估。

---

### 5. 🔲 README 补全（1 天）

**部分已完成**：Quick Start 已更新为一行命令，Linux 服务器部署说明已添加。

**剩余改动**：
1. ~~Quick Start 改为一行安装命令~~ ✅ 已完成
2. 添加 Troubleshooting 章节（Docker 未运行 / 镜像 build 失败 / 端口被占 / OpenClaw 版本不兼容）
3. Demo 视频占位替换为实际链接（Phase 1 视频完成后）
4. 验证所有 `docs/images/` 截图路径有效

---

## 次优先级（时间允许再做）

| 功能 | 关键文件 | 工作量 |
|------|----------|--------|
| 创建实例一步化（创建时选 Model/Channel/Character） | `create-dialog.js` + `handlers.go` | 2-3 天 |
| Reset 功能暴露到 Dashboard（API 已有，缺 UI） | `dashboard.js` | 半天 |
| 凭证在线测试按钮（保存前验证 API Key） | assets 相关组件 | 1-2 天 |
| 基础集成测试（当前仅 4 个单元测试文件） | 新增 `*_test.go` | 3-5 天 |
| Dashboard HTTPS（自签证书） | `internal/web/server.go` | 1-2 天 |

## 明确不做（首发后再评估）

| 功能 | 理由 |
|------|------|
| Ubuntu Desktop 镜像 | Skills 兼容性无差异（都是 npm+apt），资源代价大（+70-110% 镜像体积），与"浏览器管理"叙事矛盾 |
| 版本热升级 | 首发用户全新安装，快照系统已能保存进度 |
| Web Terminal (xterm.js) | 控制台直通已覆盖主要场景 |
| 仓库拆分（ClawSandbox 独立） | Phase 2.5 明确规划 |

---

## 时间线

| 周次 | 内容 | 产出 |
|------|------|------|
| 第 1 周 | ✅ 一键安装脚本 + ✅ 控制台直通 + 重启 bot | `curl ... \| sh` 可用 + 控制面板直通 |
| 第 2 周 | Roster 最小版 + OpenClaw 版本选择器 | bot 协作可 demo + build 稳定 |
| 第 3 周 | 首次引导 + README 补全 | 新用户体验闭环 |
| 第 4 周起 | Phase 0.5 种子用户 + Phase 1 物料 | 安装脚本跨国测试可并行 |

## 验证方式

每个功能完成后：
1. 手动冒烟测试：全新状态 → `curl` 安装 → 引导 → 创建 3 bot → 配置 → 控制面板直通 → 重启 bot → Discord 协作
2. `make test` 无回归
3. Dashboard 浏览器验证（Chrome + Safari）
4. README 在 GitHub 预览确认渲染
