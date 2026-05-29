import { test, expect } from '@playwright/test'

const UAT_BASE = process.env.BASE_URL || 'https://uat-92.neurongene.cn'

const ADMIN = { credential: '13900000002', password: 'Admin1234!' }
const USER = { credential: '13900000001', password: 'Test1234!' }

async function loginAs(request: any, user: { credential: string; password: string }) {
  const resp = await request.post(`${UAT_BASE}/api/v1/auth/login`, {
    data: { credential: user.credential, password: user.password },
  })
  expect(resp.status()).toBe(200)
  const body = await resp.json()
  const d = body.data ?? body
  expect(d.access_token).toBeDefined()
  return d
}

function decodeJWT(token: string) {
  const payload = token.split('.')[1]
  const decoded = JSON.parse(Buffer.from(payload, 'base64url').toString())
  return decoded
}

test.describe('RBAC: Login Response Contains Roles', () => {
  test('RBAC-001: Admin login response includes roles field', async ({ request }) => {
    const d = await loginAs(request, ADMIN)
    expect(d.roles).toBeDefined()
    expect(Array.isArray(d.roles)).toBe(true)
    expect(d.roles).toContain('admin')
  })

  test('RBAC-002: Normal user login response includes roles field', async ({ request }) => {
    const d = await loginAs(request, USER)
    expect(d.roles).toBeDefined()
    expect(Array.isArray(d.roles)).toBe(true)
    expect(d.roles).toContain('user')
    expect(d.roles).not.toContain('admin')
  })

  test('RBAC-003: JWT payload contains roles claim', async ({ request }) => {
    const d = await loginAs(request, ADMIN)
    const claims = decodeJWT(d.access_token)
    expect(claims.roles).toBeDefined()
    expect(Array.isArray(claims.roles)).toBe(true)
    expect(claims.roles).toContain('admin')
  })

  test('RBAC-004: Normal user JWT does not contain admin role', async ({ request }) => {
    const d = await loginAs(request, USER)
    const claims = decodeJWT(d.access_token)
    expect(claims.roles).toBeDefined()
    expect(claims.roles).not.toContain('admin')
  })
})

test.describe('RBAC: Frontend UI Role-Based Visibility', () => {
  test('RBAC-005: Admin user sees 管理后台 menu item', async ({ page }) => {
    await page.goto(`${UAT_BASE}/login`)
    await page.locator('input[placeholder="手机号 / 邮箱 / 账号"]').fill(ADMIN.credential)
    await page.locator('input[placeholder="密码"]').fill(ADMIN.password)
    await page.locator('.el-tabs__content button:has-text("登录")').first().click()
    await page.waitForURL(/\/$/, { timeout: 10000 })
    await expect(page.locator('.el-menu-item:has-text("管理后台")')).toBeVisible({ timeout: 5000 })
  })

  test('RBAC-006: Normal user does NOT see 管理后台 menu item', async ({ page }) => {
    await page.goto(`${UAT_BASE}/login`)
    await page.locator('input[placeholder="手机号 / 邮箱 / 账号"]').fill(USER.credential)
    await page.locator('input[placeholder="密码"]').fill(USER.password)
    await page.locator('.el-tabs__content button:has-text("登录")').first().click()
    await page.waitForURL(/\/$/, { timeout: 10000 })
    await expect(page.locator('.el-menu-item:has-text("管理后台")')).not.toBeVisible({ timeout: 5000 })
  })
})

test.describe('RBAC: Permission API', () => {
  test('RBAC-007: Admin user has admin-level permissions', async ({ request }) => {
    const d = await loginAs(request, ADMIN)
    const resp = await request.get(`${UAT_BASE}/api/v1/config/users/${d.account_id}/permissions`, {
      headers: { Authorization: `Bearer ${d.access_token}` },
    })
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    const perms = body.data ?? body
    expect(Array.isArray(perms)).toBe(true)
    expect(perms).toContain('admin.user.manage')
    expect(perms).toContain('admin.credit.adjust')
    expect(perms).toContain('data.dashboard')
  })

  test('RBAC-008: Normal user has only self-service permissions', async ({ request }) => {
    const d = await loginAs(request, USER)
    const resp = await request.get(`${UAT_BASE}/api/v1/config/users/${d.account_id}/permissions`, {
      headers: { Authorization: `Bearer ${d.access_token}` },
    })
    expect(resp.status()).toBe(200)
    const body = await resp.json()
    const perms = body.data ?? body
    expect(Array.isArray(perms)).toBe(true)
    expect(perms).toContain('account.self')
    expect(perms).toContain('credits.self')
    expect(perms).not.toContain('admin.user.manage')
  })
})

test.describe('RBAC: Gateway Role Guard', () => {
  test('RBAC-009: Normal user gets 403 on admin routes', async ({ request }) => {
    const d = await loginAs(request, USER)
    const adminRoutes = [
      '/api/v1/audit/logs',
      '/api/v1/blacklist',
    ]
    for (const route of adminRoutes) {
      const resp = await request.get(`${UAT_BASE}${route}`, {
        headers: { Authorization: `Bearer ${d.access_token}` },
      })
      const body = await resp.json()
      expect(body.error, `${route} should 403 for normal user`).toBe('insufficient permissions')
    }
  })

  test('RBAC-010: Admin user can access admin routes', async ({ request }) => {
    const d = await loginAs(request, ADMIN)
    const resp = await request.get(`${UAT_BASE}/api/v1/audit/logs`, {
      headers: { Authorization: `Bearer ${d.access_token}` },
    })
    expect(resp.status()).toBeLessThan(500)
    const body = await resp.json()
    expect(body.error).not.toBe('insufficient permissions')
  })
})
