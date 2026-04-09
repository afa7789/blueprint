<script setup lang="ts">
import WaitlistForm from '../components/landing/WaitlistForm.vue'
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
      <WaitlistForm />
      <div class="hero-links">
        <a href="https://github.com/afa7789/blueprint" target="_blank" rel="noopener noreferrer" class="btn btn-outline">
          GitHub
        </a>
        <router-link to="/register" class="btn btn-primary">Get Started</router-link>
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
        <div class="feature-card">
          <div class="feature-icon">🛒</div>
          <h3>E-Commerce</h3>
          <p>Products, categories, cart, orders, coupons, pre-sale, Stripe + PIX payments.</p>
        </div>
        <div class="feature-card">
          <div class="feature-icon">📝</div>
          <h3>Blog + AI</h3>
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
        <router-link to="/store" class="btn btn-outline btn-lg">Browse Store</router-link>
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

/* Responsive */
@media (max-width: 768px) {
  .hero h1 { font-size: 32px; }
  .features { grid-template-columns: 1fr; }
  .tech-grid { grid-template-columns: repeat(2, 1fr); }
  .hero-links, .cta-buttons { flex-direction: column; align-items: center; }
}
</style>
