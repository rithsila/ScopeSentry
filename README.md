# ScopeSentry (Headless & Agent-First)

English | [中文](./README_CN.md)

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Autumn-27/ScopeSentry-Scan)

## Introduction
Scope Sentry is a tool with functions such as asset mapping, subdomain enumeration, information leakage detection, vulnerability scanning, directory scanning, subdomain takeover, crawler, and page monitoring. By building multiple nodes, users can freely choose nodes to run scanning tasks. When new vulnerabilities emerge, it can quickly check whether the concerned assets have related components.

Distributed Implementation Reference Articles: [https://mp.weixin.qq.com/s/xfgRxUjljoQ8KzacblktxA](https://mp.weixin.qq.com/s/xfgRxUjljoQ8KzacblktxA)

Server Recommendation: [lightnode](https://www.lightnode.com/?inviteCode=CQ11JU&promoteWay=LINK)

## Discord:

[https://discord.gg/GWVwSBBm48](https://discord.gg/GWVwSBBm48)

## Architecture
- **Server**: Go (Gin) Headless Server with native Model Context Protocol (MCP) server & REST API
- **CLI Client**: Standalone Go CLI client (`scopesentry`)
- **Scan Node**: Go distributed scanning workers (`ScopeSentry-Scan`)
- **Databases**: MongoDB & Redis

## Agent & Automation Support
ScopeSentry is designed for AI agents and automated attack surface workflows:
- **Built-in MCP Server**: Exposes 19 security scanning and asset management tools at `/mcp`, including scan stop/resume. Discover the deployed server's tool list at runtime.
- **Standalone CLI**: `scopesentry` connects to `/mcp` with pure machine-readable JSON stdout.
- **Headless Admin**: Generate API keys without a browser via `./ScopeSentry apikey create --name <name>`.
- **Guides**: See [AI_AGENT_GUIDE.md](./AI_AGENT_GUIDE.md) and [SKILL.md](./SKILL.md).

## Website

- Official Website: [https://www.scope-sentry.top](https://www.scope-sentry.top/en/)
- Github: [https://github.com/Autumn-27/ScopeSentry](https://github.com/Autumn-27/ScopeSentry)
- Scanner source code: [https://github.com/Autumn-27/ScopeSentry-Scan](https://github.com/Autumn-27/ScopeSentry-Scan)
- UI source code: [https://github.com/Autumn-27/ScopeSentry-UI](https://github.com/Autumn-27/ScopeSentry-UI)
- Plugin Market: [Plugin Market](https://plugin.scope-sentry.top/en)
- Plugin Template：[https://github.com/Autumn-27/ScopeSentry-Plugin-Template](https://github.com/Autumn-27/ScopeSentry-Plugin-Template)

## Install
```
git clone https://github.com/Autumn-27/ScopeSentry.git
cd ScopeSentry
# Change the MongoDB and Redis account passwords in the. env file.
docker-compose -f single-host-deployment.yml up -d
```
After running, there will be four containers: mongodb, redis, scope-sentry (server), and scopesentry-scan (scan). By default, there will be one scanning node.

Create an API key or retrieve the plugin verification key from the configured server container. Store the output securely; do not paste keys into chat or commit them.

```bash
docker exec scope-sentry ScopeSentry apikey create --name artex
docker exec scope-sentry ScopeSentry pluginkey
```

**Add new nodes(Optional)**

```
git clone https://github.com/Autumn-27/ScopeSentry-Scan.git
cd ScopeSentry-Scan/build
# Edit MongoDB and Redis connection settings in the .env file.
# Set a stable, unique NodeName before starting the scanner.
docker-compose -f scan-docker-compose.yml up -d
```

## Plugin Flowchart

<img src="流程图.svg"/>

## Current Features
- Plugin System (Add any tool through extension)
- Subdomain Enumeration
- Subdomain Takeover Detection
- Port Scanning
- Asset Identification
- Directory Scanning
- Vulnerability Scanning
- Sensitive Information Leakage Detection
- URL Extraction
- Crawler
- Page Monitoring
- Custom WEB Fingerprint
- POC Import
- Asset Grouping
- Multi-Node Scanning
- Webhook

## To Do
- Weak Password Cracking
- 
## Installation

For installation instructions, see the [official website](https://www.scope-sentry.top)

## Communication

Discord:

[https://discord.gg/agsYdAyN](https://discord.gg/agsYdAyN)


## Headless CLI & MCP Usage

ScopeSentry operates as a headless scanner server with Model Context Protocol (MCP) and a dedicated CLI.

### Health Check & Diagnostics
```bash
scopesentry doctor
```

### Headless Admin Credentials
```bash
# Create an admin API key
./ScopeSentry apikey create --name admin

# List API keys
./ScopeSentry apikey list

# Fetch plugin key
./ScopeSentry pluginkey
```

### CLI Asset & Task Commands
```bash
# List projects
scopesentry project list

# List assets with query syntax
scopesentry asset list --type subdomain --search 'domain="example.com"'

# Create a project
scopesentry project create --name "example" --target "example.com"

# Request scan lifecycle changes
scopesentry task stop <task-id>
scopesentry task resume <task-id>
```

### Use from ARTEX Chat, MCP, Sync, or terminal

Reuse ARTEX's existing MCP entry named `ScopeSentry`. Asset Sync imports existing remote assets; it does not launch scans. Enable the ScopeSentry skill for **Auto** (standalone Chat) or **Main Agent** (task console), and load the skill before calling its linked MCP tools. Bind ARTEX's opt-in `sync_scopesentry_assets` tool separately when chat should import assets. Worker access is a separate permission.

Direct MCP uses `http://<host>:8082/mcp`. CLI configuration uses the base URL `http://<host>:8082`, plus `SCOPESENTRY_API_KEY` injected into the execution environment. The CLI must be installed on the ARTEX backend host/container. The server, MongoDB, Redis, and scanner nodes remain necessary.

Discover tools at runtime (19 in this version). Scan creation requires an existing template and online nodes; recurring MCP/CLI requests are rejected. A returned task ID does not prove scanner dispatch succeeded. Stop/resume confirms request acceptance rather than scanner acknowledgement. See [ARTEX setup](docs/artex-integration.md), [AI agent guide](AI_AGENT_GUIDE.md), [MCP/CLI reference](SKILL.md), and the [integration plan](docs/superpowers/plans/2026-11-03-artex-integration.md).


#License

All branches of this project follow AGPL-3.0, and additional terms need to be followed:
1. The commercial use of this software requires a separate commercial license.
2. Companies, organizations, and for-profit entities must obtain a commercial license before using, distributing, or modifying this software.
Individuals and non-profit organizations are free to use this software in accordance with the terms of AGPL-3.0.
4. If you have any commercial license inquiries, please contact rainy-autumn@outlook.com .
