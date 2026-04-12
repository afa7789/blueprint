import { createRouter, createWebHistory } from 'vue-router'
import Landing from '../views/Landing.vue'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import ForgotPassword from '../views/ForgotPassword.vue'
import { useAuthStore } from '../stores/auth'
import { fetchFeatureFlags, isFeatureEnabled } from '../services/featureFlags'

const routes = [
  {
    path: '/',
    name: 'Landing',
    component: Landing
  },
  {
    path: '/store',
    component: () => import('../views/Store/StoreLayout.vue'),
    children: [
      { path: '', name: 'StoreIndex', component: () => import('../views/Store/StoreIndex.vue') },
      { path: 'cart', name: 'Cart', component: () => import('../views/Store/Cart.vue') },
      { path: 'checkout', name: 'Checkout', component: () => import('../views/Store/Checkout.vue'), meta: { requiresAuth: true } },
      { path: 'orders', name: 'OrderHistory', component: () => import('../views/Store/OrderHistory.vue'), meta: { requiresAuth: true } },
      { path: ':id', name: 'ProductDetail', component: () => import('../views/Store/ProductDetail.vue') },
    ]
  },
  {
    path: '/blog',
    name: 'BlogIndex',
    component: () => import('../views/Blog/BlogIndex.vue'),
  },
  {
    path: '/blog/:slug',
    name: 'BlogPost',
    component: () => import('../views/Blog/BlogPost.vue'),
  },
  {
    path: '/legal/:slug',
    name: 'LegalPage',
    component: () => import('../views/LegalPage.vue'),
  },
  {
    path: '/linktree',
    name: 'Linktree',
    component: () => import('../views/Linktree.vue')
  },
  {
    path: '/brand-kit',
    name: 'BrandKit',
    component: () => import('../views/BrandKit.vue'),
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/register',
    name: 'Register',
    component: Register
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: ForgotPassword
  },
  {
    path: '/operator',
    name: 'OperatorDashboard',
    component: () => import('../views/Operator/OperatorDashboard.vue'),
    meta: { requiresAuth: true, requiresRole: 'operator' }
  },
  {
    path: '/user',
    component: () => import('../views/User/UserLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: '/user/profile' },
      { path: 'profile', name: 'UserProfile', component: () => import('../views/User/UserProfile.vue') },
      { path: 'password', name: 'UserPassword', component: () => import('../views/User/UserPassword.vue') },
      { path: 'cards', name: 'UserCards', component: () => import('../views/User/UserCards.vue') },
      { path: 'orders', name: 'UserOrders', component: () => import('../views/User/UserOrders.vue') },
    ]
  },
  {
    path: '/admin',
    component: () => import('../views/Admin/AdminLayout.vue'),
    meta: { requiresAuth: true, requiresRole: 'admin' },
    children: [
      { path: '', redirect: '/admin/users' },
      { path: 'users', name: 'AdminUsers', component: () => import('../views/Admin/AdminUsers.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'waitlist', name: 'AdminWaitlist', component: () => import('../views/Admin/AdminWaitlist.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'features', name: 'AdminFeatureFlags', component: () => import('../views/Admin/AdminFeatureFlags.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'banners', name: 'AdminBanners', component: () => import('../views/Admin/AdminBanners.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'linktree', name: 'AdminLinktree', component: () => import('../views/Admin/AdminLinktree.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'theme', name: 'AdminTheme', component: () => import('../views/Admin/AdminBrandKit.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'email-groups', name: 'AdminEmailGroups', component: () => import('../views/Admin/AdminEmailGroups.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'blog', name: 'AdminBlog', component: () => import('../views/Admin/AdminBlog.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'jobs', name: 'AdminJobs', component: () => import('../views/Admin/AdminJobs.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'tools', name: 'AdminTools', component: () => import('../views/Admin/AdminTools.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'logs', name: 'AdminLogs', component: () => import('../views/Admin/AdminLogs.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'audit', name: 'AdminAudit', component: () => import('../views/Admin/AdminAudit.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'legal', name: 'AdminLegal', component: () => import('../views/Admin/AdminLegal.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'security', name: 'AdminSecurity', component: () => import('../views/Admin/AdminSecurity.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'config', name: 'AdminConfig', component: () => import('../views/Admin/AdminConfig.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'products', name: 'AdminProducts', component: () => import('../views/Admin/AdminProducts.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'orders', name: 'AdminOrders', component: () => import('../views/Admin/AdminOrders.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'coupons', name: 'AdminCoupons', component: () => import('../views/Admin/AdminCoupons.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
      { path: 'payments', name: 'AdminPayments', component: () => import('../views/Admin/AdminPayments.vue'), meta: { requiresAuth: true, requiresRole: 'admin' } },
    ]
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to) => {
  await fetchFeatureFlags()

  if (to.path.startsWith('/store') && !isFeatureEnabled('store_enabled')) return { path: '/' }
  if (to.path.startsWith('/blog') && !isFeatureEnabled('blog_enabled')) return { path: '/' }
  if (to.path === '/linktree' && !isFeatureEnabled('linktree_enabled')) return { path: '/' }
  if (to.path === '/brand-kit' && !isFeatureEnabled('brand_kit_enabled')) return { path: '/' }

  const auth = useAuthStore()
  if (to.meta.requiresAuth) {
    if (!auth.isAuthenticated) {
      return { name: 'Login' }
    }
  }
  if (to.meta.requiresRole === 'admin') {
    if (!auth.isAdmin) {
      return { name: 'Landing' }
    }
  }
  if (to.meta.requiresRole === 'operator') {
    if (!auth.isOperator) {
      return { name: 'Landing' }
    }
  }
})

export default router
