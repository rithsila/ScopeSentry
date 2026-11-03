---
name: scopesentry-mcp
description: Manage ScopeSentry security scanning platform (projects, tasks, templates, assets, nodes) via MCP or the lightweight Go CLI client. Use when interacting with ScopeSentry, creating scan tasks, querying assets, or configuring templates.
---

# ScopeSentry MCP User Guide

ScopeSentry is a headless, agent-first attack surface management and vulnerability scanning server. Connect via the native Model Context Protocol (MCP) server or using the standalone Go CLI client (`scopesentry`).

## 1. Prerequisites & Authentication

### 1.1 Verify Service Accessibility

- MCP Endpoint: `http://<host>:8082/mcp`
- Health Endpoints: `http://<host>:8082/live` (liveness) and `http://<host>:8082/ready` (readiness)
- Swagger API Docs: `http://<host>:8082/api/swagger/index.html`

### 1.2 Create API Key Headlessly

No browser or frontend UI required. Create an API key using the administrative CLI on the server host:

```bash
./ScopeSentry apikey create --name "agent-key"
```

Example response:
```json
{
  "name": "agent-key",
  "key": "ssk_a1b2c3d4e5f6...",
  "createdAt": "2026-11-03T15:00:00Z"
}
```

Save the `ssk_...` token (displayed only upon creation).

To list or delete keys:
```bash
./ScopeSentry apikey list
./ScopeSentry apikey delete --id <key-id>
```

### 1.3 Configure MCP Client (Cursor, Claude Desktop, OpenCode)

Add the MCP server to your configuration:

```json
{
  "mcpServers": {
    "scopesentry": {
      "url": "http://<host>:8082/mcp",
      "headers": {
        "X-API-Key": "ssk_your_key"
      }
    }
  }
}
```

Authentication is accepted via `X-API-Key: ssk_...` or `Authorization: Bearer ssk_...`.

---

## 2. Standalone CLI Client (`scopesentry`)

The `scopesentry` binary provides a command-line interface that communicates directly with the MCP server:

```bash
# Set credentials
export SCOPESENTRY_URL="http://localhost:8082"
export SCOPESENTRY_API_KEY="ssk_your_key"

# Verify connectivity
scopesentry doctor

# Manage projects
scopesentry project list
scopesentry project create --name "Acme Corp" --target "acme.com,192.168.1.0/24"

# Query assets
scopesentry asset list --type subdomain --search 'subdomain.domain="acme.com"'
scopesentry asset count --type asset

# Manage scan tasks
scopesentry task list
scopesentry task wait <task-id> --timeout 15m --interval 5s

# Direct MCP tool execution
scopesentry tools list
scopesentry tools describe list_assets
scopesentry tools call list_assets --input '{"asset_type":"asset","pageIndex":1}'
```

Output convention: stdout outputs pure, parseable JSON; stderr receives progress and diagnostics.

---

## 3. Tool Overview (All 17 MCP Tools)

| Tool | Purpose | Key Parameters |
| --- | --- | --- |
| `list_projects` | Project tree grouped by tag | `{}` |
| `list_projects_data` | Paginated project list with name search | `search`, `pageIndex`, `pageSize` |
| `get_project` | Project details | `id` (ObjectID) |
| `create_project` | Create project (replaces scope if name matches) | `name`, `target`, `tag`, `template`, `node`, `allNode` |
| `list_tasks` | Paginated scan task list | `search`, `pageIndex`, `pageSize` |
| `get_task` | Scan task details & progress | `id` (ObjectID) |
| `create_scan_task` | Launch or schedule a scan task | `name`, `node`, `target`, `template`, `targetSource` |
| `list_scan_templates` | Paginated scan template list | `query`, `pageIndex`, `pageSize` |
| `get_scan_template` | Scan template details | `id` (ObjectID) |
| `create_scan_template` | Assemble scan pipeline from modules/plugins | `name`, `modules`, `parameters`, `vullist`, `template_json` |
| `list_plugin_modules` | Pipeline module names in execution order | `{}` |
| `list_plugins` | Inspect available plugins and parameter defaults | `module`, `search` |
| `list_assets` | Query discovered assets with DSL search & filtering | `asset_type`, `search`, `filter`, `pageIndex`, `pageSize` |
| `count_assets` | Count total matching assets | `asset_type`, `search`, `filter` |
| `get_asset_detail` | Asset details (asset) or vulnerability findings | `asset_type`, `id` (ObjectID or vuln hash) |
| `add_asset_tag` | Add tag to an asset | `asset_type`, `id`, `tag` |
| `list_nodes` | List distributed scan workers | `online_only` (bool) |

---

## 4. Search DSL Reference for `list_assets` & `count_assets`

Use `search` for full-text and field-level queries (supports `&&`, `||`, `!`, and parentheses):

- Subdomains: `subdomain.domain="example.com"`
- Web Assets: `asset.host="example.com"`, `asset.title="Login"`, `asset.code==200`
- URLs: `url.url="login"`, `url.status_code==200`
- Vulnerabilities: `vulnerability.name="Log4j"`, `vulnerability.level=="critical"`
- Sensitive Data: `sensitive.body="password"`
