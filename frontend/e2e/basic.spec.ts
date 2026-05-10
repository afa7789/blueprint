import { test, expect } from '@playwright/test'

test('health endpoint returns 200', async ({ request }) => {
  const res = await request.get('/healthz')
  expect(res.status()).toBe(200)
})

test('frontend loads', async ({ page }) => {
  const response = await page.goto('/')
  expect(response).not.toBeNull()
  expect(response!.status()).toBeLessThan(400)
  await expect(page.locator('#app').first()).toBeVisible()
})
