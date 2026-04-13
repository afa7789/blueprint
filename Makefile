.PHONY: up down logs backend frontend build deploy ngrok \
	test test-backend test-frontend \
	lint lint-backend lint-frontend \
	fmt fmt-backend fmt-frontend \
	typecheck tidy vet \
	check ci clean help

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

# === Quality: test, lint, format, typecheck ===

## Run all tests (backend + frontend)
test: test-backend test-frontend

test-backend:
	cd backend && go test ./... -race -count=1

test-frontend:
	cd frontend && bun run test

## Lint (static analysis). Backend uses golangci-lint; frontend uses vue-tsc typecheck.
lint: lint-backend lint-frontend

lint-backend:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install: brew install golangci-lint"; exit 1; }
	cd backend && golangci-lint run ./...

lint-frontend: typecheck

typecheck:
	cd frontend && bun run vue-tsc --noEmit

## Format code in place
fmt: fmt-backend fmt-frontend

fmt-backend:
	cd backend && gofmt -s -w . && go vet ./...

fmt-frontend:
	@echo "(no formatter configured for frontend — skipping)"

vet:
	cd backend && go vet ./...

tidy:
	cd backend && go mod tidy

## Aggregate checks used locally and in CI
check: fmt-backend lint test

ci: lint test
	@echo "CI checks passed."

clean:
	rm -rf build/ frontend/dist/ backend/tmp/

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# === Admin User ===

# Promote user to admin (user must already exist - register first at /register)
# make promote-admin email=example@e-mail.com
promote-admin:
	@docker compose exec -T postgres psql -U blueprint -d blueprint -c "UPDATE users SET role = 'admin' WHERE email = '$(email)';"
