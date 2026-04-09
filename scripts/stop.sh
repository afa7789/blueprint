#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# stop.sh — Stop Blueprint services
# Usage: ./scripts/stop.sh [all|api|health]
# ---------------------------------------------------------------------------

YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info() { echo -e "${CYAN}[stop]${NC} $*"; }
warn() { echo -e "${YELLOW}[stop]${NC} $*"; }

SERVICE="${1:-all}"

if [ "$SERVICE" = "all" ]; then
    info "Stopping blueprint-api and blueprint-health..."
    sudo systemctl stop blueprint-api blueprint-health
else
    info "Stopping blueprint-$SERVICE..."
    sudo systemctl stop "blueprint-$SERVICE"
fi

warn "Services stopped."
sudo systemctl status blueprint-api blueprint-health --no-pager || true
