#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# check-perf.sh — Performance and security header checks
# Usage: ./scripts/check-perf.sh <domain>
# ---------------------------------------------------------------------------

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

DOMAIN="${1:-}"
[ -n "$DOMAIN" ] || { echo "Usage: $0 <domain>" >&2; exit 1; }

pass() { echo -e "  ${GREEN}PASS${NC}  $*"; }
fail() { echo -e "  ${RED}FAIL${NC}  $*"; }
info() { echo -e "${CYAN}[check-perf]${NC} $*"; }

FAILURES=0

result() {
    local label="$1"
    local ok="$2"
    local detail="${3:-}"
    if [ "$ok" = "true" ]; then
        pass "$label${detail:+  ($detail)}"
    else
        fail "$label${detail:+  ($detail)}"
        (( FAILURES++ )) || true
    fi
}

echo ""
echo -e "${BOLD}Blueprint performance & security check — https://$DOMAIN${NC}"
echo "------------------------------------------------------------------------"

# ---------------------------------------------------------------------------
# Brotli compression
# ---------------------------------------------------------------------------
info "Checking Brotli compression..."
BR_ENC="$(curl -H "Accept-Encoding: br" -sI "https://$DOMAIN" 2>/dev/null | grep -i "^content-encoding" | tr -d '\r' || true)"
if echo "$BR_ENC" | grep -qi "br"; then
    result "Brotli encoding" "true" "$BR_ENC"
else
    result "Brotli encoding" "false" "content-encoding: $BR_ENC (expected br)"
fi

# ---------------------------------------------------------------------------
# gzip compression
# ---------------------------------------------------------------------------
info "Checking gzip compression..."
GZ_ENC="$(curl -H "Accept-Encoding: gzip" -sI "https://$DOMAIN" 2>/dev/null | grep -i "^content-encoding" | tr -d '\r' || true)"
if echo "$GZ_ENC" | grep -qi "gzip"; then
    result "gzip encoding" "true" "$GZ_ENC"
else
    result "gzip encoding" "false" "content-encoding: $GZ_ENC (expected gzip)"
fi

# ---------------------------------------------------------------------------
# HTTP/2
# ---------------------------------------------------------------------------
info "Checking HTTP/2..."
HTTP_VER="$(curl -sI --http2 "https://$DOMAIN" 2>/dev/null | head -1 | tr -d '\r' || true)"
if echo "$HTTP_VER" | grep -q "HTTP/2"; then
    result "HTTP/2" "true" "$HTTP_VER"
else
    result "HTTP/2" "false" "Got: $HTTP_VER"
fi

# ---------------------------------------------------------------------------
# Cache headers on a static asset (try /assets/ path, fallback to /)
# ---------------------------------------------------------------------------
info "Checking Cache-Control headers..."
# Try to find a static asset URL
ASSET_URL="https://$DOMAIN"
CACHE_CTRL="$(curl -sI "$ASSET_URL" 2>/dev/null | grep -i "^cache-control" | tr -d '\r' || true)"
if [ -n "$CACHE_CTRL" ]; then
    result "Cache-Control present" "true" "$CACHE_CTRL"
else
    result "Cache-Control present" "false" "No Cache-Control header on $ASSET_URL"
fi

# ---------------------------------------------------------------------------
# Security headers
# ---------------------------------------------------------------------------
info "Checking security headers..."
HEADERS="$(curl -sI "https://$DOMAIN" 2>/dev/null | tr -d '\r' || true)"

check_header() {
    local label="$1"
    local header="$2"
    local found
    found="$(echo "$HEADERS" | grep -i "^$header" | head -1 || true)"
    if [ -n "$found" ]; then
        result "$label" "true" "$found"
    else
        result "$label" "false" "Missing $header"
    fi
}

check_header "HSTS"                   "strict-transport-security"
check_header "X-Frame-Options"        "x-frame-options"
check_header "X-Content-Type-Options" "x-content-type-options"
check_header "Referrer-Policy"        "referrer-policy"
check_header "Content-Security-Policy" "content-security-policy"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo "------------------------------------------------------------------------"
if [ "$FAILURES" -eq 0 ]; then
    echo -e "${GREEN}All checks passed.${NC}"
else
    echo -e "${RED}$FAILURES check(s) failed.${NC}"
    exit 1
fi
