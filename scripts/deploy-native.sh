#!/usr/bin/env bash
set -euo pipefail

# Promote an artifact already built by CI. The production host does not need
# Git, the repository, or a Go toolchain.
APP_DIR=/opt/blueprint
RELEASE_ID="${1:?usage: deploy-native.sh <release-id>}"
INCOMING_DIR="$APP_DIR/releases/incoming/$RELEASE_ID"
RELEASE_DIR="$APP_DIR/releases/$RELEASE_ID"
STAGE_DIR="$APP_DIR/releases/staging-$RELEASE_ID"
SERVICES=(blueprint-api blueprint-health)
BINARIES=(blueprint-api blueprint-health)

cleanup() { rm -rf "$STAGE_DIR"; }
trap cleanup EXIT

[[ -d "$INCOMING_DIR/bin" ]] || { echo "Missing release artifacts: $INCOMING_DIR/bin" >&2; exit 1; }
for binary in "${BINARIES[@]}"; do
  [[ -f "$INCOMING_DIR/bin/$binary" ]] || { echo "Missing binary: $binary" >&2; exit 1; }
done

mkdir -p "$STAGE_DIR" "$RELEASE_DIR" "$APP_DIR/api" "$APP_DIR/health"
cp -a "$INCOMING_DIR/." "$STAGE_DIR/"

# Keep the previous binaries in the release directory for rollback.
for binary in "${BINARIES[@]}"; do
  case "$binary" in
    blueprint-api)    target_dir="$APP_DIR/api" ;;
    blueprint-health) target_dir="$APP_DIR/health" ;;
  esac
  [[ -f "$target_dir/$binary" ]] && cp "$target_dir/$binary" "$RELEASE_DIR/$binary.previous"
  install -o blueprint -g blueprint -m 0755 \
    "$STAGE_DIR/bin/$binary" "$target_dir/$binary.new"
  mv -f "$target_dir/$binary.new" "$target_dir/$binary"
done

if [[ -d "$STAGE_DIR/frontend" ]]; then
  rsync -az --delete "$STAGE_DIR/frontend/" "$APP_DIR/frontend/"
fi
chown -R blueprint:blueprint "$APP_DIR/frontend"

# Grafana dashboards and provisioning are declarative release artifacts.
# Keep this conditional so application deploys remain safe on hosts where
# monitoring has not been installed yet.
if [[ -d "$STAGE_DIR/grafana" && -d /etc/grafana ]]; then
  install -d /etc/grafana/provisioning /var/lib/grafana/dashboards
  rsync -az --delete "$STAGE_DIR/grafana/provisioning/" /etc/grafana/provisioning/
  rsync -az --delete "$STAGE_DIR/grafana/dashboards/" /var/lib/grafana/dashboards/
  if id grafana &>/dev/null; then
    chown -R grafana:grafana /etc/grafana/provisioning /var/lib/grafana/dashboards
  fi
fi

systemctl daemon-reload
systemctl restart "${SERVICES[@]}"
if systemctl list-unit-files grafana-server.service &>/dev/null; then
  systemctl restart grafana-server
fi
systemctl reload nginx

healthy=false
for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:8080/healthz >/dev/null; then
    healthy=true
    break
  fi
  sleep 2
done

if [[ "$healthy" != true ]]; then
  for binary in "${BINARIES[@]}"; do
    case "$binary" in
      blueprint-api)    target_dir="$APP_DIR/api" ;;
      blueprint-health) target_dir="$APP_DIR/health" ;;
    esac
    [[ -f "$RELEASE_DIR/$binary.previous" ]] && \
      cp "$RELEASE_DIR/$binary.previous" "$target_dir/$binary"
  done
  systemctl restart "${SERVICES[@]}"
  echo "Deploy failed; previous binaries restored" >&2
  exit 1
fi

rm -rf "$INCOMING_DIR"
find "$APP_DIR/releases" -mindepth 1 -maxdepth 1 -type d -mtime +14 -exec rm -rf {} +
echo "Blueprint artifact deploy healthy: $RELEASE_ID"
