import { ref } from 'vue'

const needsUpdate = ref(false)
let updateFn: (() => Promise<void>) | null = null

export function useServiceWorker() {
  async function checkForUpdates() {
    if ('serviceWorker' in navigator) {
      const reg = await navigator.serviceWorker.getRegistration()
      if (reg?.waiting) {
        needsUpdate.value = true
      }
    }
  }

  function onNeedRefresh(fn: () => Promise<void>) {
    updateFn = fn
    needsUpdate.value = true
  }

  async function updateApp() {
    if (updateFn) await updateFn()
    window.location.reload()
  }

  return { needsUpdate, checkForUpdates, onNeedRefresh, updateApp }
}
