import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: '../e2e',
  timeout: 30000,
  retries: 1,
  use: {
    baseURL: 'http://localhost',
    trace: 'on-first-retry',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: {
    command: 'docker compose up -d --wait',
    url: 'http://localhost/healthz',
    reuseExistingServer: true,
    timeout: 60000,
    cwd: '..',
  },
})
