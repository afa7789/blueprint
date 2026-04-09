#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# health.sh — Run health checks and output JSON
# Usage: ./scripts/health.sh [--telegram]
# Run ON the VPS.
# ---------------------------------------------------------------------------

TELEGRAM_FLAG=false
for arg in "$@"; do
    case "$arg" in
        --telegram) TELEGRAM_FLAG=true ;;
        *) echo "Unknown flag: $arg" >&2; exit 1 ;;
    esac
done

# ---------------------------------------------------------------------------
# Individual checks
# ---------------------------------------------------------------------------
check_api() {
    curl -sf http://localhost:8080/healthz > /dev/null 2>&1
}

check_postgres() {
    pg_isready -q
}

check_redis() {
    redis-cli ping 2>/dev/null | grep -q PONG
}

check_disk() {
    [ "$(df / --output=pcent | tail -1 | tr -d ' %')" -lt 80 ]
}

check_backup() {
    LATEST=$(ls -t /opt/blueprint/backups/*.dump.gz 2>/dev/null | head -1)
    [ -n "$LATEST" ] && [ "$(( $(date +%s) - $(stat -c %Y "$LATEST") ))" -lt 90000 ]
}

check_health_monitor() {
    curl -sf "http://localhost:8081/health?format=json" > /dev/null 2>&1
}

# ---------------------------------------------------------------------------
# Run checks and build JSON
# ---------------------------------------------------------------------------
CHECKS_JSON=""
OVERALL="healthy"
FAILED_NAMES=""

run_check() {
    local name="$1"
    local fn="$2"
    local status="up"

    if ! $fn 2>/dev/null; then
        status="down"
        OVERALL="unhealthy"
        FAILED_NAMES="$FAILED_NAMES $name"
    fi

    if [ -n "$CHECKS_JSON" ]; then
        CHECKS_JSON="$CHECKS_JSON,"
    fi
    CHECKS_JSON="${CHECKS_JSON}{\"name\":\"$name\",\"status\":\"$status\"}"
}

run_check "api"            check_api
run_check "postgres"       check_postgres
run_check "redis"          check_redis
run_check "disk"           check_disk
run_check "backup"         check_backup
run_check "health_monitor" check_health_monitor

OUTPUT="{\"checks\":[$CHECKS_JSON],\"overall\":\"$OVERALL\"}"
echo "$OUTPUT"

# ---------------------------------------------------------------------------
# Telegram alert
# ---------------------------------------------------------------------------
if $TELEGRAM_FLAG && [ "$OVERALL" = "unhealthy" ]; then
    BOT_TOKEN="${TELEGRAM_BOT_TOKEN:-}"
    CHAT_ID="${TELEGRAM_CHAT_ID:-}"

    if [ -z "$BOT_TOKEN" ] || [ -z "$CHAT_ID" ]; then
        echo "[health] WARNING: --telegram flag set but TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID not in env" >&2
    else
        HOSTNAME_VAL="$(hostname)"
        MESSAGE="Blueprint health alert on $HOSTNAME_VAL%0AFailed checks:$FAILED_NAMES"
        curl -s "https://api.telegram.org/bot$BOT_TOKEN/sendMessage" \
            -d "chat_id=$CHAT_ID" \
            -d "text=$MESSAGE" > /dev/null
    fi
fi

# Exit non-zero if unhealthy so cron/systemd can detect failures
[ "$OVERALL" = "healthy" ]
