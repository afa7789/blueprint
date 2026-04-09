#!/usr/bin/env bash
set -euo pipefail

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { echo -e "${GREEN}[INFO]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }

VPS="${1:?Usage: $0 user@host [db_password] [redis_password]}"
DB_PASS="${2:-}"
REDIS_PASS="${3:-}"

# Resolve the repo root relative to this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Deploying setup to $VPS ==="

# Copy setup script and systemd units
info "Uploading scripts to $VPS..."
ssh "$VPS" "mkdir -p /tmp/blueprint-setup/systemd"
scp "${SCRIPT_DIR}/setup-vps.sh" "$VPS:/tmp/blueprint-setup/setup-vps.sh"
scp "${SCRIPT_DIR}/systemd/blueprint-api.service" \
    "${SCRIPT_DIR}/systemd/blueprint-health.service" \
    "$VPS:/tmp/blueprint-setup/systemd/"

info "Running setup on $VPS (this may take several minutes)..."
# shellcheck disable=SC2087
ssh -t "$VPS" "sudo bash /tmp/blueprint-setup/setup-vps.sh ${DB_PASS} ${REDIS_PASS}"

# Download generated .env.production
info "Downloading generated .env.production..."
scp "$VPS:/opt/blueprint/.env.production" "${SCRIPT_DIR}/.env.production.local"

echo ""
echo -e "${GREEN}Done.${NC} VPS env saved to ${SCRIPT_DIR}/.env.production.local"
warn "Keep .env.production.local secure — it contains secrets."
