# ARTEX and ScopeSentry integration design

## Intent and chosen approach

Operate the headless ScopeSentry service from ARTEX through asset sync, direct MCP, terminal commands, and chat. The operator requested a plan followed by implementation using the recommended approach. Keep one ScopeSentry MCP connection in ARTEX; do not add a second credential store or replace the scanner service.

Approaches considered: (1) direct MCP plus the existing sync adapter (selected primary path: structured tool schemas and no client installation), (2) CLI-only shell access (supported secondary path: useful for scripting but needs a compatible binary and separate CLI credential configuration), (3) a new REST-only integration (not selected: duplicates the existing MCP surface). Chat is an interface to these tools, not another transport.

## Goals

- Preserve one-way asset sync, and expose that same import through an opt-in ARTEX host tool for chat agents.
- Add `stop_scan_task` and `resume_scan_task` MCP tools and matching `task stop` / `task resume` CLI commands using existing task services. Return request acceptance, not scanner acknowledgement.
- Validate template existence and selected online nodes before accepting one-shot scans; guard missing templates and initialize URL target services. Reject recurring MCP requests explicitly rather than silently running them once. Recurring-job support is outside this change.
- Link the ARTEX ScopeSentry skill to the existing `ScopeSentry` MCP name. Keep permissions opt-in. Explain Auto (standalone Chat), Main Agent (task console), and independent Worker permissions.
- Provide reproducible CLI staging for native/Docker ARTEX use without committing binaries or secrets. Correct the ScopeSentry release build to use package entrypoints and include the CLI.

## Constraints and non-goals

- Use existing Go service APIs, MCP transports, and ARTEX host-tool/visibility patterns. No new libraries, database tables, or frontend redesign.
- No real scans, deployment, credential changes, database resets, commits, or automatic broad permission grants during implementation.
- Keep existing tool names and JSON output conventions. Tool discovery becomes 19 tools after lifecycle additions; clients must discover rather than require an exact count.
- Project scope replacement remains unchanged. Document that adding targets requires reading and preserving existing scope. Multiple targets must use actual newlines.
- The ARTEX skill must require a complete successful current-scope read before adding targets to an existing project; retain every existing target, merge new entries, and stop if the read fails or is incomplete. A checked workflow fixture must demonstrate two old targets surviving a third addition.
- Asset sync still imports subdomains, services, and apps (plus related IPs), capped at 5,000 per target/type. Vulnerability finding import is not part of this work.
- CLI settings remain independent of ARTEX's stored MCP headers. CLI URL is the service base URL; direct MCP URL ends in `/mcp`.

## Acceptance criteria

1. In-memory MCP tests discover 19 tools and exercise lifecycle validation, service forwarding, and failure propagation without MongoDB/Redis.
2. CLI tests cover stop/resume mapping, invalid IDs, backend errors, and JSON output; existing command tests remain green.
3. Scan requests reject missing templates, absent/offline nodes, and unsupported scheduling before persistence; tests use dependency fakes and require no live scanner.
4. ARTEX sync UI and host tool invoke one shared import function with the same validation, response, cap, and error handling. The host tool is bound through existing opt-in permissions.
5. ARTEX skill declares `mcps: ScopeSentry`; docs explain skill loading and deferred tool discovery, direct visibility, chat workflows, CLI installation, and separate asset import.
6. Targeted tests and builds pass, or external prerequisites and existing failures are reported by name. Live end-to-end testing remains a separate operator action.
7. A fake-MCP workflow test or explicitly labeled checked transcript covers project selection/creation, scope-preserving target update, node/template selection, task creation/status, task-name/project-ID result filters, and a separate asset sync. This does not claim a live LLM end-to-end test.

## Verification and rollout

Write regression tests first. Use in-memory transports, httptest servers, and fake service boundaries. Inspect all test files for database/config initialization before running broader suites; only database-free tests or explicit isolated fixtures may run. Build both ScopeSentry binaries into temporary paths (avoid macOS case-insensitive filename collisions). Run ARTEX's database-free packages and integration tests only against an explicitly provided test database, never the live store. Stage Linux CLI builds separately for Docker; do not infer running deployment versions from source tests.

After deployment, enable the ScopeSentry skill for Auto/Main Agent and the sync host tool where needed. Refresh tool discovery and use a new chat turn. Start with projects/nodes/templates discovery, then run one explicitly chosen test scan and inspect status/results. Sync results into ARTEX as a separate action.
