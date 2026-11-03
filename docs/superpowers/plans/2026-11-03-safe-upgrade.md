# ScopeSentry Safe Upgrade Implementation Plan

> **For agentic workers:** Execute inline. Steps use checkbox syntax for tracking.

**Goal:** Commit and push the reviewed integration in both repositories, then upgrade only the existing ScopeSentry server and CLI on `10.10.23.6`.

**Architecture:** Keep the current Compose deployment and data services. Replace its bind-mounted server binary, restart only the server, and install the CLI. Archive old binaries for rollback; never delete data or change credentials. ARTEX deployment is excluded because no destination was selected.

**Tech Stack:** Go production Linux AMD64 builds, SSH, Docker Compose, Git.

**Spec / approval:** Operator selected “Safe upgrade”: preserve scanner and all data, keep rollback copy, commit/push both repositories, no ARTEX deployment to an unknown host.

## Global constraints

- Never deploy `scopesentry_unit` builds. No test execution against runtime databases.
- No database/volume deletion, credential rotation, scan dispatch, global Docker pruning, force pushes, or unrelated cleanup.
- Keep existing container configuration and key files; restart the server only.
- Changelog files remain unchanged; no separate changelog update was approved.

## Review focus

- Git remote/branch mismatch: verify current `main` and `origin`, reject non-fast-forward updates.
- Wrong binary: compile without test tag, check Linux AMD64 and SHA-256 after transfer.
- Lost credentials: use restart, not recreation; retain config/key files and fingerprint before/after without exposing contents.
- Data/scanner disruption: compare container IDs and start timestamps before/after; do not restart dependencies.
- Failed readiness: restore archived server and restart only server; do not remove rollback files until accepted.

## Tasks

- [x] Inspect both repositories and destination; rerun protected ScopeSentry suite and safe ARTEX checks.
- [ ] Review/stage integration source/docs/tests only; commit each repository and push `main` to its existing `origin`.
- [ ] Compile default production server/CLI for Linux AMD64; transfer to a versioned staging directory and verify hashes.
- [ ] Archive old binaries, install new server/CLI, restart only `scope-sentry`, and verify readiness plus MCP tool discovery without running scans.
- [ ] Record commit IDs, rollback paths, health, unchanged dependency containers, and remaining ARTEX rollout step.
