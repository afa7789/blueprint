<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchFeatureFlags, isFeatureEnabled } from '../../services/featureFlags'

const show = ref(false)
const dismissed = ref(false)
let deferredPrompt: any = null

onMounted(async () => {
  // Check if already installed or dismissed
  if (window.matchMedia('(display-mode: standalone)').matches) return
  if (localStorage.getItem('pwa-install-dismissed')) return

  await fetchFeatureFlags()
  if (!isFeatureEnabled('pwa')) return

  window.addEventListener('beforeinstallprompt', (e: Event) => {
    e.preventDefault()
    deferredPrompt = e
    show.value = true
  })
})

async function install() {
  if (!deferredPrompt) return
  deferredPrompt.prompt()
  const { outcome } = await deferredPrompt.userChoice
  if (outcome === 'accepted') {
    show.value = false
  }
  deferredPrompt = null
}

function dismiss() {
  show.value = false
  dismissed.value = true
  localStorage.setItem('pwa-install-dismissed', '1')
}
</script>

<template>
  <Transition name="slide">
    <div v-if="show && !dismissed" class="pwa-prompt">
      <div class="pwa-content">
        <img src="/icon.svg" alt="" class="pwa-icon" />
        <div class="pwa-text">
          <strong>Install Blueprint</strong>
          <span>Add to home screen for a better experience</span>
        </div>
      </div>
      <div class="pwa-actions">
        <button class="pwa-install" @click="install">
          <i class="fas fa-download"></i> Install
        </button>
        <button class="pwa-dismiss" @click="dismiss" aria-label="Dismiss">
          <i class="fas fa-times"></i>
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.pwa-prompt {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
  background: var(--bg);
  border: 1px solid var(--accent-border);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  max-width: 420px;
  width: calc(100% - 32px);
}

.pwa-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pwa-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  flex-shrink: 0;
}

.pwa-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pwa-text strong {
  font-size: 14px;
  color: var(--text-h);
}

.pwa-text span {
  font-size: 12px;
  color: var(--text);
}

.pwa-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.pwa-install {
  padding: 8px 16px;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}

.pwa-install:hover {
  opacity: 0.9;
}

.pwa-dismiss {
  background: none;
  border: none;
  color: var(--text);
  cursor: pointer;
  padding: 4px;
  font-size: 14px;
}

.pwa-dismiss:hover {
  color: var(--text-h);
}

/* Transition */
.slide-enter-active, .slide-leave-active {
  transition: all 0.3s ease;
}
.slide-enter-from, .slide-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(20px);
}
</style>
