.PHONY: up down backend frontend dev logs pgweb

# Start infrastructure (Postgres + Redis + pgweb)
up:
	docker compose up -d
	@echo ""
	@echo "  PostgreSQL: localhost:5433"
	@echo "  Redis:      localhost:6380"
	@echo "  pgweb:      http://localhost:8082"
	@echo ""

down:
	docker compose down

# Run backend (auto-migrates on boot)
backend:
	@set -a && . ./.env && set +a && cd backend && go run ./cmd/server

# Run frontend
frontend:
	cd frontend && bun run dev

# Run everything (use with: make up && make dev)
dev:
	@echo "Run in separate terminals:"
	@echo "  make up       # infrastructure"
	@echo "  make backend  # API on :8080"
	@echo "  make frontend # UI on :5173"
	@echo "  pgweb:        http://localhost:8082"

logs:
	docker compose logs -f

pgweb:
	@echo "http://localhost:8081"
	@open http://localhost:8082 2>/dev/null || xdg-open http://localhost:8082 2>/dev/null || echo "Open http://localhost:8082"
