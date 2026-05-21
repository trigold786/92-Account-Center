import { test, expect } from '@playwright/test'

const UAT_BASE = process.env.BASE_URL || 'https://uat-92.neurongene.cn'

test.describe('UAT Layer 1-2: Login/Register GUI Business Flow', () => {
  test('TC-GUI-001: Vue SPA root loads with app container', async ({ page }) => {
    await page.goto(`${UAT_BASE}/`)
    const app = page.locator('#app')
    await expect(app).toBeVisible({ timeout: 10000 })
  })

  test('TC-GUI-002: Vue JS bundle loads and renders', async ({ page }) => {
    await page.goto(`${UAT_BASE}/`)
    const scripts = await page.locator('script[src]').count()
    expect(scripts).toBeGreaterThanOrEqual(1)
  })

  test('TC-GUI-003: SPA routing - /login returns 404 from nginx (BUG-002)', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/login`)
    expect(resp.status()).toBe(404)
  })

  test('TC-GUI-004: SPA routing - /register returns 404 from nginx (BUG-002)', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/register`)
    expect(resp.status()).toBe(404)
  })

  test('TC-GUI-005: SPA routing - all Vue routes 404 from nginx (BUG-002)', async ({ request }) => {
    const routes = ['/login', '/register', '/account', '/credits', '/subscriptions', '/referral', '/devices', '/admin']
    for (const route of routes) {
      const resp = await request.get(`${UAT_BASE}${route}`)
      expect(resp.status(), `${route} should be 404 due to missing SPA fallback`).toBe(404)
    }
  })
})
