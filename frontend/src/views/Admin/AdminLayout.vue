<template>
  <div class="admin-layout">
    <button class="sidebar-toggle" @click="sidebarOpen = !sidebarOpen" aria-label="Toggle sidebar">
      &#9776;
    </button>
    <aside class="sidebar" :class="{ open: sidebarOpen }">
      <nav>
        <h3>Admin</h3>
        <router-link to="/admin/users" @click="sidebarOpen = false">Users</router-link>
        <router-link to="/admin/features" @click="sidebarOpen = false">Feature Flags</router-link>
        <router-link to="/admin/banners" @click="sidebarOpen = false">Banners</router-link>
        <router-link to="/admin/linktree" @click="sidebarOpen = false">Linktree</router-link>
        <router-link to="/admin/brand-kit" @click="sidebarOpen = false">Brand Kit</router-link>
        <router-link to="/admin/email-groups" @click="sidebarOpen = false">Email Groups</router-link>
        <router-link to="/admin/jobs" @click="sidebarOpen = false">Jobs</router-link>
        <router-link to="/admin/tools" @click="sidebarOpen = false">Tools Hub</router-link>
        <router-link to="/admin/logs" @click="sidebarOpen = false">Logs</router-link>
        <router-link to="/admin/audit" @click="sidebarOpen = false">Audit</router-link>
      </nav>
    </aside>
    <main class="admin-main">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
const sidebarOpen = ref(false)
</script>

<style scoped>
.admin-layout {
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
  display: block;
  padding: 8px 12px;
  border-radius: 6px;
  color: var(--text);
  text-decoration: none;
  font-size: 14px;
  transition: background 0.15s, color 0.15s;
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

.admin-main {
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

  .admin-main {
    padding: 16px;
    padding-top: 52px;
  }
}
</style>
