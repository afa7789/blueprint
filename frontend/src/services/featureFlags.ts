import { ref } from 'vue'
import { api } from './api'

export interface FeatureFlag {
  id: number
  key: string
  enabled: boolean
}

const featureFlags = ref<FeatureFlag[]>([])
let cached = false

export async function fetchFeatureFlags(force = false): Promise<FeatureFlag[]> {
  if (!force && cached && featureFlags.value.length > 0) {
    return featureFlags.value
  }

  try {
    const data = await api.get<FeatureFlag[]>('/api/v1/features')
    featureFlags.value = data
    cached = true
  } catch (error) {
    console.error('Failed to fetch feature flags:', error)
  }

  return featureFlags.value
}

export function setFeatureFlagEnabled(key: string, enabled: boolean) {
  const candidates = [key]
  if (!key.endsWith('_enabled')) {
    candidates.push(`${key}_enabled`)
  }

  const existing = featureFlags.value.find(flag => candidates.includes(flag.key))
  if (existing) {
    existing.enabled = enabled
    cached = true
    return
  }

  featureFlags.value.push({
    id: Date.now(),
    key: key.endsWith('_enabled') ? key : `${key}_enabled`,
    enabled,
  })
  cached = true
}

export function invalidateFeatureFlags() {
  cached = false
}

export function isFeatureEnabled(key: string): boolean {
  // Support both exact key and short name (e.g. "store" matches "store_enabled")
  const flag = featureFlags.value.find(f => f.key === key || f.key === key + '_enabled')
  return flag?.enabled ?? false
}

export function getFlags(): FeatureFlag[] {
  return featureFlags.value
}
