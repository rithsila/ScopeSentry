# ScopeSentry (Headless & Agent-First)

ScopeSentry is a headless, distributed asset mapping and vulnerability scanning platform (Go 1.24 + Gin) designed for AI agents and automation. The Vue 3 SPA has been excised in favor of a native Model Context Protocol (MCP) server, a headless administrative CLI, and a lightweight standalone Go CLI client (`scopesentry`).

## Quick Commands

- **Build Server**: `go build -o ScopeSentry ./cmd/main`
- **Build CLI**: `go build -o scopesentry ./cmd/scopesentry`
- **Run Server**: `./ScopeSentry` (default port: `8082`, requires MongoDB and Redis).
- **Headless Admin Credentials**:
  - Create API Key: `./ScopeSentry apikey create --name <key-name>`
  - List API Keys: `./ScopeSentry apikey list`
  - Delete API Key: `./ScopeSentry apikey delete --id <key-id>`
  - Get Plugin Key: `./ScopeSentry pluginkey`
- **Run CLI Commands**:
  - `scopesentry doctor` (health check for server, /live, /ready, /mcp)
  - `scopesentry project list` / `scopesentry project create --name <n> --target <t>`
  - `scopesentry task list` / `scopesentry task create ...` / `scopesentry task wait <id>`
  - `scopesentry asset list --type <type> --search <dsl>`
  - `scopesentry tools list` / `scopesentry tools describe <tool>` / `scopesentry tools call <tool>`
- **Run Tests**:
  - Full CLI test suite: `go test -v ./cmd/scopesentry/...`
  - Server & Admin test suite: `go test -v ./cmd/main`
  - In-memory MCP tools test suite: `go test -v ./internal/mcp`
  - Targeted package tests: `go test ./internal/services/project ./internal/api/handlers/sensitiveRule`

## Architecture & Entry Points

- **Headless Server Entry**: `cmd/main/main.go` & `cmd/main/admin.go`
  - Subcommands: `apikey`, `pluginkey`, `version`.
  - HTTP Server: MCP server (`/mcp` with `X-API-Key` or `Authorization: Bearer`), Node API routes, Swagger UI (`/api/swagger/*any`), scan evidence image serving (`/images`), `/live` (liveness), and `/ready` (MongoDB + Redis ping).
  - Unmatched routes return JSON `{"error": "Not found"}` (HTTP 404).
- **CLI Client Entry**: `cmd/scopesentry/main.go` & `cmd/scopesentry/commands/`
  - Standalone binary connecting exclusively over `/mcp` using `github.com/modelcontextprotocol/go-sdk`.
  - Zero backend dependencies: does NOT import mongo, redis, gin, workers, or scheduler.
  - Strict output convention: stdout outputs pure, machine-readable JSON; stderr outputs progress, diagnostics, and errors.
- **MCP Server**: `internal/mcp/`
  - Exposes all 17 security scanning tools with English descriptions and JSON schemas.
  - See `SKILL.md` for full tool documentation and search DSL reference.
- **Plugin Engine**: `internal/plugins/` & `internal/symbols/`
  - Runtime Go plugin evaluation using Yaegi (`github.com/traefik/yaegi`).

## Runtime Quirks & Environment

- **Databases**: Requires MongoDB (default `:27017`) and Redis (default `:6379`).
- **Initial Credentials**:
  - Generate an API key headlessly without a browser: `./ScopeSentry apikey create --name admin`.
  - Fetch plugin key: `./ScopeSentry pluginkey`.
- **Configuration**: Auto-generates `config.yaml` in executable directory on start. Overridable via env vars:
  - `TIMEZONE` (default: `Asia/Shanghai`)
  - `MONGODB_IP`, `MONGODB_PORT`, `MONGODB_DATABASE`, `MONGODB_USER`, `MONGODB_PASSWORD`
  - `REDIS_IP`, `REDIS_PORT`, `REDIS_PASSWORD`
  - `SCOPESENTRY_PRINT_PLUGIN_KEY` (set to `true` to print plugin key banner during server boot)
- **Docker Deployment**: `docker compose -f single-host-deployment.yml up -d` builds and runs the headless server, MongoDB, Redis, and scanner node with `/ready` healthcheck.
