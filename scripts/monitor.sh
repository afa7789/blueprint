#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# monitor.sh — Production health monitor, designed for */5 * * * * cron
# Run ON the VPS.
# ---------------------------------------------------------------------------

# ---------------------------------------------------------------------------
# Checks
# ---------------------------------------------------------------------------
check_api() {
    if curl -sf http://localhost:8080/healthz > /dev/null 2>&1; then
        echo '{"name":"api","status":"up","detail":"HTTP 200 on /healthz"}'
    else
        echo '{"name":"api","status":"down","detail":"Failed to reach http://localhost:8080/healthz"}'
        return 1
    fi
}

check_postgres() {
    local out
    out="$(pg_isready 2>&1 || true)"
    if echo "$out" | grep -q "accepting connections"; then
        echo "{\"name\":\"postgres\",\"status\":\"up\",\"detail\":\"$(echo "$out" | tr -d '\n')\"}"
    else
        echo "{\"name\":\"postgres\",\"status\":\"down\",\"detail\":\"$(echo "$out" | tr -d '\n')\"}"
        return 1
    fi
}

check_redis() {
    if redis-cli ping 2>/dev/null | grep -q PONG; then
        echo '{"name":"redis","status":"up","detail":"PONG received"}'
    else
        echo '{"name":"redis","status":"down","detail":"No PONG from redis-cli"}'
        return 1
    fi
}

check_disk() {
    local pct
    pct="$(df / --output=pcent | tail -1 | tr -d ' %')"
    if [ "$pct" -lt 80 ]; then
        echo "{\"name\":\"disk\",\"status\":\"up\",\"detail\":\"${pct}% used\"}"
    else
        echo "{\"name\":\"disk\",\"status\":\"down\",\"detail\":\"${pct}% used — above 80% threshold\"}"
        return 1
    fi
}

check_memory() {
    local total avail pct
    total="$(grep MemTotal /proc/meminfo | awk '{print $2}')"
    avail="$(grep MemAvailable /proc/meminfo | awk '{print $2}')"
    pct=$(( (total - avail) * 100 / total ))
    if [ "$pct" -lt 90 ]; then
        echo "{\"name\":\"memory\",\"status\":\"up\",\"detail\":\"${pct}% used\"}"
    else
        echo "{\"name\":\"memory\",\"status\":\"down\",\"detail\":\"${pct}% used — above 90% threshold\"}"
        return 1
    fi
}

check_ssl() {
    local domain="${DOMAIN:-}"
    if [ -z "$domain" ]; then
        echo '{"name":"ssl","status":"unknown","detail":"DOMAIN not set"}'
        return 0
    fi

    local expiry_date expiry_epoch now_epoch days_left
    expiry_date="$(echo | openssl s_client -servername "$domain" -connect "$domain:443" 2>/dev/null \
        | openssl x509 -noout -enddate 2>/dev/null \
        | cut -d= -f2 || true)"

    if [ -z "$expiry_date" ]; then
        echo "{\"name\":\"ssl\",\"status\":\"down\",\"detail\":\"Could not retrieve cert for $domain\"}"
        return 1
    fi

    expiry_epoch="$(date -d "$expiry_date" +%s 2>/dev/null || date -j -f "%b %d %T %Y %Z" "$expiry_date" +%s 2>/dev/null || echo 0)"
    now_epoch="$(date +%s)"
    days_left=$(( (expiry_epoch - now_epoch) / 86400 ))

    if [ "$days_left" -gt 14 ]; then
        echo "{\"name\":\"ssl\",\"status\":\"up\",\"detail\":\"Expires in ${days_left} days ($expiry_date)\"}"
    else
        echo "{\"name\":\"ssl\",\"status\":\"down\",\"detail\":\"Expires in ${days_left} days — renew soon ($expiry_date)\"}"
        return 1
    fi
}

# ---------------------------------------------------------------------------
# Run all checks
# ---------------------------------------------------------------------------
CHECKS_JSON=""
OVERALL="healthy"
FAILED_NAMES=""

run_check() {
    local name="$1"
    local fn="$2"
    local result

    result="$($fn 2>/dev/null || true)"
    local status
    status="$(echo "$result" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('status','unknown'))" 2>/dev/null || echo "unknown")"

    if [ "$status" = "down" ]; then
        OVERALL="unhealthy"
        FAILED_NAMES="$FAILED_NAMES $name"
    fi

    if [ -n "$CHECKS_JSON" ]; then
        CHECKS_JSON="$CHECKS_JSON,"
    fi
    CHECKS_JSON="${CHECKS_JSON}${result}"
}

run_check "api"      check_api
run_check "postgres" check_postgres
run_check "redis"    check_redis
run_check "disk"     check_disk
run_check "memory"   check_memory
run_check "ssl"      check_ssl

TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "{\"timestamp\":\"$TIMESTAMP\",\"checks\":[$CHECKS_JSON],\"overall\":\"$OVERALL\"}"

# ---------------------------------------------------------------------------
# Telegram alert on failure
# ---------------------------------------------------------------------------
if [ "$OVERALL" = "unhealthy" ]; then
    BOT_TOKEN="${TELEGRAM_BOT_TOKEN:-}"
    CHAT_ID="${TELEGRAM_CHAT_ID:-}"

    if [ -n "$BOT_TOKEN" ] && [ -n "$CHAT_ID" ]; then
        HOSTNAME_VAL="$(hostname)"
        MESSAGE="[blueprint] Monitor alert on $HOSTNAME_VAL at $TIMESTAMP%0AFailed:$FAILED_NAMES"
        curl -s "https://api.telegram.org/bot$BOT_TOKEN/sendMessage" \
            -d "chat_id=$CHAT_ID" \
            -d "text=$MESSAGE" > /dev/null
    fi
fi

[ "$OVERALL" = "healthy" ]
