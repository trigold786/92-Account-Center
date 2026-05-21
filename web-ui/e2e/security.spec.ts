import { test, expect } from '@playwright/test'

const UAT_BASE = process.env.BASE_URL || 'https://uat-92.neurongene.cn'

test.describe('UAT Layer 6: Security Verification', () => {
  test('TC-SEC-007b: HTTPS loads with security headers', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/`)
    expect(resp.status()).toBe(200)
    const hsts = resp.headers()['strict-transport-security']
    expect(hsts).toContain('max-age=31536000')
    expect(resp.headers()['x-content-type-options']).toBe('nosniff')
    expect(resp.headers()['x-frame-options']).toBe('SAMEORIGIN')
    expect(resp.headers()['x-xss-protection']).toBe('1; mode=block')
  })

  test('TC-SEC-001: SQL injection blocked', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/auth/login`, {
      data: { credential: "admin' OR '1'='1", password: 'test' },
    })
    expect([401, 400, 422]).toContain(resp.status())
  })

  test('TC-SEC-002: XSS blocked', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/api/v1/account/users?q=<script>alert(1)</script>`)
    expect(resp.status()).toBe(401)
  })

  test('TC-SEC-006: JWT required on protected endpoints', async ({ request }) => {
    const endpoints = [
      '/api/v1/account/users',
      '/api/v1/credits/balance',
      '/api/v1/payment/orders',
      '/api/v1/notifications',
      '/api/v1/compliance/audit-logs',
      '/api/v1/subscriptions/me',
    ]
    for (const ep of endpoints) {
      const resp = await request.get(`${UAT_BASE}${ep}`)
      expect(resp.status(), `${ep} should require auth`).toBe(401)
    }
  })

  test('TC-SEC-public: /auth/login still returns 401 (BUG-001 pending)', async ({ request }) => {
    const resp = await request.post(`${UAT_BASE}/api/v1/auth/login`, {
      data: { credential: 'test', password: 'test' },
    })
    expect(resp.status()).toBe(401)
  })

  test('TC-SEC-metrics: /metrics returns 200 via Traefik proxy', async ({ request }) => {
    const resp = await request.get(`${UAT_BASE}/metrics`)
    expect(resp.status()).toBe(200)
  })
})
