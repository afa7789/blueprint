<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { api } from '../../services/api'
import { loadSiteModules, siteModules } from '../../services/siteModules'
import { useTheme } from '../../composables/useTheme'

const { brandKit, loadTheme } = useTheme()

const linktreeEnabled = computed(() => siteModules.linktreeEnabled)
const storeVisible = computed(() => siteModules.storeEnabled && siteModules.hasStoreContent)
const blogVisible = computed(() => siteModules.blogEnabled && siteModules.hasBlogContent)
const brandKitEnabled = computed(() => siteModules.brandKitEnabled)

interface LegalLink {
  slug: string
  title: string
}
const legalPages = ref<LegalLink[]>([])

onMounted(async () => {
  await Promise.all([loadSiteModules(), loadTheme()])

  try {
    legalPages.value = await api.get<LegalLink[]>('/api/v1/legal')
  } catch {
    // silent — no legal pages configured
  }
})
</script>

<template>
  <footer class="footer">
    <div class="footer-content">
      <div class="footer-links">
        <div v-if="storeVisible" class="footer-section">
          <h3>Store</h3>
          <router-link to="/store">Browse Products</router-link>
        </div>
        <div v-if="blogVisible" class="footer-section">
          <h3>Blog</h3>
          <router-link to="/blog">Read Articles</router-link>
        </div>
        <div v-if="linktreeEnabled" class="footer-section">
          <h3>Linktree</h3>
          <router-link to="/linktree">View Links</router-link>
        </div>
        <div class="footer-section">
          <h3>Resources</h3>
          <router-link v-if="brandKitEnabled" to="/brand-kit">Brand Kit</router-link>
          <router-link to="/legal/terms">Terms</router-link>
        </div>
        <div v-if="brandKitEnabled && brandKit" class="footer-section footer-section-brand">
          <h3>Theme</h3>
          <div v-if="brandKit.logo_url" class="brand-logo-wrap">
            <img :src="brandKit.logo_url" alt="Brand logo" class="brand-logo" />
          </div>
          <div class="brand-swatches">
            <span class="brand-swatch" :style="{ background: brandKit.accent_color }" title="Accent"></span>
            <span class="brand-swatch" :style="{ background: brandKit.text_color }" title="Text"></span>
            <span class="brand-swatch" :style="{ background: brandKit.bg_color }" title="Background"></span>
          </div>
          <p class="brand-meta" v-if="brandKit.font_family">Body: {{ brandKit.font_family }}</p>
          <p class="brand-meta" v-if="brandKit.heading_font">Heading: {{ brandKit.heading_font }}</p>
        </div>
      </div>
      <div class="footer-bottom">
        <div v-if="legalPages.length" class="legal-links">
          <router-link v-for="page in legalPages" :key="page.slug" :to="`/legal/${page.slug}`">
            {{ page.title }}
          </router-link>
        </div>
        <p>&copy; 2026. All rights reserved.</p>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.footer {
  border-top: 1px solid var(--border);
  padding: 32px 20px;
  text-align: left;
}

.footer-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.footer-links {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
}

.footer-section h3 {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--text-h);
}

.footer-section a {
  color: var(--text);
  text-decoration: none;
  font-size: 14px;
  transition: color 0.2s;
}

.footer-section a:hover {
  color: var(--accent);
}

.footer-section-brand {
  min-width: 220px;
}

.brand-logo-wrap {
  margin-bottom: 8px;
}

.brand-logo {
  max-width: 140px;
  max-height: 44px;
  object-fit: contain;
}

.brand-swatches {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.brand-swatch {
  width: 18px;
  height: 18px;
  border-radius: 999px;
  border: 1px solid var(--border);
}

.brand-meta {
  margin: 0 0 4px;
  font-size: 13px;
  color: var(--text);
}

.footer-bottom {
  font-size: 14px;
  color: var(--text);
}

.legal-links {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.legal-links a {
  color: var(--text);
  text-decoration: none;
  font-size: 13px;
  transition: color 0.2s;
}

.legal-links a:hover {
  color: var(--accent);
}
</style>
