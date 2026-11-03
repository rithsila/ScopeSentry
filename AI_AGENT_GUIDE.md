# ScopeSentry AI Agent Guide

An operational guide for AI agents to automate and query ScopeSentry through MCP or the standalone Go CLI.

---

## 1. Headless Architecture Overview

ScopeSentry is a headless, agent-first attack surface management and security scanning platform.

- **Master Server**: Go 1.24 + Gin headless service. Exposes MCP (`/mcp`), REST API, Swagger (`/api/swagger/*any`), and image storage (`/images`).
- **CLI Client**: Standalone binary `scopesentry` (`cmd/scopesentry`) communicating over `/mcp` with zero backend dependencies. Stdout is pure JSON.
- **Scanning Nodes**: Python/Go worker daemons running scanners (subfinder, httpx, naabu, nuclei, custom plugins) synchronized via Redis and MongoDB.
- **Data Stores**:
  - **MongoDB**: Projects, tasks, scan templates, assets, and vulnerabilities.
  - **Redis**: Task queue, distributed locks, and real-time log publishing.

---

## 2. Authentication & Setup

### Creating an API Key Headlessly
No browser or UI is needed. From the server terminal or container:

```bash
./ScopeSentry apikey create --name "agent-key"
```

Output:
```json
{
  "name": "agent-key",
  "key": "ssk_xxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "createdAt": "2026-11-03T15:00:00Z"
}
```

### Option A: Direct CLI Client (`scopesentry`)
For agents with shell/command execution access:

```bash
export SCOPESENTRY_URL="http://127.0.0.1:8082"
export SCOPESENTRY_API_KEY="ssk_xxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# Verify health
scopesentry doctor
```

### Option B: MCP Server Connection
For agents running directly against the MCP protocol (Cursor, Claude Desktop, OpenCode):

- **MCP Endpoint**: `http://<HOST>:8082/mcp`
- **Headers**: `X-API-Key: ssk_...` or `Authorization: Bearer ssk_...`

```json
{
  "mcpServers": {
    "scopesentry": {
      "url": "http://127.0.0.1:8082/mcp",
      "headers": {
        "X-API-Key": "ssk_xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
      }
    }
  }
}
```

---

## 3. CLI Quick Reference for Agents

All `scopesentry` commands write machine-readable JSON to `stdout` and logs/diagnostics to `stderr`.

```bash
# Projects
scopesentry project list
scopesentry project list --by-tag
scopesentry project get <project-id>
scopesentry project create --name "Company A" --target "example.com\n1.2.3.0/24"

# Tasks
scopesentry task list --search "Recon"
scopesentry task get <task-id>
scopesentry task create --name "Initial Scan" --node "node-1" --target "example.com" --template "<template-id>"
scopesentry task wait <task-id> --timeout 15m --interval 5s

# Assets
scopesentry asset list --type subdomain --search 'subdomain.domain="example.com"'
scopesentry asset count --type asset
scopesentry asset get --type vulnerability <vuln-hash>
scopesentry asset add-tag --type asset --id <asset-id> --tag "Production"

# Templates & Plugins
scopesentry template list
scopesentry plugin modules
scopesentry plugin list --module SubdomainScan

# Direct MCP tool execution
scopesentry tools list
scopesentry tools describe create_scan_task
scopesentry tools call create_scan_task --input '{"name":"api-scan","node":["node-1"]}'
```

---

## 4. MCP Tools Catalog (All 17 Tools)

| Tool Name | What It Does | Key Arguments |
|---|---|---|
| `list_projects` | List projects grouped by tag | None |
| `list_projects_data` | Search projects with pagination | `search`, `pageIndex`, `pageSize` |
| `get_project` | Get project details | `id` (Project ObjectID) |
| `create_project` | Create project or replace scope on duplicate name | `name`, `target`, `tag`, `template`, `node` |
| `list_tasks` | List scan tasks | `search`, `pageIndex`, `pageSize` |
| `get_task` | Get task status, progress, and node execution stats | `id` (Task ObjectID) |
| `create_scan_task` | Launch or schedule a scan task | `name`, `node`, `template`, `targetSource` |
| `list_scan_templates` | List scan templates | `query`, `pageIndex`, `pageSize` |
| `get_scan_template` | Get scan template details | `id` (Template ObjectID) |
| `create_scan_template`| Build custom scan pipeline from plugin hashes | `name`, `modules`, `parameters`, `vullist` |
| `list_plugin_modules` | List all pipeline module names in run order | None |
| `list_plugins` | List scanning plugins & parameter defaults | `module` (optional filter), `search` |
| `list_assets` | Query discovered assets with DSL search & filtering | `asset_type`, `search`, `filter`, `pageIndex`, `pageSize` |
| `count_assets` | Count total matching assets | `asset_type`, `search`, `filter` |
| `get_asset_detail` | Get full asset record or vulnerability finding | `id`, `asset_type` (asset or vulnerability) |
| `add_asset_tag` | Add custom tag to an asset | `asset_type`, `id`, `tag` |
| `list_nodes` | List worker nodes & status | `online_only` (boolean) |

---

## 5. Asset Types & Search DSL

### Asset Types
- `asset`: HTTP web applications, technologies, and ports.
- `subdomain`: Discovered subdomains and DNS records.
- `RootDomain`: Monitored root domains.
- `vulnerability`: Detected vulnerabilities and Nuclei findings.
- `UrlScan`: Extracted URLs, endpoints, and paths.
- `SensitiveResult`: Leaked keys, tokens, or credentials.
- `DirScanResult`: Discovered directories and backup files.
- `PageMonitoring`: Website change detection diffs.
- `IPAsset`: Resolved IPs, ports, and banners.
- `SubdomainTakerResult`: Subdomain takeover vulnerabilities.

### Search DSL Rules
- `field="value"`: Case-insensitive substring match.
- `field=="exact"`: Exact match.
- `field!="value"`: Negated match.
- `expr1 && expr2`: Logical AND.
- `expr1 || expr2`: Logical OR.

### Examples:
```text
asset.host="api.example.com" && asset.code==200
subdomain.domain="example.com"
vulnerability.level=="critical" || vulnerability.level=="high"
url.url="login" && url.status_code==200
```

---

## 6. Important Agent Rules & Gotchas

1. **Project Scope Replacement on Duplicate Name**:
   When `create_project` is called with an existing project name and a valid `target`, ScopeSentry does not return an error; it replaces the target scope of the existing project and preserves the project ObjectID. The response contains `"replaced": true`.
2. **Project IDs in Filters**:
   When filtering assets or tasks by project, always use the MongoDB ObjectID (24 hex characters), not the project name.
3. **Task Completion Detection**:
   Tasks finish when `status == 3` or `progress >= 100`. Use `scopesentry task wait <task-id>` for automatic polling with timeout safeguards.
4. **Always Check Nodes First**:
   Always verify active nodes with `list_nodes({ online_only: true })` before creating a task. Tasks assigned to offline nodes will stall in `Waiting` state.
5. **Count Before Paging**:
   Call `count_assets` before paginating through large asset lists with `list_assets`.
