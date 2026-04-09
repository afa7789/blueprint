import { ref } from 'vue'
import { api } from './api'

export interface FeatureFlag {
  name: string
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

export function isFeatureEnabled(name: string): boolean {
  const flag = featureFlags.value.find(f => f.name === name)
  return flag?.enabled ?? false
}