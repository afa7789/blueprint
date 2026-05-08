import { test, expect } from '@playwright/test'

test('health endpoint returns 200', async ({ request }) => {
  const res = await request.get('/healthz')
  expect(res.ok()).toBeTruthy()
})

test('frontend loads', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('body')).toBeVisible()
})
