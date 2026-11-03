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
- [x] Review/stage integration source/docs/tests only; commit each repository and push `main` to its existing `origin`.
- [x] Compile default production server/CLI for Linux AMD64; transfer to a versioned staging directory and verify hashes.
- [x] Archive old binaries, install new server/CLI, restart only `scope-sentry`, and verify readiness plus MCP tool discovery without running scans.
- [x] Record commit IDs, rollback paths, health, unchanged dependency containers, and remaining ARTEX rollout step.

## Execution record

- Source pushed: ScopeSentry `bcdee3cdd7bec763d5c66de08a1f93db9030cc17`; ARTEX `761e6a9` and `4ecd1ee`.
- Build gate caught a macOS case-insensitive filename collision: `ScopeSentry` and `scopesentry` shared inode `122586772`, leaving both paths as the CLI. These artifacts were rejected before transfer.
- Ruling: use distinct artifact names `server-linux-amd64` and `cli-linux-amd64` — avoids the confirmed filesystem collision; deploying without package/hash checks could replace the server with the CLI.
- Rebuilt with empty tags, Linux AMD64, CGO disabled, and clean source revision `bcdee3c`; package metadata identifies `cmd/main` and `cmd/scopesentry` separately. Server SHA-256: `fbadcce8f2a39f1c1102521e90e429f89227e4feebdb4316c4cfdb8b80ebf1a2`; CLI SHA-256: `4d54f7b5e47a4617d7b7f00c18e95d60a5112a900a982736a31717e01e5dc5e7`.
- Pre-upgrade: `/live` and `/ready` return success; unauthenticated `/mcp` returns 401. Existing CLI configuration was not found in the documented home config path.
- Ruling: stop only the server before writing its bind-mounted binary in place, then start the same container — preserves the mounted inode and container-local config/key files. Atomic host-path replacement alone could leave an existing file bind mount on the old inode. Server-only downtime is required; restore the archived bytes on failure.
- Upgrade executed at `2026-11-03T14:20:13Z`. Same server container retained; server and CLI installed successfully. Transfer checks and `/proc/1/exe` confirm the expected server hash.
- `/live` returned `{"status":"ok"}`; `/ready` returned `{"status":"ready"}`. CLI reports `2.0.0-headless` and `server_reachable: true`. Missing and invalid API keys both return HTTP 401.
- Config and plugin-key SHA-256 fingerprints match before/after. Scanner, MongoDB, and Redis container IDs, start timestamps, and running status match before/after. No scan requests were sent.
- Rollback/staging root: `/opt/scopesentry/releases/20261103-bcdee3c`. Previous installed server and CLI are copied under `rollback/ScopeSentry` and `rollback/scopesentry-cli`; old `ScopeSentry.bak` and `ScopeSentry.bak.2` are moved there. `rollback.sha256` records archive checksums. No data directories, volumes, or credentials were removed.
- Task 4 is complete: binary installation, archive cleanup, readiness, auth rejection, preservation checks, and authenticated MCP discovery all passed. Operator connected ARTEX to `http://10.10.23.6:8082/mcp` and verified all 19 tools discovered live. ARTEX source documentation updated to 19 tools (`4ecd1ee`). ARTEX itself remains pushed but not deployed because its destination is not selected.
- Final recheck passed: health endpoints, running executable hash, preserved config/key fingerprints and dependencies, and all four rollback archive checksums. Server restart count is zero after the controlled stop/start; release and rollback directories have mode 0700.
- Independent deployment-document review found no blocking findings. The review confirmed the documented limits: no authenticated MCP discovery, no end-to-end sync claim, no ARTEX deployment, and no exercised rollback.

## Server/CLI rollback (only if needed)

Run on `10.10.23.6` as root. This stops only the server and restores the previous server and CLI. Keep the current container; do not recreate it.

```sh
set -eu
R=/opt/scopesentry/releases/20261103-bcdee3c
sha256sum -c "$R/rollback.sha256"
docker stop -t 30 scope-sentry
cat "$R/rollback/ScopeSentry" > /opt/scopesentry/ScopeSentry
chmod --reference="$R/rollback/ScopeSentry" /opt/scopesentry/ScopeSentry
cp -p "$R/rollback/scopesentry-cli" /usr/local/bin/scopesentry
docker start scope-sentry
```

After startup, check `curl -fsS http://127.0.0.1:8082/ready` and compare config/key fingerprints and dependency state against the saved `credentials.before` and `dependencies.before` files. The rollback copies remain available; rollback was not needed or exercised during this upgrade.
