# Blueprint — Deployment Guide

## Overview

Blueprint deploys as two Go binaries + Vue 3 static frontend on a VPS with Nginx reverse proxy.

```
Internet → Nginx (SSL + Brotli) → Go API :8080
                                 → Health Monitor :8081
                                 → Frontend (static files)
                                 → pgweb :8082
                                 → Grafana :3000
```

---

## 1. VPS Provisioning

### Automated (recommended)

From your local machine:

```bash
bash scripts/setup-vps-runner.sh user@your-vps [db_password] [redis_password]
```

This script:
1. Copies `setup-vps.sh` + systemd units to the VPS
2. Executes via `ssh -t` with sudo
3. Downloads the generated `.env.production` back to `scripts/.env.production.local`

If passwords are omitted, they're generated automatically with `openssl rand`.

### What gets installed

| Software | Method | Version |
|----------|--------|---------|
| Go | Official tarball | Latest stable |
| Bun | Official installer | Latest |
| PostgreSQL | pgdg apt repo | 16 |
| Redis | apt | 7 |
| Nginx | ppa:ondrej/nginx-mainline | Latest |
| Brotli module | libnginx-mod-http-brotli-* | Matches Nginx |
| Certbot | snap | Latest |
| pgweb | GitHub releases binary | Latest |
| Prometheus | Official binary | Latest LTS |
| Node Exporter | Official binary | Latest |
| Grafana | Official apt repo | Latest |

### What gets configured

- System user `blueprint` (no login shell)
- Directories: `/opt/blueprint/{api,health,frontend,backups,uploads,logs,scripts}`
- Systemd units: `blueprint-api.service`, `blueprint-health.service`
- PostgreSQL: user `blueprint` + database `blueprint`
- Redis: password authentication
- UFW firewall: 22, 80, 443 open; 8081/3000/9090 restricted
- `.env.production` with all credentials and config

### Manual provisioning

If you prefer to set up manually, run on the VPS:

```bash
sudo bash scripts/setup-vps.sh [db_password] [redis_password]
```

Or for just the dependencies (no DB/service configuration):

```bash
sudo bash scripts/install.sh
```

---

## 2. Nginx + SSL

```bash
sudo bash /opt/blueprint/scripts/setup-nginx.sh your-domain.com
```

### What it configures

- **Reverse proxy**: `/api/` → `:8080`, `/health` → `:8081`, `/pgweb/` → `:8082`, `/grafana/` → `:3000`
- **Static files**: Frontend served from `/opt/blueprint/frontend/`
- **Compression**: Brotli (level 6) + gzip fallback
- **SSL**: Certbot with auto-renewal (optional, interactive prompt)
- **HTTP/2**: Enabled
- **Rate limiting**: 30r/s for API, 5r/s for auth endpoints
- **Security headers**: HSTS, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy
- **Cache**: 1 year for static assets (immutable), no-cache for service worker
- **SSE**: Buffering disabled for `/api/v1/admin/logs/stream`
- **SPA**: Fallback to `index.html` for client-side routing
- **Uploads**: `/static/` aliased to `/opt/blueprint/uploads/`

---

## 3. Monitoring Setup

```bash
sudo bash /opt/blueprint/scripts/setup-monitoring.sh
```

### Prometheus

Scrapes:
- Node Exporter (`:9100`) — host metrics (CPU, memory, disk, network)
- Blueprint API (`:8080/healthz`) — application health
- Health Monitor (`:8081/health?format=json`) — service health checks

Config at `/etc/prometheus/prometheus.yml`.

### Grafana

- Accessible at `https://your-domain.com/grafana/`
- Default credentials: admin/admin (change on first login)
- Prometheus datasource auto-configured

### Node Exporter

Exposes host metrics for Prometheus on `:9100`.

---

## 4. Deploy

### How it works

Production deploys use **immutable artifacts**. CI builds everything (binaries, frontend, Grafana config) and uploads a complete release bundle to the VPS. A promotion script on the server activates it — no build tools needed on the host.

```
GitHub Actions (CI) → build release → rsync to VPS → deploy-native.sh → health check
```

### Automatic deploy (recommended)

Pushing to `main` triggers the pipeline automatically:

1. **CI** runs security, build, lint, test, e2e, and semantic-release.
2. On success, **deploy.yml** builds a release bundle:
   - `bin/blueprint-api` + `bin/blueprint-health` (Linux amd64, static)
   - `frontend/` (Vue 3 dist)
   - `grafana/` (dashboards + provisioning)
3. Uploads to `/opt/blueprint/releases/incoming/<sha>/`
4. Runs `deploy-native.sh <sha>` on the VPS
5. Health check passes → done. Fails → automatic rollback.

### Secrets required

| Secret | Purpose |
|--------|---------|
| `VPS_SSH_KEY` | SSH private key for the VPS |
| `VPS_HOST` | VPS hostname or IP |
| `VPS_USER` | SSH user (default: `blueprint`) |

### Manual deploy (fallback)

For emergencies or when CI is unavailable, the old local-build-and-upload flow still works:

```bash
cp scripts/.deploy.env.example scripts/.deploy.env
# edit .deploy.env with your VPS credentials

make deploy           # Deploy backend + frontend
make deploy-backend   # Backend only
make deploy-frontend  # Frontend only
make deploy-dry       # Dry run (preview)
```

### Deploy flow — automatic (deploy.yml)

1. **Checkout** code at the commit SHA that passed CI
2. **Build** Go binaries with `CGO_ENABLED=0 -trimpath -ldflags="-s -w"`
3. **Build** frontend with `bun run build`
4. **Assemble** Grafana dashboards + provisioning
5. **Upload** entire release to `/opt/blueprint/releases/incoming/<sha>/`
6. **Activate** by running `deploy-native.sh <sha>` on the VPS

### Deploy flow — deploy-native.sh (on VPS)

1. **Validate** all expected binaries exist in the incoming directory
2. **Stage** the release to a temporary staging directory
3. **Preserve** current binaries as `<binary>.previous` for rollback
4. **Swap** binaries atomically (`install` + `mv`)
5. **Sync** frontend assets (`rsync --delete`)
6. **Sync** Grafana dashboards + provisioning (if Grafana is installed)
7. **Restart** services: `blueprint-api`, `blueprint-health`, `grafana-server`, nginx
8. **Health check** — 30 retries, 2s interval, `curl /healthz`
9. **Rollback** automatically if health check fails
10. **Cleanup** incoming directory and prune releases older than 14 days

### Deploy flow — manual (deploy.sh)

1. **Build** locally: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"`
2. **Upload** via rsync to VPS `/tmp/blueprint-deploy/`
3. **Backup** current binaries to `.bak`
4. **Replace** with new binaries
5. **Restart** systemd services
6. **Health check** — 10 retries, 2s interval, `curl /healthz`
7. **Auto-rollback** if health check fails (restore `.bak`, restart)
8. **Cleanup** temporary files

### Deploy flow — Frontend (manual only)

1. **Build** locally: `bun run build`
2. **Upload** via rsync `dist/` to VPS `/opt/blueprint/frontend/`
3. **Reload** Nginx

---

## 5. Service Control

```bash
bash /opt/blueprint/scripts/start.sh          # Start all (api + health)
bash /opt/blueprint/scripts/start.sh api      # Start API only
bash /opt/blueprint/scripts/start.sh health   # Start health monitor only
bash /opt/blueprint/scripts/stop.sh           # Stop all
bash /opt/blueprint/scripts/restart.sh        # Restart all
```

---

## 6. Rollback

### Automatic (deploy-native.sh)

If the health check fails after a deploy, `deploy-native.sh` automatically restores the `.previous` binaries from the release directory and restarts the services.

### Manual

```bash
bash /opt/blueprint/scripts/rollback.sh
```

Restores previous binaries, restarts services, runs health check.

### From a specific release

Releases are kept for 14 days under `/opt/blueprint/releases/`. To rollback to a specific release:

```bash
ls /opt/blueprint/releases/          # find the SHA you want
bash /opt/blueprint/scripts/deploy-native.sh <sha>
```

---

## 7. Health Checks

```bash
bash /opt/blueprint/scripts/health.sh
```

| Check | What it tests | Fail condition |
|-------|---------------|----------------|
| API | `curl /healthz` | HTTP error or timeout |
| PostgreSQL | `pg_isready` | Connection failed |
| Redis | `redis-cli ping` | No PONG |
| Disk | `df /` | Usage > 80% |
| Backup | Last `.dump.gz` age | Older than 25 hours |
| Health Monitor | `curl :8081/health` | HTTP error |

Output: `{"checks":[...],"overall":"healthy|unhealthy"}`

Exit code: 0 = healthy, 1 = unhealthy.

### Telegram alerts

```bash
bash /opt/blueprint/scripts/health.sh --telegram
```

Requires `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` in env.

---

## 8. Backup

```bash
bash /opt/blueprint/scripts/backup.sh
```

- **Method**: `pg_dump -Fc | gzip`
- **Location**: `/opt/blueprint/backups/blueprint_YYYYMMDD_HHMMSS.dump.gz`
- **Retention**: 7 daily, 4 weekly, 12 monthly
- **S3 upload**: If `AWS_BUCKET` env is set
- **Logs**: `/opt/blueprint/logs/backup.log`

### Restore

```bash
gunzip -c /opt/blueprint/backups/blueprint_20260409_020000.dump.gz | pg_restore -d blueprint
```

---

## 9. Monitoring (Cron)

```bash
bash /opt/blueprint/scripts/monitor.sh
```

6 checks: API, PostgreSQL, Redis, Disk, Memory, SSL cert expiry. JSON output. Telegram alert on failure. Run via cron every 5 minutes.

---

## 10. Performance Audit

```bash
bash /opt/blueprint/scripts/check-perf.sh your-domain.com
```

Tests: Brotli, gzip, HTTP/2, Cache-Control, HSTS, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, CSP.

---

## 11. Cron Setup

```bash
crontab scripts/crontab.example
```

```cron
0 2 * * *    backup.sh        # Database backup daily at 2 AM
*/5 * * * *  monitor.sh       # Monitoring every 5 minutes
0 4 * * 0    certbot renew    # SSL renewal check weekly
0 3 1 * *    log cleanup      # Delete logs older than 30 days
```

---

## 12. Systemd Services

### blueprint-api.service

```ini
[Unit]
Description=Blueprint API Server
After=network.target postgresql.service redis-server.service

[Service]
Type=simple
User=blueprint
WorkingDirectory=/opt/blueprint/api
EnvironmentFile=/opt/blueprint/.env.production
ExecStart=/opt/blueprint/api/blueprint-api
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

### blueprint-health.service

```ini
[Unit]
Description=Blueprint Health Monitor
After=network.target

[Service]
Type=simple
User=blueprint
WorkingDirectory=/opt/blueprint/health
EnvironmentFile=/opt/blueprint/.env.production
ExecStart=/opt/blueprint/health/blueprint-health
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

---

## 13. Environment Variables (Production)

Generated by `setup-vps.sh` at `/opt/blueprint/.env.production`:

```env
DATABASE_URL=postgres://blueprint:PASSWORD@localhost:5432/blueprint?sslmode=disable
DATABASE_MIGRATION_URL=postgres://blueprint:PASSWORD@localhost:5432/blueprint?sslmode=disable
REDIS_URL=redis://:PASSWORD@localhost:6379
JWT_SECRET=GENERATED_SECRET
PORT=8080
ENV=production
FRONTEND_URL=https://your-domain.com
STORAGE_TYPE=local
UPLOAD_DIR=/opt/blueprint/uploads
PGWEB_URL=http://localhost:8082
GRAFANA_URL=http://localhost:3000/grafana
PROMETHEUS_URL=http://localhost:9090
SMTP_HOST=smtp.example.com
SMTP_PORT=587
STRIPE_KEY=
STRIPE_WEBHOOK_SECRET=
OPENAI_KEY=
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=

# Rate Limiting (defaults — also configurable via Admin > Security)
RATE_LIMIT_API=60
RATE_LIMIT_AUTH=10
RATE_LIMIT_REGISTER=5
RATE_LIMIT_FORGOT=3
MAX_REQUEST_BODY_MB=10

# Email Verification (also toggleable via Admin > Feature Flags)
EMAIL_VERIFICATION_REQUIRED=false
```

---

## 14. VPS Directory Structure

```
/opt/blueprint/
├── api/
│   ├── blueprint-api          # Go binary (live)
├── health/
│   ├── blueprint-health       # Health monitor binary (live)
├── frontend/                  # Vue 3 dist (static)
├── uploads/                   # User uploads
├── backups/                   # pg_dump files
├── logs/                      # App logs
├── scripts/                   # Operational scripts (including deploy-native.sh)
├── releases/
│   ├── incoming/              # Staging area for next deploy (empty after promote)
│   ├── <sha>/                 # Promoted release (keeps .previous for rollback)
│   │   ├── bin/
│   │   │   ├── blueprint-api.previous
│   │   │   └── blueprint-health.previous
│   │   ├── frontend/
│   │   └── grafana/
│   └── staging-<sha>/         # Transient, cleaned up by trap
└── .env.production            # Environment config
```

---

## Scripts Reference

| Script | Runs on | Purpose |
|--------|---------|---------|
| `setup-vps-runner.sh` | Local | Orchestrates VPS provisioning via SSH |
| `setup-vps.sh` | VPS | Full setup (packages, DB, services, firewall) |
| `install.sh` | VPS | Lightweight dependency installer |
| `setup-nginx.sh` | VPS | Nginx + SSL + Brotli config |
| `setup-monitoring.sh` | VPS | Prometheus + Grafana + Node Exporter |
| `deploy.sh` | Local | Build + upload + restart + health check (manual fallback) |
| `deploy-native.sh` | VPS | Promote CI-built artifact + health check + auto-rollback |
| `start.sh` / `stop.sh` / `restart.sh` | VPS | Service control |
| `rollback.sh` | VPS | Restore previous binaries |
| `health.sh` | VPS | 6-check health report (JSON) |
| `backup.sh` | VPS | pg_dump + retention + S3 |
| `monitor.sh` | VPS | Full monitoring + Telegram (cron) |
| `check-perf.sh` | Any | Performance + security header audit |
