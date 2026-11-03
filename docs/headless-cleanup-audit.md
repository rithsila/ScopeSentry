# ScopeSentry Headless Cleanup & Architecture Audit

**Date:** 2026-11-03  
**Status:** DRAFT / PENDING APPROVAL  
**Scope:** Post-headless architecture audit, dead code catalog, optional feature matrix, and staged removal plan.

---

## 1. Executive Summary

ScopeSentry has successfully excised its Vue 3 Single Page Application (SPA) and embedded static assets, transitioning to a headless architecture oriented around:
1. A long-running Go (Gin) scanning server.
2. An embedded Model Context Protocol (MCP) server (`/mcp`) exposing 17 security scanning tools with English schemas.
3. A headless admin CLI (`./ScopeSentry apikey ...`, `./ScopeSentry pluginkey`, `./ScopeSentry version`).
4. A lightweight, standalone CLI client binary (`scopesentry`) strictly decoupled from backend dependencies.

This audit maps all active execution paths, identifies confirmed dead code remaining from the legacy web UI era, classifies optional components, and defines a verification plan before any files are deleted.

---

## 2. Architecture & Call Paths

```
+-----------------------------------------------------------------------------------+
|                                  CLIENT LAYER                                     |
|                                                                                   |
|  +-----------------------------+          +------------------------------------+  |
|  | CLI Client (cmd/scopesentry)|          | External AI Agents / Tools (MCP)   |  |
|  | - Standalone binary (11.9MB)|          | - Claude Desktop / Custom Agents   |  |
|  | - Zero backend imports      |          | - JSON-RPC over Streamable HTTP    |  |
|  +--------------+--------------+          +-----------------+------------------+  |
|                 |                                           |                     |
+-----------------|-------------------------------------------|---------------------+
                  |                                           |
                  +---------------------+---------------------+
                                        | (HTTP /mcp with X-API-Key / Bearer)
                                        v
+-----------------------------------------------------------------------------------+
|                              HEADLESS SERVER LAYER                                |
|                                                                                   |
|  +-----------------------------------------------------------------------------+  |
|  | Entry: cmd/main/main.go & cmd/main/admin.go (68.3MB)                        |  |
|  | - Subcommands: apikey, pluginkey, version (bypasses server boot)            |  |
|  | - Router: internal/api/routes/routes.go                                     |  |
|  +-----------------------------------------------------------------------------+  |
|                                       |                                           |
|         +-----------------------------+-----------------------------+             |
|         |                             |                             |             |
|         v                             v                             v             |
|  +---------------+             +---------------+             +---------------+    |
|  | MCP Server    |             | REST API      |             | System &      |    |
|  | /mcp          |             | /api/*        |             | Health Routes |    |
|  | (17 tools)    |             | - auth, tasks |             | - /live       |    |
|  | internal/mcp/ |             | - assets, node|             | - /ready      |    |
|  +-------+-------+             +-------+-------+             | - /images     |    |
|          |                             |                     | - /api/swagger|    |
|          |                             |                     +---------------+    |
+----------|-----------------------------|------------------------------------------+
           |                             |
           +--------------+--------------+
                          |
                          v
+-----------------------------------------------------------------------------------+
|                              CORE ENGINE & WORKERS                                |
|                                                                                   |
|  +-----------------------------------------------------------------------------+  |
|  | Schedulers & Workers: internal/scheduler/, internal/worker/                 |  |
|  | Database Layer: internal/database/mongodb, internal/database/redis           |  |
|  | Plugin Engine: Yaegi runtime (internal/plugins/, internal/symbols/)         |  |
|  | CT Monitoring: internal/utils/ct/                                           |  |
|  +-----------------------------------------------------------------------------+  |
+-----------------------------------------------------------------------------------+
```

---

## 3. Component Classification Matrix

Components are categorized into four explicit buckets:
- **KEEP**: Required for headless server, MCP, CLI, database, or scanning worker operation.
- **REMOVE**: Confirmed dead code, orphaned packages, commented-out stubs, or obsolete assets.
- **OPTIONAL**: Active functional features not strictly needed for headless agent workflows.
- **DEFECT FIX**: Pre-existing compiler/vet bugs preventing full `go test ./...` execution.

### 3.1 KEEP (Core Operational Components)

| Path | Purpose | Why Retained |
|---|---|---|
| `internal/mcp/` | MCP server and 17 scanning tool handlers | Primary interface for agent automation. |
| `cmd/scopesentry/` | Standalone CLI client | Dedicated headless CLI client. |
| `cmd/main/admin.go` | Local administrative CLI | Headless API key & plugin key management. |
| `internal/api/handlers/user/` (`Login`, `ChangePassword`) | User authentication | Required for JWT auth and node communication. |
| `internal/worker/` | Distributed scan worker coordinator | Core vulnerability and asset scanning engine. |
| `internal/scheduler/` | Cron-like recurring scan scheduler | Scheduled asset discovery jobs. |
| `internal/database/mongodb/`, `redis/` | Database and message queue drivers | Primary data store and task distribution. |
| `internal/plugins/`, `internal/symbols/` | Yaegi runtime plugin engine & symbol tables | Dynamic plugin execution for scanners. |
| `internal/utils/ct/` | Certificate Transparency watcher | Subdomain discovery via CT logs. |
| `internal/api/handlers/configuration/` | Configuration handlers (rad, subfinder, deduplication) | Scanning tool configuration management. |
| Static `/images` (`GlobalConfig.System.ImgDir`) | Scan evidence screenshot hosting | Required for scan report image evidence. |
| Health `/live`, `/ready` | HTTP liveness and readiness probes | Docker, Kubernetes, and CLI health monitoring. |

### 3.2 REMOVE (Confirmed Dead Code & Legacy Assets)

| Target Path / File | Type | Proof / Evidence | Impact of Removal |
|---|---|---|---|
| `internal/api/routes/role/role.go` | Route registration | 100% commented-out. Unreferenced in `routes.go`. | Zero impact. Clean deletion. |
| `internal/api/handlers/role/role.go` | API handlers | 100% commented-out (`Create`, `List`, `Get`, `Update`, `Delete`). | Zero impact. Clean deletion. |
| `internal/repositories/role/role.go` | Database repository | Zero importers in the entire codebase. No `internal/services/role` exists. | Zero impact. Clean deletion. |
| `internal/models/role.go` | Model structs | Used only by unused role repository and Yaegi symbol table. | Remove model & symbol entry. |
| `pkg/response/response.go` | Response helper | Duplicate package with 0 callers. Entire app uses `internal/api/response`. | Zero impact. `pkg/` becomes empty. |
| `internal/models/miniprogram.go` | Duplicate model | Unused duplicate `MiniProgram` struct. Active model is `MPResult` in `internal/models/mp.go`. | Update symbol table; zero runtime impact. |
| `internal/api/middleware/cors.go` | Middleware | Unregistered in Gin router. Headless architecture has no browser clients. | Zero impact. Clean deletion. |
| `internal/api/handlers/user/user.go` (lines 22-34, 103-249) | Commented handlers & unused structs | Commented-out `Register`, `List`, `Get`, `Update`, `Delete` handlers; unused request structs. | Leaves active `Login` & `ChangePassword`. |
| `internal/api/routes/user/user.go` (lines 29-61) | Commented routes | Commented-out route definitions for CRUD user endpoints. | Leaves active `/login` & `/changePassword`. |
| `docs/images/` (31 MB) | Obsolete assets | 31MB of browser UI screenshots from excised Vue SPA referenced only in old READMEs. | Massive repository weight reduction. |

### 3.3 OPTIONAL (Active Features for User Decision)

| Feature | Files / Paths | Description & Agent Context | Trade-off / Decision |
|---|---|---|---|
| **Swagger UI** | `docs/docs.go`, `docs/swagger.*`, route in `cmd/main/main.go` | Web UI at `/api/swagger/*any` documenting REST endpoints. | Agents use MCP schema, not Swagger. Can be retained or removed to drop ~5MB build size. |
| **Excel Export** | `internal/services/export/`, `internal/api/routes/export/`, `internal/api/handlers/export/` | Exports scan assets to `.xlsx` using `github.com/xuri/excelize/v2`. | Agents consume JSON via MCP. Excel export is primarily for human spreadsheet reporting. |
| **Remote Version Check** | `internal/update/`, `internal/api/handlers/system/system.go` | Queries GitHub API (`api.github.com/repos/Autumn-27/ScopeSentry/releases/latest`) on startup. | Blocks or logs warnings in air-gapped / isolated scanner environments. |
| **Notification Webhooks** | `internal/config/notification.go`, `internal/api/handlers/configuration/notification.go` | Sends scan completion alerts to DingTalk, Feishu, Telegram, Email, Webhooks. | Active feature. Keep if external alerting is desired; remove if notification is handled by agent supervisor. |
| **i18n Middleware** | `internal/api/middleware/i18n.go`, `internal/i18n/` | Translates backend response messages between English and Chinese based on header. | Harmless, low footprint. Keep. |

### 3.4 DEFECT FIX (Pre-Existing Go Vet Bugs)

Running `go test ./...` currently encounters 4 compiler/vet defects in existing non-test files:
1. `internal/config/config.go:274:41`: `string(GlobalConfig.MongoDB.Port)` converts integer port to a rune rather than decimal string digits (must be `strconv.Itoa(GlobalConfig.MongoDB.Port)`).
2. `internal/config/notification.go:34,40`: `global.Log.Error(fmt.Sprintf(msg), ...)` calls `fmt.Sprintf` with non-constant format string (must be `global.Log.Error(msg, ...)`).
3. `internal/api/handlers/task/template.go:73:52`: `logger.Error(fmt.Sprintf("TemplateDetail error", err))` lacks `%v` format verb.
4. `internal/utils/ct/sdk.go:353:19`: `cb(fmt.Errorf(errorEvent.ErrorMessage), ...)` calls `fmt.Errorf` with non-constant format string (must be `fmt.Errorf("%s", errorEvent.ErrorMessage)`).
5. Offline subcommand database connection logs: running `./ScopeSentry version` or `./ScopeSentry pluginkey` triggers background Mongo/Redis connection dial timeouts in `internal/database/*/mongodb.go` and `redis.go` `init()` functions.

---

## 4. Staged Implementation Plan

Execution should proceed in 4 isolated stages, with full test and build verification at each step:

### Stage 1: Fix Pre-Existing Vet Defects
- Fix `internal/config/config.go` port conversion.
- Fix format string vet errors in `notification.go`, `template.go`, and `ct/sdk.go`.
- Verify: `go test ./...` compiles and runs cleanly across all 144 packages.

### Stage 2: Remove Confirmed Dead Code
- Delete `internal/api/routes/role/role.go`.
- Delete `internal/api/handlers/role/role.go`.
- Delete `internal/repositories/role/role.go`.
- Delete `internal/models/role.go` and remove `Role` entries from `internal/symbols/`.
- Delete `pkg/response/response.go` and remove `pkg/` directory.
- Delete `internal/models/miniprogram.go` and remove `MiniProgram` from `internal/symbols/`.
- Delete `internal/api/middleware/cors.go`.
- Strip commented-out code blocks from `internal/api/handlers/user/user.go` and `internal/api/routes/user/user.go`.
- Verify: `go build ./cmd/main`, `go build ./cmd/scopesentry`, `go test ./...`.

### Stage 3: Clean Documentation & Purge Obsolete Assets
- Remove `docs/images/` (reclaims 31MB disk and repository bloat).
- Update `README.md` and `README_CN.md` to reference headless CLI and MCP architecture instead of obsolete Vue screenshots.
- Verify: Git working tree clean of binary media bloat.

### Stage 4: Subcommand Database Isolation
- Prevent offline administrative subcommands (`version`, `pluginkey`) from triggering noisy connection error logs when databases are offline.
- Verify: `./bin/scopesentry-server version` and `./bin/scopesentry-server pluginkey` return pure JSON with zero error logs on stderr when databases are stopped.

---

## 5. Risk Assessment & Safety Invariants

| Invariant | Verification Method | Pass Criteria |
|---|---|---|
| **MCP Tool Preservation** | `go test -v ./internal/mcp` | All 17 tools remain registered with unchanged names, schemas, and handlers. |
| **CLI Dependency Decoupling** | `go list -deps ./cmd/scopesentry` | Zero references to `internal/database`, `internal/worker`, `internal/scheduler`, or Gin. |
| **CLI Output Purity** | Command execution with pipe check | stdout produces strictly valid JSON (`jq .`); logs and diagnostics routed exclusively to stderr. |
| **Scanning Protocols** | Symbol table & worker inspection | Yaegi symbol tables remain valid; worker and Redis queue structures intact. |
| **Health & Fallback** | `cmd/main/server_test.go` | `/live` (200), `/ready` (503 when DB offline, 200 when online), unmatched routes return 404 JSON. |

---

## 6. Action Decision Gate

Before beginning deletions, the following decisions are requested:

1. **Dead Code Cleanup (Stage 2):** Approve removal of `role/`, `pkg/response`, `miniprogram.go`, `cors.go`, and commented-out user management code?
2. **Obsolete Assets (Stage 3):** Approve deletion of 31MB `docs/images/` obsolete browser screenshots?
3. **Optional Features:**
   - Keep or Remove Swagger UI (`/api/swagger`)?
   - Keep or Remove Excel Export (`/api/export`)?
   - Keep or Remove Remote GitHub update check on startup?
