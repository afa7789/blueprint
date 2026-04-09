import { ref } from 'vue'

type ThemeMode = 'system' | 'light' | 'dark'

const themeMode = ref<ThemeMode>((localStorage.getItem('theme-mode') as ThemeMode) || 'system')
const fontSizeOffset = ref(parseInt(localStorage.getItem('font-size-offset') || '0'))
const highContrast = ref(localStorage.getItem('high-contrast') === 'true')

export function useAccessibility() {
  function setTheme(mode: ThemeMode) {
    themeMode.value = mode
    localStorage.setItem('theme-mode', mode)
    applyTheme()
  }

  function adjustFontSize(delta: number) {
    fontSizeOffset.value = Math.max(-4, Math.min(8, fontSizeOffset.value + delta))
    localStorage.setItem('font-size-offset', String(fontSizeOffset.value))
    applyFontSize()
  }

  function resetFontSize() {
    fontSizeOffset.value = 0
    localStorage.setItem('font-size-offset', '0')
    applyFontSize()
  }

  function toggleHighContrast() {
    highContrast.value = !highContrast.value
    localStorage.setItem('high-contrast', String(highContrast.value))
    applyHighContrast()
  }

  function applyTheme() {
    const root = document.documentElement
    root.removeAttribute('data-theme')
    if (themeMode.value === 'light') {
      root.setAttribute('data-theme', 'light')
    } else if (themeMode.value === 'dark') {
      root.setAttribute('data-theme', 'dark')
    }
    // 'system' = remove attribute, let @media handle it
  }

  function applyFontSize() {
    const base = 18 // default from style.css
    document.documentElement.style.fontSize = `${base + fontSizeOffset.value}px`
  }

  function applyHighContrast() {
    document.documentElement.classList.toggle('high-contrast', highContrast.value)
  }

  // Apply on load
  function init() {
    applyTheme()
    applyFontSize()
    applyHighContrast()
  }

  return { themeMode, fontSizeOffset, highContrast, setTheme, adjustFontSize, resetFontSize, toggleHighContrast, init }
}
