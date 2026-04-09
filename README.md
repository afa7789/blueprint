<p align="center">
  <img src="resources/banner.png" alt="Blueprint Banner" style="max-width: 65%; height: auto;">
</p>

# Blueprint - Service Starter Kit

**MIT License - Free to use, including commercial projects.**

Full-stack starter kit: Go backend + Vue 3 frontend. Landing page, admin panel, e-commerce, payments (Stripe + PIX), blog with AI, health monitoring, job scheduler, logs, and more.

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+ / Bun
- Docker & Docker Compose

### 1. Infrastructure

```bash
docker compose up -d
```

| Service    | Port | URL |
|------------|------|-----|
| PostgreSQL | 5433 | - |
| Redis      | 6380 | - |
| pgweb      | 8083 | http://localhost:8083 |
| Prometheus | 9090 | http://localhost:9090 |
| Grafana    | 3001 | http://localhost:3001 (admin/blueprint) |

### 2. Backend

```bash
make backend
```

Or manually:

```bash
set -a && . ./.env && set +a && cd backend && go run ./cmd/server
```

Migrations run automatically on boot. API at http://localhost:8080

### 3. Frontend

```bash
cd frontend
bun install    # first time only
bun run dev
```

Frontend at http://localhost:5173

### 4. Create Admin User

Register via http://localhost:5173/register, then promote:

```bash
docker exec blueprint-postgres-1 psql -U blueprint -d blueprint \
  -c "UPDATE users SET role='admin' WHERE email='your@email.com';"
```

Re-login to get admin token.

### 5. Stop

```bash
docker compose down       # stop infra
docker compose down -v    # stop + delete data
```

## Features

### Core
- Landing Page + Waitlist
- Dynamic Footer (shows Linktree/Store only if enabled)
- Auth with JWT + Roles (Admin, Operator, User)
- Feature Flags (13 toggles, DB-backed)

### Admin Panel
- User management + role control
- Feature flag toggles
- Banner system (targeting, scheduling, cache)
- Linktree management (drag reorder)
- Brand Kit (colors, logo, favicon, fonts)
- Email groups + subscriptions
- Blog management with AI generation
- Product/Category/Order management
- Coupon management

### E-Commerce (Store)
- Public product catalog with categories
- Pre-sale support (shows when stock exhausted)
- Client-side cart (Pinia + localStorage)
- Order creation with stock validation
- Coupon system (percentage/fixed, min purchase, expiry)
- User order history + tracking

### Payments
- Stripe (PaymentIntent + webhook)
- PIX Auto (stub, ready for gateway integration)
- PIX Manual (admin approval flow)
- Payment provider interface for extensibility

### Blog
- Public listing + slug-based detail
- Admin CRUD with draft/published
- Cover image upload (local storage, S3 ready)
- AI content generation (OpenAI placeholder)
- Auto-slug generation from title

### Health Monitor
- Standalone binary (`cmd/health/`)
- 10 checks: Redis, PostgreSQL, SMTP, Telegram, Disk, Memory, Backup, SSL, Frontend, API
- Status: healthy / degraded / unhealthy (503)
- Embedded HTML dashboard (go:embed)
- JSON endpoint for load balancers
- Telegram alerts on status change
- 60s check interval, 30s dashboard refresh

### Jobs & Cron (Module 10)
- DB-backed cron scheduler (robfig/cron)
- Job registry (handler name -> Go function)
- Pause/Resume/Run Now
- Execution history with duration + error tracking
- Retry failed executions

### Admin Tools Hub (Module 11)
- Links to external tools (pgweb, Redis, MinIO, Grafana, Prometheus)
- URLs from ENV config with DB override
- Health ping per tool
- CRUD for managing tool links

### Logs & Observability (Module 12)
- Structured logger (stdout + DB)
- Audit trail on all admin mutations (automatic middleware)
- Real-time log streaming (SSE)
- Filterable log viewer (level, source, date, search)
- Configurable retention + manual cleanup

### Operator Panel
- Orders ready to ship (status = paid)
- Mark shipped + tracking code
- PIX manual approval

### PWA
- VitePWA + Workbox
- Runtime caching (NetworkFirst for API, CacheFirst for static)
- Service Worker update toast
- Dexie (IndexedDB) for offline data
- Precaching of critical assets

## Architecture

```
Frontend :5173          Backend :8080          Health :8081
(Vue 3 + PWA)           (Go + Fiber)           (Standalone)
                             |
                    +--------+--------+
                    |                 |
              PostgreSQL :5433    Redis :6380
```

### Backend Structure

```
backend/
  cmd/
    server/main.go          # API entry (140 handlers)
    health/main.go           # Health monitor (standalone)
  internal/
    domain/
      entity.go              # 23 Go structs
      repository.go          # 19 repository interfaces
    infrastructure/
      adapter.go             # pgx implementations
    handlers/
      auth.go                # JWT auth (register/login/refresh)
      admin.go               # User/Banner/Linktree/BrandKit/Email CRUD
      store.go               # Products/Categories/Orders
      payment.go             # Stripe + PIX
      blog.go                # Blog CRUD + AI
      jobs.go                # Cron scheduler
      tools.go               # Admin tools hub
      logs.go                # Logs + Audit
      ...
  pkg/
    config/config.go         # ENV-based config
    database/                # pgxpool + redis + migrations
    middleware/               # JWT auth, RBAC, audit log
    logger/logger.go         # Structured logger (stdout + DB)
  migrations/                # 13 SQL migrations (auto-run on boot)
```

### Frontend Structure

```
frontend/src/
  services/
    api.ts                   # Fetch wrapper
    featureFlags.ts          # Flag service
  stores/
    auth.ts                  # Auth (Pinia)
    cart.ts                  # Cart (Pinia + localStorage)
  views/
    Landing.vue              # Landing + Waitlist
    Login.vue / Register.vue # Auth
    Linktree.vue             # Public linktree
    Store/                   # Product listing, detail, cart, checkout, orders
    Blog/                    # Blog listing + post detail
    Admin/                   # 11 admin views
    Operator/                # Operator dashboard
  composables/
    useServiceWorker.ts      # PWA update detection
  db/index.ts                # Dexie setup
```

## Configuration

### Backend (.env)

```env
DATABASE_URL=postgres://blueprint:blueprint@localhost:5433/blueprint?sslmode=disable
DATABASE_MIGRATION_URL=postgres://blueprint:blueprint@localhost:5433/blueprint?sslmode=disable
REDIS_URL=redis://localhost:6380
JWT_SECRET=change-me-in-production
PORT=8080
ENV=development
FRONTEND_URL=http://localhost:5173

# Storage
STORAGE_TYPE=local
UPLOAD_DIR=./uploads

# Payments (optional)
STRIPE_KEY=
STRIPE_WEBHOOK_SECRET=

# AI (optional)
OPENAI_KEY=

# Alerts (optional)
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=

# Admin Tools URLs (optional)
PGWEB_URL=http://localhost:8083
REDIS_COMMANDER_URL=
MINIO_URL=
GRAFANA_URL=
PROMETHEUS_URL=

# Email (optional)
SMTP_HOST=localhost
SMTP_PORT=587
```

### Frontend (.env)

```env
VITE_API_URL=http://localhost:8080
VITE_APP_TITLE=Blueprint
```

## Feature Flags

13 flags toggled via Admin Panel (DB-backed):

| Flag | Default |
|------|---------|
| store_enabled | true |
| blog_enabled | true |
| waitlist_enabled | true |
| payments_stripe | true |
| payments_pix | true |
| pwa_enabled | true |
| ai_blog_enabled | true |
| linktree_enabled | true |
| brand_kit_enabled | true |
| helper_boxes_enabled | true |
| pix_auto_enabled | true |
| pix_manual_enabled | true |
| email_auto_enabled | true |

## API Overview

### Public
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /healthz | DB + Redis status |
| GET | /api/v1/features | Feature flags |
| GET | /api/v1/products | Product listing |
| GET | /api/v1/products/:id | Product detail |
| GET | /api/v1/categories | Categories |
| GET | /api/v1/blog | Published posts |
| GET | /api/v1/blog/:slug | Post by slug |
| GET | /api/v1/linktree | Active links |
| GET | /api/v1/banners | Active banners |
| GET | /api/v1/brand-kit | Brand kit |
| POST | /api/v1/waitlist | Join waitlist |
| POST | /api/v1/coupons/validate | Validate coupon |

### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/auth/register | Register |
| POST | /api/v1/auth/login | Login (JWT) |
| POST | /api/v1/auth/refresh | Refresh token |
| POST | /api/v1/auth/logout | Logout |
| GET | /api/v1/auth/me | Current user |
| POST | /api/v1/auth/forgot-password | Request reset |
| POST | /api/v1/auth/reset-password | Reset password |

### Auth-Required
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/orders | Create order |
| GET | /api/v1/orders/me | My orders |
| POST | /api/v1/payments/stripe | Stripe payment |
| POST | /api/v1/payments/pix | PIX payment |

### Admin (40+ endpoints)
All under `/api/v1/admin/` — users, products, categories, orders, coupons, banners, linktree, brand-kit, email-groups, email-subscriptions, user-groups, blog, jobs, tools, logs, audit.

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go + Fiber |
| Database | PostgreSQL (pgx, no ORM) |
| Cache | Redis |
| Frontend | Vue 3 + Vite + TypeScript |
| State | Pinia |
| PWA | VitePWA (Workbox) |
| Offline Storage | Dexie (IndexedDB) |
| Auth | JWT (httpOnly cookies + Bearer) |
| Payments | Stripe + PIX |
| Cron | robfig/cron/v3 |
| Migrations | golang-migrate (embedded FS) |

## Database

24 tables across 13 migrations. All use UUID primary keys, JSONB for flexible data, and proper indexes. Migrations auto-run on server boot.

## License

MIT - See [LICENSE](LICENSE)

## Contributing

See [CONTRIBUTION.md](CONTRIBUTION.md)
