<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import WaitlistForm from '../components/landing/WaitlistForm.vue'
import { loadSiteModules, siteModules } from '../services/siteModules'

const waitlistEnabled = computed(() => siteModules.waitlistEnabled)
const storeEnabled = computed(() => siteModules.storeEnabled)
const blogEnabled = computed(() => siteModules.blogEnabled)
const storeVisible = computed(() => siteModules.storeEnabled && siteModules.hasStoreContent)
const blogVisible = computed(() => siteModules.blogEnabled && siteModules.hasBlogContent)

onMounted(() => {
  loadSiteModules()
})

interface FeatureItem {
  text: string
  done: boolean
}

interface FeatureCategory {
  id: string
  title: string
  icon: string
  features: FeatureItem[]
  open: boolean
}

const categories = ref<FeatureCategory[]>([
  {
    id: 'landing',
    title: 'Landing Page & Auth',
    icon: '🏠',
    open: true,
    features: [
      { text: 'Landing page with waitlist', done: true },
      { text: 'Dynamic footer (Linktree, Store — feature flag gated)', done: true },
      { text: 'Store visible without login', done: true },
      { text: 'Login with roles (Admin, Operator, User)', done: true },
      { text: 'JWT auth (access + refresh tokens, httpOnly cookies)', done: true },
    ],
  },
  {
    id: 'admin',
    title: 'Admin Panel',
    icon: '⚙️',
    open: false,
    features: [
      { text: 'User control + role upgrade', done: true },
      { text: 'Feature flags (13 toggles, DB-backed)', done: true },
      { text: 'Email system (groups, subscriptions, disable)', done: true },
      { text: 'Banner system (targeting by profile, scheduling, order)', done: true },
      { text: 'Blog with AI (CRUD, image upload, AI generation placeholder)', done: true },
      { text: 'Linktree system (CRUD + reorder)', done: true },
      { text: 'Brand kit (colors, logo, favicon, fonts)', done: true },
      { text: 'User groups (email and discounts per group)', done: true },
      { text: 'Legal pages — Terms of Service + Privacy Policy (admin CRUD, HTML editor, footer links)', done: true },
      { text: 'Security settings (rate limits, login attempts, password policy — admin configurable)', done: true },
    ],
  },
  {
    id: 'images',
    title: 'Image System',
    icon: '🖼️',
    open: false,
    features: [
      { text: 'Configurable storage: Local (S3-ready via interface)', done: true },
      { text: 'Upload helper with UUID naming', done: true },
      { text: 'Used by Blog (cover), extensible to Products', done: true },
    ],
  },
  {
    id: 'payments',
    title: 'Payments',
    icon: '💳',
    open: false,
    features: [
      { text: 'Stripe (PaymentIntent + webhook for confirmation)', done: true },
      { text: 'Interface/trait for extensibility', done: true },
      { text: 'PIX Auto (stub, ready for gateway)', done: true },
      { text: 'PIX Manual (admin approves → status "paid")', done: true },
      { text: 'Automatic stock rollback on cancellation', done: true },
    ],
  },
  {
    id: 'store',
    title: 'Store (E-Commerce)',
    icon: '🛒',
    open: false,
    features: [
      { text: 'Togglable via feature flag', done: true },
      { text: 'Pre-sale products (shows when stock exhausted)', done: true },
      { text: 'Public store (no login required)', done: true },
      { text: 'Client-side cart (Pinia + localStorage)', done: true },
      { text: 'User panel: orders and tracking', done: true },
      { text: 'Discount by user group', done: true },
      { text: 'Coupon system (percentage/fixed, expiry, min purchase, max uses)', done: true },
    ],
  },
  {
    id: 'user',
    title: 'User Panel',
    icon: '👤',
    open: false,
    features: [
      { text: 'Login / Register / Forgot Password', done: true },
      { text: 'Email verification (feature flag toggle, Redis token, 24h expiry)', done: true },
      { text: 'Update profile (name, phone, avatar, address)', done: true },
      { text: 'Change password (verifies current password)', done: true },
      { text: 'Saved cards (Stripe Customer + SetupIntent + list/delete)', done: true },
      { text: 'Order history + tracking (status badges, tracking link)', done: true },
      { text: 'Dedicated layout with sidebar (/user/profile, /user/password, /user/cards, /user/orders)', done: true },
      { text: 'Nav bar with "My Account" / "Admin" / "Login" / "Logout"', done: true },
      { text: 'ENV warning when Stripe is not configured (STRIPE_KEY)', done: true },
    ],
  },
  {
    id: 'health',
    title: 'Health Monitor',
    icon: '📊',
    open: false,
    features: [
      { text: 'Redis check (PING + DBSIZE)', done: true },
      { text: 'PostgreSQL check (connection + table count)', done: true },
      { text: 'SMTP check (TCP dial host:port)', done: true },
      { text: 'Telegram Bot check (HTTP GET /getMe)', done: true },
      { text: 'Disk space check (alert < 20% free)', done: true },
      { text: 'Memory check (runtime.MemStats)', done: true },
      { text: 'Backup age check (alert > 25h)', done: true },
      { text: 'SSL check (TCP dial :443)', done: true },
      { text: 'Frontend build check (dist/index.html exists)', done: true },
      { text: 'API internal health check', done: true },
      { text: 'Status: healthy / degraded / unhealthy (503)', done: true },
      { text: 'Embedded HTML dashboard (go:embed, 30s auto-refresh)', done: true },
      { text: 'JSON endpoint for load balancers', done: true },
      { text: 'Telegram alerts on status change', done: true },
    ],
  },
  {
    id: 'pwa',
    title: 'Frontend PWA',
    icon: '📱',
    open: false,
    features: [
      { text: 'VitePWA + Workbox', done: true },
      { text: 'Precaching of critical assets (54 entries)', done: true },
      { text: 'Code splitting + lazy loading (all routes)', done: true },
      { text: 'Service Worker update toast', done: true },
      { text: 'Cache strategies: NetworkFirst (API), CacheFirst (static)', done: true },
      { text: 'Dexie (IndexedDB) setup', done: true },
      { text: 'Responsive images (WebP/AVIF)', done: false },
      { text: 'Push notifications', done: false },
      { text: 'Lighthouse PWA score ≥ 90', done: false },
    ],
  },
  {
    id: 'flags',
    title: 'Feature Flags',
    icon: '🚩',
    open: false,
    features: [
      { text: 'store_enabled', done: true },
      { text: 'blog_enabled', done: true },
      { text: 'waitlist_enabled', done: true },
      { text: 'payments_stripe', done: true },
      { text: 'payments_pix', done: true },
      { text: 'pwa_enabled', done: true },
      { text: 'ai_blog_enabled', done: true },
      { text: 'linktree_enabled', done: true },
      { text: 'brand_kit_enabled', done: true },
      { text: 'helper_boxes_enabled', done: true },
      { text: 'pix_auto_enabled', done: true },
      { text: 'pix_manual_enabled', done: true },
      { text: 'email_auto_enabled', done: true },
    ],
  },
  {
    id: 'operator',
    title: 'Operator Panel',
    icon: '📦',
    open: false,
    features: [
      { text: 'View paid orders (ready to ship)', done: true },
      { text: 'Mark as shipped + tracking code', done: true },
      { text: 'Approve manual PIX', done: true },
      { text: 'Accessible for admin + operator', done: true },
    ],
  },
  {
    id: 'jobs',
    title: 'Jobs & Cron',
    icon: '⏱️',
    open: false,
    features: [
      { text: 'Job dashboard with status (active/paused)', done: true },
      { text: 'DB-backed cron scheduler (robfig/cron)', done: true },
      { text: 'Pause / Resume / Execute immediately', done: true },
      { text: 'Execution history (duration, error, output)', done: true },
      { text: 'Failure reprocessing (retry)', done: true },
      { text: 'Registerable handler registry', done: true },
    ],
  },
  {
    id: 'tools',
    title: 'Admin Tools Hub',
    icon: '🔧',
    open: false,
    features: [
      { text: 'Tool grid by category', done: true },
      { text: '5 seeded tools (pgweb, Redis, MinIO, Grafana, Prometheus)', done: true },
      { text: 'URLs via ENV config with DB override', done: true },
      { text: 'Health ping per tool', done: true },
      { text: 'CRUD for managing links', done: true },
      { text: 'Role-based access control', done: true },
    ],
  },
  {
    id: 'logs',
    title: 'Logs & Observability',
    icon: '🔍',
    open: false,
    features: [
      { text: 'Structured logger (stdout + DB)', done: true },
      { text: 'Real-time log viewer (SSE streaming)', done: true },
      { text: 'Filters: level, source, date, search', done: true },
      { text: 'Automatic audit trail (middleware on all admin mutations)', done: true },
      { text: 'Retention config (default 30 days)', done: true },
      { text: 'Manual cleanup of old logs', done: true },
      { text: 'Loki/ELK integration', done: false },
      { text: 'Distributed tracing (OpenTelemetry)', done: false },
    ],
  },
  {
    id: 'security',
    title: 'Security & Route Protection',
    icon: '🔐',
    open: false,
    features: [
      { text: 'JWT auth (access 15min + refresh 7d, httpOnly cookies)', done: true },
      { text: 'RBAC middleware (admin, operator, user)', done: true },
      { text: 'Rate limiting — Redis-backed, per IP/email/user, configurable', done: true },
      { text: 'Security headers middleware (X-Frame-Options, HSTS, CSP)', done: true },
      { text: 'Request size limiting (default 10MB)', done: true },
      { text: 'Email verification on register (feature flag toggle)', done: true },
      { text: 'Admin-configurable security settings (DB table, no restart needed)', done: true },
      { text: 'Rate limit headers (X-RateLimit-Limit, Remaining, Reset)', done: true },
      { text: 'Graceful degradation (rate limiting skipped if Redis unavailable)', done: true },
      { text: 'Audit trail on all admin mutations', done: true },
    ],
  },
  {
    id: 'legal',
    title: 'Legal Pages',
    icon: '📄',
    open: false,
    features: [
      { text: 'DB-backed legal pages (terms, privacy, cookies, etc.)', done: true },
      { text: 'Admin CRUD with HTML content editor', done: true },
      { text: 'Slug-based public URLs (/legal/terms, /legal/privacy)', done: true },
      { text: 'Active/inactive toggle per page', done: true },
      { text: 'Auto-displayed in footer (dynamic, fetched from API)', done: true },
      { text: 'Seeded with generic Terms of Service + Privacy Policy', done: true },
    ],
  },
  {
    id: 'devops',
    title: 'DevOps & Deployment',
    icon: '🚀',
    open: false,
    features: [
      { text: 'setup-vps.sh — Full VPS provisioning (Go, Bun, PG, Redis, Nginx+Brotli, Certbot, systemd, UFW)', done: true },
      { text: 'deploy.sh — Zero-downtime deploy (local build → rsync → health check → auto-rollback)', done: true },
      { text: 'backup.sh — pg_dump + retention 7d/4w/12m + optional S3', done: true },
      { text: 'rollback.sh — Rollback to previous version', done: true },
      { text: 'setup-nginx.sh — Nginx config generator (SSL, Brotli, rate limiting, SSE, SPA)', done: true },
      { text: 'setup-monitoring.sh — Grafana + Prometheus setup', done: true },
      { text: 'monitor.sh — 6 checks + Telegram (cron 5min)', done: true },
      { text: 'check-perf.sh — Brotli, gzip, HTTP/2, cache, security headers check', done: true },
      { text: 'Docker Compose local (PG, Redis, pgweb, Prometheus, Grafana)', done: true },
      { text: 'Systemd units (blueprint-api, blueprint-health)', done: true },
    ],
  },
])

function toggleCategory(id: string) {
  const cat = categories.value.find(c => c.id === id)
  if (cat) cat.open = !cat.open
}

const totalFeatures = categories.value.reduce((sum, c) => sum + c.features.length, 0)
const doneFeatures = categories.value.reduce((sum, c) => sum + c.features.filter(f => f.done).length, 0)
</script>

<template>
  <div class="landing">
    <!-- Hero with banner -->
    <section class="hero">
      <img src="/banner-thin.svg" alt="Blueprint" class="hero-banner" />
      <h1>The <span class="accent">Starter Kit</span> for Modern Services</h1>
      <p class="hero-description">
        Full-stack foundation with Go + Vue 3. Landing page, admin panel, e-commerce, payments,
        blog, health monitoring, and production-grade deployment — all in one kit.
      </p>
      <WaitlistForm v-if="waitlistEnabled" />
      <div class="hero-links">
        <a href="https://github.com/afa7789/blueprint" target="_blank" rel="noopener noreferrer" class="btn btn-outline">
          <i class="fab fa-github"></i> GitHub
        </a>
        <router-link to="/register" class="btn btn-primary">Get Started</router-link>
        <router-link v-if="storeVisible" to="/store" class="btn btn-outline">Loja</router-link>
        <router-link v-if="blogVisible" to="/blog" class="btn btn-outline">Blog</router-link>
      </div>
    </section>

    <!-- Feature grid -->
    <section class="features-section">
      <h2 class="section-title">Everything You Need</h2>
      <div class="features">
        <div class="feature-card">
          <div class="feature-icon">🔐</div>
          <h3>Auth + Roles</h3>
          <p>JWT with refresh tokens, RBAC (admin, operator, user), email verification, rate limiting.</p>
        </div>
        <div class="feature-card" :class="{ dimmed: !storeEnabled }">
          <div class="feature-icon">🛒</div>
          <h3>E-Commerce <span v-if="!storeEnabled" class="flag-off">off</span></h3>
          <p>Products, categories, cart, orders, coupons, pre-sale, Stripe + PIX payments.</p>
        </div>
        <div class="feature-card" :class="{ dimmed: !blogEnabled }">
          <div class="feature-icon">📝</div>
          <h3>Blog + AI <span v-if="!blogEnabled" class="flag-off">off</span></h3>
          <p>Admin CRUD, image uploads, AI content generation, slug-based URLs.</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">⚙️</div>
          <h3>Admin Panel</h3>
          <p>Users, banners, linktree, brand kit, email groups, feature flags, legal pages.</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">📊</div>
          <h3>Monitoring</h3>
          <p>Health monitor (10 checks), structured logs, audit trail, Grafana + Prometheus.</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">🚀</div>
          <h3>Deploy Ready</h3>
          <p>Docker Compose, VPS scripts, Nginx + SSL, zero-downtime deploys, auto-rollback.</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">📱</div>
          <h3>PWA</h3>
          <p>Offline-first with Workbox, IndexedDB, service worker updates, cache strategies.</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">🎨</div>
          <h3>Brand Kit</h3>
          <p>Customizable colors, logo, fonts. Dynamic footer, linktree, banner system.</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">🔧</div>
          <h3>14 Feature Flags</h3>
          <p>Toggle any module on/off from admin. No redeploy needed.</p>
        </div>
      </div>
    </section>

    <!-- Complete Feature List -->
    <section class="feature-list-section">
      <div class="feature-list-header">
        <h2 class="section-title">Complete Feature List</h2>
        <div class="feature-list-stats">
          <div class="stat-badge">
            <span class="stat-number">{{ categories.length }}</span>
            <span class="stat-label">modules</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-badge">
            <span class="stat-number">{{ totalFeatures }}+</span>
            <span class="stat-label">features</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-badge">
            <span class="stat-number">{{ doneFeatures }}</span>
            <span class="stat-label">implemented</span>
          </div>
        </div>
      </div>

      <div class="categories-grid">
        <div
          v-for="cat in categories"
          :key="cat.id"
          class="category-block"
          :class="{ 'category-open': cat.open }"
        >
          <button class="category-header" @click="toggleCategory(cat.id)" :aria-expanded="cat.open">
            <span class="category-icon">{{ cat.icon }}</span>
            <span class="category-title">{{ cat.title }}</span>
            <span class="category-meta">
              <span class="category-count">{{ cat.features.filter(f => f.done).length }}/{{ cat.features.length }}</span>
              <span class="category-chevron">{{ cat.open ? '▲' : '▼' }}</span>
            </span>
          </button>
          <ul v-if="cat.open" class="feature-items">
            <li
              v-for="(feat, i) in cat.features"
              :key="i"
              class="feature-item"
              :class="feat.done ? 'feat-done' : 'feat-pending'"
            >
              <span class="feat-mark">{{ feat.done ? '✓' : '–' }}</span>
              <span class="feat-text">{{ feat.text }}</span>
            </li>
          </ul>
        </div>
      </div>
    </section>

    <!-- Tech stack -->
    <section class="tech-section">
      <h2 class="section-title">Tech Stack</h2>
      <div class="tech-grid">
        <div class="tech-item"><span class="tech-label">Backend</span> Go + Fiber</div>
        <div class="tech-item"><span class="tech-label">Frontend</span> Vue 3 + Vite + TS</div>
        <div class="tech-item"><span class="tech-label">Database</span> PostgreSQL (pgx)</div>
        <div class="tech-item"><span class="tech-label">Cache</span> Redis</div>
        <div class="tech-item"><span class="tech-label">Auth</span> JWT + bcrypt</div>
        <div class="tech-item"><span class="tech-label">Payments</span> Stripe + PIX</div>
        <div class="tech-item"><span class="tech-label">State</span> Pinia</div>
        <div class="tech-item"><span class="tech-label">PWA</span> VitePWA + Workbox</div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta">
      <h2>Ready to build?</h2>
      <p>Join the waitlist or dive straight into the code.</p>
      <div class="cta-buttons">
        <router-link to="/register" class="btn btn-primary btn-lg">Create Account</router-link>
        <router-link v-if="storeVisible" to="/store" class="btn btn-outline btn-lg">Browse Store</router-link>
        <router-link v-if="blogVisible" to="/blog" class="btn btn-outline btn-lg">Read Blog</router-link>
      </div>
    </section>
  </div>
</template>

<style scoped>
.landing {
  min-height: 100%;
  display: flex;
  flex-direction: column;
}

/* Hero */
.hero {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  background: linear-gradient(180deg, var(--accent-bg) 0%, transparent 100%);
}

.hero-banner {
  max-width: 100%;
  height: auto;
  margin-bottom: 8px;
}

.hero h1 {
  font-size: 48px;
  font-weight: 700;
  letter-spacing: -1.5px;
  margin: 0;
  line-height: 1.1;
  max-width: 700px;
}

.accent {
  color: var(--accent);
}

.hero-description {
  font-size: 18px;
  color: var(--text);
  max-width: 560px;
  line-height: 1.6;
}

.hero-links {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  padding: 10px 24px;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.2s;
  cursor: pointer;
  border: none;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
}

.btn-primary:hover {
  opacity: 0.9;
  box-shadow: 0 4px 12px rgba(38, 68, 236, 0.3);
}

.btn-outline {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text-h);
}

.btn-outline:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.btn-lg {
  padding: 14px 32px;
  font-size: 16px;
}

/* Features */
.features-section {
  padding: 64px 20px;
  border-top: 1px solid var(--border);
}

.section-title {
  text-align: center;
  font-size: 32px;
  font-weight: 600;
  margin: 0 0 40px;
  letter-spacing: -0.5px;
}

.features {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  max-width: 900px;
  margin: 0 auto;
}

.feature-card {
  padding: 24px;
  border: 1px solid var(--border);
  border-radius: 12px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.feature-card:hover {
  border-color: var(--accent-border);
  box-shadow: 0 4px 16px rgba(38, 68, 236, 0.08);
}

.feature-card.dimmed {
  opacity: 0.5;
  border-style: dashed;
}

.flag-off {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--code-bg);
  color: var(--text);
  font-weight: 400;
  vertical-align: middle;
}

.feature-icon {
  font-size: 28px;
  margin-bottom: 12px;
}

.feature-card h3 {
  font-size: 17px;
  font-weight: 600;
  margin: 0 0 8px;
  color: var(--text-h);
}

.feature-card p {
  font-size: 14px;
  color: var(--text);
  line-height: 1.5;
  margin: 0;
}

/* Tech stack */
.tech-section {
  padding: 64px 20px;
  border-top: 1px solid var(--border);
  background: var(--code-bg);
}

.tech-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  max-width: 700px;
  margin: 0 auto;
}

.tech-item {
  text-align: center;
  font-size: 14px;
  color: var(--text-h);
  font-weight: 500;
}

.tech-label {
  display: block;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--accent);
  margin-bottom: 4px;
  font-weight: 600;
}

/* CTA */
.cta {
  padding: 80px 20px;
  text-align: center;
  border-top: 1px solid var(--border);
}

.cta h2 {
  font-size: 32px;
  margin: 0 0 12px;
}

.cta p {
  font-size: 18px;
  color: var(--text);
  margin: 0 0 28px;
}

.cta-buttons {
  display: flex;
  gap: 12px;
  justify-content: center;
}

/* Feature List Section */
.feature-list-section {
  padding: 64px 20px;
  border-top: 1px solid var(--border);
}

.feature-list-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  margin-bottom: 40px;
}

.feature-list-stats {
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 12px 24px;
}

.stat-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.stat-number {
  font-size: 22px;
  font-weight: 700;
  color: var(--accent);
  line-height: 1;
}

.stat-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text);
  font-weight: 500;
}

.stat-divider {
  width: 1px;
  height: 32px;
  background: var(--border);
}

.categories-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  max-width: 960px;
  margin: 0 auto;
}

.category-block {
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
  transition: border-color 0.2s;
}

.category-block.category-open {
  border-color: var(--accent-border);
}

.category-header {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  color: var(--text-h);
  font-size: 14px;
  font-weight: 600;
  transition: background 0.15s;
}

.category-header:hover {
  background: var(--code-bg);
}

.category-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.category-title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
}

.category-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.category-count {
  font-size: 12px;
  color: var(--text);
  font-weight: 400;
  background: var(--code-bg);
  padding: 2px 7px;
  border-radius: 20px;
  border: 1px solid var(--border);
}

.category-chevron {
  font-size: 10px;
  color: var(--text);
  flex-shrink: 0;
}

.feature-items {
  list-style: none;
  margin: 0;
  padding: 0 16px 14px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 13px;
  line-height: 1.45;
}

.feat-mark {
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 700;
  width: 16px;
  text-align: center;
  margin-top: 1px;
}

.feat-done .feat-mark {
  color: #22c55e;
}

.feat-pending .feat-mark {
  color: var(--text);
  opacity: 0.4;
}

.feat-done .feat-text {
  color: var(--text-h);
}

.feat-pending .feat-text {
  color: var(--text);
  opacity: 0.55;
}

/* Responsive */
@media (max-width: 768px) {
  .hero h1 { font-size: 32px; }
  .features { grid-template-columns: 1fr; }
  .tech-grid { grid-template-columns: repeat(2, 1fr); }
  .hero-links, .cta-buttons { flex-direction: column; align-items: center; }
  .categories-grid { grid-template-columns: 1fr; }
  .feature-list-stats { gap: 12px; padding: 10px 16px; }
}
</style>
