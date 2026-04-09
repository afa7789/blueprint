#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# start.sh — Start Blueprint services
# Usage: ./scripts/start.sh [all|api|health]
# ---------------------------------------------------------------------------

GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

info()    { echo -e "${CYAN}[start]${NC} $*"; }
success() { echo -e "${GREEN}[start]${NC} $*"; }

SERVICE="${1:-all}"

if [ "$SERVICE" = "all" ]; then
    info "Starting blueprint-api and blueprint-health..."
    sudo systemctl start blueprint-api blueprint-health
else
    info "Starting blueprint-$SERVICE..."
    sudo systemctl start "blueprint-$SERVICE"
fi

success "Services started."
sudo systemctl status blueprint-api blueprint-health --no-pager
