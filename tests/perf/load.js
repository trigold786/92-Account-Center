import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  stages: [
    { target: 100, duration: '1m' },
    { target: 500, duration: '3m' },
    { target: 0, duration: '1m' },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.001'],
  },
}

const BASE_URL = __ENV.BASE_URL || 'http://localhost:30300'

const endpoints = [
  () => http.get(`${BASE_URL}/health`),
  () => http.post(`${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ phone: '13800138000', password: 'test123' }),
    { headers: { 'Content-Type': 'application/json' } }),
  () => http.get(`${BASE_URL}/api/v1/account/profile`,
    { headers: { 'Authorization': 'Bearer test_token' } }),
]

export default function () {
  const idx = Math.floor(Math.random() * endpoints.length)
  const res = endpoints[idx]()
  check(res, { 'status is 2xx': (r) => r.status >= 200 && r.status < 300 })
  sleep(0.5)
}
