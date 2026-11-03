# ARTEX and ScopeSentry Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Support ARTEX sync, direct MCP, terminal, and chat control of ScopeSentry with reliable one-shot scan validation and lifecycle commands.

**Architecture:** Preserve the existing HTTP MCP server and shared ARTEX connection. Extend ScopeSentry through its task services; reuse ARTEX's import implementation for an opt-in host tool. Keep CLI distribution reproducible and permissions explicit.

**Tech Stack:** Go 1.24+ (ScopeSentry), Go 1.26+ (ARTEX), modelcontextprotocol/go-sdk, Norma host tools, existing MongoDB/Redis/PostgreSQL service boundaries.

**Spec:** `docs/superpowers/specs/2026-11-03-artex-integration-design.md`

## Global Constraints

- No new dependencies, schema changes, frontend redesign, automatic permission grants, live scans, deployment, or commits.
- Keep both checkouts in place; use disjoint file ownership for parallel work and preserve any unrelated changes.
- Keep pure JSON stdout for CLI commands; errors and diagnostics go to stderr.
- Add exactly two lifecycle MCP tools; preserve all existing names and discover the resulting tool count dynamically in clients.
- Reject recurring MCP requests explicitly; existing REST scheduling is unchanged.

## Review Focus

- Invalid/blank/multiple lifecycle IDs must not reach task services unexpectedly.
- Missing templates and offline node names must fail before background dispatch and persistence.
- Sync tool calls must respect cancellation, validate selection, and reuse the UI's cap and import behavior.
- Skill-gated MCP tools remain unavailable until the visible skill is loaded; existing permissions are never silently widened.
- Docker CLI architecture and secret handling must match the execution environment; no keys or generated binaries enter Git.

---

### Task 1: ScopeSentry lifecycle MCP and CLI

**Files:** `internal/mcp/task_lifecycle.go` (new), `internal/mcp/tools.go`, `internal/mcp/mcp_test.go`, lifecycle tests (new); `cmd/scopesentry/commands/task.go`, `root.go`, command tests.

**Interfaces:** MCP `stop_scan_task({id:string})` / `resume_scan_task({id:string})` return `{success:true,id:string,message:string}` after `task.Service.StopTasks` / `StartTasks` accepts the request. CLI `task stop <id>` / `task resume <id>` call those tools. Invalid ObjectIDs fail before service calls.

- [x] Add in-memory registry and lifecycle tests using a fake task service; assert invalid IDs, exact forwarding, and service error propagation. Run and observe failure.
- [x] Register and implement the two handlers in a focused file, reusing task services and MCP error formatting.
- [x] Add CLI fake-MCP tests for positional/flag IDs, empty/invalid IDs, stop/resume routing, and backend errors. Observe failure before implementing commands.
- [x] Implement CLI parsing and help changes; verify through `bash scripts/test-isolated.sh ./internal/mcp ./cmd/scopesentry/...` (supersedes the original unprotected test command).

### Task 2: One-shot scan preflight

**Files:** `internal/mcp/tools.go`, focused preflight tests; `internal/services/task/common/common.go`, focused service tests; release configuration `.goreleaser.yaml`.

**Interfaces:** `create_scan_task` rejects scheduling, validates template ObjectID/existence and online node names before `Insert`. `allNode` resolves only online names and does not expand to offline nodes. `GetScanTemplate` returns an error for nil template instead of dereferencing it. URL service is initialized by `NewService`.

- [x] Add tests asserting no insertion after invalid/missing template, offline/unknown/empty nodes, or recurring requests; test online-only all-node selection. Observe failure.
- [x] Implement preflight using existing template and node service dependencies. Preserve general/project source behavior and task name uniqueness.
- [x] Add service tests for nil-template error and initialized URL dependency; implement the minimum fix. Verify task construction and binding fields remain compatible.
- [x] Fix release builds to compile `./cmd/main` and `./cmd/scopesentry`; run targeted and full isolated tests, focused/full `go vet`, and build server/CLI into temporary paths. Full vet's 141 baseline diagnostics remain documented.

Suite safety gate: inspect transitive package initialization, not only test bodies. Review found MongoDB/Redis `init()` connection attempts and config/key file writes before test fakes. Earlier unguarded ScopeSentry runs do not establish isolation. Do not repeat them. Use the explicit compile-time `scopesentry_unit` bootstrap and verified isolated runner once implemented; normal server builds must retain production startup. Never execute a suite that could reach a runtime store without isolation.

### Task 3: ARTEX shared sync operation and chat tool

**Files (ARTEX):** `server/sync_scopesentry.go`, new sync-tool/test file, `server/orchestration.go`, `agent/toolcatalog.go`, existing seed tests if needed.

**Interfaces:** Shared sync operation accepts `context.Context` and the existing selection request, returning the current response structure and errors. HTTP handler and opt-in `sync_scopesentry_assets` host tool both call it. Input keeps `dimension`, `targets`, `asset_types`, `create_company`, and `page_size` semantics. No automatic tool bindings.

- [x] Add tests for empty/invalid selection, correct project/task filters, context cancellation, and host tool JSON schemas/results with fake boundaries. Observe failing behavior before edits.
- [x] Extract the existing import body into one shared operation without altering cap, ingestion, response, or company behavior. Keep HTTP errors compatible.
- [x] Add the host tool using `wrTool` and register metadata in the existing catalog with opt-in binding.
- [x] Test tool registration/schema without a database; run existing server tests only with an explicit dedicated test DSN, otherwise document the prerequisite.

### Task 4: ARTEX skill, CLI staging, and setup instructions

**Files (ARTEX):** `skills/scopesentry/SKILL.md`, `SCOPESENTRY_GUIDE.md`, new `scripts/build-scopesentry-cli.sh`, `tools/bin/.gitignore`, `Dockerfile`, integration guide/tests.

**Interfaces:** Skill frontmatter declares `mcps: ScopeSentry`. Staging script accepts a local ScopeSentry checkout, validates requested OS/architecture, and builds `scopesentry-<os>-<arch>` to an ignored staging directory. Docker optionally installs staged `scopesentry-linux-${TARGETARCH}`; missing binary preserves MCP-only builds.

- [x] Add database-free tests for skill linkage, documented flags, staging script validation, and optional Docker installation contract. Observe failure.
- [x] Link the skill and correct CLI flags, newline targets, tool count assumptions, scheduling/lifecycle semantics, and supported result filters.
- [x] Implement staging script without downloads, secrets, or committed binaries; build native and Linux artifacts as verification allows.
- [x] Document setup for Auto/Main Agent, Worker opt-in, deferred tool discovery, host-tool binding, direct MCP versus CLI URLs/credentials, and sync after scans. Include read-only verification prompts and terminal commands.
- [x] Add a checked workflow fixture/transcript covering create/select project, read complete scope, preserve two existing targets while adding a third, list online nodes and templates, create a one-shot task, inspect status, filter results by task name/project ObjectID, then explicitly sync assets. Label it as contract coverage rather than a live LLM test.

### Task 5: Review, verification, and handoff

**Files:** design/plan progress, ScopeSentry `SKILL.md`, `AI_AGENT_GUIDE.md`, `README.md`, ARTEX integration docs.

- [x] Run focused regression suites, both builds, and safe database-free ARTEX tests. Run full suites where feasible; name prerequisite failures without modifying runtime databases.
- [x] Request independent code review across both repositories. Fix concrete issues and rerun affected tests. Final startup correction independently approved; protected normal/race suites pass.
- [x] Check diffs for credentials, generated binaries, accidental permission grants, schema changes, and unrelated edits.
- [x] Update plan checkboxes with actual evidence and list remaining live configuration/deployment steps. Results/limits: `docs/artex-integration-verification.md`; rollout: `docs/artex-integration.md`. Nothing deployed or committed.

## Execution authorization

The operator asked to write the plan and then proceed with the recommendation. Execute in this session using independent subagents for disjoint repository work, with local review and regression tests. Do not stop for a second planning approval. Do not commit or deploy.
