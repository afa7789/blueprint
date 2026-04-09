<template>
  <div class="user-layout">
    <button class="sidebar-toggle" @click="sidebarOpen = !sidebarOpen" aria-label="Toggle sidebar">
      &#9776;
    </button>
    <aside class="sidebar" :class="{ open: sidebarOpen }">
      <nav>
        <h3>My Account</h3>
        <router-link to="/user/profile" @click="sidebarOpen = false"><i class="fas fa-user"></i> Profile</router-link>
        <router-link to="/user/password" @click="sidebarOpen = false"><i class="fas fa-lock"></i> Security</router-link>
        <router-link v-if="flagStripe" to="/user/cards" @click="sidebarOpen = false"><i class="fas fa-credit-card"></i> Saved Cards</router-link>
        <router-link to="/user/orders" @click="sidebarOpen = false"><i class="fas fa-box"></i> Orders</router-link>
      </nav>
    </aside>
    <main class="user-main">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchFeatureFlags, isFeatureEnabled } from '../../services/featureFlags'

const sidebarOpen = ref(false)
const flagStripe = ref(false)

onMounted(async () => {
  await fetchFeatureFlags()
  flagStripe.value = isFeatureEnabled('payments_stripe')
})
</script>

<style scoped>
.user-layout {
  display: flex;
  min-height: 100svh;
  text-align: left;
}

.sidebar {
  width: 200px;
  flex-shrink: 0;
  border-right: 1px solid var(--border);
  padding: 24px 0;
  background: var(--bg);
}

.sidebar nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0 16px;
}

.sidebar nav h3 {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text);
  margin: 0 0 12px;
  padding: 0;
}

.sidebar nav a {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 6px;
  color: var(--text);
  text-decoration: none;
  font-size: 14px;
  transition: background 0.15s, color 0.15s;
}

.sidebar nav a i {
  width: 16px;
  text-align: center;
  font-size: 13px;
  opacity: 0.7;
}

.sidebar nav a.router-link-active i {
  opacity: 1;
}

.sidebar nav a:hover {
  background: var(--code-bg);
  color: var(--text-h);
}

.sidebar nav a.router-link-active {
  background: var(--accent-bg);
  color: var(--accent);
  font-weight: 500;
}

.user-main {
  flex: 1;
  padding: 32px;
  overflow-x: auto;
}

.sidebar-toggle {
  display: none;
  position: fixed;
  top: 12px;
  left: 12px;
  z-index: 100;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 18px;
  color: var(--text-h);
}

@media (max-width: 768px) {
  .sidebar-toggle {
    display: block;
  }

  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    height: 100%;
    z-index: 99;
    transform: translateX(-100%);
    transition: transform 0.2s;
    box-shadow: var(--shadow);
    padding-top: 56px;
  }

  .sidebar.open {
    transform: translateX(0);
  }

  .user-main {
    padding: 16px;
    padding-top: 52px;
  }
}
</style>
