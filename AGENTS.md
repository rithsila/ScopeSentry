# ScopeSentry

ScopeSentry is a distributed asset mapping and vulnerability scanning server (Go 1.24 + Gin) with an embedded Vue 3 SPA and built-in MCP server.

## Quick Commands

- **Build Server**: `go build -o ScopeSentry ./cmd/main/main.go`
  - *Gotcha*: Do NOT run `go build ./cmd/main`. It fails with duplicate `main` declaration in `basic_usage.go`.
- **Run Server**: `./ScopeSentry` (default port: `8082`, requires MongoDB and Redis).
- **Run Single Test Package**:
  - `go test ./internal/services/project`
  - `go test ./internal/api/handlers/sensitiveRule`
  - *Gotcha*: `go test ./...` currently fails due to duplicate `main` in `cmd/main` and Go 1.24 vet format checks in legacy packages. Run targeted package tests instead.
- **Run Single Test**: `go test -v ./internal/services/project -run TestProjectScopeEmptyWhenBlank`
- **Build Frontend**:
  - `cd frontend && npm ci && npx vite build --mode pro`
  - `rsync -a frontend/dist-pro/ cmd/main/static/`
  - *Gotcha*: Never wipe `cmd/main/static/`. It holds non-build assets (`draco/`, `lib/`, `plugins/`, `map.html`) needed at runtime. Always overlay via `rsync -a`.
- **Frontend Dev Server**: `cd frontend && pnpm dev` (runs on `http://localhost:4000`, proxies `/api`, `/mcp`, `/images` to `http://127.0.0.1:8082`).

## Architecture & Entry Points

- **Backend Entry**: `cmd/main/main.go` — sets up Gin router, embedded static assets (`//go:embed static/*`), Swagger UI (`/api/swagger/*any`), MCP server (`/mcp`), and worker listeners.
- **MCP Server**: `internal/mcp/` — exposes tools for projects, scans, assets, and nodes via Model Context Protocol. See `SKILL.md` for usage.
- **Plugin Engine**: `internal/plugins/` & `internal/symbols/` — runtime Go plugin evaluation using Yaegi (`github.com/traefik/yaegi`). Symbol tables in `internal/symbols/` must mirror exported packages.
- **Frontend Source**: `frontend/src/` — Vue 3 + TypeScript + Vite + Element Plus admin SPA. Output is built to `frontend/dist-pro/` then merged into `cmd/main/static/`.

## Runtime Quirks & Environment

- **Databases**: Requires MongoDB (default `:27017`) and Redis (default `:6379`).
- **Initial Credentials**: On first run against a clean MongoDB, ScopeSentry auto-creates the admin user (`ScopeSentry`), writes the random password to `./PASSWORD`, and prints it to stdout.
- **Configuration**: Auto-generates `config.yaml` in executable directory on start. Overridable via env vars:
  - `TIMEZONE` (default: `Asia/Shanghai`)
  - `MONGODB_IP`, `MONGODB_PORT`, `MONGODB_DATABASE`, `MONGODB_USER`, `MONGODB_PASSWORD`
  - `REDIS_IP`, `REDIS_PORT`, `REDIS_PASSWORD`
- **Docker Deployment**: `docker compose -f single-host-deployment.yml up -d` starts MongoDB, Redis, ScopeSentry server, and the scanning node.
