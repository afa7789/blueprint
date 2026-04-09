<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../services/api'

interface BrandKitData {
  accent_color: string
  text_color: string
  text_heading_color: string
  bg_color: string
  border_color: string
  logo_url: string | null
  favicon_url: string | null
  font_family: string | null
  heading_font: string | null
  mono_font: string | null
  primary_color: string
  secondary_color: string
  code_bg_color: string
  dark_accent_color: string
  dark_text_color: string
  dark_text_heading_color: string
  dark_bg_color: string
  dark_border_color: string
}

const brand = ref<BrandKitData | null>(null)
const loading = ref(true)
const copied = ref<string | null>(null)

onMounted(async () => {
  try {
    brand.value = await api.get<BrandKitData>('/api/v1/brand-kit')
  } catch { /* ignore */ }
  loading.value = false
})

function copyColor(color: string) {
  navigator.clipboard.writeText(color)
  copied.value = color
  setTimeout(() => { copied.value = null }, 1500)
}

function downloadAsset(url: string, filename: string) {
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
}
</script>

<template>
  <div class="brandkit-page">
    <!-- Hero -->
    <section class="bk-hero">
      <h1>Brand Kit</h1>
      <p class="bk-subtitle">Download our brand assets, colors, and guidelines for use in your projects.</p>
    </section>

    <div v-if="loading" class="bk-loading">Loading brand assets...</div>

    <template v-if="!loading && brand">

      <!-- Logos -->
      <section class="bk-section">
        <h2 class="bk-section-title">Logos &amp; Icons</h2>
        <div class="bk-asset-grid">
          <div v-if="brand.logo_url" class="bk-asset-card">
            <div class="bk-asset-preview bk-asset-preview--light">
              <img :src="brand.logo_url" alt="Logo" class="bk-logo-img" />
            </div>
            <div class="bk-asset-info">
              <span class="bk-asset-name">Logo</span>
              <button class="bk-btn" @click="downloadAsset(brand.logo_url!, 'logo')">Download Logo</button>
            </div>
          </div>

          <div v-if="brand.favicon_url" class="bk-asset-card">
            <div class="bk-asset-preview bk-asset-preview--light">
              <img :src="brand.favicon_url" alt="Favicon" class="bk-favicon-img" />
            </div>
            <div class="bk-asset-info">
              <span class="bk-asset-name">Favicon</span>
              <button class="bk-btn" @click="downloadAsset(brand.favicon_url!, 'favicon')">Download Favicon</button>
            </div>
          </div>

          <div class="bk-asset-card">
            <div class="bk-asset-preview bk-asset-preview--light">
              <img src="/icon.svg" alt="App Icon" class="bk-icon-img" />
            </div>
            <div class="bk-asset-info">
              <span class="bk-asset-name">App Icon (Favicon)</span>
              <span class="bk-asset-desc">Flat icon with color background. Used as favicon and app icon.</span>
              <div class="bk-asset-actions">
                <button class="bk-btn" @click="downloadAsset('/icon.svg', 'blueprint-icon.svg')">SVG</button>
                <button class="bk-btn" @click="downloadAsset('/icon.png', 'blueprint-icon.png')">PNG</button>
              </div>
            </div>
          </div>

          <div class="bk-asset-card">
            <div class="bk-asset-preview bk-asset-preview--dark">
              <img src="/inverted-icon.svg" alt="Inverted Icon" class="bk-icon-img" />
            </div>
            <div class="bk-asset-info">
              <span class="bk-asset-name">Inverted Icon</span>
              <span class="bk-asset-desc">White version for dark backgrounds.</span>
              <div class="bk-asset-actions">
                <button class="bk-btn" @click="downloadAsset('/inverted-icon.svg', 'blueprint-inverted.svg')">SVG</button>
                <button class="bk-btn" @click="downloadAsset('/inverted-icon.png', 'blueprint-inverted.png')">PNG</button>
              </div>
            </div>
          </div>

          <div class="bk-asset-card">
            <div class="bk-asset-preview bk-asset-preview--light">
              <img src="/banner-thin.svg" alt="Banner" class="bk-banner-img" />
            </div>
            <div class="bk-asset-info">
              <span class="bk-asset-name">Banner</span>
              <span class="bk-asset-desc">Full wordmark banner for headers and README.</span>
              <div class="bk-asset-actions">
                <button class="bk-btn" @click="downloadAsset('/banner-thin.svg', 'blueprint-banner.svg')">SVG</button>
                <button class="bk-btn" @click="downloadAsset('/banner-thin.png', 'blueprint-banner.png')">PNG</button>
              </div>
            </div>
          </div>

          <div class="bk-asset-card">
            <div class="bk-asset-preview bk-asset-preview--light">
              <img src="/icon.svg" alt="Favicon" class="bk-icon-img" />
            </div>
            <div class="bk-asset-info">
              <span class="bk-asset-name">Favicon</span>
              <span class="bk-asset-desc">Flat icon with color background — used as browser favicon.</span>
              <div class="bk-asset-actions">
                <button class="bk-btn" @click="downloadAsset('/icon.svg', 'blueprint-favicon.svg')">SVG</button>
                <button class="bk-btn" @click="downloadAsset('/icon.png', 'blueprint-favicon.png')">PNG</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Color Palette -->
      <section class="bk-section">
        <h2 class="bk-section-title">Color Palette</h2>

        <div class="bk-palette-columns">
          <!-- Light Mode -->
          <div class="bk-palette-group">
            <h3 class="bk-palette-mode">Light Mode</h3>
            <div class="bk-swatch-grid">
              <div class="bk-swatch-card" v-for="swatch in [
                { label: 'Primary', value: brand.primary_color },
                { label: 'Secondary', value: brand.secondary_color },
                { label: 'Accent', value: brand.accent_color },
                { label: 'Text', value: brand.text_color },
                { label: 'Heading', value: brand.text_heading_color },
                { label: 'Background', value: brand.bg_color },
                { label: 'Border', value: brand.border_color },
                { label: 'Code Bg', value: brand.code_bg_color },
              ]" :key="swatch.label">
                <div class="bk-swatch-box" :style="{ background: swatch.value, border: '1px solid rgba(0,0,0,0.08)' }"></div>
                <div class="bk-swatch-meta">
                  <span class="bk-swatch-label">{{ swatch.label }}</span>
                  <span class="bk-swatch-hex">{{ swatch.value }}</span>
                  <button
                    class="bk-copy-btn"
                    :class="{ copied: copied === swatch.value }"
                    @click="copyColor(swatch.value)"
                  >{{ copied === swatch.value ? 'Copied!' : 'Copy' }}</button>
                </div>
              </div>
            </div>
          </div>

          <!-- Dark Mode -->
          <div class="bk-palette-group bk-palette-group--dark">
            <h3 class="bk-palette-mode">Dark Mode</h3>
            <div class="bk-swatch-grid">
              <div class="bk-swatch-card" v-for="swatch in [
                { label: 'Accent', value: brand.dark_accent_color },
                { label: 'Text', value: brand.dark_text_color },
                { label: 'Heading', value: brand.dark_text_heading_color },
                { label: 'Background', value: brand.dark_bg_color },
                { label: 'Border', value: brand.dark_border_color },
              ]" :key="swatch.label">
                <div class="bk-swatch-box" :style="{ background: swatch.value, border: '1px solid rgba(255,255,255,0.1)' }"></div>
                <div class="bk-swatch-meta">
                  <span class="bk-swatch-label">{{ swatch.label }}</span>
                  <span class="bk-swatch-hex">{{ swatch.value }}</span>
                  <button
                    class="bk-copy-btn"
                    :class="{ copied: copied === swatch.value }"
                    @click="copyColor(swatch.value)"
                  >{{ copied === swatch.value ? 'Copied!' : 'Copy' }}</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Typography -->
      <section class="bk-section">
        <h2 class="bk-section-title">Typography</h2>
        <div class="bk-type-grid">
          <div class="bk-type-card">
            <div class="bk-type-label">Body Font</div>
            <div class="bk-type-family">{{ brand.font_family || 'system-ui, Segoe UI, Roboto, sans-serif' }}</div>
            <div class="bk-type-preview" :style="{ fontFamily: brand.font_family || 'var(--sans)' }">
              Aa Bb Cc Dd Ee Ff Gg Hh Ii Jj Kk Ll Mm Nn Oo Pp Qq Rr Ss Tt Uu Vv Ww Xx Yy Zz
            </div>
            <div class="bk-type-preview bk-type-preview--sm" :style="{ fontFamily: brand.font_family || 'var(--sans)' }">
              The quick brown fox jumps over the lazy dog. 0123456789
            </div>
          </div>
          <div class="bk-type-card">
            <div class="bk-type-label">Heading Font</div>
            <div class="bk-type-family">{{ brand.heading_font || 'system-ui, Segoe UI, Roboto, sans-serif' }}</div>
            <div class="bk-type-preview bk-type-preview--heading" :style="{ fontFamily: brand.heading_font || 'var(--heading)' }">
              Heading Preview Text
            </div>
          </div>
          <div class="bk-type-card">
            <div class="bk-type-label">Monospace Font</div>
            <div class="bk-type-family">{{ brand.mono_font || 'ui-monospace, Consolas, monospace' }}</div>
            <div class="bk-type-preview bk-type-preview--mono" :style="{ fontFamily: brand.mono_font || 'var(--mono)' }">
              const blueprint = { version: '1.0', ready: true };
            </div>
          </div>
        </div>
      </section>

    </template>

    <!-- Usage Guidelines (always visible) -->
    <section class="bk-section">
      <h2 class="bk-section-title">Usage Guidelines</h2>
      <ul class="bk-guidelines">
        <li>Use the logo on light backgrounds with sufficient contrast.</li>
        <li>Maintain minimum clear space around the logo equal to the height of the logo mark.</li>
        <li>Do not distort, recolor, or add effects to the logo.</li>
        <li>Use the provided color palette for consistency across all materials.</li>
        <li>Do not use the brand assets in a way that implies partnership or endorsement without permission.</li>
      </ul>
    </section>
  </div>
</template>

<style scoped>
.brandkit-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 48px 24px 80px;
}

.bk-hero {
  margin-bottom: 56px;
  text-align: left;
}

.bk-hero h1 {
  font-size: 40px;
  font-weight: 700;
  color: var(--text-h);
  margin: 0 0 12px;
  letter-spacing: -0.5px;
}

.bk-subtitle {
  font-size: 18px;
  color: var(--text);
  margin: 0;
  max-width: 560px;
  line-height: 1.6;
}

.bk-loading {
  padding: 48px 0;
  color: var(--text);
  font-size: 16px;
}

.bk-section {
  margin-bottom: 64px;
}

.bk-section-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-h);
  margin: 0 0 24px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border);
  letter-spacing: -0.3px;
}

/* Assets */
.bk-asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
}

.bk-asset-card {
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  background: var(--bg);
}

.bk-asset-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 140px;
  padding: 24px;
}

.bk-asset-preview--light {
  background: #f9fafb;
}

.bk-asset-preview--dark {
  background: #111827;
}

.bk-logo-img {
  max-height: 80px;
  max-width: 100%;
  object-fit: contain;
}

.bk-favicon-img {
  width: 64px;
  height: 64px;
  object-fit: contain;
}

.bk-icon-img {
  max-height: 80px;
  max-width: 100%;
  object-fit: contain;
}

.bk-banner-img {
  max-height: 60px;
  max-width: 100%;
  object-fit: contain;
}

.bk-asset-info {
  padding: 14px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid var(--border);
}

.bk-asset-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-h);
}

.bk-asset-desc {
  font-size: 12px;
  color: var(--text);
  line-height: 1.4;
}

.bk-asset-actions {
  display: flex;
  gap: 6px;
  margin-top: 4px;
}

.bk-btn {
  font-size: 12px;
  padding: 5px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.bk-btn:hover {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
}

/* Palette */
.bk-palette-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
}

@media (max-width: 640px) {
  .bk-palette-columns {
    grid-template-columns: 1fr;
  }
}

.bk-palette-group {
  padding: 24px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--bg);
}

.bk-palette-group--dark {
  background: #111827;
  border-color: #374151;
}

.bk-palette-mode {
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text);
  margin: 0 0 16px;
}

.bk-palette-group--dark .bk-palette-mode {
  color: #9ca3af;
}

.bk-swatch-grid {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.bk-swatch-card {
  display: flex;
  align-items: center;
  gap: 12px;
}

.bk-swatch-box {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  flex-shrink: 0;
}

.bk-swatch-meta {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.bk-swatch-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-h);
  min-width: 70px;
}

.bk-palette-group--dark .bk-swatch-label {
  color: #f9fafb;
}

.bk-swatch-hex {
  font-size: 12px;
  color: var(--text);
  font-family: monospace;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bk-palette-group--dark .bk-swatch-hex {
  color: #9ca3af;
}

.bk-copy-btn {
  font-size: 11px;
  padding: 3px 8px;
  border: 1px solid var(--border);
  border-radius: 4px;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.15s, color 0.15s;
}

.bk-palette-group--dark .bk-copy-btn {
  border-color: #374151;
  color: #9ca3af;
}

.bk-copy-btn:hover,
.bk-copy-btn.copied {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
}

/* Typography */
.bk-type-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.bk-type-card {
  padding: 24px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--bg);
}

.bk-type-label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text);
  opacity: 0.6;
  margin-bottom: 4px;
}

.bk-type-family {
  font-size: 13px;
  color: var(--text);
  margin-bottom: 12px;
  font-family: monospace;
}

.bk-type-preview {
  font-size: 20px;
  color: var(--text-h);
  line-height: 1.4;
}

.bk-type-preview--heading {
  font-size: 28px;
  font-weight: 700;
}

.bk-type-preview--sm {
  font-size: 15px;
  color: var(--text);
  margin-top: 8px;
}

.bk-type-preview--mono {
  font-size: 16px;
}

/* Guidelines */
.bk-guidelines {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.bk-guidelines li {
  font-size: 15px;
  color: var(--text);
  line-height: 1.6;
  padding-left: 20px;
  position: relative;
}

.bk-guidelines li::before {
  content: "—";
  position: absolute;
  left: 0;
  color: var(--accent);
  font-weight: 600;
}
</style>
