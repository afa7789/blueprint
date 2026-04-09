<script setup lang="ts">
import { RouterView } from 'vue-router'
import DynamicFooter from './components/common/DynamicFooter.vue'
import UpdateToast from './components/common/UpdateToast.vue'
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

async function logout() {
  await auth.logout()
  router.push('/')
}
</script>

<template>
  <div id="app">
    <nav class="top-nav">
      <div class="top-nav-right">
        <template v-if="auth.isAuthenticated">
          <router-link to="/user">My Account</router-link>
          <router-link v-if="auth.isAdmin" to="/admin">Admin</router-link>
          <button class="nav-btn" @click="logout">Logout</button>
        </template>
        <template v-else>
          <router-link to="/login">Login</router-link>
        </template>
      </div>
    </nav>
    <main>
      <RouterView />
    </main>
    <DynamicFooter />
    <UpdateToast />
  </div>
</template>

<style scoped>
.top-nav {
  display: flex;
  justify-content: flex-end;
  padding: 8px 16px;
  border-bottom: 1px solid var(--border);
  background: var(--bg);
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