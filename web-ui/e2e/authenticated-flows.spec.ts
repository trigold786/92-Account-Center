import { test, expect } from '@playwright/test'

const UAT_BASE = process.env.BASE_URL || 'https://uat-92.neurongene.cn'

test.describe('UAT Layer 2: Frontend-Backend Consistency via API', () => {
  test('TC-CON-API-001: /auth/login passes through to auth-service', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/auth/login`, {
      data: { credential: 'test', password: 'test' },
    })
    const body = await resp.json()
    expect(body.error).toBeDefined()
    expect(body.error).not.toBe('missing authorization header')
  })

  test('TC-CON-API-002: /account/register passes through to account-service', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/account/register`, {
      data: { phone: '13800138000', password: 'test123', code: '123456' },
    })
    const body = await resp.json()
    expect(body.error).toBeDefined()
    expect(body.error).not.toBe('missing authorization header')
  })

  test('TC-CON-API-003: /sms/send passes through to notification-service', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/sms/send`, {
      data: { phone: '13800138000', scene: 'login' },
    })
    const body = await resp.json()
    expect(body.error).not.toBe('missing authorization header')
  })

  test('TC-CON-API-004: /auth/refresh passes through to auth-service', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/auth/refresh`, {
      data: { refresh_token: 'invalid_token' },
    })
    const body = await resp.json()
    expect(body.error).not.toBe('missing authorization header')
  })

  test('TC-CON-API-005: Protected routes return JWT 401', async ({ request }) => {
    const protectedRoutes = [
      '/api/v1/account/users',
      '/api/v1/credits/1/account',
      '/api/v1/subscriptions/1',
      '/api/v1/referral/1/summary',
      '/api/v1/device/user/1',
      '/api/v1/risk/history/1',
      '/api/v1/audit/logs',
    ]
    for (const route of protectedRoutes) {
      const resp = await request.get(`${UAT_BASE}${route}`)
      const body = await resp.json()
      expect(body.error, `${route} should be JWT-blocked`).toBe('missing authorization header')
    }
  })

  test('TC-CON-API-006: /health returns 200', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/health`)
    expect(resp.status()).toBe(200)
  })
})
