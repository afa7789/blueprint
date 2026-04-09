#!/usr/bin/env bash
set -euo pipefail

# Lightweight prerequisites installer.
# Assumes VPS already has basic Ubuntu 22.04+ with apt and snap available.
# Does NOT create databases, configure services, or write secrets.
# Use setup-vps.sh for a full first-time provisioning.

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
die()   { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

[[ $EUID -eq 0 ]] || die "This script must be run as root"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Blueprint Prerequisites Installer ==="
echo ""

# ---------------------------------------------------------------------------
# System packages
# ---------------------------------------------------------------------------
info "Updating apt cache..."
apt-get update -y

info "Installing base packages..."
apt-get install -y curl wget git build-essential unzip jq software-properties-common \
  apt-transport-https ca-certificates gnupg lsb-release

# ---------------------------------------------------------------------------
# Go
# ---------------------------------------------------------------------------
info "Installing Go..."
GO_VERSION=$(curl -fsSL "https://go.dev/dl/?mode=json" | jq -r '.[0].version')
GO_ARCH="linux-amd64"
GO_TAR="${GO_VERSION}.${GO_ARCH}.tar.gz"

if [[ ! -d /usr/local/go ]] || \
   [[ "$(/usr/local/go/bin/go version 2>/dev/null | awk '{print $3}')" != "${GO_VERSION}" ]]; then
  wget -q "https://go.dev/dl/${GO_TAR}" -O "/tmp/${GO_TAR}"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "/tmp/${GO_TAR}"
  rm "/tmp/${GO_TAR}"
  info "Go ${GO_VERSION} installed"
else
  info "Go ${GO_VERSION} already up to date"
fi

if [[ ! -f /etc/profile.d/go.sh ]]; then
  echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
fi
export PATH=$PATH:/usr/local/go/bin

# ---------------------------------------------------------------------------
# Bun
# ---------------------------------------------------------------------------
info "Installing Bun..."
if ! command -v bun &>/dev/null; then
  curl -fsSL https://bun.sh/install | bash
  ln -sf "$HOME/.bun/bin/bun" /usr/local/bin/bun 2>/dev/null || true
  info "Bun installed"
else
  info "Bun already installed"
fi

# ---------------------------------------------------------------------------
# PostgreSQL client tools only (no server)
# ---------------------------------------------------------------------------
info "Installing PostgreSQL client..."
if ! command -v psql &>/dev/null; then
  install -d /usr/share/postgresql-common/pgdg
  curl -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc --fail \
    https://www.postgresql.org/media/keys/ACCC4CF8.asc
  echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] \
    https://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" \
    > /etc/apt/sources.list.d/pgdg.list
  apt-get update -y
  apt-get install -y postgresql-client-16
  info "PostgreSQL client installed"
else
  info "psql already available"
fi

# ---------------------------------------------------------------------------
# Redis tools (client only)
# ---------------------------------------------------------------------------
info "Installing Redis tools..."
if ! command -v redis-cli &>/dev/null; then
  apt-get install -y redis-tools
  info "Redis tools installed"
else
  info "redis-cli already available"
fi

# ---------------------------------------------------------------------------
# Nginx with Brotli modules
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
  info "Nginx already installed"
fi

# ---------------------------------------------------------------------------
# Application directories
# ---------------------------------------------------------------------------
info "Creating system user 'blueprint'..."
if ! id blueprint &>/dev/null; then
  useradd --system --no-create-home --shell /sbin/nologin blueprint
  info "User 'blueprint' created"
else
  info "User 'blueprint' already exists"
fi

info "Creating application directories..."
for dir in api health frontend backups uploads logs scripts; do
  mkdir -p "/opt/blueprint/${dir}"
done
chown -R blueprint:blueprint /opt/blueprint
info "Directories ready under /opt/blueprint/"

# ---------------------------------------------------------------------------
# Systemd units
# ---------------------------------------------------------------------------
info "Installing Blueprint systemd units..."
for svc in blueprint-api blueprint-health; do
  SRC="${SCRIPT_DIR}/systemd/${svc}.service"
  DST="/etc/systemd/system/${svc}.service"
  if [[ -f "$SRC" ]]; then
    cp "$SRC" "$DST"
    info "Installed ${svc}.service"
  else
    warn "Unit file not found: ${SRC} — skipping"
  fi
done
systemctl daemon-reload

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo -e "${GREEN}=== Prerequisites installed ===${NC}"
echo ""
echo "  Go:               $(go version 2>/dev/null || echo 'check PATH')"
echo "  Bun:              $(bun --version 2>/dev/null || echo 'check PATH')"
echo "  Nginx:            $(nginx -v 2>&1 | head -1)"
echo ""
echo -e "${YELLOW}  Not done yet:${NC}"
echo "    - No database created (run setup-vps.sh or configure manually)"
echo "    - No Redis server installed (run setup-vps.sh or apt install redis-server)"
echo "    - No .env.production written"
echo "    - Blueprint services are installed but NOT enabled/started"
echo ""
echo "  Deploy binaries, then:"
echo "    systemctl enable --now blueprint-api blueprint-health"
echo ""
