#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# rollback.sh — Restore previous Blueprint binaries and restart services
# Run ON the VPS: /opt/blueprint/scripts/rollback.sh
# ---------------------------------------------------------------------------

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

VPS_PATH="${VPS_PATH:-/opt/blueprint}"

info()    { echo -e "${CYAN}[rollback]${NC} $*"; }
success() { echo -e "${GREEN}[rollback]${NC} $*"; }
warn()    { echo -e "${YELLOW}[rollback]${NC} $*"; }
error()   { echo -e "${RED}[rollback]${NC} $*" >&2; }
die()     { error "$*"; exit 1; }

# ---------------------------------------------------------------------------
# Verify backups exist
# ---------------------------------------------------------------------------
API_BAK="$VPS_PATH/api/blueprint-api.bak"
HEALTH_BAK="$VPS_PATH/health/blueprint-health.bak"

[ -f "$API_BAK" ]    || die "No API backup found at $API_BAK"
[ -f "$HEALTH_BAK" ] || die "No health backup found at $HEALTH_BAK"

# ---------------------------------------------------------------------------
# Restore binaries
# ---------------------------------------------------------------------------
info "Restoring API binary from backup..."
cp "$API_BAK" "$VPS_PATH/api/blueprint-api"
chmod +x "$VPS_PATH/api/blueprint-api"

info "Restoring health binary from backup..."
cp "$HEALTH_BAK" "$VPS_PATH/health/blueprint-health"
chmod +x "$VPS_PATH/health/blueprint-health"

# ---------------------------------------------------------------------------
# Restart services
# ---------------------------------------------------------------------------
info "Restarting services..."
systemctl restart blueprint-api blueprint-health

# ---------------------------------------------------------------------------
# Health check
# ---------------------------------------------------------------------------
info "Running health checks (10 retries, 2s apart)..."
RETRIES=10
HEALTHY=false
for i in $(seq 1 $RETRIES); do
    sleep 2
    if curl -sf http://localhost:8080/healthz > /dev/null 2>&1; then
        HEALTHY=true
        break
    fi
    warn "Health check attempt $i/$RETRIES failed, retrying..."
done

if ! $HEALTHY; then
    die "Health checks failed after rollback — manual intervention required"
fi

success "Rollback complete. Services are healthy."
systemctl status blueprint-api blueprint-health --no-pager
