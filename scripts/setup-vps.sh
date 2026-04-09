#!/usr/bin/env bash
set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info()    { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }
die()     { error "$*"; exit 1; }

# Arguments
DB_PASSWORD="${1:-$(openssl rand -base64 24)}"
REDIS_PASSWORD="${2:-$(openssl rand -base64 24)}"

echo "=== Blueprint VPS Setup ==="
echo "DB Password: $DB_PASSWORD"
echo "Redis Password: $REDIS_PASSWORD"
echo ""

[[ $EUID -eq 0 ]] || die "This script must be run as root"

# ---------------------------------------------------------------------------
# 1. System update
# ---------------------------------------------------------------------------
info "Updating system packages..."
apt-get update -y
apt-get upgrade -y

# ---------------------------------------------------------------------------
# 2. Essential packages
# ---------------------------------------------------------------------------
info "Installing essential packages..."
apt-get install -y curl wget git build-essential unzip jq software-properties-common \
  apt-transport-https ca-certificates gnupg lsb-release

# ---------------------------------------------------------------------------
# 3. Go — latest stable from go.dev
# ---------------------------------------------------------------------------
info "Installing Go..."
GO_VERSION=$(curl -fsSL "https://go.dev/dl/?mode=json" | jq -r '.[0].version')
GO_ARCH="linux-amd64"
GO_TAR="${GO_VERSION}.${GO_ARCH}.tar.gz"

if [[ ! -d /usr/local/go ]] || [[ "$(/usr/local/go/bin/go version 2>/dev/null | awk '{print $3}')" != "${GO_VERSION}" ]]; then
  wget -q "https://go.dev/dl/${GO_TAR}" -O "/tmp/${GO_TAR}"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "/tmp/${GO_TAR}"
  rm "/tmp/${GO_TAR}"
  info "Go ${GO_VERSION} installed"
else
  info "Go ${GO_VERSION} already installed, skipping"
fi

# Add Go to PATH for all users
if [[ ! -f /etc/profile.d/go.sh ]]; then
  echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
fi
export PATH=$PATH:/usr/local/go/bin

# ---------------------------------------------------------------------------
# 4. Bun
# ---------------------------------------------------------------------------
info "Installing Bun..."
if ! command -v bun &>/dev/null; then
  curl -fsSL https://bun.sh/install | bash
  # Make bun available system-wide
  ln -sf "$HOME/.bun/bin/bun" /usr/local/bin/bun 2>/dev/null || true
  info "Bun installed"
else
  info "Bun already installed, skipping"
fi

# ---------------------------------------------------------------------------
# 5. PostgreSQL 16
# ---------------------------------------------------------------------------
info "Installing PostgreSQL 16..."
if ! command -v psql &>/dev/null; then
  install -d /usr/share/postgresql-common/pgdg
  curl -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc --fail \
    https://www.postgresql.org/media/keys/ACCC4CF8.asc
  echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] \
    https://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" \
    > /etc/apt/sources.list.d/pgdg.list
  apt-get update -y
  apt-get install -y postgresql-16 postgresql-client-16
  systemctl enable postgresql
  systemctl start postgresql
  info "PostgreSQL 16 installed"
else
  info "PostgreSQL already installed, skipping package install"
fi

info "Configuring PostgreSQL user and database..."
su -c "psql -tc \"SELECT 1 FROM pg_roles WHERE rolname='blueprint'\" | grep -q 1 || \
  psql -c \"CREATE USER blueprint WITH PASSWORD '${DB_PASSWORD}';\"" postgres

su -c "psql -tc \"SELECT 1 FROM pg_database WHERE datname='blueprint'\" | grep -q 1 || \
  psql -c \"CREATE DATABASE blueprint OWNER blueprint;\"" postgres

su -c "psql -c \"GRANT ALL PRIVILEGES ON DATABASE blueprint TO blueprint;\"" postgres

# ---------------------------------------------------------------------------
# 6. Redis 7
# ---------------------------------------------------------------------------
info "Installing Redis..."
if ! command -v redis-server &>/dev/null; then
  apt-get install -y redis-server
  systemctl enable redis-server
  info "Redis installed"
else
  info "Redis already installed, skipping package install"
fi

info "Configuring Redis password..."
REDIS_CONF="/etc/redis/redis.conf"
if grep -q "^requirepass " "$REDIS_CONF" 2>/dev/null; then
  sed -i "s|^requirepass .*|requirepass ${REDIS_PASSWORD}|" "$REDIS_CONF"
else
  echo "requirepass ${REDIS_PASSWORD}" >> "$REDIS_CONF"
fi
systemctl restart redis-server

# ---------------------------------------------------------------------------
# 7. Nginx with Brotli
# ---------------------------------------------------------------------------
info "Installing Nginx..."
if ! command -v nginx &>/dev/null; then
  add-apt-repository -y ppa:ondrej/nginx-mainline
  apt-get update -y
  apt-get install -y nginx libnginx-mod-http-brotli-filter libnginx-mod-http-brotli-static
  systemctl enable nginx
  systemctl start nginx
  info "Nginx installed"
else
  info "Nginx already installed, skipping"
fi

# ---------------------------------------------------------------------------
# 8. Certbot
# ---------------------------------------------------------------------------
info "Installing Certbot..."
if ! command -v certbot &>/dev/null; then
  snap install --classic certbot
  ln -sf /snap/bin/certbot /usr/bin/certbot
  info "Certbot installed"
else
  info "Certbot already installed, skipping"
fi

# ---------------------------------------------------------------------------
# 9. pgweb
# ---------------------------------------------------------------------------
info "Installing pgweb..."
if [[ ! -f /usr/local/bin/pgweb ]]; then
  PGWEB_VERSION=$(curl -fsSL https://api.github.com/repos/sosedoff/pgweb/releases/latest \
    | jq -r '.tag_name')
  PGWEB_URL="https://github.com/sosedoff/pgweb/releases/download/${PGWEB_VERSION}/pgweb_linux_amd64.zip"
  wget -q "$PGWEB_URL" -O /tmp/pgweb.zip
  unzip -o /tmp/pgweb.zip -d /tmp/pgweb_extract
  install -m 0755 /tmp/pgweb_extract/pgweb_linux_amd64 /usr/local/bin/pgweb
  rm -rf /tmp/pgweb.zip /tmp/pgweb_extract
  info "pgweb ${PGWEB_VERSION} installed"
else
  info "pgweb already installed, skipping"
fi

# ---------------------------------------------------------------------------
# 10. Prometheus
# ---------------------------------------------------------------------------
info "Installing Prometheus..."
PROM_VERSION=$(curl -fsSL https://api.github.com/repos/prometheus/prometheus/releases/latest \
  | jq -r '.tag_name' | tr -d 'v')

if [[ ! -f /usr/local/bin/prometheus ]]; then
  PROM_TAR="prometheus-${PROM_VERSION}.linux-amd64.tar.gz"
  wget -q "https://github.com/prometheus/prometheus/releases/download/v${PROM_VERSION}/${PROM_TAR}" \
    -O "/tmp/${PROM_TAR}"
  tar -xzf "/tmp/${PROM_TAR}" -C /tmp
  install -m 0755 "/tmp/prometheus-${PROM_VERSION}.linux-amd64/prometheus" /usr/local/bin/prometheus
  install -m 0755 "/tmp/prometheus-${PROM_VERSION}.linux-amd64/promtool" /usr/local/bin/promtool
  rm -rf "/tmp/${PROM_TAR}" "/tmp/prometheus-${PROM_VERSION}.linux-amd64"
  info "Prometheus ${PROM_VERSION} installed"
else
  info "Prometheus already installed, skipping binary"
fi

# Prometheus config
mkdir -p /etc/prometheus /var/lib/prometheus
if [[ ! -f /etc/prometheus/prometheus.yml ]]; then
  cat > /etc/prometheus/prometheus.yml <<'PROMYML'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'node'
    static_configs:
      - targets: ['localhost:9100']

  - job_name: 'blueprint-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /healthz
PROMYML
  info "Prometheus config written"
fi

# Prometheus systemd service
if [[ ! -f /etc/systemd/system/prometheus.service ]]; then
  useradd --no-create-home --shell /bin/false prometheus 2>/dev/null || true
  chown -R prometheus:prometheus /etc/prometheus /var/lib/prometheus

  cat > /etc/systemd/system/prometheus.service <<'PROMSVC'
[Unit]
Description=Prometheus Monitoring
After=network.target

[Service]
Type=simple
User=prometheus
Group=prometheus
ExecStart=/usr/local/bin/prometheus \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/var/lib/prometheus \
  --web.listen-address=0.0.0.0:9090
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
PROMSVC

  systemctl daemon-reload
  systemctl enable prometheus
  systemctl start prometheus
  info "Prometheus service created and started"
fi

# ---------------------------------------------------------------------------
# 11. Node Exporter
# ---------------------------------------------------------------------------
info "Installing Node Exporter..."
NODE_EXP_VERSION=$(curl -fsSL https://api.github.com/repos/prometheus/node_exporter/releases/latest \
  | jq -r '.tag_name' | tr -d 'v')

if [[ ! -f /usr/local/bin/node_exporter ]]; then
  NODE_TAR="node_exporter-${NODE_EXP_VERSION}.linux-amd64.tar.gz"
  wget -q "https://github.com/prometheus/node_exporter/releases/download/v${NODE_EXP_VERSION}/${NODE_TAR}" \
    -O "/tmp/${NODE_TAR}"
  tar -xzf "/tmp/${NODE_TAR}" -C /tmp
  install -m 0755 "/tmp/node_exporter-${NODE_EXP_VERSION}.linux-amd64/node_exporter" \
    /usr/local/bin/node_exporter
  rm -rf "/tmp/${NODE_TAR}" "/tmp/node_exporter-${NODE_EXP_VERSION}.linux-amd64"
  info "Node Exporter ${NODE_EXP_VERSION} installed"
else
  info "Node Exporter already installed, skipping binary"
fi

if [[ ! -f /etc/systemd/system/node_exporter.service ]]; then
  useradd --no-create-home --shell /bin/false node_exporter 2>/dev/null || true

  cat > /etc/systemd/system/node_exporter.service <<'NESVC'
[Unit]
Description=Prometheus Node Exporter
After=network.target

[Service]
Type=simple
User=node_exporter
Group=node_exporter
ExecStart=/usr/local/bin/node_exporter
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
NESVC

  systemctl daemon-reload
  systemctl enable node_exporter
  systemctl start node_exporter
  info "Node Exporter service created and started"
fi

# ---------------------------------------------------------------------------
# 12. Grafana
# ---------------------------------------------------------------------------
info "Installing Grafana..."
if ! command -v grafana-server &>/dev/null; then
  mkdir -p /etc/apt/keyrings
  wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor \
    > /etc/apt/keyrings/grafana.gpg
  echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" \
    > /etc/apt/sources.list.d/grafana.list
  apt-get update -y
  apt-get install -y grafana
  systemctl enable grafana-server
  systemctl start grafana-server
  info "Grafana installed"
else
  info "Grafana already installed, skipping"
fi

# ---------------------------------------------------------------------------
# 13. System user
# ---------------------------------------------------------------------------
info "Creating system user 'blueprint'..."
if ! id blueprint &>/dev/null; then
  useradd --system --no-create-home --shell /sbin/nologin blueprint
  info "User 'blueprint' created"
else
  info "User 'blueprint' already exists, skipping"
fi

# ---------------------------------------------------------------------------
# 14. Directories
# ---------------------------------------------------------------------------
info "Creating application directories..."
for dir in api health frontend backups uploads logs scripts; do
  mkdir -p "/opt/blueprint/${dir}"
done
chown -R blueprint:blueprint /opt/blueprint
info "Directories created under /opt/blueprint/"

# ---------------------------------------------------------------------------
# 15. Systemd units
# ---------------------------------------------------------------------------
info "Installing Blueprint systemd units..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

for svc in blueprint-api blueprint-health; do
  SRC="${SCRIPT_DIR}/systemd/${svc}.service"
  DST="/etc/systemd/system/${svc}.service"
  if [[ -f "$SRC" ]]; then
    cp "$SRC" "$DST"
    info "Installed ${svc}.service"
  else
    warn "Systemd unit not found at ${SRC}, skipping"
  fi
done

systemctl daemon-reload

# ---------------------------------------------------------------------------
# 16. UFW firewall
# ---------------------------------------------------------------------------
info "Configuring UFW firewall..."
if command -v ufw &>/dev/null; then
  ufw --force reset
  ufw default deny incoming
  ufw default allow outgoing
  ufw allow 22/tcp comment 'SSH'
  ufw allow 80/tcp comment 'HTTP'
  ufw allow 443/tcp comment 'HTTPS'

  # Restricted ports — prompt for admin IP or skip
  ADMIN_IP="${ADMIN_IP:-}"
  if [[ -n "$ADMIN_IP" ]]; then
    ufw allow from "$ADMIN_IP" to any port 8081 comment 'Blueprint Health Monitor'
    ufw allow from "$ADMIN_IP" to any port 3000 comment 'Grafana'
    ufw allow from "$ADMIN_IP" to any port 9090 comment 'Prometheus'
    info "Restricted ports open for $ADMIN_IP"
  else
    warn "ADMIN_IP not set — ports 8081, 3000, 9090 are NOT open."
    warn "Run: ufw allow from YOUR_IP to any port 8081"
    warn "     ufw allow from YOUR_IP to any port 3000"
    warn "     ufw allow from YOUR_IP to any port 9090"
  fi

  ufw --force enable
  info "UFW configured"
else
  warn "ufw not found, skipping firewall configuration"
fi

# ---------------------------------------------------------------------------
# 17. Environment file
# ---------------------------------------------------------------------------
info "Generating /opt/blueprint/.env.production..."
JWT_SECRET="$(openssl rand -base64 32)"

cat > /opt/blueprint/.env.production <<ENVFILE
DATABASE_URL=postgres://blueprint:${DB_PASSWORD}@localhost:5432/blueprint?sslmode=disable
DATABASE_MIGRATION_URL=postgres://blueprint:${DB_PASSWORD}@localhost:5432/blueprint?sslmode=disable
REDIS_URL=redis://:${REDIS_PASSWORD}@localhost:6379
JWT_SECRET=${JWT_SECRET}
PORT=8080
ENV=production
FRONTEND_URL=https://YOUR_DOMAIN
STORAGE_TYPE=local
UPLOAD_DIR=/opt/blueprint/uploads
PGWEB_URL=http://localhost:8081
ENVFILE

chmod 600 /opt/blueprint/.env.production
chown blueprint:blueprint /opt/blueprint/.env.production
info ".env.production written"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo -e "${GREEN}=== Setup Complete ===${NC}"
echo ""
echo "  PostgreSQL user:   blueprint"
echo "  PostgreSQL db:     blueprint"
echo "  DB password:       ${DB_PASSWORD}"
echo "  Redis password:    ${REDIS_PASSWORD}"
echo "  JWT secret:        ${JWT_SECRET}"
echo ""
echo "  Services:"
echo "    PostgreSQL       :5432"
echo "    Redis            :6379"
echo "    Nginx            :80 / :443"
echo "    Prometheus       :9090"
echo "    Node Exporter    :9100"
echo "    Grafana          :3000  (admin/admin — change on first login)"
echo ""
echo "  Blueprint app dirs: /opt/blueprint/"
echo "  Env file:           /opt/blueprint/.env.production"
echo ""
echo -e "${YELLOW}  Next steps:${NC}"
echo "    1. Edit /opt/blueprint/.env.production and replace YOUR_DOMAIN"
echo "    2. Deploy API binary to /opt/blueprint/api/blueprint-api"
echo "    3. Deploy health binary to /opt/blueprint/health/blueprint-health"
echo "    4. Deploy frontend to /opt/blueprint/frontend/"
echo "    5. Configure Nginx vhost and run: certbot --nginx -d your.domain"
echo "    6. systemctl start blueprint-api blueprint-health"
echo ""
