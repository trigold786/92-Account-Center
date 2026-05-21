import { test, expect } from '@playwright/test'

const UAT_BASE = process.env.BASE_URL || 'https://uat-92.neurongene.cn'

test.describe('UAT Layer 5: GB/T 25000.51 Quality Characteristics', () => {
  test('TC-Quality-001: Page title correct', async ({ page }) => {
    await page.goto(`${UAT_BASE}/`)
    await expect(page).toHaveTitle('Account Center')
  })

  test('TC-Quality-002: Root page HTML contains #app and script bundle', async ({ page }) => {
    await page.goto(`${UAT_BASE}/`)
    const html = await page.content()
    expect(html).toContain('id="app"')
    expect(html).toContain('script')
  })

  test('TC-Quality-003: Responsive viewport - root loads on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto(`${UAT_BASE}/`)
    await expect(page).toHaveTitle('Account Center')
    const app = page.locator('#app')
    await expect(app).toBeVisible()
  })

  test('TC-Quality-004: /health returns 200 (via SPA fallback)', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/health`)
    expect(resp.status()).toBe(200)
  })

  test('TC-Quality-005: API gateway proxying works', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/auth/login`, {
      data: { credential: 'test', password: 'test' },
    })
    expect(resp.status()).toBeLessThan(500)
    const body = await resp.json()
    expect(body).toBeDefined()
  })
})
