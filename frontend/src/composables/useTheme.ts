import { ref } from 'vue'
import { api } from '../services/api'

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
  font_family: string | null
  heading_font: string | null
  mono_font: string | null
  base_font_size: string
  logo_url: string | null
  favicon_url: string | null
}

const brandKit = ref<BrandKit | null>(null)
const loaded = ref(false)

export function useTheme() {
  async function loadTheme() {
    if (loaded.value) return
    try {
      const data = await api.get<BrandKit>('/api/v1/brand-kit')
      brandKit.value = data
      applyTheme(data)
      loaded.value = true
    } catch {
      // Use CSS defaults
    }
  }

  function applyTheme(kit: BrandKit) {
    const root = document.documentElement

    // Light mode vars
    if (kit.accent_color) root.style.setProperty('--accent', kit.accent_color)
    if (kit.accent_bg) root.style.setProperty('--accent-bg', kit.accent_bg)
    if (kit.accent_border) root.style.setProperty('--accent-border', kit.accent_border)
    if (kit.text_color) root.style.setProperty('--text', kit.text_color)
    if (kit.text_heading_color) root.style.setProperty('--text-h', kit.text_heading_color)
    if (kit.bg_color) root.style.setProperty('--bg', kit.bg_color)
    if (kit.border_color) root.style.setProperty('--border', kit.border_color)
    if (kit.code_bg_color) root.style.setProperty('--code-bg', kit.code_bg_color)

    // Fonts
    if (kit.font_family) root.style.setProperty('--sans', kit.font_family)
    if (kit.heading_font) root.style.setProperty('--heading', kit.heading_font)
    if (kit.mono_font) root.style.setProperty('--mono', kit.mono_font)
    if (kit.base_font_size) root.style.setProperty('font-size', kit.base_font_size)

    // Favicon
    if (kit.favicon_url) {
      const link = document.querySelector("link[rel~='icon']") as HTMLLinkElement
      if (link) link.href = kit.favicon_url
    }

    // Dark mode — inject a style tag with @media override
    const darkCSS = `
      @media (prefers-color-scheme: dark) {
        :root {
          ${kit.dark_accent_color ? `--accent: ${kit.dark_accent_color};` : ''}
          ${kit.dark_accent_bg ? `--accent-bg: ${kit.dark_accent_bg};` : ''}
          ${kit.dark_accent_border ? `--accent-border: ${kit.dark_accent_border};` : ''}
          ${kit.dark_text_color ? `--text: ${kit.dark_text_color};` : ''}
          ${kit.dark_text_heading_color ? `--text-h: ${kit.dark_text_heading_color};` : ''}
          ${kit.dark_bg_color ? `--bg: ${kit.dark_bg_color};` : ''}
          ${kit.dark_border_color ? `--border: ${kit.dark_border_color};` : ''}
          ${kit.dark_code_bg_color ? `--code-bg: ${kit.dark_code_bg_color};` : ''}
        }
      }
    `
    let styleEl = document.getElementById('brand-kit-dark')
    if (!styleEl) {
      styleEl = document.createElement('style')
      styleEl.id = 'brand-kit-dark'
      document.head.appendChild(styleEl)
    }
    styleEl.textContent = darkCSS
  }

  return { brandKit, loadTheme, loaded, applyTheme }
}
