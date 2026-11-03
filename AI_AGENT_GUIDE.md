# ScopeSentry AI Agent Guide

A quick operational reference for AI agents to understand, query, and automate ScopeSentry.

---

## 1. What is ScopeSentry?

ScopeSentry is a distributed attack surface management and vulnerability scanning platform.

- **Master Server**: Go (Gin) backend that manages projects, tasks, schedules, and the web interface.
- **Scanning Nodes**: Python/Go worker daemons that pull scan jobs from a Redis queue and execute security tools (Subfinder, httpx, naabu, nuclei, custom plugins).
- **Storage**:
  - **MongoDB**: Stores projects, tasks, scan templates, assets, and vulnerabilities.
  - **Redis**: Handles job queuing, distributed locks, and live log pub/sub.

---

## 2. How to Connect

### Option A: Model Context Protocol (MCP) — Recommended

If your client supports MCP (Cursor, Claude Desktop, OpenCode):

- **MCP Endpoint**: `http://<HOST>:8082/mcp`
- **Authentication**: HTTP Header `X-API-Key: ssk_...` or `Authorization: Bearer ssk_...`

Configuration snippet:
```json
{
  "mcpServers": {
    "scopesentry": {
      "url": "http://10.10.23.6:8082/mcp",
      "headers": {
        "X-API-Key": "ssk_YOUR_API_KEY"
      }
    }
  }
}
```

### Option B: REST API

- **Base URL**: `http://<HOST>:8082/api`
- **Header**: `X-API-Key: ssk_...`

Key endpoints:
- `POST /api/project/data` — List projects (paginated)
- `POST /api/project/create` — Create project
- `POST /api/task/list` — List tasks
- `POST /api/task/add` — Create scan task
- `POST /api/assets/<asset_type>` — Query assets
- `POST /api/assets/common/total` — Count assets

---

## 3. Core Concepts

### Projects
A project groups targets (root domains, subdomains, IPs, CIDRs) under one container.
- **Target format**: Multiline string or comma-separated list (`example.com`, `192.168.1.0/24`).
- **Identifier**: MongoDB ObjectID (e.g. `67262c5b9f71c40f8a846b10`). Always use this ID in filters, not the project name.

### Scan Templates
Templates define which plugins execute and in what order.
- **Pipeline Modules**: `TargetHandler` -> `SubdomainScan` -> `SubdomainSecurity` -> `PortScanPreparation` -> `PortScan` -> `PortFingerprint` -> `AssetMapping` -> `AssetHandle` -> `URLScan` -> `WebCrawler` -> `URLSecurity` -> `DirScan` -> `VulnerabilityScan` -> `PassiveScan`.
- **Plugins**: Modular scripts identified by their MD5 hash.

### Scan Tasks
A task runs a Scan Template against targets on one or more Nodes.
- **Status lifecycle**: `Waiting` -> `Running` -> `Complete` / `Error`.

---

## 4. MCP Tools Catalog

| Tool Name | What It Does | Key Arguments |
|---|---|---|
| `list_projects` | List projects grouped by tag | None |
| `list_projects_data` | Search projects with pagination | `search`, `pageIndex`, `pageSize` |
| `get_project` | Get project details | `id` (Project ObjectID) |
| `create_project` | Create or update target scope | `name`, `target`, `tag`, `template` |
| `list_tasks` | List scan tasks | `pageIndex`, `pageSize` |
| `get_task` | Get task status and progress | `id` (Task ObjectID) |
| `list_scan_templates` | List scan templates | `pageIndex`, `pageSize` |
| `get_scan_template` | Get scan template details | `id` (Template ObjectID) |
| `list_plugin_modules` | List all pipeline module names | None |
| `list_plugins` | List scanning plugins & hashes | `module` (optional filter) |
| `create_scan_template`| Build a custom scan template | `name`, `modules`, `parameters` |
| `create_scan_task` | Launch a scan task | `name`, `node`, `template`, `targetSource` |
| `list_assets` | Query discovered assets | `asset_type`, `search`, `filter`, `pageIndex` |
| `count_assets` | Count total matching assets | `asset_type`, `search`, `filter` |
| `get_asset_detail` | Get full asset or bug record | `id`, `asset_type` |
| `list_nodes` | List worker nodes & status | `online_only` (boolean) |

---

## 5. Asset Types & Search Syntax

### Available Asset Types
- `asset`: HTTP web applications and open ports.
- `subdomain`: Discovered subdomains.
- `RootDomain`: Monitored root domains.
- `vulnerability`: Detected vulnerabilities (Nuclei, custom POCs).
- `UrlScan`: Extracted URLs and endpoints.
- `SensitiveResult`: Leaked keys, tokens, or credentials.
- `DirScanResult`: Discovered directories and backup files.
- `PageMonitoring`: Website change detection diffs.

### Search DSL Rules (`search` string)
- `field=value`: Fuzzy contains match (case-insensitive).
- `field=="exact"`: Exact match.
- `field!="value"`: Exclude value.
- `expr1 && expr2`: Logical AND.
- `expr1 || expr2`: Logical OR.

### Example Search Queries
```text
domain=example.com && port==443
statuscode==200 && title="Admin"
service=nginx && (port==80 || port==8080)
level==critical || level==high
```

### Filter Rules (`filter` object)
Use the `filter` JSON parameter for exact indexed lookups.
- `project`: Always an array of Project ObjectIDs: `{"project": ["67262c5b..."]}`.
- `level`: Vulnerability severity: `{"level": ["critical", "high"]}`.
- `port`: Port numbers: `{"port": ["443", "8443"]}`.

---

## 6. Standard Agent Workflows

### Workflow 1: Create a Project and Run a Scan

1. Call `list_nodes({ online_only: true })` to verify at least one node is online.
2. Call `list_scan_templates()` to select an existing template ObjectID.
3. Call `create_project()`:
   ```json
   {
     "name": "Target Company",
     "target": "example.com\n1.2.3.4/24",
     "tag": "External"
   }
   ```
4. Call `create_scan_task()` to start scanning:
   ```json
   {
     "name": "Target Company - Initial Recon",
     "node": ["node-1"],
     "template": "<TEMPLATE_OBJECT_ID>",
     "targetSource": "project",
     "project": ["<PROJECT_OBJECT_ID>"]
   }
   ```

### Workflow 2: Triaging Vulnerabilities

1. Get the total count:
   ```json
   {
     "asset_type": "vulnerability",
     "filter": {
       "level": ["critical", "high"]
     }
   }
   ```
2. Retrieve the vulnerability list:
   ```json
   {
     "asset_type": "vulnerability",
     "filter": {
       "level": ["critical", "high"]
     },
     "pageIndex": 1,
     "pageSize": 20
   }
   ```
3. Inspect details with `get_asset_detail({ id: "<VULN_ID>", asset_type: "vulnerability" })`.

### Workflow 3: Export Discovered Web Services

1. Query active web services for a project:
   ```json
   {
     "asset_type": "asset",
     "filter": {
       "project": ["<PROJECT_OBJECT_ID>"],
       "statuscode": ["200", "301", "302"]
     },
     "pageIndex": 1,
     "pageSize": 50
   }
   ```
2. Extract URLs and titles from results for further inspection.

---

## 7. Important Agent Rules & Gotchas

1. **Project ID vs Project Name**:
   - `filter.project` **must** be a MongoDB ObjectID (24 hex characters). Do not pass the project name.
2. **Task Name vs Task ID**:
   - `filter.task` accepts the task **name** string, not its ObjectID.
3. **Template ID**:
   - `create_scan_task` requires the template's **ObjectID** in the `template` parameter.
4. **Always Check Nodes**:
   - Before launching a task, check `list_nodes({ online_only: true })`. Tasks assigned to offline nodes stay in `Waiting` state indefinitely.
5. **Count Before Paging**:
   - Use `count_assets` before looping through `list_assets` to know the total record count in advance.
