import { reactive } from 'vue'
import { api, ApiError } from './api'
import { fetchFeatureFlags, isFeatureEnabled } from './featureFlags'

export const siteModules = reactive({
  ready: false,
  loading: false,
  waitlistEnabled: false,
  storeEnabled: false,
  blogEnabled: false,
  brandKitEnabled: false,
  linktreeEnabled: false,
  bannersEnabled: false,
  hasStoreContent: false,
  hasBlogContent: false,
})

let inFlight: Promise<typeof siteModules> | null = null

async function fetchTotal(path: string): Promise<number> {
  try {
    const data = await api.get<{ total?: number; data?: unknown[]; posts?: unknown[] }>(path)
    if (typeof data.total === 'number') return data.total
    if (Array.isArray(data.data)) return data.data.length
    if (Array.isArray(data.posts)) return data.posts.length
    return 0
  } catch (error) {
    if (error instanceof ApiError && (error.status === 403 || error.status === 404)) {
      return 0
    }
    throw error
  }
}

export async function loadSiteModules(force = false): Promise<typeof siteModules> {
  if (inFlight && !force) return inFlight

  if (inFlight && force) {
    await inFlight
  }

  inFlight = (async () => {
    siteModules.loading = true

    try {
      await fetchFeatureFlags(force)

      siteModules.waitlistEnabled = isFeatureEnabled('waitlist_enabled')
      siteModules.storeEnabled = isFeatureEnabled('store_enabled')
      siteModules.blogEnabled = isFeatureEnabled('blog_enabled')
      siteModules.brandKitEnabled = isFeatureEnabled('brand_kit_enabled')
      siteModules.linktreeEnabled = isFeatureEnabled('linktree_enabled')
      siteModules.bannersEnabled = isFeatureEnabled('banners_enabled')

      const [storeTotal, blogTotal] = await Promise.all([
        siteModules.storeEnabled ? fetchTotal('/api/v1/products?page=1&limit=1') : Promise.resolve(0),
        siteModules.blogEnabled ? fetchTotal('/api/v1/blog?page=1&limit=1') : Promise.resolve(0),
      ])

      siteModules.hasStoreContent = storeTotal > 0
      siteModules.hasBlogContent = blogTotal > 0
      siteModules.ready = true
    } catch (error) {
      console.error('Failed to load site modules:', error)
      siteModules.hasStoreContent = false
      siteModules.hasBlogContent = false
      siteModules.ready = true
    } finally {
      siteModules.loading = false
    }

    return siteModules
  })().finally(() => {
    inFlight = null
  })

  return inFlight
}

export function refreshSiteModules(): Promise<typeof siteModules> {
  return loadSiteModules(true)
}
