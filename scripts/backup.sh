#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# backup.sh — PostgreSQL backup with retention policy and optional S3 upload
# Run ON the VPS.
# ---------------------------------------------------------------------------

GREEN='\033[0;32m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

info()    { echo -e "${CYAN}[backup]${NC} $*"; }
success() { echo -e "${GREEN}[backup]${NC} $*"; }
die()     { echo -e "${RED}[backup]${NC} $*" >&2; exit 1; }

ENV_FILE="/opt/blueprint/.env.production"
BACKUP_DIR="/opt/blueprint/backups"

# ---------------------------------------------------------------------------
# Load env
# ---------------------------------------------------------------------------
[ -f "$ENV_FILE" ] || die "Env file not found: $ENV_FILE"
# shellcheck source=/dev/null
source "$ENV_FILE"

DATABASE_URL="${DATABASE_URL:?DATABASE_URL must be set in $ENV_FILE}"
AWS_BUCKET="${AWS_BUCKET:-}"

# ---------------------------------------------------------------------------
# Parse DATABASE_URL: postgres://user:password@host:port/dbname
# ---------------------------------------------------------------------------
# Strip scheme
_url="${DATABASE_URL#postgres://}"
_url="${_url#postgresql://}"

DB_USER="${_url%%:*}"
_url="${_url#*:}"
DB_PASS="${_url%%@*}"
_url="${_url#*@}"
DB_HOST="${_url%%:*}"
_url="${_url#*:}"
DB_PORT="${_url%%/*}"
DB_NAME="${_url#*/}"
# Strip query string if any
DB_NAME="${DB_NAME%%\?*}"

export PGPASSWORD="$DB_PASS"

# ---------------------------------------------------------------------------
# Create backup
# ---------------------------------------------------------------------------
mkdir -p "$BACKUP_DIR"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
BACKUP_FILE="$BACKUP_DIR/blueprint_${TIMESTAMP}.dump.gz"

info "Dumping database $DB_NAME to $BACKUP_FILE..."
pg_dump -Fc -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"
success "Dump complete: $BACKUP_FILE ($(du -sh "$BACKUP_FILE" | cut -f1))"

# ---------------------------------------------------------------------------
# Retention policy
# ---------------------------------------------------------------------------
info "Applying retention policy..."

# Collect all backups sorted newest first
mapfile -t ALL_BACKUPS < <(ls -t "$BACKUP_DIR"/blueprint_*.dump.gz 2>/dev/null)

KEEP=()

# Keep last 7 daily backups
DAILY_COUNT=0
for f in "${ALL_BACKUPS[@]}"; do
    if [ $DAILY_COUNT -lt 7 ]; then
        KEEP+=("$f")
        (( DAILY_COUNT++ )) || true
    fi
done

# Keep 1 per week for last 4 weeks (files between 7 and 35 days old)
NOW="$(date +%s)"
for WEEK in 1 2 3 4; do
    WEEK_START=$(( NOW - WEEK * 7 * 86400 ))
    WEEK_END=$(( NOW - (WEEK - 1) * 7 * 86400 ))
    FOUND=""
    for f in "${ALL_BACKUPS[@]}"; do
        MTIME="$(stat -c %Y "$f")"
        if [ "$MTIME" -ge "$WEEK_START" ] && [ "$MTIME" -lt "$WEEK_END" ]; then
            FOUND="$f"
            break
        fi
    done
    [ -n "$FOUND" ] && KEEP+=("$FOUND")
done

# Keep 1 per month for last 12 months
for MONTH in $(seq 1 12); do
    MONTH_START=$(( NOW - MONTH * 30 * 86400 ))
    MONTH_END=$(( NOW - (MONTH - 1) * 30 * 86400 ))
    FOUND=""
    for f in "${ALL_BACKUPS[@]}"; do
        MTIME="$(stat -c %Y "$f")"
        if [ "$MTIME" -ge "$MONTH_START" ] && [ "$MTIME" -lt "$MONTH_END" ]; then
            FOUND="$f"
            break
        fi
    done
    [ -n "$FOUND" ] && KEEP+=("$FOUND")
done

# Delete everything not in KEEP
for f in "${ALL_BACKUPS[@]}"; do
    SHOULD_KEEP=false
    for k in "${KEEP[@]}"; do
        if [ "$f" = "$k" ]; then
            SHOULD_KEEP=true
            break
        fi
    done
    if ! $SHOULD_KEEP; then
        info "Deleting old backup: $(basename "$f")"
        rm -f "$f"
    fi
done

# ---------------------------------------------------------------------------
# S3 upload
# ---------------------------------------------------------------------------
if [ -n "$AWS_BUCKET" ]; then
    info "Uploading to s3://$AWS_BUCKET/backups/..."
    aws s3 cp "$BACKUP_FILE" "s3://$AWS_BUCKET/backups/$(basename "$BACKUP_FILE")"
    success "Uploaded to S3."
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
TOTAL="$(ls "$BACKUP_DIR"/blueprint_*.dump.gz 2>/dev/null | wc -l)"
success "Backup complete."
echo "  File:           $BACKUP_FILE"
echo "  Size:           $(du -sh "$BACKUP_FILE" | cut -f1)"
echo "  Total backups:  $TOTAL"
