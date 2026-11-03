package mcp

// listAssetsToolDesc explains asset query syntax for MCP clients
const listAssetsToolDesc = `Query ScopeSentry asset list.

[Request Body]
Matches POST endpoints (e.g., /api/assets/asset) with models.SearchRequest:
- pageIndex, pageSize: Pagination controls
- search: Search expression string DSL (e.g. domain=example.com && port==443)
- filter: Exact filter object (e.g. filter.project with project ObjectIDs, filter.task with task names)
- sort: Sorting (supported on UrlScan/DirScanResult: {"length":"ascending"})
- sid: Sensitive rule name (SensitiveResult only)

asset_type values:
- asset: Web/port assets
- RootDomain: Monitored root domains
- subdomain: Discovered subdomains
- app: Mobile applications
- mp: Mini-programs
- UrlScan: URLs and endpoints
- SensitiveResult: Sensitive findings/leaks
- DirScanResult: Directory brute-force findings
- crawler: Web crawler links
- vulnerability: Detected vulnerabilities
- PageMonitoring: Webpage change detection diffs
- IPAsset: IP aggregate assets
- SubdomainTakerResult: Subdomain takeover findings

[search DSL Rules]
- field=value : Fuzzy match (case-insensitive)
- field=="value" : Exact match (quoted if containing spaces)
- field!="value" : Negative match
- expr1 && expr2 : Logical AND
- expr1 || expr2 : Logical OR
- (expr) : Grouping
- Common search fields across types: tag->tags, task->taskName, rootDomain
- Note: project cannot be queried via search; use filter.project instead.

Search keywords by asset_type:
- asset: domain, ip, port, service, app, title, statuscode, icon, banner, type, body, header
- RootDomain: domain, icp, company
- subdomain: domain, ip, type, value
- app: name, icp, company, category, description, url, apk
- mp: name, icp, company, category, description, url
- UrlScan: url, input, source, resultId, type
- SensitiveResult: url, sname, body, info, md5
- DirScanResult: url, statuscode, redirect, length
- vulnerability: url, vulname, matched, request, response, level
- crawler: url, method, body, resultId
- PageMonitoring: url, hash, diff, response
- IPAsset: ip, domain, port, service, webServer, app
- SubdomainTakerResult: domain, value, type, response

[filter Object]
JSON map where keys are dimensions and values are string arrays (OR within same key, AND across keys).
- filter.project: Array of project ObjectIDs (e.g. ["67262c5b9f71c40f8a846b10"])
- filter.task: Array of task names (not task IDs)
- filter.level: Vulnerability severity ("critical","high","medium","low","info","unknown")
- filter.status: Status code or handling status (1-6)`

// createScanTemplateToolDesc explains template creation for MCP clients
const createScanTemplateToolDesc = `Create a scan template.

A scan template consists of modules (pipeline stages), each mounting one or more plugins identified by plugin MD5 hashes.

Recommended workflow:
1. list_plugin_modules to get available module names
2. list_plugins (filtered by module) to get plugin hashes and default parameters
3. Pass modules map: {moduleName: [pluginHash1, pluginHash2]}
4. Plugin default parameters are automatically populated; override using parameters if needed

Parameters:
- name: Template name (required)
- modules: Map of module names to plugin hash arrays (e.g. {"SubdomainScan":["d60ba73c..."]})
- parameters: Optional parameter overrides {module: {pluginHash: paramString}}
- vullist: Optional Nuclei POC template IDs for VulnerabilityScan
- template_json: Optional complete ScanTemplate JSON (takes precedence)

Pipeline modules: TargetHandler, SubdomainScan, SubdomainSecurity, PortScanPreparation,
PortScan, PortFingerprint, AssetMapping, AssetHandle, URLScan, WebCrawler,
URLSecurity, DirScan, VulnerabilityScan, PassiveScan`

// createScanTaskToolDesc explains task creation for MCP clients
const createScanTaskToolDesc = `Create a scan task. Requires name and node; template is a scan template ObjectID.

targetSource determines target extraction:
- general: Direct targets in target field (multiline or comma-separated domains/IPs/URLs)
- project: Read targets from project ObjectIDs in project field
- asset / RootDomain / subdomain / UrlScan: Search targets from the asset database using search/filter
- *Source (e.g. subdomainSource): Create from asset pages using targetTp='search' or 'select'

Other parameters:
- allNode: Run on all online nodes
- ignore / duplicates: Target exclusions and deduplication policy
- bindProject: Project ObjectID to associate findings with
- scheduledTasks: Enable recurring scan with cycleType/hour/minute/day/week`

// countAssetsToolDesc explains asset counting for MCP clients
const countAssetsToolDesc = `Count total assets matching search and filter criteria without pagination.

Parameters match list_assets (asset_type, search, filter, sid).
Returns total matching count.`
