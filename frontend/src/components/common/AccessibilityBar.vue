<script setup lang="ts">
import { ref } from 'vue'
import { useAccessibility } from '../../composables/useAccessibility'

const expanded = ref(false)
const { themeMode, fontSizeOffset, highContrast, setTheme, adjustFontSize, resetFontSize, toggleHighContrast } = useAccessibility()
</script>

<template>
  <div class="a11y-widget" :class="{ expanded }">
    <button class="a11y-toggle" @click="expanded = !expanded" title="Accessibility">
      <i class="fas fa-universal-access"></i>
    </button>
    <div v-if="expanded" class="a11y-panel">
      <div class="a11y-section">
        <span class="a11y-label">Theme</span>
        <div class="a11y-buttons">
          <button :class="{ active: themeMode === 'light' }" @click="setTheme('light')" title="Light">☀️</button>
          <button :class="{ active: themeMode === 'dark' }" @click="setTheme('dark')" title="Dark">🌙</button>
          <button :class="{ active: themeMode === 'system' }" @click="setTheme('system')" title="System">💻</button>
        </div>
      </div>
      <div class="a11y-section">
        <span class="a11y-label">Font Size</span>
        <div class="a11y-buttons">
          <button @click="adjustFontSize(-2)">A-</button>
          <button @click="resetFontSize" class="size-indicator">{{ fontSizeOffset >= 0 ? '+' : '' }}{{ fontSizeOffset }}</button>
          <button @click="adjustFontSize(2)">A+</button>
        </div>
      </div>
      <div class="a11y-section">
        <span class="a11y-label">Contrast</span>
        <button :class="{ active: highContrast }" @click="toggleHighContrast">
          <i class="fas fa-circle-half-stroke"></i> {{ highContrast ? 'On' : 'Off' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.a11y-widget {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.a11y-toggle {
  width: 40px;
  height: 40px;
  border-radius: 20px;
  border: 1px solid var(--border);
  background: var(--bg);
  color: var(--text);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  box-shadow: var(--shadow, 0 2px 8px rgba(0,0,0,0.15));
  transition: color 0.2s, border-color 0.2s;
}

.a11y-toggle:hover {
  color: var(--text-h);
  border-color: var(--accent);
}

.a11y-panel {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 160px;
  box-shadow: var(--shadow, 0 4px 16px rgba(0,0,0,0.15));
  backdrop-filter: blur(8px);
  order: -1;
}

.a11y-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.a11y-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text);
  opacity: 0.7;
}

.a11y-buttons {
  display: flex;
  gap: 4px;
}

.a11y-buttons button,
.a11y-section > button {
  flex: 1;
  padding: 5px 8px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  font-size: 13px;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
  white-space: nowrap;
}

.a11y-buttons button:hover,
.a11y-section > button:hover {
  color: var(--text-h);
  border-color: var(--accent);
}

.a11y-buttons button.active,
.a11y-section > button.active {
  background: var(--accent-bg);
  border-color: var(--accent-border);
  color: var(--accent);
}

.size-indicator {
  font-family: var(--mono, monospace);
  font-size: 12px !important;
  min-width: 36px;
}
</style>
