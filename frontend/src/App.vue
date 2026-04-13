<script setup lang="ts">
import { RouterView } from 'vue-router'
import DynamicFooter from './components/common/DynamicFooter.vue'
import UpdateToast from './components/common/UpdateToast.vue'
import AccessibilityBar from './components/common/AccessibilityBar.vue'
import PwaInstallPrompt from './components/common/PwaInstallPrompt.vue'
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'
import { useTheme } from './composables/useTheme'
import { useAccessibility } from './composables/useAccessibility'
import { computed, onMounted } from 'vue'
import { loadSiteModules, siteModules } from './services/siteModules'

const auth = useAuthStore()
const router = useRouter()
const { loadTheme } = useTheme()
loadTheme()
const { init } = useAccessibility()
init()

const storeVisible = computed(() => siteModules.storeEnabled && siteModules.hasStoreContent)
const blogVisible = computed(() => siteModules.blogEnabled && siteModules.hasBlogContent)

onMounted(() => {
  loadSiteModules()
})

async function logout() {
  await auth.logout()
  router.push('/')
}
</script>

<template>
  <div id="app">
    <nav class="top-nav">
      <router-link to="/" class="top-nav-brand">
        <img src="/inverted-icon.svg" alt="Blueprint" class="brand-logo" />
        <span class="brand-name">Blueprint</span>
      </router-link>
      <div class="top-nav-center">
        <router-link v-if="storeVisible" to="/store"><i class="fas fa-store"></i> Store</router-link>
        <router-link v-if="blogVisible" to="/blog"><i class="fas fa-pen-to-square"></i> Blog</router-link>
      </div>
      <div class="top-nav-right">
        <template v-if="auth.isAuthenticated">
          <router-link to="/user"><i class="fas fa-user"></i> My Account</router-link>
          <router-link v-if="auth.isAdmin" to="/admin"><i class="fas fa-gauge-high"></i> Admin</router-link>
          <button class="nav-btn" @click="logout"><i class="fas fa-right-from-bracket"></i> Logout</button>
        </template>
        <template v-else>
          <router-link to="/login"><i class="fas fa-right-to-bracket"></i> Login</router-link>
        </template>
      </div>
    </nav>
    <main>
      <RouterView />
    </main>
    <DynamicFooter />
    <UpdateToast />
    <AccessibilityBar />
    <PwaInstallPrompt />
  </div>
</template>

<style scoped>
.top-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  border-bottom: 1px solid var(--border);
  background: var(--bg);
}

.top-nav-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
}

.brand-logo {
  height: 18px;
  width: auto;
}

.brand-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-h);
  letter-spacing: -0.3px;
}

.top-nav-center {
  display: flex;
  align-items: center;
  gap: 16px;
}

.top-nav-center a {
  font-size: 13px;
  color: var(--text);
  text-decoration: none;
  font-weight: 500;
}

.top-nav-center a:hover {
  color: var(--accent);
}

.top-nav-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.top-nav-right a {
  font-size: 13px;
  color: var(--text);
  text-decoration: none;
}

.top-nav-right a:hover {
  color: var(--text-h);
}

.nav-btn {
  background: none;
  border: none;
  font-size: 13px;
  color: var(--text);
  cursor: pointer;
  padding: 0;
}

.nav-btn:hover {
  color: var(--text-h);
}

#app {
  width: 1126px;
  max-width: 100%;
  margin: 0 auto;
  text-align: center;
  border-inline: 1px solid var(--border);
  min-height: 100svh;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

main {
  flex: 1;
  display: flex;
  flex-direction: column;
}
</style>
