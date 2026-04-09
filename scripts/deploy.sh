#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# deploy.sh — Zero-downtime deploy for Blueprint
# Usage: ./scripts/deploy.sh [--backend] [--frontend] [--all] [--dry-run]
# Reads config from scripts/.deploy.env
# ---------------------------------------------------------------------------

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEPLOY_ENV="$SCRIPT_DIR/.deploy.env"

info()    { echo -e "${CYAN}[deploy]${NC} $*"; }
success() { echo -e "${GREEN}[deploy]${NC} $*"; }
warn()    { echo -e "${YELLOW}[deploy]${NC} $*"; }
error()   { echo -e "${RED}[deploy]${NC} $*" >&2; }
die()     { error "$*"; exit 1; }

# ---------------------------------------------------------------------------
# Parse flags
# ---------------------------------------------------------------------------
DEPLOY_BACKEND=false
DEPLOY_FRONTEND=false
DRY_RUN=false

for arg in "$@"; do
    case "$arg" in
        --backend)  DEPLOY_BACKEND=true ;;
        --frontend) DEPLOY_FRONTEND=true ;;
        --all)      DEPLOY_BACKEND=true; DEPLOY_FRONTEND=true ;;
        --dry-run)  DRY_RUN=true ;;
        *) die "Unknown flag: $arg. Use --backend, --frontend, --all, --dry-run" ;;
    esac
done

if ! $DEPLOY_BACKEND && ! $DEPLOY_FRONTEND; then
    die "Specify at least one of: --backend, --frontend, --all"
fi

# ---------------------------------------------------------------------------
# Load deploy config
# ---------------------------------------------------------------------------
if [ ! -f "$DEPLOY_ENV" ]; then
    die "Missing $DEPLOY_ENV — copy .deploy.env.example and fill in values"
fi

# shellcheck source=/dev/null
source "$DEPLOY_ENV"

VPS_HOST="${VPS_HOST:?VPS_HOST must be set in .deploy.env}"
VPS_USER="${VPS_USER:?VPS_USER must be set in .deploy.env}"
VPS_PATH="${VPS_PATH:-/opt/blueprint}"
DOMAIN="${DOMAIN:-}"

SSH_TARGET="$VPS_USER@$VPS_HOST"
BUILD_DIR="$PROJECT_ROOT/build"

run() {
    if $DRY_RUN; then
        echo -e "${YELLOW}[dry-run]${NC} $*"
    else
        eval "$@"
    fi
}

ssh_run() {
    if $DRY_RUN; then
        echo -e "${YELLOW}[dry-run ssh]${NC} $*"
    else
        ssh "$SSH_TARGET" "$@"
    fi
}

# ---------------------------------------------------------------------------
# Backend deploy
# ---------------------------------------------------------------------------
deploy_backend() {
    info "Starting backend deploy..."

    info "Building blueprint-api..."
    run "cd \"$PROJECT_ROOT/backend\" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=\"-s -w\" -o \"$BUILD_DIR/blueprint-api\" ./cmd/server"

    info "Building blueprint-health..."
    run "cd \"$PROJECT_ROOT/backend\" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=\"-s -w\" -o \"$BUILD_DIR/blueprint-health\" ./cmd/health"

    info "Uploading binaries to VPS..."
    run "ssh \"$SSH_TARGET\" 'mkdir -p /tmp/blueprint-deploy'"
    run "rsync -az --progress \"$BUILD_DIR/blueprint-api\" \"$BUILD_DIR/blueprint-health\" \"$SSH_TARGET:/tmp/blueprint-deploy/\""

    info "Installing binaries on VPS..."
    ssh_run "cp $VPS_PATH/api/blueprint-api $VPS_PATH/api/blueprint-api.bak 2>/dev/null || true"
    ssh_run "cp $VPS_PATH/health/blueprint-health $VPS_PATH/health/blueprint-health.bak 2>/dev/null || true"
    ssh_run "cp /tmp/blueprint-deploy/blueprint-api $VPS_PATH/api/blueprint-api"
    ssh_run "cp /tmp/blueprint-deploy/blueprint-health $VPS_PATH/health/blueprint-health"
    ssh_run "chmod +x $VPS_PATH/api/blueprint-api $VPS_PATH/health/blueprint-health"

    info "Restarting services..."
    ssh_run "systemctl restart blueprint-api blueprint-health"

    info "Running health checks (10 retries, 2s apart)..."
    if ! $DRY_RUN; then
        RETRIES=10
        HEALTHY=false
        for i in $(seq 1 $RETRIES); do
            sleep 2
            if ssh "$SSH_TARGET" "curl -sf http://localhost:8080/healthz" > /dev/null 2>&1; then
                HEALTHY=true
                break
            fi
            warn "Health check attempt $i/$RETRIES failed, retrying..."
        done

        if ! $HEALTHY; then
            error "Health checks failed after $RETRIES attempts — rolling back"
            ssh "$SSH_TARGET" "
                cp $VPS_PATH/api/blueprint-api.bak $VPS_PATH/api/blueprint-api 2>/dev/null || true
                cp $VPS_PATH/health/blueprint-health.bak $VPS_PATH/health/blueprint-health 2>/dev/null || true
                systemctl restart blueprint-api blueprint-health
            "
            die "Rollback complete. Deploy failed."
        fi
    else
        echo -e "${YELLOW}[dry-run]${NC} health check loop (10 retries, 2s apart): curl -sf http://localhost:8080/healthz"
    fi

    info "Cleaning up..."
    ssh_run "rm -rf /tmp/blueprint-deploy"

    success "Backend deploy complete."
}

# ---------------------------------------------------------------------------
# Frontend deploy
# ---------------------------------------------------------------------------
deploy_frontend() {
    info "Starting frontend deploy..."

    info "Building frontend..."
    run "cd \"$PROJECT_ROOT/frontend\" && bun run build"

    info "Uploading frontend assets to VPS..."
    run "rsync -az --delete --progress \"$PROJECT_ROOT/frontend/dist/\" \"$SSH_TARGET:$VPS_PATH/frontend/\""

    info "Reloading nginx..."
    ssh_run "systemctl reload nginx"

    success "Frontend deploy complete."
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
if $DRY_RUN; then
    warn "DRY RUN mode — no changes will be made"
fi

mkdir -p "$BUILD_DIR"

$DEPLOY_BACKEND  && deploy_backend
$DEPLOY_FRONTEND && deploy_frontend

success "Deploy finished."
