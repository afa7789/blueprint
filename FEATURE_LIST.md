# Feature List

## 1. Landing Page / Geral
- [x] Landing Page com Waitlist
- [x] Footer dinâmico (Linktree, Loja — só exibe se ativado via feature flags)
- [x] Loja visível sem login
- [x] Login com Roles (Admin, Operador, Usuário)
- [x] JWT auth (access + refresh tokens, httpOnly cookies)

---

## 2. Painel Admin
- [x] User Control + Upgrade de roles
- [x] Feature Flags (13 toggles, DB-backed)
- [x] E-mail System (grupos, subscriptions, desabilitar)
- [x] Banner System (targeting por perfil, scheduling, order)
- [x] Blog with AI (CRUD, upload de imagens, gerar com IA placeholder)
- [x] Linktree System (CRUD + reordenar)
- [x] Brand Kit (cores, logo, favicon, fontes)
- [x] User Groups (desconto por grupo)

---

## 3. Sistema de Imagens (Global)
- [x] Storage configurável: Local (S3 ready via interface)
- [x] Upload helper com UUID naming
- [x] Usado por Blog (cover), futuramente Products

---

## 4. Sistema de Pagamentos
- [x] Stripe (PaymentIntent + webhook para confirmação)
- [x] Interface/trait para extensibilidade
- [x] PIX Auto (stub, ready para gateway)
- [x] PIX Manual (admin aprova → status "paid")
- [x] Rollback automático de estoque em cancelamento

---

## 5. Loja (Store)
- [x] Habilitável via feature flag
- [x] Produtos PRÉ-VENDA (aparece quando estoque esgotado)
- [x] Loja pública (sem login)
- [x] Carrinho client-side (Pinia + localStorage)
- [x] Painel do usuário: pedidos e rastreio
- [x] Desconto por grupos de usuário
- [x] Sistema de cupons (percentual/fixo, validade, min compra, max usos)

---

## 6. Painel do Usuário
- [x] Login / Register / Forgot Password
- [x] Histórico de pedidos + tracking
- [ ] Atualizar perfil (pendente)
- [ ] Adicionar cartão salvo (pendente — depende Stripe Customer)

---

## 7. Health Monitor (Standalone Binary)

### Check Services

| Serviço | Como verifica | Tipo |
|---------|---------------|------|
| Redis | PING + DBSIZE | Crítico |
| PostgreSQL | Conexão + contagem de tabelas | Crítico |
| SMTP | TCP dial host:porta | Degraded |
| Telegram Bot | HTTP GET /getMe | Degraded |
| Disco | % espaço livre (alerta < 20%) | Alerta |
| Memória | runtime.MemStats | Info |
| Backup | Idade do último .dump.gz (alerta > 25h) | Alerta |
| SSL | TCP dial :443 | Degraded |
| Frontend | Verifica se dist/index.html existe | Degraded |
| API | Health check interno | Crítico |

- [x] 10 checks implementados
- [x] Status: healthy / degraded / unhealthy (503)
- [x] Dashboard HTML embutido (go:embed, auto-refresh 30s)
- [x] JSON endpoint para load balancers
- [x] Telegram alerts on status change
- [x] Check interval: 60s

---

## 8. Frontend PWA
- [x] VitePWA + Workbox
- [x] Precaching de assets críticos (54 entries)
- [x] Code splitting + lazy loading (todas as rotas)
- [x] Service Worker update toast
- [x] Cache strategies: NetworkFirst (API), CacheFirst (static)
- [x] Dexie (IndexedDB) setup
- [ ] Imagens responsivas (WebP/AVIF) — pendente
- [ ] Notificações push — pendente
- [ ] Lighthouse PWA score ≥ 90 — não testado

---

## 9. Toggle Features (DB-backed, 13 flags)

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

---

## 10. Painel do Operador
- [x] Visualizar pedidos pagos (prontos para envio)
- [x] Marcar como enviado + código de rastreio
- [x] Aprovar PIX manual
- [x] Acessível para admin + operador

---

## 11. Gerenciamento de Jobs e Cron
- [x] Dashboard de jobs com status (active/paused)
- [x] Cron scheduler DB-backed (robfig/cron)
- [x] Pausar / Resumir / Executar imediatamente
- [x] Histórico de execuções (duração, erro, output)
- [x] Reprocessamento de falhas (retry)
- [x] Registry de handlers registráveis

---

## 12. Admin Tools Hub
- [x] Grid de ferramentas por categoria
- [x] 5 tools seedados (pgweb, Redis, MinIO, Grafana, Prometheus)
- [x] URLs via ENV config com override no DB
- [x] Health ping por ferramenta
- [x] CRUD para gerenciar links
- [x] Controle de acesso por role

---

## 13. Logs e Observabilidade
- [x] Structured logger (stdout + DB)
- [x] Visualizador de logs em tempo real (SSE streaming)
- [x] Filtros: level, source, data, busca
- [x] Audit trail automático (middleware em todas mutations admin)
- [x] Configuração de retenção (default 30 dias)
- [x] Cleanup manual de logs antigos
- [ ] Integração Loki/ELK — pendente (placeholder)
- [ ] Distributed Tracing (OpenTelemetry) — pendente

---

## 14. DevOps & Deployment Scripts
- [x] `setup-vps.sh` — Provisionamento completo VPS Ubuntu (Go, Bun, PG, Redis, Nginx+Brotli, Certbot, pgweb, Prometheus, Grafana, Node Exporter, systemd units, UFW)
- [x] `setup-vps-runner.sh` — Orquestrador remoto (scp + ssh)
- [x] `install.sh` — Instalação lightweight de dependências
- [x] `deploy.sh` — Deploy zero-downtime (build local → rsync → health check → rollback automático)
- [x] `start.sh` / `stop.sh` / `restart.sh` — Controle de serviço systemd
- [x] `rollback.sh` — Rollback para versão anterior
- [x] `health.sh` — 6 checks com JSON output + Telegram alert
- [x] `backup.sh` — pg_dump + retenção 7d/4w/12m + S3 opcional
- [x] `monitor.sh` — 6 checks + Telegram (cron 5min)
- [x] `check-perf.sh` — Verifica Brotli, gzip, HTTP/2, cache, security headers
- [x] `setup-nginx.sh` — Gerador de config Nginx (SSL, Brotli, rate limiting, SSE, SPA)
- [x] `setup-monitoring.sh` — Grafana + Prometheus setup
- [x] `crontab.example` — Template de cron (backup, monitor, certbot)
- [x] Systemd units (blueprint-api, blueprint-health)
- [x] Docker Compose local (PG, Redis, pgweb, Prometheus, Grafana)
