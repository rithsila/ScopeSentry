# ScopeSentry Headless & Agent-First Migration Summary

**Date**: November 3, 2026  
**Status**: Completed & Verified  

Historical migration report: the initial headless migration exposed 17 MCP tools. The later ARTEX integration adds stop/resume for 19 tools and explicitly rejects recurring MCP requests. Current behavior is documented in `AI_AGENT_GUIDE.md`; the original verification numbers below are retained as historical evidence.

---

## 1. Executive Summary

ScopeSentry has been transformed into a headless, agent-first platform for attack surface management and vulnerability scanning. The embedded Vue 3 Single Page Application (SPA), static frontend assets, and web build pipelines were excised while preserving 100% of the backend scanning engine, distributed scanning node protocol, MongoDB/Redis data structures, and all 17 Model Context Protocol (MCP) tools.

In place of the browser UI, ScopeSentry now provides:
1. **A Headless Go Server (`cmd/main`)**: Featuring native MCP over HTTP (`/mcp`), healthchecks (`/live`, `/ready`), Swagger documentation (`/api/swagger/*any`), evidence image serving (`/images`), and a 404 JSON fallback for unmatched routes.
2. **Headless Administrative Credential Management**: CLI subcommands (`./ScopeSentry apikey`, `./ScopeSentry pluginkey`) for managing API keys and secrets directly from the terminal or container without a browser.
3. **A Standalone Go CLI Client (`cmd/scopesentry`)**: A lightweight binary implementing command groups for all 17 MCP tools plus utility commands (`tools list/describe/call`, `task wait`, `doctor`, `version`), enforcing strict stdout JSON purity and stderr diagnostic separation.

---

## 2. Binary Size Impact

By eliminating the embedded Vue SPA (`cmd/main/static`) and non-essential assets, binary footprint was drastically reduced:

| Binary | Previous Baseline | New Headless Build | Delta | % Change |
|---|---|---|---|---|
| **ScopeSentry Server** | 174.5 MB | **68.3 MB** | -106.2 MB | **-60.8%** |
| **scopesentry CLI** | N/A (New) | **11.9 MB** | +11.9 MB | Standalone |

*Note on macOS Development*: Because macOS APFS is case-insensitive by default, building `./ScopeSentry` (server) and `./scopesentry` (CLI) into the same working directory will collide. Locally on macOS, build the CLI to `bin/scopesentry` (`go build -o bin/scopesentry ./cmd/scopesentry`) or `scopesentry-cli`. Linux and container environments (ext4) are case-sensitive.

---

## 3. Dependency Isolation Verification

The standalone CLI client (`cmd/scopesentry`) communicates exclusively via the MCP protocol using `github.com/modelcontextprotocol/go-sdk` and standard HTTP transports.

Dependency check command:
```bash
go list -deps ./cmd/scopesentry | grep -E 'mongo|redis|gin|worker|scheduler'
```
**Result**: 0 matches. The CLI binary contains strictly zero backend server dependencies.

---

## 4. MCP Tools & Schema Enhancements

All 17 MCP tools were reviewed, verified, and updated with clear English descriptions and JSON schema annotations:

| # | MCP Tool | Purpose | Schema Status |
|---|---|---|---|
| 1 | `list_projects` | Project tree grouped by tag | English, Tested |
| 2 | `list_projects_data` | Paginated project list with search | English, Tested |
| 3 | `get_project` | Project details by ObjectID | English, Tested |
| 4 | `create_project` | Create or update scope on duplicate name | English, Tested |
| 5 | `list_tasks` | Paginated scan task list | English, Tested |
| 6 | `get_task` | Scan task details & progress | English, Tested |
| 7 | `create_scan_task` | Launch or schedule a scan task | English, Tested |
| 8 | `list_scan_templates` | Paginated scan template list | English, Tested |
| 9 | `get_scan_template` | Scan template details | English, Tested |
| 10 | `create_scan_template`| Build custom scan pipeline from plugin hashes | English, Tested |
| 11 | `list_plugin_modules` | Pipeline module names in execution order | English, Tested |
| 12 | `list_plugins` | Inspect available plugins and parameter defaults | English, Tested |
| 13 | `list_assets` | Query discovered assets with DSL search & filtering | English, Tested |
| 14 | `count_assets` | Count total matching assets | English, Tested |
| 15 | `get_asset_detail` | Asset details (asset) or vulnerability findings | English, Tested |
| 16 | `add_asset_tag` | Add tag to an asset | English, Tested |
| 17 | `list_nodes` | List distributed scan workers | English, Tested |

### Key Behavioral Invariants Preserved
- **Duplicate Project Scope Replacement**: When `create_project` is called with an existing project name and a valid `target`, ScopeSentry updates the asset scope of the project, preserves the project ObjectID, and returns `{"success": true, "replaced": true}`.
- **Lazy Service Initialization**: Backend services and repositories in `internal/mcp` and `internal/repositories` now initialize lazily and guard against uninitialized MongoDB/Redis pointers, enabling in-memory unit tests to register tools without running databases.

---

## 5. Headless Administrative CLI

ScopeSentry server binary now handles administrative credential operations before initializing background workers:

```bash
# Create an API Key for an AI agent
./ScopeSentry apikey create --name "cursor-agent"

# List active API keys
./ScopeSentry apikey list

# Delete an API key
./ScopeSentry apikey delete --id "<key-id>"

# Retrieve current plugin signing key
./ScopeSentry pluginkey

# Check server version
./ScopeSentry version
```

Log output in `internal/logger/logger.go` was redirected to `stderr` to guarantee that all administrative commands output pure JSON on `stdout`.

---

## 6. Standalone CLI Client (`scopesentry`)

Command groups implemented in `cmd/scopesentry/commands/`:
- `project`: `list`, `get`, `create`
- `task`: `list`, `get`, `create`, `wait`
  - `task wait <task-id> [--timeout 10m] [--interval 5s]`: polls until task completes (`status == 3` or `progress >= 100`). Progress is written to `stderr`. On timeout, exits with code 3 (`output.ExitTimeout`) without canceling the task.
- `template`: `list`, `get`, `create`
- `plugin`: `modules`, `list`
- `asset`: `list`, `count`, `get`, `add-tag`
- `node`: `list`
- `tools`: `list`, `describe <tool>`, `call <tool> [--input <json>] [--input-file <path>]`
- `doctor`: checks `/live`, `/ready`, and `/mcp`, printing diagnostic report to `stderr` and JSON status to `stdout`.
- `version`: reports CLI version and server reachability.

### Exit Codes:
- `0`: Success
- `1`: MCP or execution error
- `2`: Command argument / configuration usage error
- `3`: Polling timeout (`task wait`)

---

## 7. Containerization & Deployment Updates

1. **Multi-Stage Dockerfile**:
   - Stage 1 (`golang:1.24-alpine`): Compiles both `cmd/main` and `cmd/scopesentry`.
   - Stage 2 (`alpine:latest`): Copies compiled binaries into `/usr/local/bin` and `/opt/ScopeSentry/`, sets up asset storage directories (`/images`, `/files`, `/uploads`), installs CA certificates and timezone data.
2. **Docker Compose Updates (`single-host-deployment.yml` & `single-host-deployment-fix.yml`)**:
   - Switched `scope-sentry` service from stale prebuilt images to local `build: .` context.
   - Updated healthcheck from root web path to `/ready` endpoint:
     `test: ["CMD-SHELL", "curl -f http://127.0.0.1:8082/ready || exit 1"]`

---

## 8. Test Verification Matrix

All test suites pass cleanly:

```bash
go test ./cmd/scopesentry/... ./cmd/main ./internal/mcp ./internal/services/project ./internal/api/handlers/sensitiveRule
```

- `cmd/scopesentry/client`: Mock MCP server, streamable HTTP transport, authentication headers (`X-API-Key`, `Authorization: Bearer`), error propagation.
- `cmd/scopesentry/commands`: Command mappings, JSON argument parsing, `task wait` polling and timeout exit code 3, `tools call`.
- `cmd/main`: Health endpoints (`/live`, `/ready`), 404 JSON fallback, admin CLI (`apikey`, `pluginkey`, `version`).
- `internal/mcp`: In-memory MCP server registering all 17 tools, asset index normalization.
- `internal/services/project` & `internal/api/handlers/sensitiveRule`: Target scope parsing and legacy validation tests.

**Total**: 37 tests passed across 9 packages.
