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
scopesentry project create --name "Acme Corp" --target 'acme.com
192.168.1.0/24'

# Query assets
scopesentry asset list --type subdomain --search 'domain="acme.com"'
scopesentry asset count --type asset

# Manage scan tasks
scopesentry task list
scopesentry task wait <task-id> --timeout 15m --interval 5s
scopesentry task stop <task-id>
scopesentry task resume <task-id>

# Direct MCP tool execution
scopesentry tools list
scopesentry tools describe list_assets
scopesentry tools call list_assets --input '{"asset_type":"asset","pageIndex":1}'
```

Output convention: stdout outputs pure, parseable JSON; stderr receives progress and diagnostics.

---

## 3. Tool Overview

This version registers 19 tools. Always discover the live tool list; older deployments may expose 17.

| Tool | Purpose | Key Parameters |
| --- | --- | --- |
| `list_projects` | Project tree grouped by tag | `{}` |
| `list_projects_data` | Paginated project list with name search | `search`, `pageIndex`, `pageSize` |
| `get_project` | Project details | `id` (ObjectID) |
| `create_project` | Create project (replaces scope if name matches) | `name`, `target`, `tag`, `template`, `node`, `allNode` |
| `list_tasks` | Paginated scan task list | `search`, `pageIndex`, `pageSize` |
| `get_task` | Scan task details & progress | `id` (ObjectID) |
| `create_scan_task` | Launch a one-shot scan after template/node validation | `name`, `node`, `target`, `template`, `targetSource` |
| `stop_scan_task` | Request a scan to stop/pause | `id` (task ObjectID) |
| `resume_scan_task` | Request a paused scan to resume | `id` (task ObjectID) |
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

- Subdomains: `domain="example.com"`
- Web Assets: `domain="example.com"`, `title="Login"`, `statuscode==200`
- URLs: `url="login"`
- Vulnerabilities: `vulname="Log4j"`, `level=="critical"`
- Sensitive Data: `body="password"`

Use only the fields supported by the selected asset type. Do not prefix fields with a collection name. Restrict scan results with `filter: {"task":["scan-name"]}` or project ObjectIDs with `filter: {"project":["project-id"]}`. Do not put project names into the search DSL.

## 5. Scan control and ARTEX

The MCP endpoint ends in `/mcp`; `SCOPESENTRY_URL` is the base URL without that suffix. The CLI appends `/mcp`. Global CLI flags are `--url` and `--api-key` and belong before the command. Keep keys in secret configuration, not chat messages or source files.

Check `list_nodes` with `online_only:true` and inspect an existing template before creating a scan. Pass a template ObjectID and actual online node names. `allNode:true` uses only online nodes. Recurring requests (`scheduledTasks:true`) are rejected through MCP/CLI; use one-shot tasks and a separate scheduler.

A create response confirms a saved task, not worker acknowledgement or successful completion. Poll `get_task` or `task wait`; a wait timeout does not cancel the scan. Stop/resume responses confirm acceptance of a request, not that all scanner processes have stopped or resumed. Check subsequent state and worker evidence.

ARTEX Sync and agent MCP access reuse the saved connection named `ScopeSentry`. Sync imports existing subdomain/service/app assets; it does not launch scans or import vulnerability findings. Chat can control ScopeSentry through the visible skill/MCP tools. Auto is the standalone platform chat agent; Main Agent is the task console. Grant access separately to Worker if it must call the service. Reading remote results does not automatically sync them into ARTEX.

Creating a project with an existing name replaces its entire target scope. Before adding targets, read the complete current target scope successfully, preserve every existing target, merge new entries, and submit the full intended scope using real newlines. Stop if the read fails or completeness cannot be confirmed. The CLI runs on the ARTEX backend host/container, not in the browser. The headless server, MongoDB, Redis, and scanner nodes are still required.
