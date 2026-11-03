# Headless ScopeSentry & Agent-First CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert ScopeSentry into a headless, agent-first platform by removing the Vue frontend and embedded SPA, keeping all 17 MCP tools intact with English schemas, adding a headless admin key management path, and providing a lightweight standalone Go CLI client that speaks streamable HTTP to `/mcp`.

**Architecture:** 
- The backend remains a long-running Go (Gin) service managing MongoDB, Redis, workers, scanning nodes, and the streamable HTTP `/mcp` endpoint.
- The Vue SPA and all embedded static files (112 MB) are completely excised; static file serving is limited strictly to scan evidence (`/images`), swagger docs (`/api/swagger`), and health endpoints (`/live`, `/ready`).
- A lightweight Go CLI (`cmd/scopesentry/`) acts as a dedicated client consuming `/mcp` via the official `github.com/modelcontextprotocol/go-sdk/mcp` client SDK, strictly decoupled from server runtime packages (no mongo, redis, gin, workers, or scheduler in its dependency tree).
- Server-side administrative tasks (API key creation/listing/revocation) are supported via local server subcommands (`./ScopeSentry apikey ...`), avoiding any browser dependency.

**Tech Stack:** Go 1.24, `github.com/modelcontextprotocol/go-sdk/mcp`, Gin, MongoDB, Redis, Docker/Compose.

**Spec:** Migrating ScopeSentry into a headless, agent-first platform as approved in the migration design review.

## Global Constraints
- Target binary build command for server: `go build -o ScopeSentry ./cmd/main/main.go`.
- CLI binary build command: `go build -o scopesentry ./cmd/scopesentry/main.go`.
- CLI client must not import any backend server packages (`internal/database`, `internal/bootstrap`, `internal/worker`, `internal/scheduler`, `internal/services`).
- CLI stdout must output machine-readable JSON only; human diagnostics/logs must go to stderr.
- All 17 existing MCP tools must remain registered with identical names and compatible input schemas.
- Replace behavior of `create_project` on name match must be preserved.
- Scan evidence in `/images` (`GlobalConfig.System.ImgDir`) must be preserved.
- No secrets or plugin keys logged to stdout during normal server startup.

## Review Focus
1. `create_project` name collision: verify that calling `create_project` with an existing project name updates the scope targets and returns `replaced: true` instead of failing.
2. CLI stdout purity: verify that machine-readable JSON on stdout is not corrupted by informational logs, banners, or progress text (which must go to stderr).
3. MCP `isError` handling: verify that if the server returns HTTP 200 with JSON-RPC result containing `isError: true`, the CLI treats it as a command failure (exit code 1).
4. Task wait timeout non-cancellation: verify that `scopesentry task wait` timing out exits with code 3 without sending a stop/delete request to the running scan task.
5. Headless route fallback: verify that unknown routes return HTTP 404 JSON `{"error": "Not found"}` rather than falling back to an HTML page.

---

### Task 1: Fix cmd/main Build Conflict & Baseline Cleanup

**Files:**
- Move: `cmd/main/basic_usage.go` -> `examples/ct/basic_usage.go`
- Test: Verify `go build ./cmd/main` and `go test ./cmd/main`

**Interfaces:**
- Consumes: Existing Go build toolchain.
- Produces: Clean `cmd/main` package without conflicting `func main()` declarations.

- [ ] **Step 1: Relocate `cmd/main/basic_usage.go`**
Move `cmd/main/basic_usage.go` to `examples/ct/basic_usage.go` and update its package to `main`.

- [ ] **Step 2: Verify `cmd/main` package build**
Run: `go build ./cmd/main`
Expected: PASS with no duplicate main error.

- [ ] **Step 3: Run targeted existing tests**
Run: `go test ./internal/services/project ./internal/api/handlers/sensitiveRule`
Expected: PASS (6 tests passed).

---

### Task 2: Translate MCP Tool Schemas and Descriptions to English

**Files:**
- Modify: `internal/mcp/tools.go`
- Modify: `internal/mcp/schema.go`
- Test: `internal/mcp/mcp_test.go`

**Interfaces:**
- Consumes: `github.com/modelcontextprotocol/go-sdk/mcp`, internal services.
- Produces: 17 MCP tools with clear English descriptions and JSON schema annotations while preserving exact tool names and parameter keys.

- [ ] **Step 1: Write MCP registration unit test**
Create `internal/mcp/mcp_test.go` to test that `newServer()` initializes without errors and registers all 17 expected tool names.

- [ ] **Step 2: Run test to observe baseline**
Run: `go test -v ./internal/mcp`
Expected: Verify current tool count and registration.

- [ ] **Step 3: Translate tool descriptions and schemas to English**
In `internal/mcp/tools.go` and `internal/mcp/schema.go`:
- Translate `listAssetsToolDesc`, `createScanTemplateToolDesc`, `createScanTaskToolDesc` to clear English.
- Update `jsonschema` tags on struct inputs (`listProjectsDataInput`, `getProjectInput`, `createProjectInput`, `listTasksInput`, `getTaskInput`, `listScanTemplatesInput`, `getScanTemplateInput`, `createScanTemplateInput`, `createScanTaskInput`, `listAssetsInput`, `countAssetsInput`, `getAssetDetailInput`, `addAssetTagInput`, `listNodesInput`, `listPluginsInput`) to concise English.
- Keep tool names, JSON field tags, and `create_project` replacement behavior description accurate.

- [ ] **Step 4: Run test to verify passes**
Run: `go test -v ./internal/mcp`
Expected: PASS.

---

### Task 3: Strip Frontend from Server & Add Healthcheck Endpoints

**Files:**
- Modify: `cmd/main/main.go`
- Modify: `internal/api/routes/routes.go`
- Delete: `cmd/main/static/` directory
- Test: `cmd/main/server_test.go`

**Interfaces:**
- Consumes: Gin engine, MongoDB/Redis health checks.
- Produces: Clean headless server without embedded static files, providing `/live` and `/ready` endpoints, retaining `/images` and `/api/swagger`.

- [ ] **Step 1: Write server routing and healthcheck test**
Create `cmd/main/server_test.go` testing that `/live` returns 200 `{"status":"ok"}`, unknown route `/unknown` returns 404 JSON, and static assets `/map`, `/favicon.ico` return 404 JSON.

- [ ] **Step 2: Update `cmd/main/main.go`**
- Remove `//go:embed static/*` and `embeddedFiles`.
- Remove `/assets`, `/css`, `/draco`, `/js`, `/lib`, `/plugins`, `/png`, `/map`, `/favicon.ico` routes.
- Retain `/images` pointing to `config.GlobalConfig.System.ImgDir`.
- Retain Swagger `/api/swagger/*any`.
- Replace SPA fallback in `router.NoRoute` with clean 404 JSON: `c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})`.
- Add `/live` handler: returns HTTP 200 `gin.H{"status": "ok"}`.
- Add `/ready` handler: checks MongoDB client ping and Redis client ping; returns 200 `{"status": "ready"}` or 503 `{"status": "unavailable", "error": ...}`.
- Silence `PrintPlugin()` on standard startup (or only print if `SCOPESENTRY_PRINT_PLUGIN_KEY=true`).

- [ ] **Step 3: Update `internal/api/routes/routes.go`**
Remove `/assets` and `/uploads` from `loggerWithSkipPaths` prefix list if no longer used, keeping `/images`.

- [ ] **Step 4: Remove `cmd/main/static/` directory**
Remove the 112 MB `cmd/main/static/` directory from git tracking and disk.

- [ ] **Step 5: Run tests and measure binary size**
Run: `go test -v ./cmd/main`
Run: `go build -o ScopeSentry ./cmd/main/main.go && ls -lh ScopeSentry`
Expected: Tests pass and binary size drops significantly (from ~174 MB to ~60 MB).

---

### Task 4: Headless Admin CLI for API Key Management

**Files:**
- Create: `cmd/main/admin.go`
- Modify: `cmd/main/main.go`
- Test: `cmd/main/admin_test.go`

**Interfaces:**
- Consumes: `internal/services/apikey`, `internal/config`, `internal/database/mongodb`.
- Produces: CLI commands on the server binary (`./ScopeSentry apikey create/list/delete`, `./ScopeSentry pluginkey`) for headless bootstrap and administration.

- [ ] **Step 1: Write admin command tests**
Create unit tests in `cmd/main/admin_test.go` testing argument parsing for `apikey create`, `apikey list`, `apikey delete`, and `pluginkey`.

- [ ] **Step 2: Implement `admin.go`**
Create `cmd/main/admin.go` that inspects `os.Args`:
- If `len(os.Args) > 1 && os.Args[1] == "apikey"`:
  - `apikey create --name <name>`: initializes minimal DB config, calls `apikey.NewService().Create()`, prints JSON with `key` and `id`.
  - `apikey list`: calls `apikey.NewService().List()`, prints JSON list of keys.
  - `apikey delete --id <id>`: calls `apikey.NewService().Delete()`, prints JSON confirmation.
  - Exits with 0 or 1 without starting the HTTP server or workers.
- If `len(os.Args) > 1 && os.Args[1] == "pluginkey"`:
  - Prints the plugin key from config/database securely and exits.

- [ ] **Step 3: Wire into `cmd/main/main.go`**
At top of `main()`, check if `handleAdminCommands()` handled a subcommand, and return immediately if so.

- [ ] **Step 4: Run admin tests**
Run: `go test -v ./cmd/main -run TestAdmin`
Expected: PASS.

---

### Task 5: Implement the Standalone Lightweight CLI (`cmd/scopesentry/`)

**Files:**
- Create: `cmd/scopesentry/main.go`
- Create: `cmd/scopesentry/client/mcp_client.go`
- Create: `cmd/scopesentry/config/config.go`
- Create: `cmd/scopesentry/commands/project.go`
- Create: `cmd/scopesentry/commands/task.go`
- Create: `cmd/scopesentry/commands/template.go`
- Create: `cmd/scopesentry/commands/plugin.go`
- Create: `cmd/scopesentry/commands/asset.go`
- Create: `cmd/scopesentry/commands/node.go`
- Create: `cmd/scopesentry/commands/tools.go`
- Create: `cmd/scopesentry/commands/wait.go`
- Create: `cmd/scopesentry/commands/doctor.go`
- Create: `cmd/scopesentry/output/output.go`

**Interfaces:**
- Consumes: `github.com/modelcontextprotocol/go-sdk/mcp`, `github.com/spf13/cobra` (or stdlib `flag`/minimal router).
- Produces: Standalone binary `scopesentry` mapping all 17 MCP tools + utility commands. Strictly zero backend package imports.

- [ ] **Step 1: Create client package and configuration**
Implement `cmd/scopesentry/config/config.go`:
- Load URL and API key from CLI flags (`--url`, `--api-key`), environment variables (`SCOPESENTRY_URL`, `SCOPESENTRY_API_KEY`), or config file `~/.config/scopesentry/config.yaml`.
- Default URL: `http://localhost:8082`.
Implement `cmd/scopesentry/client/mcp_client.go`:
- Initializes `mcp.StreamableClientTransport` pointing to `/mcp`.
- Adds `X-API-Key` and `Authorization: Bearer <key>` headers via an `http.RoundTripper`.
- Connects using `mcp.NewClient(...)`.
- Calls tools via `session.CallTool(...)`.
- Handles `isError` on `CallToolResult`, returning structured errors.

- [ ] **Step 2: Implement command groups for all 17 tools**
- `scopesentry project list|get|create` -> maps to `list_projects`, `list_projects_data`, `get_project`, `create_project`.
- `scopesentry task list|get|create` -> maps to `list_tasks`, `get_task`, `create_scan_task`.
- `scopesentry template list|get|create` -> maps to `list_scan_templates`, `get_scan_template`, `create_scan_template`.
- `scopesentry plugin modules|list` -> maps to `list_plugin_modules`, `list_plugins`.
- `scopesentry asset list|count|get|add-tag` -> maps to `list_assets`, `count_assets`, `get_asset_detail`, `add_asset_tag`.
- `scopesentry node list` -> maps to `list_nodes`.

- [ ] **Step 3: Implement utility commands**
- `scopesentry tools list|describe|call`:
  - `list`: calls `session.ListTools(...)` and prints available tools with schemas.
  - `describe <tool>`: prints schema and description of specific tool.
  - `call <tool> [--input <json>] [--input-file <path>]`: executes tool with arbitrary arguments.
- `scopesentry task wait <task-id> [--timeout 10m] [--interval 5s]`: polls `get_task` until `status == 3` (Complete) or 100% progress. Bounded timeout without canceling scan.
- `scopesentry doctor`: connects to `/mcp`, pings server, lists tools, prints health diagnostic to stderr and JSON summary to stdout.
- `scopesentry version`: prints client version, and connects to server to fetch server version if reachable.

- [ ] **Step 4: Implement output formatting & exit code rules**
In `cmd/scopesentry/output/output.go`:
- Stdout receives ONLY the resulting JSON object/array.
- Stderr receives progress/diagnostics/errors.
- Exit codes: 0 = Success, 1 = MCP / execution error, 2 = Argument / config error, 3 = Timeout.

- [ ] **Step 5: Verify CLI dependency graph**
Run: `go list -deps ./cmd/scopesentry | grep -E 'mongo|redis|gin|worker|scheduler'`
Expected: No matches (pure standalone client).

---

### Task 6: CLI Client Unit and Mock Tests

**Files:**
- Create: `cmd/scopesentry/client/mcp_client_test.go`
- Create: `cmd/scopesentry/commands/commands_test.go`

**Interfaces:**
- Consumes: Go standard `net/http/httptest` and `github.com/modelcontextprotocol/go-sdk/mcp`.
- Produces: Comprehensive test coverage for CLI argument parsing, JSON stdout, stderr separation, exit codes, auth headers, tool routing, and error conditions.

- [ ] **Step 1: Write Mock MCP HTTP Server tests**
Test that:
- `mcp_client` transmits `X-API-Key` header.
- `CallTool` returns JSON text content.
- `isError: true` triggers an error return.
- Invalid input JSON or file read errors are handled cleanly.

- [ ] **Step 2: Test Command Mappings**
Test CLI argument parsing and schema translation for:
- `project create` preserving replacement parameters.
- `asset list` with search expression and filter JSON.
- `template create` with module hash lists or stdin `-`.
- `tools call` with arbitrary tool name and JSON input.
- `task wait` polling behavior and timeout exit code 3.

- [ ] **Step 3: Run CLI test suite**
Run: `go test -v ./cmd/scopesentry/...`
Expected: All tests PASS.

---

### Task 7: Docker, Compose & Deployment Updates

**Files:**
- Modify: `dockerfile`
- Modify: `single-host-deployment.yml`
- Modify: `single-host-deployment-fix.yml`
- Test: Build validation

**Interfaces:**
- Consumes: Headless Go codebase.
- Produces: Multi-stage Docker build that compiles ScopeSentry from source, and Compose configurations with updated `/ready` healthchecks.

- [ ] **Step 1: Update `dockerfile`**
Update `dockerfile` to a multi-stage build:
- Stage 1 (`golang:1.24-alpine`): builds `cmd/main/main.go` and `cmd/scopesentry/main.go`.
- Stage 2 (`alpine:latest` or `debian:bookworm-slim`): installs runtime dependencies (ca-certificates, tzdata), copies binaries, sets up `/opt/ScopeSentry/images`, `/opt/ScopeSentry/files`, `/opt/ScopeSentry/uploads`.

- [ ] **Step 2: Update Compose files**
In `single-host-deployment.yml` and `single-host-deployment-fix.yml`:
- Update `scope-sentry` service to build from local `build: .` context instead of stale prebuilt `autumn27/scopesentry:latest`.
- Update healthcheck command to: `test: ["CMD-SHELL", "curl -f http://127.0.0.1:8082/ready || exit 1"]`.

---

### Task 8: Documentation, Verification & Delivery

**Files:**
- Modify: `AGENTS.md`
- Modify: `SKILL.md`
- Modify: `AI_AGENT_GUIDE.md`
- Modify: `README.md`
- Create: `MIGRATION_SUMMARY.md`

**Interfaces:**
- Consumes: Completed implementation and measured test results.
- Produces: Updated guides for human operators and AI agents, CLI usage examples, binary size comparisons, and verification report.

- [ ] **Step 1: Update Agent Guides and SKILL.md**
Update `AGENTS.md`, `SKILL.md`, and `AI_AGENT_GUIDE.md` to reflect:
- Headless architecture and removal of Vue frontend.
- CLI usage, installation, and environment variables.
- Headless API key creation workflow via `./ScopeSentry apikey create`.
- Copy-paste examples for every CLI command group.

- [ ] **Step 2: End-to-end Verification & Binary Size Comparison**
- Build both binaries:
  `go build -o ScopeSentry ./cmd/main/main.go`
  `go build -o scopesentry ./cmd/scopesentry/main.go`
- Record final sizes and deltas against the 174.5 MB baseline.
- Run targeted test suites:
  `go test ./cmd/main`
  `go test ./cmd/scopesentry/...`
  `go test ./internal/mcp`
  `go test ./internal/services/project`
- Document results in `MIGRATION_SUMMARY.md`.

---
