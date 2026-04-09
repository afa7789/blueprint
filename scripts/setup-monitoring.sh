#!/usr/bin/env bash
set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# --- Guards ---

if [[ $EUID -ne 0 ]]; then
    error "This script must be run as root."
    exit 1
fi

# --- Prometheus config ---

PROMETHEUS_CONF="/etc/prometheus/prometheus.yml"

if [[ ! -d /etc/prometheus ]]; then
    info "Creating /etc/prometheus directory ..."
    mkdir -p /etc/prometheus
fi

info "Writing ${PROMETHEUS_CONF} ..."

cat > "${PROMETHEUS_CONF}" <<'PROMEOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'node'
    static_configs:
      - targets: ['localhost:9100']

  - job_name: 'blueprint-api'
    metrics_path: /healthz
    scheme: http
    static_configs:
      - targets: ['localhost:8080']

  - job_name: 'blueprint-health'
    metrics_path: /health
    params:
      format: ['json']
    static_configs:
      - targets: ['localhost:8081']
PROMEOF

success "Wrote ${PROMETHEUS_CONF}"

# --- Grafana ini ---

GRAFANA_INI="/etc/grafana/grafana.ini"

if [[ ! -f "${GRAFANA_INI}" ]]; then
    warn "${GRAFANA_INI} not found. Is Grafana installed?"
    warn "Install with: apt install -y grafana  (after adding the Grafana APT repo)"
    warn "Skipping Grafana ini configuration."
else
    info "Configuring Grafana sub-path in ${GRAFANA_INI} ..."

    # Idempotent: update root_url and serve_from_sub_path under [server]
    # Uses a Python one-liner to avoid awk/sed complexity with ini files
    python3 - "${GRAFANA_INI}" <<'PYEOF'
import sys, re

path = sys.argv[1]
with open(path, 'r') as f:
    content = f.read()

# Ensure [server] section has root_url and serve_from_sub_path
# Replace existing values if present, otherwise append after [server]
def set_ini_key(text, section, key, value):
    # Match key = anything under the given section
    pattern = re.compile(
        r'(?m)(^\[' + re.escape(section) + r'\][^\[]*?)^(' + re.escape(key) + r'\s*=.*?)$',
        re.MULTILINE | re.DOTALL
    )
    replacement = r'\1' + key + ' = ' + value
    if pattern.search(text):
        return pattern.sub(replacement, text, count=1)
    # Key not found — append after [server] header line
    section_pattern = re.compile(r'(?m)^(\[' + re.escape(section) + r'\])')
    return section_pattern.sub(r'\1\n' + key + ' = ' + value, text, count=1)

content = set_ini_key(content, 'server', 'root_url', '%(protocol)s://%(domain)s/grafana/')
content = set_ini_key(content, 'server', 'serve_from_sub_path', 'true')

with open(path, 'w') as f:
    f.write(content)

print("  root_url and serve_from_sub_path set")
PYEOF

    success "Grafana ini updated"
fi

# --- Enable and start services ---

SERVICES=()

# Prometheus
if systemctl list-unit-files prometheus.service &>/dev/null; then
    SERVICES+=("prometheus")
else
    warn "prometheus.service not found — skipping (install prometheus package first)"
fi

# Node exporter
if systemctl list-unit-files prometheus-node-exporter.service &>/dev/null; then
    SERVICES+=("prometheus-node-exporter")
elif systemctl list-unit-files node_exporter.service &>/dev/null; then
    SERVICES+=("node_exporter")
else
    warn "node_exporter service not found — skipping (install prometheus-node-exporter package first)"
fi

# Grafana
if systemctl list-unit-files grafana-server.service &>/dev/null; then
    SERVICES+=("grafana-server")
else
    warn "grafana-server.service not found — skipping (install grafana package first)"
fi

for svc in "${SERVICES[@]}"; do
    info "Enabling and starting ${svc} ..."
    systemctl enable "${svc}"
    systemctl restart "${svc}"
    success "${svc} running"
done

# --- Add Prometheus datasource to Grafana ---

if [[ " ${SERVICES[*]} " == *" grafana-server "* ]]; then
    info "Waiting for Grafana to be ready ..."
    for i in $(seq 1 15); do
        if curl -sf http://localhost:3000/api/health &>/dev/null; then
            break
        fi
        sleep 2
        if [[ $i -eq 15 ]]; then
            warn "Grafana did not become ready in time. Skipping datasource provisioning."
            GRAFANA_READY=false
        fi
    done
    GRAFANA_READY="${GRAFANA_READY:-true}"

    if [[ "${GRAFANA_READY}" == "true" ]]; then
        info "Adding Prometheus datasource to Grafana ..."
        HTTP_STATUS=$(curl -sf -o /dev/null -w "%{http_code}" \
            -X POST http://localhost:3000/api/datasources \
            -H "Content-Type: application/json" \
            -u admin:admin \
            -d '{
                "name": "Prometheus",
                "type": "prometheus",
                "url": "http://localhost:9090",
                "access": "proxy",
                "isDefault": true
            }' || true)

        if [[ "${HTTP_STATUS}" == "200" || "${HTTP_STATUS}" == "409" ]]; then
            # 409 = already exists — idempotent
            success "Prometheus datasource configured (status: ${HTTP_STATUS})"
        else
            warn "Datasource API returned HTTP ${HTTP_STATUS}. Check Grafana credentials or configure manually."
        fi
    fi
fi

# --- Print access URLs ---

echo ""
echo -e "${GREEN}Monitoring stack is up.${NC}"
echo ""
echo "  Prometheus:    http://localhost:9090"
echo "  Node Exporter: http://localhost:9100/metrics"
echo "  Grafana:       http://localhost:3000  (also at /grafana/ via Nginx)"
echo ""
warn "Default Grafana credentials are admin/admin — change them immediately."
echo ""
