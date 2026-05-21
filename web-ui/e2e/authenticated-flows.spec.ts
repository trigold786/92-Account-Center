import { test, expect } from '@playwright/test'

const UAT_BASE = process.env.BASE_URL || 'https://uat-92.neurongene.cn'

test.describe('UAT Layer 2: Frontend-Backend Consistency via API', () => {
  test('TC-CON-API-001: /auth/login returns 401 (BUG-001: api-gateway not rebuilt)', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/auth/login`, {
      data: { credential: 'test', password: 'test' },
    })
    expect(resp.status()).toBe(401)
  })

  test('TC-CON-API-002: /account/register returns 401 (BUG-001)', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/account/register`, {
      data: { phone: '13800138000', password: 'test123', code: '123456' },
    })
    expect(resp.status()).toBe(401)
  })

  test('TC-CON-API-003: /sms/send returns 401 (BUG-001)', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/sms/send`, {
      data: { phone: '13800138000', scene: 'login' },
    })
    expect(resp.status()).toBe(401)
  })

  test('TC-CON-API-004: /auth/refresh returns 401 (BUG-001)', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/auth/refresh`, {
      data: { refresh_token: 'invalid_token' },
    })
    expect(resp.status()).toBe(401)
  })

  test('TC-CON-API-005: Protected routes return 401 without token', async ({ request }) => {
    const protectedRoutes = [
      { method: 'GET', path: '/api/v1/account/users' },
      { method: 'GET', path: '/api/v1/credits/1/account' },
      { method: 'GET', path: '/api/v1/subscriptions/1' },
      { method: 'GET', path: '/api/v1/referral/1/summary' },
      { method: 'GET', path: '/api/v1/device/user/1' },
      { method: 'GET', path: '/api/v1/risk/history/1' },
      { method: 'GET', path: '/api/v1/audit/logs' },
    ]
    for (const route of protectedRoutes) {
      const resp = await request.get(`${UAT_BASE}${route.path}`)
      expect(resp.status(), `${route.path} should be 401`).toBe(401)
    }
  })

  test('TC-CON-API-006: /health returns HTML via Traefik (SPA fallback)', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/health`)
    expect(resp.status()).toBe(200)
  })
})
