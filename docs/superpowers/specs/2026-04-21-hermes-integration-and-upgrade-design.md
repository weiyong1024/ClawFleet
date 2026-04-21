# Hermes Integration + Upgrade Flow — Design Spec

> Date: 2026-04-21 | Status: Draft

## Goal

1. ClawFleet manages Hermes containers alongside OpenClaw — unified Fleet, separate runtimes
2. `clawfleet upgrade` CLI command + Dashboard upgrade trigger with observable progress

## Scope

**In scope:**
- Create Hermes instances from Dashboard/CLI (container lifecycle only)
- Expose Hermes native ports (Dashboard 9119, Gateway 3000) for full native experience
- Unified instance list with runtime type badge
- `clawfleet upgrade` command (download binary + pull image + restart daemon)
- Dashboard upgrade banner + trigger + progress log

**Out of scope:**
- ClawFleet-managed Configure for Hermes (user configures via Hermes native Dashboard)
- Skills management for Hermes
- Roster/SOUL.md injection for Hermes
- Cross-runtime collaboration (OpenClaw ↔ Hermes)

---

## Part 1: Hermes Integration

### 1.1 Data Model

**Instance** — add `RuntimeType` field:
```go
type Instance struct {
    // ... existing fields
    RuntimeType string `json:"runtime_type,omitempty"` // "" or "openclaw" = OpenClaw, "hermes" = Hermes
}
```

Empty string = OpenClaw (backward compatible with all existing instances).

**Config** — add Hermes image config:
```go
type HermesConfig struct {
    ImageName string `yaml:"image_name"`
    ImageTag  string `yaml:"image_tag"`
}
```

Default: `nousresearch/hermes-agent:latest`

### 1.2 Container Creation

**Port allocation:** Unified pool (scheme C). Each instance gets ports from the same sequential allocator regardless of runtime.

OpenClaw instance ports:
- noVNC: 6901+N → container 6901
- Gateway LAN bridge: 18789+N → container 18790

Hermes instance ports:
- Dashboard: 9119+N → container 9119
- Gateway: 3000+N → container 3000

The allocator already uses `port.FindAvailable()` which probes availability. The per-runtime port mapping is determined at create time based on `RuntimeType`.

**Volume mapping:**
- OpenClaw: `~/.clawfleet/data/<name>/openclaw` → `/home/node/.openclaw`
- Hermes: `~/.clawfleet/data/<name>/hermes` → `/opt/data`

**Image selection:**
- `RuntimeType == "hermes"` → use `HermesConfig.ImageName:ImageTag`
- Otherwise → use existing `ImageConfig.Name:Tag`

### 1.3 Instance Card (Frontend)

- Badge: 🦞 for OpenClaw, ☤ for Hermes (or text label)
- OpenClaw cards: show Desktop, Control Panel, Configure, Skills, Save Soul, Restart Bot (existing)
- Hermes cards: show **Dashboard** (opens `localhost:{hermes_dashboard_port}`), **Suspend/Resume/Destroy** (lifecycle only)
- Hermes cards do NOT show: Configure, Skills, Save Soul, Restart Bot (these are OpenClaw-specific)

### 1.4 Create Dialog (Frontend)

Add "Runtime" dropdown above instance count:
- OpenClaw (default)
- Hermes

Selection determines which image is used. Soul Archive only available for OpenClaw (Hermes snapshots not supported in v1).

### 1.5 API Changes

**POST /api/v1/instances** — add optional `runtime_type` field:
```json
{"count": 1, "runtime_type": "hermes"}
```

Default: `"openclaw"` (backward compatible).

**GET /api/v1/instances** — response includes `runtime_type` and runtime-specific port names:
```json
{
  "name": "claw-4",
  "runtime_type": "hermes",
  "hermes_dashboard_port": 9119,
  "hermes_gateway_port": 3000
}
```

**OpenClaw-specific endpoints** — return 400 for Hermes instances:
- `POST /instances/{name}/configure`
- `POST /instances/{name}/restart-bot`
- `GET /instances/{name}/skills`
- `POST /instances/{name}/skills/install`
- `DELETE /instances/{name}/skills/{slug}`

Error: `{"error": {"message": "Not available for Hermes instances. Use the Hermes Dashboard to configure."}}`

### 1.6 Hermes Image Management

**No build support for Hermes.** Hermes uses the official `nousresearch/hermes-agent` image from Docker Hub. ClawFleet pulls it, doesn't build it.

**Auto-pull:** Same as OpenClaw — if image is missing at create time, auto-pull from registry.

**Image page:** Show Hermes image status alongside OpenClaw. Read-only (pull only, no build/version selector).

### 1.7 Files Changed

| File | Change |
|------|--------|
| `internal/state/store.go` | `Instance.RuntimeType` field |
| `internal/config/config.go` | `HermesConfig` struct + defaults |
| `internal/container/manager.go` | Runtime-aware port mapping + volume binding |
| `internal/web/handlers.go` | Create handler reads `runtime_type`, rejects OpenClaw-only ops for Hermes |
| `internal/web/handlers_configure.go` | Reject configure for Hermes instances |
| `internal/web/handlers_skills.go` | Reject skills ops for Hermes instances |
| `internal/web/static/js/components/create-dialog.js` | Runtime dropdown |
| `internal/web/static/js/components/instance-card.js` | Runtime badge + conditional buttons |
| `internal/web/static/js/api.js` | Pass `runtime_type` in create |
| `internal/web/static/js/i18n.js` | Runtime labels + Hermes strings |

---

## Part 2: Upgrade Flow

### 2.1 CLI: `clawfleet upgrade`

**Steps:**
1. Check current version (`version.Version`)
2. Fetch latest release from GitHub API (`/repos/clawfleet/ClawFleet/releases/latest`)
3. If already latest → print "Already up to date" and exit
4. Download new binary (same platform/arch detection as install.sh)
5. Verify checksum
6. Replace current binary (atomic: write temp file → rename)
7. Pull new Docker image(s) (OpenClaw + Hermes if configured)
8. Restart Dashboard daemon
9. Print summary: `v1.1.0 → v1.2.0`

**Flags:**
- `clawfleet upgrade` — upgrade to latest
- `clawfleet upgrade --version v1.2.0` — upgrade to specific version
- `clawfleet upgrade --check` — check if upgrade available, don't apply

**Error handling:**
- If binary replacement fails (permissions) → suggest `sudo clawfleet upgrade`
- If image pull fails → warn but continue (image will auto-pull on next create)
- If daemon restart fails → print manual restart command

### 2.2 Dashboard: Upgrade Banner + Trigger

**Version check:** Dashboard periodically checks GitHub API for latest release (on startup + every 6 hours). If newer version exists, show banner at top of page:

```
ClawFleet v1.2.0 available (current: v1.1.0) — [Upgrade Now]
```

**Upgrade trigger:** Click "Upgrade Now" → SSE stream endpoint (same pattern as image build/pull):
- `POST /api/v1/upgrade` — starts upgrade process, streams progress logs
- Progress: "Downloading binary..." → "Verifying checksum..." → "Replacing binary..." → "Pulling images..." → "Restarting daemon..."
- On success: banner changes to "Upgrade complete. Refresh to load new version."
- On failure: show error, suggest CLI fallback

### 2.3 API

**GET /api/v1/upgrade/check** — returns:
```json
{
  "data": {
    "current": "v1.1.0",
    "latest": "v1.2.0",
    "update_available": true,
    "release_url": "https://github.com/clawfleet/ClawFleet/releases/tag/v1.2.0"
  }
}
```

**POST /api/v1/upgrade** — SSE stream, same pattern as `/api/v1/image/build`:
```
data: Downloading clawfleet_1.2.0_darwin_arm64.tar.gz...
data: Verifying checksum...
data: Replacing binary...
data: Pulling OpenClaw image ghcr.io/clawfleet/clawfleet:v1.2.0...
data: Pulling Hermes image nousresearch/hermes-agent:latest...
data: Restarting Dashboard...
event: done
data: Upgraded from v1.1.0 to v1.2.0
```

### 2.4 Self-Upgrade Safety

The binary is replacing itself while potentially running. Safety measures:

1. Write new binary to temp file in same directory (ensures same filesystem)
2. `os.Rename()` for atomic replacement (POSIX guarantees atomicity on same fs)
3. Dashboard daemon restart uses `exec` syscall (new process from new binary)
4. If anything fails before rename → old binary untouched, no corruption

### 2.5 Files Changed

| File | Change |
|------|--------|
| `internal/cli/upgrade.go` | New CLI command |
| `internal/web/handlers_upgrade.go` | New: check + upgrade SSE endpoints |
| `internal/web/routes.go` | Register upgrade routes |
| `internal/web/static/js/components/toolbar.js` | Upgrade banner + button |
| `internal/web/static/js/api.js` | Upgrade API calls |
| `internal/web/static/js/i18n.js` | Upgrade strings |

---

## Implementation Order

1. **Instance RuntimeType** — data model + config (foundation for everything)
2. **Hermes container creation** — manager.go port/volume mapping
3. **API + handler guards** — reject OpenClaw-only ops for Hermes
4. **Frontend** — create dialog runtime dropdown + instance card badge/buttons
5. **Upgrade CLI** — `clawfleet upgrade` command
6. **Upgrade Dashboard** — banner + trigger + SSE progress
7. **Testing** — smoke test both runtimes + upgrade flow

## Verification

- Create OpenClaw instance → all existing features work (zero regression)
- Create Hermes instance → container running, Dashboard port accessible, user can configure via native Hermes Dashboard
- Mixed Fleet → both types in same list, badges correct, buttons appropriate per type
- `clawfleet upgrade --check` → shows available version
- `clawfleet upgrade` → binary replaced, images pulled, daemon restarted
- Dashboard upgrade banner → visible when update available, SSE progress works
