# ScopeSentry Troubleshooting Guide

This guide helps you diagnose, fix, and maintain ScopeSentry running on your VM (`10.10.23.6`).

---

## 1. Quick Status Checks

Run these commands on the VM to check system health in under 10 seconds:

```bash
# Check container status
cd /opt/scopesentry && docker compose ps

# Check live memory and CPU usage
docker stats --no-stream

# Check if port 8082 is answering locally
curl -I http://127.0.0.1:8082
```

---

## 2. Inspecting Container Logs

If something is not working, check the logs of the specific container:

```bash
# View ScopeSentry Server logs
docker logs scope-sentry --tail 50 -f

# View Scanner Node logs (task execution and errors)
docker logs scopesentry-scan --tail 50 -f

# View MongoDB database logs
docker logs scopesentry-mongodb --tail 50

# View Redis task queue logs
docker logs scopesentry-redis --tail 50
```

---

## 3. Top 5 Common Issues & Exact Fixes

### Issue 1: Cannot Open Web Dashboard (`http://10.10.23.6:8082`)
- **Cause**: Container is stopped, or the firewall (UFW) is blocking port 8082.
- **Fix**:
  ```bash
  # 1. Verify container is running
  cd /opt/scopesentry && docker compose up -d scope-sentry

  # 2. Check and allow firewall rule
  ufw allow 8082/tcp comment "ScopeSentry Web and MCP"
  ufw reload
  ```

---

### Issue 2: Scanner Node Shows Offline or Red in Web UI
- **Cause**: Redis connection dropped or scanner crashed from an external tool issue.
- **Fix**:
  ```bash
  # Restart the scanner container
  docker restart scopesentry-scan

  # Verify scanner registered in Redis logs
  docker logs scopesentry-scan --tail 20
  ```

---

### Issue 3: ARTEX Fails to Sync Assets
- **Cause**: Wrong MCP URL or missing/incorrect API key.
- **Fix**:
  1. Open ScopeSentry web interface (`http://10.10.23.6:8082`).
  2. Go to **API Key** -> Click **Create Key** -> Copy the string starting with `ssk_...`.
  3. In ARTEX **Asset Sync**, enter:
     - **MCP Endpoint**: `http://10.10.23.6:8082/mcp`
     - **API Key**: `ssk_...`
  4. Test connectivity from the ARTEX host:
     ```bash
     curl -s -H "X-API-Key: YOUR_API_KEY" http://10.10.23.6:8082/mcp
     ```

---

### Issue 4: High Memory Usage or Out-of-Memory (OOM)
- **Cause**: Too many scanner threads running at the same time.
- **Fix**:
  1. In the Web UI, go to **System Configuration** -> **Module Config**.
  2. Reduce concurrency settings:
     - Change `maxGoroutineCount` from `3` to `1` or `2`.
     - Change `vulnerabilityScan.goroutineCount` from `2` to `1`.
     - Change `dirScan.goroutineCount` from `3` to `1`.
  3. Save the configuration.

---

### Issue 5: Lost Admin Password or Plugin Key
- **Cause**: The initialization logs rolled off screen.
- **Fix**:
  Retrieve passwords from the persistent container files:
  ```bash
  # Print the initial admin password
  docker exec scope-sentry cat /opt/ScopeSentry/PASSWORD

  # Print the initial plugin key
  docker exec scope-sentry cat /opt/ScopeSentry/PLUGINKEY
  ```

---

## 4. Full Restart & Recovery

If the entire stack becomes unresponsive:

```bash
cd /opt/scopesentry

# 1. Stop all containers cleanly
docker compose down

# 2. Start all services fresh
docker compose up -d

# 3. Verify all 4 containers are running
docker compose ps
```

---

## 5. Persistent Data Directories

All scan data and configurations are stored on the VM host at:
- Database files: `/opt/scopesentry/data/mongodb`
- Redis queues: `/opt/scopesentry/data/redis/data`
- Exported scans: `/opt/scopesentry/data/files`
- Screenshots: `/opt/scopesentry/data/images`
- Uploaded plugins: `/opt/scopesentry/data/uploads`
- Credentials: `/opt/scopesentry/.env`
