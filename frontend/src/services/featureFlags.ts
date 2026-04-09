import { ref } from 'vue'
import { api } from './api'

export interface FeatureFlag {
  id: number
  key: string
  enabled: boolean
}

const featureFlags = ref<FeatureFlag[]>([])
let cached = false

export async function fetchFeatureFlags(): Promise<FeatureFlag[]> {
  if (cached && featureFlags.value.length > 0) {
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

export function isFeatureEnabled(key: string): boolean {
  // Support both exact key and short name (e.g. "store" matches "store_enabled")
  const flag = featureFlags.value.find(f => f.key === key || f.key === key + '_enabled')
  return flag?.enabled ?? false
}

export function getFlags(): FeatureFlag[] {
  return featureFlags.value
}
