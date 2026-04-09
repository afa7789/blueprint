.PHONY: up down logs backend frontend build deploy ngrok

# === Local Development (tudo via Docker, 1 porta) ===

up:
	@docker compose up -d --build
	@echo ""
	@echo "  ┌──────────────┬──────────────────────────────────────┐"
	@echo "  │ Service      │ URL                                  │"
	@echo "  ├──────────────┼──────────────────────────────────────┤"
	@echo "  │ App          │ http://localhost                     │"
	@echo "  │ API          │ http://localhost/api/v1/...          │"
	@echo "  │ Health       │ http://localhost/healthz             │"
	@echo "  │ pgweb        │ http://localhost/pgweb/              │"
	@echo "  │ Grafana      │ http://localhost/grafana/ (admin/bp) │"
	@echo "  │ Prometheus   │ http://localhost/prometheus/         │"
	@echo "  └──────────────┴──────────────────────────────────────┘"
	@echo ""
	@echo "  Share: ngrok http 80"
	@echo ""
	@echo "  Logs:   make logs"
	@echo "  Stop:   make down"
	@echo "  Share:  make ngrok"

down:
	docker compose down

logs:
	docker compose logs -f

logs-backend:
	docker compose logs -f backend

logs-frontend:
	docker compose logs -f frontend

# Share via ngrok (single port)
ngrok:
	@echo "Sharing http://localhost:80 via ngrok..."
	@echo "Run: ngrok http 80"
	@ngrok http 80

# === Run without Docker (local Go + Bun) ===

backend:
	@set -a && . ./.env && set +a && cd backend && go run ./cmd/server

frontend:
	cd frontend && bun run dev

# === Build (production binaries) ===

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

# === Admin User ===

# Promote user to admin (user must already exist - register first at /register)
# make promote-admin email=example@e-mail.com
promote-admin:
	@docker compose exec -T postgres psql -U blueprint -d blueprint -c "UPDATE users SET role = 'admin' WHERE email = '$(email)';"
