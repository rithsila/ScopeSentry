# ARTEX integration: Sync, MCP, terminal, and chat

## Connection model

Keep ScopeSentry's server, MongoDB, Redis, and scanner nodes running. ARTEX's agents do not replace the scanner service.

| Path | Purpose | Setup |
|---|---|---|
| Sync page | Import discovered assets into ARTEX | Existing `ScopeSentry` MCP entry |
| Direct MCP | Projects, scopes, one-shot scans, status, results, stop/resume | Per-agent visibility; when the linked skill is visible, load it to unlock tools |
| Terminal | Same remote operations through `scopesentry` | Compatible CLI installed where ARTEX backend runs; separate CLI credential configuration |
| Chat | Ask Auto/Main Agent to operate the above | Appropriate tool/skill visibility for that exact agent |

Asset Sync and direct MCP share ARTEX's saved endpoint/header configuration. CLI configuration is separate. Do not add a duplicate MCP entry just for chat or terminal use.

## Build the ScopeSentry side

From the ScopeSentry checkout:

```bash
go build -o bin/ScopeSentry-server ./cmd/main
go build -o bin/scopesentry ./cmd/scopesentry
```

These paths avoid the `ScopeSentry`/`scopesentry` case-insensitive filename collision on macOS. Building does not deploy or start either executable.

The server process needs its normal service configuration and worker deployment. The CLI only needs network access and a service API key. Run API-key administration on the ScopeSentry host using its configured server binary; avoid printing keys in agent chats or execution logs.

## Configure ARTEX

1. In **Sync**, keep the existing data source named **ScopeSentry**, with endpoint `http://<server>:8082/mcp` and its API-key header. The existing MCP card is the same saved connection.
2. In **Skills**, make the `scopesentry` skill visible to **Auto** for standalone Chat, or **Main Agent** for task-console chat. Its `mcps: ScopeSentry` binding makes tools available when the agent loads that skill. Grant Worker independently only if needed.
3. If chat should import results into ARTEX, bind the opt-in host tool **`sync_scopesentry_assets`** to that agent in the existing Tools settings. This tool is separate from ScopeSentry's remote tools and belongs to ARTEX.
4. For terminal use, install the matching CLI on the backend host or inside its container. Configure `SCOPESENTRY_URL=http://<server>:8082` and provide `SCOPESENTRY_API_KEY` through secret configuration. Stored MCP headers do not become shell environment variables.
5. Refresh MCP discovery and use a new chat turn. A deployed old server may still show 17 tools; this version exposes 19. Do not assume a successful connection means a scanner is online.

The tools remain opt-in. Source changes do not change your live permissions or credentials.

If a visible skill links `ScopeSentry`, the agent must load that skill before using its tools, even if the MCP card also grants direct visibility. Direct MCP visibility alone provides tools at the start only when no visible skill gates that server.

## Read-only chat check

Use Auto on ARTEX's Chat page:

> Load the ScopeSentry skill. Use the existing ScopeSentry MCP connection. List projects, online scanner nodes, and scan templates. Do not create or modify anything. Report the exact available tool names and any connection errors.

Direct MCP tools are deferred. ARTEX discovers their schemas through `SearchExtraTools`, then calls them with `ExecuteExtraTool`. For the standard connection name, external tools have names such as `mcp__ScopeSentry__create_project`. ARTEX's unprefixed task-management tools control ARTEX tasks, not ScopeSentry scans.

## One-shot scan workflow

Only use this workflow after choosing the intended targets, template, and nodes:

1. Read/select the project. To **add** targets to an existing project, first read its full current `target` value. Preserve every existing target, merge new entries, and send the complete scope to `create_project`. Stop if the old scope cannot be read completely. Reusing a name replaces scope; it is not append.
2. Use `list_nodes` with `online_only:true`, then `list_scan_templates` and `get_scan_template`. Select existing ObjectIDs and node names. A template's name is not its ID.
3. Create a uniquely named one-shot task with `targetSource:"project"` and `project:["<project-id>"]`, or explicitly supplied targets. Use actual newlines for multiple direct targets.
4. Read `get_task` until completion or a reported problem. A creation response only confirms the task was saved. `task wait` timeout does not cancel it. Stop with `stop_scan_task` / `task stop`, or resume a paused task with `resume_scan_task` / `task resume`. Control acceptance is not scanner acknowledgement.
5. Read results with `count_assets` / `list_assets`, filtering by **task name**, for example `filter:{"task":["artex-recon-001"]}`. Project filters take ObjectIDs. Import assets separately using Sync or the ARTEX sync host tool.

MCP/CLI rejects recurring scheduling requests; use an external scheduler for repeated one-shot scans. Asset Sync imports subdomains, services, apps, and related IPs, with a 5,000-item cap per selected project/task and asset type. It does not import vulnerability findings.

## Terminal checks

After the backend environment has a compatible CLI and credentials:

```bash
scopesentry doctor
scopesentry tools list
scopesentry node list --online-only
scopesentry template list
scopesentry project list
```

Inspect `doctor`'s JSON `status`, not only its exit code: a degraded service can exit 0. Health endpoints do not check scanner availability.

## Verification boundaries

Local check results and limits are recorded in [the verification report](artex-integration-verification.md).

Local unit tests and builds do not prove the remote deployment is upgraded, that agent visibility is enabled, or that workers receive scans. This implementation does not deploy, modify live settings, or run a real scan. After deployment, first verify the read-only chat check above, then use a deliberately selected test scope to exercise one complete scan and separate import.
