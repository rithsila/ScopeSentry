# ScopeSentry (Headless & Agent-First)

ScopeSentry is a headless, distributed asset mapping and vulnerability scanning platform (Go 1.24 + Gin) designed for AI agents and automation. The Vue 3 SPA has been excised in favor of a native Model Context Protocol (MCP) server, a headless administrative CLI, and a lightweight standalone Go CLI client (`scopesentry`).

## Quick Commands

- **Build Server**: `go build -o ScopeSentry ./cmd/main`
- **Build CLI**: `go build -o bin/scopesentry ./cmd/scopesentry` (separate directory avoids a server/client filename collision on case-insensitive macOS filesystems).
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
  - `scopesentry task stop <id>` / `scopesentry task resume <id>` (request acceptance, not scanner acknowledgement).
  - `scopesentry asset list --type <type> --search <dsl>`
  - `scopesentry tools list` / `scopesentry tools describe <tool>` / `scopesentry tools call <tool>`
- **Run Tests**:
  - Isolated unit suite on macOS: `bash scripts/test-isolated.sh` (requires `scopesentry_unit` bootstrap; never deploy this build tag).
  - CLI packages: `bash scripts/test-isolated.sh ./cmd/scopesentry/...`
  - Server/Admin/MCP: `bash scripts/test-isolated.sh ./cmd/main ./internal/mcp`
  - Race checks: `RACE=1 bash scripts/test-isolated.sh`
  - Do not run untagged backend tests: transitive package `init()` can read runtime configuration, connect to databases, and create config/key files before test fakes. On other platforms use an explicitly isolated environment; the macOS runner refuses an unsafe fallback.

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
  - Exposes 19 tools with English descriptions and JSON schemas, including scan stop/resume. Discover the live list because older deployments expose 17.
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
