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

  test('TC-GUI-003: SPA fallback works - /login returns 200', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/login`)
    expect(resp.status()).toBe(200)
  })

  test('TC-GUI-004: SPA fallback works - /register returns 200', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/register`)
    expect(resp.status()).toBe(200)
  })

  test('TC-GUI-005: SPA fallback works - all Vue routes return 200', async ({ request }) => {
    const routes = ['/login', '/register', '/account', '/credits', '/subscriptions', '/referral', '/devices', '/admin']
    for (const route of routes) {
      const resp = await request.get(`${UAT_BASE}${route}`)
      expect(resp.status(), `${route} should return 200 with SPA fallback`).toBe(200)
    }
  })

  test('TC-GUI-006: Login page renders with form elements', async ({ page }) => {
    await page.goto(`${UAT_BASE}/login`)
    await expect(page.locator('h2')).toContainText('登录')
    await expect(page.locator('input[placeholder="手机号 / 邮箱 / 账号"]')).toBeVisible()
    await expect(page.locator('input[placeholder="密码"]')).toBeVisible()
    await expect(page.locator('.el-tabs__content button:has-text("登录")').first()).toBeVisible()
  })

  test('TC-GUI-007: Login tab switch (password <-> code)', async ({ page }) => {
    await page.goto(`${UAT_BASE}/login`)
    await page.locator('[role="tab"]:has-text("验证码登录")').click()
    await expect(page.locator('input[placeholder="手机号"]')).toBeVisible()
    await page.locator('[role="tab"]:has-text("密码登录")').click()
    await expect(page.locator('input[placeholder="密码"]')).toBeVisible()
  })

  test('TC-GUI-008: Login validation - empty fields show warning', async ({ page }) => {
    await page.goto(`${UAT_BASE}/login`)
    await page.locator('.el-tabs__content button:has-text("登录")').first().click()
    await expect(page.locator('.el-message')).toBeVisible()
  })

  test('TC-GUI-009: Navigate to Register page via link', async ({ page }) => {
    await page.goto(`${UAT_BASE}/login`)
    await page.locator('text=还没有账号？立即注册').click()
    await expect(page).toHaveURL(/\/register/)
    await expect(page.locator('h2')).toContainText('注册')
  })

  test('TC-GUI-010: Register page renders with form elements', async ({ page }) => {
    await page.goto(`${UAT_BASE}/register`)
    await expect(page.locator('input[placeholder="手机号"]')).toBeVisible()
    await expect(page.locator('input[placeholder="密码（至少6位）"]')).toBeVisible()
    await expect(page.locator('input[placeholder="确认密码"]')).toBeVisible()
    await expect(page.locator('input[placeholder="验证码"]')).toBeVisible()
    await expect(page.locator('text=我已阅读并同意服务条款')).toBeVisible()
    await expect(page.locator('button:has-text("注册")')).toBeVisible()
  })

  test('TC-GUI-011: Register validation - password mismatch', async ({ page }) => {
    await page.goto(`${UAT_BASE}/register`)
    await page.locator('input[placeholder="手机号"]').fill('13800138000')
    await page.locator('input[placeholder="密码（至少6位）"]').fill('test123')
    await page.locator('input[placeholder="确认密码"]').fill('different')
    await page.locator('input[placeholder="验证码"]').fill('123456')
    await page.locator('button:has-text("注册")').click()
    await expect(page.locator('.el-message')).toContainText('密码不一致')
  })

  test('TC-GUI-012: Register validation - agree checkbox required', async ({ page }) => {
    await page.goto(`${UAT_BASE}/register`)
    await page.locator('input[placeholder="手机号"]').fill('13800138000')
    await page.locator('input[placeholder="密码（至少6位）"]').fill('test123')
    await page.locator('input[placeholder="确认密码"]').fill('test123')
    await page.locator('input[placeholder="验证码"]').fill('123456')
    await page.locator('button:has-text("注册")').click()
    await expect(page.locator('.el-message')).toContainText('同意')
  })

  test('TC-GUI-013: Navigate back to Login from Register', async ({ page }) => {
    await page.goto(`${UAT_BASE}/register`)
    await page.locator('text=已有账号？去登录').click()
    await expect(page).toHaveURL(/\/login/)
    await expect(page.locator('h2')).toContainText('登录')
  })

  test('TC-GUI-014: Protected routes redirect to login (SPA guard)', async ({ page }) => {
    const protectedRoutes = ['/', '/account', '/credits', '/subscriptions', '/referral', '/devices', '/admin']
    for (const route of protectedRoutes) {
      await page.goto(`${UAT_BASE}${route}`)
      await page.waitForTimeout(500)
      const url = page.url()
      expect(url, `${route} should redirect to /login`).toContain('/login')
    }
  })

  test('TC-GUI-015: NotFound page for invalid routes', async ({ page }) => {
    await page.goto(`${UAT_BASE}/nonexistent-page-test`)
    await expect(page.locator('text=404')).toBeVisible()
  })
})
