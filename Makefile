.PHONY: up down backend frontend dev logs pgweb deploy deploy-backend deploy-frontend

# === Local Development ===

up:
	docker compose up -d
	@echo ""
	@echo "  PostgreSQL: localhost:5433"
	@echo "  Redis:      localhost:6380"
	@echo "  pgweb:      http://localhost:8083"
	@echo ""

down:
	docker compose down

backend:
	@set -a && . ./.env && set +a && cd backend && go run ./cmd/server

frontend:
	cd frontend && bun run dev

dev:
	@echo "Run in separate terminals:"
	@echo "  make up       # infrastructure"
	@echo "  make backend  # API on :8080"
	@echo "  make frontend # UI on :5173"
	@echo "  pgweb:        http://localhost:8083"

logs:
	docker compose logs -f

pgweb:
	@open http://localhost:8083 2>/dev/null || echo "Open http://localhost:8083"

# === Build ===

build-backend:
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../build/blueprint-api ./cmd/server
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../build/blueprint-health ./cmd/health

build-frontend:
	cd frontend && bun run build

build: build-backend build-frontend

# === Deploy (VPS) ===

deploy-backend:
	bash scripts/deploy.sh --backend

deploy-frontend:
	bash scripts/deploy.sh --frontend

deploy:
	bash scripts/deploy.sh --all

deploy-dry:
	bash scripts/deploy.sh --all --dry-run

# === VPS Setup ===

setup-vps:
	@echo "Usage: bash scripts/setup-vps-runner.sh user@your-vps [db_pass] [redis_pass]"

setup-nginx:
	@echo "Usage: sudo bash scripts/setup-nginx.sh your-domain.com"

setup-monitoring:
	@echo "Usage: sudo bash scripts/setup-monitoring.sh"
