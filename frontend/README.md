# Frontend - Vue 3 + TypeScript + Vite

**Blueprint Frontend** - SPA com PWA ready, landing page, admin panel, loja, blog e mais.

## Pré-requisitos

- Node.js 18+
- Bun (opcional)

## Quick Start

```bash
cd frontend
bun install
bun run dev      # Development: http://localhost:5173
bun run build    # Production build
bun run preview  # Preview production build
```

## Variáveis de Ambiente (.env)

```env
VITE_API_URL=http://localhost:8080
VITE_APP_TITLE=My Service
VITE_ENABLE_BLOG=true
VITE_ENABLE_STORE=true
VITE_ENABLE_LINKTREE=true
```

## Features Toggle

Ative/desative módulos via variáveis de ambiente:
- `VITE_ENABLE_BLOG` - Blog com IA
- `VITE_ENABLE_STORE` - Loja com pagamentos
- `VITE_ENABLE_LINKTREE` - Sistema de links
- `VITE_ENABLE_BRAND_KIT` - Personalização de marca

## Tech Stack

- Vue 3 + Vite + TypeScript
- Pinia (state management)
- Dexie (IndexedDB offline)
- VitePWA (Workbox)
- Sass

## Licença

MIT - Livre para uso comercial. See [../LICENSE](../LICENSE)
