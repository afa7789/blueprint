<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchFeatureFlags, isFeatureEnabled } from '../../services/featureFlags'

const linktreeEnabled = ref(false)
const storeEnabled = ref(false)

onMounted(async () => {
  await fetchFeatureFlags()
  linktreeEnabled.value = isFeatureEnabled('linktree')
  storeEnabled.value = isFeatureEnabled('store')
})
</script>

<template>
  <footer class="footer">
    <div class="footer-content">
      <div class="footer-links">
        <div v-if="storeEnabled" class="footer-section">
          <h3>Store</h3>
          <router-link to="/store">Browse Products</router-link>
        </div>
        <div v-if="linktreeEnabled" class="footer-section">
          <h3>Linktree</h3>
          <router-link to="/linktree">View Links</router-link>
        </div>
      </div>
      <div class="footer-bottom">
        <p>&copy; 2026. All rights reserved.</p>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.footer {
  border-top: 1px solid var(--border);
  padding: 32px 20px;
  text-align: left;
}

.footer-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.footer-links {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
}

.footer-section h3 {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--text-h);
}

.footer-section a {
  color: var(--text);
  text-decoration: none;
  font-size: 14px;
  transition: color 0.2s;
}

.footer-section a:hover {
  color: var(--accent);
}

.footer-bottom {
  font-size: 14px;
  color: var(--text);
}
</style>