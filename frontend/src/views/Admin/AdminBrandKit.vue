<template>
  <div>
    <HelperBox
      title="Brand Kit"
      description="Customize your brand identity. Colors, logo, favicon, and fonts are applied globally across the frontend."
      featureFlag="brand_kit_enabled"
    />
    <h2>Brand Kit</h2>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="success" class="success">Saved successfully.</div>
    <div v-if="loading">Loading...</div>

    <!-- Active Theme Summary Card -->
    <div v-if="!loading" class="theme-summary-card">
      <div class="theme-summary-top">
        <div v-if="form.logo_url" class="theme-logo">
          <img :src="form.logo_url" alt="Logo" class="theme-logo-img" />
        </div>
        <div class="theme-summary-info">
          <div class="theme-swatches">
            <div class="swatch-item" v-for="swatch in colorSwatches" :key="swatch.label">
              <div class="swatch-circle" :style="{ background: swatch.value }"></div>
              <span class="swatch-label">{{ swatch.label }}</span>
            </div>
          </div>
          <div class="theme-font-previews">
            <div class="font-preview-item" :style="{ fontFamily: form.font_family || 'inherit' }">
              <span class="font-tag">Body</span> The quick brown fox jumps over the lazy dog
            </div>
            <div class="font-preview-item font-preview-heading" :style="{ fontFamily: form.heading_font || 'inherit' }">
              <span class="font-tag">Heading</span> The quick brown fox jumps over the lazy dog
            </div>
            <div class="font-preview-item font-preview-mono" :style="{ fontFamily: form.mono_font || 'monospace' }">
              <span class="font-tag">Mono</span> The quick brown fox jumps over the lazy dog
            </div>
          </div>
          <div class="theme-meta">
            <span class="theme-meta-item">Base size: <strong>{{ form.base_font_size || '16px' }}</strong></span>
          </div>
        </div>
      </div>
    </div>

    <form v-if="!loading" class="brand-form" @submit.prevent="save">

      <!-- Section 1: Brand -->
      <section class="form-section">
        <h3 class="section-title">Brand</h3>
        <label>Logo URL
          <input v-model="form.logo_url" type="text" placeholder="https://example.com/logo.png" />
          <div v-if="form.logo_url" class="url-preview">
            <img :src="form.logo_url" alt="Logo preview" class="logo-preview-img" @error="onLogoError" />
          </div>
        </label>
        <label>Favicon URL
          <input v-model="form.favicon_url" type="text" placeholder="https://example.com/favicon.ico" />
          <div v-if="form.favicon_url" class="url-preview">
            <img :src="form.favicon_url" alt="Favicon preview" class="favicon-preview-img" @error="onFaviconError" />
          </div>
        </label>
      </section>

      <!-- Section 2: Light Mode Colors -->
      <section class="form-section">
        <h3 class="section-title">Light Mode Colors</h3>
        <label>Accent Color
          <div class="color-row">
            <input v-model="form.accent_color" type="color" />
            <input v-model="form.accent_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Accent Background <span class="hint">(rgba)</span>
          <input v-model="form.accent_bg" type="text" placeholder="rgba(38, 68, 236, 0.1)" />
        </label>
        <label>Accent Border <span class="hint">(rgba)</span>
          <input v-model="form.accent_border" type="text" placeholder="rgba(38, 68, 236, 0.3)" />
        </label>
        <label>Text Color
          <div class="color-row">
            <input v-model="form.text_color" type="color" />
            <input v-model="form.text_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Heading Color
          <div class="color-row">
            <input v-model="form.text_heading_color" type="color" />
            <input v-model="form.text_heading_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Background Color
          <div class="color-row">
            <input v-model="form.bg_color" type="color" />
            <input v-model="form.bg_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Border Color
          <div class="color-row">
            <input v-model="form.border_color" type="color" />
            <input v-model="form.border_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Code Background
          <div class="color-row">
            <input v-model="form.code_bg_color" type="color" />
            <input v-model="form.code_bg_color" type="text" class="color-text" />
          </div>
        </label>
      </section>

      <!-- Section 3: Dark Mode Colors -->
      <section class="form-section dark-section">
        <h3 class="section-title">Dark Mode Colors</h3>
        <label>Dark Accent Color
          <div class="color-row">
            <input v-model="form.dark_accent_color" type="color" />
            <input v-model="form.dark_accent_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Dark Accent Background <span class="hint">(rgba)</span>
          <input v-model="form.dark_accent_bg" type="text" placeholder="rgba(38, 68, 236, 0.15)" />
        </label>
        <label>Dark Accent Border <span class="hint">(rgba)</span>
          <input v-model="form.dark_accent_border" type="text" placeholder="rgba(38, 68, 236, 0.4)" />
        </label>
        <label>Dark Text Color
          <div class="color-row">
            <input v-model="form.dark_text_color" type="color" />
            <input v-model="form.dark_text_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Dark Heading Color
          <div class="color-row">
            <input v-model="form.dark_text_heading_color" type="color" />
            <input v-model="form.dark_text_heading_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Dark Background Color
          <div class="color-row">
            <input v-model="form.dark_bg_color" type="color" />
            <input v-model="form.dark_bg_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Dark Border Color
          <div class="color-row">
            <input v-model="form.dark_border_color" type="color" />
            <input v-model="form.dark_border_color" type="text" class="color-text" />
          </div>
        </label>
        <label>Dark Code Background
          <div class="color-row">
            <input v-model="form.dark_code_bg_color" type="color" />
            <input v-model="form.dark_code_bg_color" type="text" class="color-text" />
          </div>
        </label>
      </section>

      <!-- Section 4: Typography -->
      <section class="form-section">
        <h3 class="section-title">Typography</h3>
        <label>Font Family
          <input v-model="form.font_family" type="text" placeholder="Inter, sans-serif" />
          <div v-if="form.font_family" class="inline-font-preview" :style="{ fontFamily: form.font_family }">The quick brown fox jumps over the lazy dog</div>
        </label>
        <label>Heading Font
          <input v-model="form.heading_font" type="text" placeholder="Inter, sans-serif" />
          <div v-if="form.heading_font" class="inline-font-preview inline-font-preview--heading" :style="{ fontFamily: form.heading_font }">The quick brown fox jumps over the lazy dog</div>
        </label>
        <label>Monospace Font
          <input v-model="form.mono_font" type="text" placeholder="JetBrains Mono, monospace" />
          <div v-if="form.mono_font" class="inline-font-preview inline-font-preview--mono" :style="{ fontFamily: form.mono_font }">The quick brown fox jumps over the lazy dog</div>
        </label>
        <label>Base Font Size
          <select v-model="form.base_font_size">
            <option value="14px">14px</option>
            <option value="16px">16px</option>
            <option value="18px">18px</option>
            <option value="20px">20px</option>
          </select>
        </label>
      </section>

      <!-- Section 5: Preview -->
      <section class="form-section">
        <h3 class="section-title">Preview</h3>
        <div class="preview" :style="previewStyles">
          <div class="preview-header">
            <div class="preview-brand">
              <img v-if="form.logo_url" :src="form.logo_url" alt="Logo" class="preview-logo" />
              <img v-if="form.favicon_url" :src="form.favicon_url" alt="Favicon" class="preview-favicon" />
            </div>
            <h3 :style="{ fontFamily: form.heading_font || 'inherit' }">Heading Preview</h3>
          </div>
          <p :style="{ fontFamily: form.font_family || 'inherit' }">
            Body text looks like this. Here is some <code :style="{ fontFamily: form.mono_font || 'monospace' }">inline code</code> too.
          </p>
          <div class="preview-row">
            <button type="button" class="preview-btn">Accent Button</button>
            <span class="preview-badge">Badge</span>
            <span class="preview-link">Link text</span>
          </div>
          <div class="preview-colors">
            <div class="preview-swatch" :style="{ background: form.accent_color }" title="accent"></div>
            <div class="preview-swatch" :style="{ background: form.text_color }" title="text"></div>
            <div class="preview-swatch" :style="{ background: form.bg_color, border: '1px solid ' + form.border_color }" title="bg"></div>
            <div class="preview-swatch" :style="{ background: form.border_color }" title="border"></div>
            <div class="preview-swatch" :style="{ background: form.code_bg_color }" title="code-bg"></div>
          </div>
        </div>
      </section>

      <button type="submit" class="btn-primary">Save Brand Kit</button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'
import { useTheme } from '../../composables/useTheme'

interface BrandKit {
  accent_color: string
  accent_bg: string
  accent_border: string
  text_color: string
  text_heading_color: string
  bg_color: string
  border_color: string
  code_bg_color: string
  dark_accent_color: string
  dark_accent_bg: string
  dark_accent_border: string
  dark_text_color: string
  dark_text_heading_color: string
  dark_bg_color: string
  dark_border_color: string
  dark_code_bg_color: string
  font_family: string
  heading_font: string
  mono_font: string
  base_font_size: string
  logo_url: string
  favicon_url: string
  primary_color: string
  secondary_color: string
}

const { applyTheme } = useTheme()

const form = reactive<BrandKit>({
  accent_color: '#2644ec',
  accent_bg: '',
  accent_border: '',
  text_color: '#374151',
  text_heading_color: '#111827',
  bg_color: '#ffffff',
  border_color: '#e5e7eb',
  code_bg_color: '#f3f4f6',
  dark_accent_color: '#2644ec',
  dark_accent_bg: '',
  dark_accent_border: '',
  dark_text_color: '#d1d5db',
  dark_text_heading_color: '#f9fafb',
  dark_bg_color: '#111827',
  dark_border_color: '#374151',
  dark_code_bg_color: '#1f2937',
  font_family: '',
  heading_font: '',
  mono_font: '',
  base_font_size: '16px',
  logo_url: '',
  favicon_url: '',
  primary_color: '#2644ec',
  secondary_color: '#08060d',
})

const loading = ref(false)
const error = ref('')
const success = ref(false)

const colorSwatches = computed(() => [
  { label: 'Accent', value: form.accent_color || '#2644ec' },
  { label: 'Text', value: form.text_color || '#374151' },
  { label: 'Bg', value: form.bg_color || '#ffffff' },
  { label: 'Border', value: form.border_color || '#e5e7eb' },
  { label: 'Code Bg', value: form.code_bg_color || '#f3f4f6' },
])

function onLogoError(e: Event) {
  (e.target as HTMLImageElement).style.display = 'none'
}

function onFaviconError(e: Event) {
  (e.target as HTMLImageElement).style.display = 'none'
}

const previewStyles = computed(() => ({
  background: form.bg_color || 'var(--bg)',
  color: form.text_color || 'var(--text)',
  borderColor: form.border_color || 'var(--border)',
  '--preview-accent': form.accent_color || 'var(--accent)',
  '--preview-code-bg': form.code_bg_color || 'var(--code-bg)',
  '--preview-heading': form.text_heading_color || 'var(--text-h)',
}))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<BrandKit>('/api/v1/admin/brand-kit')
    Object.assign(form, data)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load brand kit'
  } finally {
    loading.value = false
  }
}

async function save() {
  error.value = ''
  success.value = false
  try {
    await api.put('/api/v1/admin/brand-kit', form)
    success.value = true
    applyTheme(form as Parameters<typeof applyTheme>[0])
    setTimeout(() => { success.value = false }, 3000)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save brand kit'
  }
}

onMounted(load)
</script>

<style scoped>
.brand-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
  max-width: 560px;
  margin-top: 8px;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg);
}

.dark-section {
  background: #1a1a2e;
  border-color: #374151;
}

.dark-section label {
  color: #d1d5db;
}

.dark-section input[type="text"] {
  background: #111827;
  color: #f9fafb;
  border-color: #374151;
}

.section-title {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-h);
  letter-spacing: -0.2px;
}

.dark-section .section-title {
  color: #f9fafb;
}

.brand-form label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  color: var(--text);
}

.hint {
  font-size: 11px;
  opacity: 0.6;
  font-weight: normal;
}

.brand-form input[type="text"],
.brand-form select {
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text-h);
  padding: 8px 10px;
  border-radius: 4px;
  font-size: 14px;
}

.brand-form select {
  cursor: pointer;
}

.color-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.color-row input[type="color"] {
  width: 40px;
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 2px;
  cursor: pointer;
  background: var(--bg);
  flex-shrink: 0;
}

.color-text {
  flex: 1;
}

.preview {
  padding: 24px;
  border: 1px solid;
  border-radius: 8px;
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.preview-brand {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-logo {
  height: 28px;
  width: auto;
  max-width: 120px;
  object-fit: contain;
}

.preview-favicon {
  width: 20px;
  height: 20px;
  object-fit: contain;
  border-radius: 2px;
}

.preview h3 {
  margin: 0;
  font-size: 18px;
}

.preview p {
  font-size: 14px;
  margin: 0 0 12px;
  line-height: 1.5;
}

.preview code {
  font-size: 13px;
  padding: 2px 6px;
  border-radius: 4px;
}

.preview-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.preview-btn {
  padding: 6px 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}

.preview-badge {
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
}

.preview-link {
  font-size: 13px;
  text-decoration: underline;
  cursor: pointer;
}

.preview-colors {
  display: flex;
  gap: 6px;
}

.preview-swatch {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 1px solid rgba(0,0,0,0.1);
}

.preview p {
  margin: 0 0 12px;
  font-size: 14px;
}

.preview code {
  background: var(--preview-code-bg, var(--code-bg));
  padding: 2px 5px;
  border-radius: 3px;
  font-size: 13px;
}

.preview button {
  background: var(--preview-accent, var(--accent));
  color: #fff;
  border: none;
  padding: 8px 16px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 13px;
}

.btn-primary {
  background: var(--accent);
  color: #fff;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  align-self: flex-start;
}

.error {
  color: #e53e3e;
  font-size: 14px;
  margin-bottom: 8px;
}

.success {
  color: #48bb78;
  font-size: 14px;
  margin-bottom: 8px;
}

/* Theme Summary Card */
.theme-summary-card {
  max-width: 560px;
  margin: 8px 0 20px;
  padding: 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg);
}

.theme-summary-top {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.theme-logo {
  flex-shrink: 0;
}

.theme-logo-img {
  max-height: 48px;
  max-width: 120px;
  object-fit: contain;
  border-radius: 4px;
}

.theme-summary-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.theme-swatches {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.swatch-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.swatch-circle {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 1px solid var(--border);
  flex-shrink: 0;
}

.swatch-label {
  font-size: 10px;
  color: var(--text);
  opacity: 0.7;
  white-space: nowrap;
}

.theme-font-previews {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.font-preview-item {
  font-size: 13px;
  color: var(--text);
  display: flex;
  align-items: baseline;
  gap: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.font-preview-item.font-preview-heading {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-h);
}

.font-preview-item.font-preview-mono {
  font-size: 12px;
  opacity: 0.85;
}

.font-tag {
  font-size: 10px;
  font-family: inherit;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  opacity: 0.5;
  flex-shrink: 0;
  min-width: 48px;
}

.theme-meta {
  display: flex;
  gap: 12px;
}

.theme-meta-item {
  font-size: 12px;
  color: var(--text);
  opacity: 0.7;
}

/* URL image previews */
.url-preview {
  margin-top: 4px;
}

.logo-preview-img {
  max-height: 48px;
  max-width: 200px;
  object-fit: contain;
  border-radius: 4px;
  border: 1px solid var(--border);
  padding: 4px;
  background: var(--bg);
}

.favicon-preview-img {
  width: 32px;
  height: 32px;
  object-fit: contain;
  border-radius: 4px;
  border: 1px solid var(--border);
  padding: 2px;
  background: var(--bg);
}

/* Inline font previews in Typography section */
.inline-font-preview {
  font-size: 13px;
  color: var(--text);
  padding: 6px 8px;
  background: var(--code-bg, #f3f4f6);
  border-radius: 4px;
  border: 1px solid var(--border);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.inline-font-preview--heading {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-h);
}

.inline-font-preview--mono {
  font-size: 12px;
}
</style>
