#!/usr/bin/env bash
# ==============================================================================
# ScopeSentry Update & Upgrade Script
# Target: Debian 13 VM (10.10.23.6)
# Location: /opt/scopesentry/update.sh
# ==============================================================================
set -euo pipefail

DEPLOY_DIR="/opt/scopesentry"

echo "=== [1/4] Updating Debian System Packages ==="
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get dist-upgrade -y
apt-get autoremove -y
apt-get clean

echo "=== [2/4] Pulling Latest ScopeSentry Docker Images ==="
cd "$DEPLOY_DIR"
docker compose pull

echo "=== [3/4] Recreating Containers with Updated Images ==="
docker compose up -d --remove-orphans

echo "=== [4/4] Verifying Service Health ==="
sleep 5
docker compose ps

echo ""
echo "Testing local HTTP health check on port 8082..."
if curl -s -f -I http://127.0.0.1:8082 > /dev/null; then
    echo "✅ ScopeSentry Web UI is responding (HTTP 200 OK)."
else
    echo "⚠️ Warning: ScopeSentry is still starting up or port 8082 is unreachable."
fi

echo ""
echo "=== Update Completed Successfully ==="
