#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# restart.sh — Restart Blueprint services
# Usage: ./scripts/restart.sh [all|api|health]
# ---------------------------------------------------------------------------

GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

info()    { echo -e "${CYAN}[restart]${NC} $*"; }
success() { echo -e "${GREEN}[restart]${NC} $*"; }

SERVICE="${1:-all}"

if [ "$SERVICE" = "all" ]; then
    info "Restarting blueprint-api and blueprint-health..."
    sudo systemctl restart blueprint-api blueprint-health
else
    info "Restarting blueprint-$SERVICE..."
    sudo systemctl restart "blueprint-$SERVICE"
fi

success "Services restarted."
sudo systemctl status blueprint-api blueprint-health --no-pager
