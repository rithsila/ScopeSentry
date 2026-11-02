# ScopeSentry System Recommendations & Lightweight Architecture Analysis

## 1. Executive Summary

ScopeSentry is a distributed asset discovery and vulnerability scanner. It is designed to be lightweight, fast, and resource-efficient.

### Why ScopeSentry is Lightweight
1. **Single Compiled Go Binary**: The backend uses Go 1.24 (Gin). It compiles to a native binary with low memory usage (~30MB–80MB idle).
2. **Embedded Frontend**: The Vue 3 SPA is built and embedded directly into the Go binary (`//go:embed static/*`). You do not need Nginx to serve the UI.
3. **Distributed Master-Worker Model**: Heavy scanning tasks run on separate worker nodes (`ScopeSentry-Scan`). The central server only handles APIs and coordination.
4. **Fast Native Tools**: Core scanning relies on compiled Go and Rust engines (`RustScan`, `httpx`, `subfinder`, `katana`, `nuclei`) instead of heavy Python scripts.
5. **Configurable Goroutines**: Concurrency can be adjusted per module in `ModulesConfig` to prevent out-of-memory (OOM) errors on small servers.

---

## 2. Resource Footprint Breakdown

| Component | Technology | Idle RAM | Active RAM | Role |
| :--- | :--- | :--- | :--- | :--- |
| **ScopeSentry Server** | Go 1.24 + Gin (embedded Vue 3) | ~40 MB | ~100–250 MB | Web UI, API, Task Management, MCP |
| **Redis** | Redis 7 | ~15 MB | ~30–60 MB | Task Queues & Pub/Sub messaging |
| **MongoDB** | MongoDB 7 | ~150 MB | ~300–800 MB | Asset storage & scan results |
| **ScopeSentry-Scan** | Go Worker + CLI tools | ~50 MB | ~300 MB–1.5 GB | Port scan, crawling, nuclei tests |

---

## 3. Recommended System Specifications

### Option A: All-In-One Single Host (Smallest Budget / VPS)
Runs Server, MongoDB, Redis, and 1 Scan Node on the same machine.

- **CPU**: 1 vCPU or 2 vCPUs
- **RAM**: 2 GB (4 GB recommended if running concurrent Nuclei scans)
- **Disk**: 20 GB SSD
- **Swap**: 2 GB swap recommended on 2 GB RAM machines
- **Target Workload**: Personal use, continuous monitoring of 5–20 domains, light scheduled scans

### Option B: Production Single Host (Recommended)
- **CPU**: 2–4 vCPUs
- **RAM**: 4 GB–8 GB
- **Disk**: 50 GB+ SSD
- **Target Workload**: Active discovery, high concurrency, thousands of subdomains, full port scans

### Option C: Distributed Deployment (Scalable Architecture)
- **Master Node (Server + MongoDB + Redis)**:
  - 2 vCPUs, 2 GB–4 GB RAM, 40 GB SSD
- **Worker Scanning Nodes (`ScopeSentry-Scan`)**:
  - 1–2 vCPUs, 1 GB–2 GB RAM each
  - Can be added or destroyed dynamically on any cloud VPS (e.g., LightNode, DigitalOcean, Hetzner)

---

## 4. How to Keep Memory Low (`ModulesConfig`)

You can control memory and CPU usage directly in the web UI under **System Configuration** or via `ModulesConfig`.

Recommended settings for a 2 GB RAM VPS:

```yaml
maxGoroutineCount: 2       # Max simultaneous targets
subdomainScan:
  goroutineCount: 2        # Subdomain discovery threads
assetMapping:
  goroutineCount: 3        # HTTP probing concurrency
portScan:
  goroutineCount: 1        # Keep port scanner lightweight
dirScan:
  goroutineCount: 2        # Directory scanning threads
vulnerabilityScan:
  goroutineCount: 1        # Nuclei concurrency (prevents RAM spikes)
webCrawler:
  goroutineCount: 1        # Crawler threads
```

---

## 5. Deployment Verification Commands

Run these commands to verify container resource usage on your server:

```bash
# Check running containers and resource usage in real time
docker stats

# Run the single host stack
docker compose -f single-host-deployment.yml up -d
```
