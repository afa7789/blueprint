# Deploy Guide - Blueprint

## Requisitos

- Go 1.21+
- Node.js 18+
- PostgreSQL
- Redis

## Backend

### Desenvolvimento
```bash
cd backend
go mod download
go run cmd/server/main.go
# Runs on http://localhost:8080
```

### Produção
```bash
cd backend
go build -o server cmd/server/main.go
./server --config config.yaml
```

### Variáveis Obrigatórias (config.yaml)
```yaml
host: "0.0.0.0"
port: 8080

database:
  url: "postgres://user:pass@localhost:5432/db"

redis:
  url: "redis://localhost:6379"
```

## Frontend

### Build Produção
```bash
cd frontend
bun install
bun run build
# Output em dist/
```

### Servir estaticamente
```bash
# Nginx example
server {
  root /var/www/blueprint/dist;
  index index.html;
  location / {
    try_files $uri $uri/ /index.html;
  }
}
```

## Health Monitor (Separado)

Deploy em subdomain dedicado (`health.yourservice.com`):

```bash
cd backend
go build -o health-monitor cmd/health/main.go
./health-monitor --config config.yaml
```

### Endpoints
- `GET /health` — Dashboard HTML
- `GET /health?format=json` — JSON para load balancers

## Docker

```dockerfile
# Backend
FROM golang:1.21-alpine
WORKDIR /app
COPY backend/ .
RUN go build -o server cmd/server/main.go
EXPOSE 8080
CMD ["./server", "--config", "config.yaml"]
```

```dockerfile
# Frontend
FROM node:18-alpine
WORKDIR /app
COPY frontend/ .
RUN npm install && npm run build
# Servir com nginx
```

## Docker Compose

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: blueprint
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redisdata:/data

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    volumes:
      - ./config.yaml:/app/config.yaml

  frontend:
    build: ./frontend
    ports:
      - "3000:80"
    depends_on:
      - backend

volumes:
  pgdata:
  redisdata:
```

## HTTPS / SSL

Configure no nginx ou use CDN (Cloudflare, AWS CloudFront).

## Monitoramento

O Health Monitor já inclui:
- Verificações: Redis, PostgreSQL, SMTP, Telegram, Disco, Memória, Backup, SSL, Frontend, API
- Alertas Telegram automáticos
- Dashboard HTML embutido

## Licença

MIT - Livre para uso comercial. See [../LICENSE](../LICENSE)
