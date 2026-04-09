# Feature List

## 1. Landing Page / Geral
- [ ] Landing Page com possibilidade de Waitlist
- [ ] Footer dinâmico (Linktree, Loja, etc. — só exibe se ativado)
- [ ] Loja visível mesmo para quem não está logado
- [ ] Login com Roles (Admin, Operador, Usuário)

---

## 2. Painel Admin
- [ ] User Control + Upgrade de roles
- [ ] Env Change (com backup automático + aviso)
- [ ] E-mail System (grupos, histórico, contador, desabilitar automático)
- [ ] Banner System (cache configurável, não repeat, duração, por perfil, páginas automáticas)
- [ ] Blog with AI (escrever, importar artigo, upload de imagens, gerar com IA)
- [ ] Linktree System (habilitar/desabilitar + reordenar)
- [ ] Brand Kit (habilitar/desabilitar)
- [ ] Color System + tema completo
- [ ] Helper Menu (desabilitar boxes)
- [ ] Dados de backend salvos com backup

---

## 3. Sistema de Imagens (Global)
- [ ] Storage configurável: Local, S3, CDN (chaves nas ENVs)
- [ ] Utilizado por Blog, Banners, Produtos da Loja, etc.

---

## 4. Sistema de Pagamentos
- [ ] Stripe como padrão + Trait/Interface para outros gateways
- [ ] Pagamento único e recorrente
- [ ] PIX Automático (habilitar/desabilitar)
- [ ] PIX Manual (cliente envia comprovante → admin confere → OK/Cancelar com devolução automática de estoque)

---

## 5. Loja (Store)
- [ ] Habilitar/desabilitar módulo
- [ ] Produtos PRÉ-VENDA (pedido antecipado; aparece apenas quando o estoque estiver esgotado)
- [ ] Loja pública (sem login)
- [ ] Carrinho + integração com e-mail
- [ ] Painel do usuário para visualizar pagamentos, pedidos e rastreio
- [ ] Aplicar descontos para grupos (e-mails ou tipos de usuário)
- [ ] Sistema de cupoms

---

## 6. Painel do Usuário
- [ ] Atualizar perfil
- [ ] Forget Password (sempre habilitado)
- [ ] Adicionar cartão salvo
- [ ] Histórico completo de pagamentos e pedidos

**Observação:**
- O painel do usuário também estará disponível para administradores.

---

## 7. Helper Server + Health Monitor

### Check Services

| Serviço       | Como verifica                               | Tipo     |
|----------------|----------------------------------------------|----------|
| Redis          | PING + DBSIZE                               | Crítico  |
| PostgreSQL     | Conexão + contagem de tabelas               | Crítico  |
| SMTP           | TCP dial host:porta                         | Degraded |
| Telegram Bot   | HTTP GET /getMe                             | Degraded |
| Disco          | % de espaço livre (alerta < 20%)            | Alerta   |
| Memória        | runtime.MemStats                            | Info     |
| Backup         | Idade do último `.dump.gz` (alerta > 25h)   | Alerta   |
| SSL            | TCP dial :443                               | Degraded |
| Frontend       | Verifica se `dist/index.html` existe        | Degraded |
| API            | Health check interno                        | Crítico  |

### Status Levels
- **healthy** — todos os serviços críticos operacionais
- **degraded** — falha em serviço não crítico (SMTP, Telegram)
- **unhealthy** — falha em serviço crítico, retorna HTTP 503

### Endpoints
- `GET /health` — Dashboard HTML (auto-refresh a cada 30s, embutido no binário)
- `GET /health?format=json` — JSON completo para load balancers

### Alerts
- Verificação a cada 60 segundos
- Alterações de status geram alerta automático no Telegram

---

## 8. Frontend PWA Features
- [ ] Workbox + VitePWA
- [ ] Precaching de assets críticos
- [ ] Code splitting + lazy loading
- [ ] Critical CSS + Preload
- [ ] App Shell + Skeleton UI
- [ ] Lazy loading de imagens
- [ ] Otimização de fontes (swap, preload, self-hosted)
- [ ] Favicons e ícones PWA completos
- [ ] Verificação de nova versão (Service Worker update listener + UI)
- [ ] Estratégias de cache específicas (API vs assets)
- [ ] Offline-first com IndexedDB (migração de localStorage para Dexie)
- [ ] Imagens responsivas (WebP/AVIF + srcset)
- [ ] Estratégia de cache para APIs (Network First + fallback em IndexedDB)
- [ ] Badge de atualização pendente (toast após `updatefound`)
- [ ] Notificações push (futuro, requer backend)
- [ ] Testes Lighthouse PWA (score ≥ 90)

---

## 9. Toggle Features (Enable/Disable)

Lista de módulos que podem ser habilitados ou desabilitados:
- [ ] Blog
- [ ] Linktree
- [ ] Brand Kit
- [ ] Loja
- [ ] Envio automático de e-mail
- [ ] Validação automática PIX
- [ ] Validação manual PIX
- [ ] Helper Boxes

---

## 10. Painel do Operador da Loja
- [ ] Auxiliar no processo de entregas
- [ ] Visualizar pedidos prontos para envio (pagamento confirmado)
- [ ] Acessar dados de envio do cliente
- [ ] Confirmar envio com upload de comprovantes e recibos
- [ ] Registrar código de rastreio

**Observações:**
- O painel do operador também estará disponível para administradores.
- O operador poderá acessar funcionalidades do painel de usuário para realizar compras.

---

## 11. Gerenciamento de Jobs e Cron
- [ ] Dashboard de tarefas em background
- [ ] Controle de cron (pausar, executar e reiniciar)
- [ ] Monitoramento de filas
- [ ] Execução manual de jobs
- [ ] Histórico de execuções
- [ ] Reprocessamento de falhas
- [ ] Definição de prioridades e filas
- [ ] Alertas automáticos em caso de erro

---

## 12. Admin Tools Hub
- [ ] Hub de ferramentas administrativas
- [ ] Links para ferramentas externas
- [ ] Integração com PostgreSQL (pgAdmin, Adminer)
- [ ] Integração com Redis (RedisInsight)
- [ ] Integração com MinIO ou S3 Console
- [ ] Integração com Grafana e Prometheus
- [ ] Integração com Portainer
- [ ] Controle de acesso por roles
- [ ] Restrição por ambiente (Development, Staging, Production)
- [ ] Configuração via variáveis de ambiente (ENV)

---

## 13. Logs e Observabilidade (SOTA)
- [ ] Visualizador de logs em tempo real (streaming)
- [ ] Centralização de logs
- [ ] Filtros por serviço, data e nível
- [ ] Busca avançada e indexação
- [ ] Download e exportação de logs
- [ ] Alertas automatizados baseados em eventos
- [ ] Auditoria de ações administrativas
- [ ] Integração com Grafana Loki
- [ ] Integração com ELK Stack (Elasticsearch, Logstash, Kibana)
- [ ] Integração com Prometheus e Grafana
- [ ] Distributed Tracing com OpenTelemetry e Jaeger
- [ ] Retenção configurável de logs
- [ ] Monitoramento e correlação de eventos